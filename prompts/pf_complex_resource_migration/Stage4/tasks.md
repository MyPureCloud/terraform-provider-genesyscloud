# Stage 4 – Export Functionality Tasks (Complex Resources)

## Overview

This document provides step-by-step tasks for completing Stage 4 of the Plugin Framework migration for **complex resources**. Follow these tasks in order to implement export functionality for Plugin Framework resources with nested structures, multiple dependencies, and additional API calls by creating temporary export utilities.

**Reference Implementation**: 
- `genesyscloud/user/resource_genesyscloud_user_export_utils.go`

**Estimated Time**: 8-16 hours (complex resources with 10+ flatten helpers)

**Key Differences from Simple Resources**:
- Multiple flatten helper functions (10+ for complex resources)
- Additional API calls for complete state
- Complex nested structure flattening (3-level nesting)
- Multiple dependency types (simple, nested, deep nested)
- Error handling for API call failures
- Array/set handling with proper indexing

---

## Prerequisites

Before starting Stage 4 tasks, ensure:

- [ ] Stage 1 (Schema Migration) is complete and approved
- [ ] Stage 2 (Resource Migration) is complete and approved
- [ ] Stage 3 (Test Migration) is complete and approved
- [ ] `GetAll<ResourceName>SDK()` function is implemented in Stage 2
- [ ] You understand the exporter architecture
- [ ] You have read Stage 4 `requirements.md` and `design.md`
- [ ] You have studied the `user` reference implementation
- [ ] You understand nested attribute flattening patterns
- [ ] You understand error handling strategy

---

## Task Checklist

### Phase 1: File Creation and Setup
- [ ] Task 1.1: Create Export Utilities File
- [ ] Task 1.2: Add File-Level Documentation
- [ ] Task 1.3: Add Package Declaration and Imports

### Phase 2: Main Function Implementation
- [ ] Task 2.1: Implement Main Function Signature
- [ ] Task 2.2: Add Main Function Documentation
- [ ] Task 2.3: Implement Basic Attributes Mapping
- [ ] Task 2.4: Implement Optional Attributes Mapping
- [ ] Task 2.5: Implement Simple Dependency References Mapping

### Phase 3: Flatten Helper Functions
- [ ] Task 3.1: Identify All Nested Structures
- [ ] Task 3.2: Implement Simple Nested Structure Helpers
- [ ] Task 3.3: Implement Array/Set Helpers with Dependencies
- [ ] Task 3.4: Implement Simple Array Helpers (No Dependencies)
- [ ] Task 3.5: Implement Helpers with Additional API Calls
- [ ] Task 3.6: Implement Multi-Array Nested Structure Helpers

### Phase 4: Additional API Calls
- [ ] Task 4.1: Identify Required Additional API Calls
- [ ] Task 4.2: Implement Additional API Call Functions
- [ ] Task 4.3: Implement Flatten Helpers for API Response Data
- [ ] Task 4.4: Add Error Handling for API Calls

### Phase 5: Integration and Testing
- [ ] Task 5.1: Verify Integration with Stage 2
- [ ] Task 5.2: Test Export Functionality
- [ ] Task 5.3: Verify Dependency Resolution (Simple)
- [ ] Task 5.4: Verify Dependency Resolution (Nested)
- [ ] Task 5.5: Verify Dependency Resolution (Deep Nested)
- [ ] Task 5.6: Test Error Handling

### Phase 6: Validation and Review
- [ ] Task 6.1: Review Against Checklist
- [ ] Task 6.2: Code Review and Approval

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
   cd genesyscloud\user
   ```

2. **Create the export utilities file**
   ```powershell
   New-Item -ItemType File -Name "resource_genesyscloud_<resource_name>_export_utils.go"
   ```
   Example:
   ```powershell
   New-Item -ItemType File -Name "resource_genesyscloud_user_export_utils.go"
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

**Example** (user):
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
```

**Deliverable**: File-level documentation added

---

### Task 1.3: Add Package Declaration and Imports

**Objective**: Add package declaration and required imports for complex resources.

**Steps**:

1. **Add package declaration**
   ```go
   package <resource_name>
   ```

2. **Add required imports for complex resources**
   ```go
   import (
       "context"
       "encoding/json"  // If needed for API response unmarshaling
       "fmt"
       "log"
       "strconv"

       "github.com/mypurecloud/platform-client-sdk-go/v176/platformclientv2"
   )
   ```

3. **Verify imports match your needs**
   - Context: Required for API calls
   - JSON: Required if unmarshaling API responses
   - Fmt: Required for string formatting
   - Log: Required for warnings
   - Strconv: Required for type conversions
   - SDK: Required for types

**Example** (user):
```go
package user

