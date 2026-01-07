# Investigation: Extension Pool ID Set Identity Mismatch - Part 2

## Detailed Analysis of Non-Breaking vs Breaking Change Solutions

This document provides a comprehensive analysis of the two viable approaches to resolve the extension pool ID Set identity mismatch issue identified in Part 1, with detailed implementation considerations, user experience impact, and technical trade-offs.

## Solution Overview

Based on the investigation in Part 1, two primary solutions have been identified:

- **Option 1**: Computed Field Approach (Non-Breaking Change)
- **Option 2**: Dependency-Only Approach (Breaking Change)

This analysis examines both options in detail to determine the optimal path forward.

## Option 1: Computed Field Approach (Non-Breaking Change)

### Implementation Strategy

The computed field approach maintains backward compatibility by implementing a dual-field system that separates user configuration from Set identity management.

#### Schema Architecture

The computed field approach implements a **dual-field system** with fields at different schema levels:

```go
func UserResourceSchema() schema.Schema {
    return schema.Schema{
        Attributes: map[string]schema.Attribute{
            "id": schema.StringAttribute{...},
            "email": schema.StringAttribute{...},
            "name": schema.StringAttribute{...},
            
            // ✅ TOP-LEVEL: Computed field for internal pool tracking
            "phone_extension_pools": schema.MapAttribute{
                ElementType: types.StringType,
                Computed:    true,
                Description: "Computed mapping of phone identity keys to extension_pool_id. " +
                           "Used internally to prevent diffs when pool assignments change.",
                // Map structure: "MEDIA|TYPE|E164|EXT" -> "extension_pool_id"
            },
        },
        Blocks: map[string]schema.Block{
            "addresses": schema.ListNestedBlock{
                NestedObject: schema.NestedBlockObject{
                    Blocks: map[string]schema.Block{
                        "phone_numbers": schema.SetNestedBlock{
                            NestedObject: schema.NestedBlockObject{
                                Attributes: map[string]schema.Attribute{
                                    "number": schema.StringAttribute{...},
                                    "media_type": schema.StringAttribute{...},
                                    "type": schema.StringAttribute{...},
                                    "extension": schema.StringAttribute{...},
                                    
                                    // ✅ NESTED LEVEL: User-facing field (for configuration and dependencies)
                                    "extension_pool_id": schema.StringAttribute{
                                        Description: "Id of the extension pool which contains this extension.",
                                        Optional:    true,
                                        // Used ONLY for:
                                        // 1. User configuration syntax
                                        // 2. Terraform dependency tracking  
                                        // 3. NOT stored in phone_numbers state (always null to avoid Set identity issues)
                                    },
                                },
                            },
                        },
                    },
                },
            },
        },
    }
}
```

#### Field Positioning Strategy

The dual-field approach requires careful positioning of fields at different schema levels to achieve both backward compatibility and Set identity stability.

**Why `extension_pool_id` stays INSIDE `phone_numbers` (Nested Level):**
- ✅ **User Configuration Syntax**: Users configure it per phone number, exactly matching SDKv2 location
- ✅ **Dependency Tracking**: Terraform dependency graph requires it at phone number level to create proper resource ordering
- ✅ **Backward Compatibility**: Maintains identical configuration syntax as SDKv2 - no breaking changes
- ✅ **Logical Grouping**: Extension pool assignment is conceptually tied to specific phone number
- ✅ **Import/Export Consistency**: Matches where users expect to configure the relationship

**Why `phone_extension_pools` moves to TOP LEVEL (Root Level):**
- ✅ **Set Identity Isolation**: Positioned outside of `phone_numbers` Set, so changes don't affect Set element identity
- ✅ **Global Pool Tracking**: Can track all phone-to-pool mappings across all addresses in one centralized location
- ✅ **Computed Field Convention**: Top-level computed fields follow Terraform provider patterns and are easier to document
- ✅ **Avoids Nesting Complexity**: Simpler state management without deep nested computed field synchronization
- ✅ **Plan Output Clarity**: Changes appear at root level in `terraform plan`, making them more visible

**Dual-Field Interaction Pattern:**
```hcl
# User configures at nested level (for syntax compatibility and dependencies)
addresses {
  phone_numbers {
    extension = "4105"
    extension_pool_id = genesyscloud_telephony_providers_edges_extension_pool.pool1.id
  }
}

# System tracks at top level (for Set identity stability)
phone_extension_pools = {
  "PHONE|WORK||4105" = "pool1-id-abc123"
}
```

