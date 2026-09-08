---
name: resource-creation
description: Create a new Terraform resource (and its data source, exporter, unit tests, acceptance tests, examples, and docs) in terraform-provider-genesyscloud, end-to-end, following the codebase's proxy/schema/resource layering. Trigger this whenever the user asks to "create a resource", "add a new resource", "wrap an API endpoint as a resource", "build a data source", or gives a Genesys Cloud API endpoint (e.g. /api/v2/...) and asks to expose it through the provider.
---

# Resource Creation — Genesys Cloud Terraform Provider

Turns a Genesys Cloud API endpoint into a complete, tested, registered Terraform resource that follows this repo's established layering and conventions. Every code block below is drawn from the real, verified `genesyscloud_recording_settings` implementation — copy the shapes, swap the names.

## Persona

You are a senior Go engineer building HashiCorp Terraform providers, with deep knowledge of the Terraform Plugin SDK v2, the Genesys Cloud `platformclientv2` SDK, and this repository's patterns. Explain what you're doing in the context of the task; don't lecture on Go/Terraform basics unless asked.

## The layering (understand this before writing anything)

```
Terraform Core
  → SCHEMA layer   (_schema.go)      declares type name, schema, registration
  → RESOURCE layer (resource_*.go,   CRUD callbacks, retries, consistency check,
                    data_source_*.go)  build/flatten mapping (_utils.go)
  → PROXY layer    (_proxy.go)       the ONLY code that imports/calls platformclientv2
  → platformclientv2 SDK → Genesys Cloud REST API
```

Rule that falls out of this: **the resource/data-source layer must never call the SDK directly — only through the proxy.** The proxy holds swappable function pointers so unit tests can replace the API with stubs.

## Hard Rules (non-negotiable)

- **Read before you write.** Never invent SDK method names, model field names, or helper signatures. Confirm every SDK call and model field against the pinned SDK version first (Phase 1). Guessing here is the #1 cause of broken builds.
- **Proxy is the only SDK boundary.** Resource/data-source code calls proxy methods, never `platformclientv2` directly.
- **Never edit `docs/`.** It is generated. Change `examples/` + schema, then run `go generate`.
- **Deliver the full standard set:** resource + data source + exporter + unit tests + acceptance tests + examples + docs + registration. Don't stop at "it compiles."
- **Verify before claiming done.** Run gofmt, `go build ./...`, `go vet`, unit tests, `go generate`. Acceptance tests need a live org and are the user's manual step — never claim they pass unless the user ran them.
- **Match the pinned SDK major version** (from `go.mod`) in every import. A field in a newer SDK may not exist in the pinned one.
- **Confirm design decisions with the user** for anything non-obvious: the fixed singleton ID string, no-op delete, `createDefault=true` on GET, an SDK bump. These are effectively permanent once released.

---

## Phase protocol (STOP-and-confirm between every phase)

This skill runs as **gated phases, not one continuous flow**. After finishing each phase you MUST:

1. **Stop and summarize** what that phase produced (endpoint resolved, files created, tests run, etc.).
2. **Ask the user to confirm** before starting the next phase. One line is enough — e.g. "Phase 1 done: confirmed SDK v195 + these fields. Proceed to Phase 2 (decide resource shape)?"
3. **On confirmation, RE-READ this SKILL.md file** (`.kiro/skills/resource-creation/SKILL.md`) before doing the next phase's work, then continue **in order** (Phase 0 → 1 → 2 → 3 → 4 → 5 → 6). Do not skip ahead or reorder.

**Why the re-read matters (and how Kiro actually behaves):** the skill is loaded into context once, at activation — it is NOT auto-re-injected after each confirmation. Over a long session, earlier instructions can be compacted out and phase-order adherence can drift. Explicitly re-reading the file at each phase boundary reloads the exact steps so the procedure stays faithful from Phase 0 through Phase 6. If a phase boundary is critical, the user may also back this with a Kiro hook (see "Enforcing the phase protocol" at the end) so it survives context compaction regardless of model state.

Do not batch multiple phases into a single uninterrupted run unless the user explicitly says "do it all end-to-end without stopping."

---

## Phase 0 — Discover the endpoint when only a name is given

