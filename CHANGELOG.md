# Changelog

All notable changes to VENOMBusiness are documented here.

## [0.5.1-alpha] - 2026-09-14

### Changed
- Froze new business features for GitHub public-release preparation.
- Reworked README for public onboarding and accurate early-alpha positioning.
- Added contribution, security and community documentation.
- Added GitHub issue and pull-request templates.
- Hardened `.gitignore` rules for secrets, databases, backups and generated artifacts.
- Added a public release checklist and alpha release notes.

### Security
- Added explicit guidance against committing `.env`, session secrets, local database files and backups.
- Added a private-first vulnerability disclosure policy.

## [0.5.0] - 2026-09-14

### Added
- PostgreSQL database driver and engine-specific migrations.
- SQLite/PostgreSQL query compatibility layer.
- Production configuration validation.
- Database readiness diagnostics.
- Backup/restore tooling and SQLite → PostgreSQL migration utility.
- CI foundations for SQLite and PostgreSQL.

## [0.4.0]

### Added
- Business profile.
- Expense management.
- Financial reporting.
- PDF quotations, invoices and payment receipts.

## [0.3.0]

### Added
- Products/services, quotations, invoices and payments.

## [0.2.0]

### Added
- Authentication, company onboarding, RBAC foundations, dashboard and customer management.
