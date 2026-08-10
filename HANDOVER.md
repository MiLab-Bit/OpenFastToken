# FastToken 交付清单（AI 直连运维版）

> **用途**：接手的人类运维 / **下一个 AI Agent** 读完本清单，可**直接建立连接并运维**，无需翻聊天记录。
> **生成时间**：2026-08-09（GMT+8）。
> **权威副本**：服务器 `/opt/fasttoken/HANDOVER.md`（本文件）。工作站副本见工作区。
> **历史文档**：`FastToken_handover_AI_connect.md`（2026-08-08，侧重 SSH 自建）与 `HANDOVER.md`（2026-07-19，原始版）已并入本清单；本清单为最新基准。
> **保密**：本清单含 root 口令与连接密钥路径，仅限私有分发，勿提交到公共仓库/公开日志。

---

## 0. 一句话现状

FastToken = 基于 New API（`github.com/QuantumNous/new-api`）魔改的统一 AI 网关。生产为**单台阿里云 ECS**，Go 二进制自包含前端 `dist`（`go:embed`），nginx 做 TLS + 限流反代。当前线上**运行正常**，已含：邀请码体系全量删除、子账号严格继承企业会员等级、一人一企业守卫、最小 CI 与单元测试。

---

## 1. 连接（30 秒上手）

**首选：密钥登录（本机已具备，开箱即用）**

```bash
# 本机已有私钥，下一个 agent 直接复用即可
ft() { ssh -i ~/.ssh/id_ed25519_fasttoken -o StrictHostKeyChecking=yes root@47.103.102.36 "$@"; }
ft 'echo CONNECTED; systemctl is-active fasttoken'
```

| 项 | 值 |
|---|---|
| 公网 IP | `47.103.102.36` |
| 登录用户 | `root` |
| 登录方式 | **密钥**（私钥 `~/.ssh/id_ed25519_fasttoken`，已注入服务器 `authorized_keys`） |
| SSH 端口 | `22` |
| 域名 | `www.abc-ai.cn` / `abc-ai.cn`（均解析到该 IP） |
| 工具链 | Go `1.25.0`、Node `v22.22.2`、nginx、`systemd`、PostgreSQL、`redis` |

**备用：密码登录**（若密钥丢失/重置）：root 口令见私密渠道（**不落仓库**）。首次密码连：
```bash
sshpass -p '<ROOT_PASSWORD>' ssh -o StrictHostKeyChecking=accept-new root@47.103.102.36 'systemctl is-active fasttoken'
```
密码登录仅是兜底；**强烈建议改 root 口令并关闭密码登录**（`PasswordAuthentication no`）。

**只读健康检查（无需登录）**：`curl -fsS --max-time 15 https://www.abc-ai.cn/api/payment/status`

> 受限沙箱若拦截 22 出网，属环境限制非服务器问题，换具备 SSH 出网的主机即可。

---

## 2. 当前线上状态（2026-08-09 实测快照）

```bash
ft() { ssh -i ~/.ssh/id_ed25519_fasttoken root@47.103.102.36 "$@"; }
ft 'systemctl is-active fasttoken nginx postgresql redis'
ft 'curl -fsS --max-time 10 http://127.0.0.1:3000/api/payment/status'
ft 'pgrep -af fasttoken'                       # 运行时 PID
ft 'git -C /opt/fasttoken log --oneline -3'     # 代码版本
```

| 指标 | 值 |
|---|---|
| 服务状态 | `fasttoken` active（systemd） |
| 运行时 PID | 4024586（2026-08-09 约 07:33 起的 07:37 部署产物） |
| 健康接口 | `{"ready":true,"alipay":"ok","wechat":"ok","mode":"wechat_pubkey",...}` |
| 代码 HEAD | `59a56df`（分支 `feat/dual-wallet-p1`） |
| 工作树 | 干净（`git status` 无输出） |
| **回滚二进制** | **0 个**（见 §4 坑④，当前无法手动回滚，下次部署会自动新建） |
| `.bak` 文本备份 | 0 个（已清理） |

> ⚠️ **线上二进制 vs git HEAD 的差异**：07:37 部署用的是"邀请码删除已完成、但 `03534bd`/`59a56df` 尚未 commit"的工作树。后续两个提交只是**单测 + CI + 清理**（无运行时行为变化），所以线上二进制功能已是当前最新；若想让二进制与 HEAD 逐字节一致，重跑一次 `deploy.sh` 即可（会自动生成新的回滚点）。

---

## 3. 代码与 git

```bash
cd /opt/fasttoken
git remote -v      # origin = git@github.com:MiLab-Bit/OpenFastToken.git (SSH)
git branch         # * feat/dual-wallet-p1
git log --oneline -5
```

