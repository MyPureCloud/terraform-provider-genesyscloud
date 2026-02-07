# Stage 3 – Test Migration Requirements (Complex Resources)

## Overview

Stage 3 focuses on migrating test files from Terraform Plugin SDKv2 to the Terraform Plugin Framework for **complex resources**. This stage implements comprehensive acceptance tests for resources and data sources with nested structures, multiple dependencies, and advanced test scenarios while preserving existing test coverage and behavior.

**Reference Implementation**: 
- `genesyscloud/user/resource_genesyscloud_user_test.go`
- `genesyscloud/user/data_source_genesyscloud_user_test.go`
- `genesyscloud/user/genesyscloud_user_init_test.go`

**Key Differences from Simple Resources**:
- Complex nested structure testing (3-level nesting)
- Multiple dependency resources (SDKv2 and Framework)
- Advanced test scenarios (concurrent modification, API error handling, deleted resource restoration)
- Extensive helper function library for nested blocks
- Edge case testing for API asymmetries
- Validation testing for complex constraints

---

## Objectives

### Primary Goal
Migrate test files to use Plugin Framework patterns while maintaining comprehensive test coverage for complex resources with nested structures and multiple dependencies.

### Specific Objectives
1. Migrate resource acceptance tests to Framework patterns
2. Migrate data source acceptance tests to Framework patterns
3. Create test initialization file for package-level test setup with mixed SDKv2/Framework dependencies
4. Implement muxed provider factory supporting both SDKv2 and Framework resources
5. Create comprehensive helper function library for nested structures
6. Implement edge case tests for API asymmetries and complex scenarios
7. Ensure all tests pass with Framework implementation
8. Maintain or improve test coverage

---

## Scope

### In Scope for Stage 3

#### 1. Resource Test File
- Create `resource_genesyscloud_<resource_name>_test.go` file
- Migrate all resource acceptance tests
- Implement Framework-specific test patterns
- Create extensive helper function library for nested structures
- Implement muxed provider factory with multiple dependencies


#### 2. Resource Test Cases (Complex Resources)
- **Basic CRUD test**: Create, read, import, verify destroy
- **Nested structure tests**: Test 1-level, 2-level, and 3-level nested blocks
- **Multiple dependency tests**: Test with various SDKv2 and Framework dependencies
- **Update tests**: Test updates to nested structures, arrays, and complex attributes
- **Lifecycle test**: Comprehensive test covering all operations
- **Edge case tests**: API asymmetries, concurrent modifications, deleted resource restoration
- **Validation tests**: Complex constraint validation, proficiency ranges, required fields
- **Destroy verification**: Verify resources are cleaned up

#### 3. Data Source Test File
- Create `data_source_genesyscloud_<resource_name>_test.go` file
- Migrate data source lookup tests
- Test data source with dependencies
- Verify data source attributes
- Test multiple search criteria (email, name, etc.)

#### 4. Test Initialization File
- Create `genesyscloud_<resource_name>_init_test.go` file
- Implement package-level test setup
- Register SDKv2 dependencies
- Register Framework resources and data sources
- Configure test environment

#### 5. Test Helper Functions (Complex Resources)
- `generateFramework<ResourceName>Resource()`: Generate basic HCL
- `generateFramework<ResourceName>WithCustomAttrs()`: Generate with custom attributes
- `generateFramework<ResourceName>With<NestedBlock>()`: Generate with specific nested blocks
- `generate<NestedBlock>()`: Generate nested block HCL
- `generate<NestedBlock><SubBlock>()`: Generate sub-nested block HCL
- `generateFramework<ResourceName>DataSource()`: Generate data source HCL
- `testVerifyFramework<ResourceName>Destroyed()`: Verify cleanup
- `getFrameworkProviderFactories()`: Create muxed provider

#### 6. Muxed Provider Pattern (Complex Resources)
- Support both Framework and SDKv2 resources in same test
- Enable testing Framework resource with SDKv2 dependencies
- Enable testing Framework resource with other Framework dependencies
- Maintain compatibility during migration period
- Handle multiple dependency types (auth, routing, telephony, etc.)