import (
    "context"
    "encoding/json"
    "fmt"
    "log"
    "strconv"

    "github.com/mypurecloud/platform-client-sdk-go/v176/platformclientv2"
)
```

**Deliverable**: Package declaration and imports added

---

## Phase 2: Main Function Implementation

### Task 2.1: Implement Main Function Signature

**Objective**: Create the main attribute mapping function signature with context and proxy.

**Steps**:

1. **Add function signature**
   ```go
   func build<ResourceName>Attributes(ctx context.Context, <resource> *platformclientv2.<ResourceType>, proxy *<resource>Proxy) (map[string]string, error) {
       attributes := make(map[string]string)
       
       // Attributes will be added here
       
       return attributes, nil
   }
   ```

2. **Replace placeholders**
   - `<ResourceName>` → Your resource name in PascalCase (e.g., `User`)
   - `<resource>` → Your resource variable name in camelCase (e.g., `user`)
   - `<ResourceType>` → SDK resource type (e.g., `User`)

**Example** (user):
```go
func buildUserAttributes(ctx context.Context, user *platformclientv2.User, proxy *userProxy) (map[string]string, error) {
    attributes := make(map[string]string)
    
    // Attributes will be added here
    
    return attributes, nil
}
```

**Key Points**:
- Accept `context.Context` for API calls
- Accept proxy for additional API calls
- Return `map[string]string` and `error`
- Initialize empty attributes map

**Deliverable**: Function signature implemented

---

### Task 2.2: Add Main Function Documentation

**Objective**: Add comprehensive function documentation explaining parameters, return values, format, and error handling.

**Steps**:

1. **Add function documentation above signature**
   ```go
   // build<ResourceName>Attributes creates a flat attribute map from SDK <resource> object for export.
   // This function fetches ALL <resource> attributes including [additional data] via separate API calls,
   // matching SDKv2 read<ResourceName> behavior.
   //
   // Parameters:
   //   - ctx: Context for API calls
   //   - <resource>: <Resource> object from API (must include expansions: [list expansions])
   //   - proxy: <Resource> proxy for additional API calls ([list API calls])
   //
   // Returns:
   //   - map[string]string: Flat attribute map with all <resource> attributes
   //   - error: Error if any fetch operation fails (caller should skip this <resource>)
   //
   // Attribute Map Format (matching SDKv2 InstanceState):
   //   - Basic: "name", "email", "division_id"
   //   - Nested: "<nested>.0.<sub_nested>.0.attribute"
   //   - Arrays: "<array>.#" = count, "<array>.0.attribute" = value
   //
   // Error Handling:
   //   - Returns error if [critical API calls] fail (matching SDKv2 behavior)
   //   - Caller should skip <resource> and continue with others
   //   - Logs warnings for non-critical issues (e.g., [example])
   ```

2. **Customize documentation**
   - List required expansions
   - List additional API calls
   - List example attributes in format section
   - Document critical vs non-critical errors

**Example** (user):
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
```

**Deliverable**: Function documentation added

---

### Task 2.3: Implement Basic Attributes Mapping

**Objective**: Map basic required attributes (ID, name, etc.).

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

3. **Add other basic string attributes**
   ```go
   if <resource>.Email != nil {
       attributes["email"] = *<resource>.Email
   }
   if <resource>.Department != nil {
       attributes["department"] = *<resource>.Department
   }
   ```

4. **Add boolean attributes** (if applicable)
   ```go
   if <resource>.BoolAttr != nil {
       attributes["bool_attr"] = strconv.FormatBool(*<resource>.BoolAttr)
   }
   ```

5. **Verify attribute names match schema**
   - Check Stage 1 schema file
   - Use exact attribute names
   - Use snake_case (lowercase with underscores)

**Example** (user):
```go
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
if user.State != nil {
    attributes["state"] = *user.State
}
if user.Department != nil {
    attributes["department"] = *user.Department
}
if user.Title != nil {
    attributes["title"] = *user.Title
}
if user.AcdAutoAnswer != nil {
    attributes["acd_auto_answer"] = strconv.FormatBool(*user.AcdAutoAnswer)
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
       attributes["optional_int_attr"] = strconv.Itoa(*<resource>.OptionalIntAttr)
   }
   ```

4. **Add boolean attribute mapping** (if applicable)
   ```go
   if <resource>.OptionalBoolAttr != nil {
       attributes["optional_bool_attr"] = strconv.FormatBool(*<resource>.OptionalBoolAttr)
   }
   ```

**Key Points**:
- Always check nil before dereferencing
- Convert non-string types to strings
- Use exact attribute names from schema

**Deliverable**: Optional attributes mapped

---

### Task 2.5: Implement Simple Dependency References Mapping

**Objective**: Map simple (top-level) dependency references with CRITICAL markers.

**Steps**:

1. **Review RefAttrs from Stage 1 schema**
   - Open `resource_genesyscloud_<resource_name>_schema.go`
   - Find `<ResourceName>Exporter()` function
   - Review `RefAttrs` map
   - Note all simple (top-level) dependency attribute names

2. **Add dependency reference mapping with CRITICAL comment**
   ```go
   // ⭐ CRITICAL: Dependency references (used by exporter for dependency resolution)
   if <resource>.Division != nil && <resource>.Division.Id != nil {
       attributes["division_id"] = *<resource>.Division.Id
   }
   ```

3. **Add all simple dependency references**
   - One mapping per dependency
   - Check both parent and child for nil
   - Extract ID from nested object
   - Mark with CRITICAL comment

**Example** (user):
```go
// ⭐ CRITICAL: Dependency references (used by exporter for dependency resolution)
if user.Division != nil && user.Division.Id != nil {
    attributes["division_id"] = *user.Division.Id
}
if user.Manager != nil && (*user.Manager).Id != nil {
    attributes["manager"] = *(*user.Manager).Id
}
```

**Key Points**:
- CRITICAL comment is mandatory
- Check both parent and child for nil
- Attribute names must match RefAttrs exactly
- Extract ID from nested object
- Only simple (top-level) dependencies here; nested dependencies handled in flatten helpers

**Deliverable**: Simple dependency references mapped

---

## Phase 3: Flatten Helper Functions

### Task 3.1: Identify All Nested Structures

**Objective**: Identify all nested structures that need flatten helpers.

**Steps**:

1. **Review Stage 1 schema file**
   - Open `resource_genesyscloud_<resource_name>_schema.go`
   - Find all nested blocks (SingleNestedBlock, ListNestedBlock, SetNestedBlock)
   - Note nested structure names and types

2. **Review Stage 2 utils file**
   - Open `resource_genesyscloud_<resource_name>_utils.go`
   - Find all `flatten<NestedStruct>` functions
   - Note nested structure patterns

3. **Create list of flatten helpers needed**
   - One helper per nested structure type
   - Note which helpers need additional API calls
   - Note which helpers have dependencies
   - Note which helpers have arrays/sets

