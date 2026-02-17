# Stage 3 – Test Migration Tasks (Complex Resources)

## Overview

This document provides step-by-step tasks for completing Stage 3 of the Plugin Framework migration for **complex resources**. Follow these tasks in order to migrate acceptance tests from SDKv2 to Plugin Framework patterns with comprehensive coverage for nested structures, multiple dependencies, and advanced test scenarios.

**Reference Implementation**: 
- `genesyscloud/user/resource_genesyscloud_user_test.go`
- `genesyscloud/user/data_source_genesyscloud_user_test.go`
- `genesyscloud/user/genesyscloud_user_init_test.go`

**Estimated Time**: 16-32 hours (depending on complexity and number of nested structures)

---

## Prerequisites

Before starting Stage 3 tasks, ensure:

- [ ] Stage 1 (Schema Migration) is complete and approved
- [ ] Stage 2 (Resource Migration) is complete and approved
- [ ] You have reviewed the existing SDKv2 test implementation
- [ ] You understand the muxed provider pattern
- [ ] You have read Stage 3 `requirements.md` and `design.md`
- [ ] You have studied the `user` reference implementation
- [ ] Test environment is configured (credentials, etc.)
- [ ] You understand nested structure testing patterns
- [ ] You understand helper function composition patterns

---

## Task Checklist

### Phase 1: Test Initialization File Setup
- [ ] Task 1.1: Create Test Initialization File
- [ ] Task 1.2: Add Package Declaration and Imports
- [ ] Task 1.3: Define Package-Level Variables
- [ ] Task 1.4: Create Registration Instance Struct
- [ ] Task 1.5: Implement SDKv2 Resource Registration
- [ ] Task 1.6: Implement SDKv2 Data Source Registration
- [ ] Task 1.7: Implement Framework Resource Registration
- [ ] Task 1.8: Implement Framework Data Source Registration
- [ ] Task 1.9: Implement Initialization Function
- [ ] Task 1.10: Implement TestMain Function

### Phase 2: Resource Test File Setup
- [ ] Task 2.1: Create Resource Test File
- [ ] Task 2.2: Add Package Declaration and Imports
- [ ] Task 2.3: Add Init Function
- [ ] Task 2.4: Implement Provider Factory
- [ ] Task 2.5: Implement Destroy Verification Function

### Phase 3: Basic Helper Functions
- [ ] Task 3.1: Implement Basic Resource Generation Helper
- [ ] Task 3.2: Implement Custom Attributes Helper
- [ ] Task 3.3: Implement Optional Attribute Helper

### Phase 4: Nested Block Helper Functions
- [ ] Task 4.1: Identify All Nested Blocks (1-level, 2-level, 3-level)
- [ ] Task 4.2: Implement 1-Level Nested Block Helpers
- [ ] Task 4.3: Implement 2-Level Nested Block Helpers
- [ ] Task 4.4: Implement 3-Level Nested Block Helpers
- [ ] Task 4.5: Implement Attribute-Specific Helpers

### Phase 5: Basic Resource Test Cases
- [ ] Task 5.1: Implement Basic CRUD Test
- [ ] Task 5.2: Implement Update Tests (Name, Description, etc.)

### Phase 6: Nested Structure Test Cases
- [ ] Task 6.1: Implement 1-Level Nested Structure Tests
- [ ] Task 6.2: Implement 2-Level Nested Structure Tests
- [ ] Task 6.3: Implement 3-Level Nested Structure Tests
- [ ] Task 6.4: Implement Array/Set Update Tests

### Phase 7: Dependency Test Cases
- [ ] Task 7.1: Implement SDKv2 Dependency Tests
- [ ] Task 7.2: Implement Framework Dependency Tests
- [ ] Task 7.3: Implement Multiple Dependency Tests

### Phase 8: Advanced Test Cases
- [ ] Task 8.1: Implement Edge Case Tests
- [ ] Task 8.2: Implement Validation Tests
- [ ] Task 8.3: Implement Lifecycle Test
- [ ] Task 8.4: Document Known API Asymmetries

### Phase 9: Data Source Test File
- [ ] Task 9.1: Create Data Source Test File
- [ ] Task 9.2: Add Package Declaration and Imports
- [ ] Task 9.3: Add Init Function
- [ ] Task 9.4: Implement Data Source HCL Helper
- [ ] Task 9.5: Implement Basic Data Source Test
- [ ] Task 9.6: Implement Multiple Search Criteria Tests

### Phase 10: Test Execution and Validation
- [ ] Task 10.1: Run Individual Tests
- [ ] Task 10.2: Run All Tests
- [ ] Task 10.3: Verify Test Coverage
- [ ] Task 10.4: Fix Any Test Failures

### Phase 11: Review and Approval
- [ ] Task 11.1: Review Against Checklist
- [ ] Task 11.2: Code Review and Approval

---

## Detailed Tasks

## Phase 1: Test Initialization File Setup

### Task 1.1: Create Test Initialization File

**Objective**: Create the test initialization file for package-level setup.

**Steps**:

1. **Navigate to resource package directory**
   ```powershell
   cd genesyscloud\<resource_name>
   ```

2. **Create the init test file**
   ```powershell
   New-Item -ItemType File -Name "genesyscloud_<resource_name>_init_test.go"
   ```
   Example:
   ```powershell
   New-Item -ItemType File -Name "genesyscloud_user_init_test.go"
   ```

**Deliverable**: Empty init test file created

---

### Task 1.2: Add Package Declaration and Imports

**Objective**: Set up the file with correct package and imports.

**Steps**:

1. **Add package declaration**
   ```go
   package <resource_name>
   ```

2. **Add required imports**
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

3. **Adjust imports based on your resource's dependencies**
   - Add imports for all SDKv2 dependencies
   - Add imports for all Framework dependencies
   - Remove unused imports

**Deliverable**: File with package declaration and imports

---

### Task 1.3: Define Package-Level Variables

**Objective**: Define package-level variables for test resources and data sources.

**Steps**:

1. **Add package-level variables**
   ```go
   // providerDataSources holds a map of all registered SDKv2 datasources,
   // should be removed after complete migration to plugin Framework
   var providerDataSources map[string]*schema.Resource

   // providerResources holds a map of all registered SDKv2 resources
   // should be removed after complete migration to plugin Framework
   var providerResources map[string]*schema.Resource

   // frameworkResources holds a map of all registered Framework resources
   var frameworkResources map[string]func() resource.Resource

   // frameworkDataSources holds a map of all registered Framework data sources
   var frameworkDataSources map[string]func() datasource.DataSource
   ```

**Deliverable**: Package-level variables defined

---

### Task 1.4: Create Registration Instance Struct

**Objective**: Create struct with mutexes for thread-safe registration.

**Steps**:

