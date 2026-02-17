# 🚀 SDKv2 to Plugin Framework Migration Template

This template provides a comprehensive, step-by-step guide for migrating Terraform resources from SDKv2 to Plugin Framework, based on the successful `genesyscloud_routing_language` migration.

## 📋 Pre-Migration Checklist

### Resource Assessment
- [ ] **Resource Complexity**: Simple (single field) / Medium (multiple fields) / Complex (nested objects)
- [ ] **Update Support**: Create/Read/Delete only OR full CRUD operations
- [ ] **Dependencies**: List all resources that depend on this resource
- [ ] **Export Support**: Does this resource support terraform export functionality?
- [ ] **Cross-Package Usage**: Are there tests in other packages that use this resource?

### Prerequisites
- [ ] Existing SDKv2 implementation is stable and well-tested
- [ ] Proxy layer exists and is working correctly
- [ ] Understanding of Framework concepts and patterns
- [ ] Access to test environment for validation

## 🏗️ Migration Strategy: Framework-Only Approach

**Recommended Strategy**: Complete replacement of SDKv2 with Framework implementation

### Benefits:
- ✅ **Simplified Architecture**: Single implementation to maintain
- ✅ **No Muxing Complexity**: Direct Framework provider usage
- ✅ **Faster Development**: Focus on one implementation
- ✅ **Clear Migration State**: Resource is either SDKv2 OR Framework, not both

## 📁 File Structure Planning

### Before Migration:
```
genesyscloud/[resource_name]/
├── resource_genesyscloud_[resource_name]_schema.go    # Schema & Registration
├── resource_genesyscloud_[resource_name].go          # SDKv2 Resource CRUD
├── data_source_genesyscloud_[resource_name].go       # SDKv2 Data Source
├── genesyscloud_[resource_name]_proxy.go             # API Proxy Layer
├── resource_genesyscloud_[resource_name]_test.go     # SDKv2 Resource Tests
├── data_source_genesyscloud_[resource_name]_test.go  # SDKv2 Data Source Tests
└── genesyscloud_[resource_name]_init_test.go         # Test Initialization
```

### After Migration:
```
genesyscloud/[resource_name]/
├── resource_genesyscloud_[resource_name]_schema.go              # ✅ Updated (Framework-only registration)
├── genesyscloud_[resource_name]_proxy.go                       # ✅ Keep (Shared API layer)
├── framework_resource_genesyscloud_[resource_name].go          # ✅ New (Framework resource)
├── framework_data_source_genesyscloud_[resource_name].go       # ✅ New (Framework data source)
├── framework_resource_genesyscloud_[resource_name]_test.go     # ✅ New (Framework resource tests)
├── framework_data_source_genesyscloud_[resource_name]_test.go  # ✅ New (Framework data source tests)
└── genesyscloud_[resource_name]_init_test.go                   # ✅ Updated (Framework-only)

# Files to REMOVE after migration:
├── resource_genesyscloud_[resource_name].go            # ❌ Remove (SDKv2 resource)
├── data_source_genesyscloud_[resource_name].go         # ❌ Remove (SDKv2 data source)
├── resource_genesyscloud_[resource_name]_test.go       # ❌ Remove (SDKv2 tests)
└── data_source_genesyscloud_[resource_name]_test.go    # ❌ Remove (SDKv2 tests)
```

## 🔄 Migration Process

### Phase 1: Framework Implementation (Week 1)

#### Step 1.1: Framework Resource Implementation
**File**: `framework_resource_genesyscloud_[resource_name].go`

