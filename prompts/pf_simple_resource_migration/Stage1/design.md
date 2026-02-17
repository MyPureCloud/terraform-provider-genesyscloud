# Stage 1 – Schema Migration Design

## Overview

This document describes the design decisions, architecture, and patterns used in Stage 1 of the Plugin Framework migration. Stage 1 focuses on converting SDKv2 schema definitions to Plugin Framework schema definitions while maintaining backward compatibility and preparing for future stages.

**Reference Implementation**: `genesyscloud/routing_wrapupcode/resource_genesyscloud_routing_wrapupcode_schema.go`

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

### 4. Export Compatibility
**Principle**: Exporter must work with both SDKv2 and Framework resources during migration.

**Rationale**:
- Not all resources will be migrated simultaneously
- Exporter must handle mixed resource types
- Dependency resolution must work across SDKv2 and Framework resources

**Implementation**:
- Use SDK-compatible `GetAll<ResourceName>SDK()` function
- Define `RefAttrs` for dependency resolution
- Maintain flat attribute map format (Phase 1 temporary)

---

## Architecture

### File Structure

```
genesyscloud/<resource_name>/
├── resource_genesyscloud_<resource_name>_schema.go          ← Stage 1 (THIS FILE)
├── resource_genesyscloud_<resource_name>.go                 ← Stage 2
├── resource_genesyscloud_<resource_name>_test.go            ← Stage 3
├── resource_genesyscloud_<resource_name>_export_utils.go    ← Stage 4
├── data_source_genesyscloud_<resource_name>.go              ← Stage 2
├── data_source_genesyscloud_<resource_name>_test.go         ← Stage 3
├── genesyscloud_<resource_name>_init_test.go                ← Stage 3
└── genesyscloud_<resource_name>_proxy.go                    ← NOT MODIFIED
```

### Schema File Components

The schema file contains five main components:

```
┌─────────────────────────────────────────────────────────┐
│  resource_genesyscloud_<resource_name>_schema.go        │
├─────────────────────────────────────────────────────────┤
│  1. Package Constants                                   │
│     - ResourceType constant                             │
├─────────────────────────────────────────────────────────┤
│  2. SetRegistrar() Function                             │
│     - Register Framework resource                       │
│     - Register Framework data source                    │
│     - Register exporter                                 │
├─────────────────────────────────────────────────────────┤
│  3. Resource Schema Function                            │
│     - <ResourceName>ResourceSchema()                    │
│     - Returns schema.Schema                             │
├─────────────────────────────────────────────────────────┤
│  4. Data Source Schema Function                         │
│     - <ResourceName>DataSourceSchema()                  │
│     - Returns datasourceschema.Schema                   │
├─────────────────────────────────────────────────────────┤
│  5. Exporter Configuration Function                     │
│     - <ResourceName>Exporter()                          │
│     - Returns *resourceExporter.ResourceExporter        │
├─────────────────────────────────────────────────────────┤
│  6. Helper Functions                                    │
│     - Generate<ResourceName>Resource()                  │
│     - For cross-package testing                         │
└─────────────────────────────────────────────────────────┘
```

---

## Component Design

### 1. Package Constants

**Purpose**: Define the resource type identifier used throughout the package.

**Design**:
```go
const ResourceType = "genesyscloud_<resource_name>"
```

**Example** (routing_wrapupcode):
```go
const ResourceType = "genesyscloud_routing_wrapupcode"
```

**Rationale**:
- Single source of truth for resource type name
- Used in registration, tests, and exporter
- Prevents typos and inconsistencies
- Easy to reference across package

---

### 2. SetRegistrar() Function

**Purpose**: Register the Framework resource, data source, and exporter with the provider.

**Design Pattern**:
```go
func SetRegistrar(regInstance registrar.Registrar) {
    regInstance.RegisterFrameworkResource(ResourceType, New<ResourceName>FrameworkResource)
    regInstance.RegisterFrameworkDataSource(ResourceType, New<ResourceName>FrameworkDataSource)
    regInstance.RegisterExporter(ResourceType, <ResourceName>Exporter())
}
```

**Example** (routing_wrapupcode):
```go
func SetRegistrar(regInstance registrar.Registrar) {
    regInstance.RegisterFrameworkResource(ResourceType, NewRoutingWrapupcodeFrameworkResource)
    regInstance.RegisterFrameworkDataSource(ResourceType, NewRoutingWrapupcodeFrameworkDataSource)
    regInstance.RegisterExporter(ResourceType, RoutingWrapupcodeExporter())
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

---

### 3. Resource Schema Function

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
    }
}
```

