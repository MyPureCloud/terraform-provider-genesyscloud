# Stage 2 – Resource Migration Design

## Overview

This document describes the design decisions, architecture, and patterns used in Stage 2 of the Plugin Framework migration. Stage 2 focuses on implementing resource CRUD operations and data source read logic using Plugin Framework patterns while reusing existing business logic from proxy methods.

**Reference Implementation**: 
- `genesyscloud/routing_wrapupcode/resource_genesyscloud_routing_wrapupcode.go`
- `genesyscloud/routing_wrapupcode/data_source_genesyscloud_routing_wrapupcode.go`

---

## Design Principles

### 1. Framework-Native Lifecycle
**Principle**: Use Plugin Framework lifecycle patterns instead of SDKv2 callback functions.

**Rationale**:
- Framework provides better type safety and error handling
- Context-based cancellation support
- Clearer separation between plan and state
- Better diagnostics and error reporting

**Implementation**:
- Resource methods: `Create()`, `Read()`, `Update()`, `Delete()`, `ImportState()`
- Data source methods: `Read()`
- All methods use context for cancellation
- All methods use Framework diagnostics

### 2. Proxy Reuse
**Principle**: Reuse existing proxy methods without modification.

**Rationale**:
- Proxy methods contain tested business logic
- Proxy methods are shared between SDKv2 and Framework
- No need to duplicate API interaction code
- Reduces migration risk

**Implementation**:
- Call proxy methods from CRUD operations
- Convert Framework models to API requests
- Convert API responses to Framework models
- No changes to proxy signatures or behavior

### 3. Type Safety with Framework Types
**Principle**: Use Framework types (`types.String`) instead of pointers for state management.

**Rationale**:
- Framework requires specific types for null/unknown handling
- Better type safety at compile time
- Explicit null/unknown state representation
- Clearer intent in code

**Implementation**:
- Model structs use `types.String`, `types.Int64`, etc.
- Helper functions convert between Framework types and API types
- Explicit null/unknown checks before API calls

### 4. Dual GetAll Functions (Phase 1 Temporary)
**Principle**: Provide both Framework and SDK versions of GetAll functions during migration.

**Rationale**:
- Exporter currently uses SDK diagnostics
- Framework version prepared for Phase 2 future
- Smooth transition path
- No breaking changes during migration

**Implementation**:
- `GetAll<ResourceName>()`: Framework version (Phase 2 future)
- `GetAll<ResourceName>SDK()`: SDK version (Phase 1 current, used by exporter)
- Both return same data structure
- Clear comments marking temporary code

---

## Architecture

### File Structure

```
genesyscloud/<resource_name>/
├── resource_genesyscloud_<resource_name>_schema.go          ← Stage 1
├── resource_genesyscloud_<resource_name>.go                 ← Stage 2 (THIS FILE)
├── data_source_genesyscloud_<resource_name>.go              ← Stage 2 (THIS FILE)
├── resource_genesyscloud_<resource_name>_test.go            ← Stage 3
├── data_source_genesyscloud_<resource_name>_test.go         ← Stage 3
├── resource_genesyscloud_<resource_name>_export_utils.go    ← Stage 4
├── genesyscloud_<resource_name>_init_test.go                ← Stage 3
└── genesyscloud_<resource_name>_proxy.go                    ← NOT MODIFIED
```

### Resource File Components

```
┌─────────────────────────────────────────────────────────┐
│  resource_genesyscloud_<resource_name>.go               │
├─────────────────────────────────────────────────────────┤
│  1. Interface Verification (compile-time checks)        │
├─────────────────────────────────────────────────────────┤
│  2. Resource Struct                                     │
│     - clientConfig field                                │
├─────────────────────────────────────────────────────────┤
│  3. Resource Model Struct                               │
│     - Framework types (types.String, etc.)              │
│     - tfsdk struct tags                                 │
├─────────────────────────────────────────────────────────┤
│  4. Constructor Function                                │
│     - New<ResourceName>FrameworkResource()              │
├─────────────────────────────────────────────────────────┤
│  5. Resource Interface Methods                          │
│     - Metadata()                                        │
│     - Schema()                                          │
│     - Configure()                                       │
├─────────────────────────────────────────────────────────┤
│  6. CRUD Methods                                        │
│     - Create()                                          │
│     - Read()                                            │
│     - Update()                                          │
│     - Delete()                                          │
│     - ImportState()                                     │
├─────────────────────────────────────────────────────────┤
│  7. Helper Functions                                    │
│     - build<ResourceName>FromFrameworkModel()           │
│     - updateFrameworkModelFromAPI()                     │
├─────────────────────────────────────────────────────────┤
│  8. GetAll Functions (for export)                       │
│     - GetAll<ResourceName>() - Framework (Phase 2)      │
│     - GetAll<ResourceName>SDK() - SDK (Phase 1)         │
└─────────────────────────────────────────────────────────┘
```

### Data Source File Components

