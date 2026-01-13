# Solution 4: Hybrid Computed Approach - Implementation Details

## Overview

This document provides detailed implementation analysis for Solution 4 (Hybrid Computed Approach) to help finalize the decision on which solution to pursue for the extension pool migration issue.

## Solution Summary

Keep `extension_pool_id` as Optional for user input but add sophisticated computed logic to normalize/stabilize values using plan modifiers. This approach attempts to provide the best of both worlds: maintaining user control while minimizing Set identity issues through intelligent state management.

## Required Code Changes

### 1. Schema Changes (`resource_genesyscloud_user_schema.go`)

**Complexity: Medium**

#### Enhance extension_pool_id with Advanced Plan Modifiers

**Current Schema:**
```go
"extension_pool_id": schema.StringAttribute{
    Description:   "Id of the extension pool which contains this extension.",
    Optional:      true,
    PlanModifiers: []planmodifier.String{phoneplan.NullIfEmpty{}},
},
```

**Enhanced Schema:**
```go
"extension_pool_id": schema.StringAttribute{
    Description: "Id of the extension pool which contains this extension. " +
        "When specified, the extension will be assigned to this pool. " +
        "When omitted, the extension will be auto-assigned based on pool ranges. " +
        "Changes to pool assignments may cause phone number Set element replacement.",
    Optional: true,
    Computed: true, // Allow computed values for normalization
    PlanModifiers: []planmodifier.String{
        phoneplan.NullIfEmpty{},
        phoneplan.ExtensionPoolNormalizer{}, // New custom plan modifier
        phoneplan.UseStateForUnknownIfMatches{}, // New custom plan modifier
    },
},
```

**Estimated Effort: 2-3 hours**

### 2. Custom Plan Modifier Development

**Complexity: High**

#### Create ExtensionPoolNormalizer Plan Modifier

```go
// File: genesyscloud/util/phoneplan/extension_pool_normalizer.go
package phoneplan

import (
    "context"
    "fmt"
    "log"
    "strconv"
    "strings"

    "github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
    "github.com/hashicorp/terraform-plugin-framework/types"
)

type ExtensionPoolNormalizer struct{}

func (m ExtensionPoolNormalizer) Description(_ context.Context) string {
    return "Normalizes extension pool assignments to reduce Set identity changes"
}

func (m ExtensionPoolNormalizer) MarkdownDescription(_ context.Context) string {
    return m.Description(context.Background())
}

func (m ExtensionPoolNormalizer) PlanModifyString(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
    // Skip if config value is null or unknown
    if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
        return
    }

    // Get the extension value from the same phone number object
    extension := getExtensionFromPhoneNumber(req.Path, req.Plan)
    if extension == "" {
        return
    }

    configPoolId := req.ConfigValue.ValueString()
    
    // Validate that the extension is actually in the specified pool range
    if !isExtensionInPoolRange(ctx, extension, configPoolId) {
        resp.Diagnostics.AddError(
            "Extension Pool Validation Failed",
            fmt.Sprintf("Extension %s is not within the range of pool %s", extension, configPoolId),
        )
        return
    }

    // If state has a different pool ID but extension hasn't changed,
    // check if we should preserve state to avoid unnecessary diffs
    if !req.StateValue.IsNull() && !req.StateValue.IsUnknown() {
        statePoolId := req.StateValue.ValueString()
        
        // If both pools contain the extension, prefer state to minimize diffs
        if statePoolId != configPoolId && isExtensionInPoolRange(ctx, extension, statePoolId) {
            log.Printf("[DEBUG] Extension %s is valid in both pools, preserving state pool %s to minimize diffs", extension, statePoolId)
            resp.PlanValue = req.StateValue
            return
        }
    }

    // Use config value as-is
    resp.PlanValue = req.ConfigValue
}

func getExtensionFromPhoneNumber(path path.Path, plan tfsdk.Plan) string {
    // Navigate to the extension field in the same phone number object
    // This is complex path manipulation to get sibling attribute
    // Implementation would need to parse the path and extract extension value
    // Simplified for this example
    return ""
}

func isExtensionInPoolRange(ctx context.Context, extension, poolId string) bool {
    // This would need access to the provider's API client to fetch pool details
    // Complex implementation required to validate extension against pool range
    // Simplified for this example
    return true
}
```

