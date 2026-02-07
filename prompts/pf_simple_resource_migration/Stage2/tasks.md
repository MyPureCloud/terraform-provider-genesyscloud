# Stage 2 – Resource Migration Tasks

## Overview

This document provides step-by-step tasks for completing Stage 2 of the Plugin Framework migration. Follow these tasks in order to migrate resource CRUD operations and data source logic from SDKv2 to Plugin Framework.

**Reference Implementation**: 
- `genesyscloud/routing_wrapupcode/resource_genesyscloud_routing_wrapupcode.go`
- `genesyscloud/routing_wrapupcode/data_source_genesyscloud_routing_wrapupcode.go`

**Estimated Time**: 6-10 hours (depending on resource complexity)

---

## Prerequisites

Before starting Stage 2 tasks, ensure:

- [ ] Stage 1 (Schema Migration) is complete and approved
- [ ] You have reviewed the existing SDKv2 resource CRUD implementation
- [ ] You understand the proxy methods available
- [ ] You have read Stage 2 `requirements.md` and `design.md`
- [ ] You have studied the `routing_wrapupcode` reference implementation
- [ ] Development environment is set up and ready

---

## Task Checklist

### Phase 1: Resource File Setup
- [ ] Task 1.1: Create Resource Implementation File
- [ ] Task 1.2: Add Package Declaration and Imports
- [ ] Task 1.3: Add Interface Verification
- [ ] Task 1.4: Define Resource Struct and Model
- [ ] Task 1.5: Implement Constructor Function

### Phase 2: Resource Interface Methods
- [ ] Task 2.1: Implement Metadata Method
- [ ] Task 2.2: Implement Schema Method
- [ ] Task 2.3: Implement Configure Method

### Phase 3: CRUD Operations
- [ ] Task 3.1: Implement Create Method
- [ ] Task 3.2: Implement Read Method
- [ ] Task 3.3: Implement Update Method
- [ ] Task 3.4: Implement Delete Method
- [ ] Task 3.5: Implement ImportState Method

### Phase 4: Helper Functions
- [ ] Task 4.1: Implement build<ResourceName>FromFrameworkModel
- [ ] Task 4.2: Implement updateFrameworkModelFromAPI

### Phase 5: GetAll Functions
- [ ] Task 5.1: Implement GetAll<ResourceName> (Framework Version)
- [ ] Task 5.2: Implement GetAll<ResourceName>SDK (SDK Version)

### Phase 6: Data Source Implementation
- [ ] Task 6.1: Create Data Source File
- [ ] Task 6.2: Implement Data Source Struct and Model
- [ ] Task 6.3: Implement Data Source Methods
- [ ] Task 6.4: Implement Data Source Read with Retry

### Phase 7: Validation and Review
- [ ] Task 7.1: Compile and Verify
- [ ] Task 7.2: Review Against Checklist
- [ ] Task 7.3: Code Review and Approval

---

## Detailed Tasks

## Phase 1: Resource File Setup

### Task 1.1: Create Resource Implementation File

**Objective**: Create the new resource implementation file.

**Steps**:

1. **Navigate to resource package directory**
   ```powershell
   cd genesyscloud\<resource_name>
   ```
   Example:
   ```powershell
   cd genesyscloud\routing_wrapupcode
   ```

2. **Create the resource file**
   ```powershell
   New-Item -ItemType File -Name "resource_genesyscloud_<resource_name>.go"
   ```
   Example:
   ```powershell
   New-Item -ItemType File -Name "resource_genesyscloud_routing_wrapupcode.go"
   ```

**Deliverable**: Empty resource file created in correct location

---

### Task 1.2: Add Package Declaration and Imports

**Objective**: Set up the file with correct package and imports.

**Steps**:

1. **Add package declaration**
   ```go
   package <resource_name>
   ```
   Example:
   ```go
   package routing_wrapupcode
   ```

2. **Add required imports**
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

3. **Adjust imports based on needs**
   - Add additional imports if using nested attributes
   - Add imports for other packages if needed for dependencies

