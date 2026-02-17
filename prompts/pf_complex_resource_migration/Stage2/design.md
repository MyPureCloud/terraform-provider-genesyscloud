# Stage 2 – Resource Migration Design

## Overview

This document describes the design decisions, architecture, and patterns used in Stage 2 of the Plugin Framework migration for **complex resources**. Stage 2 focuses on implementing resource CRUD operations, data source read logic, and comprehensive helper functions for nested structures using Plugin Framework patterns while reusing existing business logic from proxy methods.

**Reference Implementation**: 
- `genesyscloud/user/resource_genesyscloud_user.go`
- `genesyscloud/user/resource_genesyscloud_user_utils.go`
- `genesyscloud/user/data_source_genesyscloud_user.go`

**Resource Complexity**: Complex resource with multiple nested blocks (1-level, 2-level, 3-level nesting), flatten/build functions for nested structures, element type helper usage, and orchestrated update operations.

---

## Design Principles

### 1. Framework-Native Lifecycle
**Principle**: Use Plugin Framework lifecycle patterns instead of SDKv2 callback functions.

**Rationale**:
- Framework provides better type safety and error handling
- Context-based cancellation support
- Clearer separation between plan and state
- Better diagnostics and error reporting
- Explicit handling of nested structures

**Implementation**:
- Resource methods: `Create()`, `Read()`, `Update()`, `Delete()`, `ImportState()`
- Data source methods: `Read()`
- All methods use context for cancellation
- All methods use Framework diagnostics
- Nested structure models with Framework types

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

### 3. Nested Structure Handling
**Principle**: Use separate model structs and dedicated flatten/build functions for each nested block level.

**Rationale**:
- Complex resources have multiple levels of nesting (1-level, 2-level, 3-level)
- Each nested level requires type-safe conversion
- Flatten/build functions ensure data preservation
- Element type helpers provide compile-time type safety

**Implementation**:
- Define nested structure models (e.g., `VoicemailUserpoliciesModel`, `RoutingUtilizationModel`)
- Create flatten functions for SDK → Framework conversion
- Create build functions for Framework → SDK conversion
- Use element type helpers for type-safe operations
- Handle null/unknown values at each nesting level

### 4. Shared Read Logic Pattern
**Principle**: Centralize read logic in a helper function to avoid duplication across CRUD operations.

**Rationale**:
- Create, Read, and Update all need to fetch current state
- Retry logic for eventual consistency should be consistent
- Reduces code duplication and maintenance burden
- Ensures consistent state handling

**Implementation**:
- `readUser()` helper function called from Create, Read, Update
- Implements retry logic with Framework diagnostics
- Handles 404 errors gracefully
- Flattens all nested structures
- Updates model in place

### 5. Complex Update Orchestration
**Principle**: Coordinate multiple update operations in correct order with proper error handling.

**Rationale**:
- Complex resources require multiple API calls for updates
- Some updates must happen in specific order
- Each update may fail independently
- Need to track which attributes are managed vs unmanaged

**Implementation**:
- `executeAllUpdates()` orchestrates all update operations
- Handles division updates, skills, languages, utilization, voicemail, password
- Detects managed vs unmanaged attributes
- Returns diagnostics for error reporting
- Maintains correct update ordering

### 6. Type Safety with Framework Types
**Principle**: Use Framework types (`types.String`) instead of pointers for state management.

**Rationale**:
- Framework requires specific types for null/unknown handling
- Better type safety at compile time
- Explicit null/unknown state representation
- Clearer intent in code
- Nested structures require consistent type usage

**Implementation**:
- Model structs use `types.String`, `types.Int64`, `types.Set`, `types.List`
- Nested models use Framework types consistently
- Helper functions convert between Framework types and API types
- Explicit null/unknown checks before API calls
- Element type helpers ensure type consistency

### 7. Dual GetAll Functions (Phase 1 Temporary)
**Principle**: Provide both Framework and SDK versions of GetAll functions during migration.

**Rationale**:
- Exporter currently uses SDK diagnostics
- Framework version prepared for Phase 2 future
- Smooth transition path
- No breaking changes during migration
- Lazy fetch pattern for performance

**Implementation**:
- `GetAll<ResourceName>()`: Framework version (Phase 2 future)
- `GetAll<ResourceName>SDK()`: SDK version (Phase 1 current, used by exporter)
- SDK version implements lazy fetch pattern with `LazyFetchAttributes` callback
- Both return same data structure
- Clear comments marking temporary code

---

## Architecture

### File Structure

```
genesyscloud/<resource_name>/
├── resource_genesyscloud_<resource_name>_schema.go          ← Stage 1
├── resource_genesyscloud_<resource_name>.go                 ← Stage 2 (THIS FILE)
├── resource_genesyscloud_<resource_name>_utils.go           ← Stage 2 (THIS FILE - Complex)
├── data_source_genesyscloud_<resource_name>.go              ← Stage 2 (THIS FILE)
├── resource_genesyscloud_<resource_name>_test.go            ← Stage 3
├── data_source_genesyscloud_<resource_name>_test.go         ← Stage 3
├── resource_genesyscloud_<resource_name>_export_utils.go    ← Stage 4
├── genesyscloud_<resource_name>_init_test.go                ← Stage 3
└── genesyscloud_<resource_name>_proxy.go                    ← NOT MODIFIED
```

**Note**: Complex resources use a separate `_utils.go` file for helper functions to improve maintainability.

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
│     - Framework types (types.String, types.List, etc.)  │
│     - tfsdk struct tags                                 │
├─────────────────────────────────────────────────────────┤
│  4. Nested Structure Models                             │
│     - VoicemailUserpoliciesModel                        │
│     - RoutingUtilizationModel                           │
│     - MediaUtilizationModel                             │
│     - LabelUtilizationModel                             │
├─────────────────────────────────────────────────────────┤
│  5. Constructor Function                                │
│     - New<ResourceName>FrameworkResource()              │
├─────────────────────────────────────────────────────────┤
│  6. Resource Interface Methods                          │
│     - Metadata()                                        │
│     - Schema()                                          │
│     - Configure()                                       │
├─────────────────────────────────────────────────────────┤
│  7. CRUD Methods (Complex)                              │
│     - Create() - handles restore, nested structures     │
│     - Read() - retry logic, extensive logging           │
│     - Update() - plan vs state, orchestration           │
│     - Delete() - soft delete with verification          │
│     - ImportState() - full nested structure import      │
├─────────────────────────────────────────────────────────┤
│  8. GetAll Functions (for export)                       │
│     - GetAll<ResourceName>() - Framework (Phase 2)      │
│     - GetAll<ResourceName>SDK() - SDK (Phase 1)         │
│     - Lazy fetch pattern with LazyFetchAttributes       │
└─────────────────────────────────────────────────────────┘
```

### Utils File Components (Complex Resources)

```
┌─────────────────────────────────────────────────────────┐
│  resource_genesyscloud_<resource_name>_utils.go         │
├─────────────────────────────────────────────────────────┤
│  1. Package-Level Variables                             │
│     - utilizationMediaTypes map                         │
│     - Internal structs (mediaUtilization, etc.)         │
├─────────────────────────────────────────────────────────┤
│  2. Element Type Helper Functions                       │
│     - routingSkillsElementType()                        │
│     - routingLanguagesElementType()                     │
│     - locationsElementType()                            │
│     - voicemailUserpoliciesElementType()                │
│     - getMediaUtilizationAttrTypes()                    │
│     - getLabelUtilizationAttrTypes()                    │
├─────────────────────────────────────────────────────────┤
│  3. Shared Read Logic                                   │
│     - readUser() - called from Create, Read, Update     │
│     - Helper functions for read logic                   │
│     - Retry logic with Framework diagnostics            │
├─────────────────────────────────────────────────────────┤
│  4. Update Orchestration                                │
│     - updateUser() - shared update logic                │
│     - executeAllUpdates() - coordinates updates         │
│     - Update functions for each attribute type          │
├─────────────────────────────────────────────────────────┤
│  5. Flatten Functions (SDK → Framework)                 │
│     - flattenUserSkills() - 1-level nesting             │
│     - flattenUserLanguages() - 1-level nesting          │
│     - flattenUserLocations() - 1-level nesting          │
│     - flattenUserAddresses() - 2-level nesting          │
│     - flattenRoutingUtilization() - 3-level nesting     │
│     - flattenVoicemailUserpolicies() - 1-level nesting  │
│     - flattenEmployerInfo() - 1-level nesting           │
├─────────────────────────────────────────────────────────┤
│  6. Build Functions (Framework → SDK)                   │
│     - buildSdkUserSkills() - 1-level nesting            │
│     - buildSdkUserLanguages() - 1-level nesting         │
│     - buildSdkLocations() - 1-level nesting             │
│     - buildSdkAddresses() - 2-level nesting             │
│     - buildSdkRoutingUtilization() - 3-level nesting    │
│     - buildSdkVoicemailUserpolicies() - 1-level nesting │
│     - buildSdkEmployerInfo() - 1-level nesting          │
├─────────────────────────────────────────────────────────┤
│  7. Utility Functions                                   │
│     - hasChanges() - detect attribute changes           │
│     - getDeletedUserId() - check for deleted resources  │
│     - restoreDeletedUser() - restore pattern            │
│     - waitForExtensionPoolActivation() - timing helper  │
│     - executeUpdateUser() - retry wrapper               │
├─────────────────────────────────────────────────────────┤
│  8. Conversion Helpers                                  │
│     - convertSDKDiagnosticsToFramework()                │
│     - Logging helpers (invMustJSON, invStr, etc.)       │
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
│     - Read() - with retry and cache                     │
├─────────────────────────────────────────────────────────┤
│  6. Helper Functions                                    │
│     - getUserByName() - lookup with retry               │
│     - hydrateUserCache() - cache population             │
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

