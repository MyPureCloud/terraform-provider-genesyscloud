# Solution 1: Top-Level Computed Map - Implementation Details

## Overview

This document provides detailed implementation analysis for Solution 1 (Top-Level Computed Map approach) to help finalize the decision on which solution to pursue for the extension pool migration issue.

## Solution Summary

Move `extension_pool_id` out of the `phone_numbers` Set entirely and store it as a top-level computed `phone_extension_pools` map attribute. This eliminates Set identity issues while preserving pool assignment visibility.

## Required Code Changes

### 1. Schema Changes (`resource_genesyscloud_user_schema.go`)

**Complexity: Medium**

#### Add Top-Level Map Attribute
```go
// Add to UserResourceSchema() Attributes map
"phone_extension_pools": schema.MapAttribute{
    ElementType: types.StringType,
    Computed:    true,
    Description: "Computed mapping of phone identity keys (MEDIA|TYPE|E164|EXT) to extension_pool_id. " +
        "Used internally to prevent diffs when pool assignments change.",
},
```

#### Remove extension_pool_id from phone_numbers
```go
// Remove this entire block from phone_numbers NestedObject
"extension_pool_id": schema.StringAttribute{
    Description:   "Id of the extension pool which contains this extension.",
    Optional:      true,
    PlanModifiers: []planmodifier.String{phoneplan.NullIfEmpty{}},
},
```

**Estimated Effort: 2-3 hours**

### 2. Model Updates (`resource_genesyscloud_user.go`)

**Complexity: Low**

#### Update UserFrameworkResourceModel
```go
type UserFrameworkResourceModel struct {
    // ... existing fields ...
    PhoneExtensionPools types.Map    `tfsdk:"phone_extension_pools"`
    // ... rest of fields ...
}
```

#### Remove from PhoneNumberModel
```go
type PhoneNumberModel struct {
    Number          types.String `tfsdk:"number"`
    MediaType       types.String `tfsdk:"media_type"`
    Type            types.String `tfsdk:"type"`
    Extension       types.String `tfsdk:"extension"`
    // Remove: ExtensionPoolId types.String `tfsdk:"extension_pool_id"`
}
```

**Estimated Effort: 1 hour**

### 3. State Management Updates (`resource_genesyscloud_user_utils.go`)

**Complexity: High**

#### Update flattenUserAddresses Function
```go
func flattenUserAddresses(ctx context.Context, addresses *[]platformclientv2.Contact, proxy *userProxy) (types.List, types.Map, pfdiag.Diagnostics) {
    var diagnostics pfdiag.Diagnostics
    
    // ... existing phone processing logic ...
    
    // NEW: Build extension pools map
    extensionPoolsMap := make(map[string]attr.Value)
    
    for _, address := range *addresses {
        if address.MediaType != nil && (*address.MediaType == "SMS" || *address.MediaType == "PHONE") {
            // Check if this is an extension with pool assignment
            if address.Extension != nil && address.Display != nil && *address.Extension == *address.Display {
                extensionNum := strings.Trim(*address.Extension, "()")
                if extensionNum != "" {
                    // Fetch pool ID for this extension
                    poolId := fetchExtensionPoolId(ctx, extensionNum, proxy)
                    if poolId != "" {
                        // Create identity key: MEDIA|TYPE|EXT
                        media := "PHONE"
                        if address.MediaType != nil {
                            media = *address.MediaType
                        }
                        phoneType := "WORK"
                        if address.VarType != nil {
                            phoneType = *address.VarType
                        }
                        
                        identityKey := fmt.Sprintf("%s|%s|%s", media, phoneType, extensionNum)
                        extensionPoolsMap[identityKey] = types.StringValue(poolId)
                    }
                }
            }
        }
    }
    
    // Convert map to types.Map
    extensionPoolsMapValue, mapDiags := types.MapValue(types.StringType, extensionPoolsMap)
    diagnostics.Append(mapDiags...)
    
    return addressesList, extensionPoolsMapValue, diagnostics
}
```

