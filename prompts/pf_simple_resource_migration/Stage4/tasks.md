# Stage 4 – Export Functionality Tasks

## Overview

This document provides step-by-step tasks for completing Stage 4 of the Plugin Framework migration. Follow these tasks in order to implement export functionality for Plugin Framework resources by creating temporary export utilities.

**Reference Implementation**: 
- `genesyscloud/routing_wrapupcode/resource_genesyscloud_routing_wrapupcode_export_utils.go`

**Estimated Time**: 1-2 hours (simple resources)

---

## Prerequisites

Before starting Stage 4 tasks, ensure:

- [ ] Stage 1 (Schema Migration) is complete and approved
- [ ] Stage 2 (Resource Migration) is complete and approved
- [ ] Stage 3 (Test Migration) is complete and approved
- [ ] `GetAll<ResourceName>SDK()` function is implemented in Stage 2
- [ ] You understand the exporter architecture
- [ ] You have read Stage 4 `requirements.md` and `design.md`
- [ ] You have studied the `routing_wrapupcode` reference implementation

---

## Task Checklist

### Phase 1: File Creation and Setup
- [ ] Task 1.1: Create Export Utilities File
- [ ] Task 1.2: Add File-Level Documentation
- [ ] Task 1.3: Add Package Declaration and Imports

### Phase 2: Attribute Mapping Implementation
- [ ] Task 2.1: Implement Attribute Mapping Function Signature
- [ ] Task 2.2: Add Function Documentation
- [ ] Task 2.3: Implement Basic Attributes Mapping
- [ ] Task 2.4: Implement Optional Attributes Mapping
- [ ] Task 2.5: Implement Dependency References Mapping

### Phase 3: Integration and Testing
- [ ] Task 3.1: Verify Integration with Stage 2
- [ ] Task 3.2: Test Export Functionality
- [ ] Task 3.3: Verify Dependency Resolution

### Phase 4: Validation and Review
- [ ] Task 4.1: Review Against Checklist
- [ ] Task 4.2: Code Review and Approval

---

## Detailed Tasks

## Phase 1: File Creation and Setup

### Task 1.1: Create Export Utilities File

**Objective**: Create the export utilities file.

**Steps**:

1. **Navigate to resource package directory**
   ```powershell
   cd genesyscloud\<resource_name>
   ```
   Example:
   ```powershell
   cd genesyscloud\routing_wrapupcode
   ```

2. **Create the export utilities file**
   ```powershell
   New-Item -ItemType File -Name "resource_genesyscloud_<resource_name>_export_utils.go"
   ```
   Example:
   ```powershell
   New-Item -ItemType File -Name "resource_genesyscloud_routing_wrapupcode_export_utils.go"
   ```

**Deliverable**: Empty export utilities file created

---

### Task 1.2: Add File-Level Documentation

**Objective**: Add comprehensive file-level documentation explaining Phase 1 temporary nature.

**Steps**:

1. **Add file-level comment block**
   ```go
   // Package <resource_name> contains temporary export utilities for Plugin Framework <resource> resource.
   //
   // IMPORTANT: This file contains migration scaffolding that converts SDK types to flat
   // attribute maps for the legacy exporter's dependency resolution logic.
   //
   // TODO: Remove this entire file once all resources are migrated to Plugin Framework
   // and the exporter is updated to work natively with Framework types (Phase 2).
   // This is Phase 1 temporary code - resource-specific implementation.
   //
   // File: genesyscloud/<resource_name>/resource_genesyscloud_<resource_name>_export_utils.go
   ```

2. **Customize for your resource**
   - Replace `<resource_name>` with your resource name
   - Replace `<resource>` with human-readable resource name
   - Verify file path is correct

**Example** (routing_wrapupcode):
```go
// Package routing_wrapupcode contains temporary export utilities for Plugin Framework routing wrapupcode resource.
//
// IMPORTANT: This file contains migration scaffolding that converts SDK types to flat
// attribute maps for the legacy exporter's dependency resolution logic.
//
// TODO: Remove this entire file once all resources are migrated to Plugin Framework
// and the exporter is updated to work natively with Framework types (Phase 2).
// This is Phase 1 temporary code - resource-specific implementation.
//
// File: genesyscloud/routing_wrapupcode/resource_genesyscloud_routing_wrapupcode_export_utils.go
```