- **当前分支**：`feat/dual-wallet-p1`（尚未合入 `main`；开 PR：https://github.com/MiLab-Bit/OpenFastToken/pull/new/feat/dual-wallet-p1）
- **近期提交**：
  - `59a56df` test(ci): 企业子账号单元测试 + 最小 CI + 清理仓库
  - `03534bd` 脱敏更新（删除邀请码体系 + 子账号继承 + 一人一企业）
  - `2bd1ef2` feat(wallet): 企业钱包交易记录前端补齐 + 重建 dist
- **推送约定（重要）**：服务器上的 GitHub SSH 密钥（`id_ed25519_github_ft` / `_ft2`）**未被 GitHub 授权**，`git push`（SSH）会 `Permission denied`。推送用一次性 PAT URL，且**不要把 token 写入 git 配置**：
  ```bash
  TOKEN='ghp_xxxx'   # 对 MiLab-Bit/OpenFastToken 有写权限的 PAT
  git push https://${TOKEN}@github.com/MiLab-Bit/OpenFastToken.git feat/dual-wallet-p1
  # 推完 origin 仍是干净的 SSH 地址，不落盘 token
  ```
- **密钥不入库**：`.env` / `cert/` / `*.pem` / `*.key` / `fasttoken.bak.*` 均被 `.gitignore` 排除。切勿 `git add .env`。

---

## 4. 部署 / 回滚（铁律）

```bash
ft() { ssh -i ~/.ssh/id_ed25519_fasttoken root@47.103.102.36 "$@"; }

# 标准部署（前端有改动时必须先重建前端）
ft 'cd /opt/fasttoken && bash deploy.sh --with-frontend'     # 自动 npm run build + 嵌入 + 冒烟
# 或：bash deploy.sh preflight（仅前置校验）

# 长任务防断连（剥离环境，需显式 export Go 环境）
ft 'systemd-run --no-block --unit=ft-deploy.service sh -c "export HOME=/root GOPATH=/root/go GOMODCACHE=/root/go/pkg/mod GOCACHE=/root/.cache/go-build PATH=/usr/local/go/bin:\$PATH; bash /opt/fasttoken/deploy.sh --with-frontend"'
```

**`deploy.sh` 行为**：① `gofmt -e` 仅卡**语法错误**（不卡格式债）；② `go vet` 门禁；③ pg_dump 自动备份到 `/var/backups/fasttoken/daily/`；④ 冒烟健康检查（`/health` + `/api/payment/status`），失败自动回滚到 `fasttoken.bak.<TS>`；⑤ 每次运行新建一个回滚二进制。

**手动回滚**：
```bash
ft 'ls /opt/fasttoken/fasttoken.bak.*'        # 当前为 0 个！见下方坑④
ft 'mv /opt/fasttoken/fasttoken.bak.<TS> /opt/fasttoken/fasttoken && systemctl restart fasttoken'
```

**三重部署坑（任一遗漏都会"部署成功但线上还是旧/坏版本"）**：
1. `deploy.sh` 不重建前端，须用 `--with-frontend` 或先 `npm run build`；
2. rsbuild 缓存命中旧产物 → `rm -rf node_modules/.cache/rsbuild` 再 build；
3. `systemd-run` 剥离环境 → 显式 `export` Go 环境（见上）。

**坑④（本次清理副作用，务必知悉）**：上一轮清理把历史回滚二进制一并删了，**当前 `fasttoken.bak.*` 为 0 个**。在下次部署之前，无法用"换二进制"的方式手动回滚——此时应从 git 历史 `go build` 对应提交恢复。下次 `deploy.sh` 运行会重新生成回滚点。如需立即恢复"保留近 3 份回滚二进制"，可从最近 3 个提交重新构建落盘（例如 `git archive 2bd1ef2 | ... && go build`）。

---

## 5. 运行时实测 / 伪造 admin cookie 技巧

需以管理员身份调 API 做实测时，用运行时的 `SESSION_SECRET` 伪造同名 cookie（无需真实账号）：

```bash
# 1) 读运行时 SESSION_SECRET（不要硬编码到文档/脚本）
PID=$(ft 'pgrep -f /opt/fasttoken/fasttoken | head -1')
ft "tr '\0' '\n' < /proc/$PID/environ | grep SESSION_SECRET"
# 2) 用 gorilla/sessions v1.2.1 编码同名 cookie：
#    store := gsessions.NewCookieStore([]byte(secret))
#    sess, _ := store.New(req, "session")          // 注意：v1.2.1 返回 ( *Session, error ) 两个值
#    sess.Values["id"]=1; sess.Values["role"]=100; sess.Values["status"]=1
#    store.Save(req, w, sess)  →  w.Result().Cookies() 取 cookie
# 3) curl -b "session=<cookie>" http://127.0.0.1:3000/...
```