### Out of Scope for Stage 3

❌ **Schema Modifications**
- No changes to schema file from Stage 1
- Schema is already complete

❌ **Resource Implementation Changes**
- No changes to resource implementation from Stage 2
- Implementation is already complete

❌ **Export Utilities**
- No export-related test changes
- Export testing is covered in Stage 4

❌ **Proxy Modifications**
- No changes to proxy files
- Proxy files remain unchanged

---

## Success Criteria

### Functional Requirements

#### FR1: Resource Test File
- ✅ File created: `resource_genesyscloud_<resource_name>_test.go`
- ✅ All SDKv2 test cases migrated to Framework patterns
- ✅ Test naming follows convention: `TestAccFrameworkResource<ResourceName>*`
- ✅ Tests use `ProtoV6ProviderFactories` or `GetMuxedProviderFactories()`


#### FR2: Resource Test Coverage (Complex Resources)
- ✅ Basic CRUD test implemented
- ✅ Nested structure tests implemented (1-level, 2-level, 3-level)
- ✅ Multiple dependency tests implemented
- ✅ Update tests for nested structures implemented
- ✅ Array/set update tests implemented (add, remove, modify elements)
- ✅ Lifecycle test implemented (comprehensive)
- ✅ Edge case tests implemented (API asymmetries, concurrent modifications)
- ✅ Validation tests implemented (constraints, ranges, required fields)
- ✅ Import test included in all test cases
- ✅ Destroy verification implemented

#### FR3: Data Source Test File
- ✅ File created: `data_source_genesyscloud_<resource_name>_test.go`
- ✅ Basic data source lookup test implemented
- ✅ Multiple search criteria tests implemented (email, name, etc.)
- ✅ Data source with dependencies test implemented (if applicable)
- ✅ Test naming follows convention: `TestAccFrameworkDataSource<ResourceName>*`

#### FR4: Test Initialization File
- ✅ File created: `genesyscloud_<resource_name>_init_test.go`
- ✅ Package-level test setup implemented
- ✅ SDKv2 dependencies registered
- ✅ Framework resources registered
- ✅ Framework data sources registered
- ✅ Test utilities configured

#### FR5: Test Helper Functions (Complex Resources)
- ✅ Basic resource generation helper implemented
- ✅ Custom attributes helper implemented
- ✅ Nested block helpers implemented (all levels)
- ✅ Sub-nested block helpers implemented
- ✅ Data source helper implemented
- ✅ Destroy verification helper implemented
- ✅ Provider factory helper implemented
- ✅ Helper functions generate valid HCL
- ✅ Helper functions handle optional attributes
- ✅ Helper functions handle nested structures

#### FR6: Muxed Provider Factory (Complex Resources)
- ✅ Factory creates muxed provider with Framework and SDKv2 resources
- ✅ Framework resource under test is included
- ✅ All SDKv2 dependencies are included (auth, routing, telephony, etc.)
- ✅ Other Framework dependencies are included (if any)
- ✅ Factory returns `ProtoV6ProviderServer`
- ✅ Factory supports multiple dependency types

#### FR7: Test Execution
- ✅ All tests pass successfully
- ✅ Tests run in isolation (no interdependencies)
- ✅ Tests clean up resources properly
- ✅ No test flakiness
- ✅ Tests handle API rate limiting
- ✅ Tests handle eventual consistency

### Non-Functional Requirements

#### NFR1: Test Quality
- ✅ Tests follow Go testing best practices
- ✅ Tests use descriptive names
- ✅ Tests have clear assertions
- ✅ Tests include helpful error messages
- ✅ Tests document edge cases and API asymmetries

#### NFR2: Test Coverage
- ✅ Test coverage matches or exceeds SDKv2 version
- ✅ All CRUD operations are tested
- ✅ All nested structures are tested
- ✅ Edge cases are covered
- ✅ Error conditions are tested
- ✅ Validation scenarios are tested

