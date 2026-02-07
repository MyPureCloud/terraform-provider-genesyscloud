# Stage 4 – Export Functionality Requirements

## Overview

Stage 4 focuses on implementing export functionality for the migrated Plugin Framework resource. This stage creates a separate export utilities file that converts SDK resource objects to flat attribute maps for the legacy exporter's dependency resolution logic. This is a Phase 1 temporary implementation that will be removed once all resources are migrated to Plugin Framework (Phase 2).

**Reference Implementation**: 
- `genesyscloud/routing_wrapupcode/resource_genesyscloud_routing_wrapupcode_export_utils.go`

---

## Objectives

### Primary Goal
Create export utilities that enable the legacy exporter to work with Plugin Framework resources by providing flat attribute maps for dependency resolution.

### Specific Objectives
1. Create export utilities file with attribute mapping function
2. Implement flat attribute map conversion from SDK objects
3. Ensure all dependency references are included in attribute map
4. Document Phase 1 temporary nature with TODO comments
5. Maintain compatibility with existing exporter behavior
6. Preserve export structure and dependency resolution

---

## Scope

### In Scope for Stage 4

#### 1. Export Utilities File
- Create `resource_genesyscloud_<resource_name>_export_utils.go` file
- Implement `build<ResourceName>Attributes()` function
- Convert SDK resource objects to flat attribute maps
- Include all resource attributes in map
- Include all dependency references in map

#### 2. Attribute Mapping
- Map all schema attributes to flat string map
- Handle optional attributes (include if present)
- Handle nested objects (extract IDs)
- Handle dependency references (critical for exporter)
- Use consistent naming with schema

#### 3. Documentation
- Add file-level comment explaining Phase 1 temporary nature
- Add function-level comment explaining purpose
- Add TODO comment for Phase 2 removal
- Document attribute map format
- Document dependency references

### Out of Scope for Stage 4

❌ **Exporter Core Changes**
- No changes to exporter core logic
- No changes to dependency resolution algorithm
- Exporter remains unchanged

❌ **Schema Modifications**
- No changes to schema file from Stage 1
- Schema is already complete

❌ **Resource Implementation Changes**
- No changes to resource implementation from Stage 2
- Implementation is already complete

❌ **Test Changes**
- No changes to test files from Stage 3
- Tests are already complete

❌ **Proxy Modifications**
- No changes to proxy files
- Proxy files remain unchanged

---

## Success Criteria

### Functional Requirements

#### FR1: Export Utilities File
- ✅ File created: `resource_genesyscloud_<resource_name>_export_utils.go`
- ✅ Package declaration matches directory name
- ✅ All required imports are present
- ✅ File includes Phase 1 temporary documentation

#### FR2: Attribute Mapping Function
- ✅ `build<ResourceName>Attributes()` function implemented
- ✅ Function accepts SDK resource object pointer
- ✅ Function returns `map[string]string`
- ✅ Function includes comprehensive documentation

#### FR3: Attribute Map Completeness
- ✅ All schema attributes are included in map
- ✅ ID attribute is included
- ✅ Name attribute is included
- ✅ All optional attributes are included (if present)
- ✅ All dependency references are included

#### FR4: Dependency References
- ✅ All dependency attributes are included in map
- ✅ Dependency IDs are extracted from nested objects
- ✅ Attribute names match schema exactly
- ✅ Exporter can resolve dependencies correctly

#### FR5: Documentation
- ✅ File-level comment explains Phase 1 temporary nature
- ✅ Function comment explains purpose and parameters
- ✅ TODO comment marks code for Phase 2 removal
- ✅ Attribute map format is documented
- ✅ Dependency references are highlighted

#### FR6: Integration
- ✅ Function is called from `GetAll<ResourceName>SDK()` in Stage 2
- ✅ Attribute map is added to export metadata
- ✅ Exporter can successfully export resources
- ✅ Dependency resolution works correctly

### Non-Functional Requirements

#### NFR1: Code Quality
- ✅ Code follows Go best practices
- ✅ Code follows existing codebase conventions
- ✅ Proper nil checks for optional attributes
- ✅ Clear and consistent naming

#### NFR2: Documentation Quality
- ✅ Comments are clear and comprehensive
- ✅ Phase 1/Phase 2 distinction is clear
- ✅ Purpose and rationale are explained
- ✅ Migration path is documented

