# Stage 1 – Schema Migration Tasks

## Overview

This document provides step-by-step tasks for completing Stage 1 of the Plugin Framework migration. Follow these tasks in order to migrate schema definitions from SDKv2 to Plugin Framework.

**Reference Implementation**: `genesyscloud/routing_wrapupcode/resource_genesyscloud_routing_wrapupcode_schema.go`

**Estimated Time**: 2-4 hours (depending on schema complexity)

---

## Prerequisites

Before starting Stage 1 tasks, ensure:

- [ ] You have reviewed the existing SDKv2 resource implementation
- [ ] You understand all resource attributes and their properties
- [ ] You have read Stage 1 `requirements.md` and `design.md`
- [ ] You have studied the `routing_wrapupcode` reference implementation
- [ ] Development environment is set up and ready

---

## Task Checklist

### Phase 1: Preparation and Analysis
- [ ] Task 1.1: Analyze Existing SDKv2 Schema
- [ ] Task 1.2: Identify Dependencies and References
- [ ] Task 1.3: Plan Attribute Migration

### Phase 2: File Creation and Setup
- [ ] Task 2.1: Create Schema File
- [ ] Task 2.2: Add Package Declaration and Imports
- [ ] Task 2.3: Define Package Constants

### Phase 3: Schema Implementation
- [ ] Task 3.1: Implement Resource Schema Function
- [ ] Task 3.2: Implement Data Source Schema Function
- [ ] Task 3.3: Add Plan Modifiers

### Phase 4: Registration and Export
- [ ] Task 4.1: Implement SetRegistrar Function
- [ ] Task 4.2: Implement Exporter Configuration
- [ ] Task 4.3: Define Dependency References

### Phase 5: Helper Functions
- [ ] Task 5.1: Implement Generate Resource Helper Function
- [ ] Task 5.2: Test Helper Function Output

### Phase 6: Validation and Review
- [ ] Task 6.1: Compile and Verify
- [ ] Task 6.2: Review Against Checklist
- [ ] Task 6.3: Code Review and Approval

---

## Detailed Tasks

## Phase 1: Preparation and Analysis

### Task 1.1: Analyze Existing SDKv2 Schema

**Objective**: Understand the current SDKv2 schema structure and attributes.

**Steps**:

1. **Locate the SDKv2 resource file**
   - File pattern: `resource_genesyscloud_<resource_name>.go`
   - Example: `genesyscloud/routing_wrapupcode/resource_genesyscloud_routing_wrapupcode.go` (old SDKv2 version)

2. **Find the schema definition**
   - Look for `func Resource<ResourceName>() *schema.Resource`
   - Locate the `Schema: map[string]*schema.Schema` section

3. **Document all attributes**
   - Create a list of all attributes
   - Note their types (TypeString, TypeInt, TypeBool, etc.)
   - Note their properties (Required, Optional, Computed)
   - Note any validation rules
   - Note any default values

4. **Example attribute documentation**:
   ```
   Attribute: id
   - Type: TypeString
   - Computed: true
   - Description: The globally unique identifier
   
   Attribute: name
   - Type: TypeString
   - Required: true
   - Description: Resource name
   
   Attribute: division_id
   - Type: TypeString
   - Optional: true
   - Computed: true
   - Description: Division assignment
   
   Attribute: description
   - Type: TypeString
   - Optional: true
   - Description: Resource description
   ```

**Deliverable**: Documented list of all attributes with properties

---

### Task 1.2: Identify Dependencies and References

**Objective**: Identify attributes that reference other resources for exporter configuration.

**Steps**:

1. **Review attribute names for common patterns**
   - Attributes ending in `_id` often reference other resources
   - Examples: `division_id`, `queue_id`, `flow_id`, `skill_id`

2. **Determine referenced resource types**
   - `division_id` → `genesyscloud_auth_division`
   - `queue_id` → `genesyscloud_routing_queue`
   - `flow_id` → `genesyscloud_flow`
   - `skill_id` → `genesyscloud_routing_skill`

