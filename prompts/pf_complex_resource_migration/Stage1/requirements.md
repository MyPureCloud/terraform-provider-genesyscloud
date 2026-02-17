# Stage 1 – Schema Migration Requirements

## Overview

Stage 1 focuses exclusively on migrating schema definitions from Terraform Plugin SDKv2 to the Terraform Plugin Framework. This stage establishes the foundation for the resource migration by defining the data structure and validation rules without implementing any resource lifecycle logic.

**Reference Implementation**: `genesyscloud/user/resource_genesyscloud_user_schema.go`

**Resource Complexity**: Complex resource with multiple nested blocks, custom validators, plan modifiers, and managed attribute patterns.

---

## Objectives

### Primary Goal
Convert SDKv2 schema definitions to Plugin Framework schema definitions while preserving existing behavior and validation rules for a complex resource with nested structures.

### Specific Objectives
1. Create a new schema file using Plugin Framework types
2. Define resource schema with all attributes, nested blocks, and their properties
3. Migrate complex nested block structures (blocks within blocks)
4. Define data source schema with required attributes
5. Register the resource and data source with the provider
6. Configure exporter settings for dependency resolution
7. Define element type helper functions for complex nested structures
8. Implement validators for all validation rules
9. Configure plan modifiers for computed attributes and defaults

---

## Scope

### In Scope for Stage 1

#### 1. Schema File Creation
- Create `resource_genesyscloud_<resource_name>_schema.go` file
- Example: `resource_genesyscloud_user_schema.go`

#### 2. Resource Schema Definition
- Convert all SDKv2 attributes to Plugin Framework attributes
- Define attribute properties:
  - `Required`, `Optional`, `Computed`, `Sensitive`
  - `Description` text
  - Plan modifiers (e.g., `UseStateForUnknown`, defaults)
- Migrate nested block structures:
  - `SetNestedBlock` for set-based collections
  - `ListNestedBlock` for list-based collections
  - Nested blocks within blocks (e.g., `addresses.phone_numbers`)
- Preserve existing validation rules
- Maintain backward compatibility

#### 3. Validators
- Migrate SDKv2 `ValidateFunc` to Plugin Framework validators
- String validators: `OneOf`, custom validators
- Numeric validators: `Between` for ranges
- List/Set validators: `SizeAtMost`, `SizeAtLeast`
- Custom validators from validators package

#### 4. Plan Modifiers
- `UseStateForUnknown()` for computed attributes
- Default values for optional attributes
- Custom plan modifiers (if needed)

#### 5. Element Type Definitions
- Create helper functions for complex nested object types
- Define reusable element types for consistency
- Example: `routingSkillsElementType()`, `routingLanguagesElementType()`

#### 6. Data Source Schema Definition
- Define data source schema separately from resource schema
- Include only attributes needed for data source lookup
- Typically includes: `id`, `name`, `email`, and lookup criteria

#### 7. Registration Function
- Implement `SetRegistrar()` function
- Register Framework resource using `RegisterFrameworkResource()`
- Register Framework data source using `RegisterFrameworkDataSource()`
- Register exporter using `RegisterExporter()`

#### 8. Exporter Configuration
- Define `<ResourceName>Exporter()` function
- Configure `GetResourcesFunc` to use SDK-compatible function
- Define `RefAttrs` for dependency resolution
- Configure `RemoveIfMissing` for conditional removal
- Configure `AllowEmptyArrays` for empty collections
- Configure `AllowZeroValues` for zero-value fields
- Map dependency references to other resource types

### Out of Scope for Stage 1

❌ **Resource Implementation**
- No CRUD operations (Create, Read, Update, Delete)
- No resource lifecycle logic
- No API interactions

❌ **Helper Functions**
- No flatten functions (e.g., `flattenUserSkills`)
- No build functions (e.g., `buildSdkAddresses`)
- No state management helpers
- Helper functions are covered in Stage 2

❌ **Test Files**
- No test implementation
- No test cases
- Tests are covered in Stage 3

❌ **Export Utilities**
- No export attribute mapping
- No flat attribute conversion
- Export utilities are covered in Stage 4

❌ **Data Source Implementation**
- No data source read logic
- No data source API calls
- Data source implementation is covered in Stage 2

❌ **Proxy Modifications**
- No changes to `genesyscloud_<resource>_proxy.go`
- Proxy files remain unchanged throughout migration

---

## Success Criteria

### Functional Requirements

#### FR1: Schema Completeness
- ✅ All SDKv2 attributes are converted to Plugin Framework attributes
- ✅ No attributes are missing or incorrectly defined
- ✅ Attribute types match SDKv2 definitions (String, Int64, Float64, Bool, etc.)
- ✅ All nested blocks are properly defined
- ✅ Nested blocks within blocks are correctly structured