#### NFR3: Test Maintainability
- ✅ Tests are easy to understand
- ✅ Test helper functions reduce duplication
- ✅ Test data is clearly defined
- ✅ Tests are well-documented
- ✅ Helper functions are modular and reusable

#### NFR4: Test Performance
- ✅ Tests run in reasonable time
- ✅ No unnecessary API calls
- ✅ Proper use of test parallelization (when safe)
- ✅ Tests handle cleanup efficiently

---

## Dependencies and Prerequisites

### Prerequisites

#### 1. Stage 1 and 2 Completion
- Schema file must be complete (Stage 1)
- Resource implementation must be complete (Stage 2)
- Data source implementation must be complete (Stage 2)
- Utils file must be complete (Stage 2)

#### 2. Understanding of Framework Testing
- Familiarity with Framework test patterns
- Understanding of muxed provider concept
- Knowledge of ProtoV6 provider factories
- Understanding of nested structure testing

#### 3. Reference Implementation
- Study `user` test files
- Understand test patterns used
- Review muxed provider factory implementation
- Review helper function patterns


### Dependencies

#### 1. Package Imports (Resource Test File - Complex Resources)
```go
import (
    "context"
    "fmt"
    "regexp"
    "strconv"
    "strings"
    "testing"
    "time"

    "github.com/google/uuid"
    "github.com/hashicorp/terraform-plugin-framework/datasource"
    frameworkresource "github.com/hashicorp/terraform-plugin-framework/resource"
    "github.com/hashicorp/terraform-plugin-go/tfprotov6"
    "github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
    "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
    "github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
    "github.com/mypurecloud/platform-client-sdk-go/v176/platformclientv2"
    "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
    // Add imports for all SDKv2 dependencies
    authDivision "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/auth_division"
    authRole "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/auth_role"
    // Add imports for all Framework dependencies
    "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/routing_language"
    "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/routing_skill"
    // Add other dependency imports as needed
    "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
)
```

#### 2. Package Imports (Data Source Test File)
```go
import (
    "fmt"
    "log"
    "testing"
    "time"

    "github.com/google/uuid"
    "github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
    "github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
    "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
)
```

#### 3. Package Imports (Test Initialization File - Complex Resources)
```go
import (
    "sync"
    "testing"

    "github.com/hashicorp/terraform-plugin-framework/datasource"
    "github.com/hashicorp/terraform-plugin-framework/resource"
    "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
    "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud"
    // Import all SDKv2 dependencies
    authDivision "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/auth_division"
    authRole "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/auth_role"
    // Import all Framework dependencies
    routinglanguage "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/routing_language"
    // Add other dependency imports as needed
)
```

#### 4. Test Utilities
- `util.TestAccPreCheck(t)` for test prerequisites
- `uuid.NewString()` for unique test names
- `resource.Test()` for acceptance test framework
- `strconv.Quote()` for string quoting in HCL
- `util.NullValue` for omitted attributes
- `util.TrueValue` / `util.FalseValue` for boolean attributes

#### 5. Dependency Resources (Complex Resources)
- **SDKv2 resources** needed for testing (e.g., `auth_division`, `auth_role`, `location`, `telephony_providers_edges_extension_pool`)
- **Framework resources** needed for testing (e.g., `routing_language`, `routing_skill`)
- Must be included in muxed provider factory
- Must be registered in test initialization file

---

## Constraints

### Technical Constraints

#### TC1: No Resource Implementation Changes
- **Constraint**: Tests MUST NOT require changes to resource implementation
- **Rationale**: Tests verify existing implementation, not drive new features
- **Impact**: Tests must work with Stage 2 implementation as-is

#### TC2: Muxed Provider Required
- **Constraint**: Tests MUST use muxed provider to support SDKv2 and Framework dependencies
- **Rationale**: Not all resources migrated yet, need both SDKv2 and Framework
- **Impact**: Custom provider factory needed for each test file