**Template Structure**:
```go
package [resource_name]

import (
    "context"
    "fmt"
    "time"

    "github.com/hashicorp/terraform-plugin-framework/resource"
    "github.com/hashicorp/terraform-plugin-framework/resource/schema"
    "github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
    "github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
    "github.com/hashicorp/terraform-plugin-framework/types"
    "github.com/hashicorp/terraform-plugin-log/tflog"
    "github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
    "github.com/mypurecloud/platform-client-sdk-go/v176/platformclientv2"
    "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
    "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
)

// Ensure the implementation satisfies the expected interfaces
var (
    _ resource.Resource                = &[resourceName]FrameworkResource{}
    _ resource.ResourceWithConfigure   = &[resourceName]FrameworkResource{}
    _ resource.ResourceWithImportState = &[resourceName]FrameworkResource{}
)

func New[ResourceName]FrameworkResource() resource.Resource {
    return &[resourceName]FrameworkResource{}
}

type [resourceName]FrameworkResource struct {
    clientConfig *platformclientv2.Configuration
}

type [resourceName]FrameworkResourceModel struct {
    Id   types.String `tfsdk:"id"`
    Name types.String `tfsdk:"name"`
    // Add other fields as needed
}

func (r *[resourceName]FrameworkResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
    resp.TypeName = req.ProviderTypeName + "_[resource_type]"
}

func (r *[resourceName]FrameworkResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
    resp.Schema = schema.Schema{
        Description: "[Resource description]",
        Attributes: map[string]schema.Attribute{
            "id": schema.StringAttribute{
                Description: "The globally unique identifier for the resource.",
                Computed:    true,
                PlanModifiers: []planmodifier.String{
                    stringplanmodifier.UseStateForUnknown(),
                },
            },
            "name": schema.StringAttribute{
                Description: "The name of the [resource].",
                Required:    true,
                PlanModifiers: []planmodifier.String{
                    stringplanmodifier.RequiresReplace(),
                },
            },
            // Add other attributes as needed
        },
    }
}

func (r *[resourceName]FrameworkResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
    if req.ProviderData == nil {
        return
    }

    providerMeta, ok := req.ProviderData.(*provider.ProviderMeta)
    if !ok {
        resp.Diagnostics.AddError(
            "Unexpected Resource Configure Type",
            fmt.Sprintf("Expected *provider.ProviderMeta, got: %T. Please report this issue to the provider developers.", req.ProviderData),
        )
        return
    }

    r.clientConfig = providerMeta.ClientConfig
}

func (r *[resourceName]FrameworkResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
    var plan [resourceName]FrameworkResourceModel
    resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
    if resp.Diagnostics.HasError() {
        return
    }

    proxy := get[ResourceName]Proxy(r.clientConfig)
    
    // Create the resource using proxy
    [resourceVar], _, err := proxy.create[ResourceName](ctx, &platformclientv2.[ResourceType]{
        Name: plan.Name.ValueStringPointer(),
    })
    
    if err != nil {
        resp.Diagnostics.AddError(
            "Failed to create [resource]",
            fmt.Sprintf("Failed to create [resource] %s: %s", plan.Name.ValueString(), err),
        )
        return
    }

    // Set the state
    plan.Id = types.StringValue(*[resourceVar].Id)
    plan.Name = types.StringValue(*[resourceVar].Name)

    resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *[resourceName]FrameworkResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
    var state [resourceName]FrameworkResourceModel
    resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
    if resp.Diagnostics.HasError() {
        return
    }

    proxy := get[ResourceName]Proxy(r.clientConfig)
    
    [resourceVar], apiResp, err := proxy.get[ResourceName]ById(ctx, state.Id.ValueString())
    if err != nil {
        if util.IsStatus404(apiResp) {
            resp.State.RemoveResource(ctx)
            return
        }
        resp.Diagnostics.AddError(
            "Failed to read [resource]",
            fmt.Sprintf("Failed to read [resource] %s: %s", state.Id.ValueString(), err),
        )
        return
    }

    // Update state with current values
    state.Name = types.StringValue(*[resourceVar].Name)

    resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *[resourceName]FrameworkResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
    // Most Genesys Cloud resources don't support updates - they require replacement
    resp.Diagnostics.AddError(
        "Update not supported",
        "[Resource] does not support updates. All changes require resource replacement.",
    )
}

func (r *[resourceName]FrameworkResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
    var state [resourceName]FrameworkResourceModel
    resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
    if resp.Diagnostics.HasError() {
        return
    }

    proxy := get[ResourceName]Proxy(r.clientConfig)
    
    // Delete with retry logic
    err := retry.RetryContext(ctx, 30*time.Second, func() *retry.RetryError {
        _, apiResp, err := proxy.delete[ResourceName](ctx, state.Id.ValueString())
        if err != nil {
            if util.IsStatus404(apiResp) {
                return nil // Already deleted
            }
            return retry.NonRetryableError(fmt.Errorf("failed to delete [resource] %s: %s", state.Id.ValueString(), err))
        }
        return nil
    })

    if err != nil {
        resp.Diagnostics.AddError(
            "Failed to delete [resource]",
            err.Error(),
        )
        return
    }
}

func (r *[resourceName]FrameworkResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
    resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
```

