# Stage 1 – Schema Migration Tasks

## Overview

This document provides step-by-step tasks for completing Stage 1 of the Plugin Framework migration for complex resources. Follow these tasks in order to migrate schema definitions from SDKv2 to Plugin Framework.

**Reference Implementation**: `genesyscloud/user/resource_genesyscloud_user_schema.go`

**Resource Complexity**: Complex resource with multiple nested blocks, custom validators, plan modifiers, and element type helpers.

**Estimated Time**: 6-10 hours (depending on schema complexity and nesting levels)

---

## Prerequisites

Before starting Stage 1 tasks, ensure:

- [ ] You have reviewed the existing SDKv2 resource implementation
- [ ] You understand all resource attributes, nested blocks, and their properties
- [ ] You have read Stage 1 `requirements.md` and `design.md`
- [ ] You have studied the `user` reference implementation
- [ ] You understand the 25 patterns documented in requirements.md
- [ ] Development environment is set up and ready

---

## Task Checklist

### Phase 1: Preparation and Analysis
- [ ] Task 1.1: Analyze Existing SDKv2 Schema
- [ ] Task 1.2: Identify Nested Block Structures
- [ ] Task 1.3: Identify Dependencies and References
- [ ] Task 1.4: Plan Attribute and Block Migration
- [ ] Task 1.5: Identify Custom Validators and Plan Modifiers

### Phase 2: File Creation and Setup
- [ ] Task 2.1: Create Schema File
- [ ] Task 2.2: Add Package Declaration and Imports
- [ ] Task 2.3: Define Package Constants and Variables

### Phase 3: Schema Implementation - Attributes
- [ ] Task 3.1: Implement Resource Schema Function (Attributes Only)
- [ ] Task 3.2: Add Simple Attributes (Required, Optional, Computed)
- [ ] Task 3.3: Add Attributes with Defaults
- [ ] Task 3.4: Add Managed Attribute Pattern (Primitive Sets/Lists)
- [ ] Task 3.5: Add Plan Modifiers to Attributes

### Phase 4: Schema Implementation - Nested Blocks
- [ ] Task 4.1: Implement Simple Nested Blocks (SetNestedBlock/ListNestedBlock)
- [ ] Task 4.2: Implement Two-Level Nested Blocks (Blocks within Blocks)
- [ ] Task 4.3: Implement Three-Level Nested Blocks (if applicable)
- [ ] Task 4.4: Add Validators to Nested Block Attributes
- [ ] Task 4.5: Add Plan Modifiers to Nested Blocks

### Phase 5: Data Source and Helper Functions
- [ ] Task 5.1: Implement Data Source Schema Function
- [ ] Task 5.2: Implement Helper Functions (Dynamic Descriptions)

### Phase 6: Element Type Helpers
- [ ] Task 6.1: Identify Required Element Type Helpers
- [ ] Task 6.2: Implement Element Type Helper Functions
- [ ] Task 6.3: Verify Element Types Match Schema

### Phase 7: Registration and Export
- [ ] Task 7.1: Implement SetRegistrar Function
- [ ] Task 7.2: Implement Exporter Configuration
- [ ] Task 7.3: Define Dependency References (RefAttrs)
- [ ] Task 7.4: Configure RemoveIfMissing
- [ ] Task 7.5: Configure AllowEmptyArrays and AllowZeroValues

### Phase 8: Validation and Review
- [ ] Task 8.1: Compile and Verify
- [ ] Task 8.2: Review Against Checklist
- [ ] Task 8.3: Cross-Verify with Source Schema
- [ ] Task 8.4: Code Review and Approval

---

## Detailed Tasks

## Phase 1: Preparation and Analysis

### Task 1.1: Analyze Existing SDKv2 Schema

**Objective**: Understand the current SDKv2 schema structure, attributes, and nested blocks.

**Steps**:

1. **Locate the SDKv2 resource file**
   - File pattern: `resource_genesyscloud_<resource_name>.go`
   - Example: `genesyscloud/user/resource_genesyscloud_user.go` (old SDKv2 version if exists)

2. **Find the schema definition**
   - Look for `func Resource<ResourceName>() *schema.Resource`
   - Locate the `Schema: map[string]*schema.Schema` section

3. **Document all top-level attributes**
   - Create a list of all attributes
   - Note their types (TypeString, TypeInt, TypeBool, TypeFloat, etc.)
   - Note their properties (Required, Optional, Computed, Sensitive)
   - Note any validation rules
   - Note any default values

4. **Document all nested structures**
   - Identify TypeSet with Elem: &schema.Resource (nested blocks)
   - Identify TypeList with Elem: &schema.Resource (nested blocks)
   - Note the nesting level (1-level, 2-level, 3-level)

5. **Example attribute documentation**:
   ```
   Attribute: id
   - Type: TypeString
   - Computed: true
   - Description: The ID of the user
   
   Attribute: name
   - Type: TypeString
   - Required: true
   - Description: User's full name
   
   Attribute: state
   - Type: TypeString
   - Optional: true
   - Default: "active"
   - ValidateFunc: StringInSlice(["active", "inactive"])
   
   Attribute: profile_skills
   - Type: TypeSet
   - Optional: true
   - Computed: true
   - Elem: &schema.Schema{Type: TypeString}
   - Description: Profile skills (managed attribute)
   
   Nested Block: routing_skills
   - Type: TypeSet
   - Optional: true
   - Computed: true
   - Elem: &schema.Resource with:
     - skill_id (TypeString, Required)
     - proficiency (TypeFloat, Required, ValidateFunc: FloatBetween(0, 5))
   ```

**Deliverable**: Comprehensive documented list of all attributes and nested blocks with properties

---

### Task 1.2: Identify Nested Block Structures

**Objective**: Map out the nested block hierarchy and structure.

**Steps**:

1. **Create a hierarchy diagram**
   ```
   Resource: user
   ├── Attributes (top-level)
   │   ├── id (Computed)
   │   ├── name (Required)
   │   ├── email (Required)
   │   ├── state (Optional + Computed + Default)
   │   ├── profile_skills (Set of strings, managed)
   │   └── certifications (Set of strings, managed)
   ├── Nested Blocks (1-level)
   │   ├── routing_skills (SetNestedBlock)
   │   │   ├── skill_id (Required)
   │   │   └── proficiency (Required, validator)
   │   ├── routing_languages (SetNestedBlock)
   │   │   ├── language_id (Required)
   │   │   └── proficiency (Required, validator)
   │   └── locations (SetNestedBlock)
   │       ├── location_id (Required)
   │       └── notes (Optional)
   ├── Nested Blocks (2-level)
   │   └── addresses (ListNestedBlock, MaxItems: 1)
   │       ├── other_emails (SetNestedBlock)
   │       │   ├── address (Required)
   │       │   └── type (Optional + Computed + Default)
   │       └── phone_numbers (SetNestedBlock)
   │           ├── number (Optional, validator)
   │           ├── media_type (Optional + Computed + Default)
   │           ├── type (Optional + Computed + Default)
   │           ├── extension (Optional)
   │           └── extension_pool_id (Optional, custom plan modifier)
   └── Nested Blocks (3-level)
       └── routing_utilization (ListNestedBlock, MaxItems: 1)
           ├── call (ListNestedBlock, MaxItems: 1)
           │   ├── maximum_capacity (Required, validator)
           │   ├── interruptible_media_types (Optional, Set of strings)
           │   └── include_non_acd (Optional + Computed + Default)
           ├── callback (ListNestedBlock, MaxItems: 1)
           ├── message (ListNestedBlock, MaxItems: 1)
           ├── email (ListNestedBlock, MaxItems: 1)
           ├── chat (ListNestedBlock, MaxItems: 1)
           └── label_utilizations (ListNestedBlock)
               ├── label_id (Required)
               ├── maximum_capacity (Required, validator)
               └── interrupting_label_ids (Optional, Set of strings)
   ```

2. **Identify nesting patterns**
   - 1-level: Simple nested blocks (routing_skills, routing_languages)
   - 2-level: Blocks within blocks (addresses → phone_numbers)
   - 3-level: Blocks within blocks within blocks (routing_utilization → call → attributes)

3. **Note special characteristics**
   - MaxItems: 1 (use listvalidator.SizeAtMost(1))
   - Custom hash functions (Set identity behavior change)
   - Managed attribute pattern (Optional + Computed, no plan modifier for primitives)

**Deliverable**: Complete hierarchy diagram with nesting levels and special characteristics

---

### Task 1.3: Identify Dependencies and References

**Objective**: Identify attributes that reference other resources for exporter configuration.

**Steps**:

1. **Review attribute names for common patterns**
   - Attributes ending in `_id` often reference other resources
   - Examples: `division_id`, `manager`, `skill_id`, `language_id`, `location_id`

2. **Determine referenced resource types**
   - `division_id` → `genesyscloud_auth_division`
   - `manager` → `genesyscloud_user` (self-reference)
   - `routing_skills.skill_id` → `genesyscloud_routing_skill`
   - `routing_languages.language_id` → `genesyscloud_routing_language`
   - `locations.location_id` → `genesyscloud_location`
   - `addresses.phone_numbers.extension_pool_id` → `genesyscloud_telephony_providers_edges_extension_pool`

3. **Note nested references**
   - Use dot notation: `"routing_skills.skill_id"`
   - Deeply nested: `"addresses.phone_numbers.extension_pool_id"`

4. **Document dependencies**:
   ```
   Top-level:
   - manager → genesyscloud_user
   - division_id → genesyscloud_auth_division
   
   Nested (1-level):
   - routing_skills.skill_id → genesyscloud_routing_skill
   - routing_languages.language_id → genesyscloud_routing_language
   - locations.location_id → genesyscloud_location
   
   Nested (2-level):
   - addresses.phone_numbers.extension_pool_id → genesyscloud_telephony_providers_edges_extension_pool
   ```

**Deliverable**: Complete list of dependency attributes with dot notation and referenced resource types

---

### Task 1.4: Plan Attribute and Block Migration

**Objective**: Plan how each SDKv2 attribute and block will be converted to Framework.

**Steps**:

1. **Map SDKv2 types to Framework types**:
   - `TypeString` → `schema.StringAttribute`
   - `TypeInt` → `schema.Int64Attribute` (Note: Int → Int64)
   - `TypeBool` → `schema.BoolAttribute`
   - `TypeFloat` → `schema.Float64Attribute`
   - `TypeList` (primitives) → `schema.ListAttribute`
   - `TypeSet` (primitives) → `schema.SetAttribute`
   - `TypeList` (nested) → `schema.ListNestedBlock`
   - `TypeSet` (nested) → `schema.SetNestedBlock`

2. **Identify attributes needing plan modifiers**:
   - Computed attributes that don't change → `UseStateForUnknown()`
   - ID attribute → Always use `UseStateForUnknown()`
   - Optional + Computed attributes → May need `UseStateForUnknown()`
   - Nested blocks (emulate SDKv2 Optional + Computed) → `UseStateForUnknown()`
   - Custom plan modifiers → `phoneplan.NullIfEmpty{}` for extension_pool_id

3. **Identify attributes needing defaults**:
   - SDKv2 `Default: "value"` → Framework `Default: stringdefault.StaticString("value")`
   - Must be `Optional + Computed + Default`

4. **Plan validation migration**:
   - `ValidateFunc: StringInSlice(...)` → `stringvalidator.OneOf(...)`
   - `ValidateFunc: FloatBetween(...)` → `float64validator.Between(...)`
   - `ValidateFunc: IntBetween(...)` → `int64validator.Between(...)`
   - Custom validators → `validators.FWValidatePhoneNumber()`, `validators.FWValidateDate()`
   - `MaxItems: 1` → `listvalidator.SizeAtMost(1)`

5. **Plan nested block migration**:
   - SetNestedBlock: NO Optional/Required at block level, use plan modifier
   - ListNestedBlock: NO Optional/Required at block level, use plan modifier
   - Inner attributes: CAN be Required/Optional
   - Blocks within blocks: Use `Blocks` map, not `Attributes` map