```
┌─────────────────────────────────────────────────────────┐
│  data_source_genesyscloud_<resource_name>.go            │
├─────────────────────────────────────────────────────────┤
│  1. Interface Verification (compile-time checks)        │
├─────────────────────────────────────────────────────────┤
│  2. Data Source Struct                                  │
│     - clientConfig field                                │
├─────────────────────────────────────────────────────────┤
│  3. Data Source Model Struct                            │
│     - Framework types (types.String, etc.)              │
│     - tfsdk struct tags                                 │
├─────────────────────────────────────────────────────────┤
│  4. Constructor Function                                │
│     - New<ResourceName>FrameworkDataSource()            │
├─────────────────────────────────────────────────────────┤
│  5. Data Source Interface Methods                       │
│     - Metadata()                                        │
│     - Schema()                                          │
│     - Configure()                                       │
│     - Read()                                            │
└─────────────────────────────────────────────────────────┘
```

---

## Component Design

## Part 1: Resource Implementation

### 1. Interface Verification

**Purpose**: Compile-time verification that resource implements required interfaces.

**Design Pattern**:
```go
// Ensure <resource>FrameworkResource satisfies various resource interfaces.
var (
    _ resource.Resource                = &<resource>FrameworkResource{}
    _ resource.ResourceWithConfigure   = &<resource>FrameworkResource{}
    _ resource.ResourceWithImportState = &<resource>FrameworkResource{}
)
```

**Example** (routing_wrapupcode):
```go
var (
    _ resource.Resource                = &routingWrapupcodeFrameworkResource{}
    _ resource.ResourceWithConfigure   = &routingWrapupcodeFrameworkResource{}
    _ resource.ResourceWithImportState = &routingWrapupcodeFrameworkResource{}
)
```

**Why Important**:
- Catches missing interface methods at compile time
- Documents which interfaces are implemented
- Prevents runtime errors from missing methods

---

### 2. Resource Struct

**Purpose**: Hold resource-level configuration and dependencies.

**Design Pattern**:
```go
// <resource>FrameworkResource defines the resource implementation for Plugin Framework.
type <resource>FrameworkResource struct {
    clientConfig *platformclientv2.Configuration
}
```

**Example** (routing_wrapupcode):
```go
type routingWrapupcodeFrameworkResource struct {
    clientConfig *platformclientv2.Configuration
}
```

**Design Decisions**:

| Decision | Rationale |
|----------|-----------|
| Store `clientConfig` | Needed to create proxy instances in CRUD methods |
| No other fields | Keep resource stateless; all state in Terraform state |
| Lowercase struct name | Internal implementation detail, not exported |

---

### 3. Resource Model Struct

**Purpose**: Define the data structure for Terraform state and plan.

**Design Pattern**:
```go
// <resource>FrameworkResourceModel describes the resource data model.
type <resource>FrameworkResourceModel struct {
    Id          types.String `tfsdk:"id"`
    Name        types.String `tfsdk:"name"`
    // ... other attributes
}
```

**Example** (routing_wrapupcode):
```go
type routingWrapupcodeFrameworkResourceModel struct {
    Id          types.String `tfsdk:"id"`
    Name        types.String `tfsdk:"name"`
    DivisionId  types.String `tfsdk:"division_id"`
    Description types.String `tfsdk:"description"`
}
```

**Design Decisions**:

| Decision | Rationale |
|----------|-----------|
| Use Framework types | Required by Framework for null/unknown handling |
| `tfsdk` struct tags | Map struct fields to schema attribute names |
| Match schema exactly | Every schema attribute has corresponding model field |
| No pointers | Framework types handle null/unknown internally |

**Framework Type Usage**:
- `types.String` for string attributes
- `types.Int64` for integer attributes
- `types.Bool` for boolean attributes
- `types.Float64` for float attributes
- `types.List` for list attributes
- `types.Set` for set attributes
- `types.Map` for map attributes

---

### 4. Constructor Function

**Purpose**: Create new resource instances for the provider.

**Design Pattern**:
```go
// New<ResourceName>FrameworkResource is a helper function to simplify the provider implementation.
func New<ResourceName>FrameworkResource() resource.Resource {
    return &<resource>FrameworkResource{}
}
```

**Example** (routing_wrapupcode):
```go
func NewRoutingWrapupcodeFrameworkResource() resource.Resource {
    return &routingWrapupcodeFrameworkResource{}
}
```

**Why This Pattern**:
- Provider calls this function to create resource instances
- Returns interface type (`resource.Resource`) for flexibility
- Simple factory pattern
- No initialization logic needed (configuration happens in Configure method)

---

### 5. Resource Interface Methods

#### 5.1 Metadata Method

**Purpose**: Provide resource type name to the provider.

**Design Pattern**:
```go
func (r *<resource>FrameworkResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
    resp.TypeName = req.ProviderTypeName + "_<resource_name>"
}
```

**Example** (routing_wrapupcode):
```go
func (r *routingWrapupcodeFrameworkResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
    resp.TypeName = req.ProviderTypeName + "_routing_wrapupcode"
}
```

