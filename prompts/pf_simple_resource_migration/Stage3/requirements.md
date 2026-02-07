# Stage 3 – Test Migration Requirements

## Overview

Stage 3 focuses on migrating test files from Terraform Plugin SDKv2 to the Terraform Plugin Framework. This stage implements acceptance tests for resources and data sources using Framework-compatible patterns while preserving existing test coverage and behavior.

**Reference Implementation**: 
- `genesyscloud/routing_wrapupcode/resource_genesyscloud_routing_wrapupcode_test.go`
- `genesyscloud/routing_wrapupcode/data_source_genesyscloud_routing_wrapupcode_test.go`
- `genesyscloud/routing_wrapupcode/genesyscloud_routing_wrapupcode_init_test.go`

---

## Objectives

### Primary Goal
Migrate test files to use Plugin Framework patterns while maintaining comprehensive test coverage and preserving existing test behavior.

### Specific Objectives
1. Migrate resource acceptance tests to Framework patterns
2. Migrate data source acceptance tests to Framework patterns
3. Create test initialization file for package-level test setup
4. Implement muxed provider factory for mixed SDKv2/Framework testing
5. Update test helper functions to use Framework patterns
6. Ensure all tests pass with Framework implementation
7. Maintain or improve test coverage

---

## Scope

### In Scope for Stage 3

#### 1. Resource Test File
- Create `resource_genesyscloud_<resource_name>_test.go` file
- Migrate all resource acceptance tests
- Implement Framework-specific test patterns
- Create test helper functions
- Implement muxed provider factory

#### 2. Resource Test Cases
- **Basic CRUD test**: Create, read, import, verify destroy
- **Division assignment test**: Test with division reference
- **Update tests**: Test name updates, description updates
- **Lifecycle test**: Comprehensive test covering all operations
- **Destroy verification**: Verify resources are cleaned up

#### 3. Data Source Test File
- Create `data_source_genesyscloud_<resource_name>_test.go` file
- Migrate data source lookup tests
- Test data source with dependencies
- Verify data source attributes

#### 4. Test Initialization File
- Create `genesyscloud_<resource_name>_init_test.go` file
- Implement package-level test setup
- Configure test environment
- Set up test utilities

#### 5. Test Helper Functions
- `generateFramework<ResourceName>Resource()`: Generate HCL for tests
- `generateFramework<ResourceName>DataSource()`: Generate data source HCL
- `testVerifyFramework<ResourceName>Destroyed()`: Verify cleanup
- `getFrameworkProviderFactories()`: Create muxed provider

#### 6. Muxed Provider Pattern
- Support both Framework and SDKv2 resources in same test
- Enable testing Framework resource with SDKv2 dependencies
- Maintain compatibility during migration period

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
- ✅ Tests use `ProtoV6ProviderFactories` instead of SDKv2 factories

#### FR2: Resource Test Coverage
- ✅ Basic CRUD test implemented
- ✅ Division/dependency test implemented (if applicable)
- ✅ Update tests implemented (name, description, etc.)
- ✅ Lifecycle test implemented (comprehensive)
- ✅ Import test included in all test cases
- ✅ Destroy verification implemented

#### FR3: Data Source Test File
- ✅ File created: `data_source_genesyscloud_<resource_name>_test.go`
- ✅ Basic data source lookup test implemented
- ✅ Data source with dependencies test implemented (if applicable)
- ✅ Test naming follows convention: `TestAccFrameworkDataSource<ResourceName>*`

#### FR4: Test Initialization File
- ✅ File created: `genesyscloud_<resource_name>_init_test.go`
- ✅ Package-level test setup implemented
- ✅ Test utilities configured

#### FR5: Test Helper Functions
- ✅ `generateFramework<ResourceName>Resource()` implemented
- ✅ `generateFramework<ResourceName>DataSource()` implemented
- ✅ `testVerifyFramework<ResourceName>Destroyed()` implemented
- ✅ `getFrameworkProviderFactories()` implemented
- ✅ Helper functions generate valid HCL

#### FR6: Muxed Provider Factory
- ✅ Factory creates muxed provider with Framework and SDKv2 resources
- ✅ Framework resource under test is included
- ✅ SDKv2 dependencies are included (e.g., auth_division)
- ✅ Factory returns `ProtoV6ProviderServer`

#### FR7: Test Execution
- ✅ All tests pass successfully
- ✅ Tests run in isolation (no interdependencies)
- ✅ Tests clean up resources properly
- ✅ No test flakiness

### Non-Functional Requirements

#### NFR1: Test Quality
- ✅ Tests follow Go testing best practices
- ✅ Tests use descriptive names
- ✅ Tests have clear assertions
- ✅ Tests include helpful error messages

#### NFR2: Test Coverage
- ✅ Test coverage matches or exceeds SDKv2 version
- ✅ All CRUD operations are tested
- ✅ Edge cases are covered
- ✅ Error conditions are tested

#### NFR3: Test Maintainability
- ✅ Tests are easy to understand
- ✅ Test helper functions reduce duplication
- ✅ Test data is clearly defined
- ✅ Tests are well-documented