**Example** (user):
```go
var (
    _ resource.Resource                = &UserFrameworkResource{}
    _ resource.ResourceWithConfigure   = &UserFrameworkResource{}
    _ resource.ResourceWithImportState = &UserFrameworkResource{}
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

**Example** (user):
```go
// UserFrameworkResource is the main resource struct that manages Genesys Cloud user lifecycle operations.
// It holds the API client configuration needed to communicate with the Genesys Cloud platform.
type UserFrameworkResource struct {
    clientConfig *platformclientv2.Configuration
}
```

**Design Decisions**:

| Decision | Rationale |
|----------|-----------|
| Store `clientConfig` | Needed to create proxy instances in CRUD methods |
| No other fields | Keep resource stateless; all state in Terraform state |
| Exported struct name | Public API for provider registration |

---

### 3. Resource Model Struct (Complex)

**Purpose**: Define the data structure for Terraform state and plan, including nested structures.

**Design Pattern**:
```go
// <resource>FrameworkResourceModel describes the resource data model.
type <resource>FrameworkResourceModel struct {
    Id                    types.String `tfsdk:"id"`
    Name                  types.String `tfsdk:"name"`
    // ... simple attributes
    RoutingSkills         types.Set    `tfsdk:"routing_skills"`
    RoutingLanguages      types.Set    `tfsdk:"routing_languages"`
    Addresses             types.List   `tfsdk:"addresses"`
    RoutingUtilization    types.List   `tfsdk:"routing_utilization"`
    VoicemailUserpolicies types.List   `tfsdk:"voicemail_userpolicies"`
}
```

**Example** (user):
```go
// UserFrameworkResourceModel represents the complete Terraform state for a Genesys Cloud user.
// This model maps directly to the Terraform configuration schema and is used to marshal/unmarshal
// state data during CRUD operations.
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
```

**Design Decisions**:

| Decision | Rationale |
|----------|-----------|
| Use Framework types | Required by Framework for null/unknown handling |
| `tfsdk` struct tags | Map struct fields to schema attribute names |
| Match schema exactly | Every schema attribute has corresponding model field |
| No pointers | Framework types handle null/unknown internally |
| Nested structures as types.List/Set | Complex nested blocks use collection types |

---

### 4. Nested Structure Models (Complex)

**Purpose**: Define separate model structs for each nested block level.

**Design Pattern**:
```go
// Nested structure models for complex resources
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

**Example** (user):
```go
// VoicemailUserpoliciesModel represents voicemail configuration settings for a user.
type VoicemailUserpoliciesModel struct {
    AlertTimeoutSeconds    types.Int64 `tfsdk:"alert_timeout_seconds"`
    SendEmailNotifications types.Bool  `tfsdk:"send_email_notifications"`
}

// RoutingUtilizationModel defines the capacity settings for different communication channels.
type RoutingUtilizationModel struct {
    Call              types.List `tfsdk:"call"`
    Callback          types.List `tfsdk:"callback"`
    Message           types.List `tfsdk:"message"`
    Email             types.List `tfsdk:"email"`
    Chat              types.List `tfsdk:"chat"`
    LabelUtilizations types.List `tfsdk:"label_utilizations"`
}

// MediaUtilizationModel defines capacity settings for a specific media type.
type MediaUtilizationModel struct {
    MaximumCapacity         types.Int64 `tfsdk:"maximum_capacity"`
    IncludeNonAcd           types.Bool  `tfsdk:"include_non_acd"`
    InterruptibleMediaTypes types.Set   `tfsdk:"interruptible_media_types"`
}

// LabelUtilizationModel defines capacity settings for label-based routing.
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
- Nested models referenced by parent model
- Supports 1-level, 2-level, and 3-level nesting

---

### 5. Constructor Function

**Purpose**: Create new resource instances for the provider.

**Design Pattern**:
```go
// New<ResourceName>FrameworkResource is a helper function to simplify the provider implementation.
func New<ResourceName>FrameworkResource() resource.Resource {
    return &<resource>FrameworkResource{}
}
```

**Example** (user):
```go
// NewUserFrameworkResource is a factory function that creates a new instance of the user resource.
func NewUserFrameworkResource() resource.Resource {
    return &UserFrameworkResource{}
}
```

**Why This Pattern**:
- Provider calls this function to create resource instances
- Returns interface type (`resource.Resource`) for flexibility
- Simple factory pattern
- No initialization logic needed (configuration happens in Configure method)

---

### 6. Resource Interface Methods

#### 6.1 Metadata Method

**Purpose**: Provide resource type name to the provider.

**Design Pattern**:
```go
func (r *<resource>FrameworkResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
    resp.TypeName = req.ProviderTypeName + "_<resource_name>"
}
```

**Example** (user):
```go
// Metadata sets the resource type name that will be used in Terraform configurations.
func (r *UserFrameworkResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
    resp.TypeName = req.ProviderTypeName + "_user"
}
```

**Key Points**:
- Concatenates provider name with resource name
- Results in full type: `genesyscloud_user`
- Must match ResourceType constant from schema file

#### 6.2 Schema Method

**Purpose**: Provide resource schema to the provider.

**Design Pattern**:
```go
func (r *<resource>FrameworkResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
    resp.Schema = <ResourceName>ResourceSchema()
}
```

**Example** (user):
```go
// Schema defines the complete resource schema including all attributes, their types, and validation rules.
func (r *UserFrameworkResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
    resp.Schema = UserResourceSchema()
}
```

**Key Points**:
- Calls schema function from Stage 1
- No schema logic in resource file
- Keeps schema and implementation separate

#### 6.3 Configure Method

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

### 7. CRUD Methods (Complex)

#### 7.1 Create Method (Complex)

**Purpose**: Create a new resource via API, handle nested structures, and store in Terraform state.

**Design Pattern** (Complex):
```go
func (r *<resource>FrameworkResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
    var plan <resource>FrameworkResourceModel

    // Read Terraform plan data into the model
    resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
    if resp.Diagnostics.HasError() {
        return
    }

    proxy := Get<ResourceName>Proxy(r.clientConfig)
    email := plan.Email.ValueString()

    // Build nested structures from plan
    addresses, addressDiags := buildSdkAddresses(ctx, plan.Addresses)
    resp.Diagnostics.Append(addressDiags...)
    if resp.Diagnostics.HasError() {
        return
    }

    // Check for deleted resource before creating (restore pattern)
    id, diagErr := getDeletedUserId(email, proxy)
    if diagErr.HasError() {
        resp.Diagnostics.Append(diagErr...)
        return
    }

    if id != nil {
        // Found deleted resource - restore and configure
        plan.Id = types.StringValue(*id)
        restoreDeletedUser(ctx, &plan, proxy, r.clientConfig, &resp.Diagnostics)
        if resp.Diagnostics.HasError() {
            return
        }
        resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
        return
    }

    // No deleted resource - create new one
    createRequest := platformclientv2.Createuser{
        Name:      platformclientv2.String(plan.Name.ValueString()),
        Email:     &email,
        Addresses: addresses,
    }

    apiResponse, proxyPostResponse, postErr := proxy.createUser(ctx, &createRequest)
    if postErr != nil {
        resp.Diagnostics.Append(util.BuildFrameworkAPIDiagnosticError(ResourceType,
            fmt.Sprintf("Failed to create user %s error: %s", email, postErr), proxyPostResponse)...)
        return
    }

    plan.Id = types.StringValue(*apiResponse.Id)

    // Set attributes that can only be modified in a patch
    if hasChanges(&plan, "manager", "locations", "acd_auto_answer") {
        additionalAttrsUpdate := &platformclientv2.Updateuser{
            Manager:       platformclientv2.String(plan.Manager.ValueString()),
            AcdAutoAnswer: platformclientv2.Bool(plan.AcdAutoAnswer.ValueBool()),
            Locations:     buildSdkLocations(ctx, plan.Locations),
            Version:       apiResponse.Version,
        }

        _, proxyPatchResponse, patchErr := proxy.patchUserWithState(ctx, *apiResponse.Id, additionalAttrsUpdate)
        if patchErr != nil {
            resp.Diagnostics.Append(util.BuildFrameworkAPIDiagnosticError(ResourceType,
                fmt.Sprintf("Failed to update user %s error: %s", plan.Id.ValueString(), patchErr), proxyPatchResponse)...)
            return
        }
    }

    // Apply skills, languages, utilization (orchestrated updates)
    frameworkDiags := executeAllUpdates(ctx, &plan, proxy, r.clientConfig, false)
    if frameworkDiags.HasError() {
        resp.Diagnostics.Append(frameworkDiags...)
        return
    }

    // Read back the created resource to populate state with server-generated values
    readUser(ctx, &plan, proxy, &resp.Diagnostics)
    if resp.Diagnostics.HasError() {
        return
    }

    resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}
