# Extension Pool Migration Analysis: SDKv2 to Plugin Framework - RESOLVED ✅

**Status**: RESOLVED - January 18, 2026  
**Resolution**: Solution 2 implemented with manual testing validation  
**Time to Resolution**: 4 weeks investigation + implementation  
**Final Result**: All CRUD operations working correctly

---

## Executive Summary

The migration of the `genesyscloud_user` resource from Terraform SDKv2 to Plugin Framework encountered a critical issue with `extension_pool_id` functionality. Users experienced "Provider produced inconsistent result after apply" errors when using extension pools.

**Initial Hypothesis**: Set identity architectural issue due to Plugin Framework limitations  
**Actual Root Cause**: Three distinct bugs in the migrated code  
**Resolution Approach**: Solution 2 (Accept Set Identity Changes) with bug fixes  
**Validation**: Comprehensive manual testing of all CRUD operations

---

## Problem Statement

### Initial Symptoms
- ❌ "Provider produced inconsistent result after apply" errors
- ❌ `extension_pool_id` showing as `null` in state file
- ❌ Extensions created but not appearing in Extensions Assignments UI
- ❌ Extension pool functionality completely broken in Plugin Framework

### Impact
Users could not use extension pool functionality in Plugin Framework, blocking production usage for organizations that rely on extension pools for telephony management.

---

## Investigation Process

### Phase 1: Initial Analysis (Week 1)

**Hypothesis**: Set identity issue due to Plugin Framework's inability to exclude fields from Set hash

**Analysis**:
- SDKv2 uses custom hash functions to exclude `extension_pool_id` from Set identity
- Plugin Framework doesn't support custom Set hash functions
- Assumed this architectural difference was causing the errors

**Proposed Solutions**:
1. Solution 1: Move pool ID to top-level map (46-91 hours, breaking changes)
2. Solution 2: Accept Set identity changes (22 hours, no breaking changes)
3. Solution 3: Remove pool ID entirely (breaking changes)
4. Solution 4: Complex normalization logic (very complex)

**Decision**: Proceed with Solution 2 as lowest risk approach

### Phase 2: Deep Dive Investigation (Week 2-3)

**Approach**: Detailed comparison of SDKv2 vs Plugin Framework behavior

**Key Activities**:
1. Read SDKv2 source code in detail
2. Compare SDKv2 and PF state files
3. Analyze API call sequences in logs
4. Test edge cases (boundary conditions)

**Critical Discoveries**:

#### Discovery #1: State Storage Bug
**Location**: `resource_genesyscloud_user_utils.go` lines 768-775

**Finding**: `fetchExtensionPoolId()` call was commented out with misleading explanation

**Original Code**:
```go
// NOTE: We do not store the extension_pool_id in the state. The extension_pool_id
// is not part of the Set identity and is only used to validate the extension number.
// Storing it would cause unnecessary diffs.
// if extPoolId := fetchExtensionPoolId(ctx, proxy, ext); extPoolId != "" {
//     phone.ExtensionPoolId = types.StringValue(extPoolId)
// }
```

**Reality Check**: SDKv2 **DOES** store `extension_pool_id` in state file!

**Root Cause**: Developer confused "state storage" with "Set identity". SDKv2 stores pool ID in state but excludes it from Set hash calculation. These are two different concepts.

#### Discovery #2: Boundary Condition Bug
**Location**: `genesyscloud_user_proxy.go` line 343

**Finding**: Extension pool lookup used strict inequality instead of inclusive comparison

**Original Code**:
```go
if extNumInt > startNum && extNumInt < endNum {
```

**Problem**: Extensions at exact pool boundaries (e.g., 7700 in pool 7700-7799) were not found

**Impact**: User's test case with extension 7700 in pool 7700-7799 failed

#### Discovery #3: Missing Migration
**Location**: Multiple files

**Finding**: `waitForExtensionPoolActivation()` function was not migrated from SDKv2

**Impact**: Extensions assigned too quickly (1.7s) before telephony backend activated pool (~5s needed)

**Result**: Extensions appeared in user profile but NOT in Extensions Assignments UI

### Phase 3: Log Analysis (Week 3)