#### Create UseStateForUnknownIfMatches Plan Modifier

```go
// File: genesyscloud/util/phoneplan/use_state_for_unknown_if_matches.go
package phoneplan

import (
    "context"
    "github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
    "github.com/hashicorp/terraform-plugin-framework/types"
)

type UseStateForUnknownIfMatches struct{}

func (m UseStateForUnknownIfMatches) Description(_ context.Context) string {
    return "Uses state value when plan is unknown but state matches expected pool assignment"
}

func (m UseStateForUnknownIfMatches) MarkdownDescription(_ context.Context) string {
    return m.Description(context.Background())
}

func (m UseStateForUnknownIfMatches) PlanModifyString(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
    // Only act when plan value is unknown (computed)
    if !req.PlanValue.IsUnknown() {
        return
    }

    // If state has a value and it's still valid, use it
    if !req.StateValue.IsNull() && !req.StateValue.IsUnknown() {
        extension := getExtensionFromPhoneNumber(req.Path, req.Plan)
        statePoolId := req.StateValue.ValueString()
        
        // If state pool still contains the extension, preserve it
        if isExtensionInPoolRange(ctx, extension, statePoolId) {
            resp.PlanValue = req.StateValue
            return
        }
    }

    // Otherwise, let it remain unknown (will be computed during apply)
}
```

**Estimated Effort: 12-15 hours**

### 3. Model Updates (`resource_genesyscloud_user.go`)

**Complexity: Low**

**No changes required** - the existing model structure is already correct:

```go
type PhoneNumberModel struct {
    Number          types.String `tfsdk:"number"`
    MediaType       types.String `tfsdk:"media_type"`
    Type            types.String `tfsdk:"type"`
    Extension       types.String `tfsdk:"extension"`
    ExtensionPoolId types.String `tfsdk:"extension_pool_id"`
}
```

**Estimated Effort: 0 hours**

### 4. State Management Updates (`resource_genesyscloud_user_utils.go`)

**Complexity: Medium**

#### Enhance flattenUserAddresses Function

**Current Issue (line 738):**
```go
"extension_pool_id": types.StringNull(), // <- always null in state -- TODO
```

**Enhanced Implementation:**
```go
func flattenUserAddresses(ctx context.Context, addresses *[]platformclientv2.Contact, proxy *userProxy) (types.List, pfdiag.Diagnostics) {
    // ... existing code ...
    
    // Case 2: Extension == Display → true internal extension (extension mapped to pool)
    if address.Extension != nil && address.Display != nil && *address.Extension == *address.Display {
        extensionNum := strings.Trim(*address.Extension, "()")
        if extensionNum != "" {
            phoneNumber["extension"] = types.StringValue(extensionNum)
        }
        
        // ENHANCED: Intelligent pool ID handling
        poolId := fetchExtensionPoolId(ctx, extensionNum, proxy)
        if poolId != "" {
            phoneNumber["extension_pool_id"] = types.StringValue(poolId)
        } else {
            // If no pool found, check if this is a planned assignment
            // that hasn't been processed yet (during create/update)
            phoneNumber["extension_pool_id"] = types.StringUnknown()
        }
        
        phoneNumber["number"] = types.StringNull()
    }
    
    // ... rest of existing code unchanged ...
}
```

#### Add Pool Assignment Validation

```go
func validateAndAssignExtensionPool(ctx context.Context, extension, requestedPoolId string, proxy *userProxy) (string, pfdiag.Diagnostics) {
    var diagnostics pfdiag.Diagnostics
    
    if requestedPoolId == "" {
        // Auto-assign based on extension range
        actualPoolId := fetchExtensionPoolId(ctx, extension, proxy)
        return actualPoolId, diagnostics
    }
    
    // Validate requested pool assignment
    pool, _, err := proxy.getTelephonyExtensionPoolById(ctx, requestedPoolId)
    if err != nil {
        diagnostics.AddError("Extension Pool Not Found", 
            fmt.Sprintf("Requested extension pool %s not found: %v", requestedPoolId, err))
        return "", diagnostics
    }
    
    // Validate extension is in pool range
    extNum, err := strconv.Atoi(extension)
    if err != nil {
        diagnostics.AddError("Invalid Extension", 
            fmt.Sprintf("Extension %s is not a valid number: %v", extension, err))
        return "", diagnostics
    }
    
    startNum, _ := strconv.Atoi(*pool.StartNumber)
    endNum, _ := strconv.Atoi(*pool.EndNumber)
    
    if extNum < startNum || extNum > endNum {
        diagnostics.AddError("Extension Out of Range", 
            fmt.Sprintf("Extension %s is not within pool range %s-%s", extension, *pool.StartNumber, *pool.EndNumber))
        return "", diagnostics
    }
    
    return requestedPoolId, diagnostics
}
```

