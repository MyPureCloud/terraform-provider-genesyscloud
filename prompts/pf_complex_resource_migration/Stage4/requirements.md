# Stage 4 – Export Functionality Requirements (Complex Resources)

## Overview

Stage 4 focuses on implementing export functionality for migrated Plugin Framework **complex resources**. This stage creates a separate export utilities file that converts SDK resource objects to flat attribute maps for the legacy exporter's dependency resolution logic. For complex resources, this includes handling nested structures, multiple dependencies, and additional API calls to fetch complete resource state.

**Reference Implementation**: 
- `genesyscloud/user/resource_genesyscloud_user_export_utils.go`

**Key Differences from Simple Resources**:
- Multiple flatten helper functions for nested structures
- Additional API calls to fetch complete state (voicemail, routing utilization)
- Complex dependency references (extension pools, skills, languages, locations)
- Nested attribute flattening (addresses, employer_info, routing_utilization)
- Array/set attribute handling (profile_skills, certifications)
- Error handling for API call failures

---

## Objectives

### Primary Goal
Create export utilities that enable the legacy exporter to work with Plugin Framework complex resources by providing flat attribute maps with complete resource state including nested structures and dependencies.

### Specific Objectives
1. Create export utilities file with main attribute mapping function
2. Implement flatten helper functions for nested structures
3. Make additional API calls to fetch complete resource state
4. Handle all dependency references (simple and nested)
5. Convert nested structures to flat attribute maps
6. Handle arrays and sets with proper indexing
7. Document Phase 1 temporary nature with TODO comments
8. Maintain compatibility with existing exporter behavior
9. Preserve export structure and dependency resolution

---

## Scope

### In Scope for Stage 4

#### 1. Export Utilities File
- Create `resource_genesyscloud_<resource_name>_export_utils.go` file
- Implement `build<ResourceName>Attributes()` main function
- Implement flatten helper functions for nested structures
- Make additional API calls for complete state
- Convert SDK resource objects to flat attribute maps
- Include all resource attributes in map
- Include all dependency references (simple and nested) in map

#### 2. Main Attribute Mapping Function
- Accept context, SDK resource object, and proxy
- Make additional API calls if needed (voicemail, utilization, etc.)
- Call flatten helpers for nested structures
- Return flat attribute map
- Return error if critical API calls fail
- Handle optional nested structures

#### 3. Flatten Helper Functions
- One helper per nested structure type
- Convert nested SDK objects to flat attribute format
- Handle arrays/sets with proper indexing
- Extract dependency IDs from nested objects
- Set count attributes (e.g., `addresses.#`)
- Handle optional nested attributes

#### 4. Dependency References
- Simple dependencies (division_id, manager, etc.)
- Nested dependencies (skill_id, language_id, location_id)
- Deep nested dependencies (extension_pool_id in addresses)
- Mark all dependencies with CRITICAL comments
- Ensure attribute names match RefAttrs from Stage 1

#### 5. Additional API Calls
- Fetch voicemail policies if needed
- Fetch routing utilization if needed
- Fetch extension pool IDs if needed
- Handle API call failures gracefully
- Return errors for critical failures
- Log warnings for non-critical failures

#### 6. Documentation
- Add file-level comment explaining Phase 1 temporary nature
- Add function-level comments for all functions
- Add TODO comment for Phase 2 removal
- Document attribute map format
- Document dependency references
- Document error handling strategy

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

#### FR2: Main Attribute Mapping Function
- ✅ `build<ResourceName>Attributes()` function implemented
- ✅ Function accepts context, SDK resource object pointer, and proxy
- ✅ Function returns `map[string]string` and error
- ✅ Function includes comprehensive documentation
- ✅ Function makes additional API calls if needed
- ✅ Function calls flatten helpers for nested structures
- ✅ Function handles errors appropriately

#### FR3: Flatten Helper Functions
- ✅ One helper per nested structure type
- ✅ Helpers convert nested objects to flat format
- ✅ Helpers handle arrays/sets with indexing
- ✅ Helpers extract dependency IDs
- ✅ Helpers set count attributes
- ✅ Helpers include comprehensive documentation