```

**Flow** (Complex):
1. Read plan from request
2. Check for diagnostics errors
3. Get proxy instance
4. Build nested structures from plan
5. Check for deleted resource (restore pattern)
6. If deleted, restore and configure
7. If not deleted, create new resource
8. Handle attributes requiring separate PATCH
9. Apply orchestrated updates (skills, languages, utilization)
10. Read final state to ensure consistency
11. Save model to state

**Key Patterns for Complex Resources**:
- **Restore Pattern**: Check for deleted resources before creating
- **Multiple API Calls**: Create, PATCH for additional attributes, separate updates for nested structures
- **Nested Structure Building**: Convert Framework types to SDK types for each nested level
- **Orchestrated Updates**: `executeAllUpdates()` coordinates multiple update operations
- **Read After Write**: Call `readUser()` to get final state with all nested structures

---

#### 7.2 Read Method (Complex)

**Purpose**: Fetch current resource state from API, flatten all nested structures, and update Terraform state.

**Design Pattern** (Complex):
```go
func (r *<resource>FrameworkResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
    var state <resource>FrameworkResourceModel

    // Read Terraform prior state data into the model
    resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
    if resp.Diagnostics.HasError() {
        return
    }

    proxy := Get<ResourceName>Proxy(r.clientConfig)

    // Fetch current user state from API to detect drift and refresh Terraform state.
    // Uses shared helper that implements retry logic and flattens all nested structures.
    readUser(ctx, &state, proxy, &resp.Diagnostics, false) // false = normal read mode
    if resp.Diagnostics.HasError() {
        return
    }

    // Set the state
    resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}
```

**Flow** (Complex):
1. Read current state from request
2. Check for diagnostics errors
3. Get proxy instance
4. Call shared `readUser()` helper
5. Helper implements retry logic for eventual consistency
6. Helper handles 404 (removes from state)
7. Helper flattens all nested structures
8. Save model to state

**Shared Read Logic Pattern**:
```go
func readUser(ctx context.Context, model *UserFrameworkResourceModel, proxy *userProxy, diagnostics *pfdiag.Diagnostics, isImport ...bool) {
    log.Printf("Reading user %s", model.Id.ValueString())

    // Determine if this is an import operation
    importMode := len(isImport) > 0 && isImport[0]

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

        // Set basic user attributes
        setBasicUserAttributes(model, currentUser)
        setManagerAttribute(model, currentUser)

        // Flatten addresses (2-level nesting)
        var addressDiags pfdiag.Diagnostics
        model.Addresses, addressDiags = flattenUserAddresses(ctx, currentUser.Addresses, proxy)
        if addressDiags.HasError() {
            return false, fmt.Errorf("Failed to flatten addresses: %v", addressDiags)
        }

        // Handle managed attributes with consistent pattern
        handleManagedRoutingSkills(model, currentUser, &addressDiags)
        handleManagedRoutingLanguages(model, currentUser, &addressDiags)
        handleManagedLocations(model, currentUser, &addressDiags)

        // Flatten profile skills and certifications
        model.ProfileSkills = flattenUserData(currentUser.ProfileSkills)
        model.Certifications = flattenUserData(currentUser.Certifications)

        // Handle employer info
        handleManagedEmployerInfo(model, currentUser, &addressDiags)

        // Get and handle voicemail userpolicies
        if !handleVoicemailUserpolicies(ctx, model, proxy, &addressDiags, importMode) {
            return true, fmt.Errorf("Failed to read voicemail userpolicies")
        }

        // Get routing utilization (3-level nesting)
        if !handleRoutingUtilization(ctx, model, proxy, &addressDiags) {
            return false, fmt.Errorf("Failed to read routing utilization")
        }

        return false, nil
    })

    diagnostics.Append(retryDiags...)
}
```

**Key Patterns for Complex Resources**:
- **Shared Read Helper**: `readUser()` called from Create, Read, Update
- **Retry Logic**: `util.PFWithRetriesForRead()` handles eventual consistency
- **404 Handling**: Retryable error for eventual consistency
- **Expansions**: Fetch nested data with `expand` query parameter
- **Managed vs Unmanaged**: Only populate attributes that are managed in config
- **Flatten All Nested Structures**: Convert SDK types to Framework types for all levels
- **Import Mode**: Special handling during import to avoid populating defaults
- **Extensive Logging**: Debug information for troubleshooting complex operations

---

#### 7.3 Update Method (Complex)

**Purpose**: Update existing resource via API, handle nested structures, and update Terraform state.

**Design Pattern** (Complex):
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

    proxy := Get<ResourceName>Proxy(r.clientConfig)

    // Call the helper function that contains all update logic
    // This matches SDKv2 pattern where Update calls updateUser()
    updateUser(ctx, &plan, proxy, r.clientConfig, &resp.Diagnostics, &state)
    if resp.Diagnostics.HasError() {
        return
    }

    resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}
```

**Shared Update Logic Pattern**:
```go
func updateUser(ctx context.Context, plan *UserFrameworkResourceModel, proxy *userProxy, clientConfig *platformclientv2.Configuration, diagnostics *pfdiag.Diagnostics, state ...*UserFrameworkResourceModel) {
    // Wait for extension pool activation if needed
    waitForExtensionPoolActivation(ctx, plan, proxy)

    // Build addresses from plan
    addresses, addressDiags := buildSdkAddresses(ctx, plan.Addresses)
    *diagnostics = append(*diagnostics, addressDiags...)
    if diagnostics.HasError() {
        return
    }

    email := plan.Email.ValueString()
    log.Printf("Updating user %s", email)

    // Get current state for change detection
    var currentState *UserFrameworkResourceModel
    if len(state) > 0 {
        currentState = state[0]
    }

    // If state changes, it must be updated separately
    if currentState != nil && !plan.State.Equal(currentState.State) {
        log.Printf("Updating state for user %s", email)
        updateUserRequestBody := platformclientv2.Updateuser{
            State: platformclientv2.String(plan.State.ValueString()),
        }
        diagErr := executeUpdateUser(ctx, plan, proxy, updateUserRequestBody)
        if diagErr.HasError() {
            *diagnostics = append(*diagnostics, diagErr...)
            return
        }
    }

    // Update all other attributes
    updateUserRequestBody := platformclientv2.Updateuser{
        Name:           platformclientv2.String(plan.Name.ValueString()),
        Department:     platformclientv2.String(plan.Department.ValueString()),
        Title:          platformclientv2.String(plan.Title.ValueString()),
        Manager:        platformclientv2.String(plan.Manager.ValueString()),
        AcdAutoAnswer:  platformclientv2.Bool(plan.AcdAutoAnswer.ValueBool()),
        Email:          &email,
        Addresses:      addresses,
        Locations:      buildSdkLocations(ctx, plan.Locations),
        Certifications: buildSdkCertifications(ctx, plan.Certifications),
        ProfileSkills:  buildSdkProfileSkills(ctx, plan.ProfileSkills),
        EmployerInfo:   buildSdkEmployerInfo(ctx, plan.EmployerInfo),
    }

    // PATCH core user attributes
    diagErr := executeUpdateUser(ctx, plan, proxy, updateUserRequestBody)
    if diagErr.HasError() {
        *diagnostics = append(*diagnostics, diagErr...)
        return
    }

    // Apply updates requiring separate API endpoints
    frameworkDiags := executeAllUpdates(ctx, plan, proxy, clientConfig, true, currentState)
    if frameworkDiags.HasError() {
        *diagnostics = append(*diagnostics, frameworkDiags...)
        return
    }

    // Read back final state
    readUser(ctx, plan, proxy, diagnostics)
}
```

**Flow** (Complex):
1. Read plan and state from request
2. Check for diagnostics errors
3. Get proxy instance
4. Call shared `updateUser()` helper
5. Helper waits for extension pool activation (if needed)
6. Helper builds nested structures from plan
7. Helper handles state transitions separately
8. Helper updates core attributes via PATCH
9. Helper applies orchestrated updates (skills, languages, utilization)
10. Helper reads final state
11. Save model to state

**Key Patterns for Complex Resources**:
- **Shared Update Helper**: `updateUser()` called from Create, Update, restore operations
- **Change Detection**: Compare plan vs state to detect changes
- **State Transitions**: Handle state changes separately (special API requirement)
- **Multiple API Calls**: PATCH for core attributes, separate calls for nested structures
- **Orchestrated Updates**: `executeAllUpdates()` coordinates multiple update operations
- **Managed vs Unmanaged**: Only update attributes that are managed in config
- **Read After Write**: Call `readUser()` to get final state with all nested structures
- **Retry Logic**: `executeUpdateUser()` wraps PATCH with version mismatch retry

---

#### 7.4 Delete Method (Complex)

**Purpose**: Delete resource via API and verify deletion (soft delete pattern).

**Design Pattern** (Complex):
```go
func (r *<resource>FrameworkResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
    var state <resource>FrameworkResourceModel

    // Read Terraform prior state data into the model
    resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
    if resp.Diagnostics.HasError() {
        return
    }

    proxy := Get<ResourceName>Proxy(r.clientConfig)
    email := state.Email.ValueString()

    log.Printf("Deleting user %s", email)

    err := util.PFRetryWhen(util.IsVersionMismatch, func() (*platformclientv2.APIResponse, pfdiag.Diagnostics) {
        // Directory occasionally returns version errors on deletes if an object was updated at the same time.
        _, proxyDelResponse, err := proxy.deleteUser(ctx, state.Id.ValueString())
        if err != nil {
            return proxyDelResponse, util.BuildFrameworkAPIDiagnosticError(ResourceType,
                fmt.Sprintf("Failed to delete user %s error: %s", state.Id.ValueString(), err), proxyDelResponse)
        }
        log.Printf("Deleted user %s", email)
        return nil, nil
    })
    if err != nil {
        resp.Diagnostics.Append(err...)
        return
    }

    // Verify user in deleted state and search index has been updated
    verifyDiags := util.PFWithRetries(ctx, 3*time.Minute, func() (bool, error) {
        id, getErr := getDeletedUserId(email, proxy)
        if getErr.HasError() {
            // Non-retryable error
            return false, fmt.Errorf("error searching for deleted user %s: %v", email, getErr)
        }
        if id == nil {
            // Retryable - user not yet in deleted state
            return true, fmt.Errorf("user %s not yet in deleted state", email)
        }
        // Success - user is deleted
        return false, nil
    })

    if verifyDiags.HasError() {
        resp.Diagnostics.Append(verifyDiags...)
        return
    }
}
```

