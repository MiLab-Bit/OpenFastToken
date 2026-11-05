# OpenFastToken — Security policy and Code Review context

## Actual security boundaries
High-priority threats: wallet debits/refunds atomically bound to original funding source, payment webhook signature/replay, tenant isolation, admin role checks, redeem concurrency, idempotent recharge, rate limits, token relay and API key/credential exposure. Do not enable real payment tests.

## Review requirements
- Use isolated/synthetic test data. Never commit or publish credentials, financial data, user data, payment keys, wallet private keys or unapproved vendor content.
- Findings need reviewed SHA, vulnerable input boundary, affected invariant, concrete safe reproduction, severity and proposed narrow correction. No speculative blanket security claims.
- Do not remove verified defense-in-depth validation merely as duplication. Green CI is not evidence of licensing, production readiness or effective independent human approval.
- Privately disclose sensitive exploitable issues to repository owner using an established channel or GitHub private advisory if enabled; public PR discussions must omit secrets and exploitation details. No response SLA is implied.
- ChatGPT implements; Codex performs second-agent review; owner retains merge, release, external credential and production authority.