Tickets often give just a name ("create a resource for Quality Recording Management") with no `/api/v2/...` path. Do NOT guess and start coding. Resolve the endpoint first:

1. **Infer candidate paths from the name and check the SDK module cache** (fastest authoritative source):
   ```bash
   SDK="$(go env GOPATH)/pkg/mod/github.com/mypurecloud/platform-client-sdk-go/vNNN@vNNN.0.0/platformclientv2"
   ls "$SDK" | grep -i "<keyword>"                                       # models: recording*, quality*, ...
   grep -rn 'Configuration.BasePath + "/api/v2/<guess>' "$SDK"/*api.go   # confirm real paths + verbs
   ```
2. **Search the Genesys Cloud API Explorer / developer docs** to confirm the endpoint, its verbs, request/response schema, and required OAuth scopes. Use the web tools against `developer.genesys.cloud` (API Explorer at `developer.genesys.cloud/devapps/api-explorer`). Cross-check against the SDK from step 1 — the pinned SDK is the source of truth for what you can actually call; web docs may describe a newer surface.
3. **If still ambiguous** (several plausible endpoints, or the name maps to a whole UI area spanning multiple APIs), **ask the user** which endpoint(s) to wrap. List the candidates you found. Do not silently pick one.
4. **Confirm the resolved endpoint(s) with the user before coding** when you inferred them: one line — "I'll wrap `GET/PUT /api/v2/...`, sound right?"

---

## Phase 1 — Understand the API and confirm the SDK (no code yet)

1. **Pin the SDK version:** `grep platform-client-sdk-go go.mod` → e.g. `v195`. Use that exact major in every import.
2. **Read the model and API methods from the module cache** (SDK is not vendored):
   ```bash
   SDK="$(go env GOPATH)/pkg/mod/github.com/mypurecloud/platform-client-sdk-go/vNNN@vNNN.0.0/platformclientv2"
   cat "$SDK/<model>.go"                            # field names, pointer types, read-only comments
   grep -n "func (a XxxApi) .*<Verb>" "$SDK/xxxapi.go"   # exact method names + signatures + query params
   ```
   Note pointer types (`*int`, `*bool`) — they map to `resourcedata.SetNillableValue` / `GetNillableValue`. Note any `createDefault`-style query params.
3. **Map API fields → schema attributes** in a table. For each field record: `snake_case` attribute name, type, **writable vs read-only** (read-only ⇒ `Computed` only, and MUST NOT be sent in the PUT body), and documented constraints (ranges/enums → validators).
4. **Check for scope creep.** A UI page often aggregates several endpoints (e.g. "Recording Management" = `/recording/settings` + `/orphanrecordings` + `/recording/mediaretentionpolicies`). Only model fields that belong to *your* endpoint's schema. Fields backed by other endpoints belong to other resources (or are runtime data that belongs in no resource).

---

## Phase 2 — Decide the resource shape

Choose based on which verbs the endpoint exposes:

### A) Standard collection resource (POST/GET/PUT/DELETE, per-item IDs)
Normal CRUD, server-assigned ID from the create response. Delete calls the real DELETE. Exporter enumerates the collection in `getAll`.

### B) Singleton / org-settings resource (only GET + PUT, one object per org)
Use the singleton pattern (this is what `recording_settings` and `organization_authentication_settings` are):
- **Create** = `d.SetId("<fixed_constant>")` then delegate to update. There is no POST.
- **Read** = GET the singleton.
- **Update** = PUT, then call read.
- **Delete** = **no-op**, `return nil`. Nothing to delete server-side; Terraform drops it from state.
- **Exporter** = `IsSingleton: true`, `ExportId: ResourceType`, `getAll` returns exactly one entry (empty on 404).
- **Acceptance test** = NO destructive `CheckDestroy` (the object always exists). Note it mutates a real org-wide setting.
- **Reference template** = `genesyscloud/organization_authentication_settings/`.

---

## Phase 3 — Create the package

Create `genesyscloud/<name>/` where `<name>` = resource type minus the `genesyscloud_` prefix (e.g. `recording_settings`). Templates: `organization_authentication_settings` (singleton), `recording_media_retention_policy` (RecordingApi + data source + init_test), `telephony_providers_edges_trunkbasesettings` (settings-style schema/exporter).

