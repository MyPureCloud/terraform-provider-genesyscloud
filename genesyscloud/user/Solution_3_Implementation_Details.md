# Solution 3: Dependency-Only Approach - Implementation Details

## Overview

This document provides detailed implementation analysis for Solution 3 (Dependency-Only Approach) to help finalize the decision on which solution to pursue for the extension pool migration issue.

## Solution Summary

Remove `extension_pool_id` field entirely from the user resource schema and rely on API auto-assignment based on extension number ranges. Users manage extension pools as separate resources with explicit Terraform dependencies, and the API automatically assigns extensions to the appropriate pools based on number ranges.

## Required Code Changes

### 1. Schema Changes (`resource_genesyscloud_user_schema.go`)

**Complexity: Very Low**

#### Remove extension_pool_id Field Entirely
```go
// REMOVE this entire block from phone_numbers NestedObject
"extension_pool_id": schema.StringAttribute{
    Description:   "Id of the extension pool which contains this extension.",
    Optional:      true,
    PlanModifiers: []planmodifier.String{phoneplan.NullIfEmpty{}},
},
```

**After removal, phone_numbers schema becomes:**
```go
"phone_numbers": schema.SetNestedBlock{
    NestedObject: schema.NestedBlockObject{
        Attributes: map[string]schema.Attribute{
            "number": schema.StringAttribute{
                Description: "Phone number. Phone number must be in an E.164 number format.",
                Optional:    true,
                Validators:  []validator.String{validators.FWValidatePhoneNumber()},
            },
            "media_type": schema.StringAttribute{
                Description: "Media type of phone number (SMS | PHONE).",
                Optional:    true,
                Computed:    true,
                Default:     stringdefault.StaticString("PHONE"),
                Validators: []validator.String{
                    stringvalidator.OneOf("PHONE", "SMS"),
                },
            },
            "type": schema.StringAttribute{
                Description: "Type of number (WORK | WORK2 | WORK3 | WORK4 | HOME | MOBILE | OTHER).",
                Optional:    true,
                Computed:    true,
                Default:     stringdefault.StaticString("WORK"),
                Validators: []validator.String{
                    stringvalidator.OneOf("WORK", "WORK2", "WORK3", "WORK4", "HOME", "MOBILE", "OTHER"),
                },
            },
            "extension": schema.StringAttribute{
                Description: "Phone number extension",
                Optional:    true,
            },
            // extension_pool_id REMOVED - no longer exists
        },
    },
},
```

**Estimated Effort: 0.5 hours**

### 2. Model Updates (`resource_genesyscloud_user_utils.go`)

**Complexity: Very Low**

#### Remove ExtensionPoolId from PhoneNumberModel
```go
type PhoneNumberModel struct {
    Number          types.String `tfsdk:"number"`
    MediaType       types.String `tfsdk:"media_type"`
    Type            types.String `tfsdk:"type"`
    Extension       types.String `tfsdk:"extension"`
    // REMOVE: ExtensionPoolId types.String `tfsdk:"extension_pool_id"`
}
```

**Estimated Effort: 0.5 hours**

### 3. State Management Updates (`resource_genesyscloud_user_utils.go`)

**Complexity: Low**

#### Update flattenUserAddresses Function

**Current Code (lines 738-774):**
```go
// TODO Note: extension_pool_id is ALWAYS null in state to keep it out of Set identity
// and match SDKv2's hash behavior (which ignored extension_pool_id).
phoneNumber := map[string]attr.Value{
    "media_type":        types.StringValue(media),
    "number":            types.StringNull(),
    "extension":         types.StringNull(),
    "type":              types.StringNull(),
    "extension_pool_id": types.StringNull(), // <- always null in state -- TODO
}
```

**Updated Code:**
```go
// Extension pool assignment is handled automatically by API based on extension ranges
phoneNumber := map[string]attr.Value{
    "media_type": types.StringValue(media),
    "number":     types.StringNull(),
    "extension":  types.StringNull(),
    "type":       types.StringNull(),
    // extension_pool_id field removed entirely
}
```

#### Update phoneObjType Definition
```go
phoneObjType := types.ObjectType{AttrTypes: map[string]attr.Type{
    "number":     types.StringType,
    "media_type": types.StringType,
    "type":       types.StringType,
    "extension":  types.StringType,
    // REMOVE: "extension_pool_id": types.StringType,
}}
```