**Key Implementation Points**:
- [ ] Use existing proxy for all API calls
- [ ] Implement proper Framework error handling (no `diag.Diagnostics`)
- [ ] Use `types.String` for all string attributes
- [ ] Add proper plan modifiers (`RequiresReplace`, `UseStateForUnknown`)
- [ ] Implement retry logic using `retry.RetryContext()`
- [ ] Handle 404 responses properly in Read and Delete methods

#### Step 1.2: Framework Data Source Implementation
**File**: `framework_data_source_genesyscloud_[resource_name].go`

**Template Structure**:
```go
package [resource_name]

import (
    "context"
    "fmt"
    "time"

    "github.com/hashicorp/terraform-plugin-framework/datasource"
    "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
    "github.com/hashicorp/terraform-plugin-framework/types"
    "github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
    "github.com/mypurecloud/platform-client-sdk-go/v176/platformclientv2"
    "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
)

// Ensure the implementation satisfies the expected interfaces
var (
    _ datasource.DataSource              = &[resourceName]FrameworkDataSource{}
    _ datasource.DataSourceWithConfigure = &[resourceName]FrameworkDataSource{}
)

func New[ResourceName]FrameworkDataSource() datasource.DataSource {
    return &[resourceName]FrameworkDataSource{}
}

type [resourceName]FrameworkDataSource struct {
    clientConfig *platformclientv2.Configuration
}

type [resourceName]FrameworkDataSourceModel struct {
    Id   types.String `tfsdk:"id"`
    Name types.String `tfsdk:"name"`
    // Add other fields as needed
}

func (d *[resourceName]FrameworkDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
    resp.TypeName = req.ProviderTypeName + "_[resource_type]"
}

func (d *[resourceName]FrameworkDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
    resp.Schema = schema.Schema{
        Description: "[Resource] data source",
        Attributes: map[string]schema.Attribute{
            "id": schema.StringAttribute{
                Description: "The globally unique identifier for the [resource].",
                Computed:    true,
            },
            "name": schema.StringAttribute{
                Description: "The name of the [resource].",
                Required:    true,
            },
            // Add other attributes as needed
        },
    }
}

func (d *[resourceName]FrameworkDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
    if req.ProviderData == nil {
        return
    }

    providerMeta, ok := req.ProviderData.(*provider.ProviderMeta)
    if !ok {
        resp.Diagnostics.AddError(
            "Unexpected Data Source Configure Type",
            fmt.Sprintf("Expected *provider.ProviderMeta, got: %T. Please report this issue to the provider developers.", req.ProviderData),
        )
        return
    }

    d.clientConfig = providerMeta.ClientConfig
}

func (d *[resourceName]FrameworkDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
    var config [resourceName]FrameworkDataSourceModel
    resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
    if resp.Diagnostics.HasError() {
        return
    }

    proxy := get[ResourceName]Proxy(d.clientConfig)
    
    // Use retry logic for eventual consistency
    var [resourceVar]Id string
    err := retry.RetryContext(ctx, 15*time.Second, func() *retry.RetryError {
        id, _, retryable, err := proxy.get[ResourceName]IdByName(ctx, config.Name.ValueString())
        if err != nil && !retryable {
            return retry.NonRetryableError(err)
        }
        if err != nil {
            return retry.RetryableError(err)
        }
        [resourceVar]Id = id
        return nil
    })

    if err != nil {
        resp.Diagnostics.AddError(
            "Failed to find [resource]",
            fmt.Sprintf("Failed to find [resource] with name '%s': %s", config.Name.ValueString(), err),
        )
        return
    }

    config.Id = types.StringValue([resourceVar]Id)
    resp.Diagnostics.Append(resp.State.Set(ctx, config)...)
}
```

