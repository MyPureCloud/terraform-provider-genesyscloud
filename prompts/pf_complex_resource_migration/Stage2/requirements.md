# Stage 2 – Resource Migration Requirements

## Overview

Stage 2 focuses on migrating resource implementation from Terraform Plugin SDKv2 to the Terraform Plugin Framework for complex resources. This stage implements the CRUD (Create, Read, Update, Delete) lifecycle operations, data source read logic, and complex helper functions for nested structures using the schema definitions created in Stage 1.

**Reference Implementation**: 
- `genesyscloud/user/resource_genesyscloud_user.go`
- `genesyscloud/user/resource_genesyscloud_user_utils.go`
- `genesyscloud/user/data_source_genesyscloud_user.go`

**Resource Complexity**: Complex resource with multiple nested blocks, flatten/build functions for 3-level nesting, and element type helper usage.

---

## Objectives

### Primary Goal
Implement Plugin Framework resource and data source logic for complex resources while reusing existing business logic from proxy methods without modification.

### Specific Objectives
1. Create Framework resource struct and complex model with nested structures
2. Implement CRUD lifecycle methods with complex attribute handling
3. Create comprehensive helper functions for flatten/build operations
4. Implement element type helper usage for type-safe conversions
5. Create Framework data source struct and model
6. Implement data source read logic with retry patterns
7. Implement GetAll functions for export support (both Framework and SDK versions)
8. Handle complex nested structures (1-level, 2-level, 3-level nesting)
9. Ensure backward compatibility with existing behavior

---

## Scope

### In Scope for Stage 2

#### 1. Resource Implementation File
- Create `resource_genesyscloud_<resource_name>.go` file
- Implement Framework resource struct
- Implement Framework resource model with nested structure models
- Implement all CRUD methods with complex attribute handling
- Implement ImportState method
- Implement helper functions for complex model conversion

#### 2. Utils File (for Complex Resources)
- Create `resource_genesyscloud_<resource_name>_utils.go` file
- Implement flatten functions for nested structures
- Implement build functions for nested structures
- Implement element type helper functions (if not in schema file)
- Implement utility functions for complex operations
- Implement retry wrappers and consistency helpers

#### 3. CRUD Operations (Complex)
- **Create**: Handle nested structures, multiple API calls, restore deleted resources
- **Read**: Fetch all nested data, handle eventual consistency, extensive logging
- **Update**: Compare plan vs state, update nested structures, handle ordering
- **Delete**: Call proxy delete, verify deletion with retry
- **ImportState**: Enable resource import by ID with full nested structure

#### 4. Data Source Implementation File
- Create `data_source_genesyscloud_<resource_name>.go` file
- Implement Framework data source struct
- Implement Framework data source model
- Implement Read method with retry logic for eventual consistency
- Support multiple lookup criteria (email, name, etc.)

#### 5. GetAll Functions
- **GetAll<ResourceName>**: Framework version with Plugin Framework diagnostics (Phase 2 future)
- **GetAll<ResourceName>SDK**: SDK version with SDK diagnostics (Phase 1 current, used by exporter)
- Implement lazy fetch pattern for performance optimization
- Include flat attribute map building for export

#### 6. Helper Functions (Complex)
- **Flatten Functions**: Convert SDK types to Framework types
  - `flattenUserSkills()`, `flattenUserLanguages()`, `flattenLocations()`
  - `flattenAddresses()`, `flattenPhoneNumbers()`, `flattenOtherEmails()`
  - `flattenRoutingUtilization()`, `flattenMediaUtilization()`
  - `flattenVoicemailUserpolicies()`, `flattenEmployerInfo()`
- **Build Functions**: Convert Framework types to SDK types
  - `buildSdkUserSkills()`, `buildSdkUserLanguages()`, `buildSdkLocations()`
  - `buildSdkAddresses()`, `buildSdkPhoneNumbers()`, `buildSdkOtherEmails()`
  - `buildSdkRoutingUtilization()`, `buildSdkMediaUtilization()`
  - `buildSdkVoicemailUserpolicies()`, `buildSdkEmployerInfo()`
- **Utility Functions**:
  - `hasChanges()`: Detect which attributes changed
  - `executeAllUpdates()`: Orchestrate multiple update operations
  - `readUser()`: Shared read logic across CRUD operations
  - Element type helpers (if not in schema file)

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