**Method**: Compared API call sequences between SDKv2 and Plugin Framework

**SDKv2 Behavior**:
```
1. Extension pool created at T
2. Wait 5 seconds for pool activation
3. User updated at T+5 seconds
4. Extension properly registered in telephony system
```

**PF Behavior (Before Fix)**:
```
1. Extension pool created at T
2. No wait logic
3. User updated at T+1.7 seconds
4. Extension NOT registered (too fast)
```

**Key Finding**: `waitForExtensionPoolActivation()` and its dependency `getTelephonyExtensionPoolById()` were never migrated to PF

### Phase 4: Implementation (Week 4)

**Bugs Identified**:
1. **Bug #1**: State storage - `fetchExtensionPoolId()` commented out
2. **Bug #2**: Boundary condition - strict inequality instead of inclusive
3. **Migration Gap**: `waitForExtensionPoolActivation()` not migrated

---

## Technical Resolution

### Fix #1: State Storage (Priority 1)

**File**: `genesyscloud/user/resource_genesyscloud_user_utils.go`

**Changes**:
```go
// Fetch and store extension_pool_id in state (matches SDKv2 behavior)
// NOTE: While extension_pool_id is not part of the Set identity (excluded from hash in SDKv2),
// it MUST be stored in state for Terraform to track the complete resource state.
// SDKv2 stores this value in state but excludes it from the Set hash using custom hash function.
// Plugin Framework includes all fields in Set identity (no custom hash support).
if extPoolId := fetchExtensionPoolId(ctx, proxy, ext); extPoolId != "" {
    phone.ExtensionPoolId = types.StringValue(extPoolId)
} else {
    phone.ExtensionPoolId = types.StringNull()
}
```

**Result**: ✅ `extension_pool_id` now correctly stored in state file

### Fix #2: Boundary Condition (Priority 2)

**File**: `genesyscloud/user/genesyscloud_user_proxy.go`

**Changes**:
```go
// FIX: Use inclusive comparison (>= and <=) to include extensions at boundaries
// Previous bug: Used strict inequality (> and <) which excluded start/end numbers
// Example: Pool 7700-7799 should include both 7700 and 7799
if extNumInt >= startNum && extNumInt <= endNum {
    return &pool, apiResponse, nil
}
```

**Result**: ✅ Extensions at pool boundaries now found correctly

### Fix #3: Migration Gap (Priority 1)

**Files Modified**:
1. `genesyscloud/user/genesyscloud_user_proxy.go` - Added `getTelephonyExtensionPoolById()` method
2. `genesyscloud/user/resource_genesyscloud_user_utils.go` - Added `waitForExtensionPoolActivation()` function

**Implementation**:
```go
// Migrated from SDKv2 - waits for newly created extension pools to activate
func waitForExtensionPoolActivation(ctx context.Context, plan *UserFrameworkResourceModel, proxy *userProxy) {
    // 1. Parse addresses and phone numbers
    // 2. Check if extension_pool_id is specified
    // 3. Fetch pool from API to check DateCreated
    // 4. If pool created < 5 seconds ago → Sleep 5 seconds
    // 5. Track waited pools to avoid duplicate waits
}

// Called at beginning of updateUser()
func updateUser(...) {
    waitForExtensionPoolActivation(ctx, plan, proxy)
    // ... rest of function
}
```

**Result**: ✅ Extensions now appear in Extensions Assignments UI

---

## Validation and Testing

### Manual Testing Performed

**Test Environment**: Real Genesys Cloud organization  
**Test Scope**: All CRUD operations with extension pools  
**Test Duration**: Comprehensive testing over multiple days

#### Test Case 1: Create User with Extension Pool ✅
**Configuration**:
```hcl
resource "genesyscloud_telephony_providers_edges_extension_pool" "test_pool" {
  start_number = "7700"
  end_number   = "7799"
}

resource "genesyscloud_user" "test_user" {
  email = "test@example.com"
  name  = "Test User"
  addresses {
    phone_numbers {
      extension          = "7700"
      extension_pool_id  = genesyscloud_telephony_providers_edges_extension_pool.test_pool.id
      media_type         = "PHONE"
      type              = "WORK"
    }
  }
}
```

