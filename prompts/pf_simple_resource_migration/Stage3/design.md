# Stage 3 – Test Migration Design

## Overview

This document describes the design decisions, architecture, and patterns used in Stage 3 of the Plugin Framework migration. Stage 3 focuses on migrating acceptance tests to use Plugin Framework patterns while maintaining comprehensive test coverage and enabling mixed SDKv2/Framework testing through muxed providers.

**Reference Implementation**: 
- `genesyscloud/routing_wrapupcode/resource_genesyscloud_routing_wrapupcode_test.go`
- `genesyscloud/routing_wrapupcode/data_source_genesyscloud_routing_wrapupcode_test.go`
- `genesyscloud/routing_wrapupcode/genesyscloud_routing_wrapupcode_init_test.go`

---

## Design Principles

### 1. Muxed Provider Pattern
**Principle**: Use muxed provider to support both Framework and SDKv2 resources in the same test.

**Rationale**:
- Not all resources are migrated to Framework yet
- Framework resources often depend on SDKv2 resources (e.g., auth_division)
- Enables gradual migration without breaking existing tests
- Allows testing Framework resources with real dependencies

**Implementation**:
- Create custom provider factory for each test file
- Include Framework resource under test
- Include SDKv2 dependencies (e.g., auth_division)
- Return ProtoV6ProviderServer

### 2. Test Naming Convention
**Principle**: Use clear naming convention to distinguish Framework tests from SDKv2 tests.

**Rationale**:
- Easy to identify which tests are migrated
- Prevents naming conflicts during migration
- Clear intent in test reports
- Helps track migration progress

**Implementation**:
- Resource tests: `TestAccFrameworkResource<ResourceName>*`
- Data source tests: `TestAccFrameworkDataSource<ResourceName>*`
- Example: `TestAccFrameworkResourceRoutingWrapupcodeBasic`

### 3. Test Coverage Preservation
**Principle**: Maintain or improve test coverage from SDKv2 version.

**Rationale**:
- Verify migration didn't break functionality
- Maintain confidence in implementation
- Catch regressions early
- Document expected behavior

**Implementation**:
- Migrate all SDKv2 test cases
- Add additional test cases if needed
- Verify all CRUD operations
- Test edge cases and error conditions

### 4. Test Helper Functions
**Principle**: Create reusable helper functions to reduce duplication and improve maintainability.

