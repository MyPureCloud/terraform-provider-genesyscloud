# Stage 2 – Resource Migration Requirements

## Overview

Stage 2 focuses on migrating resource implementation from Terraform Plugin SDKv2 to the Terraform Plugin Framework. This stage implements the CRUD (Create, Read, Update, Delete) lifecycle operations and data source read logic using the schema definitions created in Stage 1.

**Reference Implementation**: 
- `genesyscloud/routing_wrapupcode/resource_genesyscloud_routing_wrapupcode.go`
- `genesyscloud/routing_wrapupcode/data_source_genesyscloud_routing_wrapupcode.go`

---

## Objectives

### Primary Goal
Implement Plugin Framework resource and data source logic while reusing existing business logic from proxy methods without modification.

### Specific Objectives
1. Create Framework resource struct and model
2. Implement CRUD lifecycle methods using Framework patterns
3. Create Framework data source struct and model
4. Implement data source read logic with retry patterns
5. Implement GetAll functions for export support (both Framework and SDK versions)
6. Create helper functions for model conversion
7. Ensure backward compatibility with existing behavior

---

## Scope

### In Scope for Stage 2

#### 1. Resource Implementation File
- Create `resource_genesyscloud_<resource_name>.go` file
- Implement Framework resource struct
- Implement Framework resource model
- Implement all CRUD methods
- Implement ImportState method
- Implement helper functions for model conversion

#### 2. CRUD Operations
- **Create**: Convert plan to API request, call proxy, update state
- **Read**: Fetch from API via proxy, update state, handle 404
- **Update**: Convert plan to API request, call proxy, update state
- **Delete**: Call proxy delete, verify deletion with retry
- **ImportState**: Enable resource import by ID

#### 3. Data Source Implementation File
- Create `data_source_genesyscloud_<resource_name>.go` file
- Implement Framework data source struct
- Implement Framework data source model
- Implement Read method with retry logic for eventual consistency

#### 4. GetAll Functions
- **GetAll<ResourceName>**: Framework version with Plugin Framework diagnostics (Phase 2 future)
- **GetAll<ResourceName>SDK**: SDK version with SDK diagnostics (Phase 1 current, used by exporter)
- Implement lazy fetch pattern for performance optimization

#### 5. Helper Functions
- `build<ResourceName>FromFrameworkModel()`: Convert Framework model to API request
- `updateFrameworkModelFromAPI()`: Convert API response to Framework model
- `buildWrapupcodeAttributes()`: Build flat attribute map for export (temporary Phase 1)

### Out of Scope for Stage 2

❌ **Test Files**
- No test implementation
- No test cases
- No test helper functions
- Tests are covered in Stage 3

❌ **Export Utilities File**
- No separate export utilities file
- Export attribute building is inline in resource file (temporary)
- Separate export utilities file is covered in Stage 4

❌ **Schema Modifications**
- No changes to schema file created in Stage 1
- Schema is already complete

❌ **Proxy Modifications**
- No changes to `genesyscloud_<resource>_proxy.go`
- Proxy files remain unchanged throughout migration

---

## Success Criteria

### Functional Requirements

#### FR1: Resource Struct Implementation
- ✅ Framework resource struct implements required interfaces:
  - `resource.Resource`
  - `resource.ResourceWithConfigure`
  - `resource.ResourceWithImportState`
- ✅ Resource struct contains `clientConfig` field
- ✅ Constructor function `New<ResourceName>FrameworkResource()` is implemented

#### FR2: Resource Model Implementation
- ✅ Framework resource model struct is defined
- ✅ Model uses Framework types (`types.String`, `types.Int64`, etc.)
- ✅ Model fields have `tfsdk` struct tags matching schema attribute names
- ✅ Model includes all attributes from schema

#### FR3: CRUD Operations
- ✅ **Create**: Creates resource via proxy, updates state with response
- ✅ **Read**: Fetches resource via proxy, handles 404 gracefully
- ✅ **Update**: Updates resource via proxy, updates state with response
- ✅ **Delete**: Deletes resource via proxy, verifies deletion with retry
- ✅ All operations use context for cancellation
- ✅ All operations handle errors with clear diagnostics

#### FR4: ImportState
- ✅ ImportState method is implemented
- ✅ Uses `resource.ImportStatePassthroughID` pattern
- ✅ Imports by resource ID

#### FR5: Data Source Implementation
- ✅ Framework data source struct implements required interfaces:
  - `datasource.DataSource`
  - `datasource.DataSourceWithConfigure`
- ✅ Data source model struct is defined
- ✅ Read method implements lookup by name with retry logic
- ✅ Constructor function `New<ResourceName>FrameworkDataSource()` is implemented

#### FR6: GetAll Functions
- ✅ `GetAll<ResourceName>()` implemented with Framework diagnostics (Phase 2 future)
- ✅ `GetAll<ResourceName>SDK()` implemented with SDK diagnostics (Phase 1 current)
- ✅ Both functions return `resourceExporter.ResourceIDMetaMap`
- ✅ SDK version includes flat attribute map for export
- ✅ Lazy fetch pattern implemented for performance

