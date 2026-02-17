# Stage 2 – Resource Migration Tasks (Complex Resources)

## Overview

This document provides step-by-step tasks for completing Stage 2 of the Plugin Framework migration for **complex resources**. Follow these tasks in order to migrate resource CRUD operations, data source logic, and comprehensive helper functions for nested structures from SDKv2 to Plugin Framework.

**Reference Implementation**: 
- `genesyscloud/user/resource_genesyscloud_user.go`
- `genesyscloud/user/resource_genesyscloud_user_utils.go`
- `genesyscloud/user/data_source_genesyscloud_user.go`

**Resource Complexity**: Complex resource with multiple nested blocks (1-level, 2-level, 3-level nesting), flatten/build functions, element type helpers, and orchestrated update operations.

**Estimated Time**: 16-24 hours (complex resources require significantly more time than simple resources)

---

## Prerequisites

Before starting Stage 2 tasks, ensure:

- [ ] Stage 1 (Schema Migration) is complete and approved
- [ ] You have reviewed the existing SDKv2 resource CRUD implementation
- [ ] You understand the proxy methods available
- [ ] You have read Stage 2 `requirements.md` and `design.md`
- [ ] You have studied the `user` reference implementation
- [ ] You understand nested structure patterns (1-level, 2-level, 3-level)
- [ ] You understand flatten/build function patterns
- [ ] You understand element type helper usage
- [ ] Development environment is set up and ready

---

## Task Checklist

### Phase 1: Resource File Setup
- [ ] Task 1.1: Create Resource Implementation File
- [ ] Task 1.2: Add Package Declaration and Imports
- [ ] Task 1.3: Add Interface Verification
- [ ] Task 1.4: Define Resource Struct and Models (Complex)
- [ ] Task 1.5: Implement Constructor Function

### Phase 2: Utils File Setup (Complex Resources)
- [ ] Task 2.1: Create Utils File
- [ ] Task 2.2: Add Package-Level Variables
- [ ] Task 2.3: Implement Element Type Helpers
- [ ] Task 2.4: Implement Internal Structs

### Phase 3: Resource Interface Methods
- [ ] Task 3.1: Implement Metadata Method
- [ ] Task 3.2: Implement Schema Method
- [ ] Task 3.3: Implement Configure Method

### Phase 4: Flatten Functions (SDK → Framework)
- [ ] Task 4.1: Implement 1-Level Flatten Functions
- [ ] Task 4.2: Implement 2-Level Flatten Functions
- [ ] Task 4.3: Implement 3-Level Flatten Functions
- [ ] Task 4.4: Test Flatten Functions

### Phase 5: Build Functions (Framework → SDK)
- [ ] Task 5.1: Implement 1-Level Build Functions
- [ ] Task 5.2: Implement 2-Level Build Functions
- [ ] Task 5.3: Implement 3-Level Build Functions
- [ ] Task 5.4: Test Build Functions

### Phase 6: Shared Read Logic
- [ ] Task 6.1: Implement readUser Helper
- [ ] Task 6.2: Implement Helper Functions for Read
- [ ] Task 6.3: Test Read Logic

### Phase 7: Update Orchestration
- [ ] Task 7.1: Implement updateUser Helper
- [ ] Task 7.2: Implement executeAllUpdates
- [ ] Task 7.3: Implement Individual Update Functions
- [ ] Task 7.4: Test Update Logic

### Phase 8: CRUD Operations (Complex)
- [ ] Task 8.1: Implement Create Method
- [ ] Task 8.2: Implement Read Method
- [ ] Task 8.3: Implement Update Method
- [ ] Task 8.4: Implement Delete Method
- [ ] Task 8.5: Implement ImportState Method

### Phase 9: Utility Functions
- [ ] Task 9.1: Implement hasChanges
- [ ] Task 9.2: Implement getDeletedUserId
- [ ] Task 9.3: Implement restoreDeletedUser
- [ ] Task 9.4: Implement Other Utilities

### Phase 10: GetAll Functions (Complex)
- [ ] Task 10.1: Implement GetAll<ResourceName> (Framework Version)
- [ ] Task 10.2: Implement GetAll<ResourceName>SDK (SDK Version with Lazy Fetch)
- [ ] Task 10.3: Implement buildUserAttributes Helper

### Phase 11: Data Source Implementation (Complex)
- [ ] Task 11.1: Create Data Source File
- [ ] Task 11.2: Implement Data Source Struct and Model
- [ ] Task 11.3: Implement Data Source Methods
- [ ] Task 11.4: Implement Data Source Read with Cache