**Flow** (Complex):
1. Read state from request
2. Check for diagnostics errors
3. Get proxy instance
4. Call proxy delete method with retry (version mismatch handling)
5. Handle errors
6. Verify deletion with retry logic (soft delete pattern)
7. Handle verification errors

**Key Patterns for Complex Resources**:
- **Soft Delete**: Resource moves to "deleted" state, not permanently removed
- **Version Mismatch Retry**: `util.PFRetryWhen()` handles concurrent modification errors
- **Deletion Verification**: Search for resource in deleted state to confirm
- **Extended Timeout**: 3 minutes for eventual consistency (longer than simple resources)
- **Why Verify**: Ensures resource is truly deleted before Terraform considers it gone

---

#### 7.5 ImportState Method

**Purpose**: Enable importing existing resources into Terraform state with full nested structures.

**Design Pattern**:
```go
func (r *<resource>FrameworkResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
    // Retrieve import ID and save to id attribute
    resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)

    // Read the resource with import flag
    var state <resource>FrameworkResourceModel
    resp.Diagnostics.Append(resp.State.Get(ctx, &state)...)
    if resp.Diagnostics.HasError() {
        return
    }

    proxy := Get<ResourceName>Proxy(r.clientConfig)
    readUser(ctx, &state, proxy, &resp.Diagnostics, true) // true = import mode
    if resp.Diagnostics.HasError() {
        return
    }

    resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}
```

**How It Works**:
- User runs: `terraform import genesyscloud_user.example <user_id>`
- Framework calls ImportState with ID
- `ImportStatePassthroughID` sets ID in state
- Call `readUser()` with import mode flag to populate all attributes
- Import mode avoids populating default values that weren't explicitly configured

**Key Points for Complex Resources**:
- **Import Mode Flag**: Passed to `readUser()` to handle defaults differently
- **Full Nested Structure**: All nested blocks are fetched and flattened
- **Managed vs Unmanaged**: Import mode only populates non-default values
- **After Import**: User runs `terraform plan` to see what configuration needs to be added

---

### 8. Helper Functions (Complex)