#### Step 1.3: Framework Testing Implementation
**File**: `framework_resource_genesyscloud_[resource_name]_test.go`

**Template Structure**:
```go
package [resource_name]

import (
    "fmt"
    "testing"

    "github.com/google/uuid"
    "github.com/hashicorp/terraform-plugin-framework/datasource"
    frameworkresource "github.com/hashicorp/terraform-plugin-framework/resource"
    "github.com/hashicorp/terraform-plugin-go/tfprotov6"
    "github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
    "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
    "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
)

func TestAccFrameworkResource[ResourceName]Basic(t *testing.T) {
    var (
        resourceLabel = "test_[resource_name]"
        name         = "test-[resource-name]-" + uuid.NewString()
    )

    resource.Test(t, resource.TestCase{
        PreCheck:                 func() { util.TestAccPreCheck(t) },
        ProtoV6ProviderFactories: getMuxedProviderFactories(),
        Steps: []resource.TestStep{
            {
                Config: generate[ResourceName]Resource(resourceLabel, name),
                Check: resource.ComposeTestCheckFunc(
                    resource.TestCheckResourceAttr("genesyscloud_[resource_type]."+resourceLabel, "name", name),
                    resource.TestCheckResourceAttrSet("genesyscloud_[resource_type]."+resourceLabel, "id"),
                ),
            },
            // Import test
            {
                ResourceName:      "genesyscloud_[resource_type]." + resourceLabel,
                ImportState:       true,
                ImportStateVerify: true,
            },
        },
    })
}

func TestAccFrameworkResource[ResourceName]ForceNew(t *testing.T) {
    var (
        resourceLabel = "test_[resource_name]"
        name1        = "test-[resource-name]-" + uuid.NewString()
        name2        = "test-[resource-name]-" + uuid.NewString()
    )

    resource.Test(t, resource.TestCase{
        PreCheck:                 func() { util.TestAccPreCheck(t) },
        ProtoV6ProviderFactories: getMuxedProviderFactories(),
        Steps: []resource.TestStep{
            {
                Config: generate[ResourceName]Resource(resourceLabel, name1),
                Check: resource.ComposeTestCheckFunc(
                    resource.TestCheckResourceAttr("genesyscloud_[resource_type]."+resourceLabel, "name", name1),
                ),
            },
            {
                Config: generate[ResourceName]Resource(resourceLabel, name2),
                Check: resource.ComposeTestCheckFunc(
                    resource.TestCheckResourceAttr("genesyscloud_[resource_type]."+resourceLabel, "name", name2),
                ),
            },
        },
    })
}

// getMuxedProviderFactories returns muxed provider factories for Framework testing
func getMuxedProviderFactories() map[string]func() (tfprotov6.ProviderServer, error) {
    return provider.GetMuxedProviderFactories(
        providerResources,
        providerDataSources,
        map[string]func() frameworkresource.Resource{
            ResourceType: New[ResourceName]FrameworkResource,
        },
        map[string]func() datasource.DataSource{
            ResourceType: New[ResourceName]FrameworkDataSource,
        },
    )
}

func generate[ResourceName]Resource(resourceLabel, name string) string {
    return fmt.Sprintf(`
resource "genesyscloud_[resource_type]" "%s" {
    name = "%s"
}
`, resourceLabel, name)
}
```