### Phase 12: Validation and Review
- [ ] Task 12.1: Compile and Verify
- [ ] Task 12.2: Review Against Checklist
- [ ] Task 12.3: Code Review and Approval

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
   cd genesyscloud\user
   ```

2. **Create the resource file**
   ```powershell
   New-Item -ItemType File -Name "resource_genesyscloud_<resource_name>.go"
   ```
   Example:
   ```powershell
   New-Item -ItemType File -Name "resource_genesyscloud_user.go"
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
   package user
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
       "github.com/mypurecloud/platform-client-sdk-go/v176/platformclientv2"
       "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
       resourceExporter "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_exporter"
       "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
   )
   ```

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
   - `<resource>` → Your resource name in PascalCase
   - Example: `UserFrameworkResource`

**Example** (user):
```go
var (
    _ resource.Resource                = &UserFrameworkResource{}
    _ resource.ResourceWithConfigure   = &UserFrameworkResource{}
    _ resource.ResourceWithImportState = &UserFrameworkResource{}
)
```

**Deliverable**: Interface verification added

---

### Task 1.4: Define Resource Struct and Models (Complex)

**Objective**: Define the resource struct and all model structs for state management.

**Steps**:

1. **Define resource struct**
   ```go
   // <resource>FrameworkResource is the main resource struct that manages Genesys Cloud <resource> lifecycle operations.
   type <resource>FrameworkResource struct {
       clientConfig *platformclientv2.Configuration
   }
   ```

2. **Define main resource model struct**
   ```go
   // <resource>FrameworkResourceModel represents the complete Terraform state for a Genesys Cloud <resource>.
   type <resource>FrameworkResourceModel struct {
       Id                    types.String `tfsdk:"id"`
       Name                  types.String `tfsdk:"name"`
       // ... simple attributes
       NestedBlock1          types.Set    `tfsdk:"nested_block_1"`
       NestedBlock2          types.List   `tfsdk:"nested_block_2"`
       ComplexNestedBlock    types.List   `tfsdk:"complex_nested_block"`
   }
   ```

3. **Define nested structure models (1-level)**
   ```go
   // <NestedBlock>Model represents <description>.
   type <NestedBlock>Model struct {
       Attribute1 types.String `tfsdk:"attribute_1"`
       Attribute2 types.Int64  `tfsdk:"attribute_2"`
   }
   ```

4. **Define nested structure models (2-level)**
   ```go
   // <ParentBlock>Model represents <description>.
   type <ParentBlock>Model struct {
       ChildBlock types.Set `tfsdk:"child_block"`
   }

   // <ChildBlock>Model represents <description>.
   type <ChildBlock>Model struct {
       Attribute types.String `tfsdk:"attribute"`
   }
   ```

5. **Define nested structure models (3-level)**
   ```go
   // <GrandparentBlock>Model represents <description>.
   type <GrandparentBlock>Model struct {
       ParentBlock types.List `tfsdk:"parent_block"`
   }

   // <ParentBlock>Model represents <description>.
   type <ParentBlock>Model struct {
       ChildBlock types.List `tfsdk:"child_block"`
   }

   // <ChildBlock>Model represents <description>.
   type <ChildBlock>Model struct {
       Attribute types.String `tfsdk:"attribute"`
   }
   ```

**Example** (user):
```go
type UserFrameworkResource struct {
    clientConfig *platformclientv2.Configuration
}

type UserFrameworkResourceModel struct {
    Id                    types.String `tfsdk:"id"`
    Email                 types.String `tfsdk:"email"`
    Name                  types.String `tfsdk:"name"`
    RoutingSkills         types.Set    `tfsdk:"routing_skills"`
    RoutingLanguages      types.Set    `tfsdk:"routing_languages"`
    Addresses             types.List   `tfsdk:"addresses"`
    RoutingUtilization    types.List   `tfsdk:"routing_utilization"`
    VoicemailUserpolicies types.List   `tfsdk:"voicemail_userpolicies"`
}

// 1-level nested
type VoicemailUserpoliciesModel struct {
    AlertTimeoutSeconds    types.Int64 `tfsdk:"alert_timeout_seconds"`
    SendEmailNotifications types.Bool  `tfsdk:"send_email_notifications"`
}

// 3-level nested
type RoutingUtilizationModel struct {
    Call              types.List `tfsdk:"call"`
    Callback          types.List `tfsdk:"callback"`
    LabelUtilizations types.List `tfsdk:"label_utilizations"`
}

type MediaUtilizationModel struct {
    MaximumCapacity         types.Int64 `tfsdk:"maximum_capacity"`
    IncludeNonAcd           types.Bool  `tfsdk:"include_non_acd"`
    InterruptibleMediaTypes types.Set   `tfsdk:"interruptible_media_types"`
}