**Estimated Effort: 6-8 hours**

### 5. Request Building Updates (`resource_genesyscloud_user_utils.go`)

**Complexity: Medium**

#### Enhance buildSdkPhoneNumbers Function

**Current Issue:**
The function completely ignores `extension_pool_id` from the configuration.

**Enhanced Implementation:**
```go
func buildSdkPhoneNumbers(configPhoneNumbers types.Set, proxy *userProxy) ([]platformclientv2.Contact, pfdiag.Diagnostics) {
    var diagnostics pfdiag.Diagnostics
    
    // ... existing code until extension processing ...
    
    // Process extension with intelligent pool assignment
    if !phone.Extension.IsNull() && !phone.Extension.IsUnknown() {
        phoneExt := phone.Extension.ValueString()
        if phoneExt != "" {
            contact.Extension = &phoneExt
            
            // Handle extension pool assignment
            var requestedPoolId string
            if !phone.ExtensionPoolId.IsNull() && !phone.ExtensionPoolId.IsUnknown() {
                requestedPoolId = phone.ExtensionPoolId.ValueString()
            }
            
            // Validate and assign pool
            actualPoolId, poolDiags := validateAndAssignExtensionPool(
                context.Background(), phoneExt, requestedPoolId, proxy)
            diagnostics.Append(poolDiags...)
            
            if poolDiags.HasError() {
                continue // Skip this phone number if pool assignment failed
            }
            
            // Store the actual pool assignment for state correlation
            // Note: This might require additional API calls or metadata storage
            if actualPoolId != "" {
                log.Printf("[DEBUG] Extension %s assigned to pool %s", phoneExt, actualPoolId)
                // The actual pool ID will be retrieved during read/refresh
            }
        }
    }
    
    sdkContacts[i] = contact
}
```

**Estimated Effort: 5-6 hours**

### 6. Advanced Set Identity Management

**Complexity: Very High**

#### Create Set Element Correlation Logic

The most complex part of Solution 4 is ensuring Set elements can be correlated correctly despite pool ID changes:

```go
// File: genesyscloud/user/set_correlation_helper.go
package user

import (
    "context"
    "crypto/md5"
    "fmt"
    "sort"
    "strings"
    
    "github.com/hashicorp/terraform-plugin-framework/attr"
    "github.com/hashicorp/terraform-plugin-framework/types"
)

type PhoneNumberIdentity struct {
    MediaType string
    Type      string
    Number    string
    Extension string
}

func (p PhoneNumberIdentity) StableHash() string {
    // Create a stable hash that excludes extension_pool_id
    parts := []string{
        p.MediaType,
        p.Type,
        p.Number,
        p.Extension,
    }
    
    // Sort to ensure consistent ordering
    sort.Strings(parts)
    combined := strings.Join(parts, "|")
    
    hash := md5.Sum([]byte(combined))
    return fmt.Sprintf("%x", hash)
}

func correlatePhoneNumberSets(planSet, stateSet types.Set) (map[string]attr.Value, map[string]attr.Value, error) {
    planElements := make(map[string]attr.Value)
    stateElements := make(map[string]attr.Value)
    
    // Extract plan elements with stable identities
    planList := planSet.Elements()
    for _, element := range planList {
        identity := extractPhoneNumberIdentity(element)
        stableHash := identity.StableHash()
        planElements[stableHash] = element
    }
    
    // Extract state elements with stable identities
    stateList := stateSet.Elements()
    for _, element := range stateList {
        identity := extractPhoneNumberIdentity(element)
        stableHash := identity.StableHash()
        stateElements[stableHash] = element
    }
    
    return planElements, stateElements, nil
}

func extractPhoneNumberIdentity(element attr.Value) PhoneNumberIdentity {
    // Complex logic to extract identity fields from Set element
    // This would need to handle the object structure and extract relevant fields
    // Simplified for this example
    return PhoneNumberIdentity{}
}

// Custom Set plan modifier that uses stable correlation
type StablePhoneNumberSetModifier struct{}

func (m StablePhoneNumberSetModifier) PlanModifySet(ctx context.Context, req planmodifier.SetRequest, resp *planmodifier.SetResponse) {
    // Implement sophisticated Set correlation logic
    // This would be extremely complex and error-prone
}
```

