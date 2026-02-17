package routing_wrapupcode

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
				// Create
				Config: generateFrameworkRoutingWrapupcodeResource(resourceLabel, name, util.NullValue, description),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_routing_wrapupcode."+resourceLabel, "name", name),
					resource.TestCheckResourceAttr("genesyscloud_routing_wrapupcode."+resourceLabel, "description", description),
					resource.TestCheckResourceAttrSet("genesyscloud_routing_wrapupcode."+resourceLabel, "id"),
				),
			},
			{
				// Import/Read
				ResourceName:      "genesyscloud_routing_wrapupcode." + resourceLabel,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
		CheckDestroy: testVerifyFrameworkWrapupcodesDestroyed,
	})
}

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
				// Create with division
				Config: authDivision.GenerateAuthDivisionBasic(divResourceLabel, divName) +
					generateFrameworkRoutingWrapupcodeResource(resourceLabel, name, "genesyscloud_auth_division."+divResourceLabel+".id", description),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_routing_wrapupcode."+resourceLabel, "name", name),
					resource.TestCheckResourceAttr("genesyscloud_routing_wrapupcode."+resourceLabel, "description", description),
					resource.TestCheckResourceAttrPair("genesyscloud_routing_wrapupcode."+resourceLabel, "division_id", "genesyscloud_auth_division."+divResourceLabel, "id"),
					resource.TestCheckResourceAttrSet("genesyscloud_routing_wrapupcode."+resourceLabel, "id"),
				),
			},
			{
				// Import/Read
				ResourceName:      "genesyscloud_routing_wrapupcode." + resourceLabel,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
		CheckDestroy: testVerifyFrameworkWrapupcodesDestroyed,
	})
}

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
				// Create
				Config: generateFrameworkRoutingWrapupcodeResource(resourceLabel, name1, util.NullValue, description),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_routing_wrapupcode."+resourceLabel, "name", name1),
					resource.TestCheckResourceAttr("genesyscloud_routing_wrapupcode."+resourceLabel, "description", description),
					resource.TestCheckResourceAttrSet("genesyscloud_routing_wrapupcode."+resourceLabel, "id"),
				),
			},
			{
				// Update name (should be in-place update, not replacement)
				Config: generateFrameworkRoutingWrapupcodeResource(resourceLabel, name2, util.NullValue, description),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_routing_wrapupcode."+resourceLabel, "name", name2),
					resource.TestCheckResourceAttr("genesyscloud_routing_wrapupcode."+resourceLabel, "description", description),
					resource.TestCheckResourceAttrSet("genesyscloud_routing_wrapupcode."+resourceLabel, "id"),
				),
			},
		},
		CheckDestroy: testVerifyFrameworkWrapupcodesDestroyed,
	})
}

func TestAccFrameworkResourceRoutingWrapupcodeDescriptionUpdate(t *testing.T) {
	var (
		resourceLabel = "test_routing_wrapupcode_desc_update"
		name          = "Terraform Framework Wrapupcode " + uuid.NewString()
		description1  = "Test wrapupcode description 1"
		description2  = "Test wrapupcode description 2"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { util.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: getFrameworkProviderFactories(),
		Steps: []resource.TestStep{
			{
				// Create
				Config: generateFrameworkRoutingWrapupcodeResource(resourceLabel, name, util.NullValue, description1),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_routing_wrapupcode."+resourceLabel, "name", name),
					resource.TestCheckResourceAttr("genesyscloud_routing_wrapupcode."+resourceLabel, "description", description1),
					resource.TestCheckResourceAttrSet("genesyscloud_routing_wrapupcode."+resourceLabel, "id"),
				),
			},
			{
				// Update description (should be in-place update)
				Config: generateFrameworkRoutingWrapupcodeResource(resourceLabel, name, util.NullValue, description2),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_routing_wrapupcode."+resourceLabel, "name", name),
					resource.TestCheckResourceAttr("genesyscloud_routing_wrapupcode."+resourceLabel, "description", description2),
					resource.TestCheckResourceAttrSet("genesyscloud_routing_wrapupcode."+resourceLabel, "id"),
				),
			},
		},
		CheckDestroy: testVerifyFrameworkWrapupcodesDestroyed,
	})
}