#### FR2: Resource Model Implementation (Complex)
- ✅ Framework resource model struct is defined with all attributes
- ✅ Model uses Framework types (`types.String`, `types.Int64`, `types.Set`, `types.List`)
- ✅ Model fields have `tfsdk` struct tags matching schema attribute names
- ✅ Nested structure models are defined (e.g., `VoicemailUserpoliciesModel`, `RoutingUtilizationModel`)
- ✅ All nested models use Framework types
- ✅ Model structure matches schema hierarchy exactly

#### FR3: CRUD Operations (Complex)
- ✅ **Create**: 
  - Handles nested structures correctly
  - Checks for deleted resources (restore pattern)
  - Makes multiple API calls in correct order
  - Updates attributes that require separate PATCH
  - Applies routing skills, languages, utilization
  - Reads final state to ensure consistency
- ✅ **Read**: 
  - Fetches all nested data with expands
  - Handles 404 gracefully
  - Implements retry logic for eventual consistency
  - Extensive logging for troubleshooting
  - Flattens all nested structures correctly
- ✅ **Update**: 
  - Compares plan vs state to detect changes
  - Handles state transitions separately
  - Updates nested structures correctly
  - Maintains correct update ordering
  - Reads final state to verify changes
- ✅ **Delete**: 
  - Deletes resource via proxy
  - Verifies deletion with retry
- ✅ All operations use context for cancellation
- ✅ All operations handle errors with clear diagnostics

#### FR4: ImportState
- ✅ ImportState method is implemented
- ✅ Uses `resource.ImportStatePassthroughID` pattern
- ✅ Imports by resource ID
- ✅ Reads all nested structures during import

#### FR5: Data Source Implementation
- ✅ Framework data source struct implements required interfaces:
  - `datasource.DataSource`
  - `datasource.DataSourceWithConfigure`
- ✅ Data source model struct is defined
- ✅ Read method implements lookup with retry logic
- ✅ Supports multiple lookup criteria (email, name, etc.)
- ✅ Constructor function `New<ResourceName>FrameworkDataSource()` is implemented

#### FR6: GetAll Functions
- ✅ `GetAll<ResourceName>()` implemented with Framework diagnostics (Phase 2 future)
- ✅ `GetAll<ResourceName>SDK()` implemented with SDK diagnostics (Phase 1 current)
- ✅ Both functions return `resourceExporter.ResourceIDMetaMap`
- ✅ SDK version includes flat attribute map for export
- ✅ Lazy fetch pattern implemented for performance

#### FR7: Flatten Functions (Complex)
- ✅ Flatten functions for all nested structures:
  - 1-level nested blocks (skills, languages, locations)
  - 2-level nested blocks (addresses → phone_numbers, other_emails)
  - 3-level nested blocks (routing_utilization → media types → attributes)
- ✅ Functions convert SDK types to Framework types correctly
- ✅ Functions use element type helpers for type safety
- ✅ Functions handle null/empty values correctly
- ✅ Functions preserve all attribute values

#### FR8: Build Functions (Complex)
- ✅ Build functions for all nested structures:
  - 1-level nested blocks
  - 2-level nested blocks
  - 3-level nested blocks
- ✅ Functions convert Framework types to SDK types correctly
- ✅ Functions handle null/unknown values correctly
- ✅ Functions use pointer methods appropriately
- ✅ Functions validate data before API calls

#### FR9: Utility Functions
- ✅ `hasChanges()` detects attribute changes correctly
- ✅ `executeAllUpdates()` orchestrates updates in correct order
- ✅ `readUser()` shared read logic works across CRUD operations
- ✅ Element type helpers match schema definitions exactly
- ✅ Retry wrappers handle eventual consistency

#### FR10: Behavior Preservation
- ✅ All CRUD operations behave identically to SDKv2 version
- ✅ Error handling matches SDKv2 patterns
- ✅ Retry logic matches SDKv2 patterns
- ✅ API calls use same proxy methods as SDKv2
- ✅ Nested structure handling preserves all data

### Non-Functional Requirements

#### NFR1: Code Quality
- ✅ Code follows Go best practices
- ✅ Code follows existing codebase conventions
- ✅ Proper error handling with clear messages
- ✅ Logging at appropriate levels (extensive for complex operations)
- ✅ No unused imports or variables
- ✅ Complex logic is well-commented

