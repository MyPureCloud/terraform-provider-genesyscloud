# Code Templates — Resource Creation

Verbatim shapes from the verified `genesyscloud_recording_settings` (singleton) implementation. Copy the shape, swap the names. Prefer reading the real files under `genesyscloud/recording_settings/` for the most current form.

## Schema layer — `resource_genesyscloud_<name>_schema.go`

Annotation header (right after `package`), `ResourceType` const, `SetRegistrar`, and the resource/data-source/exporter builders.

```go
package recording_settings

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
- **Writable** => `Optional: true, Computed: true`.
- **Read-only** => `Computed: true` only.
- **Constraints** => `ValidateFunc: validation.IntBetween(...)` / `validation.StringInSlice(...)` (SDK) or `validators.ValidateIntMin(...)` (repo helpers).
- Raw-JSON field => `DiffSuppressFunc: util.SuppressEquivalentJsonDiffs` and add it to the exporter's `JsonEncodeAttributes`.

## Proxy layer — `genesyscloud_<name>_proxy.go` (the ONLY SDK boundary)

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

When a proxy needs an endpoint without a generated SDK method, use the `custom_api_client` package (`Do[T]`, `DoNoResponse`, `DoRaw`, `DoWithAcceptHeader`) instead of raw HTTP — see the README "Custom API Client" section.

## Resource CRUD + getAll — `resource_genesyscloud_<name>.go`

Read the real file `genesyscloud/recording_settings/resource_genesyscloud_recording_settings.go` for exact error-wrapping/retry code. Shapes:

- `getAll<Name>(ctx, clientConfig)` -> `resourceExporter.ResourceIDMetaMap`; on 404 return an empty map (don't export); otherwise one entry keyed by the fixed ID with `BlockLabel: ResourceType`.
- `create<Name>` (singleton) -> `d.SetId(<fixedId>)` then `return update<Name>(...)`.
- `read<Name>` -> build `consistency_checker.NewConsistencyCheck(...)`, wrap in `util.WithRetriesForRead`, 404 => `retry.RetryableError`, other => `retry.NonRetryableError` (both via `util.BuildWithRetriesApiDiagnosticError`), call `set<Name>ToResourceData(d, settings)`, return `cc.CheckState(d)`.
- `update<Name>` -> `settings := get<Name>FromResourceData(d)`, PUT via proxy, error via `util.BuildAPIDiagnosticError`, then `return read<Name>(...)`.
- `delete<Name>` (singleton) -> `return nil`.

## Mapping helpers — `resource_genesyscloud_<name>_utils.go` (two functions)

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

For `TypeList`/`TypeSet` blocks, add `flatten<Thing>` / `build<Thing>` helpers and use `resourcedata.SetNillableValueWithInterfaceArrayWithFunc`.

## Data source — `data_source_genesyscloud_<name>.go`

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

## init_test — `genesyscloud_<name>_init_test.go`

Package-level `sdkConfig`, `providerResources`, `providerDataSources`; a `registerTestInstance` with two mutexes; `registerTestResources()` / `registerTestDataSources()` populating the maps with `ResourceType`; `initTestResources()` using `provider.SdkConfigurationForTests()`; and `TestMain(m)` that calls `initTestResources()` then `m.Run()`. Register your own resource plus any dependency resources your acceptance tests reference. Copy verbatim from `recording_media_retention_policy`'s init_test and swap names.

## Unit test — `resource_genesyscloud_<name>_unit_test.go`

Stub the proxy func pointers, set `internalProxy = proxy`, and always `defer func(){ internalProxy = nil }()` so tests don't leak state. Cover **create, read, update, delete, and data source read**. The read-only-field-`nil` assertion in create/update is the single most valuable unit assertion for this kind of resource — always include it.

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

## Acceptance tests — `resource_..._test.go` + `data_source_..._test.go`

- `resource.Test` with `PreCheck: func(){ util.TestAccPreCheck(t) }`, `ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources)`.
- Steps: create step (check attrs incl. `TestCheckResourceAttrSet` for the read-only field), update step (change a couple values, check they applied), and a final import step `ImportState: true, ImportStateVerify: true`.
- Singleton: **no `CheckDestroy`**.
- Data source test: apply the resource + a `data` block with `depends_on`, then `TestCheckResourceAttrPair` to prove the data source exposes the same values as the resource.
- Use `util.TrueValue` / `util.FalseValue` for bool HCL literals.

## Conditional / less-common files

- **`resource_genesyscloud_<name>_schema_upgrade.go`** — only if you change the schema shape after release and need a `StateUpgrader` (add `StateUpgraders` + bump `SchemaVersion`). See `recording_media_retention_policy`. Not needed for a brand-new resource.

## Examples + apis.md — `examples/resources/genesyscloud_<name>/`

`resource.tf` — a working HCL block with realistic values (writable fields only; never include read-only/computed fields).

`apis.md`:
```markdown
<!-- sources
genesyscloud/<name>/genesyscloud_<name>_proxy.go
-->
* [GET /api/v2/...](https://developer.genesys.cloud/devapps/api-explorer#get-api-v2-...)
* [PUT /api/v2/...](https://developer.genesys.cloud/devapps/api-explorer#put-api-v2-...)
```

If a data source is registered: `examples/data-sources/genesyscloud_<name>/data-source.tf` (empty block for a singleton) **and its own `apis.md`** — the data-source example folder needs its own `apis.md` or `go generate` errors with "Missing APIs file".

## Registrar wiring — `genesyscloud/provider_registrar/provider_registrar.go`

Two edits — both required:
1. Import alias in the import block:
   ```go
   recordingSettings "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/recording_settings"
   ```
2. A call inside `registerResources()`:
   ```go
   recordingSettings.SetRegistrar(regInstance) //Registering recording settings
   ```
After editing, grep-verify BOTH landed (the import alias is easy to lose):
```bash
grep -n "recording_settings" genesyscloud/provider_registrar/provider_registrar.go
```
No `main.go` change is needed; it aggregates via `registerResources()`.