#### TC3: Test Isolation
- **Constraint**: Tests MUST run independently without shared state
- **Rationale**: Parallel test execution and reliability
- **Impact**: Each test creates and cleans up its own resources

#### TC4: Backward Compatibility
- **Constraint**: Test behavior MUST match SDKv2 test behavior
- **Rationale**: Verify migration didn't change functionality
- **Impact**: Test assertions and expectations remain the same

#### TC5: API Asymmetry Handling
- **Constraint**: Tests MUST handle known API asymmetries (e.g., deletion behavior differences)
- **Rationale**: API may behave differently for different attribute types
- **Impact**: Some tests may need TODO comments or conditional logic


### Process Constraints

#### PC1: Stage Isolation
- **Constraint**: Stage 3 MUST NOT include export testing
- **Rationale**: Clear separation of concerns
- **Impact**: Export tests are deferred to Stage 4

#### PC2: Test Coverage Requirement
- **Constraint**: Test coverage MUST NOT decrease from SDKv2 version
- **Rationale**: Maintain quality and confidence
- **Impact**: All SDKv2 tests must be migrated

#### PC3: Edge Case Documentation
- **Constraint**: Known API asymmetries and edge cases MUST be documented with TODO comments
- **Rationale**: Track issues for future resolution
- **Impact**: Tests may include commented-out steps with detailed explanations

---

## Validation Checklist

Use this checklist to verify Stage 3 completion:

### Resource Test File
- [ ] File created: `resource_genesyscloud_<resource_name>_test.go`
- [ ] Package declaration matches directory name
- [ ] All required imports are present
- [ ] No unused imports
- [ ] Init function ensures test resources are initialized

### Resource Test Cases (Complex Resources)
- [ ] Basic CRUD test implemented
- [ ] Nested structure tests implemented (1-level, 2-level, 3-level)
- [ ] Multiple dependency tests implemented
- [ ] Update tests for nested structures implemented
- [ ] Array/set update tests implemented
- [ ] Lifecycle test implemented
- [ ] Edge case tests implemented
- [ ] Validation tests implemented
- [ ] All tests include import verification
- [ ] All tests use `ProtoV6ProviderFactories` or `GetMuxedProviderFactories()`
- [ ] Test names follow `TestAccFrameworkResource*` convention
- [ ] Tests use `t.Parallel()` where appropriate

### Data Source Test File
- [ ] File created: `data_source_genesyscloud_<resource_name>_test.go`
- [ ] Basic lookup test implemented
- [ ] Multiple search criteria tests implemented
- [ ] Lookup with dependencies test implemented (if applicable)
- [ ] Tests use `ProtoV6ProviderFactories` or `getFrameworkProviderFactories()`
- [ ] Test names follow `TestAccFrameworkDataSource*` convention
- [ ] Init function ensures test resources are initialized

### Test Initialization File
- [ ] File created: `genesyscloud_<resource_name>_init_test.go`
- [ ] Package-level setup implemented
- [ ] SDKv2 resources registered
- [ ] SDKv2 data sources registered
- [ ] Framework resources registered
- [ ] Framework data sources registered
- [ ] Test utilities configured
- [ ] TestMain function implemented

### Test Helper Functions (Complex Resources)
- [ ] `generateFramework<ResourceName>Resource()` implemented
- [ ] `generateFramework<ResourceName>WithCustomAttrs()` implemented
- [ ] Nested block helpers implemented (all levels)
- [ ] Sub-nested block helpers implemented
- [ ] `generateFramework<ResourceName>DataSource()` implemented
- [ ] `testVerifyFramework<ResourceName>Destroyed()` implemented
- [ ] `getFrameworkProviderFactories()` implemented
- [ ] Helper functions generate valid HCL
- [ ] Helper functions handle optional attributes
- [ ] Helper functions handle nested structures
- [ ] Helper functions are modular and reusable