#### FR7: Helper Functions
- ✅ `build<ResourceName>FromFrameworkModel()` converts model to API request
- ✅ `updateFrameworkModelFromAPI()` converts API response to model
- ✅ Helper functions handle null/unknown values correctly
- ✅ Helper functions use pointer methods appropriately

#### FR8: Behavior Preservation
- ✅ All CRUD operations behave identically to SDKv2 version
- ✅ Error handling matches SDKv2 patterns
- ✅ Retry logic matches SDKv2 patterns
- ✅ API calls use same proxy methods as SDKv2

### Non-Functional Requirements

#### NFR1: Code Quality
- ✅ Code follows Go best practices
- ✅ Code follows existing codebase conventions
- ✅ Proper error handling with clear messages
- ✅ Logging at appropriate levels
- ✅ No unused imports or variables

#### NFR2: Documentation
- ✅ All functions have clear comments
- ✅ Complex logic is explained with inline comments
- ✅ Phase 1/Phase 2 temporary code is marked with TODO comments
- ✅ Export-related code includes migration notes

#### NFR3: Type Safety
- ✅ Use Framework types (`types.String`) instead of pointers
- ✅ Proper null/unknown value handling
- ✅ Type conversions are explicit and safe

#### NFR4: Performance
- ✅ Lazy fetch pattern for GetAll functions
- ✅ Efficient API calls (no unnecessary requests)
- ✅ Proper use of context for cancellation

---

## Dependencies and Prerequisites

### Prerequisites

#### 1. Stage 1 Completion
- Schema file must be complete and approved
- `SetRegistrar()` function references constructor functions created in Stage 2
- Exporter configuration references GetAll functions created in Stage 2

#### 2. Understanding of Framework Patterns
- Familiarity with Framework resource lifecycle
- Understanding of Framework types and diagnostics
- Knowledge of context usage in Framework

#### 3. Reference Implementation
- Study `routing_wrapupcode` resource implementation
- Understand CRUD patterns used
- Review helper function implementations

### Dependencies

#### 1. Package Imports (Resource File)
```go
import (
    "context"
    "fmt"
    "log"
    "time"

    pfdiag "github.com/hashicorp/terraform-plugin-framework/diag"
    "github.com/hashicorp/terraform-plugin-framework/path"
    "github.com/hashicorp/terraform-plugin-framework/resource"
    "github.com/hashicorp/terraform-plugin-framework/types"
    sdkdiag "github.com/hashicorp/terraform-plugin-sdk/v2/diag"
    "github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
    "github.com/mypurecloud/platform-client-sdk-go/v176/platformclientv2"
    "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
    resourceExporter "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_exporter"
    "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
)
```

#### 2. Package Imports (Data Source File)
```go
import (
    "context"
    "fmt"
    "log"
    "time"

    "github.com/hashicorp/terraform-plugin-framework/datasource"
    "github.com/hashicorp/terraform-plugin-framework/types"
    "github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
    "github.com/mypurecloud/platform-client-sdk-go/v176/platformclientv2"
    "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
)
```

#### 3. Proxy Package
- Proxy package must exist with all required methods
- Proxy methods are NOT modified during migration
- Proxy provides: create, read, update, delete, getAll, getByName methods

#### 4. Utility Functions
- `util.IsStatus404()` for 404 detection
- `util.QuickHashFields()` for export hash calculation
- `retry.RetryContext()` for retry logic

---

## Constraints

### Technical Constraints

#### TC1: No Proxy Modifications
- **Constraint**: `genesyscloud_<resource>_proxy.go` files MUST NOT be modified
- **Rationale**: Proxy files contain generic shared implementations
- **Impact**: All API interactions must use existing proxy methods

#### TC2: No Business Logic Changes
- **Constraint**: Resource migration MUST NOT change existing business logic
- **Rationale**: Migration is a framework translation only
- **Impact**: CRUD behavior must remain identical to SDKv2

#### TC3: Reuse Existing Proxy Methods
- **Constraint**: Must use existing proxy methods without modification
- **Rationale**: Proxy methods are shared between SDKv2 and Framework
- **Impact**: Cannot add new proxy methods or change signatures

#### TC4: Framework Type Usage
- **Constraint**: Must use Framework types (`types.String`) not pointers
- **Rationale**: Framework requires specific types for state management
- **Impact**: Model structs use Framework types, conversion needed for API calls

### Process Constraints

#### PC1: Stage Isolation
- **Constraint**: Stage 2 MUST NOT include test implementation
- **Rationale**: Clear separation of concerns for easier review
- **Impact**: Tests are deferred to Stage 3

#### PC2: Export Compatibility
- **Constraint**: GetAll functions must support both Framework and SDK patterns
- **Rationale**: Exporter currently uses SDK diagnostics
- **Impact**: Two GetAll functions needed (Framework and SDK versions)

---

## Validation Checklist

Use this checklist to verify Stage 2 completion:

### Resource Implementation File
- [ ] File created: `resource_genesyscloud_<resource_name>.go`
- [ ] Package declaration matches directory name
- [ ] All required imports are present
- [ ] No unused imports