#### 8.1 Flatten Functions (SDK → Framework)

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
            // Build phone number object
            phoneObj, phoneDiags := buildPhoneObject(address, proxy)
            diagnostics.Append(phoneDiags...)
            phoneElements = append(phoneElements, phoneObj)
            
        case "EMAIL":
            // Build email object
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
    
    // Return as a list with one element (matching schema: ListNestedBlock with SizeAtMost(1))
    addressesList, listDiags := types.ListValue(addressesElementType(), []attr.Value{addressesObj})
    diagnostics.Append(listDiags...)
    
    return addressesList, diagnostics
}
```

**Pattern** (3-level nesting):
```go
func readUserRoutingUtilization(ctx context.Context, state *UserFrameworkResourceModel, proxy *userProxy) (*platformclientv2.APIResponse, pfdiag.Diagnostics) {
    var diagnostics pfdiag.Diagnostics
    
    // Make API call
    apiClient := &proxy.routingApi.Configuration.APIClient
    path := fmt.Sprintf("%s/api/v2/routing/users/%s/utilization",
        proxy.routingApi.Configuration.BasePath, state.Id.ValueString())
    
    response, err := apiClient.CallAPI(path, "GET", nil, buildHeaderParams(proxy.routingApi),
        nil, nil, "", nil, "")
    if err != nil {
        return response, util.BuildFrameworkAPIDiagnosticError(ResourceType,
            fmt.Sprintf("Failed to read routing utilization for user %s error: %s", state.Id.ValueString(), err), response)
    }
    
    // Unmarshal response
    agentUtilization := &agentUtilizationWithLabels{}
    if err = json.Unmarshal(response.RawBody, &agentUtilization); err != nil {
        diagnostics.AddError("JSON Unmarshal Error",
            fmt.Sprintf("Failed to unmarshal routing utilization: %s", err.Error()))
        return response, diagnostics
    }
    
    if agentUtilization.Level == "Organization" {
        // If the settings are org-wide, set to null
        state.RoutingUtilization = types.ListNull(routingUtilizationElementType())
        return response, diagnostics
    }
    
    // Build the settings object
    allSettingsAttrs := map[string]attr.Value{
        "call":               types.ListNull(mediaUtilizationElementType()),
        "callback":           types.ListNull(mediaUtilizationElementType()),
        "message":            types.ListNull(mediaUtilizationElementType()),
        "email":              types.ListNull(mediaUtilizationElementType()),
        "chat":               types.ListNull(mediaUtilizationElementType()),
        "label_utilizations": types.ListNull(labelUtilizationElementType()),
    }
    
    // Flatten media utilization settings
    if agentUtilization.Utilization != nil {
        for sdkType, schemaType := range getUtilizationMediaTypes() {
            if mediaSettings, ok := agentUtilization.Utilization[sdkType]; ok {
                flattenedMedia, diags := flattenUtilizationSetting(mediaSettings)
                diagnostics.Append(diags...)
                allSettingsAttrs[schemaType] = flattenedMedia
            }
        }
    }
    
    // Flatten label utilizations
    if agentUtilization.LabelUtilizations != nil {
        filteredLabels, diags := filterAndFlattenLabelUtilizations(ctx,
            agentUtilization.LabelUtilizations, state.RoutingUtilization)
        diagnostics.Append(diags...)
        allSettingsAttrs["label_utilizations"] = filteredLabels
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
- Use element type helpers for type safety
- Handle null/empty cases explicitly
- Return diagnostics for error handling
- Preserve all attribute values
- Use `types.SetValue()`, `types.ListValue()`, `types.ObjectValue()`
- For 2-level nesting: Process parent, then children
- For 3-level nesting: Process parent, then children, then grandchildren

---

#### 8.2 Build Functions (Framework → SDK)

**Purpose**: Convert Framework types to SDK types for API calls.

**Pattern** (1-level nesting):
```go
func buildSdkLocations(ctx context.Context, locations types.Set) *[]platformclientv2.Location {
    // Check if locations is null or unknown
    if locations.IsNull() || locations.IsUnknown() {
        return nil
    }
    
    sdkLocations := make([]platformclientv2.Location, 0)
    
    // Extract locations from Framework Set
    locationElements := locations.Elements()
    
    for _, locElement := range locationElements {
        locObj, ok := locElement.(types.Object)
        if !ok {
            continue
        }
        
        locAttrs := locObj.Attributes()
        
        var locID string
        if locIDAttr, exists := locAttrs["location_id"]; exists && !locIDAttr.IsNull() {
            locID = locIDAttr.(types.String).ValueString()
        }
        
        var locNotes string
        if locNotesAttr, exists := locAttrs["notes"]; exists && !locNotesAttr.IsNull() {
            locNotes = locNotesAttr.(types.String).ValueString()
        }
        
        sdkLocations = append(sdkLocations, platformclientv2.Location{
            Id:    &locID,
            Notes: &locNotes,
        })
    }
    
    return &sdkLocations
}
```

**Pattern** (2-level nesting):
```go
func buildSdkAddresses(ctx context.Context, addresses types.List) (*[]platformclientv2.Contact, pfdiag.Diagnostics) {
    var diagnostics pfdiag.Diagnostics
    sdkAddresses := make([]platformclientv2.Contact, 0)
    
    // Check if addresses is null or unknown
    if addresses.IsNull() || addresses.IsUnknown() {
        return &sdkAddresses, diagnostics
    }
    
    // Define the model for addresses block
    type AddressesModel struct {
        OtherEmails  types.Set `tfsdk:"other_emails"`
        PhoneNumbers types.Set `tfsdk:"phone_numbers"`
    }
    
    // Extract addresses into typed model
    var addressesBlocks []AddressesModel
    diags := addresses.ElementsAs(ctx, &addressesBlocks, false)
    diagnostics.Append(diags...)
    if diagnostics.HasError() {
        return &sdkAddresses, diagnostics
    }
    
    // Check if we have at least one addresses block
    if len(addressesBlocks) == 0 {
        return &sdkAddresses, diagnostics
    }
    
    // Get the first (and only) addresses block
    addressBlock := addressesBlocks[0]
    
    // Build emails
    if !addressBlock.OtherEmails.IsNull() && !addressBlock.OtherEmails.IsUnknown() {
        emailContacts, emailDiags := buildSdkEmails(addressBlock.OtherEmails)
        diagnostics.Append(emailDiags...)
        sdkAddresses = append(sdkAddresses, emailContacts...)
    }
    
    // Build phone numbers
    if !addressBlock.PhoneNumbers.IsNull() && !addressBlock.PhoneNumbers.IsUnknown() {
        phoneContacts, phoneDiags := buildSdkPhoneNumbers(addressBlock.PhoneNumbers)
        diagnostics.Append(phoneDiags...)
        sdkAddresses = append(sdkAddresses, phoneContacts...)
    }
    
    return &sdkAddresses, diagnostics
}
```

**Key Points**:
- Use `ElementsAs()` to extract Framework types
- Use `As()` to extract nested attributes
- Handle null/unknown values explicitly
- Use pointer methods for SDK types (`ValueStringPointer()`)
- Return SDK types ready for API calls
- For 2-level nesting: Build parent, then children
- For 3-level nesting: Build parent, then children, then grandchildren

---

#### 8.3 Update Orchestration Functions

**Purpose**: Coordinate multiple update operations in correct order with proper error handling.

**Pattern**:
```go
func executeAllUpdates(ctx context.Context, plan *UserFrameworkResourceModel, proxy *userProxy, sdkConfig *platformclientv2.Configuration, updateObjectDivision bool, state ...*UserFrameworkResourceModel) pfdiag.Diagnostics {
    var diagnostics pfdiag.Diagnostics
    
    var currentState *UserFrameworkResourceModel
    if len(state) > 0 {
        currentState = state[0]
    }
    
    if updateObjectDivision {
        diagErr := util.UpdateObjectDivisionPF(ctx, plan, currentState, "USER", sdkConfig)
        if diagErr.HasError() {
            diagnostics.Append(diagErr...)
            return diagnostics
        }
    }
    
    // Update user skills
    diagErr := updateUserSkills(ctx, plan, currentState, proxy)
    if diagErr.HasError() {
        diagnostics.Append(diagErr...)
        return diagnostics
    }
    
    // Update user languages
    diagErr = updateUserLanguages(ctx, plan, currentState, proxy)
    if diagErr.HasError() {
        diagnostics.Append(diagErr...)
        return diagnostics
    }
    
    // Update profile skills
    diagErr = updateUserProfileSkills(ctx, plan, proxy)
    if diagErr.HasError() {
        diagnostics.Append(diagErr...)
        return diagnostics
    }
    
    // Update routing utilization
    diagErr = updateUserRoutingUtilization(ctx, plan, proxy)
    if diagErr.HasError() {
        diagnostics.Append(diagErr...)
        return diagnostics
    }
    
    // Update voicemail policies
    diagErr = updateUserVoicemailPolicies(ctx, plan, currentState, proxy)
    if diagErr.HasError() {
        diagnostics.Append(diagErr...)
        return diagnostics
    }
    
    // Update password
    diagErr = updatePassword(ctx, plan, proxy)
    if diagErr.HasError() {
        diagnostics.Append(diagErr...)
        return diagnostics
    }
    
    return diagnostics
}
```

**Individual Update Function Pattern**:
```go
func updateUserSkills(ctx context.Context, plan *UserFrameworkResourceModel, state *UserFrameworkResourceModel, proxy *userProxy) pfdiag.Diagnostics {
    var diagnostics pfdiag.Diagnostics
    
    // Was this attribute previously managed?
    wasManaged := state != nil && !state.RoutingSkills.IsNull() && !state.RoutingSkills.IsUnknown()
    
    // Check if routing_skills is null or unknown
    isConfigured := !plan.RoutingSkills.IsNull() && !plan.RoutingSkills.IsUnknown()
    
    // Case 1: never managed and still not configured → do nothing
    if !wasManaged && !isConfigured {
        log.Printf("Skills are null/unknown for user %s, skipping skill updates", plan.Id.ValueString())
        return diagnostics
    }
    
    // Skills are configured or were previously managed - process them
    log.Printf("Updating skills for user %s (wasManaged=%v, isConfigured=%v)", plan.Email.ValueString(), wasManaged, isConfigured)
    
    // Build new skills map from Framework types
    newSkillProfs := make(map[string]float64)
    newSkillIds := []string{}
    
    if isConfigured {
        skillElements := plan.RoutingSkills.Elements()
        for _, skillElement := range skillElements {
            skillObj, ok := skillElement.(types.Object)
            if !ok {
                continue
            }
            
            skillAttrs := skillObj.Attributes()
            var skillId string
            var proficiency float64
            
            if skillIdAttr, exists := skillAttrs["skill_id"]; exists && !skillIdAttr.IsNull() {
                skillId = skillIdAttr.(types.String).ValueString()
            }
            
            if proficiencyAttr, exists := skillAttrs["proficiency"]; exists && !proficiencyAttr.IsNull() {
                proficiency = proficiencyAttr.(types.Float64).ValueFloat64()
            }
            
            if skillId == "" {
                continue
            }
            
            newSkillIds = append(newSkillIds, skillId)
            newSkillProfs[skillId] = proficiency
        }
    }
    
    // Get current skills from API
    oldSdkSkills, getErr := getUserRoutingSkills(plan.Id.ValueString(), proxy)
    if getErr != nil {
        return getErr
    }
    
    // Build old skills map
    oldSkillIds := make([]string, 0, len(oldSdkSkills))
    oldSkillProfs := make(map[string]float64)
    for _, skill := range oldSdkSkills {
        oldSkillIds = append(oldSkillIds, *skill.Id)
        oldSkillProfs[*skill.Id] = *skill.Proficiency
    }
    
    // Remove skills that are no longer in configuration
    if len(oldSkillIds) > 0 {
        var skillsToRemove []string
        
        if !isConfigured {
            // Block removed in config but was managed before → clear everything
            skillsToRemove = oldSkillIds
        } else {
            // Normal diff behavior
            skillsToRemove = lists.SliceDifference(oldSkillIds, newSkillIds)
        }
        
        for _, skillId := range skillsToRemove {
            diagErr := util.PFRetryWhen(util.IsVersionMismatch, func() (*platformclientv2.APIResponse, pfdiag.Diagnostics) {
                resp, err := proxy.userApi.DeleteUserRoutingskill(plan.Id.ValueString(), skillId)
                if err != nil {
                    return resp, util.BuildFrameworkAPIDiagnosticError(ResourceType,
                        fmt.Sprintf("Failed to remove skill from user %s error: %s", plan.Id.ValueString(), err), resp)
                }
                return nil, nil
            })
            if diagErr != nil {
                return diagErr
            }
        }
    }
    
    // Add or update skills only when attribute is configured in plan
    if isConfigured && len(newSkillIds) > 0 {
        // Skills to add
        skillsToAddOrUpdate := lists.SliceDifference(newSkillIds, oldSkillIds)
        
        // Check for existing proficiencies to update
        for skillID, newProf := range newSkillProfs {
            if oldProf, found := oldSkillProfs[skillID]; found && newProf != oldProf {
                skillsToAddOrUpdate = append(skillsToAddOrUpdate, skillID)
            }
        }
        
        if len(skillsToAddOrUpdate) > 0 {
            if diagErr := updateUserRoutingSkills(plan.Id.ValueString(), skillsToAddOrUpdate,
                newSkillProfs, proxy); diagErr != nil {
                return diagErr
            }
        }
    }
    
    return diagnostics
}
```

**Key Points**:
- **Orchestration**: `executeAllUpdates()` coordinates all update operations
- **Managed vs Unmanaged**: Track whether attribute was previously managed
- **Change Detection**: Compare plan vs state to detect changes
- **Removal Logic**: Clear attributes that were managed but now removed from config
- **Batch Operations**: Use bulk APIs where available (skills, languages)
- **Retry Logic**: Wrap API calls with version mismatch retry
- **Error Handling**: Return diagnostics immediately on error
- **Correct Ordering**: Updates happen in specific order (division, skills, languages, utilization, voicemail, password)

---

#### 8.4 Utility Functions

**Purpose**: Provide reusable helper functions for common operations.

**hasChanges Pattern**:
```go
func hasChanges(plan *UserFrameworkResourceModel, attributes ...string) bool {
    // For create operations, we consider all non-null values as changes
    for _, attr := range attributes {
        switch attr {
        case "manager":
            if !plan.Manager.IsNull() && !plan.Manager.IsUnknown() && plan.Manager.ValueString() != "" {
                return true
            }
        case "locations":
            if !plan.Locations.IsNull() && !plan.Locations.IsUnknown() {
                elements := plan.Locations.Elements()
                if len(elements) > 0 {
                    return true
                }
            }
        case "acd_auto_answer":
            if !plan.AcdAutoAnswer.IsNull() && !plan.AcdAutoAnswer.IsUnknown() {
                return true
            }
        // ... other attributes
        }
    }
    return false
}
```

**getDeletedUserId Pattern**:
```go
func getDeletedUserId(email string, proxy *userProxy) (*string, pfdiag.Diagnostics) {
    var diagnostics pfdiag.Diagnostics
    
    exactType := "EXACT"
    results, resp, getErr := proxy.userApi.PostUsersSearch(platformclientv2.Usersearchrequest{
        Query: &[]platformclientv2.Usersearchcriteria{
            {
                Fields:  &[]string{"email"},
                Value:   &email,
                VarType: &exactType,
            },
            {
                Fields:  &[]string{"state"},
                Values:  &[]string{"deleted"},
                VarType: &exactType,
            },
        },
    })
    
    if getErr != nil {
        return nil, util.BuildFrameworkAPIDiagnosticError(ResourceType,
            fmt.Sprintf("Failed to search for user %s error: %s", email, getErr), resp)
    }
    
    if results.Results != nil && len(*results.Results) > 0 {
        // User found
        return (*results.Results)[0].Id, diagnostics
    }
    
    return nil, diagnostics
}
```

**restoreDeletedUser Pattern**:
```go
func restoreDeletedUser(ctx context.Context, plan *UserFrameworkResourceModel, proxy *userProxy, clientConfig *platformclientv2.Configuration, diagnostics *pfdiag.Diagnostics) {
    email := plan.Email.ValueString()
    state := plan.State.ValueString()
    
    log.Printf("Restoring deleted user %s", email)
    
    err := util.PFRetryWhen(util.IsVersionMismatch, func() (*platformclientv2.APIResponse, pfdiag.Diagnostics) {
        // Get current user (with version)
        currentUser, proxyResponse, err := proxy.getUserById(ctx, plan.Id.ValueString(), nil, "deleted")
        if err != nil {
            return nil, util.BuildFrameworkAPIDiagnosticError(ResourceType,
                fmt.Sprintf("Failed to read user %s error: %s", plan.Id.ValueString(), err), proxyResponse)
        }
        
        // Restore user by updating state
        restoredUser, proxyPatchResponse, patchErr := proxy.patchUserWithState(ctx, plan.Id.ValueString(),
            &platformclientv2.Updateuser{
                State:   &state,
                Version: currentUser.Version,
            })
        if patchErr != nil {
            return proxyPatchResponse, util.BuildFrameworkAPIDiagnosticError(ResourceType,
                fmt.Sprintf("Failed to restore deleted user %s | Error: %s.", email, patchErr), proxyPatchResponse)
        }
        
        // Apply full configuration (equivalent to SDKv2's updateUser call)
        updateUser(ctx, plan, proxy, clientConfig, diagnostics)
        
        if diagnostics.HasError() {
            return nil, *diagnostics
        }
        
        return nil, nil
    })
    
    if err != nil {
        *diagnostics = append(*diagnostics, err...)
    }
}
```

**Key Points**:
- **hasChanges**: Detect which attributes have non-null values
- **getDeletedUserId**: Search for resource in deleted state
- **restoreDeletedUser**: Restore deleted resource and apply full configuration
- **waitForExtensionPoolActivation**: Wait for newly created extension pools to be ready
- **executeUpdateUser**: Wrap PATCH with version mismatch retry
- **convertSDKDiagnosticsToFramework**: Convert between diagnostic types

---

### 9. GetAll Functions (Complex)

#### 9.1 GetAll<ResourceName> (Framework Version - Phase 2 Future)

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
    proxy := Get<ResourceName>Proxy(clientConfig)
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
        hashedUniqueFields, err := util.QuickHashFields(*resource.Name, *resource.Department,
            resource.PrimaryContactInfo, resource.Addresses)
        if err != nil {
            diagErr.AddError("Failed to hash <resource> fields", err.Error())
            return nil, diagErr
        }
        exportMap[*resource.Id] = &resourceExporter.ResourceMeta{
            BlockLabel: *resource.Email,
            BlockHash:  hashedUniqueFields,
        }
    }
    return exportMap, nil
}
```

**Key Points**:
- Uses Plugin Framework diagnostics (`pfdiag.Diagnostics`)
- Returns clean export map without flat attributes
- Marked as Phase 2 future with comment
- Not currently used by exporter
- Hash includes multiple fields for complex resources

---

#### 9.2 GetAll<ResourceName>SDK (SDK Version - Phase 1 Current)

**Purpose**: Fetch all resources for export using SDK diagnostics with lazy fetch pattern (currently used by exporter).

**Design Pattern** (Complex with Lazy Fetch):
```go
// GetAll<ResourceName>SDK retrieves all <resources> for export using SDK diagnostics.
// This is the Phase 1 implementation that implements the lazy fetch pattern for performance.
//
// IMPORTANT: This function is CURRENTLY USED by the exporter (see <ResourceName>Exporter).
// It implements the lazy fetch pattern for performance optimization.
//
// Returns:
//   - resourceExporter.ResourceIDMetaMap: Map of resource IDs to metadata with lazy fetch callbacks
//   - sdkdiag.Diagnostics: SDK diagnostics (required by current exporter)
//
// Lazy Fetch Pattern:
//   - Phase 1: Fetch all resource IDs and basic info (lightweight - 1 API call)
//   - Filter: Exporter applies filters to determine which resources to export
//   - Phase 2: Fetch full details ONLY for filtered resources via LazyFetchAttributes callback
//
// For Plugin Framework resources, this function sets a lazy fetch callback in ResourceMeta
// that will be invoked AFTER filtering, only for resources that will be exported.
//
// TODO: Remove this function once all resources are migrated to Plugin Framework
// and the exporter is updated to use GetAll<ResourceName> (Phase 2).
func GetAll<ResourceName>SDK(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, sdkdiag.Diagnostics) {
    proxy := Get<ResourceName>Proxy(clientConfig)
    resources := make(resourceExporter.ResourceIDMetaMap)

    // Step 1: Fetch all resource IDs (lightweight - 1 API call)
    users, resp, err := proxy.GetAllUser(ctx)
    if err != nil {
        return nil, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to get users: %s", err), resp)
    }

    log.Printf("[INFO] Found %d users for export", len(*users))

    // Step 2: Create ResourceMeta with lazy fetch callback for each resource
    for _, user := range *users {
        if user.Id == nil || user.Email == nil {
            continue
        }

        hashedUniqueFields, err := util.QuickHashFields(user.Name, user.Department,
            user.PrimaryContactInfo, user.Addresses)
        if err != nil {
            return nil, sdkdiag.FromErr(err)
        }

        // Capture user ID for closure
        userId := *user.Id

        resources[*user.Id] = &resourceExporter.ResourceMeta{
            BlockLabel: *user.Email,
            BlockHash:  hashedUniqueFields,

            // Set lazy fetch callback - will be called AFTER filtering, only for exported users
            // This callback fetches full user details and builds the complete attribute map
            LazyFetchAttributes: func(fetchCtx context.Context) (map[string]string, error) {
                log.Printf("[DEBUG] EXPORT: Lazy fetch callback CALLED for user %s", userId)

                // Fetch full user details with expansions
                expansions := []string{"skills", "languages", "locations", "profileSkills", "certifications", "employerInfo"}
                fullUser, _, err := proxy.getUserById(fetchCtx, userId, expansions, "")
                if err != nil {
                    log.Printf("[ERROR] EXPORT: Failed to fetch user details for %s: %v", userId, err)
                    return nil, fmt.Errorf("failed to fetch user details: %w", err)
                }
                log.Printf("[DEBUG] EXPORT: Successfully fetched user details for %s", userId)

                // Build complete attribute map (reuses existing function)
                attributes, err := buildUserAttributes(fetchCtx, fullUser, proxy)
                if err != nil {
                    log.Printf("[ERROR] EXPORT: Failed to build attributes for %s: %v", userId, err)
                    return nil, fmt.Errorf("failed to build attributes: %w", err)
                }

                log.Printf("[DEBUG] EXPORT: Built %d attributes for user %s", len(attributes), userId)
                return attributes, nil
            },
        }
    }

    log.Printf("[INFO] EXPORT: Created lazy fetch callbacks for %d users", len(resources))
    return resources, nil
}
```

**buildUserAttributes Pattern** (Complex):
```go
// buildUserAttributes builds a flat attribute map for export (Phase 1 temporary).
// This function is called by the lazy fetch callback in GetAllUsersSDK.
func buildUserAttributes(ctx context.Context, user *platformclientv2.User, proxy *userProxy) (map[string]string, error) {
    attributes := make(map[string]string)
    
    // Basic attributes
    if user.Name != nil {
        attributes["name"] = *user.Name
    }
    if user.Email != nil {
        attributes["email"] = *user.Email
    }
    if user.State != nil {
        attributes["state"] = *user.State
    }
    if user.Department != nil {
        attributes["department"] = *user.Department
    }
    if user.Title != nil {
        attributes["title"] = *user.Title
    }
    
    // Division
    if user.Division != nil && user.Division.Id != nil {
        attributes["division_id"] = *user.Division.Id
    }
    
    // Manager
    if user.Manager != nil && (*user.Manager) != nil && (*user.Manager).Id != nil {
        attributes["manager"] = *(*user.Manager).Id
    }
    
    // ACD Auto Answer
    if user.AcdAutoAnswer != nil {
        attributes["acd_auto_answer"] = fmt.Sprintf("%v", *user.AcdAutoAnswer)
    }
    
    // Routing Skills (flatten to JSON)
    if user.Skills != nil && len(*user.Skills) > 0 {
        skillsJSON, err := json.Marshal(*user.Skills)
        if err == nil {
            attributes["routing_skills"] = string(skillsJSON)
        }
    }
    
    // Routing Languages (flatten to JSON)
    if user.Languages != nil && len(*user.Languages) > 0 {
        languagesJSON, err := json.Marshal(*user.Languages)
        if err == nil {
            attributes["routing_languages"] = string(languagesJSON)
        }
    }
    
    // Locations (flatten to JSON)
    if user.Locations != nil && len(*user.Locations) > 0 {
        locationsJSON, err := json.Marshal(*user.Locations)
        if err == nil {
            attributes["locations"] = string(locationsJSON)
        }
    }
    
    // Addresses (flatten to JSON)
    if user.Addresses != nil && len(*user.Addresses) > 0 {
        addressesJSON, err := json.Marshal(*user.Addresses)
        if err == nil {
            attributes["addresses"] = string(addressesJSON)
        }
    }
    
    // Fetch and flatten routing utilization
    utilization, _, err := proxy.getUserRoutingUtilization(ctx, *user.Id)
    if err == nil && utilization != nil {
        utilizationJSON, err := json.Marshal(utilization)
        if err == nil {
            attributes["routing_utilization"] = string(utilizationJSON)
        }
    }
    
    // Fetch and flatten voicemail policies
    voicemail, _, err := proxy.getVoicemailUserpoliciesById(ctx, *user.Id)
    if err == nil && voicemail != nil {
        voicemailJSON, err := json.Marshal(voicemail)
        if err == nil {
            attributes["voicemail_userpolicies"] = string(voicemailJSON)
        }
    }
    
    return attributes, nil
}
```

**Key Points**:
- Uses SDK diagnostics (`sdkdiag.Diagnostics`)
- Implements lazy fetch pattern with `LazyFetchAttributes` callback
- **Phase 1**: Fetch all resource IDs (lightweight)
- **Filter**: Exporter applies filters
- **Phase 2**: Callback fetches full details only for exported resources
- Builds flat attribute map for each resource
- Marked with TODO for Phase 2 removal
- Complex resources require multiple API calls in callback (user details, utilization, voicemail)
- Nested structures flattened to JSON strings for export

**Lazy Fetch Benefits**:
- Reduces API calls when exporting subset of resources
- Improves performance for large organizations
- Only fetches full details for resources that will be exported
- Matches SDKv2 pattern for backward compatibility

---

## Part 2: Data Source Implementation (Complex)

### 1. Data Source Struct

**Purpose**: Hold data source-level configuration.

**Design Pattern**:
```go
// <resource>FrameworkDataSource defines the data source implementation for Plugin Framework.
type <resource>FrameworkDataSource struct {
    clientConfig *platformclientv2.Configuration
}
```

**Example** (user):
```go
// UserFrameworkDataSource implements the Terraform Plugin Framework data source for Genesys Cloud Users.
type UserFrameworkDataSource struct {
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
    Id    types.String `tfsdk:"id"`
    Email types.String `tfsdk:"email"`
    Name  types.String `tfsdk:"name"`
}
```

**Example** (user):
```go
// UserFrameworkDataSourceModel describes the data source data model
type UserFrameworkDataSourceModel struct {
    Id    types.String `tfsdk:"id"`
    Email types.String `tfsdk:"email"`
    Name  types.String `tfsdk:"name"`
}
```

**Key Points**:
- Simpler than resource model (only lookup criteria and result)
- Typically includes `id` and lookup fields (email, name)
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

#### 4.2 Read Method (Data Source - Complex)

**Purpose**: Look up resource by name or email and return ID with retry logic and caching.

**Design Pattern** (Complex with Cache):
```go
func (d *<resource>FrameworkDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
    var config <resource>FrameworkDataSourceModel

    // Read Terraform configuration data into the model
    resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
    if resp.Diagnostics.HasError() {
        return
    }

    // Validate that at least one search field is provided
    if config.Email.IsNull() && config.Name.IsNull() {
        resp.Diagnostics.Append(
            util.BuildFrameworkDiagnosticError(
                ResourceType,
                "no user search field specified",
                nil,
            )...,
        )
        return
    }

    // Determine search key
    var searchKey string
    if !config.Email.IsNull() {
        searchKey = config.Email.ValueString()
    }
    if !config.Name.IsNull() {
        searchKey = config.Name.ValueString()
    }

    log.Printf("Searching for user with key: %s", searchKey)

    // Initialize cache if not already initialized
    if dataSourceUserCache == nil {
        dataSourceUserCache = rc.NewDataSourceCache(d.clientConfig, hydrateUserCache, getUserByName)
    }

    // Retrieve user ID from cache or API
    userId, sdkDiags := rc.RetrieveId(dataSourceUserCache, ResourceType, searchKey, ctx)
    if sdkDiags.HasError() {
        frameworkDiags := util.ConvertSDKDiagnosticsToFramework(sdkDiags)
        resp.Diagnostics.Append(frameworkDiags...)
        return
    }

    // Set the ID in the state
    config.Id = types.StringValue(userId)

    // Fetch the full user details to populate name and email
    proxy := GetUserProxy(d.clientConfig)
    user, response, err := proxy.getUserById(ctx, userId, []string{}, "")
    if err != nil {
        resp.Diagnostics.Append(
            util.BuildFrameworkAPIDiagnosticError(
                ResourceType,
                fmt.Sprintf("Failed to retrieve user details for ID %s: %s", userId, err),
                response,
            )...,
        )
        return
    }

    // Populate name and email from the API response
    if user.Name != nil {
        config.Name = types.StringValue(*user.Name)
    }
    if user.Email != nil {
        config.Email = types.StringValue(*user.Email)
    }

    log.Printf("Found user with ID: %s, Name: %s, Email: %s", userId, config.Name.ValueString(), config.Email.ValueString())

    // Save data into Terraform state
    resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
```

**Helper Functions**:
```go
// getUserByName retrieves a user ID by searching for a user by name or email.
// Note: Returns SDKv2 diag.Diagnostics for compatibility with resource_cache infrastructure.
func getUserByName(c *rc.DataSourceCache, searchField string, ctx context.Context) (string, sdkdiag.Diagnostics) {
    log.Printf("getUserByName for data source %s", ResourceType)
    proxy := GetUserProxy(c.ClientConfig)
    userId := ""
    exactSearchType := "EXACT"
    sortOrderAsc := "ASC"
    emailField := "email"

    searchCriteria := platformclientv2.Usersearchcriteria{
        VarType: &exactSearchType,
    }
    searchFieldValue, searchFieldType := emailorNameDisambiguation(searchField)
    searchCriteria.Fields = &[]string{searchFieldType}
    searchCriteria.Value = &searchFieldValue

    sdkDiags := util.WithRetries(ctx, 15*time.Second, func() *retry.RetryError {
        users, resp, getErr := proxy.getUserByName(ctx, platformclientv2.Usersearchrequest{
            SortBy:    &emailField,
            SortOrder: &sortOrderAsc,
            Query:     &[]platformclientv2.Usersearchcriteria{searchCriteria},
        })
        if getErr != nil {
            return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Error requesting users: %s", getErr), resp))
        }

        if users.Results == nil || len(*users.Results) == 0 {
            return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("No users found with search criteria %v", searchCriteria), resp))
        }

        // Select first user in the list
        userId = *(*users.Results)[0].Id
        return nil
    })

    log.Printf("getUserByName completed for data source %s", ResourceType)
    return userId, sdkDiags
}