### Muxed Provider Factory (Complex Resources)
- [ ] Factory creates muxed provider
- [ ] Framework resource included
- [ ] All SDKv2 dependencies included
- [ ] All Framework dependencies included
- [ ] Returns `ProtoV6ProviderServer`
- [ ] Factory is reusable across tests

### Test Execution
- [ ] All tests compile without errors
- [ ] All tests pass successfully
- [ ] Tests run in isolation
- [ ] Resources are cleaned up properly
- [ ] No test flakiness observed
- [ ] Tests handle API rate limiting
- [ ] Tests handle eventual consistency

### Code Quality
- [ ] Tests follow Go conventions
- [ ] Test names are descriptive
- [ ] Assertions are clear
- [ ] Error messages are helpful
- [ ] Edge cases are documented with TODO comments
- [ ] No unnecessary TODO or FIXME comments

---

## Example: user Test Migration

### File Structure
```
genesyscloud/user/
├── resource_genesyscloud_user_schema.go                     (Stage 1)
├── resource_genesyscloud_user.go                            (Stage 2)
├── resource_genesyscloud_user_utils.go                      (Stage 2)
├── data_source_genesyscloud_user.go                         (Stage 2)
├── resource_genesyscloud_user_test.go                       (Stage 3 - THIS)
├── data_source_genesyscloud_user_test.go                    (Stage 3 - THIS)
└── genesyscloud_user_init_test.go                           (Stage 3 - THIS)
```


### Key Test Patterns (Complex Resources)

#### 1. Test Function Naming
```go
func TestAccFrameworkResourceUserBasic(t *testing.T) { ... }
func TestAccFrameworkResourceUserWithProfileSkillsAndCertifications(t *testing.T) { ... }
func TestAccFrameworkResourceUserAddresses(t *testing.T) { ... }
func TestAccFrameworkResourceUserSkillsAndLanguages(t *testing.T) { ... }
func TestAccFrameworkResourceUserRoutingUtilizationBasic(t *testing.T) { ... }
func TestAccFrameworkResourceUserRoutingUtilizationWithLabels(t *testing.T) { ... }
func TestAccFrameworkResourceUserValidation(t *testing.T) { ... }
func TestAccFrameworkResourceUserDeletedUserRestoration(t *testing.T) { ... }
func TestAccFrameworkResourceUserConcurrentModification(t *testing.T) { ... }
func TestAccFrameworkDataSourceUser(t *testing.T) { ... }
```

#### 2. Test Structure (Complex Resource)
```go
resource.Test(t, resource.TestCase{
    PreCheck: func() { util.TestAccPreCheck(t) },
    ProtoV6ProviderFactories: provider.GetMuxedProviderFactories(
        map[string]*schema.Resource{
            // SDKv2 dependencies
            authDivision.ResourceType: authDivision.ResourceAuthDivision(),
        },
        nil, // SDKv2 data sources
        map[string]func() frameworkresource.Resource{
            ResourceType: NewUserFrameworkResource,
        },
        map[string]func() datasource.DataSource{
            ResourceType: NewUserFrameworkDataSource,
        },
    ),
    Steps: []resource.TestStep{
        {
            Config: generateFrameworkUserResource(...),
            Check: resource.ComposeTestCheckFunc(
                resource.TestCheckResourceAttr("genesyscloud_user."+resourceLabel, "name", name),
                resource.TestCheckResourceAttrSet("genesyscloud_user."+resourceLabel, "id"),
                // Check nested attributes
                resource.TestCheckResourceAttr("genesyscloud_user."+resourceLabel, "addresses.0.phone_numbers.0.number", phone),
            ),
        },
        {
            ResourceName:            "genesyscloud_user." + resourceLabel,
            ImportState:             true,
            ImportStateVerify:       true,
            ImportStateVerifyIgnore: []string{"password"}, // Password not returned by API
        },
    },
    CheckDestroy: testVerifyUsersDestroyed,
})
```

