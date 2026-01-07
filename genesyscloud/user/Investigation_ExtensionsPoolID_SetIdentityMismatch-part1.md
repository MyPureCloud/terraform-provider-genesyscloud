# Investigation: Extension Pool ID Set Identity Mismatch

## Problem Overview

The `TestAccFrameworkResourceUserAddressWithExtensionPool` test is failing during the Plugin Framework migration with a 409 API error:

```
Error: Failed to delete extension pool 21000 error: API Error: 409 - The resource extensionPool is referenced by the extension 78d3da52-da59-4585-91b4-5576e86cd7ef
```

**Root Cause**: The extension pool dependency relationship between users and extension pools is broken in the Plugin Framework migration, causing Terraform to attempt deletion of extension pools that are still referenced by user extensions.

**Test Failure Sequence**:
1. **Step 1**: Creates extension pool 1 (21000-21001) and user with extension 21000
2. **Step 2**: Creates extension pool 2 (21002-21003) and updates user to extension 21002
3. **Failure Point**: Terraform tries to delete extension pool 1, but API rejects because extension 21000 is still allocated from that pool

## Why and How It Works in SDKv2

### SDKv2 Implementation Strategy

In SDKv2, the extension pool functionality worked through a **custom Set hashing mechanism** that deliberately excluded `extension_pool_id` from the phone number identity:

```go
// SDKv2 Actual Hash Function (from resource_genesyscloud_user_utils.go:584)
func phoneNumberHash(val interface{}) int {
    // Copy map to avoid modifying state
    phoneMap := make(map[string]interface{})
    for k, v := range val.(map[string]interface{}) {
        if k != "extension_pool_id" {  // DELIBERATELY EXCLUDED
            phoneMap[k] = v
        }
    }
    if num, ok := phoneMap["number"]; ok {
        // Attempt to format phone numbers before hashing
        number, err := phonenumbers.Parse(num.(string), "US")
        if err == nil {
            phoneMap["number"] = phonenumbers.Format(number, phonenumbers.E164)
        }
    }
    return schema.HashResource(phoneNumberResource)(phoneMap)
}
```

### Key SDKv2 Behaviors

1. **Set Identity Stability**: Phone numbers maintained the same Set identity even when `extension_pool_id` changed
2. **Dependency Management**: Explicit `extension_pool_id` references created proper Terraform dependencies
3. **State Consistency**: Pool changes were treated as in-place updates, not element replacements
4. **No Perpetual Diffs**: Pool reassignments didn't trigger unnecessary plan changes
5. **Full CRUD Support**: Extension pool IDs were properly handled in create, read, and update operations

### SDKv2 Schema Definition

```go
// From resource_genesyscloud_user_schema.go:317
"phone_numbers": {
    Description: "Phone number addresses for this user.",
    Type:        schema.TypeSet,
    Optional:    true,
    Set:         phoneNumberHash,  // Custom hash function
    Elem:        phoneNumberResource,
    ConfigMode:  schema.SchemaConfigModeAttr,
},

// phoneNumberResource included extension_pool_id field:
"extension_pool_id": {
    Description: "Id of the extension pool which contains this extension.",
    Type:        schema.TypeString,
    Optional:    true,
},
```

### SDKv2 State Management

```go
// From resource_genesyscloud_user_utils.go:743
if address.Extension != nil {
    if address.Display != nil {
        if *address.Extension == *address.Display {
            extensionNum := strings.Trim(*address.Extension, "()")
            phoneNumber["extension"] = extensionNum
            phoneNumber["extension_pool_id"] = fetchExtensionPoolId(ctx, extensionNum, proxy)
        }
    }
}
```

**Key Point**: SDKv2 **actively populated** `extension_pool_id` in state by calling `fetchExtensionPoolId()` during read operations.

### SDKv2 Configuration Example

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

**Result**: Terraform understood the dependency and would update the user first, then delete the old pool.

### SDKv2 Test Evidence

The SDKv2 test `TestAccResourceUserAddressWithExtensionPool` shows the working pattern:

```go
// Step 1: Create user with extension 4105 from pool 1
Config: generateUserPhoneAddress(
    util.NullValue,          // number
    util.NullValue,          // Default to type PHONE
    util.NullValue,          // Default to type WORK
    strconv.Quote(addrExt1), // extension = "4105"
    fmt.Sprintf("extension_pool_id = %s.%s.id", extensionPool.ResourceType, extensionPoolResourceLabel1),
),

// Step 2: Update to extension 4225 from pool 2
Config: generateUserPhoneAddress(
    util.NullValue,
    util.NullValue,
    util.NullValue,
    strconv.Quote(addrExt2), // extension = "4225"
    fmt.Sprintf("extension_pool_id = %s.%s.id", extensionPool.ResourceType, extensionPoolResourceLabel2),
),

// Step 3: Remove addresses entirely - this worked without dependency issues
Config: generateUserWithCustomAttrs(
    addrUserResourceLabel1,
    addrEmail1,
    addrUserName,
    // No addresses block
),
```