**Rationale**:
- DRY (Don't Repeat Yourself) principle
- Easier to update test patterns
- Consistent test structure
- Improved readability

**Implementation**:
- HCL generation helpers
- Destroy verification helpers
- Provider factory helpers
- Assertion helpers

---

## Architecture

### File Structure

```
genesyscloud/<resource_name>/
├── resource_genesyscloud_<resource_name>_schema.go          ← Stage 1
├── resource_genesyscloud_<resource_name>.go                 ← Stage 2
├── data_source_genesyscloud_<resource_name>.go              ← Stage 2
├── resource_genesyscloud_<resource_name>_test.go            ← Stage 3 (THIS)
├── data_source_genesyscloud_<resource_name>_test.go         ← Stage 3 (THIS)
├── genesyscloud_<resource_name>_init_test.go                ← Stage 3 (THIS)
├── resource_genesyscloud_<resource_name>_export_utils.go    ← Stage 4
└── genesyscloud_<resource_name>_proxy.go                    ← NOT MODIFIED
```

### Resource Test File Components

```
┌─────────────────────────────────────────────────────────┐
│  resource_genesyscloud_<resource_name>_test.go          │
├─────────────────────────────────────────────────────────┤
│  1. Test Functions                                      │
│     - TestAccFrameworkResource<ResourceName>Basic       │
│     - TestAccFrameworkResource<ResourceName>Division    │
│     - TestAccFrameworkResource<ResourceName>Update      │
│     - TestAccFrameworkResource<ResourceName>Lifecycle   │
├─────────────────────────────────────────────────────────┤
│  2. Destroy Verification Function                       │
│     - testVerifyFramework<ResourceName>Destroyed()      │
├─────────────────────────────────────────────────────────┤
│  3. HCL Generation Helper                               │
│     - generateFramework<ResourceName>Resource()         │
├─────────────────────────────────────────────────────────┤
│  4. Provider Factory                                    │
│     - getFrameworkProviderFactories()                   │
└─────────────────────────────────────────────────────────┘
```

### Data Source Test File Components

```
┌─────────────────────────────────────────────────────────┐
│  data_source_genesyscloud_<resource_name>_test.go       │
├─────────────────────────────────────────────────────────┤
│  1. Test Functions                                      │
│     - TestAccFrameworkDataSource<ResourceName>          │
│     - TestAccFrameworkDataSource<ResourceName>Division  │
├─────────────────────────────────────────────────────────┤
│  2. HCL Generation Helper                               │
│     - generateFramework<ResourceName>DataSource()       │
└─────────────────────────────────────────────────────────┘
```

---

## Component Design

## Part 1: Resource Test File

### 1. Test Function Structure

**Purpose**: Define acceptance test cases for resource CRUD operations.

**Design Pattern**:
```go
func TestAccFrameworkResource<ResourceName><TestCase>(t *testing.T) {
    var (
        resourceLabel = "test_<resource>"
        name          = "Terraform Framework <Resource> " + uuid.NewString()
        // ... other test variables
    )

    resource.Test(t, resource.TestCase{
        PreCheck:                 func() { util.TestAccPreCheck(t) },
        ProtoV6ProviderFactories: getFrameworkProviderFactories(),
        Steps: []resource.TestStep{
            {
                // Create
                Config: generateFramework<ResourceName>Resource(...),
                Check: resource.ComposeTestCheckFunc(
                    resource.TestCheckResourceAttr("genesyscloud_<resource>."+resourceLabel, "name", name),
                    resource.TestCheckResourceAttrSet("genesyscloud_<resource>."+resourceLabel, "id"),
                ),
            },
            {
                // Import/Read
                ResourceName:      "genesyscloud_<resource>." + resourceLabel,
                ImportState:       true,
                ImportStateVerify: true,
            },
        },
        CheckDestroy: testVerifyFramework<ResourceName>Destroyed,
    })
}
```

**Example** (routing_wrapupcode basic test):
```go
func TestAccFrameworkResourceRoutingWrapupcodeBasic(t *testing.T) {
    var (
        resourceLabel = "test_routing_wrapupcode"
        name          = "Terraform Framework Wrapupcode " + uuid.NewString()
        description   = "Test wrapupcode description"
    )

    resource.Test(t, resource.TestCase{
        PreCheck:                 func() { util.TestAccPreCheck(t) },
        ProtoV6ProviderFactories: getFrameworkProviderFactories(),
        Steps: []resource.TestStep{
            {
                Config: generateFrameworkRoutingWrapupcodeResource(resourceLabel, name, util.NullValue, description),
                Check: resource.ComposeTestCheckFunc(
                    resource.TestCheckResourceAttr("genesyscloud_routing_wrapupcode."+resourceLabel, "name", name),
                    resource.TestCheckResourceAttr("genesyscloud_routing_wrapupcode."+resourceLabel, "description", description),
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
}
```

**Key Components**:

| Component | Purpose |
|-----------|---------|
| `PreCheck` | Verify test prerequisites (credentials, environment) |
| `ProtoV6ProviderFactories` | Provide muxed provider factory |
| `Steps` | Define test steps (create, update, import) |
| `Config` | Terraform HCL configuration for step |
| `Check` | Assertions to verify expected state |
| `ImportState` | Test import functionality |
| `CheckDestroy` | Verify resources are cleaned up |

---

### 2. Test Case Types

#### 2.1 Basic CRUD Test

**Purpose**: Test basic create, read, import, and destroy operations.

**Pattern**:
```go
func TestAccFrameworkResource<ResourceName>Basic(t *testing.T) {
    // Test with minimal required attributes
    // Verify create, read, import
    // Verify destroy
}
```

**What to Test**:
- Create resource with required attributes only
- Verify all attributes are set correctly
- Test import by ID
- Verify resource is destroyed

#### 2.2 Division/Dependency Test

**Purpose**: Test resource with dependencies (e.g., division assignment).

**Pattern**:
```go
func TestAccFrameworkResource<ResourceName>Division(t *testing.T) {
    // Create dependency resource (e.g., division)
    // Create resource with dependency reference
    // Verify dependency is correctly set
}
```

**Example** (routing_wrapupcode with division):
```go
func TestAccFrameworkResourceRoutingWrapupcodeDivision(t *testing.T) {
    var (
        resourceLabel    = "test_routing_wrapupcode_division"
        name             = "Terraform Framework Wrapupcode " + uuid.NewString()
        description      = "Test wrapupcode with division"
        divResourceLabel = "test_division"
        divName          = "terraform-" + uuid.NewString()
    )

    resource.Test(t, resource.TestCase{
        PreCheck:                 func() { util.TestAccPreCheck(t) },
        ProtoV6ProviderFactories: getFrameworkProviderFactories(),
        Steps: []resource.TestStep{
            {
                Config: authDivision.GenerateAuthDivisionBasic(divResourceLabel, divName) +
                    generateFrameworkRoutingWrapupcodeResource(resourceLabel, name, "genesyscloud_auth_division."+divResourceLabel+".id", description),
                Check: resource.ComposeTestCheckFunc(
                    resource.TestCheckResourceAttr("genesyscloud_routing_wrapupcode."+resourceLabel, "name", name),
                    resource.TestCheckResourceAttrPair("genesyscloud_routing_wrapupcode."+resourceLabel, "division_id", "genesyscloud_auth_division."+divResourceLabel, "id"),
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
}
```

**Key Points**:
- Use `TestCheckResourceAttrPair` to verify dependency reference
- Include dependency resource in HCL configuration
- Dependency resource must be in muxed provider factory

#### 2.3 Update Tests

**Purpose**: Test in-place updates of resource attributes.

**Pattern**:
```go
func TestAccFrameworkResource<ResourceName>Update(t *testing.T) {
    // Create resource with initial values
    // Update attribute values
    // Verify updates are applied
}
```

**Example** (routing_wrapupcode name update):
```go
func TestAccFrameworkResourceRoutingWrapupcodeNameUpdate(t *testing.T) {
    var (
        resourceLabel = "test_routing_wrapupcode_name_update"
        name1         = "Terraform Framework Wrapupcode " + uuid.NewString()
        name2         = "Terraform Framework Wrapupcode Updated " + uuid.NewString()
        description   = "Test wrapupcode name update"
    )

    resource.Test(t, resource.TestCase{
        PreCheck:                 func() { util.TestAccPreCheck(t) },
        ProtoV6ProviderFactories: getFrameworkProviderFactories(),
        Steps: []resource.TestStep{
            {
                Config: generateFrameworkRoutingWrapupcodeResource(resourceLabel, name1, util.NullValue, description),
                Check: resource.ComposeTestCheckFunc(
                    resource.TestCheckResourceAttr("genesyscloud_routing_wrapupcode."+resourceLabel, "name", name1),
                ),
            },
            {
                Config: generateFrameworkRoutingWrapupcodeResource(resourceLabel, name2, util.NullValue, description),
                Check: resource.ComposeTestCheckFunc(
                    resource.TestCheckResourceAttr("genesyscloud_routing_wrapupcode."+resourceLabel, "name", name2),
                ),
            },
        },
        CheckDestroy: testVerifyFrameworkWrapupcodesDestroyed,
    })
}
```

**What to Test**:
- Name updates
- Description updates
- Optional attribute updates
- Verify no resource replacement (in-place update)

#### 2.4 Lifecycle Test

**Purpose**: Comprehensive test covering multiple scenarios in one test.

**Pattern**:
```go
func TestAccFrameworkResource<ResourceName>Lifecycle(t *testing.T) {
    // Create without optional attributes
    // Add optional attributes
    // Update attributes
    // Verify all operations
}
```

**What to Test**:
- Create with minimal attributes
- Add optional attributes
- Update various attributes
- Test with and without dependencies
- Comprehensive coverage in single test

---

### 3. Destroy Verification Function

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

        resource, resp, err := api.Get<Resource>(rs.Primary.ID)
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

**Example** (routing_wrapupcode):
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

**Key Points**:
- Check only resources of the specific type
- Verify resource returns 404 (not found)
- Return error if resource still exists
- Return nil if all resources destroyed

---

### 4. HCL Generation Helper

**Purpose**: Generate Terraform HCL configuration for tests.

**Design Pattern**:
```go
// generateFramework<ResourceName>Resource generates a <resource> resource for Framework testing
func generateFramework<ResourceName>Resource(
    resourceLabel string,
    name string,
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
        name = "%s"%s%s
    }
    `, resourceLabel, name, optionalAttr1Str, optionalAttr2Str)
}
```

**Example** (routing_wrapupcode):
```go
func generateFrameworkRoutingWrapupcodeResource(
    resourceLabel string,
    name string,
    divisionId string,
    description string,
) string {
    divisionIdAttr := ""
    if divisionId != util.NullValue {
        divisionIdAttr = fmt.Sprintf(`
        division_id = %s`, divisionId)
    }

    descriptionAttr := ""
    if description != "" {
        descriptionAttr = fmt.Sprintf(`
        description = "%s"`, description)
    }

    return fmt.Sprintf(`resource "genesyscloud_routing_wrapupcode" "%s" {
        name = "%s"%s%s
    }
    `, resourceLabel, name, divisionIdAttr, descriptionAttr)
}
```

**Key Points**:
- Same pattern as Stage 1 helper function
- Handles optional attributes
- Uses `util.NullValue` for omitted references
- No quotes for references, quotes for literals

---

### 5. Provider Factory

**Purpose**: Create muxed provider that supports both Framework and SDKv2 resources.

**Design Pattern**:
```go
// getFrameworkProviderFactories returns provider factories for Framework testing
func getFrameworkProviderFactories() map[string]func() (tfprotov6.ProviderServer, error) {
    return map[string]func() (tfprotov6.ProviderServer, error){
        "genesyscloud": func() (tfprotov6.ProviderServer, error) {
            // Create Framework provider with <resource> resource and data source
            frameworkResources := map[string]func() frameworkresource.Resource{
                ResourceType: New<ResourceName>FrameworkResource,
            }
            frameworkDataSources := map[string]func() datasource.DataSource{
                ResourceType: New<ResourceName>FrameworkDataSource,
            }

            // Create muxed provider that includes both Framework and SDKv2 resources
            // This allows the test to use SDKv2 dependencies alongside Framework resource
            muxFactory := provider.NewMuxedProvider(
                "test",
                map[string]*schema.Resource{
                    // Add SDKv2 dependencies here
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

**Example** (routing_wrapupcode):
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

**Key Components**:

| Component | Purpose |
|-----------|---------|
| `frameworkResources` | Map of Framework resources to include |
| `frameworkDataSources` | Map of Framework data sources to include |
| `NewMuxedProvider` | Creates muxed provider with both SDKv2 and Framework |
| First map parameter | SDKv2 resources (dependencies) |
| Second map parameter | SDKv2 data sources (usually empty) |
| Third map parameter | Framework resources |
| Fourth map parameter | Framework data sources |

**Why Muxed Provider**:
- Enables testing Framework resource with SDKv2 dependencies
- Gradual migration without breaking tests
- Single provider instance for both types
- Transparent to Terraform test framework

---

## Part 2: Data Source Test File

### 1. Data Source Test Function

**Purpose**: Test data source lookup functionality.

**Design Pattern**:
```go
func TestAccFrameworkDataSource<ResourceName>(t *testing.T) {
    var (
        resourceLabel = "test_<resource>"
        dataLabel     = "test_data_<resource>"
        name          = "Terraform Framework Data <Resource> " + uuid.NewString()
        // ... other variables
    )

    resource.Test(t, resource.TestCase{
        PreCheck:                 func() { util.TestAccPreCheck(t) },
        ProtoV6ProviderFactories: getFrameworkProviderFactories(),
        Steps: []resource.TestStep{
            {
                Config: generateFramework<ResourceName>Resource(resourceLabel, name, ...) +
                    generateFramework<ResourceName>DataSource(dataLabel, name, "genesyscloud_<resource>."+resourceLabel),
                Check: resource.ComposeTestCheckFunc(
                    resource.TestCheckResourceAttrPair("data.genesyscloud_<resource>."+dataLabel, "id", "genesyscloud_<resource>."+resourceLabel, "id"),
                    resource.TestCheckResourceAttr("data.genesyscloud_<resource>."+dataLabel, "name", name),
                ),
            },
        },
    })
}
```

**Example** (routing_wrapupcode):
```go
func TestAccFrameworkDataSourceRoutingWrapupcode(t *testing.T) {
    var (
        resourceLabel = "test_routing_wrapupcode"
        dataLabel     = "test_data_wrapupcode"
        name          = "Terraform Framework Data Wrapupcode " + uuid.NewString()
        description   = "Test wrapupcode for data source"
    )

    resource.Test(t, resource.TestCase{
        PreCheck:                 func() { util.TestAccPreCheck(t) },
        ProtoV6ProviderFactories: getFrameworkProviderFactories(),
        Steps: []resource.TestStep{
            {
                Config: generateFrameworkRoutingWrapupcodeResource(resourceLabel, name, util.NullValue, description) +
                    generateFrameworkRoutingWrapupcodeDataSource(dataLabel, name, "genesyscloud_routing_wrapupcode."+resourceLabel),
                Check: resource.ComposeTestCheckFunc(
                    resource.TestCheckResourceAttrPair("data.genesyscloud_routing_wrapupcode."+dataLabel, "id", "genesyscloud_routing_wrapupcode."+resourceLabel, "id"),
                    resource.TestCheckResourceAttr("data.genesyscloud_routing_wrapupcode."+dataLabel, "name", name),
                ),
            },
        },
    })
}
```

**Key Points**:
- Create resource first, then look it up with data source
- Use `TestCheckResourceAttrPair` to verify IDs match
- Include `depends_on` in data source to ensure resource exists

---

### 2. Data Source HCL Generation Helper

**Purpose**: Generate data source HCL configuration.

**Design Pattern**:
```go
// generateFramework<ResourceName>DataSource generates a <resource> data source for Framework testing
func generateFramework<ResourceName>DataSource(
    dataLabel string,
    name string,
    dependsOnResource string,
) string {
    return fmt.Sprintf(`data "genesyscloud_<resource>" "%s" {
        name = "%s"
        depends_on = [%s]
    }
    `, dataLabel, name, dependsOnResource)
}
```

**Example** (routing_wrapupcode):
```go
func generateFrameworkRoutingWrapupcodeDataSource(
    dataLabel string,
    name string,
    dependsOnResource string,
) string {
    return fmt.Sprintf(`data "genesyscloud_routing_wrapupcode" "%s" {
        name = "%s"
        depends_on = [%s]
    }
    `, dataLabel, name, dependsOnResource)
}
```

**Key Points**:
- Include `depends_on` to ensure resource exists before lookup
- Use exact name for lookup
- Simple pattern for most data sources

---

## Part 3: Test Initialization File

### Purpose

The test initialization file provides package-level setup for tests. This is typically minimal for simple resources.

**Design Pattern**:
```go
package <resource_name>

import (
    "testing"
)

// Package-level test setup can go here if needed
// Most simple resources don't need additional setup
```

**When to Use**:
- Package-level test fixtures
- Shared test utilities
- Test environment configuration
- Usually minimal or empty for simple resources

---

## SDKv2 vs Plugin Framework Test Comparison

### Test Structure

**SDKv2**:
```go
func TestAccResourceRoutingWrapupcode_basic(t *testing.T) {
    resource.Test(t, resource.TestCase{
        PreCheck:          func() { util.TestAccPreCheck(t) },
        ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
        Steps: []resource.TestStep{
            {
                Config: generateRoutingWrapupcodeResource(...),
                Check: resource.ComposeTestCheckFunc(...),
            },
        },
        CheckDestroy: testVerifyWrapupcodesDestroyed,
    })
}
```

**Plugin Framework**:
```go
func TestAccFrameworkResourceRoutingWrapupcodeBasic(t *testing.T) {
    resource.Test(t, resource.TestCase{
        PreCheck:                 func() { util.TestAccPreCheck(t) },
        ProtoV6ProviderFactories: getFrameworkProviderFactories(),
        Steps: []resource.TestStep{
            {
                Config: generateFrameworkRoutingWrapupcodeResource(...),
                Check: resource.ComposeTestCheckFunc(...),
            },
        },
        CheckDestroy: testVerifyFrameworkWrapupcodesDestroyed,
    })
}
```

**Key Differences**:

| Aspect | SDKv2 | Plugin Framework |
|--------|-------|------------------|
| Test name | `TestAccResource*` | `TestAccFrameworkResource*` |
| Provider factory | `ProviderFactories` | `ProtoV6ProviderFactories` |
| Factory type | SDKv2 provider | Muxed provider (SDKv2 + Framework) |
| Helper functions | `generate*Resource()` | `generateFramework*Resource()` |
| Destroy check | `testVerify*Destroyed` | `testVerifyFramework*Destroyed` |

---

## Design Patterns and Best Practices

### Pattern 1: Unique Test Names

**Pattern**:
```go
name := "Terraform Framework <Resource> " + uuid.NewString()
```

**Why**:
- Prevents conflicts between parallel tests
- Enables test parallelization
- Avoids cleanup issues
- Unique identifier for debugging

### Pattern 2: Import Verification

**Pattern**:
```go
{
    ResourceName:      "genesyscloud_<resource>." + resourceLabel,
    ImportState:       true,
    ImportStateVerify: true,
}
```

**Why**:
- Verifies import functionality works
- Ensures state is correctly populated after import
- Tests complete resource lifecycle
- Required for all resources

### Pattern 3: Dependency Testing

**Pattern**:
```go
Config: authDivision.GenerateAuthDivisionBasic(divResourceLabel, divName) +
    generateFrameworkRoutingWrapupcodeResource(resourceLabel, name, "genesyscloud_auth_division."+divResourceLabel+".id", description),
Check: resource.ComposeTestCheckFunc(
    resource.TestCheckResourceAttrPair("genesyscloud_routing_wrapupcode."+resourceLabel, "division_id", "genesyscloud_auth_division."+divResourceLabel, "id"),
),
```

**Why**:
- Tests resource with real dependencies
- Verifies dependency references work
- Tests exporter dependency resolution
- Realistic test scenario

### Pattern 4: Update Testing

**Pattern**:
```go
Steps: []resource.TestStep{
    {
        Config: generateFramework<ResourceName>Resource(resourceLabel, name1, ...),
        Check: resource.ComposeTestCheckFunc(
            resource.TestCheckResourceAttr("genesyscloud_<resource>."+resourceLabel, "name", name1),
        ),
    },
    {
        Config: generateFramework<ResourceName>Resource(resourceLabel, name2, ...),
        Check: resource.ComposeTestCheckFunc(
            resource.TestCheckResourceAttr("genesyscloud_<resource>."+resourceLabel, "name", name2),
        ),
    },
}
```

**Why**:
- Verifies in-place updates work
- Tests that updates don't cause replacement
- Validates update logic
- Common user scenario

### Pattern 5: Comprehensive Lifecycle Test

**Pattern**:
```go
func TestAccFrameworkResource<ResourceName>Lifecycle(t *testing.T) {
    // Step 1: Create with minimal attributes
    // Step 2: Add optional attributes
    // Step 3: Update attributes
    // Step 4: Test with dependencies
    // Step 5: Import verification
}
```

**Why**:
- Single test covers multiple scenarios
- Reduces test execution time
- Comprehensive coverage
- Realistic user workflow

---

## Muxed Provider Design

### Architecture

```
┌─────────────────────────────────────────────────────────┐
│  Muxed Provider                                         │
├─────────────────────────────────────────────────────────┤
│  ┌───────────────────────────────────────────────────┐ │
│  │  Framework Resources                              │ │
│  │  - genesyscloud_routing_wrapupcode (Framework)    │ │
│  └───────────────────────────────────────────────────┘ │
│  ┌───────────────────────────────────────────────────┐ │
│  │  Framework Data Sources                           │ │
│  │  - genesyscloud_routing_wrapupcode (Framework)    │ │
│  └───────────────────────────────────────────────────┘ │
│  ┌───────────────────────────────────────────────────┐ │
│  │  SDKv2 Resources (Dependencies)                   │ │
│  │  - genesyscloud_auth_division (SDKv2)             │ │
│  └───────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────┘
```

### How It Works

1. **Test requests provider**: Terraform test framework requests "genesyscloud" provider
2. **Muxed provider created**: Factory creates muxed provider with both SDKv2 and Framework
3. **Resource routing**: Muxed provider routes requests to correct implementation
4. **Framework resource**: Requests for `genesyscloud_routing_wrapupcode` go to Framework
5. **SDKv2 dependency**: Requests for `genesyscloud_auth_division` go to SDKv2
6. **Transparent operation**: Test doesn't know about muxing

### Benefits

- **Gradual migration**: Migrate resources one at a time
- **Real dependencies**: Test with actual dependency resources
- **No mocking**: Use real SDKv2 resources as dependencies
- **Compatibility**: Works with existing test infrastructure

---

## Test Coverage Strategy

### Minimum Test Cases

#### Resource Tests
1. **Basic CRUD**: Essential functionality
2. **With Dependencies**: Real-world usage
3. **Update Tests**: Common operations
4. **Lifecycle**: Comprehensive coverage

#### Data Source Tests
1. **Basic Lookup**: Essential functionality
2. **With Dependencies**: Real-world usage

### Test Assertions

Each test should verify:
- [ ] Resource/data source ID is set
- [ ] All required attributes match expected values
- [ ] Optional attributes match when provided
- [ ] Computed attributes are populated
- [ ] Dependencies are correctly referenced
- [ ] Import works correctly
- [ ] Resources are destroyed properly

### Edge Cases to Test

- Creating with minimal attributes
- Creating with all attributes
- Updating each attribute individually
- Removing optional attributes
- Adding optional attributes
- Testing with various dependency combinations

---

## Migration Considerations

### Test Behavior Preservation

When migrating tests, verify:
- [ ] Test assertions are identical to SDKv2 version
- [ ] Test steps match SDKv2 version
- [ ] Test coverage is maintained or improved
- [ ] Test execution time is similar
- [ ] Test reliability is maintained

### Common Migration Pitfalls

#### Pitfall 1: Missing Dependency in Muxed Provider
**Problem**: Test fails because dependency resource not included in factory.
**Solution**: Add all SDKv2 dependencies to muxed provider factory.

#### Pitfall 2: Wrong Provider Factory Type
**Problem**: Using `ProviderFactories` instead of `ProtoV6ProviderFactories`.
**Solution**: Always use `ProtoV6ProviderFactories` for Framework tests.

#### Pitfall 3: Test Name Conflicts
**Problem**: Framework test has same name as SDKv2 test.
**Solution**: Use `TestAccFramework*` prefix for Framework tests.

#### Pitfall 4: Missing Import Test
**Problem**: Test doesn't verify import functionality.
**Solution**: Always include import step in tests.

#### Pitfall 5: Incorrect Destroy Verification
**Problem**: Destroy check doesn't filter by resource type.
**Solution**: Check `rs.Type` before verifying destruction.

---

## Summary

### Key Design Decisions

1. **Muxed Provider Pattern**: Support both Framework and SDKv2 in same test
2. **Test Naming Convention**: Clear distinction with `TestAccFramework*` prefix
3. **Test Coverage Preservation**: Maintain or improve coverage from SDKv2
4. **Helper Functions**: Reusable functions for HCL generation and verification
5. **Import Testing**: Always include import verification

### Test File Structure

```
Resource Test File:
├── Test functions (Basic, Division, Update, Lifecycle)
├── Destroy verification function
├── HCL generation helper
└── Provider factory

Data Source Test File:
├── Test functions (Basic, With Dependencies)
└── HCL generation helper

Test Initialization File:
└── Package-level setup (usually minimal)
```

### Next Steps

After completing Stage 3 test migration:
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
- **Muxed Providers**: https://developer.hashicorp.com/terraform/plugin/mux
