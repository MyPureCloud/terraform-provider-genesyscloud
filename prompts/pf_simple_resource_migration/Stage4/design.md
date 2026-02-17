# Stage 4 – Export Functionality Design

## Overview

This document describes the design decisions, architecture, and patterns used in Stage 4 of the Plugin Framework migration. Stage 4 focuses on implementing export functionality for Plugin Framework resources by creating temporary export utilities that convert SDK objects to flat attribute maps for the legacy exporter's dependency resolution logic.

**Reference Implementation**: 
- `genesyscloud/routing_wrapupcode/resource_genesyscloud_routing_wrapupcode_export_utils.go`

---

## Design Principles

### 1. Phase 1 Temporary Scaffolding
**Principle**: Create temporary export utilities that will be removed in Phase 2.

**Rationale**:
- Exporter currently uses SDK diagnostics and flat attribute maps
- Enables gradual migration without breaking existing export functionality
- Allows Framework resources to work with legacy exporter
- Clear migration path to Phase 2 native Framework support

**Implementation**:
- Separate export utilities file per resource
- Clearly marked as Phase 1 temporary with TODO comments
- Self-contained and easy to remove
- No complex dependencies

### 2. Flat Attribute Map Compatibility
**Principle**: Convert SDK objects to flat `map[string]string` matching SDKv2 InstanceState format.

**Rationale**:
- Exporter expects flat attribute maps for dependency resolution
- Must match SDKv2 format exactly for compatibility
- Enables exporter to work with both SDKv2 and Framework resources
- Preserves existing export behavior

**Implementation**:
- Function accepts SDK resource object pointer
- Returns `map[string]string`
- All attributes converted to strings
- Attribute names match schema exactly

### 3. Dependency Reference Preservation
**Principle**: Ensure all dependency references are included in attribute map.

**Rationale**:
- Exporter uses dependency references for ordering
- Exporter uses dependency references for HCL generation
- Missing dependencies break export functionality
- Critical for correct export output

**Implementation**:
- Extract dependency IDs from nested objects
- Include all dependency attributes in map
- Mark dependency references with CRITICAL comments
- Verify attribute names match RefAttrs from Stage 1

### 4. Clear Documentation
**Principle**: Clearly document Phase 1 temporary nature and removal plan.

**Rationale**:
- Prevents confusion about purpose of code
- Provides clear migration path
- Helps future developers understand context
- Facilitates Phase 2 cleanup

**Implementation**:
- File-level comment explains Phase 1/Phase 2
- TODO comment marks code for removal
- Function documentation explains purpose
- Attribute map format is documented

---

## Architecture

### File Structure

```
genesyscloud/<resource_name>/
├── resource_genesyscloud_<resource_name>_schema.go          ← Stage 1
├── resource_genesyscloud_<resource_name>.go                 ← Stage 2
├── data_source_genesyscloud_<resource_name>.go              ← Stage 2
├── resource_genesyscloud_<resource_name>_test.go            ← Stage 3
├── data_source_genesyscloud_<resource_name>_test.go         ← Stage 3
├── genesyscloud_<resource_name>_init_test.go                ← Stage 3
├── resource_genesyscloud_<resource_name>_export_utils.go    ← Stage 4 (THIS)
└── genesyscloud_<resource_name>_proxy.go                    ← NOT MODIFIED
```

### Export Utilities File Components

```
┌─────────────────────────────────────────────────────────┐
│  resource_genesyscloud_<resource_name>_export_utils.go  │
├─────────────────────────────────────────────────────────┤
│  1. File-Level Documentation                            │
│     - Phase 1 temporary explanation                     │
│     - TODO for Phase 2 removal                          │
│     - Purpose and rationale                             │
├─────────────────────────────────────────────────────────┤
│  2. Package Declaration and Imports                     │
│     - Package name                                      │
│     - SDK import                                        │
├─────────────────────────────────────────────────────────┤
│  3. Attribute Mapping Function                          │
│     - build<ResourceName>Attributes()                   │
│     - Accepts SDK object pointer                        │
│     - Returns flat attribute map                        │
│     - Includes all attributes                           │
│     - Includes all dependency references                │
└─────────────────────────────────────────────────────────┘
```

### Integration with Stage 2

```
┌─────────────────────────────────────────────────────────┐
│  GetAll<ResourceName>SDK() in Stage 2 Resource File     │
├─────────────────────────────────────────────────────────┤
│  1. Fetch all resources from API                        │
│  2. Build initial export map with IDs and names         │
│  3. For each resource:                                  │
│     a. Call build<ResourceName>Attributes()             │
│     b. Add attribute map to ResourceMeta                │
│  4. Return export map with attributes                   │
└─────────────────────────────────────────────────────────┘
```

