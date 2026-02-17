# Stage 4 – Export Functionality Design (Complex Resources)

## Overview

This document describes the design decisions, architecture, and patterns used in Stage 4 of the Plugin Framework migration for **complex resources**. Stage 4 focuses on implementing export functionality for Plugin Framework resources with nested structures, multiple dependencies, and additional API calls by creating temporary export utilities that convert SDK objects to flat attribute maps for the legacy exporter's dependency resolution logic.

**Reference Implementation**: 
- `genesyscloud/user/resource_genesyscloud_user_export_utils.go`

**Key Differences from Simple Resources**:
- Multiple flatten helper functions (10+ for complex resources)
- Additional API calls for complete state
- Complex nested structure flattening (3-level nesting)
- Multiple dependency types (simple, nested, deep nested)
- Error handling for API call failures
- Array/set handling with proper indexing

---

## Design Principles

### 1. Phase 1 Temporary Scaffolding (Complex Resources)
**Principle**: Create temporary export utilities with modular flatten helpers that will be removed in Phase 2.

**Rationale**:
- Exporter currently uses SDK diagnostics and flat attribute maps
- Complex resources require multiple flatten helpers for nested structures
- Enables gradual migration without breaking existing export functionality
- Allows Framework resources to work with legacy exporter
- Clear migration path to Phase 2 native Framework support

**Implementation**:
- Separate export utilities file per resource
- Main attribute mapping function
- Multiple flatten helper functions (one per nested structure)
- Clearly marked as Phase 1 temporary with TODO comments
- Self-contained and easy to remove
- Modular design for maintainability

### 2. Modular Flatten Helper Pattern
**Principle**: Create one flatten helper function per nested structure type.

**Rationale**:
- Complex resources have many nested structures
- Each nested structure has different attributes and dependencies
- Modular helpers improve code organization and maintainability
- Easier to test and debug individual helpers
- Reduces complexity of main function

**Implementation**:
- One helper per nested structure (addresses, skills, languages, etc.)
- Helpers convert nested SDK objects to flat format
- Helpers handle arrays/sets with proper indexing
- Helpers extract dependency IDs
- Helpers set count attributes
- Helpers are reusable and testable

### 3. Additional API Calls for Complete State
**Principle**: Make additional API calls to fetch complete resource state when needed.

**Rationale**:
- Some attributes not included in main API response
- Matches SDKv2 behavior (e.g., voicemail, routing utilization)
- Ensures export includes all resource attributes
- Maintains compatibility with SDKv2 export output

**Implementation**:
- Accept context parameter for API calls
- Accept proxy parameter for API access
- Make additional API calls in main function
- Handle API call failures appropriately
- Return errors for critical failures
- Log warnings for non-critical failures

### 4. Comprehensive Dependency Reference Handling
**Principle**: Extract and include all dependency references (simple, nested, and deep nested).

**Rationale**:
- Complex resources have multiple dependency types
- Dependencies may be nested in structures
- Some dependencies require additional API calls (e.g., extension pool ID)
- Exporter needs all dependency IDs for ordering and HCL generation
- Missing dependencies break export functionality

**Implementation**:
- Extract simple dependencies from top-level attributes
- Extract nested dependencies from nested structures
- Fetch deep nested dependencies via API calls
- Mark all dependencies with CRITICAL comments
- Verify attribute names match RefAttrs from Stage 1

### 5. Error Handling Strategy
**Principle**: Return errors for critical failures, log warnings for non-critical issues.

**Rationale**:
- Matches SDKv2 error handling behavior
- Allows caller to skip failed resources and continue
- Prevents partial/incomplete export data
- Provides helpful error messages for debugging

**Implementation**:
- Return error if critical API calls fail (voicemail, utilization)
- Return error if critical flatten operations fail
- Log warnings for non-critical issues (extension pool not found)
- Include error context in error messages
- Caller skips resource on error and continues with others

