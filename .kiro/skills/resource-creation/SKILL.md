---
name: resource-creation
description: Create a new Terraform resource (and its data source, exporter, unit tests, acceptance tests, examples, and docs) in terraform-provider-genesyscloud, end-to-end, following the codebase's proxy/schema/resource layering. Trigger this whenever the user asks to "create a resource", "add a new resource", "wrap an API endpoint as a resource", "build a data source", or gives a Genesys Cloud API endpoint (e.g. /api/v2/...) and asks to expose it through the provider.
---

# Resource Creation — Genesys Cloud Terraform Provider

Turn a Genesys Cloud API endpoint into a complete, tested, registered Terraform resource that follows this repo's layering.

You are a senior Go engineer fluent in Terraform Plugin SDK v2, the `platformclientv2` SDK, and this repo's patterns. Explain decisions in context; don't lecture on Go/Terraform basics unless asked.

Verbatim code shapes live in `references/code-templates.md` (drawn from the verified `genesyscloud_recording_settings` implementation). Load it when you reach Phase 3. Verification steps live in `references/verification.md` — load it at Phase 5/6.

## Layering

```
Terraform Core
  -> SCHEMA   (_schema.go)      type name, schema, registration
  -> RESOURCE (resource_*.go,   CRUD, retries, consistency check
               data_source_*.go, build/flatten in _utils.go)
  -> PROXY    (_proxy.go)       the ONLY code that imports platformclientv2
  -> platformclientv2 SDK -> Genesys Cloud REST API
```

The resource/data-source layer must never call the SDK directly — only through the proxy. The proxy holds swappable function pointers so unit tests can stub the API.

## Hard rules (non-negotiable)

- **Read before you write.** Never invent SDK method names, model fields, or helper signatures. Confirm every SDK call/field against the pinned SDK (Phase 1). Guessing is the #1 cause of broken builds.
- **Proxy is the only SDK boundary.** Resource/data-source code calls proxy methods, never `platformclientv2`.
- **Match the pinned SDK major** (from `go.mod`) in every import. A field in a newer SDK may not exist in the pinned one.
- **Never edit `docs/`.** It's generated — change `examples/` + schema, then `go generate`.
- **Deliver the full set:** resource + data source + exporter + unit tests + acceptance tests + examples + docs + registration. Don't stop at "it compiles."
- **Verify before claiming done.** gofmt, `go build ./...`, `go vet`, unit tests, `go generate`. Acceptance tests need a live org — that's the user's manual step; never claim they pass unless the user ran them.
- **Confirm permanent design decisions with the user:** the fixed singleton ID string, no-op delete, `createDefault=true` on GET, an SDK bump. These are effectively permanent once released.

## Phase protocol — STOP and confirm between every phase

Run as gated phases, not one continuous flow. After each phase you MUST:

1. **Summarize** what the phase produced.
2. **Ask the user to confirm** before the next phase (one line, e.g. "Phase 1 done: SDK v195 + these fields confirmed. Proceed to Phase 2?").
3. **On confirmation, RE-READ this SKILL.md** (`.kiro/skills/resource-creation/SKILL.md`), then continue in order (0 -> 1 -> 2 -> 3 -> 4 -> 5 -> 6). Don't skip or reorder.

Why re-read: the skill loads once at activation and is NOT re-injected per turn. Over a long session earlier steps can be compacted out and phase order drifts. Re-reading at each boundary reloads the exact steps. Don't batch phases unless the user says "do it all end-to-end without stopping."

## Phase 0 — Resolve the endpoint (when only a name is given)

Tickets often give a name ("Quality Recording Management") with no `/api/v2/...` path. Don't guess and code. Resolve first:

1. Infer candidate paths and check the SDK module cache (fastest authoritative source):
   ```bash
   SDK="$(go env GOPATH)/pkg/mod/github.com/mypurecloud/platform-client-sdk-go/vNNN@vNNN.0.0/platformclientv2"
   ls "$SDK" | grep -i "<keyword>"
   grep -rn 'Configuration.BasePath + "/api/v2/<guess>' "$SDK"/*api.go
   ```