#### 3. Muxed Provider Factory (Complex Resource)
```go
func getFrameworkProviderFactories() map[string]func() (tfprotov6.ProviderServer, error) {
    return map[string]func() (tfprotov6.ProviderServer, error){
        "genesyscloud": func() (tfprotov6.ProviderServer, error) {
            frameworkResources := map[string]func() frameworkresource.Resource{
                ResourceType: NewUserFrameworkResource,
            }
            frameworkDataSources := map[string]func() datasource.DataSource{
                ResourceType: NewUserFrameworkDataSource,
            }

            muxFactory := provider.NewMuxedProvider(
                "test",
                map[string]*schema.Resource{}, // SDKv2 resources (add dependencies if needed)
                map[string]*schema.Resource{}, // SDKv2 data sources (add dependencies if needed)
                frameworkResources,
                frameworkDataSources,
            )

            serverFactory, err := muxFactory()
            if err != nil {
                return nil, err
            }

            return serverFactory(), nil
        },
    }
}
```

#### 4. Destroy Verification (Complex Resource)
```go
func testVerifyUsersDestroyed(state *terraform.State) error {
    usersAPI := platformclientv2.NewUsersApi()
    for _, rs := range state.RootModule().Resources {
        if rs.Type != "genesyscloud_user" {
            continue
        }

        user, resp, err := usersAPI.GetUser(rs.Primary.ID, nil, "", "")
        if user != nil {
            return fmt.Errorf("Framework user (%s) still exists", rs.Primary.ID)
        } else if util.IsStatus404(resp) {
            continue
        } else {
            return fmt.Errorf("Unexpected error checking Framework user: %s", err)
        }
    }
    return nil
}
```

#### 5. Helper Functions (Complex Resource Examples)

**Basic Resource Generation**:
```go
func generateFrameworkUserResource(
    resourceLabel string,
    email string,
    name string,
    state string,
    title string,
    department string,
    manager string,
    acdAutoAnswer string,
) string {
    // Generate HCL with optional attributes
}
```

**Custom Attributes Generation**:
```go
func generateFrameworkUserWithCustomAttrs(resourceLabel, email, name string, attrs ...string) string {
    return fmt.Sprintf(`resource "%s" "%s" {
        email = "%s"
        name = "%s"
        %s
    }`, ResourceType, resourceLabel, email, name, strings.Join(attrs, "\n"))
}
```

**Nested Block Generation**:
```go
func generateFrameworkUserAddresses(nestedBlocks ...string) string {
    var phoneBlocks []string
    for _, block := range nestedBlocks {
        phoneBlocks = append(phoneBlocks, fmt.Sprintf("phone_numbers {\n%s\n}", block))
    }
    return fmt.Sprintf("addresses {\n%s\n}", strings.Join(phoneBlocks, "\n"))
}
```

**Sub-Nested Block Generation**:
```go
func generateFrameworkUserPhoneAddress(phoneNum, phoneMediaType, phoneType, extension string, extras ...string) string {
    return fmt.Sprintf(`number = %s
        media_type = %s
        type = %s
        extension = %s
        %s`, phoneNum, phoneMediaType, phoneType, extension, strings.Join(extras, "\n"))
}
```

---

## Test Coverage Requirements (Complex Resources)

### Minimum Test Cases

#### Resource Tests
1. **Basic CRUD**: Create, read, import, destroy with required attributes
2. **Nested Structures**: Test 1-level, 2-level, and 3-level nested blocks
3. **With Dependencies**: Test with SDKv2 and Framework dependencies
4. **Update Tests**: Name, description, nested structures, arrays
5. **Lifecycle**: Comprehensive test covering all scenarios
6. **Edge Cases**: API asymmetries, concurrent modifications, deleted resource restoration
7. **Validation**: Complex constraints, proficiency ranges, required fields

#### Data Source Tests
1. **Basic Lookup**: Find resource by primary identifier (email)
2. **Alternative Lookup**: Find resource by alternative identifier (name)
3. **With Dependencies**: Lookup with dependencies (if applicable)


### Test Assertions (Complex Resources)