#### NFR2: Documentation
- ✅ All functions have clear comments
- ✅ Complex logic is explained with inline comments
- ✅ Phase 1/Phase 2 temporary code is marked with TODO comments
- ✅ Export-related code includes migration notes
- ✅ Nested structure handling is documented
- ✅ Element type usage is documented

#### NFR3: Type Safety
- ✅ Use Framework types (`types.String`, `types.Set`, `types.List`) instead of pointers
- ✅ Proper null/unknown value handling
- ✅ Type conversions are explicit and safe
- ✅ Element type helpers ensure type consistency
- ✅ No type assertion panics

#### NFR4: Performance
- ✅ Lazy fetch pattern for GetAll functions
- ✅ Efficient API calls (no unnecessary requests)
- ✅ Proper use of context for cancellation
- ✅ Batch operations where possible (skills, languages)
- ✅ Minimal memory allocations in hot paths

#### NFR5: Maintainability
- ✅ Flatten/build functions are paired and consistent
- ✅ Element type helpers are reused across functions
- ✅ Utility functions reduce code duplication
- ✅ Clear separation between resource and utils files
- ✅ Consistent naming conventions

---

## Dependencies and Prerequisites

### Prerequisites

#### 1. Stage 1 Completion
- Schema file must be complete and approved
- Element type helpers must be defined (in schema or utils file)
- `SetRegistrar()` function references constructor functions created in Stage 2
- Exporter configuration references GetAll functions created in Stage 2

#### 2. Understanding of Framework Patterns
- Familiarity with Framework resource lifecycle
- Understanding of Framework types and diagnostics
- Knowledge of context usage in Framework
- Understanding of nested attribute handling
- Knowledge of element type usage

#### 3. Reference Implementation
- Study `user` resource implementation
- Understand CRUD patterns for complex resources
- Review flatten/build function implementations
- Understand element type helper usage
- Review nested structure handling

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

#### 2. Package Imports (Utils File)
```go
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

#### 3. Package Imports (Data Source File)
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

#### 4. Proxy Package
- Proxy package must exist with all required methods
- Proxy methods are NOT modified during migration
- Proxy provides: create, read, update, delete, getAll, getByName methods
- Proxy may provide specialized methods (e.g., updateUserSkills, updateUserLanguages)

#### 5. Utility Functions
- `util.IsStatus404()` for 404 detection
- `util.QuickHashFields()` for export hash calculation
- `util.PFWithRetriesForRead()` for retry logic with Framework diagnostics
- `retry.RetryContext()` for retry logic
- `chunksProcess` for batch operations
- `lists` for list utilities

#### 6. Element Type Helpers
- Must be defined in schema file or utils file
- Used in flatten/build functions for type-safe conversions
- Examples: `routingSkillsElementType()`, `routingLanguagesElementType()`

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
- **Constraint**: Must use Framework types (`types.String`, `types.Set`, `types.List`) not pointers
- **Rationale**: Framework requires specific types for state management
- **Impact**: Model structs use Framework types, conversion needed for API calls

#### TC5: Element Type Consistency
- **Constraint**: Element type helpers MUST match schema definitions exactly
- **Rationale**: Type mismatches cause runtime errors
- **Impact**: Flatten/build functions must use correct element types

#### TC6: Nested Structure Handling
- **Constraint**: Must preserve all nested structure data during conversions
- **Rationale**: Data loss breaks user configurations
- **Impact**: Flatten/build functions must handle all attributes correctly

### Process Constraints

#### PC1: Stage Isolation
- **Constraint**: Stage 2 MUST NOT include test implementation
- **Rationale**: Clear separation of concerns for easier review
- **Impact**: Tests are deferred to Stage 3

#### PC2: Export Compatibility
- **Constraint**: GetAll functions must support both Framework and SDK patterns
- **Rationale**: Exporter currently uses SDK diagnostics
- **Impact**: Two GetAll functions needed (Framework and SDK versions)

#### PC3: File Organization
- **Constraint**: Complex resources should separate utils into dedicated file
- **Rationale**: Improves maintainability and readability
- **Impact**: Create `resource_genesyscloud_<resource_name>_utils.go` for helper functions

---

## Complex Resource Patterns

### Pattern 1: Nested Structure Models

**Purpose**: Define separate model structs for each nested block level.

**Pattern**:
```go
// Main resource model
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