**Deliverable**: Comprehensive migration plan for each attribute and block

---

### Task 1.5: Identify Custom Validators and Plan Modifiers

**Objective**: Identify where custom validators and plan modifiers are needed.

**Steps**:

1. **Review SDKv2 ValidateFunc**
   - Standard validators (StringInSlice, FloatBetween) → Framework equivalents
   - Custom validators (phone numbers, dates) → Custom Framework validators

2. **Identify custom validator needs**:
   ```
   Phone number validation → validators.FWValidatePhoneNumber()
   Date validation → validators.FWValidateDate()
   ```

3. **Identify custom plan modifier needs**:
   ```
   extension_pool_id → phoneplan.NullIfEmpty{}
   (Future) E.164 canonicalization → phoneplan.E164{DefaultRegion: "US"}
   ```

4. **Verify custom validators exist**
   - Check `genesyscloud/validators` package
   - If missing, note for creation (out of scope for Stage 1, but document)

5. **Verify custom plan modifiers exist**
   - Check `genesyscloud/util/phoneplan` package
   - If missing, note for creation (out of scope for Stage 1, but document)

**Deliverable**: List of custom validators and plan modifiers needed

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
   cd genesyscloud\user
   ```

2. **Create the schema file**
   ```powershell
   New-Item -ItemType File -Name "resource_genesyscloud_<resource_name>_schema.go"
   ```
   Example:
   ```powershell
   New-Item -ItemType File -Name "resource_genesyscloud_user_schema.go"
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
   package user
   ```

2. **Add required imports** (comprehensive list for complex resources)
   ```go
   import (
       "fmt"      // For dynamic descriptions
       "strings"  // For string manipulation in descriptions
       
       // Framework validators (add as needed)
       "github.com/hashicorp/terraform-plugin-framework-validators/float64validator"
       "github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
       listvalidator "github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
       "github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
       
       // Framework schema packages
       datasourceschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
       "github.com/hashicorp/terraform-plugin-framework/resource/schema"
       
       // Framework defaults (add as needed)
       "github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
       "github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
       
       // Framework plan modifiers (add as needed)
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
       "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/phoneplan"  // If custom plan modifiers needed
       "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/validators"
   )
   ```

3. **Adjust imports based on actual needs**
   - Remove unused imports
   - Add additional validator imports as needed
   - Add custom plan modifier imports as needed

**Deliverable**: File with package declaration and imports

---

### Task 2.3: Define Package Constants and Variables

**Objective**: Define the resource type constant and any package-level variables.

**Steps**:

1. **Add ResourceType constant**
   ```go
   const ResourceType = "genesyscloud_<resource_name>"
   ```
   Example:
   ```go
   const ResourceType = "genesyscloud_user"
   ```

2. **Add package-level variables** (if needed)
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

3. **Verify constants and variables are needed**
   - Only add if used in multiple places
   - Used for dynamic descriptions or validation
   - Reduces magic strings

**Deliverable**: ResourceType constant and package variables defined

---

## Phase 3: Schema Implementation - Attributes

### Task 3.1: Implement Resource Schema Function (Attributes Only)

**Objective**: Create the resource schema function structure with top-level attributes only (no blocks yet).

**Steps**:

1. **Create the function signature**
   ```go
   // <ResourceName>ResourceSchema returns the schema for the <resource_name> resource
   func <ResourceName>ResourceSchema() schema.Schema {
       return schema.Schema{
           Description: `Resource description.

Export block label: "{<label_field>}"`,
           Attributes: map[string]schema.Attribute{
               // Attributes will be added here
           },
           Blocks: map[string]schema.Block{
               // Blocks will be added in Phase 4
           },
       }
   }
   ```

2. **Add export block label**
   - Specify which field is used as the label in exports
   - Example: `Export block label: "{email}"` for user resource

**Deliverable**: Resource schema function structure created

---

### Task 3.2: Add Simple Attributes (Required, Optional, Computed)

**Objective**: Add all simple top-level attributes to the schema.

**Steps**:

1. **Add the ID attribute** (always first)
   ```go
   "id": schema.StringAttribute{
       Description: "The ID of the <resource>.",
       Computed:    true,
       PlanModifiers: []planmodifier.String{
           stringplanmodifier.UseStateForUnknown(),
       },
   },
   ```

2. **Add required attributes**
   ```go
   "name": schema.StringAttribute{
       Description: "User's full name.",
       Required:    true,
   },
   
   "email": schema.StringAttribute{
       Description: "User's primary email and username.",
       Required:    true,
   },
   ```

3. **Add optional attributes**
   ```go
   "department": schema.StringAttribute{
       Description: "User's department.",
       Optional:    true,
   },
   
   "title": schema.StringAttribute{
       Description: "User's title.",
       Optional:    true,
   },
   
   "manager": schema.StringAttribute{
       Description: "User ID of this user's manager.",
       Optional:    true,
   },
   ```

4. **Add computed attributes** (if any beyond ID)
   ```go
   "created_date": schema.StringAttribute{
       Description: "The date the user was created.",
       Computed:    true,
       PlanModifiers: []planmodifier.String{
           stringplanmodifier.UseStateForUnknown(),
       },
   },
   ```

5. **Add optional + computed attributes** (no default)
   ```go
   "division_id": schema.StringAttribute{
       Description: "The division to which this user will belong. If not set, the home division will be used.",
       Optional:    true,
       Computed:    true,
       // NO Default - API will compute if not provided
       // NO UseStateForUnknown - we want API to recompute if config changes
   },
   ```

6. **Add sensitive attributes** (if any)
   ```go
   "password": schema.StringAttribute{
       Description: "User's password. If specified, this is only set on user create.",
       Optional:    true,
       Sensitive:   true,
   },
   ```

**Deliverable**: All simple attributes added to schema

---

### Task 3.3: Add Attributes with Defaults

**Objective**: Add attributes that have default values.

**Steps**:

1. **Add string attributes with defaults**
   ```go
   "state": schema.StringAttribute{
       Description: "User's state (active | inactive). Default is 'active'.",
       Optional:    true,
       Computed:    true,  // REQUIRED when using Default
       Default:     stringdefault.StaticString("active"),
       Validators: []validator.String{
           stringvalidator.OneOf("active", "inactive"),
       },
   },
   ```

2. **Add boolean attributes with defaults**
   ```go
   "acd_auto_answer": schema.BoolAttribute{
       Description: "Enable ACD auto-answer.",
       Optional:    true,
       Computed:    true,  // REQUIRED when using Default
       Default:     booldefault.StaticBool(false),
   },
   ```

3. **Verify all attributes with defaults are Optional + Computed**
   - Framework requires both
   - Will cause compilation error if missing Computed

**Deliverable**: All attributes with defaults added

---

### Task 3.4: Add Managed Attribute Pattern (Primitive Sets/Lists)

**Objective**: Add set/list attributes that use the managed attribute pattern.

**Steps**:

1. **Add set of strings (managed attribute)**
   ```go
   "profile_skills": schema.SetAttribute{
       Description: "Profile skills for this user. If not set, this resource will not manage profile skills.",
       Optional:    true,
       Computed:    true,
       ElementType: types.StringType,
   },
   
   "certifications": schema.SetAttribute{
       Description: "Certifications for this user. If not set, this resource will not manage certifications.",
       Optional:    true,
       Computed:    true,
       ElementType: types.StringType,
   },
   ```

2. **Key characteristics**:
   - Optional + Computed (no plan modifier needed for primitive sets)
   - Description explicitly states "If not set, this resource will not manage..."
   - ElementType specified

**Deliverable**: Managed attribute pattern sets/lists added

---

### Task 3.5: Add Plan Modifiers to Attributes

**Objective**: Ensure all attributes have appropriate plan modifiers.

**Steps**:

1. **Review all computed attributes**
   - ID attribute → `UseStateForUnknown()` (already added)
   - Other computed attributes → `UseStateForUnknown()` if they don't change

2. **Review optional + computed attributes**
   - With Default → No additional plan modifier needed (default handles it)
   - Without Default → May need `UseStateForUnknown()` depending on behavior

3. **Add custom plan modifiers** (if needed)
   ```go
   "extension_pool_id": schema.StringAttribute{
       Description:   "Id of the extension pool which contains this extension.",
       Optional:      true,
       PlanModifiers: []planmodifier.String{
           phoneplan.NullIfEmpty{},
       },
   },
   ```

4. **Verify plan modifier imports**
   - `stringplanmodifier` for string attributes
   - `int64planmodifier` for int64 attributes
   - `boolplanmodifier` for bool attributes
   - Custom plan modifiers from internal packages

**Deliverable**: All attributes have appropriate plan modifiers

---

## Phase 4: Schema Implementation - Nested Blocks

### Task 4.1: Implement Simple Nested Blocks (SetNestedBlock/ListNestedBlock)

**Objective**: Add 1-level nested blocks to the schema.

**Steps**:

1. **Add SetNestedBlock for collections**
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
   },
   ```