#### Remove Extension Pool Logic
```go
// REMOVE all this logic since we no longer track pool IDs:
// poolId := fetchExtensionPoolId(ctx, extensionNum, proxy)
// if poolId != "" {
//     phoneNumber["extension_pool_id"] = types.StringValue(poolId)
// }
```

**Estimated Effort: 2-3 hours**

### 4. Request Building Updates (`resource_genesyscloud_user_utils.go`)

**Complexity: Very Low**

#### Update buildSdkPhoneNumbers Function

**Current Code:**
```go
type PhoneNumberModel struct {
    Number          types.String `tfsdk:"number"`
    MediaType       types.String `tfsdk:"media_type"`
    Type            types.String `tfsdk:"type"`
    Extension       types.String `tfsdk:"extension"`
    ExtensionPoolId types.String `tfsdk:"extension_pool_id"` // This field is removed
}
```

**Updated Code:**
```go
type PhoneNumberModel struct {
    Number    types.String `tfsdk:"number"`
    MediaType types.String `tfsdk:"media_type"`
    Type      types.String `tfsdk:"type"`
    Extension types.String `tfsdk:"extension"`
    // ExtensionPoolId removed - no longer processed
}
```

**Processing Logic (no changes needed):**
```go
// Optional field: extension
if !phone.Extension.IsNull() && !phone.Extension.IsUnknown() {
    phoneExt := phone.Extension.ValueString()
    if phoneExt != "" {
        contact.Extension = &phoneExt
        // API will automatically assign to appropriate pool based on extension number
    }
}

// No extension_pool_id processing needed - field doesn't exist
```

**Estimated Effort: 1 hour**

### 5. Remove Extension Pool Helper Functions

**Complexity: Low**

#### Functions to Remove/Modify:
```go
// This function can be removed since we no longer fetch pool IDs
func fetchExtensionPoolId(ctx context.Context, extNum string, proxy *userProxy) string {
    // ENTIRE FUNCTION CAN BE REMOVED
}
```

#### Proxy Functions to Remove:
```go
// These can be removed from userProxy if not used elsewhere:
type getTelephonyExtensionPoolByExtensionFunc func(ctx context.Context, p *userProxy, extNum string) (*platformclientv2.Extensionpool, *platformclientv2.APIResponse, error)

func (p *userProxy) getTelephonyExtensionPoolByExtension(ctx context.Context, extNum string) (*platformclientv2.Extensionpool, *platformclientv2.APIResponse, error) {
    // CAN BE REMOVED IF NOT USED ELSEWHERE
}

func getTelephonyExtensionPoolByExtensionFn(_ context.Context, p *userProxy, extNum string) (*platformclientv2.Extensionpool, *platformclientv2.APIResponse, error) {
    // CAN BE REMOVED IF NOT USED ELSEWHERE
}
```

**Note:** Need to verify these functions aren't used by other resources before removing.

**Estimated Effort: 2-3 hours (including verification)**

### 6. Test Updates (`resource_genesyscloud_user_test.go`)

**Complexity: Medium**

#### Update Test Configurations

**Current Test Config:**
```go
phone_numbers {
    extension         = "8501"
    media_type        = "PHONE"
    type              = "WORK"
    extension_pool_id = "${genesyscloud_telephony_providers_edges_extension_pool.suresh_EPID2.id}"
}
```

**Updated Test Config:**
```go
resource "genesyscloud_user" "test_user" {
    name  = "Test User"
    email = "test@example.com"
    
    addresses {
        phone_numbers {
            extension  = "8501"  # Must be within pool range
            media_type = "PHONE"
            type       = "WORK"
            # extension_pool_id removed
        }
    }
    
    # Explicit dependency to ensure pool exists before user creation
    depends_on = [genesyscloud_telephony_providers_edges_extension_pool.test_pool]
}

resource "genesyscloud_telephony_providers_edges_extension_pool" "test_pool" {
    start_number = "8500"
    end_number   = "8699"
}
```

#### Update Test Assertions

