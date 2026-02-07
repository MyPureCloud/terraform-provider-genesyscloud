# Stage 1 – Schema Migration Design

## Overview

This document describes the design decisions, architecture, and patterns used in Stage 1 of the Plugin Framework migration for complex resources. Stage 1 focuses on converting SDKv2 schema definitions to Plugin Framework schema definitions while maintaining backward compatibility and preparing for future stages.

**Reference Implementation**: `genesyscloud/user/resource_genesyscloud_user_schema.go`

**Resource Complexity**: Complex resource with multiple nested blocks (including 3-level nesting), custom validators, plan modifiers, element type helpers, and managed attribute patterns.

---

## Design Principles

### 1. Separation of Concerns
**Principle**: Schema definitions are isolated in a dedicated file, separate from resource implementation.

**Rationale**:
- Clear separation between data structure (schema) and behavior (CRUD operations)
- Easier to review and maintain schema changes
- Supports staged migration approach
- Reduces cognitive load during development

**Implementation**:
- Schema file: `resource_genesyscloud_<resource_name>_schema.go`
- Resource file: `resource_genesyscloud_<resource_name>.go` (Stage 2)
- Utils file: `resource_genesyscloud_<resource_name>_utils.go` (Stage 2)
- Export utilities: `resource_genesyscloud_<resource_name>_export_utils.go` (Stage 4)

### 2. Framework-Native Patterns
**Principle**: Use Plugin Framework idioms and patterns rather than SDKv2 workarounds.

**Rationale**:
- Plugin Framework provides better type safety
- Native support for plan modifiers eliminates custom diff logic
- Improved validation and error handling
- Better performance and maintainability

**Implementation**:
- Use `schema.Schema` instead of `map[string]*schema.Schema`
- Use plan modifiers instead of `DiffSuppressFunc`
- Use Framework types (`types.String`) instead of pointers
- Use element type helpers for complex nested structures

### 3. Backward Compatibility
**Principle**: Schema migration must not break existing Terraform configurations.

**Rationale**:
- Users should be able to upgrade provider without modifying configurations
- State file compatibility must be maintained
- Attribute names and behavior must remain unchanged

**Implementation**:
- Preserve all attribute names exactly as in SDKv2
- Maintain Required/Optional/Computed properties
- Keep validation rules identical
- No changes to attribute types or structure
- Document acceptable behavior changes (e.g., Set identity)

### 4. Export Compatibility
**Principle**: Exporter must work with both SDKv2 and Framework resources during migration.

**Rationale**:
- Not all resources will be migrated simultaneously
- Exporter must handle mixed resource types
- Dependency resolution must work across SDKv2 and Framework resources

**Implementation**:
- Use SDK-compatible `GetAll<ResourceName>SDK()` function
- Define `RefAttrs` for dependency resolution
- Configure `RemoveIfMissing` for conditional removal
- Configure `AllowEmptyArrays` for empty collections
- Configure `AllowZeroValues` for zero-value fields
- Maintain flat attribute map format (Phase 1 temporary)

### 5. Type Safety and Reusability
**Principle**: Define element types once and reuse throughout the codebase.

**Rationale**:
- Ensures consistency between schema and utils functions
- Prevents type mismatches
- Easier to maintain and update
- Reduces code duplication

**Implementation**:
- Create element type helper functions for complex nested objects
- Use these helpers in schema definitions and Stage 2 utils
- Define at package level for cross-file access

---

## Architecture

### File Structure

```
genesyscloud/<resource_name>/
├── resource_genesyscloud_<resource_name>_schema.go          ← Stage 1 (THIS FILE)
├── resource_genesyscloud_<resource_name>.go                 ← Stage 2
├── resource_genesyscloud_<resource_name>_utils.go           ← Stage 2
├── resource_genesyscloud_<resource_name>_test.go            ← Stage 3
├── resource_genesyscloud_<resource_name>_export_utils.go    ← Stage 4
├── data_source_genesyscloud_<resource_name>.go              ← Stage 2
├── data_source_genesyscloud_<resource_name>_test.go         ← Stage 3
├── genesyscloud_<resource_name>_init_test.go                ← Stage 3
└── genesyscloud_<resource_name>_proxy.go                    ← NOT MODIFIED
```

### Schema File Components

The schema file for complex resources contains eight main components:

```
┌─────────────────────────────────────────────────────────┐
│  resource_genesyscloud_<resource_name>_schema.go        │
├─────────────────────────────────────────────────────────┤
│  1. Package Constants                                   │
│     - ResourceType constant                             │
│     - Other package-level constants                     │
├─────────────────────────────────────────────────────────┤
│  2. Package-Level Variables                             │
│     - Shared variables used across functions            │
├─────────────────────────────────────────────────────────┤
│  3. SetRegistrar() Function                             │
│     - Register Framework resource                       │
│     - Register Framework data source                    │
│     - Register exporter                                 │
├─────────────────────────────────────────────────────────┤
│  4. Helper Functions                                    │
│     - getSdkUtilizationTypes() (dynamic descriptions)   │
│     - Other schema-related helpers                      │
├─────────────────────────────────────────────────────────┤
│  5. Data Source Schema Function                         │
│     - <ResourceName>DataSourceSchema()                  │
│     - Returns datasourceschema.Schema                   │
├─────────────────────────────────────────────────────────┤
│  6. Resource Schema Function                            │
│     - <ResourceName>ResourceSchema()                    │
│     - Returns schema.Schema                             │
├─────────────────────────────────────────────────────────┤
│  7. Exporter Configuration Function                     │
│     - <ResourceName>Exporter()                          │
│     - Returns *resourceExporter.ResourceExporter        │
├─────────────────────────────────────────────────────────┤
│  8. Element Type Helper Functions                       │
│     - routingSkillsElementType()                        │
│     - routingLanguagesElementType()                     │
│     - locationsElementType()                            │
│     - employerInfoElementType()                         │
│     - voicemailUserpoliciesElementType()                │
│     - For cross-package usage in Stage 2                │
└─────────────────────────────────────────────────────────┘
```

**Key Difference from Simple Resources**: Complex resources include element type helper functions and package-level variables for managing complex nested structures.

---

## Component Design

### 1. Package Constants

**Purpose**: Define the resource type identifier and other constants used throughout the package.

**Design**:
```go
const ResourceType = "genesyscloud_<resource_name>"
```

**Example** (user):
```go
const ResourceType = "genesyscloud_user"
```

**Rationale**:
- Single source of truth for resource type name
- Used in registration, tests, and exporter
- Prevents typos and inconsistencies
- Easy to reference across package

---

### 2. Package-Level Variables

**Purpose**: Define shared variables used across schema and utils functions.

**Design Pattern**:
```go
var (
    contactTypeEmail = "EMAIL"
    
    utilizationMediaTypes = map[string]string{
        "call":     "call",
        "callback": "callback",
        "chat":     "chat",
        "email":    "email",
        "message":  "message",
    }
)
```

**Example** (user):
```go
var (
    contactTypeEmail = "EMAIL"
)
```

**Rationale**:
- Reduces magic strings throughout codebase
- Provides single source of truth for constants
- Can be used in schema descriptions and validation
- Improves maintainability

**When to Use**:
- Values used in multiple places (schema, utils, tests)
- Enumerations or lookup maps
- Values that might change in future

---

### 3. SetRegistrar() Function

**Purpose**: Register the Framework resource, data source, and exporter with the provider.

**Design Pattern**:
```go
func SetRegistrar(regInstance registrar.Registrar) {
    regInstance.RegisterFrameworkResource(ResourceType, New<ResourceName>FrameworkResource)
    regInstance.RegisterFrameworkDataSource(ResourceType, New<ResourceName>FrameworkDataSource)
    regInstance.RegisterExporter(ResourceType, <ResourceName>Exporter())
}
```

**Example** (user):
```go
func SetRegistrar(l registrar.Registrar) {
    l.RegisterFrameworkDataSource(ResourceType, NewUserFrameworkDataSource)
    l.RegisterFrameworkResource(ResourceType, NewUserFrameworkResource)
    l.RegisterExporter(ResourceType, UserExporter())
}
```