#### FR4: Attribute Map Completeness
- ✅ All schema attributes are included in map
- ✅ ID attribute is included
- ✅ Name attribute is included
- ✅ All optional attributes are included (if present)
- ✅ All nested structures are flattened
- ✅ All dependency references are included (simple and nested)
- ✅ Array/set counts are included

#### FR5: Dependency References
- ✅ All simple dependency attributes are included
- ✅ All nested dependency attributes are included
- ✅ Dependency IDs are extracted from nested objects
- ✅ Attribute names match schema exactly
- ✅ All dependencies marked with CRITICAL comments
- ✅ Exporter can resolve dependencies correctly

#### FR6: Additional API Calls
- ✅ Voicemail policies fetched if needed
- ✅ Routing utilization fetched if needed
- ✅ Extension pool IDs fetched if needed
- ✅ API call failures handled appropriately
- ✅ Errors returned for critical failures
- ✅ Warnings logged for non-critical failures

#### FR7: Documentation
- ✅ File-level comment explains Phase 1 temporary nature
- ✅ Function comments explain purpose and parameters
- ✅ TODO comment marks code for Phase 2 removal
- ✅ Attribute map format is documented
- ✅ Dependency references are highlighted
- ✅ Error handling is documented

#### FR8: Integration
- ✅ Function is called from `GetAll<ResourceName>SDK()` in Stage 2
- ✅ Attribute map is added to export metadata
- ✅ Errors are handled appropriately (skip user on error)
- ✅ Exporter can successfully export resources
- ✅ Dependency resolution works correctly

### Non-Functional Requirements

#### NFR1: Code Quality
- ✅ Code follows Go best practices
- ✅ Code follows existing codebase conventions
- ✅ Proper nil checks for optional attributes
- ✅ Clear and consistent naming
- ✅ Modular helper functions

#### NFR2: Documentation Quality
- ✅ Comments are clear and comprehensive
- ✅ Phase 1/Phase 2 distinction is clear
- ✅ Purpose and rationale are explained
- ✅ Migration path is documented
- ✅ Error handling is documented

#### NFR3: Maintainability
- ✅ Code is easy to understand
- ✅ Code is easy to remove in Phase 2
- ✅ Helper functions are reusable
- ✅ Self-contained in single file
- ✅ Minimal coupling with other code

#### NFR4: Error Handling
- ✅ Critical errors are returned
- ✅ Non-critical errors are logged
- ✅ Caller can skip failed resources
- ✅ Error messages are helpful
- ✅ Matches SDKv2 error handling behavior

---

## Dependencies and Prerequisites

### Prerequisites

#### 1. Stage 1, 2, and 3 Completion
- Schema file must be complete (Stage 1)
- Resource implementation must be complete (Stage 2)
- Utils file must be complete (Stage 2)
- `GetAll<ResourceName>SDK()` function must be implemented (Stage 2)
- Tests must be complete (Stage 3)

#### 2. Understanding of Export Mechanism
- Familiarity with exporter architecture
- Understanding of dependency resolution
- Knowledge of flat attribute map format
- Understanding of Phase 1/Phase 2 migration strategy
- Understanding of nested attribute flattening

#### 3. Reference Implementation
- Study `user` export utilities file
- Understand attribute mapping patterns
- Review flatten helper patterns
- Review additional API call patterns
- Review dependency reference handling

### Dependencies

#### 1. Package Imports (Complex Resources)
```go
import (
    "context"
    "encoding/json"
    "fmt"
    "log"
    "strconv"

    "github.com/mypurecloud/platform-client-sdk-go/v176/platformclientv2"
)
```

#### 2. SDK Resource Object
- Function receives SDK resource object from `GetAll<ResourceName>SDK()`
- Object contains all resource attributes
- Object includes nested dependency objects
- Object may require expansions for complete state

#### 3. Proxy Object
- Function receives proxy for additional API calls
- Proxy provides methods for voicemail, utilization, etc.
- Proxy provides access to API clients

#### 4. Context
- Function receives context for API calls
- Context enables cancellation and timeouts

#### 5. Exporter Integration
- Function is called from `GetAll<ResourceName>SDK()` in Stage 2
- Attribute map is added to `ResourceMeta.ExportAttributes`
- Errors are handled (skip resource on error)
- Exporter uses attribute map for dependency resolution

---

## Constraints

### Technical Constraints