#### Update Function Signatures
```go
// Update all callers to handle the new return value
addresses, extensionPools, addressDiags := flattenUserAddresses(ctx, user.Addresses, proxy)
state.Addresses = addresses
state.PhoneExtensionPools = extensionPools
```

**Estimated Effort: 8-10 hours**

### 4. Request Building Updates (`buildSdkPhoneNumbers`)

**Complexity: Medium**

#### Current Issue
The function currently ignores `extension_pool_id` completely. With Solution 1, we need a different approach.

#### Option A: Separate Extension Pool Assignment Resource
```go
// Create new resource type for pool assignments
resource "genesyscloud_user_extension_pool_assignment" "example" {
    user_id           = genesyscloud_user.example.id
    extension         = "8501"
    extension_pool_id = genesyscloud_telephony_providers_edges_extension_pool.pool.id
}
```

#### Option B: Data Source Lookup
```go
// In buildSdkPhoneNumbers, lookup pool ID based on extension
func buildSdkPhoneNumbers(configPhoneNumbers types.Set, proxy *userProxy) ([]platformclientv2.Contact, pfdiag.Diagnostics) {
    // ... existing logic ...
    
    // For extensions, lookup and assign pool
    if !phone.Extension.IsNull() && !phone.Extension.IsUnknown() {
        phoneExt := phone.Extension.ValueString()
        if phoneExt != "" {
            contact.Extension = &phoneExt
            
            // NEW: Auto-assign to appropriate pool
            poolId := fetchExtensionPoolId(context.Background(), phoneExt, proxy)
            if poolId != "" {
                // Set pool assignment via separate API call
                assignExtensionToPool(phoneExt, poolId, proxy)
            }
        }
    }
}
```

**Estimated Effort: 12-15 hours (Option A) or 6-8 hours (Option B)**

### 5. New Resource Type (Option A)

**Complexity: Very High**

If we choose Option A, we need to create an entirely new resource:

#### Files to Create:
- `genesyscloud/user_extension_pool_assignment/resource_genesyscloud_user_extension_pool_assignment.go`
- `genesyscloud/user_extension_pool_assignment/resource_genesyscloud_user_extension_pool_assignment_schema.go`
- `genesyscloud/user_extension_pool_assignment/resource_genesyscloud_user_extension_pool_assignment_utils.go`
- `genesyscloud/user_extension_pool_assignment/genesyscloud_user_extension_pool_assignment_proxy.go`
- Test files and documentation

#### Resource Schema:
```go
func UserExtensionPoolAssignmentResourceSchema() schema.Schema {
    return schema.Schema{
        Description: "Assigns a user extension to a specific extension pool",
        Attributes: map[string]schema.Attribute{
            "id": schema.StringAttribute{
                Computed: true,
            },
            "user_id": schema.StringAttribute{
                Required:    true,
                Description: "ID of the user",
            },
            "extension": schema.StringAttribute{
                Required:    true,
                Description: "Extension number",
            },
            "extension_pool_id": schema.StringAttribute{
                Required:    true,
                Description: "ID of the extension pool",
            },
        },
    }
}
```

**Estimated Effort: 25-30 hours**

### 6. Test Updates

**Complexity: Medium**

#### Update Existing Tests
```go
// Remove extension_pool_id assertions from phone_numbers
// Add new assertions for phone_extension_pools map

func TestAccResourceUser_extensionPools(t *testing.T) {
    // ... test setup ...
    
    resource.Test(t, resource.TestCase{
        Steps: []resource.TestStep{
            {
                Config: config,
                Check: resource.ComposeTestCheckFunc(
                    // OLD: resource.TestCheckResourceAttr(resourceName, "addresses.0.phone_numbers.0.extension_pool_id", poolId),
                    // NEW: Check computed map
                    resource.TestCheckResourceAttr(resourceName, "phone_extension_pools.PHONE|WORK|8501", poolId),
                ),
            },
        },
    })
}
```

#### New Tests for Assignment Resource (Option A)
- Create assignment resource tests
- Test assignment lifecycle (CRUD)
- Test dependency handling
- Test error scenarios

**Estimated Effort: 8-12 hours (Option B) or 15-20 hours (Option A)**