Each test should verify:
- [ ] Resource/data source ID is set
- [ ] All required attributes match expected values
- [ ] Optional attributes match when provided
- [ ] Computed attributes are populated
- [ ] Nested attributes are correctly set (all levels)
- [ ] Array/set elements are correctly populated
- [ ] Dependencies are correctly referenced
- [ ] Import works correctly
- [ ] Resources are destroyed properly
- [ ] Edge cases are handled correctly

### Edge Cases to Test (Complex Resources)

#### 1. Nested Structure Edge Cases
- Creating with minimal nested attributes
- Creating with all nested attributes
- Updating nested attributes individually
- Removing nested attributes
- Adding nested attributes
- Testing with various nesting levels

#### 2. Array/Set Edge Cases
- Empty arrays/sets
- Single element arrays/sets
- Multiple element arrays/sets
- Adding elements to arrays/sets
- Removing elements from arrays/sets
- Modifying elements in arrays/sets
- Reordering elements (for sets, verify order-independence)

#### 3. API Asymmetry Edge Cases
- Deletion behavior differences (e.g., phone_numbers vs other_emails)
- Extension-only phone numbers
- Number without extension to number with extension
- Extension-only to number without extension
- E.164 format validation

#### 4. Validation Edge Cases
- Proficiency ranges (0.0 to 5.0)
- Required field validation
- Format validation (email, phone, etc.)
- Constraint validation (max capacity, etc.)

#### 5. Concurrent Modification Edge Cases
- Concurrent updates to same resource
- Retry logic for version conflicts
- State consistency after conflicts

#### 6. Deleted Resource Edge Cases
- Restoring deleted resources
- Handling 404 errors gracefully
- State cleanup after deletion

---

## Complex Resource Test Patterns

### Pattern 1: Testing Nested Structures

**1-Level Nesting** (e.g., `addresses`):
```go
resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "addresses.#", "1")
resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "addresses.0.phone_numbers.#", "1")
```

**2-Level Nesting** (e.g., `addresses.phone_numbers`):
```go
resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "addresses.0.phone_numbers.0.number", phone)
resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "addresses.0.phone_numbers.0.media_type", "PHONE")
```

**3-Level Nesting** (e.g., `routing_utilization.call.attributes`):
```go
resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "routing_utilization.0.call.0.maximum_capacity", "3")
resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "routing_utilization.0.call.0.include_non_acd", "true")
resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "routing_utilization.0.call.0.interruptible_media_types.#", "2")
```

### Pattern 2: Testing Arrays/Sets

**Check Array Length**:
```go
resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "profile_skills.#", "2")
```

**Check Array Elements**:
```go
resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "profile_skills.0", "Java")
resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "profile_skills.1", "Go")
```

**Check Set Elements** (order-independent):
```go
resource.TestCheckTypeSetElemNestedAttrs(
    ResourceType+"."+resourceLabel,
    "addresses.0.phone_numbers.*",
    map[string]string{
        "number":     phone1,
        "media_type": "PHONE",
        "type":       "WORK",
    },
)
```

### Pattern 3: Testing Dependencies

**SDKv2 Dependency**:
```go
Config: authDivision.GenerateAuthDivisionBasic(divResourceLabel, divName) +
    generateFrameworkUserResource(...),
Check: resource.ComposeTestCheckFunc(
    resource.TestCheckResourceAttrPair(
        ResourceType+"."+resourceLabel,
        "division_id",
        "genesyscloud_auth_division."+divResourceLabel,
        "id",
    ),
),
```

**Framework Dependency**:
```go
Config: routing_skill.GenerateRoutingSkillResource(skillResourceLabel, skillName) +
    generateFrameworkUserWithSkillsAndLanguages(...),
Check: resource.ComposeTestCheckFunc(
    resource.TestCheckResourceAttrPair(
        ResourceType+"."+resourceLabel,
        "routing_skills.0.skill_id",
        "genesyscloud_routing_skill."+skillResourceLabel,
        "id",
    ),
),
```

### Pattern 4: Testing Updates