func hydrateUserCache(c *rc.DataSourceCache, ctx context.Context) error {
    log.Printf("hydrating cache for data source %s", ResourceType)
    proxy := GetUserProxy(c.ClientConfig)
    const pageSize = 100
    users, response, err := proxy.hydrateUserCache(ctx, pageSize, 1)
    if err != nil {
        return fmt.Errorf("failed to get first page of users: %v %v", err, response)
    }

    if users.Entities == nil || len(*users.Entities) == 0 {
        return nil
    }

    for _, user := range *users.Entities {
        c.Cache[*user.Name] = *user.Id
        c.Cache[*user.Email] = *user.Id
    }

    for pageNum := 2; pageNum <= *users.PageCount; pageNum++ {
        users, response, err := proxy.hydrateUserCache(ctx, pageSize, pageNum)

        log.Printf("hydrating cache for data source %s with page number: %v", ResourceType, pageNum)
        if err != nil {
            return fmt.Errorf("failed to get page of users: %v %v", err, response)
        }
        if users.Entities == nil || len(*users.Entities) == 0 {
            break
        }
        // Add ids to cache
        for _, user := range *users.Entities {
            c.Cache[*user.Name] = *user.Id
            c.Cache[*user.Email] = *user.Id
        }
    }
    log.Printf("cache hydration completed for data source %s", ResourceType)
    return nil
}
```

**Key Points for Complex Resources**:
- **Multiple Lookup Criteria**: Support email and name lookup
- **Cache Pattern**: Use resource_cache for performance
- **Retry Logic**: 15 seconds for eventual consistency
- **Email/Name Disambiguation**: Detect whether search field is email or name
- **Cache Hydration**: Populate cache with all resources on first use
- **Fetch Full Details**: After ID lookup, fetch full resource to populate all fields
- **SDK Diagnostics**: Cache infrastructure uses SDK diagnostics (will be updated in Phase 2)

---

## SDKv2 vs Plugin Framework Comparison (Complex Resources)

### Resource Lifecycle

**SDKv2**:
```go
func ResourceUser() *schema.Resource {
    return &schema.Resource{
        CreateContext: createUser,
        ReadContext:   readUser,
        UpdateContext: updateUser,
        DeleteContext: deleteUser,
        Importer: &schema.ResourceImporter{
            StateContext: schema.ImportStatePassthroughContext,
        },
        Schema: map[string]*schema.Schema{
            "routing_skills": {
                Type:     schema.TypeSet,
                Optional: true,
                Elem: &schema.Resource{
                    Schema: map[string]*schema.Schema{
                        "skill_id": {
                            Type:     schema.TypeString,
                            Required: true,
                        },
                        "proficiency": {
                            Type:     schema.TypeFloat,
                            Required: true,
                        },
                    },
                },
            },
            // ...
        },
    }
}

