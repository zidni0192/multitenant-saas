# Architecture

## Projects storage

`PROJECTS_DB_PATH` selects the JSON storage file; when unset, it defaults to `data/projects.json` relative to the service working directory. Parent directories are created with owner-only permissions. Each write serializes tenant-keyed projects and the next project ID to a temporary file in the same directory, syncs and closes it, then renames it into place. The store serializes access within one process and rolls back an in-memory create if persistence fails.

The JSON file is intended for local development and a single service process. It does not coordinate multiple processes, provide database transactions or migrations, or replace a production database and backup strategy. Tenant IDs are caller-provided isolation keys; they are not authenticated or authorized by this service.

## Request flow

The `/projects` handler rejects blank `X-Tenant-ID` values, then performs create/list operations against the tenant-keyed store. Request validation errors return `400`; storage failures return a generic `500` without exposing filesystem paths or OS error details.