**Example** (routing_wrapupcode):
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
            "division_id": schema.StringAttribute{
                Description: "The division to which this routing wrapupcode will belong.",
                Optional:    true,
                Computed:    true,
                PlanModifiers: []planmodifier.String{
                    stringplanmodifier.UseStateForUnknown(),
                },
            },
            "description": schema.StringAttribute{
                Description: "The wrap-up code description.",
                Optional:    true,
            },
        },
    }
}
```

#### Attribute Properties

**Required Attributes**:
- Must be provided by user in Terraform configuration
- Cannot be null or omitted
- Example: `name` (every resource needs a name)

**Optional Attributes**:
- User can provide or omit
- Can be null
- Example: `description` (not always needed)

**Computed Attributes**:
- Calculated by provider or API
- User cannot set directly
- Example: `id` (generated by API)

**Optional + Computed Attributes**:
- User can provide OR provider will compute
- If user provides, use that value
- If user omits, provider computes default
- Example: `division_id` (defaults to home division if not specified)

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

**Why Important**:
- Prevents unnecessary resource replacement
- Avoids "known after apply" in plan when value won't actually change
- Improves user experience by showing accurate plan

**Comparison with SDKv2**:
- SDKv2: Used `DiffSuppressFunc` or `CustomizeDiff`
- Framework: Uses plan modifiers (cleaner, more explicit)

---

### 4. Data Source Schema Function

**Purpose**: Define the schema for the Terraform data source (lookup by name).

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
                Required:    true,
            },
        },
    }
}
```

**Example** (routing_wrapupcode):
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

**Design Decisions**:

| Decision | Rationale |
|----------|-----------|
| Separate schema function | Data source schema is different from resource schema |
| Use `datasourceschema` package | Framework requires different types for data sources |
| Minimal attributes | Data sources only need lookup criteria and result ID |
| No plan modifiers | Data sources don't have plans (read-only) |

**Key Differences from Resource Schema**:
- Uses `datasourceschema.Schema` instead of `schema.Schema`
- Uses `datasourceschema.StringAttribute` instead of `schema.StringAttribute`
- No plan modifiers (data sources are read-only)
- Typically only includes `id` and `name` (lookup criteria)

**Common Pattern**:
- `name`: Required (user provides for lookup)
- `id`: Computed (returned after lookup)
- Additional attributes can be added if needed for lookup

---

### 5. Exporter Configuration Function

**Purpose**: Configure the exporter for this resource, including dependency resolution.

**Design Pattern**:
```go
func <ResourceName>Exporter() *resourceExporter.ResourceExporter {
    return &resourceExporter.ResourceExporter{
        GetResourcesFunc: provider.GetAllWithPooledClient(GetAll<ResourceName>SDK),
        RefAttrs: map[string]*resourceExporter.RefAttrSettings{
            "<dependency_attr>": {RefType: "<dependency_resource_type>"},
        },
    }
}
```

**Example** (routing_wrapupcode):
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

#### GetResourcesFunc Design

**Purpose**: Provide a function to fetch all resources for export.

**Why SDK Version?**:
- Exporter currently uses SDKv2 diagnostics and flat attribute maps
- Framework version (`GetAllRoutingWrapupcodes`) exists but not used yet
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

**Example** (routing_wrapupcode):
```go
RefAttrs: map[string]*resourceExporter.RefAttrSettings{
    "division_id": {RefType: "genesyscloud_auth_division"},
}
```

**How It Works**:
1. Exporter reads `division_id` attribute from resource
2. Looks up the division resource by ID
3. Generates HCL reference: `division_id = genesyscloud_auth_division.division_label.id`
4. Ensures division is exported before wrapupcode (dependency ordering)

**Key Points**:
- Attribute name must match schema attribute name exactly
- RefType must match the Terraform resource type
- Multiple dependencies can be defined
- Exporter uses this for both ordering and HCL generation

**Common Dependency Attributes**:
- `division_id` → `genesyscloud_auth_division`
- `queue_id` → `genesyscloud_routing_queue`
- `flow_id` → `genesyscloud_flow`
- `skill_id` → `genesyscloud_routing_skill`

---

### 6. Helper Functions

**Purpose**: Provide utility functions for generating Terraform HCL configurations in tests.