type LabelUtilizationModel struct {
    LabelId              types.String `tfsdk:"label_id"`
    MaximumCapacity      types.Int64  `tfsdk:"maximum_capacity"`
    InterruptingLabelIds types.Set    `tfsdk:"interrupting_label_ids"`
}
```

**Key Points**:
- One model struct per nested block level
- All use Framework types consistently
- All have `tfsdk` struct tags
- Add descriptive comments for each model

**Deliverable**: Resource struct and all model structs defined

---

### Task 1.5: Implement Constructor Function

**Objective**: Create the constructor function for the resource.

**Steps**:

1. **Add constructor function**
   ```go
   // New<ResourceName>FrameworkResource is a factory function that creates a new instance of the <resource> resource.
   func New<ResourceName>FrameworkResource() resource.Resource {
       return &<resource>FrameworkResource{}
   }
   ```

**Example** (user):
```go
func NewUserFrameworkResource() resource.Resource {
    return &UserFrameworkResource{}
}
```

**Deliverable**: Constructor function implemented

---

## Phase 2: Utils File Setup (Complex Resources)

### Task 2.1: Create Utils File

**Objective**: Create a separate utils file for helper functions.

**Steps**:

1. **Navigate to resource package directory**
   ```powershell
   cd genesyscloud\<resource_name>
   ```

2. **Create the utils file**
   ```powershell
   New-Item -ItemType File -Name "resource_genesyscloud_<resource_name>_utils.go"
   ```
   Example:
   ```powershell
   New-Item -ItemType File -Name "resource_genesyscloud_user_utils.go"
   ```

3. **Add package declaration and imports**
   ```go
   package <resource_name>

   import (
       "context"
       "encoding/json"
       "fmt"
       "log"
       "sort"
       "strings"
       "time"

       "github.com/hashicorp/terraform-plugin-framework/attr"
       pfdiag "github.com/hashicorp/terraform-plugin-framework/diag"
       "github.com/hashicorp/terraform-plugin-framework/types"
       "github.com/hashicorp/terraform-plugin-framework/types/basetypes"
       "github.com/hashicorp/terraform-plugin-sdk/v2/diag"
       "github.com/mypurecloud/platform-client-sdk-go/v176/platformclientv2"
       "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
       chunksProcess "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/chunks"
       "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/lists"
   )
   ```

**Deliverable**: Utils file created with package and imports

---

### Task 2.2: Add Package-Level Variables

**Objective**: Define package-level variables and constants.

**Steps**:

1. **Add package-level variables**
   ```go
   var (
       // Map of SDK type name to schema type name
       <mapping>MediaTypes = map[string]string{
           "sdk_type_1": "schema_type_1",
           "sdk_type_2": "schema_type_2",
       }
   )
   ```

**Example** (user):
```go
var (
    // Map of SDK media type name to schema media type name
    utilizationMediaTypes = map[string]string{
        "call":     "call",
        "callback": "callback",
        "chat":     "chat",
        "email":    "email",
        "message":  "message",
    }
)
```

**Deliverable**: Package-level variables defined

---

### Task 2.3: Implement Element Type Helpers

**Objective**: Define element type helper functions for type-safe operations.

**Steps**:

1. **Implement element type helpers for each nested block**
   ```go
   // <nestedBlock>ElementType returns the element type for <nested_block> set/list.
   func <nestedBlock>ElementType() types.ObjectType {
       return types.ObjectType{
           AttrTypes: map[string]attr.Type{
               "attribute_1": types.StringType,
               "attribute_2": types.Int64Type,
           },
       }
   }
   ```

2. **Implement attribute type helpers for complex nested blocks**
   ```go
   // get<NestedBlock>AttrTypes returns the attribute types for <nested_block>.
   func get<NestedBlock>AttrTypes() map[string]attr.Type {
       return map[string]attr.Type{
           "attribute_1": types.StringType,
           "attribute_2": types.Int64Type,
           "nested_set":  types.SetType{ElemType: types.StringType},
       }
   }
   ```

**Example** (user):
```go
func routingSkillsElementType() types.ObjectType {
    return types.ObjectType{
        AttrTypes: map[string]attr.Type{
            "skill_id":    types.StringType,
            "proficiency": types.Float64Type,
        },
    }
}

func getMediaUtilizationAttrTypes() map[string]attr.Type {
    return map[string]attr.Type{
        "maximum_capacity":          types.Int64Type,
        "include_non_acd":           types.BoolType,
        "interruptible_media_types": types.SetType{ElemType: types.StringType},
    }
}
```

**Key Points**:
- One element type helper per nested block
- Reuse these helpers in flatten/build functions
- Ensures type consistency across all operations

**Deliverable**: Element type helpers implemented for all nested blocks

---

### Task 2.4: Implement Internal Structs

**Objective**: Define internal structs for API response unmarshaling.

**Steps**:

1. **Add internal structs for complex API responses**
   ```go
   // <internalStruct> mirrors the SDK response structure for <api> API responses.
   type <internalStruct> struct {
       Field1 string                      `json:"field1"`
       Field2 map[string]<nestedStruct>   `json:"field2"`
   }

   type <nestedStruct> struct {
       Attribute1 int32    `json:"attribute1"`
       Attribute2 []string `json:"attribute2"`
   }
   ```

**Example** (user):
```go
// agentUtilizationWithLabels mirrors the SDK response structure for agent utilization API responses.
type agentUtilizationWithLabels struct {
    Utilization       map[string]mediaUtilization `json:"utilization"`
    LabelUtilizations map[string]labelUtilization `json:"labelUtilizations"`
    Level             string                      `json:"level"`
}

type mediaUtilization struct {
    MaximumCapacity         int32    `json:"maximumCapacity"`
    InterruptableMediaTypes []string `json:"interruptableMediaTypes"`
    IncludeNonAcd           bool     `json:"includeNonAcd"`
}

type labelUtilization struct {
    MaximumCapacity      int32    `json:"maximumCapacity"`
    InterruptingLabelIds []string `json:"interruptingLabelIds"`
}
```

**Deliverable**: Internal structs defined for API responses

---

## Phase 3: Resource Interface Methods

### Task 3.1: Implement Metadata Method

**Objective**: Provide resource type name to the provider.

**Steps**:

1. **Add Metadata method**
   ```go
   // Metadata sets the resource type name that will be used in Terraform configurations.
   func (r *<resource>FrameworkResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
       resp.TypeName = req.ProviderTypeName + "_<resource_name>"
   }
   ```

**Example** (user):
```go
func (r *UserFrameworkResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
    resp.TypeName = req.ProviderTypeName + "_user"
}
```

**Deliverable**: Metadata method implemented

---

### Task 3.2: Implement Schema Method

**Objective**: Provide resource schema to the provider.

**Steps**:

1. **Add Schema method**
   ```go
   // Schema defines the complete resource schema including all attributes, their types, and validation rules.
   func (r *<resource>FrameworkResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
       resp.Schema = <ResourceName>ResourceSchema()
   }
   ```

**Example** (user):
```go
func (r *UserFrameworkResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
    resp.Schema = UserResourceSchema()
}
```

**Deliverable**: Schema method implemented

---

### Task 3.3: Implement Configure Method

**Objective**: Receive and store provider configuration.

**Steps**:

1. **Add Configure method**
   ```go
   // Configure receives the provider's configured API client and stores it in the resource instance.
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