This positioning strategy ensures that:
1. **User Experience**: Configuration syntax remains unchanged from SDKv2
2. **Dependency Management**: Terraform can track resource dependencies correctly
3. **Set Identity**: Phone number Set elements remain stable when pool assignments change
4. **State Consistency**: Both representations stay synchronized through custom logic

#### Set Identity Logic

```go
// Phone number Set identity excludes extension_pool_id
type PhoneNumberIdentity struct {
    Number    string  // ✓ Included in Set identity
    MediaType string  // ✓ Included in Set identity
    Type      string  // ✓ Included in Set identity
    Extension string  // ✓ Included in Set identity
    // extension_pool_id: EXCLUDED from Set identity (not stored in phone_numbers state)
}

// Identity key generation for computed map
func generatePhoneIdentityKey(phone PhoneNumber) string {
    return fmt.Sprintf("%s|%s|%s|%s", 
        phone.MediaType, phone.Type, phone.Number, phone.Extension)
}
```

#### State Management

The dual-field system requires careful state synchronization:

```go
// During Read operation - populate both field systems
func flattenUserAddresses(addresses *[]platformclientv2.Contact) {
    // 1. Build phone_numbers Set (excludes extension_pool_id from state)
    phoneNumbers := []map[string]attr.Value{}
    phoneExtensionPools := map[string]attr.Value{}
    
    for _, address := range addresses {
        // Populate phone_numbers Set (extension_pool_id always null)
        phoneNumber := map[string]attr.Value{
            "number":            types.StringValue(number),
            "media_type":        types.StringValue(mediaType),
            "type":              types.StringValue(phoneType),
            "extension":         types.StringValue(extension),
            "extension_pool_id": types.StringNull(), // ❗ Always null in Set state
        }
        phoneNumbers = append(phoneNumbers, phoneNumber)
        
        // Populate top-level computed map (actual pool tracking)
        if extension != "" {
            poolId := fetchExtensionPoolId(ctx, extension, proxy)
            identityKey := fmt.Sprintf("%s|%s|%s|%s", 
                mediaType, phoneType, number, extension)
            phoneExtensionPools[identityKey] = types.StringValue(poolId)
        }
    }
    
    // Set both representations in state
    state.Addresses = buildAddressesWithPhoneNumbers(phoneNumbers)
    state.PhoneExtensionPools = types.MapValueMust(types.StringType, phoneExtensionPools)
}
```

#### How Set Identity Works

```go
// Plugin Framework Set identity calculation (cannot be changed)
type PhoneNumberSetElement struct {
    Number          string  // ✓ Included in Set identity
    MediaType       string  // ✓ Included in Set identity
    Type            string  // ✓ Included in Set identity
    Extension       string  // ✓ Included in Set identity
    ExtensionPoolId string  // ❗ Always null in state → doesn't affect identity
}

// Separate tracking outside Set identity
type PhoneExtensionPoolsMap map[string]string {
    "PHONE|WORK||4105": "pool1-id-abc123",  // Identity key → Pool ID
    "PHONE|WORK||4225": "pool2-id-def456",  // Identity key → Pool ID
}
```

### Configuration Examples

#### User Configuration (Unchanged from SDKv2)

```hcl
# Extension Pool Resources
resource "genesyscloud_telephony_providers_edges_extension_pool" "pool1" {
  start_number = "4100"
  end_number   = "4199"
  description  = "Primary extension pool"
}

resource "genesyscloud_telephony_providers_edges_extension_pool" "pool2" {
  start_number = "4200"
  end_number   = "4299"
  description  = "Secondary extension pool"
}

# User Resource - IDENTICAL to SDKv2 syntax
resource "genesyscloud_user" "example" {
  email = "user@example.com"
  name  = "Test User"
  
  addresses {
    phone_numbers {
      extension         = "4105"
      media_type        = "PHONE"
      type              = "WORK"
      extension_pool_id = genesyscloud_telephony_providers_edges_extension_pool.pool1.id
    }
  }
}
```

#### Terraform State Representation

**What users see in `terraform show` output:**

```hcl
# terraform show output
resource "genesyscloud_user" "example" {
  id    = "user-123"
  email = "user@example.com"
  name  = "Test User"
  
  addresses {
    phone_numbers {
      extension    = "4105"
      media_type   = "PHONE"
      type         = "WORK"
      # ❗ extension_pool_id = null (always null in state despite user configuration)
    }
  }
  
  # ✅ Top-level computed field shows actual pool assignments
  phone_extension_pools = {
    "PHONE|WORK||4105" = "4d55ecc4-5012-4f04-9612-d7ad0dd4f490"
  }
}
```

#### User Experience Impact

**Configuration vs State Mismatch Explained:**