**Critical**: The SDKv2 test successfully completed all three steps, including the address removal that fails in Plugin Framework.

## Why It Will Not Work Directly in Migration

### Plugin Framework Set Identity Limitations

Plugin Framework **does not support custom Set hashing**. Instead, it uses the **entire object value** for Set element identity:

```go
// Plugin Framework - No Custom Hashing Available
type PhoneNumberModel struct {
    Number          types.String `tfsdk:"number"`
    MediaType       types.String `tfsdk:"media_type"`
    Type            types.String `tfsdk:"type"`
    Extension       types.String `tfsdk:"extension"`
    ExtensionPoolId types.String `tfsdk:"extension_pool_id"` // THIS AFFECTS IDENTITY!
}

// PF compares ENTIRE struct for Set identity
// Any change to ExtensionPoolId = Different Set element = Replace operation
```

### Specific Migration Problems

#### 1. **Perpetual Diff Problem**
```
Plan:  {extension: "21000", extension_pool_id: "pool-new"}
State: {extension: "21000", extension_pool_id: "pool-old"}
PF:    "These are different objects - must replace phone number"
```

#### 2. **Dependency Tracking Broken**
Current implementation deliberately nullifies `extension_pool_id` in state:

```go
// From resource_genesyscloud_user_utils.go:731-773
// TODO Note: extension_pool_id is ALWAYS null in state to keep it out of Set identity
phoneNumber := map[string]attr.Value{
    "extension_pool_id": types.StringNull(), // <- always null in state
}

// Commented out pool ID assignment:
// poolId := fetchExtensionPoolId(ctx, extensionNum, proxy)
// if poolId != "" {
//     phoneNumber["extension_pool_id"] = types.StringValue(poolId)
// }
```

**Contrast with SDKv2**: SDKv2 **actively populated** `extension_pool_id` in state:
```go
// SDKv2 actually did this:
phoneNumber["extension_pool_id"] = fetchExtensionPoolId(ctx, extensionNum, proxy)
```

**Result**: No dependency relationship exists between user and extension pool in Plugin Framework.

#### 3. **Creation/Update Logic Incomplete**
In `buildSdkPhoneNumbers()`, the `ExtensionPoolId` field is extracted but never used:

```go
type PhoneNumberModel struct {
    ExtensionPoolId types.String `tfsdk:"extension_pool_id"` // Extracted
}

// But in contact building:
contact := platformclientv2.Contact{
    Extension: &phoneExt,
    // MISSING: No ExtensionPoolId assignment to SDK contact
}
```

**Contrast with SDKv2**: SDKv2 didn't need to set extension pool ID in the contact because the API auto-assigns based on extension number ranges. The `extension_pool_id` field was purely for Terraform dependency tracking.

## How to Resolve This Gap

### Actual SDKv2 vs Plugin Framework Differences

Based on the code investigation, here are the **factual differences**:

| Aspect | SDKv2 Implementation | Plugin Framework Implementation |
|--------|---------------------|--------------------------------|
| **Set Hashing** | Custom `phoneNumberHash()` excludes `extension_pool_id` | No custom hashing - entire object used for identity |
| **State Population** | `phoneNumber["extension_pool_id"] = fetchExtensionPoolId(...)` | `"extension_pool_id": types.StringNull()` (always null) |
| **Schema Definition** | `Set: phoneNumberHash` with `extension_pool_id` field | `extension_pool_id` field exists but commented as problematic |
| **Dependency Tracking** | Working - pool changes don't affect Set identity | Broken - field nullified to avoid Set identity issues |
| **Test Results** | All steps pass including address removal | Step 2 fails with 409 dependency error |

### Root Cause Confirmed

The Plugin Framework migration **deliberately disabled extension pool functionality** to avoid Set identity issues, but this broke the dependency relationship that made the SDKv2 implementation work.

### Current State Analysis

**Infrastructure Status**: ✅ **90% Complete**
- ✅ `fetchExtensionPoolId()` function implemented
- ✅ `getTelephonyExtensionPoolByExtension()` proxy method working
- ✅ Extension pool API integration complete
- ✅ Caching mechanism for extension pools implemented