**Design Decisions**:

| Decision | Rationale |
|----------|-----------|
| Use `RegisterFrameworkResource()` | Distinguishes Framework resources from SDKv2 resources during migration |
| Pass constructor function | Allows provider to create resource instances on demand |
| Register in single function | Centralizes all registration logic for the package |
| Register exporter here | Keeps all registration in one place, even though exporter works with both SDKv2 and Framework |

**Key Points**:
- Constructor functions (`New<ResourceName>FrameworkResource`) are defined in Stage 2
- Exporter configuration is defined in this file (Stage 1)
- Registration happens during provider initialization
- Order doesn't matter (data source can be registered before resource)

---

### 4. Helper Functions

**Purpose**: Provide utility functions for generating dynamic schema content.

**Design Pattern**:
```go
// getSdkUtilizationTypes returns sorted list of utilization media types
// Used in schema descriptions for consistency
func getSdkUtilizationTypes() []string {
    types := make([]string, 0, len(utilizationMediaTypes))
    for t := range utilizationMediaTypes {
        types = append(types, t)
    }
    sort.Strings(types)
    return types
}
```

**Example** (user):
```go
func getSdkUtilizationTypes() []string {
    types := make([]string, 0, len(utilizationMediaTypes))
    for t := range utilizationMediaTypes {
        types = append(types, t)
    }
    sort.Strings(types)
    return types
}
```

**Usage in Schema**:
```go
"interruptible_media_types": schema.SetAttribute{
    Description: fmt.Sprintf("Set of other media types that can interrupt this media type (%s).", 
        strings.Join(getSdkUtilizationTypes(), " | ")),
    Optional:    true,
    ElementType: types.StringType,
}
```

**Rationale**:
- Descriptions stay in sync with code
- Reduces duplication
- Easier to maintain
- Ensures consistency

**When to Create Helper Functions**:
- Dynamic description generation
- Repeated logic in schema definitions
- Complex validation or transformation logic
- Shared between schema and other functions

---

### 5. Data Source Schema Function

**Purpose**: Define the schema for the Terraform data source (lookup by name or other criteria).

**Design Pattern**:
```go
func <ResourceName>DataSourceSchema() datasourceschema.Schema {
    return datasourceschema.Schema{
        Description: "Data source description",
        Attributes: map[string]datasourceschema.Attribute{
            "id": datasourceschema.StringAttribute{
                Description: "Unique identifier",
                Computed:    true,
            },
            "name": datasourceschema.StringAttribute{
                Description: "Resource name for lookup",
                Optional:    true,
            },
            // Additional lookup criteria
        },
    }
}
```

**Example** (user):
```go
func UserDataSourceSchema() datasourceschema.Schema {
    return datasourceschema.Schema{
        Description: "Data source for Genesys Cloud Users. Select a user by email or name. If both email & name are specified, the name won't be used for user lookup",
        Attributes: map[string]datasourceschema.Attribute{
            "id": datasourceschema.StringAttribute{
                Description: "The ID of the user.",
                Computed:    true,
            },
            "email": datasourceschema.StringAttribute{
                Description: "User email.",
                Optional:    true,
            },
            "name": datasourceschema.StringAttribute{
                Description: "User name.",
                Optional:    true,
            },
        },
    }
}
```

**Design Decisions**:

| Decision | Rationale |
|----------|-----------|
| Separate schema function | Data source schema is different from resource schema |
| Use `datasourceschema` package | Framework requires different types for data sources |
| Minimal attributes | Data sources only need lookup criteria and result ID |
| No plan modifiers | Data sources don't have plans (read-only) |
| Multiple lookup criteria | Allows flexible lookups (email OR name) |

**Key Differences from Resource Schema**:
- Uses `datasourceschema.Schema` instead of `schema.Schema`
- Uses `datasourceschema.StringAttribute` instead of `schema.StringAttribute`
- No plan modifiers (data sources are read-only)
- Typically only includes `id` and lookup criteria
- No nested blocks usually needed

**Common Pattern**:
- `id`: Computed (returned after lookup)
- `name`: Optional (user provides for lookup)
- `email`: Optional (alternative lookup criteria)
- Additional attributes can be added if needed for lookup

---

### 6. Resource Schema Function

**Purpose**: Define the complete schema for the Terraform resource.

**Design Pattern**:
```go
func <ResourceName>ResourceSchema() schema.Schema {
    return schema.Schema{
        Description: "Resource description",
        Attributes: map[string]schema.Attribute{
            "id": schema.StringAttribute{
                Description: "Unique identifier",
                Computed:    true,
                PlanModifiers: []planmodifier.String{
                    stringplanmodifier.UseStateForUnknown(),
                },
            },
            "name": schema.StringAttribute{
                Description: "Resource name",
                Required:    true,
            },
            // ... other attributes
        },
        Blocks: map[string]schema.Block{
            // Nested blocks
        },
    }
}
```

**Example** (user - simplified):
```go
func UserResourceSchema() schema.Schema {
    return schema.Schema{
        Description: `Genesys Cloud User.

Export block label: "{email}"`,
        Attributes: map[string]schema.Attribute{
            "id": schema.StringAttribute{
                Description: "The ID of the user.",
                Computed:    true,
                PlanModifiers: []planmodifier.String{
                    stringplanmodifier.UseStateForUnknown(),
                },
            },
            "email": schema.StringAttribute{
                Description: "User's primary email and username.",
                Required:    true,
            },
            "name": schema.StringAttribute{
                Description: "User's full name.",
                Required:    true,
            },
            "state": schema.StringAttribute{
                Description: "User's state (active | inactive). Default is 'active'.",
                Optional:    true,
                Computed:    true,
                Default:     stringdefault.StaticString("active"),
                Validators: []validator.String{
                    stringvalidator.OneOf("active", "inactive"),
                },
            },
            "profile_skills": schema.SetAttribute{
                Description: "Profile skills for this user. If not set, this resource will not manage profile skills.",
                Optional:    true,
                Computed:    true,
                ElementType: types.StringType,
            },
        },
        Blocks: map[string]schema.Block{
            "routing_skills": schema.SetNestedBlock{
                Description: "Skills and proficiencies for this user. If not set, this resource will not manage user skills.",
                PlanModifiers: []planmodifier.Set{
                    setplanmodifier.UseStateForUnknown(),
                },
                NestedObject: schema.NestedBlockObject{
                    Attributes: map[string]schema.Attribute{
                        "skill_id": schema.StringAttribute{
                            Description: "ID of routing skill.",
                            Required:    true,
                        },
                        "proficiency": schema.Float64Attribute{
                            Description: "Rating from 0.0 to 5.0.",
                            Required:    true,
                            Validators: []validator.Float64{
                                float64validator.Between(0, 5),
                            },
                        },
                    },
                },
            },
        },
    }
}
```

#### Attribute Properties

**Required Attributes**:
- Must be provided by user in Terraform configuration
- Cannot be null or omitted
- Example: `name`, `email` (every user needs these)

**Optional Attributes**:
- User can provide or omit
- Can be null
- Example: `department`, `title` (not always needed)

**Computed Attributes**:
- Calculated by provider or API
- User cannot set directly
- Example: `id` (generated by API)

**Optional + Computed Attributes**:
- User can provide OR provider will compute
- If user provides, use that value
- If user omits, provider computes default
- Example: `division_id` (defaults to home division if not specified)

**Optional + Computed + Default Attributes**:
- User can provide OR use default value
- Must be both Optional and Computed when using Default
- Example: `state` (defaults to "active" if not specified)

#### Plan Modifiers

**UseStateForUnknown**:
```go
PlanModifiers: []planmodifier.String{
    stringplanmodifier.UseStateForUnknown(),
}
```

**Purpose**: Preserve state value when new value is unknown during plan.

**When to Use**:
- Computed attributes that don't change on update
- Attributes set by API on create and never change
- Examples: `id`, `division_id` (when computed)
- Nested blocks (emulates SDKv2 Optional + Computed behavior)

**Why Important**:
- Prevents unnecessary resource replacement
- Avoids "known after apply" in plan when value won't actually change
- Improves user experience by showing accurate plan