#### TC1: Flat Attribute Map Format
- **Constraint**: Must use `map[string]string` format
- **Rationale**: Matches SDKv2 InstanceState format used by exporter
- **Impact**: All values must be converted to strings, nested structures must be flattened

#### TC2: Attribute Naming
- **Constraint**: Attribute names must match schema exactly
- **Rationale**: Exporter uses attribute names to resolve dependencies
- **Impact**: Any mismatch breaks dependency resolution

#### TC3: Nested Attribute Flattening
- **Constraint**: Nested structures must be flattened with dot notation and indexing
- **Rationale**: Matches SDKv2 InstanceState format
- **Impact**: Complex flattening logic required

#### TC4: Dependency References
- **Constraint**: All dependency references must be included (simple and nested)
- **Rationale**: Exporter needs dependency IDs for ordering and HCL generation
- **Impact**: Missing dependencies break export

#### TC5: Additional API Calls
- **Constraint**: May need additional API calls for complete state
- **Rationale**: Some attributes not included in main API response
- **Impact**: More complex implementation, error handling required

#### TC6: Phase 1 Temporary Code
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

#### PC3: Error Handling Compatibility
- **Constraint**: Error handling must match SDKv2 behavior
- **Rationale**: Consistent behavior across SDKv2 and Framework
- **Impact**: Return errors for critical failures, log warnings for non-critical

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

### Main Attribute Mapping Function
- [ ] `build<ResourceName>Attributes()` function implemented
- [ ] Function signature is correct (accepts context, pointer, proxy, returns map and error)
- [ ] Function includes comprehensive documentation
- [ ] Function comment explains parameters and return values
- [ ] Function comment documents attribute map format
- [ ] Function comment documents error handling

### Flatten Helper Functions
- [ ] One helper per nested structure type
- [ ] Helpers convert nested objects to flat format
- [ ] Helpers handle arrays/sets with indexing
- [ ] Helpers extract dependency IDs
- [ ] Helpers set count attributes
- [ ] Helpers include comprehensive documentation
- [ ] Helpers handle optional nested structures

### Attribute Map Implementation
- [ ] ID attribute is included
- [ ] Name attribute is included
- [ ] All optional attributes are included (with nil checks)
- [ ] All nested structures are flattened
- [ ] All dependency references are included (simple and nested)
- [ ] Nested objects are handled correctly (IDs extracted)
- [ ] Arrays/sets are handled with proper indexing
- [ ] Count attributes are set (e.g., `addresses.#`)
- [ ] Attribute names match schema exactly

### Dependency References
- [ ] All simple dependency attributes are marked as CRITICAL
- [ ] All nested dependency attributes are marked as CRITICAL
- [ ] Dependency IDs are extracted from nested objects
- [ ] Dependency attribute names match RefAttrs from Stage 1
- [ ] Comment explains dependency usage by exporter

### Additional API Calls
- [ ] Voicemail policies fetched if needed
- [ ] Routing utilization fetched if needed
- [ ] Extension pool IDs fetched if needed
- [ ] API call failures handled appropriately
- [ ] Errors returned for critical failures
- [ ] Warnings logged for non-critical failures

### Integration
- [ ] Function is called from `GetAll<ResourceName>SDK()` in Stage 2
- [ ] Attribute map is added to `ResourceMeta.ExportAttributes`
- [ ] Errors are handled (skip resource on error)
- [ ] Export functionality works correctly
- [ ] Dependency resolution works correctly

### Code Quality
- [ ] Code compiles without errors
- [ ] Code follows Go conventions
- [ ] Nil checks are present for optional attributes
- [ ] Helper functions are modular and reusable
- [ ] Code is self-contained in single file
- [ ] Error handling is appropriate

---

## Example: user Export Utilities

### File Structure
```
genesyscloud/user/
├── resource_genesyscloud_user_schema.go                     (Stage 1)
├── resource_genesyscloud_user.go                            (Stage 2)
├── resource_genesyscloud_user_utils.go                      (Stage 2)
├── data_source_genesyscloud_user.go                         (Stage 2)
├── resource_genesyscloud_user_test.go                       (Stage 3)
├── data_source_genesyscloud_user_test.go                    (Stage 3)
├── genesyscloud_user_init_test.go                           (Stage 3)
└── resource_genesyscloud_user_export_utils.go               (Stage 4 - THIS)
```