**Deliverable**: File-level documentation added

---

### Task 1.3: Add Package Declaration and Imports

**Objective**: Add package declaration and minimal imports.

**Steps**:

1. **Add package declaration**
   ```go
   package <resource_name>
   ```

2. **Add SDK import**
   ```go
   import (
       "github.com/mypurecloud/platform-client-sdk-go/v176/platformclientv2"
   )
   ```

3. **Verify no other imports needed**
   - Export utilities should only need SDK import
   - No Framework imports
   - No other dependencies

**Deliverable**: Package declaration and imports added

---

## Phase 2: Attribute Mapping Implementation

### Task 2.1: Implement Attribute Mapping Function Signature

**Objective**: Create the attribute mapping function signature.

**Steps**:

1. **Add function signature**
   ```go
   func build<ResourceName>Attributes(<resource> *platformclientv2.<ResourceType>) map[string]string {
       attributes := make(map[string]string)
       
       // Attributes will be added here
       
       return attributes
   }
   ```

2. **Replace placeholders**
   - `<ResourceName>` → Your resource name in PascalCase (e.g., `RoutingWrapupcode`)
   - `<resource>` → Your resource variable name in camelCase (e.g., `wrapupcode`)
   - `<ResourceType>` → SDK resource type (e.g., `Wrapupcode`)

**Example** (routing_wrapupcode):
```go
func buildWrapupcodeAttributes(wrapupcode *platformclientv2.Wrapupcode) map[string]string {
    attributes := make(map[string]string)
    
    // Attributes will be added here
    
    return attributes
}
```

**Deliverable**: Function signature implemented

---

### Task 2.2: Add Function Documentation

**Objective**: Add comprehensive function documentation.

**Steps**:

1. **Add function documentation above signature**
   ```go
   // build<ResourceName>Attributes creates a flat attribute map from SDK <resource> object for export.
   // This function converts the SDK <resource> object to a flat map matching SDKv2 InstanceState format.
   //
   // Parameters:
   //   - <resource>: <Resource> object from API
   //
   // Returns:
   //   - map[string]string: Flat attribute map with all <resource> attributes
   //
   // Attribute Map Format (matching SDKv2 InstanceState):
   //   - "id" = <resource> ID
   //   - "name" = <resource> name
   //   - "attribute1" = attribute value
   //   - "dependency_id" = dependency ID (dependency reference)
   //
   // Note: [Add resource-specific notes about complexity, nested attributes, etc.]
   ```

2. **Customize documentation**
   - List all attributes in format section
   - Mark dependency references
   - Add resource-specific notes if needed

**Example** (routing_wrapupcode):
```go
// buildWrapupcodeAttributes creates a flat attribute map from SDK wrapupcode object for export.
// This function converts the SDK wrapupcode object to a flat map matching SDKv2 InstanceState format.
//
// Parameters:
//   - wrapupcode: Wrapupcode object from API
//
// Returns:
//   - map[string]string: Flat attribute map with all wrapupcode attributes
//
// Attribute Map Format (matching SDKv2 InstanceState):
//   - "id" = wrapupcode ID
//   - "name" = wrapupcode name
//   - "division_id" = division ID (dependency reference)
//   - "description" = wrapupcode description
//
// Note: Unlike user resource, wrapupcode is simple with no nested attributes or additional API calls.
```

**Deliverable**: Function documentation added

---

### Task 2.3: Implement Basic Attributes Mapping

**Objective**: Map basic required attributes (ID, name).

**Steps**:

1. **Add ID attribute mapping**
   ```go
   // Basic attributes
   if <resource>.Id != nil {
       attributes["id"] = *<resource>.Id
   }
   ```

2. **Add name attribute mapping**
   ```go
   if <resource>.Name != nil {
       attributes["name"] = *<resource>.Name
   }
   ```

3. **Verify attribute names match schema**
   - Check Stage 1 schema file
   - Use exact attribute names
   - Use snake_case (lowercase with underscores)

**Example** (routing_wrapupcode):
```go
// Basic attributes
if wrapupcode.Id != nil {
    attributes["id"] = *wrapupcode.Id
}
if wrapupcode.Name != nil {
    attributes["name"] = *wrapupcode.Name
}
```

**Deliverable**: Basic attributes mapped

---