2. Confirm verbs, request/response schema, and OAuth scopes in the Genesys Cloud API Explorer (`developer.genesys.cloud/devapps/api-explorer`). The pinned SDK is the source of truth for what you can actually call; web docs may describe a newer surface.
3. If ambiguous (multiple plausible endpoints, or a UI area spanning several APIs), **ask the user** which endpoint(s) to wrap — list your candidates, don't silently pick.
4. Confirm the resolved endpoint(s) before coding.

## Phase 1 — Understand the API, confirm the SDK (no code)

1. Pin the SDK: `grep platform-client-sdk-go go.mod` -> e.g. `v195`. Use that exact major everywhere.
2. Read the model + API methods from the module cache (SDK is not vendored):
   ```bash
   SDK="$(go env GOPATH)/pkg/mod/github.com/mypurecloud/platform-client-sdk-go/vNNN@vNNN.0.0/platformclientv2"
   cat "$SDK/<model>.go"                                # fields, pointer types, read-only comments
   grep -n "func (a XxxApi) .*<Verb>" "$SDK/xxxapi.go"  # exact method names, signatures, query params
   ```
   Pointer types (`*int`, `*bool`) map to `resourcedata.SetNillableValue`/`GetNillableValue`. Note `createDefault`-style query params.
3. Map API fields -> schema attributes in a table: `snake_case` name, type, **writable vs read-only** (read-only => `Computed` only, MUST NOT be in the PUT body), and constraints (ranges/enums -> validators).
4. **Check scope creep.** A UI page often aggregates several endpoints. Only model fields that belong to *your* endpoint. Fields backed by other endpoints belong to other resources.

## Phase 2 — Decide the resource shape

- **A) Standard collection** (POST/GET/PUT/DELETE, per-item IDs): normal CRUD, server-assigned ID from create, real DELETE, exporter enumerates the collection in `getAll`.
- **B) Singleton / org-settings** (only GET + PUT, one object per org) — what `recording_settings` and `organization_authentication_settings` are:
  - Create = `d.SetId("<fixed_constant>")` then delegate to update (no POST).
  - Read = GET the singleton. Update = PUT then read.
  - Delete = **no-op**, `return nil`. Terraform drops it from state.
  - Exporter = `IsSingleton: true`, `ExportId: ResourceType`, `getAll` returns one entry (empty on 404).
  - Acceptance test = NO destructive `CheckDestroy`; note it mutates a real org-wide setting.
  - Template: `genesyscloud/organization_authentication_settings/`.

## Phase 3 — Create the package

Create `genesyscloud/<name>/` where `<name>` = resource type minus the `genesyscloud_` prefix. **Open `references/code-templates.md` now** and follow it — create files in this order so each compiles against the last:

1. `resource_genesyscloud_<name>_schema.go` — annotation header, `ResourceType`, `SetRegistrar`, resource/data-source/exporter builders + schema field rules.
2. `genesyscloud_<name>_proxy.go` — the SDK boundary (func-pointer pattern; `custom_api_client` for endpoints with no SDK method).
3. `resource_genesyscloud_<name>.go` — CRUD + `getAll`.
4. `resource_genesyscloud_<name>_utils.go` — `get...FromResourceData` (omit read-only) and `set...ToResourceData` (all fields, shared with data source).
5. `data_source_genesyscloud_<name>.go` — read via the same proxy.
6. `genesyscloud_<name>_init_test.go` — test registration (copy from `recording_media_retention_policy`).
7. `resource_genesyscloud_<name>_unit_test.go` — stub proxy pointers; cover create/read/update/delete + data source read; assert read-only field is `nil` in the PUT body.
8. `resource_..._test.go` + `data_source_..._test.go` — acceptance tests.

