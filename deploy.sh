#!/usr/bin/env bash
# FastToken 部署脚本：配置安全 + 幂等 + 自带冒烟与回滚 + 部署前置校验
# v2 (2026-08-02)：加 gofmt/vet 门禁 + 强制 pg_dump 备份 + git 版本戳 + -a 但限核
# 原则：仅替换二进制，绝不触碰 .env / DB / nginx / systemd / cert/
#
# 用法：
#   bash deploy.sh                 # 标准部署（仅嵌入当前 dist，不重建前端）
#   bash deploy.sh --with-frontend # 部署前自动重建前端（清 rsbuild 缓存 + npm run build）
#   bash deploy.sh preflight       # 仅跑前置校验，不构建/不重启（CI / 排错用）
#
# 前置校验固化了历史三大部署坑（见 HANDOVER §7-6）：
#   ① deploy.sh 不重建前端 -> 用 --with-frontend 可自动重建；否则校验 dist 含关键修复标记
#   ② rsbuild 缓存嵌旧 dist -> 每次都先清 node_modules/.cache/rsbuild
#   ③ systemd-run 剥离环境  -> 显式 export Go + node/npm 全套环境变量
set -eu

BIN=/opt/fasttoken/fasttoken
WEB=/opt/fasttoken/web/default
TS=$(date +%Y%m%d%H%M%S)
BAK=$BIN.bak.$TS
BACKUP_SCRIPT=/usr/local/bin/ft-backup-db.sh

# ---- 环境（同时兼容直接执行与 systemd-run；后者会剥离环境变量）----
export HOME=/root GOPATH=/root/go GOMODCACHE=/root/go/pkg/mod GOCACHE=/root/.cache/go-build
export PATH=/usr/local/go/bin:/usr/local/node22/bin:/usr/local/bin:/usr/bin:/usr/local/sbin:/usr/sbin:$PATH
export GOGC=50 GOFLAGS=-p=2 GOPROXY=https://goproxy.cn,direct GIN_MODE=release

# 关键修复标记：i18n 解析修复（HANDOVER §3.3-1）。必须进包，否则 5453 个 key 集体失效。
I18N_MARK="keySeparator"

preflight() {
  echo "[preflight] clearing rsbuild cache ($WEB/node_modules/.cache/rsbuild)"
  rm -rf "$WEB/node_modules/.cache/rsbuild"

  echo "[preflight] checking dist exists"
  if [ ! -d "$WEB/dist" ]; then
    echo "PREFLIGHT_FAIL: $WEB/dist 不存在 —— 请先构建前端"
    return 1
  fi

  echo "[preflight] verifying i18n fix marker ('$I18N_MARK') present in dist"
  if ! grep -rq "$I18N_MARK" "$WEB/dist" 2>/dev/null; then
    echo "PREFLIGHT_FAIL: dist 中未找到 '$I18N_MARK'（前端可能未构建或未含 i18n 修复）"
    return 1
  fi

  # 若启用自动前端构建，确认 node/npm 可用
  if [ "${WITH_FRONTEND:-0}" = "1" ]; then
    if ! command -v npm >/dev/null 2>&1; then
      echo "PREFLIGHT_FAIL: --with-frontend 需要 npm，但未在 PATH 中找到"
      return 1
    fi
  fi

  echo "[preflight] OK"
  return 0
}

# ---- v2 新增：测试门禁（gofmt 语法 + go vet 静态检查）----
code_gate() {
  echo "[gate] gofmt syntax check (-e: 语法错误必拦; -l: 格式债仅警告)"
  # 2026-08-02: 实测全仓库 226 个未格式化文件但 0 语法错误。
  # 格式债是存量问题，不应阻塞 deploy；语法错误才拦。
  SYNTAX_ERR=$(gofmt -e $(git ls-files '*.go' | grep -v '^migrations/') 2>&1 | grep -c '^ERR' || true)
  if [ "${SYNTAX_ERR:-0}" != "0" ]; then
    echo "GATE_FAIL: gofmt 语法错误 \ 处："
    gofmt -e $(git ls-files '*.go' | grep -v '^migrations/') 2>&1 | grep '^ERR' | head -20
    return 1
  fi
  echo "[gate] gofmt syntax OK (0 errors; 存量格式债不影响编译)"

  echo "[gate] go vet (model/controller/middleware/router/di/repository)"
  # 2026-08-02: 2 核生产机，go vet 也会编译，必须 -p 1 限核 + nice，避免压垮线上
  # migrations/ 下有独立 package main 的 .go（migrate_oauth_v2.go 等），不参与主模块编译
  nice -n 15 go vet -p 1 ./model/... ./controller/... ./middleware/... ./router/... ./di/... ./repository/... 2>&1 | head -30 || {
    echo "GATE_FAIL: go vet 未通过（见上）"
    return 1
  }
  echo "[gate] OK"
  return 0
}