### Task 2.4: Implement Optional Attributes Mapping

**Objective**: Map optional attributes with nil checks.

**Steps**:

1. **Review schema for optional attributes**
   - Check Stage 1 schema file
   - Identify all optional attributes
   - Note attribute types (string, int, bool)

2. **Add string attribute mapping**
   ```go
   if <resource>.OptionalStringAttr != nil {
       attributes["optional_string_attr"] = *<resource>.OptionalStringAttr
   }
   ```

3. **Add integer attribute mapping** (if applicable)
   ```go
   if <resource>.OptionalIntAttr != nil {
       attributes["optional_int_attr"] = fmt.Sprintf("%d", *<resource>.OptionalIntAttr)
   }
   ```
   Note: Add `"fmt"` to imports if using `fmt.Sprintf`

4. **Add boolean attribute mapping** (if applicable)
   ```go
   if <resource>.OptionalBoolAttr != nil {
       if *<resource>.OptionalBoolAttr {
           attributes["optional_bool_attr"] = "true"
       } else {
           attributes["optional_bool_attr"] = "false"
       }
   }
   ```

**Example** (routing_wrapupcode):
```go
if wrapupcode.Description != nil {
    attributes["description"] = *wrapupcode.Description
}
```

**Key Points**:
- Always check nil before dereferencing
- Convert non-string types to strings
- Use exact attribute names from schema

**Deliverable**: Optional attributes mapped

---

### Task 2.5: Implement Dependency References Mapping

**Objective**: Map dependency references with CRITICAL markers.

**Steps**:

1. **Review RefAttrs from Stage 1 schema**
   - Open `resource_genesyscloud_<resource_name>_schema.go`
   - Find `<ResourceName>Exporter()` function
   - Review `RefAttrs` map
   - Note all dependency attribute names

2. **Add dependency reference mapping with CRITICAL comment**
   ```go
   // ⭐ CRITICAL: Dependency reference (used by exporter for dependency resolution)
   if <resource>.Dependency != nil && <resource>.Dependency.Id != nil {
       attributes["dependency_id"] = *<resource>.Dependency.Id
   }
   ```

3. **Add all dependency references**
   - One mapping per dependency
   - Check both parent and child for nil
   - Extract ID from nested object
   - Mark each with CRITICAL comment

**Example** (routing_wrapupcode):
```go
// ⭐ CRITICAL: Dependency reference (used by exporter for dependency resolution)
if wrapupcode.Division != nil && wrapupcode.Division.Id != nil {
    attributes["division_id"] = *wrapupcode.Division.Id
}
```

**Example with multiple dependencies**:
```go
// ⭐ CRITICAL: Dependency references (used by exporter for dependency resolution)
if resource.Division != nil && resource.Division.Id != nil {
    attributes["division_id"] = *resource.Division.Id
}
if resource.Queue != nil && resource.Queue.Id != nil {
    attributes["queue_id"] = *resource.Queue.Id
}
if resource.Flow != nil && resource.Flow.Id != nil {
    attributes["flow_id"] = *resource.Flow.Id
}
```

**Key Points**:
- CRITICAL comment is mandatory
- Check both parent and child for nil
- Attribute names must match RefAttrs exactly
- Extract ID from nested object

**Deliverable**: Dependency references mapped

---

## Phase 3: Integration and Testing

### Task 3.1: Verify Integration with Stage 2

**Objective**: Verify the function is called from `GetAll<ResourceName>SDK()`.

**Steps**:

1. **Open Stage 2 resource file**
   - File: `resource_genesyscloud_<resource_name>.go`
   - Find `GetAll<ResourceName>SDK()` function

2. **Verify function is called**
   - Look for loop over resources
   - Find call to `build<ResourceName>Attributes(&resource)`
   - Verify attribute map is added to `ResourceMeta.ExportAttributes`

3. **Expected pattern in Stage 2**:
   ```go
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
   ```

4. **If not present, add integration**
   - This should have been done in Stage 2
   - If missing, add the integration code
   - Follow the pattern above

**Deliverable**: Integration verified or added

---

### Task 3.2: Test Export Functionality

**Objective**: Test that export works correctly.

**Steps**:

1. **Compile the code**
   ```powershell
   go build ./genesyscloud/<resource_name>
   ```