### 7. Documentation Updates

**Complexity: Low**

#### Update Resource Documentation
- Remove `extension_pool_id` from phone_numbers examples
- Add `phone_extension_pools` computed attribute documentation
- Update migration guide
- Add examples for new usage patterns

**Estimated Effort: 3-4 hours**

### 8. Migration Considerations

**Complexity: High**

#### State Migration
Users upgrading from SDKv2 will have `extension_pool_id` in their state. We need:

```go
// State upgrader to migrate old state format
func (r *UserFrameworkResource) UpgradeState(ctx context.Context) map[int64]resource.StateUpgrader {
    return map[int64]resource.StateUpgrader{
        0: {
            PriorSchema: &schema.Schema{
                // Old schema with extension_pool_id in phone_numbers
            },
            StateUpgrader: func(ctx context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
                // Extract extension_pool_id from phone_numbers
                // Move to top-level phone_extension_pools map
                // Remove from phone_numbers
            },
        },
    }
}
```

**Estimated Effort: 6-8 hours**

## Total Implementation Effort

### Option A (New Assignment Resource):
- **Schema Changes**: 3 hours
- **Model Updates**: 1 hour  
- **State Management**: 10 hours
- **Request Building**: 15 hours
- **New Resource**: 30 hours
- **Test Updates**: 20 hours
- **Documentation**: 4 hours
- **Migration**: 8 hours
- **Total: 91 hours (~2.3 weeks)**

### Option B (Auto-Assignment):
- **Schema Changes**: 3 hours
- **Model Updates**: 1 hour
- **State Management**: 10 hours  
- **Request Building**: 8 hours
- **Test Updates**: 12 hours
- **Documentation**: 4 hours
- **Migration**: 8 hours
- **Total: 46 hours (~1.2 weeks)**

## Implementation Risks

### High Risk Areas:
1. **State Migration**: Complex logic to move data between schema structures
2. **Set Identity Key Generation**: Must be consistent and unique
3. **API Integration**: Ensuring pool assignments work correctly
4. **Backward Compatibility**: Handling existing configurations during upgrade

### Medium Risk Areas:
1. **Test Coverage**: Ensuring all edge cases are covered
2. **Performance**: Map operations and pool lookups
3. **Error Handling**: Graceful handling of pool assignment failures

### Low Risk Areas:
1. **Schema Definition**: Straightforward attribute additions/removals
2. **Documentation**: Standard documentation updates

## Breaking Change Impact

### User Configuration Changes Required:
```hcl
# BEFORE (SDKv2 and current PF)
addresses {
  phone_numbers {
    extension         = "8501"
    extension_pool_id = genesyscloud_telephony_providers_edges_extension_pool.pool.id
  }
}

# AFTER Option A (New Resource)
addresses {
  phone_numbers {
    extension = "8501"
    # extension_pool_id removed
  }
}

resource "genesyscloud_user_extension_pool_assignment" "example" {
  user_id           = genesyscloud_user.example.id
  extension         = "8501"
  extension_pool_id = genesyscloud_telephony_providers_edges_extension_pool.pool.id
}

# AFTER Option B (Auto-Assignment)
addresses {
  phone_numbers {
    extension = "8501"
    # extension_pool_id removed - auto-assigned based on pool ranges
  }
}
# Must ensure extension pools exist with depends_on
depends_on = [genesyscloud_telephony_providers_edges_extension_pool.pool]
```

## Recommendation

**Solution 1 is NOT recommended** due to:

1. **Very High Implementation Complexity**: 46-91 hours of development
2. **Major Breaking Changes**: Requires all users to restructure configurations
3. **High Risk**: Complex state migration and new resource management
4. **User Experience Impact**: More complex configuration patterns
5. **Timeline Impact**: 1.2-2.3 weeks of development vs. days for Solution 2

While Solution 1 is technically the "cleanest" approach, the implementation cost and user impact make it impractical for the current migration timeline. **Solution 2 (Accept Set Identity Changes)** remains the recommended approach for immediate functionality restoration with minimal risk and effort.