**Comparison with SDKv2**:
- SDKv2: Used `DiffSuppressFunc` or `CustomizeDiff`
- Framework: Uses plan modifiers (cleaner, more explicit)

**Custom Plan Modifiers**:
```go
PlanModifiers: []planmodifier.String{
    phoneplan.NullIfEmpty{},  // Custom modifier
}
```

**When to Create Custom Plan Modifiers**:
- Standard plan modifiers don't meet your needs
- Complex state manipulation required
- Consistent behavior needed across multiple attributes
- Example: `NullIfEmpty` for `extension_pool_id`

---

### 7. Exporter Configuration Function

**Purpose**: Configure the exporter for this resource, including dependency resolution and special handling.

**Design Pattern**:
```go
func <ResourceName>Exporter() *resourceExporter.ResourceExporter {
    return &resourceExporter.ResourceExporter{
        GetResourcesFunc: provider.GetAllWithPooledClient(GetAll<ResourceName>SDK),
        RefAttrs: map[string]*resourceExporter.RefAttrSettings{
            "<dependency_attr>": {RefType: "<dependency_resource_type>"},
        },
        RemoveIfMissing: map[string][]string{
            "<block_name>": {"<required_field>"},
        },
        AllowEmptyArrays: []string{"<array_field>"},
        AllowZeroValues:  []string{"<numeric_field>"},
    }
}
```

**Example** (user):
```go
func UserExporter() *resourceExporter.ResourceExporter {
    return &resourceExporter.ResourceExporter{
        GetResourcesFunc: provider.GetAllWithPooledClient(GetAllUsersSDK),
        RefAttrs: map[string]*resourceExporter.RefAttrSettings{
            "manager":                                   {RefType: ResourceType},
            "division_id":                               {RefType: "genesyscloud_auth_division"},
            "routing_skills.skill_id":                   {RefType: "genesyscloud_routing_skill"},
            "routing_languages.language_id":             {RefType: "genesyscloud_routing_language"},
            "locations.location_id":                     {RefType: "genesyscloud_location"},
            "addresses.phone_numbers.extension_pool_id": {RefType: "genesyscloud_telephony_providers_edges_extension_pool"},
        },
        RemoveIfMissing: map[string][]string{
            "routing_skills":         {"skill_id"},
            "routing_languages":      {"language_id"},
            "locations":              {"location_id"},
            "voicemail_userpolicies": {"alert_timeout_seconds"},
        },
        AllowEmptyArrays: []string{
            "routing_skills",
            "routing_languages",
        },
        AllowZeroValues: []string{
            "routing_skills.proficiency",
            "routing_languages.proficiency",
            "routing_utilization.call.maximum_capacity",
            "routing_utilization.callback.maximum_capacity",
            "routing_utilization.chat.maximum_capacity",
            "routing_utilization.email.maximum_capacity",
            "routing_utilization.message.maximum_capacity",
        },
    }
}
```

#### GetResourcesFunc Design

**Purpose**: Provide a function to fetch all resources for export.

**Why SDK Version?**:
- Exporter currently uses SDKv2 diagnostics and flat attribute maps
- Framework version (`GetAllUsers`) exists but not used yet
- Phase 1 temporary: Use SDK-compatible function
- Phase 2 future: Migrate exporter to use Framework version

**Pattern**:
```go
GetResourcesFunc: provider.GetAllWithPooledClient(GetAll<ResourceName>SDK)
```

**Explanation**:
- `provider.GetAllWithPooledClient()`: Wrapper that provides pooled API client
- `GetAll<ResourceName>SDK`: Function that fetches all resources (defined in Stage 2)
- Returns `resourceExporter.ResourceIDMetaMap` with flat attributes

#### RefAttrs Design

**Purpose**: Define dependency references for export ordering and HCL generation.

**Pattern**:
```go
RefAttrs: map[string]*resourceExporter.RefAttrSettings{
    "<attribute_name>": {RefType: "<referenced_resource_type>"},
}
```

**Rules**:
- Use dot notation for nested attributes: `"routing_skills.skill_id"`
- Self-references use `ResourceType` constant
- External references use full resource type: `"genesyscloud_routing_skill"`
- Deeply nested references: `"addresses.phone_numbers.extension_pool_id"`

**How It Works**:
1. Exporter reads attribute value from resource
2. Looks up the referenced resource by ID
3. Generates HCL reference: `skill_id = genesyscloud_routing_skill.skill_label.id`
4. Ensures referenced resource is exported before this resource (dependency ordering)

**Common Dependency Attributes**:
- `division_id` → `genesyscloud_auth_division`
- `queue_id` → `genesyscloud_routing_queue`
- `flow_id` → `genesyscloud_flow`
- `skill_id` → `genesyscloud_routing_skill`
- `manager` → `genesyscloud_user` (self-reference)

#### RemoveIfMissing Design

**Purpose**: Remove entire blocks if required fields are missing.

**Pattern**:
```go
RemoveIfMissing: map[string][]string{
    "<block_name>": {"<required_field>"},
}
```

**Example**:
```go
RemoveIfMissing: map[string][]string{
    "routing_skills":         {"skill_id"},
    "routing_languages":      {"language_id"},
    "locations":              {"location_id"},
    "voicemail_userpolicies": {"alert_timeout_seconds"},
}
```

**How It Works**:
- Key = block name
- Value = array of required fields
- If any required field is missing, remove the entire block from export
- Prevents invalid HCL generation

**When to Use**:
- Blocks where certain fields are essential
- API returns partial data that would be invalid in Terraform
- Prevents export errors

#### AllowEmptyArrays Design

**Purpose**: Allow arrays to be explicitly empty (not null).

**Pattern**:
```go
AllowEmptyArrays: []string{"<array_field>"}
```

**Example**:
```go
AllowEmptyArrays: []string{
    "routing_skills",
    "routing_languages",
}
```

**Why Important**:
- Distinguishes between "not set" (null) and "explicitly empty" ([])
- Some resources need to explicitly clear arrays
- Example: User with no skills should export as `routing_skills = []`, not omitted

#### AllowZeroValues Design

**Purpose**: Allow numeric fields to legitimately be 0.

**Pattern**:
```go
AllowZeroValues: []string{"<numeric_field>"}
```

**Example**:
```go
AllowZeroValues: []string{
    "routing_skills.proficiency",
    "routing_languages.proficiency",
    "routing_utilization.call.maximum_capacity",
    "routing_utilization.callback.maximum_capacity",
    "routing_utilization.chat.maximum_capacity",
    "routing_utilization.email.maximum_capacity",
    "routing_utilization.message.maximum_capacity",
}
```

**Why Important**:
- Prevents 0 from being treated as "not set"
- Critical for proficiency ratings (0 is valid)
- Critical for capacity values (0 means disabled)
- Without this, 0 values would be omitted from export

---

### 8. Element Type Helper Functions

**Purpose**: Define reusable element types for complex nested structures.

**Design Pattern**:
```go
func <blockName>ElementType() types.ObjectType {
    return types.ObjectType{
        AttrTypes: map[string]attr.Type{
            "<attribute_name>": types.StringType,
            "<attribute_name>": types.Int64Type,
            // ... other attributes
        },
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

func routingLanguagesElementType() types.ObjectType {
    return types.ObjectType{
        AttrTypes: map[string]attr.Type{
            "language_id": types.StringType,
            "proficiency": types.Int64Type,
        },
    }
}

func locationsElementType() types.ObjectType {
    return types.ObjectType{
        AttrTypes: map[string]attr.Type{
            "location_id": types.StringType,
            "notes":       types.StringType,
        },
    }
}

func employerInfoElementType() types.ObjectType {
    return types.ObjectType{
        AttrTypes: map[string]attr.Type{
            "official_name": types.StringType,
            "employee_id":   types.StringType,
            "employee_type": types.StringType,
            "date_hire":     types.StringType,
        },
    }
}

func voicemailUserpoliciesElementType() types.ObjectType {
    return types.ObjectType{
        AttrTypes: map[string]attr.Type{
            "alert_timeout_seconds":    types.Int64Type,
            "send_email_notifications": types.BoolType,
        },
    }
}
```

