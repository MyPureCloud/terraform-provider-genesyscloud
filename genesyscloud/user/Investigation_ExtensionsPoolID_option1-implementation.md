# Investigation: Option 1 Implementation - Computed Field Approach

## Document Overview

This document provides a detailed implementation investigation for Option 1 (Computed Field Approach) to resolve the extension pool ID Set identity mismatch issue in the Plugin Framework migration. This analysis examines the technical requirements, implementation challenges, and code changes needed to implement the dual-field system while maintaining backward compatibility.

**Investigation Date**: January 8, 2026  
**Approach**: Option 1 - Computed Field Approach (Non-Breaking Change)  
**Complexity Assessment**: HIGH  
**Implementation Status**: INVESTIGATED (Not Implemented)

---

## Implementation Strategy Overview

Option 1 implements a **dual-field system** that separates user configuration from Set identity management:

- **User Configuration Field**: `extension_pool_id` remains in `phone_numbers` for backward compatibility
- **Computed Tracking Field**: `phone_extension_pools` at top level for actual pool tracking
- **Set Identity Solution**: Keep `extension_pool_id` null in state to exclude from Set identity
- **Dependency Management**: Use configuration field for Terraform dependencies

## Current Codebase Analysis

### Existing Schema Structure

**Current Schema (in `resource_genesyscloud_user_schema.go`):**
```go
// Top-level computed field is commented out (lines 71-82)
/*
"phone_extension_pools": schema.MapAttribute{
    ElementType: types.StringType,
    Computed:    true,
    Description: "Id of the extension pool which contains this extension." +
        "Computed mapping of phone identity keys to  (MEDIA|TYPE|E164|EXT) to extension_pool_id." +
        "Used internally to prevent diffs when pool assignments change.",
},
*/

// Phone numbers schema (lines 259-295)
"phone_numbers": schema.SetNestedBlock{
    NestedObject: schema.NestedBlockObject{
        Attributes: map[string]schema.Attribute{
            "extension_pool_id": schema.StringAttribute{
                Description:   "Id of the extension pool which contains this extension.",
                Optional:      true,
                PlanModifiers: []planmodifier.String{phoneplan.NullIfEmpty{}},
            },
        },
    },
},
```

**Current State Management (in `resource_genesyscloud_user_utils.go`):**
```go
// flattenUserAddresses() already sets extension_pool_id to null (line 734)
"extension_pool_id": types.StringNull(), // <- always null in state
```

### Existing Implementation Gaps

1. **Missing Computed Field**: Top-level `phone_extension_pools` field is commented out
2. **No Identity Key Generation**: Missing logic to generate phone identity keys
3. **No Pool ID Lookup**: Missing API integration to fetch extension pool IDs
4. **No Dual-Field Synchronization**: Missing logic to coordinate both fields
5. **Incomplete State Management**: `flattenUserAddresses()` doesn't populate computed field

---

## Detailed Implementation Requirements

### 1. Schema Changes

#### **A. Uncomment and Configure Top-Level Computed Field**

**File**: `resource_genesyscloud_user_schema.go`  
**Location**: Lines 71-82

```go
// CHANGE: Uncomment and refine the phone_extension_pools field
"phone_extension_pools": schema.MapAttribute{
    ElementType: types.StringType,
    Computed:    true,
    Description: "Computed mapping of phone identity keys to extension_pool_id. " +
                "Used internally to prevent diffs when pool assignments change. " +
                "Map structure: 'MEDIA|TYPE|E164|EXT' -> 'extension_pool_id'. " +
                "This field tracks actual pool assignments while keeping phone_numbers Set identity stable.",
},
```

#### **B. Modify Phone Numbers Field Behavior**

**File**: `resource_genesyscloud_user_schema.go`  
**Location**: Lines 295-301

```go
// CHANGE: Update extension_pool_id field description and plan modifiers
"extension_pool_id": schema.StringAttribute{
    Description: "Id of the extension pool which contains this extension. " +
                "Used for configuration syntax and Terraform dependencies only. " +
                "Actual pool assignments are tracked in the phone_extension_pools computed field. " +
                "This field will always appear as null in terraform show output.",
    Optional: true,
    // REMOVE: phoneplan.NullIfEmpty{} - replace with custom plan modifier
    PlanModifiers: []planmodifier.String{
        &ExtensionPoolIdAlwaysNullModifier{},
    },
},
```

