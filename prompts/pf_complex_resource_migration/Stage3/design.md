# Stage 3 – Test Migration Design (Complex Resources)

## Overview

This document describes the design decisions, architecture, and patterns used in Stage 3 of the Plugin Framework migration for **complex resources**. Stage 3 focuses on migrating acceptance tests to use Plugin Framework patterns while maintaining comprehensive test coverage for resources with nested structures, multiple dependencies, and advanced test scenarios.

**Reference Implementation**: 
- `genesyscloud/user/resource_genesyscloud_user_test.go`
- `genesyscloud/user/data_source_genesyscloud_user_test.go`
- `genesyscloud/user/genesyscloud_user_init_test.go`

**Key Differences from Simple Resources**:
- Complex nested structure testing (3-level nesting)
- Multiple dependency management (SDKv2 and Framework)
- Extensive helper function library
- Edge case testing for API asymmetries
- Advanced test scenarios (concurrent modification, deleted resource restoration)
- Validation testing for complex constraints

---

## Design Principles

### 1. Muxed Provider Pattern (Complex Resources)
**Principle**: Use muxed provider to support both Framework and SDKv2 resources with multiple dependencies.

**Rationale**:
- Complex resources often depend on multiple other resources
- Some dependencies are SDKv2, some are Framework
- Enables testing Framework resources with real dependencies
- Allows gradual migration without breaking existing tests
- Supports mixed provider scenarios during migration

**Implementation**:
- Create custom provider factory for each test file
- Include Framework resource under test
- Include all SDKv2 dependencies (auth, routing, telephony, etc.)
- Include all Framework dependencies (routing_language, routing_skill, etc.)
- Return ProtoV6ProviderServer

**Example** (user resource with multiple dependencies):
```go
ProtoV6ProviderFactories: provider.GetMuxedProviderFactories(
    map[string]*schema.Resource{
        // SDKv2 dependencies
        authDivision.ResourceType: authDivision.ResourceAuthDivision(),
        authRole.ResourceType: authRole.ResourceAuthRole(),
        location.ResourceType: location.ResourceLocation(),
        extensionPool.ResourceType: extensionPool.ResourceTelephonyExtensionPool(),
    },
    nil, // SDKv2 data sources
    map[string]func() frameworkresource.Resource{
        // Framework resources
        ResourceType: NewUserFrameworkResource,
        routinglanguage.ResourceType: routinglanguage.NewFrameworkRoutingLanguageResource,
    },
    map[string]func() datasource.DataSource{
        // Framework data sources
        ResourceType: NewUserFrameworkDataSource,
        routinglanguage.ResourceType: routinglanguage.NewFrameworkRoutingLanguageDataSource,
    },
),
```

### 2. Modular Helper Function Library
**Principle**: Create extensive, modular helper function library for generating HCL configurations.

**Rationale**:
- Complex resources have many nested structures
- Tests need to generate various combinations of attributes
- Modular helpers enable flexible test configuration
- Reduces duplication and improves maintainability
- Enables testing specific scenarios without full resource configuration

**Implementation**:
- Basic resource generation helper
- Custom attributes helper (variadic pattern)
- Nested block helpers (one per nested block type)
- Sub-nested block helpers (for 2-level and 3-level nesting)
- Attribute-specific helpers (for complex attributes)

**Example** (modular helper pattern):
```go
// Basic resource generation
generateFrameworkUserResource(resourceLabel, email, name, state, title, department, manager, acdAutoAnswer)

// Custom attributes (flexible)
generateFrameworkUserWithCustomAttrs(resourceLabel, email, name, 
    generateFrameworkUserAddresses(...),
    generateFrameworkRoutingUtilization(...),
)

// Nested block generation
generateFrameworkUserAddresses(
    generateFrameworkUserPhoneAddress(...),
    generateFrameworkUserEmailAddress(...),
)

// Sub-nested block generation
generateFrameworkUserPhoneAddress(phoneNum, phoneMediaType, phoneType, extension, extras...)
```


### 3. Test Naming Convention (Complex Resources)
**Principle**: Use descriptive naming convention that clearly identifies test purpose and complexity.

**Rationale**:
- Easy to identify which tests are migrated
- Clear indication of what each test covers
- Prevents naming conflicts during migration
- Helps track migration progress
- Enables filtering tests by scenario

**Implementation**:
- Resource tests: `TestAccFrameworkResource<ResourceName><Scenario>`
- Data source tests: `TestAccFrameworkDataSource<ResourceName><Scenario>`
- Scenario names describe what is being tested
- Examples: `Basic`, `WithProfileSkillsAndCertifications`, `Addresses`, `RoutingUtilizationWithLabels`, `Validation`, `DeletedUserRestoration`

### 4. Test Coverage Preservation and Enhancement
**Principle**: Maintain or improve test coverage from SDKv2 version, with focus on complex scenarios.

**Rationale**:
- Verify migration didn't break functionality
- Maintain confidence in implementation
- Catch regressions early
- Document expected behavior for complex scenarios
- Test edge cases and API asymmetries

**Implementation**:
- Migrate all SDKv2 test cases
- Add tests for nested structure scenarios
- Add tests for edge cases (API asymmetries, concurrent modifications)
- Add validation tests for complex constraints
- Test all CRUD operations on nested structures

### 5. Edge Case Documentation
**Principle**: Document known API asymmetries and edge cases with detailed TODO comments.

**Rationale**:
- Track issues for future resolution
- Provide context for test failures
- Document API behavior differences
- Help future developers understand constraints
- Enable informed decision-making

**Implementation**:
- Use TODO comments with issue identifiers
- Explain the problem in detail
- Document why it worked in SDKv2 but fails in Framework
- Describe potential resolution approaches
- Comment out failing test steps with full context

**Example**:
```go
// TODO (ADDRESSES-DELETION-ASYMMETRY): This test step expects addresses to be fully removed
// when omitted from config (addresses.# = 0). However, the Genesys Cloud API exhibits
// asymmetric deletion behavior: when an empty Addresses array is sent via PATCH,
// phone_numbers (PHONE/SMS media types) ARE deleted, but other_emails (EMAIL media type)
// are NOT deleted. This may be intentional API behavior rather than a bug.
//
// Why this worked in SDK v2 but fails in Plugin Framework:
// - SDK v2: After update, the Read function populated other_emails back into state from the API.
//   The state management silently accepted this mismatch between config (no addresses) and
//   state (has other_emails). No drift was detected, but state was inconsistent with config.
// - Plugin Framework: PF has stricter state consistency checks. After update, when Read tries
//   to populate other_emails from API but config says addresses should be null, PF detects
//   the inconsistency and throws error: "Provider produced inconsistent result after apply -
//   block count changed from 0 to 1". This is actually BETTER behavior - it catches the problem
//   instead of silently accepting inconsistent state.
//
// Resolution options:
// 1. Confirm with Genesys Cloud if this is intentional API behavior or a bug
// 2. If API behavior won't change: Implement explicit deletion logic in updateUser()
// 3. Alternative: Use UseStateForUnknown() plan modifier on addresses block
```

### 6. Test Initialization Pattern
**Principle**: Use package-level initialization to register all test dependencies.

**Rationale**:
- Ensures all dependencies are available for tests
- Centralizes dependency management
- Prevents runtime errors from missing dependencies
- Supports both SDKv2 and Framework dependencies
- Thread-safe registration with mutexes

**Implementation**:
- Create init_test.go file
- Register SDKv2 resources and data sources
- Register Framework resources and data sources
- Use sync.RWMutex for thread-safe registration
- Call initialization in TestMain and init functions

---

## Architecture

### File Structure

