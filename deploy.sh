#!/usr/bin/env bash
# FastToken 部署脚本：配置安全 + 幂等 + 自带冒烟与回滚 + 部署前置校验
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

# ---- 环境（同时兼容直接执行与 systemd-run；后者会剥离环境变量）----
# 显式写出 Go 与 node/npm 路径，避免 systemd-run 下 PATH 被清空导致命令找不到。
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
# 必须用 -a 强制无缓存重建：go embed 的构建缓存有时会不随 dist 变更失效，
# 导致二进制嵌的还是旧 dist（翻译缺失、页面回退成英文 key）。-a 保证重新嵌入当前 dist。
go build -a -o /tmp/ft.new . || { echo "BUILD_FAIL"; exit 1; }

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

echo "[deploy] smoke: /api/payment/status (retry up to 5x)"
ok=0
for n in 1 2 3 4 5; do
  if curl -fsS --max-time 12 http://127.0.0.1:3000/api/payment/status | grep -q '"ready":true'; then
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
echo "DEPLOY_OK ($TS)"
