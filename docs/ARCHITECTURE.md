# Architecture

VENOMBusiness is a **modular monolith**. Modules own their domain logic and expose narrow service/repository boundaries. The first release intentionally avoids microservices.

## Backend layers

- `cmd/server`: composition root
- `internal/config`: environment configuration
- `internal/httpx`: HTTP middleware and JSON helpers
- `internal/platform`: database, security, sessions and cross-cutting infrastructure
- `internal/modules`: business modules
- `migrations`: ordered schema migrations

## Module rule

A module may depend on platform abstractions and explicitly exported services from another module. It must not reach into another module's persistence details.

## Data strategy

SQLite is the default for easy self-hosting. Database access is written against `database/sql` and kept behind repositories so a PostgreSQL adapter can be introduced without rewriting business rules.

## Authentication

Passwords are hashed using Argon2id. Browser sessions use high-entropy opaque tokens stored in HttpOnly cookies. Only a SHA-256 digest of the session token is stored server-side.

## Versioning

Public endpoints live under `/api/v1`. Database changes are additive migrations. Breaking API changes require a new major API namespace.