#### FR2: Schema Accuracy
- ✅ Required/Optional/Computed properties match SDKv2 behavior
- ✅ Descriptions are preserved or improved
- ✅ Default values are maintained (if any)
- ✅ Validation rules are preserved
- ✅ Sensitive attributes are marked correctly

#### FR3: Nested Block Migration
- ✅ SetNestedBlock used for set-based collections
- ✅ ListNestedBlock used for list-based collections
- ✅ Plan modifiers applied to emulate SDKv2 Optional + Computed behavior
- ✅ Inner attributes have correct Required/Optional properties
- ✅ Nested blocks within blocks properly structured (e.g., addresses.phone_numbers)

#### FR4: Validators
- ✅ All SDKv2 ValidateFunc converted to Framework validators
- ✅ String validators (OneOf, custom) implemented
- ✅ Numeric validators (Between) implemented
- ✅ List/Set validators (SizeAtMost) implemented
- ✅ Custom validators from validators package used correctly

#### FR5: Plan Modifiers
- ✅ Computed attributes use `UseStateForUnknown()` plan modifier
- ✅ ID attribute uses `UseStateForUnknown()` plan modifier
- ✅ Attributes with defaults use Optional + Computed + Default pattern
- ✅ Nested blocks use plan modifiers to emulate SDKv2 behavior
- ✅ Custom plan modifiers implemented (if needed)

#### FR6: Element Type Definitions
- ✅ Helper functions created for complex nested object types
- ✅ Element types are reusable and consistent
- ✅ Element types match schema definitions exactly

#### FR7: Registration
- ✅ `SetRegistrar()` function is implemented
- ✅ Resource is registered with correct type name
- ✅ Data source is registered with correct type name
- ✅ Exporter is registered

#### FR8: Exporter Configuration
- ✅ `GetResourcesFunc` points to SDK-compatible function (e.g., `GetAllUsersSDK`)
- ✅ `RefAttrs` includes all dependency references
- ✅ Dependency types are correctly mapped (e.g., `division_id` → `genesyscloud_auth_division`)
- ✅ `RemoveIfMissing` configured for conditional removal
- ✅ `AllowEmptyArrays` configured for empty collections
- ✅ `AllowZeroValues` configured for zero-value fields

### Non-Functional Requirements

#### NFR1: Code Quality
- ✅ Code follows Go best practices
- ✅ Code follows existing codebase conventions
- ✅ Imports are organized and minimal
- ✅ No unused imports or variables
- ✅ Functions have clear comments

#### NFR2: Documentation
- ✅ All attributes have Description fields
- ✅ Schema descriptions are clear and accurate
- ✅ Complex logic is explained with inline comments
- ✅ Managed attribute pattern documented in descriptions

#### NFR3: Backward Compatibility
- ✅ Schema changes do not break existing Terraform configurations
- ✅ Attribute names remain unchanged
- ✅ Attribute behavior remains unchanged
- ✅ Validation rules remain unchanged

#### NFR4: File Organization
- ✅ Schema file is placed in correct package directory
- ✅ File naming follows convention: `resource_genesyscloud_<resource_name>_schema.go`
- ✅ Package declaration matches directory name
- ✅ Imports are grouped logically (validators, schema, types, internal)

---

## Dependencies and Prerequisites

### Prerequisites

#### 1. Understanding of Resource Structure
- Review existing SDKv2 resource implementation
- Identify all attributes and their properties
- Understand nested block structures
- Identify attribute relationships and dependencies
- Understand managed vs unmanaged attribute patterns

#### 2. Knowledge of Plugin Framework
- Familiarity with Plugin Framework schema types
- Understanding of plan modifiers and their purposes
- Knowledge of Framework registration patterns
- Understanding of nested block patterns
- Knowledge of validator types and usage

#### 3. Reference Implementation
- Study `user` schema implementation
- Understand the migration patterns used
- Review element type helper functions
- Understand complex nested block patterns

### Dependencies

#### 1. Package Imports
```go
import (
    // Validators
    "github.com/hashicorp/terraform-plugin-framework-validators/float64validator"
    "github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
    listvalidator "github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
    "github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
    
    // Schema packages
    datasourceschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
    "github.com/hashicorp/terraform-plugin-framework/resource/schema"
    
    // Defaults
    "github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
    "github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
    
    // Plan modifiers
    "github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
    "github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
    "github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
    "github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
    
    // Validators
    "github.com/hashicorp/terraform-plugin-framework/schema/validator"
    
    // Types
    "github.com/hashicorp/terraform-plugin-framework/types"
    
    // Internal packages
    "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
    resourceExporter "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_exporter"
    registrar "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_register"
    "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/validators"
)
```

#### 2. Existing Functions
- `GetAll<ResourceName>SDK()` function must exist (will be created in Stage 2 if needed)
- Proxy package must exist (not modified during migration)

#### 3. Utility Functions
- Custom validators from `validators` package (e.g., `FWValidatePhoneNumber()`, `FWValidateDate()`)
- Custom plan modifiers (if needed)