**Deliverable**: File with package declaration and imports

---

### Task 1.3: Add Interface Verification

**Objective**: Add compile-time interface verification.

**Steps**:

1. **Add interface verification comment and variables**
   ```go
   // Ensure <resource>FrameworkResource satisfies various resource interfaces.
   var (
       _ resource.Resource                = &<resource>FrameworkResource{}
       _ resource.ResourceWithConfigure   = &<resource>FrameworkResource{}
       _ resource.ResourceWithImportState = &<resource>FrameworkResource{}
   )
   ```

2. **Replace placeholders**
   - `<resource>` → Your resource name in camelCase
   - Example: `routingWrapupcode`

**Example** (routing_wrapupcode):
```go
var (
    _ resource.Resource                = &routingWrapupcodeFrameworkResource{}
    _ resource.ResourceWithConfigure   = &routingWrapupcodeFrameworkResource{}
    _ resource.ResourceWithImportState = &routingWrapupcodeFrameworkResource{}
)
```

**Deliverable**: Interface verification added

---

### Task 1.4: Define Resource Struct and Model

**Objective**: Define the resource struct and model for state management.

**Steps**:

1. **Define resource struct**
   ```go
   // <resource>FrameworkResource defines the resource implementation for Plugin Framework.
   type <resource>FrameworkResource struct {
       clientConfig *platformclientv2.Configuration
   }
   ```

2. **Define resource model struct**
   ```go
   // <resource>FrameworkResourceModel describes the resource data model.
   type <resource>FrameworkResourceModel struct {
       Id          types.String `tfsdk:"id"`
       Name        types.String `tfsdk:"name"`
       // Add all other attributes from schema
   }
   ```

3. **Map all schema attributes to model fields**
   - Use Framework types (`types.String`, `types.Int64`, `types.Bool`, etc.)
   - Use `tfsdk` struct tags matching schema attribute names exactly
   - Include all attributes from Stage 1 schema

**Example** (routing_wrapupcode):
```go
type routingWrapupcodeFrameworkResource struct {
    clientConfig *platformclientv2.Configuration
}

type routingWrapupcodeFrameworkResourceModel struct {
    Id          types.String `tfsdk:"id"`
    Name        types.String `tfsdk:"name"`
    DivisionId  types.String `tfsdk:"division_id"`
    Description types.String `tfsdk:"description"`
}
```

**Deliverable**: Resource struct and model defined

---

### Task 1.5: Implement Constructor Function

**Objective**: Create the constructor function for the resource.

**Steps**:

1. **Add constructor function**
   ```go
   // New<ResourceName>FrameworkResource is a helper function to simplify the provider implementation.
   func New<ResourceName>FrameworkResource() resource.Resource {
       return &<resource>FrameworkResource{}
   }
   ```

2. **Replace placeholders**
   - `<ResourceName>` → Your resource name in PascalCase
   - `<resource>` → Your resource name in camelCase

**Example** (routing_wrapupcode):
```go
func NewRoutingWrapupcodeFrameworkResource() resource.Resource {
    return &routingWrapupcodeFrameworkResource{}
}
```

**Deliverable**: Constructor function implemented

---

## Phase 2: Resource Interface Methods

### Task 2.1: Implement Metadata Method

**Objective**: Provide resource type name to the provider.

**Steps**:

1. **Add Metadata method**
   ```go
   func (r *<resource>FrameworkResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
       resp.TypeName = req.ProviderTypeName + "_<resource_name>"
   }
   ```

2. **Verify type name matches ResourceType constant from Stage 1**

**Example** (routing_wrapupcode):
```go
func (r *routingWrapupcodeFrameworkResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
    resp.TypeName = req.ProviderTypeName + "_routing_wrapupcode"
}
```

**Deliverable**: Metadata method implemented

---

### Task 2.2: Implement Schema Method

**Objective**: Provide resource schema to the provider.

**Steps**:

1. **Add Schema method**
   ```go
   func (r *<resource>FrameworkResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
       resp.Schema = <ResourceName>ResourceSchema()
   }
   ```

2. **Verify function name matches schema function from Stage 1**

**Example** (routing_wrapupcode):
```go
func (r *routingWrapupcodeFrameworkResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
    resp.Schema = RoutingWrapupcodeResourceSchema()
}
```

**Deliverable**: Schema method implemented

---

### Task 2.3: Implement Configure Method

**Objective**: Receive and store provider configuration.

**Steps**:

1. **Add Configure method**
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

**Deliverable**: Configure method implemented

---

## Phase 3: CRUD Operations

### Task 3.1: Implement Create Method

**Objective**: Implement resource creation logic.

**Steps**:

1. **Add Create method signature**
   ```go
   func (r *<resource>FrameworkResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
   ```

2. **Read plan from request**
   ```go
   var plan <resource>FrameworkResourceModel

   // Read Terraform plan data into the model
   resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
   if resp.Diagnostics.HasError() {
       return
   }
   ```

3. **Get proxy and build API request**
   ```go
   proxy := get<ResourceName>Proxy(r.clientConfig)
   apiRequest := build<ResourceName>FromFrameworkModel(plan)

   log.Printf("Creating <resource> %s", plan.Name.ValueString())
   ```

4. **Call proxy create method**
   ```go
   apiResponse, _, err := proxy.create<ResourceName>(ctx, apiRequest)
   if err != nil {
       resp.Diagnostics.AddError(
           "Error Creating <Resource>",
           fmt.Sprintf("Could not create <resource> %s: %s", plan.Name.ValueString(), err),
       )
       return
   }
   ```

5. **Update model and save to state**
   ```go
   // Update model with response data
   updateFrameworkModelFromAPI(&plan, apiResponse)

   log.Printf("Created <resource> %s with ID %s", plan.Name.ValueString(), *apiResponse.Id)

   // Save data into Terraform state
   resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
   }
   ```

**Example** (routing_wrapupcode - see reference file for complete implementation)

**Deliverable**: Create method implemented

---

### Task 3.2: Implement Read Method

**Objective**: Implement resource read logic with 404 handling.

**Steps**:

1. **Add Read method signature**
   ```go
   func (r *<resource>FrameworkResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
   ```

2. **Read state from request**
   ```go
   var state <resource>FrameworkResourceModel

   // Read Terraform prior state data into the model
   resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
   if resp.Diagnostics.HasError() {
       return
   }
   ```

3. **Get proxy and call read method**
   ```go
   proxy := get<ResourceName>Proxy(r.clientConfig)
   id := state.Id.ValueString()

   log.Printf("Reading <resource> %s", id)

   apiResponse, apiResp, err := proxy.get<ResourceName>ById(ctx, id)
   ```

4. **Handle 404 (resource deleted outside Terraform)**
   ```go
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
   ```

5. **Update model and save to state**
   ```go
   // Update the state with the latest data
   updateFrameworkModelFromAPI(&state, apiResponse)

   log.Printf("Read <resource> %s %s", id, *apiResponse.Name)

   // Save updated data into Terraform state
   resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
   }
   ```

**Key Points**:
- Always handle 404 by removing from state (not an error)
- Use `util.IsStatus404()` to check for 404
- Call `resp.State.RemoveResource(ctx)` for 404

**Deliverable**: Read method implemented with 404 handling

---

### Task 3.3: Implement Update Method

**Objective**: Implement resource update logic.

**Steps**:

1. **Add Update method signature**
   ```go
   func (r *<resource>FrameworkResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
   ```

2. **Read plan and state from request**
   ```go
   var plan <resource>FrameworkResourceModel
   var state <resource>FrameworkResourceModel

   // Read Terraform plan and current state data into the models
   resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
   resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
   if resp.Diagnostics.HasError() {
       return
   }
   ```