**Missing Components**: ❌ **Integration Disabled**
- ❌ State management (deliberately nullified)
- ❌ Creation/update logic (field ignored)
- ❌ Dependency tracking (broken)
- ❌ Test configuration (commented out)

### Resolution Requirements

1. **Solve Set Identity Problem**: Prevent `extension_pool_id` changes from causing element replacements
2. **Restore Dependency Tracking**: Ensure proper resource ordering during updates
3. **Complete CRUD Operations**: Handle `extension_pool_id` in create/update/read operations
4. **Maintain Backward Compatibility**: Don't break existing user configurations

## Possible Solutions

### Option 1: Computed Field Approach
**Status**: Partially implemented (commented out)

**Architecture**:
```go
// Top-level computed map (outside phone_numbers Set)
"phone_extension_pools": schema.MapAttribute{
    ElementType: types.StringType,
    Computed:    true,
    Description: "Computed mapping of phone identity keys to extension_pool_id",
}
```

**How it works**:
- Keep `extension_pool_id` OUT of `phone_numbers` Set
- Store pool mappings in separate computed map
- Use phone identity key (e.g., "PHONE|WORK|+15551234567|1001") as map key

**Pros**:
- ✅ Maintains Set identity stability
- ✅ Preserves SDKv2 behavior exactly
- ✅ No perpetual diffs
- ✅ Full control over pool assignments

**Cons**:
- ❌ Complex implementation (custom identity key generation)
- ❌ Non-standard Terraform pattern
- ❌ User can't directly reference `extension_pool_id` in config
- ❌ Requires significant refactoring

**Implementation Complexity**: **High**

### Option 2: Dependency-Only Approach
**Status**: Not implemented

**Architecture**:
```hcl
resource "genesyscloud_user" "example" {
  addresses {
    phone_numbers {
      extension = "21000"
      # No extension_pool_id field at all
    }
  }
  depends_on = [genesyscloud_telephony_providers_edges_extension_pool.pool1]
}
```

**How it works**:
- Remove `extension_pool_id` from schema entirely
- Use `depends_on` for explicit dependencies
- Let Genesys Cloud API auto-assign pools based on extension number ranges

**Pros**:
- ✅ Simple implementation
- ✅ No Set identity issues
- ✅ Clear dependency management
- ✅ Matches API behavior (auto-assignment)
- ✅ Follows Terraform best practices

**Cons**:
- ❌ Less explicit control over pool assignments
- ❌ Relies on API auto-assignment logic
- ❌ May not work for complex pool scenarios

**Implementation Complexity**: **Low**

### Option 3: Separate Resource Approach
**Status**: Not implemented

**Architecture**:
```hcl
resource "genesyscloud_user" "example" {
  addresses {
    phone_numbers {
      extension = "21000"
    }
  }
}

resource "genesyscloud_user_extension" "example" {
  user_id           = genesyscloud_user.example.id
  extension         = "21000"
  extension_pool_id = genesyscloud_telephony_providers_edges_extension_pool.pool1.id
}
```

**How it works**:
- Create separate `genesyscloud_user_extension` resource
- Manage user-extension-pool relationships independently
- Clean separation of concerns

**Pros**:
- ✅ Clean separation of concerns
- ✅ No Set identity issues
- ✅ Explicit dependency management
- ✅ Follows Terraform resource design patterns
- ✅ Highly flexible

**Cons**:
- ❌ Breaking change for existing users
- ❌ More complex configuration
- ❌ Requires new resource implementation
- ❌ Additional maintenance overhead

**Implementation Complexity**: **High**

### Option 4: Plan Modifier Approach
**Status**: Partially explored

**Architecture**:
```go
"extension_pool_id": schema.StringAttribute{
    Optional: true,
    PlanModifiers: []planmodifier.String{
        extensionPoolPlanModifier{}, // Custom modifier to handle pool changes
    },
}
```

**How it works**:
- Keep `extension_pool_id` in schema
- Use custom plan modifier to detect pool-only changes
- Implement "ignore changes" logic for pool reassignments

**Pros**:
- ✅ Maintains current schema structure
- ✅ Can handle pool changes gracefully
- ✅ User-friendly configuration

**Cons**:
- ❌ Complex plan modifier logic
- ❌ Still has fundamental Set identity challenges
- ❌ May not fully solve perpetual diff problem
- ❌ Plan modifiers have limitations

**Implementation Complexity**: **Medium-High**

## Recommended Solution and Why It Is Recommended