#### NFR3: Maintainability
- ✅ Code is easy to understand
- ✅ Code is easy to remove in Phase 2
- ✅ No complex logic or dependencies
- ✅ Self-contained in single file

---

## Dependencies and Prerequisites

### Prerequisites

#### 1. Stage 1, 2, and 3 Completion
- Schema file must be complete (Stage 1)
- Resource implementation must be complete (Stage 2)
- `GetAll<ResourceName>SDK()` function must be implemented (Stage 2)
- Tests must be complete (Stage 3)

#### 2. Understanding of Export Mechanism
- Familiarity with exporter architecture
- Understanding of dependency resolution
- Knowledge of flat attribute map format
- Understanding of Phase 1/Phase 2 migration strategy

#### 3. Reference Implementation
- Study `routing_wrapupcode` export utilities file
- Understand attribute mapping patterns
- Review dependency reference handling

### Dependencies

#### 1. Package Imports
```go
import (
    "github.com/mypurecloud/platform-client-sdk-go/v176/platformclientv2"
)
```

#### 2. SDK Resource Object
- Function receives SDK resource object from `GetAll<ResourceName>SDK()`
- Object contains all resource attributes
- Object includes nested dependency objects

#### 3. Exporter Integration
- Function is called from `GetAll<ResourceName>SDK()` in Stage 2
- Attribute map is added to `ResourceMeta.ExportAttributes`
- Exporter uses attribute map for dependency resolution

---

## Constraints

### Technical Constraints

#### TC1: Flat Attribute Map Format
- **Constraint**: Must use `map[string]string` format
- **Rationale**: Matches SDKv2 InstanceState format used by exporter
- **Impact**: All values must be converted to strings

#### TC2: Attribute Naming
- **Constraint**: Attribute names must match schema exactly
- **Rationale**: Exporter uses attribute names to resolve dependencies
- **Impact**: Any mismatch breaks dependency resolution

#### TC3: Dependency References
- **Constraint**: All dependency references must be included
- **Rationale**: Exporter needs dependency IDs for ordering and HCL generation
- **Impact**: Missing dependencies break export

#### TC4: Phase 1 Temporary Code
- **Constraint**: Code is temporary and will be removed in Phase 2
- **Rationale**: Exporter will be updated to work natively with Framework types
- **Impact**: Code must be clearly marked as temporary

### Process Constraints

#### PC1: No Exporter Changes
- **Constraint**: Cannot modify exporter core logic
- **Rationale**: Exporter must work with both SDKv2 and Framework resources
- **Impact**: Must adapt to existing exporter interface

#### PC2: Backward Compatibility
- **Constraint**: Export behavior must match SDKv2 version
- **Rationale**: Users expect consistent export output
- **Impact**: Attribute map must match SDKv2 format exactly

---

## Validation Checklist

Use this checklist to verify Stage 4 completion:

### Export Utilities File
- [ ] File created: `resource_genesyscloud_<resource_name>_export_utils.go`
- [ ] Package declaration matches directory name
- [ ] All required imports are present
- [ ] No unused imports

### File Documentation
- [ ] File-level comment explains Phase 1 temporary nature
- [ ] File-level comment includes TODO for Phase 2 removal
- [ ] File-level comment explains purpose
- [ ] File path is included in comment

### Attribute Mapping Function
- [ ] `build<ResourceName>Attributes()` function implemented
- [ ] Function signature is correct (accepts pointer, returns map)
- [ ] Function includes comprehensive documentation
- [ ] Function comment explains parameters and return value
- [ ] Function comment documents attribute map format

### Attribute Map Implementation
- [ ] ID attribute is included
- [ ] Name attribute is included
- [ ] All optional attributes are included (with nil checks)
- [ ] All dependency references are included
- [ ] Nested objects are handled correctly (IDs extracted)
- [ ] Attribute names match schema exactly

### Dependency References
- [ ] All dependency attributes are marked as CRITICAL in comments
- [ ] Dependency IDs are extracted from nested objects
- [ ] Dependency attribute names match RefAttrs from Stage 1
- [ ] Comment explains dependency usage by exporter

### Integration
- [ ] Function is called from `GetAll<ResourceName>SDK()` in Stage 2
- [ ] Attribute map is added to `ResourceMeta.ExportAttributes`
- [ ] Export functionality works correctly
- [ ] Dependency resolution works correctly