#### NFR4: Test Performance
- ✅ Tests run in reasonable time
- ✅ No unnecessary API calls
- ✅ Proper use of test parallelization (when safe)

---

## Dependencies and Prerequisites

### Prerequisites

#### 1. Stage 1 and 2 Completion
- Schema file must be complete (Stage 1)
- Resource implementation must be complete (Stage 2)
- Data source implementation must be complete (Stage 2)

#### 2. Understanding of Framework Testing
- Familiarity with Framework test patterns
- Understanding of muxed provider concept
- Knowledge of ProtoV6 provider factories

#### 3. Reference Implementation
- Study `routing_wrapupcode` test files
- Understand test patterns used
- Review muxed provider factory implementation

### Dependencies

#### 1. Package Imports (Resource Test File)
```go
import (
    "fmt"
    "testing"

    "github.com/google/uuid"
    "github.com/hashicorp/terraform-plugin-framework/datasource"
    frameworkresource "github.com/hashicorp/terraform-plugin-framework/resource"
    "github.com/hashicorp/terraform-plugin-go/tfprotov6"
    "github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
    "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
    "github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
    "github.com/mypurecloud/platform-client-sdk-go/v176/platformclientv2"
    authDivision "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/auth_division"
    "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
    "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
)
```

#### 2. Package Imports (Data Source Test File)
```go
import (
    "fmt"
    "testing"

    "github.com/google/uuid"
    "github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
    authDivision "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/auth_division"
    "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
)
```

#### 3. Test Utilities
- `util.TestAccPreCheck(t)` for test prerequisites
- `uuid.NewString()` for unique test names
- `resource.Test()` for acceptance test framework

#### 4. Dependency Resources
- SDKv2 resources needed for testing (e.g., `auth_division`)
- Must be included in muxed provider factory

---

## Constraints

### Technical Constraints

#### TC1: No Resource Implementation Changes
- **Constraint**: Tests MUST NOT require changes to resource implementation
- **Rationale**: Tests verify existing implementation, not drive new features
- **Impact**: Tests must work with Stage 2 implementation as-is

#### TC2: Muxed Provider Required
- **Constraint**: Tests MUST use muxed provider to support SDKv2 dependencies
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

### Process Constraints

#### PC1: Stage Isolation
- **Constraint**: Stage 3 MUST NOT include export testing
- **Rationale**: Clear separation of concerns
- **Impact**: Export tests are deferred to Stage 4

#### PC2: Test Coverage Requirement
- **Constraint**: Test coverage MUST NOT decrease from SDKv2 version
- **Rationale**: Maintain quality and confidence
- **Impact**: All SDKv2 tests must be migrated

---

## Validation Checklist

Use this checklist to verify Stage 3 completion:

### Resource Test File
- [ ] File created: `resource_genesyscloud_<resource_name>_test.go`
- [ ] Package declaration matches directory name
- [ ] All required imports are present
- [ ] No unused imports

### Resource Test Cases
- [ ] Basic CRUD test implemented
- [ ] Division/dependency test implemented (if applicable)
- [ ] Name update test implemented
- [ ] Description update test implemented
- [ ] Lifecycle test implemented
- [ ] All tests include import verification
- [ ] All tests use `ProtoV6ProviderFactories`
- [ ] Test names follow `TestAccFrameworkResource*` convention

### Data Source Test File
- [ ] File created: `data_source_genesyscloud_<resource_name>_test.go`
- [ ] Basic lookup test implemented
- [ ] Lookup with dependencies test implemented (if applicable)
- [ ] Tests use `ProtoV6ProviderFactories`
- [ ] Test names follow `TestAccFrameworkDataSource*` convention

### Test Initialization File
- [ ] File created: `genesyscloud_<resource_name>_init_test.go`
- [ ] Package-level setup implemented
- [ ] Test utilities configured

### Test Helper Functions
- [ ] `generateFramework<ResourceName>Resource()` implemented
- [ ] `generateFramework<ResourceName>DataSource()` implemented
- [ ] `testVerifyFramework<ResourceName>Destroyed()` implemented
- [ ] `getFrameworkProviderFactories()` implemented
- [ ] Helper functions generate valid HCL
- [ ] Helper functions handle optional attributes

### Muxed Provider Factory
- [ ] Factory creates muxed provider
- [ ] Framework resource included
- [ ] SDKv2 dependencies included
- [ ] Returns `ProtoV6ProviderServer`
- [ ] Factory is reusable across tests

### Test Execution
- [ ] All tests compile without errors
- [ ] All tests pass successfully
- [ ] Tests run in isolation
- [ ] Resources are cleaned up properly
- [ ] No test flakiness observed

### Code Quality
- [ ] Tests follow Go conventions
- [ ] Test names are descriptive
- [ ] Assertions are clear
- [ ] Error messages are helpful
- [ ] No TODO or FIXME comments (unless intentional)

---

## Example: routing_wrapupcode Test Migration

