# Gotchas & Known Pitfalls

Institutional knowledge that has caused real bugs or wasted time. Check these before assuming.

## SDK

- **SDK version mismatch.** A field shown in the web API docs may not exist in the pinned SDK major (from `go.mod`). Confirm against the pinned version; if a needed field is missing, flag an SDK bump rather than silently dropping it. Never invent SDK method names, model field names, or types — read them from the SDK source.

## Schema / mapping

- **Read-only fields must be `Computed`-only** and must NOT be included in the create/update request body. If a read-only/computed field is emitted as settable config (e.g. by the exporter), re-apply breaks. Assert the field is `nil` in the PUT body in a unit test.
- **`resourcedata.GetNillableValue` uses `GetOk` semantics** — a literal `false` or `0` reads as "unset" and returns nil. Fine for `Computed` fields with validated ranges; call it out explicitly if the resource genuinely needs to force an explicit `false`/`0`.

## Singleton / org-settings resources (GET+PUT only, one per org)

- **Create** = `d.SetId("<fixed_constant>")` then delegate to update (no POST).
- **Delete** is a **no-op** (`return nil`); the object always exists server-side.
- Acceptance tests have **no destructive `CheckDestroy`**.
- The **fixed singleton ID string is permanent** — changing it after release orphans users' state.
- `createDefault=true` on the singleton GET auto-materializes defaults for a fresh org — confirm that behavior is acceptable before relying on it.

## Registration

- The **import alias in `provider_registrar.go` silently goes missing** — always grep-verify BOTH the import alias and the `SetRegistrar` call landed.

## Docs / examples

- **A data-source example folder needs its own `apis.md`** or `make docs` / `go generate` fails with "Missing APIs file".
- **Never hand-edit `docs/`** — it is regenerated.

## Local testing

- **Stale sideload** — after every code change, `make sideload` and confirm the binary contains the resource (`strings <mirror-binary> | grep genesyscloud_<name>`). Running an old binary against a config wastes time chasing phantom bugs.

## Exporter (tfexporter)

- The exporter runs highly concurrent. Shared state (resource maps, slices) must go through the mutex-guarded accessors (`GetSanitizedResourceMap` / `SetSanitizedResourceMap` / `RemoveFromSanitizedResourceMap`, etc.). Do NOT mutate shared maps/slices directly inside goroutines — that caused intermittent export failures (see `EXPORTER_FIXES_SUMMARY.md`).
- Concurrency is bounded by `max_concurrent_operations`. Timeout errors are retried with exponential backoff; non-timeout errors are not.

## Pagination

- When migrating `getAll*` to `provider.FetchPagesConcurrently`, create a **new proxy per page** (`newXxxProxy(clientConfig)`), pass the **same** filters/expands/args as page 1, return an empty slice (not nil) for empty pages, and keep cache population AFTER the full fetch (not inside the callback). Nil-check `PageCount` (default to 1). See `CONCURRENT_PAGINATION.md`.
- `token_pool_size` higher than `max_concurrent_pages` adds OAuth prefill cost without speeding pagination — tune for total apply time, and set the pool to match concurrency, not higher.