## Phase 4 — Examples, wiring, docs

Follow the examples/registrar sections in `references/code-templates.md`:
1. `examples/resources/genesyscloud_<name>/resource.tf` + `apis.md`; if a data source exists, add `examples/data-sources/genesyscloud_<name>/` with its own `apis.md` (or `go generate` fails).
2. Register in `genesyscloud/provider_registrar/provider_registrar.go` — import alias **and** `SetRegistrar` call; grep-verify both landed.
3. `go generate ./...` — must exit 0 with no "Missing APIs file". Never hand-edit `docs/`.

## Phase 5 — Automated verification

Run these yourself; state plainly which passed. Full detail in `references/verification.md`.

```bash
gofmt -w genesyscloud/<name>/ && gofmt -l genesyscloud/<name>/   # empty output = clean
go vet ./genesyscloud/<name>/...
make testunit                                                   # go build + whole-repo unit suite
make docs                                                       # apidocs generator + go generate; exit 0, no "Missing APIs file"
```

## Phase 6 — Live verification (the user's step; run only if asked)

This repo uses a **filesystem plugin mirror**, NOT `dev_overrides`. Open `references/verification.md` for the test config + export HCL. Sequence:

1. `make sideload` — build + copy the binary to `~/.terraform.d/plugins/genesys.com/mypurecloud/genesyscloud/0.1.0/<os>_<arch>/`.
2. `strings <mirror>/terraform-provider-genesyscloud | grep genesyscloud_<name>` — confirm the new resource is actually in the binary (catches a stale sideload).
3. In a clean dir (`source = "genesys.com/mypurecloud/genesyscloud"`, version `0.1.0`), `terraform init && terraform apply` -> CREATE+READ; edit values + apply -> UPDATE must be `~ in-place`; `plan` again -> **"No changes"** (drift check); out-of-range value -> validation fails at plan time; `terraform import ... <fixedId>` -> no diff after.
4. Add a `genesyscloud_tf_export` block and apply -> exported `.tf` has all writable fields, NO read-only/computed fields, and round-trips with a clean `terraform plan`.

## Common pitfalls (learned the hard way)

- **SDK version mismatch** — a field in web docs may not exist in the pinned SDK. Confirm against the pinned major; flag an SDK bump rather than silently dropping a field.
- **Read-only fields** — `Computed`-only; never in PUT/create body; assert `nil` in a unit test. If exported as settable config, the round-trip breaks.
- **`resourcedata.GetNillableValue` uses `GetOk` semantics** — literal `false`/`0` reads as unset (nil). Fine for `Computed` fields with validated ranges; call it out if the resource must force an explicit `false`/`0`.
- **Singleton delete is a no-op** — no destructive `CheckDestroy`.
- **`provider_registrar.go` import alias** silently goes missing — grep-verify both edits.
- **Data source example folder needs its own `apis.md`** or `go generate` fails.
- **Stale sideload** — `make sideload` after every change and `strings | grep` the binary to confirm the resource is in it.
- **Fixed singleton ID is permanent** — changing it after release orphans users' state.
- **`createDefault=true` on singleton GET** auto-materializes defaults for a fresh org — confirm that's acceptable.

## Reference files (repo — read for exact, current conventions)

- Singleton pattern (create-via-update, no-op delete, IsSingleton exporter): `genesyscloud/organization_authentication_settings/`
- Full working example of this skill's output: `genesyscloud/recording_settings/`
- RecordingApi + data source + init_test: `genesyscloud/recording_media_retention_policy/`
- Settings-style schema + exporter (JSON attrs, ExportAsDataFunc): `genesyscloud/telephony_providers_edges_trunkbasesettings/`
- Registrar wiring: `genesyscloud/provider_registrar/provider_registrar.go`
- Export test helpers: `genesyscloud/tfexporter/resource_genesyscloud_tf_export_utils_test.go`
- Repo's short checklist: `.amazonq/rules/new-resource.md`