### 2. Custom Plan Modifier Implementation

#### **A. Create ExtensionPoolIdAlwaysNullModifier**

**New File**: `resource_genesyscloud_user_plan_modifiers.go`

```go
package user

import (
    "context"
    "github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
    "github.com/hashicorp/terraform-plugin-framework/types"
)

// ExtensionPoolIdAlwaysNullModifier ensures extension_pool_id is always null in state
// This excludes it from Set identity calculation while preserving config for dependencies
type ExtensionPoolIdAlwaysNullModifier struct{}

func (m *ExtensionPoolIdAlwaysNullModifier) Description(ctx context.Context) string {
    return "Always sets extension_pool_id to null in state to exclude from Set identity"
}

func (m *ExtensionPoolIdAlwaysNullModifier) MarkdownDescription(ctx context.Context) string {
    return "Always sets extension_pool_id to null in state to exclude from Set identity"
}

func (m *ExtensionPoolIdAlwaysNullModifier) PlanModifyString(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
    // Always set planned value to null, regardless of configuration
    // This ensures the field never appears in state, excluding it from Set identity
    resp.PlanValue = types.StringNull()
    
    // Log the configuration value for debugging (but don't store it)
    if !req.ConfigValue.IsNull() && !req.ConfigValue.IsUnknown() {
        log.Printf("[DEBUG] extension_pool_id configured as %q but will be stored as null", req.ConfigValue.ValueString())
    }
}
```

### 3. State Management Changes

#### **A. Modify flattenUserAddresses Function Signature**

**File**: `resource_genesyscloud_user_utils.go`  
**Location**: Line 675

```go
// CHANGE: Update function signature to return computed field
func flattenUserAddresses(ctx context.Context, addresses *[]platformclientv2.Contact, proxy *userProxy) (types.List, types.Map, pfdiag.Diagnostics) {
    var diagnostics pfdiag.Diagnostics
    
    // Initialize computed field tracking
    phoneExtensionPools := map[string]attr.Value{}
    
    // ... existing address processing logic ...
    
    // NEW: For each phone number, generate identity key and track pool ID
    for _, address := range *addresses {
        if address.MediaType != nil && (*address.MediaType == "PHONE" || *address.MediaType == "SMS") {
            // Generate identity key for this phone number
            identityKey := generatePhoneIdentityKey(&address)
            
            // Fetch actual extension pool ID from API if extension exists
            if address.Extension != nil && *address.Extension != "" {
                poolId, poolDiags := fetchExtensionPoolIdForExtension(ctx, *address.Extension, proxy)
                diagnostics.Append(poolDiags...)
                
                if poolId != "" {
                    phoneExtensionPools[identityKey] = types.StringValue(poolId)
                }
            }
        }
    }
    
    // Build addresses list (existing logic, extension_pool_id remains null)
    addressesList, listDiags := buildAddressesList(...)
    diagnostics.Append(listDiags...)
    
    // Build computed map
    phoneExtensionPoolsMap := types.MapValueMust(types.StringType, phoneExtensionPools)
    
    return addressesList, phoneExtensionPoolsMap, diagnostics
}
```

#### **B. Add Identity Key Generation Function**

**File**: `resource_genesyscloud_user_utils.go`  
**Location**: New function

```go
// generatePhoneIdentityKey creates a unique identity key for phone number tracking
// Format: "MEDIA|TYPE|E164|EXT" - matches the key structure used in computed field
func generatePhoneIdentityKey(contact *platformclientv2.Contact) string {
    media := ""
    if contact.MediaType != nil {
        media = *contact.MediaType
    }
    
    phoneType := ""
    if contact.VarType != nil {
        phoneType = *contact.VarType
    }
    
    // Use E.164 formatted number if available
    number := ""
    if contact.Address != nil {
        // TODO: Consider E.164 normalization for consistency
        number = *contact.Address
    }
    
    extension := ""
    if contact.Extension != nil {
        extension = *contact.Extension
    }
    
    return fmt.Sprintf("%s|%s|%s|%s", media, phoneType, number, extension)
}
```