**Estimated Effort: 20-25 hours**

### 7. Test Updates (`resource_genesyscloud_user_test.go`)

**Complexity: High**

#### Test Plan Modifier Behavior

```go
func TestAccResourceUser_extensionPoolPlanModifiers(t *testing.T) {
    // Test that plan modifiers work correctly
    // Test pool assignment normalization
    // Test state preservation when appropriate
    // Test validation errors
    // Test Set element correlation
    
    resource.Test(t, resource.TestCase{
        Steps: []resource.TestStep{
            {
                Config: generateUserWithExtensionPool("user1", "pool1"),
                Check: resource.ComposeTestCheckFunc(
                    resource.TestCheckResourceAttr("genesyscloud_user.user1", "addresses.0.phone_numbers.0.extension", "8501"),
                    resource.TestCheckResourceAttrPair("genesyscloud_user.user1", "addresses.0.phone_numbers.0.extension_pool_id", "genesyscloud_telephony_providers_edges_extension_pool.pool1", "id"),
                ),
            },
            {
                // Test pool change behavior
                Config: generateUserWithExtensionPool("user1", "pool2"),
                Check: resource.ComposeTestCheckFunc(
                    // Verify that Set element correlation works
                    resource.TestCheckResourceAttr("genesyscloud_user.user1", "addresses.0.phone_numbers.0.extension", "8501"),
                    resource.TestCheckResourceAttrPair("genesyscloud_user.user1", "addresses.0.phone_numbers.0.extension_pool_id", "genesyscloud_telephony_providers_edges_extension_pool.pool2", "id"),
                ),
            },
            {
                // Test auto-assignment when pool ID is removed
                Config: generateUserWithExtensionNoPool("user1"),
                Check: resource.ComposeTestCheckFunc(
                    resource.TestCheckResourceAttr("genesyscloud_user.user1", "addresses.0.phone_numbers.0.extension", "8501"),
                    resource.TestCheckResourceAttrSet("genesyscloud_user.user1", "addresses.0.phone_numbers.0.extension_pool_id"),
                ),
            },
        },
    })
}

func TestAccResourceUser_extensionPoolValidation(t *testing.T) {
    // Test validation errors
    resource.Test(t, resource.TestCase{
        Steps: []resource.TestStep{
            {
                Config:      generateUserWithInvalidExtensionPool(),
                ExpectError: regexp.MustCompile("Extension .* is not within pool range"),
            },
        },
    })
}
```

#### Test Edge Cases

```go
func TestAccResourceUser_extensionPoolEdgeCases(t *testing.T) {
    // Test overlapping pool ranges
    // Test pool deletion scenarios
    // Test concurrent pool assignments
    // Test plan modifier error handling
}
```

**Estimated Effort: 12-15 hours**

### 8. Documentation Updates

**Complexity: Medium**

#### Document Plan Modifier Behavior

