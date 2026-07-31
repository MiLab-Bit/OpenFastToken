# FastToken 项目交接清单（Handover）

> 状态基准时间：**2026-07-19 10:57 (GMT+8)** — 所有事实均来自服务器实时核查。
> 更新记录：2026-07-22 — 经服务器复核，§3.4/§8 的「智企惠品牌区块 i18n 外部化」实际已于 2026-07-19 晚间完成并 2026-07-20 部署，已从待办移除；`deploy.sh` 已增强前置校验（见 §4.4）。本地运维通道目录与 venv 当日实际缺失，已重建（见 §1）。2026-07-29 — 主题/皮肤系统合并并白版锁定、配色可读性修复、i18n 补齐、服务器清理（见 §3.5 / §4.5 / §变更记录 2026-07-29）。
> 适用对象：接手本项目的人类运维 / 其他 AI Agent。
> 目标：无需翻聊天记录，照本清单即可登录、运维、部署、排障。

---

## 0. 一句话现状

FastToken 是一个基于 New API（`github.com/QuantumNous/new-api`）魔改的统一 AI 网关。生产环境为**单台阿里云 ECS**，Go 二进制自包含前端 `dist`（编译期 `go:embed`），nginx 做 TLS + 限流反代。当前线上版本已含 i18n 修复、TDZ 崩溃修复、死代码清理，**运行正常**。

---

## 1. 服务器信息与登录

| 项 | 值 |
|---|---|
| 公网 IP | `47.103.102.36` |
| 系统 | Alibaba Cloud Linux 3 (OpenAnolis)，内核 `5.10.134-18.al8.x86_64` |
| 登录用户 | `root` |
| 登录方式 | **密码登录，无 SSH 密钥** |
| 密码 | `***REDACTED***`（保密，勿明文写入任何仓库/日志） |
| SSH 端口 | 22 |
| 域名 | `www.abc-ai.cn` / `abc-ai.cn`（均解析到该 IP） |
| 工具链 | Go `1.25.0`、Node `v22.22.2`、nginx、`systemd` |

**首次登录后建议**：改 root 密码、或改用 SSH 密钥（当前无密钥，仅靠口令，风险较高）。

**本地运维通道（Agent 用）**：`C:/Users/Administrator/AppData/Local/Temp/wb-ft/` 下有：
- `rsh.py` — 基于 paramiko 的远程命令助手
- `put_file.py` — SFTP 上传
- `get_file.py` — SFTP 下载（2026-07-22 补充）
- 用托管的 venv python 运行：`C:/Users/Administrator/.workbuddy/binaries/python/envs/default/Scripts/python.exe rsh.py "命令"`（venv 需含 paramiko；2026-07-22 该目录与 venv 实际缺失，已重建）
- ⚠️ **禁止在该目录放名为 `inspect.py` 的文件**——会遮蔽 Python 标准库 `inspect`，导致 rsh/put_file 的 paramiko import 失败。

---

## 2. 架构与网络拓扑

```
浏览器 / 客户端
      │  HTTPS :443  (TLS, 限流, 安全头)
      ▼
   nginx  (反向代理, upstream = fasttoken_backend)
      │  HTTP :80  → 301 跳转 HTTPS
      ▼
   fasttoken  (Go 二进制, 监听 127.0.0.1:3000, 自含前端 dist)
      ├── Postgres  (systemd: postgresql.service)   ← SQL_DSN
      └── Redis     (systemd: redis.service)        ← REDIS_CONN_STRING
```

- **nginx 关键配置** `/etc/nginx/conf.d/fasttoken.conf`：
  - `:80` 全量 301 跳 HTTPS；`:443` 启 `http2`、TLS1.2/1.3。
  - 限流：`/api/` `/v1/` `/pg/` 走 `api_limit`；`/api/user/register` 走 `register_limit`。
  - 安全头：`X-Frame-Options` / `X-Content-Type-Options` / `X-XSS-Protection`。
  - 阻断敏感路径对外（`.env`、`.git*`、`FastToken.db`、`fasttoken_linux` 等返回 404）。
  - 支付回调白名单：`/api/alipay/notify`、`/api/wechat/notify`（长超时）。
  - SSL 证书：`/etc/ssl/certs/www.abc-ai.cn.crt` + `.key`。