Files (create in this order so each compiles against the last):

### 3.1 `resource_genesyscloud_<name>_schema.go`
Annotation header (right after `package`), `ResourceType` const, `SetRegistrar`, and the resource/data-source/exporter builders.

```go
package recording_settings

// @team: Recording and Policies
// @pm: <PM Name>
// @jira: <TICKET-KEY>
// @description: <one line describing what this manages>

import (
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	registrar "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_register"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

const ResourceType = "genesyscloud_recording_settings"

func SetRegistrar(l registrar.Registrar) {
	l.RegisterResource(ResourceType, ResourceRecordingSettings())
	l.RegisterDataSource(ResourceType, DataSourceRecordingSettings())
	l.RegisterExporter(ResourceType, RecordingSettingsExporter())
}

func ResourceRecordingSettings() *schema.Resource {
	return &schema.Resource{
		Description:   "Genesys Cloud Organization Recording Settings. This is a singleton resource; only one instance exists per organization.",
		CreateContext: provider.CreateWithPooledClient(createRecordingSettings),
		ReadContext:   provider.ReadWithPooledClient(readRecordingSettings),
		UpdateContext: provider.UpdateWithPooledClient(updateRecordingSettings),
		DeleteContext: provider.DeleteWithPooledClient(deleteRecordingSettings),
		Importer:      &schema.ResourceImporter{StateContext: schema.ImportStatePassthroughContext},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			// writable field: Optional + Computed (server always has a value)
			"max_simultaneous_streams": {
				Type: schema.TypeInt, Optional: true, Computed: true,
				Description: "Maximum number of simultaneous screen recording streams.",
			},
			// read-only field: Computed ONLY (user can never set it)
			"max_configurable_screen_recording_streams": {
				Type: schema.TypeInt, Computed: true,
				Description: "Upper limit that max_simultaneous_streams can be configured to. Read-only.",
			},
			// writable + validated range
			"recording_playback_url_ttl": {
				Type: schema.TypeInt, Optional: true, Computed: true,
				ValidateFunc: validation.IntBetween(2, 60),
				Description:  "Playback URL validity in minutes (2-60, default 60).",
			},
			// ... remaining fields ...
		},
	}
}

func DataSourceRecordingSettings() *schema.Resource {
	return &schema.Resource{
		Description: "Data source for Genesys Cloud Organization Recording Settings. Singleton, so no arguments are required.",
		ReadContext: provider.ReadWithPooledClient(dataSourceRecordingSettingsRead),
		// For a singleton: NO input args. Expose every field as Computed so the data source is useful.
		Schema: map[string]*schema.Schema{
			"max_simultaneous_streams":                  {Type: schema.TypeInt, Computed: true},
			"max_configurable_screen_recording_streams": {Type: schema.TypeInt, Computed: true},
			// ... all remaining fields, all Computed ...
		},
	}
}

func RecordingSettingsExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: provider.GetAllWithPooledClient(getAllRecordingSettings),
		IsSingleton:      true, // singleton only
		ExportId:         ResourceType,
		RefAttrs:         map[string]*resourceExporter.RefAttrSettings{},
		// If any attribute references another resource, add it to RefAttrs: {"foo_id": {RefType: "genesyscloud_bar"}}
	}
}
```

Schema field rules:
- **Writable** ⇒ `Optional: true, Computed: true`.
- **Read-only** ⇒ `Computed: true` only.
- **Constraints** ⇒ `ValidateFunc: validation.IntBetween(...)` / `validation.StringInSlice(...)` (SDK) or `validators.ValidateIntMin(...)` (repo helpers).
- If a field is raw JSON, use `DiffSuppressFunc: util.SuppressEquivalentJsonDiffs` and add it to the exporter's `JsonEncodeAttributes`.

