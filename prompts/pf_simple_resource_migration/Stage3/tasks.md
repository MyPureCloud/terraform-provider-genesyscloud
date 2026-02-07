# Stage 3 – Test Migration Tasks

## Overview

This document provides step-by-step tasks for completing Stage 3 of the Plugin Framework migration. Follow these tasks in order to migrate acceptance tests from SDKv2 to Plugin Framework patterns.

**Reference Implementation**: 
- `genesyscloud/routing_wrapupcode/resource_genesyscloud_routing_wrapupcode_test.go`
- `genesyscloud/routing_wrapupcode/data_source_genesyscloud_routing_wrapupcode_test.go`
- `genesyscloud/routing_wrapupcode/genesyscloud_routing_wrapupcode_init_test.go`

**Estimated Time**: 4-8 hours (depending on test complexity)

---

## Prerequisites

Before starting Stage 3 tasks, ensure:

- [ ] Stage 1 (Schema Migration) is complete and approved
- [ ] Stage 2 (Resource Migration) is complete and approved
- [ ] You have reviewed the existing SDKv2 test implementation
- [ ] You understand the muxed provider pattern
- [ ] You have read Stage 3 `requirements.md` and `design.md`
- [ ] You have studied the `routing_wrapupcode` reference implementation
- [ ] Test environment is configured (credentials, etc.)

---

## Task Checklist

### Phase 1: Resource Test File Setup
- [ ] Task 1.1: Create Resource Test File
- [ ] Task 1.2: Add Package Declaration and Imports
- [ ] Task 1.3: Implement Provider Factory
- [ ] Task 1.4: Implement HCL Generation Helper
- [ ] Task 1.5: Implement Destroy Verification Function

### Phase 2: Resource Test Cases
- [ ] Task 2.1: Implement Basic CRUD Test
- [ ] Task 2.2: Implement Division/Dependency Test
- [ ] Task 2.3: Implement Name Update Test
- [ ] Task 2.4: Implement Description Update Test
- [ ] Task 2.5: Implement Lifecycle Test

### Phase 3: Data Source Test File
- [ ] Task 3.1: Create Data Source Test File
- [ ] Task 3.2: Add Package Declaration and Imports
- [ ] Task 3.3: Implement Data Source HCL Helper
- [ ] Task 3.4: Implement Basic Data Source Test
- [ ] Task 3.5: Implement Data Source with Dependencies Test

### Phase 4: Test Initialization File
- [ ] Task 4.1: Create Test Initialization File
- [ ] Task 4.2: Add Package-Level Setup (if needed)

### Phase 5: Test Execution and Validation
- [ ] Task 5.1: Run Individual Tests
- [ ] Task 5.2: Run All Tests
- [ ] Task 5.3: Verify Test Coverage
- [ ] Task 5.4: Fix Any Test Failures

### Phase 6: Review and Approval
- [ ] Task 6.1: Review Against Checklist
- [ ] Task 6.2: Code Review and Approval

---

## Detailed Tasks

## Phase 1: Resource Test File Setup

### Task 1.1: Create Resource Test File

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
   Example:
   ```powershell
   New-Item -ItemType File -Name "resource_genesyscloud_routing_wrapupcode_test.go"
   ```

**Deliverable**: Empty test file created

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

3. **Adjust imports based on dependencies**
   - Add imports for other dependency packages if needed
   - Remove unused imports

**Deliverable**: File with package declaration and imports

---

### Task 1.3: Implement Provider Factory

**Objective**: Create muxed provider factory for tests.

**Steps**:

