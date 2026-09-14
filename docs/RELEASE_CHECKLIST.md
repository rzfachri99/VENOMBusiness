# Public Alpha Release Checklist

Use this checklist before publishing `v0.5.1-alpha`.

## Repository hygiene

- [ ] Repository name and description are set.
- [ ] Default branch is `main`.
- [ ] `.env`, local databases, backups and logs are not tracked.
- [ ] No credentials, tokens, real customer data or private URLs exist in history.
- [ ] README renders correctly on GitHub.
- [ ] LICENSE, SECURITY, CONTRIBUTING and CODE_OF_CONDUCT are present.

## Validation

- [ ] Fresh SQLite onboarding tested.
- [ ] Upgrade from an existing v0.5 SQLite DB tested.
- [ ] PostgreSQL fresh migration tested.
- [ ] Customer → quotation → invoice → payment flow tested.
- [ ] Expense and financial-report flow tested.
- [ ] PDF quotation, invoice and receipt tested.
- [ ] `go test ./...` passes.
- [ ] `npm run build` passes.
- [ ] GitHub Actions are green.

## Release

- [ ] Add real application screenshots to README.
- [ ] Confirm release is marked **pre-release / alpha**.
- [ ] Tag: `v0.5.1-alpha`.
- [ ] Attach source archive if desired.
- [ ] Publish release notes from `docs/RELEASE_NOTES_v0.5.1-alpha.md`.