2. **Add ListNestedBlock with MaxItems: 1**
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
   },
   ```

3. **Key points for simple nested blocks**:
   - NO Optional/Required at block level
   - Use `UseStateForUnknown()` plan modifier
   - Inner attributes CAN be Required/Optional
   - Add validators to inner attributes as needed
   - Use `listvalidator.SizeAtMost(1)` for MaxItems: 1

4. **Repeat for all 1-level nested blocks**
   - routing_skills
   - routing_languages
   - locations
   - employer_info
   - voicemail_userpolicies

**Deliverable**: All 1-level nested blocks added

---

### Task 4.2: Implement Two-Level Nested Blocks (Blocks within Blocks)

**Objective**: Add nested blocks that contain other nested blocks.

**Steps**:

1. **Create parent block structure**
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
               // Child blocks will be added here
           },
       },
   },
   ```

2. **Add child blocks within parent**
   ```go
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
   ```

3. **Critical rule**: When nesting blocks within blocks, use `Blocks: map[string]schema.Block{}` in the parent's `NestedBlockObject`, NOT `Attributes`

4. **Common mistake to avoid**:
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

**Deliverable**: All 2-level nested blocks added

---

### Task 4.3: Implement Three-Level Nested Blocks (if applicable)

**Objective**: Add deeply nested blocks (blocks within blocks within blocks).

**Steps**:

1. **Create level 1 block**
   ```go
   "routing_utilization": schema.ListNestedBlock{
       Description: "The routing utilization settings for this user. If empty list, the org default settings are used. If not set, this resource will not manage the users's utilization settings.",
       PlanModifiers: []planmodifier.List{
           listplanmodifier.UseStateForUnknown(),
       },
       Validators: []validator.List{
           listvalidator.SizeAtMost(1),
       },
       NestedObject: schema.NestedBlockObject{
           Blocks: map[string]schema.Block{  // Level 2 blocks go here
               // Media type blocks will be added here
           },
       },
   },
   ```

2. **Add level 2 blocks (media types)**
   ```go
   "call": schema.ListNestedBlock{
       Description: "Call media settings. If not set, this reverts to the default media type settings.",
       PlanModifiers: []planmodifier.List{
           listplanmodifier.UseStateForUnknown(),
       },
       Validators: []validator.List{
           listvalidator.SizeAtMost(1),
       },
       NestedObject: schema.NestedBlockObject{
           Attributes: map[string]schema.Attribute{  // Level 3 attributes
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
   ```

3. **Repeat for all media types**
   - call
   - callback
   - message
   - email
   - chat

4. **Add label_utilizations block** (if applicable)
   ```go
   "label_utilizations": schema.ListNestedBlock{
       Description: "Label utilization settings. If not set, default label settings will be applied. This is in PREVIEW and should not be used unless the feature is available to your organization.",
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
   ```

5. **Key points for 3-level nesting**:
   - Level 1: Parent block (routing_utilization)
   - Level 2: Media type blocks (call, callback, etc.) - use `Blocks` map
   - Level 3: Attributes (maximum_capacity, etc.) - use `Attributes` map
   - Each level can have its own plan modifiers and validators

**Deliverable**: All 3-level nested blocks added (if applicable)

---