```markdown
## Extension Pool Assignment

The `extension_pool_id` field supports both explicit assignment and automatic assignment:

### Explicit Assignment
```hcl
addresses {
  phone_numbers {
    extension         = "8501"
    extension_pool_id = genesyscloud_telephony_providers_edges_extension_pool.pool.id
  }
}
```

### Automatic Assignment
```hcl
addresses {
  phone_numbers {
    extension = "8501"
    # extension_pool_id omitted - will be auto-assigned based on pool ranges
  }
}
```

### Important Behaviors

- **Plan Modifiers**: The provider uses intelligent plan modifiers to minimize unnecessary diffs when pool assignments change
- **Validation**: Extensions are validated against pool ranges during planning
- **State Preservation**: When possible, existing pool assignments are preserved to reduce plan churn
- **Set Identity**: Pool assignment changes may cause phone number Set elements to be replaced rather than updated in-place

### Advanced Configuration

For complex scenarios with overlapping pool ranges, you can use explicit assignment to ensure deterministic behavior:

```hcl
# Multiple pools with overlapping ranges
resource "genesyscloud_telephony_providers_edges_extension_pool" "sales" {
  start_number = "1000"
  end_number   = "1999"
}

resource "genesyscloud_telephony_providers_edges_extension_pool" "support" {
  start_number = "1500"  # Overlaps with sales pool
  end_number   = "2000"
}

# Explicit assignment ensures extension goes to intended pool
resource "genesyscloud_user" "sales_user" {
  addresses {
    phone_numbers {
      extension         = "1750"  # Could go to either pool
      extension_pool_id = genesyscloud_telephony_providers_edges_extension_pool.sales.id
    }
  }
}
```
```

**Estimated Effort: 4-5 hours**

### 9. Error Handling and Edge Cases

**Complexity: High**

#### Handle Complex Scenarios

```go
func handleExtensionPoolConflicts(ctx context.Context, extension string, requestedPoolId string, proxy *userProxy) pfdiag.Diagnostics {
    var diagnostics pfdiag.Diagnostics
    
    // Check for conflicts with existing assignments
    existingAssignment := findExistingExtensionAssignment(ctx, extension, proxy)
    if existingAssignment != nil && existingAssignment.PoolId != requestedPoolId {
        diagnostics.AddWarning("Extension Pool Conflict",
            fmt.Sprintf("Extension %s is already assigned to pool %s, reassigning to %s", 
                extension, existingAssignment.PoolId, requestedPoolId))
    }
    
    // Check for overlapping pool ranges
    overlappingPools := findOverlappingPools(ctx, extension, proxy)
    if len(overlappingPools) > 1 {
        diagnostics.AddWarning("Multiple Pool Matches",
            fmt.Sprintf("Extension %s matches multiple pools: %v. Using explicit assignment.", 
                extension, overlappingPools))
    }
    
    return diagnostics
}

func recoverFromPoolAssignmentFailure(ctx context.Context, extension string, proxy *userProxy) (string, error) {
    // Attempt to recover by finding any available pool
    availablePools := findAvailablePoolsForExtension(ctx, extension, proxy)
    if len(availablePools) > 0 {
        return availablePools[0], nil
    }
    
    return "", fmt.Errorf("no available pools found for extension %s", extension)
}
```

**Estimated Effort: 8-10 hours**

## Total Implementation Effort

### Summary:
- **Schema Changes**: 3 hours (enhanced plan modifiers)
- **Custom Plan Modifiers**: 15 hours (complex logic development)
- **Model Updates**: 0 hours (no changes needed)
- **State Management**: 8 hours (enhanced flatten function)
- **Request Building**: 6 hours (enhanced build function)
- **Set Identity Management**: 25 hours (very complex correlation logic)
- **Test Updates**: 15 hours (comprehensive test scenarios)
- **Documentation**: 5 hours (detailed behavior documentation)
- **Error Handling**: 10 hours (complex edge case handling)
- **Total: 87 hours (~2.2 weeks)**

## Implementation Risks

### Very High Risk Areas:
1. **Set Element Correlation**: Extremely complex logic to correlate Set elements with changing pool IDs
2. **Plan Modifier Complexity**: Custom plan modifiers with access to API data are very difficult to implement correctly
3. **State Management Complexity**: Sophisticated logic to determine when to preserve vs. update state
4. **Edge Case Handling**: Numerous complex scenarios with overlapping pools, conflicts, and failures

### High Risk Areas:
1. **API Integration**: Plan modifiers need access to provider API, which is architecturally challenging
2. **Performance**: Multiple API calls during planning phase could impact performance
3. **Debugging**: Complex plan modifier logic is very difficult to debug and troubleshoot
4. **Testing**: Comprehensive testing of all plan modifier scenarios is extremely challenging