---

## Component Design

### 1. File-Level Documentation

**Purpose**: Explain Phase 1 temporary nature and provide context.

**Design Pattern**:
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

package <resource_name>
```

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

package routing_wrapupcode
```

**Key Elements**:

| Element | Purpose |
|---------|---------|
| "IMPORTANT" marker | Draws attention to temporary nature |
| Phase 1/Phase 2 explanation | Provides migration context |
| TODO comment | Marks code for removal |
| File path | Helps locate file for cleanup |
| "migration scaffolding" term | Clearly indicates temporary nature |

---

### 2. Package Declaration and Imports

**Purpose**: Minimal imports for SDK object handling.

**Design Pattern**:
```go
package <resource_name>

import (
    "github.com/mypurecloud/platform-client-sdk-go/v176/platformclientv2"
)
```

**Key Points**:
- Only SDK import needed
- No Framework imports (works with SDK objects)
- No other dependencies
- Self-contained file

---

### 3. Attribute Mapping Function

**Purpose**: Convert SDK resource object to flat attribute map.

**Design Pattern**:
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
// Note: [Resource-specific notes about complexity, nested attributes, etc.]
func build<ResourceName>Attributes(<resource> *platformclientv2.<ResourceType>) map[string]string {
    attributes := make(map[string]string)

    // Basic attributes
    if <resource>.Id != nil {
        attributes["id"] = *<resource>.Id
    }
    if <resource>.Name != nil {
        attributes["name"] = *<resource>.Name
    }

    // Optional attributes
    if <resource>.OptionalAttr != nil {
        attributes["optional_attr"] = *<resource>.OptionalAttr
    }

    // ⭐ CRITICAL: Dependency reference (used by exporter for dependency resolution)
    if <resource>.Dependency != nil && <resource>.Dependency.Id != nil {
        attributes["dependency_id"] = *<resource>.Dependency.Id
    }

    return attributes
}
```

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

**Function Design Decisions**:

| Decision | Rationale |
|----------|-----------|
| Accept pointer | SDK objects are pointers |
| Return `map[string]string` | Matches SDKv2 InstanceState format |
| Nil checks for all attributes | SDK pointers may be nil |
| String conversion | All values must be strings in map |
| CRITICAL comment for dependencies | Highlights importance for exporter |
| Attribute names match schema | Required for exporter dependency resolution |

---

## Attribute Mapping Patterns

### Pattern 1: Basic Attributes

**Purpose**: Map simple string attributes.

**Pattern**:
```go
if resource.Attribute != nil {
    attributes["attribute"] = *resource.Attribute
}
```

**Example**:
```go
if wrapupcode.Name != nil {
    attributes["name"] = *wrapupcode.Name
}
if wrapupcode.Description != nil {
    attributes["description"] = *wrapupcode.Description
}
```

**Key Points**:
- Always check nil before dereferencing
- Attribute name must match schema exactly
- Use lowercase with underscores (snake_case)

---

### Pattern 2: Dependency References

**Purpose**: Extract dependency IDs from nested objects.

**Pattern**:
```go
// ⭐ CRITICAL: Dependency reference (used by exporter for dependency resolution)
if resource.Dependency != nil && resource.Dependency.Id != nil {
    attributes["dependency_id"] = *resource.Dependency.Id
}
```

**Example**:
```go
// ⭐ CRITICAL: Dependency reference (used by exporter for dependency resolution)
if wrapupcode.Division != nil && wrapupcode.Division.Id != nil {
    attributes["division_id"] = *wrapupcode.Division.Id
}
```

**Key Points**:
- Check both parent and child for nil
- Extract ID from nested object
- Mark with CRITICAL comment
- Attribute name must match RefAttrs from Stage 1

---

### Pattern 3: Multiple Dependencies

**Purpose**: Handle resources with multiple dependencies.

**Pattern**:
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
- Each dependency gets separate nil check
- Each dependency marked as CRITICAL
- All dependencies must be in RefAttrs from Stage 1

---

### Pattern 4: Optional Attributes

**Purpose**: Include optional attributes only if present.

**Pattern**:
```go
if resource.OptionalAttr != nil {
    attributes["optional_attr"] = *resource.OptionalAttr
}
```

**Key Points**:
- Nil check prevents panic
- Attribute not included if nil
- Matches SDKv2 behavior (omit if not set)

---

### Pattern 5: Boolean Attributes

**Purpose**: Convert boolean to string.

**Pattern**:
```go
if resource.BoolAttr != nil {
    if *resource.BoolAttr {
        attributes["bool_attr"] = "true"
    } else {
        attributes["bool_attr"] = "false"
    }
}
```

**Key Points**:
- Convert bool to "true" or "false" string
- Nil check for optional booleans
- Consistent string representation

---

### Pattern 6: Integer Attributes

**Purpose**: Convert integer to string.

**Pattern**:
```go
if resource.IntAttr != nil {
    attributes["int_attr"] = fmt.Sprintf("%d", *resource.IntAttr)
}
```

**Key Points**:
- Use `fmt.Sprintf` for conversion
- Nil check for optional integers
- Format as decimal number

---

## Integration with Exporter

### Export Flow

```
1. User runs: terraform export genesyscloud_routing_wrapupcode
   ↓