1. **Add provider factory function**
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
               // This allows the test to use <dependency> (SDKv2) alongside <resource> (Framework)
               muxFactory := provider.NewMuxedProvider(
                   "test",
                   map[string]*schema.Resource{
                       authDivision.ResourceType: authDivision.ResourceAuthDivision(),
                       // Add other SDKv2 dependencies here
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

2. **Replace placeholders**
   - `<resource>` → Your resource name
   - `<ResourceName>` → Your resource name in PascalCase
   - Add all SDKv2 dependencies your tests need

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

**Deliverable**: Provider factory function implemented

---

### Task 1.4: Implement HCL Generation Helper

**Objective**: Create helper function to generate Terraform HCL for tests.

**Steps**:

1. **Add HCL generation function**
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

2. **Customize for your resource**
   - Add parameters for all attributes
   - Handle optional attributes with null checks
   - Use `util.NullValue` for reference attributes
   - Use empty string check for literal attributes

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

**Deliverable**: HCL generation helper implemented

---

### Task 1.5: Implement Destroy Verification Function

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

2. **Customize for your resource**
   - Use correct API client
   - Use correct resource type string
   - Use correct API method to check existence

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

**Deliverable**: Destroy verification function implemented

---

## Phase 2: Resource Test Cases

### Task 2.1: Implement Basic CRUD Test

**Objective**: Test basic create, read, import, and destroy operations.

**Steps**:

1. **Add test function**
   ```go
   func TestAccFrameworkResource<ResourceName>Basic(t *testing.T) {
       var (
           resourceLabel = "test_<resource>"
           name          = "Terraform Framework <Resource> " + uuid.NewString()
           description   = "Test <resource> description"
       )

       resource.Test(t, resource.TestCase{
           PreCheck:                 func() { util.TestAccPreCheck(t) },
           ProtoV6ProviderFactories: getFrameworkProviderFactories(),
           Steps: []resource.TestStep{
               {
                   // Create
                   Config: generateFramework<ResourceName>Resource(resourceLabel, name, util.NullValue, description),
                   Check: resource.ComposeTestCheckFunc(
                       resource.TestCheckResourceAttr("genesyscloud_<resource>."+resourceLabel, "name", name),
                       resource.TestCheckResourceAttr("genesyscloud_<resource>."+resourceLabel, "description", description),
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

**Deliverable**: Basic CRUD test implemented

---

### Task 2.2: Implement Division/Dependency Test

**Objective**: Test resource with dependencies (e.g., division).

**Steps**:

1. **Add test function** (if resource has dependencies)
   ```go
   func TestAccFrameworkResource<ResourceName>Division(t *testing.T) {
       var (
           resourceLabel    = "test_<resource>_division"
           name             = "Terraform Framework <Resource> " + uuid.NewString()
           description      = "Test <resource> with division"
           divResourceLabel = "test_division"
           divName          = "terraform-" + uuid.NewString()
       )

       resource.Test(t, resource.TestCase{
           PreCheck:                 func() { util.TestAccPreCheck(t) },
           ProtoV6ProviderFactories: getFrameworkProviderFactories(),
           Steps: []resource.TestStep{
               {
                   // Create with division
                   Config: authDivision.GenerateAuthDivisionBasic(divResourceLabel, divName) +
                       generateFramework<ResourceName>Resource(resourceLabel, name, "genesyscloud_auth_division."+divResourceLabel+".id", description),
                   Check: resource.ComposeTestCheckFunc(
                       resource.TestCheckResourceAttr("genesyscloud_<resource>."+resourceLabel, "name", name),
                       resource.TestCheckResourceAttr("genesyscloud_<resource>."+resourceLabel, "description", description),
                       resource.TestCheckResourceAttrPair("genesyscloud_<resource>."+resourceLabel, "division_id", "genesyscloud_auth_division."+divResourceLabel, "id"),
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

2. **Key points**
   - Use `TestCheckResourceAttrPair` to verify dependency reference
   - Include dependency resource HCL in config
   - Ensure dependency is in muxed provider factory

**Deliverable**: Division/dependency test implemented (if applicable)

---

### Task 2.3: Implement Name Update Test

**Objective**: Test in-place name updates.

**Steps**:

1. **Add test function**
   ```go
   func TestAccFrameworkResource<ResourceName>NameUpdate(t *testing.T) {
       var (
           resourceLabel = "test_<resource>_name_update"
           name1         = "Terraform Framework <Resource> " + uuid.NewString()
           name2         = "Terraform Framework <Resource> Updated " + uuid.NewString()
           description   = "Test <resource> name update"
       )

       resource.Test(t, resource.TestCase{
           PreCheck:                 func() { util.TestAccPreCheck(t) },
           ProtoV6ProviderFactories: getFrameworkProviderFactories(),
           Steps: []resource.TestStep{
               {
                   // Create
                   Config: generateFramework<ResourceName>Resource(resourceLabel, name1, util.NullValue, description),
                   Check: resource.ComposeTestCheckFunc(
                       resource.TestCheckResourceAttr("genesyscloud_<resource>."+resourceLabel, "name", name1),
                       resource.TestCheckResourceAttr("genesyscloud_<resource>."+resourceLabel, "description", description),
                       resource.TestCheckResourceAttrSet("genesyscloud_<resource>."+resourceLabel, "id"),
                   ),
               },
               {
                   // Update name (should be in-place update, not replacement)
                   Config: generateFramework<ResourceName>Resource(resourceLabel, name2, util.NullValue, description),
                   Check: resource.ComposeTestCheckFunc(
                       resource.TestCheckResourceAttr("genesyscloud_<resource>."+resourceLabel, "name", name2),
                       resource.TestCheckResourceAttr("genesyscloud_<resource>."+resourceLabel, "description", description),
                       resource.TestCheckResourceAttrSet("genesyscloud_<resource>."+resourceLabel, "id"),
                   ),
               },
           },
           CheckDestroy: testVerifyFramework<ResourceName>Destroyed,
       })
   }
   ```

**Deliverable**: Name update test implemented

---

### Task 2.4: Implement Description Update Test

**Objective**: Test in-place description updates.

**Steps**:

1. **Add test function**
   ```go
   func TestAccFrameworkResource<ResourceName>DescriptionUpdate(t *testing.T) {
       var (
           resourceLabel = "test_<resource>_desc_update"
           name          = "Terraform Framework <Resource> " + uuid.NewString()
           description1  = "Test <resource> description 1"
           description2  = "Test <resource> description 2"
       )

       resource.Test(t, resource.TestCase{
           PreCheck:                 func() { util.TestAccPreCheck(t) },
           ProtoV6ProviderFactories: getFrameworkProviderFactories(),
           Steps: []resource.TestStep{
               {
                   // Create
                   Config: generateFramework<ResourceName>Resource(resourceLabel, name, util.NullValue, description1),
                   Check: resource.ComposeTestCheckFunc(
                       resource.TestCheckResourceAttr("genesyscloud_<resource>."+resourceLabel, "name", name),
                       resource.TestCheckResourceAttr("genesyscloud_<resource>."+resourceLabel, "description", description1),
                       resource.TestCheckResourceAttrSet("genesyscloud_<resource>."+resourceLabel, "id"),
                   ),
               },
               {
                   // Update description (should be in-place update)
                   Config: generateFramework<ResourceName>Resource(resourceLabel, name, util.NullValue, description2),
                   Check: resource.ComposeTestCheckFunc(
                       resource.TestCheckResourceAttr("genesyscloud_<resource>."+resourceLabel, "name", name),
                       resource.TestCheckResourceAttr("genesyscloud_<resource>."+resourceLabel, "description", description2),
                       resource.TestCheckResourceAttrSet("genesyscloud_<resource>."+resourceLabel, "id"),
                   ),
               },
           },
           CheckDestroy: testVerifyFramework<ResourceName>Destroyed,
       })
   }
   ```

**Deliverable**: Description update test implemented

---

### Task 2.5: Implement Lifecycle Test

**Objective**: Comprehensive test covering multiple scenarios.

**Steps**:

1. **Add comprehensive lifecycle test**
   ```go
   func TestAccFrameworkResource<ResourceName>Lifecycle(t *testing.T) {
       var (
           resourceLabel    = "test_<resource>_lifecycle"
           name1            = "Terraform Framework <Resource> " + uuid.NewString()
           name2            = "Terraform Framework <Resource> Updated " + uuid.NewString()
           description1     = "Test <resource> lifecycle 1"
           description2     = "Test <resource> lifecycle 2"
           divResourceLabel = "test_division"
           divName          = "terraform-" + uuid.NewString()
       )

       resource.Test(t, resource.TestCase{
           PreCheck:                 func() { util.TestAccPreCheck(t) },
           ProtoV6ProviderFactories: getFrameworkProviderFactories(),
           Steps: []resource.TestStep{
               {
                   // Create without division
                   Config: generateFramework<ResourceName>Resource(resourceLabel, name1, util.NullValue, description1),
                   Check: resource.ComposeTestCheckFunc(
                       resource.TestCheckResourceAttr("genesyscloud_<resource>."+resourceLabel, "name", name1),
                       resource.TestCheckResourceAttr("genesyscloud_<resource>."+resourceLabel, "description", description1),
                       resource.TestCheckResourceAttrSet("genesyscloud_<resource>."+resourceLabel, "id"),
                   ),
               },
               {
                   // Create with division
                   Config: authDivision.GenerateAuthDivisionBasic(divResourceLabel, divName) +
                       generateFramework<ResourceName>Resource(resourceLabel, name1, "genesyscloud_auth_division."+divResourceLabel+".id", description1),
                   Check: resource.ComposeTestCheckFunc(
                       resource.TestCheckResourceAttr("genesyscloud_<resource>."+resourceLabel, "name", name1),
                       resource.TestCheckResourceAttr("genesyscloud_<resource>."+resourceLabel, "description", description1),
                       resource.TestCheckResourceAttrPair("genesyscloud_<resource>."+resourceLabel, "division_id", "genesyscloud_auth_division."+divResourceLabel, "id"),
                       resource.TestCheckResourceAttrSet("genesyscloud_<resource>."+resourceLabel, "id"),
                   ),
               },
               {
                   // Update name
                   Config: authDivision.GenerateAuthDivisionBasic(divResourceLabel, divName) +
                       generateFramework<ResourceName>Resource(resourceLabel, name2, "genesyscloud_auth_division."+divResourceLabel+".id", description1),
                   Check: resource.ComposeTestCheckFunc(
                       resource.TestCheckResourceAttr("genesyscloud_<resource>."+resourceLabel, "name", name2),
                       resource.TestCheckResourceAttr("genesyscloud_<resource>."+resourceLabel, "description", description1),
                       resource.TestCheckResourceAttrPair("genesyscloud_<resource>."+resourceLabel, "division_id", "genesyscloud_auth_division."+divResourceLabel, "id"),
                       resource.TestCheckResourceAttrSet("genesyscloud_<resource>."+resourceLabel, "id"),
                   ),
               },
               {
                   // Update description
                   Config: authDivision.GenerateAuthDivisionBasic(divResourceLabel, divName) +
                       generateFramework<ResourceName>Resource(resourceLabel, name2, "genesyscloud_auth_division."+divResourceLabel+".id", description2),
                   Check: resource.ComposeTestCheckFunc(
                       resource.TestCheckResourceAttr("genesyscloud_<resource>."+resourceLabel, "name", name2),
                       resource.TestCheckResourceAttr("genesyscloud_<resource>."+resourceLabel, "description", description2),
                       resource.TestCheckResourceAttrPair("genesyscloud_<resource>."+resourceLabel, "division_id", "genesyscloud_auth_division."+divResourceLabel, "id"),
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

**Deliverable**: Lifecycle test implemented

---

## Phase 3: Data Source Test File

### Task 3.1: Create Data Source Test File

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

### Task 3.2: Add Package Declaration and Imports

**Objective**: Set up the file with correct package and imports.

**Steps**:

1. **Add package declaration and imports**
   ```go
   package <resource_name>

   import (
       "fmt"
       "testing"

       "github.com/google/uuid"
       "github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
       authDivision "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/auth_division"
       "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
   )
   ```

**Deliverable**: File with package declaration and imports

---

### Task 3.3: Implement Data Source HCL Helper

**Objective**: Create helper function to generate data source HCL.

**Steps**:

1. **Add data source HCL generation function**
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

**Deliverable**: Data source HCL helper implemented

---

### Task 3.4: Implement Basic Data Source Test

**Objective**: Test basic data source lookup.

**Steps**:

1. **Add test function**
   ```go
   func TestAccFrameworkDataSource<ResourceName>(t *testing.T) {
       var (
           resourceLabel = "test_<resource>"
           dataLabel     = "test_data_<resource>"
           name          = "Terraform Framework Data <Resource> " + uuid.NewString()
           description   = "Test <resource> for data source"
       )

       resource.Test(t, resource.TestCase{
           PreCheck:                 func() { util.TestAccPreCheck(t) },
           ProtoV6ProviderFactories: getFrameworkProviderFactories(),
           Steps: []resource.TestStep{
               {
                   Config: generateFramework<ResourceName>Resource(resourceLabel, name, util.NullValue, description) +
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

**Deliverable**: Basic data source test implemented

---

### Task 3.5: Implement Data Source with Dependencies Test

**Objective**: Test data source lookup with dependencies.

**Steps**:

1. **Add test function** (if applicable)
   ```go
   func TestAccFrameworkDataSource<ResourceName>WithDivision(t *testing.T) {
       var (
           resourceLabel    = "test_<resource>_div"
           dataLabel        = "test_data_<resource>_div"
           name             = "Terraform Framework Data <Resource> Div " + uuid.NewString()
           description      = "Test <resource> with division for data source"
           divResourceLabel = "test_division"
           divName          = "terraform-" + uuid.NewString()
       )

       resource.Test(t, resource.TestCase{
           PreCheck:                 func() { util.TestAccPreCheck(t) },
           ProtoV6ProviderFactories: getFrameworkProviderFactories(),
           Steps: []resource.TestStep{
               {
                   Config: authDivision.GenerateAuthDivisionBasic(divResourceLabel, divName) +
                       generateFramework<ResourceName>Resource(resourceLabel, name, "genesyscloud_auth_division."+divResourceLabel+".id", description) +
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

**Deliverable**: Data source with dependencies test implemented (if applicable)

---

## Phase 4: Test Initialization File

### Task 4.1: Create Test Initialization File

**Objective**: Create test initialization file for package-level setup.

**Steps**:

1. **Navigate to resource package directory**
   ```powershell
   cd genesyscloud\<resource_name>
   ```

2. **Create the init test file**
   ```powershell
   New-Item -ItemType File -Name "genesyscloud_<resource_name>_init_test.go"
   ```

**Deliverable**: Empty init test file created

---

### Task 4.2: Add Package-Level Setup (if needed)

**Objective**: Add package-level test setup if needed.

**Steps**:

1. **Add minimal package declaration**
   ```go
   package <resource_name>

   import (
       "testing"
   )

   // Package-level test setup can go here if needed
   // Most simple resources don't need additional setup
   ```

2. **Add setup code only if needed**
   - Most simple resources don't need package-level setup
   - Add only if you have shared test fixtures or utilities

**Deliverable**: Init test file with minimal setup

---

## Phase 5: Test Execution and Validation

### Task 5.1: Run Individual Tests

**Objective**: Run each test individually to verify they pass.

**Steps**:

1. **Run basic test**
   ```powershell
   go test -v -run TestAccFrameworkResource<ResourceName>Basic ./genesyscloud/<resource_name>
   ```

2. **Run division test**
   ```powershell
   go test -v -run TestAccFrameworkResource<ResourceName>Division ./genesyscloud/<resource_name>
   ```

3. **Run update tests**
   ```powershell
   go test -v -run TestAccFrameworkResource<ResourceName>NameUpdate ./genesyscloud/<resource_name>
   go test -v -run TestAccFrameworkResource<ResourceName>DescriptionUpdate ./genesyscloud/<resource_name>
   ```

4. **Run lifecycle test**
   ```powershell
   go test -v -run TestAccFrameworkResource<ResourceName>Lifecycle ./genesyscloud/<resource_name>
   ```

5. **Run data source tests**
   ```powershell
   go test -v -run TestAccFrameworkDataSource<ResourceName> ./genesyscloud/<resource_name>
   ```

6. **Fix any failures**
   - Review error messages
   - Check test configuration
   - Verify resource implementation
   - Update tests as needed

**Deliverable**: All individual tests pass

---

### Task 5.2: Run All Tests

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

### Task 5.3: Verify Test Coverage

**Objective**: Ensure test coverage matches or exceeds SDKv2 version.

**Steps**:

1. **Review test coverage**
   - [ ] Basic CRUD operations tested
   - [ ] Division/dependency assignment tested (if applicable)
   - [ ] Name updates tested
   - [ ] Description updates tested
   - [ ] Lifecycle scenarios tested
   - [ ] Data source lookup tested
   - [ ] Import functionality tested
   - [ ] Destroy verification tested

2. **Compare with SDKv2 tests**
   - Verify all SDKv2 test cases are migrated
   - Identify any missing test cases
   - Add additional tests if needed

**Deliverable**: Test coverage verified and complete

---

### Task 5.4: Fix Any Test Failures

**Objective**: Address any test failures or issues.

**Steps**:

1. **Common test failures and solutions**:

   **Failure: Provider not found**
   - Check provider factory is correctly implemented
   - Verify resource type is registered in factory
   - Ensure imports are correct

   **Failure: Resource not destroyed**
   - Check destroy verification function
   - Verify API call is correct
   - Check for resource type filter

   **Failure: Attribute mismatch**
   - Verify HCL generation is correct
   - Check attribute names match schema
   - Verify test assertions are correct

   **Failure: Dependency not found**
   - Ensure dependency is in muxed provider factory
   - Verify dependency HCL is included in config
   - Check dependency import is present

   **Failure: Import fails**
   - Verify ImportState method is implemented
   - Check import step configuration
   - Verify resource can be read after import

2. **Debug test failures**
   - Add logging to tests
   - Check Terraform output
   - Verify API responses
   - Review resource implementation

**Deliverable**: All test failures resolved

---

## Phase 6: Review and Approval

### Task 6.1: Review Against Checklist

**Objective**: Verify all requirements are met.

**Steps**:

1. **Use the validation checklist from requirements.md**

   **Resource Test File**:
   - [ ] File created: `resource_genesyscloud_<resource_name>_test.go`
   - [ ] Package declaration matches directory name
   - [ ] All required imports are present
   - [ ] No unused imports

   **Resource Test Cases**:
   - [ ] Basic CRUD test implemented
   - [ ] Division/dependency test implemented (if applicable)
   - [ ] Name update test implemented
   - [ ] Description update test implemented
   - [ ] Lifecycle test implemented
   - [ ] All tests include import verification
   - [ ] All tests use `ProtoV6ProviderFactories`
   - [ ] Test names follow `TestAccFrameworkResource*` convention

   **Data Source Test File**:
   - [ ] File created: `data_source_genesyscloud_<resource_name>_test.go`
   - [ ] Basic lookup test implemented
   - [ ] Lookup with dependencies test implemented (if applicable)
   - [ ] Tests use `ProtoV6ProviderFactories`
   - [ ] Test names follow `TestAccFrameworkDataSource*` convention

   **Test Initialization File**:
   - [ ] File created: `genesyscloud_<resource_name>_init_test.go`
   - [ ] Package-level setup implemented (if needed)

   **Test Helper Functions**:
   - [ ] `generateFramework<ResourceName>Resource()` implemented
   - [ ] `generateFramework<ResourceName>DataSource()` implemented
   - [ ] `testVerifyFramework<ResourceName>Destroyed()` implemented
   - [ ] `getFrameworkProviderFactories()` implemented
   - [ ] Helper functions generate valid HCL
   - [ ] Helper functions handle optional attributes

   **Muxed Provider Factory**:
   - [ ] Factory creates muxed provider
   - [ ] Framework resource included
   - [ ] SDKv2 dependencies included
   - [ ] Returns `ProtoV6ProviderServer`
   - [ ] Factory is reusable across tests

   **Test Execution**:
   - [ ] All tests compile without errors
   - [ ] All tests pass successfully
   - [ ] Tests run in isolation
   - [ ] Resources are cleaned up properly
   - [ ] No test flakiness observed

   **Code Quality**:
   - [ ] Tests follow Go conventions
   - [ ] Test names are descriptive
   - [ ] Assertions are clear
   - [ ] Error messages are helpful
   - [ ] No TODO or FIXME comments (unless intentional)

2. **Fix any issues found**

**Deliverable**: All checklist items verified

---

### Task 6.2: Code Review and Approval

**Objective**: Get peer review and approval before proceeding to Stage 4.

**Steps**:

1. **Create pull request or review request**
   - Include link to Stage 3 requirements and design docs
   - Highlight test coverage
   - Note any deviations from standard pattern

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
- Verify provider.NewMuxedProvider parameters

### Issue 2: Test Timeout

**Problem**: Test times out waiting for resource.

**Solution**:
- Check API credentials are configured
- Verify resource is actually created
- Increase timeout if needed
- Check for API rate limiting

### Issue 3: Resource Not Destroyed

**Problem**: Destroy verification fails.

**Solution**:
- Verify destroy verification function checks correct resource type
- Check API call is correct
- Ensure 404 is handled correctly
- Verify resource is actually deleted

### Issue 4: Import Fails

**Problem**: Import step fails.

**Solution**:
- Verify ImportState method is implemented
- Check resource can be read by ID
- Verify all attributes are populated after import
- Check for computed attributes that may differ

### Issue 5: Attribute Mismatch

**Problem**: Test assertion fails for attribute value.

**Solution**:
- Verify HCL generation is correct
- Check attribute name matches schema
- Verify expected value is correct
- Check for type conversion issues

### Issue 6: Dependency Not Found

**Problem**: Test fails because dependency resource not found.

**Solution**:
- Add dependency to muxed provider factory
- Verify dependency import is present
- Check dependency HCL is included in config
- Ensure dependency is created before resource

---

## Completion Criteria

Stage 3 is complete when:

- [ ] All tasks in this document are completed
- [ ] All checklist items are verified
- [ ] All tests compile without errors
- [ ] All tests pass successfully
- [ ] Test coverage is verified
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
| Phase 1: Resource Test File Setup | 1-2 hours |
| Phase 2: Resource Test Cases | 2-3 hours |
| Phase 3: Data Source Test File | 1-2 hours |
| Phase 4: Test Initialization File | 15-30 minutes |
| Phase 5: Test Execution and Validation | 1-2 hours |
| Phase 6: Review and Approval | 30-60 minutes |
| **Total** | **5-10 hours** |

*Note: Times vary based on test complexity and familiarity with patterns.*

---

## References

- **Reference Implementation**: 
  - `genesyscloud/routing_wrapupcode/resource_genesyscloud_routing_wrapupcode_test.go`
  - `genesyscloud/routing_wrapupcode/data_source_genesyscloud_routing_wrapupcode_test.go`
  - `genesyscloud/routing_wrapupcode/genesyscloud_routing_wrapupcode_init_test.go`
- **Stage 3 Requirements**: `prompts/pf_simple_resource_migration/Stage3/requirements.md`
- **Stage 3 Design**: `prompts/pf_simple_resource_migration/Stage3/design.md`
- **Plugin Framework Testing**: https://developer.hashicorp.com/terraform/plugin/framework/acctests
- **Terraform Testing**: https://developer.hashicorp.com/terraform/plugin/testing