#### **C. Add Extension Pool ID Lookup Function**

**File**: `resource_genesyscloud_user_utils.go`  
**Location**: New function

```go
// fetchExtensionPoolIdForExtension queries the API to find which extension pool contains the given extension
// This is a complex operation that requires checking all extension pools and their ranges
func fetchExtensionPoolIdForExtension(ctx context.Context, extension string, proxy *userProxy) (string, pfdiag.Diagnostics) {
    var diagnostics pfdiag.Diagnostics
    
    // Convert extension to integer for range checking
    extNum, err := strconv.Atoi(extension)
    if err != nil {
        diagnostics.AddWarning(
            "Extension Pool Lookup Failed",
            fmt.Sprintf("Could not parse extension %q as integer: %v", extension, err),
        )
        return "", diagnostics
    }
    
    // Query all extension pools from API
    // NOTE: This is expensive - consider caching strategy
    pools, resp, getErr := proxy.telephonyApi.GetTelephonyProvidersEdgesExtensionpools(&platformclientv2.GetTelephonyProvidersEdgesExtensionpoolsOpts{})
    if getErr != nil {
        diagnostics.AddError(
            "Extension Pool Query Failed",
            fmt.Sprintf("Failed to query extension pools: %v", getErr),
        )
        return "", diagnostics
    }
    
    if pools.Entities == nil {
        return "", diagnostics
    }
    
    // Check each pool to see if extension falls within its range
    for _, pool := range *pools.Entities {
        if pool.StartNumber == nil || pool.EndNumber == nil {
            continue
        }
        
        startNum, startErr := strconv.Atoi(*pool.StartNumber)
        endNum, endErr := strconv.Atoi(*pool.EndNumber)
        
        if startErr != nil || endErr != nil {
            continue
        }
        
        // Check if extension falls within this pool's range
        if extNum >= startNum && extNum <= endNum {
            if pool.Id != nil {
                log.Printf("[DEBUG] Extension %s found in pool %s (range %s-%s)", 
                    extension, *pool.Id, *pool.StartNumber, *pool.EndNumber)
                return *pool.Id, diagnostics
            }
        }
    }
    
    // Extension not found in any pool
    log.Printf("[DEBUG] Extension %s not found in any extension pool", extension)
    return "", diagnostics
}
```

### 4. CRUD Operation Changes

#### **A. Modify Read Operation**

**File**: `resource_genesyscloud_user.go`  
**Location**: Read function

```go
func (r *userResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
    // ... existing read logic ...
    
    // CHANGE: Call modified flattenUserAddresses that returns both values
    addressesList, phoneExtensionPoolsMap, diags := flattenUserAddresses(ctx, user.Addresses, r.proxy)
    resp.Diagnostics.Append(diags...)
    if resp.Diagnostics.HasError() {
        return
    }
    
    // Set both fields in state
    state.Addresses = addressesList
    state.PhoneExtensionPools = phoneExtensionPoolsMap
    
    // ... rest of read logic ...
}
```

#### **B. Modify Create Operation**

**File**: `resource_genesyscloud_user.go`  
**Location**: Create function

```go
func (r *userResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
    // ... existing create logic ...
    
    // After user creation, read back state to populate computed field
    // This ensures the computed field is properly synchronized
    
    // Re-read user to get current state including extension pool assignments
    user, getResp, getErr := r.proxy.userApi.GetUser(userId, nil, "", "")
    if getErr != nil {
        // ... error handling ...
    }
    
    // Populate both address fields
    addressesList, phoneExtensionPoolsMap, diags := flattenUserAddresses(ctx, user.Addresses, r.proxy)
    resp.Diagnostics.Append(diags...)
    
    state.Addresses = addressesList
    state.PhoneExtensionPools = phoneExtensionPoolsMap
    
    // ... rest of create logic ...
}
```

#### **C. Modify Update Operation**

**File**: `resource_genesyscloud_user.go`  
**Location**: Update function