### File Structure
```
genesyscloud/routing_wrapupcode/
├── resource_genesyscloud_routing_wrapupcode_schema.go       (Stage 1)
├── resource_genesyscloud_routing_wrapupcode.go              (Stage 2)
├── data_source_genesyscloud_routing_wrapupcode.go           (Stage 2)
├── resource_genesyscloud_routing_wrapupcode_test.go         (Stage 3 - THIS)
├── data_source_genesyscloud_routing_wrapupcode_test.go      (Stage 3 - THIS)
└── genesyscloud_routing_wrapupcode_init_test.go             (Stage 3 - THIS)
```

### Key Test Patterns

#### 1. Test Function Naming
```go
func TestAccFrameworkResourceRoutingWrapupcodeBasic(t *testing.T) { ... }
func TestAccFrameworkResourceRoutingWrapupcodeDivision(t *testing.T) { ... }
func TestAccFrameworkDataSourceRoutingWrapupcode(t *testing.T) { ... }
```

#### 2. Test Structure
```go
resource.Test(t, resource.TestCase{
    PreCheck:                 func() { util.TestAccPreCheck(t) },
    ProtoV6ProviderFactories: getFrameworkProviderFactories(),
    Steps: []resource.TestStep{
        {
            Config: generateFrameworkRoutingWrapupcodeResource(...),
            Check: resource.ComposeTestCheckFunc(
                resource.TestCheckResourceAttr("genesyscloud_routing_wrapupcode."+resourceLabel, "name", name),
                resource.TestCheckResourceAttrSet("genesyscloud_routing_wrapupcode."+resourceLabel, "id"),
            ),
        },
        {
            ResourceName:      "genesyscloud_routing_wrapupcode." + resourceLabel,
            ImportState:       true,
            ImportStateVerify: true,
        },
    },
    CheckDestroy: testVerifyFrameworkWrapupcodesDestroyed,
})
```

#### 3. Muxed Provider Factory
```go
func getFrameworkProviderFactories() map[string]func() (tfprotov6.ProviderServer, error) {
    return map[string]func() (tfprotov6.ProviderServer, error){
        "genesyscloud": func() (tfprotov6.ProviderServer, error) {
            frameworkResources := map[string]func() frameworkresource.Resource{
                ResourceType: NewRoutingWrapupcodeFrameworkResource,
            }
            frameworkDataSources := map[string]func() datasource.DataSource{
                ResourceType: NewRoutingWrapupcodeFrameworkDataSource,
            }

            muxFactory := provider.NewMuxedProvider(
                "test",
                map[string]*schema.Resource{
                    authDivision.ResourceType: authDivision.ResourceAuthDivision(),
                },
                map[string]*schema.Resource{},
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

#### 4. Destroy Verification
```go
func testVerifyFrameworkWrapupcodesDestroyed(state *terraform.State) error {
    routingAPI := platformclientv2.NewRoutingApi()
    for _, rs := range state.RootModule().Resources {
        if rs.Type != "genesyscloud_routing_wrapupcode" {
            continue
        }

        wrapupcode, resp, err := routingAPI.GetRoutingWrapupcode(rs.Primary.ID)
        if wrapupcode != nil {
            return fmt.Errorf("Framework routing wrapupcode (%s) still exists", rs.Primary.ID)
        } else if util.IsStatus404(resp) {
            continue
        } else {
            return fmt.Errorf("Unexpected error checking Framework routing wrapupcode: %s", err)
        }
    }
    return nil
}
```

---

## Test Coverage Requirements

### Minimum Test Cases

#### Resource Tests
1. **Basic CRUD**: Create, read, import, destroy
2. **With Dependencies**: Test with division or other dependencies
3. **Update Name**: Verify in-place update
4. **Update Description**: Verify in-place update
5. **Lifecycle**: Comprehensive test covering all scenarios

#### Data Source Tests
1. **Basic Lookup**: Find resource by name
2. **With Dependencies**: Lookup with division or other dependencies

### Test Assertions

Each test should verify:
- [ ] Resource/data source ID is set
- [ ] All required attributes match expected values
- [ ] Optional attributes match when provided
- [ ] Computed attributes are populated
- [ ] Dependencies are correctly referenced
- [ ] Import works correctly
- [ ] Resources are destroyed properly

---

## Next Steps

After Stage 3 completion and approval:
1. Run all tests and verify they pass
2. Review test coverage
3. Address any test failures
4. Proceed to **Stage 4 – Export Functionality**

---

## References

- **Reference Implementation**: 
  - `genesyscloud/routing_wrapupcode/resource_genesyscloud_routing_wrapupcode_test.go`
  - `genesyscloud/routing_wrapupcode/data_source_genesyscloud_routing_wrapupcode_test.go`
  - `genesyscloud/routing_wrapupcode/genesyscloud_routing_wrapupcode_init_test.go`
- **Plugin Framework Testing**: https://developer.hashicorp.com/terraform/plugin/framework/acctests
- **Terraform Testing**: https://developer.hashicorp.com/terraform/plugin/testing