## Phase 4: Flatten Functions (SDK → Framework)

### Task 4.1: Implement 1-Level Flatten Functions

**Objective**: Implement flatten functions for 1-level nested blocks.

**Steps**:

1. **For each 1-level nested block, implement flatten function**
   ```go
   // flatten<NestedBlock> converts SDK types to Framework types for <nested_block>.
   func flatten<NestedBlock>(ctx context.Context, sdkData *[]platformclientv2.<SdkType>) (types.Set, pfdiag.Diagnostics) {
       var diags pfdiag.Diagnostics
       
       if sdkData == nil || len(*sdkData) == 0 {
           return types.SetNull(<nestedBlock>ElementType()), diags
       }
       
       elements := make([]attr.Value, 0)
       for _, item := range *sdkData {
           obj, objDiags := types.ObjectValue(
               <nestedBlock>ElementType().AttrTypes,
               map[string]attr.Value{
                   "attribute_1": types.StringValue(*item.Attribute1),
                   "attribute_2": types.Int64Value(int64(*item.Attribute2)),
               },
           )
           diags.Append(objDiags...)
           elements = append(elements, obj)
       }
       
       set, setDiags := types.SetValue(<nestedBlock>ElementType(), elements)
       diags.Append(setDiags...)
       
       return set, diags
   }
   ```

**Example** (user - routing_skills):
```go
func flattenUserSkills(ctx context.Context, skills *[]platformclientv2.Userroutingskill) (types.Set, pfdiag.Diagnostics) {
    var diags pfdiag.Diagnostics
    
    if skills == nil || len(*skills) == 0 {
        return types.SetNull(routingSkillsElementType()), diags
    }
    
    skillValues := make([]attr.Value, 0)
    for _, skill := range *skills {
        skillObj, objDiags := types.ObjectValue(
            routingSkillsElementType().AttrTypes,
            map[string]attr.Value{
                "skill_id":    types.StringValue(*skill.Id),
                "proficiency": types.Float64Value(*skill.Proficiency),
            },
        )
        diags.Append(objDiags...)
        skillValues = append(skillValues, skillObj)
    }
    
    skillSet, setDiags := types.SetValue(routingSkillsElementType(), skillValues)
    diags.Append(setDiags...)
    
    return skillSet, diags
}
```

**Key Points**:
- Use element type helper for type safety
- Handle null/empty cases
- Return diagnostics for error handling
- Use `types.SetValue()` or `types.ListValue()`

**Deliverable**: All 1-level flatten functions implemented

---

### Task 4.2: Implement 2-Level Flatten Functions

**Objective**: Implement flatten functions for 2-level nested blocks.

**Steps**:

1. **For each 2-level nested block, implement parent flatten function**
   ```go
   // flatten<ParentBlock> converts SDK types to Framework types for <parent_block> (2-level nesting).
   func flatten<ParentBlock>(ctx context.Context, sdkData *[]platformclientv2.<SdkType>) (types.List, pfdiag.Diagnostics) {
       var diagnostics pfdiag.Diagnostics
       
       if sdkData == nil || len(*sdkData) == 0 {
           return types.ListNull(<parentBlock>ElementType()), diagnostics
       }
       
       // Flatten child blocks
       childElements, childDiags := flatten<ChildBlock>(ctx, sdkData)
       diagnostics.Append(childDiags...)
       
       // Build parent object
       parentObj, objDiags := types.ObjectValue(
           <parentBlock>ElementType().AttrTypes,
           map[string]attr.Value{
               "child_block": childElements,
           },
       )
       diagnostics.Append(objDiags...)
       
       // Wrap in list (MaxItems: 1)
       parentList, listDiags := types.ListValue(<parentBlock>ElementType(), []attr.Value{parentObj})
       diagnostics.Append(listDiags...)
       
       return parentList, diagnostics
   }
   ```

2. **Implement child flatten function**
   ```go
   // flatten<ChildBlock> converts SDK types to Framework types for <child_block>.
   func flatten<ChildBlock>(ctx context.Context, sdkData *[]platformclientv2.<SdkType>) (types.Set, pfdiag.Diagnostics) {
       // Similar to 1-level flatten function
   }
   ```

**Example** (user - addresses with phone_numbers and other_emails):
```go
func flattenUserAddresses(ctx context.Context, addresses *[]platformclientv2.Contact, proxy *userProxy) (types.List, pfdiag.Diagnostics) {
    var diagnostics pfdiag.Diagnostics
    
    if addresses == nil || len(*addresses) == 0 {
        return types.ListNull(addressesElementType()), diagnostics
    }
    
    // Initialize collections for emails and phone numbers
    emailElements := make([]attr.Value, 0)
    phoneElements := make([]attr.Value, 0)
    
    // Process each contact from API
    for _, address := range *addresses {
        switch *address.MediaType {
        case "SMS", "PHONE":
            phoneObj, phoneDiags := buildPhoneObject(address, proxy)
            diagnostics.Append(phoneDiags...)
            phoneElements = append(phoneElements, phoneObj)
            
        case "EMAIL":
            emailObj, emailDiags := buildEmailObject(address)
            diagnostics.Append(emailDiags...)
            emailElements = append(emailElements, emailObj)
        }
    }
    
    // Create email set
    emailSet, setDiags := types.SetValue(emailElementType(), emailElements)
    diagnostics.Append(setDiags...)
    
    // Create phone number set
    phoneSet, setDiags := types.SetValue(phoneElementType(), phoneElements)
    diagnostics.Append(setDiags...)
    
    // Create the addresses object containing both sets
    addressesObj, objDiags := types.ObjectValue(addressesElementType().AttrTypes, map[string]attr.Value{
        "other_emails":  emailSet,
        "phone_numbers": phoneSet,
    })
    diagnostics.Append(objDiags...)
    
    // Return as a list with one element
    addressesList, listDiags := types.ListValue(addressesElementType(), []attr.Value{addressesObj})
    diagnostics.Append(listDiags...)
    
    return addressesList, diagnostics
}
```

