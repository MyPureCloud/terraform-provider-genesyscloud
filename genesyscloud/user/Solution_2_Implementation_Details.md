# Solution 2: Accept Set Identity Changes - Implementation Details

## Overview

This document provides detailed implementation analysis for Solution 2 (Accept Set Identity Changes approach) to help finalize the decision on which solution to pursue for the extension pool migration issue.

## Solution Summary

Keep `extension_pool_id` in the `phone_numbers` Set and implement proper processing in build/flatten functions. Accept that pool changes will cause Set element replacements, but restore full functionality with minimal code changes and no breaking changes for users.

## Required Code Changes

### 1. Schema Changes (`resource_genesyscloud_user_schema.go`)

**Complexity: None**

**No changes required** - the schema already has the correct structure:

```go
"extension_pool_id": schema.StringAttribute{
    Description:   "Id of the extension pool which contains this extension.",
    Optional:      true,
    PlanModifiers: []planmodifier.String{phoneplan.NullIfEmpty{}},
},
```

The existing schema is already correct and functional.

**Estimated Effort: 0 hours**

### 2. Model Updates (`resource_genesyscloud_user.go`)

**Complexity: None**

**No changes required** - the model already has the correct structure:

```go
type PhoneNumberModel struct {
    Number          types.String `tfsdk:"number"`
    MediaType       types.String `tfsdk:"media_type"`
    Type            types.String `tfsdk:"type"`
    Extension       types.String `tfsdk:"extension"`
    ExtensionPoolId types.String `tfsdk:"extension_pool_id"`
}
```

The existing model is already correct and functional.

**Estimated Effort: 0 hours**

### 3. State Management Updates (`resource_genesyscloud_user_utils.go`)

**Complexity: Low**

#### Fix flattenUserAddresses Function

**Current Issue (line 738):**
```go
"extension_pool_id": types.StringNull(), // <- always null in state -- TODO
```

**Fix Required:**
```go
func flattenUserAddresses(ctx context.Context, addresses *[]platformclientv2.Contact, proxy *userProxy) (types.List, pfdiag.Diagnostics) {
    // ... existing code ...
    
    // Case 2: Extension == Display → true internal extension (extension mapped to pool)
    if address.Extension != nil && address.Display != nil && *address.Extension == *address.Display {
        extensionNum := strings.Trim(*address.Extension, "()")
        if extensionNum != "" {
            phoneNumber["extension"] = types.StringValue(extensionNum)
        }
        
        // FIX: Actually fetch and store the extension pool ID
        poolId := fetchExtensionPoolId(ctx, extensionNum, proxy)
        if poolId != "" {
            phoneNumber["extension_pool_id"] = types.StringValue(poolId)
        } else {
            phoneNumber["extension_pool_id"] = types.StringNull()
        }
        
        // Keep existing logic for number normalization
        phoneNumber["number"] = types.StringNull()
    }
    
    // ... rest of existing code unchanged ...
}
```

**Changes Required:**
1. Remove the line that always sets `extension_pool_id` to null
2. Add logic to fetch and store the actual pool ID
3. Handle cases where pool ID is not found (set to null)

**Estimated Effort: 2-3 hours**

### 4. Request Building Updates (`resource_genesyscloud_user_utils.go`)

**Complexity: Low**

#### Fix buildSdkPhoneNumbers Function

**Current Issue:**
The function completely ignores `extension_pool_id` from the configuration.

**Current Code:**
```go
func buildSdkPhoneNumbers(configPhoneNumbers types.Set) ([]platformclientv2.Contact, pfdiag.Diagnostics) {
    // ... existing code ...
    
    // Optional field: extension
    if !phone.Extension.IsNull() && !phone.Extension.IsUnknown() {
        phoneExt := phone.Extension.ValueString()
        if phoneExt != "" {
            contact.Extension = &phoneExt
        }
    }
    
    // ExtensionPoolId is completely ignored - THIS IS THE BUG
    
    sdkContacts[i] = contact
}
```

**Fix Required:**
```go
func buildSdkPhoneNumbers(configPhoneNumbers types.Set) ([]platformclientv2.Contact, pfdiag.Diagnostics) {
    // ... existing code unchanged until extension handling ...
    
    // Optional field: extension
    if !phone.Extension.IsNull() && !phone.Extension.IsUnknown() {
        phoneExt := phone.Extension.ValueString()
        if phoneExt != "" {
            contact.Extension = &phoneExt
        }
    }
    
    // FIX: Process extension_pool_id from configuration
    if !phone.ExtensionPoolId.IsNull() && !phone.ExtensionPoolId.IsUnknown() {
        poolId := phone.ExtensionPoolId.ValueString()
        if poolId != "" {
            // Note: The Genesys API doesn't directly accept extension_pool_id in Contact
            // Instead, we need to ensure the extension is assigned to the correct pool
            // This happens automatically when the extension is within the pool's range
            // and the pool exists. The pool ID is used for validation/reference only.
            
            // For now, we store it as metadata but the actual assignment
            // happens server-side based on extension number and existing pools
            log.Printf("[DEBUG] Extension %s requested for pool %s", phoneExt, poolId)
        }
    }
    
    sdkContacts[i] = contact
}
```