```go
func (r *userResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
    // ... existing update logic ...
    
    // IMPORTANT: During update, we need to handle the dual-field system carefully
    // 1. Process user configuration (including extension_pool_id references for dependencies)
    // 2. Send update to API (buildSdkPhoneNumbers handles this)
    // 3. Re-read state to populate computed field
    // 4. Ensure extension_pool_id remains null in final state
    
    // After update, re-read to synchronize computed field
    user, getResp, getErr := r.proxy.userApi.GetUser(userId, nil, "", "")
    if getErr != nil {
        // ... error handling ...
    }
    
    // Populate both address fields
    addressesList, phoneExtensionPoolsMap, diags := flattenUserAddresses(ctx, user.Addresses, r.proxy)
    resp.Diagnostics.Append(diags...)
    
    state.Addresses = addressesList
    state.PhoneExtensionPools = phoneExtensionPoolsMap
    
    // ... rest of update logic ...
}
```

### 5. Build SDK Functions Changes

#### **A. Modify buildSdkPhoneNumbers Function**

**File**: `resource_genesyscloud_user_utils.go`  
**Location**: Line 1328

```go
func buildSdkPhoneNumbers(configPhoneNumbers types.Set) ([]platformclientv2.Contact, pfdiag.Diagnostics) {
    var diagnostics pfdiag.Diagnostics
    
    // ... existing validation logic ...
    
    // Build contacts - IMPORTANT: Use extension_pool_id from config for API calls
    for i, phone := range phoneNumbers {
        // ... existing field processing ...
        
        // CHANGE: Handle extension_pool_id from configuration
        // Even though this field is null in state, it may have a value in config
        // We need to use the config value for API operations
        if !phone.ExtensionPoolId.IsNull() && !phone.ExtensionPoolId.IsUnknown() {
            poolId := phone.ExtensionPoolId.ValueString()
            if poolId != "" {
                // TODO: Determine how to use extension_pool_id in API calls
                // The Genesys Cloud API may need this for proper extension assignment
                log.Printf("[DEBUG] Extension pool ID from config: %s", poolId)
            }
        }
        
        sdkContacts[i] = contact
    }
    
    return sdkContacts, diagnostics
}
```

### 6. Testing Changes Required

#### **A. Update Existing Test Cases**

**File**: `resource_genesyscloud_user_test.go`  
**Location**: All extension pool tests

```go
// CHANGE: Update all test checks to expect dual-field behavior
func TestAccFrameworkResourceUserAddressWithExtensionPool(t *testing.T) {
    // ... test setup ...
    
    resource.TestStep{
        Config: generateFrameworkUserWithExtensionPool(...),
        Check: resource.ComposeTestCheckFunc(
            // OLD: Check extension_pool_id field directly
            // resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel, "addresses.0.phone_numbers.0.extension_pool_id", poolId),
            
            // NEW: Check that extension_pool_id is null in state
            resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel, "addresses.0.phone_numbers.0.extension_pool_id", ""),
            
            // NEW: Check computed field contains actual pool assignment
            resource.TestCheckResourceAttrSet(ResourceType+"."+userResourceLabel, "phone_extension_pools.PHONE|WORK||21000"),
            
            // NEW: Verify computed field has correct pool ID
            resource.TestCheckResourceAttrPair(
                ResourceType+"."+userResourceLabel, "phone_extension_pools.PHONE|WORK||21000",
                "genesyscloud_telephony_providers_edges_extension_pool.test-extension-pool-1", "id",
            ),
        ),
    },
}
```

#### **B. Add New Test Cases**

**File**: `resource_genesyscloud_user_test.go`  
**Location**: New test functions

```go
// NEW: Test dual-field synchronization
func TestAccFrameworkResourceUserDualFieldSync(t *testing.T) {
    // Test that computed field updates when extension pool changes
    // Test that configuration field remains for dependencies
    // Test that state field is always null
}

// NEW: Test plan output behavior
func TestAccFrameworkResourceUserPlanOutput(t *testing.T) {
    // Test that plan shows changes in computed field, not config field
    // Test user experience with configuration vs state mismatch
}

// NEW: Test import behavior
func TestAccFrameworkResourceUserImportWithExtensionPools(t *testing.T) {
    // Test that import populates computed field correctly
    // Test that imported state has null extension_pool_id fields
}
```

### 7. Documentation Changes Required

#### **A. Update Resource Documentation**

**File**: `docs/resources/user.md`