---

## Constraints

### Technical Constraints

#### TC1: No Proxy Modifications
- **Constraint**: `genesyscloud_<resource>_proxy.go` files MUST NOT be modified
- **Rationale**: Proxy files contain generic shared implementations used by both SDKv2 and Framework
- **Impact**: All API interactions must use existing proxy methods

#### TC2: No Business Logic Changes
- **Constraint**: Schema migration MUST NOT change existing business logic
- **Rationale**: Migration is a framework translation only, not a refactoring opportunity
- **Impact**: Attribute behavior, validation, and defaults must remain identical

#### TC3: Backward Compatibility
- **Constraint**: Schema changes MUST NOT break existing Terraform configurations
- **Rationale**: Users should be able to upgrade provider without modifying their configurations
- **Impact**: Attribute names, types, and behavior must remain unchanged

#### TC4: Export Compatibility
- **Constraint**: Exporter MUST work with both SDKv2 and Framework resources during migration
- **Rationale**: Not all resources will be migrated simultaneously
- **Impact**: Exporter uses SDK-compatible functions (e.g., `GetAllUsersSDK`)

#### TC5: Set Identity Behavior Change
- **Constraint**: Plugin Framework uses ALL fields for Set element identity (no custom hash functions)
- **Rationale**: Framework limitation - cannot exclude fields from Set identity like SDKv2
- **Impact**: Changes to previously excluded fields (e.g., `extension_pool_id`) will cause Set element replacement
- **Mitigation**: Document this behavior change; it's acceptable and matches AWS provider patterns

### Process Constraints

#### PC1: Stage Isolation
- **Constraint**: Stage 1 MUST NOT include resource implementation or tests
- **Rationale**: Clear separation of concerns for easier review and debugging
- **Impact**: Only schema-related code is created in Stage 1

#### PC2: Review and Approval
- **Constraint**: Stage 1 completion requires review and approval before proceeding to Stage 2
- **Rationale**: Schema is the foundation; errors here propagate to later stages
- **Impact**: Thorough review of schema definitions is critical

---

## Schema Conversion Patterns

This section provides detailed technical patterns for converting SDKv2 schemas to Plugin Framework schemas, with concrete examples from the `user` resource migration.

### Type Conversion Table

| SDKv2 Type | Plugin Framework Type | Notes |
|------------|----------------------|-------|
| `schema.TypeString` | `schema.StringAttribute` | Direct conversion |
| `schema.TypeInt` | `schema.Int64Attribute` | **Important**: Int → Int64 |
| `schema.TypeBool` | `schema.BoolAttribute` | Direct conversion |
| `schema.TypeFloat` | `schema.Float64Attribute` | Direct conversion |
| `schema.TypeList` (primitives) | `schema.ListAttribute` | Requires `ElementType` |
| `schema.TypeSet` (primitives) | `schema.SetAttribute` | Requires `ElementType` |
| `schema.TypeList` (nested) | `schema.ListNestedBlock` | Use `NestedObject` |
| `schema.TypeSet` (nested) | `schema.SetNestedBlock` | Use `NestedObject` |

---

## Validation Checklist

Use this checklist to verify Stage 1 completion:

### Schema File
- [ ] File created: `resource_genesyscloud_<resource_name>_schema.go`
- [ ] Package declaration matches directory name
- [ ] All required imports are present
- [ ] No unused imports
- [ ] Imports are organized logically

### Resource Schema
- [ ] `<ResourceName>ResourceSchema()` function implemented
- [ ] Returns `schema.Schema` type
- [ ] All attributes from SDKv2 are present
- [ ] Attribute properties (Required/Optional/Computed/Sensitive) match SDKv2
- [ ] Descriptions are clear and accurate
- [ ] Plan modifiers are applied to computed attributes
- [ ] ID attribute uses `UseStateForUnknown()` modifier
- [ ] Attributes with defaults use Optional + Computed + Default pattern

### Nested Blocks
- [ ] SetNestedBlock used for set-based collections
- [ ] ListNestedBlock used for list-based collections
- [ ] Plan modifiers applied to blocks (UseStateForUnknown)
- [ ] Inner attributes have correct Required/Optional properties
- [ ] Nested blocks within blocks properly structured
- [ ] Validators applied to nested block attributes

### Validators
- [ ] All SDKv2 ValidateFunc converted to Framework validators
- [ ] String validators (OneOf) implemented
- [ ] Numeric validators (Between) implemented
- [ ] List/Set validators (SizeAtMost) implemented
- [ ] Custom validators from validators package used

### Element Type Definitions
- [ ] Helper functions created for complex nested object types
- [ ] Element types match schema definitions
- [ ] Element types are reusable