**File**: `framework_data_source_genesyscloud_[resource_name]_test.go`

```go
func TestAccFrameworkDataSource[ResourceName](t *testing.T) {
    var (
        resourceLabel    = "test_[resource_name]"
        dataSourceLabel  = "test_[resource_name]_ds"
        name            = "test-[resource-name]-" + uuid.NewString()
    )

    resource.Test(t, resource.TestCase{
        PreCheck:                 func() { util.TestAccPreCheck(t) },
        ProtoV6ProviderFactories: getMuxedProviderFactories(),
        Steps: []resource.TestStep{
            {
                Config: generate[ResourceName]Resource(resourceLabel, name) + 
                       generate[ResourceName]DataSource(dataSourceLabel, name),
                Check: resource.ComposeTestCheckFunc(
                    resource.TestCheckResourceAttrPair(
                        "data.genesyscloud_[resource_type]."+dataSourceLabel, "id",
                        "genesyscloud_[resource_type]."+resourceLabel, "id",
                    ),
                    resource.TestCheckResourceAttr("data.genesyscloud_[resource_type]."+dataSourceLabel, "name", name),
                ),
            },
        },
    })
}

func generate[ResourceName]DataSource(dataSourceLabel, name string) string {
    return fmt.Sprintf(`
data "genesyscloud_[resource_type]" "%s" {
    name = "%s"
}
`, dataSourceLabel, name)
}
```

### Phase 2: Migration Execution (Week 2)

#### Step 2.1: Update Registration to Framework-Only
**File**: `resource_genesyscloud_[resource_name]_schema.go`

**Before**:
```go
func SetRegistrar(regInstance registrar.Registrar) {
    regInstance.RegisterResource(ResourceType, Resource[ResourceName]())
    regInstance.RegisterDataSource(ResourceType, DataSource[ResourceName]())
    regInstance.RegisterExporter(ResourceType, [ResourceName]Exporter())
}
```

**After**:
```go
func SetRegistrar(regInstance registrar.Registrar) {
    // Framework-only registration
    regInstance.RegisterFrameworkResource(ResourceType, New[ResourceName]FrameworkResource)
    regInstance.RegisterFrameworkDataSource(ResourceType, New[ResourceName]FrameworkDataSource)
    regInstance.RegisterExporter(ResourceType, [ResourceName]Exporter())
}
```

**Checklist**:
- [ ] Remove SDKv2 resource registration
- [ ] Remove SDKv2 data source registration
- [ ] Add Framework resource registration
- [ ] Add Framework data source registration
- [ ] Keep exporter registration (works with both)

#### Step 2.2: Update Export Function (if needed)
If the resource supports export functionality, ensure the export function exists:

```go
// GetAll[ResourceName]s retrieves all [resource]s for export using the proxy
func GetAll[ResourceName]s(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
    proxy := get[ResourceName]Proxy(clientConfig)
    [resources], _, err := proxy.getAll[ResourceName]s(ctx, "")
    if err != nil {
        return nil, diag.Errorf("Failed to get [resource]s for export: %v", err)
    }

    if [resources] == nil {
        return resourceExporter.ResourceIDMetaMap{}, nil
    }

    exportMap := make(resourceExporter.ResourceIDMetaMap)
    for _, [resource] := range *[resources] {
        exportMap[*[resource].Id] = &resourceExporter.ResourceMeta{
            BlockLabel: *[resource].Name,
        }
    }
    return exportMap, nil
}

func [ResourceName]Exporter() *resourceExporter.ResourceExporter {
    return &resourceExporter.ResourceExporter{
        GetResourcesFunc: provider.GetAllWithPooledClient(GetAll[ResourceName]s),
    }
}
```

