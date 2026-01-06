# Configuration Comparison: SDKv2 vs Plugin Framework

## Extension Pool Configuration Changes

This document shows the exact configuration differences between SDKv2 and Plugin Framework implementations for extension pool scenarios.

## 1. Basic Extension Pool Configuration

### SDKv2 Configuration (Working)

```hcl
# Extension Pool Resource
resource "genesyscloud_telephony_providers_edges_extension_pool" "test-extension-pool-1" {
  start_number = "4100"
  end_number   = "4199"
  description  = "Test extension pool for user integration"
}

# User Resource with Extension Pool Reference
resource "genesyscloud_user" "test-user-extension-pool" {
  email = "terraform-ext-pool-user@example.com"
  name  = "Extension Pool User"
  
  addresses {
    phone_numbers {
      extension         = "4105"
      media_type        = "PHONE"
      type              = "WORK"
      extension_pool_id = genesyscloud_telephony_providers_edges_extension_pool.test-extension-pool-1.id
    }
  }
}
```

### Plugin Framework Configuration (Current - Broken)

```hcl
# Extension Pool Resource (Same)
resource "genesyscloud_telephony_providers_edges_extension_pool" "test-extension-pool-1" {
  start_number = "4100"
  end_number   = "4199"
  description  = "Test extension pool for user integration"
}

# User Resource - extension_pool_id commented out due to Set identity issues
resource "genesyscloud_user" "test-user-extension-pool" {
  email = "terraform-ext-pool-user@example.com"
  name  = "Extension Pool User"
  
  addresses {
    phone_numbers {
      extension  = "4105"
      media_type = "PHONE"
      type       = "WORK"
      # TODO: Temporarily commented - extension_pool_id causes Set identity mismatch in PF
      # extension_pool_id = genesyscloud_telephony_providers_edges_extension_pool.test-extension-pool-1.id
    }
  }
}
```

**Problem**: No dependency relationship exists, causing Terraform to try deleting extension pools while they're still referenced.

### Plugin Framework Configuration (Recommended Fix - Option 2)

```hcl
# Extension Pool Resource (Same)
resource "genesyscloud_telephony_providers_edges_extension_pool" "test-extension-pool-1" {
  start_number = "4100"
  end_number   = "4199"
  description  = "Test extension pool for user integration"
}

# User Resource - extension_pool_id field removed, depends_on added
resource "genesyscloud_user" "test-user-extension-pool" {
  email = "terraform-ext-pool-user@example.com"
  name  = "Extension Pool User"
  
  addresses {
    phone_numbers {
      extension  = "4105"
      media_type = "PHONE"
      type       = "WORK"
      # extension_pool_id field removed from schema
    }
  }
  
  # Explicit dependency ensures proper resource ordering
  depends_on = [genesyscloud_telephony_providers_edges_extension_pool.test-extension-pool-1]
}
```

**Solution**: Explicit `depends_on` ensures proper resource ordering without Set identity issues.

## 2. Extension Pool Update Scenario

### SDKv2 Configuration (Working)

```hcl
# Step 1: Create user with extension from pool 1
resource "genesyscloud_telephony_providers_edges_extension_pool" "pool-1" {
  start_number = "4100"
  end_number   = "4199"
}

resource "genesyscloud_user" "example" {
  email = "user@example.com"
  name  = "Test User"
  
  addresses {
    phone_numbers {
      extension         = "4105"
      extension_pool_id = genesyscloud_telephony_providers_edges_extension_pool.pool-1.id
    }
  }
}

# Step 2: Update to extension from pool 2
resource "genesyscloud_telephony_providers_edges_extension_pool" "pool-2" {
  start_number = "4200"
  end_number   = "4299"
}

resource "genesyscloud_user" "example" {
  email = "user@example.com"
  name  = "Test User"
  
  addresses {
    phone_numbers {
      extension         = "4225"  # Changed extension
      extension_pool_id = genesyscloud_telephony_providers_edges_extension_pool.pool-2.id  # Changed pool
    }
  }
}
```

**Result**: ✅ Works - Custom hash function ignores `extension_pool_id` changes, dependency tracking ensures proper update order.

### Plugin Framework Configuration (Current - Fails)

```hcl
# Same configuration but extension_pool_id commented out
resource "genesyscloud_user" "example" {
  addresses {
    phone_numbers {
      extension = "4225"
      # extension_pool_id = genesyscloud_telephony_providers_edges_extension_pool.pool-2.id
    }
  }
}
```

**Result**: ❌ Fails - No dependency relationship, Terraform tries to delete pool-1 while extension 4105 is still allocated.

### Plugin Framework Configuration (Recommended Fix)

```hcl
# Step 1: Create user with extension from pool 1
resource "genesyscloud_telephony_providers_edges_extension_pool" "pool-1" {
  start_number = "4100"
  end_number   = "4199"
}

resource "genesyscloud_user" "example" {
  email = "user@example.com"
  name  = "Test User"
  
  addresses {
    phone_numbers {
      extension = "4105"
    }
  }
  
  depends_on = [genesyscloud_telephony_providers_edges_extension_pool.pool-1]
}

# Step 2: Update to extension from pool 2
resource "genesyscloud_telephony_providers_edges_extension_pool" "pool-2" {
  start_number = "4200"
  end_number   = "4299"
}

resource "genesyscloud_user" "example" {
  email = "user@example.com"
  name  = "Test User"
  
  addresses {
    phone_numbers {
      extension = "4225"  # API auto-assigns to pool-2 based on range
    }
  }
  
  depends_on = [
    genesyscloud_telephony_providers_edges_extension_pool.pool-1,
    genesyscloud_telephony_providers_edges_extension_pool.pool-2
  ]
}
```

