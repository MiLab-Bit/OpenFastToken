package main

import (
	"bytes"
	"context"
	"embed"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"www.abc-ai.cn/FastToken/common"
	"www.abc-ai.cn/FastToken/common/circuitbreaker"
	"www.abc-ai.cn/FastToken/common/cooldown"
	"www.abc-ai.cn/FastToken/common/credential"
	semcache "www.abc-ai.cn/FastToken/common/semantic_cache"
	"www.abc-ai.cn/FastToken/common/weightedlb"
	"www.abc-ai.cn/FastToken/common/workerpool"
	"www.abc-ai.cn/FastToken/constant"
	"www.abc-ai.cn/FastToken/controller"
	"www.abc-ai.cn/FastToken/di"
	"www.abc-ai.cn/FastToken/i18n"
	"www.abc-ai.cn/FastToken/logger"
	"www.abc-ai.cn/FastToken/middleware"
	"www.abc-ai.cn/FastToken/model"
	"www.abc-ai.cn/FastToken/oauth"
	perfmetrics "www.abc-ai.cn/FastToken/pkg/perf_metrics"
	"www.abc-ai.cn/FastToken/relay"
	"www.abc-ai.cn/FastToken/router"
	"www.abc-ai.cn/FastToken/service"
	_ "www.abc-ai.cn/FastToken/setting/performance_setting"
	"www.abc-ai.cn/FastToken/setting/ratio_setting"

	"github.com/bytedance/gopkg/util/gopool"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	_ "net/http/pprof"
	"www.abc-ai.cn/FastToken/common/tracing"
)

//go:embed web/default/dist
var buildFS embed.FS

//go:embed web/default/dist/index.html
var indexPage []byte

// runPeriodic runs fn on each ticker tick, exiting when ctx is cancelled.
func runPeriodic(ctx context.Context, interval time.Duration, fn func()) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			common.SysLog("runPeriodic: shutting down")
			return
		case <-ticker.C:
			fn()
		}
	}
}

