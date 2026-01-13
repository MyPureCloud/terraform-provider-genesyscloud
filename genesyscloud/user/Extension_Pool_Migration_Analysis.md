# Extension Pool Migration Analysis: SDKv2 to Plugin Framework

## Executive Summary

The migration of the `genesyscloud_user` resource from Terraform SDKv2 to Plugin Framework has encountered a critical issue with `extension_pool_id` functionality. While regular phone number operations work correctly in Plugin Framework, any configuration involving extension pools fails with "Provider produced inconsistent result after apply" errors. This document analyzes the root cause and presents viable solutions for team decision-making.

**Impact**: Users cannot use extension pool functionality in Plugin Framework, making the migration incomplete and blocking production usage for organizations that rely on extension pools.

## Problem Overview

### Root Cause
The issue stems from fundamental differences in how SDKv2 and Plugin Framework handle Set element identity:

- **SDKv2**: Uses custom hash functions that can exclude specific fields from Set identity calculations
- **Plugin Framework**: Uses complete object values for Set identity, with no custom hash support

In SDKv2, `extension_pool_id` was deliberately excluded from the Set hash to prevent plan diffs when pool assignments changed. Plugin Framework cannot replicate this behavior, causing Set element correlation failures.

## Why It Works in SDKv2

### Example Configuration
```hcl
resource "genesyscloud_user" "example" {
  name  = "John Doe"
  email = "john@example.com"
  
  addresses {
    phone_numbers {
      extension         = "8501"
      media_type        = "PHONE"
      type              = "WORK"
      extension_pool_id = genesyscloud_telephony_providers_edges_extension_pool.pool.id
    }
  }
}

resource "genesyscloud_telephony_providers_edges_extension_pool" "pool" {
  start_number = "8500"
  end_number   = "8699"
}
```

### SDKv2 State File
```json
{
  "addresses": [
    {
      "phone_numbers": [
        {
          "extension": "8501",
          "extension_pool_id": "7d0135e7-82fa-4f46-9fba-686289b4e0f7",
          "media_type": "PHONE",
          "number": null,
          "type": "WORK"
        }
      ]
    }
  ]
}
```

### Why SDKv2 Works
1. **Custom Hash Function**: Excludes `extension_pool_id` from Set identity
2. **State Storage**: Pool ID is stored and visible in state
3. **API Integration**: Pool ID is sent to API and properly assigned
4. **Stable Identity**: Set elements remain stable even when pool assignments change

## Why It Fails in Plugin Framework

### Example Configuration (Same as SDKv2)
```hcl
resource "genesyscloud_user" "example" {
  name  = "John Doe"
  email = "john@example.com"
  
  addresses {
    phone_numbers {
      extension         = "8501"
      media_type        = "PHONE"
      type              = "WORK"
      extension_pool_id = genesyscloud_telephony_providers_edges_extension_pool.pool.id
    }
  }
}
```

### Plugin Framework State File
```json
{
  "addresses": [
    {
      "phone_numbers": [
        {
          "extension": "8501",
          "extension_pool_id": null,
          "media_type": "PHONE",
          "number": null,
          "type": "WORK"
        }
      ]
    }
  ]
}
```

### Why Plugin Framework Fails
1. **No Custom Hash**: Set identity includes all fields, including `extension_pool_id`
2. **Null in State**: Current implementation sets `extension_pool_id` to null in state
3. **Missing API Integration**: `extension_pool_id` from config is ignored and never sent to API
4. **Set Correlation Failure**: Planned element (with pool ID) doesn't match actual element (null pool ID)

## Test Case Comparison

### Test Case: Create User with Extension Pool

**Configuration Used in Both SDKv2 and PF:**
```hcl
phone_numbers {
  extension         = "8501"
  media_type        = "PHONE"
  type              = "WORK"
  extension_pool_id = "${genesyscloud_telephony_providers_edges_extension_pool.suresh_EPID2.id}"
}
```

**SDKv2 Result:**
- ✅ **Success**: User created with extension assigned to specified pool
- ✅ **State**: Shows `extension_pool_id` with actual pool UUID
- ✅ **Remote API**: Extension properly associated with pool

**Plugin Framework Result:**
- ❌ **Failure**: "Provider produced inconsistent result after apply"
- ❌ **State**: Shows `extension_pool_id: null`
- ❌ **Remote API**: Extension created but not associated with any pool