### Task 4.4: Add Validators to Nested Block Attributes

**Objective**: Ensure all nested block attributes have appropriate validators.

**Steps**:

1. **Review all nested block attributes**
   - Identify attributes with validation rules in SDKv2

2. **Add standard validators**
   ```go
   // String OneOf
   Validators: []validator.String{
       stringvalidator.OneOf("WORK", "HOME"),
   }
   
   // Float Between
   Validators: []validator.Float64{
       float64validator.Between(0, 5),
   }
   
   // Int Between
   Validators: []validator.Int64{
       int64validator.Between(0, 25),
   }
   
   // List SizeAtMost
   Validators: []validator.List{
       listvalidator.SizeAtMost(1),
   }
   ```

3. **Add custom validators**
   ```go
   // Phone number validation
   Validators: []validator.String{
       validators.FWValidatePhoneNumber(),
   }
   
   // Date validation
   Validators: []validator.String{
       validators.FWValidateDate(),
   }
   ```

4. **Verify validator imports**
   - Standard validators from Framework
   - Custom validators from internal packages

**Deliverable**: All nested block attributes have appropriate validators

---

### Task 4.5: Add Plan Modifiers to Nested Blocks

**Objective**: Ensure all nested blocks have appropriate plan modifiers.

**Steps**:

1. **Add UseStateForUnknown to all nested blocks**
   ```go
   // SetNestedBlock
   PlanModifiers: []planmodifier.Set{
       setplanmodifier.UseStateForUnknown(),
   }
   
   // ListNestedBlock
   PlanModifiers: []planmodifier.List{
       listplanmodifier.UseStateForUnknown(),
   }
   ```

2. **Verify plan modifier imports**
   - `setplanmodifier` for SetNestedBlock
   - `listplanmodifier` for ListNestedBlock

3. **Add custom plan modifiers to nested attributes** (if needed)
   ```go
   "extension_pool_id": schema.StringAttribute{
       PlanModifiers: []planmodifier.String{
           phoneplan.NullIfEmpty{},
       },
   },
   ```

**Deliverable**: All nested blocks have appropriate plan modifiers

---

## Phase 5: Data Source and Helper Functions

### Task 5.1: Implement Data Source Schema Function

**Objective**: Create the data source schema for resource lookup.

**Steps**:

1. **Create the function signature**
   ```go
   // <ResourceName>DataSourceSchema returns the schema for the <resource_name> data source
   func <ResourceName>DataSourceSchema() datasourceschema.Schema {
       return datasourceschema.Schema{
           Description: "Data source for <Resource>. Select a <resource> by <criteria>",
           Attributes: map[string]datasourceschema.Attribute{
               // Attributes will be added here
           },
       }
   }
   ```

2. **Add ID attribute** (computed)
   ```go
   "id": datasourceschema.StringAttribute{
       Description: "The ID of the <resource>.",
       Computed:    true,
   },
   ```

3. **Add lookup attributes** (optional)
   ```go
   "email": datasourceschema.StringAttribute{
       Description: "<Resource> email.",
       Optional:    true,
   },
   
   "name": datasourceschema.StringAttribute{
       Description: "<Resource> name.",
       Optional:    true,
   },
   ```

4. **Add description explaining lookup behavior**
   - Example: "If both email & name are specified, the name won't be used for user lookup"

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

**Deliverable**: Complete data source schema function

---

### Task 5.2: Implement Helper Functions (Dynamic Descriptions)

**Objective**: Create helper functions for generating dynamic schema content.

**Steps**:

1. **Identify need for helper functions**
   - Dynamic descriptions that include lists of valid values
   - Repeated logic in schema definitions

2. **Create helper function** (if needed)
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

3. **Use helper function in schema**
   ```go
   "interruptible_media_types": schema.SetAttribute{
       Description: fmt.Sprintf("Set of other media types that can interrupt this media type (%s).", 
           strings.Join(getSdkUtilizationTypes(), " | ")),
       Optional:    true,
       ElementType: types.StringType,
   },
   ```

4. **Verify imports**
   - `fmt` for string formatting
   - `strings` for string manipulation
   - `sort` for sorting (if needed)

**Deliverable**: Helper functions implemented (if needed)

---

## Phase 6: Element Type Helpers

### Task 6.1: Identify Required Element Type Helpers

**Objective**: Determine which element type helper functions are needed.

**Steps**:

1. **Review all nested blocks**
   - Every SetNestedBlock needs an element type helper
   - Every ListNestedBlock needs an element type helper

2. **Create list of required helpers**
   ```
   For user resource:
   - routingSkillsElementType()
   - routingLanguagesElementType()
   - locationsElementType()
   - employerInfoElementType()
   - voicemailUserpoliciesElementType()
   - (Additional helpers for other nested blocks)
   ```

3. **Note the attributes in each block**
   - Document attribute names and types
   - Will be used to create element type definitions

**Deliverable**: List of required element type helper functions

---

### Task 6.2: Implement Element Type Helper Functions

**Objective**: Create element type helper functions for all nested blocks.

**Steps**:

1. **Create helper function for each nested block**
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

2. **Example implementations**:
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

3. **Naming convention**:
   - Function name: `<blockName>ElementType()`
   - Use camelCase, not snake_case
   - Match the block name from schema

4. **Type mapping**:
   - StringAttribute → types.StringType
   - Int64Attribute → types.Int64Type
   - Float64Attribute → types.Float64Type
   - BoolAttribute → types.BoolType

5. **Add import for attr package** (if not already present)
   ```go
   "github.com/hashicorp/terraform-plugin-framework/attr"
   ```

**Deliverable**: All element type helper functions implemented

---

### Task 6.3: Verify Element Types Match Schema

**Objective**: Ensure element type definitions exactly match schema definitions.

**Steps**:

1. **For each element type helper**:
   - Compare AttrTypes map with schema NestedBlockObject Attributes
   - Verify attribute names match exactly
   - Verify attribute types match exactly