func main() {
	startTime := time.Now()

	// 创建全局应用上下文，统一管理所有后台 goroutine 生命周期
	appCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// 初始化 OpenTelemetry 追踪器
	tp, err := tracing.InitTracer()
	if err != nil {
		log.Fatalf("Failed to initialize tracer: %v", err)
	}
	defer tracing.ShutdownTracer(tp)

	err = InitResources(appCtx)
	if err != nil {
		common.FatalLog("failed to initialize resources: " + err.Error())
		return
	}

	// 初始化断路器管理器
	if common.CircuitBreakerEnabled {
		circuitbreaker.InitManager(
			common.CircuitBreakerFailureThreshold,
			common.CircuitBreakerSuccessThreshold,
			time.Duration(common.CircuitBreakerTimeout)*time.Second,
		)
		common.SysLog("Circuit breaker manager initialized")
	}

	// 初始化异步 Worker Pool
	workerpool.Init(10, 1000) // 10 workers, 1000 queue size
	common.SysLog("Async worker pool initialized (10 workers, 1000 queue)")

	// 初始化 Provider 全局冷却管理器 (Phase 2: 429/5xx 触发全 Provider 级别冷却)
	cooldown.Init(nil)

	// 初始化 Key Health Tracker (Phase 3: 跟踪 API Key 健康状态并自动封禁)
	credential.InitTracker(0, 0) // use default thresholds

	// 初始化动态加权负载均衡器
	weightedlb.InitWeightedLB()
	common.SysLog("Dynamic weighted load balancer initialized")

	common.SysLog("FastToken " + common.Version + " started")
	common.SysLog(fmt.Sprintf("Embedded index page: %d bytes, buildFS files: checking...", len(indexPage)))
	if os.Getenv("GIN_MODE") != "debug" {
		gin.SetMode(gin.ReleaseMode)
	}
	if common.DebugEnabled {
		common.SysLog("running in debug mode")
	}

	defer func() {
		err := model.CloseDB()
		if err != nil {
			common.FatalLog("failed to close database: " + err.Error())
		}
	}()

	defer workerpool.Shutdown()

	if common.RedisEnabled {
		// for compatibility with old versions
		common.MemoryCacheEnabled = true
	}
	if common.MemoryCacheEnabled {
		common.SysLog("memory cache enabled")
		common.SysLog(fmt.Sprintf("sync frequency: %d seconds", common.SyncFrequency))

		// Add panic recovery and retry for InitChannelCache
		func() {
			defer func() {
				if r := recover(); r != nil {
					common.SysLog(fmt.Sprintf("InitChannelCache panic: %v, retrying once", r))
					// Retry once
					_, _, fixErr := model.FixAbility()
					if fixErr != nil {
						common.FatalLog(fmt.Sprintf("InitChannelCache failed: %s", fixErr.Error()))
					}
				}
			}()
			model.InitChannelCache()
		}()

		go runPeriodic(appCtx, time.Duration(common.SyncFrequency)*time.Second, model.InitChannelCache)
	}

	// 热更新配置（受 appCtx 控制，服务关闭时自动退出）
	go runPeriodic(appCtx, time.Duration(common.SyncFrequency)*time.Second, func() { model.LoadOptionsFromDatabase() })

	// 数据看板
	go runPeriodic(appCtx, time.Duration(common.DataExportInterval)*time.Minute, model.SaveQuotaDataCache)

	if os.Getenv("CHANNEL_UPDATE_FREQUENCY") != "" {
		frequency, err := strconv.Atoi(os.Getenv("CHANNEL_UPDATE_FREQUENCY"))
		if err != nil {
			common.FatalLog("failed to parse CHANNEL_UPDATE_FREQUENCY: " + err.Error())
		}
		go runPeriodic(appCtx, time.Duration(frequency)*time.Minute, func() {
			controller.UpdateAllChannelBalances()
		})
	}

	go controller.AutomaticallyTestChannels()

	// Codex credential auto-refresh check every 10 minutes, refresh when expires within 1 day
	service.StartCodexCredentialAutoRefreshTask()

	// Wire task polling adaptor factory (breaks service -> relay import cycle)
	service.GetTaskAdaptorFunc = func(platform constant.TaskPlatform) service.TaskPollingAdaptor {
		a := relay.GetTaskAdaptor(platform)
		if a == nil {
			return nil
		}
		return a
	}

	// Channel upstream model update check task
	controller.StartChannelUpstreamModelUpdateTask()

	if common.IsMasterNode && constant.UpdateTask {
		gopool.Go(func() {
			controller.UpdateMidjourneyTaskBulk()
		})
		gopool.Go(func() {
			controller.UpdateTaskBulk()
		})
	}
	if os.Getenv("BATCH_UPDATE_ENABLED") == "true" {
		common.BatchUpdateEnabled = true
		common.SysLog("batch update enabled with interval " + strconv.Itoa(common.BatchUpdateInterval) + "s")
		model.InitBatchUpdater()
	}

	if os.Getenv("ENABLE_PPROF") == "true" {
		// Bind to localhost only and optionally require basic auth
		handler := http.Handler(http.DefaultServeMux)
		pprofPassword := os.Getenv("PPROF_PASSWORD")
		if pprofPassword != "" {
			handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_, pass, ok := r.BasicAuth()
				if !ok || pass != pprofPassword {
					w.Header().Set("WWW-Authenticate", `Basic realm="pprof"`)
					w.WriteHeader(http.StatusUnauthorized)
					return
				}
				http.DefaultServeMux.ServeHTTP(w, r)
			})
		}
		gopool.Go(func() {
			log.Println(http.ListenAndServe("127.0.0.1:8005", handler))
		})
		go common.Monitor()
		common.SysLog("pprof enabled on 127.0.0.1:8005")
	}

	err = common.StartPyroScope()
	if err != nil {
		common.SysError(fmt.Sprintf("start pyroscope error : %v", err))
	}

	// Initialize HTTP server
	server := gin.New()
	server.Use(gin.CustomRecovery(func(c *gin.Context, err any) {
		common.SysLog(fmt.Sprintf("panic detected: %v", err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"message": fmt.Sprintf("Panic detected, error: %v. Please submit a issue here: https://www.abc-ai.cn/FastToken", err),
				"type":    "FastToken_panic",
			},
		})
	}))
	// This will cause SSE not to work!!!
	//server.Use(gzip.Gzip(gzip.DefaultCompression))
	server.Use(middleware.RequestId())
	server.Use(middleware.PoweredBy())
	server.Use(middleware.I18n())
	middleware.SetUpLogger(server)
	// Initialize session store
	store := cookie.NewStore([]byte(common.SessionSecret))
	store.Options(sessions.Options{
		Path:     "/",
		MaxAge:   2592000, // 30 days
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode, // 修改为 Lax 模式，允许同站请求
	})
	server.Use(sessions.Sessions("session", store))

	// Health check endpoint (k8s/lb probe, diagnostics)
	server.GET("/health", func(c *gin.Context) {
		err := model.PingDB()
		if err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status":  "unhealthy",
				"db":      err.Error(),
				"version": common.Version,
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"version": common.Version,
		})
	})


	// 微信公众号网站归属验证文件
	server.GET("/4ff871073676d3e06648043e143771cb.txt", func(c *gin.Context) {
		c.String(http.StatusOK, "b22233c4a28beb32ef5ad63cad946bc208d87342")
	})
	InjectUmamiAnalytics()
	InjectGoogleAnalytics()

	// 启动期支付密钥自检（非致命；结果见 /api/payment/status，供部署冒烟判定）
	controller.InitPaymentAtStartup()

	// 设置路由
	router.SetRouter(server, router.ThemeAssets{
		DefaultBuildFS:   buildFS,
		DefaultIndexPage: indexPage,
		// ClassiBuildFS:   classicBuildFS,
		// // ClassicIndexPage: classicIndexPage,
	})
	var port = os.Getenv("PORT")
	if port == "" {
		port = strconv.Itoa(*common.Port)
	}

	// Log startup success message
	common.LogStartupSuccess(startTime, port)

	// Graceful shutdown: use appCtx (created at startup) for coordinated shutdown
	srv := &http.Server{
		Addr:    "127.0.0.1:" + port,
		Handler: server,
	}

	// Start server in background
	go func() {
		common.SysLog(fmt.Sprintf("HTTP server listening on 127.0.0.1:%s", port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			common.FatalLog("failed to start HTTP server: " + err.Error())
		}
	}()

	// Wait for shutdown signal (via appCtx)
	<-appCtx.Done()
	common.SysLog("Shutting down server...")

	// Create a deadline for graceful shutdown; all background goroutines will also stop via appCtx
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		common.FatalLog("server forced to shutdown: " + err.Error())
	}

	common.SysLog("Server exited gracefully")
}