```
genesyscloud/<resource_name>/
├── resource_genesyscloud_<resource_name>_schema.go          ← Stage 1
├── resource_genesyscloud_<resource_name>.go                 ← Stage 2
├── resource_genesyscloud_<resource_name>_utils.go           ← Stage 2
├── data_source_genesyscloud_<resource_name>.go              ← Stage 2
├── resource_genesyscloud_<resource_name>_test.go            ← Stage 3 (THIS)
├── data_source_genesyscloud_<resource_name>_test.go         ← Stage 3 (THIS)
├── genesyscloud_<resource_name>_init_test.go                ← Stage 3 (THIS)
├── resource_genesyscloud_<resource_name>_export_utils.go    ← Stage 4
└── genesyscloud_<resource_name>_proxy.go                    ← NOT MODIFIED
```

### Resource Test File Components (Complex Resources)

```
┌─────────────────────────────────────────────────────────────────┐
│  resource_genesyscloud_<resource_name>_test.go                  │
├─────────────────────────────────────────────────────────────────┤
│  1. Init Function                                               │
│     - Ensures test resources are initialized                    │
├─────────────────────────────────────────────────────────────────┤
│  2. Provider Factory Function                                   │
│     - getFrameworkProviderFactories()                           │
│     - Creates muxed provider with all dependencies              │
├─────────────────────────────────────────────────────────────────┤
│  3. Test Functions (15+ tests for complex resources)            │
│     - TestAccFrameworkResource<ResourceName>Basic               │
│     - TestAccFrameworkResource<ResourceName>WithNestedStructure │
│     - TestAccFrameworkResource<ResourceName>Dependencies        │
│     - TestAccFrameworkResource<ResourceName>Updates             │
│     - TestAccFrameworkResource<ResourceName>Lifecycle           │
│     - TestAccFrameworkResource<ResourceName>EdgeCases           │
│     - TestAccFrameworkResource<ResourceName>Validation          │
├─────────────────────────────────────────────────────────────────┤
│  4. Destroy Verification Function                               │
│     - testVerifyFramework<ResourceName>Destroyed()              │
├─────────────────────────────────────────────────────────────────┤
│  5. Helper Function Library (20+ helpers)                       │
│     - Basic resource generation                                 │
│     - Custom attributes generation                              │
│     - Nested block generation (1-level, 2-level, 3-level)       │
│     - Sub-nested block generation                               │
│     - Attribute-specific helpers                                │
└─────────────────────────────────────────────────────────────────┘
```

### Data Source Test File Components

```
┌─────────────────────────────────────────────────────────────────┐
│  data_source_genesyscloud_<resource_name>_test.go               │
├─────────────────────────────────────────────────────────────────┤
│  1. Init Function                                               │
│     - Ensures test resources are initialized                    │
├─────────────────────────────────────────────────────────────────┤
│  2. Test Functions                                              │
│     - TestAccFrameworkDataSource<ResourceName>                  │
│     - Multiple search criteria tests                            │
├─────────────────────────────────────────────────────────────────┤
│  3. HCL Generation Helper                                       │
│     - generateFramework<ResourceName>DataSource()               │
└─────────────────────────────────────────────────────────────────┘
```

### Test Initialization File Components

```
┌─────────────────────────────────────────────────────────────────┐
│  genesyscloud_<resource_name>_init_test.go                      │
├─────────────────────────────────────────────────────────────────┤
│  1. Package-Level Variables                                     │
│     - providerResources (SDKv2)                                 │
│     - providerDataSources (SDKv2)                               │
│     - frameworkResources (Framework)                            │
│     - frameworkDataSources (Framework)                          │
├─────────────────────────────────────────────────────────────────┤
│  2. Registration Instance                                       │
│     - registerTestInstance struct with mutexes                  │
├─────────────────────────────────────────────────────────────────┤
│  3. Registration Methods                                        │
│     - registerTestResources() - SDKv2 resources                 │
│     - registerTestDataSources() - SDKv2 data sources            │
│     - registerFrameworkTestResources() - Framework resources    │
│     - registerFrameworkTestDataSources() - Framework DS         │
├─────────────────────────────────────────────────────────────────┤
│  4. Initialization Function                                     │
│     - initTestResources() - Calls all registration methods      │
├─────────────────────────────────────────────────────────────────┤
│  5. TestMain Function                                           │
│     - Calls initTestResources() before running tests            │
└─────────────────────────────────────────────────────────────────┘
```

---

## Component Design

## Part 1: Resource Test File

### 1. Init Function

**Purpose**: Ensure test resources are initialized before any tests run.

**Design Pattern**:
```go
// Ensure test resources are initialized for Framework tests
func init() {
    if frameworkResources == nil || frameworkDataSources == nil {
        initTestResources()
    }
}
```

**Why Needed**:
- Tests may run before TestMain
- Ensures dependencies are registered
- Prevents nil pointer errors
- Idempotent initialization


### 2. Provider Factory Function (Complex Resources)

**Purpose**: Create muxed provider that supports both Framework and SDKv2 resources with multiple dependencies.