1. **Add registration instance struct**
   ```go
   type registerTestInstance struct {
       resourceMapMutex            sync.RWMutex
       datasourceMapMutex          sync.RWMutex
       frameworkResourceMapMutex   sync.RWMutex
       frameworkDataSourceMapMutex sync.RWMutex
   }
   ```

**Deliverable**: Registration instance struct defined


---

### Task 1.5: Implement SDKv2 Resource Registration

**Objective**: Register all SDKv2 resources needed for tests.

**Steps**:

1. **Add registerTestResources method**
   ```go
   // registerTestResources registers all SDKv2 resources used in the tests
   func (r *registerTestInstance) registerTestResources() {
       r.resourceMapMutex.Lock()
       defer r.resourceMapMutex.Unlock()

       // Register SDKv2 resources needed for Framework tests
       providerResources[authRole.ResourceType] = authRole.ResourceAuthRole()
       providerResources[authDivision.ResourceType] = authDivision.ResourceAuthDivision()
       // Add all other SDKv2 dependencies
   }
   ```

2. **Identify all SDKv2 dependencies**
   - Review your test cases
   - List all SDKv2 resources referenced
   - Add registration for each

**Example** (user resource):
```go
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
```

**Deliverable**: SDKv2 resource registration method implemented

---

### Task 1.6: Implement SDKv2 Data Source Registration

**Objective**: Register all SDKv2 data sources needed for tests.

**Steps**:

1. **Add registerTestDataSources method**
   ```go
   // registerTestDataSources registers all SDKv2 data sources used in the tests
   func (r *registerTestInstance) registerTestDataSources() {
       r.datasourceMapMutex.Lock()
       defer r.datasourceMapMutex.Unlock()

       // Register SDKv2 data sources needed for Framework tests
       providerDataSources[authRole.ResourceType] = authRole.DataSourceAuthRole()
       providerDataSources["genesyscloud_auth_division_home"] = genesyscloud.DataSourceAuthDivisionHome()
       // Add all other SDKv2 data source dependencies
   }
   ```

2. **Identify all SDKv2 data source dependencies**
   - Review your test cases
   - List all SDKv2 data sources referenced
   - Add registration for each

**Deliverable**: SDKv2 data source registration method implemented

---

### Task 1.7: Implement Framework Resource Registration

**Objective**: Register all Framework resources needed for tests.

**Steps**:

1. **Add registerFrameworkTestResources method**
   ```go
   // registerFrameworkTestResources registers all Framework resources used in the tests
   func (r *registerTestInstance) registerFrameworkTestResources() {
       r.frameworkResourceMapMutex.Lock()
       defer r.frameworkResourceMapMutex.Unlock()

       frameworkResources[ResourceType] = New<ResourceName>FrameworkResource
       // Add all other Framework resource dependencies
   }
   ```

2. **Identify all Framework dependencies**
   - Your resource under test
   - Any other Framework resources referenced in tests

**Example** (user resource):
```go
func (r *registerTestInstance) registerFrameworkTestResources() {
    r.frameworkResourceMapMutex.Lock()
    defer r.frameworkResourceMapMutex.Unlock()

    frameworkResources[ResourceType] = NewUserFrameworkResource
    frameworkResources[routinglanguage.ResourceType] = routinglanguage.NewFrameworkRoutingLanguageResource
}
```

**Deliverable**: Framework resource registration method implemented

---

### Task 1.8: Implement Framework Data Source Registration

**Objective**: Register all Framework data sources needed for tests.

**Steps**:

1. **Add registerFrameworkTestDataSources method**
   ```go
   // registerFrameworkTestDataSources registers all Framework data sources used in the tests
   func (r *registerTestInstance) registerFrameworkTestDataSources() {
       r.frameworkDataSourceMapMutex.Lock()
       defer r.frameworkDataSourceMapMutex.Unlock()

       frameworkDataSources[ResourceType] = New<ResourceName>FrameworkDataSource
       // Add all other Framework data source dependencies
   }
   ```

**Example** (user resource):
```go
func (r *registerTestInstance) registerFrameworkTestDataSources() {
    r.frameworkDataSourceMapMutex.Lock()
    defer r.frameworkDataSourceMapMutex.Unlock()

    frameworkDataSources[ResourceType] = NewUserFrameworkDataSource
    frameworkDataSources[routinglanguage.ResourceType] = routinglanguage.NewFrameworkRoutingLanguageDataSource
}
```

**Deliverable**: Framework data source registration method implemented

---

### Task 1.9: Implement Initialization Function

**Objective**: Create function that initializes all test resources and data sources.

**Steps**:

1. **Add initTestResources function**
   ```go
   // initTestResources initializes all test resources and data sources.
   func initTestResources() {
       // Initialize both SDKv2 and Framework resources for mixed provider tests
       providerResources = make(map[string]*schema.Resource)
       providerDataSources = make(map[string]*schema.Resource)
       frameworkResources = make(map[string]func() resource.Resource)
       frameworkDataSources = make(map[string]func() datasource.DataSource)

       regInstance := &registerTestInstance{}

       // Register SDKv2 resources and data sources (needed for dependencies)
       regInstance.registerTestResources()
       regInstance.registerTestDataSources()

       // Register Framework resources and data sources
       regInstance.registerFrameworkTestResources()
       regInstance.registerFrameworkTestDataSources()
   }
   ```

**Deliverable**: Initialization function implemented

---

### Task 1.10: Implement TestMain Function

**Objective**: Create TestMain function that runs before all tests.

**Steps**:

1. **Add TestMain function**
   ```go
   // TestMain is a "setup" function called by the testing framework when run the test
   func TestMain(m *testing.M) {
       // Run setup function before starting the test suite for <resource> package
       initTestResources()

       // Run the test suite for the <resource> package
       m.Run()
   }
   ```

**Deliverable**: TestMain function implemented

**Verification**: Test initialization file is complete

---

## Phase 2: Resource Test File Setup

### Task 2.1: Create Resource Test File

**Objective**: Create the resource test file.

**Steps**:

1. **Navigate to resource package directory**
   ```powershell
   cd genesyscloud\<resource_name>
   ```

2. **Create the test file**
   ```powershell
   New-Item -ItemType File -Name "resource_genesyscloud_<resource_name>_test.go"
   ```

**Deliverable**: Empty test file created

---

### Task 2.2: Add Package Declaration and Imports

**Objective**: Set up the file with correct package and imports.

**Steps**:

1. **Add package declaration**
   ```go
   package <resource_name>
   ```