func InjectUmamiAnalytics() {
	analyticsInjectBuilder := &strings.Builder{}
	if os.Getenv("UMAMI_WEBSITE_ID") != "" {
		umamiSiteID := os.Getenv("UMAMI_WEBSITE_ID")
		umamiScriptURL := os.Getenv("UMAMI_SCRIPT_URL")
		if umamiScriptURL == "" {
			umamiScriptURL = "https://analytics.umami.is/script.js"
		}
		analyticsInjectBuilder.WriteString("<script defer src=\"")
		analyticsInjectBuilder.WriteString(umamiScriptURL)
		analyticsInjectBuilder.WriteString("\" data-website-id=\"")
		analyticsInjectBuilder.WriteString(umamiSiteID)
		analyticsInjectBuilder.WriteString("\"></script>")
	}
	analyticsInjectBuilder.WriteString("<!--Umami FastToken-->\n")
	analyticsInject := []byte(analyticsInjectBuilder.String())
	placeholder := []byte("<!--umami-->\n")
	indexPage = bytes.ReplaceAll(indexPage, placeholder, analyticsInject)
}

func InjectGoogleAnalytics() {
	analyticsInjectBuilder := &strings.Builder{}
	if os.Getenv("GOOGLE_ANALYTICS_ID") != "" {
		gaID := os.Getenv("GOOGLE_ANALYTICS_ID")
		// Google Analytics 4 (gtag.js)
		analyticsInjectBuilder.WriteString("<script async src=\"https://www.googletagmanager.com/gtag/js?id=")
		analyticsInjectBuilder.WriteString(gaID)
		analyticsInjectBuilder.WriteString("\"></script>")
		analyticsInjectBuilder.WriteString("<script>")
		analyticsInjectBuilder.WriteString("window.dataLayer = window.dataLayer || [];")
		analyticsInjectBuilder.WriteString("function gtag(){dataLayer.push(arguments);}")
		analyticsInjectBuilder.WriteString("gtag('js', new Date());")
		analyticsInjectBuilder.WriteString("gtag('config', '")
		analyticsInjectBuilder.WriteString(gaID)
		analyticsInjectBuilder.WriteString("');")
		analyticsInjectBuilder.WriteString("</script>")
	}
	analyticsInjectBuilder.WriteString("<!--Google Analytics FastToken-->\n")
	analyticsInject := []byte(analyticsInjectBuilder.String())
	placeholder := []byte("<!--Google Analytics-->\n")
	indexPage = bytes.ReplaceAll(indexPage, placeholder, analyticsInject)
}