**Design Pattern**:
```go
// getFrameworkProviderFactories returns provider factories for Framework testing.
// This creates a muxed provider that includes:
//   - Framework resources: genesyscloud_<resource> (for creating test resources)
//   - Framework data sources: genesyscloud_<resource> (for testing data source lookups)
//   - SDKv2 resources: Any dependencies needed (e.g., auth_division, location, etc.)
//
// The muxed provider allows tests to use both Framework and SDKv2 resources together.
func getFrameworkProviderFactories() map[string]func() (tfprotov6.ProviderServer, error) {
    return map[string]func() (tfprotov6.ProviderServer, error){
        "genesyscloud": func() (tfprotov6.ProviderServer, error) {
            // Define Framework resources for testing
            frameworkResources := map[string]func() frameworkresource.Resource{
                ResourceType: New<ResourceName>FrameworkResource,
            }

            // Define Framework data sources for testing
            frameworkDataSources := map[string]func() datasource.DataSource{
                ResourceType: New<ResourceName>FrameworkDataSource,
            }

            // Create muxed provider that includes both Framework and SDKv2 resources
            // This allows the test to use SDKv2 dependencies alongside Framework resource
            muxFactory := provider.NewMuxedProvider(
                "test",
                map[string]*schema.Resource{}, // SDKv2 resources (add dependencies here if needed)
                map[string]*schema.Resource{}, // SDKv2 data sources (add dependencies here if needed)
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

**Example** (user resource with multiple dependencies):
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
                map[string]*schema.Resource{}, // Empty - no SDKv2 dependencies in this factory
                map[string]*schema.Resource{}, // Empty - no SDKv2 data sources
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

**Alternative Pattern** (using provider.GetMuxedProviderFactories):
```go
ProtoV6ProviderFactories: provider.GetMuxedProviderFactories(
    map[string]*schema.Resource{
        // SDKv2 dependencies
        authDivision.ResourceType: authDivision.ResourceAuthDivision(),
        authRole.ResourceType: authRole.ResourceAuthRole(),
        location.ResourceType: location.ResourceLocation(),
        extensionPool.ResourceType: extensionPool.ResourceTelephonyExtensionPool(),
    },
    nil, // SDKv2 data sources
    map[string]func() frameworkresource.Resource{
        ResourceType: NewUserFrameworkResource,
        routinglanguage.ResourceType: routinglanguage.NewFrameworkRoutingLanguageResource,
    },
    map[string]func() datasource.DataSource{
        ResourceType: NewUserFrameworkDataSource,
        routinglanguage.ResourceType: routinglanguage.NewFrameworkRoutingLanguageDataSource,
    },
),
```

**Key Points**:
- Use `getFrameworkProviderFactories()` for simple cases (no SDKv2 dependencies)
- Use `provider.GetMuxedProviderFactories()` inline for complex cases (multiple dependencies)
- Include all SDKv2 dependencies needed by tests
- Include all Framework dependencies needed by tests
- Return ProtoV6ProviderServer

### 3. Test Function Structure (Complex Resources)

**Purpose**: Define acceptance test cases for complex resource CRUD operations with nested structures.

**Design Pattern**:
```go
func TestAccFrameworkResource<ResourceName><Scenario>(t *testing.T) {
    t.Parallel() // Enable parallel execution
    var (
        resourceLabel = "test_<resource>"
        // Define test variables
        name          = "Terraform Framework <Resource> " + uuid.NewString()
        // ... other test variables
    )

    resource.Test(t, resource.TestCase{
        PreCheck: func() { util.TestAccPreCheck(t) },
        ProtoV6ProviderFactories: provider.GetMuxedProviderFactories(
            map[string]*schema.Resource{
                // SDKv2 dependencies
            },
            nil, // SDKv2 data sources
            map[string]func() frameworkresource.Resource{
                ResourceType: New<ResourceName>FrameworkResource,
            },
            map[string]func() datasource.DataSource{
                ResourceType: New<ResourceName>FrameworkDataSource,
            },
        ),
        Steps: []resource.TestStep{
            {
                // Create
                Config: generateFramework<ResourceName>Resource(...),
                Check: resource.ComposeTestCheckFunc(
                    resource.TestCheckResourceAttr("genesyscloud_<resource>."+resourceLabel, "name", name),
                    resource.TestCheckResourceAttrSet("genesyscloud_<resource>."+resourceLabel, "id"),
                    // Check nested attributes
                    resource.TestCheckResourceAttr("genesyscloud_<resource>."+resourceLabel, "nested_block.0.attribute", value),
                ),
            },
            {
                // Import/Read
                ResourceName:            "genesyscloud_<resource>." + resourceLabel,
                ImportState:             true,
                ImportStateVerify:       true,
                ImportStateVerifyIgnore: []string{"password"}, // Attributes not returned by API
            },
        },
        CheckDestroy: testVerifyFramework<ResourceName>Destroyed,
    })
}
```

**Example** (user basic test):
```go
func TestAccFrameworkResourceUserBasic(t *testing.T) {
    t.Parallel()
    var (
        userResourceLabel = "test-user-framework"
        email1            = "terraform-framework-" + uuid.NewString() + "@user.com"
        email2            = "terraform-framework-" + uuid.NewString() + "@user.com"
        userName1         = "John Framework"
        userName2         = "Jane Framework"
        stateActive       = "active"
        stateInactive     = "inactive"
        title1            = "Senior Developer"
        title2            = "Project Lead"
        department1       = "Engineering"
        department2       = "Product"
    )

    resource.Test(t, resource.TestCase{
        PreCheck: func() { util.TestAccPreCheck(t) },
        ProtoV6ProviderFactories: provider.GetMuxedProviderFactories(
            nil, // SDKv2 resources removed
            nil, // SDKv2 data sources removed
            map[string]func() frameworkresource.Resource{
                ResourceType: NewUserFrameworkResource,
            },
            map[string]func() datasource.DataSource{
                ResourceType: NewUserFrameworkDataSource,
            },
        ),
        Steps: []resource.TestStep{
            {
                // Create basic user
                Config: generateFrameworkUserResource(
                    userResourceLabel,
                    email1,
                    userName1,
                    util.NullValue, // Defaults to active
                    strconv.Quote(title1),
                    strconv.Quote(department1),
                    util.NullValue, // No manager
                    util.NullValue, // Default acdAutoAnswer
                ),
                Check: resource.ComposeTestCheckFunc(
                    resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel, "email", email1),
                    resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel, "name", userName1),
                    resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel, "state", stateActive),
                    resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel, "title", title1),
                    resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel, "department", department1),
                    resource.TestCheckNoResourceAttr(ResourceType+"."+userResourceLabel, "manager"),
                    resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel, "acd_auto_answer", "false"),
                    resource.TestCheckResourceAttrSet(ResourceType+"."+userResourceLabel, "id"),
                    resource.TestCheckResourceAttrSet(ResourceType+"."+userResourceLabel, "division_id"),
                ),
            },
            {
                // Update user attributes
                Config: generateFrameworkUserResource(
                    userResourceLabel,
                    email2,
                    userName2,
                    strconv.Quote(stateInactive),
                    strconv.Quote(title2),
                    strconv.Quote(department2),
                    util.NullValue, // No manager
                    util.TrueValue, // AcdAutoAnswer
                ),
                Check: resource.ComposeTestCheckFunc(
                    resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel, "email", email2),
                    resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel, "name", userName2),
                    resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel, "state", stateInactive),
                    resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel, "title", title2),
                    resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel, "department", department2),
                    resource.TestCheckResourceAttr(ResourceType+"."+userResourceLabel, "acd_auto_answer", "true"),
                    resource.TestCheckResourceAttrSet(ResourceType+"."+userResourceLabel, "id"),
                ),
            },
            {
                // Import state verification
                ResourceName:            ResourceType + "." + userResourceLabel,
                ImportState:             true,
                ImportStateVerify:       true,
                ImportStateVerifyIgnore: []string{"password"}, // Password not returned by API
            },
        },
        CheckDestroy: testVerifyUsersDestroyed,
    })
}
```

**Key Components**:

| Component | Purpose |
|-----------|---------|
| `t.Parallel()` | Enable parallel test execution |
| `PreCheck` | Verify test prerequisites (credentials, environment) |
| `ProtoV6ProviderFactories` | Provide muxed provider factory |
| `Steps` | Define test steps (create, update, import) |
| `Config` | Terraform HCL configuration for step |
| `Check` | Assertions to verify expected state |
| `ImportState` | Test import functionality |
| `ImportStateVerifyIgnore` | Attributes to ignore during import verification |
| `CheckDestroy` | Verify resources are cleaned up |


### 4. Test Case Types (Complex Resources)

#### 4.1 Basic CRUD Test

**Purpose**: Test basic create, read, import, and destroy operations with required and optional attributes.

**Pattern**:
```go
func TestAccFrameworkResource<ResourceName>Basic(t *testing.T) {
    // Test with required attributes
    // Test with optional attributes
    // Test updates to attributes
    // Verify create, read, import
    // Verify destroy
}
```

**What to Test**:
- Create resource with required attributes only
- Update to add optional attributes
- Update to modify attributes
- Test import by ID
- Verify resource is destroyed

#### 4.2 Nested Structure Tests

**Purpose**: Test resources with 1-level, 2-level, and 3-level nested blocks.

**Pattern**:
```go
func TestAccFrameworkResource<ResourceName><NestedBlock>(t *testing.T) {
    // Create with nested block
    // Update nested block attributes
    // Add/remove nested block elements
    // Verify nested attributes
}
```

**Example** (user addresses - 2-level nesting):
```go
func TestAccFrameworkResourceUserAddresses(t *testing.T) {
    // Test phone_numbers (nested in addresses)
    // Test other_emails (nested in addresses)
    // Test multiple phone numbers
    // Test extension-only phone numbers
    // Test E.164 format validation
}
```

**Example** (user routing_utilization - 3-level nesting):
```go
func TestAccFrameworkResourceUserRoutingUtilizationWithLabels(t *testing.T) {
    // Test routing_utilization block
    // Test call/callback/chat/email/message blocks (nested in routing_utilization)
    // Test label_utilizations (nested in call/callback/etc.)
    // Test interruptible_media_types (nested in call/callback/etc.)
}
```

**What to Test**:
- Create with nested blocks
- Update nested block attributes
- Add elements to nested arrays/sets
- Remove elements from nested arrays/sets
- Test all nesting levels
- Verify nested attribute values

#### 4.3 Multiple Dependency Tests

**Purpose**: Test resource with various SDKv2 and Framework dependencies.

**Pattern**:
```go
func TestAccFrameworkResource<ResourceName><Dependency>(t *testing.T) {
    // Create dependency resources
    // Create resource with dependency references
    // Verify dependency is correctly set
    // Test with multiple dependencies
}
```

**Example** (user with skills and languages):
```go
func TestAccFrameworkResourceUserSkillsAndLanguages(t *testing.T) {
    // Create routing_skill resources (Framework)
    // Create routing_language resources (Framework)
    // Create user with skill and language references
    // Verify references are correct
    // Test proficiency values
}
```

**Key Points**:
- Use `TestCheckResourceAttrPair` to verify dependency reference
- Include dependency resource HCL in config
- Ensure dependency is in muxed provider factory
- Test both SDKv2 and Framework dependencies

#### 4.4 Update Tests (Complex Resources)

**Purpose**: Test in-place updates of nested structures and arrays.

**Pattern**:
```go
func TestAccFrameworkResource<ResourceName>Updates(t *testing.T) {
    // Create resource with initial values
    // Update nested attributes
    // Update array elements
    // Verify updates are applied
}
```

**What to Test**:
- Update nested block attributes
- Add elements to arrays/sets
- Remove elements from arrays/sets
- Modify elements in arrays/sets
- Update dependency references
- Verify no resource replacement (in-place update)

#### 4.5 Edge Case Tests

**Purpose**: Test API asymmetries, concurrent modifications, and other edge cases.

**Pattern**:
```go
func TestAccFrameworkResource<ResourceName><EdgeCase>(t *testing.T) {
    // Test specific edge case
    // Document API behavior
    // Verify handling is correct
}
```

**Examples**:

**API Asymmetry** (user addresses deletion):
```go
// TODO (ADDRESSES-DELETION-ASYMMETRY): Documents known API behavior
// Test may be commented out until resolution
```

**Extension-Only Phone Number**:
```go
// Test phone number with extension but no number
// Test transition from extension-only to number with extension
// Test transition from number to extension-only
```

**Concurrent Modification**:
```go
func TestAccFrameworkResourceUserConcurrentModification(t *testing.T) {
    // Test concurrent updates to same resource
    // Verify retry logic for version conflicts
    // Verify state consistency after conflicts
}
```

**Deleted Resource Restoration**:
```go
func TestAccFrameworkResourceUserDeletedUserRestoration(t *testing.T) {
    // Create user
    // Delete user outside Terraform
    // Attempt to update (should fail gracefully)
    // Verify error handling
}
```

#### 4.6 Validation Tests

**Purpose**: Test complex constraint validation.

**Pattern**:
```go
func TestAccFrameworkResource<ResourceName>Validation(t *testing.T) {
    // Test invalid values
    // Verify validation errors
    // Test boundary conditions
}
```

**Example** (user skill proficiency validation):
```go
func TestAccFrameworkResourceUserSkillProficiencyValidation(t *testing.T) {
    // Test proficiency < 0.0 (invalid)
    // Test proficiency > 5.0 (invalid)
    // Test proficiency = 0.0 (valid)
    // Test proficiency = 5.0 (valid)
    // Verify error messages
}
```

**What to Test**:
- Proficiency ranges (0.0 to 5.0)
- Required field validation
- Format validation (email, phone, etc.)
- Constraint validation (max capacity, etc.)
- Error message clarity

#### 4.7 Lifecycle Test

**Purpose**: Comprehensive test covering multiple scenarios in one test.

**Pattern**:
```go
func TestAccFrameworkResource<ResourceName>Lifecycle(t *testing.T) {
    // Create without optional attributes
    // Add optional attributes
    // Update attributes
    // Add nested blocks
    // Update nested blocks
    // Remove nested blocks
    // Verify all operations
}
```

**What to Test**:
- Create with minimal attributes
- Add optional attributes
- Update various attributes
- Add nested blocks
- Update nested blocks
- Remove nested blocks
- Test with and without dependencies
- Comprehensive coverage in single test

---

### 5. Destroy Verification Function (Complex Resources)

**Purpose**: Verify that resources are properly cleaned up after test.

**Design Pattern**:
```go
// testVerifyFramework<ResourceName>Destroyed checks that all <resources> have been destroyed
func testVerifyFramework<ResourceName>Destroyed(state *terraform.State) error {
    api := platformclientv2.New<ResourceAPI>()
    for _, rs := range state.RootModule().Resources {
        if rs.Type != "genesyscloud_<resource>" {
            continue
        }

        resource, resp, err := api.Get<Resource>(rs.Primary.ID, ...)
        if resource != nil {
            return fmt.Errorf("Framework <resource> (%s) still exists", rs.Primary.ID)
        } else if util.IsStatus404(resp) {
            // Resource not found as expected
            continue
        } else {
            // Unexpected error
            return fmt.Errorf("Unexpected error checking Framework <resource>: %s", err)
        }
    }
    // Success. All resources destroyed
    return nil
}
```

**Example** (user):
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

**Key Points**:
- Check only resources of the specific type
- Verify resource returns 404 (not found)
- Return error if resource still exists
- Return nil if all resources destroyed
- Handle API errors gracefully


### 6. Helper Function Library (Complex Resources)

**Purpose**: Create extensive, modular helper function library for generating HCL configurations.

#### 6.1 Basic Resource Generation Helper

**Design Pattern**:
```go
// generateFramework<ResourceName>Resource generates a <resource> resource for Framework testing
func generateFramework<ResourceName>Resource(
    resourceLabel string,
    requiredAttr1 string,
    requiredAttr2 string,
    optionalAttr1 string,
    optionalAttr2 string,
) string {
    optionalAttr1Str := ""
    if optionalAttr1 != util.NullValue {
        optionalAttr1Str = fmt.Sprintf(`
        optional_attr1 = %s`, optionalAttr1)
    }

    optionalAttr2Str := ""
    if optionalAttr2 != "" {
        optionalAttr2Str = fmt.Sprintf(`
        optional_attr2 = "%s"`, optionalAttr2)
    }

    return fmt.Sprintf(`resource "genesyscloud_<resource>" "%s" {
        required_attr1 = "%s"
        required_attr2 = "%s"%s%s
    }
    `, resourceLabel, requiredAttr1, requiredAttr2, optionalAttr1Str, optionalAttr2Str)
}
```

**Example** (user):
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
    stateAttr := ""
    if state != util.NullValue {
        stateAttr = fmt.Sprintf(`
        state = %s`, state)
    }

    titleAttr := ""
    if title != util.NullValue {
        titleAttr = fmt.Sprintf(`
        title = %s`, title)
    }

    departmentAttr := ""
    if department != util.NullValue {
        departmentAttr = fmt.Sprintf(`
        department = %s`, department)
    }

    managerAttr := ""
    if manager != util.NullValue {
        managerAttr = fmt.Sprintf(`
        manager = %s`, manager)
    }

    acdAutoAnswerAttr := ""
    if acdAutoAnswer != util.NullValue {
        acdAutoAnswerAttr = fmt.Sprintf(`
        acd_auto_answer = %s`, acdAutoAnswer)
    }

    return fmt.Sprintf(`resource "%s" "%s" {
        email = "%s"
        name = "%s"%s%s%s%s%s
    }
    `, ResourceType, resourceLabel, email, name, stateAttr, titleAttr, departmentAttr, managerAttr, acdAutoAnswerAttr)
}
```