func TestAccFrameworkResourceRoutingWrapupcodeLifecycle(t *testing.T) {
	var (
		resourceLabel    = "test_routing_wrapupcode_lifecycle"
		name1            = "Terraform Framework Wrapupcode " + uuid.NewString()
		name2            = "Terraform Framework Wrapupcode Updated " + uuid.NewString()
		description1     = "Test wrapupcode lifecycle 1"
		description2     = "Test wrapupcode lifecycle 2"
		divResourceLabel = "test_division"
		divName          = "terraform-" + uuid.NewString()
	)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { util.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: getFrameworkProviderFactories(),
		Steps: []resource.TestStep{
			{
				// Create without division
				Config: generateFrameworkRoutingWrapupcodeResource(resourceLabel, name1, util.NullValue, description1),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_routing_wrapupcode."+resourceLabel, "name", name1),
					resource.TestCheckResourceAttr("genesyscloud_routing_wrapupcode."+resourceLabel, "description", description1),
					resource.TestCheckResourceAttrSet("genesyscloud_routing_wrapupcode."+resourceLabel, "id"),
				),
			},
			{
				// Create with division
				Config: authDivision.GenerateAuthDivisionBasic(divResourceLabel, divName) +
					generateFrameworkRoutingWrapupcodeResource(resourceLabel, name1, "genesyscloud_auth_division."+divResourceLabel+".id", description1),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_routing_wrapupcode."+resourceLabel, "name", name1),
					resource.TestCheckResourceAttr("genesyscloud_routing_wrapupcode."+resourceLabel, "description", description1),
					resource.TestCheckResourceAttrPair("genesyscloud_routing_wrapupcode."+resourceLabel, "division_id", "genesyscloud_auth_division."+divResourceLabel, "id"),
					resource.TestCheckResourceAttrSet("genesyscloud_routing_wrapupcode."+resourceLabel, "id"),
				),
			},
			{
				// Update name
				Config: authDivision.GenerateAuthDivisionBasic(divResourceLabel, divName) +
					generateFrameworkRoutingWrapupcodeResource(resourceLabel, name2, "genesyscloud_auth_division."+divResourceLabel+".id", description1),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_routing_wrapupcode."+resourceLabel, "name", name2),
					resource.TestCheckResourceAttr("genesyscloud_routing_wrapupcode."+resourceLabel, "description", description1),
					resource.TestCheckResourceAttrPair("genesyscloud_routing_wrapupcode."+resourceLabel, "division_id", "genesyscloud_auth_division."+divResourceLabel, "id"),
					resource.TestCheckResourceAttrSet("genesyscloud_routing_wrapupcode."+resourceLabel, "id"),
				),
			},
			{
				// Update description
				Config: authDivision.GenerateAuthDivisionBasic(divResourceLabel, divName) +
					generateFrameworkRoutingWrapupcodeResource(resourceLabel, name2, "genesyscloud_auth_division."+divResourceLabel+".id", description2),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_routing_wrapupcode."+resourceLabel, "name", name2),
					resource.TestCheckResourceAttr("genesyscloud_routing_wrapupcode."+resourceLabel, "description", description2),
					resource.TestCheckResourceAttrPair("genesyscloud_routing_wrapupcode."+resourceLabel, "division_id", "genesyscloud_auth_division."+divResourceLabel, "id"),
					resource.TestCheckResourceAttrSet("genesyscloud_routing_wrapupcode."+resourceLabel, "id"),
				),
			},
			{
				// Import/Read
				ResourceName:      "genesyscloud_routing_wrapupcode." + resourceLabel,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
		CheckDestroy: testVerifyFrameworkWrapupcodesDestroyed,
	})
}

// testVerifyFrameworkWrapupcodesDestroyed checks that all routing wrapupcodes have been destroyed
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
			// Wrapupcode not found as expected
			continue
		} else {
			// Unexpected error
			return fmt.Errorf("Unexpected error checking Framework routing wrapupcode: %s", err)
		}
	}
	// Success. All wrapupcodes destroyed
	return nil
}

// generateFrameworkRoutingWrapupcodeResource generates a routing wrapupcode resource for Framework testing
func generateFrameworkRoutingWrapupcodeResource(resourceLabel string, name string, divisionId string, description string) string {
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

// getFrameworkProviderFactories returns provider factories for Framework testing
func getFrameworkProviderFactories() map[string]func() (tfprotov6.ProviderServer, error) {
	return map[string]func() (tfprotov6.ProviderServer, error){
		"genesyscloud": func() (tfprotov6.ProviderServer, error) {
			// Create Framework provider with routing_wrapupcode resource and data source
			frameworkResources := map[string]func() frameworkresource.Resource{
				ResourceType: NewRoutingWrapupcodeFrameworkResource,
			}
			frameworkDataSources := map[string]func() datasource.DataSource{
				ResourceType: NewRoutingWrapupcodeFrameworkDataSource,
			}

			// Create muxed provider that includes both Framework and SDKv2 resources
			// This allows the test to use auth_division (SDKv2) alongside routing_wrapupcode (Framework)
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