**Usage in Stage 2** (flatten/build functions):
```go
// Setting null value with correct type
model.RoutingSkills = types.SetNull(routingSkillsElementType())

// Creating empty set with correct type
emptySet, _ := types.SetValue(routingSkillsElementType(), []attr.Value{})

// Converting SDK objects to Framework types
skillValues := make([]attr.Value, 0)
for _, skill := range sdkSkills {
    skillObj, _ := types.ObjectValue(
        routingSkillsElementType().AttrTypes,
        map[string]attr.Value{
            "skill_id":    types.StringValue(*skill.Id),
            "proficiency": types.Float64Value(*skill.Proficiency),
        },
    )
    skillValues = append(skillValues, skillObj)
}
model.RoutingSkills, _ = types.SetValue(routingSkillsElementType(), skillValues)
```

**Benefits**:
1. **Type Safety**: Ensures consistency between schema and utils
2. **Reusability**: Define once, use in multiple places
3. **Maintainability**: Change type definition in one place
4. **Error Prevention**: Compile-time type checking prevents mismatches

**When to Create Element Type Helpers**:
- For every SetNestedBlock or ListNestedBlock
- For complex nested structures
- When the same structure is used in multiple places
- When Stage 2 utils need to create typed null/empty values

**Naming Convention**:
- Function name: `<blockName>ElementType()`
- Example: `routingSkillsElementType()`, `locationsElementType()`
- Use camelCase, not snake_case
- Match the block name from schema

**Critical Rule**: The `AttrTypes` map MUST exactly match the attributes defined in the corresponding `NestedBlockObject` in the schema. Any mismatch will cause runtime errors.

---

## Complex Schema Patterns

### Pattern 1: Three-Level Nested Blocks

**Use Case**: Deeply nested structures like `routing_utilization` → `call`/`callback`/etc. → attributes.

**Design**:
```go
"routing_utilization": schema.ListNestedBlock{
    Description: "The routing utilization settings for this user.",
    PlanModifiers: []planmodifier.List{
        listplanmodifier.UseStateForUnknown(),
    },
    Validators: []validator.List{
        listvalidator.SizeAtMost(1),
    },
    NestedObject: schema.NestedBlockObject{
        Blocks: map[string]schema.Block{  // Level 2: Use Blocks, not Attributes
            "call": schema.ListNestedBlock{
                Description: "Call media settings. If not set, this reverts to the default media type settings.",
                PlanModifiers: []planmodifier.List{
                    listplanmodifier.UseStateForUnknown(),
                },
                Validators: []validator.List{
                    listvalidator.SizeAtMost(1),
                },
                NestedObject: schema.NestedBlockObject{
                    Attributes: map[string]schema.Attribute{  // Level 3: Final attributes
                        "maximum_capacity": schema.Int64Attribute{
                            Description: "Maximum capacity of conversations of this media type. Value must be between 0 and 25.",
                            Required:    true,
                            Validators: []validator.Int64{
                                int64validator.Between(0, 25),
                            },
                        },
                        "interruptible_media_types": schema.SetAttribute{
                            Description: fmt.Sprintf("Set of other media types that can interrupt this media type (%s).", 
                                strings.Join(getSdkUtilizationTypes(), " | ")),
                            Optional:    true,
                            ElementType: types.StringType,
                        },
                        "include_non_acd": schema.BoolAttribute{
                            Description: "Block this media type when on a non-ACD conversation.",
                            Optional:    true,
                            Computed:    true,
                            Default:     booldefault.StaticBool(false),
                        },
                    },
                },
            },
            "callback": schema.ListNestedBlock{
                // Same structure as call
            },
            "message": schema.ListNestedBlock{
                // Same structure as call
            },
            "email": schema.ListNestedBlock{
                // Same structure as call
            },
            "chat": schema.ListNestedBlock{
                // Same structure as call
            },
        },
    },
}
```

**Key Points**:
- Level 1: `routing_utilization` (ListNestedBlock)
- Level 2: `call`, `callback`, etc. (ListNestedBlock within `Blocks` map)
- Level 3: `maximum_capacity`, `include_non_acd`, etc. (Attributes)
- Each level can have its own plan modifiers and validators
- Use `Blocks` map for nested blocks, `Attributes` map for final attributes
- **Critical**: When nesting blocks within blocks, use `Blocks: map[string]schema.Block{}` in the parent's `NestedBlockObject`, not `Attributes`

**Common Mistake**:
```go
// ❌ WRONG - Using Attributes for nested blocks
NestedObject: schema.NestedBlockObject{
    Attributes: map[string]schema.Attribute{
        "call": schema.ListNestedBlock{...},  // This won't work!
    },
}

// ✅ CORRECT - Use Blocks for nested blocks
NestedObject: schema.NestedBlockObject{
    Blocks: map[string]schema.Block{
        "call": schema.ListNestedBlock{...},
    },
}
```

---

### Pattern 2: Nested Blocks Within Blocks (Two Levels)

**Use Case**: Structures like `addresses` → `phone_numbers`/`other_emails`.

**Design**:
```go
"addresses": schema.ListNestedBlock{
    Description: "The address settings for this user. If not set, this resource will not manage addresses.",
    PlanModifiers: []planmodifier.List{
        listplanmodifier.UseStateForUnknown(),
    },
    Validators: []validator.List{
        listvalidator.SizeAtMost(1),
    },
    NestedObject: schema.NestedBlockObject{
        Blocks: map[string]schema.Block{  // Note: Blocks, not Attributes
            "other_emails": schema.SetNestedBlock{
                Description: "Other Email addresses for this user.",
                NestedObject: schema.NestedBlockObject{
                    Attributes: map[string]schema.Attribute{
                        "address": schema.StringAttribute{
                            Description: "Email address.",
                            Required:    true,
                        },
                        "type": schema.StringAttribute{
                            Description: "Type of email address (WORK | HOME).",
                            Optional:    true,
                            Computed:    true,
                            Default:     stringdefault.StaticString("WORK"),
                            Validators: []validator.String{
                                stringvalidator.OneOf("WORK", "HOME"),
                            },
                        },
                    },
                },
            },
            "phone_numbers": schema.SetNestedBlock{
                Description: "Phone number addresses for this user.",
                NestedObject: schema.NestedBlockObject{
                    Attributes: map[string]schema.Attribute{
                        "number": schema.StringAttribute{
                            Description: "Phone number. Phone number must be in an E.164 number format.",
                            Optional:    true,
                            Validators:  []validator.String{validators.FWValidatePhoneNumber()},
                        },
                        "media_type": schema.StringAttribute{
                            Description: "Media type of phone number (SMS | PHONE).",
                            Optional:    true,
                            Computed:    true,
                            Default:     stringdefault.StaticString("PHONE"),
                            Validators: []validator.String{
                                stringvalidator.OneOf("PHONE", "SMS"),
                            },
                        },
                        "type": schema.StringAttribute{
                            Description: "Type of number (WORK | WORK2 | WORK3 | WORK4 | HOME | MOBILE | OTHER).",
                            Optional:    true,
                            Computed:    true,
                            Default:     stringdefault.StaticString("WORK"),
                            Validators: []validator.String{
                                stringvalidator.OneOf("WORK", "WORK2", "WORK3", "WORK4", "HOME", "MOBILE", "OTHER"),
                            },
                        },
                        "extension": schema.StringAttribute{
                            Description: "Phone number extension",
                            Optional:    true,
                        },
                        "extension_pool_id": schema.StringAttribute{
                            Description:   "Id of the extension pool which contains this extension.",
                            Optional:      true,
                            PlanModifiers: []planmodifier.String{phoneplan.NullIfEmpty{}},
                        },
                    },
                },
            },
        },
    },
}
```

**Key Points**:
- Parent block (`addresses`) uses `Blocks` map for nested blocks
- Child blocks (`phone_numbers`, `other_emails`) use `Attributes` map for final attributes
- Each nested block can have its own structure
- Plan modifiers and validators at each level

---

### Pattern 3: SetNestedBlock with Managed Attribute Pattern