// Nested structure models
type VoicemailUserpoliciesModel struct {
    AlertTimeoutSeconds    types.Int64 `tfsdk:"alert_timeout_seconds"`
    SendEmailNotifications types.Bool  `tfsdk:"send_email_notifications"`
}

type RoutingUtilizationModel struct {
    Call              types.List `tfsdk:"call"`
    Callback          types.List `tfsdk:"callback"`
    Message           types.List `tfsdk:"message"`
    Email             types.List `tfsdk:"email"`
    Chat              types.List `tfsdk:"chat"`
    LabelUtilizations types.List `tfsdk:"label_utilizations"`
}

type MediaUtilizationModel struct {
    MaximumCapacity         types.Int64 `tfsdk:"maximum_capacity"`
    IncludeNonAcd           types.Bool  `tfsdk:"include_non_acd"`
    InterruptibleMediaTypes types.Set   `tfsdk:"interruptible_media_types"`
}
```

**Key Points**:
- One model struct per nested block level
- All use Framework types
- All have `tfsdk` struct tags
- Nested models referenced by parent model

---

### Pattern 2: Flatten Functions for Nested Structures

**Purpose**: Convert SDK types to Framework types for nested structures.

**Pattern** (1-level nesting):
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

**Pattern** (2-level nesting):
```go
func flattenAddresses(ctx context.Context, addresses *[]platformclientv2.Contact) (types.List, pfdiag.Diagnostics) {
    var diags pfdiag.Diagnostics
    
    if addresses == nil || len(*addresses) == 0 {
        return types.ListNull(addressesElementType()), diags
    }
    
    // Flatten phone_numbers
    phoneNumbers, phoneDiags := flattenPhoneNumbers(ctx, addresses)
    diags.Append(phoneDiags...)
    
    // Flatten other_emails
    otherEmails, emailDiags := flattenOtherEmails(ctx, addresses)
    diags.Append(emailDiags...)
    
    // Build addresses object
    addressObj, objDiags := types.ObjectValue(
        addressesElementType().AttrTypes,
        map[string]attr.Value{
            "phone_numbers": phoneNumbers,
            "other_emails":  otherEmails,
        },
    )
    diags.Append(objDiags...)
    
    // Wrap in list (MaxItems: 1)
    addressList, listDiags := types.ListValue(addressesElementType(), []attr.Value{addressObj})
    diags.Append(listDiags...)
    
    return addressList, diags
}
```

**Pattern** (3-level nesting):
```go
func flattenRoutingUtilization(ctx context.Context, utilization *agentUtilizationWithLabels) (types.List, pfdiag.Diagnostics) {
    var diags pfdiag.Diagnostics
    
    if utilization == nil {
        return types.ListNull(routingUtilizationElementType()), diags
    }
    
    // Flatten each media type
    call, callDiags := flattenMediaUtilization(ctx, utilization.Utilization["call"])
    diags.Append(callDiags...)
    
    callback, callbackDiags := flattenMediaUtilization(ctx, utilization.Utilization["callback"])
    diags.Append(callbackDiags...)
    
    // ... other media types
    
    // Build routing_utilization object
    utilizationObj, objDiags := types.ObjectValue(
        routingUtilizationElementType().AttrTypes,
        map[string]attr.Value{
            "call":     call,
            "callback": callback,
            // ... other media types
        },
    )
    diags.Append(objDiags...)
    
    // Wrap in list (MaxItems: 1)
    utilizationList, listDiags := types.ListValue(routingUtilizationElementType(), []attr.Value{utilizationObj})
    diags.Append(listDiags...)
    
    return utilizationList, diags
}
```

**Key Points**:
- Use element type helpers for type safety
- Handle null/empty cases
- Return diagnostics for error handling
- Preserve all attribute values
- Use `types.SetValue()`, `types.ListValue()`, `types.ObjectValue()`

---

### Pattern 3: Build Functions for Nested Structures

**Purpose**: Convert Framework types to SDK types for API calls.

**Pattern** (1-level nesting):
```go
func buildSdkUserSkills(ctx context.Context, skills types.Set) []platformclientv2.Userroutingskill {
    if skills.IsNull() || skills.IsUnknown() {
        return nil
    }
    
    var skillObjs []types.Object
    skills.ElementsAs(ctx, &skillObjs, false)
    
    sdkSkills := make([]platformclientv2.Userroutingskill, 0, len(skillObjs))
    for _, skillObj := range skillObjs {
        var skillId types.String
        var proficiency types.Float64
        
        skillObj.As(ctx, &struct {
            SkillId    types.String  `tfsdk:"skill_id"`
            Proficiency types.Float64 `tfsdk:"proficiency"`
        }{
            SkillId:    skillId,
            Proficiency: proficiency,
        }, basetypes.ObjectAsOptions{})
        
        sdkSkills = append(sdkSkills, platformclientv2.Userroutingskill{
            Id:          skillId.ValueStringPointer(),
            Proficiency: proficiency.ValueFloat64Pointer(),
        })
    }
    
    return sdkSkills
}
```

**Pattern** (2-level nesting):
```go
func buildSdkAddresses(ctx context.Context, addresses types.List) (*[]platformclientv2.Contact, pfdiag.Diagnostics) {
    var diags pfdiag.Diagnostics
    
    if addresses.IsNull() || addresses.IsUnknown() {
        return nil, diags
    }
    
    var addressObjs []types.Object
    addresses.ElementsAs(ctx, &addressObjs, false)
    
    if len(addressObjs) == 0 {
        return nil, diags
    }
    
    // Extract nested blocks
    var phoneNumbers types.Set
    var otherEmails types.Set
    
    addressObjs[0].As(ctx, &struct {
        PhoneNumbers types.Set `tfsdk:"phone_numbers"`
        OtherEmails  types.Set `tfsdk:"other_emails"`
    }{
        PhoneNumbers: phoneNumbers,
        OtherEmails:  otherEmails,
    }, basetypes.ObjectAsOptions{})
    
    // Build SDK contacts
    contacts := make([]platformclientv2.Contact, 0)
    
    // Build phone numbers
    if !phoneNumbers.IsNull() && !phoneNumbers.IsUnknown() {
        phoneContacts := buildSdkPhoneNumbers(ctx, phoneNumbers)
        contacts = append(contacts, phoneContacts...)
    }
    
    // Build other emails
    if !otherEmails.IsNull() && !otherEmails.IsUnknown() {
        emailContacts := buildSdkOtherEmails(ctx, otherEmails)
        contacts = append(contacts, emailContacts...)
    }
    
    return &contacts, diags
}
```

**Key Points**:
- Use `ElementsAs()` to extract Framework types
- Use `As()` to extract nested attributes
- Handle null/unknown values
- Use pointer methods for SDK types (`ValueStringPointer()`)
- Return SDK types ready for API calls

---

### Pattern 4: Shared Read Logic

**Purpose**: Centralize read logic to avoid duplication across CRUD operations.

**Pattern**:
```go
func readUser(ctx context.Context, model *UserFrameworkResourceModel, proxy *userProxy, diagnostics *pfdiag.Diagnostics, isImport ...bool) {
    log.Printf("Reading user %s", model.Id.ValueString())
    
    // Determine if this is an import operation
    importMode := len(isImport) > 0 && isImport[0]
    
    retryDiags := util.PFWithRetriesForRead(ctx, func() (bool, error) {
        // Fetch user from API
        currentUser, proxyResponse, getUserErr := proxy.getUserById(ctx, model.Id.ValueString(), []string{
            "skills", "languages", "locations", "profileSkills", "certifications", "employerInfo",
        }, "")
        
        if getUserErr != nil {
            if util.IsStatus404(proxyResponse) {
                return true, fmt.Errorf("API Error: 404")
            }
            return false, fmt.Errorf("Failed to read user %s | error: %s", model.Id.ValueString(), getUserErr)
        }
        
        // Update model from API response
        model.Email = types.StringPointerValue(currentUser.Email)
        model.Name = types.StringPointerValue(currentUser.Name)
        model.State = types.StringPointerValue(currentUser.State)
        // ... other attributes
        
        // Flatten nested structures
        model.RoutingSkills, _ = flattenUserSkills(ctx, currentUser.Skills)
        model.RoutingLanguages, _ = flattenUserLanguages(ctx, currentUser.Languages)
        model.Addresses, _ = flattenAddresses(ctx, currentUser.Addresses)
        // ... other nested structures
        
        return false, nil
    })
    
    diagnostics.Append(retryDiags...)
}
```

**Key Points**:
- Used by Create, Read, Update methods
- Implements retry logic for eventual consistency
- Handles 404 errors
- Flattens all nested structures
- Updates model in place

---

### Pattern 5: Complex Update Orchestration

**Purpose**: Coordinate multiple update operations in correct order.

**Pattern**:
```go
func executeAllUpdates(ctx context.Context, plan *UserFrameworkResourceModel, proxy *userProxy, clientConfig *platformclientv2.Configuration, diagnostics *pfdiag.Diagnostics) pfdiag.Diagnostics {
    var diags pfdiag.Diagnostics
    
    // Update routing skills
    if !plan.RoutingSkills.IsNull() && !plan.RoutingSkills.IsUnknown() {
        skills := buildSdkUserSkills(ctx, plan.RoutingSkills)
        _, err := proxy.updateUserSkills(ctx, plan.Id.ValueString(), skills)
        if err != nil {
            diags.AddError("Failed to update user skills", err.Error())
            return diags
        }
    }
    
    // Update routing languages
    if !plan.RoutingLanguages.IsNull() && !plan.RoutingLanguages.IsUnknown() {
        languages := buildSdkUserLanguages(ctx, plan.RoutingLanguages)
        _, err := proxy.updateUserLanguages(ctx, plan.Id.ValueString(), languages)
        if err != nil {
            diags.AddError("Failed to update user languages", err.Error())
            return diags
        }
    }
    
    // Update routing utilization
    if !plan.RoutingUtilization.IsNull() && !plan.RoutingUtilization.IsUnknown() {
        utilization := buildSdkRoutingUtilization(ctx, plan.RoutingUtilization)
        _, err := proxy.updateUserUtilization(ctx, plan.Id.ValueString(), utilization)
        if err != nil {
            diags.AddError("Failed to update user utilization", err.Error())
            return diags
        }
    }
    
    return diags
}
```

**Key Points**:
- Orchestrates multiple API calls
- Maintains correct ordering
- Handles errors gracefully
- Returns diagnostics for error reporting

---

## Validation Checklist

Use this checklist to verify Stage 2 completion:

### Resource Implementation File
- [ ] File created: `resource_genesyscloud_<resource_name>.go`
- [ ] Package declaration matches directory name
- [ ] All required imports are present
- [ ] No unused imports

### Utils File (Complex Resources)
- [ ] File created: `resource_genesyscloud_<resource_name>_utils.go`
- [ ] Package declaration matches directory name
- [ ] All required imports are present
- [ ] Element type helpers defined (if not in schema file)

### Resource Struct and Model
- [ ] Resource struct implements required interfaces
- [ ] Resource struct has `clientConfig` field
- [ ] Resource model struct defined with all attributes
- [ ] Nested structure models defined
- [ ] All models use Framework types
- [ ] All models have correct `tfsdk` struct tags
- [ ] Constructor function implemented

### CRUD Methods
- [ ] Metadata() method implemented
- [ ] Schema() method implemented (calls schema function from Stage 1)
- [ ] Configure() method implemented
- [ ] Create() method implemented with complex handling
- [ ] Read() method implemented with retry logic
- [ ] Update() method implemented with change detection
- [ ] Delete() method implemented with verification
- [ ] ImportState() method implemented

### CRUD Method Behavior (Complex)
- [ ] Create: Handles nested structures, multiple API calls, restore pattern
- [ ] Read: Fetches all nested data, handles 404, extensive logging
- [ ] Update: Compares plan vs state, updates nested structures, correct ordering
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
- [ ] Supports multiple lookup criteria (if applicable)

### GetAll Functions
- [ ] `GetAll<ResourceName>()` implemented (Framework version)
- [ ] `GetAll<ResourceName>SDK()` implemented (SDK version)
- [ ] Both return `resourceExporter.ResourceIDMetaMap`
- [ ] SDK version includes flat attribute map
- [ ] Functions include Phase 1/Phase 2 comments
- [ ] Lazy fetch pattern implemented (if applicable)

### Flatten Functions
- [ ] Flatten functions for all 1-level nested blocks
- [ ] Flatten functions for all 2-level nested blocks (if applicable)
- [ ] Flatten functions for all 3-level nested blocks (if applicable)
- [ ] Functions use element type helpers
- [ ] Functions handle null/empty values
- [ ] Functions return diagnostics
- [ ] Functions preserve all attribute values

### Build Functions
- [ ] Build functions for all 1-level nested blocks
- [ ] Build functions for all 2-level nested blocks (if applicable)
- [ ] Build functions for all 3-level nested blocks (if applicable)
- [ ] Functions use element type helpers
- [ ] Functions handle null/unknown values
- [ ] Functions return SDK types
- [ ] Functions validate data

### Utility Functions
- [ ] `hasChanges()` implemented (if needed)
- [ ] `executeAllUpdates()` implemented (if needed)
- [ ] `readUser()` shared read logic implemented
- [ ] Element type helpers match schema exactly
- [ ] Retry wrappers implemented (if needed)

### Code Quality
- [ ] Code compiles without errors
- [ ] Code follows Go conventions
- [ ] Functions have clear comments
- [ ] Error messages are clear and actionable
- [ ] Logging is appropriate and extensive
- [ ] Complex logic is well-documented

---

## Example: user Resource Migration

### File Structure
```
genesyscloud/user/
├── resource_genesyscloud_user_schema.go  (Stage 1)
├── resource_genesyscloud_user.go         (Stage 2 - THIS)
├── resource_genesyscloud_user_utils.go   (Stage 2 - THIS)
└── data_source_genesyscloud_user.go      (Stage 2 - THIS)
```

### Key Components

#### 1. Resource Struct
```go
type UserFrameworkResource struct {
    clientConfig *platformclientv2.Configuration
}
```

#### 2. Resource Model (Complex)
```go
type UserFrameworkResourceModel struct {
    Id                    types.String `tfsdk:"id"`
    Email                 types.String `tfsdk:"email"`
    Name                  types.String `tfsdk:"name"`
    Password              types.String `tfsdk:"password"`
    State                 types.String `tfsdk:"state"`
    DivisionId            types.String `tfsdk:"division_id"`
    Department            types.String `tfsdk:"department"`
    Title                 types.String `tfsdk:"title"`
    Manager               types.String `tfsdk:"manager"`
    AcdAutoAnswer         types.Bool   `tfsdk:"acd_auto_answer"`
    RoutingSkills         types.Set    `tfsdk:"routing_skills"`
    RoutingLanguages      types.Set    `tfsdk:"routing_languages"`
    Locations             types.Set    `tfsdk:"locations"`
    Addresses             types.List   `tfsdk:"addresses"`
    ProfileSkills         types.Set    `tfsdk:"profile_skills"`
    Certifications        types.Set    `tfsdk:"certifications"`
    EmployerInfo          types.List   `tfsdk:"employer_info"`
    RoutingUtilization    types.List   `tfsdk:"routing_utilization"`
    VoicemailUserpolicies types.List   `tfsdk:"voicemail_userpolicies"`
}

