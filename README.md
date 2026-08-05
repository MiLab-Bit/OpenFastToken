# OpenFastToken

> Unified AI Gateway with **dual-wallet billing** (personal + enterprise), built on the [new-api](https://github.com/QuantumNous/new-api) codebase.

OpenFastToken aggregates multiple model providers behind OpenAI / Claude / Gemini compatible interfaces, and adds production-grade billing features on top of New API — most notably a **dual-wallet system** for enterprise accounts, plus check-in, redemption codes, online top-ups, deployments, prefill groups, group ratios and performance/rankings dashboards.

## ✨ Highlights

- **Dual-wallet billing (Phase 1, 2026-08)** — every user has a personal wallet (`users.quota`) and, if they belong to an enterprise, an enterprise wallet (member balance granted from the enterprise main wallet):
  - Billing prefers the **enterprise wallet** and falls back to the personal wallet when the enterprise balance is insufficient; a single request is **never split** across wallets.
  - **Refunds always return to the original wallet.**
  - Enterprise certification auto-upgrades the submitting user to the highest membership tier (platinum) — no separate upgrade logic.
  - Enterprise admins can **grant / recycle / view history** for member quota, and **self-recharge** the main wallet via WeChat / Alipay.
  - Platform admins can credit (授信) the enterprise main wallet.
  - All quota updates are atomic and never go negative (conditional UPDATE), with cross-tenant operations rejected server-side.
- **WeChat Pay (Native QR) & Alipay** top-up with async notify, order polling, and automatic reconciliation (补单).
- Check-in (签到), redemption codes (兑换), referral rebates (推荐返利), recharge gift bonuses (充值赠送).
- Multi-language frontend (zh / en / fr / ja / ru / vi / ar) with runtime "config-as-data" i18n overrides.
- Enterprise registration / certification / member management (企业认证与成员管理).
- Skin theming system, performance dashboards, rankings, deployments management.

## 🏗 Architecture

```
Browser / Client
      │  HTTPS (nginx: TLS, rate-limit, security headers)
      ▼
   FastToken (Go binary, embeds frontend dist via go:embed)
      ├── PostgreSQL   ← SQL_DSN
      └── Redis        ← REDIS_CONN_STRING
```

- Backend: Go 1.25, GORM, Gin, PostgreSQL / MySQL / SQLite (SQLite for tests).
- Frontend: React 18 + rsbuild (web/default), shadcn-style UI, react-i18next.
- Deployment: single binary embeds the frontend `dist`; `deploy.sh` provides preflight checks, DB backup, smoke test and automatic rollback.

## 🚀 Quick Start

### 1. Configure

Copy `.env.example` to `.env` and fill in your values:

```bash
cp .env.example .env
# edit .env: SQL_DSN / REDIS_CONN_STRING / payment keys, etc.
```

### 2. Build frontend (optional — dist is prebuilt and embedded)

```bash
cd web/default
npm install
npm run build
```

### 3. Build & run backend

```bash
go build -o fasttoken .
./fasttoken          # listens on :3000 (configurable via PORT / bind address)
```

Health check: `curl http://127.0.0.1:3000/api/payment/status`

### 4. Deploy to production

```bash
bash deploy.sh                # standard deploy (preflight + backup + smoke + rollback)
bash deploy.sh --with-frontend  # also rebuild the frontend first
bash deploy.sh preflight      # only run preflight checks
```

## 🔑 Dual-Wallet API Reference

### Platform admin

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/enterprise/:id/wallet` | View enterprise main wallet |
| POST | `/api/enterprise/:id/wallet/recharge` | Credit (授信) the main wallet `{"quota":10000000,"remark":"..."}` |

### Enterprise admin / member (tenant self-service)

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/user/tenant/wallet` | My wallet view (member quota + admin main wallet) |
| POST | `/api/user/tenant/wallet/grant` | Grant quota to a member (admin) |
| POST | `/api/user/tenant/wallet/recycle` | Recycle quota from a member (admin) |
| GET | `/api/user/tenant/wallet/txns` | Wallet transaction history (admin, paginated) |
| POST | `/api/user/tenant/wallet/topup` | Self-recharge main wallet (WeChat/Alipay, admin) |
| GET | `/api/user/tenant/wallet/topup/status` | Poll top-up order status |

> Security: the enterprise id is always resolved server-side from the authenticated session — the client never sends it. Cross-tenant grants are rejected by the backend.

## 📦 Project Layout

| Path | Description |
|------|-------------|
| `controller/` | HTTP handlers (billing, wallet, top-up, enterprise, admin) |
| `model/` | GORM models & DB access (dual-wallet, top-up, tasks, logs) |
| `service/` | Billing engine: `FundingSource` abstraction, `CompositeFunding` (enterprise-first fallback), billing sessions |
| `relay/` | Model relay channels & task billing (all 4 deduction paths unified on dual-wallet) |
| `setting/` | Runtime configuration (config-as-data: options / pricing / i18n from DB) |
| `web/default/` | React frontend |
| `migrations/` | SQL migrations |
| `common/` | Shared utilities, quota math, validators |

## 🧪 Tests

```bash
go test ./model/ ./service/ ./controller/ ./router/
```

The dual-wallet test suite (`model/dual_wallet_test.go`, `service/dual_wallet_test.go`) covers the full wallet lifecycle: credit → grant → consume → refund → recycle, negative-balance protection, callback routing (enterprise vs personal), refund-original-path, cross-tenant rejection, and composite funding selection (enterprise-first / personal fallback / pure-enterprise user / streaming reserve & rollback).

## 📄 License

[AGPL-3.0](LICENSE) — see the LICENSE file. This project is a fork of the AGPL-licensed [new-api](https://github.com/QuantumNous/new-api); AGPL obligations apply to derivative works.

---

*OpenFastToken — production-grade dual-wallet billing for AI gateways.*
