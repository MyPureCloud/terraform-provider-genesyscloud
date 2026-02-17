**Contributor & Maintainer Design Notes (SDKv2 → PF)**

> **Purpose of this file**
> This document captures **technical gaps, behavioral differences, and design constraints** observed while migrating this resource from **Terraform SDK v2** to the **Terraform Plugin Framework (PF)**.
>
> **Audience**
>
> * Terraform provider contributors
> * Maintainers reviewing or extending this resource
>
> ⚠️ **Important**
>
> * Do **not** add raw Go code unless strictly necessary
> * Prefer **design explanations over implementation details**
> * Focus on **why** behaviors differ, not just *what* differs
> * Be precise, factual, and forward-looking

---

## 1. Migration Scope & Current State

**Migration Type:** Full Plugin Framework migration

**Current Status:**
* All CRUD operations migrated to Plugin Framework
* Schema fully converted to PF schema definitions
* State management using PF types and models
* Custom diff logic from SDKv2 not yet replicated in PF

**Known Issues & Status:**
* **DEVTOOLING-1238:** Address deletion functionality fails when EMAIL contacts exist (⏭️ IN PROGRESS)
  - Test case `TestAccFrameworkResourceUserAddresses` fails at steps 9/10 (address removal scenarios)
  - Details: See Sections 2-11
* **Export Flat Format Issue** (✅ RESOLVED - January 22, 2026)
  - Problem: PF resources exported in flat dot notation instead of nested HCL
  - Solution: Built CTY type converter, fixed 3 bugs (ElementType, NestedObject, CTY bypass)
  - Result: Export now produces nested HCL format identical to SDKv2
  - Details: See Section 12

**Future Enhancements:**
* **Phase 2 Native PF Export** (⏭️ PLANNED - 9-13 weeks)
  - Goal: Eliminate temporary export code (~500 lines), achieve true Framework integration
  - Approach: Use Framework Read method directly, no CTY dependency
  - Benefits: Better maintainability, no flatten functions, proper type handling
  - Details: See Section 13

---

## 2. How This Resource Behaved in SDKv2

**Custom Diff Logic:**
* SDKv2 used `customizeDiffAddressRemoval` function to detect when addresses block was omitted
* When addresses block removed, function forced explicit change to empty array `[]` in the diff
* This triggered API call with empty addresses array during update

**Lenient State Handling:**
* After API update, SDKv2 accepted whatever state the API returned
* No strict validation that desired state matched actual state
* `flattenUserAddresses()` populated state with actual API response, including any remaining contacts

**Why It "Worked":**
* Test cases passed because no consistency validation occurred
* `terraform apply` succeeded even when EMAIL contacts remained after deletion attempt
* State file correctly reflected API reality, but users expected different behavior
* Silent acceptance of partial deletion created expectation mismatch

---

## 3. Plugin Framework Behavioral Model

**Strict State Consistency:**
* Plugin Framework validates that planned state matches actual state after apply
* If configuration expects no addresses block, state must have no addresses
* Mismatch between expected and actual state causes "inconsistent result" error

**Declarative Lifecycle:**
* No custom diff functions like SDKv2's `CustomizeDiff`
* State changes must be handled through standard CRUD operations
* Plan modifiers available but more limited than SDKv2 diff customization

**Change Detection:**
* Framework automatically detects configuration changes
* Stricter validation of block count changes (0 to 1, 1 to 0)
* Less tolerance for "close enough" state matching

---

## 4. Key Behavior Differences (SDKv2 vs PF)

| Area | SDKv2 | Plugin Framework |
|------|-------|------------------|
| **Diff Customization** | `CustomizeDiff` function forces empty array when addresses block omitted | No equivalent mechanism - relies on standard plan/apply cycle |
| **State Validation** | Accepts any state returned by API after update | Validates planned state matches actual state - fails on mismatch |
| **Address Deletion** | Sends empty array, accepts partial deletion (EMAIL contacts remain) | Sends empty array, detects partial deletion, fails with consistency error |
| **Error Reporting** | Silent acceptance of API limitations | Explicit error when state doesn't match expectations |
| **Test Behavior** | Tests pass despite EMAIL contacts remaining in state | Tests fail when EMAIL contacts remain (correct behavior) |