The dual-field approach creates an intentional mismatch between what users configure and what appears in Terraform state. This is the core trade-off of maintaining backward compatibility while solving Set identity issues.

```hcl
# What user configures (unchanged from SDKv2):
addresses {
  phone_numbers {
    extension = "4105"
    extension_pool_id = genesyscloud_telephony_providers_edges_extension_pool.pool1.id
  }
}

# What appears in Terraform state:
addresses {
  phone_numbers {
    extension = "4105"
    extension_pool_id = null  # ❗ Always null despite user configuration
  }
}
phone_extension_pools = {
  "PHONE|WORK||4105" = "pool1-id-abc123"  # ✅ Actual value stored here
}
```

**Why This Mismatch Occurs:**
1. **Set Identity Requirement**: Plugin Framework uses full object value for Set identity
2. **Stability Need**: Including `extension_pool_id` in state would make Set elements unstable when pools change
3. **Backward Compatibility**: User configuration field must remain for dependency tracking
4. **Solution**: Store actual value in computed field outside the Set, keep configured field null in state

**Impact on User Workflows:**

**Terraform Plan Output:**
```bash
# When changing extension pool, plan shows:
Terraform will perform the following actions:

  # genesyscloud_user.example will be updated in-place
  ~ resource "genesyscloud_user" "example" {
      id = "user-123"
      
      # ❗ No change shown in phone_numbers (despite user configuring extension_pool_id)
      
      # ✅ Change shown in computed field
    ~ phone_extension_pools = {
        ~ "PHONE|WORK||4105" = "pool1-id-abc" -> "pool2-id-xyz"
      }
    }
```

**User Confusion Points:**
1. **State Inspection**: `terraform show` displays `extension_pool_id = null` even though user configured a value
2. **Plan Understanding**: Changes appear in `phone_extension_pools` field, not in the field user actually modified
3. **Debugging Complexity**: Users must understand that actual pool assignment is tracked in computed field
4. **Documentation Burden**: Requires extensive explanation of why state differs from configuration

#### Extension Pool Update Scenario

```hcl
# Step 1: User updates configuration to use different pool
resource "genesyscloud_user" "example" {
  addresses {
    phone_numbers {
      extension         = "4105"
      extension_pool_id = genesyscloud_telephony_providers_edges_extension_pool.pool2.id  # Changed
    }
  }
}

# Step 2: Terraform plan shows change in computed field only
# terraform plan output:
Terraform will perform the following actions:

  # genesyscloud_user.example will be updated in-place
  ~ resource "genesyscloud_user" "example" {
      id = "user-123"
      
      # No change in phone_numbers Set (identity preserved)
      
      # Change shown in computed field
    ~ phone_extension_pools = {
        ~ "PHONE|WORK||4105" = "pool1-id-abc" -> "pool2-id-def"
      }
    }

# Step 3: After apply, state reflects new pool assignment
phone_extension_pools = {
  "PHONE|WORK||4105" = "pool2-id-def"
}
```

### Implementation Requirements

#### Code Changes Required

1. **Schema Definition**
   - Uncomment `phone_extension_pools` field
   - Modify `extension_pool_id` handling to exclude from state

2. **State Management Functions**
   - `flattenUserAddresses()`: Implement dual-field population
   - `buildSdkPhoneNumbers()`: Handle user configuration field
   - Identity key generation logic

3. **CRUD Operations**
   - Create: Populate computed field after user creation
   - Read: Synchronize both fields from API response
   - Update: Handle changes in both representations
   - Delete: Clean up both field types

4. **Plan/Apply Logic**
   - Custom plan handling for dual-field system
   - Dependency resolution using user configuration field
   - State synchronization between fields

5. **Import/Export**
   - Import: Generate computed field from API data
   - Export: Reference user configuration field for dependencies

#### Testing Requirements

1. **Unit Tests**
   - Identity key generation
   - Dual-field state management
   - Plan/apply logic with both fields

2. **Integration Tests**
   - Extension pool creation/update/deletion scenarios
   - Multiple extension pools per user
   - Address removal scenarios
   - Import/export functionality

3. **Regression Tests**
   - All existing SDKv2 test scenarios
   - Backward compatibility validation
   - State migration testing

### Advantages

1. **✅ Zero Breaking Changes**
   - **Identical Configuration Syntax**: Users keep exact same configuration syntax as SDKv2
   - **No Migration Required**: Existing Terraform configurations work without any changes
   - **Preserved Dependencies**: All `extension_pool_id` references continue to create proper Terraform dependencies
   - **Full SDKv2 Compatibility**: All extension pool functionality from SDKv2 is preserved