### Data Source Schema
- [ ] `<ResourceName>DataSourceSchema()` function implemented
- [ ] Returns `datasourceschema.Schema` type
- [ ] Includes required lookup attributes
- [ ] Descriptions are clear and accurate

### Registration
- [ ] `SetRegistrar()` function implemented
- [ ] Calls `RegisterFrameworkResource()` with correct type name
- [ ] Calls `RegisterFrameworkDataSource()` with correct type name
- [ ] Calls `RegisterExporter()` with exporter configuration

### Exporter Configuration
- [ ] `<ResourceName>Exporter()` function implemented
- [ ] Returns `*resourceExporter.ResourceExporter`
- [ ] `GetResourcesFunc` uses SDK-compatible function
- [ ] `RefAttrs` includes all dependency references
- [ ] Dependency types are correctly mapped
- [ ] `RemoveIfMissing` configured (if needed)
- [ ] `AllowEmptyArrays` configured (if needed)
- [ ] `AllowZeroValues` configured (if needed)

### Code Quality
- [ ] Code compiles without errors
- [ ] Code follows Go conventions
- [ ] Functions have clear comments
- [ ] No TODO or FIXME comments (unless intentional)
- [ ] Complex patterns are documented

### Attribute Property Patterns

#### Pattern 1: Required String Attribute
```go
// SDKv2 (before)
"name": {
    Type:     schema.TypeString,
    Required: true,
}

// Plugin Framework (after)
"name": schema.StringAttribute{
    Description: "User's full name.",
    Required:    true,
}
```

#### Pattern 2: Optional String Attribute
```go
// SDKv2 (before)
"department": {
    Type:     schema.TypeString,
    Optional: true,
}

// Plugin Framework (after)
"department": schema.StringAttribute{
    Description: "User's department.",
    Optional:    true,
}
```

#### Pattern 3: Computed ID Attribute
```go
// SDKv2 (before)
"id": {
    Type:     schema.TypeString,
    Computed: true,
}

// Plugin Framework (after)
"id": schema.StringAttribute{
    Description: "The ID of the user.",
    Computed:    true,
    PlanModifiers: []planmodifier.String{
        stringplanmodifier.UseStateForUnknown(),
    },
}
```

**Critical**: All computed attributes MUST use `UseStateForUnknown()` plan modifier to preserve state values.

#### Pattern 4: Optional + Computed + Default
```go
// SDKv2 (before)
"state": {
    Type:     schema.TypeString,
    Optional: true,
    Default:  "active",
}

// Plugin Framework (after)
"state": schema.StringAttribute{
    Description: "User's state (active | inactive). Default is 'active'.",
    Optional:    true,
    Computed:    true,  // REQUIRED when using Default
    Default:     stringdefault.StaticString("active"),
    Validators: []validator.String{
        stringvalidator.OneOf("active", "inactive"),
    },
}
```

**Critical Rule**: In Plugin Framework, attributes with `Default` MUST be both `Optional` and `Computed`.

#### Pattern 5: Sensitive Attribute
```go
// Plugin Framework
"password": schema.StringAttribute{
    Description: "User's password. If specified, this is only set on user create.",
    Optional:    true,
    Sensitive:   true,
}
```

#### Pattern 6: Optional + Computed (No Default)
```go
// SDKv2 (before)
"division_id": {
    Type:     schema.TypeString,
    Optional: true,
    Computed: true,
}

// Plugin Framework (after)
"division_id": schema.StringAttribute{
    Description: "The division to which this user will belong. If not set, the home division will be used.",
    Optional:    true,
    Computed:    true,
}
```

**Note**: No plan modifier needed here - the API will compute the value if not provided.

#### Pattern 7: Set of Strings (Managed Attribute)
```go
// SDKv2 (before)
"profile_skills": {
    Type:     schema.TypeSet,
    Optional: true,
    Computed: true,
    Elem:     &schema.Schema{Type: schema.TypeString},
}

// Plugin Framework (after)
"profile_skills": schema.SetAttribute{
    Description: "Profile skills for this user. If not set, this resource will not manage profile skills.",
    Optional:    true,
    Computed:    true,
    ElementType: types.StringType,
}
```

**Managed Attribute Pattern**: The description explicitly states "If not set, this resource will not manage..." to indicate optional management.

### Nested Block Patterns

