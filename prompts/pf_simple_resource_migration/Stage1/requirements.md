# Stage 1 – Schema Migration Requirements

## Overview

Stage 1 focuses exclusively on migrating schema definitions from Terraform Plugin SDKv2 to the Terraform Plugin Framework. This stage establishes the foundation for the resource migration by defining the data structure and validation rules without implementing any resource lifecycle logic.

**Reference Implementation**: `genesyscloud/routing_wrapupcode/resource_genesyscloud_routing_wrapupcode_schema.go`

---

## Objectives

### Primary Goal
Convert SDKv2 schema definitions to Plugin Framework schema definitions while preserving existing behavior and validation rules.

### Specific Objectives
1. Create a new schema file using Plugin Framework types
2. Define resource schema with all attributes and their properties
3. Define data source schema with required attributes
4. Register the resource and data source with the provider
5. Configure exporter settings for dependency resolution
6. Provide helper functions for cross-package testing

---

## Scope

### In Scope for Stage 1

#### 1. Schema File Creation
- Create `resource_genesyscloud_<resource_name>_schema.go` file
- Example: `resource_genesyscloud_routing_wrapupcode_schema.go`

#### 2. Resource Schema Definition
- Convert all SDKv2 attributes to Plugin Framework attributes
- Define attribute properties:
  - `Required`, `Optional`, `Computed`
  - `Description` text
  - Plan modifiers (e.g., `UseStateForUnknown`)
- Preserve existing validation rules
- Maintain backward compatibility

#### 3. Data Source Schema Definition
- Define data source schema separately from resource schema
- Include only attributes needed for data source lookup
- Typically includes: `id`, `name`, and lookup criteria

#### 4. Registration Function
- Implement `SetRegistrar()` function
- Register Framework resource using `RegisterFrameworkResource()`
- Register Framework data source using `RegisterFrameworkDataSource()`
- Register exporter using `RegisterExporter()`

#### 5. Exporter Configuration
- Define `<ResourceName>Exporter()` function
- Configure `GetResourcesFunc` to use SDK-compatible function
- Define `RefAttrs` for dependency resolution
- Map dependency references to other resource types

#### 6. Helper Functions
- Create `Generate<ResourceName>Resource()` function
- Support cross-package testing
- Generate Terraform HCL configuration strings
- Handle optional attributes with null values

### Out of Scope for Stage 1

❌ **Resource Implementation**
- No CRUD operations (Create, Read, Update, Delete)
- No resource lifecycle logic
- No API interactions

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
- ✅ Attribute types match SDKv2 definitions (String, Int, Bool, etc.)

#### FR2: Schema Accuracy
- ✅ Required/Optional/Computed properties match SDKv2 behavior
- ✅ Descriptions are preserved or improved
- ✅ Default values are maintained (if any)
- ✅ Validation rules are preserved

#### FR3: Plan Modifiers
- ✅ Computed attributes use `UseStateForUnknown()` plan modifier
- ✅ ID attribute uses `UseStateForUnknown()` plan modifier
- ✅ Other computed attributes (e.g., `division_id`) use appropriate modifiers

#### FR4: Registration
- ✅ `SetRegistrar()` function is implemented
- ✅ Resource is registered with correct type name
- ✅ Data source is registered with correct type name
- ✅ Exporter is registered

#### FR5: Exporter Configuration
- ✅ `GetResourcesFunc` points to SDK-compatible function (e.g., `GetAllRoutingWrapupcodesSDK`)
- ✅ `RefAttrs` includes all dependency references
- ✅ Dependency types are correctly mapped (e.g., `division_id` → `genesyscloud_auth_division`)

#### FR6: Helper Functions
- ✅ `Generate<ResourceName>Resource()` function is implemented
- ✅ Function generates valid Terraform HCL configuration
- ✅ Function handles optional attributes correctly
- ✅ Function supports null values using `util.NullValue`

### Non-Functional Requirements

#### NFR1: Code Quality
- ✅ Code follows Go best practices
- ✅ Code follows existing codebase conventions
- ✅ Imports are organized and minimal
- ✅ No unused imports or variables

#### NFR2: Documentation
- ✅ All functions have clear comments
- ✅ Schema descriptions are clear and accurate
- ✅ Complex logic is explained with inline comments

#### NFR3: Backward Compatibility
- ✅ Schema changes do not break existing Terraform configurations
- ✅ Attribute names remain unchanged
- ✅ Attribute behavior remains unchanged

#### NFR4: File Organization
- ✅ Schema file is placed in correct package directory
- ✅ File naming follows convention: `resource_genesyscloud_<resource_name>_schema.go`
- ✅ Package declaration matches directory name

---

## Dependencies and Prerequisites

### Prerequisites

#### 1. Understanding of Resource Structure
- Review existing SDKv2 resource implementation
- Identify all attributes and their properties
- Understand attribute relationships and dependencies

#### 2. Knowledge of Plugin Framework
- Familiarity with Plugin Framework schema types
- Understanding of plan modifiers
- Knowledge of Framework registration patterns

#### 3. Reference Implementation
- Study `routing_wrapupcode` schema implementation
- Understand the migration patterns used
- Review helper function implementations

### Dependencies

#### 1. Package Imports
```go
import (
    datasourceschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
    "github.com/hashicorp/terraform-plugin-framework/resource/schema"
    "github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
    "github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
    "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
    resourceExporter "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_exporter"
    registrar "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_register"
    "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
)
```