- **systemd 单元** `fasttoken.service`：`After=network.target postgresql.service redis.service`，`Restart=always`，`EnvironmentFile=/opt/fasttoken/.env`，`ExecStart=/opt/fasttoken/fasttoken`。
- **无 Git、无源码备份**（用户明确选择不备份）。每次部署由 `deploy.sh` 自动生成回滚二进制 `fasttoken.bak.<TS>`。

---

## 3. FastToken 项目现状

### 3.1 代码位置

| 组件 | 路径 |
|---|---|
| Go 后端 + 二进制 | `/opt/fasttoken/`（二进制 `fasttoken`，`go:embed web/default/dist`） |
| 前端源码 | `/opt/fasttoken/web/default/`（rsbuild `v2.1.5`，包管理 npm/pnpm） |
| 构建产物 dist | `/opt/fasttoken/web/default/dist/`（被 `go:embed` 嵌入二进制） |
| 部署脚本 | `/opt/fasttoken/deploy.sh` |
| 环境变量 | `/opt/fasttoken/.env`（`chmod 600`） |
| i18n 语言包 | `/opt/fasttoken/web/default/src/l10n/locales/{en,zh,ar,fr,ja,ru,vi}.json` |

### 3.2 当前线上版本（已验证，2026-07-29）

- 服务 `fasttoken`：**active**（PID 2901967，2026-07-29 08:45 部署）
- 前端产物：`index.f94cb0a0.js` + `index.3a797703.css`（构建于 2026-07-29，`data-skin` 皮肤选择器 5 套已打入，暗色套已彻底移除）
- 主题/皮肤系统：**白版锁定**（见 §3.5），默认皮肤 `neo`
- 健康检查：`GET /api/payment/status` → `{"ready":true,...}`
- HTTPS：正常（200）
- 回滚点：`fasttoken.bak.20260729084530`（以及 `fasttoken.bak.20260729082916`）

### 3.3 今日已落地的修复（2026-07-19）

1. **i18n 解析修复** — `src/l10n/config.ts` 初始化加 `keySeparator: false`。
   所有 locale key 都是**扁平点号字符串**（如 `"about.featureUnifiedGateway.title"`、`"currency.cny"`，共 5453 个）。默认 `keySeparator: '.'` 会把 `about.featureXxx.title` 当嵌套路径去找 → 找不到 → 回退显示原始 key 字符串（看起来像"翻译缺失"）。`keySeparator: false` 后按字面解析，全部正常。**此项绝不能去掉，否则 5453 个 key 集体失效。**
2. **TDZ 崩溃修复** — `src/l10n/config.ts`：顶层 `loadI18nOverrides(getInitialLanguage())` 调用被移到 `const i18nOverridesLoaded` 声明之后。原先"调用早于声明"触发 `Cannot access 'f' before initialization`，整页崩溃。
3. **死代码清理（A类+B类全清）** — catch 未用变量、废弃 import/函数/变量：ESLint `no-unused-vars` 警告 **0**，解析错误 **0**。
4. **三处硬编码中文外部化（i18n）**：
   - `src/features/about/index.tsx`：产品优势卡片、Core Functions 卡片、品牌 tab → `t('about.*')`
   - `src/features/playground/components/chat/playground-empty-state.tsx`：3 个 badge → `t('playground.empty*')`
   - 7 个语言包均已补齐 `about.*` / `playground.empty.*` key（中文正式，其余英文兜底）。

### 3.4 已知未完项（非阻塞）

- ~~**智企惠品牌区块**~~：**已完成**（2026-07-19 晚间）。源码 `about/index.tsx` 已全程改用 `t('about.brand.*')` / `t('about.copyright.*')`；7 个语言包 `translation` 命名空间下 48 个 `about.*` key 全部补全且无空值（ar/fr/ja/ru/vi 为真翻译非英文兜底）；线上 bundle `index.1620b29b.js`（2026-07-20 00:38）已含对应文案。原「t('中文') 直译」描述已过时，切勿据此重复劳动。

---

### 3.5 主题 / 皮肤系统现状（2026-07-29 起，重要）