**Example** (user):
```
Flatten Helpers Needed:
1. flattenSDKAddressesToAttributes - Multi-array nested (phone_numbers, other_emails), needs API call for extension pools
2. flattenSDKSkillsToAttributes - Array with dependencies (skill_id)
3. flattenSDKLanguagesToAttributes - Array with dependencies (language_id)
4. flattenSDKLocationsToAttributes - Array with dependencies (location_id)
5. flattenSDKProfileSkillsToAttributes - Simple string array
6. flattenSDKCertificationsToAttributes - Simple string array
7. flattenSDKEmployerInfoToAttributes - Simple nested structure
8. flattenSDKVoicemailToAttributes - Simple nested structure (from API call)
9. flattenSDKRoutingUtilizationToAttributes - Complex nested with API call
```

**Deliverable**: List of flatten helpers needed

---

### Task 3.2: Implement Simple Nested Structure Helpers

**Objective**: Implement flatten helpers for simple nested structures (no arrays, no API calls).

**Steps**:

1. **For each simple nested structure, create helper function**
   ```go
   // flatten<NestedStruct>ToAttributes converts SDK <nested_struct> to flat attribute map.
   //
   // Attribute Format (matching SDKv2 InstanceState):
   //   - "<nested>.#" = "1"
   //   - "<nested>.0.attribute" = value
   //   - "<nested>.0.dependency_id" = "guid" ⭐ DEPENDENCY (if applicable)
   func flattenSDK<NestedStruct>ToAttributes(<nestedStruct> *platformclientv2.<NestedStructType>, attributes map[string]string) {
       if <nestedStruct> == nil {
           return
       }

       prefix := "<nested>.0"

       if <nestedStruct>.Attribute != nil {
           attributes[prefix+".attribute"] = *<nestedStruct>.Attribute
       }

       // ⭐ CRITICAL: Dependency reference (if applicable)
       if <nestedStruct>.Dependency != nil && <nestedStruct>.Dependency.Id != nil {
           attributes[prefix+".dependency_id"] = *<nestedStruct>.Dependency.Id
       }

       attributes["<nested>.#"] = "1"
   }
   ```

2. **Add helper to main function**
   ```go
   // In build<ResourceName>Attributes():
   if <resource>.NestedStruct != nil {
       flattenSDK<NestedStruct>ToAttributes(<resource>.NestedStruct, attributes)
   }
   ```

**Example** (user employer_info):
```go
// flattenSDKEmployerInfoToAttributes converts employer info to flat attribute map.
//
// Attribute Format:
//   - "employer_info.#" = "1"
//   - "employer_info.0.official_name" = "John Smith"
//   - "employer_info.0.employee_id" = "EMP-12345"
//   - "employer_info.0.employee_type" = "Full-time"
//   - "employer_info.0.date_hire" = "2020-01-15"
func flattenSDKEmployerInfoToAttributes(employerInfo *platformclientv2.Employerinfo, attributes map[string]string) {
    if employerInfo == nil {
        return
    }

    prefix := "employer_info.0"

    if employerInfo.OfficialName != nil {
        attributes[prefix+".official_name"] = *employerInfo.OfficialName
    }
    if employerInfo.EmployeeId != nil {
        attributes[prefix+".employee_id"] = *employerInfo.EmployeeId
    }
    if employerInfo.EmployeeType != nil {
        attributes[prefix+".employee_type"] = *employerInfo.EmployeeType
    }
    if employerInfo.DateHire != nil {
        attributes[prefix+".date_hire"] = *employerInfo.DateHire
    }

    attributes["employer_info.#"] = "1"
}

// In buildUserAttributes():
if user.EmployerInfo != nil {
    flattenSDKEmployerInfoToAttributes(user.EmployerInfo, attributes)
}
```

**Key Points**:
- Check nil at start
- Use prefix for all attributes
- Set count attribute at end
- Mark dependencies with CRITICAL

**Deliverable**: Simple nested structure helpers implemented

---

### Task 3.3: Implement Array/Set Helpers with Dependencies

**Objective**: Implement flatten helpers for arrays/sets that contain dependency references.

**Steps**:

1. **For each array/set with dependencies, create helper function**
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

2. **Add helper to main function**
   ```go
   // In build<ResourceName>Attributes():
   if <resource>.Array != nil {
       flattenSDK<Array>ToAttributes(*<resource>.Array, attributes)
   }
   ```

**Example** (user skills):
```go
// flattenSDKSkillsToAttributes converts SDK skills to flat attribute map.
// Each skill includes skill_id (dependency reference) and proficiency.
//
// Attribute Format:
//   - "routing_skills.#" = count
//   - "routing_skills.0.skill_id" = "skill-guid" ⭐ DEPENDENCY
//   - "routing_skills.0.proficiency" = "4.5"
func flattenSDKSkillsToAttributes(skills []platformclientv2.Userroutingskill, attributes map[string]string) {
    if len(skills) == 0 {
        return
    }

    for i, skill := range skills {
        prefix := fmt.Sprintf("routing_skills.%d", i)

        // ⭐ CRITICAL: skill_id is the dependency reference
        if skill.Id != nil {
            attributes[prefix+".skill_id"] = *skill.Id
        }

        if skill.Proficiency != nil {
            attributes[prefix+".proficiency"] = fmt.Sprintf("%.1f", *skill.Proficiency)
        }
    }

    attributes["routing_skills.#"] = strconv.Itoa(len(skills))
}

// In buildUserAttributes():
if user.Skills != nil {
    flattenSDKSkillsToAttributes(*user.Skills, attributes)
}
```

**Key Points**:
- Check array length at start
- Use `fmt.Sprintf` for indexed attributes
- Mark dependencies with CRITICAL
- Set count attribute at end
- Use `strconv.Itoa` for count