**Alternative Approach (if API supports direct pool assignment):**
```go
// If the API supports direct pool assignment via a separate call
if !phone.ExtensionPoolId.IsNull() && !phone.ExtensionPoolId.IsUnknown() {
    poolId := phone.ExtensionPoolId.ValueString()
    phoneExt := phone.Extension.ValueString()
    if poolId != "" && phoneExt != "" {
        // Make separate API call to assign extension to specific pool
        err := assignExtensionToPool(phoneExt, poolId, proxy)
        if err != nil {
            diagnostics.AddError("Extension Pool Assignment Failed", 
                fmt.Sprintf("Failed to assign extension %s to pool %s: %v", phoneExt, poolId, err))
        }
    }
}
```

**Estimated Effort: 3-4 hours**

### 5. API Integration Research

**Complexity: Medium**

We need to understand exactly how the Genesys API handles extension pool assignments:

#### Research Required:
1. **Does the Contact API accept pool assignments directly?**
2. **Is there a separate API endpoint for extension-to-pool assignment?**
3. **Does the API auto-assign based on extension ranges?**

#### Likely Implementation:
Based on the existing `fetchExtensionPoolId` function, the API likely auto-assigns extensions to pools based on number ranges. Our job is to:
1. Send the extension to the API (already working)
2. Verify it gets assigned to the expected pool
3. Store the actual pool ID in state (the fix above)

**Estimated Effort: 4-6 hours (research + implementation)**

### 6. Test Updates (`resource_genesyscloud_user_test.go`)

**Complexity: Low**

#### Fix Existing Test TODOs

**Current Test Code:**
```go
// NOTE: extension_pool_id handling depends on Option 1 vs 2 - leave as TODO for now
"# TODO: extension_pool_id field handling",

// TODO: Add appropriate extension_pool_id assertion based on chosen option
```

**Fix Required:**
```go
func TestAccResourceUser_extensionPools(t *testing.T) {
    // ... existing test setup ...
    
    resource.Test(t, resource.TestCase{
        Steps: []resource.TestStep{
            {
                Config: generateUserWithExtensionPool(userResourceLabel, extensionPoolResourceLabel),
                Check: resource.ComposeTestCheckFunc(
                    resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel, "addresses.0.phone_numbers.0.extension", "8501"),
                    resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel, "addresses.0.phone_numbers.0.media_type", "PHONE"),
                    resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel, "addresses.0.phone_numbers.0.type", "WORK"),
                    // FIX: Add actual extension_pool_id assertion
                    resource.TestCheckResourceAttrPair(ResourceType+"."+userResourceLabel, "addresses.0.phone_numbers.0.extension_pool_id",
                        extensionPool.ResourceType+"."+extensionPoolResourceLabel, "id"),
                ),
            },
        },
    })
}
```

#### Update Test Configuration Generation:
```go
func generateFrameworkUserWithExtensionPool(userLabel, poolLabel string) string {
    return fmt.Sprintf(`
resource "genesyscloud_user" "%s" {
    name  = "Test User"
    email = "test@example.com"
    
    addresses {
        phone_numbers {
            extension         = "8501"
            media_type        = "PHONE"
            type              = "WORK"
            extension_pool_id = genesyscloud_telephony_providers_edges_extension_pool.%s.id
        }
    }
}

resource "genesyscloud_telephony_providers_edges_extension_pool" "%s" {
    start_number = "8500"
    end_number   = "8699"
}
`, userLabel, poolLabel, poolLabel)
}
```

**Estimated Effort: 2-3 hours**

### 7. Documentation Updates

**Complexity: Low**

#### Update Examples
The existing documentation examples should already be correct since we're not changing the schema. We just need to:

1. **Verify examples work** with the fixed implementation
2. **Add notes about Set identity behavior** when pool assignments change
3. **Update migration guide** to mention the diff behavior difference from SDKv2

**Example Documentation Addition:**
```markdown
## Extension Pool Behavior

When using `extension_pool_id` in phone numbers, note that:

- Pool assignments are stored in state and visible to users
- Changing pool assignments will cause the phone number Set element to be replaced (not updated in-place)
- This may result in more plan diffs compared to SDKv2, but functionality is preserved
- Extensions are automatically validated against the specified pool's number range
```

**Estimated Effort: 1-2 hours**

### 8. Edge Case Handling

**Complexity: Medium**

#### Handle Error Scenarios:
1. **Extension not in pool range** - validation error
2. **Pool doesn't exist** - reference error  
3. **Extension already assigned elsewhere** - conflict error
4. **Pool assignment fails** - API error