---

## 5. Root Causes of Current Gaps

**API Asymmetric Deletion Behavior:**
* Genesys Cloud API `/api/v2/users/{userId}` PATCH endpoint with empty addresses array `[]` deletes PHONE/SMS contacts but NOT EMAIL contacts
* This is API-level behavior, not provider issue
* Confirmed through manual testing and log analysis

**Framework Design Constraint:**
* Plugin Framework enforces strict state consistency by design
* No equivalent to SDKv2's `CustomizeDiff` for post-apply state manipulation
* Cannot "accept" partial deletion as valid state when configuration expects complete deletion

**Backward Compatibility Risk:**
* SDKv2 behavior (silent acceptance) is technically incorrect but customers rely on it
* Fixing to match PF behavior would be breaking change for SDKv2 users
* PF correctly exposes the API limitation that SDKv2 was hiding

**Design Trade-off:**
* **Option A:** Maintain backward compatibility - hide API limitation, allow inconsistent state
* **Option B:** Provide correct functionality - expose API limitation, fail on inconsistency
* **Option C:** Implement workaround - explicit EMAIL deletion logic to achieve true deletion

---

## 6. Known Technical Gaps

**Address Deletion Logic:**
* No PF equivalent to SDKv2's `customizeDiffAddressRemoval` function
* Cannot force empty array in plan when addresses block omitted
* Standard update flow sends empty array but cannot handle partial API deletion

**State Consistency Handling:**
* `flattenUserAddresses()` correctly returns API state (including remaining EMAIL contacts)
* PF detects mismatch: config expects 0 addresses, API returns 1 address block
* No mechanism to "accept" this mismatch as valid like SDKv2 did

**EMAIL Contact Deletion:**
* No explicit deletion logic for EMAIL media type contacts
* Relying solely on empty addresses array doesn't delete EMAIL contacts
* Would require separate API calls to delete EMAIL contacts individually

**Test Coverage:**
* `TestAccFrameworkResourceUserAddresses` step 9/10 fails on address removal
* Test expects all addresses deleted when block omitted
* Current implementation cannot achieve this due to API limitation

---

## 7. Why These Gaps Are Hard to Fix

**Rejected Approach: Disable State Validation**
* Could suppress PF consistency error to mimic SDKv2 behavior
* **Risk:** Violates PF design principles, creates technical debt
* **Risk:** State would show EMAIL contacts when config expects none
* **Risk:** Users wouldn't know deletion failed partially

**Rejected Approach: Modify flattenUserAddresses to Hide EMAIL Contacts**
* Could filter out EMAIL contacts from state when addresses block omitted
* **Risk:** State wouldn't match API reality (data hiding)
* **Risk:** Drift between Terraform state and actual Genesys Cloud state
* **Risk:** Subsequent applies could have unexpected behavior

**Complexity: Explicit EMAIL Deletion**
* Requires detecting when addresses block removed AND EMAIL contacts exist
* Must make additional API calls to delete EMAIL contacts individually
* **Risk:** Multiple API calls increase failure surface area
* **Risk:** Partial failure scenarios (PHONE deleted, EMAIL deletion fails)
* **Complexity:** Need to handle rollback if EMAIL deletion fails

**Backward Compatibility Concern:**
* Any fix that changes behavior from SDKv2 could break existing customer workflows
* Customers may have workarounds built around current SDKv2 behavior
* Need to balance correctness vs. compatibility

---

## 8. Considered & Possible Future Approaches

**Approach 1: Implement Explicit EMAIL Deletion (Recommended)**

*Implementation:*
* Detect when `plan.Addresses.IsNull()` and `currentState.Addresses` has EMAIL contacts
* After sending empty addresses array, make additional API calls to delete remaining EMAIL contacts
* Validate all addresses deleted before completing update