**Deliverable**: Array/set helpers with dependencies implemented

---

### Task 3.4: Implement Simple Array Helpers (No Dependencies)

**Objective**: Implement flatten helpers for simple string arrays (no dependencies).

**Steps**:

1. **For each simple string array, create helper function**
   ```go
   // flattenSDK<Array>ToAttributes converts <array> to flat attribute map.
   //
   // Attribute Format:
   //   - "<array>.#" = count
   //   - "<array>.0" = "value1"
   //   - "<array>.1" = "value2"
   func flattenSDK<Array>ToAttributes(array []string, attributes map[string]string) {
       if len(array) == 0 {
           return
       }

       for i, item := range array {
           attributes[fmt.Sprintf("<array>.%d", i)] = item
       }

       attributes["<array>.#"] = strconv.Itoa(len(array))
   }
   ```

2. **Add helper to main function**
   ```go
   // In build<ResourceName>Attributes():
   if <resource>.Array != nil {
       flattenSDK<Array>ToAttributes(*<resource>.Array, attributes)
   }
   ```

**Example** (user profile_skills):
```go
// flattenSDKProfileSkillsToAttributes converts profile skills to flat attribute map.
//
// Attribute Format:
//   - "profile_skills.#" = count
//   - "profile_skills.0" = "Java"
//   - "profile_skills.1" = "Python"
func flattenSDKProfileSkillsToAttributes(skills []string, attributes map[string]string) {
    if len(skills) == 0 {
        return
    }

    for i, skill := range skills {
        attributes[fmt.Sprintf("profile_skills.%d", i)] = skill
    }

    attributes["profile_skills.#"] = strconv.Itoa(len(skills))
}

// In buildUserAttributes():
if user.ProfileSkills != nil {
    flattenSDKProfileSkillsToAttributes(*user.ProfileSkills, attributes)
}
```

**Key Points**:
- Simpler than array with dependencies
- No nested attributes, just values
- Still need count attribute

**Deliverable**: Simple array helpers implemented

---

### Task 3.5: Implement Helpers with Additional API Calls

**Objective**: Implement flatten helpers that need additional API calls to fetch dependency IDs.

**Steps**:

1. **For each helper needing API calls, create helper function**
   ```go
   // flattenSDK<NestedStruct>ToAttributes converts SDK <nested_struct> to flat attribute map.
   // Handles [specific complexity] with [additional API call description].
   //
   // Attribute Format:
   //   - "<nested>.#" = "1"
   //   - "<nested>.0.<sub_nested>.#" = count
   //   - "<nested>.0.<sub_nested>.0.attribute" = value
   //   - "<nested>.0.<sub_nested>.0.dependency_id" = "guid" ⭐ DEPENDENCY
   //
   // [Additional API Call Description]:
   //   - [When and why the API call is made]
   //   - [What happens if API call fails]
   func flattenSDK<NestedStruct>ToAttributes(ctx context.Context, nestedStruct []platformclientv2.<NestedStructType>, attributes map[string]string, proxy *<resource>Proxy) error {
       if len(nestedStruct) == 0 {
           return nil
       }

       index := 0

       for _, item := range nestedStruct {
           prefix := fmt.Sprintf("<nested>.0.<sub_nested>.%d", index)

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

           index++
       }

       attributes["<nested>.#"] = "1"
       attributes["<nested>.0.<sub_nested>.#"] = strconv.Itoa(index)

       return nil
   }
   ```

2. **Add helper to main function with error handling**
   ```go
   // In build<ResourceName>Attributes():
   if <resource>.NestedStruct != nil {
       if err := flattenSDK<NestedStruct>ToAttributes(ctx, *<resource>.NestedStruct, attributes, proxy); err != nil {
           return nil, fmt.Errorf("failed to flatten <nested_struct>: %w", err)
       }
   }
   ```

**Example** (user addresses with extension pool lookup - abbreviated):
```go
// flattenSDKAddressesToAttributes converts SDK addresses to flat attribute map.
// Handles both phone numbers (with extension pool lookup) and other emails.
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

    for _, address := range addresses {
        if address.MediaType != nil && *address.MediaType == "PHONE" {
            prefix := fmt.Sprintf("addresses.0.phone_numbers.%d", phoneIndex)

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
        }
    }

    attributes["addresses.#"] = "1"
    attributes["addresses.0.phone_numbers.#"] = strconv.Itoa(phoneIndex)

    return nil
}

// In buildUserAttributes():
if user.Addresses != nil {
    if err := flattenSDKAddressesToAttributes(ctx, *user.Addresses, attributes, proxy); err != nil {
        return nil, fmt.Errorf("failed to flatten addresses: %w", err)
    }
}
```

**Key Points**:
- Accept context and proxy parameters
- Return error for critical failures
- Log warnings for non-critical issues
- Reuse existing helper functions if available

**Deliverable**: Helpers with additional API calls implemented

---

### Task 3.6: Implement Multi-Array Nested Structure Helpers

**Objective**: Implement flatten helpers for nested structures with multiple arrays at same level.

**Steps**:

1. **For each multi-array nested structure, create helper function**
   ```go
   func flattenSDK<NestedStruct>ToAttributes(ctx context.Context, nestedStruct []platformclientv2.<NestedStructType>, attributes map[string]string, proxy *<resource>Proxy) error {
       if len(nestedStruct) == 0 {
           return nil
       }

       array1Index := 0
       array2Index := 0

       for _, item := range nestedStruct {
           if item.Type == "TYPE1" {
               prefix := fmt.Sprintf("<nested>.0.<array1>.%d", array1Index)
               // ... set attributes
               array1Index++
           } else if item.Type == "TYPE2" {
               prefix := fmt.Sprintf("<nested>.0.<array2>.%d", array2Index)
               // ... set attributes
               array2Index++
           }
       }

       attributes["<nested>.#"] = "1"
       attributes["<nested>.0.<array1>.#"] = strconv.Itoa(array1Index)
       attributes["<nested>.0.<array2>.#"] = strconv.Itoa(array2Index)

       return nil
   }
   ```