**Use Case**: Optional collections that may or may not be managed by the resource.

**Design**:
```go
"routing_skills": schema.SetNestedBlock{
    Description: "Skills and proficiencies for this user. If not set, this resource will not manage user skills.",
    // NO Optional/Required field at block level
    PlanModifiers: []planmodifier.Set{
        setplanmodifier.UseStateForUnknown(), // Emulates SDKv2 Computed behavior
    },
    NestedObject: schema.NestedBlockObject{
        Attributes: map[string]schema.Attribute{
            "skill_id": schema.StringAttribute{
                Description: "ID of routing skill.",
                Required:    true,
            },
            "proficiency": schema.Float64Attribute{
                Description: "Rating from 0.0 to 5.0 on how competent an agent is for a particular skill.",
                Required:    true,
                Validators: []validator.Float64{
                    float64validator.Between(0, 5),
                },
            },
        },
    },
}
```

**Key Points**:
- SetNestedBlock/ListNestedBlock do NOT support `Optional` or `Required` at the block level
- Use `UseStateForUnknown()` plan modifier to emulate SDKv2's `Optional + Computed` behavior
- Inner attributes CAN be `Required` or `Optional`
- Description explicitly states "If not set, this resource will not manage..." (managed attribute pattern)

**Why UseStateForUnknown?**:
- When block is omitted from config, Terraform preserves existing state
- Emulates SDKv2 behavior where `Optional: true, Computed: true` meant "keep existing value if not specified"
- Without this, omitting the block would set it to null, potentially deleting existing data

---

### Pattern 4: ListNestedBlock with MaxItems: 1

**Use Case**: Single-item blocks that use list syntax (common for API compatibility).

**Design**:
```go
"employer_info": schema.ListNestedBlock{
    Description: "The employer info for this user. If not set, this resource will not manage employer info.",
    PlanModifiers: []planmodifier.List{
        listplanmodifier.UseStateForUnknown(),
    },
    Validators: []validator.List{
        listvalidator.SizeAtMost(1), // Replaces SDKv2 MaxItems
    },
    NestedObject: schema.NestedBlockObject{
        Attributes: map[string]schema.Attribute{
            "official_name": schema.StringAttribute{
                Description: "User's official name.",
                Optional:    true,
            },
            "employee_id": schema.StringAttribute{
                Description: "Employee ID.",
                Optional:    true,
            },
            "employee_type": schema.StringAttribute{
                Description: "Employee type (Full-time | Part-time | Contractor).",
                Optional:    true,
                Validators: []validator.String{
                    stringvalidator.OneOf("Full-time", "Part-time", "Contractor"),
                },
            },
            "date_hire": schema.StringAttribute{
                Description: "Hiring date. Dates must be an ISO-8601 string. For example: yyyy-MM-dd.",
                Optional:    true,
                Validators: []validator.String{
                    validators.FWValidateDate(),
                },
            },
        },
    },
}
```

**Key Points**:
- `MaxItems: 1` in SDKv2 → `listvalidator.SizeAtMost(1)` in Framework
- Still uses ListNestedBlock (not a single object)
- All inner attributes are Optional (user can provide any combination)
- Custom validators for specific fields (date format, enum values)

---

### Pattern 5: Optional + Computed + Default

**Use Case**: Attributes with default values that users can override.

**Design**:
```go
"state": schema.StringAttribute{
    Description: "User's state (active | inactive). Default is 'active'.",
    Optional:    true,
    Computed:    true,  // REQUIRED when using Default
    Default:     stringdefault.StaticString("active"),
    Validators: []validator.String{
        stringvalidator.OneOf("active", "inactive"),
    },
}

"acd_auto_answer": schema.BoolAttribute{
    Description: "Enable ACD auto-answer.",
    Optional:    true,
    Computed:    true,  // REQUIRED when using Default
    Default:     booldefault.StaticBool(false),
}
```

**Critical Rule**: In Plugin Framework, attributes with `Default` MUST be both `Optional` and `Computed`.

**Why Computed is Required**:
- Framework needs to know the value can be computed (by the default)
- Without Computed, Framework will error
- This is different from SDKv2 where Default didn't require Computed

**Common Mistake**:
```go
// ❌ WRONG - Will cause error
"state": schema.StringAttribute{
    Optional: true,
    Default:  stringdefault.StaticString("active"),
}

// ✅ CORRECT
"state": schema.StringAttribute{
    Optional: true,
    Computed: true,  // REQUIRED
    Default:  stringdefault.StaticString("active"),
}
```

---

### Pattern 6: Optional + Computed Without Default

**Use Case**: Attributes that can be set by user OR computed by API, with no specific default.

**Design**:
```go
"division_id": schema.StringAttribute{
    Description: "The division to which this user will belong. If not set, the home division will be used.",
    Optional:    true,
    Computed:    true,
    // NO Default - API will compute if not provided
    // NO UseStateForUnknown - we want API to recompute if config changes
}

"send_email_notifications": schema.BoolAttribute{
    Description: "Whether email notifications are sent to the user when a new voicemail is received.",
    Optional:    true,
    Computed:    true,
    // NO Default - API will compute if not provided
}
```

**Key Points**:
- User can provide value OR API will compute
- No specific default value to set
- API determines the value if not provided
- Different from Pattern 5 (which has explicit Default)
- Do NOT add `UseStateForUnknown()` plan modifier - we want the API to recompute if config changes

**When to Use**:
- API has server-side defaults
- Default value depends on other factors (e.g., organization settings)
- Value is computed based on context

---

### Pattern 7: Managed Attribute Pattern (Set/List of Primitives)

**Use Case**: Collections that may or may not be managed by the resource.

**Design**:
```go
"profile_skills": schema.SetAttribute{
    Description: "Profile skills for this user. If not set, this resource will not manage profile skills.",
    Optional:    true,
    Computed:    true,
    ElementType: types.StringType,
}

"certifications": schema.SetAttribute{
    Description: "Certifications for this user. If not set, this resource will not manage certifications.",
    Optional:    true,
    Computed:    true,
    ElementType: types.StringType,
}
```

**Key Points**:
- Optional + Computed (no plan modifier needed for primitive sets)
- Description explicitly states "If not set, this resource will not manage..."
- Allows resource to ignore certain attributes if not specified
- User can explicitly set to empty array `[]` to clear values

**Difference from Nested Blocks**:
- Primitive sets (strings, ints) use SetAttribute
- Nested objects use SetNestedBlock
- SetAttribute doesn't need UseStateForUnknown plan modifier

---

### Pattern 8: Custom Validators

**Use Case**: Complex validation logic beyond standard validators.

**Design**:
```go
"number": schema.StringAttribute{
    Description: "Phone number. Phone number must be in an E.164 number format.",
    Optional:    true,
    Validators:  []validator.String{
        validators.FWValidatePhoneNumber(),  // Custom validator
    },
}

"date_hire": schema.StringAttribute{
    Description: "Hiring date. Dates must be an ISO-8601 string. For example: yyyy-MM-dd.",
    Optional:    true,
    Validators: []validator.String{
        validators.FWValidateDate(),  // Custom validator
    },
}
```

**When to Create Custom Validators**:
- Standard validators don't meet your needs
- Complex validation logic (phone numbers, dates, etc.)
- Validation requires external data or API calls
- Consistent validation needed across multiple resources

**Location**: Create in `genesyscloud/validators` package for reusability.

**Implementation Pattern**:
```go
// In genesyscloud/validators package
func FWValidatePhoneNumber() validator.String {
    return &phoneNumberValidator{}
}

type phoneNumberValidator struct{}

func (v phoneNumberValidator) Description(ctx context.Context) string {
    return "value must be a valid E.164 phone number"
}

func (v phoneNumberValidator) MarkdownDescription(ctx context.Context) string {
    return "value must be a valid E.164 phone number"
}

func (v phoneNumberValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
    if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
        return
    }
    
    // Validation logic here
    // Add diagnostics to resp.Diagnostics if validation fails
}
```

---

### Pattern 9: Dynamic Description Generation

**Use Case**: Descriptions that include dynamic content from code.