### 3.2 `genesyscloud_<name>_proxy.go` — the SDK boundary
```go
package recording_settings

import (
	"context"
	"fmt"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/platform-client-sdk-go/v195/platformclientv2"
)

var internalProxy *recordingSettingsProxy

// one func type per operation -> enables stubbing in unit tests
type getRecordingSettingsFunc func(ctx context.Context, p *recordingSettingsProxy) (*platformclientv2.Recordingsettings, *platformclientv2.APIResponse, error)
type updateRecordingSettingsFunc func(ctx context.Context, p *recordingSettingsProxy, settings *platformclientv2.Recordingsettings) (*platformclientv2.Recordingsettings, *platformclientv2.APIResponse, error)

type recordingSettingsProxy struct {
	clientConfig                *platformclientv2.Configuration
	recordingApi                *platformclientv2.RecordingApi
	getRecordingSettingsAttr    getRecordingSettingsFunc
	updateRecordingSettingsAttr updateRecordingSettingsFunc
}

func newRecordingSettingsProxy(clientConfig *platformclientv2.Configuration) *recordingSettingsProxy {
	api := platformclientv2.NewRecordingApiWithConfig(clientConfig)
	return &recordingSettingsProxy{
		clientConfig:                clientConfig,
		recordingApi:                api,
		getRecordingSettingsAttr:    getRecordingSettingsFn,
		updateRecordingSettingsAttr: updateRecordingSettingsFn,
	}
}

func getRecordingSettingsProxy(clientConfig *platformclientv2.Configuration) *recordingSettingsProxy {
	if internalProxy == nil {
		internalProxy = newRecordingSettingsProxy(clientConfig)
	}
	return internalProxy
}

// wrapper methods delegate to the func pointers
func (p *recordingSettingsProxy) getRecordingSettings(ctx context.Context) (*platformclientv2.Recordingsettings, *platformclientv2.APIResponse, error) {
	return p.getRecordingSettingsAttr(ctx, p)
}
func (p *recordingSettingsProxy) updateRecordingSettings(ctx context.Context, settings *platformclientv2.Recordingsettings) (*platformclientv2.Recordingsettings, *platformclientv2.APIResponse, error) {
	return p.updateRecordingSettingsAttr(ctx, p, settings)
}

// concrete implementations make the real SDK calls
func getRecordingSettingsFn(ctx context.Context, p *recordingSettingsProxy) (*platformclientv2.Recordingsettings, *platformclientv2.APIResponse, error) {
	ctx = provider.EnsureResourceContext(ctx, ResourceType) // SDK debug logging attribution
	settings, resp, err := p.recordingApi.GetRecordingSettings(true) // createDefault=true: materialize defaults if none exist
	if err != nil {
		return nil, resp, fmt.Errorf("failed to retrieve recording settings: %s", err)
	}
	return settings, resp, nil
}
func updateRecordingSettingsFn(ctx context.Context, p *recordingSettingsProxy, settings *platformclientv2.Recordingsettings) (*platformclientv2.Recordingsettings, *platformclientv2.APIResponse, error) {
	ctx = provider.EnsureResourceContext(ctx, ResourceType)
	updated, resp, err := p.recordingApi.PutRecordingSettings(*settings)
	if err != nil {
		return nil, resp, fmt.Errorf("failed to update recording settings: %s", err)
	}
	return updated, resp, nil
}
```