### 6. Clear Documentation (Complex Resources)
**Principle**: Clearly document Phase 1 temporary nature, flatten helpers, and error handling.

**Rationale**:
- Complex resources have many helper functions
- Error handling strategy needs explanation
- Prevents confusion about purpose of code
- Provides clear migration path
- Helps future developers understand context
- Facilitates Phase 2 cleanup

**Implementation**:
- File-level comment explains Phase 1/Phase 2
- TODO comment marks code for removal
- Main function documentation explains parameters, return values, error handling
- Helper function documentation explains format and dependencies
- Attribute map format is documented with examples

---

## Architecture

### File Structure

```
genesyscloud/<resource_name>/
├── resource_genesyscloud_<resource_name>_schema.go          ← Stage 1
├── resource_genesyscloud_<resource_name>.go                 ← Stage 2
├── resource_genesyscloud_<resource_name>_utils.go           ← Stage 2
├── data_source_genesyscloud_<resource_name>.go              ← Stage 2
├── resource_genesyscloud_<resource_name>_test.go            ← Stage 3
├── data_source_genesyscloud_<resource_name>_test.go         ← Stage 3
├── genesyscloud_<resource_name>_init_test.go                ← Stage 3
├── resource_genesyscloud_<resource_name>_export_utils.go    ← Stage 4 (THIS)
└── genesyscloud_<resource_name>_proxy.go                    ← NOT MODIFIED
```

### Export Utilities File Components (Complex Resources)

```
┌─────────────────────────────────────────────────────────────────┐
│  resource_genesyscloud_<resource_name>_export_utils.go          │
├─────────────────────────────────────────────────────────────────┤
│  1. File-Level Documentation                                    │
│     - Phase 1 temporary explanation                             │
│     - TODO for Phase 2 removal                                  │
│     - Purpose and rationale                                     │
├─────────────────────────────────────────────────────────────────┤
│  2. Package Declaration and Imports                             │
│     - Package name                                              │
│     - SDK import                                                │
│     - Context, fmt, log, strconv imports                        │
│     - JSON import (if needed for API calls)                     │
├─────────────────────────────────────────────────────────────────┤
│  3. Main Attribute Mapping Function                             │
│     - build<ResourceName>Attributes()                           │
│     - Accepts context, SDK object pointer, proxy                │
│     - Returns flat attribute map and error                      │
│     - Maps basic attributes                                     │
│     - Calls flatten helpers for nested structures               │
│     - Makes additional API calls if needed                      │
│     - Handles errors appropriately                              │
├─────────────────────────────────────────────────────────────────┤
│  4. Flatten Helper Functions (10+ for complex resources)        │
│     - flattenSDK<NestedStruct>ToAttributes()                    │
│     - One helper per nested structure type                      │
│     - Convert nested objects to flat format                     │
│     - Handle arrays/sets with indexing                          │
│     - Extract dependency IDs                                    │
│     - Set count attributes                                      │
└─────────────────────────────────────────────────────────────────┘
```

### Integration with Stage 2 (Complex Resources)

```
┌─────────────────────────────────────────────────────────────────┐
│  GetAll<ResourceName>SDK() in Stage 2 Resource File             │
├─────────────────────────────────────────────────────────────────┤
│  1. Fetch all resources from API with expansions                │
│  2. Build initial export map with IDs and names                 │
│  3. For each resource:                                          │
│     a. Call build<ResourceName>Attributes(ctx, &resource, proxy)│
│     b. If error, log and skip resource (continue with others)   │
│     c. Add attribute map to ResourceMeta                        │
│  4. Return export map with attributes                           │
└─────────────────────────────────────────────────────────────────┘
```

---

## Component Design

### 1. File-Level Documentation (Complex Resources)

**Purpose**: Explain Phase 1 temporary nature and provide context for complex resource.

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

package user
```

---

### 2. Package Declaration and Imports (Complex Resources)

**Purpose**: Import necessary packages for SDK handling, API calls, and formatting.

**Design Pattern**:
```go
package <resource_name>