**Key Points**:
- Handle all required attributes
- Handle all optional attributes with null checks
- Use `util.NullValue` for omitted reference attributes
- Use empty string check for omitted literal attributes
- Use `strconv.Quote()` for string values in HCL

#### 6.2 Custom Attributes Helper (Variadic Pattern)

**Design Pattern**:
```go
// generateFramework<ResourceName>WithCustomAttrs generates a <resource> with custom attributes
// This allows flexible test configuration by accepting variadic attributes
func generateFramework<ResourceName>WithCustomAttrs(resourceLabel, requiredAttr1, requiredAttr2 string, attrs ...string) string {
    return fmt.Sprintf(`resource "genesyscloud_<resource>" "%s" {
        required_attr1 = "%s"
        required_attr2 = "%s"
        %s
    }`, resourceLabel, requiredAttr1, requiredAttr2, strings.Join(attrs, "\n"))
}
```

**Example** (user):
```go
func generateFrameworkUserWithCustomAttrs(resourceLabel, email, name string, attrs ...string) string {
    return fmt.Sprintf(`resource "%s" "%s" {
        email = "%s"
        name = "%s"
        %s
    }`, ResourceType, resourceLabel, email, name, strings.Join(attrs, "\n"))
}
```

**Usage**:
```go
generateFrameworkUserWithCustomAttrs(
    userResourceLabel, email, name,
    generateFrameworkUserAddresses(...),
    generateFrameworkRoutingUtilization(...),
    `title = "Senior Developer"`,
    `department = "Engineering"`,
)
```