```markdown
## Extension Pool Integration

The `genesyscloud_user` resource supports extension pool integration through a dual-field system:

### Configuration Field: `extension_pool_id`
- **Purpose**: Used for configuration syntax and Terraform dependencies
- **Location**: Inside `phone_numbers` block
- **Behavior**: Always appears as `null` in `terraform show` output
- **Usage**: Configure this field to create proper resource dependencies

### Computed Field: `phone_extension_pools`
- **Purpose**: Tracks actual extension pool assignments
- **Location**: Top-level resource attribute
- **Behavior**: Automatically populated by the provider
- **Format**: Map of phone identity keys to extension pool IDs

### Example Configuration
```hcl
resource "genesyscloud_user" "example" {
  addresses {
    phone_numbers {
      extension = "4105"
      extension_pool_id = genesyscloud_telephony_providers_edges_extension_pool.pool1.id
    }
  }
}
```

### Example State Output
```hcl
# terraform show output
resource "genesyscloud_user" "example" {
  addresses {
    phone_numbers {
      extension = "4105"
      extension_pool_id = null  # Always null in state
    }
  }
  
  phone_extension_pools = {
    "PHONE|WORK||4105" = "pool-id-abc123"  # Actual assignment
  }
}
```

### Important Notes
- Changes to extension pools appear in `phone_extension_pools`, not in `phone_numbers`
- The `extension_pool_id` field is required for proper Terraform dependency tracking
- This design maintains backward compatibility with existing configurations
```

---

## Implementation Challenges Analysis

### 1. Technical Complexity Challenges

#### **A. Extension Pool ID Lookup Complexity**
- **Challenge**: Determining which extension pool contains a given extension
- **API Limitation**: No direct API to lookup pool by extension
- **Performance Impact**: Must query all extension pools for each extension
- **Reliability Issues**: Extension pools can change, making lookups unreliable
- **Caching Complexity**: Need sophisticated caching to avoid performance issues

#### **B. State Synchronization Complexity**
- **Challenge**: Keeping configuration field and computed field synchronized
- **Edge Cases**: What happens when API returns unexpected data?
- **Consistency Issues**: Config and computed field might become out of sync
- **Debugging Difficulty**: Two sources of truth make troubleshooting complex

#### **C. Plan/Apply Logic Complexity**
- **Challenge**: Predicting computed field values during plan phase
- **Limitation**: Plan phase doesn't have access to API data
- **User Experience**: Plan output might not match apply results
- **Dependency Issues**: Computed field changes might not trigger dependent resources

### 2. User Experience Challenges

#### **A. Configuration vs State Mismatch**
- **Confusion**: Users configure `extension_pool_id` but see `null` in state
- **Debugging**: Users must understand dual-field system to troubleshoot
- **Documentation Burden**: Extensive explanation required
- **Support Overhead**: Increased support requests due to confusion

#### **B. Plan Output Confusion**
- **Unexpected Behavior**: Changes appear in computed field, not configured field
- **User Expectations**: Users expect changes to appear where they configured them
- **Learning Curve**: Users must learn new mental model for extension pools

### 3. Maintenance Challenges

#### **A. Permanent Complexity**
- **Technical Debt**: Dual-field system must be maintained indefinitely
- **Code Complexity**: Every change must consider both field representations
- **Testing Overhead**: All tests must verify both field behaviors
- **Documentation Maintenance**: Complex system requires ongoing documentation updates

#### **B. Future Enhancement Risk**
- **Feature Development**: New features must work with dual-field system
- **Migration Difficulty**: Future changes become more complex
- **Backward Compatibility**: Must maintain complex system forever

---

## Risk Assessment

### High Risk Factors

1. **Implementation Complexity**: Very high complexity with multiple failure points
2. **User Experience**: Confusing behavior that violates user expectations
3. **API Integration**: Complex and unreliable extension pool lookup logic
4. **Performance Impact**: Multiple API calls for extension pool lookups
5. **Maintenance Burden**: Permanent technical debt affecting all future development

### Medium Risk Factors

1. **Testing Complexity**: Extensive test scenarios required for dual-field system
2. **Documentation Overhead**: Complex system requires extensive documentation
3. **Support Impact**: Increased support burden due to user confusion

### Low Risk Factors