2. **Example verification**:
   ```
   Schema:
   "routing_skills": schema.SetNestedBlock{
       NestedObject: schema.NestedBlockObject{
           Attributes: map[string]schema.Attribute{
               "skill_id": schema.StringAttribute{...},      // StringType
               "proficiency": schema.Float64Attribute{...},  // Float64Type
           },
       },
   }
   
   Element Type:
   func routingSkillsElementType() types.ObjectType {
       return types.ObjectType{
           AttrTypes: map[string]attr.Type{
               "skill_id":    types.StringType,    // ✅ Matches
               "proficiency": types.Float64Type,   // ✅ Matches
           },
       }
   }
   ```

3. **Common mistakes to avoid**:
   - Wrong type (Int64 vs Float64)
   - Missing attributes
   - Extra attributes
   - Typos in attribute names

4. **Create checklist**:
   - [ ] routingSkillsElementType matches routing_skills schema
   - [ ] routingLanguagesElementType matches routing_languages schema
   - [ ] locationsElementType matches locations schema
   - [ ] employerInfoElementType matches employer_info schema
   - [ ] voicemailUserpoliciesElementType matches voicemail_userpolicies schema
   - [ ] (Additional element types as needed)

**Deliverable**: All element types verified to match schema

---

## Phase 7: Registration and Export

### Task 7.1: Implement SetRegistrar Function

**Objective**: Register the Framework resource, data source, and exporter.

**Steps**:

1. **Create the function**
   ```go
   // SetRegistrar registers all the resources and exporters in the package
   func SetRegistrar(l registrar.Registrar) {
       l.RegisterFrameworkDataSource(ResourceType, New<ResourceName>FrameworkDataSource)
       l.RegisterFrameworkResource(ResourceType, New<ResourceName>FrameworkResource)
       l.RegisterExporter(ResourceType, <ResourceName>Exporter())
   }
   ```

2. **Replace placeholders**
   - `<ResourceName>` → Your resource name in PascalCase
   - Example: `User`

**Example** (user):
```go
func SetRegistrar(l registrar.Registrar) {
    l.RegisterFrameworkDataSource(ResourceType, NewUserFrameworkDataSource)
    l.RegisterFrameworkResource(ResourceType, NewUserFrameworkResource)
    l.RegisterExporter(ResourceType, UserExporter())
}
```

**Note**: Constructor functions (`New<ResourceName>FrameworkResource`) will be created in Stage 2.

**Deliverable**: SetRegistrar function implemented

---

### Task 7.2: Implement Exporter Configuration

**Objective**: Configure the exporter for this resource.

**Steps**:

1. **Create the exporter function**
   ```go
   func <ResourceName>Exporter() *resourceExporter.ResourceExporter {
       return &resourceExporter.ResourceExporter{
           GetResourcesFunc: provider.GetAllWithPooledClient(GetAll<ResourceName>SDK),
           RefAttrs: map[string]*resourceExporter.RefAttrSettings{
               // Dependency references will be added in Task 7.3
           },
           RemoveIfMissing: map[string][]string{
               // Will be added in Task 7.4
           },
           AllowEmptyArrays: []string{
               // Will be added in Task 7.5
           },
           AllowZeroValues: []string{
               // Will be added in Task 7.5
           },
       }
   }
   ```

2. **Set GetResourcesFunc**
   - Use `provider.GetAllWithPooledClient(GetAll<ResourceName>SDK)`
   - The `GetAll<ResourceName>SDK` function will be created in Stage 2
   - Example: `GetAllUsersSDK`

3. **Verify function naming**
   - Function name: `GetAll<ResourceName>SDK` (PascalCase, plural, SDK suffix)
   - Example: `GetAllUsersSDK`

**Example** (user - structure only):
```go
func UserExporter() *resourceExporter.ResourceExporter {
    return &resourceExporter.ResourceExporter{
        GetResourcesFunc: provider.GetAllWithPooledClient(GetAllUsersSDK),
        RefAttrs: map[string]*resourceExporter.RefAttrSettings{
            // Will be added in next task
        },
        RemoveIfMissing: map[string][]string{
            // Will be added in Task 7.4
        },
        AllowEmptyArrays: []string{
            // Will be added in Task 7.5
        },
        AllowZeroValues: []string{
            // Will be added in Task 7.5
        },
    }
}
```

**Deliverable**: Exporter configuration function structure implemented

---

### Task 7.3: Define Dependency References (RefAttrs)

**Objective**: Configure RefAttrs for dependency resolution in exporter.

**Steps**:

1. **Review dependency list from Task 1.3**
   - Identify all attributes that reference other resources
   - Include nested references with dot notation

2. **Add each dependency to RefAttrs**
   ```go
   RefAttrs: map[string]*resourceExporter.RefAttrSettings{
       "<attribute_name>": {RefType: "<terraform_resource_type>"},
   }
   ```

3. **Add top-level dependencies**
   ```go
   "manager":     {RefType: ResourceType},  // Self-reference
   "division_id": {RefType: "genesyscloud_auth_division"},
   ```

4. **Add nested dependencies (1-level)**
   ```go
   "routing_skills.skill_id":     {RefType: "genesyscloud_routing_skill"},
   "routing_languages.language_id": {RefType: "genesyscloud_routing_language"},
   "locations.location_id":       {RefType: "genesyscloud_location"},
   ```

5. **Add deeply nested dependencies (2-level)**
   ```go
   "addresses.phone_numbers.extension_pool_id": {RefType: "genesyscloud_telephony_providers_edges_extension_pool"},
   ```

6. **Complete example** (user):
   ```go
   RefAttrs: map[string]*resourceExporter.RefAttrSettings{
       "manager":                                   {RefType: ResourceType},
       "division_id":                               {RefType: "genesyscloud_auth_division"},
       "routing_skills.skill_id":                   {RefType: "genesyscloud_routing_skill"},
       "routing_languages.language_id":             {RefType: "genesyscloud_routing_language"},
       "locations.location_id":                     {RefType: "genesyscloud_location"},
       "addresses.phone_numbers.extension_pool_id": {RefType: "genesyscloud_telephony_providers_edges_extension_pool"},
   }
   ```

7. **Key rules**:
   - Use dot notation for nested attributes
   - Self-references use `ResourceType` constant
   - External references use full Terraform resource type
   - Attribute names must match schema exactly

