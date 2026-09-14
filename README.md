# VENOMBusiness

**Open-source Business OS for small businesses and startups.**

VENOMBusiness is a self-hosted business management platform built with **Go**, **React**, and **TypeScript**. It is designed to stay simple for small teams while keeping an architecture that can grow with more demanding deployments.

> **Early Alpha — v0.5.1**
>
> VENOMBusiness is under active development. Do not use it as the sole system of record for critical financial data without maintaining regular backups and validating the deployment for your environment.

## Why VENOMBusiness?

- Self-hosted first; no mandatory cloud dependency
- SQLite for simple deployments
- PostgreSQL for production and multi-user deployments
- Modular-monolith backend architecture
- Server-side opaque sessions and Argon2id password hashing
- Company-scoped business data and RBAC foundations
- Built for gradual growth from a small business tool into a broader Business OS

## Current modules

- Authentication and business onboarding
- Customer management
- Products & services
- Quotations
- Invoices
- Payments
- Business profile
- Expense management
- Financial summary/reporting
- PDF quotations, invoices, and payment receipts
- Audit-log foundation
- SQLite/PostgreSQL database layer
- Backup/restore and SQLite → PostgreSQL migration utilities

## Technology

| Layer | Technology |
| --- | --- |
| Backend | Go 1.27 |
| Frontend | React 19 + TypeScript + Vite |
| Database | SQLite / PostgreSQL |
| Auth | Opaque server-side sessions |
| Password hashing | Argon2id |
| Architecture | Modular monolith |

## Quick start with SQLite

Requirements: Go 1.27+, Node.js LTS, and npm.

```bash
git clone https://github.com/YOUR_USERNAME/VENOMBusiness.git
cd VENOMBusiness
```

Start the API:

```bash
cd apps/api
cp .env.example .env
go mod tidy
go run ./cmd/server
```

In another terminal, start the frontend:

```bash
cd apps/web
npm install
npm run dev
```

Open `http://localhost:5173`.

> Windows CMD users can use `copy .env.example .env` instead of `cp`.

## PostgreSQL development

Start PostgreSQL with Docker:

```bash
docker compose -f docker-compose.dev.yml up -d postgres
```

Then set these values in `apps/api/.env`:

```env
VENOM_DATABASE_DRIVER=postgres
VENOM_DATABASE_URL=postgres://venom:venom_dev_password@localhost:5432/venombusiness?sslmode=disable
```

Start the API normally; PostgreSQL migrations are applied automatically.

For production, use TLS and a strong secret configuration. See [`docs/DATABASES.md`](docs/DATABASES.md).

## Screenshots

Screenshots will be added as the public alpha UI is finalized. Contributions that improve onboarding documentation and screenshots are welcome.

## Repository layout

```text
apps/
├── api/          Go API, modules, migrations and venomctl
└── web/          React + TypeScript frontend

docs/             Architecture, database and roadmap docs
.github/           CI, issue templates and pull request template
```

## Database operations

The `venomctl` utility supports operational tasks including backup/restore and SQLite → PostgreSQL migration.

```bash
cd apps/api
go build -o venomctl ./cmd/venomctl
```

See [`docs/DATABASES.md`](docs/DATABASES.md) before running database operations.

## Project status

VENOMBusiness is currently an **early alpha**. APIs, schemas, module boundaries and UI flows may change before `v1.0.0`.

Roadmap highlights:

- `v0.6` — Inventory & Stock Engine
- `v0.7` — CRM
- `v0.8` — Purchasing
- `v0.9` — broader production hardening and beta stabilization
- `v1.0` — stable release target

See [`docs/ROADMAP.md`](docs/ROADMAP.md).

## Security

Please do not disclose suspected vulnerabilities in public issues. Follow [`SECURITY.md`](SECURITY.md) for reporting guidance.

## Contributing

Bug reports, documentation improvements and code contributions are welcome. Start with [`CONTRIBUTING.md`](CONTRIBUTING.md).

## License

Licensed under the [Apache License 2.0](LICENSE).