**Result**: ✅ Should work - Dependencies ensure both pools exist before user update, API handles pool assignment automatically.

## 3. Address Removal Scenario

### SDKv2 Configuration (Working)

```hcl
# Step 3: Remove addresses entirely
resource "genesyscloud_user" "example" {
  email = "user@example.com"
  name  = "Test User"
  # No addresses block
}

# Extension pools can remain - no dependency issues
resource "genesyscloud_telephony_providers_edges_extension_pool" "pool-1" {
  start_number = "4100"
  end_number   = "4199"
}

resource "genesyscloud_telephony_providers_edges_extension_pool" "pool-2" {
  start_number = "4200"
  end_number   = "4299"
}
```

**Result**: ✅ Works - User addresses removed first, then pools can be deleted safely.

### Plugin Framework Configuration (Current - Fails)

```hcl
# Same configuration - addresses removed
resource "genesyscloud_user" "example" {
  email = "user@example.com"
  name  = "Test User"
  # No addresses block
}
```

**Result**: ❌ Fails - No dependency tracking, Terraform tries to delete pools before user addresses are cleaned up.

### Plugin Framework Configuration (Recommended Fix)

```hcl
# Step 3: Remove addresses entirely
resource "genesyscloud_user" "example" {
  email = "user@example.com"
  name  = "Test User"
  # No addresses block
  
  # Keep dependencies to ensure proper cleanup order
  depends_on = [
    genesyscloud_telephony_providers_edges_extension_pool.pool-1,
    genesyscloud_telephony_providers_edges_extension_pool.pool-2
  ]
}
```

**Result**: ✅ Should work - Dependencies ensure user is updated (addresses removed) before pools are deleted.

## 4. Multiple Extension Pools Scenario

### SDKv2 Configuration (Working)

```hcl
resource "genesyscloud_user" "example" {
  addresses {
    phone_numbers {
      extension         = "4105"
      extension_pool_id = genesyscloud_telephony_providers_edges_extension_pool.pool-1.id
    }
    phone_numbers {
      extension         = "4225"
      extension_pool_id = genesyscloud_telephony_providers_edges_extension_pool.pool-2.id
    }
  }
}
```

### Plugin Framework Configuration (Recommended Fix)

```hcl
resource "genesyscloud_user" "example" {
  addresses {
    phone_numbers {
      extension = "4105"  # Auto-assigned to pool-1
    }
    phone_numbers {
      extension = "4225"  # Auto-assigned to pool-2
    }
  }
  
  depends_on = [
    genesyscloud_telephony_providers_edges_extension_pool.pool-1,
    genesyscloud_telephony_providers_edges_extension_pool.pool-2
  ]
}
```

## 5. Migration Path for Existing Users

### Current SDKv2 Configuration

```hcl
resource "genesyscloud_user" "existing_user" {
  email = "existing@example.com"
  name  = "Existing User"
  
  addresses {
    phone_numbers {
      extension         = "1001"
      extension_pool_id = genesyscloud_telephony_providers_edges_extension_pool.main_pool.id
    }
  }
}
```

### Migration to Plugin Framework

```hcl
resource "genesyscloud_user" "existing_user" {
  email = "existing@example.com"
  name  = "Existing User"
  
  addresses {
    phone_numbers {
      extension = "1001"
      # Remove: extension_pool_id = genesyscloud_telephony_providers_edges_extension_pool.main_pool.id
    }
  }
  
  # Add: Explicit dependency
  depends_on = [genesyscloud_telephony_providers_edges_extension_pool.main_pool]
}
```

## Summary of Configuration Changes

| Aspect | SDKv2 | Plugin Framework (Current) | Plugin Framework (Recommended) |
|--------|-------|---------------------------|-------------------------------|
| **Extension Pool Reference** | `extension_pool_id = pool.id` | `# extension_pool_id = pool.id` (commented) | Field removed from schema |
| **Dependency Management** | Implicit via `extension_pool_id` | None (broken) | Explicit via `depends_on` |
| **Pool Assignment** | Explicit reference | No reference | API auto-assignment |
| **Configuration Complexity** | Medium | Low (but broken) | Low |
| **Terraform Plan Behavior** | Stable (no perpetual diffs) | Broken (dependency errors) | Stable |

## Key Benefits of Recommended Approach

1. **Simpler Configuration**: No need to manage `extension_pool_id` references
2. **No Set Identity Issues**: Removing the problematic field eliminates Plugin Framework conflicts
3. **Explicit Dependencies**: Clear `depends_on` relationships are easier to understand
4. **API Alignment**: Leverages Genesys Cloud's automatic extension-to-pool assignment
5. **Backward Compatible**: Existing functionality preserved with simpler syntax

## Migration Effort

- **Low Impact**: Only need to remove `extension_pool_id` lines and add `depends_on`
- **No Functional Loss**: All extension pool scenarios continue to work
- **Clear Path**: Simple find-and-replace operation for most configurations
- **Validation**: Can be tested incrementally before full migration