### Medium Risk Areas:
1. **Documentation**: Complex behavior is difficult to document clearly for users
2. **Maintenance**: Sophisticated plan modifier logic is hard to maintain and extend
3. **User Understanding**: Users may not understand the complex behaviors and edge cases

## Breaking Change Impact

**Zero Breaking Changes:**

Users can keep their existing configurations exactly as they are:

```hcl
# This configuration works before and after implementation
addresses {
  phone_numbers {
    extension         = "8501"
    extension_pool_id = genesyscloud_telephony_providers_edges_extension_pool.pool.id
  }
}
```

However, users may experience:
- **Different diff patterns** due to plan modifier behavior
- **More complex error messages** when validation fails
- **Unexpected state preservation** in some scenarios

## Architectural Concerns

### Plan Modifier Limitations:
1. **No API Access**: Plan modifiers typically don't have access to provider API clients
2. **Context Limitations**: Limited context about other resources and their states
3. **Performance Impact**: API calls during planning can significantly slow down operations
4. **Error Handling**: Limited ability to handle API errors gracefully during planning

### Set Identity Challenges:
1. **Framework Limitations**: Plugin Framework doesn't provide hooks for custom Set identity logic
2. **Correlation Complexity**: Matching Set elements across plan/state boundaries is extremely complex
3. **Race Conditions**: Concurrent modifications could cause correlation failures
4. **Debugging Difficulty**: Set correlation issues are very hard to diagnose and fix

## Comparison with Other Solutions

| Aspect | Solution 1 | Solution 2 | Solution 3 | Solution 4 |
|--------|------------|------------|------------|------------|
| **Development Time** | 46-91 hours | 22 hours | 27 hours | 87 hours |
| **Breaking Changes** | Yes (major) | No | Yes (major) | No |
| **User Impact** | High | None | High | Low |
| **Implementation Risk** | High | Low | Medium | Very High |
| **Code Complexity** | Very High | Low | Low | Extremely High |
| **Functionality** | Full | Full | Reduced | Full+ |
| **Maintainability** | Medium | High | High | Very Low |

## Pros and Cons Summary

### Pros:
- ✅ **No breaking changes** for users
- ✅ **Maintains full functionality** with explicit pool control
- ✅ **Intelligent diff reduction** through plan modifiers
- ✅ **Sophisticated validation** and error handling
- ✅ **Flexible assignment patterns** (explicit + automatic)
- ✅ **Advanced edge case handling**

### Cons:
- ❌ **Extremely high implementation complexity** (87 hours)
- ❌ **Very high risk** of bugs and edge case failures
- ❌ **Architectural challenges** with plan modifier API access
- ❌ **Performance concerns** with API calls during planning
- ❌ **Very difficult to debug** and troubleshoot
- ❌ **Hard to maintain** and extend over time
- ❌ **Complex user experience** with sophisticated behaviors

## Recommendation

**Solution 4 is NOT recommended** for the following reasons:

### **Critical Issues:**
1. **Extreme Complexity**: 87 hours of development with very high risk of bugs
2. **Architectural Problems**: Plan modifiers with API access are very difficult to implement correctly
3. **Maintenance Burden**: Extremely complex code that will be hard to maintain and debug
4. **Diminishing Returns**: Marginal benefits over Solution 2 don't justify the massive complexity

### **Better Alternatives:**
- **Solution 2**: Achieves 95% of the same functionality in 22 hours with low risk
- **Solution 1**: Provides cleaner architecture if breaking changes are acceptable

### **When Solution 4 Might Be Considered:**
- If the team has extensive experience with complex Plugin Framework plan modifiers
- If perfect diff minimization is absolutely critical for user experience
- If there's significant time available for thorough testing and debugging
- As a future enhancement after Solution 2 proves the basic functionality

**Conclusion:** While Solution 4 represents the theoretical "perfect" solution, the implementation complexity and risk make it impractical. The sophisticated plan modifier logic required is at the edge of what's possible with the Plugin Framework, and the 87-hour development time with very high risk makes it unsuitable for immediate deployment.

**Solution 2 remains the clear winner** for restoring functionality quickly and safely, with the option to enhance with Solution 4 concepts in the future if the complexity proves worthwhile.