**Key Points**:
- Accepts variadic attributes for flexibility
- Enables modular composition of nested blocks
- Reduces duplication in test code
- Allows mixing generated blocks with inline attributes

#### 6.3 Nested Block Helpers (1-Level)

**Design Pattern**:
```go
// generate<NestedBlock> generates a <nested_block> block for Framework testing
func generate<NestedBlock>(nestedContent ...string) string {
    if len(nestedContent) == 0 {
        return ""
    }
    return fmt.Sprintf(`<nested_block> {
        %s
    }`, strings.Join(nestedContent, "\n"))
}
```

**Example** (user addresses):
```go
func generateFrameworkUserAddresses(nestedBlocks ...string) string {
    var phoneBlocks []string
    for _, block := range nestedBlocks {
        phoneBlocks = append(phoneBlocks, fmt.Sprintf("phone_numbers {\n%s\n}", block))
    }
    return fmt.Sprintf("addresses {\n%s\n}", strings.Join(phoneBlocks, "\n"))
}
```

**Usage**:
```go
generateFrameworkUserAddresses(
    generateFrameworkUserPhoneAddress(phone1, mediaType1, type1, extension1),
    generateFrameworkUserPhoneAddress(phone2, mediaType2, type2, extension2),
)
```

#### 6.4 Sub-Nested Block Helpers (2-Level and 3-Level)

**Design Pattern** (2-level):
```go
// generate<SubNestedBlock> generates a <sub_nested_block> block content (not wrapped)
func generate<SubNestedBlock>(attr1, attr2, attr3 string, extras ...string) string {
    return fmt.Sprintf(`attr1 = %s
        attr2 = %s
        attr3 = %s
        %s`, attr1, attr2, attr3, strings.Join(extras, "\n"))
}
```

**Example** (user phone address - 2-level):
```go
func generateFrameworkUserPhoneAddress(phoneNum, phoneMediaType, phoneType, extension string, extras ...string) string {
    return fmt.Sprintf(`number = %s
        media_type = %s
        type = %s
        extension = %s
        %s`, phoneNum, phoneMediaType, phoneType, extension, strings.Join(extras, "\n"))
}
```

**Usage**:
```go
generateFrameworkUserPhoneAddress(
    strconv.Quote(phone),
    strconv.Quote("PHONE"),
    strconv.Quote("WORK"),
    util.NullValue,
    fmt.Sprintf("extension_pool_id = genesyscloud_telephony_providers_edges_extension_pool.%s.id", poolLabel),
)
```

**Design Pattern** (3-level):
```go
// generate<Level2Block> generates a <level2_block> block with nested <level3_block> blocks
func generate<Level2Block>(attr1, attr2 string, level3Blocks ...string) string {
    level3Content := ""
    if len(level3Blocks) > 0 {
        level3Content = strings.Join(level3Blocks, "\n")
    }
    return fmt.Sprintf(`<level2_block> {
        attr1 = %s
        attr2 = %s
        %s
    }`, attr1, attr2, level3Content)
}
```

**Example** (user routing_utilization call block - 3-level):
```go
func generateFrameworkRoutingUtilizationCall(maxCapacity, includeNonAcd string) string {
    return fmt.Sprintf(`call {
        maximum_capacity = %s
        include_non_acd = %s
    }`, maxCapacity, includeNonAcd)
}
```

#### 6.5 Attribute-Specific Helpers

**Design Pattern**:
```go
// generate<Attribute> generates a specific attribute configuration
func generate<Attribute>(values ...string) string {
    if len(values) == 0 {
        return ""
    }
    var quotedValues []string
    for _, v := range values {
        quotedValues = append(quotedValues, strconv.Quote(v))
    }
    return fmt.Sprintf(`<attribute> = [%s]`, strings.Join(quotedValues, ", "))
}
```

**Example** (user profile_skills):
```go
func generateProfileSkills(skills ...string) string {
    if len(skills) == 0 {
        return ""
    }
    var skillStrings []string
    for _, skill := range skills {
        skillStrings = append(skillStrings, strconv.Quote(skill))
    }
    return fmt.Sprintf(`profile_skills = [%s]`, strings.Join(skillStrings, ", "))
}
```

**Usage**:
```go
generateProfileSkills("Java", "Go", "Python")
// Output: profile_skills = ["Java", "Go", "Python"]
```

#### 6.6 Helper Function Composition

**Pattern**: Compose helpers to build complex configurations

**Example** (user with multiple nested blocks):
```go
generateFrameworkUserWithCustomAttrs(
    userResourceLabel, email, name,
    // Addresses block (2-level nesting)
    generateFrameworkUserAddresses(
        generateFrameworkUserPhoneAddress(phone, mediaType, phoneType, extension),
        generateFrameworkUserEmailAddress(otherEmail, emailType),
    ),
    // Routing utilization block (3-level nesting)
    generateFrameworkRoutingUtilization(
        generateFrameworkRoutingUtilizationCall(maxCapacity, includeNonAcd),
        generateFrameworkRoutingUtilizationCallback(maxCapacity, includeNonAcd),
    ),
    // Profile skills (array attribute)
    generateProfileSkills("Java", "Go"),
    // Certifications (array attribute)
    generateCertifications("AWS Developer", "AWS Architect"),
)
```

**Key Points**:
- Helpers are modular and composable
- Each helper handles one level of nesting
- Variadic patterns enable flexible composition
- Extras parameter allows adding custom attributes


---

## Part 2: Data Source Test File

### 1. Data Source Test Function (Complex Resources)

**Purpose**: Test data source lookup functionality with multiple search criteria.