2. **Use separate index variables for each array**
   - Initialize all index variables at start
   - Increment only the relevant index
   - Set all count attributes at end

**Example** (user addresses with phone_numbers and other_emails):
```go
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
            // ... set phone attributes
            phoneIndex++

        case "EMAIL":
            // Skip primary email
            if address.VarType != nil && *address.VarType == "PRIMARY" {
                continue
            }

            prefix := fmt.Sprintf("addresses.0.other_emails.%d", emailIndex)
            // ... set email attributes
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

**Key Points**:
- Separate index variable for each array
- Increment only relevant index
- Set all count attributes at end
- Handle filtering (e.g., skip primary email)

**Deliverable**: Multi-array nested structure helpers implemented

---

## Phase 4: Additional API Calls

### Task 4.1: Identify Required Additional API Calls

**Objective**: Identify all additional API calls needed for complete resource state.

**Steps**:

1. **Review Stage 2 resource file**
   - Open `resource_genesyscloud_<resource_name>.go`
   - Find `read<ResourceName>()` function
   - Note all API calls made (beyond main resource fetch)

2. **Review Stage 2 utils file**
   - Open `resource_genesyscloud_<resource_name>_utils.go`
   - Find helper functions that make API calls
   - Note which API calls are needed for export

3. **Create list of additional API calls needed**
   - Note API endpoint
   - Note when to call (always vs conditional)
   - Note error handling (critical vs non-critical)

**Example** (user):
```
Additional API Calls Needed:
1. Voicemail policies - Always call, critical (return error if fails)
2. Routing utilization - Always call, critical (return error if fails)
3. Extension pool lookup - Conditional (per phone number), non-critical (log warning if fails)
```

**Deliverable**: List of additional API calls needed

---

### Task 4.2: Implement Additional API Call Functions

**Objective**: Implement functions to make additional API calls (or reuse existing functions).

**Steps**:

1. **Check if helper functions already exist**
   - Review Stage 2 utils file
   - Look for functions like `fetchExtensionPoolId()`
   - Reuse existing functions if available

2. **If helper doesn't exist, create it**
   ```go
   func fetchDependencyId(ctx context.Context, referenceValue string, proxy *<resource>Proxy) string {
       // Make API call
       // Return dependency ID or empty string
   }
   ```

3. **For separate API calls, implement in flatten helper**
   - See Task 4.3 below

**Example** (user - reuses existing function):
```go
// Reuses existing fetchExtensionPoolId function from resource_genesyscloud_user_utils.go
poolId := fetchExtensionPoolId(ctx, *address.Extension, proxy)
```

**Deliverable**: Additional API call functions implemented or identified for reuse

---

### Task 4.3: Implement Flatten Helpers for API Response Data

**Objective**: Implement flatten helpers for data fetched via separate API calls.

**Steps**:

1. **For each separate API call, create flatten helper**
   ```go
   // flattenSDK<Data>ToAttributes fetches and converts <data> to flat map.
   // Makes separate API call to get <data> (matching SDKv2 behavior).
   //
   // Attribute Format:
   //   - "<data>.#" = "1"
   //   - "<data>.0.<nested>.#" = "1"
   //   - "<data>.0.<nested>.0.attribute" = value
   //
   // Returns error if API call fails (matching SDKv2 behavior - caller should skip resource).
   func flattenSDK<Data>ToAttributes(ctx context.Context, resourceId string, proxy *<resource>Proxy, attributes map[string]string) error {
       // Make API call to get data
       apiClient := &proxy.api.Configuration.APIClient
       path := fmt.Sprintf("%s/api/v2/<endpoint>/%s/<data>",
           proxy.api.Configuration.BasePath, resourceId)

       response, err := apiClient.CallAPI(path, "GET", nil, buildHeaderParams(proxy.api),
           nil, nil, "", nil, "")
       if err != nil {
           return fmt.Errorf("failed to fetch <data>: %w", err)
       }

       // Unmarshal response
       var data <DataType>
       if err := json.Unmarshal(response.RawBody, &data); err != nil {
           return fmt.Errorf("failed to unmarshal <data>: %w", err)
       }

       // Flatten data to attributes
       // ... flatten logic

       return nil
   }
   ```

2. **Add helper to main function with error handling**
   ```go
   // In build<ResourceName>Attributes():
   if err := flattenSDK<Data>ToAttributes(ctx, *<resource>.Id, proxy, attributes); err != nil {
       log.Printf("Failed to fetch <data> for <resource> %s: %v", *<resource>.Id, err)
       return nil, fmt.Errorf("failed to fetch <data>: %w", err)
   }
   ```

**Example** (user routing utilization - abbreviated):
```go
// flattenSDKRoutingUtilizationToAttributes fetches and converts routing utilization to flat map.
// Makes separate API call to get utilization settings (matching SDKv2 behavior).
//
// Returns error if API call fails (matching SDKv2 behavior - caller should skip user).
func flattenSDKRoutingUtilizationToAttributes(ctx context.Context, userId string, proxy *userProxy, attributes map[string]string) error {
    // Make API call to get routing utilization
    apiClient := &proxy.routingApi.Configuration.APIClient
    path := fmt.Sprintf("%s/api/v2/routing/users/%s/utilization",
        proxy.routingApi.Configuration.BasePath, userId)

    response, err := apiClient.CallAPI(path, "GET", nil, buildHeaderParams(proxy.routingApi),
        nil, nil, "", nil, "")
    if err != nil {
        return fmt.Errorf("failed to fetch routing utilization: %w", err)
    }

    // Unmarshal response
    var agentUtilization agentUtilizationWithLabels
    if err := json.Unmarshal(response.RawBody, &agentUtilization); err != nil {
        return fmt.Errorf("failed to unmarshal routing utilization: %w", err)
    }

    // If organization-level settings, don't export
    if agentUtilization.Level == "Organization" {
        return nil
    }

    // Flatten media utilization settings
    for sdkType, schemaType := range getUtilizationMediaTypes() {
        if mediaSettings, ok := agentUtilization.Utilization[sdkType]; ok {
            prefix := fmt.Sprintf("routing_utilization.0.%s.0", schemaType)
            attributes[prefix+".maximum_capacity"] = strconv.Itoa(int(mediaSettings.MaximumCapacity))
            attributes[prefix+".include_non_acd"] = strconv.FormatBool(mediaSettings.IncludeNonAcd)
            attributes[fmt.Sprintf("routing_utilization.0.%s.#", schemaType)] = "1"
        }
    }

    attributes["routing_utilization.#"] = "1"
    return nil
}

