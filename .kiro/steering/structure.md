# Structure & Architecture

## Layering (respect this in all provider code)

```
Terraform Core
  -> SCHEMA layer   (_schema.go)      type name, schema, registration
  -> RESOURCE layer (resource_*.go,   CRUD callbacks, retries, consistency check
                     data_source_*.go, build/flatten mapping in _utils.go)
  -> PROXY layer    (_proxy.go)       the ONLY code that imports/calls platformclientv2
  -> platformclientv2 SDK -> Genesys Cloud REST API
```

**Core invariant:** the resource and data-source layers must never call the `platformclientv2` SDK directly — only through the proxy. The proxy holds swappable function pointers (one func type per operation) so unit tests can replace the API with stubs.

## Repository layout

- `genesyscloud/` — one sub-package per resource type. The package name matches the resource type minus the `genesyscloud_` prefix (e.g. `genesyscloud_routing_queue` lives in `genesyscloud/routing_queue/`).
- `genesyscloud/provider/` — provider setup, pooled SDK client, `EnsureResourceContext`, concurrent pagination.
- `genesyscloud/provider_registrar/` — aggregates all resource registrations.
- `genesyscloud/util/` — shared helpers (`resourcedata`, `errors`, retries, diagnostics, JSON, e164, etc.).
- `genesyscloud/validators/` — reusable schema validators.
- `genesyscloud/consistency_checker/` — eventual-consistency retry helper used in reads.
- `genesyscloud/resource_exporter/`, `genesyscloud/tfexporter/` — export framework and the `genesyscloud_tf_export` resource.
- `genesyscloud/custom_api_client/` — typed HTTP client for endpoints with no generated SDK method.
- `examples/` — per-resource `resource.tf` + `apis.md` used to generate docs.
- `docs/` — **generated**; never hand-edit.
- `main.go` — provider entry point; aggregates via the registrar.

## Per-resource file convention (package `genesyscloud/<name>/`)

- `resource_genesyscloud_<name>_schema.go` — `ResourceType` const, `SetRegistrar`, resource/data-source/exporter builders.
- `genesyscloud_<name>_proxy.go` — the SDK boundary (private).
- `resource_genesyscloud_<name>.go` — CRUD callbacks + `getAll` for the exporter.
- `resource_genesyscloud_<name>_utils.go` — build (schema->SDK) / flatten (SDK->schema) helpers.
- `data_source_genesyscloud_<name>.go` — data source read via the same proxy.
- `resource_..._test.go` / `data_source_..._test.go` — acceptance tests.
- `genesyscloud_<name>_unit_test.go` — unit tests (stub the proxy func pointers).
- `genesyscloud_<name>_init_test.go` — test registration.

## Registration

A new resource is registered in `genesyscloud/provider_registrar/provider_registrar.go` (import alias + a `SetRegistrar(regInstance)` call). No `main.go` change is needed.