2. **✅ Solves Set Identity Issues**
   - **Stable Set Elements**: Pool changes don't affect phone number Set identity calculation
   - **No Perpetual Diffs**: Extension pool reassignments don't trigger unwanted plan changes
   - **Set Element Preservation**: Phone numbers maintain consistent identity across plan/apply cycles
   - **Framework Compliance**: Works within Plugin Framework Set identity constraints

3. **✅ Maintains Terraform Dependencies**
   - **Resource Ordering**: `extension_pool_id` field still creates proper Terraform dependencies
   - **Dependency Graph**: Resource creation/deletion order remains correct
   - **Implicit Dependencies**: Terraform automatically detects resource relationships
   - **Explicit Dependencies**: Users can still use `depends_on` when needed

4. **✅ Complete Functionality Coverage**
   - **All Extension Pool Scenarios**: Supports all SDKv2 extension pool use cases
   - **Multiple Pools Per User**: Handles complex scenarios with multiple extension pools
   - **Pool Update Scenarios**: Supports changing pools, adding/removing extensions
   - **Edge Cases**: Handles extension-only phone numbers and complex address updates

### Disadvantages

1. **❌ High Implementation Complexity**
   - **Dual-Field System**: Requires maintaining two separate field representations with complex synchronization logic
   - **State Management Complexity**: Must synchronize user configuration field with computed tracking field on every operation
   - **Custom Synchronization Logic**: Requires complex logic to keep both representations consistent
   - **Higher Bug Risk**: Multiple state representations increase probability of synchronization bugs and edge cases

2. **❌ Confusing User Experience**
   - **State Mismatch**: User configures `extension_pool_id = pool.id` but Terraform state shows `extension_pool_id = null`
   - **Plan Output Confusion**: Changes appear in `phone_extension_pools` computed field, not in the field user actually configured
   - **Debugging Complexity**: Users must understand dual-field system to troubleshoot issues effectively
   - **Non-Intuitive Behavior**: Configuration doesn't match state representation, violating user expectations

3. **❌ Documentation and Support Burden**
   - **Complex Documentation**: Requires extensive explanation of why state differs from configuration
   - **User Training**: Users need to understand dual-field concept to use effectively
   - **Support Complexity**: Support teams must explain non-intuitive behavior to confused users
   - **FAQ Maintenance**: Ongoing documentation updates to address user confusion

4. **❌ Import/Export Complications**
   - **Import Mismatch**: `terraform import` shows computed field, but user must write configuration using nested field
   - **Export Complexity**: Must handle dual representations in export logic, choosing which field to reference
   - **State Migration**: Complex migration logic needed between field representations
   - **Tool Integration**: Third-party tools may not understand dual-field pattern

5. **❌ Long-term Maintenance Burden**
   - **Perpetual Complexity**: Two field systems must be maintained indefinitely
   - **Complex Edge Cases**: Dual-field synchronization creates numerous edge cases to handle
   - **Future Enhancement Risk**: New features must consider both field representations, increasing development complexity
   - **Technical Debt**: Complex workaround creates ongoing maintenance overhead

6. **❌ Performance and Resource Overhead**
   - **Dual-Field Synchronization**: Additional processing on every read/write operation
   - **Memory Usage**: Duplicate field tracking increases memory consumption
   - **Plan/Apply Complexity**: More complex state calculations slow down plan/apply operations
   - **API Call Overhead**: May require additional API calls to synchronize field representations

### Implementation Complexity: HIGH

**Estimated Development Effort**: 3-4 weeks
**Risk Level**: High (complex state management, multiple edge cases)
**Maintenance Impact**: Significant ongoing complexity

## Option 2: Dependency-Only Approach (Breaking Change)

### Implementation Strategy

The dependency-only approach removes the problematic `extension_pool_id` field entirely and relies on explicit `depends_on` relationships for proper resource ordering.

#### Schema Architecture

The dependency-only approach **removes the problematic field entirely** and simplifies the schema:

```go
"phone_numbers": schema.SetNestedBlock{
    Description: "Phone number addresses for this user.",
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
            },
            "type": schema.StringAttribute{
                Description: "Type of number (WORK | WORK2 | WORK3 | WORK4 | HOME | MOBILE | OTHER).",
                Optional:    true,
                Computed:    true,
                Default:     stringdefault.StaticString("WORK"),
            },
            "extension": schema.StringAttribute{
                Description: "Phone number extension",
                Optional:    true,
            },
            // ❌ extension_pool_id field REMOVED entirely - problem eliminated
        },
    },
},
```

#### Simplified Set Identity Logic