### Code Quality
- [ ] Code compiles without errors
- [ ] Code follows Go conventions
- [ ] Nil checks are present for optional attributes
- [ ] No complex logic or dependencies
- [ ] Code is self-contained in single file

---

## Example: routing_wrapupcode Export Utilities

### File Structure
```
genesyscloud/routing_wrapupcode/
├── resource_genesyscloud_routing_wrapupcode_schema.go       (Stage 1)
├── resource_genesyscloud_routing_wrapupcode.go              (Stage 2)
├── data_source_genesyscloud_routing_wrapupcode.go           (Stage 2)
├── resource_genesyscloud_routing_wrapupcode_test.go         (Stage 3)
├── data_source_genesyscloud_routing_wrapupcode_test.go      (Stage 3)
├── genesyscloud_routing_wrapupcode_init_test.go             (Stage 3)
└── resource_genesyscloud_routing_wrapupcode_export_utils.go (Stage 4 - THIS)
```

### File-Level Documentation
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

package routing_wrapupcode
```

### Attribute Mapping Function
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
func buildWrapupcodeAttributes(wrapupcode *platformclientv2.Wrapupcode) map[string]string {
    attributes := make(map[string]string)

    // Basic attributes
    if wrapupcode.Id != nil {
        attributes["id"] = *wrapupcode.Id
    }
    if wrapupcode.Name != nil {
        attributes["name"] = *wrapupcode.Name
    }
    if wrapupcode.Description != nil {
        attributes["description"] = *wrapupcode.Description
    }

    // ⭐ CRITICAL: Dependency reference (used by exporter for dependency resolution)
    if wrapupcode.Division != nil && wrapupcode.Division.Id != nil {
        attributes["division_id"] = *wrapupcode.Division.Id
    }

    return attributes
}
```

### Key Elements

| Element | Purpose |
|---------|---------|
| File-level comment | Explains Phase 1 temporary nature and removal plan |
| TODO comment | Marks code for Phase 2 removal |
| Function documentation | Explains purpose, parameters, return value, format |
| Attribute map format | Documents expected structure |
| CRITICAL comment | Highlights dependency references |
| Nil checks | Handles optional attributes safely |

---

## Phase 1 vs Phase 2

### Phase 1 (Current - Temporary)

**Approach**: Convert SDK objects to flat attribute maps

**Files**:
- `resource_genesyscloud_<resource_name>_export_utils.go` (temporary)
- `GetAll<ResourceName>SDK()` in resource file (temporary)

**Exporter**:
- Uses SDK diagnostics
- Uses flat attribute maps for dependency resolution
- Works with both SDKv2 and Framework resources

**Pros**:
- Enables gradual migration
- No exporter changes needed
- Works with existing infrastructure

**Cons**:
- Temporary code that must be maintained
- Duplication of attribute mapping logic
- Extra file per resource

### Phase 2 (Future - Permanent)

**Approach**: Exporter works natively with Framework types

**Files**:
- Export utilities file removed
- `GetAll<ResourceName>SDK()` removed
- Only `GetAll<ResourceName>()` remains

**Exporter**:
- Uses Framework diagnostics
- Works directly with Framework types
- No flat attribute map conversion needed

**Pros**:
- Cleaner, more maintainable code
- No temporary scaffolding
- Native Framework integration

**Cons**:
- Requires exporter refactoring
- Can only happen after all resources migrated

---

## Migration Path

### Current State (Phase 1)
```
Resource (Framework) → GetAll SDK → build Attributes → Flat Map → Exporter
```

### Future State (Phase 2)
```
Resource (Framework) → GetAll → Framework Types → Exporter
```

### Transition Plan
1. **Phase 1**: All resources migrated with export utilities (current)
2. **Exporter Update**: Refactor exporter to work with Framework types
3. **Phase 2**: Remove all export utilities files
4. **Cleanup**: Remove all `GetAll<ResourceName>SDK()` functions

---

## Next Steps

After Stage 4 completion and approval:
1. Test export functionality
2. Verify dependency resolution works
3. Confirm export output matches SDKv2 version
4. Complete migration for this resource
5. Begin migration of next resource

---

## References

- **Reference Implementation**: `genesyscloud/routing_wrapupcode/resource_genesyscloud_routing_wrapupcode_export_utils.go`
- **Stage 2 GetAll Functions**: `genesyscloud/routing_wrapupcode/resource_genesyscloud_routing_wrapupcode.go`
- **Exporter Documentation**: Internal exporter architecture docs
