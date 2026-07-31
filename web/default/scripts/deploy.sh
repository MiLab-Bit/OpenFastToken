#!/usr/bin/env bash
#
# deploy.sh — FastToken 安全发版脚本（零停机 / 优雅停机）
# ===================================================================
# 解决的问题：
#   之前每次重启都有 3~5s 空窗，期间 Nginx 仍把新请求打向正在关闭的端口，
#   连接被重置 -> 用户看到「注册失败 / 充值提交失败」。
#
# 本脚本的发版链路（planned deploy，完全无缝）：
#   1) 发版前先把 Nginx upstream 标为 down 并 reload
#      -> Nginx 立即停止给后端派新请求（新请求立即 502/可重试，而非撞关闭中的端口）
#   2) 等待在途请求与 Nginx 旧 worker 排空
#   3) systemctl restart（应用收到 SIGTERM 后 srv.Shutdown 等待在途请求处理完，<=10s 优雅退出）
#   4) 探活 /api/status 直到返回 200
#   5) 把 upstream 恢复为 up 并 reload
#
# 用法：
#   ./deploy.sh                  # 仅对当前二进制做「摘除 -> 优雅重启 -> 恢复」（常用于验证/配置热更）
#   ./deploy.sh /path/to/new     # 先把新二进制放到 /opt/fasttoken/fasttoken，再安全重启
#   ./deploy.sh --build          # 在服务器上用本地 Go 源码重新编译后再安全重启
#   ./deploy.sh /path/to/new --build
#
# 注意：必须在 root 下运行（需要 cp 二进制 / systemctl / nginx -s reload）。
# ===================================================================

set -uo pipefail

APP_DIR=/opt/fasttoken
BIN="$APP_DIR/fasttoken"
SVC=fasttoken
UPSTREAM_INC=/etc/nginx/conf.d/backend_servers.inc
PORT="${PORT:-3000}"
HEALTH_URL="http://127.0.0.1:${PORT}/api/status"
DRAIN_WAIT=4            # 摘除后等待在途请求排空（秒）
HEALTH_TIMEOUT=60       # 等待新进程就绪的最长秒数

NEW_BIN=""
DO_BUILD=0
for a in "$@"; do
  case "$a" in
    --build) DO_BUILD=1 ;;
    -*) echo "未知参数: $a" >&2; exit 2 ;;
    *) NEW_BIN="$a" ;;
  esac
done

log(){ echo "[$(date +%H:%M:%S)] $*"; }
die(){
  echo "[$(date +%H:%M:%S)] ERROR: $*" >&2
  apply_upstream up || true      # 兜底恢复流量，避免卡在 down 状态
  exit 1
}

# 写 upstream include 并 reload（best-effort，不自行 die，由调用方决定）
apply_upstream(){
  local state="$1"
  if [ "$state" = "down" ]; then
    printf "server 127.0.0.1:%s down;\n" "$PORT" > "$UPSTREAM_INC"
  else
    printf "server 127.0.0.1:%s max_fails=3 fail_timeout=15s;\n" "$PORT" > "$UPSTREAM_INC"
  fi
  if ! nginx -t >/dev/null 2>&1; then
    echo "[$(date +%H:%M:%S)] nginx -t 失败，回滚 include" >&2
    printf "server 127.0.0.1:%s max_fails=3 fail_timeout=15s;\n" "$PORT" > "$UPSTREAM_INC"
    return 1
  fi
  nginx -s reload 2>/dev/null
  log "Nginx upstream -> $state"
  return 0
}

wait_health(){
  local t=0
  until curl -fsS "$HEALTH_URL" >/dev/null 2>&1; do
    t=$((t+1))
    [ "$t" -ge "$HEALTH_TIMEOUT" ] && return 1
    sleep 1
  done
  return 0
}

# ---- 0. 前置检查 ----
command -v curl  >/dev/null 2>&1 || die "缺少 curl"
command -v nginx >/dev/null 2>&1 || die "缺少 nginx"
[ -f "$BIN" ] || die "找不到二进制 $BIN"

# ---- 1. 替换二进制（可选）----
if [ -n "$NEW_BIN" ]; then
  [ -f "$NEW_BIN" ] || die "新二进制不存在: $NEW_BIN"
  cp -f "$BIN" "$BIN.bak.deploy-$(date +%Y%m%d%H%M%S)"
  cp -f "$NEW_BIN" "$BIN"
  chmod +x "$BIN"
  log "已用 $NEW_BIN 替换二进制（旧版本已备份）"
fi

# ---- 2. 编译（可选）----
if [ "$DO_BUILD" = "1" ]; then
  command -v /usr/local/go/bin/go >/dev/null 2>&1 || die "服务器无 Go 工具链 (/usr/local/go/bin/go)"
  cp -f "$BIN" "$BIN.bak.build-$(date +%Y%m%d%H%M%S)"
  log "在服务器编译中 ..."
  ( cd "$APP_DIR" && /usr/local/go/bin/go build -o "$BIN" . ) || die "go build 失败"
  log "已在服务器编译新二进制"
fi

# ---- 3. 摘除流量（关键）----
apply_upstream down || die "Nginx 摘除失败，中止发版"
log "等待在途请求排空 ${DRAIN_WAIT}s ..."
sleep "$DRAIN_WAIT"

# ---- 4. 优雅重启 ----
log "systemctl restart $SVC（SIGTERM 优雅停机，在途请求处理完再退出）"
systemctl restart "$SVC" || die "systemctl restart 失败"

# ---- 5. 探活 ----
log "等待服务就绪（最多 ${HEALTH_TIMEOUT}s）..."
if ! wait_health; then
  die "新进程在 ${HEALTH_TIMEOUT}s 内未就绪"
fi
log "服务已就绪：$HEALTH_URL -> 200"

# ---- 6. 恢复流量 ----
apply_upstream up || log "警告：恢复 upstream 失败，请手动检查 $UPSTREAM_INC"
log "发版完成 ✅（链路：摘除 -> 优雅重启 -> 探活 -> 恢复）"