### Resource Struct and Model
- [ ] Resource struct implements required interfaces
- [ ] Resource struct has `clientConfig` field
- [ ] Resource model struct defined with all attributes
- [ ] Model uses Framework types (`types.String`, etc.)
- [ ] Model has correct `tfsdk` struct tags
- [ ] Constructor function implemented

### CRUD Methods
- [ ] Metadata() method implemented
- [ ] Schema() method implemented (calls schema function from Stage 1)
- [ ] Configure() method implemented
- [ ] Create() method implemented
- [ ] Read() method implemented
- [ ] Update() method implemented
- [ ] Delete() method implemented
- [ ] ImportState() method implemented

### CRUD Method Behavior
- [ ] Create: Reads plan, calls proxy, updates state
- [ ] Read: Calls proxy, handles 404, updates state
- [ ] Update: Reads plan and state, calls proxy, updates state
- [ ] Delete: Calls proxy, verifies deletion with retry
- [ ] All methods handle errors with clear diagnostics
- [ ] All methods use context appropriately

### Data Source Implementation File
- [ ] File created: `data_source_genesyscloud_<resource_name>.go`
- [ ] Package declaration matches directory name
- [ ] All required imports are present

### Data Source Struct and Model
- [ ] Data source struct implements required interfaces
- [ ] Data source struct has `clientConfig` field
- [ ] Data source model struct defined
- [ ] Model uses Framework types
- [ ] Model has correct `tfsdk` struct tags
- [ ] Constructor function implemented

### Data Source Methods
- [ ] Metadata() method implemented
- [ ] Schema() method implemented (calls schema function from Stage 1)
- [ ] Configure() method implemented
- [ ] Read() method implemented with retry logic

### GetAll Functions
- [ ] `GetAll<ResourceName>()` implemented (Framework version)
- [ ] `GetAll<ResourceName>SDK()` implemented (SDK version)
- [ ] Both return `resourceExporter.ResourceIDMetaMap`
- [ ] SDK version includes flat attribute map
- [ ] Functions include Phase 1/Phase 2 comments
- [ ] Lazy fetch pattern implemented (if applicable)

### Helper Functions
- [ ] `build<ResourceName>FromFrameworkModel()` implemented
- [ ] `updateFrameworkModelFromAPI()` implemented
- [ ] Helper functions handle null/unknown values
- [ ] Helper functions use pointers appropriately

### Code Quality
- [ ] Code compiles without errors
- [ ] Code follows Go conventions
- [ ] Functions have clear comments
- [ ] Error messages are clear and actionable
- [ ] Logging is appropriate

---

## Example: routing_wrapupcode Resource Migration

### File Structure
```
genesyscloud/routing_wrapupcode/
├── resource_genesyscloud_routing_wrapupcode_schema.go  (Stage 1)
├── resource_genesyscloud_routing_wrapupcode.go         (Stage 2 - THIS)
└── data_source_genesyscloud_routing_wrapupcode.go      (Stage 2 - THIS)
```

### Key Components

#### 1. Resource Struct
```go
type routingWrapupcodeFrameworkResource struct {
    clientConfig *platformclientv2.Configuration
}
```

#### 2. Resource Model
```go
type routingWrapupcodeFrameworkResourceModel struct {
    Id          types.String `tfsdk:"id"`
    Name        types.String `tfsdk:"name"`
    DivisionId  types.String `tfsdk:"division_id"`
    Description types.String `tfsdk:"description"`
}
```

#### 3. Constructor Function
```go
func NewRoutingWrapupcodeFrameworkResource() resource.Resource {
    return &routingWrapupcodeFrameworkResource{}
}
```

#### 4. Create Method Pattern
```go
func (r *routingWrapupcodeFrameworkResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
    var plan routingWrapupcodeFrameworkResourceModel
    resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
    if resp.Diagnostics.HasError() {
        return
    }

    proxy := getRoutingWrapupcodeProxy(r.clientConfig)
    wrapupcodeRequest := buildWrapupcodeFromFrameworkModel(plan)

    wrapupcode, _, err := proxy.createRoutingWrapupcode(ctx, wrapupcodeRequest)
    if err != nil {
        resp.Diagnostics.AddError("Error Creating Resource", err.Error())
        return
    }

    updateFrameworkModelFromAPI(&plan, wrapupcode)
    resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}
```

#### 5. Helper Function Pattern
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

---

## Next Steps

After Stage 2 completion and approval:
1. Review resource implementation with team
2. Verify CRUD operations work correctly
3. Confirm GetAll functions are correct
4. Proceed to **Stage 3 – Test Migration**

---

## References

- **Reference Implementation**: 
  - `genesyscloud/routing_wrapupcode/resource_genesyscloud_routing_wrapupcode.go`
  - `genesyscloud/routing_wrapupcode/data_source_genesyscloud_routing_wrapupcode.go`
- **Plugin Framework Resources**: https://developer.hashicorp.com/terraform/plugin/framework/resources
- **Plugin Framework Data Sources**: https://developer.hashicorp.com/terraform/plugin/framework/data-sources
