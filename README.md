# VENOMBusiness

<p align="center">
  <strong>Open-source Business OS for small businesses and startups.</strong>
</p>

<p align="center">
  A modern, self-hosted platform for managing customers, products, quotations, invoices, payments, expenses, and business finances from one place.
</p>

<p align="center">
  <img src="https://img.shields.io/badge/version-v0.5.1--alpha-orange" alt="Version">
  <img src="https://img.shields.io/badge/Go-Backend-00ADD8?logo=go&logoColor=white" alt="Go">
  <img src="https://img.shields.io/badge/React-Frontend-61DAFB?logo=react&logoColor=black" alt="React">
  <img src="https://img.shields.io/badge/TypeScript-Ready-3178C6?logo=typescript&logoColor=white" alt="TypeScript">
  <img src="https://img.shields.io/badge/SQLite-Supported-003B57?logo=sqlite&logoColor=white" alt="SQLite">
  <img src="https://img.shields.io/badge/PostgreSQL-Supported-4169E1?logo=postgresql&logoColor=white" alt="PostgreSQL">
  <img src="https://img.shields.io/badge/Self--Hosted-Yes-success" alt="Self Hosted">
</p>

---

## ⚠️ Early Alpha

VENOMBusiness is currently under active development.

The project is available for testing, development, experimentation, and community feedback, but it should not yet be used as the sole system of record for critical financial or business data without maintaining regular backups.

The public API, database schema, and application behavior may still change before the stable `v1.0` release.

---

## 🖥️ Preview

<p align="center">
  <img src="img/dashboard.png" alt="VENOMBusiness Dashboard" width="100%">
</p>

VENOMBusiness provides a clean and modern workspace for managing day-to-day business operations without requiring a complicated enterprise ERP setup.

The project follows a simple philosophy:

> **Powerful business software without the DevOps headache.**

---

## ✨ Features

VENOMBusiness currently provides:

* 🏢 Company / workspace management
* 👥 Customer management
* 📦 Products & services
* 📄 Quotations
* 🧾 Invoices
* 💳 Payment recording
* 💸 Expense management
* 📊 Financial overview
* 📑 PDF quotations
* 📑 PDF invoices
* 🧾 Payment receipts
* 🔐 Authentication
* 🛡️ Role-based access control
* 📝 Audit logging
* 💾 Backup & restore utilities
* ❤️ Health & readiness endpoints
* 🗃️ SQLite support
* 🐘 PostgreSQL support
* 🔄 SQLite → PostgreSQL migration utility
* 🏠 Self-hosted deployment

---

# 📸 Screenshots

## 🔐 Secure Login

<p align="center">
  <img src="img/login.png" alt="VENOMBusiness Login" width="100%">
</p>

VENOMBusiness uses server-side authentication sessions with security controls designed for self-hosted business applications.

---

## 📊 Business Dashboard

<p align="center">
  <img src="img/dashboard.png" alt="VENOMBusiness Dashboard" width="100%">
</p>

Get a quick overview of customers, catalog items, invoiced revenue, collected payments, outstanding receivables, and other business metrics.

---

## 🏢 Business Profile

<p align="center">
  <img src="img/business-profile.png" alt="VENOMBusiness Business Profile" width="100%">
</p>

Configure your business identity, legal information, contact details, tax information, banking information, and payment instructions.

---

## 👥 Customer Management

<p align="center">
  <img src="img/customer-record.png" alt="VENOMBusiness Customer Management" width="100%">
</p>

Keep customer information organized inside your company workspace.

Customer data is scoped to the authenticated company to provide proper workspace isolation.

---

## 📦 Products & Services

<p align="center">
  <img src="img/product-record.png" alt="VENOMBusiness Products and Services" width="100%">
</p>

Manage reusable products and services with SKU, descriptions, pricing, tax configuration, and active status.

---

## 📄 Quotations

<p align="center">
  <img src="img/quotation-record.png" alt="VENOMBusiness Quotation Management" width="100%">
</p>

Create professional quotations, manage their lifecycle, and convert accepted quotations directly into invoices.

---

# 🏗️ Architecture

VENOMBusiness uses a **modular monolith** architecture.

The goal is to keep deployment simple while maintaining clear boundaries between business domains.

```text
VENOMBusiness
│
├── React + TypeScript
│        │
│        │ REST / JSON
│        ▼
│
├── Go Backend
│
│   ├── Authentication
│   ├── Company / Workspace
│   ├── Customers
│   ├── Products
│   ├── Quotations
│   ├── Invoices
│   ├── Payments
│   ├── Expenses
│   ├── Reports
│   └── Audit
│
└── Database Layer
     │
     ├── SQLite
     │
     └── PostgreSQL
```

We intentionally avoid premature microservices.

