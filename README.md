# multitenant-saas

A local-first multi-tenant SaaS starter with an in-memory, tenant-scoped projects API.

## Local setup

Requires Go 1.27.1 or newer.

```sh
go test ./...
go run .
curl http://localhost:8080/health
```

## Projects API

Both routes require a non-blank `X-Tenant-ID` header; otherwise the service returns `400 Bad Request`.

- `POST /projects` with JSON `{"name":"Example"}` creates a project and returns `201 Created`.
- `GET /projects` lists only projects belonging to that tenant.

The project store is in-memory and is reset when the process restarts. This slice demonstrates request-level tenant isolation; it does not provide authentication or persistent storage.

## Architecture

The initial slice is a standard-library Go HTTP service with a health route. Future issue-driven slices will establish tenant-scoped persistence, RBAC, audit events, and a React UI. No database, Docker, identity provider, or production credentials are required to run tests.

## License

MIT