**Update Nested Attribute**:
```go
Steps: []resource.TestStep{
    {
        Config: generateFrameworkUserWithAddresses(..., phone1, ...),
        Check: resource.ComposeTestCheckFunc(
            resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "addresses.0.phone_numbers.0.number", phone1),
        ),
    },
    {
        Config: generateFrameworkUserWithAddresses(..., phone2, ...),
        Check: resource.ComposeTestCheckFunc(
            resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "addresses.0.phone_numbers.0.number", phone2),
        ),
    },
}
```

**Update Array Elements**:
```go
Steps: []resource.TestStep{
    {
        Config: generateFrameworkUserWithProfileAttrs(..., generateProfileSkills("Java"), ...),
        Check: resource.ComposeTestCheckFunc(
            resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "profile_skills.0", "Java"),
        ),
    },
    {
        Config: generateFrameworkUserWithProfileAttrs(..., generateProfileSkills("Go"), ...),
        Check: resource.ComposeTestCheckFunc(
            resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "profile_skills.0", "Go"),
        ),
    },
}
```

### Pattern 5: Testing Edge Cases

**API Asymmetry with TODO Comment**:
```go
// TODO (ADDRESSES-DELETION-ASYMMETRY): This test step expects addresses to be fully removed
// when omitted from config (addresses.# = 0). However, the Genesys Cloud API exhibits
// asymmetric deletion behavior: when an empty Addresses array is sent via PATCH,
// phone_numbers (PHONE/SMS media types) ARE deleted, but other_emails (EMAIL media type)
// are NOT deleted. This may be intentional API behavior rather than a bug.
//
// Current status: Test may fail until resolution approach is decided and implemented.
/*{
    Config: generateFrameworkUserResource(...), // No addresses
    Check: resource.ComposeTestCheckFunc(
        resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "addresses.#", "0"),
    ),
},*/
```

**Extension-Only Phone Number**:
```go
{
    Config: generateFrameworkUserWithAddresses(
        ...,
        generateFrameworkUserPhoneAddress(
            util.NullValue,        // No number
            util.NullValue,        // Default to PHONE
            util.NullValue,        // Default to WORK
            strconv.Quote(phone1), // Extension using phone1 value
        ),
        ...,
    ),
    Check: resource.ComposeTestCheckFunc(
        resource.TestCheckNoResourceAttr(ResourceType+"."+resourceLabel, "addresses.0.phone_numbers.0.number"),
        resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "addresses.0.phone_numbers.0.extension", phone1),
    ),
}
```

### Pattern 6: Testing Validation

**Proficiency Range Validation**:
```go
{
    Config: generateFrameworkUserWithSkillsAndLanguages(
        ...,
        generateFrameworkUserRoutingSkill(skillID, "6.0"), // Invalid: > 5.0
        ...,
    ),
    ExpectError: regexp.MustCompile(`expected routing_skills\[0\]\.proficiency to be in the range \(0\.000000 - 5\.000000\)`),
}
```

**Required Field Validation**:
```go
{
    Config: generateFrameworkUserResource(
        ...,
        "", // Empty email (required field)
        ...,
    ),
    ExpectError: regexp.MustCompile(`expected "email" to not be an empty string`),
}
```

---

## Next Steps

After Stage 3 completion and approval:
1. Run all tests and verify they pass
2. Review test coverage
3. Address any test failures
4. Document any known API asymmetries
5. Proceed to **Stage 4 – Export Functionality**

---

## References

- **Reference Implementation**: 
  - `genesyscloud/user/resource_genesyscloud_user_test.go`
  - `genesyscloud/user/data_source_genesyscloud_user_test.go`
  - `genesyscloud/user/genesyscloud_user_init_test.go`
- **Simple Resource Reference**: `prompts/pf_simple_resource_migration/Stage3/requirements.md`
- **Plugin Framework Testing**: https://developer.hashicorp.com/terraform/plugin/framework/acctests
- **Terraform Testing**: https://developer.hashicorp.com/terraform/plugin/testing