#### Pattern 8: SetNestedBlock (Managed Attribute)
```go
// SDKv2 (before)
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
}

// Plugin Framework (after)
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

**Critical Differences**:
1. SetNestedBlock/ListNestedBlock do NOT support `Optional` or `Required` at the block level
2. Use `UseStateForUnknown()` plan modifier to emulate SDKv2's `Optional + Computed` behavior
3. Inner attributes CAN be `Required` or `Optional`
4. Validators migrate from `ValidateFunc` to Framework validators

#### Pattern 9: ListNestedBlock (Single Item)
```go
// SDKv2 (before)
"employer_info": {
    Type:     schema.TypeList,
    Optional: true,
    Computed: true,
    MaxItems: 1,
    Elem: &schema.Resource{
        Schema: map[string]*schema.Schema{
            "official_name": {
                Type:     schema.TypeString,
                Optional: true,
            },
            "employee_id": {
                Type:     schema.TypeString,
                Optional: true,
            },
            "employee_type": {
                Type:         schema.TypeString,
                Optional:     true,
                ValidateFunc: validation.StringInSlice([]string{"Full-time", "Part-time", "Contractor"}, false),
            },
            "date_hire": {
                Type:         schema.TypeString,
                Optional:     true,
                ValidateFunc: validation.IsRFC3339Time,
            },
        },
    },
}