**Design Pattern**:
```go
func TestAccFrameworkDataSource<ResourceName>(t *testing.T) {
    t.Parallel()
    var (
        resourceLabel   = "test_<resource>_resource"
        dataSourceLabel = "test_<resource>_data_source"
        // Define test variables
        identifier1     = "value1"
        identifier2     = "value2"
    )

    resource.Test(t, resource.TestCase{
        PreCheck:                 func() { util.TestAccPreCheck(t) },
        ProtoV6ProviderFactories: getFrameworkProviderFactories(),
        Steps: []resource.TestStep{
            {
                // Search by primary identifier
                Config: generateFramework<ResourceName>Resource(resourceLabel, ...) +
                    generateFramework<ResourceName>DataSource(dataSourceLabel, identifier1, util.NullValue, resourceLabel),
                Check: resource.ComposeTestCheckFunc(
                    resource.TestCheckResourceAttrPair("data.genesyscloud_<resource>."+dataSourceLabel, "id", "genesyscloud_<resource>."+resourceLabel, "id"),
                    resource.TestCheckResourceAttrPair("data.genesyscloud_<resource>."+dataSourceLabel, "name", "genesyscloud_<resource>."+resourceLabel, "name"),
                ),
            },
            {
                // Search by alternative identifier
                Config: generateFramework<ResourceName>Resource(resourceLabel, ...) +
                    generateFramework<ResourceName>DataSource(dataSourceLabel, util.NullValue, identifier2, resourceLabel),
                Check: resource.ComposeTestCheckFunc(
                    resource.TestCheckResourceAttrPair("data.genesyscloud_<resource>."+dataSourceLabel, "id", "genesyscloud_<resource>."+resourceLabel, "id"),
                    resource.TestCheckResourceAttrPair("data.genesyscloud_<resource>."+dataSourceLabel, "name", "genesyscloud_<resource>."+resourceLabel, "name"),
                ),
            },
        },
    })
}
```

**Example** (user data source):
```go
func TestAccFrameworkDataSourceUser(t *testing.T) {
    t.Parallel()
    var (
        userResourceLabel   = "test-user-resource"
        userDataSourceLabel = "test-user-data-source"
        randomString        = uuid.NewString()
        userEmail           = "framework_user_" + randomString + "@example.com"
        userName            = "Framework_User_" + randomString
        userID              string
    )

    resource.Test(t, resource.TestCase{
        PreCheck:                 func() { util.TestAccPreCheck(t) },
        ProtoV6ProviderFactories: getFrameworkProviderFactories(),
        Steps: []resource.TestStep{
            {
                // Search by email
                Config: GenerateBasicUserResource(
                    userResourceLabel,
                    userEmail,
                    userName,
                ) + generateUserDataSource(
                    userDataSourceLabel,
                    ResourceType+"."+userResourceLabel+".email",
                    util.NullValue,
                    ResourceType+"."+userResourceLabel,
                ),
                Check: resource.ComposeTestCheckFunc(
                    resource.TestCheckResourceAttrPair("data."+ResourceType+"."+userDataSourceLabel, "id", ResourceType+"."+userResourceLabel, "id"),
                    resource.TestCheckResourceAttrPair("data."+ResourceType+"."+userDataSourceLabel, "name", ResourceType+"."+userResourceLabel, "name"),
                    func(s *terraform.State) error {
                        rs, ok := s.RootModule().Resources[ResourceType+"."+userResourceLabel]
                        if !ok {
                            return fmt.Errorf("not found: %s", ResourceType+"."+userResourceLabel)
                        }
                        if rs.Primary.ID == "" {
                            return fmt.Errorf("user ID is empty")
                        }
                        userID = rs.Primary.ID
                        log.Printf("User ID: %s\n", userID)
                        return nil
                    },
                ),
            },
            {
                // Search by name
                Config: GenerateBasicUserResource(
                    userResourceLabel,
                    userEmail,
                    userName,
                ) + generateUserDataSource(
                    userDataSourceLabel,
                    util.NullValue,
                    ResourceType+"."+userResourceLabel+".name",
                    ResourceType+"."+userResourceLabel,
                ),
                Check: resource.ComposeTestCheckFunc(
                    resource.TestCheckResourceAttrPair("data."+ResourceType+"."+userDataSourceLabel, "id", ResourceType+"."+userResourceLabel, "id"),
                    resource.TestCheckResourceAttrPair("data."+ResourceType+"."+userDataSourceLabel, "name", ResourceType+"."+userResourceLabel, "name"),
                    func(s *terraform.State) error {
                        time.Sleep(30 * time.Second) // Wait for proper cleanup
                        return nil
                    },
                ),
            },
        },
        CheckDestroy: func(state *terraform.State) error {
            time.Sleep(45 * time.Second)
            return testVerifyUsersDestroyed(state)
        },
    })
}
```

**Key Points**:
- Create resource first, then look it up with data source
- Test multiple search criteria (email, name, etc.)
- Use `TestCheckResourceAttrPair` to verify IDs match
- Include `depends_on` in data source to ensure resource exists
- Handle eventual consistency with sleep if needed

### 2. Data Source HCL Generation Helper

**Design Pattern**:
```go
// generateFramework<ResourceName>DataSource generates a <resource> data source for Framework testing
func generateFramework<ResourceName>DataSource(
    dataLabel string,
    identifier1 string,
    identifier2 string,
    dependsOnResource string,
) string {
    identifier1Attr := ""
    if identifier1 != util.NullValue {
        identifier1Attr = fmt.Sprintf(`
        identifier1 = %s`, identifier1)
    }

    identifier2Attr := ""
    if identifier2 != util.NullValue {
        identifier2Attr = fmt.Sprintf(`
        identifier2 = %s`, identifier2)
    }

    return fmt.Sprintf(`data "genesyscloud_<resource>" "%s" {%s%s
        depends_on = [%s]
    }
    `, dataLabel, identifier1Attr, identifier2Attr, dependsOnResource)
}
```

**Example** (user data source):
```go
func generateUserDataSource(
    resourceLabel string,
    email string,
    name string,
    dependsOnResource string,
) string {
    emailAttr := ""
    if email != util.NullValue {
        emailAttr = fmt.Sprintf(`
        email = %s`, email)
    }

    nameAttr := ""
    if name != util.NullValue {
        nameAttr = fmt.Sprintf(`
        name = %s`, name)
    }

    return fmt.Sprintf(`data "%s" "%s" {%s%s
        depends_on = [%s]
    }
    `, ResourceType, resourceLabel, emailAttr, nameAttr, dependsOnResource)
}
```

**Key Points**:
- Support multiple search criteria
- Include `depends_on` to ensure resource exists before lookup
- Handle optional search criteria with null checks
- Use exact values for lookup

---

## Part 3: Test Initialization File

### Purpose

The test initialization file provides package-level setup for tests with multiple dependencies.