**Remove Extension Pool ID Assertions:**
```go
// REMOVE these assertions:
// resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel, "addresses.0.phone_numbers.0.extension_pool_id", poolId),
// resource.TestCheckResourceAttrPair(ResourceType+"."+userResourceLabel, "addresses.0.phone_numbers.0.extension_pool_id", extensionPool.ResourceType+"."+extensionPoolResourceLabel, "id"),
```

**Keep Extension Assertions:**
```go
resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel, "addresses.0.phone_numbers.0.extension", "8501"),
resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel, "addresses.0.phone_numbers.0.media_type", "PHONE"),
resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel, "addresses.0.phone_numbers.0.type", "WORK"),
```

#### Add Dependency Validation Tests
```go
func TestAccResourceUser_extensionDependency(t *testing.T) {
    // Test that extension assignment works with dependency-only approach
    // Test that extension gets assigned to correct pool based on number range
    // Test error handling when no pool exists for extension range
}
```

**Estimated Effort: 6-8 hours**

### 7. Documentation Updates

**Complexity: Medium**

#### Update Resource Documentation

**Remove extension_pool_id References:**
```markdown
<!-- REMOVE this section -->
- `extension_pool_id` - (Optional) Id of the extension pool which contains this extension.
```

**Add Dependency Pattern Documentation:**
```markdown
## Extension Pool Assignment

Extensions are automatically assigned to extension pools based on the extension number and the pool's number range. To ensure proper assignment:

1. Create extension pools with appropriate number ranges
2. Use `depends_on` to ensure pools exist before creating users with extensions
3. Ensure extension numbers fall within the desired pool's range

### Example:

```hcl
resource "genesyscloud_telephony_providers_edges_extension_pool" "sales_pool" {
  start_number = "1000"
  end_number   = "1999"
}

resource "genesyscloud_user" "sales_user" {
  name  = "Sales User"
  email = "sales@example.com"
  
  addresses {
    phone_numbers {
      extension  = "1001"  # Will be assigned to sales_pool automatically
      media_type = "PHONE"
      type       = "WORK"
    }
  }
  
  depends_on = [genesyscloud_telephony_providers_edges_extension_pool.sales_pool]
}
```

### Important Notes:

- Extension pool assignment is automatic based on number ranges
- Multiple pools with overlapping ranges may cause unpredictable assignment
- Extensions outside any pool range will fail to be assigned
- Use explicit `depends_on` to ensure proper creation order
```

#### Update Migration Guide
```markdown
## Breaking Change: Extension Pool Assignment

The `extension_pool_id` field has been removed from phone numbers. Extension pool assignment is now handled automatically based on extension number ranges.

### Migration Required:

**Before:**
```hcl
addresses {
  phone_numbers {
    extension         = "1001"
    extension_pool_id = genesyscloud_telephony_providers_edges_extension_pool.pool.id
  }
}
```

**After:**
```hcl
addresses {
  phone_numbers {
    extension = "1001"  # Pool assignment is automatic
  }
}

# Ensure pool exists before user creation
depends_on = [genesyscloud_telephony_providers_edges_extension_pool.pool]
```
```

**Estimated Effort: 4-5 hours**

### 8. Validation and Error Handling

**Complexity: Medium**

#### Add Extension Range Validation

Since users can no longer explicitly specify pools, we need better validation:

```go
func validateExtensionInPoolRange(extension string, proxy *userProxy) pfdiag.Diagnostics {
    var diagnostics pfdiag.Diagnostics
    
    // Get all extension pools
    pools, err := getAllExtensionPools(proxy)
    if err != nil {
        diagnostics.AddError("Failed to fetch extension pools", err.Error())
        return diagnostics
    }
    
    extNum, err := strconv.Atoi(extension)
    if err != nil {
        diagnostics.AddError("Invalid Extension", 
            fmt.Sprintf("Extension %s is not a valid number", extension))
        return diagnostics
    }
    
    // Check if extension falls within any pool range
    var matchingPools []string
    for _, pool := range pools {
        startNum, _ := strconv.Atoi(*pool.StartNumber)
        endNum, _ := strconv.Atoi(*pool.EndNumber)
        
        if extNum >= startNum && extNum <= endNum {
            matchingPools = append(matchingPools, *pool.Id)
        }
    }
    
    if len(matchingPools) == 0 {
        diagnostics.AddWarning("Extension Not in Pool Range", 
            fmt.Sprintf("Extension %s does not fall within any existing extension pool range", extension))
    } else if len(matchingPools) > 1 {
        diagnostics.AddWarning("Multiple Pool Matches", 
            fmt.Sprintf("Extension %s matches multiple extension pools - assignment may be unpredictable", extension))
    }
    
    return diagnostics
}
```