#### Step 2.3: Remove SDKv2 Implementation Files
**Files to Delete**:
- [ ] `resource_genesyscloud_[resource_name].go` - SDKv2 resource implementation
- [ ] `data_source_genesyscloud_[resource_name].go` - SDKv2 data source implementation
- [ ] `resource_genesyscloud_[resource_name]_test.go` - SDKv2 resource tests
- [ ] `data_source_genesyscloud_[resource_name]_test.go` - SDKv2 data source tests

**Verification Steps**:
- [ ] Confirm no imports reference removed files
- [ ] Verify Framework tests still pass after removal
- [ ] Check that registration only includes Framework implementations

#### Step 2.4: Update Test Infrastructure
**File**: `genesyscloud_[resource_name]_init_test.go`

**Before**:
```go
func initTestResources() {
    providerResources = make(map[string]*schema.Resource)
    providerDataSources = make(map[string]*schema.Resource)
    
    regInstance := &registerTestInstance{}
    regInstance.registerTestResources()
    regInstance.registerTestDataSources()
}

func (r *registerTestInstance) registerTestResources() {
    providerResources[ResourceType] = Resource[ResourceName]()
}

func (r *registerTestInstance) registerTestDataSources() {
    providerDataSources[ResourceType] = DataSource[ResourceName]()
}
```

**After**:
```go
func initTestResources() {
    providerResources = make(map[string]*schema.Resource)
    providerDataSources = make(map[string]*schema.Resource)
    
    // Framework-only - no test registration needed
    // Framework resources are handled by muxed provider in tests
}
```

#### Step 2.5: Fix Cross-Package Test Dependencies ⚠️ **CRITICAL DISCOVERY**

**🔥 Major Anti-Pattern Discovered**: Multiple packages often have nearly identical custom `getMuxedProviderFactoriesFor[Package]()` functions that create code duplication and maintenance burden.

**Analysis Required**:
1. **Search for duplicate functions** across all packages:
   ```bash
   grep -r "getMuxedProviderFactoriesFor" genesyscloud/*/
   ```
2. **Identify code duplication patterns** in test files
3. **Assess maintenance burden** and consistency risks

**❌ Anti-Pattern (Code Duplication)**:
```go
// Each package has its own nearly identical custom function
func getMuxedProviderFactoriesForOutboundCampaign() map[string]func() (tfprotov6.ProviderServer, error) {
    return provider.GetMuxedProviderFactories(
        providerResources, providerDataSources,
        map[string]func() frameworkresource.Resource{
            routingWrapupcode.ResourceType: routingWrapupcode.NewRoutingWrapupcodeFrameworkResource,
        },
        map[string]func() datasource.DataSource{
            routingWrapupcode.ResourceType: routingWrapupcode.NewRoutingWrapupcodeFrameworkDataSource,
        },
    )
}

func getMuxedProviderFactoriesForRoutingQueue() map[string]func() (tfprotov6.ProviderServer, error) {
    return provider.GetMuxedProviderFactories(
        providerResources, providerDataSources,
        map[string]func() frameworkresource.Resource{
            routingWrapupcode.ResourceType: routingWrapupcode.NewRoutingWrapupcodeFrameworkResource,
        },
        map[string]func() datasource.DataSource{
            routingWrapupcode.ResourceType: routingWrapupcode.NewRoutingWrapupcodeFrameworkDataSource,
        },
    )
}
// ... 4+ more nearly identical functions across packages
```

**✅ Correct Pattern (DRY Principle)**:
```go
// Direct usage in each test case - no custom functions needed
resource.Test(t, resource.TestCase{
    PreCheck: func() { util.TestAccPreCheck(t) },
    ProtoV6ProviderFactories: provider.GetMuxedProviderFactories(
        providerResources,
        providerDataSources,
        map[string]func() frameworkresource.Resource{
            [resourceName].ResourceType: [resourceName].New[ResourceName]FrameworkResource,
        },
        map[string]func() datasource.DataSource{
            [resourceName].ResourceType: [resourceName].New[ResourceName]FrameworkDataSource,
        },
    ),
    // ... test steps
})
```