### 3.3 `resource_genesyscloud_<name>.go` — CRUD + getAll
Use the exact shapes from the real file:
- `getAll<Name>(ctx, clientConfig)` → `resourceExporter.ResourceIDMetaMap`; on 404 return an empty map (don't export); otherwise one entry keyed by the fixed ID with `BlockLabel: ResourceType`.
- `create<Name>` (singleton) → `d.SetId(<fixedId>)` then `return update<Name>(...)`.
- `read<Name>` → build `consistency_checker.NewConsistencyCheck(...)`, wrap in `util.WithRetriesForRead`, 404 ⇒ `retry.RetryableError`, other ⇒ `retry.NonRetryableError` (both via `util.BuildWithRetriesApiDiagnosticError`), call `set<Name>ToResourceData(d, settings)`, return `cc.CheckState(d)`.
- `update<Name>` → `settings := get<Name>FromResourceData(d)`, PUT via proxy, error via `util.BuildAPIDiagnosticError`, then `return read<Name>(...)`.
- `delete<Name>` (singleton) → `return nil`.

Reference the working file `genesyscloud/recording_settings/resource_genesyscloud_recording_settings.go` for the precise error-wrapping and retry code.

### 3.4 `resource_genesyscloud_<name>_utils.go` — mapping helpers (two functions)
```go
// schema -> SDK model. OMIT read-only fields from the update body.
func getRecordingSettingsFromResourceData(d *schema.ResourceData) platformclientv2.Recordingsettings {
	return platformclientv2.Recordingsettings{
		MaxSimultaneousStreams:          resourcedata.GetNillableValue[int](d, "max_simultaneous_streams"),
		RegionalRecordingStorageEnabled: resourcedata.GetNillableValue[bool](d, "regional_recording_storage_enabled"),
		// NOTE: MaxConfigurableScreenRecordingStreams (read-only) is intentionally NOT set here.
		// ... other writable fields ...
	}
}

// SDK model -> schema. Include ALL fields (incl. read-only). Shared by resource read AND data source read.
func setRecordingSettingsToResourceData(d *schema.ResourceData, settings *platformclientv2.Recordingsettings) {
	resourcedata.SetNillableValue(d, "max_simultaneous_streams", settings.MaxSimultaneousStreams)
	resourcedata.SetNillableValue(d, "max_configurable_screen_recording_streams", settings.MaxConfigurableScreenRecordingStreams)
	// ... every field ...
}
```

### 3.5 `data_source_genesyscloud_<name>.go`
Read via the same proxy, set the ID, flatten with the shared helper. No direct SDK calls.
```go
func dataSourceRecordingSettingsRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	sdkConfig := m.(*provider.ProviderMeta).ClientConfig
	proxy := getRecordingSettingsProxy(sdkConfig)
	return util.WithRetries(ctx, 15*time.Second, func() *retry.RetryError {
		settings, resp, err := proxy.getRecordingSettings(ctx)
		if err != nil {
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("error requesting recording settings | error: %s", err), resp))
		}
		d.SetId(recordingSettingsId)                       // singleton: fixed ID
		setRecordingSettingsToResourceData(d, settings)    // expose values
		return nil
	})
}
```

### 3.6 `genesyscloud_<name>_init_test.go`
Package-level `sdkConfig`, `providerResources`, `providerDataSources`; a `registerTestInstance` with two mutexes; `registerTestResources()` / `registerTestDataSources()` populating the maps with `ResourceType`; `initTestResources()` using `provider.SdkConfigurationForTests()`; and `TestMain(m)` that calls `initTestResources()` then `m.Run()`. Register your own resource plus any dependency resources your acceptance tests reference. Copy verbatim from `recording_media_retention_policy`'s init_test and swap names.

### 3.7 `resource_genesyscloud_<name>_unit_test.go`
Stub the proxy func pointers, set `internalProxy = proxy`, and always `defer func(){ internalProxy = nil }()` so tests don't leak state. Cover **create, read, update, delete, and data source read**.
```go
func TestUnitResourceRecordingSettingsUpdate(t *testing.T) {
	testSettings := generateRecordingSettingsData()
	var capturedPutBody *platformclientv2.Recordingsettings
	proxy := &recordingSettingsProxy{}
	proxy.getRecordingSettingsAttr = func(ctx context.Context, p *recordingSettingsProxy) (*platformclientv2.Recordingsettings, *platformclientv2.APIResponse, error) {
		s := testSettings; return &s, &platformclientv2.APIResponse{StatusCode: http.StatusOK}, nil
	}
	proxy.updateRecordingSettingsAttr = func(ctx context.Context, p *recordingSettingsProxy, s *platformclientv2.Recordingsettings) (*platformclientv2.Recordingsettings, *platformclientv2.APIResponse, error) {
		capturedPutBody = s
		return s, &platformclientv2.APIResponse{StatusCode: http.StatusOK}, nil
	}
	internalProxy = proxy
	defer func() { internalProxy = nil }()

	d := schema.TestResourceDataRaw(t, ResourceRecordingSettings().Schema, buildRecordingSettingsDataMap(testSettings))
	d.SetId(recordingSettingsId)
	gcloud := &provider.ProviderMeta{ClientConfig: &platformclientv2.Configuration{}}

	diag := updateRecordingSettings(context.Background(), d, gcloud)
	assert.Equal(t, false, diag.HasError())
	assert.Nil(t, capturedPutBody.MaxConfigurableScreenRecordingStreams) // read-only field MUST NOT be sent
}
```
The read-only-field-`nil` assertion in create/update is the single most valuable unit assertion for this kind of resource — always include it.

### 3.8 `resource_genesyscloud_<name>_test.go` + `data_source_genesyscloud_<name>_test.go`
Acceptance tests:
- `resource.Test` with `PreCheck: func(){ util.TestAccPreCheck(t) }`, `ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources)`.
- Steps: create step (check attrs incl. `TestCheckResourceAttrSet` for the read-only field), update step (change a couple values, check they applied), and a final import step `ImportState: true, ImportStateVerify: true`.
- Singleton: **no `CheckDestroy`**.
- Data source test: apply the resource + a `data` block with `depends_on`, then `TestCheckResourceAttrPair` to prove the data source exposes the same values as the resource.
- Use `util.TrueValue` / `util.FalseValue` for bool HCL literals.

### Conditional / less-common files
- **`resource_genesyscloud_<name>_schema_upgrade.go`** — only if you change the schema shape after release and need a `StateUpgrader` (add `StateUpgraders` + bump `SchemaVersion`). See `recording_media_retention_policy`. Not needed for a brand-new resource.
- **Nested-object helpers** in `_utils.go` — for `TypeList`/`TypeSet` blocks, add `flatten<Thing>` / `build<Thing>` helpers and use `resourcedata.SetNillableValueWithInterfaceArrayWithFunc`.

---

## Phase 4 — Examples, wiring, docs

### 4.1 Examples
- `examples/resources/genesyscloud_<name>/resource.tf` — a working HCL block with realistic values (writable fields only; never include read-only/computed fields).
- `examples/resources/genesyscloud_<name>/apis.md`:
  ```markdown
  <!-- sources
  genesyscloud/<name>/genesyscloud_<name>_proxy.go
  -->
  * [GET /api/v2/...](https://developer.genesys.cloud/devapps/api-explorer#get-api-v2-...)
  * [PUT /api/v2/...](https://developer.genesys.cloud/devapps/api-explorer#put-api-v2-...)
  ```
- If a data source is registered: `examples/data-sources/genesyscloud_<name>/data-source.tf` (empty block for a singleton) **and its own `apis.md`** — the data-source example folder needs its own `apis.md` or `go generate` errors with "Missing APIs file".

### 4.2 Register in `genesyscloud/provider_registrar/provider_registrar.go`
Two edits — both required:
1. Import alias in the import block:
   ```go
   recordingSettings "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/recording_settings"
   ```
2. A call inside `registerResources()`:
   ```go
   recordingSettings.SetRegistrar(regInstance) //Registering recording settings
   ```
After editing, **grep-verify BOTH landed** — the import alias is easy to lose:
```bash
grep -n "recording_settings" genesyscloud/provider_registrar/provider_registrar.go
```
No `main.go` change is needed; it aggregates via `registerResources()`.

### 4.3 Docs
```bash
go generate ./...
```
Must exit 0 with no "Missing APIs file" errors. The generator auto-fills the permissions/scopes section from the API Explorer. **Never hand-edit `docs/`.**

---

## Phase 5 — Verify (automated)

```bash
gofmt -w genesyscloud/<name>/ && gofmt -l genesyscloud/<name>/   # empty output = clean
go build ./...
go vet ./genesyscloud/<name>/...
go test ./genesyscloud/<name>/... -run TestUnit -v              # all unit tests green
go generate ./...                                               # exit 0
```
State plainly which checks passed. If the whole-repo unit run is wanted: `make testunit` (`TF_UNIT=1 go test ./... -run TestUnit ...`).

---

## Phase 6 — Manual live verification (hand to the user, or run if they ask)

This repo uses a **filesystem plugin mirror**, NOT `dev_overrides`.

1. **Build + sideload** from the correct branch:
   ```bash
   make sideload   # builds dist/ and copies to ~/.terraform.d/plugins/genesys.com/mypurecloud/genesyscloud/0.1.0/<os>_<arch>/
   ```
2. **Confirm the binary actually contains the resource** (catches a stale sideload — a common wasted-hour mistake):
   ```bash
   strings ~/.terraform.d/plugins/genesys.com/mypurecloud/genesyscloud/0.1.0/$(go env GOOS)_$(go env GOARCH)/terraform-provider-genesyscloud | grep genesyscloud_<name>
   ```
3. **Test config** in a clean dir (source MUST be the mirror path):
   ```hcl
   terraform {
     required_providers {
       genesyscloud = { source = "genesys.com/mypurecloud/genesyscloud", version = "0.1.0" }
     }
   }
   provider "genesyscloud" {}   # creds from GENESYSCLOUD_OAUTHCLIENT_ID / _SECRET / _REGION env vars

   resource "genesyscloud_<name>" "this" { /* writable fields */ }
   data "genesyscloud_<name>" "current" { depends_on = [genesyscloud_<name>.this] }
   ```
4. **Run and check each path** (with the filesystem mirror you DO run `terraform init`):
   - `terraform init && terraform apply` → **CREATE (PUT) + READ (GET)**. Outputs should show writable values round-tripped and any read-only field populated by the API.
   - Change a couple values, `terraform apply` again → **UPDATE**. Plan must show `~ update in-place` (not destroy/recreate).
   - `terraform plan` right after apply → **drift check**. Must say **"No changes."** A perpetual diff means a field isn't round-tripping — fix it.
   - Set a field out of range (e.g. a TTL to 90), `terraform plan` → **validation** must fail at plan time, before any API call.
   - `terraform import genesyscloud_<name>.this <fixedId>` (singleton) → should succeed; a follow-up plan shows no diff.
5. **Test export** — add a `genesyscloud_tf_export` block and apply:
   ```hcl
   resource "genesyscloud_tf_export" "export" {
     directory                = "./exported"
     include_state_file       = true
     export_format            = "hcl"
     include_filter_resources = ["genesyscloud_<name>"]
   }
   ```
   Then verify the exported `.tf`:
   - Exactly one resource block for a singleton (not a broken enumerated collection).
   - All **writable** fields present with the org's real values.
   - **Read-only/computed fields are ABSENT** (if a computed field is emitted as settable config, re-apply breaks — that's a real exporter bug to fix).
   - Round-trip it: drop the exported `.tf` in a clean dir and `terraform plan` — no errors about unknown/invalid attributes.

---

## Common pitfalls (learned the hard way)

- **SDK version mismatch.** A field in the web API docs may not exist in the pinned SDK (e.g. a field present in v195 but not v193). Confirm against the pinned major; if a needed field is missing, flag an SDK bump to the user rather than silently dropping the field.
- **Read-only fields.** `Computed`-only in schema; never in the PUT/create body; assert `nil` in a unit test. If exported as settable config, the export round-trip breaks.
- **`resourcedata.GetNillableValue` uses `GetOk` semantics** — a literal `false`/`0` is treated as "unset" and returns nil. Fine for `Computed` fields with validated ranges; call it out if the resource genuinely needs to force an explicit `false`/`0`.
- **Singleton delete is a no-op** and its acceptance test can't assert destruction — no destructive `CheckDestroy`.
- **`provider_registrar.go` import alias** silently goes missing — always grep-verify both edits.
- **Data source example folder needs its own `apis.md`** or `go generate` fails.
- **Stale sideload** — always `make sideload` after every code change and `strings | grep` the binary to confirm the new resource is in it. Running the wrong/old binary against an unrelated config wastes time.
- **Fixed singleton ID is permanent** — pick a stable string; changing it after release orphans users' state.
- **`createDefault=true` on the singleton GET** auto-materializes default settings for a fresh org — confirm that behavior is acceptable with the user.

---

## Reference files (read these for exact, current conventions)

- Singleton pattern (create-via-update, no-op delete, IsSingleton exporter): `genesyscloud/organization_authentication_settings/`
- Full working example of THIS skill's output: `genesyscloud/recording_settings/`
- RecordingApi construction + data source + init_test: `genesyscloud/recording_media_retention_policy/`
- Settings-style schema + exporter (JSON attrs, ExportAsDataFunc): `genesyscloud/telephony_providers_edges_trunkbasesettings/`
- Registrar wiring: `genesyscloud/provider_registrar/provider_registrar.go`
- Export test helpers: `genesyscloud/tfexporter/resource_genesyscloud_tf_export_utils_test.go`
- Repo's own short checklist: `.amazonq/rules/new-resource.md`