**Design Pattern**:
```go
func Generate<ResourceName>Resource(
    resourceLabel string,
    name string,
    // ... other required attributes
    optionalAttr1 string,
    optionalAttr2 string,
) string {
    // Build optional attribute strings
    optionalAttr1Str := ""
    if optionalAttr1 != util.NullValue {
        optionalAttr1Str = fmt.Sprintf(`
        optional_attr1 = %s`, optionalAttr1)
    }

    optionalAttr2Str := ""
    if optionalAttr2 != "" {
        optionalAttr2Str = fmt.Sprintf(`
        optional_attr2 = "%s"`, optionalAttr2)
    }

    return fmt.Sprintf(`resource "genesyscloud_<resource_name>" "%s" {
        name = "%s"%s%s
    }
    `, resourceLabel, name, optionalAttr1Str, optionalAttr2Str)
}
```

**Example** (routing_wrapupcode):
```go
func GenerateRoutingWrapupcodeResource(
    resourceLabel string,
    name string,
    divisionId string,
    description string,
) string {
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

**Design Decisions**:

| Decision | Rationale |
|----------|-----------|
| String return type | Returns HCL configuration as string for test concatenation |
| Handle optional attributes | Allows tests to omit optional attributes |
| Use `util.NullValue` | Distinguishes between "omit attribute" and "empty string" |
| No quotes for references | Allows passing Terraform references (e.g., `genesyscloud_auth_division.div.id`) |
| Add quotes for literals | String literals need quotes in HCL |

**Usage in Tests**:
```go
// Test with all attributes
config := GenerateRoutingWrapupcodeResource(
    "test_wrapupcode",
    "My Wrapupcode",
    "genesyscloud_auth_division.div.id",
    "Test description",
)

// Test with minimal attributes (omit optional)
config := GenerateRoutingWrapupcodeResource(
    "test_wrapupcode",
    "My Wrapupcode",
    util.NullValue,  // Omit division_id
    "",              // Omit description
)
```

**Cross-Package Usage**:
- Other packages can import and use this function
- Enables testing resources that depend on this resource
- Example: Testing routing queue that references wrapupcode

---

## SDKv2 vs Plugin Framework Comparison

### Schema Definition

**SDKv2**:
```go
func ResourceRoutingWrapupcode() *schema.Resource {
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
            "division_id": {
                Type:     schema.TypeString,
                Optional: true,
                Computed: true,
            },
            "description": {
                Type:     schema.TypeString,
                Optional: true,
            },
        },
    }
}
```

**Plugin Framework**:
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
            "division_id": schema.StringAttribute{
                Description: "The division to which this routing wrapupcode will belong.",
                Optional:    true,
                Computed:    true,
                PlanModifiers: []planmodifier.String{
                    stringplanmodifier.UseStateForUnknown(),
                },
            },
            "description": schema.StringAttribute{
                Description: "The wrap-up code description.",
                Optional:    true,
            },
        },
    }
}
```

**Key Differences**:

| Aspect | SDKv2 | Plugin Framework |
|--------|-------|------------------|
| Schema type | `map[string]*schema.Schema` | `schema.Schema` struct |
| Attribute type | `*schema.Schema` | `schema.StringAttribute`, etc. |
| Type definition | `Type: schema.TypeString` | `schema.StringAttribute` |
| Description | Optional, often omitted | Required, always included |
| Plan modifiers | `DiffSuppressFunc`, `CustomizeDiff` | `PlanModifiers` array |
| Validation | `ValidateFunc` | `Validators` array |
| Type safety | Runtime type checking | Compile-time type safety |

---

## Design Patterns and Best Practices

### Pattern 1: Computed Attributes with UseStateForUnknown

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

### Pattern 2: Optional + Computed Attributes

**When to Use**:
- Attributes that user can provide OR provider will compute default
- Attributes that have API-side defaults

**Example**:
```go
"division_id": schema.StringAttribute{
    Optional: true,
    Computed: true,
    PlanModifiers: []planmodifier.String{
        stringplanmodifier.UseStateForUnknown(),
    },
},
```

**Behavior**:
- User provides value → Use that value
- User omits value → API computes default (e.g., home division)
- On update, if still omitted → Preserve existing value (UseStateForUnknown)

### Pattern 3: Separate Resource and Data Source Schemas

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

### Pattern 4: Dependency Reference Configuration

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
- Order doesn't matter (exporter handles ordering)

