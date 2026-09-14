# Contributing to VENOMBusiness

Thanks for considering a contribution. VENOMBusiness is an early-alpha open-source project, so clear bug reports, reproducible test cases, documentation and focused pull requests are especially valuable.

## Before opening a pull request

1. Search existing issues and pull requests first.
2. Keep a pull request focused on one problem or feature.
3. Do not include secrets, `.env` files, local databases, backups, generated logs, or customer data.
4. Preserve company/workspace data isolation in every new query and endpoint.
5. Prefer backwards-compatible schema changes and add migrations when the database changes.

## Development

Backend:

```bash
cd apps/api
cp .env.example .env
go mod tidy
go test ./...
go run ./cmd/server
```

Frontend:

```bash
cd apps/web
npm install
npm run build
npm run dev
```

## Database compatibility

Changes touching persistence should be considered for both SQLite and PostgreSQL. Engine-specific migration SQL belongs in the appropriate migration directory.

## Pull request checklist

- [ ] The change has a clear purpose.
- [ ] Existing behavior is not unintentionally broken.
- [ ] Database changes include migrations where required.
- [ ] Sensitive information is not committed.
- [ ] Documentation is updated when behavior or setup changes.
- [ ] Backend tests/build and frontend build have been run where possible.

## Style

Use `gofmt` for Go code and keep TypeScript strict and readable. Avoid unnecessary abstractions; VENOMBusiness intentionally uses a modular-monolith architecture.