#### Add to buildSdkPhoneNumbers:
```go
// Validate extension is in a pool range
if !phone.Extension.IsNull() && !phone.Extension.IsUnknown() {
    phoneExt := phone.Extension.ValueString()
    if phoneExt != "" {
        contact.Extension = &phoneExt
        
        // Validate extension is in a pool range
        validationDiags := validateExtensionInPoolRange(phoneExt, proxy)
        diagnostics.Append(validationDiags...)
    }
}
```

**Estimated Effort: 5-6 hours**

### 9. Remove Exporter References

**Complexity: Low**

#### Update UserExporter

**Current Code:**
```go
RefAttrs: map[string]*resourceExporter.RefAttrSettings{
    "manager":                                   {RefType: ResourceType},
    "division_id":                               {RefType: "genesyscloud_auth_division"},
    "routing_skills.skill_id":                   {RefType: "genesyscloud_routing_skill"},
    "routing_languages.language_id":             {RefType: "genesyscloud_routing_language"},
    "locations.location_id":                     {RefType: "genesyscloud_location"},
    "addresses.phone_numbers.extension_pool_id": {RefType: "genesyscloud_telephony_providers_edges_extension_pool"},
},
```

**Updated Code:**
```go
RefAttrs: map[string]*resourceExporter.RefAttrSettings{
    "manager":                         {RefType: ResourceType},
    "division_id":                     {RefType: "genesyscloud_auth_division"},
    "routing_skills.skill_id":         {RefType: "genesyscloud_routing_skill"},
    "routing_languages.language_id":   {RefType: "genesyscloud_routing_language"},
    "locations.location_id":           {RefType: "genesyscloud_location"},
    // REMOVE: "addresses.phone_numbers.extension_pool_id": {RefType: "genesyscloud_telephony_providers_edges_extension_pool"},
},
```

**Estimated Effort: 0.5 hours**

## Total Implementation Effort

### Summary:
- **Schema Changes**: 0.5 hours (remove field)
- **Model Updates**: 0.5 hours (remove field)
- **State Management**: 3 hours (clean up flatten function)
- **Request Building**: 1 hour (remove processing)
- **Remove Helper Functions**: 3 hours (cleanup + verification)
- **Test Updates**: 8 hours (rewrite test patterns)
- **Documentation**: 5 hours (major updates + migration guide)
- **Validation Logic**: 6 hours (new validation functions)
- **Exporter Updates**: 0.5 hours (remove references)
- **Total: 27 hours (~3.5 days)**

## Implementation Risks

### High Risk Areas:
1. **Breaking Changes**: 100% of users with extension pools must restructure configs
2. **Loss of Explicit Control**: Users can't specify which pool to use for overlapping ranges
3. **Dependency Management**: Users must manually manage creation order
4. **Validation Complexity**: Need robust validation since explicit assignment is removed

### Medium Risk Areas:
1. **API Behavior**: Relying on API auto-assignment behavior that may not be deterministic
2. **Multiple Pool Conflicts**: Unpredictable assignment when extension ranges overlap
3. **Error Messages**: Less clear error messages when assignments fail
4. **Testing**: Need comprehensive tests for various pool range scenarios

### Low Risk Areas:
1. **Code Simplification**: Removing code is generally safer than adding complex logic
2. **Set Identity**: Completely eliminates Set identity issues
3. **Performance**: Fewer API calls and simpler state management

## Breaking Change Impact

### Configuration Changes Required:

**Before (SDKv2 and current PF):**
```hcl
resource "genesyscloud_user" "example" {
  addresses {
    phone_numbers {
      extension         = "1001"
      extension_pool_id = genesyscloud_telephony_providers_edges_extension_pool.pool.id
    }
  }
}

resource "genesyscloud_telephony_providers_edges_extension_pool" "pool" {
  start_number = "1000"
  end_number   = "1999"
}
```