```go
// Clean Set identity without extension_pool_id complications
type PhoneNumberIdentity struct {
    Number    string  // ✓ Included in Set identity
    MediaType string  // ✓ Included in Set identity
    Type      string  // ✓ Included in Set identity
    Extension string  // ✓ Included in Set identity
    // No extension_pool_id field - Set identity issues completely eliminated
}
```

#### Simplified State Management

```go
// Single-field system - no synchronization complexity
func flattenUserAddresses(addresses *[]platformclientv2.Contact) {
    phoneNumbers := []map[string]attr.Value{}
    
    for _, address := range addresses {
        phoneNumber := map[string]attr.Value{
            "number":     types.StringValue(number),
            "media_type": types.StringValue(mediaType),
            "type":       types.StringValue(phoneType),
            "extension":  types.StringValue(extension),
            // No extension_pool_id field to manage or synchronize
        }
        phoneNumbers = append(phoneNumbers, phoneNumber)
    }
    
    // No computed field synchronization needed
    // No dual-field complexity
    // Single source of truth
}
```

### Configuration Examples

#### Migration from SDKv2 to Plugin Framework

**Before (SDKv2):**
```hcl
resource "genesyscloud_user" "example" {
  email = "user@example.com"
  name  = "Test User"
  
  addresses {
    phone_numbers {
      extension         = "4105"
      extension_pool_id = genesyscloud_telephony_providers_edges_extension_pool.pool1.id
    }
  }
}
```

**After (Plugin Framework):**
```hcl
resource "genesyscloud_user" "example" {
  email = "user@example.com"
  name  = "Test User"
  
  addresses {
    phone_numbers {
      extension = "4105"
      # extension_pool_id field removed
    }
  }
  
  # Explicit dependency ensures proper resource ordering
  depends_on = [genesyscloud_telephony_providers_edges_extension_pool.pool1]
}
```

#### Multiple Extension Pools Scenario

```hcl
resource "genesyscloud_telephony_providers_edges_extension_pool" "pool1" {
  start_number = "4100"
  end_number   = "4199"
}

resource "genesyscloud_telephony_providers_edges_extension_pool" "pool2" {
  start_number = "4200"
  end_number   = "4299"
}

resource "genesyscloud_user" "multi_ext" {
  email = "multi@example.com"
  name  = "Multi Extension User"
  
  addresses {
    phone_numbers {
      extension = "4105"  # Auto-assigned to pool1 by API
    }
    phone_numbers {
      extension = "4225"  # Auto-assigned to pool2 by API
    }
  }
  
  # Dependencies ensure both pools exist before user creation
  depends_on = [
    genesyscloud_telephony_providers_edges_extension_pool.pool1,
    genesyscloud_telephony_providers_edges_extension_pool.pool2
  ]
}
```

#### Extension Pool Update Scenario

```hcl
# Step 1: User with extension from pool1 range
resource "genesyscloud_user" "example" {
  addresses {
    phone_numbers {
      extension = "4105"  # In pool1 range (4100-4199)
    }
  }
  depends_on = [genesyscloud_telephony_providers_edges_extension_pool.pool1]
}

# Step 2: Update to extension from pool2 range
resource "genesyscloud_user" "example" {
  addresses {
    phone_numbers {
      extension = "4225"  # In pool2 range (4200-4299) - API auto-assigns
    }
  }
  depends_on = [
    genesyscloud_telephony_providers_edges_extension_pool.pool1,
    genesyscloud_telephony_providers_edges_extension_pool.pool2
  ]
}
```

#### Address Removal Scenario

```hcl
# Step 3: Remove addresses entirely
resource "genesyscloud_user" "example" {
  email = "user@example.com"
  name  = "Test User"
  # No addresses block
  
  # Keep dependencies to ensure proper cleanup order
  depends_on = [
    genesyscloud_telephony_providers_edges_extension_pool.pool1,
    genesyscloud_telephony_providers_edges_extension_pool.pool2
  ]
}
```

### Implementation Requirements

#### Code Changes Required

1. **Schema Definition**
   - Remove `extension_pool_id` field from phone_numbers schema
   - Update field descriptions and documentation

2. **State Management Functions**
   - Simplify `flattenUserAddresses()` - remove extension pool handling
   - Simplify `buildSdkPhoneNumbers()` - no pool ID processing
   - Remove extension pool synchronization logic

3. **CRUD Operations**
   - Rely on API auto-assignment for extension-to-pool mapping
   - Remove extension pool ID handling from all operations
   - Simplify error handling (no pool-related errors)

