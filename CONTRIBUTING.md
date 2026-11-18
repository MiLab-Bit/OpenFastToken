# Contributing to OpenFastToken

Thanks for your interest in improving OpenFastToken. This repository builds on [New API](https://github.com/QuantumNous/new-api); please review the existing README and upstream project before proposing changes.

## Before opening a pull request

1. Search existing issues and pull requests to avoid duplicate work.
2. For feature requests or changes to billing behavior, open an issue first describing the use case, expected behavior, and compatibility implications.
3. Keep changes focused. Avoid unrelated formatting, generated artifacts, or dependency upgrades.
4. Document changes to user-visible behavior, configuration, API compatibility, and migrations.

## Development and verification

Follow the setup and verification commands in the repository's current README and relevant component directories. Do not assume that a single local test covers the complete gateway.

For each pull request, include:
- A short problem statement and scope of the change.
- How you verified it, with the exact commands and results; clearly mark checks you could not run.
- Relevant screenshots or API examples for UI/API changes.
- Any breaking changes, migration steps, and rollback considerations.

## Billing and security-sensitive changes

Changes to personal/enterprise wallet selection, quota deduction, refunds, redemption, or top-ups require extra care. Describe the expected accounting invariants, concurrency and retry behavior, and tests for failure paths. Never include production tokens, payment credentials, customer data, or private logs in issues or pull requests.

For suspected security vulnerabilities, **do not publish exploit details in a public issue**. Check the repository's security reporting options and contact the maintainers privately through an available channel.

## Licensing and attribution

This repository is licensed under AGPL-3.0 and builds on upstream code. Preserve applicable copyright and license notices, and identify upstream-derived changes where relevant. Contributions must be compatible with the project's licensing obligations.

## Review

Maintainers may request tests, documentation, or scope adjustments before merging. Opening a pull request does not guarantee acceptance.