# ---- preflight-only 模式 ----
if [ "${1:-}" = "preflight" ]; then
  preflight || exit 1
  echo "PREFLIGHT_OK"
  exit 0
fi

WITH_FRONTEND=0
if [ "${1:-}" = "--with-frontend" ]; then
  WITH_FRONTEND=1
fi

# ---- 前置校验 ----
preflight || { echo "PREFLIGHT_FAIL"; exit 1; }

# ---- v2 新增：代码门禁（改完代码先过这里，别直接 build）----
code_gate || { echo "CODE_GATE_FAIL"; exit 1; }

# ---- v2 新增：强制数据库备份（build 前，任何一次部署前都要有备份）----
if [ -x "$BACKUP_SCRIPT" ]; then
  echo "[deploy] pre-deploy DB backup (ft-backup-db.sh)"
  systemctl start ft-backup-db.service || bash "$BACKUP_SCRIPT" || {
    echo "PRE_DEPLOY_BACKUP_FAIL: 数据库备份失败，中止部署"
    exit 1
  }
  LATEST=$(ls -1t /var/backups/fasttoken/daily/*.dump 2>/dev/null | head -1)
  [ -n "$LATEST" ] && echo "[deploy] backup OK: $(basename "$LATEST")"
else
  echo "PRE_DEPLOY_BACKUP_SKIP: $BACKUP_SCRIPT 不存在（备份未配置，强烈建议先装）"
fi

# 可选：部署前自动重建前端（消除“忘记 npm run build”这一最常见失误）
if [ "$WITH_FRONTEND" = "1" ]; then
  echo "[deploy] building frontend (--with-frontend)"
  cd "$WEB"
  npm run build || { echo "FRONTEND_BUILD_FAIL"; exit 1; }
  cd /opt/fasttoken
  if ! grep -rq "$I18N_MARK" "$WEB/dist" 2>/dev/null; then
    echo "FRONTEND_BUILD_FAIL: 重建后 dist 仍缺 '$I18N_MARK'"
    exit 1
  fi
fi

echo "[deploy] backup binary -> $BAK"
cp -p "$BIN" "$BAK"

echo "[deploy] build (embeds web/default/dist via go embed)"
cd /opt/fasttoken
# 必须用 -a 强制无缓存重建：go embed 的构建缓存有时会不随 dist 变更失效。
# v2: -p 1 + nice 限核（2026-08-02 实测 go vet 在 2 核上 load 冲到 11，build 同理）
# 版本戳：git commit short hash 注入 common.Version（取代恒为 v0.0.0）
GIT_SHA=$(git rev-parse --short HEAD 2>/dev/null || echo "nogit-$TS")
nice -n 15 go build -a -p 1 -ldflags "-X common.Version=$GIT_SHA" -o /tmp/ft.new . || { echo "BUILD_FAIL"; exit 1; }

# 构建后校验：新二进制内嵌的 dist 必须含 i18n 标记（确认 embed 成功，未嵌旧/坏 dist）
echo "[deploy] verifying embedded dist contains i18n marker"
if ! grep -q "$I18N_MARK" /tmp/ft.new 2>/dev/null; then
  echo "BUILD_FAIL: 新二进制未嵌入 '$I18N_MARK'，dist 可能未更新"
  rm -f /tmp/ft.new
  exit 1
fi

echo "[deploy] swap binary"
mv /tmp/ft.new "$BIN"
chmod +x "$BIN"

# 收紧密钥文件权限（非破坏性，仅提权）
chmod 600 /opt/fasttoken/.env 2>/dev/null || true
chmod 600 /opt/fasttoken/cert/wechat/pub_key.pem 2>/dev/null || true

echo "[deploy] restart fasttoken"
systemctl restart fasttoken
sleep 8

# v2: smoke 改为真实回源健康检查 + 业务探针双保险
echo "[deploy] smoke: /health (real upstream) + /api/payment/status (retry up to 5x)"
ok=0
for n in 1 2 3 4 5; do
  if curl -fsS --max-time 12 http://127.0.0.1:3000/api/payment/status | grep -q '"ready":true' \
     && curl -fsS --max-time 5 http://127.0.0.1:3000/api/status >/dev/null 2>&1; then
    ok=1; break
  fi
  sleep 2
done
if [ "$ok" != "1" ]; then
  echo "SMOKE_FAIL -> rollback to $BAK"
  mv "$BAK" "$BIN"
  systemctl restart fasttoken
  sleep 8
  curl -fsS --max-time 12 http://127.0.0.1:3000/api/payment/status || echo "ROLLBACK_FAIL"
  exit 1
fi
echo "DEPLOY_OK ($TS) git=$GIT_SHA"