4. **Testing**
   - Update all tests to use `depends_on` pattern
   - Remove extension pool ID assertions
   - Add dependency ordering tests

5. **Documentation**
   - Migration guide for users
   - Updated examples and tutorials
   - Deprecation notices for SDKv2 patterns

#### Migration Support

1. **Documentation**
   - Clear migration guide with before/after examples
   - Automated migration script (optional)
   - FAQ for common migration scenarios

2. **Validation**
   - Schema validation to detect old `extension_pool_id` usage
   - Clear error messages with migration guidance
   - Terraform plan warnings for deprecated patterns

### Advantages

1. **✅ Simple Implementation**
   - **Remove Problematic Field**: Eliminates complexity at the source by removing the problematic field entirely
   - **Single Field System**: No synchronization complexity between multiple field representations
   - **Straightforward CRUD Operations**: No dual-field logic required in create/read/update/delete operations
   - **Lower Development Risk**: Removing complexity is inherently safer than adding complex workarounds

2. **✅ Clear User Experience**
   - **Configuration Matches State**: What users configure is exactly what appears in Terraform state
   - **Intuitive Plan Output**: Changes appear where users expect them, making plan output predictable
   - **No Confusing Dual-Field System**: Users see exactly what they configure, eliminating confusion
   - **Easier Debugging**: Single source of truth for phone number data simplifies troubleshooting

3. **✅ Aligns with API Behavior**
   - **Leverages API Auto-Assignment**: Uses Genesys Cloud's automatic extension-to-pool assignment as designed
   - **Natural API Integration**: Works with API's built-in logic rather than fighting against it
   - **Reduces Configuration Complexity**: Fewer fields for users to manage and understand
   - **API-First Approach**: Trusts API to handle extension pool mapping based on extension ranges

4. **✅ Plugin Framework Best Practices**
   - **Standard Dependency Management**: Uses standard `depends_on` for resource ordering, following Terraform conventions
   - **No Custom Set Identity Workarounds**: Works with Plugin Framework as designed, not against it
   - **Follows Terraform Conventions**: Explicit dependencies are standard practice in Terraform providers
   - **Framework Compliance**: No framework limitations or constraints to work around

5. **✅ Long-term Maintainability**
   - **Simpler Codebase**: Fewer edge cases to handle, less code to maintain and debug
   - **Reduced Future Complexity**: Clean foundation makes future enhancements easier to implement
   - **Lower Bug Risk**: Fewer moving parts means fewer potential failure points
   - **Easier Developer Onboarding**: New team members can understand the system quickly

6. **✅ Performance Benefits**
   - **No Dual-Field Synchronization**: Single field operations are faster and more efficient
   - **Simpler State Management**: Faster read/write operations without complex synchronization
   - **Optimized Plan/Apply**: Less complex state calculations improve plan/apply performance
   - **Reduced Memory Usage**: No duplicate field tracking reduces memory consumption

### Disadvantages

1. **❌ Breaking Change Required**
   - **Configuration Migration**: Users must update their Terraform configurations to remove `extension_pool_id` references
   - **Deployment Disruption**: Potential disruption during upgrade process as configurations need to be updated
   - **Migration Effort**: Existing deployments require careful migration planning and execution
   - **Rollback Complexity**: Rolling back to previous provider version requires configuration changes

2. **❌ Less Explicit Control**
   - **API Auto-Assignment Dependency**: Relies on Genesys Cloud API's automatic extension-to-pool assignment logic
   - **Limited Pool Selection Control**: May not work for complex pool assignment scenarios requiring explicit control
   - **Reduced Granular Control**: Less fine-grained control over which specific pool an extension uses
   - **API Behavior Dependency**: Success depends on API's extension range matching logic working correctly

3. **❌ Migration Planning Complexity**
   - **Dependency Identification**: Users need to identify all extension pool dependencies in their configurations
   - **Complex Configuration Updates**: Large configurations may require careful migration planning and testing
   - **Risk of Missing Dependencies**: Potential for missing dependencies during migration, causing resource ordering issues
   - **Testing Requirements**: Extensive testing needed to ensure migration doesn't break existing deployments

4. **❌ Potential Functionality Gaps**
   - **Edge Case Coverage**: May not cover all complex extension pool scenarios that worked in SDKv2
   - **Advanced Use Cases**: Some advanced pool assignment patterns may not be supported
   - **API Limitation Exposure**: Exposes any limitations in API's auto-assignment logic
   - **Reduced Flexibility**: Less flexibility for users who need explicit pool control

### Implementation Complexity: LOW

**Estimated Development Effort**: 1-2 weeks
**Risk Level**: Low (removing complexity, not adding it)
**Maintenance Impact**: Reduced ongoing complexity