### File-Level Documentation
```go
// Package user contains temporary export utilities for Plugin Framework user resource.
//
// IMPORTANT: This file contains migration scaffolding that converts SDK types to flat
// attribute maps for the legacy exporter's dependency resolution logic.
//
// TODO: Remove this entire file once all resources are migrated to Plugin Framework
// and the exporter is updated to work natively with Framework types (Phase 2).
// This is Phase 1 temporary code - resource-specific implementation.
//
// File: genesyscloud/user/resource_genesyscloud_user_export_utils.go

package user
```

### Main Attribute Mapping Function (Complex Resource)
```go
// buildUserAttributes creates a flat attribute map from SDK user object for export.
// This function fetches ALL user attributes including voicemail and routing utilization
// via separate API calls, matching SDKv2 readUser behavior.
//
// Parameters:
//   - ctx: Context for API calls
//   - user: User object from API (must include expansions: skills, languages, locations, profileSkills, certifications, employerInfo)
//   - proxy: User proxy for additional API calls (voicemail, utilization, extension pools)
//
// Returns:
//   - map[string]string: Flat attribute map with all user attributes
//   - error: Error if any fetch operation fails (caller should skip this user)
//
// Attribute Map Format (matching SDKv2 InstanceState):
//   - Basic: "name", "email", "division_id", "manager"
//   - Nested: "addresses.0.phone_numbers.0.extension_pool_id"
//   - Arrays: "routing_skills.#" = count, "routing_skills.0.skill_id" = value
//
// Error Handling:
//   - Returns error if voicemail or utilization fetch fails (matching SDKv2 behavior)
//   - Caller should skip user and continue with others
//   - Logs warnings for non-critical issues (e.g., extension pool not found)
func buildUserAttributes(ctx context.Context, user *platformclientv2.User, proxy *userProxy) (map[string]string, error) {
    attributes := make(map[string]string)

    // Basic attributes
    if user.Id != nil {
        attributes["id"] = *user.Id
    }
    if user.Name != nil {
        attributes["name"] = *user.Name
    }
    if user.Email != nil {
        attributes["email"] = *user.Email
    }

    // ⭐ CRITICAL: Dependency references (used by exporter for dependency resolution)
    if user.Division != nil && user.Division.Id != nil {
        attributes["division_id"] = *user.Division.Id
    }
    if user.Manager != nil && (*user.Manager).Id != nil {
        attributes["manager"] = *(*user.Manager).Id
    }

    // Complex nested attributes
    if user.Addresses != nil {
        if err := flattenSDKAddressesToAttributes(ctx, *user.Addresses, attributes, proxy); err != nil {
            return nil, fmt.Errorf("failed to flatten addresses: %w", err)
        }
    }

    if user.Skills != nil {
        flattenSDKSkillsToAttributes(*user.Skills, attributes)
    }

    // Fetch voicemail policies (separate API call, matching SDKv2)
    voicemail, _, err := proxy.getVoicemailUserpoliciesById(ctx, *user.Id)
    if err != nil {
        return nil, fmt.Errorf("failed to fetch voicemail policies: %w", err)
    }
    if voicemail != nil {
        flattenSDKVoicemailToAttributes(voicemail, attributes)
    }

    return attributes, nil
}
```