2. Exporter calls: GetAllRoutingWrapupcodesSDK()
   ↓
3. GetAllRoutingWrapupcodesSDK() fetches resources from API
   ↓
4. For each resource:
   a. Call buildWrapupcodeAttributes(resource)
   b. Get flat attribute map
   c. Add to ResourceMeta.ExportAttributes
   ↓
5. Exporter receives ResourceIDMetaMap with attributes
   ↓
6. Exporter resolves dependencies using attribute map
   ↓
7. Exporter generates HCL with correct references
   ↓
8. Export complete
```

### Dependency Resolution

**How Exporter Uses Attribute Map**:

1. **Read dependency attribute**: `divisionId := attributes["division_id"]`
2. **Look up dependency resource**: Find division with that ID
3. **Generate HCL reference**: `division_id = genesyscloud_auth_division.division_label.id`
4. **Order resources**: Ensure division exported before wrapupcode

**Why Attribute Names Matter**:
- Exporter uses attribute names from RefAttrs (Stage 1)
- Attribute name mismatch breaks dependency resolution
- Must match schema attribute names exactly

---

## Phase 1 vs Phase 2 Design

### Phase 1 Design (Current - Temporary)

**Architecture**:
```
┌─────────────────────────────────────────────────────────┐
│  Framework Resource                                     │
├─────────────────────────────────────────────────────────┤
│  GetAll<ResourceName>SDK()                              │
│  ├─ Fetch SDK objects from API                          │
│  ├─ Build export map                                    │
│  └─ For each resource:                                  │
│     ├─ Call build<ResourceName>Attributes()            │
│     └─ Add flat map to ResourceMeta                     │
└─────────────────────────────────────────────────────────┘
                    ↓
┌─────────────────────────────────────────────────────────┐
│  Export Utilities (Temporary)                           │
├─────────────────────────────────────────────────────────┤
│  build<ResourceName>Attributes()                        │
│  ├─ Accept SDK object                                   │
│  ├─ Convert to flat map[string]string                   │
│  └─ Return attribute map                                │
└─────────────────────────────────────────────────────────┘
                    ↓
┌─────────────────────────────────────────────────────────┐
│  Exporter (Legacy)                                      │
├─────────────────────────────────────────────────────────┤
│  ├─ Receive ResourceIDMetaMap with flat attributes      │
│  ├─ Resolve dependencies using attribute map            │
│  └─ Generate HCL                                        │
└─────────────────────────────────────────────────────────┘
```

**Files**:
- Export utilities file (temporary)
- `GetAll<ResourceName>SDK()` (temporary)
- `GetAll<ResourceName>()` (future, not used yet)

**Pros**:
- Works with existing exporter
- Enables gradual migration
- No exporter changes needed

**Cons**:
- Extra file per resource
- Duplication of logic
- Temporary code to maintain

### Phase 2 Design (Future - Permanent)

**Architecture**:
```
┌─────────────────────────────────────────────────────────┐
│  Framework Resource                                     │
├─────────────────────────────────────────────────────────┤
│  GetAll<ResourceName>()                                 │
│  ├─ Fetch resources from API                            │
│  ├─ Build export map                                    │
│  └─ Return ResourceIDMetaMap (no flat attributes)       │
└─────────────────────────────────────────────────────────┘
                    ↓