3. **Check API documentation**
   - Verify which attributes are foreign keys
   - Confirm the referenced resource types

4. **Document dependencies**:
   ```
   division_id → genesyscloud_auth_division
   ```

**Deliverable**: List of dependency attributes and their referenced resource types

---

### Task 1.3: Plan Attribute Migration

**Objective**: Plan how each SDKv2 attribute will be converted to Framework.

**Steps**:

1. **Map SDKv2 types to Framework types**:
   - `TypeString` → `schema.StringAttribute`
   - `TypeInt` → `schema.Int64Attribute`
   - `TypeBool` → `schema.BoolAttribute`
   - `TypeFloat` → `schema.Float64Attribute`
   - `TypeList` → `schema.ListAttribute` or `schema.ListNestedAttribute`
   - `TypeSet` → `schema.SetAttribute` or `schema.SetNestedAttribute`
   - `TypeMap` → `schema.MapAttribute`

2. **Identify attributes needing plan modifiers**:
   - Computed attributes that don't change → `UseStateForUnknown()`
   - ID attribute → Always use `UseStateForUnknown()`
   - Optional + Computed attributes → Use `UseStateForUnknown()`

3. **Plan validation migration** (if any):
   - SDKv2 `ValidateFunc` → Framework `Validators`
   - Note: Most simple resources don't have complex validation

**Deliverable**: Migration plan for each attribute

---

## Phase 2: File Creation and Setup

### Task 2.1: Create Schema File

**Objective**: Create the new schema file in the correct location.

**Steps**:

1. **Navigate to resource package directory**
   ```powershell
   cd genesyscloud\<resource_name>
   ```
   Example:
   ```powershell
   cd genesyscloud\routing_wrapupcode
   ```

2. **Create the schema file**
   ```powershell
   New-Item -ItemType File -Name "resource_genesyscloud_<resource_name>_schema.go"
   ```
   Example:
   ```powershell
   New-Item -ItemType File -Name "resource_genesyscloud_routing_wrapupcode_schema.go"
   ```

**Deliverable**: Empty schema file created in correct location

---

### Task 2.2: Add Package Declaration and Imports

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
       "fmt"

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

3. **Adjust imports based on attribute types**
   - If using Int64Attribute: Add `"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"`
   - If using BoolAttribute: Add `"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"`
   - If using nested attributes: Add appropriate nested schema imports

**Deliverable**: File with package declaration and imports

---

### Task 2.3: Define Package Constants

**Objective**: Define the resource type constant.

**Steps**:

1. **Add ResourceType constant**
   ```go
   const ResourceType = "genesyscloud_<resource_name>"
   ```
   Example:
   ```go
   const ResourceType = "genesyscloud_routing_wrapupcode"
   ```

2. **Verify constant matches Terraform resource type exactly**
   - Must match the resource type used in Terraform configurations
   - Must match the type used in SDKv2 implementation

**Deliverable**: ResourceType constant defined

---

## Phase 3: Schema Implementation

### Task 3.1: Implement Resource Schema Function

**Objective**: Convert SDKv2 resource schema to Plugin Framework schema.

**Steps**:

1. **Create the function signature**
   ```go
   // <ResourceName>ResourceSchema returns the schema for the <resource_name> resource
   func <ResourceName>ResourceSchema() schema.Schema {
       return schema.Schema{
           Description: "Resource description",
           Attributes: map[string]schema.Attribute{
               // Attributes will be added here
           },
       }
   }
   ```

2. **Add the ID attribute** (always first)
   ```go
   "id": schema.StringAttribute{
       Description: "The globally unique identifier for the <resource>.",
       Computed:    true,
       PlanModifiers: []planmodifier.String{
           stringplanmodifier.UseStateForUnknown(),
       },
   },
   ```

3. **Add required attributes**
   - Convert each required attribute from SDKv2
   - Example:
   ```go
   "name": schema.StringAttribute{
       Description: "<Resource> name.",
       Required:    true,
   },
   ```