2. **Add required imports**
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
       // Add imports for all dependencies
       "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
   )
   ```

3. **Adjust imports based on dependencies**
   - Add imports for all SDKv2 dependencies
   - Add imports for all Framework dependencies
   - Remove unused imports

**Deliverable**: File with package declaration and imports

---

### Task 2.3: Add Init Function

**Objective**: Ensure test resources are initialized before tests run.

**Steps**:

1. **Add init function**
   ```go
   // Ensure test resources are initialized for Framework tests
   func init() {
       if frameworkResources == nil || frameworkDataSources == nil {
           initTestResources()
       }
   }
   ```

**Deliverable**: Init function added

---

### Task 2.4: Implement Provider Factory

**Objective**: Create muxed provider factory for tests.

**Steps**:

1. **Decide on factory approach**
   - Simple case (no SDKv2 dependencies): Use `getFrameworkProviderFactories()`
   - Complex case (multiple dependencies): Use `provider.GetMuxedProviderFactories()` inline

2. **For simple case, add provider factory function**
   ```go
   // getFrameworkProviderFactories returns provider factories for Framework testing.
   // This creates a muxed provider that includes:
   //   - Framework resources: genesyscloud_<resource> (for creating test resources)
   //   - Framework data sources: genesyscloud_<resource> (for testing data source lookups)
   //   - SDKv2 resources: Any dependencies needed (add if needed)
   //
   // The muxed provider allows tests to use both Framework and SDKv2 resources together.
   func getFrameworkProviderFactories() map[string]func() (tfprotov6.ProviderServer, error) {
       return map[string]func() (tfprotov6.ProviderServer, error){
           "genesyscloud": func() (tfprotov6.ProviderServer, error) {
               frameworkResources := map[string]func() frameworkresource.Resource{
                   ResourceType: New<ResourceName>FrameworkResource,
               }
               frameworkDataSources := map[string]func() datasource.DataSource{
                   ResourceType: New<ResourceName>FrameworkDataSource,
               }

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

3. **For complex case, use inline in tests**
   ```go
   ProtoV6ProviderFactories: provider.GetMuxedProviderFactories(
       map[string]*schema.Resource{
           // SDKv2 dependencies
           authDivision.ResourceType: authDivision.ResourceAuthDivision(),
       },
       nil, // SDKv2 data sources
       map[string]func() frameworkresource.Resource{
           ResourceType: New<ResourceName>FrameworkResource,
       },
       map[string]func() datasource.DataSource{
           ResourceType: New<ResourceName>FrameworkDataSource,
       },
   ),
   ```

**Deliverable**: Provider factory implemented


---

### Task 2.5: Implement Destroy Verification Function

**Objective**: Create function to verify resources are destroyed after test.

**Steps**:

1. **Add destroy verification function**
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

2. **Customize for your resource**
   - Use correct API client
   - Use correct resource type string
   - Use correct API method to check existence

**Deliverable**: Destroy verification function implemented

---

## Phase 3: Basic Helper Functions

### Task 3.1: Implement Basic Resource Generation Helper

**Objective**: Create helper function to generate basic resource HCL.

**Steps**:

1. **Add basic resource generation function**
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

2. **Customize for your resource**
   - Add parameters for all required attributes
   - Add parameters for common optional attributes
   - Handle optional attributes with null checks
   - Use `util.NullValue` for reference attributes
   - Use empty string check for literal attributes

**Deliverable**: Basic resource generation helper implemented

---

### Task 3.2: Implement Custom Attributes Helper

**Objective**: Create helper function for flexible attribute composition.

**Steps**:

1. **Add custom attributes helper**
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

**Deliverable**: Custom attributes helper implemented

---

### Task 3.3: Implement Optional Attribute Helper

**Objective**: Create helper function for generating optional attributes.

**Steps**:

1. **Add optional attribute helper**
   ```go
   // generateOptionalAttr generates an optional attribute if value is not null
   func generateOptionalAttr(attrName string, value string) string {
       if value == util.NullValue || value == "" {
           return ""
       }
       return fmt.Sprintf(`%s = %s`, attrName, value)
   }
   ```

**Deliverable**: Optional attribute helper implemented

---

## Phase 4: Nested Block Helper Functions

### Task 4.1: Identify All Nested Blocks

**Objective**: Identify all nested blocks in your resource schema.

**Steps**:

1. **Review your schema file from Stage 1**
   - Identify all 1-level nested blocks
   - Identify all 2-level nested blocks
   - Identify all 3-level nested blocks

2. **Create a list of nested blocks**
   - Example for user resource:
     - 1-level: `addresses`, `employer_info`, `voicemail_userpolicies`, `routing_utilization`
     - 2-level: `addresses.phone_numbers`, `addresses.other_emails`, `routing_utilization.call`, `routing_utilization.callback`
     - 3-level: `routing_utilization.call.label_utilizations`, `routing_utilization.call.interruptible_media_types`

**Deliverable**: List of all nested blocks by level

---

### Task 4.2: Implement 1-Level Nested Block Helpers

**Objective**: Create helper functions for 1-level nested blocks.

**Steps**:

1. **For each 1-level nested block, add a helper function**
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

2. **Customize for each nested block**
   - Handle sub-nested blocks if applicable
   - Handle arrays/sets of nested blocks

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

**Deliverable**: 1-level nested block helpers implemented

---

### Task 4.3: Implement 2-Level Nested Block Helpers

**Objective**: Create helper functions for 2-level nested blocks.

**Steps**:

1. **For each 2-level nested block, add a helper function**
   ```go
   // generate<SubNestedBlock> generates a <sub_nested_block> block content (not wrapped)
   func generate<SubNestedBlock>(attr1, attr2, attr3 string, extras ...string) string {
       return fmt.Sprintf(`attr1 = %s
           attr2 = %s
           attr3 = %s
           %s`, attr1, attr2, attr3, strings.Join(extras, "\n"))
   }
   ```

2. **Customize for each sub-nested block**
   - Return content only (not wrapped in block)
   - Handle optional attributes
   - Support extras parameter for additional attributes

**Example** (user phone address):
```go
func generateFrameworkUserPhoneAddress(phoneNum, phoneMediaType, phoneType, extension string, extras ...string) string {
    return fmt.Sprintf(`number = %s
        media_type = %s
        type = %s
        extension = %s
        %s`, phoneNum, phoneMediaType, phoneType, extension, strings.Join(extras, "\n"))
}
```

**Deliverable**: 2-level nested block helpers implemented

---

### Task 4.4: Implement 3-Level Nested Block Helpers

**Objective**: Create helper functions for 3-level nested blocks.

**Steps**:

1. **For each 3-level nested block, add a helper function**
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

**Example** (user routing_utilization call block):
```go
func generateFrameworkRoutingUtilizationCall(maxCapacity, includeNonAcd string) string {
    return fmt.Sprintf(`call {
        maximum_capacity = %s
        include_non_acd = %s
    }`, maxCapacity, includeNonAcd)
}
```

**Deliverable**: 3-level nested block helpers implemented

---

### Task 4.5: Implement Attribute-Specific Helpers

**Objective**: Create helper functions for complex attributes (arrays, etc.).

**Steps**:

1. **For each complex attribute, add a helper function**
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

**Deliverable**: Attribute-specific helpers implemented

---

## Phase 5: Basic Resource Test Cases

### Task 5.1: Implement Basic CRUD Test

**Objective**: Test basic create, read, import, and destroy operations.

**Steps**:

1. **Add test function**
   ```go
   func TestAccFrameworkResource<ResourceName>Basic(t *testing.T) {
       t.Parallel()
       var (
           resourceLabel = "test_<resource>"
           // Define test variables
           requiredAttr1 = "value1"
           requiredAttr2 = "value2"
       )

       resource.Test(t, resource.TestCase{
           PreCheck: func() { util.TestAccPreCheck(t) },
           ProtoV6ProviderFactories: getFrameworkProviderFactories(),
           Steps: []resource.TestStep{
               {
                   // Create
                   Config: generateFramework<ResourceName>Resource(resourceLabel, requiredAttr1, requiredAttr2, util.NullValue, util.NullValue),
                   Check: resource.ComposeTestCheckFunc(
                       resource.TestCheckResourceAttr("genesyscloud_<resource>."+resourceLabel, "required_attr1", requiredAttr1),
                       resource.TestCheckResourceAttr("genesyscloud_<resource>."+resourceLabel, "required_attr2", requiredAttr2),
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

2. **Customize for your resource**
   - Add test variables for all attributes
   - Add assertions for all attributes
   - Use unique names with `uuid.NewString()`
   - Add `ImportStateVerifyIgnore` if needed

**Deliverable**: Basic CRUD test implemented


---

### Task 5.2: Implement Update Tests

**Objective**: Test in-place updates of resource attributes.

**Steps**:

1. **Add update test for each updatable attribute**
   - Name update test
   - Description update test
   - Other attribute update tests

2. **Example pattern**
   ```go
   func TestAccFrameworkResource<ResourceName>NameUpdate(t *testing.T) {
       t.Parallel()
       var (
           resourceLabel = "test_<resource>_name_update"
           name1         = "Name 1 " + uuid.NewString()
           name2         = "Name 2 " + uuid.NewString()
       )

       resource.Test(t, resource.TestCase{
           PreCheck:                 func() { util.TestAccPreCheck(t) },
           ProtoV6ProviderFactories: getFrameworkProviderFactories(),
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
           },
           CheckDestroy: testVerifyFramework<ResourceName>Destroyed,
       })
   }
   ```

**Deliverable**: Update tests implemented

---

## Phase 6: Nested Structure Test Cases

### Task 6.1: Implement 1-Level Nested Structure Tests

**Objective**: Test resources with 1-level nested blocks.

**Steps**:

1. **For each 1-level nested block, add a test**
   ```go
   func TestAccFrameworkResource<ResourceName><NestedBlock>(t *testing.T) {
       t.Parallel()
       var (
           resourceLabel = "test_<resource>_<nested_block>"
           // Define test variables
       )

       resource.Test(t, resource.TestCase{
           PreCheck:                 func() { util.TestAccPreCheck(t) },
           ProtoV6ProviderFactories: getFrameworkProviderFactories(),
           Steps: []resource.TestStep{
               {
                   // Create with nested block
                   Config: generateFramework<ResourceName>WithCustomAttrs(
                       resourceLabel, requiredAttr1, requiredAttr2,
                       generate<NestedBlock>(...),
                   ),
                   Check: resource.ComposeTestCheckFunc(
                       resource.TestCheckResourceAttr("genesyscloud_<resource>."+resourceLabel, "<nested_block>.#", "1"),
                       resource.TestCheckResourceAttr("genesyscloud_<resource>."+resourceLabel, "<nested_block>.0.attribute", value),
                   ),
               },
               {
                   // Update nested block
                   Config: generateFramework<ResourceName>WithCustomAttrs(
                       resourceLabel, requiredAttr1, requiredAttr2,
                       generate<NestedBlock>(...), // Updated values
                   ),
                   Check: resource.ComposeTestCheckFunc(
                       resource.TestCheckResourceAttr("genesyscloud_<resource>."+resourceLabel, "<nested_block>.0.attribute", newValue),
                   ),
               },
               {
                   // Import
                   ResourceName:      "genesyscloud_<resource>." + resourceLabel,
                   ImportState:       true,
                   ImportStateVerify: true,
               },
           },
           CheckDestroy: testVerifyFramework<ResourceName>Destroyed,
       })
   }
   ```

**Deliverable**: 1-level nested structure tests implemented

---

### Task 6.2: Implement 2-Level Nested Structure Tests

**Objective**: Test resources with 2-level nested blocks.

**Steps**:

1. **For each 2-level nested structure, add a test**
   - Test creating with 2-level nested blocks
   - Test updating 2-level nested attributes
   - Test multiple elements in nested arrays/sets

2. **Use appropriate assertion methods**
   - `TestCheckResourceAttr` for specific elements
   - `TestCheckTypeSetElemNestedAttrs` for set elements

**Example pattern**:
```go
Config: generateFramework<ResourceName>WithCustomAttrs(
    resourceLabel, requiredAttr1, requiredAttr2,
    generate<Level1Block>(
        generate<Level2Block>(attr1, attr2, attr3),
    ),
),
Check: resource.ComposeTestCheckFunc(
    resource.TestCheckResourceAttr("genesyscloud_<resource>."+resourceLabel, "<level1>.0.<level2>.0.attr1", value1),
    resource.TestCheckResourceAttr("genesyscloud_<resource>."+resourceLabel, "<level1>.0.<level2>.0.attr2", value2),
),
```

**Deliverable**: 2-level nested structure tests implemented

---

### Task 6.3: Implement 3-Level Nested Structure Tests

**Objective**: Test resources with 3-level nested blocks.

**Steps**:

1. **For each 3-level nested structure, add a test**
   - Test creating with 3-level nested blocks
   - Test updating 3-level nested attributes
   - Test complex scenarios with multiple elements

2. **Use appropriate assertion methods**
   ```go
   resource.TestCheckResourceAttr("genesyscloud_<resource>."+resourceLabel, "<level1>.0.<level2>.0.<level3>.#", "1")
   resource.TestCheckResourceAttr("genesyscloud_<resource>."+resourceLabel, "<level1>.0.<level2>.0.<level3>.0.attr", value)
   ```

**Deliverable**: 3-level nested structure tests implemented

---

### Task 6.4: Implement Array/Set Update Tests

**Objective**: Test adding, removing, and modifying elements in arrays/sets.

**Steps**:

1. **Add test for array/set operations**
   ```go
   func TestAccFrameworkResource<ResourceName><ArrayAttribute>(t *testing.T) {
       t.Parallel()
       var (
           resourceLabel = "test_<resource>_<array>"
           value1        = "value1"
           value2        = "value2"
       )

       resource.Test(t, resource.TestCase{
           PreCheck:                 func() { util.TestAccPreCheck(t) },
           ProtoV6ProviderFactories: getFrameworkProviderFactories(),
           Steps: []resource.TestStep{
               {
                   // Create with single element
                   Config: generateFramework<ResourceName>WithCustomAttrs(
                       resourceLabel, requiredAttr1, requiredAttr2,
                       generate<Attribute>(value1),
                   ),
                   Check: resource.ComposeTestCheckFunc(
                       resource.TestCheckResourceAttr("genesyscloud_<resource>."+resourceLabel, "<attribute>.#", "1"),
                       resource.TestCheckResourceAttr("genesyscloud_<resource>."+resourceLabel, "<attribute>.0", value1),
                   ),
               },
               {
                   // Update to different element
                   Config: generateFramework<ResourceName>WithCustomAttrs(
                       resourceLabel, requiredAttr1, requiredAttr2,
                       generate<Attribute>(value2),
                   ),
                   Check: resource.ComposeTestCheckFunc(
                       resource.TestCheckResourceAttr("genesyscloud_<resource>."+resourceLabel, "<attribute>.#", "1"),
                       resource.TestCheckResourceAttr("genesyscloud_<resource>."+resourceLabel, "<attribute>.0", value2),
                   ),
               },
               {
                   // Remove elements (empty array)
                   Config: generateFramework<ResourceName>WithCustomAttrs(
                       resourceLabel, requiredAttr1, requiredAttr2,
                       "<attribute> = []",
                   ),
                   Check: resource.ComposeTestCheckFunc(
                       resource.TestCheckResourceAttr("genesyscloud_<resource>."+resourceLabel, "<attribute>.#", "0"),
                   ),
               },
           },
           CheckDestroy: testVerifyFramework<ResourceName>Destroyed,
       })
   }
   ```

**Deliverable**: Array/set update tests implemented

---

## Phase 7: Dependency Test Cases

### Task 7.1: Implement SDKv2 Dependency Tests

**Objective**: Test resource with SDKv2 dependencies.

**Steps**:

1. **For each SDKv2 dependency, add a test**
   ```go
   func TestAccFrameworkResource<ResourceName>With<Dependency>(t *testing.T) {
       t.Parallel()
       var (
           resourceLabel = "test_<resource>_<dependency>"
           depLabel      = "test_<dependency>"
           depName       = "Dependency " + uuid.NewString()
       )

       resource.Test(t, resource.TestCase{
           PreCheck: func() { util.TestAccPreCheck(t) },
           ProtoV6ProviderFactories: provider.GetMuxedProviderFactories(
               map[string]*schema.Resource{
                   dependency.ResourceType: dependency.ResourceDependency(),
               },
               nil,
               map[string]func() frameworkresource.Resource{
                   ResourceType: New<ResourceName>FrameworkResource,
               },
               map[string]func() datasource.DataSource{
                   ResourceType: New<ResourceName>FrameworkDataSource,
               },
           ),
           Steps: []resource.TestStep{
               {
                   Config: dependency.GenerateDependencyResource(depLabel, depName) +
                       generateFramework<ResourceName>WithCustomAttrs(
                           resourceLabel, requiredAttr1, requiredAttr2,
                           fmt.Sprintf(`dependency_id = genesyscloud_<dependency>.%s.id`, depLabel),
                       ),
                   Check: resource.ComposeTestCheckFunc(
                       resource.TestCheckResourceAttrPair(
                           "genesyscloud_<resource>."+resourceLabel,
                           "dependency_id",
                           "genesyscloud_<dependency>."+depLabel,
                           "id",
                       ),
                   ),
               },
           },
           CheckDestroy: testVerifyFramework<ResourceName>Destroyed,
       })
   }
   ```

**Deliverable**: SDKv2 dependency tests implemented

---

### Task 7.2: Implement Framework Dependency Tests

**Objective**: Test resource with Framework dependencies.

**Steps**:

1. **For each Framework dependency, add a test**
   - Similar pattern to SDKv2 dependency tests
   - Include Framework dependency in muxed provider factory

**Deliverable**: Framework dependency tests implemented

---

### Task 7.3: Implement Multiple Dependency Tests

**Objective**: Test resource with multiple dependencies.

**Steps**:

1. **Add test with multiple dependencies**
   - Include all dependencies in muxed provider factory
   - Test with various combinations of dependencies

**Deliverable**: Multiple dependency tests implemented

---

## Phase 8: Advanced Test Cases

### Task 8.1: Implement Edge Case Tests

**Objective**: Test API asymmetries and edge cases.

**Steps**:

1. **Identify edge cases from SDKv2 tests**
   - Review SDKv2 test file
   - List all edge cases tested

2. **For each edge case, add a test**
   - Document the edge case with comments
   - Test the specific scenario
   - Verify handling is correct

3. **Document known API asymmetries**
   - Add detailed TODO comments
   - Explain the issue
   - Describe potential resolutions
   - Comment out failing test steps if needed

**Example** (API asymmetry):
```go
// TODO (<ISSUE-ID>): Detailed explanation
// - Expected behavior
// - Actual behavior
// - Why it worked in SDKv2 but fails in Framework
// - Potential resolution approaches
/*{
    Config: ...,
    Check: ...,
},*/
```

**Deliverable**: Edge case tests implemented and documented

---

### Task 8.2: Implement Validation Tests

**Objective**: Test complex constraint validation.

**Steps**:

1. **For each validation constraint, add a test**
   ```go
   func TestAccFrameworkResource<ResourceName>Validation(t *testing.T) {
       t.Parallel()
       var (
           resourceLabel = "test_<resource>_validation"
       )

       resource.Test(t, resource.TestCase{
           PreCheck:                 func() { util.TestAccPreCheck(t) },
           ProtoV6ProviderFactories: getFrameworkProviderFactories(),
           Steps: []resource.TestStep{
               {
                   // Test invalid value
                   Config: generateFramework<ResourceName>WithInvalidValue(...),
                   ExpectError: regexp.MustCompile(`expected <attribute> to be in the range \(min - max\)`),
               },
               {
                   // Test boundary value (min)
                   Config: generateFramework<ResourceName>WithBoundaryValue(..., minValue),
                   Check: resource.ComposeTestCheckFunc(
                       resource.TestCheckResourceAttr("genesyscloud_<resource>."+resourceLabel, "<attribute>", minValue),
                   ),
               },
               {
                   // Test boundary value (max)
                   Config: generateFramework<ResourceName>WithBoundaryValue(..., maxValue),
                   Check: resource.ComposeTestCheckFunc(
                       resource.TestCheckResourceAttr("genesyscloud_<resource>."+resourceLabel, "<attribute>", maxValue),
                   ),
               },
           },
       })
   }
   ```

**Deliverable**: Validation tests implemented


---

### Task 8.3: Implement Lifecycle Test

**Objective**: Comprehensive test covering multiple scenarios.

**Steps**:

1. **Add lifecycle test**
   ```go
   func TestAccFrameworkResource<ResourceName>Lifecycle(t *testing.T) {
       t.Parallel()
       var (
           resourceLabel = "test_<resource>_lifecycle"
           // Define test variables for all scenarios
       )

       resource.Test(t, resource.TestCase{
           PreCheck:                 func() { util.TestAccPreCheck(t) },
           ProtoV6ProviderFactories: getFrameworkProviderFactories(),
           Steps: []resource.TestStep{
               {
                   // Create with minimal attributes
                   Config: generateFramework<ResourceName>Resource(resourceLabel, requiredAttr1, requiredAttr2, util.NullValue, util.NullValue),
                   Check: resource.ComposeTestCheckFunc(
                       resource.TestCheckResourceAttr("genesyscloud_<resource>."+resourceLabel, "required_attr1", requiredAttr1),
                   ),
               },
               {
                   // Add optional attributes
                   Config: generateFramework<ResourceName>Resource(resourceLabel, requiredAttr1, requiredAttr2, optionalAttr1, optionalAttr2),
                   Check: resource.ComposeTestCheckFunc(
                       resource.TestCheckResourceAttr("genesyscloud_<resource>."+resourceLabel, "optional_attr1", optionalAttr1),
                   ),
               },
               {
                   // Add nested blocks
                   Config: generateFramework<ResourceName>WithCustomAttrs(
                       resourceLabel, requiredAttr1, requiredAttr2,
                       generate<NestedBlock>(...),
                   ),
                   Check: resource.ComposeTestCheckFunc(
                       resource.TestCheckResourceAttr("genesyscloud_<resource>."+resourceLabel, "<nested_block>.#", "1"),
                   ),
               },
               {
                   // Update nested blocks
                   Config: generateFramework<ResourceName>WithCustomAttrs(
                       resourceLabel, requiredAttr1, requiredAttr2,
                       generate<NestedBlock>(...), // Updated values
                   ),
                   Check: resource.ComposeTestCheckFunc(
                       resource.TestCheckResourceAttr("genesyscloud_<resource>."+resourceLabel, "<nested_block>.0.attribute", newValue),
                   ),
               },
               {
                   // Import
                   ResourceName:      "genesyscloud_<resource>." + resourceLabel,
                   ImportState:       true,
                   ImportStateVerify: true,
               },
           },
           CheckDestroy: testVerifyFramework<ResourceName>Destroyed,
       })
   }
   ```

**Deliverable**: Lifecycle test implemented

---

### Task 8.4: Document Known API Asymmetries

**Objective**: Document all known API asymmetries with detailed TODO comments.

**Steps**:

1. **Review all edge cases**
   - Identify API asymmetries
   - Identify behaviors that differ from SDKv2

2. **For each asymmetry, add detailed TODO comment**
   - Issue identifier (e.g., ADDRESSES-DELETION-ASYMMETRY)
   - Expected behavior
   - Actual behavior
   - Why it worked in SDKv2 but fails in Framework
   - Potential resolution approaches
   - Current status

3. **Comment out failing test steps**
   - Keep the test step code
   - Add detailed explanation above
   - Mark as TODO for future resolution

**Deliverable**: All API asymmetries documented

---

## Phase 9: Data Source Test File

### Task 9.1: Create Data Source Test File

**Objective**: Create the data source test file.

**Steps**:

1. **Navigate to resource package directory**
   ```powershell
   cd genesyscloud\<resource_name>
   ```

2. **Create the test file**
   ```powershell
   New-Item -ItemType File -Name "data_source_genesyscloud_<resource_name>_test.go"
   ```

**Deliverable**: Empty data source test file created

---

### Task 9.2: Add Package Declaration and Imports

**Objective**: Set up the file with correct package and imports.

**Steps**:

1. **Add package declaration and imports**
   ```go
   package <resource_name>

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

**Deliverable**: File with package declaration and imports

---

### Task 9.3: Add Init Function

**Objective**: Ensure test resources are initialized.

**Steps**:

1. **Add init function**
   ```go
   // Ensure test resources are initialized for Framework tests
   func init() {
       if frameworkResources == nil || frameworkDataSources == nil {
           initTestResources()
       }
   }
   ```

**Deliverable**: Init function added

---

### Task 9.4: Implement Data Source HCL Helper

**Objective**: Create helper function to generate data source HCL.

**Steps**:

1. **Add data source HCL generation function**
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

**Deliverable**: Data source HCL helper implemented

---

### Task 9.5: Implement Basic Data Source Test

**Objective**: Test basic data source lookup.

**Steps**:

1. **Add test function**
   ```go
   func TestAccFrameworkDataSource<ResourceName>(t *testing.T) {
       t.Parallel()
       var (
           resourceLabel   = "test_<resource>_resource"
           dataSourceLabel = "test_<resource>_data_source"
           identifier1     = "value1"
       )

       resource.Test(t, resource.TestCase{
           PreCheck:                 func() { util.TestAccPreCheck(t) },
           ProtoV6ProviderFactories: getFrameworkProviderFactories(),
           Steps: []resource.TestStep{
               {
                   Config: generateFramework<ResourceName>Resource(resourceLabel, identifier1, ...) +
                       generateFramework<ResourceName>DataSource(dataSourceLabel, identifier1, util.NullValue, "genesyscloud_<resource>."+resourceLabel),
                   Check: resource.ComposeTestCheckFunc(
                       resource.TestCheckResourceAttrPair("data.genesyscloud_<resource>."+dataSourceLabel, "id", "genesyscloud_<resource>."+resourceLabel, "id"),
                       resource.TestCheckResourceAttrPair("data.genesyscloud_<resource>."+dataSourceLabel, "name", "genesyscloud_<resource>."+resourceLabel, "name"),
                   ),
               },
           },
       })
   }
   ```

**Deliverable**: Basic data source test implemented

---

### Task 9.6: Implement Multiple Search Criteria Tests

**Objective**: Test data source lookup with multiple search criteria.

**Steps**:

1. **Add test steps for each search criterion**
   ```go
   {
       // Search by identifier1
       Config: generateFramework<ResourceName>Resource(resourceLabel, identifier1, ...) +
           generateFramework<ResourceName>DataSource(dataSourceLabel, identifier1, util.NullValue, "genesyscloud_<resource>."+resourceLabel),
       Check: resource.ComposeTestCheckFunc(
           resource.TestCheckResourceAttrPair("data.genesyscloud_<resource>."+dataSourceLabel, "id", "genesyscloud_<resource>."+resourceLabel, "id"),
       ),
   },
   {
       // Search by identifier2
       Config: generateFramework<ResourceName>Resource(resourceLabel, identifier1, ...) +
           generateFramework<ResourceName>DataSource(dataSourceLabel, util.NullValue, identifier2, "genesyscloud_<resource>."+resourceLabel),
       Check: resource.ComposeTestCheckFunc(
           resource.TestCheckResourceAttrPair("data.genesyscloud_<resource>."+dataSourceLabel, "id", "genesyscloud_<resource>."+resourceLabel, "id"),
       ),
   },
   ```

**Deliverable**: Multiple search criteria tests implemented

---

## Phase 10: Test Execution and Validation

### Task 10.1: Run Individual Tests

**Objective**: Run each test individually to verify they pass.

**Steps**:

1. **Run basic test**
   ```powershell
   go test -v -run TestAccFrameworkResource<ResourceName>Basic ./genesyscloud/<resource_name>
   ```

2. **Run nested structure tests**
   ```powershell
   go test -v -run TestAccFrameworkResource<ResourceName><NestedBlock> ./genesyscloud/<resource_name>
   ```

3. **Run dependency tests**
   ```powershell
   go test -v -run TestAccFrameworkResource<ResourceName>With<Dependency> ./genesyscloud/<resource_name>
   ```

4. **Run edge case tests**
   ```powershell
   go test -v -run TestAccFrameworkResource<ResourceName><EdgeCase> ./genesyscloud/<resource_name>
   ```

5. **Run validation tests**
   ```powershell
   go test -v -run TestAccFrameworkResource<ResourceName>Validation ./genesyscloud/<resource_name>
   ```

6. **Run lifecycle test**
   ```powershell
   go test -v -run TestAccFrameworkResource<ResourceName>Lifecycle ./genesyscloud/<resource_name>
   ```

7. **Run data source tests**
   ```powershell
   go test -v -run TestAccFrameworkDataSource<ResourceName> ./genesyscloud/<resource_name>
   ```

8. **Fix any failures**
   - Review error messages
   - Check test configuration
   - Verify resource implementation
   - Update tests as needed

**Deliverable**: All individual tests pass

---

### Task 10.2: Run All Tests

**Objective**: Run all tests together to verify they pass.

**Steps**:

1. **Run all Framework tests for the package**
   ```powershell
   go test -v -run TestAccFramework ./genesyscloud/<resource_name>
   ```

2. **Verify all tests pass**
   - Check for any failures
   - Verify resources are cleaned up
   - Check for test flakiness

3. **Run tests multiple times to check for flakiness**
   ```powershell
   go test -v -run TestAccFramework ./genesyscloud/<resource_name> -count=3
   ```

**Deliverable**: All tests pass consistently

---

### Task 10.3: Verify Test Coverage

**Objective**: Ensure test coverage matches or exceeds SDKv2 version.

**Steps**:

1. **Review test coverage**
   - [ ] Basic CRUD operations tested
   - [ ] All nested structures tested (1-level, 2-level, 3-level)
   - [ ] All dependencies tested (SDKv2 and Framework)
   - [ ] Array/set operations tested
   - [ ] Update scenarios tested
   - [ ] Edge cases tested
   - [ ] Validation scenarios tested
   - [ ] Lifecycle scenarios tested
   - [ ] Data source lookup tested
   - [ ] Multiple search criteria tested
   - [ ] Import functionality tested
   - [ ] Destroy verification tested

2. **Compare with SDKv2 tests**
   - Verify all SDKv2 test cases are migrated
   - Identify any missing test cases
   - Add additional tests if needed

**Deliverable**: Test coverage verified and complete

---

### Task 10.4: Fix Any Test Failures

**Objective**: Address any test failures or issues.

**Steps**:

1. **Common test failures and solutions**:

   **Failure: Provider not found**
   - Check provider factory is correctly implemented
   - Verify resource type is registered in factory
   - Ensure imports are correct
   - Verify init function is called

   **Failure: Resource not destroyed**
   - Check destroy verification function
   - Verify API call is correct
   - Check for resource type filter
   - Verify 404 handling

   **Failure: Attribute mismatch**
   - Verify HCL generation is correct
   - Check attribute names match schema
   - Verify test assertions are correct
   - Check for type conversion issues

   **Failure: Dependency not found**
   - Ensure dependency is in muxed provider factory
   - Verify dependency HCL is included in config
   - Check dependency import is present
   - Verify dependency is registered in init test file

   **Failure: Import fails**
   - Verify ImportState method is implemented
   - Check import step configuration
   - Verify resource can be read after import
   - Check ImportStateVerifyIgnore list

   **Failure: Nested attribute path error**
   - Verify attribute path is correct
   - Check indexing for nested structures
   - Use correct syntax: `nested.0.sub_nested.0.attribute`

   **Failure: Set ordering issue**
   - Use `TestCheckTypeSetElemNestedAttrs` for sets
   - Don't assume specific order for set elements

   **Failure: Nil pointer error**
   - Verify init function is called
   - Check test resources are initialized
   - Verify all dependencies are registered

2. **Debug test failures**
   - Add logging to tests
   - Check Terraform output
   - Verify API responses
   - Review resource implementation

**Deliverable**: All test failures resolved

---

## Phase 11: Review and Approval

### Task 11.1: Review Against Checklist

**Objective**: Verify all requirements are met.

**Steps**:

1. **Use the validation checklist from requirements.md**

   **Test Initialization File**:
   - [ ] File created: `genesyscloud_<resource_name>_init_test.go`
   - [ ] Package-level variables defined
   - [ ] Registration instance struct defined
   - [ ] SDKv2 resource registration implemented
   - [ ] SDKv2 data source registration implemented
   - [ ] Framework resource registration implemented
   - [ ] Framework data source registration implemented
   - [ ] Initialization function implemented
   - [ ] TestMain function implemented

   **Resource Test File**:
   - [ ] File created: `resource_genesyscloud_<resource_name>_test.go`
   - [ ] Package declaration matches directory name
   - [ ] All required imports are present
   - [ ] No unused imports
   - [ ] Init function ensures test resources are initialized

   **Resource Test Cases**:
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

   **Data Source Test File**:
   - [ ] File created: `data_source_genesyscloud_<resource_name>_test.go`
   - [ ] Basic lookup test implemented
   - [ ] Multiple search criteria tests implemented
   - [ ] Tests use `ProtoV6ProviderFactories` or `getFrameworkProviderFactories()`
   - [ ] Test names follow `TestAccFrameworkDataSource*` convention
   - [ ] Init function ensures test resources are initialized

   **Test Helper Functions**:
   - [ ] Basic resource generation helper implemented
   - [ ] Custom attributes helper implemented
   - [ ] Nested block helpers implemented (all levels)
   - [ ] Sub-nested block helpers implemented
   - [ ] Attribute-specific helpers implemented
   - [ ] Data source helper implemented
   - [ ] Destroy verification helper implemented
   - [ ] Provider factory helper implemented (if needed)
   - [ ] Helper functions generate valid HCL
   - [ ] Helper functions handle optional attributes
   - [ ] Helper functions handle nested structures
   - [ ] Helper functions are modular and reusable

   **Test Execution**:
   - [ ] All tests compile without errors
   - [ ] All tests pass successfully
   - [ ] Tests run in isolation
   - [ ] Resources are cleaned up properly
   - [ ] No test flakiness observed
   - [ ] Tests handle API rate limiting
   - [ ] Tests handle eventual consistency

   **Code Quality**:
   - [ ] Tests follow Go conventions
   - [ ] Test names are descriptive
   - [ ] Assertions are clear
   - [ ] Error messages are helpful
   - [ ] Edge cases are documented with TODO comments
   - [ ] No unnecessary TODO or FIXME comments

2. **Fix any issues found**

**Deliverable**: All checklist items verified

---

### Task 11.2: Code Review and Approval

**Objective**: Get peer review and approval before proceeding to Stage 4.

**Steps**:

1. **Create pull request or review request**
   - Include link to Stage 3 requirements and design docs
   - Highlight test coverage
   - Note any deviations from standard pattern
   - Document any known API asymmetries

2. **Address review comments**
   - Make requested changes
   - Re-run tests
   - Re-verify checklist

3. **Get approval**
   - Obtain approval from reviewer
   - Merge or mark as ready for Stage 4

**Deliverable**: Stage 3 approved and ready for Stage 4

---

## Common Issues and Solutions

### Issue 1: Provider Factory Error

**Problem**: Error creating muxed provider.

**Solution**:
- Verify all imports are correct
- Check resource constructor function names
- Ensure SDKv2 resources are correctly referenced
- Verify Framework resources are correctly referenced
- Verify provider.NewMuxedProvider or provider.GetMuxedProviderFactories parameters

### Issue 2: Test Timeout

**Problem**: Test times out waiting for resource.

**Solution**:
- Check API credentials are configured
- Verify resource is actually created
- Increase timeout if needed
- Check for API rate limiting
- Add eventual consistency handling

### Issue 3: Resource Not Destroyed

**Problem**: Destroy verification fails.

**Solution**:
- Verify destroy verification function checks correct resource type
- Check API call is correct
- Ensure 404 is handled correctly
- Verify resource is actually deleted
- Check for dependent resources blocking deletion

### Issue 4: Import Fails

**Problem**: Import step fails.

**Solution**:
- Verify ImportState method is implemented
- Check resource can be read by ID
- Verify all attributes are populated after import
- Check for computed attributes that may differ
- Add attributes to ImportStateVerifyIgnore if needed

### Issue 5: Nested Attribute Path Error

**Problem**: Test assertion fails for nested attribute.

**Solution**:
- Verify attribute path is correct
- Check indexing: `nested.0.sub_nested.0.attribute`
- Verify nested block exists before accessing attributes
- Check for optional nested blocks

### Issue 6: Set Ordering Issue

**Problem**: Test fails due to set element order.

**Solution**:
- Use `TestCheckTypeSetElemNestedAttrs` for sets
- Don't use index-based assertions for sets
- Verify all elements exist without assuming order

### Issue 7: Dependency Not Found

**Problem**: Test fails because dependency resource not found.

**Solution**:
- Add dependency to muxed provider factory
- Verify dependency import is present
- Check dependency HCL is included in config
- Ensure dependency is registered in init test file
- Verify dependency resource type is correct

### Issue 8: Nil Pointer Error

**Problem**: Test fails with nil pointer error.

**Solution**:
- Verify init function is called
- Check test resources are initialized
- Verify all dependencies are registered
- Check TestMain is implemented
- Verify frameworkResources and frameworkDataSources are not nil

### Issue 9: Helper Function Not Generating Valid HCL

**Problem**: Test fails due to invalid HCL syntax.

**Solution**:
- Verify helper function generates valid HCL
- Check for missing quotes, commas, or braces
- Verify attribute names match schema
- Test helper function output manually
- Check for proper indentation

### Issue 10: API Asymmetry Causing Test Failure

**Problem**: Test fails due to API behavior difference.

**Solution**:
- Document the asymmetry with detailed TODO comment
- Explain expected vs actual behavior
- Describe why it worked in SDKv2 but fails in Framework
- List potential resolution approaches
- Comment out failing test step if needed
- Track issue for future resolution

---

## Completion Criteria

Stage 3 is complete when:

- [ ] All tasks in this document are completed
- [ ] All checklist items are verified
- [ ] All tests compile without errors
- [ ] All tests pass successfully
- [ ] Test coverage is verified
- [ ] Known API asymmetries are documented
- [ ] Code review is approved

---

## Next Steps

After Stage 3 completion:

1. **Review and approval**
   - Get team review
   - Address any feedback
   - Get final approval

2. **Proceed to Stage 4**
   - Begin export functionality implementation
   - Create export utilities file
   - Implement export attribute mapping

3. **Reference Stage 4 documentation**
   - Read Stage 4 `requirements.md`
   - Read Stage 4 `design.md`
   - Follow Stage 4 `tasks.md`

---

## Time Estimates

| Phase | Estimated Time |
|-------|----------------|
| Phase 1: Test Initialization File Setup | 2-3 hours |
| Phase 2: Resource Test File Setup | 1-2 hours |
| Phase 3: Basic Helper Functions | 2-3 hours |
| Phase 4: Nested Block Helper Functions | 4-6 hours |
| Phase 5: Basic Resource Test Cases | 2-3 hours |
| Phase 6: Nested Structure Test Cases | 4-6 hours |
| Phase 7: Dependency Test Cases | 2-4 hours |
| Phase 8: Advanced Test Cases | 4-6 hours |
| Phase 9: Data Source Test File | 2-3 hours |
| Phase 10: Test Execution and Validation | 3-5 hours |
| Phase 11: Review and Approval | 1-2 hours |
| **Total** | **27-43 hours** |

*Note: Times vary based on complexity, number of nested structures, and familiarity with patterns.*

---

## References

- **Reference Implementation**: 
  - `genesyscloud/user/resource_genesyscloud_user_test.go`
  - `genesyscloud/user/data_source_genesyscloud_user_test.go`
  - `genesyscloud/user/genesyscloud_user_init_test.go`
- **Stage 3 Requirements**: `prompts/pf_complex_resource_migration/Stage3/requirements.md`
- **Stage 3 Design**: `prompts/pf_complex_resource_migration/Stage3/design.md`
- **Simple Resource Reference**: `prompts/pf_simple_resource_migration/Stage3/tasks.md`
- **Plugin Framework Testing**: https://developer.hashicorp.com/terraform/plugin/framework/acctests
- **Terraform Testing**: https://developer.hashicorp.com/terraform/plugin/testing

