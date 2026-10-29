# multitenant-saas

A local-first multi-tenant SaaS starter with a tenant-scoped projects API backed by a local JSON file.

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

Data is stored at `data/projects.json` by default. Set `PROJECTS_DB_PATH` to choose another path. The service creates parent directories as needed and writes a replacement file atomically. Keep the data file on a local filesystem and back it up as needed. This lightweight single-process JSON store is not designed for concurrent multi-process access, high write volume, or production database workloads. Tenant IDs are isolation keys, not authenticated identities; authentication and authorization are out of scope.

## Architecture

The standard-library HTTP handler validates the tenant header before calling a file-backed store. The store persists tenant-keyed project records and a monotonically increasing project ID in JSON. Writes use a temporary file and rename to avoid exposing partially written state. The implementation requires no database, Docker, external identity provider, or production credentials.

See [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md) for storage details and limitations.

## License

MIT