// Nested structure models
type VoicemailUserpoliciesModel struct {
    AlertTimeoutSeconds    types.Int64 `tfsdk:"alert_timeout_seconds"`
    SendEmailNotifications types.Bool  `tfsdk:"send_email_notifications"`
}

type RoutingUtilizationModel struct {
    Call              types.List `tfsdk:"call"`
    Callback          types.List `tfsdk:"callback"`
    Message           types.List `tfsdk:"message"`
    Email             types.List `tfsdk:"email"`
    Chat              types.List `tfsdk:"chat"`
    LabelUtilizations types.List `tfsdk:"label_utilizations"`
}

type MediaUtilizationModel struct {
    MaximumCapacity         types.Int64 `tfsdk:"maximum_capacity"`
    IncludeNonAcd           types.Bool  `tfsdk:"include_non_acd"`
    InterruptibleMediaTypes types.Set   `tfsdk:"interruptible_media_types"`
}
```

#### 3. Constructor Function
```go
func NewUserFrameworkResource() resource.Resource {
    return &UserFrameworkResource{}
}
```

#### 4. Create Method Pattern (Complex)
```go
func (r *UserFrameworkResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
    var plan UserFrameworkResourceModel
    resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
    if resp.Diagnostics.HasError() {
        return
    }

    proxy := GetUserProxy(r.clientConfig)
    email := plan.Email.ValueString()

    // Build addresses from plan
    addresses, addressDiags := buildSdkAddresses(ctx, plan.Addresses)
    resp.Diagnostics.Append(addressDiags...)
    if resp.Diagnostics.HasError() {
        return
    }

    // Check for deleted user before creating
    id, diagErr := getDeletedUserId(email, proxy)
    if diagErr.HasError() {
        resp.Diagnostics.Append(diagErr...)
        return
    }

    if id != nil {
        // Found deleted user - restore and configure
        plan.Id = types.StringValue(*id)
        restoreDeletedUser(ctx, &plan, proxy, r.clientConfig, &resp.Diagnostics)
        if resp.Diagnostics.HasError() {
            return
        }
        resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
        return
    }

    // No deleted user - create new one
    createUser := platformclientv2.Createuser{
        Name:      platformclientv2.String(plan.Name.ValueString()),
        Email:     &email,
        Addresses: addresses,
    }

    userResponse, proxyPostResponse, postErr := proxy.createUser(ctx, &createUser)
    if postErr != nil {
        resp.Diagnostics.Append(util.BuildFrameworkAPIDiagnosticError(ResourceType,
            fmt.Sprintf("Failed to create user %s error: %s", email, postErr), proxyPostResponse)...)
        return
    }

    plan.Id = types.StringValue(*userResponse.Id)

    // Set attributes that can only be modified in a patch
    if hasChanges(&plan, "manager", "locations", "acd_auto_answer") {
        additionalAttrsUpdate := &platformclientv2.Updateuser{
            Manager:       platformclientv2.String(plan.Manager.ValueString()),
            AcdAutoAnswer: platformclientv2.Bool(plan.AcdAutoAnswer.ValueBool()),
            Locations:     buildSdkLocations(ctx, plan.Locations),
            Version:       userResponse.Version,
        }

        _, proxyPatchResponse, patchErr := proxy.patchUserWithState(ctx, *userResponse.Id, additionalAttrsUpdate)
        if patchErr != nil {
            resp.Diagnostics.Append(util.BuildFrameworkAPIDiagnosticError(ResourceType,
                fmt.Sprintf("Failed to update user %s error: %s", plan.Id.ValueString(), patchErr), proxyPatchResponse)...)
            return
        }
    }

    // Apply skills, languages, utilization
    frameworkDiags := executeAllUpdates(ctx, &plan, proxy, r.clientConfig, false)
    if frameworkDiags.HasError() {
        resp.Diagnostics.Append(frameworkDiags...)
        return
    }

    // Read back the created user
    readUser(ctx, &plan, proxy, &resp.Diagnostics)
    if resp.Diagnostics.HasError() {
        return
    }

    resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}