// In buildUserAttributes():
if err := flattenSDKRoutingUtilizationToAttributes(ctx, *user.Id, proxy, attributes); err != nil {
    log.Printf("Failed to fetch routing utilization for user %s: %v", *user.Id, err)
    return nil, fmt.Errorf("failed to fetch routing utilization: %w", err)
}
```

**Key Points**:
- Make API call using proxy
- Unmarshal response
- Flatten to attributes
- Return error if API call fails
- Reuse existing helper functions if available (e.g., `buildHeaderParams`, `getUtilizationMediaTypes`)

**Deliverable**: Flatten helpers for API response data implemented

---

### Task 4.4: Add Error Handling for API Calls

**Objective**: Add appropriate error handling for all API calls.

**Steps**:

1. **For critical API calls, return error**
   ```go
   data, _, err := proxy.getCriticalData(ctx, *resource.Id)
   if err != nil {
       log.Printf("Failed to fetch critical data for resource %s: %v", *resource.Id, err)
       return nil, fmt.Errorf("failed to fetch critical data: %w", err)
   }
   ```

2. **For non-critical API calls, log warning**
   ```go
   optionalData := fetchOptionalData(ctx, *resource.Id, proxy)
   if optionalData == "" {
       log.Printf("Warning: Optional data not found for resource %s", *resource.Id)
       // Continue without error
   }
   ```

3. **Verify error handling matches SDKv2 behavior**
   - Review Stage 2 resource file
   - Check which API call failures cause read to fail
   - Match that behavior in export utilities

**Example** (user):
```go
// Critical: Voicemail policies
voicemail, _, err := proxy.getVoicemailUserpoliciesById(ctx, *user.Id)
if err != nil {
    log.Printf("Failed to fetch voicemail policies for user %s: %v", *user.Id, err)
    return nil, fmt.Errorf("failed to fetch voicemail policies: %w", err)
}

// Critical: Routing utilization
if err := flattenSDKRoutingUtilizationToAttributes(ctx, *user.Id, proxy, attributes); err != nil {
    log.Printf("Failed to fetch routing utilization for user %s: %v", *user.Id, err)
    return nil, fmt.Errorf("failed to fetch routing utilization: %w", err)
}