#### Implementation:
```go
func validateExtensionPoolAssignment(extension, poolId string, proxy *userProxy) pfdiag.Diagnostics {
    var diagnostics pfdiag.Diagnostics
    
    // Fetch pool details
    pool, _, err := proxy.getTelephonyExtensionPoolById(context.Background(), poolId)
    if err != nil {
        diagnostics.AddError("Extension Pool Not Found", 
            fmt.Sprintf("Extension pool %s not found: %v", poolId, err))
        return diagnostics
    }
    
    // Validate extension is in pool range
    extNum, err := strconv.Atoi(extension)
    if err != nil {
        diagnostics.AddError("Invalid Extension", 
            fmt.Sprintf("Extension %s is not a valid number: %v", extension, err))
        return diagnostics
    }
    
    startNum, _ := strconv.Atoi(*pool.StartNumber)
    endNum, _ := strconv.Atoi(*pool.EndNumber)
    
    if extNum < startNum || extNum > endNum {
        diagnostics.AddError("Extension Out of Range", 
            fmt.Sprintf("Extension %s is not within pool range %s-%s", extension, *pool.StartNumber, *pool.EndNumber))
        return diagnostics
    }
    
    return diagnostics
}
```

**Estimated Effort: 3-4 hours**

## Total Implementation Effort

### Summary:
- **Schema Changes**: 0 hours (no changes needed)
- **Model Updates**: 0 hours (no changes needed)
- **State Management**: 3 hours (fix flatten function)
- **Request Building**: 4 hours (fix build function)
- **API Integration**: 6 hours (research + implementation)
- **Test Updates**: 3 hours (fix existing tests)
- **Documentation**: 2 hours (minor updates)
- **Edge Case Handling**: 4 hours (validation logic)
- **Total: 22 hours (~3 days)**

## Implementation Risks

### Low Risk Areas:
1. **Schema/Model**: No changes required - existing structure is correct
2. **Test Updates**: Straightforward assertion fixes
3. **Documentation**: Minor updates to existing content

### Medium Risk Areas:
1. **API Integration**: Need to understand exact API behavior for pool assignments
2. **Edge Cases**: Proper error handling for various failure scenarios
3. **State Correlation**: Ensuring Set elements correlate correctly after fixes

### Minimal Risk Areas:
1. **Breaking Changes**: None - users keep existing configurations
2. **Migration**: No state migration required
3. **Backward Compatibility**: Fully maintained

## Breaking Change Impact

**Zero Breaking Changes Required:**

Users can keep their existing configurations exactly as they are:

```hcl
# This configuration works before and after the fix
addresses {
  phone_numbers {
    extension         = "8501"
    media_type        = "PHONE"
    type              = "WORK"
    extension_pool_id = genesyscloud_telephony_providers_edges_extension_pool.pool.id
  }
}
```

The only difference users will notice is:
- ✅ **Extension pools now work** (currently broken)
- ⚠️ **More diffs when pool assignments change** (Set element replacements vs. in-place updates)

## Expected Behavior Changes

### Before Fix (Current State):
```bash
# terraform plan shows
+ extension_pool_id = (known after apply)

# terraform apply fails with
Error: Provider produced inconsistent result after apply

# State shows
"extension_pool_id": null
```

### After Fix:
```bash
# terraform plan shows  
+ extension_pool_id = "7d0135e7-82fa-4f46-9fba-686289b4e0f7"

# terraform apply succeeds
Apply complete! Resources: 1 added, 0 changed, 0 destroyed.

# State shows
"extension_pool_id": "7d0135e7-82fa-4f46-9fba-686289b4e0f7"
```

### When Pool Assignment Changes:
```bash
# SDKv2 behavior (in-place update):
~ phone_numbers {
    ~ extension_pool_id = "old-pool-id" -> "new-pool-id"
  }

# Plugin Framework behavior (element replacement):
- phone_numbers {
    - extension_pool_id = "old-pool-id"
  }
+ phone_numbers {
    + extension_pool_id = "new-pool-id"
  }
```

## Comparison with Solution 1

| Aspect | Solution 1 | Solution 2 |
|--------|------------|------------|
| **Development Time** | 46-91 hours | 22 hours |
| **Breaking Changes** | Yes (major) | No |
| **User Impact** | High | None |
| **Implementation Risk** | High | Low |
| **Code Complexity** | Very High | Low |
| **Migration Required** | Yes | No |
| **Timeline** | 1.2-2.3 weeks | 3 days |

## Recommendation

**Solution 2 is strongly recommended** because:

### **Immediate Benefits:**
- ✅ **Fastest time to working solution** (3 days vs. 1.2-2.3 weeks)
- ✅ **Zero user impact** - no configuration changes required
- ✅ **Low implementation risk** - simple bug fixes vs. architectural changes
- ✅ **Preserves all functionality** - users get exactly what they expect

### **Acceptable Trade-offs:**
- ⚠️ **More diffs when pools change** - acceptable for restored functionality
- ⚠️ **Set element replacements** - cosmetic issue, no functional impact

### **Strategic Value:**
- 🎯 **Unblocks users immediately** - critical for migration success
- 🎯 **Maintains SDKv2 compatibility** - smooth migration experience  
- 🎯 **Allows future improvements** - can still implement Solution 1 later if needed

**Conclusion:** Solution 2 provides the optimal balance of speed, safety, and functionality restoration with minimal effort and zero user disruption.