import (
    "context"
    "encoding/json"  // If needed for API response unmarshaling
    "fmt"
    "log"
    "strconv"

    "github.com/mypurecloud/platform-client-sdk-go/v176/platformclientv2"
)
```

**Key Points**:
- Context for API calls
- JSON for unmarshaling API responses (if needed)
- Fmt for string formatting
- Log for warnings
- Strconv for type conversions
- SDK import for types

---

### 3. Main Attribute Mapping Function (Complex Resources)

**Purpose**: Convert SDK resource object to flat attribute map with complete state.

**Design Pattern**:
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
func build<ResourceName>Attributes(ctx context.Context, <resource> *platformclientv2.<ResourceType>, proxy *<resource>Proxy) (map[string]string, error) {
    attributes := make(map[string]string)

    // Basic attributes
    if <resource>.Id != nil {
        attributes["id"] = *<resource>.Id
    }
    if <resource>.Name != nil {
        attributes["name"] = *<resource>.Name
    }

    // ⭐ CRITICAL: Dependency references (used by exporter for dependency resolution)
    if <resource>.Division != nil && <resource>.Division.Id != nil {
        attributes["division_id"] = *<resource>.Division.Id
    }

    // Complex nested attributes
    if <resource>.NestedStruct != nil {
        if err := flattenSDK<NestedStruct>ToAttributes(ctx, *<resource>.NestedStruct, attributes, proxy); err != nil {
            return nil, fmt.Errorf("failed to flatten <nested_struct>: %w", err)
        }
    }

    // Fetch additional data (separate API call, matching SDKv2)
    // Return error if fetch fails (matching SDKv2 behavior - skip resource)
    additionalData, _, err := proxy.getAdditionalData(ctx, *<resource>.Id)
    if err != nil {
        log.Printf("Failed to fetch additional data for <resource> %s: %v", *<resource>.Id, err)
        return nil, fmt.Errorf("failed to fetch additional data: %w", err)
    }
    if additionalData != nil {
        flattenSDK<AdditionalData>ToAttributes(additionalData, attributes)
    }

    return attributes, nil
}
```

**Example** (user - abbreviated):
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

    // ⭐ CRITICAL: Dependency references
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

    // Fetch voicemail policies (separate API call)
    voicemail, _, err := proxy.getVoicemailUserpoliciesById(ctx, *user.Id)
    if err != nil {
        log.Printf("Failed to fetch voicemail policies for user %s: %v", *user.Id, err)
        return nil, fmt.Errorf("failed to fetch voicemail policies: %w", err)
    }
    if voicemail != nil {
        flattenSDKVoicemailToAttributes(voicemail, attributes)
    }

    return attributes, nil
}
```

**Function Design Decisions**:

| Decision | Rationale |
|----------|-----------|
| Accept context | Enables API calls with cancellation/timeout |
| Accept pointer | SDK objects are pointers |
| Accept proxy | Provides access to API clients for additional calls |
| Return map and error | Allows caller to skip failed resources |
| Call flatten helpers | Modular design, easier to maintain |
| Make additional API calls | Fetch complete state matching SDKv2 |
| Return error for critical failures | Prevents incomplete export data |
| Log warnings for non-critical | Provides debugging info without failing |

---

### 4. Flatten Helper Functions (Complex Resources)

#### 4.1 Flatten Helper for Simple Nested Structure

**Purpose**: Convert simple nested structure to flat format.

**Design Pattern**:
```go
// flattenSDK<NestedStruct>ToAttributes converts SDK <nested_struct> to flat attribute map.
//
// Attribute Format (matching SDKv2 InstanceState):
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
```


#### 4.2 Flatten Helper for Array/Set with Dependencies

**Purpose**: Convert array/set to flat format with dependency extraction.

**Design Pattern**:
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
```