**Example with Multiple Dependencies**:
```go
RefAttrs: map[string]*resourceExporter.RefAttrSettings{
    "division_id": {RefType: "genesyscloud_auth_division"},
    "queue_id":    {RefType: "genesyscloud_routing_queue"},
    "flow_id":     {RefType: "genesyscloud_flow"},
}
```

### Pattern 5: Helper Function Naming

**Convention**: `Generate<ResourceName>Resource()`

**Examples**:
- `GenerateRoutingWrapupcodeResource()`
- `GenerateRoutingQueueResource()`
- `GenerateAuthDivisionBasic()` (simplified variant)

**Parameters**:
1. `resourceLabel` (always first) - Terraform resource label
2. Required attributes (in logical order)
3. Optional attributes (in logical order)

**Return**: HCL configuration string

---

## Migration Considerations

### Backward Compatibility Checklist

When migrating schema from SDKv2 to Framework, verify:

- [ ] All attribute names are identical
- [ ] All attribute types are identical (String → StringAttribute, etc.)
- [ ] Required/Optional/Computed properties match exactly
- [ ] Validation rules are preserved
- [ ] Default values are preserved (if any)
- [ ] Sensitive attributes are marked as sensitive
- [ ] Deprecated attributes are handled (if any)

### Common Migration Pitfalls

#### Pitfall 1: Changing Attribute Names
**Problem**: Changing attribute names breaks existing configurations.
**Solution**: Keep attribute names exactly as in SDKv2.

#### Pitfall 2: Missing Plan Modifiers
**Problem**: Computed attributes show "known after apply" unnecessarily.
**Solution**: Add `UseStateForUnknown()` to computed attributes that don't change.

#### Pitfall 3: Wrong Schema Type
**Problem**: Using `schema.Schema` for data source or vice versa.
**Solution**: Use `schema.Schema` for resources, `datasourceschema.Schema` for data sources.

#### Pitfall 4: Missing Descriptions
**Problem**: Framework encourages descriptions, SDKv2 often omitted them.
**Solution**: Add clear descriptions to all attributes (improves documentation).

#### Pitfall 5: Incorrect RefAttrs
**Problem**: Wrong attribute name or resource type in RefAttrs.
**Solution**: Verify attribute names match schema exactly, resource types match Terraform types.

---

## Testing Considerations

### Schema Validation

**What to Test** (in Stage 3):
- Schema compiles without errors
- All attributes are accessible
- Plan modifiers work correctly
- Validation rules work correctly

**How to Test**:
- Create resource with all attributes
- Create resource with minimal attributes
- Update attributes
- Verify plan shows correct changes

### Helper Function Testing

**What to Test**:
- Generated HCL is valid
- Optional attributes can be omitted
- References work correctly
- String literals are quoted correctly

**How to Test**:
- Use helper function in test cases
- Verify Terraform accepts generated HCL
- Test with and without optional attributes

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

## Summary

### Key Design Decisions

1. **Separate Schema File**: Isolates schema definitions for clarity and maintainability
2. **Framework-Native Patterns**: Uses Plugin Framework idioms (plan modifiers, type safety)
3. **Backward Compatibility**: Preserves all attribute names and behavior
4. **Export Compatibility**: Uses SDK-compatible functions during migration (Phase 1)
5. **Helper Functions**: Provides utilities for cross-package testing

### Schema File Structure

```
resource_genesyscloud_<resource_name>_schema.go
├── Package Constants (ResourceType)
├── SetRegistrar() - Register resource, data source, exporter
├── <ResourceName>ResourceSchema() - Resource schema definition
├── <ResourceName>DataSourceSchema() - Data source schema definition
├── <ResourceName>Exporter() - Exporter configuration
└── Generate<ResourceName>Resource() - Helper function for tests
```

### Next Steps

After completing Stage 1 schema migration:
1. Review schema definitions for accuracy
2. Verify all attributes are correctly defined
3. Confirm exporter configuration is correct
4. Proceed to **Stage 2 – Resource Migration**

---

## References

- **Reference Implementation**: `genesyscloud/routing_wrapupcode/resource_genesyscloud_routing_wrapupcode_schema.go`
- **Plugin Framework Schema Guide**: https://developer.hashicorp.com/terraform/plugin/framework/handling-data/schemas
- **Plan Modifiers Guide**: https://developer.hashicorp.com/terraform/plugin/framework/resources/plan-modification
- **Attribute Types**: https://developer.hashicorp.com/terraform/plugin/framework/handling-data/attributes