**Design**:
```go
// Package-level variable
var (
    utilizationMediaTypes = map[string]string{
        "call":     "call",
        "callback": "callback",
        "chat":     "chat",
        "email":    "email",
        "message":  "message",
    }
)

// Helper function
func getSdkUtilizationTypes() []string {
    types := make([]string, 0, len(utilizationMediaTypes))
    for t := range utilizationMediaTypes {
        types = append(types, t)
    }
    sort.Strings(types)
    return types
}

// Usage in schema
"interruptible_media_types": schema.SetAttribute{
    Description: fmt.Sprintf("Set of other media types that can interrupt this media type (%s).", 
        strings.Join(getSdkUtilizationTypes(), " | ")),
    Optional:    true,
    ElementType: types.StringType,
}
```

**Benefits**:
- Descriptions stay in sync with code
- Reduces duplication
- Easier to maintain
- Ensures consistency

**Requirements**:
- Import `fmt` and `strings` packages
- Define helper function in schema file
- Use package-level variables for data

---

### Pattern 10: Custom Plan Modifiers

**Use Case**: Complex state manipulation beyond standard plan modifiers.

**Design**:
```go
"extension_pool_id": schema.StringAttribute{
    Description:   "Id of the extension pool which contains this extension.",
    Optional:      true,
    PlanModifiers: []planmodifier.String{
        phoneplan.NullIfEmpty{},  // Custom modifier
    },
}
```

**Custom Plan Modifier Implementation** (in `genesyscloud/util/phoneplan` package):
```go
package phoneplan

import (
    "context"
    "github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
    "github.com/hashicorp/terraform-plugin-framework/types"
)

// NullIfEmpty sets the value to null if it's an empty string
type NullIfEmpty struct{}

func (m NullIfEmpty) Description(ctx context.Context) string {
    return "Sets the value to null if it's an empty string"
}

func (m NullIfEmpty) MarkdownDescription(ctx context.Context) string {
    return "Sets the value to null if it's an empty string"
}

func (m NullIfEmpty) PlanModifyString(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
    if req.PlanValue.IsNull() || req.PlanValue.IsUnknown() {
        return
    }
    
    if req.PlanValue.ValueString() == "" {
        resp.PlanValue = types.StringNull()
    }
}
```

**When to Create Custom Plan Modifiers**:
- Standard plan modifiers don't meet your needs
- Complex state manipulation required
- Consistent behavior needed across multiple attributes
- Workarounds for Framework limitations

**Location**: Create in `genesyscloud/util/<package>` for reusability.

---

## Migration Considerations

### Set Identity Behavior Change

**SDKv2 Behavior**:
```go
Set: &schema.Set{
    F: func(v interface{}) int {
        // Custom hash function
        // Could exclude fields like extension_pool_id from hash
        // Changes to excluded fields didn't create new set elements
    },
}
```

**Plugin Framework Behavior**:
- ALL fields are included in Set element identity
- No custom hash function support
- Changes to any field (including previously excluded fields) will cause Set element replacement

**Example Impact** (user resource):
```go
"extension_pool_id": schema.StringAttribute{
    //TODO
    //Issue: In SDKv2 hashing you explicitly removed extension_pool_id before hashing
    // a phone element, so pool changes didn't create new set elements or diffs.
    // In PF, set element identity is the full object value; including extension_pool_id
    // means a pool change will look like an element replacement and cause perpetual diffs.
    Description:   "Id of the extension pool which contains this extension.",
    Optional:      true,
    PlanModifiers: []planmodifier.String{phoneplan.NullIfEmpty{}},
}
```

**Mitigation Strategies**:
1. **Document the behavior change**: Users should be aware that pool changes will cause element replacement
2. **Use custom plan modifiers**: `NullIfEmpty` helps minimize impact by treating empty strings as null
3. **Accept the limitation**: This is an acceptable behavior change that matches AWS provider patterns during PF migration
4. **Future consideration**: Alternative approaches (like top-level computed maps) can be considered later

**Status**: Documented as TC5 in Constraints section. Acceptable behavior change.

---

### Future Enhancements (TODO Comments)

#### Alternative Approach for Extension Pool IDs

**Current Implementation**:
```go
"extension_pool_id": schema.StringAttribute{
    Optional:      true,
    PlanModifiers: []planmodifier.String{phoneplan.NullIfEmpty{}},
}
```

**Alternative Approach (Commented Out)**:
```go
//TODO
/* will handel this condition later
"phone_extension_pools": schema.MapAttribute{
    ElementType: types.StringType,
    Computed:    true,
    Description: "Id of the extension pool which contains this extension." +
        "Computed mapping of phone identity keys to  (MEDIA|TYPE|E164|EXT) to extension_pool_id." +
        "Used internally to prevent diffs when pool assignments change.",
},
*/
```

**Rationale for Alternative**:
- Top-level computed map would be identity-insensitive
- Pool changes wouldn't affect phone_numbers Set identity
- More complex to implement and maintain
- Current approach is simpler and acceptable

**Status**: Commented out for potential future optimization. Current implementation is sufficient.

---

#### Plan Modifiers vs Validators for Phone Numbers

**Current Implementation**:
```go
"number": schema.StringAttribute{
    Description: "Phone number. Phone number must be in an E.164 number format.",
    Optional:    true,
    Validators:  []validator.String{validators.FWValidatePhoneNumber()},
    //TODO
    // PlanModifiers for now as Validators will do all the required check
    // if required we can review and enable it later.
    // Safe E.164 canonicalization: noop if null/unknown/empty or parse fails
    //PlanModifiers: []planmodifier.String{phoneplan.E164{DefaultRegion: "US"}},
}
```

**Decision**:
- Use validators for validation (current)
- Plan modifiers for normalization (commented out for now)
- Can be enabled later if E.164 canonicalization is needed
- Validators are sufficient for current requirements

**Rationale**:
- Validation is the primary concern
- Normalization can be added later if needed
- Simpler implementation for initial migration
- Reduces risk of unexpected behavior changes

**Status**: Validators only. Plan modifiers commented out for potential future enhancement.

---

### Import Organization

**Recommended Import Order**:
```go
import (
    "fmt"      // Standard library
    "strings"  // Standard library
    
    // Framework validators (grouped)
    "github.com/hashicorp/terraform-plugin-framework-validators/float64validator"
    "github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
    listvalidator "github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
    "github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
    
    // Framework schema packages
    datasourceschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
    "github.com/hashicorp/terraform-plugin-framework/resource/schema"
    
    // Framework defaults
    "github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
    "github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
    
    // Framework plan modifiers
    "github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
    "github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
    "github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
    "github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
    
    // Framework validators and types
    "github.com/hashicorp/terraform-plugin-framework/schema/validator"
    "github.com/hashicorp/terraform-plugin-framework/types"
    
    // Internal packages
    "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
    resourceExporter "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_exporter"
    registrar "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_register"
    "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/phoneplan"
    "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/validators"
)
```

**Grouping**:
1. Standard library
2. Framework validators
3. Framework schema packages
4. Framework defaults
5. Framework plan modifiers
6. Framework validators and types
7. Internal packages

**Aliasing**:
- `datasourceschema` - Avoids conflict with resource schema
- `listvalidator` - Clearer than importing as `.`
- `resourceExporter` - Clearer than `resource_exporter`
- `registrar` - Clearer than `resource_register`

**Rationale**:
- Logical grouping improves readability
- Aliases prevent naming conflicts
- Consistent with Go best practices
- Easier to maintain

---

## SDKv2 vs Plugin Framework Comparison

### Schema Definition

**SDKv2**:
```go
func ResourceUser() *schema.Resource {
    return &schema.Resource{
        Schema: map[string]*schema.Schema{
            "id": {
                Type:     schema.TypeString,
                Computed: true,
            },
            "name": {
                Type:     schema.TypeString,
                Required: true,
            },
            "state": {
                Type:     schema.TypeString,
                Optional: true,
                Default:  "active",
            },
            "routing_skills": {
                Type:     schema.TypeSet,
                Optional: true,
                Computed: true,
                Elem: &schema.Resource{
                    Schema: map[string]*schema.Schema{
                        "skill_id": {
                            Type:     schema.TypeString,
                            Required: true,
                        },
                        "proficiency": {
                            Type:         schema.TypeFloat,
                            Required:     true,
                            ValidateFunc: validation.FloatBetween(0, 5),
                        },
                    },
                },
            },
        },
    }
}
```