func createUser(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
    // Implementation with d.Get(), d.Set()
}
```

**Plugin Framework**:
```go
type UserFrameworkResource struct {
    clientConfig *platformclientv2.Configuration
}

type UserFrameworkResourceModel struct {
    Id            types.String `tfsdk:"id"`
    RoutingSkills types.Set    `tfsdk:"routing_skills"`
    // ...
}

type RoutingSkillModel struct {
    SkillId     types.String  `tfsdk:"skill_id"`
    Proficiency types.Float64 `tfsdk:"proficiency"`
}

func (r *UserFrameworkResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
    var plan UserFrameworkResourceModel
    resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
    // Implementation with typed models
}
```

**Key Differences**:

| Aspect | SDKv2 | Plugin Framework |
|--------|-------|------------------|
| Structure | Function-based callbacks | Method-based interface |
| State access | `d *schema.ResourceData` | `req.Plan`, `req.State`, `resp.State` |
| Nested structures | `d.Get("routing_skills").([]interface{})` | Typed models with `types.Set` |
| Error handling | Return `diag.Diagnostics` | Add to `resp.Diagnostics` |
| Type safety | Runtime (interface{}) | Compile-time (typed models) |
| Null handling | Pointer nil checks | `IsNull()`, `IsUnknown()` methods |
| Nested block access | Type assertions and loops | `ElementsAs()` with typed models |

---

## Design Patterns and Best Practices (Complex Resources)

### Pattern 1: Nested Structure Models

**Pattern**:
```go
// Define separate model for each nested level
type ParentModel struct {
    NestedBlock types.List `tfsdk:"nested_block"`
}