**After (Solution 3):**
```hcl
resource "genesyscloud_telephony_providers_edges_extension_pool" "pool" {
  start_number = "1000"
  end_number   = "1999"
}

resource "genesyscloud_user" "example" {
  addresses {
    phone_numbers {
      extension = "1001"  # Must be within pool range
      # extension_pool_id removed
    }
  }
  
  depends_on = [genesyscloud_telephony_providers_edges_extension_pool.pool]
}
```

### User Impact:
- **100% of extension pool users** must modify their configurations
- **Loss of explicit pool control** - can't choose specific pool for overlapping ranges
- **New dependency management** - must use `depends_on` explicitly
- **Different mental model** - from explicit assignment to range-based auto-assignment

## API Behavior Analysis

### Current Understanding:
Based on the existing `getTelephonyExtensionPoolByExtensionFn` function:

```go
// API searches all pools and returns first match where:
// extNum > startNum && extNum < endNum
for _, pool := range allPools {
    startNum, _ := strconv.Atoi(*pool.StartNumber)
    endNum, _ := strconv.Atoi(*pool.EndNumber)
    
    if extNumInt > startNum && extNumInt < endNum {
        return &pool, apiResponse, nil  // Returns FIRST match
    }
}
```

### Potential Issues:
1. **Non-deterministic assignment** when multiple pools have overlapping ranges
2. **API may not auto-assign** - might require explicit pool specification
3. **Error handling** when no pool matches extension range
4. **Pool creation timing** - extension assignment might fail if pool doesn't exist yet

### Research Required:
- **Does the API actually auto-assign extensions to pools?**
- **What happens when multiple pools match an extension?**
- **How does the API handle extensions outside any pool range?**
- **Is there a separate API call needed for pool assignment?**

**Estimated Research Effort: 4-6 hours**

## Comparison with Other Solutions

| Aspect | Solution 1 | Solution 2 | Solution 3 |
|--------|------------|------------|------------|
| **Development Time** | 46-91 hours | 22 hours | 27 hours |
| **Breaking Changes** | Yes (major) | No | Yes (major) |
| **User Impact** | High | None | High |
| **Implementation Risk** | High | Low | Medium |
| **Code Complexity** | Very High | Low | Low |
| **Functionality Loss** | None | None | Explicit pool control |
| **Set Identity Issues** | Eliminated | Accepted | Eliminated |

## Pros and Cons Summary

### Pros:
- ✅ **Eliminates Set identity issues completely**
- ✅ **Simplest code implementation** (mostly removing code)
- ✅ **Clean Plugin Framework approach**
- ✅ **Leverages API auto-assignment behavior**
- ✅ **No complex state management needed**
- ✅ **Follows Terraform dependency patterns**

### Cons:
- ❌ **Major breaking change** requiring config restructuring
- ❌ **Loss of explicit pool control** for users
- ❌ **Potential ambiguity** with overlapping pool ranges
- ❌ **Dependency management complexity** for users
- ❌ **API behavior uncertainty** - need to verify auto-assignment works
- ❌ **Different user experience** from SDKv2 expectations
- ❌ **No visibility** of pool assignments in state

## Recommendation

**Solution 3 is NOT recommended** for the following reasons:

### **Major Concerns:**
1. **Breaking Change Impact**: Requires 100% of extension pool users to restructure configurations
2. **Loss of Functionality**: Users lose explicit control over pool assignments
3. **API Uncertainty**: Unclear if API actually supports reliable auto-assignment
4. **User Experience Degradation**: More complex dependency management, less visibility

### **Better Alternatives:**
- **Solution 2**: Restores functionality in 22 hours with zero breaking changes
- **Solution 1**: Provides clean architecture but with much higher effort (46-91 hours)

### **When Solution 3 Might Be Considered:**
- If API research confirms reliable auto-assignment behavior
- If organization prioritizes clean architecture over user experience
- If breaking changes are acceptable for long-term architectural benefits
- As a future enhancement after Solution 2 restores immediate functionality

**Conclusion:** While Solution 3 has technical merit and eliminates Set identity issues, the combination of breaking changes, functionality loss, and API uncertainty make it unsuitable for immediate deployment. **Solution 2 remains the recommended approach** for restoring functionality quickly with minimal risk and zero user impact.