// Plugin Framework (after)
"employer_info": schema.ListNestedBlock{
    Description: "The employer info for this user. If not set, this resource will not manage employer info.",
    PlanModifiers: []planmodifier.List{
        listplanmodifier.UseStateForUnknown(),
    },
    Validators: []validator.List{
        listvalidator.SizeAtMost(1), // Replaces MaxItems
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
- `MaxItems` → `listvalidator.SizeAtMost(1)`
- `ValidateFunc` → Framework validators
- Custom validators use validators package

#### Pattern 10: Nested Blocks Within Blocks
```go
// Plugin Framework - addresses block contains phone_numbers and other_emails blocks
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

**Critical**: When nesting blocks within blocks, use `Blocks: map[string]schema.Block{}` in the parent's `NestedBlockObject`, not `Attributes`.

### Validator Patterns

#### Pattern 11: String OneOf Validator
```go
// SDKv2 (before)
ValidateFunc: validation.StringInSlice([]string{"active", "inactive"}, false)

// Plugin Framework (after)
Validators: []validator.String{
    stringvalidator.OneOf("active", "inactive"),
}
```

#### Pattern 12: Numeric Range Validators
```go
// SDKv2 (before) - Float
ValidateFunc: validation.FloatBetween(0, 5)

// Plugin Framework (after)
Validators: []validator.Float64{
    float64validator.Between(0, 5),
}

// SDKv2 (before) - Int
ValidateFunc: validation.IntBetween(0, 25)

// Plugin Framework (after)
Validators: []validator.Int64{
    int64validator.Between(0, 25),
}
```

#### Pattern 13: Custom Validators
```go
// Plugin Framework - Phone number validation
"number": schema.StringAttribute{
    Validators: []validator.String{
        validators.FWValidatePhoneNumber(),
    },
}

// Plugin Framework - Date validation
"date_hire": schema.StringAttribute{
    Validators: []validator.String{
        validators.FWValidateDate(),
    },
}
```

**Note**: Custom validators must be implemented in the `validators` package and follow Framework validator interface.

#### Pattern 14: List/Set Size Validators
```go
// SDKv2 (before)
MaxItems: 1

// Plugin Framework (after)
Validators: []validator.List{
    listvalidator.SizeAtMost(1),
}

// Other size validators
listvalidator.SizeAtLeast(1)
listvalidator.SizeBetween(1, 10)
```

### Plan Modifier Patterns

#### Pattern 15: UseStateForUnknown (Most Common)
```go
// For computed ID
"id": schema.StringAttribute{
    Computed: true,
    PlanModifiers: []planmodifier.String{
        stringplanmodifier.UseStateForUnknown(),
    },
}

// For nested blocks (emulates Optional + Computed)
"routing_skills": schema.SetNestedBlock{
    PlanModifiers: []planmodifier.Set{
        setplanmodifier.UseStateForUnknown(),
    },
}

// For lists
"addresses": schema.ListNestedBlock{
    PlanModifiers: []planmodifier.List{
        listplanmodifier.UseStateForUnknown(),
    },
}
```

**Purpose**: Preserves state value when attribute is not specified in config. This emulates SDKv2's `Computed: true` behavior.

#### Pattern 16: Default Values
```go
// String default
"state": schema.StringAttribute{
    Optional: true,
    Computed: true,  // REQUIRED with Default
    Default:  stringdefault.StaticString("active"),
}

// Boolean default
"acd_auto_answer": schema.BoolAttribute{
    Optional: true,
    Computed: true,  // REQUIRED with Default
    Default:  booldefault.StaticBool(false),
}

// Int64 default (if needed)
"timeout": schema.Int64Attribute{
    Optional: true,
    Computed: true,
    Default:  int64default.StaticInt64(30),
}
```

**Critical Rule**: Attributes with `Default` MUST be both `Optional` and `Computed`.

#### Pattern 17: Custom Plan Modifiers
```go
// Example: NullIfEmpty for extension_pool_id
"extension_pool_id": schema.StringAttribute{
    Optional: true,
    PlanModifiers: []planmodifier.String{
        phoneplan.NullIfEmpty{}, // Custom modifier
    },
}
```

**Note**: Custom plan modifiers must implement the appropriate plan modifier interface.

### Element Type Definition Patterns

#### Pattern 18: Element Type Helper Functions
```go
// Define once, reuse everywhere
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
```

**Benefits**:
1. Type safety - ensures consistency between schema and utils
2. Reusability - define once, use in multiple places
3. Maintainability - change type definition in one place

### Data Source Schema Pattern

#### Pattern 19: Data Source Schema
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

**Key Points**:
- Use `datasourceschema` package, not `schema`
- Typically includes: `id` (Computed), lookup fields (Optional)
- Much simpler than resource schema
- No nested blocks usually needed

### Exporter Configuration Pattern

#### Pattern 20: Exporter Configuration
```go
func UserExporter() *resourceExporter.ResourceExporter {
    return &resourceExporter.ResourceExporter{
        GetResourcesFunc: provider.GetAllWithPooledClient(GetAllUsersSDK),
        RefAttrs: map[string]*resourceExporter.RefAttrSettings{
            // Self-reference
            "manager": {RefType: ResourceType},
            
            // External references
            "division_id":                               {RefType: "genesyscloud_auth_division"},
            "routing_skills.skill_id":                   {RefType: "genesyscloud_routing_skill"},
            "routing_languages.language_id":             {RefType: "genesyscloud_routing_language"},
            "locations.location_id":                     {RefType: "genesyscloud_location"},
            
            // Nested references (use dot notation)
            "addresses.phone_numbers.extension_pool_id": {RefType: "genesyscloud_telephony_providers_edges_extension_pool"},
        },
        RemoveIfMissing: map[string][]string{
            // Remove block if required field is missing
            "routing_skills":         {"skill_id"},
            "routing_languages":      {"language_id"},
            "locations":              {"location_id"},
            "voicemail_userpolicies": {"alert_timeout_seconds"},
        },
        AllowEmptyArrays: []string{
            // These arrays can be explicitly empty (not null)
            "routing_skills",
            "routing_languages",
        },
        AllowZeroValues: []string{
            // These numeric fields can legitimately be 0
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

**RefAttrs Rules**:
- Use dot notation for nested attributes: `"routing_skills.skill_id"`
- Self-references use `ResourceType` constant
- External references use full resource type: `"genesyscloud_routing_skill"`

**RemoveIfMissing Rules**:
- Key = block name
- Value = array of required fields
- If any required field is missing, remove the entire block

**AllowZeroValues Rules**:
- Critical for proficiency ratings, capacity values
- Prevents 0 from being treated as "not set"

### Common Pitfalls and Solutions

#### Pitfall 1: Forgetting Computed with Default
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

#### Pitfall 2: Using Optional on Nested Blocks
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

#### Pitfall 3: Missing UseStateForUnknown on Computed Attributes
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

#### Pitfall 4: Wrong Type for Nested Blocks
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

#### Pitfall 5: Set Identity Behavior Change
```go
// SDKv2 - Could exclude fields from Set hash
Set: &schema.Set{
    F: func(v interface{}) int {
        // Custom hash excluding extension_pool_id
    },
}

// Plugin Framework - ALL fields included in Set identity
// No custom hash function support
// Changes to extension_pool_id will cause Set element replacement
```

**Mitigation**: Document this behavior change. It's acceptable and matches AWS provider patterns during PF migration.

### Advanced Patterns

#### Pattern 21: Three-Level Nested Blocks
```go
// routing_utilization → call/callback/email/chat/message → attributes
"routing_utilization": schema.ListNestedBlock{
    Description: "The routing utilization settings for this user.",
    PlanModifiers: []planmodifier.List{
        listplanmodifier.UseStateForUnknown(),
    },
    Validators: []validator.List{
        listvalidator.SizeAtMost(1),
    },
    NestedObject: schema.NestedBlockObject{
        Blocks: map[string]schema.Block{  // Level 2: Media type blocks
            "call": schema.ListNestedBlock{
                Description: "Call media settings. If not set, this reverts to the default media type settings.",
                PlanModifiers: []planmodifier.List{
                    listplanmodifier.UseStateForUnknown(),
                },
                Validators: []validator.List{
                    listvalidator.SizeAtMost(1),
                },
                NestedObject: schema.NestedBlockObject{
                    Attributes: map[string]schema.Attribute{  // Level 3: Actual attributes
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
            "label_utilizations": schema.ListNestedBlock{
                Description: "Label utilization settings. If not set, default label settings will be applied.",
                PlanModifiers: []planmodifier.List{
                    listplanmodifier.UseStateForUnknown(),
                },
                NestedObject: schema.NestedBlockObject{
                    Attributes: map[string]schema.Attribute{
                        "label_id": schema.StringAttribute{
                            Description: "Id of the label being configured.",
                            Required:    true,
                        },
                        "maximum_capacity": schema.Int64Attribute{
                            Description: "Maximum capacity of conversations with this label. Value must be between 0 and 25.",
                            Required:    true,
                            Validators: []validator.Int64{
                                int64validator.Between(0, 25),
                            },
                        },
                        "interrupting_label_ids": schema.SetAttribute{
                            Description: "Set of other labels that can interrupt this label.",
                            Optional:    true,
                            ElementType: types.StringType,
                        },
                    },
                },
            },
        },
    },
}
```

**Key Points**:
- Level 1: `routing_utilization` (ListNestedBlock)
- Level 2: `call`, `callback`, etc. (ListNestedBlock within Blocks)
- Level 3: `maximum_capacity`, `include_non_acd`, etc. (Attributes)
- Each level can have its own plan modifiers and validators
- Use `Blocks` map for nested blocks, `Attributes` map for final attributes

#### Pattern 22: Optional + Computed Without Default
```go
// Special case: Computed by API, but user can override
"send_email_notifications": schema.BoolAttribute{
    Description: "Whether email notifications are sent to the user when a new voicemail is received.",
    Optional:    true,
    Computed:    true,
    // NO Default - API will compute if not provided
}

"division_id": schema.StringAttribute{
    Description: "The division to which this user will belong. If not set, the home division will be used.",
    Optional:    true,
    Computed:    true,
    // NO Default - API will compute if not provided
}
```

**When to use**:
- Attribute can be set by user OR computed by API
- No specific default value to set
- API will determine the value if not provided
- Different from Pattern 4 (which has explicit Default)
- Different from Pattern 6 (which is just Optional + Computed for simple cases)

**Critical**: Do NOT add `UseStateForUnknown()` plan modifier here - we want the API to recompute if config changes.

#### Pattern 23: Dynamic Description Generation
```go
// Using helper function to generate dynamic descriptions
"interruptible_media_types": schema.SetAttribute{
    Description: fmt.Sprintf("Set of other media types that can interrupt this media type (%s).", 
        strings.Join(getSdkUtilizationTypes(), " | ")),
    Optional:    true,
    ElementType: types.StringType,
}

// Helper function (defined in same file)
func getSdkUtilizationTypes() []string {
    types := make([]string, 0, len(utilizationMediaTypes))
    for t := range utilizationMediaTypes {
        types = append(types, t)
    }
    sort.Strings(types)
    return types
}

// Variable (defined at package level)
var (
    utilizationMediaTypes = map[string]string{
        "call":     "call",
        "callback": "callback",
        "chat":     "chat",
        "email":    "email",
        "message":  "message",
    }
)
```

**Benefits**:
- Descriptions stay in sync with code
- Reduces duplication
- Easier to maintain

**Usage**: Import `fmt` and `strings` packages for string manipulation.

#### Pattern 24: Custom Plan Modifiers
```go
// Using custom plan modifier from internal package
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

**When to create custom plan modifiers**:
- Standard plan modifiers don't meet your needs
- Complex state manipulation required
- Consistent behavior needed across multiple attributes

**Location**: Create in `genesyscloud/util/<package>` for reusability.

#### Pattern 25: Package-Level Variables and Constants
```go
// Constants
const ResourceType = "genesyscloud_user"

// Package-level variables
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

**Purpose**:
- `ResourceType`: Used in registration, exporter, and error messages
- Package variables: Shared across schema and utils functions
- Reduces magic strings and improves maintainability

### Migration Considerations and TODO Comments

#### Consideration 1: Set Identity Behavior Change
```go
// TODO comment from actual schema
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

**Resolution**: 
- Documented as TC5 in Constraints section
- Acceptable behavior change
- Matches AWS provider patterns during PF migration
- Custom plan modifier helps minimize impact

#### Consideration 2: Future Enhancements
```go
// TODO comment from actual schema
//TODO
/* will handel this condition later
// NOTE: phone_extension_pools is defined at the top level (outside addresses.phone_numbers)
// instead of being nested under each phone number. In SDKv2, extension_pool_id was excluded
// from the custom hash used for the phone_numbers Set to prevent plan diffs when only the
// pool mapping changed. Since Plugin Framework does not support custom Set hash functions,
// this top-level Computed map preserves the same behavior: extension pool IDs are
// represented as computed metadata (identity-insensitive) rather than user-managed fields.
// This avoids unwanted diffs when pool assignments change while keeping phone_numbers
// identity stable across plans and refreshes.
"phone_extension_pools": schema.MapAttribute{
    ElementType: types.StringType,
    Computed:    true,
    Description: "Id of the extension pool which contains this extension." +
        "Computed mapping of phone identity keys to  (MEDIA|TYPE|E164|EXT) to extension_pool_id." +
        "Used internally to prevent diffs when pool assignments change.",
},
*/
```

**Status**: 
- Commented out in current implementation
- Alternative approach considered for future
- Current implementation uses `NullIfEmpty` plan modifier instead
- Keep TODO for potential future optimization

#### Consideration 3: Plan Modifiers vs Validators
```go
// TODO comment from actual schema
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
- Use validators for validation
- Plan modifiers for normalization (commented out for now)
- Can be enabled later if E.164 canonicalization is needed
- Validators are sufficient for current requirements

### Helper Functions for Schema

#### Helper Function Pattern
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

**Purpose**:
- Generate dynamic content for schema descriptions
- Ensure consistency between code and documentation
- Reduce duplication

**Location**: Define in schema file, use in schema definitions.

### Import Organization

#### Recommended Import Order
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

### Complete Schema File Structure

```go
package user

import (
    // ... imports as shown above
)

// Constants
const ResourceType = "genesyscloud_user"

// Package-level variables
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

// SetRegistrar registers all the resources and exporters in the package
func SetRegistrar(l registrar.Registrar) {
    l.RegisterFrameworkDataSource(ResourceType, NewUserFrameworkDataSource)
    l.RegisterFrameworkResource(ResourceType, NewUserFrameworkResource)
    l.RegisterExporter(ResourceType, UserExporter())
}

// Helper functions
func getSdkUtilizationTypes() []string {
    // ... implementation
}

// UserDataSourceSchema returns the schema for the user data source
func UserDataSourceSchema() datasourceschema.Schema {
    // ... implementation
}

// UserResourceSchema returns the schema for the user resource
func UserResourceSchema() schema.Schema {
    // ... implementation
}

// UserExporter returns the exporter configuration
func UserExporter() *resourceExporter.ResourceExporter {
    // ... implementation
}
```

**File Organization**:
1. Package declaration
2. Imports (organized by group)
3. Constants
4. Package-level variables
5. SetRegistrar function
6. Helper functions
7. Data source schema function
8. Resource schema function
9. Exporter configuration function

**Mitigation**: Document this behavior change. It's acceptable and matches AWS provider patterns during PF migration.

---

## Example: user Schema Migration

### File Structure
```
genesyscloud/user/
└── resource_genesyscloud_user_schema.go
```

### Key Components

#### 1. Constants and Registration
```go
package user

const ResourceType = "genesyscloud_user"

func SetRegistrar(l registrar.Registrar) {
    l.RegisterFrameworkDataSource(ResourceType, NewUserFrameworkDataSource)
    l.RegisterFrameworkResource(ResourceType, NewUserFrameworkResource)
    l.RegisterExporter(ResourceType, UserExporter())
}
```

#### 2. Element Type Helper Functions
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
```

#### 3. Data Source Schema
```go
func UserDataSourceSchema() datasourceschema.Schema {
    return datasourceschema.Schema{
        Description: "Data source for Genesys Cloud Users. Select a user by email or name.",
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

#### 4. Resource Schema (Simplified Example)
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
            "addresses": schema.ListNestedBlock{
                Description: "The address settings for this user. If not set, this resource will not manage addresses.",
                PlanModifiers: []planmodifier.List{
                    listplanmodifier.UseStateForUnknown(),
                },
                Validators: []validator.List{
                    listvalidator.SizeAtMost(1),
                },
                NestedObject: schema.NestedBlockObject{
                    Blocks: map[string]schema.Block{
                        "phone_numbers": schema.SetNestedBlock{
                            Description: "Phone number addresses for this user.",
                            NestedObject: schema.NestedBlockObject{
                                Attributes: map[string]schema.Attribute{
                                    "number": schema.StringAttribute{
                                        Description: "Phone number in E.164 format.",
                                        Optional:    true,
                                        Validators:  []validator.String{validators.FWValidatePhoneNumber()},
                                    },
                                    "extension": schema.StringAttribute{
                                        Description: "Phone number extension",
                                        Optional:    true,
                                    },
                                },
                            },
                        },
                    },
                },
            },
        },
    }
}
```

#### 5. Exporter Configuration
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
        AllowEmptyArrays: []string{"routing_skills", "routing_languages"},
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

---

## Next Steps

After Stage 1 completion and approval:
1. Review schema definitions with team
2. Verify all attributes are correctly defined
3. Confirm nested block structures are correct
4. Verify exporter configuration is complete
5. Ensure all validators and plan modifiers are properly configured
6. Proceed to **Stage 2 – Resource Migration**

---

## References

- **Reference Implementation**: `genesyscloud/user/resource_genesyscloud_user_schema.go`
- **Plugin Framework Schema Documentation**: https://developer.hashicorp.com/terraform/plugin/framework/handling-data/schemas
- **Plan Modifiers Documentation**: https://developer.hashicorp.com/terraform/plugin/framework/resources/plan-modification
- **Validators Documentation**: https://developer.hashicorp.com/terraform/plugin/framework/validation
- **Nested Attributes and Blocks**: https://developer.hashicorp.com/terraform/plugin/framework/handling-data/attributes/list-nested
