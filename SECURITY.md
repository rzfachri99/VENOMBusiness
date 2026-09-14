# Security Policy

VENOMBusiness handles authentication and business/financial data, so security reports are taken seriously.

## Supported versions

During early alpha, security fixes are applied to the latest release line only.

| Version | Security fixes |
| --- | --- |
| 0.5.x | Yes |
| < 0.5 | No |

## Reporting a vulnerability

Please **do not open a public GitHub issue** for an undisclosed vulnerability.

Until a dedicated private vulnerability-reporting channel is published for the repository, contact the repository owner privately through an available GitHub contact method and include:

- affected component/version
- impact
- reproduction steps or proof of concept
- suggested remediation, if known

Do not include real customer data, credentials, session tokens, production database dumps, or other third-party secrets in a report.

## Security expectations for deployments

- Use HTTPS in production.
- Use secure cookies in production.
- Keep secrets outside the repository.
- Use PostgreSQL TLS for remote production databases.
- Maintain tested backups.
- Restrict database and host access using least privilege.
- Keep Go, Node.js and dependencies patched.

VENOMBusiness is early alpha and should not be the only copy of critical financial records.