**Error Message:**
```
planned set element cty.ObjectVal(map[string]cty.Value{
  "extension":cty.StringVal("8501"),
  "extension_pool_id":cty.StringVal("0d22c483-399f-4f2a-b638-c8e48e735d29"), 
  "media_type":cty.StringVal("PHONE"),
  "number":cty.NullVal(cty.String), 
  "type":cty.StringVal("WORK")
}) does not correlate with any element in actual.
```

## Possible Solutions

### Solution 1: Top-Level Computed Map

**Schema Changes:**
- Remove `extension_pool_id` from `phone_numbers` Set
- Add top-level computed `phone_extension_pools` map attribute

**Configuration:**
```hcl
# User config (extension_pool_id removed from phone_numbers)
addresses {
  phone_numbers {
    extension  = "8501"
    media_type = "PHONE"
    type       = "WORK"
    # extension_pool_id removed - no longer user-configurable
  }
}

# Pool assignment handled via separate resource or data source
resource "genesyscloud_user_extension_pool_assignment" "example" {
  user_id           = genesyscloud_user.example.id
  extension         = "8501"
  extension_pool_id = genesyscloud_telephony_providers_edges_extension_pool.pool.id
}
```

**State File:**
```json
{
  "phone_extension_pools": {
    "PHONE|WORK|8501": "7d0135e7-82fa-4f46-9fba-686289b4e0f7"
  },
  "addresses": [
    {
      "phone_numbers": [
        {
          "extension": "8501",
          "media_type": "PHONE",
          "number": null,
          "type": "WORK"
        }
      ]
    }
  ]
}
```

**Breaking Change:** Yes - requires config restructuring
**Complexity:** High - requires new resource type and complex state management

**Pros:**
- ✅ Completely eliminates Set identity issues
- ✅ Matches Plugin Framework best practices
- ✅ Clean separation of concerns (phone numbers vs pool assignments)
- ✅ No perpetual diffs when pool assignments change
- ✅ Extensible for future pool-related features

**Cons:**
- ❌ Major breaking change requiring user config migration
- ❌ High implementation complexity (new resource type needed)
- ❌ Different user experience from SDKv2
- ❌ Requires additional Terraform resources for pool assignments
- ❌ Potential confusion during migration period

### Solution 2: Accept Set Identity Changes

**Schema Changes:**
- Keep `extension_pool_id` in `phone_numbers` Set
- Implement proper processing in build/flatten functions

**Configuration:**
```hcl
# Same as current - no changes required
addresses {
  phone_numbers {
    extension         = "8501"
    media_type        = "PHONE"
    type              = "WORK"
    extension_pool_id = genesyscloud_telephony_providers_edges_extension_pool.pool.id
  }
}
```

**State File:**
```json
{
  "addresses": [
    {
      "phone_numbers": [
        {
          "extension": "8501",
          "extension_pool_id": "7d0135e7-82fa-4f46-9fba-686289b4e0f7",
          "media_type": "PHONE",
          "number": null,
          "type": "WORK"
        }
      ]
    }
  ]
}
```

**Breaking Change:** No - maintains current configuration pattern
**Complexity:** Low - fix existing implementation, accept more diffs when pools change

**Pros:**
- ✅ No breaking changes for users
- ✅ Maintains SDKv2 configuration compatibility
- ✅ Lowest implementation risk and complexity
- ✅ Preserves state visibility of pool assignments
- ✅ Quick fix to restore functionality
- ✅ Users keep explicit control over pool assignments

**Cons:**
- ❌ More diffs when extension pool assignments change
- ❌ Set element replacements when pools change (not just updates)
- ❌ Potential for plan instability in complex scenarios
- ❌ Doesn't follow Plugin Framework best practices for Set identity
- ❌ May cause confusion when pool changes trigger phone number "replacements"

### Solution 3: Dependency-Only Approach

**Schema Changes:**
- Remove `extension_pool_id` field entirely from schema
- Rely on API auto-assignment based on extension number ranges