#### 4.3 Flatten Helper for Simple Array (No Dependencies)

**Purpose**: Convert simple string array to flat format.

**Design Pattern**:
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
```

#### 4.4 Flatten Helper with Additional API Call

**Purpose**: Flatten nested structure with additional API call for dependency ID.

**Design Pattern**:
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

**Example** (user addresses with extension pool lookup):
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
            if address.MediaType != nil {
                attributes[prefix+".media_type"] = *address.MediaType
            }
            if address.VarType != nil {
                attributes[prefix+".type"] = *address.VarType
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
            if address.VarType != nil {
                attributes[prefix+".type"] = *address.VarType
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

#### 4.5 Flatten Helper with Separate API Call

**Purpose**: Make separate API call to fetch data and flatten to attributes.

**Design Pattern**:
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

**Example** (user routing utilization - abbreviated):
```go
// flattenSDKRoutingUtilizationToAttributes fetches and converts routing utilization to flat map.
// Makes separate API call to get utilization settings (matching SDKv2 behavior).
//
// Attribute Format:
//   - "routing_utilization.#" = "1"
//   - "routing_utilization.0.call.#" = "1"
//   - "routing_utilization.0.call.0.maximum_capacity" = "3"
//   - "routing_utilization.0.call.0.include_non_acd" = "false"
//   - ... similar for callback, message, email, chat
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
```

---

## Attribute Flattening Patterns (Complex Resources)

### Pattern 1: Simple Nested Structure (1-Level)

**Format**: `<nested>.0.attribute`

**Example**:
```
employer_info.# = "1"
employer_info.0.official_name = "John Smith"
employer_info.0.employee_id = "EMP-12345"
```

**Code**:
```go
prefix := "employer_info.0"
attributes[prefix+".official_name"] = *employerInfo.OfficialName
attributes["employer_info.#"] = "1"
```

### Pattern 2: Array/Set (1-Level)

**Format**: `<array>.#` and `<array>.0`, `<array>.1`, etc.

**Example**:
```
profile_skills.# = "2"
profile_skills.0 = "Java"
profile_skills.1 = "Python"
```

**Code**:
```go
for i, skill := range skills {
    attributes[fmt.Sprintf("profile_skills.%d", i)] = skill
}
attributes["profile_skills.#"] = strconv.Itoa(len(skills))
```

### Pattern 3: Nested Array/Set (2-Level)

**Format**: `<nested>.0.<array>.#` and `<nested>.0.<array>.0.attribute`

**Example**:
```
addresses.# = "1"
addresses.0.phone_numbers.# = "2"
addresses.0.phone_numbers.0.number = "+13175559001"
addresses.0.phone_numbers.0.extension = "8701"
addresses.0.phone_numbers.1.number = "+13175559002"
```

**Code**:
```go
for i, phone := range phoneNumbers {
    prefix := fmt.Sprintf("addresses.0.phone_numbers.%d", i)
    attributes[prefix+".number"] = *phone.Address
    attributes[prefix+".extension"] = *phone.Extension
}
attributes["addresses.#"] = "1"
attributes["addresses.0.phone_numbers.#"] = strconv.Itoa(len(phoneNumbers))
```

### Pattern 4: Deep Nested Structure (3-Level)

**Format**: `<level1>.0.<level2>.0.<level3>.#` and `<level1>.0.<level2>.0.<level3>.0.attribute`

**Example**:
```
routing_utilization.# = "1"
routing_utilization.0.call.# = "1"
routing_utilization.0.call.0.maximum_capacity = "3"
routing_utilization.0.call.0.interruptible_media_types.# = "2"
routing_utilization.0.call.0.interruptible_media_types.0 = "email"
routing_utilization.0.call.0.interruptible_media_types.1 = "chat"
```

