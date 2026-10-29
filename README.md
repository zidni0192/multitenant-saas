# multitenant-saas

A local-first multi-tenant SaaS starter. The initial API exposes a health endpoint; tenant isolation, authorization, audit logging, and the React client are planned as scoped follow-up work.

## Local setup

Requires Go 1.27.1 or newer.

```sh
go test ./...
go run .
curl http://localhost:8080/health
```

## Architecture

The initial slice is a standard-library Go HTTP service with a health route. Future issue-driven slices will establish tenant-scoped persistence, RBAC, audit events, and a React UI. No database, Docker, identity provider, or production credentials are required to run tests.

## License

MIT