**Key Points**:
- Concatenates provider name with resource name
- Results in full type: `genesyscloud_routing_wrapupcode`
- Must match ResourceType constant from schema file

#### 5.2 Schema Method

**Purpose**: Provide resource schema to the provider.

**Design Pattern**:
```go
func (r *<resource>FrameworkResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
    resp.Schema = <ResourceName>ResourceSchema()
}
```

**Example** (routing_wrapupcode):
```go
func (r *routingWrapupcodeFrameworkResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
    resp.Schema = RoutingWrapupcodeResourceSchema()
}
```

**Key Points**:
- Calls schema function from Stage 1
- No schema logic in resource file
- Keeps schema and implementation separate

#### 5.3 Configure Method

**Purpose**: Receive provider configuration and store client config.

**Design Pattern**:
```go
func (r *<resource>FrameworkResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
    // Prevent panic if the provider has not been configured.
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
```

**Key Points**:
- Called once during provider initialization
- Stores `clientConfig` for use in CRUD methods
- Type assertion with error handling
- Graceful handling of nil provider data

---

### 6. CRUD Methods

#### 6.1 Create Method

**Purpose**: Create a new resource via API and store in Terraform state.

**Design Pattern**:
```go
func (r *<resource>FrameworkResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
    var plan <resource>FrameworkResourceModel

    // Read Terraform plan data into the model
    resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
    if resp.Diagnostics.HasError() {
        return
    }

    proxy := get<ResourceName>Proxy(r.clientConfig)
    apiRequest := build<ResourceName>FromFrameworkModel(plan)

    log.Printf("Creating <resource> %s", plan.Name.ValueString())

    apiResponse, _, err := proxy.create<ResourceName>(ctx, apiRequest)
    if err != nil {
        resp.Diagnostics.AddError(
            "Error Creating <Resource>",
            fmt.Sprintf("Could not create <resource> %s: %s", plan.Name.ValueString(), err),
        )
        return
    }

    // Update model with response data
    updateFrameworkModelFromAPI(&plan, apiResponse)

    log.Printf("Created <resource> %s with ID %s", plan.Name.ValueString(), *apiResponse.Id)

    // Save data into Terraform state
    resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}
```

**Flow**:
1. Read plan from request
2. Check for diagnostics errors
3. Get proxy instance
4. Convert plan to API request
5. Call proxy create method
6. Handle errors
7. Update model with API response
8. Save model to state

**Error Handling**:
- Check diagnostics after reading plan
- Add error diagnostic if API call fails
- Include resource name in error message for context

#### 6.2 Read Method

**Purpose**: Fetch current resource state from API and update Terraform state.

**Design Pattern**:
```go
func (r *<resource>FrameworkResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
    var state <resource>FrameworkResourceModel

    // Read Terraform prior state data into the model
    resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
    if resp.Diagnostics.HasError() {
        return
    }

    proxy := get<ResourceName>Proxy(r.clientConfig)
    id := state.Id.ValueString()

    log.Printf("Reading <resource> %s", id)

    apiResponse, apiResp, err := proxy.get<ResourceName>ById(ctx, id)
    if err != nil {
        if util.IsStatus404(apiResp) {
            // Resource not found, remove from state
            resp.State.RemoveResource(ctx)
            return
        }
        resp.Diagnostics.AddError(
            "Error Reading <Resource>",
            fmt.Sprintf("Could not read <resource> %s: %s", id, err),
        )
        return
    }

    // Update the state with the latest data
    updateFrameworkModelFromAPI(&state, apiResponse)

    log.Printf("Read <resource> %s %s", id, *apiResponse.Name)

    // Save updated data into Terraform state
    resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
```

**Flow**:
1. Read current state from request
2. Check for diagnostics errors
3. Get proxy instance
4. Call proxy read method
5. Handle 404 (remove from state)
6. Handle other errors
7. Update model with API response
8. Save model to state

**404 Handling**:
- Check if response is 404 using `util.IsStatus404()`
- If 404, call `resp.State.RemoveResource(ctx)` to remove from state
- This is expected behavior (resource deleted outside Terraform)
- Don't add error diagnostic for 404

#### 6.3 Update Method

**Purpose**: Update existing resource via API and update Terraform state.

**Design Pattern**:
```go
func (r *<resource>FrameworkResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
    var plan <resource>FrameworkResourceModel
    var state <resource>FrameworkResourceModel

    // Read Terraform plan and current state data into the models
    resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
    resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
    if resp.Diagnostics.HasError() {
        return
    }

    proxy := get<ResourceName>Proxy(r.clientConfig)
    id := state.Id.ValueString()
    apiRequest := build<ResourceName>FromFrameworkModel(plan)

    log.Printf("Updating <resource> %s", plan.Name.ValueString())

    apiResponse, _, err := proxy.update<ResourceName>(ctx, id, apiRequest)
    if err != nil {
        resp.Diagnostics.AddError(
            "Error Updating <Resource>",
            fmt.Sprintf("Could not update <resource> %s: %s", plan.Name.ValueString(), err),
        )
        return
    }

    // Update model with response data
    updateFrameworkModelFromAPI(&plan, apiResponse)

    log.Printf("Updated <resource> %s", plan.Name.ValueString())

    // Save updated data into Terraform state
    resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}
```