For small and medium deployments, a modular monolith provides simpler operations, easier development, and fewer infrastructure requirements.

---

# 🛠️ Technology Stack

### Backend

* Go
* REST / JSON API
* Server-side sessions
* Argon2id password hashing
* RBAC
* Database migrations
* Audit logging

### Frontend

* React
* TypeScript
* Vite
* Tailwind CSS
* Lucide Icons

### Database

VENOMBusiness supports two database strategies.

### SQLite

Recommended for:

* Personal use
* Freelancers
* Small businesses
* Local installations
* Small teams
* Development

SQLite requires almost zero database administration.

### PostgreSQL

Recommended for:

* Production deployments
* Startups
* Multi-user installations
* Growing businesses
* Larger datasets
* Higher concurrency environments

The application is designed so businesses can start small with SQLite and migrate to PostgreSQL later.

---

# 🚀 Quick Start

## Requirements

For development you will need:

```text
Go
Node.js
npm
Git
```

PostgreSQL is optional when using SQLite.

---

## 1. Clone Repository

```bash
git clone https://github.com/rzfachri99/VENOMBusiness.git
cd VENOMBusiness
```

---

# ⚙️ Backend Setup

Enter the API directory:

```bash
cd apps/api
```

Create your environment configuration.

### Windows

```bat
copy .env.example .env
```

### Linux / macOS

```bash
cp .env.example .env
```

Install Go dependencies:

```bash
go mod tidy
```

Start the API:

```bash
go run ./cmd/server
```

The backend will normally be available at:

```text
http://localhost:8080
```

Database migrations are executed automatically when the application starts.

---

# 🎨 Frontend Setup

Open another terminal:

```bash
cd apps/web
```

Install dependencies:

```bash
npm install
```

Start the development server:

```bash
npm run dev
```

Open:

```text
http://localhost:5173
```

You can now create your first VENOMBusiness account.

---

# 🗃️ SQLite Configuration

SQLite is the default and easiest way to start VENOMBusiness.

Example:

```env
VENOM_DATABASE_DRIVER=sqlite
VENOM_DATABASE_URL=file:venombusiness.db?_pragma=foreign_keys(1)&_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)
```

No separate database server is required.

---

# 🐘 PostgreSQL Configuration

For PostgreSQL:

```env
VENOM_DATABASE_DRIVER=postgres
VENOM_DATABASE_URL=postgres://venom:YOUR_PASSWORD@localhost:5432/venombusiness?sslmode=disable
```

For production environments, TLS/SSL should be enabled.

Example:

```env
VENOM_ENV=production

VENOM_APP_ORIGIN=https://business.example.com

VENOM_COOKIE_SECURE=true

VENOM_DATABASE_DRIVER=postgres

VENOM_DATABASE_URL=postgres://venom:STRONG_PASSWORD@db.example.com:5432/venombusiness?sslmode=require
```

VENOMBusiness performs additional configuration validation when running in production mode.

---

# 🐳 PostgreSQL Development with Docker

A development PostgreSQL service is included.

Run:

```bash
docker compose -f docker-compose.dev.yml up -d postgres
```

Then configure VENOMBusiness to use PostgreSQL through `.env`.

Docker is optional.

VENOMBusiness does **not** require Docker for normal SQLite installations.

---

# 🔄 SQLite → PostgreSQL Migration

VENOMBusiness includes an administration utility called `venomctl`.

Build it:

### Windows

```bat
cd apps\api
go build -o venomctl.exe ./cmd/venomctl
```

### Linux / macOS

```bash
cd apps/api
go build -o venomctl ./cmd/venomctl
```

Configure your existing SQLite database and PostgreSQL destination.

Example:

```env
VENOM_DATABASE_DRIVER=sqlite

VENOM_DATABASE_URL=file:venombusiness.db?_pragma=foreign_keys(1)

VENOM_TARGET_DATABASE_URL=postgres://venom:password@localhost:5432/venombusiness?sslmode=disable
```

Then run:

### Windows

```bat
venomctl.exe migrate-to-postgres
```

### Linux / macOS

```bash
./venomctl migrate-to-postgres
```

The migration runs inside a PostgreSQL transaction.

The destination database should be empty before migration.

Always create a backup before performing database migration.

---

# 💾 Backup & Restore

## SQLite Backup

```bash
venomctl backup venom-backup.db
```

Windows:

```bat
venomctl.exe backup venom-backup.db
```

SQLite backups use a consistent database snapshot.

---

## SQLite Restore

Stop the VENOMBusiness API before restoring the database.

```bash
venomctl restore venom-backup.db
```

Windows:

```bat
venomctl.exe restore venom-backup.db
```

---

## PostgreSQL Backup

PostgreSQL backups use `pg_dump`.

```bash
venomctl backup venom-production.backup
```

`pg_dump` must be available in your system `PATH`.