4. **Add optional attributes**
   - Convert each optional attribute from SDKv2
   - Example:
   ```go
   "description": schema.StringAttribute{
       Description: "The <resource> description.",
       Optional:    true,
   },
   ```

5. **Add computed attributes**
   - Convert each computed attribute from SDKv2
   - Add plan modifiers as needed
   - Example:
   ```go
   "created_date": schema.StringAttribute{
       Description: "The date the <resource> was created.",
       Computed:    true,
       PlanModifiers: []planmodifier.String{
           stringplanmodifier.UseStateForUnknown(),
       },
   },
   ```

6. **Add optional + computed attributes**
   - These can be provided by user OR computed by API
   - Example:
   ```go
   "division_id": schema.StringAttribute{
       Description: "The division to which this <resource> will belong.",
       Optional:    true,
       Computed:    true,
       PlanModifiers: []planmodifier.String{
           stringplanmodifier.UseStateForUnknown(),
       },
   },
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

**Deliverable**: Complete resource schema function

---

### Task 3.2: Implement Data Source Schema Function

**Objective**: Create the data source schema for resource lookup.

**Steps**:

1. **Create the function signature**
   ```go
   // <ResourceName>DataSourceSchema returns the schema for the <resource_name> data source
   func <ResourceName>DataSourceSchema() datasourceschema.Schema {
       return datasourceschema.Schema{
           Description: "Data source for <Resource>. Select a <resource> by name",
           Attributes: map[string]datasourceschema.Attribute{
               // Attributes will be added here
           },
       }
   }
   ```

2. **Add ID attribute** (computed)
   ```go
   "id": datasourceschema.StringAttribute{
       Description: "The globally unique identifier for the <resource>.",
       Computed:    true,
   },
   ```

3. **Add name attribute** (required for lookup)
   ```go
   "name": datasourceschema.StringAttribute{
       Description: "<Resource> name.",
       Required:    true,
   },
   ```

4. **Add additional lookup attributes if needed**
   - Most data sources only need `id` and `name`
   - Some may need additional criteria (e.g., `division_id` for scoped lookup)

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

**Deliverable**: Complete data source schema function

---

### Task 3.3: Add Plan Modifiers

**Objective**: Ensure plan modifiers are correctly applied to computed attributes.

**Steps**:

1. **Review all computed attributes**
   - Identify attributes that are computed on create and don't change
   - Identify optional + computed attributes

2. **Add UseStateForUnknown to ID attribute**
   ```go
   "id": schema.StringAttribute{
       Computed: true,
       PlanModifiers: []planmodifier.String{
           stringplanmodifier.UseStateForUnknown(),
       },
   },
   ```

3. **Add UseStateForUnknown to other computed attributes**
   - Apply to attributes that don't change after creation
   - Apply to optional + computed attributes

4. **Verify plan modifier imports**
   - Ensure `stringplanmodifier` is imported for string attributes
   - Add other plan modifier imports as needed (int64, bool, etc.)

**Deliverable**: All computed attributes have appropriate plan modifiers

---

## Phase 4: Registration and Export

### Task 4.1: Implement SetRegistrar Function

**Objective**: Register the Framework resource, data source, and exporter.

**Steps**:

1. **Create the function**
   ```go
   // SetRegistrar registers all of the resources, datasources and exporters in the package
   func SetRegistrar(regInstance registrar.Registrar) {
       regInstance.RegisterFrameworkResource(ResourceType, New<ResourceName>FrameworkResource)
       regInstance.RegisterFrameworkDataSource(ResourceType, New<ResourceName>FrameworkDataSource)
       regInstance.RegisterExporter(ResourceType, <ResourceName>Exporter())
   }
   ```

2. **Replace placeholders**
   - `<ResourceName>` → Your resource name in PascalCase
   - Example: `RoutingWrapupcode`

**Example** (routing_wrapupcode):
```go
func SetRegistrar(regInstance registrar.Registrar) {
    regInstance.RegisterFrameworkResource(ResourceType, NewRoutingWrapupcodeFrameworkResource)
    regInstance.RegisterFrameworkDataSource(ResourceType, NewRoutingWrapupcodeFrameworkDataSource)
    regInstance.RegisterExporter(ResourceType, RoutingWrapupcodeExporter())
}
```

**Note**: Constructor functions (`New<ResourceName>FrameworkResource`) will be created in Stage 2.

**Deliverable**: SetRegistrar function implemented

---

### Task 4.2: Implement Exporter Configuration

**Objective**: Configure the exporter for this resource.

**Steps**:

1. **Create the exporter function**
   ```go
   func <ResourceName>Exporter() *resourceExporter.ResourceExporter {
       return &resourceExporter.ResourceExporter{
           GetResourcesFunc: provider.GetAllWithPooledClient(GetAll<ResourceName>SDK),
           RefAttrs: map[string]*resourceExporter.RefAttrSettings{
               // Dependency references will be added here
           },
       }
   }
   ```

2. **Set GetResourcesFunc**
   - Use `provider.GetAllWithPooledClient(GetAll<ResourceName>SDK)`
   - The `GetAll<ResourceName>SDK` function will be created in Stage 2
   - Example: `GetAllRoutingWrapupcodesSDK`

3. **Verify function naming**
   - Function name: `GetAll<ResourceName>SDK` (PascalCase, plural, SDK suffix)
   - Example: `GetAllRoutingWrapupcodesSDK`

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

**Deliverable**: Exporter configuration function implemented

---

### Task 4.3: Define Dependency References

**Objective**: Configure RefAttrs for dependency resolution in exporter.

**Steps**:

1. **Review dependency list from Task 1.2**
   - Identify all attributes that reference other resources

2. **Add each dependency to RefAttrs**
   ```go
   RefAttrs: map[string]*resourceExporter.RefAttrSettings{
       "<attribute_name>": {RefType: "<terraform_resource_type>"},
   }
   ```

3. **Common dependency mappings**:
   - `division_id` → `genesyscloud_auth_division`
   - `queue_id` → `genesyscloud_routing_queue`
   - `flow_id` → `genesyscloud_flow`
   - `skill_id` → `genesyscloud_routing_skill`
   - `wrapupcode_id` → `genesyscloud_routing_wrapupcode`
   - `language_id` → `genesyscloud_routing_language`

4. **Example with multiple dependencies**:
   ```go
   RefAttrs: map[string]*resourceExporter.RefAttrSettings{
       "division_id": {RefType: "genesyscloud_auth_division"},
       "queue_id":    {RefType: "genesyscloud_routing_queue"},
       "flow_id":     {RefType: "genesyscloud_flow"},
   }
   ```

5. **If no dependencies**:
   ```go
   RefAttrs: map[string]*resourceExporter.RefAttrSettings{},
   ```

**Deliverable**: RefAttrs configured with all dependencies

---

## Phase 5: Helper Functions

### Task 5.1: Implement Generate Resource Helper Function

**Objective**: Create a helper function for generating Terraform HCL in tests.

**Steps**:

1. **Create function signature**
   ```go
   // Generate<ResourceName>Resource generates a <resource_name> resource for cross-package testing
   // This function is used by other packages that need to create <resource_name> resources in their tests
   func Generate<ResourceName>Resource(
       resourceLabel string,
       name string,
       // Add other required attributes
       optionalAttr1 string,
       optionalAttr2 string,
   ) string {
   ```

2. **Add parameter for each attribute**
   - `resourceLabel` (always first) - Terraform resource label
   - `name` (usually second) - Resource name
   - Required attributes (in logical order)
   - Optional attributes (in logical order)

3. **Build optional attribute strings**
   ```go
   // Handle optional attributes
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
   ```

4. **Return HCL configuration**
   ```go
   return fmt.Sprintf(`resource "genesyscloud_<resource_name>" "%s" {
       name = "%s"%s%s
   }
   `, resourceLabel, name, optionalAttr1Str, optionalAttr2Str)
   ```

5. **Key points**:
   - Use `util.NullValue` to check if reference attributes should be omitted
   - Don't add quotes for reference attributes (e.g., `division_id = genesyscloud_auth_division.div.id`)
   - Add quotes for string literal attributes (e.g., `description = "My description"`)
   - Use empty string `""` to check if string literals should be omitted

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

**Deliverable**: Generate resource helper function implemented

---

### Task 5.2: Test Helper Function Output

**Objective**: Verify the helper function generates valid HCL.

**Steps**:

1. **Manually test the function output**
   - Call the function with various parameter combinations
   - Verify the generated HCL is valid

2. **Test with all attributes**:
   ```go
   config := GenerateRoutingWrapupcodeResource(
       "test",
       "My Wrapupcode",
       "genesyscloud_auth_division.div.id",
       "Test description",
   )
   // Expected output:
   // resource "genesyscloud_routing_wrapupcode" "test" {
   //     name = "My Wrapupcode"
   //     division_id = genesyscloud_auth_division.div.id
   //     description = "Test description"
   // }
   ```

3. **Test with minimal attributes**:
   ```go
   config := GenerateRoutingWrapupcodeResource(
       "test",
       "My Wrapupcode",
       util.NullValue,
       "",
   )
   // Expected output:
   // resource "genesyscloud_routing_wrapupcode" "test" {
   //     name = "My Wrapupcode"
   // }
   ```

4. **Verify**:
   - [ ] Required attributes are always included
   - [ ] Optional attributes can be omitted
   - [ ] Reference attributes don't have quotes
   - [ ] String literal attributes have quotes
   - [ ] Indentation is correct
   - [ ] No trailing commas

**Deliverable**: Helper function tested and verified

---

## Phase 6: Validation and Review

### Task 6.1: Compile and Verify

**Objective**: Ensure the code compiles without errors.

**Steps**:

1. **Run Go build**
   ```powershell
   go build ./genesyscloud/<resource_name>
   ```
   Example:
   ```powershell
   go build ./genesyscloud/routing_wrapupcode
   ```

2. **Fix any compilation errors**
   - Missing imports
   - Syntax errors
   - Type mismatches

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

### Task 6.2: Review Against Checklist

**Objective**: Verify all requirements are met.

**Steps**:

1. **Use the validation checklist from requirements.md**

   **Schema File**:
   - [ ] File created: `resource_genesyscloud_<resource_name>_schema.go`
   - [ ] Package declaration matches directory name
   - [ ] All required imports are present
   - [ ] No unused imports

   **Resource Schema**:
   - [ ] `<ResourceName>ResourceSchema()` function implemented
   - [ ] Returns `schema.Schema` type
   - [ ] All attributes from SDKv2 are present
   - [ ] Attribute properties (Required/Optional/Computed) match SDKv2
   - [ ] Descriptions are clear and accurate
   - [ ] Plan modifiers are applied to computed attributes
   - [ ] ID attribute uses `UseStateForUnknown()` modifier

   **Data Source Schema**:
   - [ ] `<ResourceName>DataSourceSchema()` function implemented
   - [ ] Returns `datasourceschema.Schema` type
   - [ ] Includes `id` and `name` attributes (at minimum)
   - [ ] Descriptions are clear and accurate

   **Registration**:
   - [ ] `SetRegistrar()` function implemented
   - [ ] Calls `RegisterFrameworkResource()` with correct type name
   - [ ] Calls `RegisterFrameworkDataSource()` with correct type name
   - [ ] Calls `RegisterExporter()` with exporter configuration

   **Exporter Configuration**:
   - [ ] `<ResourceName>Exporter()` function implemented
   - [ ] Returns `*resourceExporter.ResourceExporter`
   - [ ] `GetResourcesFunc` uses SDK-compatible function
   - [ ] `RefAttrs` includes all dependency references
   - [ ] Dependency types are correctly mapped

   **Helper Functions**:
   - [ ] `Generate<ResourceName>Resource()` function implemented
   - [ ] Generates valid Terraform HCL configuration
   - [ ] Handles optional attributes correctly
   - [ ] Supports null values using `util.NullValue`
   - [ ] Includes function comment describing usage

   **Code Quality**:
   - [ ] Code compiles without errors
   - [ ] Code follows Go conventions
   - [ ] Functions have clear comments
   - [ ] No TODO or FIXME comments (unless intentional)

2. **Fix any issues found**

**Deliverable**: All checklist items verified

---

### Task 6.3: Code Review and Approval

**Objective**: Get peer review and approval before proceeding to Stage 2.

**Steps**:

1. **Create pull request or review request**
   - Include link to Stage 1 requirements and design docs
   - Highlight any deviations from standard pattern

2. **Address review comments**
   - Make requested changes
   - Re-verify checklist

3. **Get approval**
   - Obtain approval from reviewer
   - Merge or mark as ready for Stage 2

**Deliverable**: Stage 1 approved and ready for Stage 2

---

## Common Issues and Solutions

### Issue 1: Import Errors

**Problem**: Cannot find package or import errors.

**Solution**:
- Run `go mod tidy` to update dependencies
- Verify import paths are correct
- Check that all required packages are installed

### Issue 2: Type Mismatch Errors

**Problem**: Type mismatch between schema and attribute types.

**Solution**:
- Verify using correct attribute type (e.g., `schema.StringAttribute` not `datasourceschema.StringAttribute`)
- Check that plan modifiers match attribute type (e.g., `stringplanmodifier` for string attributes)

### Issue 3: Missing Plan Modifiers

**Problem**: Computed attributes show "known after apply" unnecessarily.

**Solution**:
- Add `UseStateForUnknown()` plan modifier to computed attributes
- Verify plan modifier import is present

### Issue 4: Helper Function Generates Invalid HCL

**Problem**: Generated HCL has syntax errors or incorrect formatting.

**Solution**:
- Check quote handling (references vs. literals)
- Verify indentation and spacing
- Test with various parameter combinations

### Issue 5: Exporter RefAttrs Not Working

**Problem**: Exporter doesn't resolve dependencies correctly.

**Solution**:
- Verify attribute name matches schema exactly
- Verify resource type matches Terraform resource type
- Check that referenced resource exists

---

## Completion Criteria

Stage 1 is complete when:

- [ ] All tasks in this document are completed
- [ ] All checklist items are verified
- [ ] Code compiles without errors
- [ ] Code review is approved
- [ ] Documentation is updated (if needed)

---

## Next Steps

After Stage 1 completion:

1. **Review and approval**
   - Get team review
   - Address any feedback
   - Get final approval

2. **Proceed to Stage 2**
   - Begin resource implementation
   - Implement CRUD operations
   - Create GetAll functions

3. **Reference Stage 2 documentation**
   - Read Stage 2 `requirements.md`
   - Read Stage 2 `design.md`
   - Follow Stage 2 `tasks.md`

---

## Time Estimates

| Phase | Estimated Time |
|-------|----------------|
| Phase 1: Preparation | 30-60 minutes |
| Phase 2: File Setup | 15-30 minutes |
| Phase 3: Schema Implementation | 60-90 minutes |
| Phase 4: Registration and Export | 30-45 minutes |
| Phase 5: Helper Functions | 30-45 minutes |
| Phase 6: Validation and Review | 30-60 minutes |
| **Total** | **3-5 hours** |

*Note: Times vary based on schema complexity and familiarity with patterns.*

---

## References

- **Reference Implementation**: `genesyscloud/routing_wrapupcode/resource_genesyscloud_routing_wrapupcode_schema.go`
- **Stage 1 Requirements**: `prompts/pf_simple_resource_migration/Stage1/requirements.md`
- **Stage 1 Design**: `prompts/pf_simple_resource_migration/Stage1/design.md`
- **Plugin Framework Documentation**: https://developer.hashicorp.com/terraform/plugin/framework