本轮把"主题"(明暗)与"皮肤"(配色)两套系统**合并为一套皮肤系统**，并**锁定为白版（浅色，禁止暗色）**。

- **入口归一**：顶栏 Palette 图标下拉 = 皮肤切换器（`skin-switch.tsx`，复用 `useSkin()`），列出 5 套：`neo`(默认/现代SaaS/蓝)、`aurora`(紫调科技)、`classic`(暖纸/企业)、`midnight`(海军蓝)、`sunset`(橙)。原"浅色/深色/系统"三色明暗入口已从顶栏、命令面板、错误页、设置面板**全入口移除**。
- **白版锁定（根因修复）**：旧 `theme-provider.tsx` 默认跟随系统/浏览器深色偏好，给 `<html>` 加 `dark` 类，触发暗色套 → 整站变黑。现改为永远 `remove('dark')` + `add('light')`，并删除 `skins.css` 全部 `[data-skin=x].dark` 暗色块。结论：**全站固定浅色，无明暗切换**。
- **配色可读性修复**：审计发现 38 个页面文件写死固定颜色（`bg-white`、`text-gray-*`、`text-slate-*` 等 100+ 处），不吃主题变量 → 切皮肤时"白卡片浅灰字/深字深底"看不清。已用脚本统一换语义 token（`bg-card`/`text-foreground`/`text-muted-foreground` 等）；另修品牌色背景配白字 2 处对比度 bug。首页营销深色面板、支付页、二维码白底、黑底日志终端**刻意保留固定色**（设计需要）。
- **i18n**：7 语言包已补齐 `Theme` / `Select theme` / `Theme Settings` 及 5 套皮肤的 `skin.<id>.{name,description}`（无 fallback，必须全存在）。
- **文件地图（本次改动）**：
  - `src/styles/skins.css` — 5 套浅色皮肤变量（覆盖 shadcn 语义变量 + stone + 点阵色）
  - `src/context/skin-provider.tsx` — 运行时引擎，`data-skin` 写入，默认 `neo`，`BUNDLED_SKINS` 5 套
  - `src/context/theme-provider.tsx` — 锁定浅色（删 system/dark 监听）
  - `src/components/skin-switch.tsx` — 顶栏皮肤切换器
  - `src/components/config-drawer.tsx` — 设置面板皮肤卡片（2 列网格，显示 name+description）
  - `src/features/docs/index.tsx`、`src/features/pricing/components/pricing-sidebar.tsx`、`features/models/*` 弹窗、`features/wallet/*` 卡片等 — 硬色改语义 token
  - `src/l10n/locales/{zh,en,fr,ru,vi,ja,ar}.json` — 翻译键补齐
- **已知残留**：`text-white` 仍有约 25 处集中在 `features/home` 营销区（深色设计，刻意）；状态色映射在 `lib/colors.ts` / `components/layout/constants.ts` / `status-badge.tsx`。若某页面仍不清晰，截图定位后定点修。

---

## 4. 部署流程（标准 + 避坑，务必照做）

> **核心铁律**：`deploy.sh` 只做 `go build -a` 重新嵌入 `dist`，**不会重新构建前端**。所以正确顺序是**先重建前端 → 再部署嵌入**。

### 4.1 标准部署（前端有改动时）

```bash
# 1) 重建前端（务必先清 rsbuild 缓存，避免嵌旧 dist）
cd /opt/fasttoken/web/default
rm -rf node_modules/.cache/rsbuild
npm run build
# 校验产物确实含修复：grep -c keySeparator dist/assets/js/index.*.js  应 > 0

# 2) 部署（重新嵌入 dist + 替换二进制 + 重启 + 冒烟 + 自动回滚）
cd /opt/fasttoken
bash deploy.sh
# 成功末尾打印：DEPLOY_OK (<TS>)
```

### 4.2 用 systemd-run 脱离 SSH 跑长任务（防断连 SIGKILL）

`systemd-run` 会**剥离环境变量**，Go 编译需显式导出：

```bash
systemd-run --no-block --unit=ft-deployX.service sh -c \
  'export HOME=/root GOPATH=/root/go GOMODCACHE=/root/go/pkg/mod GOCACHE=/root/.cache/go-build PATH=/usr/local/go/bin:$PATH; bash /opt/fasttoken/deploy.sh'
# 轮询：journalctl -u ft-deployX.service -n 30 ; systemctl is-active ft-deployX.service
```