*Trade-offs:*
* ✅ Provides true DEVTOOLING-1238 functionality
* ✅ Maintains backward compatibility (no workflow changes)
* ✅ Better than SDKv2 (actually deletes all addresses)
* ⚠️ Requires 2-3 days development effort
* ⚠️ Additional API calls increase complexity

**Approach 2: API Team Fix (Preferred Long-term)**

*Request:*
* Ask API team to fix asymmetric deletion behavior in `/api/v2/users/{userId}` endpoint
* Empty addresses array should delete ALL media types consistently

*Trade-offs:*
* ✅ Fixes root cause for all API consumers
* ✅ No provider code changes needed
* ✅ Eliminates technical debt
* ⚠️ Depends on API team prioritization
* ⚠️ Unknown timeline

**Approach 3: Replicate SDKv2 Behavior (Backward Compatible)**

*Implementation:*
* Suppress PF consistency validation for addresses block
* Accept partial deletion as valid state (EMAIL contacts remain)

*Trade-offs:*
* ✅ Perfect backward compatibility
* ✅ Quick implementation (1 day)
* ❌ Perpetuates incorrect behavior
* ❌ Violates PF design principles
* ❌ Creates technical debt

**Approach 4: Enhanced Error with Workaround Guidance**

*Implementation:*
* Detect EMAIL contacts when addresses block removed
* Provide clear error message with workaround instructions

*Trade-offs:*
* ✅ Quick implementation (1 day)
* ✅ Honest about API limitation
* ❌ Breaking change from SDKv2
* ❌ Requires customer workflow changes

---

## 9. Risks & Side Effects

**If Implementing Explicit EMAIL Deletion:**
* **Risk:** Additional API calls could fail independently, leaving resource in partial state
* **Risk:** Need proper error handling and potentially rollback logic
* **Maintenance:** More complex update logic to maintain
* **Testing:** Need comprehensive test coverage for failure scenarios

**If Replicating SDKv2 Behavior:**
* **Risk:** State drift - Terraform state won't match API reality
* **Risk:** Violates PF design principles, may cause issues with future PF versions
* **Risk:** Technical debt that will need to be addressed eventually
* **Maintenance:** Suppressing framework validation is fragile

**If Changing Behavior from SDKv2:**
* **Risk:** Breaking change for existing customers
* **Risk:** Customer confusion during migration
* **Risk:** Support burden from customers reporting "new bugs"
* **Migration:** Need clear communication and migration guide

**General Risks:**
* **API Changes:** If API team fixes asymmetric deletion, provider code may need updates
* **Framework Evolution:** PF validation may become stricter in future versions
* **Customer Expectations:** Any solution must manage customer expectations clearly

---

## 10. Contribution Guidelines for This Resource

**Critical Areas - Require Maintainer Review:**
* Any changes to address deletion logic
* Modifications to `flattenUserAddresses()` function
* Changes to state validation or consistency checking
* Updates to `updateUser()` function related to addresses

**Mandatory Testing:**
* `TestAccFrameworkResourceUserAddresses` must pass all steps
* Test address deletion with both PHONE and EMAIL contacts
* Test address deletion with only EMAIL contacts
* Test address deletion with only PHONE contacts
* Verify state consistency after all operations

**What Must Not Be Changed Casually:**
* State flattening logic - must accurately reflect API reality
* Schema definitions for addresses block
* Update logic flow without considering API asymmetric behavior

**Before Making Changes:**
* Review DEVTOOLING-1238 analysis document
* Understand API asymmetric deletion behavior
* Consider backward compatibility impact
* Test with actual Genesys Cloud API, not just mocks

---

## 12. Export Flat Format Issue (RESOLVED)

### 12.1 Problem Summary

**Issue:** PF resources exported in flat dot notation instead of nested HCL format

**Example of Problem:**
```hcl
# Wrong (flat format)
addresses.0.phone_numbers.0.type = "WORK"
addresses.0.phone_numbers.0.number = "+13175559002"
addresses.0.phone_numbers.1.type = "WORK2"
```

**Expected Behavior:**
```hcl
# Correct (nested format)
addresses {
  phone_numbers {
    type = "WORK"
    number = "+13175559002"
  }
  phone_numbers {
    type = "WORK2"
  }
}
```