### Flatten Helper Function Example (Nested Structure with Dependencies)
```go
// flattenSDKAddressesToAttributes converts SDK addresses to flat attribute map.
// Handles both phone numbers (with extension pool lookup) and other emails.
//
// Attribute Format (matching SDKv2 InstanceState):
//   - "addresses.#" = "1"
//   - "addresses.0.phone_numbers.#" = count
//   - "addresses.0.phone_numbers.0.number" = "+13175559001"
//   - "addresses.0.phone_numbers.0.extension" = "8701"
//   - "addresses.0.phone_numbers.0.extension_pool_id" = "pool-guid" ⭐ DEPENDENCY
//   - "addresses.0.phone_numbers.0.media_type" = "PHONE"
//   - "addresses.0.phone_numbers.0.type" = "WORK"
//   - "addresses.0.other_emails.#" = count
//   - "addresses.0.other_emails.0.address" = "alt@example.com"
//   - "addresses.0.other_emails.0.type" = "WORK"
//
// Extension Pool Lookup:
//   - For phone numbers with extensions, fetches extension_pool_id via API
//   - Logs warning if pool not found (edge case: extension exists but pool deleted)
//   - Sets empty string if lookup fails (exporter will handle gracefully)
func flattenSDKAddressesToAttributes(ctx context.Context, addresses []platformclientv2.Contact, attributes map[string]string, proxy *userProxy) error {
    if len(addresses) == 0 {
        return nil
    }

    phoneIndex := 0
    emailIndex := 0

    for _, address := range addresses {
        if address.MediaType == nil {
            continue
        }

        switch *address.MediaType {
        case "PHONE", "SMS":
            prefix := fmt.Sprintf("addresses.0.phone_numbers.%d", phoneIndex)

            if address.Address != nil {
                attributes[prefix+".number"] = *address.Address
            }
            if address.Extension != nil {
                attributes[prefix+".extension"] = *address.Extension

                // ⭐ CRITICAL: Fetch extension pool ID (dependency reference)
                poolId := fetchExtensionPoolId(ctx, *address.Extension, proxy)
                if poolId != "" {
                    attributes[prefix+".extension_pool_id"] = poolId
                } else {
                    log.Printf("Warning: Extension pool not found for extension %s", *address.Extension)
                }
            }

            phoneIndex++

        case "EMAIL":
            // Skip primary email
            if address.VarType != nil && *address.VarType == "PRIMARY" {
                continue
            }

            prefix := fmt.Sprintf("addresses.0.other_emails.%d", emailIndex)

            if address.Address != nil {
                attributes[prefix+".address"] = *address.Address
            }

            emailIndex++
        }
    }

    // Set counts
    attributes["addresses.#"] = "1"
    attributes["addresses.0.phone_numbers.#"] = strconv.Itoa(phoneIndex)
    attributes["addresses.0.other_emails.#"] = strconv.Itoa(emailIndex)

    return nil
}
```

### Key Elements (Complex Resources)

| Element | Purpose |
|---------|---------|
| File-level comment | Explains Phase 1 temporary nature and removal plan |
| TODO comment | Marks code for Phase 2 removal |
| Main function documentation | Explains purpose, parameters, return values, format, error handling |
| Flatten helper documentation | Explains nested structure flattening and format |
| Attribute map format | Documents expected structure with examples |
| CRITICAL comments | Highlights dependency references (simple and nested) |
| Nil checks | Handles optional attributes safely |
| Error handling | Returns errors for critical failures, logs warnings for non-critical |
| Additional API calls | Fetches complete state (voicemail, utilization) |
| Count attributes | Sets array/set counts (e.g., `addresses.#`) |
| Indexing | Uses proper indexing for arrays/sets (e.g., `addresses.0.phone_numbers.0`) |

---

## Complex Resource Patterns

### Pattern 1: Main Function with Additional API Calls

**Purpose**: Fetch complete resource state including data from separate API endpoints.

**Pattern**:
```go
func build<ResourceName>Attributes(ctx context.Context, resource *platformclientv2.<ResourceType>, proxy *<resource>Proxy) (map[string]string, error) {
    attributes := make(map[string]string)

    // Basic attributes
    // ... map basic attributes

    // Nested structures
    if resource.NestedStruct != nil {
        if err := flattenSDK<NestedStruct>ToAttributes(ctx, *resource.NestedStruct, attributes, proxy); err != nil {
            return nil, fmt.Errorf("failed to flatten <nested_struct>: %w", err)
        }
    }

    // Additional API call for complete state
    additionalData, _, err := proxy.getAdditionalData(ctx, *resource.Id)
    if err != nil {
        return nil, fmt.Errorf("failed to fetch additional data: %w", err)
    }
    if additionalData != nil {
        flattenSDK<AdditionalData>ToAttributes(additionalData, attributes)
    }

    return attributes, nil
}
```

### Pattern 2: Flatten Helper for Nested Structure

**Purpose**: Convert nested SDK object to flat attribute map with proper indexing.