2. **Fix any compilation errors**
   - Missing imports
   - Syntax errors
   - Type mismatches

3. **Run export command** (if export tool available)
   ```powershell
   terraform export genesyscloud_<resource>
   ```

4. **Verify export output**
   - Check that resources are exported
   - Verify HCL is generated
   - Check for any errors

5. **If export tool not available**
   - Verify code compiles
   - Review code manually
   - Proceed to code review

**Deliverable**: Export functionality tested

---

### Task 3.3: Verify Dependency Resolution

**Objective**: Verify that dependency references are resolved correctly.

**Steps**:

1. **Review exported HCL** (if export tool available)
   - Find resources with dependencies
   - Verify dependency references are correct
   - Example: `division_id = genesyscloud_auth_division.division_label.id`

2. **Check dependency ordering**
   - Dependencies should be exported before dependent resources
   - Example: Division exported before wrapupcode

3. **Verify attribute names**
   - Dependency attribute names match schema
   - Dependency attribute names match RefAttrs from Stage 1

4. **If issues found**
   - Check attribute names in `build<ResourceName>Attributes()`
   - Verify RefAttrs in Stage 1 schema
   - Ensure dependency IDs are extracted correctly

**Deliverable**: Dependency resolution verified

---

## Phase 4: Validation and Review

### Task 4.1: Review Against Checklist

**Objective**: Verify all requirements are met.

**Steps**:

1. **Use the validation checklist from requirements.md**

   **Export Utilities File**:
   - [ ] File created: `resource_genesyscloud_<resource_name>_export_utils.go`
   - [ ] Package declaration matches directory name
   - [ ] All required imports are present
   - [ ] No unused imports

   **File Documentation**:
   - [ ] File-level comment explains Phase 1 temporary nature
   - [ ] File-level comment includes TODO for Phase 2 removal
   - [ ] File-level comment explains purpose
   - [ ] File path is included in comment

   **Attribute Mapping Function**:
   - [ ] `build<ResourceName>Attributes()` function implemented
   - [ ] Function signature is correct (accepts pointer, returns map)
   - [ ] Function includes comprehensive documentation
   - [ ] Function comment explains parameters and return value
   - [ ] Function comment documents attribute map format

   **Attribute Map Implementation**:
   - [ ] ID attribute is included
   - [ ] Name attribute is included
   - [ ] All optional attributes are included (with nil checks)
   - [ ] All dependency references are included
   - [ ] Nested objects are handled correctly (IDs extracted)
   - [ ] Attribute names match schema exactly

   **Dependency References**:
   - [ ] All dependency attributes are marked as CRITICAL in comments
   - [ ] Dependency IDs are extracted from nested objects
   - [ ] Dependency attribute names match RefAttrs from Stage 1
   - [ ] Comment explains dependency usage by exporter

   **Integration**:
   - [ ] Function is called from `GetAll<ResourceName>SDK()` in Stage 2
   - [ ] Attribute map is added to `ResourceMeta.ExportAttributes`
   - [ ] Export functionality works correctly
   - [ ] Dependency resolution works correctly

   **Code Quality**:
   - [ ] Code compiles without errors
   - [ ] Code follows Go conventions
   - [ ] Nil checks are present for optional attributes
   - [ ] No complex logic or dependencies
   - [ ] Code is self-contained in single file

2. **Fix any issues found**

**Deliverable**: All checklist items verified

---

### Task 4.2: Code Review and Approval

**Objective**: Get peer review and approval.

**Steps**:

1. **Create pull request or review request**
   - Include link to Stage 4 requirements and design docs
   - Highlight Phase 1 temporary nature
   - Note dependency references

2. **Address review comments**
   - Make requested changes
   - Re-verify checklist
   - Re-test export functionality

3. **Get approval**
   - Obtain approval from reviewer
   - Merge or mark as complete

**Deliverable**: Stage 4 approved and migration complete

---

## Common Issues and Solutions

### Issue 1: Missing Dependency Reference

**Problem**: Dependency not resolved in export.

**Solution**:
- Verify dependency attribute is in `build<ResourceName>Attributes()`
- Check attribute name matches RefAttrs from Stage 1
- Ensure dependency ID is extracted from nested object
- Verify both parent and child nil checks

### Issue 2: Attribute Name Mismatch

**Problem**: Export fails or dependency not resolved.