**Code**:
```go
prefix := "routing_utilization.0.call.0"
attributes[prefix+".maximum_capacity"] = strconv.Itoa(int(mediaSettings.MaximumCapacity))

for i, mediaType := range mediaSettings.InterruptableMediaTypes {
    attributes[fmt.Sprintf("%s.interruptible_media_types.%d", prefix, i)] = mediaType
}
attributes[prefix+".interruptible_media_types.#"] = strconv.Itoa(len(mediaSettings.InterruptableMediaTypes))
attributes["routing_utilization.0.call.#"] = "1"
attributes["routing_utilization.#"] = "1"
```

### Pattern 5: Multiple Arrays in Same Nested Structure

**Format**: Multiple arrays at same level with separate indexing

**Example**:
```
addresses.# = "1"
addresses.0.phone_numbers.# = "2"
addresses.0.phone_numbers.0.number = "+13175559001"
addresses.0.phone_numbers.1.number = "+13175559002"
addresses.0.other_emails.# = "1"
addresses.0.other_emails.0.address = "alt@example.com"
```

**Code**:
```go
phoneIndex := 0
emailIndex := 0

for _, address := range addresses {
    switch *address.MediaType {
    case "PHONE":
        prefix := fmt.Sprintf("addresses.0.phone_numbers.%d", phoneIndex)
        attributes[prefix+".number"] = *address.Address
        phoneIndex++
    case "EMAIL":
        prefix := fmt.Sprintf("addresses.0.other_emails.%d", emailIndex)
        attributes[prefix+".address"] = *address.Address
        emailIndex++
    }
}

attributes["addresses.#"] = "1"
attributes["addresses.0.phone_numbers.#"] = strconv.Itoa(phoneIndex)
attributes["addresses.0.other_emails.#"] = strconv.Itoa(emailIndex)
```

---

## Error Handling Design (Complex Resources)

### Error Handling Strategy

```
┌─────────────────────────────────────────────────────────┐
│  Error Handling Decision Tree                           │
├─────────────────────────────────────────────────────────┤
│  Is the operation critical?                             │
│  ├─ YES: Return error (caller skips resource)           │
│  │   Examples:                                          │
│  │   - Voicemail policies fetch fails                   │
│  │   - Routing utilization fetch fails                  │
│  │   - Nested structure flatten fails                   │
│  │                                                       │
│  └─ NO: Log warning and continue                        │
│      Examples:                                          │
│      - Extension pool not found                         │
│      - Optional nested structure is nil                 │
│      - Non-critical attribute missing                   │
└─────────────────────────────────────────────────────────┘
```

### Critical Error Pattern

**When to use**: Operation is required for complete export data.

**Pattern**:
```go
data, _, err := proxy.getCriticalData(ctx, *resource.Id)
if err != nil {
    log.Printf("Failed to fetch critical data for resource %s: %v", *resource.Id, err)
    return nil, fmt.Errorf("failed to fetch critical data: %w", err)
}
```

**Result**: Function returns error, caller skips resource and continues with others.

### Non-Critical Warning Pattern

**When to use**: Operation is optional or has fallback behavior.

**Pattern**:
```go
optionalData := fetchOptionalData(ctx, *resource.Id, proxy)
if optionalData == "" {
    log.Printf("Warning: Optional data not found for resource %s", *resource.Id)
    // Continue without error
}
```

**Result**: Warning logged, function continues normally.

### Error Context Pattern

**When to use**: Provide helpful context for debugging.

**Pattern**:
```go
if err := flattenNestedStruct(ctx, nestedStruct, attributes, proxy); err != nil {
    return nil, fmt.Errorf("failed to flatten nested_struct: %w", err)
}
```

**Result**: Error includes context about which operation failed.

---

## Integration with Exporter (Complex Resources)

### Export Flow