**Deliverable**: RefAttrs configured with all dependencies

---

### Task 7.4: Configure RemoveIfMissing

**Objective**: Configure blocks to be removed if required fields are missing.

**Steps**:

1. **Identify blocks with required fields**
   - Review nested blocks
   - Identify which fields are essential

2. **Add RemoveIfMissing configuration**
   ```go
   RemoveIfMissing: map[string][]string{
       "<block_name>": {"<required_field>"},
   }
   ```

3. **Example** (user):
   ```go
   RemoveIfMissing: map[string][]string{
       "routing_skills":         {"skill_id"},
       "routing_languages":      {"language_id"},
       "locations":              {"location_id"},
       "voicemail_userpolicies": {"alert_timeout_seconds"},
   }
   ```

4. **How it works**:
   - Key = block name
   - Value = array of required fields
   - If any required field is missing, remove the entire block from export

5. **When to use**:
   - Blocks where certain fields are essential
   - API returns partial data that would be invalid in Terraform
   - Prevents export errors

**Deliverable**: RemoveIfMissing configured

---

### Task 7.5: Configure AllowEmptyArrays and AllowZeroValues

**Objective**: Configure special handling for empty arrays and zero values.

**Steps**:

1. **Configure AllowEmptyArrays**
   - Identify arrays that can be explicitly empty (not null)
   - Example: User with no skills should export as `routing_skills = []`, not omitted

   ```go
   AllowEmptyArrays: []string{
       "routing_skills",
       "routing_languages",
   }
   ```

2. **Configure AllowZeroValues**
   - Identify numeric fields where 0 is a valid value
   - Critical for proficiency ratings and capacity values

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

3. **Why important**:
   - AllowEmptyArrays: Distinguishes between "not set" (null) and "explicitly empty" ([])
   - AllowZeroValues: Prevents 0 from being treated as "not set"

4. **Complete example** (user):
   ```go
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
   ```

**Deliverable**: AllowEmptyArrays and AllowZeroValues configured

---

## Phase 8: Validation and Review

### Task 8.1: Compile and Verify

**Objective**: Ensure the code compiles without errors.

**Steps**:

1. **Run Go build**
   ```powershell
   go build ./genesyscloud/<resource_name>
   ```
   Example:
   ```powershell
   go build ./genesyscloud/user
   ```

2. **Fix any compilation errors**
   - Missing imports
   - Syntax errors
   - Type mismatches
   - Undefined functions (element type helpers, etc.)

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

### Task 8.2: Review Against Checklist

**Objective**: Verify all requirements are met.

**Steps**:

1. **Use the validation checklist from requirements.md**

   **Schema File**:
   - [ ] File created: `resource_genesyscloud_<resource_name>_schema.go`
   - [ ] Package declaration matches directory name
   - [ ] All required imports are present
   - [ ] No unused imports
   - [ ] Imports are organized logically

   **Resource Schema**:
   - [ ] `<ResourceName>ResourceSchema()` function implemented
   - [ ] Returns `schema.Schema` type
   - [ ] All attributes from SDKv2 are present
   - [ ] Attribute properties (Required/Optional/Computed/Sensitive) match SDKv2
   - [ ] Descriptions are clear and accurate
   - [ ] Plan modifiers are applied to computed attributes
   - [ ] ID attribute uses `UseStateForUnknown()` modifier
   - [ ] Attributes with defaults use Optional + Computed + Default pattern

   **Nested Blocks**:
   - [ ] SetNestedBlock used for set-based collections
   - [ ] ListNestedBlock used for list-based collections
   - [ ] Plan modifiers applied to blocks (UseStateForUnknown)
   - [ ] Inner attributes have correct Required/Optional properties
   - [ ] Nested blocks within blocks properly structured (use Blocks map)
   - [ ] Validators applied to nested block attributes
   - [ ] Three-level nesting implemented correctly (if applicable)

   **Validators**:
   - [ ] All SDKv2 ValidateFunc converted to Framework validators
   - [ ] String validators (OneOf) implemented
   - [ ] Numeric validators (Between) implemented
   - [ ] List/Set validators (SizeAtMost) implemented
   - [ ] Custom validators from validators package used

   **Element Type Definitions**:
   - [ ] Helper functions created for complex nested object types
   - [ ] Element types match schema definitions exactly
   - [ ] Element types are reusable

   **Data Source Schema**:
   - [ ] `<ResourceName>DataSourceSchema()` function implemented
   - [ ] Returns `datasourceschema.Schema` type
   - [ ] Includes required lookup attributes
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
   - [ ] `RemoveIfMissing` configured (if needed)
   - [ ] `AllowEmptyArrays` configured (if needed)
   - [ ] `AllowZeroValues` configured (if needed)

   **Code Quality**:
   - [ ] Code compiles without errors
   - [ ] Code follows Go conventions
   - [ ] Functions have clear comments
   - [ ] No TODO or FIXME comments (unless intentional and documented)
   - [ ] Complex patterns are documented

2. **Fix any issues found**

**Deliverable**: All checklist items verified

---

### Task 8.3: Cross-Verify with Source Schema

**Objective**: Ensure nothing was missed from the original SDKv2 schema.

**Steps**:

1. **Open SDKv2 schema side-by-side with new schema**
   - SDKv2: `resource_genesyscloud_<resource_name>.go` (old version)
   - Framework: `resource_genesyscloud_<resource_name>_schema.go` (new)

2. **Verify each SDKv2 attribute is migrated**
   - [ ] All top-level attributes present
   - [ ] All nested blocks present
   - [ ] All nested block attributes present
   - [ ] All validation rules migrated
   - [ ] All default values migrated

3. **Check for differences**
   - Acceptable: Set identity behavior change (documented)
   - Acceptable: Plan modifier differences (Framework pattern)
   - Not acceptable: Missing attributes
   - Not acceptable: Wrong types
   - Not acceptable: Missing validation

4. **Document any intentional differences**
   - Set identity behavior change
   - Custom plan modifiers
   - TODO comments for future enhancements