#### 2. Existing Functions
- `GetAll<ResourceName>SDK()` function must exist (will be created in Stage 2 if needed)
- Proxy package must exist (not modified during migration)

#### 3. Utility Functions
- `util.NullValue` for handling null values in helper functions
- `util.QuickHashFields()` for export hash calculation (used in Stage 2)

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
- **Impact**: Exporter uses SDK-compatible functions (e.g., `GetAllRoutingWrapupcodesSDK`)

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

## Validation Checklist

Use this checklist to verify Stage 1 completion:

### Schema File
- [ ] File created: `resource_genesyscloud_<resource_name>_schema.go`
- [ ] Package declaration matches directory name
- [ ] All required imports are present
- [ ] No unused imports

### Resource Schema
- [ ] `<ResourceName>ResourceSchema()` function implemented
- [ ] Returns `schema.Schema` type
- [ ] All attributes from SDKv2 are present
- [ ] Attribute properties (Required/Optional/Computed) match SDKv2
- [ ] Descriptions are clear and accurate
- [ ] Plan modifiers are applied to computed attributes
- [ ] ID attribute uses `UseStateForUnknown()` modifier

### Data Source Schema
- [ ] `<ResourceName>DataSourceSchema()` function implemented
- [ ] Returns `datasourceschema.Schema` type
- [ ] Includes `id` and `name` attributes (at minimum)
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

### Helper Functions
- [ ] `Generate<ResourceName>Resource()` function implemented
- [ ] Generates valid Terraform HCL configuration
- [ ] Handles optional attributes correctly
- [ ] Supports null values using `util.NullValue`
- [ ] Includes function comment describing usage

### Code Quality
- [ ] Code compiles without errors
- [ ] Code follows Go conventions
- [ ] Functions have clear comments
- [ ] No TODO or FIXME comments (unless intentional)

---

## Example: routing_wrapupcode Schema Migration

### File Structure
```
genesyscloud/routing_wrapupcode/
└── resource_genesyscloud_routing_wrapupcode_schema.go
```

### Key Components

#### 1. Resource Schema
```go
func RoutingWrapupcodeResourceSchema() schema.Schema {
    return schema.Schema{
        Description: "Genesys Cloud Routing Wrapup Code",
        Attributes: map[string]schema.Attribute{
            "id": schema.StringAttribute{
                Description: "The globally unique identifier for the wrapup code.",
                Computed:    true,
                PlanModifiers: []planmodifier.String{
                    stringplanmodifier.UseStateForUnknown(),
                },
            },
            "name": schema.StringAttribute{
                Description: "Wrapup Code name.",
                Required:    true,
            },
            // ... other attributes
        },
    }
}
```

#### 2. Data Source Schema
```go
func RoutingWrapupcodeDataSourceSchema() datasourceschema.Schema {
    return datasourceschema.Schema{
        Description: "Data source for Genesys Cloud Wrap-up Code. Select a wrap-up code by name",
        Attributes: map[string]datasourceschema.Attribute{
            "id": datasourceschema.StringAttribute{
                Description: "The globally unique identifier for the wrapup code.",
                Computed:    true,
            },
            "name": datasourceschema.StringAttribute{
                Description: "Wrap-up code name.",
                Required:    true,
            },
        },
    }
}
```

#### 3. Registration
```go
func SetRegistrar(regInstance registrar.Registrar) {
    regInstance.RegisterFrameworkResource(ResourceType, NewRoutingWrapupcodeFrameworkResource)
    regInstance.RegisterFrameworkDataSource(ResourceType, NewRoutingWrapupcodeFrameworkDataSource)
    regInstance.RegisterExporter(ResourceType, RoutingWrapupcodeExporter())
}
```

#### 4. Exporter Configuration
```go
func RoutingWrapupcodeExporter() *resourceExporter.ResourceExporter {
    return &resourceExporter.ResourceExporter{
        GetResourcesFunc: provider.GetAllWithPooledClient(GetAllRoutingWrapupcodesSDK),
        RefAttrs: map[string]*resourceExporter.RefAttrSettings{
            "division_id": {RefType: "genesyscloud_auth_division"},
        },
    }
}
```

#### 5. Helper Function
```go
func GenerateRoutingWrapupcodeResource(resourceLabel string, name string, divisionId string, description string) string {
    divisionIdAttr := ""
    if divisionId != util.NullValue {
        divisionIdAttr = fmt.Sprintf(`
        division_id = %s`, divisionId)
    }

    descriptionAttr := ""
    if description != "" {
        descriptionAttr = fmt.Sprintf(`
        description = "%s"`, description)
    }

    return fmt.Sprintf(`resource "genesyscloud_routing_wrapupcode" "%s" {
        name = "%s"%s%s
    }
    `, resourceLabel, name, divisionIdAttr, descriptionAttr)
}
```

---

## Next Steps

After Stage 1 completion and approval:
1. Review schema definitions with team
2. Verify all attributes are correctly defined
3. Confirm exporter configuration is correct
4. Proceed to **Stage 2 – Resource Migration**

---

## References

- **Reference Implementation**: `genesyscloud/routing_wrapupcode/resource_genesyscloud_routing_wrapupcode_schema.go`
- **Plugin Framework Schema Documentation**: https://developer.hashicorp.com/terraform/plugin/framework/handling-data/schemas
- **Plan Modifiers Documentation**: https://developer.hashicorp.com/terraform/plugin/framework/resources/plan-modification