**Results**:
- ✅ User created successfully
- ✅ Extension 7700 assigned to pool
- ✅ `extension_pool_id` stored in state file
- ✅ Extension appears in user profile
- ✅ Extension appears in Extensions Assignments UI
- ✅ No "Provider produced inconsistent result" errors

#### Test Case 2: Update User ✅
**Test**: Modify user attributes while maintaining extension pool assignment

**Results**:
- ✅ Updates applied successfully
- ✅ Extension pool assignment maintained
- ✅ No unexpected diffs or errors

#### Test Case 3: Delete User ✅
**Test**: Delete user with extension pool assignment

**Results**:
- ✅ User deleted successfully
- ✅ Extension released from pool
- ✅ Clean state removal
- ✅ No errors during deletion

#### Test Case 4: Boundary Conditions ✅
**Test**: Extensions at exact pool boundaries (7700 and 7799 in pool 7700-7799)

**Results**:
- ✅ Both boundary extensions work correctly
- ✅ Pool lookup successful
- ✅ No correlation errors

#### Test Case 5: Multiple CRUD Combinations ✅
**Test**: Various combinations of create, read, update, delete operations

**Results**:
- ✅ All combinations work correctly
- ✅ No Set identity issues observed
- ✅ State management consistent
- ✅ No unexpected behaviors

### Set Identity Validation

**Initial Concern**: Plugin Framework's inability to exclude fields from Set hash would cause issues

**Testing Result**: **NO SET IDENTITY ISSUES OBSERVED**

**Explanation**:
- The feared "Set identity problem" was based on incorrect assumptions
- SDKv2 stores pool ID in state (just excludes from hash)
- PF stores pool ID in state (includes in Set identity)
- In practice, pool assignments rarely change, so Set identity difference has no impact
- When pool assignments do change, PF shows delete+add instead of update (acceptable)

**Conclusion**: The "Set identity issue" was a misdiagnosis. The real problems were simple bugs.

---

## Solution 2 Implementation Summary

### What Was Implemented

**Approach**: Fix the bugs in the migrated code while accepting that Plugin Framework includes all fields in Set identity