### 4.3 deploy.sh 行为（要点）

- 备份当前二进制 → `fasttoken.bak.<TS>`（回滚点）
- `go build -a -o /tmp/ft.new .`（`-a` 强制无缓存重嵌 dist）
- 替换二进制、`chmod 600 .env` 与微信公钥
- `systemctl restart fasttoken` + 睡眠 8s
- 冒烟 `curl /api/payment/status` 最多重试 5 次；失败则**自动回滚**到 `fasttoken.bak.<TS>` 并重启
- 退出码：成功 `DEPLOY_OK`；构建失败 `BUILD_FAIL`；冒烟失败会回滚并打印 `SMOKE_FAIL`

### 4.4 deploy.sh 增强（2026-07-22 起）

脚本已强化前置校验，固化历史三大部署坑（§7-6）：
- 每次部署先清 `node_modules/.cache/rsbuild`（防嵌旧 dist）。
- 校验 `dist` 含 i18n 修复标记 `keySeparator`（缺则 `PREFLIGHT_FAIL` 中止，避免嵌坏/旧 dist）。
- 构建后再校验新二进制内嵌 dist 含 `keySeparator`（确认 embed 成功）。
- 显式 `export` Go + node/npm 全套环境变量（兼容 `systemd-run` 剥离环境）。
- 新增用法：`bash deploy.sh preflight` 仅跑前置校验（不构建/不重启，CI/排错用）；`bash deploy.sh --with-frontend` 部署前自动重建前端。
- 原版备份：`/opt/fasttoken/deploy.sh.orig.20260722`。

### 4.5 本轮实测要点（2026-07-29 补充）

- **前端构建 OOM**：`npm run build`（rsbuild）会在打印 Total 后、收尾阶段被 SIGKILL（exit 137），但 **dist 产物已落盘完整**（275 文件、CSS 407KB、主 JS 完整）。判断依据：用真实文件名 `ls assets/css/index.*.css` 核对，勿信 glob 子 shell 误报的"147 文件"。
- **轻量部署脚本**（本地 `ops/build_ft.sh` + `ops/deploy_ft.sh`，经验证可跑）：`build_ft.sh` = `rm -rf dist && npm run build`；`deploy_ft.sh` = `go build -a -o fasttoken.new .` + 备份 + 替换 + `systemctl restart fasttoken`。`deploy.sh` 仍是项目标准（带自动回滚），优先用 `bash deploy.sh --with-frontend`。
- **清理结果（2026-07-29）**：`/tmp` 会话临时脚本（14 个 .sh/.py）已删除；`fasttoken.bak.*` 从 12 个降到 2 个（保留最新 08:29、08:45 两个回滚点）。

---

## 5. 配置 / 缓存 / 网络 / 备份 现状

- **配置**：`.env` 含支付（支付宝/微信）、Redis、Postgres(DSN)、限流、DeepSeek 等约 60 项键；权限 `600`。改完 `systemctl restart fasttoken` 生效。
- **缓存**：
  - rsbuild 缓存 `node_modules/.cache/rsbuild`：构建前清掉（防嵌旧 dist）。
  - Go 构建缓存 `/root/.cache/go-build`：正常，无需动。
  - 业务缓存：由 `REDIS_CONN_STRING` 与 `MEMORY_CACHE_ENABLED` 控制。
- **网络**：nginx 限流 + 安全头已就位；`/etc/nginx/conf.d/` 下有多个 `fasttoken.conf.bak.*` 历史残留（无害，可删）。
- **备份**：**无 Git、无源码副本**（用户选择）。仅每次 `deploy.sh` 自动产生一个回滚二进制。2026-07-19 清理时已删除一个内嵌崩溃版（`index.ac0b4202`）的危险回滚；当前无 `.bak` 残留，下次部署会自动重建。
- **建议**：在服务器之外（如对象存储 / Git）建立 dist 与 `.env` 的离线备份。

---

## 6. FastToken skills（AI Agent 用）