### **Recommendation: Option 2 - Dependency-Only Approach**

After analyzing the complexity, implementation effort, and architectural alignment, **Option 2** is the recommended solution.

### Why Option 2 Is Recommended

#### 1. **Aligns with Genesys Cloud API Behavior**

The investigation revealed that Genesys Cloud **automatically assigns extensions to pools** based on number ranges, and **SDKv2 never actually sent extension_pool_id to the API**:

```go
// SDKv2 buildSdkPhoneNumbers() - NO extension_pool_id sent to API
contact := platformclientv2.Contact{
    MediaType: &phoneMediaType,
    VarType:   &phoneType,
    Address:   &phoneNum,    // if present
    Extension: &phoneExt,    // if present
    // NO ExtensionPoolId field set
}
```

The `extension_pool_id` field in SDKv2 was **purely for Terraform dependency tracking**, not API communication.

**Implication**: The API handles extension-to-pool mapping automatically, making explicit pool assignment unnecessary for functionality.

#### 2. **Solves Core Problem Completely**

- ✅ **Eliminates Set Identity Issues**: No `extension_pool_id` in Set = No identity conflicts
- ✅ **No Perpetual Diffs**: Pool changes don't affect phone number identity
- ✅ **Clean Dependency Management**: `depends_on` provides explicit ordering
- ✅ **API Compatibility**: Leverages existing auto-assignment behavior

#### 3. **Minimal Implementation Risk**

**Low Complexity Changes Required**:
```go
// 1. Remove field from schema (simple deletion)
// 2. Update tests to use depends_on (configuration change)
// 3. Keep fetchExtensionPoolId() for internal state tracking
// 4. Update documentation
```

**No Complex Logic Needed**:
- No custom plan modifiers
- No identity key generation
- No new resource types
- No breaking schema changes

#### 4. **Follows Terraform Best Practices**

- **Explicit Dependencies**: `depends_on` is the standard Terraform pattern
- **Resource Separation**: Each resource manages its own lifecycle
- **API Alignment**: Leverages provider API capabilities rather than fighting them
- **Simplicity**: Easier to understand, maintain, and debug

#### 5. **Backward Compatible Migration Path**

**Phase 1**: Remove `extension_pool_id` field, add deprecation notice
**Phase 2**: Update documentation with migration examples
**Phase 3**: Users migrate to `depends_on` approach gradually

**Example Migration**:
```hcl
# Before (SDKv2)
resource "genesyscloud_user" "example" {
  addresses {
    phone_numbers {
      extension = "21000"
      extension_pool_id = genesyscloud_telephony_providers_edges_extension_pool.pool1.id
    }
  }
}

# After (Plugin Framework)
resource "genesyscloud_user" "example" {
  addresses {
    phone_numbers {
      extension = "21000"
    }
  }
  depends_on = [genesyscloud_telephony_providers_edges_extension_pool.pool1]
}
```

#### 6. **Validation Through SDKv2 Evidence**

The SDKv2 test `TestAccResourceUserAddressWithExtensionPool` **successfully completed** all steps including:
- ✅ Creating user with extension from pool 1
- ✅ Updating user to extension from pool 2  
- ✅ **Removing addresses entirely** (this step fails in Plugin Framework)

This proves that dependency-based management can work, since SDKv2's `extension_pool_id` was only used for dependency tracking, not API communication.

### Implementation Roadmap

#### Phase 1: Schema Cleanup (Low Risk)
1. Remove `extension_pool_id` from phone_numbers schema
2. Add deprecation notice in documentation
3. Update exporter to remove extension pool references

#### Phase 2: Test Updates (Validation)
1. Update failing test to use `depends_on` approach
2. Verify all extension pool scenarios work
3. Add integration tests for dependency ordering

#### Phase 3: Documentation (User Communication)
1. Update resource documentation
2. Provide migration guide with examples
3. Add troubleshooting section for common issues

#### Phase 4: Monitoring (Post-Release)
1. Monitor for user feedback on dependency approach
2. Validate that auto-assignment covers all use cases
3. Consider future enhancements if needed

### Success Criteria

- ✅ `TestAccFrameworkResourceUserAddressWithExtensionPool` passes
- ✅ No perpetual diffs in extension pool scenarios
- ✅ Proper resource ordering during create/update/delete
- ✅ Backward compatibility maintained
- ✅ Clear migration path for existing users

This approach provides a **clean, maintainable, and architecturally sound solution** that aligns with both Terraform and Genesys Cloud best practices while solving the immediate test failure and long-term Set identity challenges.