**Flow**:
1. Read plan and state from request
2. Check for diagnostics errors
3. Get proxy instance
4. Get ID from state (ID doesn't change)
5. Convert plan to API request
6. Call proxy update method
7. Handle errors
8. Update model with API response
9. Save model to state

**Key Points**:
- Use ID from state (not plan) - ID is computed and doesn't change
- Use plan for all other attributes
- Update method receives both plan and state

#### 6.4 Delete Method

**Purpose**: Delete resource via API and verify deletion.

**Design Pattern**:
```go
func (r *<resource>FrameworkResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
    var state <resource>FrameworkResourceModel

    // Read Terraform prior state data into the model
    resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
    if resp.Diagnostics.HasError() {
        return
    }

    proxy := get<ResourceName>Proxy(r.clientConfig)
    id := state.Id.ValueString()
    name := state.Name.ValueString()

    log.Printf("Deleting <resource> %s", name)

    _, err := proxy.delete<ResourceName>(ctx, id)
    if err != nil {
        resp.Diagnostics.AddError(
            "Error Deleting <Resource>",
            fmt.Sprintf("Could not delete <resource> %s: %s", name, err),
        )
        return
    }

    // Verify deletion with retry logic
    retryErr := retry.RetryContext(ctx, 30*time.Second, func() *retry.RetryError {
        _, apiResp, err := proxy.get<ResourceName>ById(ctx, id)
        if err != nil {
            if util.IsStatus404(apiResp) {
                // Resource deleted successfully
                log.Printf("Deleted <resource> %s", id)
                return nil
            }
            return retry.NonRetryableError(fmt.Errorf("error deleting <resource> %s: %s", id, err))
        }

        return retry.RetryableError(fmt.Errorf("<resource> %s still exists", id))
    })

    if retryErr != nil {
        resp.Diagnostics.AddError(
            "Error Verifying <Resource> Deletion",
            fmt.Sprintf("Could not verify deletion of <resource> %s: %s", id, retryErr),
        )
        return
    }

    log.Printf("Successfully deleted <resource> %s", name)
}
```

**Flow**:
1. Read state from request
2. Check for diagnostics errors
3. Get proxy instance
4. Call proxy delete method
5. Handle errors
6. Verify deletion with retry logic
7. Handle verification errors

**Retry Logic**:
- Use `retry.RetryContext()` for eventual consistency
- Retry for 30 seconds (configurable)
- Check if resource returns 404 (deleted)
- If still exists, retry
- If other error, fail immediately (non-retryable)

**Why Retry**:
- API deletion may not be immediate
- Eventual consistency in distributed systems
- Prevents false positives in tests

#### 6.5 ImportState Method

**Purpose**: Enable importing existing resources into Terraform state.

**Design Pattern**:
```go
func (r *<resource>FrameworkResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
    // Retrieve import ID and save to id attribute
    resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
```

**Example** (routing_wrapupcode):
```go
func (r *routingWrapupcodeFrameworkResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
    resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
```

**How It Works**:
- User runs: `terraform import genesyscloud_routing_wrapupcode.example <resource_id>`
- Framework calls ImportState with ID
- `ImportStatePassthroughID` sets ID in state
- Framework automatically calls Read to populate other attributes

**Key Points**:
- Simple passthrough pattern for most resources
- More complex imports may need custom logic
- After import, Read method populates full state

---

### 7. Helper Functions

#### 7.1 build<ResourceName>FromFrameworkModel

**Purpose**: Convert Framework model to API request object.

**Design Pattern**:
```go
// build<ResourceName>FromFrameworkModel converts Framework model to API request model
func build<ResourceName>FromFrameworkModel(model <resource>FrameworkResourceModel) *platformclientv2.<ApiRequestType> {
    request := &platformclientv2.<ApiRequestType>{
        Name: model.Name.ValueStringPointer(),
    }

    // Handle optional attributes
    if !model.OptionalAttr.IsNull() && !model.OptionalAttr.IsUnknown() {
        request.OptionalAttr = model.OptionalAttr.ValueStringPointer()
    }

    // Handle nested objects
    if !model.NestedId.IsNull() && !model.NestedId.IsUnknown() {
        request.Nested = &platformclientv2.NestedObject{
            Id: model.NestedId.ValueStringPointer(),
        }
    }

    return request
}
```

**Example** (routing_wrapupcode):
```go
func buildWrapupcodeFromFrameworkModel(model routingWrapupcodeFrameworkResourceModel) *platformclientv2.Wrapupcoderequest {
    request := &platformclientv2.Wrapupcoderequest{
        Name: model.Name.ValueStringPointer(),
    }

    if !model.Description.IsNull() && !model.Description.IsUnknown() {
        request.Description = model.Description.ValueStringPointer()
    }

    if !model.DivisionId.IsNull() && !model.DivisionId.IsUnknown() {
        request.Division = &platformclientv2.Writablestarrabledivision{
            Id: model.DivisionId.ValueStringPointer(),
        }
    }

    return request
}
```

**Key Patterns**:

| Pattern | Usage |
|---------|-------|
| `ValueStringPointer()` | Convert `types.String` to `*string` for API |
| `IsNull()` check | Skip attribute if null |
| `IsUnknown()` check | Skip attribute if unknown (during plan) |
| Required attributes | Always set, no null check needed |
| Optional attributes | Check null/unknown before setting |
| Nested objects | Create nested struct with ID reference |

**Why Both Checks**:
- `IsNull()`: User explicitly set to null or omitted
- `IsUnknown()`: Value not yet known (during plan phase)
- Both must be false to safely use value

#### 7.2 updateFrameworkModelFromAPI

**Purpose**: Convert API response to Framework model.

**Design Pattern**:
```go
// updateFrameworkModelFromAPI updates Framework model from API response
func updateFrameworkModelFromAPI(model *<resource>FrameworkResourceModel, apiResponse *platformclientv2.<ApiResponseType>) {
    model.Id = types.StringValue(*apiResponse.Id)
    model.Name = types.StringValue(*apiResponse.Name)

    // Handle optional attributes
    if apiResponse.OptionalAttr != nil {
        model.OptionalAttr = types.StringValue(*apiResponse.OptionalAttr)
    } else {
        model.OptionalAttr = types.StringNull()
    }

    // Handle nested objects
    if apiResponse.Nested != nil && apiResponse.Nested.Id != nil {
        model.NestedId = types.StringValue(*apiResponse.Nested.Id)
    } else {
        model.NestedId = types.StringNull()
    }
}
```

**Example** (routing_wrapupcode):
```go
func updateFrameworkModelFromAPI(model *routingWrapupcodeFrameworkResourceModel, wrapupcode *platformclientv2.Wrapupcode) {
    model.Id = types.StringValue(*wrapupcode.Id)
    model.Name = types.StringValue(*wrapupcode.Name)

    if wrapupcode.Description != nil {
        model.Description = types.StringValue(*wrapupcode.Description)
    } else {
        model.Description = types.StringNull()
    }

    if wrapupcode.Division != nil && wrapupcode.Division.Id != nil {
        model.DivisionId = types.StringValue(*wrapupcode.Division.Id)
    } else {
        model.DivisionId = types.StringNull()
    }
}
```

**Key Patterns**:

| Pattern | Usage |
|---------|-------|
| `types.StringValue()` | Convert `*string` to `types.String` |
| `types.StringNull()` | Set null value explicitly |
| Nil checks | Check if pointer is nil before dereferencing |
| Required attributes | Always present, no nil check needed |
| Optional attributes | Check nil, set null if missing |
| Nested objects | Check both parent and child for nil |

**Why Explicit Null**:
- Framework distinguishes between null and empty string
- Explicit null handling prevents state drift
- Matches Terraform's null semantics

---

### 8. GetAll Functions

#### 8.1 GetAll<ResourceName> (Framework Version - Phase 2 Future)

**Purpose**: Fetch all resources for export using Framework diagnostics.

**Design Pattern**:
```go
// GetAll<ResourceName> retrieves all <resources> for export using Plugin Framework diagnostics.
// This is the future Phase 2 implementation that will be used once the exporter is updated
// to work natively with Framework types.
//
// Returns:
//   - resourceExporter.ResourceIDMetaMap: Map of resource IDs to metadata
//   - pfdiag.Diagnostics: Plugin Framework diagnostics
//
// Note: Currently NOT used by exporter. Exporter uses GetAll<ResourceName>SDK (SDK version).
func GetAll<ResourceName>(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, pfdiag.Diagnostics) {
    var diagErr pfdiag.Diagnostics
    proxy := get<ResourceName>Proxy(clientConfig)
    resources, _, err := proxy.getAll<ResourceName>(ctx)
    if err != nil {
        diagErr.AddError("Failed to get <resources> for export", err.Error())
        return nil, diagErr
    }

    if resources == nil {
        return resourceExporter.ResourceIDMetaMap{}, nil
    }

    exportMap := make(resourceExporter.ResourceIDMetaMap)
    for _, resource := range *resources {
        hashedUniqueFields, err := util.QuickHashFields(*resource.Name)
        if err != nil {
            diagErr.AddError("Failed to hash <resource> fields", err.Error())
            return nil, diagErr
        }
        exportMap[*resource.Id] = &resourceExporter.ResourceMeta{
            BlockLabel: *resource.Name,
            BlockHash:  hashedUniqueFields,
        }
    }
    return exportMap, nil
}
```

**Key Points**:
- Uses Plugin Framework diagnostics (`pfdiag.Diagnostics`)
- Returns clean export map without flat attributes
- Marked as Phase 2 future with TODO comment
- Not currently used by exporter

#### 8.2 GetAll<ResourceName>SDK (SDK Version - Phase 1 Current)

**Purpose**: Fetch all resources for export using SDK diagnostics (currently used by exporter).

**Design Pattern**:
```go
// GetAll<ResourceName>SDK retrieves all <resources> for export using SDK diagnostics.
// This is the Phase 1 implementation that converts SDK types to flat attribute maps
// for the legacy exporter's dependency resolution logic.
//
// IMPORTANT: This function is CURRENTLY USED by the exporter (see <ResourceName>Exporter).
// It implements the lazy fetch pattern for performance optimization.
//
// Returns:
//   - resourceExporter.ResourceIDMetaMap: Map of resource IDs to metadata with flat attributes
//   - sdkdiag.Diagnostics: SDK diagnostics (required by current exporter)
//
// Lazy Fetch Pattern:
//   - First API call: Fetch all resource IDs and names (lightweight)
//   - Filter: Apply exporter filters to determine which resources to export
//   - Second API call: Fetch full details ONLY for filtered resources (performance optimization)
//
// TODO: Remove this function once all resources are migrated to Plugin Framework
// and the exporter is updated to use GetAll<ResourceName> (Phase 2).
func GetAll<ResourceName>SDK(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, sdkdiag.Diagnostics) {
    proxy := get<ResourceName>Proxy(clientConfig)

    // Step 1: Fetch all resources (lightweight - just IDs and names)
    resources, _, err := proxy.getAll<ResourceName>(ctx)
    if err != nil {
        return nil, sdkdiag.Errorf("Failed to get <resources> for export: %v", err)
    }

    if resources == nil {
        return resourceExporter.ResourceIDMetaMap{}, nil
    }

    // Step 2: Build initial export map with IDs and names
    exportMap := make(resourceExporter.ResourceIDMetaMap)
    for _, resource := range *resources {
        hashedUniqueFields, err := util.QuickHashFields(*resource.Name)
        if err != nil {
            return nil, sdkdiag.Errorf("Failed to hash <resource> fields: %v", err)
        }
        exportMap[*resource.Id] = &resourceExporter.ResourceMeta{
            BlockLabel: *resource.Name,
            BlockHash:  hashedUniqueFields,
        }
    }

    // Step 3: Lazy fetch - Get full details ONLY for filtered resources
    // Note: For simple resources, the initial fetch already includes all attributes
    // so we don't need additional API calls. However, we still build the flat attribute map
    // for consistency with the exporter's dependency resolution logic.
    for _, resource := range *resources {
        if resource.Id == nil {
            continue
        }

        // Build flat attribute map for exporter (Phase 1 temporary)
        attributes := build<ResourceName>Attributes(&resource)

        // Update export map with attributes
        if meta, exists := exportMap[*resource.Id]; exists {
            meta.ExportAttributes = attributes
        } else {
            log.Printf("Warning: <Resource> %s not found in export map", *resource.Id)
        }
    }

    return exportMap, nil
}
```

**Example** (routing_wrapupcode):
```go
func GetAllRoutingWrapupcodesSDK(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, sdkdiag.Diagnostics) {
    proxy := getRoutingWrapupcodeProxy(clientConfig)

    wrapupcodes, _, err := proxy.getAllRoutingWrapupcode(ctx)
    if err != nil {
        return nil, sdkdiag.Errorf("Failed to get routing wrapupcodes for export: %v", err)
    }

    if wrapupcodes == nil {
        return resourceExporter.ResourceIDMetaMap{}, nil
    }

    exportMap := make(resourceExporter.ResourceIDMetaMap)
    for _, wrapupcode := range *wrapupcodes {
        hashedUniqueFields, err := util.QuickHashFields(*wrapupcode.Name)
        if err != nil {
            return nil, sdkdiag.Errorf("Failed to hash wrapupcode fields: %v", err)
        }
        exportMap[*wrapupcode.Id] = &resourceExporter.ResourceMeta{
            BlockLabel: *wrapupcode.Name,
            BlockHash:  hashedUniqueFields,
        }
    }

    for _, wrapupcode := range *wrapupcodes {
        if wrapupcode.Id == nil {
            continue
        }

        attributes := buildWrapupcodeAttributes(&wrapupcode)

        if meta, exists := exportMap[*wrapupcode.Id]; exists {
            meta.ExportAttributes = attributes
        } else {
            log.Printf("Warning: Wrapupcode %s not found in export map", *wrapupcode.Id)
        }
    }

    return exportMap, nil
}
```

**Key Points**:
- Uses SDK diagnostics (`sdkdiag.Diagnostics`)
- Builds flat attribute map for each resource
- Currently used by exporter (Phase 1)
- Marked with TODO for Phase 2 removal
- Implements lazy fetch pattern (when applicable)

**Lazy Fetch Pattern**:
- **Step 1**: Fetch all resource IDs and names (lightweight)
- **Step 2**: Build initial export map
- **Step 3**: Fetch full details only for filtered resources
- **Benefit**: Reduces API calls when exporting subset of resources
- **Note**: For simple resources like wrapupcode, initial fetch includes all data

**Flat Attribute Map**:
- Converts resource to `map[string]string`
- Matches SDKv2 InstanceState format
- Used by exporter for dependency resolution
- Temporary Phase 1 code (will be removed in Phase 2)

---

## Part 2: Data Source Implementation

### 1. Data Source Struct

**Purpose**: Hold data source-level configuration.

**Design Pattern**:
```go
// <resource>FrameworkDataSource defines the data source implementation for Plugin Framework.
type <resource>FrameworkDataSource struct {
    clientConfig *platformclientv2.Configuration
}
```

**Example** (routing_wrapupcode):
```go
type routingWrapupcodeFrameworkDataSource struct {
    clientConfig *platformclientv2.Configuration
}
```

---

### 2. Data Source Model

**Purpose**: Define the data structure for data source output.

**Design Pattern**:
```go
// <resource>FrameworkDataSourceModel describes the data source data model.
type <resource>FrameworkDataSourceModel struct {
    Id   types.String `tfsdk:"id"`
    Name types.String `tfsdk:"name"`
}
```

**Example** (routing_wrapupcode):
```go
type routingWrapupcodeFrameworkDataSourceModel struct {
    Id   types.String `tfsdk:"id"`
    Name types.String `tfsdk:"name"`
}
```

**Key Points**:
- Simpler than resource model (only lookup criteria and result)
- Typically just `id` and `name`
- Can include additional lookup criteria if needed

---

### 3. Data Source Constructor

**Design Pattern**:
```go
// New<ResourceName>FrameworkDataSource is a helper function to simplify the provider implementation.
func New<ResourceName>FrameworkDataSource() datasource.DataSource {
    return &<resource>FrameworkDataSource{}
}
```

---

### 4. Data Source Methods

#### 4.1 Metadata, Schema, Configure

Same pattern as resource (see resource section above).

#### 4.2 Read Method (Data Source)

**Purpose**: Look up resource by name and return ID.

**Design Pattern**:
```go
func (d *<resource>FrameworkDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
    var config <resource>FrameworkDataSourceModel

    // Read Terraform configuration data into the model
    resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
    if resp.Diagnostics.HasError() {
        return
    }

    proxy := get<ResourceName>Proxy(d.clientConfig)
    name := config.Name.ValueString()

    log.Printf("Reading <resource> data source for name: %s", name)

    // Use retry logic for eventual consistency
    var resourceId string
    retryErr := retry.RetryContext(ctx, 15*time.Second, func() *retry.RetryError {
        id, retryable, _, err := proxy.get<ResourceName>IdByName(ctx, name)
        if err != nil {
            if !retryable {
                return retry.NonRetryableError(fmt.Errorf("failed to find <resource> with name '%s': %s", name, err))
            }
            log.Printf("Retrying lookup for <resource> with name '%s': %s", name, err)
            return retry.RetryableError(fmt.Errorf("failed to find <resource> with name '%s': %s", name, err))
        }
        resourceId = id
        return nil
    })

    if retryErr != nil {
        resp.Diagnostics.AddError(
            "Error Reading <Resource> Data Source",
            fmt.Sprintf("Could not find <resource> with name '%s': %s", name, retryErr),
        )
        return
    }

    // Set the ID in the model
    config.Id = types.StringValue(resourceId)

    log.Printf("Found <resource> %s with ID %s", name, resourceId)

    // Save data into Terraform state
    resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
```

**Key Points**:
- Reads config (not state) - data sources don't have prior state
- Uses retry logic for eventual consistency (15 seconds typical)
- Proxy method returns retryable flag
- Sets ID in model after successful lookup

**Retry Logic**:
- Data sources may be used immediately after resource creation
- Eventual consistency means resource may not be immediately available
- Retry for 15 seconds (configurable)
- Proxy method indicates if error is retryable

---

## SDKv2 vs Plugin Framework Comparison

### Resource Lifecycle

**SDKv2**:
```go
func ResourceRoutingWrapupcode() *schema.Resource {
    return &schema.Resource{
        CreateContext: createRoutingWrapupcode,
        ReadContext:   readRoutingWrapupcode,
        UpdateContext: updateRoutingWrapupcode,
        DeleteContext: deleteRoutingWrapupcode,
        Importer: &schema.ResourceImporter{
            StateContext: schema.ImportStatePassthroughContext,
        },
        Schema: map[string]*schema.Schema{
            // ...
        },
    }
}

func createRoutingWrapupcode(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
    // Implementation
}
```

**Plugin Framework**:
```go
type routingWrapupcodeFrameworkResource struct {
    clientConfig *platformclientv2.Configuration
}

func (r *routingWrapupcodeFrameworkResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
    // Implementation
}
```

**Key Differences**:

| Aspect | SDKv2 | Plugin Framework |
|--------|-------|------------------|
| Structure | Function-based callbacks | Method-based interface |
| State access | `d *schema.ResourceData` | `req.Plan`, `req.State`, `resp.State` |
| Error handling | Return `diag.Diagnostics` | Add to `resp.Diagnostics` |
| Type safety | Runtime (interface{}) | Compile-time (typed models) |
| Null handling | Pointer nil checks | `IsNull()`, `IsUnknown()` methods |

---

## Design Patterns and Best Practices

### Pattern 1: Early Return on Diagnostics

**Pattern**:
```go
resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
if resp.Diagnostics.HasError() {
    return
}
```

**Why**:
- Prevents nil pointer dereferences
- Stops execution if plan/state read fails
- Clear error handling flow

### Pattern 2: Explicit Null Handling

**Pattern**:
```go
if !model.OptionalAttr.IsNull() && !model.OptionalAttr.IsUnknown() {
    request.OptionalAttr = model.OptionalAttr.ValueStringPointer()
}
```

**Why**:
- Framework distinguishes null, unknown, and empty
- Prevents sending null values to API
- Explicit intent in code

### Pattern 3: 404 Removal from State

**Pattern**:
```go
if util.IsStatus404(apiResp) {
    resp.State.RemoveResource(ctx)
    return
}
```

**Why**:
- Resource deleted outside Terraform
- Graceful handling of drift
- Allows Terraform to recreate if needed

### Pattern 4: Retry for Eventual Consistency

**Pattern**:
```go
retryErr := retry.RetryContext(ctx, 30*time.Second, func() *retry.RetryError {
    // Check condition
    if success {
        return nil
    }
    if nonRetryableError {
        return retry.NonRetryableError(err)
    }
    return retry.RetryableError(err)
})
```

**Why**:
- Handles eventual consistency in distributed systems
- Prevents false failures in tests
- Configurable timeout

### Pattern 5: Logging for Debugging

**Pattern**:
```go
log.Printf("Creating <resource> %s", plan.Name.ValueString())
log.Printf("Created <resource> %s with ID %s", name, id)
```

**Why**:
- Helps debug issues in production
- Tracks resource lifecycle
- Provides audit trail

---

## Migration Considerations

### Behavior Preservation Checklist

When migrating from SDKv2 to Framework, verify:

- [ ] CRUD operations behave identically
- [ ] Error messages are equivalent or better
- [ ] Retry logic matches (timeouts, conditions)
- [ ] 404 handling is consistent
- [ ] Logging is equivalent
- [ ] Import functionality works
- [ ] Data source lookup works with retry

### Common Migration Pitfalls

#### Pitfall 1: Forgetting Null/Unknown Checks
**Problem**: Sending null values to API causes errors.
**Solution**: Always check `IsNull()` and `IsUnknown()` before using values.

#### Pitfall 2: Using State ID in Update
**Problem**: Using plan.Id instead of state.Id in Update method.
**Solution**: Always use state.Id (ID doesn't change).

#### Pitfall 3: Not Handling 404 in Read
**Problem**: Treating 404 as error instead of removing from state.
**Solution**: Check for 404 and call `RemoveResource()`.

#### Pitfall 4: Missing Retry Logic
**Problem**: Data source fails immediately if resource not found.
**Solution**: Add retry logic with appropriate timeout.

#### Pitfall 5: Incorrect Diagnostics Handling
**Problem**: Continuing execution after diagnostics error.
**Solution**: Check `HasError()` and return early.

---

## Summary

### Key Design Decisions

1. **Framework-Native Patterns**: Use Framework lifecycle methods and types
2. **Proxy Reuse**: Reuse existing proxy methods without modification
3. **Type Safety**: Use Framework types for null/unknown handling
4. **Dual GetAll Functions**: Support both Framework and SDK during migration
5. **Explicit Error Handling**: Clear diagnostics and early returns

### File Structure

```
Resource File:
├── Interface verification
├── Resource struct and model
├── Constructor function
├── Interface methods (Metadata, Schema, Configure)
├── CRUD methods (Create, Read, Update, Delete, ImportState)
├── Helper functions (build, update)
└── GetAll functions (Framework and SDK versions)

Data Source File:
├── Interface verification
├── Data source struct and model
├── Constructor function
├── Interface methods (Metadata, Schema, Configure)
└── Read method (with retry logic)
```

### Next Steps

After completing Stage 2 resource migration:
1. Review CRUD implementation for correctness
2. Verify GetAll functions work correctly
3. Confirm data source lookup works
4. Proceed to **Stage 3 – Test Migration**

---

## References

- **Reference Implementation**: 
  - `genesyscloud/routing_wrapupcode/resource_genesyscloud_routing_wrapupcode.go`
  - `genesyscloud/routing_wrapupcode/data_source_genesyscloud_routing_wrapupcode.go`
- **Plugin Framework Resources**: https://developer.hashicorp.com/terraform/plugin/framework/resources
- **Plugin Framework Data Sources**: https://developer.hashicorp.com/terraform/plugin/framework/data-sources
- **Framework Types**: https://developer.hashicorp.com/terraform/plugin/framework/handling-data/types