**Pattern**:
```go
// flattenSDK<NestedStruct>ToAttributes converts SDK <nested_struct> to flat attribute map.
//
// Attribute Format:
//   - "<nested>.#" = "1"
//   - "<nested>.0.attribute" = value
//   - "<nested>.0.dependency_id" = "guid" ⭐ DEPENDENCY
func flattenSDK<NestedStruct>ToAttributes(nestedStruct *platformclientv2.<NestedStructType>, attributes map[string]string) {
    if nestedStruct == nil {
        return
    }

    prefix := "<nested>.0"

    if nestedStruct.Attribute != nil {
        attributes[prefix+".attribute"] = *nestedStruct.Attribute
    }

    // ⭐ CRITICAL: Dependency reference
    if nestedStruct.Dependency != nil && nestedStruct.Dependency.Id != nil {
        attributes[prefix+".dependency_id"] = *nestedStruct.Dependency.Id
    }

    attributes["<nested>.#"] = "1"
}
```

### Pattern 3: Flatten Helper for Array/Set

**Purpose**: Convert array/set to flat attribute map with indexing.

**Pattern**:
```go
// flattenSDK<Array>ToAttributes converts SDK <array> to flat attribute map.
//
// Attribute Format:
//   - "<array>.#" = count
//   - "<array>.0.attribute" = value
//   - "<array>.0.dependency_id" = "guid" ⭐ DEPENDENCY
func flattenSDK<Array>ToAttributes(array []platformclientv2.<ArrayType>, attributes map[string]string) {
    if len(array) == 0 {
        return
    }

    for i, item := range array {
        prefix := fmt.Sprintf("<array>.%d", i)

        if item.Attribute != nil {
            attributes[prefix+".attribute"] = *item.Attribute
        }

        // ⭐ CRITICAL: Dependency reference
        if item.Dependency != nil && item.Dependency.Id != nil {
            attributes[prefix+".dependency_id"] = *item.Dependency.Id
        }
    }

    attributes["<array>.#"] = strconv.Itoa(len(array))
}
```

### Pattern 4: Flatten Helper with Additional API Call

**Purpose**: Fetch additional data for nested structure (e.g., extension pool ID).

**Pattern**:
```go
func flattenSDK<NestedStruct>ToAttributes(ctx context.Context, nestedStruct []platformclientv2.<NestedStructType>, attributes map[string]string, proxy *<resource>Proxy) error {
    if len(nestedStruct) == 0 {
        return nil
    }

    for i, item := range nestedStruct {
        prefix := fmt.Sprintf("<nested>.%d", i)

        if item.Attribute != nil {
            attributes[prefix+".attribute"] = *item.Attribute
        }

        // ⭐ CRITICAL: Fetch dependency ID via API call
        if item.ReferenceValue != nil {
            dependencyId := fetchDependencyId(ctx, *item.ReferenceValue, proxy)
            if dependencyId != "" {
                attributes[prefix+".dependency_id"] = dependencyId
            } else {
                log.Printf("Warning: Dependency not found for reference %s", *item.ReferenceValue)
            }
        }
    }

    attributes["<nested>.#"] = strconv.Itoa(len(nestedStruct))
    return nil
}
```

### Pattern 5: Error Handling

**Purpose**: Handle API call failures appropriately.

**Pattern**:
```go
// Critical API call - return error if fails
criticalData, _, err := proxy.getCriticalData(ctx, *resource.Id)
if err != nil {
    log.Printf("Failed to fetch critical data for resource %s: %v", *resource.Id, err)
    return nil, fmt.Errorf("failed to fetch critical data: %w", err)
}

// Non-critical API call - log warning if fails
nonCriticalData := fetchNonCriticalData(ctx, *resource.Id, proxy)
if nonCriticalData == "" {
    log.Printf("Warning: Non-critical data not found for resource %s", *resource.Id)
    // Continue without error
}
```

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

- **Reference Implementation**: `genesyscloud/user/resource_genesyscloud_user_export_utils.go`
- **Simple Resource Reference**: `prompts/pf_simple_resource_migration/Stage4/requirements.md`
- **Stage 2 GetAll Functions**: `genesyscloud/user/resource_genesyscloud_user.go`
- **Stage 2 Utils Functions**: `genesyscloud/user/resource_genesyscloud_user_utils.go`
- **Exporter Documentation**: Internal exporter architecture docs