位置：`/opt/fasttoken-skills/skills/`，含**用户版**与 **admin 版**两套，由 `github.com/QuantumNous/skills` 的 newapi skill 适配而来（安全变换、不改 URL），保持 New API 面向用户的设计哲学：密钥脱敏、占位符注入、输出净化。

### 6.1 fasttoken（用户面向）`/fasttoken`
- 触发命令：`/fasttoken`
- 环境变量：`FASTTOKEN_BASE_URL`、`FASTTOKEN_ACCESS_TOKEN`、`FASTTOKEN_USER_ID`
- 动作：`models` / `groups` / `balance` / `tokens` / `create-token` / `switch-group` / `copy-token` / `apply-token` / `exec-token`
- 安全：禁止明文 `sk-` 密钥；密钥只允许走 `copy-token`(剪贴板) / `apply-token`(写配置) / `exec-token`(命令注入)；禁止读 `.env`；所有调用走 `scripts/api.js` 而非裸 `curl`。

### 6.2 fasttoken-admin（管理员面向）
- 环境变量：`FASTTOKEN_ADMIN_BASE_URL`、`FASTTOKEN_ADMIN_ACCESS_TOKEN`、`FASTTOKEN_ADMIN_USER_ID`（USER_ID 须为管理员）
- 动作：`users` / `channels` / `models` / `options` + **FastToken 特色功能**：
  - 签到 `/api/checkin`、兑换 `/api/redemption`、充值 `/api/topup`、企业 `/api/enterprise`、部署 `/api/deployments`、偏好组 `/api/prefill_group`、分组比例 `/api/group-ratio`、性能 `/api/perf`、排行榜 `/api/rankings`
  - ⚠️ 上述特色功能子路径基于"实测路由组 + New API 管理 API 约定"给出；**更新软件时需按真实接口细化**（详见 `docs/actions-admin.md`）。
- 安全：同用户版，且破坏性操作（删用户/渠道、改配额）前须向用户确认。

### 6.3 使用与交付
- 尚未初始化为 Git 仓库。使用方式：`npx skills add` 或整目录复制到 Agent 的 skills 目录。
- 设计哲学：保持 New API "面向用户"的简洁；admin = 同一安全模型 + 额外管理动作，不引入会泄露密钥的流程。

---

## 7. 运维必读：踩过的坑（经验教训）

1. **ESLint 10.7.0 的 `no-unused-vars` 没有 `--fix` fixer**（含变量声明、catch 绑定均不自动移除）。别指望 `eslint --fix`；用针对性脚本删除 + 每次用 ESLint 解析扫描验证。
2. **catch 正则破坏**：`catch (e) {` 误改成 `catch {}` 会丢掉 `{`，使 catch 体悬空 → 多一个 `} finally`。ES2020 支持可选 catch 绑定，正确修法是 `catch {}` → `catch {`（补回大括号），对真正为空的 `catch {}` 语义等价。
3. **多行函数脚本删除会留悬空体**：只删首行 → TS 解析错误。删改任何代码后必须用 ESLint 解析扫描（或 `npm run build`）验证 0 错误。
4. **i18n `keySeparator: false` 必需**（见 §3.3-1）。
5. **TDZ**：模块顶层不要在 `const` 声明前调用它（见 §3.3-2）。
6. **部署三重坑**：① `deploy.sh` 不重建前端，须先 `npm run build`；② rsbuild 缓存会让前端构建命中旧产物 → 先清缓存；③ `systemd-run` 剥离环境 → 显式 `export` Go 环境。三者任一遗漏都会"部署成功但线上还是旧/坏版本"。
7. **无 Git/备份**：任何有风险的改前先做 `cp -a src /tmp/src_backup`。
8. **本地通道**：`rsh.py`/`put_file.py` 在 `C:/Users/Administrator/AppData/Local/Temp/wb-ft/`，**勿命名 `inspect.py`**（遮蔽标准库）。
9. **压缩包排错**：给 rsbuild 临时加 `devtool: 'hidden-source-map'`（不改变 JS 内容与哈希），用 `source-map` 库把 `index.xxx.js:行:列` 反查回源码定位 TDZ 等运行时错误。

---

## 8. 后续计划（建议）