┌─────────────────────────────────────────────────────────┐
│  Exporter (Updated)                                     │
├─────────────────────────────────────────────────────────┤
│  ├─ Receive ResourceIDMetaMap                           │
│  ├─ Work natively with Framework types                  │
│  ├─ Resolve dependencies using Framework model          │
│  └─ Generate HCL                                        │
└─────────────────────────────────────────────────────────┘
```

**Files**:
- Export utilities file removed
- `GetAll<ResourceName>SDK()` removed
- Only `GetAll<ResourceName>()` remains

**Pros**:
- Cleaner code
- No temporary scaffolding
- Native Framework integration

**Cons**:
- Requires exporter refactoring
- Can only happen after all resources migrated

---

## Design Patterns and Best Practices

### Pattern 1: Comprehensive Documentation

**Pattern**:
```go
// File-level: Explain Phase 1 temporary nature
// Function-level: Explain purpose and format
// Inline: Mark dependency references as CRITICAL
```

**Why**:
- Prevents confusion
- Provides context
- Facilitates cleanup
- Helps future developers

### Pattern 2: Nil Safety

**Pattern**:
```go
if resource.Attribute != nil {
    attributes["attribute"] = *resource.Attribute
}
```

**Why**:
- Prevents panics
- Handles optional attributes
- Matches SDKv2 behavior

### Pattern 3: Dependency Highlighting

**Pattern**:
```go
// ⭐ CRITICAL: Dependency reference (used by exporter for dependency resolution)
if resource.Dependency != nil && resource.Dependency.Id != nil {
    attributes["dependency_id"] = *resource.Dependency.Id
}
```

**Why**:
- Draws attention to critical code
- Explains importance
- Helps with debugging
- Documents exporter usage

### Pattern 4: Attribute Name Consistency

**Pattern**:
```go
// Attribute name must match schema exactly
attributes["division_id"] = *wrapupcode.Division.Id  // Matches schema "division_id"
```

**Why**:
- Required for exporter dependency resolution
- Must match RefAttrs from Stage 1
- Prevents export failures

### Pattern 5: Self-Contained File

**Pattern**:
- Single file per resource
- Minimal imports (only SDK)
- No complex dependencies
- Easy to remove

**Why**:
- Simplifies Phase 2 cleanup
- Reduces coupling
- Clear boundaries
- Easy to understand

---

## Migration Considerations

### Cleanup Checklist for Phase 2

When removing export utilities in Phase 2:

- [ ] Delete `resource_genesyscloud_<resource_name>_export_utils.go` file
- [ ] Remove `GetAll<ResourceName>SDK()` function from resource file
- [ ] Remove `build<ResourceName>Attributes()` calls from `GetAll<ResourceName>SDK()`
- [ ] Update exporter to use `GetAll<ResourceName>()` (Framework version)
- [ ] Remove SDK diagnostics import if no longer needed
- [ ] Update exporter to work with Framework types
- [ ] Test export functionality with Framework types
- [ ] Verify dependency resolution still works

### Common Pitfalls

#### Pitfall 1: Missing Dependency Reference
**Problem**: Dependency attribute not included in map.
**Solution**: Verify all RefAttrs from Stage 1 are included.

#### Pitfall 2: Incorrect Attribute Name
**Problem**: Attribute name doesn't match schema.
**Solution**: Use exact attribute name from schema.

#### Pitfall 3: Missing Nil Check
**Problem**: Panic when dereferencing nil pointer.
**Solution**: Always check nil before dereferencing.

#### Pitfall 4: Wrong String Conversion
**Problem**: Boolean or integer not converted to string correctly.
**Solution**: Use proper conversion (`fmt.Sprintf` for numbers, "true"/"false" for booleans).

---

## Summary

### Key Design Decisions

1. **Phase 1 Temporary Scaffolding**: Separate export utilities file per resource
2. **Flat Attribute Map**: Convert SDK objects to `map[string]string`
3. **Dependency Preservation**: Include all dependency references with CRITICAL markers
4. **Clear Documentation**: Explain Phase 1/Phase 2 and mark for removal
5. **Self-Contained**: Minimal dependencies, easy to remove

### File Structure

```
Export Utilities File:
├── File-level documentation (Phase 1 temporary, TODO)
├── Package declaration and imports (SDK only)
└── Attribute mapping function
    ├── Function documentation
    ├── Basic attributes
    ├── Optional attributes
    └── Dependency references (marked CRITICAL)
```

### Integration Points

- Called from `GetAll<ResourceName>SDK()` in Stage 2
- Attribute map added to `ResourceMeta.ExportAttributes`
- Exporter uses attribute map for dependency resolution
- Enables export functionality for Framework resources

### Next Steps

After completing Stage 4 export utilities:
1. Test export functionality
2. Verify dependency resolution
3. Confirm export output matches SDKv2
4. Complete migration for this resource
5. Begin next resource migration

---

## References

- **Reference Implementation**: `genesyscloud/routing_wrapupcode/resource_genesyscloud_routing_wrapupcode_export_utils.go`
- **Stage 2 GetAll Functions**: `genesyscloud/routing_wrapupcode/resource_genesyscloud_routing_wrapupcode.go`
- **Stage 1 Exporter Config**: `genesyscloud/routing_wrapupcode/resource_genesyscloud_routing_wrapupcode_schema.go`