1. **Backward Compatibility**: Maintains existing configuration syntax
2. **Dependency Management**: Preserves Terraform dependency behavior

---

## Cost-Benefit Analysis

### Implementation Costs

| Cost Category | Estimate | Impact |
|---------------|----------|---------|
| **Development Time** | 3-4 weeks | High |
| **Testing Effort** | 1-2 weeks | High |
| **Documentation** | 1 week | Medium |
| **Code Review** | 1 week | Medium |
| **Total Initial Cost** | 6-8 weeks | Very High |

### Ongoing Costs

| Cost Category | Annual Impact | Description |
|---------------|---------------|-------------|
| **Maintenance** | 2-3 weeks | Dual-field system maintenance |
| **Support** | 1-2 weeks | User confusion support |
| **Documentation** | 1 week | Keeping complex docs updated |
| **Feature Development** | 20% overhead | All features must consider dual fields |

### Benefits

| Benefit | Value | Notes |
|---------|-------|-------|
| **Backward Compatibility** | High | No breaking changes required |
| **Dependency Preservation** | High | Maintains Terraform dependencies |
| **Set Identity Resolution** | High | Solves the core technical issue |

### Net Assessment

**Costs Significantly Outweigh Benefits**
- Very high implementation and maintenance costs
- Poor user experience creates ongoing support burden
- Complex system creates permanent technical debt
- Better alternatives available (Option 2)

---

## Comparison with Option 2

| Aspect | Option 1 (Computed Field) | Option 2 (Dependency-Only) | Winner |
|--------|---------------------------|----------------------------|---------|
| **Implementation Time** | 6-8 weeks | 1-2 days | Option 2 |
| **Code Complexity** | Very High | Low | Option 2 |
| **User Experience** | Poor (confusing) | Excellent (clear) | Option 2 |
| **Maintenance Burden** | High (permanent) | Low (none) | Option 2 |
| **Breaking Changes** | None | Minimal | Option 1 |
| **Technical Debt** | Significant | None | Option 2 |
| **Performance** | Poor (API lookups) | Good (no lookups) | Option 2 |
| **Testing Complexity** | High | Low | Option 2 |
| **Documentation Burden** | High | Low | Option 2 |
| **Future Enhancement Risk** | High | Low | Option 2 |

**Overall Winner: Option 2** (9 out of 10 categories)

---

## Final Recommendation

### DO NOT IMPLEMENT Option 1

**Primary Reasons:**

1. **Excessive Complexity**: Implementation requires 6-8 weeks vs 1-2 days for Option 2
2. **Poor User Experience**: Configuration vs state mismatch will confuse users
3. **High Risk**: Multiple complex systems that can fail independently
4. **Permanent Technical Debt**: Dual-field system must be maintained forever
5. **Better Alternative Available**: Option 2 achieves same result with 1% of the complexity

### Implement Option 2 Instead

**Justification:**

1. **Simple Implementation**: Remove problematic field, use `depends_on`
2. **Clear User Experience**: Configuration matches state representation
3. **No Technical Debt**: Clean, maintainable solution
4. **Proven Pattern**: Follows established Terraform best practices
5. **Future-Proof**: Provides solid foundation for future development

### Strategic Assessment

Option 1 represents a **classic over-engineering scenario** where:
- The solution is far more complex than the problem
- Implementation costs exceed the value delivered
- The cure is worse than the disease

Option 2 demonstrates **elegant problem-solving** by:
- Eliminating the problem at its source
- Providing a clean, maintainable solution
- Following established best practices
- Delivering superior user experience

---

## Conclusion

This investigation confirms that **Option 1 should not be implemented** due to its excessive complexity, poor user experience, and high maintenance burden. The analysis reveals that Option 1 would require:

- **6-8 weeks of development time** vs 1-2 days for Option 2
- **Permanent technical debt** vs clean implementation
- **Complex dual-field system** vs simple single-field approach
- **Poor user experience** vs intuitive behavior

**Option 2 (Dependency-Only Approach) is clearly superior** and should be implemented instead. This investigation validates the original recommendation and provides detailed evidence for why the simpler approach is the correct choice.

The key insight is that **removing complexity is often better than managing it** - Option 2 eliminates the problem at its source rather than creating elaborate workarounds to manage it.