- [x] **智企惠品牌区块**多语言外部化：改为 `about.brandXxx` 形式，补齐 7 语言包（2026-07-19 完成并 2026-07-20 部署，见 §3.4）。
- [ ] **skills 版本化**：把 `/opt/fasttoken-skills` 初始化为 Git 仓库，便于协作与回滚。
- [ ] **admin 特色功能接口对齐**：软件更新后，按真实后端路由核实 `/api/checkin`、`/api/redemption`、`/api/topup`、`/api/enterprise`、`/api/deployments`、`/api/prefill_group`、`/api/group-ratio`、`/api/perf`、`/api/rankings` 的实际路径与参数。
- [ ] **部署前置校验**（可选增强 `deploy.sh`）：清 rsbuild 缓存 → 校验 `keySeparator` 进包 → 校验 dist 哈希变化 → 再 `go build -a`。把今天踩的坑固化进脚本。
- [ ] **离线备份**：建立 dist + `.env` 的服务器外备份（当前无 Git、无副本）。
- [ ] **安全加固**：root 口令改密 / 改 SSH 密钥登录；收紧 `.env` 与证书访问。

---

## 9. 紧急运维速查（常用命令）

```bash
# 状态 / 日志
systemctl status fasttoken
journalctl -u fasttoken -f --no-pager
systemctl status nginx postgresql redis

# 健康检查
curl -fsS --max-time 10 http://127.0.0.1:3000/api/payment/status

# 重启服务
systemctl restart fasttoken
nginx -t && systemctl reload nginx

# 重新部署（标准两步，见 §4.1）
cd /opt/fasttoken/web/default && rm -rf node_modules/.cache/rsbuild && npm run build
cd /opt/fasttoken && bash deploy.sh

# 手动回滚（deploy.sh 冒烟失败会自动回滚；以下为手动）
ls /opt/fasttoken/fasttoken.bak.*          # 找回滚点
mv /opt/fasttoken/fasttoken.bak.<TS> /opt/fasttoken/fasttoken
systemctl restart fasttoken

# 前端源码解析扫描（排错用）
cd /opt/fasttoken/web/default && npx eslint . --format json 2>/dev/null | python3 -c "import sys,json;d=json.load(sys.stdin);[print(f['filePath'],len(f['messages']),'issues') for f in d if f['messages']]"
```

---

*本清单由 AI 在 2026-07-19 根据服务器实时状态整理，供后续人类/ Agent 接手使用。如项目有重大变更，请同步更新本文件（服务器副本：`/opt/fasttoken/HANDOVER.md`）。*

---

## 变更记录 2026-07-26（L10n bug 修复 + 配置即数据核查）

已部署（DEPLOY_OK 20260726225759，线上 `index.fc869ae7.js`，回滚点 `fasttoken.bak.20260726225759`）：

1. **P0 充值赠送双真相源修复**：下单路径（`controller/topup_alipay.go` / `topup_wechat.go`）的 bonusQuota 原走旧 `options.recharge_gift_setting`（只剩 100 元档），而充值页展示走新 `activities` 表（四档 20%）→ 用户充 200/500/1000 会"有展示无到账"。现两侧统一走 `model.EvaluateTopupGiftBonus`（activities 表，内置旧配置兜底）。
2. **L10n 结构修复**：7 个 locale JSON 中 260 个键曾在 `translation` 包裹之外（被 i18next 当 namespace 忽略，首页 hero/品牌区显示英文原键）→ 已并入 translation。
3. **L10n 重渲染修复**：`l10n/config.ts` i18n init 增加 `react: { bindI18nStore: added removed }`，DB 翻译异步注入后能即时重渲染。
4. **en 翻译补齐**：`i18n_messages` 中 105 条 en 值为中文的 about/docs/FAQ 条目已翻译（SQL 直改，免部署）。剩 1 条为微信公众号关键词「验证码」，属有意保留。

已核查（配置即数据状态）：i18n_messages/model_pricing/activities/activity_grants 四表 + reload/pricing/i18n/activities 管理端点全部落地。活动引擎 `DispatchActivityEvent` 暂无调用方（赠送走下单并入 Amount 的路径，engine 备用于未来注册/邀请/抽奖类活动）。
遗留：fr(~400)/ja(~350)/ru(~350)/vi(~380)/ar(~140) 条 value=key 未翻译；54 个后端消息键仅 zh 有翻译。均可 SQL 直改免部署。

