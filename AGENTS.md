# OpenFastToken — Agent and Codex PR review guide

## Repository contract
Go/Gin/GORM + React AI gateway with personal and enterprise wallets, prepaid quota, relay, billing, payment callbacks, Postgres/Redis, OAuth and provider integrations. Keep authority in server-side financial records, not browser or LLM output.

## Implementation and review
- Inspect actual sources, contract documentation, prior test or CI before changing code. Do not invent missing business capabilities, new layers or obsolete compatibility requirements.
- Describe intent and exclusions, state owner, domain invariants, actual validation with SHA, and justified complexity. Review meaningful changes, not arbitrary line count.
- Identify demonstrable defects with file/function, reproduction, inputs, observed impact and minimal fix. When evidence is incomplete, label it. No finding is a valid review outcome.
- Preserve existing release/approval gates, scoped data access and tests. Keep AI-authored implementation distinct from Codex second-agent review, and from owner merge/production authorization.
- Consult [SECURITY.md](SECURITY.md) for this repository's threat model. Do not ask Codex to auto-fix, merge or deploy.