**Impact:**
* Export output was unusable
* Missing nested block structure
* All 35 attributes fetched but only 1 appeared in output
* Dependencies not exported

**Status:** ✅ RESOLVED (January 22, 2026)

---

### 12.2 Root Cause

**Primary Cause:** Framework resources lacked CTY type converter

**Three Specific Issues:**

1. **Missing CTY Type Converter**
   - Framework resources don't have `CoreConfigSchema()` or `ImpliedType()` methods like SDKv2
   - Exporter couldn't build proper CTY type for Framework schemas
   - Fell back to basic stub type `{id, name}` which filtered out all other attributes

2. **ElementType Field Access Bug**
   - Collection attributes (List, Set, Map) have `ElementType` field, not `GetType()` method
   - Code tried to call non-existent method
   - Converter failed on collection attributes

3. **NestedObject Field Access Bug**
   - Block types (ListNestedBlock, SetNestedBlock, SingleNestedBlock) have `NestedObject` field, not `GetNestedObject()` method
   - Code tried to call non-existent method
   - Converter failed on nested blocks

4. **CTY Conversion Bypass**
   - Code intentionally bypassed CTY conversion for Framework resources
   - Written when only basic CTY type existed
   - Produced flat output even when CTY type was available

---

### 12.3 Solution Implemented

**Approach:** Built Framework Schema → CTY Type converter

**Functions Added (6 total):**
1. `getFrameworkResourceFactories()` - Get Framework resource factories
2. `getFrameworkSchema()` - Retrieve schema from Framework resource
3. `frameworkSchemaToCtyType()` - Convert schema to CTY type
4. `frameworkAttributeToCtyType()` - Convert attributes (12 types supported)
5. `frameworkBlockToCtyType()` - Convert blocks (3 types supported)
6. `frameworkTypeToCtType()` - Convert element types (8 types supported)

**Bugs Fixed:**
1. **ElementType Bug** - Used reflection to access `ElementType` field instead of calling `GetType()` method
2. **NestedObject Bug** - Used reflection to access `NestedObject` field instead of calling `GetNestedObject()` method
3. **CTY Bypass Bug** - Removed bypass code, now uses CTY conversion for Framework resources

**Code Location:**
* File: `genesyscloud/tfexporter/genesyscloud_resource_exporter.go`
* Lines added: ~365 net
* Approach: Reflection-based type detection to handle Framework API inconsistencies

**Implementation Characteristics:**
* ✅ Generic - works for ALL Framework resources, not just user
* ✅ Handles all 23 Framework types (12 attributes + 3 blocks + 8 elements)
* ✅ No resource-specific code in exporter
* ✅ Backward compatible - SDKv2 resources unaffected

---

### 12.4 Current Behavior After Fix

**What Works:**
* ✅ Nested HCL format (not flat dot notation)
* ✅ All 35 user attributes exported
* ✅ All nested blocks formatted correctly (addresses, phone_numbers, employer_info, etc.)
* ✅ Dependencies exported (division_id, extension_pool_id, etc.)
* ✅ Boolean values as boolean (not string "false")
* ✅ Export time same as SDKv2 (~22 seconds)
* ✅ Functionally equivalent to SDKv2 output

**Known Cosmetic Difference:**
* ⚠️ Attribute ordering differs from SDKv2
* Cause: Go map iteration is non-deterministic
* Example: SDKv2 may have `email` near top, PF may have it near bottom
* Impact: None - Terraform doesn't care about attribute order
* Both outputs produce identical state when applied

**Lazy Fetch Implementation:**
* User resource uses lazy fetch callback pattern (same concept as SDKv2 RefreshWithoutUpgrade)
* Callback captures userId in closure, fetches full details only for filtered users
* Builds flat attribute map (35 attributes) from SDK types
* Functions: `GetAllUsersSDK()`, `buildUserAttributes()`, 10 flatten functions
* Location: `genesyscloud/user/resource_genesyscloud_user.go` and `resource_genesyscloud_user_export_utils.go`

---