## Comparative Analysis

### Technical Comparison

| Aspect | Option 1 (Computed Field) | Option 2 (Dependency-Only) |
|--------|---------------------------|----------------------------|
| **Set Identity Issues** | ✅ Solved (dual-field system) | ✅ Solved (field removed) |
| **Backward Compatibility** | ✅ Full compatibility | ❌ Breaking change required |
| **Implementation Complexity** | ❌ High (dual-field system) | ✅ Low (remove field) |
| **Code Maintainability** | ❌ Complex (synchronization logic) | ✅ Simple (single field system) |
| **User Experience** | ❌ Confusing (dual representation) | ✅ Clear (matches configuration) |
| **API Alignment** | ⚠️ Workaround approach | ✅ Natural alignment |
| **Performance** | ❌ Overhead (dual tracking) | ✅ Optimal (simplified) |
| **Testing Complexity** | ❌ High (multiple scenarios) | ✅ Low (straightforward) |
| **Documentation Burden** | ❌ High (explain dual system) | ✅ Low (standard patterns) |
| **Future Enhancement Risk** | ❌ High (complex foundation) | ✅ Low (simple foundation) |

### User Impact Comparison

| User Scenario | Option 1 Impact | Option 2 Impact |
|---------------|------------------|------------------|
| **Existing Configurations** | ✅ No changes required | ❌ Migration required |
| **New Configurations** | ⚠️ Confusing state representation | ✅ Clear, intuitive syntax |
| **Debugging Issues** | ❌ Complex (dual fields) | ✅ Simple (single representation) |
| **Import/Export** | ❌ Non-intuitive mapping | ✅ Straightforward |
| **Plan Understanding** | ❌ Changes in computed field | ✅ Changes in configured field |
| **Learning Curve** | ❌ High (explain dual system) | ✅ Low (standard Terraform patterns) |

### Risk Assessment

#### Option 1 Risks

1. **High Implementation Risk**
   - Complex dual-field synchronization logic
   - Multiple edge cases to handle
   - Higher probability of bugs

2. **User Confusion Risk**
   - State doesn't match configuration
   - Non-intuitive plan output
   - Support burden increase

3. **Maintenance Risk**
   - Complex codebase to maintain
   - Future enhancements complicated
   - Technical debt accumulation

#### Option 2 Risks

1. **Migration Risk**
   - Users must update configurations
   - Potential for missed dependencies
   - Upgrade disruption

2. **Functionality Risk**
   - Relies on API auto-assignment
   - May not cover all use cases
   - Less explicit control

## Recommendation

### Primary Recommendation: Option 2 (Dependency-Only Approach)

Based on the comprehensive analysis, **Option 2 (Dependency-Only Approach)** is strongly recommended despite being a breaking change.

#### Justification

1. **Long-term Benefits Outweigh Short-term Pain**
   - **Simple, Maintainable Implementation**: Clean codebase that's easy to understand, debug, and enhance
   - **Clear User Experience**: Configuration matches state, eliminating confusion and support burden
   - **Terraform Best Practices**: Aligns with standard Terraform patterns and Plugin Framework design principles
   - **Future-Proof Foundation**: Provides solid foundation for future enhancements without technical debt

2. **Breaking Change is Manageable**
   - **Clear Migration Path**: Simple, well-defined steps to update configurations
   - **Straightforward Configuration Changes**: Users only need to remove field and add `depends_on`
   - **Migration Tooling Potential**: Can provide automated migration scripts and validation tools
   - **One-Time Impact**: Breaking change pain is temporary, but benefits are permanent

3. **Technical Superiority**
   - **Eliminates Root Cause**: Removes the problematic field entirely rather than working around it
   - **Better Foundation**: Provides clean foundation for future development and maintenance
   - **Reduced Risk**: Fewer edge cases and synchronization bugs compared to dual-field approach
   - **Performance Benefits**: Simpler operations with better performance characteristics

4. **User Experience Priority**
   - **Intuitive Behavior**: Configuration matches state representation, meeting user expectations
   - **Predictable Plan Output**: Changes appear where users expect them to appear
   - **Easier Debugging**: Single source of truth simplifies troubleshooting and understanding
   - **Reduced Learning Curve**: Standard Terraform patterns are easier for users to understand

#### Implementation Strategy

1. **Phase 1: Preparation (2-3 weeks)**
   - **Migration Documentation**: Develop comprehensive migration guide with before/after examples
   - **Migration Validation Tools**: Create tools to help users identify configurations that need updates
   - **Communication Plan**: Communicate breaking change timeline and migration resources to community
   - **Testing Framework**: Develop extensive test suite for migration scenarios