**Design Pattern**:
```go
package <resource_name>

import (
    "sync"
    "testing"

    "github.com/hashicorp/terraform-plugin-framework/datasource"
    "github.com/hashicorp/terraform-plugin-framework/resource"
    "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
    // Import all dependencies
)

// Package-level variables
var providerDataSources map[string]*schema.Resource
var providerResources map[string]*schema.Resource
var frameworkResources map[string]func() resource.Resource
var frameworkDataSources map[string]func() datasource.DataSource

type registerTestInstance struct {
    resourceMapMutex            sync.RWMutex
    datasourceMapMutex          sync.RWMutex
    frameworkResourceMapMutex   sync.RWMutex
    frameworkDataSourceMapMutex sync.RWMutex
}

// registerTestResources registers all SDKv2 resources used in the tests
func (r *registerTestInstance) registerTestResources() {
    r.resourceMapMutex.Lock()
    defer r.resourceMapMutex.Unlock()

    // Register SDKv2 resources needed for Framework tests
    providerResources[dependency1.ResourceType] = dependency1.ResourceDependency1()
    providerResources[dependency2.ResourceType] = dependency2.ResourceDependency2()
    // ... register all SDKv2 dependencies
}

// registerTestDataSources registers all SDKv2 data sources used in the tests
func (r *registerTestInstance) registerTestDataSources() {
    r.datasourceMapMutex.Lock()
    defer r.datasourceMapMutex.Unlock()

    // Register SDKv2 data sources needed for Framework tests
    providerDataSources[dependency1.ResourceType] = dependency1.DataSourceDependency1()
    // ... register all SDKv2 data source dependencies
}

// registerFrameworkTestResources registers all Framework resources used in the tests
func (r *registerTestInstance) registerFrameworkTestResources() {
    r.frameworkResourceMapMutex.Lock()
    defer r.frameworkResourceMapMutex.Unlock()

    frameworkResources[ResourceType] = New<ResourceName>FrameworkResource
    frameworkResources[frameworkDep1.ResourceType] = frameworkDep1.NewFrameworkDep1Resource
    // ... register all Framework resource dependencies
}

// registerFrameworkTestDataSources registers all Framework data sources used in the tests
func (r *registerTestInstance) registerFrameworkTestDataSources() {
    r.frameworkDataSourceMapMutex.Lock()
    defer r.frameworkDataSourceMapMutex.Unlock()

    frameworkDataSources[ResourceType] = New<ResourceName>FrameworkDataSource
    frameworkDataSources[frameworkDep1.ResourceType] = frameworkDep1.NewFrameworkDep1DataSource
    // ... register all Framework data source dependencies
}

// initTestResources initializes all test resources and data sources
func initTestResources() {
    providerResources = make(map[string]*schema.Resource)
    providerDataSources = make(map[string]*schema.Resource)
    frameworkResources = make(map[string]func() resource.Resource)
    frameworkDataSources = make(map[string]func() datasource.DataSource)

    regInstance := &registerTestInstance{}

    regInstance.registerTestResources()
    regInstance.registerTestDataSources()
    regInstance.registerFrameworkTestResources()
    regInstance.registerFrameworkTestDataSources()
}

// TestMain is called by the testing framework when running the test suite
func TestMain(m *testing.M) {
    initTestResources()
    m.Run()
}
```

**Example** (user init test):
```go
package user

import (
    "sync"
    "testing"

    "github.com/hashicorp/terraform-plugin-framework/datasource"
    "github.com/hashicorp/terraform-plugin-framework/resource"
    "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
    "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud"
    authDivision "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/auth_division"
    authRole "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/auth_role"
    "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/location"
    routinglanguage "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/routing_language"
    routingSkill "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/routing_skill"
    routingUtilizationLabel "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/routing_utilization_label"
    extensionPool "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/telephony_providers_edges_extension_pool"
)

var providerDataSources map[string]*schema.Resource
var providerResources map[string]*schema.Resource
var frameworkResources map[string]func() resource.Resource
var frameworkDataSources map[string]func() datasource.DataSource

type registerTestInstance struct {
    resourceMapMutex            sync.RWMutex
    datasourceMapMutex          sync.RWMutex
    frameworkResourceMapMutex   sync.RWMutex
    frameworkDataSourceMapMutex sync.RWMutex
}

func (r *registerTestInstance) registerTestResources() {
    r.resourceMapMutex.Lock()
    defer r.resourceMapMutex.Unlock()

    providerResources[authRole.ResourceType] = authRole.ResourceAuthRole()
    providerResources[authDivision.ResourceType] = authDivision.ResourceAuthDivision()
    providerResources[location.ResourceType] = location.ResourceLocation()
    providerResources[routingSkill.ResourceType] = routingSkill.ResourceRoutingSkill()
    providerResources[routingUtilizationLabel.ResourceType] = routingUtilizationLabel.ResourceRoutingUtilizationLabel()
    providerResources[extensionPool.ResourceType] = extensionPool.ResourceTelephonyExtensionPool()
}

func (r *registerTestInstance) registerTestDataSources() {
    r.datasourceMapMutex.Lock()
    defer r.datasourceMapMutex.Unlock()

    providerDataSources[authRole.ResourceType] = authRole.DataSourceAuthRole()
    providerDataSources["genesyscloud_auth_division_home"] = genesyscloud.DataSourceAuthDivisionHome()
    providerDataSources[location.ResourceType] = location.DataSourceLocation()
    providerDataSources[routingSkill.ResourceType] = routingSkill.DataSourceRoutingSkill()
    providerDataSources[routingUtilizationLabel.ResourceType] = routingUtilizationLabel.DataSourceRoutingUtilizationLabel()
}

func (r *registerTestInstance) registerFrameworkTestResources() {
    r.frameworkResourceMapMutex.Lock()
    defer r.frameworkResourceMapMutex.Unlock()

    frameworkResources[ResourceType] = NewUserFrameworkResource
    frameworkResources[routinglanguage.ResourceType] = routinglanguage.NewFrameworkRoutingLanguageResource
}

func (r *registerTestInstance) registerFrameworkTestDataSources() {
    r.frameworkDataSourceMapMutex.Lock()
    defer r.frameworkDataSourceMapMutex.Unlock()

    frameworkDataSources[ResourceType] = NewUserFrameworkDataSource
    frameworkDataSources[routinglanguage.ResourceType] = routinglanguage.NewFrameworkRoutingLanguageDataSource
}

func initTestResources() {
    providerResources = make(map[string]*schema.Resource)
    providerDataSources = make(map[string]*schema.Resource)
    frameworkResources = make(map[string]func() resource.Resource)
    frameworkDataSources = make(map[string]func() datasource.DataSource)

    regInstance := &registerTestInstance{}

    regInstance.registerTestResources()
    regInstance.registerTestDataSources()
    regInstance.registerFrameworkTestResources()
    regInstance.registerFrameworkTestDataSources()
}

func TestMain(m *testing.M) {
    initTestResources()
    m.Run()
}
```

**Key Points**:
- Register all SDKv2 dependencies
- Register all Framework dependencies
- Use mutexes for thread-safe registration
- Initialize in TestMain
- Also call from init() in test files

---

## Design Patterns and Best Practices

### Pattern 1: Parallel Test Execution

**Pattern**:
```go
func TestAccFrameworkResource<ResourceName><Scenario>(t *testing.T) {
    t.Parallel()
    // ... test code
}
```

**Why**:
- Enables parallel test execution
- Reduces total test time
- Requires unique resource names (use uuid.NewString())
- Safe when tests are isolated


### Pattern 2: Unique Test Names with UUID

**Pattern**:
```go
email := "terraform-framework-" + uuid.NewString() + "@user.com"
name := "Terraform Framework User " + uuid.NewString()
```

**Why**:
- Prevents conflicts between parallel tests
- Enables test parallelization
- Avoids cleanup issues
- Unique identifier for debugging

### Pattern 3: Import Verification with Ignore List

**Pattern**:
```go
{
    ResourceName:            "genesyscloud_<resource>." + resourceLabel,
    ImportState:             true,
    ImportStateVerify:       true,
    ImportStateVerifyIgnore: []string{"password", "sensitive_attr"}, // Attributes not returned by API
}
```

**Why**:
- Verifies import functionality works
- Ensures state is correctly populated after import
- Ignores attributes not returned by API (passwords, computed values)
- Tests complete resource lifecycle

### Pattern 4: Nested Attribute Testing

**1-Level Nesting**:
```go
resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "nested_block.#", "1")
resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "nested_block.0.attribute", value)
```

**2-Level Nesting**:
```go
resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "level1.0.level2.#", "1")
resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "level1.0.level2.0.attribute", value)
```

**3-Level Nesting**:
```go
resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "level1.0.level2.0.level3.#", "1")
resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "level1.0.level2.0.level3.0.attribute", value)
```

### Pattern 5: Set Element Testing (Order-Independent)