### 12.5 Technical Implementation Details

**Type Coverage:**

*Attributes (12 types):*
* StringAttribute, BoolAttribute, Int64Attribute, Float64Attribute, NumberAttribute
* ListAttribute, SetAttribute, MapAttribute
* SingleNestedAttribute, ListNestedAttribute, SetNestedAttribute, MapNestedAttribute

*Blocks (3 types):*
* ListNestedBlock, SetNestedBlock, SingleNestedBlock

*Element Types (8 types):*
* StringType, BoolType, Int64Type, Float64Type, NumberType
* ListType, SetType, MapType, ObjectType

**Reflection Usage:**
* Framework API is inconsistent - some types use methods, others use fields
* Nested attributes have `GetNestedObject()` method ✅
* Collection attributes have `ElementType` field (not method) ❌
* Block types have `NestedObject` field (not method) ❌
* Solution: Use reflection to access fields dynamically

**Example Reflection Code:**
```go
// Access ElementType field
elemTypeValue := reflect.ValueOf(attr).FieldByName("ElementType")
elemType := elemTypeValue.Interface()

// Access NestedObject field
nestedObjValue := reflect.ValueOf(block).FieldByName("NestedObject")
nestedObj, ok := nestedObjValue.Interface().(fwschema.NestedBlockObject)
```

---

### 12.6 Testing & Verification

**Test Results:**
* ✅ Export produces nested HCL format
* ✅ All attributes present in output
* ✅ Dependencies exported correctly
* ✅ Output can be applied with Terraform
* ✅ No errors in export logs

**Test Files:**
* Output: `genesyscloud/user/PF-terraform-genesyscloud_user/genesyscloud.tf`
* Reference: `genesyscloud/user/sdkv2-terraform-genesyscloud_user/genesyscloud.tf`
* Comparison shows functional equivalence (only attribute order differs)

---

### 12.7 Lessons Learned

**What Went Wrong:**
* Assumed Framework API consistency (methods vs fields)
* Didn't test incrementally after each type category
* Forgot own bypass code from earlier implementation

**What Went Right:**
* Generic design - works for all Framework resources
* Systematic debugging - each bug revealed a pattern
* Complete audit - covered all Framework types
* Reflection approach - handles dynamic type access

---

## 13. Phase 2 Native PF Export (FUTURE ENHANCEMENT)

### 13.1 Enhancement Summary

**Goal:** Eliminate temporary export code, achieve true Framework integration

**Current Approach (Phase 1 - Implemented):**
```
PF Resource → Lazy Callback → Flat Map → CTY Type → Nested JSON → HCL
```

**Proposed Approach (Phase 2 - Future):**
```
PF Resource → Framework Read → Framework Types → Framework Schema → Nested JSON → HCL
```

**Key Difference:** Phase 2 uses Framework types and methods natively throughout, no CTY conversion needed

**Status:** ⏭️ PLANNED (9-13 weeks implementation)

---

### 13.2 Why Consider This Enhancement

**Current Phase 1 Has Temporary Code:**
* ~500 lines of temporary code (flatten functions, lazy callbacks)
* Duplicates logic that Framework Read already provides
* Requires maintenance for each resource
* Uses flat maps instead of Framework types

**Phase 2 Would Eliminate:**
* ✅ All flatten functions (10 functions for user resource)
* ✅ All lazy fetch callbacks
* ✅ CTY type converter dependency
* ✅ Resource-specific export code

**Phase 2 Would Provide:**
* ✅ True Framework integration (uses Framework Read method)
* ✅ Better type safety (boolean as boolean, not string)
* ✅ Easier maintenance (generic implementation)
* ✅ Cleaner codebase (no temporary code)

---

### 13.3 Proposed Technical Approach

**Architecture:**

1. **Use Framework Read Method Directly**
   - Call resource's Read method instead of lazy callback
   - Get Framework types (types.String, types.Bool, etc.) directly
   - No need to duplicate API call logic

2. **Extract Data Using Framework Schema**
   - Use Framework schema to navigate nested structure
   - Convert Framework types to native Go types
   - Proper type handling (boolean, number, string)