**Plugin Framework**:
```go
func UserResourceSchema() schema.Schema {
    return schema.Schema{
        Description: "Genesys Cloud User.",
        Attributes: map[string]schema.Attribute{
            "id": schema.StringAttribute{
                Description: "The ID of the user.",
                Computed:    true,
                PlanModifiers: []planmodifier.String{
                    stringplanmodifier.UseStateForUnknown(),
                },
            },
            "name": schema.StringAttribute{
                Description: "User's full name.",
                Required:    true,
            },
            "state": schema.StringAttribute{
                Description: "User's state (active | inactive). Default is 'active'.",
                Optional:    true,
                Computed:    true,
                Default:     stringdefault.StaticString("active"),
                Validators: []validator.String{
                    stringvalidator.OneOf("active", "inactive"),
                },
            },
        },
        Blocks: map[string]schema.Block{
            "routing_skills": schema.SetNestedBlock{
                Description: "Skills and proficiencies for this user. If not set, this resource will not manage user skills.",
                PlanModifiers: []planmodifier.Set{
                    setplanmodifier.UseStateForUnknown(),
                },
                NestedObject: schema.NestedBlockObject{
                    Attributes: map[string]schema.Attribute{
                        "skill_id": schema.StringAttribute{
                            Description: "ID of routing skill.",
                            Required:    true,
                        },
                        "proficiency": schema.Float64Attribute{
                            Description: "Rating from 0.0 to 5.0.",
                            Required:    true,
                            Validators: []validator.Float64{
                                float64validator.Between(0, 5),
                            },
                        },
                    },
                },
            },
        },
    }
}
```

### Key Differences

| Aspect | SDKv2 | Plugin Framework |
|--------|-------|------------------|
| Schema type | `map[string]*schema.Schema` | `schema.Schema` struct |
| Attribute type | `*schema.Schema` | `schema.StringAttribute`, etc. |
| Type definition | `Type: schema.TypeString` | `schema.StringAttribute` |
| Description | Optional, often omitted | Required, always included |
| Plan modifiers | `DiffSuppressFunc`, `CustomizeDiff` | `PlanModifiers` array |
| Validation | `ValidateFunc` | `Validators` array |
| Type safety | Runtime type checking | Compile-time type safety |
| Nested objects | `Elem: &schema.Resource{}` | `SetNestedBlock` / `ListNestedBlock` |
| Default values | `Default: "value"` | `Default: stringdefault.StaticString("value")` |
| Computed + Default | `Default` only | `Optional + Computed + Default` |
| Set hash function | Custom `Set.F` function | No custom hash (all fields used) |

### Nested Block Differences

**SDKv2**:
```go
"routing_skills": {
    Type:     schema.TypeSet,
    Optional: true,
    Computed: true,
    Elem: &schema.Resource{
        Schema: map[string]*schema.Schema{
            "skill_id": {...},
        },
    },
}
```

**Plugin Framework**:
```go
"routing_skills": schema.SetNestedBlock{
    // NO Optional field - use plan modifier instead
    PlanModifiers: []planmodifier.Set{
        setplanmodifier.UseStateForUnknown(),
    },
    NestedObject: schema.NestedBlockObject{
        Attributes: map[string]schema.Attribute{
            "skill_id": {...},
        },
    },
}
```

**Key Difference**: SetNestedBlock/ListNestedBlock do NOT support `Optional` or `Required` at the block level. Use plan modifiers to emulate SDKv2 behavior.

---

## Design Patterns and Best Practices

### Pattern: Computed Attributes with UseStateForUnknown

**When to Use**:
- Attributes that are computed on create and don't change
- Attributes that have default values computed by API
- Attributes that should preserve state value during update

**Example**:
```go
"id": schema.StringAttribute{
    Computed: true,
    PlanModifiers: []planmodifier.String{
        stringplanmodifier.UseStateForUnknown(),
    },
},
```

**Why**:
- Prevents "known after apply" in plan when value won't change
- Avoids unnecessary resource replacement
- Improves plan accuracy

---

### Pattern: Separate Resource and Data Source Schemas

**Why Separate**:
- Different attribute requirements
- Different Framework types
- Data sources are simpler (lookup only)
- Clearer separation of concerns

**Pattern**:
```go
// Resource schema - full attributes
func <ResourceName>ResourceSchema() schema.Schema { ... }

// Data source schema - minimal attributes
func <ResourceName>DataSourceSchema() datasourceschema.Schema { ... }
```

---

### Pattern: Dependency Reference Configuration

**Purpose**: Enable exporter to resolve dependencies and generate correct HCL.

**Pattern**:
```go
RefAttrs: map[string]*resourceExporter.RefAttrSettings{
    "<attribute_name>": {RefType: "<resource_type>"},
}
```

**Guidelines**:
- Include all attributes that reference other resources
- Use exact attribute name from schema
- Use exact Terraform resource type
- Use dot notation for nested attributes
- Order doesn't matter (exporter handles ordering)

**Example with Multiple Dependencies**:
```go
RefAttrs: map[string]*resourceExporter.RefAttrSettings{
    "manager":                                   {RefType: ResourceType},
    "division_id":                               {RefType: "genesyscloud_auth_division"},
    "routing_skills.skill_id":                   {RefType: "genesyscloud_routing_skill"},
    "routing_languages.language_id":             {RefType: "genesyscloud_routing_language"},
    "addresses.phone_numbers.extension_pool_id": {RefType: "genesyscloud_telephony_providers_edges_extension_pool"},
}
```

---

### Pattern: Element Type Helper Functions

**Convention**: `<blockName>ElementType()`

**Examples**:
- `routingSkillsElementType()`
- `routingLanguagesElementType()`
- `locationsElementType()`
- `employerInfoElementType()`

**Usage**:
1. Define in schema file (Stage 1)
2. Use in schema definitions (Stage 1)
3. Use in flatten/build functions (Stage 2)
4. Use in tests (Stage 3)

**Benefits**:
- Type safety across all stages
- Single source of truth for element types
- Prevents type mismatches
- Easier to maintain

---

### Pattern: Managed Attribute Descriptions

**Convention**: Explicitly state when attributes are optionally managed.

**Pattern**:
```go
Description: "<description>. If not set, this resource will not manage <attribute>."
```

**Examples**:
```go
"routing_skills": schema.SetNestedBlock{
    Description: "Skills and proficiencies for this user. If not set, this resource will not manage user skills.",
}

"profile_skills": schema.SetAttribute{
    Description: "Profile skills for this user. If not set, this resource will not manage profile skills.",
}
```

**Why Important**:
- Clarifies optional management behavior
- Helps users understand when to omit vs set to empty
- Documents the "managed attribute" pattern
- Improves user experience

---

## Common Migration Pitfalls

### Pitfall 1: Forgetting Computed with Default

```go
// ❌ WRONG - Will cause error
"state": schema.StringAttribute{
    Optional: true,
    Default:  stringdefault.StaticString("active"),
}

// ✅ CORRECT
"state": schema.StringAttribute{
    Optional: true,
    Computed: true,  // REQUIRED
    Default:  stringdefault.StaticString("active"),
}
```

**Why**: Framework requires `Computed: true` when using `Default`.

---

### Pitfall 2: Using Optional on Nested Blocks

```go
// ❌ WRONG - SetNestedBlock doesn't support Optional
"routing_skills": schema.SetNestedBlock{
    Optional: true,  // This field doesn't exist!
}

// ✅ CORRECT - Use plan modifier instead
"routing_skills": schema.SetNestedBlock{
    PlanModifiers: []planmodifier.Set{
        setplanmodifier.UseStateForUnknown(),
    },
}
```

**Why**: SetNestedBlock/ListNestedBlock do NOT have `Optional` or `Required` fields.

---

### Pitfall 3: Missing UseStateForUnknown on Computed Attributes