3. **Get proxy and build API request**
   ```go
   proxy := get<ResourceName>Proxy(r.clientConfig)
   id := state.Id.ValueString()  // Use ID from state, not plan
   apiRequest := build<ResourceName>FromFrameworkModel(plan)

   log.Printf("Updating <resource> %s", plan.Name.ValueString())
   ```

4. **Call proxy update method**
   ```go
   apiResponse, _, err := proxy.update<ResourceName>(ctx, id, apiRequest)
   if err != nil {
       resp.Diagnostics.AddError(
           "Error Updating <Resource>",
           fmt.Sprintf("Could not update <resource> %s: %s", plan.Name.ValueString(), err),
       )
       return
   }
   ```

5. **Update model and save to state**
   ```go
   // Update model with response data
   updateFrameworkModelFromAPI(&plan, apiResponse)

   log.Printf("Updated <resource> %s", plan.Name.ValueString())

   // Save updated data into Terraform state
   resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
   }
   ```

**Key Points**:
- Read both plan and state
- Use ID from state (ID doesn't change)
- Use plan for all other attributes

**Deliverable**: Update method implemented

---

### Task 3.4: Implement Delete Method

**Objective**: Implement resource deletion logic with verification.

**Steps**:

1. **Add Delete method signature**
   ```go
   func (r *<resource>FrameworkResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
   ```

2. **Read state from request**
   ```go
   var state <resource>FrameworkResourceModel

   // Read Terraform prior state data into the model
   resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
   if resp.Diagnostics.HasError() {
       return
   }
   ```

3. **Get proxy and call delete method**
   ```go
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
   ```

4. **Verify deletion with retry logic**
   ```go
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

**Key Points**:
- Call delete method first
- Verify deletion with retry (30 seconds typical)
- Check for 404 to confirm deletion
- Retry if resource still exists

**Deliverable**: Delete method implemented with verification

---

### Task 3.5: Implement ImportState Method

**Objective**: Enable resource import by ID.

**Steps**:

1. **Add ImportState method**
   ```go
   func (r *<resource>FrameworkResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
       // Retrieve import ID and save to id attribute
       resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
   }
   ```

**Key Points**:
- Simple passthrough pattern for most resources
- Framework will call Read to populate other attributes
- More complex imports may need custom logic

**Deliverable**: ImportState method implemented

---

## Phase 4: Helper Functions

### Task 4.1: Implement build<ResourceName>FromFrameworkModel

**Objective**: Convert Framework model to API request object.

**Steps**:

1. **Add function signature**
   ```go
   // build<ResourceName>FromFrameworkModel converts Framework model to API request model
   func build<ResourceName>FromFrameworkModel(model <resource>FrameworkResourceModel) *platformclientv2.<ApiRequestType> {
   ```

2. **Create API request object with required attributes**
   ```go
   request := &platformclientv2.<ApiRequestType>{
       Name: model.Name.ValueStringPointer(),
   }
   ```

3. **Add optional attributes with null/unknown checks**
   ```go
   if !model.OptionalAttr.IsNull() && !model.OptionalAttr.IsUnknown() {
       request.OptionalAttr = model.OptionalAttr.ValueStringPointer()
   }
   ```

4. **Add nested objects with null/unknown checks**
   ```go
   if !model.NestedId.IsNull() && !model.NestedId.IsUnknown() {
       request.Nested = &platformclientv2.NestedObject{
           Id: model.NestedId.ValueStringPointer(),
       }
   }
   ```

5. **Return request object**
   ```go
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

**Key Points**:
- Use `ValueStringPointer()` to convert Framework types to API pointers
- Always check `IsNull()` and `IsUnknown()` for optional attributes
- Required attributes don't need null checks
- Create nested objects as needed

**Deliverable**: build function implemented

---

### Task 4.2: Implement updateFrameworkModelFromAPI

**Objective**: Convert API response to Framework model.

**Steps**:

1. **Add function signature**
   ```go
   // updateFrameworkModelFromAPI updates Framework model from API response
   func updateFrameworkModelFromAPI(model *<resource>FrameworkResourceModel, apiResponse *platformclientv2.<ApiResponseType>) {
   ```

2. **Update required attributes**
   ```go
   model.Id = types.StringValue(*apiResponse.Id)
   model.Name = types.StringValue(*apiResponse.Name)
   ```

3. **Update optional attributes with nil checks**
   ```go
   if apiResponse.OptionalAttr != nil {
       model.OptionalAttr = types.StringValue(*apiResponse.OptionalAttr)
   } else {
       model.OptionalAttr = types.StringNull()
   }
   ```

4. **Update nested objects with nil checks**
   ```go
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

**Key Points**:
- Use `types.StringValue()` to convert API pointers to Framework types
- Use `types.StringNull()` for missing optional attributes
- Check nil before dereferencing pointers
- Check both parent and child for nested objects

**Deliverable**: update function implemented

---

## Phase 5: GetAll Functions

### Task 5.1: Implement GetAll<ResourceName> (Framework Version)

**Objective**: Implement Framework version of GetAll for Phase 2 future.

**Steps**:

1. **Add function with documentation**
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
   ```

2. **Get proxy and fetch all resources**
   ```go
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
   ```

3. **Build export map**
   ```go
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

**Deliverable**: Framework GetAll function implemented

---

### Task 5.2: Implement GetAll<ResourceName>SDK (SDK Version)

**Objective**: Implement SDK version of GetAll for Phase 1 current use.

**Steps**:

1. **Add function with documentation**
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
   ```

2. **Get proxy and fetch all resources**
   ```go
   proxy := get<ResourceName>Proxy(clientConfig)

   // Step 1: Fetch all resources (lightweight - just IDs and names)
   resources, _, err := proxy.getAll<ResourceName>(ctx)
   if err != nil {
       return nil, sdkdiag.Errorf("Failed to get <resources> for export: %v", err)
   }

   if resources == nil {
       return resourceExporter.ResourceIDMetaMap{}, nil
   }
   ```

3. **Build initial export map**
   ```go
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
   ```

4. **Build flat attribute maps (Phase 1 temporary)**
   ```go
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

5. **Implement build<ResourceName>Attributes helper**
   ```go
   // build<ResourceName>Attributes creates a flat attribute map from SDK object for export.
   // This function converts the SDK object to a flat map matching SDKv2 InstanceState format.
   //
   // Parameters:
   //   - resource: Resource object from API
   //
   // Returns:
   //   - map[string]string: Flat attribute map with all resource attributes
   //
   // Note: This is Phase 1 temporary code. Will be removed in Phase 2.
   func build<ResourceName>Attributes(resource *platformclientv2.<ResourceType>) map[string]string {
       attributes := make(map[string]string)

       // Basic attributes
       if resource.Id != nil {
           attributes["id"] = *resource.Id
       }
       if resource.Name != nil {
           attributes["name"] = *resource.Name
       }
       if resource.Description != nil {
           attributes["description"] = *resource.Description
       }

       // ⭐ CRITICAL: Dependency reference (used by exporter for dependency resolution)
       if resource.Division != nil && resource.Division.Id != nil {
           attributes["division_id"] = *resource.Division.Id
       }

       return attributes
   }
   ```

**Example** (routing_wrapupcode - see reference file for complete implementation)

**Key Points**:
- SDK version is currently used by exporter
- Builds flat attribute map for each resource
- Includes all dependency references in attribute map
- Marked as Phase 1 temporary with TODO comment

**Deliverable**: SDK GetAll function and build attributes helper implemented

---

## Phase 6: Data Source Implementation

### Task 6.1: Create Data Source File

**Objective**: Create the data source implementation file.

**Steps**:

1. **Navigate to resource package directory**
   ```powershell
   cd genesyscloud\<resource_name>
   ```

2. **Create the data source file**
   ```powershell
   New-Item -ItemType File -Name "data_source_genesyscloud_<resource_name>.go"
   ```
   Example:
   ```powershell
   New-Item -ItemType File -Name "data_source_genesyscloud_routing_wrapupcode.go"
   ```

3. **Add package declaration and imports**
   ```go
   package <resource_name>

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

**Deliverable**: Data source file created with package and imports

---

### Task 6.2: Implement Data Source Struct and Model

**Objective**: Define data source struct and model.

**Steps**:

1. **Add interface verification**
   ```go
   // Ensure <resource>FrameworkDataSource satisfies various data source interfaces.
   var (
       _ datasource.DataSource              = &<resource>FrameworkDataSource{}
       _ datasource.DataSourceWithConfigure = &<resource>FrameworkDataSource{}
   )
   ```

2. **Define data source struct**
   ```go
   // <resource>FrameworkDataSource defines the data source implementation for Plugin Framework.
   type <resource>FrameworkDataSource struct {
       clientConfig *platformclientv2.Configuration
   }
   ```

3. **Define data source model**
   ```go
   // <resource>FrameworkDataSourceModel describes the data source data model.
   type <resource>FrameworkDataSourceModel struct {
       Id   types.String `tfsdk:"id"`
       Name types.String `tfsdk:"name"`
   }
   ```

4. **Add constructor function**
   ```go
   // New<ResourceName>FrameworkDataSource is a helper function to simplify the provider implementation.
   func New<ResourceName>FrameworkDataSource() datasource.DataSource {
       return &<resource>FrameworkDataSource{}
   }
   ```

**Example** (routing_wrapupcode):
```go
var (
    _ datasource.DataSource              = &routingWrapupcodeFrameworkDataSource{}
    _ datasource.DataSourceWithConfigure = &routingWrapupcodeFrameworkDataSource{}
)

type routingWrapupcodeFrameworkDataSource struct {
    clientConfig *platformclientv2.Configuration
}

type routingWrapupcodeFrameworkDataSourceModel struct {
    Id   types.String `tfsdk:"id"`
    Name types.String `tfsdk:"name"`
}

func NewRoutingWrapupcodeFrameworkDataSource() datasource.DataSource {
    return &routingWrapupcodeFrameworkDataSource{}
}
```

**Deliverable**: Data source struct and model defined

---

### Task 6.3: Implement Data Source Methods

**Objective**: Implement Metadata, Schema, and Configure methods.

**Steps**:

1. **Add Metadata method**
   ```go
   func (d *<resource>FrameworkDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
       resp.TypeName = req.ProviderTypeName + "_<resource_name>"
   }
   ```

2. **Add Schema method**
   ```go
   func (d *<resource>FrameworkDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
       resp.Schema = <ResourceName>DataSourceSchema()
   }
   ```

3. **Add Configure method**
   ```go
   func (d *<resource>FrameworkDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
       // Prevent panic if the provider has not been configured.
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
   ```

**Deliverable**: Data source interface methods implemented

---

### Task 6.4: Implement Data Source Read with Retry

**Objective**: Implement data source read logic with retry for eventual consistency.

**Steps**:

1. **Add Read method signature**
   ```go
   func (d *<resource>FrameworkDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
   ```

2. **Read config from request**
   ```go
   var config <resource>FrameworkDataSourceModel

   // Read Terraform configuration data into the model
   resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
   if resp.Diagnostics.HasError() {
       return
   }
   ```

3. **Get proxy and name**
   ```go
   proxy := get<ResourceName>Proxy(d.clientConfig)
   name := config.Name.ValueString()

   log.Printf("Reading <resource> data source for name: %s", name)
   ```

4. **Implement retry logic**
   ```go
   // Use retry logic for eventual consistency (15-second timeout)
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
   ```

5. **Set ID and save to state**
   ```go
   // Set the ID in the model
   config.Id = types.StringValue(resourceId)

   log.Printf("Found <resource> %s with ID %s", name, resourceId)

   // Save data into Terraform state
   resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
   }
   ```

**Key Points**:
- Use retry logic for eventual consistency
- Retry for 15 seconds (typical timeout)
- Proxy method returns retryable flag
- Log retry attempts for debugging

**Deliverable**: Data source Read method implemented with retry

---

## Phase 7: Validation and Review

### Task 7.1: Compile and Verify

**Objective**: Ensure the code compiles without errors.

**Steps**:

1. **Run Go build**
   ```powershell
   go build ./genesyscloud/<resource_name>
   ```

2. **Fix any compilation errors**
   - Missing imports
   - Syntax errors
   - Type mismatches
   - Undefined functions

3. **Run Go fmt**
   ```powershell
   go fmt ./genesyscloud/<resource_name>/...
   ```

4. **Run Go vet**
   ```powershell
   go vet ./genesyscloud/<resource_name>/...
   ```

5. **Verify no errors or warnings**

**Deliverable**: Code compiles successfully with no errors

---

### Task 7.2: Review Against Checklist

**Objective**: Verify all requirements are met.

**Steps**:

1. **Use the validation checklist from requirements.md**

   **Resource Implementation File**:
   - [ ] File created: `resource_genesyscloud_<resource_name>.go`
   - [ ] Package declaration matches directory name
   - [ ] All required imports are present
   - [ ] No unused imports

   **Resource Struct and Model**:
   - [ ] Resource struct implements required interfaces
   - [ ] Resource struct has `clientConfig` field
   - [ ] Resource model struct defined with all attributes
   - [ ] Model uses Framework types
   - [ ] Model has correct `tfsdk` struct tags
   - [ ] Constructor function implemented

   **CRUD Methods**:
   - [ ] Metadata() method implemented
   - [ ] Schema() method implemented
   - [ ] Configure() method implemented
   - [ ] Create() method implemented
   - [ ] Read() method implemented
   - [ ] Update() method implemented
   - [ ] Delete() method implemented
   - [ ] ImportState() method implemented

   **CRUD Method Behavior**:
   - [ ] Create: Reads plan, calls proxy, updates state
   - [ ] Read: Calls proxy, handles 404, updates state
   - [ ] Update: Reads plan and state, calls proxy, updates state
   - [ ] Delete: Calls proxy, verifies deletion with retry
   - [ ] All methods handle errors with clear diagnostics
   - [ ] All methods use context appropriately

   **Data Source Implementation**:
   - [ ] File created: `data_source_genesyscloud_<resource_name>.go`
   - [ ] Data source struct and model defined
   - [ ] All interface methods implemented
   - [ ] Read method implemented with retry logic

   **GetAll Functions**:
   - [ ] `GetAll<ResourceName>()` implemented (Framework version)
   - [ ] `GetAll<ResourceName>SDK()` implemented (SDK version)
   - [ ] Both return `resourceExporter.ResourceIDMetaMap`
   - [ ] SDK version includes flat attribute map
   - [ ] Functions include Phase 1/Phase 2 comments

   **Helper Functions**:
   - [ ] `build<ResourceName>FromFrameworkModel()` implemented
   - [ ] `updateFrameworkModelFromAPI()` implemented
   - [ ] `build<ResourceName>Attributes()` implemented (for SDK GetAll)
   - [ ] Helper functions handle null/unknown values

   **Code Quality**:
   - [ ] Code compiles without errors
   - [ ] Code follows Go conventions
   - [ ] Functions have clear comments
   - [ ] Error messages are clear and actionable
   - [ ] Logging is appropriate

2. **Fix any issues found**

**Deliverable**: All checklist items verified

---

### Task 7.3: Code Review and Approval

**Objective**: Get peer review and approval before proceeding to Stage 3.

**Steps**:

1. **Create pull request or review request**
   - Include link to Stage 2 requirements and design docs
   - Highlight any deviations from standard pattern
   - Note any complex logic or special cases

2. **Address review comments**
   - Make requested changes
   - Re-verify checklist
   - Re-test compilation

3. **Get approval**
   - Obtain approval from reviewer
   - Merge or mark as ready for Stage 3

**Deliverable**: Stage 2 approved and ready for Stage 3

---

## Common Issues and Solutions

### Issue 1: Null Pointer Dereference

**Problem**: Panic when accessing API response fields.

**Solution**:
- Always check if pointer is nil before dereferencing
- Use nil checks in `updateFrameworkModelFromAPI()`
- Example: `if apiResponse.Field != nil { ... }`

### Issue 2: Incorrect Type Conversion

**Problem**: Type mismatch between Framework types and API types.

**Solution**:
- Use `ValueStringPointer()` to convert Framework → API
- Use `types.StringValue()` to convert API → Framework
- Use `types.StringNull()` for missing optional values

### Issue 3: Missing Null/Unknown Checks

**Problem**: Sending null values to API causes errors.

**Solution**:
- Always check `IsNull()` and `IsUnknown()` before using values
- Example: `if !model.Field.IsNull() && !model.Field.IsUnknown() { ... }`

### Issue 4: 404 Not Handled in Read

**Problem**: Read method fails when resource deleted outside Terraform.

**Solution**:
- Check for 404 using `util.IsStatus404(apiResp)`
- Call `resp.State.RemoveResource(ctx)` for 404
- Don't add error diagnostic for 404

### Issue 5: Delete Verification Fails

**Problem**: Delete method reports error even though resource is deleted.

**Solution**:
- Use retry logic to wait for eventual consistency
- Check for 404 in retry loop (indicates success)
- Increase retry timeout if needed (30 seconds typical)

### Issue 6: Data Source Fails Immediately

**Problem**: Data source fails to find resource created in same apply.

**Solution**:
- Add retry logic with 15-second timeout
- Use proxy method's retryable flag
- Log retry attempts for debugging

### Issue 7: Export Attributes Missing

**Problem**: Exporter doesn't resolve dependencies correctly.

**Solution**:
- Verify `build<ResourceName>Attributes()` includes all attributes
- Ensure dependency attributes are included (e.g., `division_id`)
- Check attribute names match schema exactly

---

## Completion Criteria

Stage 2 is complete when:

- [ ] All tasks in this document are completed
- [ ] All checklist items are verified
- [ ] Code compiles without errors
- [ ] Code review is approved
- [ ] Resource CRUD operations are implemented
- [ ] Data source is implemented
- [ ] GetAll functions are implemented
- [ ] Helper functions are implemented

---

## Next Steps

After Stage 2 completion:

1. **Review and approval**
   - Get team review
   - Address any feedback
   - Get final approval

2. **Proceed to Stage 3**
   - Begin test migration
   - Implement resource tests
   - Implement data source tests
   - Create test helper functions

3. **Reference Stage 3 documentation**
   - Read Stage 3 `requirements.md`
   - Read Stage 3 `design.md`
   - Follow Stage 3 `tasks.md`

---

## Time Estimates

| Phase | Estimated Time |
|-------|----------------|
| Phase 1: Resource File Setup | 30-60 minutes |
| Phase 2: Resource Interface Methods | 30-45 minutes |
| Phase 3: CRUD Operations | 2-3 hours |
| Phase 4: Helper Functions | 1-2 hours |
| Phase 5: GetAll Functions | 1-2 hours |
| Phase 6: Data Source Implementation | 1-2 hours |
| Phase 7: Validation and Review | 1-2 hours |
| **Total** | **6-12 hours** |

*Note: Times vary based on resource complexity and familiarity with patterns.*

---

## References

- **Reference Implementation**: 
  - `genesyscloud/routing_wrapupcode/resource_genesyscloud_routing_wrapupcode.go`
  - `genesyscloud/routing_wrapupcode/data_source_genesyscloud_routing_wrapupcode.go`
- **Stage 2 Requirements**: `prompts/pf_simple_resource_migration/Stage2/requirements.md`
- **Stage 2 Design**: `prompts/pf_simple_resource_migration/Stage2/design.md`
- **Plugin Framework Documentation**: https://developer.hashicorp.com/terraform/plugin/framework