**Configuration:**
```hcl
# Pool must be created first
resource "genesyscloud_telephony_providers_edges_extension_pool" "pool" {
  start_number = "8500"
  end_number   = "8699"
}

# User config (no extension_pool_id)
resource "genesyscloud_user" "example" {
  addresses {
    phone_numbers {
      extension  = "8501"  # Must be within pool range
      media_type = "PHONE"
      type       = "WORK"
      # No extension_pool_id - auto-assigned by API
    }
  }
  
  depends_on = [genesyscloud_telephony_providers_edges_extension_pool.pool]
}
```

**State File:**
```json
{
  "addresses": [
    {
      "phone_numbers": [
        {
          "extension": "8501",
          "media_type": "PHONE",
          "number": null,
          "type": "WORK"
        }
      ]
    }
  ]
}
```

**Breaking Change:** Yes - requires removing field and restructuring dependencies
**Complexity:** Low - simplest implementation, but loses explicit pool control

**Pros:**
- ✅ Eliminates Set identity issues completely
- ✅ Simplest implementation (just remove the field)
- ✅ Leverages existing API auto-assignment behavior
- ✅ Follows Terraform dependency patterns
- ✅ No complex state management needed
- ✅ Clean Plugin Framework implementation

**Cons:**
- ❌ Major breaking change requiring config restructuring
- ❌ Loss of explicit pool control for users
- ❌ Potential ambiguity with overlapping pool ranges
- ❌ Dependency management complexity for users
- ❌ Different behavior from SDKv2 expectations
- ❌ No visibility of pool assignments in state

### Solution 4: Hybrid Computed Approach

**Schema Changes:**
- Keep `extension_pool_id` as Optional for user input
- Add computed logic to normalize/stabilize values
- Use plan modifiers for intelligent handling

**Configuration:**
```hcl
# Same as current - no changes required
addresses {
  phone_numbers {
    extension         = "8501"
    media_type        = "PHONE"
    type              = "WORK"
    extension_pool_id = genesyscloud_telephony_providers_edges_extension_pool.pool.id
  }
}
```

**State File:**
```json
{
  "addresses": [
    {
      "phone_numbers": [
        {
          "extension": "8501",
          "extension_pool_id": "7d0135e7-82fa-4f46-9fba-686289b4e0f7",
          "media_type": "PHONE",
          "number": null,
          "type": "WORK"
        }
      ]
    }
  ]
}
```

**Breaking Change:** No - maintains current configuration pattern
**Complexity:** Medium - requires sophisticated plan modifiers and state correlation logic

**Pros:**
- ✅ No breaking changes for users
- ✅ Maintains SDKv2 configuration compatibility
- ✅ Preserves state visibility of pool assignments
- ✅ Intelligent handling of Set identity issues
- ✅ Users keep explicit control over pool assignments
- ✅ Potentially fewer diffs than Solution 2

**Cons:**
- ❌ Medium implementation complexity
- ❌ Requires sophisticated plan modifier logic
- ❌ Risk of edge cases in state correlation
- ❌ More complex debugging and maintenance
- ❌ Potential for unexpected behavior with complex configurations
- ❌ May still have some diff issues in certain scenarios

## Recommendation

Based on the analysis, **Solution 2 (Accept Set Identity Changes)** is recommended because:

### Why Solution 2 is Preferred:

**Immediate Value:**
- Restores broken functionality quickly with minimal risk
- No user impact or migration required
- Maintains expected user experience from SDKv2

**Risk Assessment:**
- Lowest implementation risk among all solutions
- Well-understood trade-offs (more diffs vs broken functionality)
- Easy to implement and test

**Strategic Considerations:**
- Preserves investment in existing user configurations
- Allows time for future architectural improvements
- Provides working solution while evaluating long-term options

### Trade-off Analysis:

| Solution | Breaking Change | Implementation Risk | User Impact | Functionality |
|----------|----------------|-------------------|-------------|---------------|
| Solution 1 | ❌ Yes | 🔴 High | 🔴 High | ✅ Full |
| **Solution 2** | ✅ **No** | 🟢 **Low** | 🟢 **None** | ✅ **Full** |
| Solution 3 | ❌ Yes | 🟡 Medium | 🔴 High | 🟡 Reduced |
| Solution 4 | ✅ No | 🟡 Medium | 🟢 None | ✅ Full |

**Conclusion:** The increased diffs when extension pool assignments change are an acceptable trade-off for maintaining functionality and avoiding breaking changes during the SDKv2 to Plugin Framework migration. This approach allows the team to deliver a working solution immediately while keeping options open for future architectural improvements.