// Non-critical: Extension pool lookup
poolId := fetchExtensionPoolId(ctx, *address.Extension, proxy)
if poolId == "" {
    log.Printf("Warning: Extension pool not found for extension %s", *address.Extension)
    // Continue without error
}
```

**Deliverable**: Error handling added for all API calls

---

## Phase 5: Integration and Testing

### Task 5.1: Verify Integration with Stage 2

**Objective**: Verify the function is called from `GetAll<ResourceName>SDK()`.

**Steps**:

1. **Open Stage 2 resource file**
   - File: `resource_genesyscloud_<resource_name>.go`
   - Find `GetAll<ResourceName>SDK()` function

2. **Verify function is called with context and proxy**
   - Look for loop over resources
   - Find call to `build<ResourceName>Attributes(ctx, &resource, proxy)`
   - Verify error is handled (skip resource on error)
   - Verify attribute map is added to `ResourceMeta.ExportAttributes`

3. **Expected pattern in Stage 2**:
   ```go
   for _, resource := range *resources {
       if resource.Id == nil {
           continue
       }

       // Build flat attribute map for exporter (Phase 1 temporary)
       attributes, err := build<ResourceName>Attributes(ctx, &resource, proxy)
       if err != nil {
           log.Printf("Error building attributes for <resource> %s: %v", *resource.Id, err)
           continue // Skip this <resource> and continue with others
       }

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

### Task 5.2: Test Export Functionality

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
   - Undefined functions

3. **Run export command** (if export tool available)
   ```powershell
   terraform export genesyscloud_<resource>
   ```

4. **Verify export output**
   - Check that resources are exported
   - Verify HCL is generated
   - Check for any errors
   - Verify nested structures are flattened correctly

5. **If export tool not available**
   - Verify code compiles
   - Review code manually
   - Proceed to code review

**Deliverable**: Export functionality tested

---

### Task 5.3: Verify Dependency Resolution (Simple)

**Objective**: Verify that simple (top-level) dependency references are resolved correctly.

**Steps**:

1. **Review exported HCL** (if export tool available)
   - Find resources with simple dependencies
   - Verify dependency references are correct
   - Example: `division_id = genesyscloud_auth_division.division_label.id`

2. **Check dependency ordering**
   - Dependencies should be exported before dependent resources
   - Example: Division exported before user

3. **Verify attribute names**
   - Dependency attribute names match schema
   - Dependency attribute names match RefAttrs from Stage 1

4. **If issues found**
   - Check attribute names in `build<ResourceName>Attributes()`
   - Verify RefAttrs in Stage 1 schema
   - Ensure dependency IDs are extracted correctly

**Deliverable**: Simple dependency resolution verified

---

### Task 5.4: Verify Dependency Resolution (Nested)

**Objective**: Verify that nested dependency references (in arrays/sets) are resolved correctly.

**Steps**:

1. **Review exported HCL** (if export tool available)
   - Find resources with nested dependencies
   - Verify nested dependency references are correct
   - Example: `skill_id = genesyscloud_routing_skill.skill_label.id` in routing_skills array

2. **Check nested attribute paths**
   - Verify paths match schema
   - Example: `routing_skills.0.skill_id`

3. **Verify all array elements**
   - Check multiple array elements
   - Verify indexing is correct

4. **If issues found**
   - Check flatten helper implementation
   - Verify attribute paths
   - Ensure dependency IDs are extracted correctly

**Deliverable**: Nested dependency resolution verified

---

### Task 5.5: Verify Dependency Resolution (Deep Nested)

**Objective**: Verify that deep nested dependency references (fetched via API) are resolved correctly.

**Steps**:

1. **Review exported HCL** (if export tool available)
   - Find resources with deep nested dependencies
   - Verify deep nested dependency references are correct
   - Example: `extension_pool_id = genesyscloud_telephony_providers_edges_extension_pool.pool_label.id` in phone_numbers

2. **Check deep nested attribute paths**
   - Verify paths match schema
   - Example: `addresses.0.phone_numbers.0.extension_pool_id`

3. **Verify API call results**
   - Check that API calls are made
   - Verify dependency IDs are fetched correctly
   - Check warning logs for failed lookups

4. **If issues found**
   - Check flatten helper with API call
   - Verify API call implementation
   - Ensure dependency IDs are extracted correctly

**Deliverable**: Deep nested dependency resolution verified

---

### Task 5.6: Test Error Handling

**Objective**: Verify that error handling works correctly.

**Steps**:

1. **Test critical API call failure** (if possible)
   - Simulate voicemail fetch failure
   - Verify error is returned
   - Verify resource is skipped
   - Verify export continues with other resources

2. **Test non-critical API call failure** (if possible)
   - Simulate extension pool lookup failure
   - Verify warning is logged
   - Verify export continues
   - Verify resource is exported (without extension_pool_id)

3. **Review error logs**
   - Check for helpful error messages
   - Verify error context is included
   - Verify warnings are logged appropriately

4. **If issues found**
   - Check error handling in main function
   - Check error handling in flatten helpers
   - Verify error messages are helpful

**Deliverable**: Error handling tested

---

## Phase 6: Validation and Review

### Task 6.1: Review Against Checklist

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

   **Main Attribute Mapping Function**:
   - [ ] `build<ResourceName>Attributes()` function implemented
   - [ ] Function signature is correct (accepts context, pointer, proxy, returns map and error)
   - [ ] Function includes comprehensive documentation
   - [ ] Function comment explains parameters and return values
   - [ ] Function comment documents attribute map format
   - [ ] Function comment documents error handling

   **Flatten Helper Functions**:
   - [ ] One helper per nested structure type
   - [ ] Helpers convert nested objects to flat format
   - [ ] Helpers handle arrays/sets with indexing
   - [ ] Helpers extract dependency IDs
   - [ ] Helpers set count attributes
   - [ ] Helpers include comprehensive documentation
   - [ ] Helpers handle optional nested structures

   **Attribute Map Implementation**:
   - [ ] ID attribute is included
   - [ ] Name attribute is included
   - [ ] All optional attributes are included (with nil checks)
   - [ ] All nested structures are flattened
   - [ ] All dependency references are included (simple and nested)
   - [ ] Nested objects are handled correctly (IDs extracted)
   - [ ] Arrays/sets are handled with proper indexing
   - [ ] Count attributes are set (e.g., `addresses.#`)
   - [ ] Attribute names match schema exactly

   **Dependency References**:
   - [ ] All simple dependency attributes are marked as CRITICAL
   - [ ] All nested dependency attributes are marked as CRITICAL
   - [ ] Dependency IDs are extracted from nested objects
   - [ ] Dependency attribute names match RefAttrs from Stage 1
   - [ ] Comment explains dependency usage by exporter

   **Additional API Calls**:
   - [ ] Voicemail policies fetched if needed
   - [ ] Routing utilization fetched if needed
   - [ ] Extension pool IDs fetched if needed
   - [ ] API call failures handled appropriately
   - [ ] Errors returned for critical failures
   - [ ] Warnings logged for non-critical failures

   **Integration**:
   - [ ] Function is called from `GetAll<ResourceName>SDK()` in Stage 2
   - [ ] Attribute map is added to `ResourceMeta.ExportAttributes`
   - [ ] Errors are handled (skip resource on error)
   - [ ] Export functionality works correctly
   - [ ] Dependency resolution works correctly

   **Code Quality**:
   - [ ] Code compiles without errors
   - [ ] Code follows Go conventions
   - [ ] Nil checks are present for optional attributes
   - [ ] Helper functions are modular and reusable
   - [ ] Code is self-contained in single file
   - [ ] Error handling is appropriate

2. **Fix any issues found**

**Deliverable**: All checklist items verified

---

### Task 6.2: Code Review and Approval

**Objective**: Get peer review and approval.

**Steps**:

1. **Create pull request or review request**
   - Include link to Stage 4 requirements and design docs
   - Highlight Phase 1 temporary nature
   - Note dependency references (simple, nested, deep nested)
   - Note additional API calls
   - Note error handling strategy

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

### Issue 1: Missing Flatten Helper

**Problem**: Nested structure not flattened, attributes missing.

**Solution**:
- Create flatten helper for nested structure
- Call helper from main function
- Verify attribute paths match schema

### Issue 2: Incorrect Indexing

**Problem**: Array elements have wrong indices, export fails.

**Solution**:
- Use separate index variables for each array
- Increment only relevant index
- Set all count attributes at end

### Issue 3: Missing Count Attribute

**Problem**: Exporter can't iterate over array, export fails.

**Solution**:
- Always set `<array>.#` count attribute
- Set count for nested structures: `<nested>.#` = "1"
- Set count at end of flatten helper

### Issue 4: Missing Dependency Reference

**Problem**: Dependency not resolved, export fails.

**Solution**:
- Verify all RefAttrs from Stage 1 are included
- Mark with CRITICAL comment
- Check both simple and nested dependencies
- Verify attribute names match exactly

### Issue 5: Incorrect Attribute Path

**Problem**: Nested attribute has wrong path, export fails.

**Solution**:
- Use correct dot notation: `<nested>.0.<sub_nested>.0.attribute`
- Use proper indexing for arrays: `<array>.0`, `<array>.1`
- Verify paths match schema

### Issue 6: API Call Failure Not Handled

**Problem**: API call fails, incomplete data exported or panic.

**Solution**:
- Return error for critical failures
- Log warning for non-critical failures
- Match SDKv2 error handling behavior

### Issue 7: Missing Nil Check

**Problem**: Panic when dereferencing nil pointer.

**Solution**:
- Always check nil before dereferencing
- Check both parent and child for nested objects
- Example: `if resource.Nested != nil && resource.Nested.Id != nil`

### Issue 8: Wrong Error Handling Strategy

**Problem**: Non-critical failure causes entire export to fail.

**Solution**:
- Return errors only for critical failures
- Log warnings for non-critical failures
- Review SDKv2 behavior to determine criticality

### Issue 9: Missing Additional API Call

**Problem**: Some attributes missing from export.

**Solution**:
- Review Stage 2 read function
- Identify all additional API calls
- Implement flatten helpers for API response data

### Issue 10: Compilation Errors

**Problem**: Code doesn't compile.

**Solution**:
- Check imports (context, json, fmt, log, strconv)
- Verify function signatures
- Check for undefined functions
- Verify type conversions

---

## Completion Criteria

Stage 4 is complete when:

- [ ] All tasks in this document are completed
- [ ] All checklist items are verified
- [ ] Code compiles without errors
- [ ] Export functionality works correctly
- [ ] Dependency resolution works correctly (simple, nested, deep nested)
- [ ] Error handling works correctly
- [ ] Code review is approved
- [ ] **Migration is complete for this resource!** 🎉

---

## Next Steps

After Stage 4 completion:

1. **Celebrate!** 🎉
   - You've successfully migrated a complex resource to Plugin Framework
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
| Phase 1: File Creation and Setup | 30 minutes |
| Phase 2: Main Function Implementation | 1-2 hours |
| Phase 3: Flatten Helper Functions | 3-6 hours |
| Phase 4: Additional API Calls | 2-4 hours |
| Phase 5: Integration and Testing | 1-2 hours |
| Phase 6: Validation and Review | 1-2 hours |
| **Total** | **8-16 hours** |

*Note: Times vary based on resource complexity (number of nested structures, dependencies, API calls).*

**Complexity Factors**:
- Number of flatten helpers (10+ for complex resources)
- Number of additional API calls (2-3 for complex resources)
- Depth of nesting (up to 3 levels)
- Number of dependencies (simple, nested, deep nested)
- Error handling complexity

---

## Complete Migration Summary

### All Stages Complete! 🎉

You have successfully completed all 4 stages of the Plugin Framework migration for a complex resource:

- ✅ **Stage 1**: Schema Migration (schema definitions with nested blocks)
- ✅ **Stage 2**: Resource Migration (CRUD operations with nested models)
- ✅ **Stage 3**: Test Migration (acceptance tests with nested structures)
- ✅ **Stage 4**: Export Functionality (export utilities with flatten helpers)

### Files Created

```
genesyscloud/<resource_name>/
├── resource_genesyscloud_<resource_name>_schema.go          ✅ Stage 1
├── resource_genesyscloud_<resource_name>.go                 ✅ Stage 2
├── resource_genesyscloud_<resource_name>_utils.go           ✅ Stage 2
├── data_source_genesyscloud_<resource_name>.go              ✅ Stage 2
├── resource_genesyscloud_<resource_name>_test.go            ✅ Stage 3
├── data_source_genesyscloud_<resource_name>_test.go         ✅ Stage 3
├── genesyscloud_<resource_name>_init_test.go                ✅ Stage 3
├── resource_genesyscloud_<resource_name>_export_utils.go    ✅ Stage 4
└── genesyscloud_<resource_name>_proxy.go                    (unchanged)
```

### What's Next?

1. **Use as reference** for migrating other complex resources
2. **Share knowledge** with team members
3. **Track progress** on overall migration
4. **Plan for Phase 2** (exporter update and cleanup)

---

## References

- **Reference Implementation**: `genesyscloud/user/resource_genesyscloud_user_export_utils.go`
- **Stage 4 Requirements**: `prompts/pf_complex_resource_migration/Stage4/requirements.md`
- **Stage 4 Design**: `prompts/pf_complex_resource_migration/Stage4/design.md`
- **Stage 2 GetAll Functions**: `genesyscloud/user/resource_genesyscloud_user.go`
- **Stage 2 Utils Functions**: `genesyscloud/user/resource_genesyscloud_user_utils.go`
- **Stage 1 Exporter Config**: `genesyscloud/user/resource_genesyscloud_user_schema.go`
- **Simple Resource Reference**: `prompts/pf_simple_resource_migration/Stage4/tasks.md`