PostgreSQL restoration uses `pg_restore`.

---

# ❤️ Health Checks

VENOMBusiness provides health endpoints for deployment monitoring.

Basic health:

```text
GET /healthz
```

Readiness:

```text
GET /readyz
```

The readiness endpoint can report information such as:

```json
{
  "status": "ready",
  "database": {
    "driver": "postgres",
    "migration_version": 3,
    "open_connections": 3,
    "in_use": 1,
    "idle": 2
  }
}
```

These endpoints can be used by reverse proxies, container platforms, monitoring systems, and orchestration environments.

---

# 🔐 Security

Security is treated as a core part of VENOMBusiness architecture.

Current security foundations include:

* Argon2id password hashing
* Server-side sessions
* Secure session cookies
* Role-based authorization
* Company/workspace isolation
* Login rate limiting
* CSRF same-origin protection
* HTTP security headers
* Request body size limits
* Database foreign keys
* Audit logging
* Production configuration validation

Please do not publicly disclose security vulnerabilities through GitHub Issues.

See:

```text
SECURITY.md
```

for responsible disclosure instructions.

---

# 📁 Project Structure

```text
VENOMBusiness/
│
├── .github/
│   ├── ISSUE_TEMPLATE/
│   ├── workflows/
│   └── pull_request_template.md
│
├── apps/
│   ├── api/
│   │   ├── cmd/
│   │   ├── internal/
│   │   └── migrations/
│   │
│   └── web/
│       └── src/
│
├── docs/
│
├── img/
│
├── CHANGELOG.md
├── CODE_OF_CONDUCT.md
├── CONTRIBUTING.md
├── LICENSE
├── README.md
├── SECURITY.md
└── VERSION
```

---

# 🗺️ Roadmap

VENOMBusiness is being developed incrementally.

### v0.5.x

Core Business OS foundation.

* Authentication
* Workspace
* Customers
* Products
* Quotations
* Invoices
* Payments
* Expenses
* Reports
* SQLite
* PostgreSQL
* Backup & restore

### v0.6

Inventory & Stock Engine.

Planned:

* Warehouses
* Stock in
* Stock out
* Stock adjustments
* Stock movement ledger
* Low-stock alerts
* Suppliers
* Purchasing foundation
* COGS
* Inventory valuation

### v0.7

CRM.

Planned:

* Leads
* Opportunities
* Sales pipeline
* Activities
* Follow-up reminders
* Customer timeline

### v0.8

Purchasing & operational expansion.

### v0.9

Production hardening and stabilization.

### v1.0

First stable release.

---

# 🤝 Contributing

Contributions are welcome.

You can contribute by:

* Reporting bugs
* Suggesting features
* Improving documentation
* Testing SQLite and PostgreSQL
* Improving UI/UX
* Reviewing security
* Submitting pull requests

Please read:

```text
CONTRIBUTING.md
```

before submitting a pull request.

---

# 🐛 Bug Reports

Found a bug?

Please use the GitHub Bug Report template and include:

* VENOMBusiness version
* Operating system
* Database driver
* Steps to reproduce
* Expected behavior
* Actual behavior
* Relevant logs

Never include passwords, tokens, `.env` contents, production database dumps, or other secrets.

---

# 💡 Feature Requests

Ideas are welcome.

VENOMBusiness aims to remain useful for small businesses without becoming unnecessarily complicated.

Feature requests should ideally describe:

1. The business problem.
2. Who experiences the problem.
3. The proposed workflow.
4. Why the feature belongs in the core application.

---

# 🌍 Project Philosophy

VENOMBusiness is built around a few principles:

### Self-hosted first

Your business application should be able to run without mandatory third-party cloud services.

### Simple deployment

Small businesses should not need a DevOps team just to run business software.

### Start small, grow later

Start with SQLite.

Move to PostgreSQL when your business grows.

### Modern engineering without unnecessary complexity

VENOMBusiness uses modern technology while avoiding architecture that creates operational complexity without clear benefits.

### Open development

Development happens publicly so users and contributors can follow the evolution of the project.

---

# 📜 License

VENOMBusiness is open-source software.

See the included:

```text
LICENSE
```

file for the complete license terms.

---

# ⭐ Support the Project

If you find VENOMBusiness useful, consider giving the repository a ⭐ on GitHub.

It helps more developers, freelancers, startup founders, and small-business owners discover the project.

You can also help by:

* Testing the application
* Reporting bugs
* Improving documentation
* Suggesting business workflows
* Contributing code
* Sharing the project

---

<p align="center">
  <strong>VENOMBusiness</strong>
</p>

<p align="center">
  Open-source Business OS for small businesses and startups.
</p>

<p align="center">
  Built with Go, React, TypeScript, SQLite and PostgreSQL.
</p>