**Required Imports for Framework Integration**:
```go
import (
    "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
    [resourceName] "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/[resource_name]"
    "github.com/hashicorp/terraform-plugin-framework/datasource"
    frameworkresource "github.com/hashicorp/terraform-plugin-framework/resource"
)
```

**Migration Impact**:
- **Eliminate all custom `getMuxedProviderFactoriesFor[Package]()` functions**
- **Update all affected test files** with consistent direct usage pattern
- **Add required imports** to all affected files
- **Establish reusable template** for future Framework migrations
- **Improve overall codebase architecture** by removing technical debt

### Phase 3: Validation & Testing (Week 2)

#### Step 3.1: Compilation Verification
- [ ] `go build ./genesyscloud/[resource_name]/` - No compilation errors
- [ ] `go vet ./genesyscloud/[resource_name]/` - No vet warnings
- [ ] Search for any remaining references to deleted SDKv2 functions

#### Step 3.2: Framework Testing
- [ ] Run Framework resource tests: `go test ./genesyscloud/[resource_name]/ -run TestAccFrameworkResource`
- [ ] Run Framework data source tests: `go test ./genesyscloud/[resource_name]/ -run TestAccFrameworkDataSource`
- [ ] Verify all test scenarios pass (basic, force new, error handling, import)

#### Step 3.3: Cross-Package Testing
- [ ] Run tests in packages that depend on this resource
- [ ] Verify muxed provider factories work correctly
- [ ] Check that Framework resources are accessible from other tests

#### Step 3.4: Export Functionality Testing
If the resource supports export:
- [ ] Test terraform export functionality
- [ ] Verify exported configurations are valid
- [ ] Check that Framework resources appear in export output

## 🚨 Common Issues & Solutions

### Issue 1: Compilation Errors After SDKv2 Removal
**Symptoms**: `undefined: [resourceName].Resource[ResourceName]`
**Solution**: Remove all SDKv2 registrations from test files across packages

### Issue 2: Framework Resources Not Found in Tests
**Symptoms**: `The provider does not support resource type`
**Solution**: Use muxed provider factories in tests that need Framework resources

### Issue 3: Export Functionality Broken
**Symptoms**: `undefined: getAll[ResourceName]s`
**Solution**: Create new export function using proxy layer

### Issue 4: Circular Import Dependencies
**Symptoms**: `import cycle not allowed`
**Solution**: Use `resource_register` package instead of `provider_registrar` in test files

### Issue 5: Empty Registrar Functions
**Symptoms**: Framework resources not accessible to export system
**Solution**: Implement proper Registrar interface methods that store resources in global maps

### Issue 6: Cross-Package Code Duplication (Major Discovery)
**Symptoms**: Multiple packages have nearly identical `getMuxedProviderFactoriesFor[Package]()` functions
**Impact**: Code duplication, maintenance burden, inconsistency risk across 6+ packages
**Solution**: 
1. **Eliminate all custom functions** - Remove duplicate `getMuxedProviderFactoriesFor[Package]()` functions
2. **Use direct pattern** - Replace with direct `provider.GetMuxedProviderFactories()` calls in test cases
3. **Add required imports** - Ensure all files have necessary Framework imports
4. **Establish consistency** - Create standardized approach for all packages
5. **Document pattern** - Make this the standard for future Framework migrations

**Files Typically Affected**:
- `genesyscloud/routing_queue/*_test.go`
- `genesyscloud/outbound_campaign/*_test.go`
- `genesyscloud/outbound_sequence/*_test.go`
- `genesyscloud/outbound_campaignrule/*_test.go`
- `genesyscloud/outbound_callanalysisresponseset/*_test.go`
- `genesyscloud/outbound_wrapupcode_mappings/*_test.go`

## ✅ Success Criteria

### Functional Requirements
- [ ] Framework resource behaves identically to previous SDKv2 implementation
- [ ] All CRUD operations work correctly
- [ ] Data source lookup works properly
- [ ] Import functionality works
- [ ] Error handling provides clear messages
- [ ] Export functionality works (if applicable)