**Changes Made**:
1. ✅ Uncommented `fetchExtensionPoolId()` call (Bug #1)
2. ✅ Fixed boundary condition to use >= and <= (Bug #2)
3. ✅ Migrated `waitForExtensionPoolActivation()` from SDKv2 (Migration Gap)
4. ✅ Migrated `getTelephonyExtensionPoolById()` proxy method (Migration Gap)

**Total Code Changes**: ~115 lines across 2 files

**Breaking Changes**: None

**User Impact**: Zero - existing configurations work without modification

### Why Solution 2 Was Correct

**Advantages**:
- ✅ No breaking changes for users
- ✅ Fastest implementation (actual time: ~40 minutes of coding)
- ✅ Low risk (simple bug fixes)
- ✅ Follows AWS provider patterns during SDKv2→PF migration
- ✅ Validated through comprehensive manual testing

**Trade-offs Accepted**:
- ⚠️ Pool assignment changes show as Set element replacement (delete+add) instead of in-place update
- ⚠️ This is cosmetic only - functionality works correctly
- ⚠️ In practice, pool assignments rarely change

**Alternatives Rejected**:
- Solution 1: Too complex (46-91 hours), breaking changes
- Solution 3: Functionality loss, breaking changes
- Solution 4: Very complex, uncertain benefit

---

## Before vs After Comparison

### Before Fixes

**State File**:
```json
{
  "extension_pool_id": null  ❌
}
```

**Behavior**:
- ❌ "Provider produced inconsistent result after apply" errors
- ❌ Extension pool ID not stored
- ❌ Extensions not in Extensions Assignments UI
- ❌ Boundary extensions (7700 in 7700-7799) failed

**Timing**:
- Pool created at T
- User updated at T+1.7 seconds (too fast)

### After Fixes

**State File**:
```json
{
  "extension_pool_id": "617a6999-462c-4691-a943-17c39ecabf1b"  ✅
}
```

**Behavior**:
- ✅ No errors
- ✅ Extension pool ID correctly stored
- ✅ Extensions appear in Extensions Assignments UI
- ✅ Boundary extensions work correctly
- ✅ All CRUD operations successful

**Timing**:
- Pool created at T
- Wait 5 seconds (if pool is new)
- User updated at T+5 seconds (matches SDKv2)

---

## Lessons Learned

### Technical Insights

1. **State Storage ≠ Set Identity**: These are two different concepts
   - State storage: What goes in terraform.tfstate
   - Set identity: What determines if two Set elements are "the same"
   - SDKv2 stores pool ID but excludes from hash
   - PF stores pool ID and includes in Set identity

2. **Comments Can Be Misleading**: Always verify against actual code behavior
   - Comment said "We do not store extension_pool_id in state"
   - Reality: SDKv2 DOES store it in state
   - Lesson: Trust code over comments

3. **Boundary Conditions Matter**: Always test edge cases
   - Using > and < instead of >= and <= caused failures
   - Extension 7700 in pool 7700-7799 revealed the bug
   - Lesson: Test boundaries, not just middle values

4. **Complete Migration Required**: Check for all dependencies
   - `waitForExtensionPoolActivation()` was skipped
   - Its dependency `getTelephonyExtensionPoolById()` was also skipped
   - Lesson: Verify all related functions are migrated

5. **Timing Is Critical**: Don't skip wait/sleep logic
   - 5-second wait isn't "optimization" - it's required
   - Telephony backend needs time to activate pools
   - Lesson: Respect timing-sensitive logic

### Process Improvements

1. **Verify Assumptions Early**: Check actual SDKv2 behavior before theorizing
2. **Try Simple Fixes First**: Look for commented code, typos, boundary issues
3. **Test Incrementally**: Validate each fix independently
4. **Compare Actual Behavior**: Run same tests in SDKv2 and PF, compare logs
5. **Manual Testing Is Essential**: Automated tests missed the real-world issues

---

## Metrics

### Development Effort

**Investigation**: 4 weeks  
**Implementation**: 40 minutes of actual coding  
**Testing**: Comprehensive manual testing over multiple days  
**Documentation**: Ongoing

### Code Changes

**Files Modified**: 2  
**Lines Added/Modified**: ~115  
**Breaking Changes**: 0  
**Test Cases**: 5+ manual test scenarios

### Success Metrics

**Functional Requirements**:
- ✅ extension_pool_id stored in state file
- ✅ Extension appears in user profile
- ✅ Extension appears in Extensions Assignments UI
- ✅ No "Provider produced inconsistent result" errors
- ✅ All CRUD operations work correctly
- ✅ Boundary extensions work correctly

**Non-Functional Requirements**:
- ✅ No breaking changes
- ✅ Backward compatible
- ✅ Performance matches SDKv2
- ✅ Code quality maintained

---

## Conclusion

The extension pool migration issue has been **successfully resolved** through Solution 2 implementation with comprehensive bug fixes. The initial hypothesis of a "Set identity architectural issue" was incorrect - the real problems were three simple bugs in the migrated code.

**Key Achievements**:
1. ✅ All bugs identified and fixed
2. ✅ Comprehensive manual testing completed
3. ✅ All CRUD operations validated
4. ✅ No breaking changes introduced
5. ✅ Production-ready implementation

**Status**: **RESOLVED** - Ready for production use

**Validation**: Confirmed through extensive manual testing by development team

**Recommendation**: Proceed with Plugin Framework deployment for extension pool functionality

---

## Appendix: Related Documentation

- `FINAL_SUCCESS_SUMMARY.md` - Complete fix summary with all bugs
- `SET_IDENTITY_MYSTERY_SOLVED.md` - Analysis of the misdiagnosis
- `LESSONS_LEARNED_4_WEEKS_WASTED.md` - Process improvement insights
- `IMPLEMENTATION_COMPLETE.md` - Implementation checklist and verification
- `MIGRATION_GAP_ANALYSIS.md` - Detailed analysis of what was missed in migration

---

**Document Version**: 2.0 (Updated after resolution)  
**Last Updated**: January 18, 2026  
**Status**: RESOLVED ✅  
**Validated By**: Development team through comprehensive manual testing  
**Approved For**: Production deployment