**Key Points**:
- Process parent, then children
- Handle null/empty at each level
- Use element type helpers consistently
- Return diagnostics for error handling

**Deliverable**: All 2-level flatten functions implemented

---

### Task 4.3: Implement 3-Level Flatten Functions

**Objective**: Implement flatten functions for 3-level nested blocks.

**Steps**:

1. **For each 3-level nested block, implement grandparent flatten function**
   ```go
   // flatten<GrandparentBlock> converts SDK types to Framework types for <grandparent_block> (3-level nesting).
   func flatten<GrandparentBlock>(ctx context.Context, sdkData *<SdkType>) (types.List, pfdiag.Diagnostics) {
       var diagnostics pfdiag.Diagnostics
       
       if sdkData == nil {
           return types.ListNull(<grandparentBlock>ElementType()), diagnostics
       }
       
       // Build grandparent attributes
       grandparentAttrs := map[string]attr.Value{
           "parent_block_1": types.ListNull(<parentBlock>ElementType()),
           "parent_block_2": types.ListNull(<parentBlock>ElementType()),
       }
       
       // Flatten each parent block
       if sdkData.ParentBlock1 != nil {
           parent1, parent1Diags := flatten<ParentBlock1>(ctx, sdkData.ParentBlock1)
           diagnostics.Append(parent1Diags...)
           grandparentAttrs["parent_block_1"] = parent1
       }
       
       // Create grandparent object
       grandparentObj, objDiags := types.ObjectValue(<grandparentBlock>ElementType().AttrTypes, grandparentAttrs)
       diagnostics.Append(objDiags...)
       
       // Wrap in list
       grandparentList, listDiags := types.ListValue(<grandparentBlock>ElementType(), []attr.Value{grandparentObj})
       diagnostics.Append(listDiags...)
       
       return grandparentList, diagnostics
   }
   ```

2. **Implement parent flatten functions**
3. **Implement child flatten functions**

**Example** (user - routing_utilization):
```go
func readUserRoutingUtilization(ctx context.Context, state *UserFrameworkResourceModel, proxy *userProxy) (*platformclientv2.APIResponse, pfdiag.Diagnostics) {
    var diagnostics pfdiag.Diagnostics
    
    // Make API call and unmarshal
    // ... (API call code)
    
    // Build the settings object (grandparent level)
    allSettingsAttrs := map[string]attr.Value{
        "call":               types.ListNull(mediaUtilizationElementType()),
        "callback":           types.ListNull(mediaUtilizationElementType()),
        "label_utilizations": types.ListNull(labelUtilizationElementType()),
    }
    
    // Flatten media utilization settings (parent level)
    if agentUtilization.Utilization != nil {
        for sdkType, schemaType := range getUtilizationMediaTypes() {
            if mediaSettings, ok := agentUtilization.Utilization[sdkType]; ok {
                flattenedMedia, diags := flattenUtilizationSetting(mediaSettings)
                diagnostics.Append(diags...)
                allSettingsAttrs[schemaType] = flattenedMedia
            }
        }
    }
    
    // Create the settings object
    settingsObj, diags := types.ObjectValue(routingUtilizationElementType().AttrTypes, allSettingsAttrs)
    diagnostics.Append(diags...)
    
    // Create the list with one element
    utilizationList, diags := types.ListValue(routingUtilizationElementType(), []attr.Value{settingsObj})
    diagnostics.Append(diags...)
    state.RoutingUtilization = utilizationList
    
    return response, diagnostics
}
```

**Key Points**:
- Process grandparent, then parents, then children
- Handle null/empty at each level
- Use element type helpers consistently
- Complex 3-level nesting may require direct API calls

**Deliverable**: All 3-level flatten functions implemented

---

### Task 4.4: Test Flatten Functions

**Objective**: Verify flatten functions work correctly.

**Steps**:

1. **Create test SDK data**
2. **Call flatten function**
3. **Verify Framework types are correct**
4. **Verify all attributes are preserved**
5. **Verify null/empty cases handled**

**Deliverable**: All flatten functions tested and working

---

## Phase 5: Build Functions (Framework → SDK)

### Task 5.1: Implement 1-Level Build Functions

**Objective**: Implement build functions for 1-level nested blocks.

**Steps**:

1. **For each 1-level nested block, implement build function**
   ```go
   // buildSdk<NestedBlock> converts Framework types to SDK types for <nested_block>.
   func buildSdk<NestedBlock>(ctx context.Context, data types.Set) *[]platformclientv2.<SdkType> {
       if data.IsNull() || data.IsUnknown() {
           return nil
       }
       
       elements := data.Elements()
       sdkItems := make([]platformclientv2.<SdkType>, 0, len(elements))
       
       for _, elem := range elements {
           obj, ok := elem.(types.Object)
           if !ok {
               continue
           }
           
           attrs := obj.Attributes()
           var attr1 types.String
           var attr2 types.Int64
           
           if val, exists := attrs["attribute_1"]; exists && !val.IsNull() {
               attr1 = val.(types.String)
           }
           if val, exists := attrs["attribute_2"]; exists && !val.IsNull() {
               attr2 = val.(types.Int64)
           }
           
           sdkItems = append(sdkItems, platformclientv2.<SdkType>{
               Attribute1: attr1.ValueStringPointer(),
               Attribute2: platformclientv2.Int(int(attr2.ValueInt64())),
           })
       }
       
       return &sdkItems
   }
   ```

**Deliverable**: All 1-level build functions implemented

---

### Task 5.2: Implement 2-Level Build Functions

**Objective**: Implement build functions for 2-level nested blocks.

**Steps**:

1. **Implement parent build function that calls child build functions**
2. **Extract nested blocks using `ElementsAs()`**
3. **Build SDK types for each level**

**Deliverable**: All 2-level build functions implemented

---

### Task 5.3: Implement 3-Level Build Functions

**Objective**: Implement build functions for 3-level nested blocks.

**Steps**:

1. **Implement grandparent build function**
2. **Call parent build functions**
3. **Call child build functions**
4. **Handle null/unknown at each level**

**Deliverable**: All 3-level build functions implemented

---

### Task 5.4: Test Build Functions

**Objective**: Verify build functions work correctly.

**Steps**:

1. **Create test Framework data**
2. **Call build function**
3. **Verify SDK types are correct**
4. **Verify all attributes are preserved**
5. **Test round-trip: SDK → Framework → SDK**

**Deliverable**: All build functions tested and working

---

## Phase 6: Shared Read Logic

### Task 6.1: Implement readUser Helper

**Objective**: Implement shared read logic called from Create, Read, Update.

**Steps**:

1. **Add readUser function signature**
   ```go
   // readUser reads the user data from the API and populates the model with all attributes.
   func readUser(ctx context.Context, model *UserFrameworkResourceModel, proxy *userProxy, diagnostics *pfdiag.Diagnostics, isImport ...bool) {
   ```

2. **Implement retry logic**
   ```go
   retryDiags := util.PFWithRetriesForRead(ctx, func() (bool, error) {
       // Fetch user from API with expansions
       currentUser, proxyResponse, getUserErr := proxy.getUserById(ctx, model.Id.ValueString(), []string{
           "skills", "languages", "locations", "profileSkills", "certifications", "employerInfo",
       }, "")
       
       if getUserErr != nil {
           if util.IsStatus404(proxyResponse) {
               return true, fmt.Errorf("API Error: 404")
           }
           return false, fmt.Errorf("Failed to read user %s | error: %s", model.Id.ValueString(), getUserErr)
       }
       
       // Set basic attributes
       // Flatten nested structures
       // Handle managed vs unmanaged attributes
       
       return false, nil
   })
   
   diagnostics.Append(retryDiags...)
   }
   ```

3. **Implement attribute setting**
4. **Implement nested structure flattening**
5. **Implement managed vs unmanaged logic**

**Deliverable**: readUser helper implemented

---

### Task 6.2: Implement Helper Functions for Read

**Objective**: Implement helper functions used by readUser.

**Steps**:

1. **Implement setBasicUserAttributes**
2. **Implement handleManagedRoutingSkills**
3. **Implement handleManagedRoutingLanguages**
4. **Implement handleManagedLocations**
5. **Implement handleVoicemailUserpolicies**
6. **Implement handleRoutingUtilization**

**Deliverable**: All read helper functions implemented

---

### Task 6.3: Test Read Logic

**Objective**: Verify read logic works correctly.

**Steps**:

1. **Test basic attribute reading**
2. **Test nested structure flattening**
3. **Test managed vs unmanaged logic**
4. **Test import mode handling**
5. **Test retry logic**

**Deliverable**: Read logic tested and working

---

## Phase 7: Update Orchestration

### Task 7.1: Implement updateUser Helper

**Objective**: Implement shared update logic called from Create, Update, restore.

**Steps**:

1. **Add updateUser function**
2. **Implement state transition handling**
3. **Implement core attribute updates**
4. **Call executeAllUpdates**
5. **Call readUser to get final state**

**Deliverable**: updateUser helper implemented

---

### Task 7.2: Implement executeAllUpdates

**Objective**: Implement orchestration function for all update operations.

**Steps**:

1. **Add executeAllUpdates function**
   ```go
   func executeAllUpdates(ctx context.Context, plan *UserFrameworkResourceModel, proxy *userProxy, sdkConfig *platformclientv2.Configuration, updateObjectDivision bool, state ...*UserFrameworkResourceModel) pfdiag.Diagnostics {
       var diagnostics pfdiag.Diagnostics
       
       // Update division
       // Update skills
       // Update languages
       // Update profile skills
       // Update routing utilization
       // Update voicemail policies
       // Update password
       
       return diagnostics
   }
   ```

**Deliverable**: executeAllUpdates implemented

---

### Task 7.3: Implement Individual Update Functions

**Objective**: Implement update functions for each attribute type.

**Steps**:

1. **Implement updateUserSkills**
   - Handle managed vs unmanaged
   - Remove skills no longer in config
   - Add/update skills in config
   - Use bulk API

2. **Implement updateUserLanguages**
   - Similar pattern to skills

3. **Implement updateUserProfileSkills**
4. **Implement updateUserRoutingUtilization**
5. **Implement updateUserVoicemailPolicies**
6. **Implement updatePassword**

**Deliverable**: All individual update functions implemented

---

### Task 7.4: Test Update Logic

**Objective**: Verify update logic works correctly.

**Steps**:

1. **Test managed vs unmanaged logic**
2. **Test attribute removal**
3. **Test attribute addition**
4. **Test attribute modification**
5. **Test update ordering**

**Deliverable**: Update logic tested and working

---

## Phase 8: CRUD Operations (Complex)

### Task 8.1: Implement Create Method

**Objective**: Implement complex resource creation logic.

**Steps**:

1. **Read plan from request**
2. **Build nested structures from plan**
3. **Check for deleted resource (restore pattern)**
4. **If deleted, restore and configure**
5. **If not deleted, create new resource**
6. **Handle attributes requiring separate PATCH**
7. **Apply orchestrated updates**
8. **Read final state**
9. **Save to state**

**Deliverable**: Create method implemented

---

### Task 8.2: Implement Read Method

**Objective**: Implement complex resource read logic.

**Steps**:

1. **Read state from request**
2. **Call shared readUser helper**
3. **Save to state**

**Deliverable**: Read method implemented

---

### Task 8.3: Implement Update Method

**Objective**: Implement complex resource update logic.

**Steps**:

1. **Read plan and state from request**
2. **Call shared updateUser helper**
3. **Save to state**

**Deliverable**: Update method implemented

---

### Task 8.4: Implement Delete Method

**Objective**: Implement complex resource deletion logic.

**Steps**:

1. **Read state from request**
2. **Call proxy delete with retry (version mismatch)**
3. **Verify deletion with retry (soft delete pattern)**
4. **Handle errors**

**Deliverable**: Delete method implemented

---

### Task 8.5: Implement ImportState Method

**Objective**: Enable resource import with full nested structures.

**Steps**:

1. **Add ImportState method**
2. **Call readUser with import mode flag**
3. **Save to state**

**Deliverable**: ImportState method implemented

---

## Phase 9: Utility Functions

### Task 9.1: Implement hasChanges

**Objective**: Detect which attributes have non-null values.

**Steps**:

1. **Add hasChanges function**
2. **Check each attribute for null/unknown**
3. **Return true if any attribute has value**

**Deliverable**: hasChanges implemented

---

### Task 9.2: Implement getDeletedUserId

**Objective**: Search for resource in deleted state.

**Steps**:

1. **Add getDeletedUserId function**
2. **Call search API with deleted state filter**
3. **Return ID if found, nil otherwise**

**Deliverable**: getDeletedUserId implemented

---

### Task 9.3: Implement restoreDeletedUser

**Objective**: Restore deleted resource and apply full configuration.

**Steps**:

1. **Add restoreDeletedUser function**
2. **Get current user with version**
3. **PATCH to restore state**
4. **Call updateUser to apply full configuration**

**Deliverable**: restoreDeletedUser implemented

---

### Task 9.4: Implement Other Utilities

**Objective**: Implement remaining utility functions.

**Steps**:

1. **Implement waitForExtensionPoolActivation** (if needed)
2. **Implement executeUpdateUser** (retry wrapper)
3. **Implement convertSDKDiagnosticsToFramework**
4. **Implement logging helpers** (invMustJSON, invStr, etc.)

**Deliverable**: All utility functions implemented

---

## Phase 10: GetAll Functions (Complex)

### Task 10.1: Implement GetAll<ResourceName> (Framework Version)

**Objective**: Implement Framework version for Phase 2 future.

**Steps**:

1. **Add GetAll function with Framework diagnostics**
2. **Fetch all resources**
3. **Build export map with hash**
4. **Return map and diagnostics**

**Deliverable**: Framework GetAll implemented

---

### Task 10.2: Implement GetAll<ResourceName>SDK (SDK Version with Lazy Fetch)

**Objective**: Implement SDK version with lazy fetch pattern.

**Steps**:

1. **Add GetAllSDK function with SDK diagnostics**
2. **Fetch all resource IDs (lightweight)**
3. **Build initial export map**
4. **Set LazyFetchAttributes callback for each resource**
5. **Callback fetches full details with expansions**
6. **Callback builds complete attribute map**

**Deliverable**: SDK GetAll with lazy fetch implemented

---

### Task 10.3: Implement buildUserAttributes Helper

**Objective**: Build flat attribute map for export.

**Steps**:

1. **Add buildUserAttributes function**
2. **Flatten basic attributes**
3. **Flatten nested structures to JSON**
4. **Fetch and flatten routing utilization**
5. **Fetch and flatten voicemail policies**
6. **Return attribute map**

**Deliverable**: buildUserAttributes implemented

---

## Phase 11: Data Source Implementation (Complex)

### Task 11.1: Create Data Source File

**Objective**: Create data source file with package and imports.

**Steps**:

1. **Create file**: `data_source_genesyscloud_<resource_name>.go`
2. **Add package declaration and imports**

**Deliverable**: Data source file created

---

### Task 11.2: Implement Data Source Struct and Model

**Objective**: Define data source struct and model.

**Steps**:

1. **Add interface verification**
2. **Define data source struct**
3. **Define data source model**
4. **Add constructor function**