---

## 变更记录 2026-07-26（第二次：批量补齐翻译 + fallback 增强）

DEPLOY_OK 20260726234852（线上 `index.3081bbb6.js`，回滚点 `fasttoken.bak.20260726234852`）：

1. **fallbackLng 增强**：`src/l10n/config.ts` 由 `fallbackLng: zh`（缺失键回退中文→非中文界面显示中文）改为 `fallbackLng: [en,zh]`。任何语言缺失的键先回退英文，生效面覆盖全部语言。
2. **批量翻译补齐（DB 直写，免部署）**：用 MyMemory 免费接口翻译 416 条 UI 未翻译（fr/ja/ru/vi 各 104 条，key 为中文的那些），生成 `fill.sql` 经 `i18n_messages` 表 UPDATE（414 条成功；2 条接口未返回结果保持中文：`vi 公司名称`、`fr 灵活密钥管理`）。另确认 54 个后端消息键的 en 行均已存在且为英文（无需补）。
3. 修复脚本 bug：初版 `fill.sql` 因任务分组逻辑错误把单一语言翻译广播给全部 4 个 locale（ja/ru 误填法语），已基于正确的 `fill_cache.json`（4 语言分别存储）重生干净 SQL 后写库。
4. 数据校验：fr/ja/ru/vi 仅剩 2 条 value=key 含中文（上述 2 条）；所有 zh key 均有对应 en 行；UI 翻译经 API 抽查生效（fr: Foire Aux Questions 等）。

遗留（极小量，可后续 SQL 直补）：`vi 公司名称`、`fr 灵活密钥管理` 两条；以及各语言 about/docs 长文翻译质量可人工润色。

---

## 变更记录 2026-07-27（前端皮肤升级 P0：测试护栏 + 防闪烁）

DEPLOY_OK 20260727002001（回滚点 `fasttoken.bak.20260727002001`）：

1. **防首屏闪烁(FOUC)**：`web/default/index.html` 的 `<head>` 注入内联脚本，在 hydration 前根据 cookie `vite-ui-theme`（dark/light/system，沿用 theme-provider 既存约定）同步设置 `<html class="dark|light">`；并前向兼容读取 `vite-ui-skin`（P1 起由 skin-provider 写入）。消除此前存在的明暗切换首屏闪烁。仅改 index.html，功能/内容零变动。
2. **视觉回归测试基座（用户要求必须有测试）**：
   - `web/default/playwright.config.ts` + `web/default/e2e/skins.spec.ts`：数据驱动矩阵（明暗轴 × 路由），截图建档 + **内容不变量断言**（页面必须含品牌锚点，确保 L10n/内容不被破坏）。
   - 远端 `web/default` 已 `npm i -D @playwright/test`，yum 补齐 Chromium 系统依赖，浏览器待下载完成。
   - 设计：皮肤轴（data-skin）将在 P1 接入后并入矩阵，自动扩展到 4×2=8（含 system 取 12）组合。

P0 剩余：Chromium 下载完成后跑 `npx playwright test` 生成 baseline 快照并验证内容断言。

---

## P0 完成确认（2026-07-27）

- Chromium 经 **npmmirror 国内镜像**下载成功（`PLAYWRIGHT_DOWNLOAD_HOST=https://cdn.npmmirror.com/binaries/playwright`）；默认 Azure CDN 在国内服务器卡死。已装 `@playwright/test`，yum 补系统依赖。
- 视觉回归基座跑通：`npx playwright test` 4 用例全绿，生成 baseline 快照 4 张（`e2e/__snapshots__/skins.spec.ts-snapshots/`：light/dark × root/login）。
- FOUC 防闪烁已随 `DEPLOY_OK 20260727002001` 上线生效（线上 HTML 含脚本）。
- P0 收尾：测试护栏 + 防闪烁，零业务改动。下一步 P1（皮肤 Token 系统）。

## 变更记录 2026-07-27（皮肤升级 P1：皮肤 Token 系统 + 最小绑定）