```

#### 5. Flatten Function Pattern (Complex)
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

#### 6. Build Function Pattern (Complex)
```go
func buildSdkUserSkills(ctx context.Context, skills types.Set) []platformclientv2.Userroutingskill {
    if skills.IsNull() || skills.IsUnknown() {
        return nil
    }
    
    var skillObjs []types.Object
    skills.ElementsAs(ctx, &skillObjs, false)
    
    sdkSkills := make([]platformclientv2.Userroutingskill, 0, len(skillObjs))
    for _, skillObj := range skillObjs {
        var skillId types.String
        var proficiency types.Float64
        
        skillObj.As(ctx, &struct {
            SkillId     types.String  `tfsdk:"skill_id"`
            Proficiency types.Float64 `tfsdk:"proficiency"`
        }{
            SkillId:     skillId,
            Proficiency: proficiency,
        }, basetypes.ObjectAsOptions{})
        
        sdkSkills = append(sdkSkills, platformclientv2.Userroutingskill{
            Id:          skillId.ValueStringPointer(),
            Proficiency: proficiency.ValueFloat64Pointer(),
        })
    }
    
    return sdkSkills
}
```

---

## Next Steps

After Stage 2 completion and approval:
1. Review resource implementation with team
2. Verify CRUD operations work correctly
3. Verify flatten/build functions preserve all data
4. Confirm GetAll functions are correct
5. Test complex nested structure handling
6. Proceed to **Stage 3 – Test Migration**

---

## References

- **Reference Implementation**: 
  - `genesyscloud/user/resource_genesyscloud_user.go`
  - `genesyscloud/user/resource_genesyscloud_user_utils.go`
  - `genesyscloud/user/data_source_genesyscloud_user.go`
- **Stage 1 Requirements**: `prompts/pf_complex_resource_migration/Stage1/requirements.md`
- **Stage 1 Design**: `prompts/pf_complex_resource_migration/Stage1/design.md`
- **Plugin Framework Resources**: https://developer.hashicorp.com/terraform/plugin/framework/resources
- **Plugin Framework Data Sources**: https://developer.hashicorp.com/terraform/plugin/framework/data-sources
- **Plugin Framework Types**: https://developer.hashicorp.com/terraform/plugin/framework/handling-data/types