**Deliverable**: Data source struct and model defined

---

### Task 11.3: Implement Data Source Methods

**Objective**: Implement Metadata, Schema, Configure methods.

**Steps**:

1. **Add Metadata method**
2. **Add Schema method**
3. **Add Configure method**

**Deliverable**: Data source interface methods implemented

---

### Task 11.4: Implement Data Source Read with Cache

**Objective**: Implement read logic with cache and retry.

**Steps**:

1. **Add Read method**
2. **Validate search fields**
3. **Initialize cache if needed**
4. **Retrieve ID from cache or API**
5. **Fetch full details**
6. **Populate name and email**
7. **Save to state**

8. **Implement getUserByName helper**
9. **Implement hydrateUserCache helper**

**Deliverable**: Data source Read with cache implemented

---

## Phase 12: Validation and Review

### Task 12.1: Compile and Verify

**Objective**: Ensure code compiles without errors.

**Steps**:

1. **Run Go build**
   ```powershell
   go build ./genesyscloud/<resource_name>
   ```

2. **Fix compilation errors**
3. **Run Go fmt**
   ```powershell
   go fmt ./genesyscloud/<resource_name>/...
   ```

4. **Run Go vet**
   ```powershell
   go vet ./genesyscloud/<resource_name>/...
   ```

**Deliverable**: Code compiles successfully

---

### Task 12.2: Review Against Checklist

**Objective**: Verify all requirements are met.

**Steps**:

1. **Use validation checklist from requirements.md**
2. **Verify all files created**
3. **Verify all functions implemented**
4. **Verify all nested structures handled**
5. **Verify managed vs unmanaged logic**
6. **Fix any issues found**

**Deliverable**: All checklist items verified

---

### Task 12.3: Code Review and Approval

**Objective**: Get peer review and approval.

**Steps**:

1. **Create pull request**
2. **Address review comments**
3. **Get approval**

**Deliverable**: Stage 2 approved and ready for Stage 3

---

## Common Issues and Solutions (Complex Resources)

### Issue 1: Incorrect Element Type Usage
**Problem**: Using wrong element type causes runtime errors.
**Solution**: Define element type helpers and reuse consistently.

### Issue 2: Missing Null/Unknown Checks in Nested Structures
**Problem**: Not checking null/unknown at each level causes panics.
**Solution**: Check null/unknown before accessing nested attributes.

### Issue 3: Forgetting Managed vs Unmanaged Logic
**Problem**: Clearing unmanaged attributes causes state drift.
**Solution**: Track whether attribute was previously managed.

### Issue 4: Incorrect Update Ordering
**Problem**: Updates in wrong order cause API errors.
**Solution**: Follow SDKv2 update order in executeAllUpdates.

### Issue 5: Not Handling 3-Level Nesting
**Problem**: Forgetting to flatten/build grandchild attributes.
**Solution**: Process each nesting level explicitly.

### Issue 6: Missing Import Mode Handling
**Problem**: Populating defaults during import causes diffs.
**Solution**: Pass import mode flag and skip defaults.

### Issue 7: Incomplete Lazy Fetch
**Problem**: Not fetching all nested structures in callback.
**Solution**: Fetch all expansions and build complete attribute map.

### Issue 8: Not Preserving Extensive Logging
**Problem**: Losing debugging information.
**Solution**: Preserve all log statements for complex operations.

---

## Completion Criteria

Stage 2 is complete when:

- [ ] All tasks in this document are completed
- [ ] All checklist items are verified
- [ ] Code compiles without errors
- [ ] All nested structures handled correctly
- [ ] Flatten/build functions preserve all data
- [ ] Managed vs unmanaged logic works
- [ ] Update orchestration works correctly
- [ ] Code review is approved

---

## Next Steps

After Stage 2 completion:

1. **Review and approval**
2. **Proceed to Stage 3 – Test Migration**
3. **Reference Stage 3 documentation**

---

## Time Estimates (Complex Resources)

| Phase | Estimated Time |
|-------|----------------|
| Phase 1: Resource File Setup | 1-2 hours |
| Phase 2: Utils File Setup | 1-2 hours |
| Phase 3: Resource Interface Methods | 30-45 minutes |
| Phase 4: Flatten Functions | 4-6 hours |
| Phase 5: Build Functions | 3-4 hours |
| Phase 6: Shared Read Logic | 2-3 hours |
| Phase 7: Update Orchestration | 3-4 hours |
| Phase 8: CRUD Operations | 2-3 hours |
| Phase 9: Utility Functions | 1-2 hours |
| Phase 10: GetAll Functions | 2-3 hours |
| Phase 11: Data Source | 2-3 hours |
| Phase 12: Validation and Review | 2-3 hours |
| **Total** | **24-38 hours** |

*Note: Complex resources require significantly more time than simple resources due to nested structures, flatten/build functions, and orchestrated updates.*

---

## References

- **Reference Implementation**: 
  - `genesyscloud/user/resource_genesyscloud_user.go`
  - `genesyscloud/user/resource_genesyscloud_user_utils.go`
  - `genesyscloud/user/data_source_genesyscloud_user.go`
- **Stage 2 Requirements**: `prompts/pf_complex_resource_migration/Stage2/requirements.md`
- **Stage 2 Design**: `prompts/pf_complex_resource_migration/Stage2/design.md`
- **Plugin Framework Documentation**: https://developer.hashicorp.com/terraform/plugin/framework