```go
// ❌ WRONG - Will cause perpetual diffs
"id": schema.StringAttribute{
    Computed: true,
}

// ✅ CORRECT
"id": schema.StringAttribute{
    Computed: true,
    PlanModifiers: []planmodifier.String{
        stringplanmodifier.UseStateForUnknown(),
    },
}
```

**Why**: Without `UseStateForUnknown`, computed attributes show "known after apply" unnecessarily.

---

### Pitfall 4: Wrong Type for Nested Blocks

```go
// ❌ WRONG - Using Attributes for nested blocks
NestedObject: schema.NestedBlockObject{
    Attributes: map[string]schema.Attribute{
        "phone_numbers": schema.SetNestedBlock{...},  // This won't work!
    },
}

// ✅ CORRECT - Use Blocks for nested blocks
NestedObject: schema.NestedBlockObject{
    Blocks: map[string]schema.Block{
        "phone_numbers": schema.SetNestedBlock{...},
    },
}
```

**Why**: Nested blocks must be in `Blocks` map, not `Attributes` map.

---

### Pitfall 5: Element Type Mismatch

```go
// ❌ WRONG - Element type doesn't match schema
func routingSkillsElementType() types.ObjectType {
    return types.ObjectType{
        AttrTypes: map[string]attr.Type{
            "skill_id":    types.StringType,
            "proficiency": types.Int64Type,  // WRONG - schema uses Float64
        },
    }
}

// ✅ CORRECT - Element type matches schema exactly
func routingSkillsElementType() types.ObjectType {
    return types.ObjectType{
        AttrTypes: map[string]attr.Type{
            "skill_id":    types.StringType,
            "proficiency": types.Float64Type,  // Matches schema
        },
    }
}
```

**Why**: Element type MUST exactly match the schema definition. Mismatches cause runtime errors.

---

### Pitfall 6: Missing Validators on Nested Attributes

```go
// ❌ WRONG - Missing validator
"proficiency": schema.Float64Attribute{
    Required: true,
}

// ✅ CORRECT - Include validator
"proficiency": schema.Float64Attribute{
    Required: true,
    Validators: []validator.Float64{
        float64validator.Between(0, 5),
    },
}
```

**Why**: Validators from SDKv2 must be migrated to Framework validators.

---

### Pitfall 7: Incorrect RefAttrs Syntax

```go
// ❌ WRONG - Missing dot notation for nested attributes
RefAttrs: map[string]*resourceExporter.RefAttrSettings{
    "skill_id": {RefType: "genesyscloud_routing_skill"},
}

// ✅ CORRECT - Use dot notation for nested attributes
RefAttrs: map[string]*resourceExporter.RefAttrSettings{
    "routing_skills.skill_id": {RefType: "genesyscloud_routing_skill"},
}
```

**Why**: Nested attributes require dot notation to specify the full path.

---

### Pitfall 8: Missing AllowZeroValues for Proficiency

```go
// ❌ WRONG - Zero proficiency will be omitted from export
AllowZeroValues: []string{
    // Missing proficiency fields
}

// ✅ CORRECT - Include proficiency fields
AllowZeroValues: []string{
    "routing_skills.proficiency",
    "routing_languages.proficiency",
}
```

**Why**: Without `AllowZeroValues`, 0 proficiency ratings will be treated as "not set" and omitted from export.

---

## Testing Considerations

### Schema Validation

**What to Test** (in Stage 3):
- Schema compiles without errors
- All attributes are accessible
- Plan modifiers work correctly
- Validation rules work correctly
- Nested blocks are properly structured
- Element types match schema definitions

**How to Test**:
- Create resource with all attributes
- Create resource with minimal attributes
- Update attributes
- Verify plan shows correct changes
- Test nested block creation and updates
- Test validation rules trigger correctly

### Element Type Validation

**What to Test**:
- Element type helpers return correct types
- Types match schema definitions exactly
- Null values can be created with correct types
- Empty collections can be created with correct types

**How to Test**:
- Call element type helper functions
- Verify returned types match schema
- Create null/empty values using helpers
- Verify no type mismatch errors

---

## Future Considerations

### Phase 2: Native Framework Export

**Current State** (Phase 1):
- Exporter uses `GetAll<ResourceName>SDK()` with SDK diagnostics
- Returns flat attribute maps for dependency resolution
- Temporary compatibility layer

**Future State** (Phase 2):
- Exporter uses `GetAll<ResourceName>()` with Framework diagnostics
- Works natively with Framework types
- No flat attribute map conversion needed
- Cleaner, more maintainable code

**Migration Path**:
1. Complete all resource migrations (Stages 1-4)
2. Update exporter to work with Framework types
3. Remove SDK-compatible functions
4. Remove export utility files

**Impact on Schema**:
- Schema file remains unchanged
- Exporter configuration updated to use Framework function
- No changes to schema definitions

---

### Alternative Approaches for Set Identity

**Current Limitation**: Plugin Framework uses ALL fields for Set element identity.

**Potential Future Solutions**:
1. **Top-level computed maps**: Store identity-sensitive data separately
2. **Custom equality functions**: If Framework adds support in future
3. **Restructure schema**: Avoid Sets for objects with identity-insensitive fields

**Current Mitigation**:
- Document behavior change
- Use custom plan modifiers (e.g., `NullIfEmpty`)
- Accept as acceptable behavior change

---

## Summary

### Key Design Decisions

1. **Separate Schema File**: Isolates schema definitions for clarity and maintainability
2. **Framework-Native Patterns**: Uses Plugin Framework idioms (plan modifiers, type safety)
3. **Element Type Helpers**: Defines reusable element types for complex nested structures
4. **Backward Compatibility**: Preserves all attribute names and behavior (with documented exceptions)
5. **Export Compatibility**: Uses SDK-compatible functions during migration (Phase 1)
6. **Managed Attribute Pattern**: Explicitly documents optional management in descriptions
7. **Custom Validators and Plan Modifiers**: Extends Framework capabilities for complex requirements

### Schema File Structure

```
resource_genesyscloud_<resource_name>_schema.go
├── Package Constants (ResourceType, etc.)
├── Package-Level Variables (shared data)
├── SetRegistrar() - Register resource, data source, exporter
├── Helper Functions (dynamic descriptions, etc.)
├── <ResourceName>DataSourceSchema() - Data source schema definition
├── <ResourceName>ResourceSchema() - Resource schema definition
├── <ResourceName>Exporter() - Exporter configuration
└── Element Type Helper Functions - For complex nested structures
```

### Complex Resource Characteristics

**What Makes a Resource Complex**:
- Multiple nested blocks (2+ levels)
- Three-level nesting (block → block → attributes)
- Custom validators and plan modifiers
- Element type helper functions
- Managed attribute patterns
- Dynamic description generation
- Package-level variables and constants
- Complex exporter configuration (RemoveIfMissing, AllowZeroValues, etc.)

**Differences from Simple Resources**:
- More component types (8 vs 6)
- Element type helper functions
- Package-level variables
- More complex exporter configuration
- Custom plan modifiers
- Dynamic descriptions

### Next Steps

After completing Stage 1 schema migration:
1. Review schema definitions for accuracy
2. Verify all attributes are correctly defined
3. Confirm exporter configuration is correct
4. Verify element type helpers match schema exactly
5. Test schema compiles without errors
6. Proceed to **Stage 2 – Resource Migration**

---

## References

- **Reference Implementation**: `genesyscloud/user/resource_genesyscloud_user_schema.go`
- **Requirements Document**: `prompts/pf_complex_resource_migration/Stage1/requirements.md`
- **Plugin Framework Schema Guide**: https://developer.hashicorp.com/terraform/plugin/framework/handling-data/schemas
- **Plan Modifiers Guide**: https://developer.hashicorp.com/terraform/plugin/framework/resources/plan-modification
- **Attribute Types**: https://developer.hashicorp.com/terraform/plugin/framework/handling-data/attributes
- **Nested Attributes**: https://developer.hashicorp.com/terraform/plugin/framework/handling-data/blocks
- **Validators**: https://developer.hashicorp.com/terraform/plugin/framework/validation