5. **Create comparison checklist**:
   ```
   Top-level attributes:
   - [ ] id
   - [ ] email
   - [ ] name
   - [ ] password
   - [ ] state
   - [ ] division_id
   - [ ] department
   - [ ] title
   - [ ] manager
   - [ ] acd_auto_answer
   - [ ] profile_skills
   - [ ] certifications
   
   Nested blocks (1-level):
   - [ ] routing_skills
   - [ ] routing_languages
   - [ ] locations
   - [ ] employer_info
   - [ ] voicemail_userpolicies
   
   Nested blocks (2-level):
   - [ ] addresses
   - [ ] addresses.other_emails
   - [ ] addresses.phone_numbers
   
   Nested blocks (3-level):
   - [ ] routing_utilization
   - [ ] routing_utilization.call
   - [ ] routing_utilization.callback
   - [ ] routing_utilization.message
   - [ ] routing_utilization.email
   - [ ] routing_utilization.chat
   - [ ] routing_utilization.label_utilizations
   ```

**Deliverable**: Cross-verification completed, no missing elements

---

### Task 8.4: Code Review and Approval

**Objective**: Get peer review and approval before proceeding to Stage 2.

**Steps**:

1. **Prepare for review**
   - Ensure all checklist items are complete
   - Document any deviations from standard pattern
   - Prepare summary of changes

2. **Create pull request or review request**
   - Include link to Stage 1 requirements and design docs
   - Highlight complex patterns used (3-level nesting, custom plan modifiers, etc.)
   - Note any TODO comments and their rationale

3. **Address review comments**
   - Make requested changes
   - Re-verify checklist
   - Re-compile and test

4. **Get approval**
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
- Verify custom packages exist (validators, phoneplan)

---

### Issue 2: Type Mismatch Errors

**Problem**: Type mismatch between schema and attribute types.

**Solution**:
- Verify using correct attribute type (e.g., `schema.StringAttribute` not `datasourceschema.StringAttribute`)
- Check that plan modifiers match attribute type (e.g., `stringplanmodifier` for string attributes)
- Verify element types match schema definitions exactly

---

### Issue 3: Missing Plan Modifiers

**Problem**: Computed attributes show "known after apply" unnecessarily.

**Solution**:
- Add `UseStateForUnknown()` plan modifier to computed attributes
- Add `UseStateForUnknown()` to nested blocks
- Verify plan modifier import is present

---

### Issue 4: Nested Blocks Not Compiling

**Problem**: Nested blocks within blocks cause compilation errors.

**Solution**:
- Use `Blocks: map[string]schema.Block{}` for nested blocks, not `Attributes`
- Verify correct nesting structure (Blocks → Blocks → Attributes)
- Check that all block types are correct (SetNestedBlock vs ListNestedBlock)

---

### Issue 5: Element Type Mismatch

**Problem**: Runtime errors about type mismatches in Stage 2.

**Solution**:
- Verify element type helper functions match schema exactly
- Check attribute names match exactly (case-sensitive)
- Check attribute types match exactly (Int64 vs Float64, etc.)
- Re-verify Task 6.3

---

### Issue 6: Validator Not Found

**Problem**: Custom validator not found or undefined.

**Solution**:
- Verify custom validator exists in `genesyscloud/validators` package
- Check import statement
- If validator doesn't exist, note for creation (may be out of scope)

---

### Issue 7: Exporter RefAttrs Not Working

**Problem**: Exporter doesn't resolve dependencies correctly.

**Solution**:
- Verify attribute name matches schema exactly
- Use dot notation for nested attributes
- Verify resource type matches Terraform resource type
- Check that referenced resource exists

---

### Issue 8: Default Value Compilation Error

**Problem**: Error when using Default without Computed.

**Solution**:
- Add `Computed: true` to all attributes with `Default`
- Framework requires both Optional and Computed when using Default
- This is different from SDKv2

---

## Completion Criteria

Stage 1 is complete when:

- [ ] All tasks in this document are completed
- [ ] All checklist items are verified
- [ ] Code compiles without errors
- [ ] Cross-verification with SDKv2 schema is complete
- [ ] Element type helpers match schema exactly
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
   - Implement flatten/build functions using element type helpers

3. **Reference Stage 2 documentation**
   - Read Stage 2 `requirements.md`
   - Read Stage 2 `design.md`
   - Follow Stage 2 `tasks.md`

---

## Time Estimates

| Phase | Estimated Time |
|-------|----------------|
| Phase 1: Preparation and Analysis | 90-120 minutes |
| Phase 2: File Setup | 15-30 minutes |
| Phase 3: Schema Implementation - Attributes | 60-90 minutes |
| Phase 4: Schema Implementation - Nested Blocks | 120-180 minutes |
| Phase 5: Data Source and Helper Functions | 30-45 minutes |
| Phase 6: Element Type Helpers | 45-60 minutes |
| Phase 7: Registration and Export | 60-90 minutes |
| Phase 8: Validation and Review | 60-90 minutes |
| **Total** | **6-10 hours** |

*Note: Times vary significantly based on:*
- *Number of nested blocks*
- *Nesting depth (1-level vs 3-level)*
- *Number of custom validators/plan modifiers*
- *Familiarity with patterns*

**Complexity Factors**:
- Simple resource (no nested blocks): 3-5 hours
- Medium resource (1-2 nested blocks): 4-6 hours
- Complex resource (3+ nested blocks, 2-3 level nesting): 6-10 hours
- Very complex resource (many nested blocks, 3-level nesting, custom modifiers): 8-12 hours

---

## References

- **Reference Implementation**: `genesyscloud/user/resource_genesyscloud_user_schema.go`
- **Stage 1 Requirements**: `prompts/pf_complex_resource_migration/Stage1/requirements.md`
- **Stage 1 Design**: `prompts/pf_complex_resource_migration/Stage1/design.md`
- **Plugin Framework Documentation**: https://developer.hashicorp.com/terraform/plugin/framework
- **Nested Attributes Guide**: https://developer.hashicorp.com/terraform/plugin/framework/handling-data/blocks
- **Plan Modifiers Guide**: https://developer.hashicorp.com/terraform/plugin/framework/resources/plan-modification