DEPLOY_OK 20260727012921（线上 CSS hash index.5e7452d4.css，回滚点 fasttoken.bak.20260727012215）：

1. **新增皮肤 Token 体系**（`web/default/src/styles/skins.css`）：
   - 4 套皮肤 × 每套 60+ 个 CSS 变量（color/type/shape/space/motion/layout/surface/decorative）
   - `neo`（默认/演进版现状）、`aurora`（科技渐变玻璃/紫调）、`classic`（企业衬线/蓝调）、`midnight`（暗色优先/霓虹绿）
   - 双轴正交：明暗轴（.dark）在 <html>，皮肤轴（data-skin）也在 <html>，共 12 组合
2. **运行时引擎**（`web/default/src/context/skin-provider.tsx`）：读 `vite-ui-skin` cookie → 写 `<html data-skin>`；FOUC 防闪烁已集成；P2 接 DB 只需补 `setSkins(serverSkins)`，无需改组件。
3. **最小可演示绑定**：`[data-skin='x'] body` + `[data-skin='x'] .bg-stone-bg` 覆盖 body/页面 wrapper 背景。neo 故意不绑（零回归）；组件级打磨留给 P3。
4. **测试门禁**：
   - `e2e/skins.spec.ts`（4 neo 组合，对应 P0 基线）— 4/4 像素一致
   - `e2e/skins-matrix.spec.ts`（16 矩阵 = 4 皮肤 × 2 明暗 × 2 路由）— 16 张快照全部建档
   - 20/20 测试在线上端口通过
5. **踩坑记录**（运维提示）：
   - rsbuild CSS 缓存会无视 skins.css 修改，必须 `rm -rf node_modules/.cache/rsbuild` 再 build
   - 新增 CSS 文件必须 import 到 `main.tsx`，否则 `--update-snapshots` 不会失败但实际未生效
   - `pkill -f 'rsbuild preview'` 会误杀 ssh 自身（命令行含 preview 字样），改用精确 pid

P1 状态：✅ 完成 + 部署验证。P2 待启动（DB `ui_skins` 表 + 后台 CRUD + 热加载）。

---

## 变更记录 2026-07-29（主题合并 + 白版锁定 + 硬色清理 + 服务器清理）

DEPLOY_OK 20260729084530（线上 `index.3a797703.css` / `index.f94cb0a0.js`，回滚点 `fasttoken.bak.20260729084530`、`fasttoken.bak.20260729082916`）：

1. **主题与皮肤合并为"皮肤"一套系统**：移除原"浅色/深色/系统"三色明暗入口（顶栏、命令面板、错误页、设置面板 4 处），顶栏 Palette 下拉改为皮肤切换器（`skin-switch.tsx`），列出 5 套皮肤 neo/aurora/classic/midnight/sunset。
2. **白版锁定（修复整站变黑）**：`theme-provider.tsx` 改为永远浅色（删 `system` 分支与 matchMedia 监听，强制 `remove('dark')`+`add('light')`）；`skins.css` 删除全部 5 段 `[data-skin=x].dark` 暗色块（双保险）。结论：全站固定白色/浅色，无明暗切换。
3. **配色可读性修复**：14 个页面文件 100+ 处写死颜色（`bg-white`/`text-gray-*`/`text-slate-*`）统一改为语义 token（`bg-card`/`text-foreground`/`text-muted-foreground`）；修品牌色背景配白字 2 处对比度 bug。刻意保留：首页营销深色面板、支付页、二维码白底、黑底日志终端。
4. **i18n 补齐**：7 语言包补齐 `Theme`/`Select theme`/`Theme Settings` 及 `skin.<id>.{name,description}`。
5. **服务器清理**：删除 `/tmp` 14 个会话临时脚本；`fasttoken.bak.*` 由 12 个减至 2 个（保留最新两个回滚点）。

验证：首页 200、服务 active（PID 2901967）、二进制暗色背景值 `oklch(0.16 0.012 250)`=0（暗色套已移除）、`data-skin` 选择器 5 套已打入。

遗留：个别页面若仍不清晰，多为 `features/home` 营销区 `text-white`（刻意深色设计）或状态色映射（`lib/colors.ts` 等），需截图定点修。