**Solution**:
- Check attribute names match schema exactly
- Use snake_case (lowercase with underscores)
- Verify RefAttrs in Stage 1 schema
- Compare with schema attribute names

### Issue 3: Nil Pointer Panic

**Problem**: Panic when accessing optional attribute.

**Solution**:
- Add nil check before dereferencing
- Check both parent and child for nested objects
- Example: `if resource.Nested != nil && resource.Nested.Id != nil`

### Issue 4: Type Conversion Error

**Problem**: Non-string attribute not converted correctly.

**Solution**:
- Use `fmt.Sprintf("%d", *intAttr)` for integers
- Use `"true"` or `"false"` strings for booleans
- Add `"fmt"` to imports if needed

### Issue 5: Function Not Called

**Problem**: Export doesn't include attributes.

**Solution**:
- Verify function is called from `GetAll<ResourceName>SDK()`
- Check attribute map is added to `ResourceMeta.ExportAttributes`
- Review Stage 2 integration code

---

## Completion Criteria

Stage 4 is complete when:

- [ ] All tasks in this document are completed
- [ ] All checklist items are verified
- [ ] Code compiles without errors
- [ ] Export functionality works correctly
- [ ] Dependency resolution works correctly
- [ ] Code review is approved
- [ ] **Migration is complete for this resource!** 🎉

---

## Next Steps

After Stage 4 completion:

1. **Celebrate!** 🎉
   - You've successfully migrated a resource to Plugin Framework
   - All 4 stages are complete

2. **Document lessons learned**
   - Note any challenges encountered
   - Document any deviations from standard pattern
   - Share knowledge with team

3. **Begin next resource migration**
   - Select next resource to migrate
   - Follow same 4-stage process
   - Use this resource as reference

4. **Track migration progress**
   - Update migration tracking document
   - Note completed resources
   - Estimate remaining work

---

## Time Estimates

| Phase | Estimated Time |
|-------|----------------|
| Phase 1: File Creation and Setup | 15-30 minutes |
| Phase 2: Attribute Mapping Implementation | 30-60 minutes |
| Phase 3: Integration and Testing | 15-30 minutes |
| Phase 4: Validation and Review | 15-30 minutes |
| **Total** | **1-2.5 hours** |

*Note: Times vary based on resource complexity (number of attributes, dependencies).*

---

## Complete Migration Summary

### All Stages Complete! 🎉

You have successfully completed all 4 stages of the Plugin Framework migration:

- ✅ **Stage 1**: Schema Migration (schema definitions)
- ✅ **Stage 2**: Resource Migration (CRUD operations)
- ✅ **Stage 3**: Test Migration (acceptance tests)
- ✅ **Stage 4**: Export Functionality (export utilities)

### Files Created

```
genesyscloud/<resource_name>/
├── resource_genesyscloud_<resource_name>_schema.go          ✅ Stage 1
├── resource_genesyscloud_<resource_name>.go                 ✅ Stage 2
├── data_source_genesyscloud_<resource_name>.go              ✅ Stage 2
├── resource_genesyscloud_<resource_name>_test.go            ✅ Stage 3
├── data_source_genesyscloud_<resource_name>_test.go         ✅ Stage 3
├── genesyscloud_<resource_name>_init_test.go                ✅ Stage 3
├── resource_genesyscloud_<resource_name>_export_utils.go    ✅ Stage 4
└── genesyscloud_<resource_name>_proxy.go                    (unchanged)
```

### What's Next?

1. **Use as reference** for migrating other resources
2. **Share knowledge** with team members
3. **Track progress** on overall migration
4. **Plan for Phase 2** (exporter update and cleanup)

---

## References

- **Reference Implementation**: `genesyscloud/routing_wrapupcode/resource_genesyscloud_routing_wrapupcode_export_utils.go`
- **Stage 4 Requirements**: `prompts/pf_simple_resource_migration/Stage4/requirements.md`
- **Stage 4 Design**: `prompts/pf_simple_resource_migration/Stage4/design.md`
- **Stage 2 GetAll Functions**: `genesyscloud/routing_wrapupcode/resource_genesyscloud_routing_wrapupcode.go`
- **Stage 1 Exporter Config**: `genesyscloud/routing_wrapupcode/resource_genesyscloud_routing_wrapupcode_schema.go`