```
1. User runs: terraform export genesyscloud_user
   ↓
2. Exporter calls: GetAllUsersSDK(ctx)
   ↓
3. GetAllUsersSDK() fetches resources from API with expansions
   ↓
4. For each resource:
   a. Call buildUserAttributes(ctx, &user, proxy)
   b. If error:
      - Log error
      - Skip user
      - Continue with others
   c. If success:
      - Get flat attribute map
      - Add to ResourceMeta.ExportAttributes
   ↓
5. Exporter receives ResourceIDMetaMap with attributes
   ↓
6. Exporter resolves dependencies using attribute map
   ↓
7. Exporter generates HCL with correct references
   ↓
8. Export complete
```

### Dependency Resolution (Complex Resources)

**Simple Dependency**:
```
1. Read: divisionId := attributes["division_id"]
2. Look up: Find division with that ID
3. Generate: division_id = genesyscloud_auth_division.division_label.id
```

**Nested Dependency**:
```
1. Read: skillId := attributes["routing_skills.0.skill_id"]
2. Look up: Find skill with that ID
3. Generate: skill_id = genesyscloud_routing_skill.skill_label.id
```

**Deep Nested Dependency**:
```
1. Read: poolId := attributes["addresses.0.phone_numbers.0.extension_pool_id"]
2. Look up: Find extension pool with that ID
3. Generate: extension_pool_id = genesyscloud_telephony_providers_edges_extension_pool.pool_label.id
```

---

## Design Patterns and Best Practices (Complex Resources)

### Pattern 1: Modular Flatten Helpers

**Pattern**:
```go
// Main function calls helpers
if resource.NestedStruct1 != nil {
    flattenSDKNestedStruct1ToAttributes(*resource.NestedStruct1, attributes)
}
if resource.NestedStruct2 != nil {
    flattenSDKNestedStruct2ToAttributes(*resource.NestedStruct2, attributes)
}
```

**Why**:
- Improves code organization
- Easier to test individual helpers
- Reduces complexity of main function
- Reusable helpers

### Pattern 2: Count Attributes

**Pattern**:
```go
// Always set count attribute for arrays/sets
attributes["<array>.#"] = strconv.Itoa(len(array))

// Always set count for nested structures
attributes["<nested>.#"] = "1"
```

**Why**:
- Matches SDKv2 InstanceState format
- Exporter uses counts for iteration
- Required for proper HCL generation

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

### Pattern 4: Error Wrapping

**Pattern**:
```go
if err := flattenNestedStruct(ctx, nestedStruct, attributes, proxy); err != nil {
    return nil, fmt.Errorf("failed to flatten nested_struct: %w", err)
}
```

**Why**:
- Provides error context
- Enables error chain inspection
- Helps with debugging
- Clear error messages

### Pattern 5: Nil Safety

**Pattern**:
```go
if resource.OptionalAttr != nil {
    attributes["optional_attr"] = *resource.OptionalAttr
}

// For nested objects, check both parent and child
if resource.Nested != nil && resource.Nested.Id != nil {
    attributes["nested_id"] = *resource.Nested.Id
}
```

**Why**:
- Prevents panics
- Handles optional attributes
- Matches SDKv2 behavior

### Pattern 6: Separate Indexing for Multiple Arrays

**Pattern**:
```go
phoneIndex := 0
emailIndex := 0

for _, item := range items {
    if item.Type == "PHONE" {
        prefix := fmt.Sprintf("addresses.0.phone_numbers.%d", phoneIndex)
        // ... set attributes
        phoneIndex++
    } else if item.Type == "EMAIL" {
        prefix := fmt.Sprintf("addresses.0.other_emails.%d", emailIndex)
        // ... set attributes
        emailIndex++
    }
}

attributes["addresses.0.phone_numbers.#"] = strconv.Itoa(phoneIndex)
attributes["addresses.0.other_emails.#"] = strconv.Itoa(emailIndex)
```

**Why**:
- Correct indexing for each array
- Matches SDKv2 format
- Prevents index conflicts

---

## Migration Considerations (Complex Resources)

### Cleanup Checklist for Phase 2

When removing export utilities in Phase 2:

- [ ] Delete `resource_genesyscloud_<resource_name>_export_utils.go` file
- [ ] Remove `GetAll<ResourceName>SDK()` function from resource file
- [ ] Remove `build<ResourceName>Attributes()` calls from `GetAll<ResourceName>SDK()`
- [ ] Remove all flatten helper functions
- [ ] Update exporter to use `GetAll<ResourceName>()` (Framework version)
- [ ] Remove SDK diagnostics import if no longer needed
- [ ] Update exporter to work with Framework types
- [ ] Test export functionality with Framework types
- [ ] Verify dependency resolution still works
- [ ] Verify nested structure handling

### Common Pitfalls (Complex Resources)

#### Pitfall 1: Missing Flatten Helper
**Problem**: Nested structure not flattened, attributes missing.
**Solution**: Create flatten helper for each nested structure type.

#### Pitfall 2: Incorrect Indexing
**Problem**: Array elements have wrong indices, export fails.
**Solution**: Use separate index variables for each array, increment correctly.

#### Pitfall 3: Missing Count Attribute
**Problem**: Exporter can't iterate over array, export fails.
**Solution**: Always set `<array>.#` count attribute.

#### Pitfall 4: Missing Dependency Reference
**Problem**: Dependency not resolved, export fails.
**Solution**: Verify all RefAttrs from Stage 1 are included, mark with CRITICAL.

#### Pitfall 5: Incorrect Attribute Path
**Problem**: Nested attribute has wrong path, export fails.
**Solution**: Use correct dot notation and indexing: `<nested>.0.<sub_nested>.0.attribute`.

#### Pitfall 6: Missing Error Handling
**Problem**: API call fails, incomplete data exported.
**Solution**: Return error for critical failures, log warnings for non-critical.

#### Pitfall 7: Missing Nil Check
**Problem**: Panic when dereferencing nil pointer.
**Solution**: Always check nil before dereferencing, especially for nested objects.

#### Pitfall 8: Wrong Error Handling Strategy
**Problem**: Non-critical failure causes entire export to fail.
**Solution**: Return errors only for critical failures, log warnings for non-critical.

---

## Summary

### Key Design Decisions (Complex Resources)

1. **Modular Flatten Helpers**: One helper per nested structure type
2. **Additional API Calls**: Fetch complete state matching SDKv2
3. **Comprehensive Dependency Handling**: Simple, nested, and deep nested dependencies
4. **Error Handling Strategy**: Return errors for critical, log warnings for non-critical
5. **Flat Attribute Map**: Convert nested structures with proper indexing
6. **Clear Documentation**: Explain Phase 1/Phase 2, helpers, error handling

### File Structure (Complex Resources)

```
Export Utilities File:
├── File-level documentation (Phase 1 temporary, TODO)
├── Package declaration and imports (SDK, context, fmt, log, strconv, json)
├── Main attribute mapping function
│   ├── Accept context, SDK object pointer, proxy
│   ├── Return map and error
│   ├── Map basic attributes
│   ├── Call flatten helpers
│   ├── Make additional API calls
│   └── Handle errors
└── Flatten helper functions (10+ for complex resources)
    ├── Simple nested structure helpers
    ├── Array/set helpers
    ├── Helpers with additional API calls
    └── Helpers with separate API calls
```

### Integration Points

- Called from `GetAll<ResourceName>SDK()` in Stage 2
- Errors handled (skip resource on error)
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

- **Reference Implementation**: `genesyscloud/user/resource_genesyscloud_user_export_utils.go`
- **Simple Resource Reference**: `prompts/pf_simple_resource_migration/Stage4/design.md`
- **Stage 2 GetAll Functions**: `genesyscloud/user/resource_genesyscloud_user.go`
- **Stage 2 Utils Functions**: `genesyscloud/user/resource_genesyscloud_user_utils.go`
- **Stage 1 Exporter Config**: `genesyscloud/user/resource_genesyscloud_user_schema.go`