**Pattern**:
```go
resource.TestCheckTypeSetElemNestedAttrs(
    ResourceType+"."+resourceLabel,
    "nested_set.*",
    map[string]string{
        "attribute1": value1,
        "attribute2": value2,
    },
)
```

**Why**:
- Tests set elements without depending on order
- Verifies specific element exists in set
- Handles Terraform's set ordering

### Pattern 6: Dependency Testing

**SDKv2 Dependency**:
```go
Config: dependency.GenerateDependencyResource(depLabel, depName) +
    generateFrameworkResourceWithDependency(resourceLabel, name, "genesyscloud_dependency."+depLabel+".id"),
Check: resource.ComposeTestCheckFunc(
    resource.TestCheckResourceAttrPair(
        ResourceType+"."+resourceLabel,
        "dependency_id",
        "genesyscloud_dependency."+depLabel,
        "id",
    ),
),
```

**Framework Dependency**:
```go
Config: frameworkDep.GenerateFrameworkDependencyResource(depLabel, depName) +
    generateFrameworkResourceWithDependency(resourceLabel, name, "genesyscloud_framework_dep."+depLabel+".id"),
Check: resource.ComposeTestCheckFunc(
    resource.TestCheckResourceAttrPair(
        ResourceType+"."+resourceLabel,
        "dependency_id",
        "genesyscloud_framework_dep."+depLabel,
        "id",
    ),
),
```

### Pattern 7: Edge Case Documentation

**Pattern**:
```go
// TODO (<ISSUE-ID>): Detailed explanation of the edge case
// - What the expected behavior is
// - What the actual behavior is
// - Why it worked in SDKv2 but fails in Framework
// - Potential resolution approaches
// - Current status
/*{
    // Commented-out test step
    Config: ...,
    Check: ...,
},*/
```

**Why**:
- Documents known issues
- Provides context for future developers
- Tracks resolution approaches
- Enables informed decision-making

### Pattern 8: Validation Testing

**Pattern**:
```go
{
    Config: generateFrameworkResourceWithInvalidValue(...),
    ExpectError: regexp.MustCompile(`expected <attribute> to be in the range \(min - max\)`),
}
```

**Why**:
- Tests validation logic
- Verifies error messages
- Documents constraints
- Ensures user-friendly errors

### Pattern 9: Pre-Test Cleanup

**Pattern**:
```go
func TestAccFrameworkResource<ResourceName><Scenario>(t *testing.T) {
    t.Parallel()
    
    // Pre-test cleanup (matching SDKv2 pattern)
    t.Logf("Attempting to cleanup resource with identifier %s", identifier)
    err := cleanup.DeleteResourceWithIdentifier(identifier)
    if err != nil {
        t.Log(err)
    }
    
    resource.Test(t, resource.TestCase{
        // ... test code
    })
}
```

**Why**:
- Cleans up resources from previous failed tests
- Prevents test failures due to existing resources
- Matches SDKv2 cleanup pattern
- Logs cleanup attempts

### Pattern 10: Eventual Consistency Handling

**Pattern**:
```go
Check: resource.ComposeTestCheckFunc(
    // ... assertions
    func(s *terraform.State) error {
        time.Sleep(30 * time.Second) // Wait for eventual consistency
        return nil
    },
),
```

**Why**:
- Handles API eventual consistency
- Prevents flaky tests
- Allows time for resource propagation
- Common in distributed systems

---

## Migration Considerations

### Test Behavior Preservation

When migrating tests, verify:
- [ ] Test assertions are identical to SDKv2 version
- [ ] Test steps match SDKv2 version
- [ ] Test coverage is maintained or improved
- [ ] Test execution time is similar
- [ ] Test reliability is maintained
- [ ] Edge cases are documented

### Common Migration Pitfalls (Complex Resources)

#### Pitfall 1: Missing Dependency in Muxed Provider
**Problem**: Test fails because dependency resource not included in factory.
**Solution**: Add all SDKv2 and Framework dependencies to muxed provider factory.

#### Pitfall 2: Wrong Provider Factory Type
**Problem**: Using `ProviderFactories` instead of `ProtoV6ProviderFactories`.
**Solution**: Always use `ProtoV6ProviderFactories` or `GetMuxedProviderFactories()` for Framework tests.

#### Pitfall 3: Test Name Conflicts
**Problem**: Framework test has same name as SDKv2 test.
**Solution**: Use `TestAccFramework*` prefix for Framework tests.

#### Pitfall 4: Missing Import Test
**Problem**: Test doesn't verify import functionality.
**Solution**: Always include import step in tests.

#### Pitfall 5: Incorrect Destroy Verification
**Problem**: Destroy check doesn't filter by resource type.
**Solution**: Check `rs.Type` before verifying destruction.

#### Pitfall 6: Nested Attribute Path Errors
**Problem**: Incorrect attribute path for nested structures.
**Solution**: Use correct indexing: `nested.0.sub_nested.0.attribute`.

#### Pitfall 7: Set Ordering Assumptions
**Problem**: Test assumes specific order for set elements.
**Solution**: Use `TestCheckTypeSetElemNestedAttrs` for order-independent checks.

#### Pitfall 8: Missing Init Function
**Problem**: Tests fail with nil pointer errors.
**Solution**: Add init function to ensure test resources are initialized.

#### Pitfall 9: API Asymmetry Not Documented
**Problem**: Test fails due to API behavior difference, no documentation.
**Solution**: Add detailed TODO comment explaining the issue.

#### Pitfall 10: Insufficient Helper Modularity
**Problem**: Helper functions are too rigid, can't test specific scenarios.
**Solution**: Use variadic patterns and modular composition.

---

## Summary

### Key Design Decisions

1. **Muxed Provider Pattern**: Support both Framework and SDKv2 with multiple dependencies
2. **Modular Helper Library**: Extensive, composable helper functions for nested structures
3. **Test Naming Convention**: Clear distinction with `TestAccFramework*` prefix
4. **Test Coverage Enhancement**: Maintain or improve coverage with focus on complex scenarios
5. **Edge Case Documentation**: Detailed TODO comments for known API asymmetries
6. **Test Initialization Pattern**: Package-level setup with thread-safe registration

### Test File Structure (Complex Resources)

```
Resource Test File:
├── Init function (ensures test resources initialized)
├── Provider factory function (muxed provider with dependencies)
├── Test functions (15+ tests covering all scenarios)
├── Destroy verification function
└── Helper function library (20+ helpers for nested structures)

Data Source Test File:
├── Init function (ensures test resources initialized)
├── Test functions (multiple search criteria)
└── HCL generation helper

Test Initialization File:
├── Package-level variables (SDKv2 and Framework)
├── Registration instance (with mutexes)
├── Registration methods (SDKv2 and Framework)
├── Initialization function
└── TestMain function
```

### Helper Function Patterns

1. **Basic Resource Generation**: Fixed parameters for common scenarios
2. **Custom Attributes**: Variadic pattern for flexible composition
3. **Nested Block Helpers**: One per nested block type
4. **Sub-Nested Block Helpers**: For 2-level and 3-level nesting
5. **Attribute-Specific Helpers**: For complex attributes (arrays, etc.)
6. **Composition**: Combine helpers to build complex configurations

### Next Steps

After completing Stage 3 test migration:
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
- **Simple Resource Reference**: `prompts/pf_simple_resource_migration/Stage3/design.md`
- **Plugin Framework Testing**: https://developer.hashicorp.com/terraform/plugin/framework/acctests
- **Terraform Testing**: https://developer.hashicorp.com/terraform/plugin/testing
- **Muxed Providers**: https://developer.hashicorp.com/terraform/plugin/mux