func InitResources(ctx context.Context) error {
	// Initialize resources here if needed
	// This is a placeholder function for future resource initialization
	err := godotenv.Load(".env")
	if err != nil {
		if common.DebugEnabled {
			common.SysLog("No .env file found, using default environment variables. If needed, please create a .env file and set the relevant variables.")
		}
	}

	// 加载环境变量
	common.InitEnv()

	logger.SetupLogger()

	// Initialize model settings
	ratio_setting.InitRatioSettings()

	service.InitHttpClient()

	service.InitTokenEncoders()

	// Initialize SQL Database
	err = model.InitDB()
	if err != nil {
		common.FatalLog("failed to initialize database: " + err.Error())
		return err
	}

	model.CheckSetup()

	// Initialize options, should after model.InitDB()
	model.InitOptionMap()

	// 配置即数据：把编译期默认值 seed 进库并加载定价/本地化（免部署改配置）
	model.InitConfigData()

	// Periodic cleanup: expire stale pending topup orders every 10 minutes
	go func() {
		ticker := time.NewTicker(time.Duration(model.GetIntOption("PendingTopupCleanupIntervalMin", 10)) * time.Minute)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				expired, err := model.ExpireStaleTopUps()
				if err != nil {
					common.SysLog(fmt.Sprintf("topup expiration cleanup failed: %s", err.Error()))
				} else if expired > 0 {
					common.SysLog(fmt.Sprintf("topup expiration cleanup: %d stale order(s) expired", expired))
				}
			}
		}
	}()

	// Periodic Alipay reconciliation: auto-complete paid-but-not-notified orders every 60s.
	// Acts as a safety net so top-ups arrive even if Alipay async notify is missed.
	go func() {
		ticker := time.NewTicker(60 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				items, err := controller.ReconcilePendingAlipayOrders()
				if err != nil {
					common.SysLog(fmt.Sprintf("alipay reconcile failed: %s", err.Error()))
					continue
				}
				for _, it := range items {
					if it.Completed {
						common.SysLog(fmt.Sprintf("alipay reconcile: auto-completed out_trade_no=%s user_id=%d", it.OutTradeNo, it.UserId))
					}
				}
			}
		}
	}()

	// Periodic Wechat reconciliation: auto-complete paid-but-not-notified orders every 60s.
	// Acts as a safety net so top-ups arrive even if Wechat async notify is missed.
	go func() {
		ticker := time.NewTicker(60 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				items, err := controller.ReconcilePendingWechatOrders()
				if err != nil {
					common.SysLog(fmt.Sprintf("wechat reconcile failed: %s", err.Error()))
					continue
				}
				for _, it := range items {
					if it.Completed {
						common.SysLog(fmt.Sprintf("wechat reconcile: auto-completed out_trade_no=%s user_id=%d", it.OutTradeNo, it.UserId))
					}
				}
			}
		}
	}()

	// 清理旧的磁盘缓存文件
	common.CleanupOldCacheFiles()

	// 初始化模型
	model.GetPricing()

	// Initialize SQL Database
	err = model.InitLogDB()
	if err != nil {
		return err
	}

	// Initialize Redis
	err = common.InitRedisClient()
	if err != nil {
		return err
	}

	// Initialize Repository DI (Phase C: Repository pattern)
	err = di.Init()
	if err != nil {
		return err
	}

	// Initialize semantic cache (Phase 1: Redis-based exact-match cache)
	semcache.Init(nil)

	perfmetrics.Init()

	// 启动系统监控
	common.StartSystemMonitor()

	// init sms service (from db options table, not env vars)
	smsProvider := common.SMSProvider
	if smsProvider == "" {
		smsProvider = "placeholder"
	}
	if smsProvider != "placeholder" {
		// aliyun requires all config fields present
		if smsProvider == "aliyun" {
			akId := common.SMSAccessKeyId
			akSecret := common.SMSAccessKeySecret
			tmpl := common.SMSTemplateCode
			sign := common.SMSSignName
			if akId == "" || akSecret == "" || tmpl == "" || sign == "" {
				common.SysError("SMS provider is aliyun but config in DB is incomplete, fallback to placeholder")
				smsProvider = "placeholder"
			} else {
				service.InitSMSService(service.SMSConfig{
					Provider:   "aliyun",
					SecretId:   akId,
					SecretKey:  akSecret,
					TemplateId: tmpl,
					SignName:   sign,
				})
			}
		}
	}
	// default placeholder
	if smsProvider == "placeholder" {
		service.InitSMSService(service.SMSConfig{Provider: "placeholder"})
	}

	// 定时清理过期验证码（每小时执行一次）
	gopool.Go(func() {
		for {
			time.Sleep(1 * time.Hour)
			if err := model.CleanExpiredSMSCodes(); err != nil {
				common.SysError("clean expired SMS codes failed: " + err.Error())
			} else {
				common.SysLog("expired SMS codes cleaned")
			}
		}
	})

	// Initialize i18n
	err = i18n.Init()
	if err != nil {
		common.SysError("failed to initialize i18n: " + err.Error())
		// Don't return error, i18n is not critical
	} else {
		common.SysLog("i18n initialized with languages: " + strings.Join(i18n.SupportedLanguages(), ", "))
	}
	// Register user language loader for lazy loading
	i18n.SetUserLangLoader(model.GetUserLanguage)

	// Load custom OAuth providers from database
	err = oauth.LoadCustomProviders()
	if err != nil {
		common.SysError("failed to load custom OAuth providers: " + err.Error())
		// Don't return error, custom OAuth is not critical
	}

	return nil
}