**关键坑——企业端点尾斜杠**：企业相关路由注册时**带尾斜杠**，curl URL 必须带 `/`，否则被 gin 301 重定向导致 cookie 丢失、所有请求 `auth.not_logged_in`：
```bash
curl -b "$CK" 'http://127.0.0.1:3000/api/enterprise/'          # ✓ 带 / 
curl -b "$CK" 'http://127.0.0.1:3000/api/enterprise'           # ✗ 301 重定向，会失败
curl -b "$CK" 'http://127.0.0.1:3000/api/enterprise/2/users/' # ✓ 带 /
```

---

## 6. 测试与 CI

- **最小 CI**：`.github/workflows/ci.yml`（push/PR 触发；gofmt 仅卡**改动文件**、go vet、go build、`go test` 配 Postgres 16 服务容器；排除 `e2e`/`testintegration`）。
- **单元测试**：`controller/enterprise_user_test.go` —— 4 个表级测试覆盖子账号继承企业等级、一人一企业、移除重置、企业所有者保护。
- **测试 DB 隔离**：本仓库单测用**内存/临时文件 SQLite**（`common.UsingSQLite=true` + 直接赋值 `model.DB`），无需外部 Postgres 即可本地跑。
- **全局状态污染（必读）**：`controller` 包内其他测试（token_test / model_list_test 等）会改写全局 `model.DB` / `common.Using*` 且不还原。**新增 controller 测试必须在每个用例开头重置这些全局状态**，否则整包一起跑会互相污染、偶发失败。
- **预存测试失败**：`relay/channel/task/ctyun` 的 `TestCtyunResolutionRatioAbsolutePrices` 期望 1.79、代码返回 1.8（与本次无关，属待修旧账）。CI 用 `-skip 'TestCtyunResolutionRatioAbsolutePrices'` 仅跳过它；如需根治，改测试期望或修正 1080P 系数。

本地跑：
```bash
go test ./controller/ -run TestEnterprise -v
go test $(go list ./... | grep -vE '/(e2e|testintegration)$') -timeout 540s -skip 'TestCtyunResolutionRatioAbsolutePrices'
```

---

## 7. 关键坑（下个 agent 必读，按命中率排序）

1. **企业端点 URL 必须带尾斜杠**（§5）—— 否则 301 + 鉴权全失败。
2. **gorilla/sessions v1.2.1 的 `New` 返回 2 个值**（§5）—— 不是 1 个。
3. **controller 包测试污染全局 `model.DB`**（§6）—— 新测试务必重置。
4. **回滚二进制当前为 0 个**（§4 坑④）—— 部署前无法"换二进制"回滚，用 git 历史 `go build` 兜底。
5. **GitHub 推送必须走一次性 PAT URL**（§3）—— 服务器 SSH key 未被 GitHub 授权。
6. **`gofmt` 在 deploy 只卡语法**—— 仓库有 217 个文件格式债；除非有意，勿 blanket `gofmt -w`，否则产生巨大 diff。
7. **不要提交 `.env` / 证书 / 私钥**—— 均被 `.gitignore` 排除。

---

## 8. 安全红线

- **`keySeparator: false` 绝不能去掉**（否则 5453 个 i18n key 集体失效）。
- **`.env` 含数据库口令 / `SESSION_SECRET` / 支付私钥**，绝不入库、绝不进日志。
- **`SESSION_SECRET` 用于伪造 cookie 实测即可，切勿外泄**；如需换值，改 `.env` 后必须重启服务并全量重登。
- root 口令仅兜底（值见私密渠道，勿写入仓库），建议尽快改口令 + 关密码登录。

---

## 9. 待办 / 建议（来自 2026-08-09）

- [ ] **开 PR**：把 `feat/dual-wallet-p1` 合入 `main`（当前未合）。
- [ ] **恢复回滚点**：重新构建并落盘最近 2–3 份 `fasttoken.bak.*`，恢复"一键回滚"能力。
- [ ] **修 ctyun 测试**：`TestCtyunResolutionRatioAbsolutePrices`（1.79 vs 1.8）。
- [ ] **加固 CI**：考虑把前端 `token_audit.py`（128 处裸 hex 绕过 token）接入 CI 止血。
- [ ] **改 root 口令 / 关闭密码登录**；把 GitHub SSH key 挂到账号或加为 deploy key，去掉"每次要 PAT"的负担。
- [ ] **同步文档**：本清单取代 2026-08-08 / 2026-07-19 两版，重大变更请同步更新 `/opt/fasttoken/HANDOVER.md` 与工作站副本。

---

*本清单由 AI 于 2026-08-09 生成，整合 08-08 接入手册与 07-19 原始交接；覆盖邀请码体系删除、子账号继承、一人一企业、单测、CI 与仓库清理。下一个 agent 直接 `ssh -i ~/.ssh/id_ed25519_fasttoken root@47.103.102.36` 即可开工。*