### Testing Requirements
- [ ] Framework tests provide equivalent coverage to previous SDKv2 tests
- [ ] All test scenarios pass (basic, force new, error handling, import)
- [ ] Cross-package tests work with muxed provider factories
- [ ] No test regressions

### Architectural Requirements
- [ ] Clean Framework implementation following best practices
- [ ] Proper error handling and state management
- [ ] Single implementation to maintain (Framework only)
- [ ] Zero breaking changes for existing Terraform configurations

## 📊 Migration Checklist

### Pre-Migration
- [ ] Resource assessment completed
- [ ] Dependencies identified
- [ ] Test environment prepared
- [ ] Backup of current implementation

### Implementation Phase
- [ ] Framework resource implemented
- [ ] Framework data source implemented
- [ ] Framework tests created
- [ ] Export function updated (if needed)

### Migration Phase
- [ ] Registration updated to Framework-only
- [ ] SDKv2 files removed
- [ ] Test infrastructure updated
- [ ] Cross-package dependencies fixed

### Validation Phase
- [ ] Compilation successful
- [ ] Framework tests passing
- [ ] Cross-package tests passing
- [ ] Export functionality working
- [ ] No regressions identified

### Documentation Phase
- [ ] Migration documented
- [ ] Lessons learned captured
- [ ] Template updated with new insights

## 🎯 Expected Timeline

### Simple Resources (1-2 weeks)
- **Week 1**: Framework implementation and testing
- **Week 2**: Migration execution and validation

### Medium Resources (2-3 weeks)
- **Week 1-2**: Framework implementation and comprehensive testing
- **Week 3**: Migration execution, validation, and cross-package fixes

### Complex Resources (3-4 weeks)
- **Week 1-2**: Framework implementation with complex schema/logic
- **Week 3**: Comprehensive testing and integration
- **Week 4**: Migration execution, validation, and system-wide testing

## 🏆 Success Metrics

- ✅ **Zero breaking changes** to existing Terraform configurations
- ✅ **Framework resource behaves identically** to previous SDKv2 implementation
- ✅ **Complete SDKv2 removal** - no legacy code remaining
- ✅ **Simplified architecture** - single implementation per resource
- ✅ **Comprehensive testing** - Framework tests cover all scenarios
- ✅ **Export functionality preserved** (if applicable)
- ✅ **Cross-package compatibility** maintained

## 📚 Additional Resources

### Framework Documentation
- [Terraform Plugin Framework](https://developer.hashicorp.com/terraform/plugin/framework)
- [Framework Resource Implementation](https://developer.hashicorp.com/terraform/plugin/framework/resources)
- [Framework Data Source Implementation](https://developer.hashicorp.com/terraform/plugin/framework/data-sources)

### Internal References
- `genesyscloud/routing_language/` - Complete Framework migration example
- `genesyscloud/provider/provider_utils.go` - Centralized muxed provider factory
- `genesyscloud/resource_register/` - Framework resource registration system

### Migration Examples
- **Simple Resource**: `genesyscloud_routing_language` - Single field, basic CRUD
- **Medium Resource**: [Add examples as more resources are migrated]
- **Complex Resource**: [Add examples as more resources are migrated]

---

## 🎉 Conclusion

This template provides a comprehensive guide for migrating Terraform resources from SDKv2 to Plugin Framework using the proven Framework-Only approach. The template is based on the successful migration of `genesyscloud_routing_language` and incorporates all lessons learned from that experience.

**Key Success Factors**:
1. **Complete replacement approach** - Eliminates complexity of parallel implementations
2. **Proper test infrastructure** - Framework resources properly registered and accessible
3. **Cross-package compatibility** - Muxed provider factories enable integration
4. **Export system integration** - Framework resources work with terraform export
5. **Clean architecture** - Single implementation, modern patterns, maintainable code

Follow this template step-by-step for a successful Framework migration that maintains functionality while modernizing the codebase.