type NestedBlockModel struct {
    Attribute types.String `tfsdk:"attribute"`
}
```

**Why**:
- Type safety for nested structures
- Clear model hierarchy
- Easier to maintain and understand
- Compile-time validation

### Pattern 2: Element Type Helpers

**Pattern**:
```go
func routingSkillsElementType() types.ObjectType {
    return types.ObjectType{
        AttrTypes: map[string]attr.Type{
            "skill_id":    types.StringType,
            "proficiency": types.Float64Type,
        },
    }
}
```

**Why**:
- Reusable type definitions
- Consistent type usage across flatten/build functions
- Single source of truth for element types
- Prevents type mismatches

### Pattern 3: Shared Read Logic

**Pattern**:
```go
func readUser(ctx context.Context, model *UserFrameworkResourceModel, proxy *userProxy, diagnostics *pfdiag.Diagnostics, isImport ...bool) {
    // Shared read logic called from Create, Read, Update
}
```

**Why**:
- Avoids code duplication
- Consistent state handling
- Single place for retry logic
- Easier to maintain

### Pattern 4: Update Orchestration

**Pattern**:
```go
func executeAllUpdates(ctx context.Context, plan *UserFrameworkResourceModel, proxy *userProxy, clientConfig *platformclientv2.Configuration, updateObjectDivision bool, state ...*UserFrameworkResourceModel) pfdiag.Diagnostics {
    // Coordinate multiple update operations
}
```

**Why**:
- Correct update ordering
- Centralized error handling
- Managed vs unmanaged attribute tracking
- Easier to add new update operations

### Pattern 5: Managed vs Unmanaged Attributes

**Pattern**:
```go
wasManaged := state != nil && !state.RoutingSkills.IsNull() && !state.RoutingSkills.IsUnknown()
isConfigured := !plan.RoutingSkills.IsNull() && !plan.RoutingSkills.IsUnknown()

if !wasManaged && !isConfigured {
    // Never managed and still not configured → do nothing
    return diagnostics
}

if wasManaged && !isConfigured {
    // Was managed, now removed → clear remote
    // ... removal logic
}

if isConfigured {
    // Currently configured → update
    // ... update logic
}
```

**Why**:
- Respects user intent
- Avoids clearing unmanaged attributes
- Handles attribute removal correctly
- Matches Terraform semantics

### Pattern 6: Lazy Fetch for Export

**Pattern**:
```go
resources[*user.Id] = &resourceExporter.ResourceMeta{
    BlockLabel: *user.Email,
    BlockHash:  hashedUniqueFields,
    LazyFetchAttributes: func(fetchCtx context.Context) (map[string]string, error) {
        // Fetch full details only when needed
        fullUser, _, err := proxy.getUserById(fetchCtx, userId, expansions, "")
        // Build attribute map
        return attributes, nil
    },
}
```

**Why**:
- Performance optimization
- Reduces API calls
- Only fetches details for exported resources
- Matches SDKv2 pattern

### Pattern 7: Flatten/Build Pairing

**Pattern**:
```go
// Flatten: SDK → Framework
func flattenUserSkills(ctx context.Context, skills *[]platformclientv2.Userroutingskill) (types.Set, pfdiag.Diagnostics) {
    // Convert SDK to Framework
}

// Build: Framework → SDK
func buildSdkUserSkills(ctx context.Context, skills types.Set) []platformclientv2.Userroutingskill {
    // Convert Framework to SDK
}
```

**Why**:
- Symmetric operations
- Easier to maintain
- Clear conversion logic
- Consistent naming

### Pattern 8: Extensive Logging

**Pattern**:
```go
log.Printf("[INV] CREATE payload.Addresses=%s", invMustJSON(createUser.Addresses))
log.Printf("[INV] CREATE echo (server user.Addresses)=%s", invMustJSON(userResponse.Addresses))
log.Printf("[INV] FINAL STATE effective addresses: %s", invMustJSON(plan.Addresses))
```

**Why**:
- Complex resources need detailed debugging
- Track data through entire lifecycle
- Identify where data changes
- Troubleshoot production issues

---

## Migration Considerations (Complex Resources)

### Behavior Preservation Checklist

When migrating complex resources from SDKv2 to Framework, verify:

- [ ] CRUD operations behave identically
- [ ] All nested structures are preserved
- [ ] Flatten/build functions handle all attributes
- [ ] Managed vs unmanaged attributes work correctly
- [ ] Update orchestration maintains correct ordering
- [ ] Restore deleted resource pattern works
- [ ] Error messages are equivalent or better
- [ ] Retry logic matches (timeouts, conditions)
- [ ] 404 handling is consistent
- [ ] Logging is extensive and helpful
- [ ] Import functionality works with all nested structures
- [ ] Data source lookup works with retry and cache
- [ ] Export lazy fetch pattern works correctly

### Common Migration Pitfalls (Complex Resources)

#### Pitfall 1: Incorrect Element Type Usage
**Problem**: Using wrong element type in flatten/build functions causes runtime errors.
**Solution**: Define element type helpers and reuse them consistently.

#### Pitfall 2: Missing Null/Unknown Checks in Nested Structures
**Problem**: Not checking null/unknown at each nesting level causes panics.
**Solution**: Check null/unknown before accessing nested attributes.

#### Pitfall 3: Forgetting Managed vs Unmanaged Logic
**Problem**: Clearing unmanaged attributes on read causes state drift.
**Solution**: Track whether attribute was previously managed before updating.

#### Pitfall 4: Incorrect Update Ordering
**Problem**: Updates in wrong order cause API errors.
**Solution**: Follow SDKv2 update order in `executeAllUpdates()`.

#### Pitfall 5: Not Handling 3-Level Nesting
**Problem**: Forgetting to flatten/build grandchild attributes.
**Solution**: Process each nesting level explicitly.

#### Pitfall 6: Missing Import Mode Handling
**Problem**: Populating default values during import causes unnecessary diffs.
**Solution**: Pass import mode flag to read logic and skip defaults.

#### Pitfall 7: Incomplete Lazy Fetch Implementation
**Problem**: Not fetching all nested structures in lazy fetch callback.
**Solution**: Fetch all expansions and build complete attribute map.

#### Pitfall 8: Not Preserving Extensive Logging
**Problem**: Losing debugging information from SDKv2.
**Solution**: Preserve all log statements, especially for complex operations.

---

## Summary

### Key Design Decisions (Complex Resources)

1. **Nested Structure Models**: Separate model structs for each nesting level
2. **Element Type Helpers**: Reusable type definitions for type safety
3. **Shared Read Logic**: `readUser()` helper called from Create, Read, Update
4. **Update Orchestration**: `executeAllUpdates()` coordinates multiple updates
5. **Managed vs Unmanaged**: Track attribute management state
6. **Flatten/Build Pairing**: Symmetric conversion functions for each nested level
7. **Lazy Fetch Pattern**: Performance optimization for export
8. **Extensive Logging**: Detailed debugging for complex operations
9. **Utils File Separation**: Dedicated file for helper functions
10. **Proxy Reuse**: No modifications to proxy methods

### File Structure (Complex Resources)

```
Resource File (resource_genesyscloud_<resource_name>.go):
├── Interface verification
├── Resource struct and model
├── Nested structure models
├── Constructor function
├── Interface methods (Metadata, Schema, Configure)
├── CRUD methods (Create, Read, Update, Delete, ImportState)
└── GetAll functions (Framework and SDK versions)

Utils File (resource_genesyscloud_<resource_name>_utils.go):
├── Package-level variables
├── Element type helper functions
├── Shared read logic (readUser)
├── Update orchestration (updateUser, executeAllUpdates)
├── Flatten functions (1-level, 2-level, 3-level)
├── Build functions (1-level, 2-level, 3-level)
├── Utility functions (hasChanges, getDeletedUserId, restoreDeletedUser)
└── Conversion helpers

Data Source File (data_source_genesyscloud_<resource_name>.go):
├── Interface verification
├── Data source struct and model
├── Constructor function
├── Interface methods (Metadata, Schema, Configure)
├── Read method (with retry and cache)
└── Helper functions (getUserByName, hydrateUserCache)
```

### Next Steps

After completing Stage 2 complex resource migration:
1. Review resource implementation for correctness
2. Verify all nested structures are handled correctly
3. Confirm flatten/build functions preserve all data
4. Test managed vs unmanaged attribute logic
5. Verify GetAll functions work correctly with lazy fetch
6. Confirm data source lookup works with cache
7. Proceed to **Stage 3 – Test Migration**

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
- **Framework Types**: https://developer.hashicorp.com/terraform/plugin/framework/handling-data/types
- **Nested Attributes**: https://developer.hashicorp.com/terraform/plugin/framework/handling-data/attributes/list-nested
