# VENOMBusiness Databases

VENOMBusiness v0.5 officially supports SQLite and PostgreSQL.

## SQLite
Use SQLite for personal use, prototypes, small businesses, and simple single-server deployments.

```env
VENOM_DATABASE_DRIVER=sqlite
VENOM_DATABASE_URL=file:venombusiness.db?_pragma=foreign_keys(1)&_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)
```

## PostgreSQL
Use PostgreSQL for production teams, higher concurrency, managed database deployments, and businesses that expect to scale.

Development example:

```env
VENOM_DATABASE_DRIVER=postgres
VENOM_DATABASE_URL=postgres://venom:venom_dev_password@localhost:5432/venombusiness?sslmode=disable
```

Production must use TLS. `sslmode=disable` is rejected when `VENOM_ENV=production`.

## Local PostgreSQL

```bash
docker compose -f docker-compose.dev.yml up -d postgres
```

Then use the development PostgreSQL URL shown above.

## Health diagnostics

`GET /readyz` verifies the database and reports the current driver, migration version, and connection-pool state.

## Backups

Build the administration utility:

```bash
cd apps/api
go build -o venomctl ./cmd/venomctl
```

SQLite online backup:

```bash
./venomctl backup backup.db
```

SQLite backup uses `VACUUM INTO` to create a consistent snapshot. Stop the API before restoring:

```bash
./venomctl restore backup.db
```

For PostgreSQL, `venomctl` delegates to `pg_dump` and `pg_restore`, which must be installed and available in PATH.

## Migrate SQLite to PostgreSQL

Create an empty PostgreSQL database, keep the SQLite database configured as the source, and set:

```env
VENOM_DATABASE_DRIVER=sqlite
VENOM_DATABASE_URL=file:venombusiness.db?_pragma=foreign_keys(1)&_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)
VENOM_TARGET_DATABASE_URL=postgres://venom:password@localhost:5432/venombusiness?sslmode=disable
```

Then run:

```bash
./venomctl migrate-to-postgres
```

The destination must be empty. VENOMBusiness applies PostgreSQL migrations first and copies data in foreign-key-safe order inside one transaction.