3. **No CTY Conversion Needed**
   - Work with nested data directly
   - No CTY type building required
   - Simpler architecture

4. **Generic Implementation**
   - Works for all Framework resources
   - No resource-specific code
   - Schema changes automatically reflected

**Example Flow:**
```go
// Create Framework resource instance
resource := &UserFrameworkResource{}

// Call Framework Read method
resource.Read(ctx, readRequest, &readResponse)

// Extract data using Framework schema
data := extractFrameworkData(readResponse.State, resource.Schema())

// Generate HCL directly from nested data
hcl := generateHCL(resourceType, blockLabel, data, schema)
```

---

### 13.4 Trade-offs Analysis

**Benefits:**
* ✅ Eliminates ~500 lines of temporary code per resource
* ✅ True Framework integration (uses Framework methods)
* ✅ Better type safety (proper boolean, number handling)
* ✅ Easier maintenance (generic, no resource-specific code)
* ✅ No CTY dependency
* ✅ Schema changes automatically reflected
* ✅ Cleaner codebase

**Costs:**
* ❌ More complex initial implementation (new export handler architecture)
* ❌ Longer development time (9-13 weeks vs 1-2 weeks for Phase 1)
* ❌ Need to maintain two export paths (SDKv2 + PF) during transition
* ❌ More testing required
* ❌ Higher initial risk

**Performance:**
* Same as Phase 1 (~22 seconds)
* Both make identical API calls
* No performance difference

---

### 13.5 Implementation Timeline

**Phase 1: Proof of Concept (2-3 weeks)**
* Design Framework export handler interface
* Implement Framework resource reader
* Implement schema-based data extraction
* Test with user resource
* Compare output with SDKv2

**Phase 2: Core Implementation (3-4 weeks)**
* Build production-ready Framework export handler
* Handle all Framework attribute and block types
* Implement dependency and reference resolution
* Add error handling and logging

**Phase 3: Integration (2-3 weeks)**
* Integrate with existing exporter
* Route SDKv2 vs PF resources
* Test coexistence
* Performance optimization

**Phase 4: Rollout (2-3 weeks)**
* Test with all Framework resources
* Beta testing
* Documentation
* Production deployment

**Total: 9-13 weeks**

---

### 13.6 Decision & Recommendation

**Decision: Hybrid Approach**

1. **Phase 1 (Implemented):** CTY converter for immediate fix
   - Timeline: 1-2 weeks ✅ COMPLETE
   - Risk: Low
   - Benefit: Immediate unblocking

2. **Phase 2 (Future):** Native PF export for long-term solution
   - Timeline: 9-13 weeks ⏭️ PLANNED
   - Risk: Medium
   - Benefit: Long-term maintainability

**Rationale:**
* Phase 1 unblocks current exports immediately
* Provides time to build proper Phase 2 solution
* Lower risk (gradual transition)
* Can validate both approaches
* Flexibility to adjust timeline

**When to Start Phase 2:**
* After all Framework resources migrated (user, language, wrapupcode)
* When team has bandwidth for larger project
* When temporary code maintenance becomes burden
* When other Framework features need similar patterns

---

## 14. Related References & Tracking

**Related Issues:**
* DEVTOOLING-1238: Address deletion functionality (⏭️ IN PROGRESS)
* Export flat format issue (✅ RESOLVED - January 22, 2026)
* Phase 2 Native PF Export (⏭️ PLANNED - 9-13 weeks)

**Related Documents:**
* None - This document is self-contained

**API Documentation:**
* Genesys Cloud API: `PATCH /api/v2/users/{userId}`
* Addresses field behavior and media type handling

**Terraform Documentation:**
* [Plugin Framework State Management](https://developer.hashicorp.com/terraform/plugin/framework/handling-data/state)
* [Plugin Framework Plan Modification](https://developer.hashicorp.com/terraform/plugin/framework/resources/plan-modification)

**Test Files:**
* `TestAccFrameworkResourceUserAddresses` - Address CRUD operations test
* Steps 9/10 currently failing on address deletion scenarios

---