2. **Phase 2: Implementation (1-2 weeks)**
   - **Schema Updates**: Remove `extension_pool_id` field from phone_numbers schema
   - **Code Simplification**: Remove all extension pool handling logic from user resource
   - **Test Updates**: Update all tests to use `depends_on` pattern instead of `extension_pool_id`
   - **Error Messages**: Implement clear error messages for old syntax with migration guidance

3. **Phase 3: Migration Support (4-6 weeks)**
   - **Migration Assistance**: Provide community support for migration questions and issues
   - **Issue Monitoring**: Monitor for migration-related issues and provide quick resolution
   - **Documentation Updates**: Update all community resources, examples, and tutorials
   - **Feedback Collection**: Gather feedback on migration experience for future improvements

### Alternative Recommendation: Hybrid Approach

If breaking changes are absolutely unacceptable in the current release cycle, consider a **phased approach**:

1. **Phase 1: Interim Solution (Current Release)**
   - Implement Option 1 (Computed Field) as temporary backward-compatible solution
   - Provide full functionality while maintaining SDKv2 compatibility
   - Document the dual-field system thoroughly for user understanding

2. **Phase 2: Deprecation Period (Next Release)**
   - Add deprecation warnings for `extension_pool_id` field usage
   - Provide migration documentation and tooling
   - Encourage users to migrate to `depends_on` pattern

3. **Phase 3: Breaking Change (Major Version)**
   - Migrate to Option 2 (Dependency-Only) in next major version
   - Remove deprecated `extension_pool_id` field and dual-field complexity
   - Achieve long-term clean implementation

**Hybrid Approach Benefits:**
- **Immediate Solution**: Solves current test failures without breaking changes
- **Migration Runway**: Gives users time to plan and execute migration
- **Community Preparation**: Allows community to adapt gradually
- **Long-term Clean Solution**: Eventually achieves optimal implementation

**Hybrid Approach Drawbacks:**
- **Extended Complexity**: Maintains complex dual-field system longer
- **Multiple Migration Phases**: Users experience multiple changes over time
- **Increased Development Effort**: Requires implementing both solutions
- **Technical Debt Period**: Carries complex workaround for extended time

## Conclusion

While **Option 1 (Computed Field)** achieves the technical goal of maintaining backward compatibility, it introduces significant complexity and user experience issues that may cause more long-term problems than the breaking change it aims to avoid.

**Option 2 (Dependency-Only Approach)** provides a cleaner, more maintainable solution that aligns with Terraform best practices and Plugin Framework design principles. It provides a better foundation for future development and eliminates the root cause of the problem rather than working around it.

**Key Decision Factors:**

1. **Technical Debt vs. Breaking Change Trade-off**
   - Option 1 creates permanent technical debt to avoid temporary migration pain
   - Option 2 accepts short-term migration effort for long-term technical benefits

2. **User Experience Considerations**
   - Option 1 creates confusing dual-field system that violates user expectations
   - Option 2 provides intuitive, predictable behavior that matches Terraform conventions

3. **Maintenance and Development Impact**
   - Option 1 increases ongoing maintenance burden and development complexity
   - Option 2 simplifies codebase and reduces future development risk

4. **Community and Ecosystem Alignment**
   - Option 1 creates provider-specific patterns that differ from Terraform standards
   - Option 2 follows established Terraform and Plugin Framework best practices

**Final Recommendation:**

**Accept the breaking change** and invest in **excellent migration support** rather than implementing a complex workaround that creates long-term technical debt and user confusion. The short-term migration effort is justified by the significant long-term benefits in maintainability, user experience, and technical foundation.

**Success Metrics:**
- Clean, maintainable codebase that follows Plugin Framework best practices
- Intuitive user experience with configuration matching state representation
- Comprehensive migration documentation and tooling
- Strong community support during migration period
- Solid foundation for future feature development

## Success Criteria

### Option 1 Success Criteria
- ✅ All existing SDKv2 configurations work without changes
- ✅ No perpetual diffs in extension pool scenarios
- ✅ Proper resource ordering maintained
- ⚠️ User understanding of dual-field system
- ⚠️ Maintainable codebase complexity

### Option 2 Success Criteria
- ✅ Clean, intuitive configuration syntax
- ✅ No Set identity issues
- ✅ Simple, maintainable implementation
- ✅ Clear migration path provided
- ✅ Excellent migration documentation and tooling

The analysis strongly supports **Option 2** as the path forward for a robust, maintainable, and user-friendly solution.