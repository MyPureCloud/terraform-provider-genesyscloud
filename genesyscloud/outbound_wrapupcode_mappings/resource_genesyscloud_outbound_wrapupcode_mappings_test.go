package outbound_wrapupcode_mappings

import (
	"fmt"
	authDivision "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/auth_division"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	routingWrapupcode "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/routing_wrapupcode"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
	lists "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/lists"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccResourceOutboundWrapupCodeMapping(t *testing.T) {
	var (
		resourceLabel            = "wrapupcodemappings"
		wrapupCode1ResourceLabel = "wrapupcode1"
		wrapupCode1Name          = "tf test wrapupcode" + uuid.NewString()
		wrapupCode2ResourceLabel = "wrapupcode2"
		wrapupCode2Name          = "tf test wrapupcode" + uuid.NewString()
		wrapupCode3ResourceLabel = "wrapupcode3"
		wrapupCode3Name          = "tf test wrapupcode" + uuid.NewString()
		divResourceLabel         = "test-division"
		divName                  = "terraform-" + uuid.NewString()
		description              = "Terraform test description"
	)
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					time.Sleep(30 * time.Second)
				},
				Config: authDivision.GenerateAuthDivisionBasic(divResourceLabel, divName) + routingWrapupcode.GenerateRoutingWrapupcodeResource(
					wrapupCode1ResourceLabel,
					wrapupCode1Name,
					"genesyscloud_auth_division."+divResourceLabel+".id",
					description,
				) +
					fmt.Sprintf(`
resource "genesyscloud_outbound_wrapupcodemappings"	"%s" {
	default_set = ["Number_UnCallable", "Contact_UnCallable"]
	mappings {
		wrapup_code_id = genesyscloud_routing_wrapupcode.%s.id
		flags          = ["Contact_UnCallable"]
	}
}
`, resourceLabel, wrapupCode1ResourceLabel),
				Check: resource.ComposeTestCheckFunc(
					func(s *terraform.State) error {
						time.Sleep(30 * time.Second) // Wait for 45 seconds to get proper response
						return nil
					},
					util.ValidateStringInArray("genesyscloud_outbound_wrapupcodemappings."+resourceLabel, "default_set", "Contact_UnCallable"),
					util.ValidateStringInArray("genesyscloud_outbound_wrapupcodemappings."+resourceLabel, "default_set", "Number_UnCallable"),
					verifyWrapupCodeMappingsMappingValues("genesyscloud_outbound_wrapupcodemappings."+resourceLabel,
						"genesyscloud_routing_wrapupcode."+wrapupCode1ResourceLabel, []string{"Contact_UnCallable"}),
				),
			},
			// Update
			{
				PreConfig: func() {
					time.Sleep(30 * time.Second)
				},
				Config: authDivision.GenerateAuthDivisionBasic(divResourceLabel, divName) +
					routingWrapupcode.GenerateRoutingWrapupcodeResource(wrapupCode1ResourceLabel, wrapupCode1Name, "genesyscloud_auth_division."+divResourceLabel+".id", description) +
					routingWrapupcode.GenerateRoutingWrapupcodeResource(wrapupCode2ResourceLabel, wrapupCode2Name, "genesyscloud_auth_division."+divResourceLabel+".id", description) +
					fmt.Sprintf(`
resource "genesyscloud_outbound_wrapupcodemappings"	"%s" {
	default_set = ["Right_Party_Contact", "Contact_UnCallable", "Business_Success"]
	mappings {
		wrapup_code_id = genesyscloud_routing_wrapupcode.%s.id
		flags          = ["Contact_UnCallable"]
	}
	mappings {
		wrapup_code_id = genesyscloud_routing_wrapupcode.%s.id
		flags          = ["Number_UnCallable", "Right_Party_Contact", "Business_Failure"]
	}
}
`, resourceLabel, wrapupCode1ResourceLabel, wrapupCode2ResourceLabel),
				Check: resource.ComposeTestCheckFunc(
					func(s *terraform.State) error {
						time.Sleep(45 * time.Second) // Wait for 45 seconds to get proper response
						return nil
					},
					util.ValidateStringInArray("genesyscloud_outbound_wrapupcodemappings."+resourceLabel, "default_set", "Contact_UnCallable"),
					util.ValidateStringInArray("genesyscloud_outbound_wrapupcodemappings."+resourceLabel, "default_set", "Right_Party_Contact"),
					util.ValidateStringInArray("genesyscloud_outbound_wrapupcodemappings."+resourceLabel, "default_set", "Business_Success"),
					verifyWrapupCodeMappingsMappingValues("genesyscloud_outbound_wrapupcodemappings."+resourceLabel,
						"genesyscloud_routing_wrapupcode."+wrapupCode1ResourceLabel, []string{"Contact_UnCallable"}),
					verifyWrapupCodeMappingsMappingValues("genesyscloud_outbound_wrapupcodemappings."+resourceLabel,
						"genesyscloud_routing_wrapupcode."+wrapupCode2ResourceLabel, []string{"Number_UnCallable", "Right_Party_Contact", "Business_Failure"}),
				),
			},
			// Update
			{
				PreConfig: func() {
					time.Sleep(30 * time.Second)
				},
				Config: authDivision.GenerateAuthDivisionBasic(divResourceLabel, divName) +
					routingWrapupcode.GenerateRoutingWrapupcodeResource(wrapupCode1ResourceLabel, wrapupCode1Name, "genesyscloud_auth_division."+divResourceLabel+".id", description) +
					routingWrapupcode.GenerateRoutingWrapupcodeResource(wrapupCode2ResourceLabel, wrapupCode2Name, "genesyscloud_auth_division."+divResourceLabel+".id", description) +
					routingWrapupcode.GenerateRoutingWrapupcodeResource(wrapupCode3ResourceLabel, wrapupCode3Name, "genesyscloud_auth_division."+divResourceLabel+".id", description) +
					fmt.Sprintf(`
resource "genesyscloud_outbound_wrapupcodemappings"	"%s" {
	default_set = ["Right_Party_Contact", "Number_UnCallable", "Contact_UnCallable", "Business_Neutral"]
	mappings {
		wrapup_code_id = genesyscloud_routing_wrapupcode.%s.id
		flags          = ["Contact_UnCallable"]
	}
	mappings {
		wrapup_code_id = genesyscloud_routing_wrapupcode.%s.id
		flags          = ["Number_UnCallable", "Right_Party_Contact", "Business_Neutral"]
	}
	mappings {
		wrapup_code_id = genesyscloud_routing_wrapupcode.%s.id
		flags          = ["Number_UnCallable", "Contact_UnCallable", "Right_Party_Contact", "Business_Success"]
	}
}
`, resourceLabel, wrapupCode1ResourceLabel, wrapupCode2ResourceLabel, wrapupCode3ResourceLabel),
				Check: resource.ComposeTestCheckFunc(
					func(s *terraform.State) error {
						time.Sleep(45 * time.Second) // Wait for 45 seconds to get proper response
						return nil
					},
					util.ValidateStringInArray("genesyscloud_outbound_wrapupcodemappings."+resourceLabel, "default_set", "Contact_UnCallable"),
					util.ValidateStringInArray("genesyscloud_outbound_wrapupcodemappings."+resourceLabel, "default_set", "Number_UnCallable"),
					util.ValidateStringInArray("genesyscloud_outbound_wrapupcodemappings."+resourceLabel, "default_set", "Right_Party_Contact"),
					util.ValidateStringInArray("genesyscloud_outbound_wrapupcodemappings."+resourceLabel, "default_set", "Business_Neutral"),
					verifyWrapupCodeMappingsMappingValues("genesyscloud_outbound_wrapupcodemappings."+resourceLabel,
						"genesyscloud_routing_wrapupcode."+wrapupCode1ResourceLabel, []string{"Contact_UnCallable"}),
					verifyWrapupCodeMappingsMappingValues("genesyscloud_outbound_wrapupcodemappings."+resourceLabel,
						"genesyscloud_routing_wrapupcode."+wrapupCode2ResourceLabel, []string{"Number_UnCallable", "Right_Party_Contact", "Business_Neutral"}),
					verifyWrapupCodeMappingsMappingValues("genesyscloud_outbound_wrapupcodemappings."+resourceLabel,
						"genesyscloud_routing_wrapupcode."+wrapupCode3ResourceLabel, []string{"Number_UnCallable", "Contact_UnCallable", "Right_Party_Contact", "Business_Success"}),
				),
			},
			{
				ResourceName:            "genesyscloud_outbound_wrapupcodemappings." + resourceLabel,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"default_set", "mappings", "placeholder"},
			},
		},
	})
}

// verifyWrapupCodeMappingsMappingValues checks that the mapping attribute has the correct flags set.
func verifyWrapupCodeMappingsMappingValues(resourcePath string, wrapupCodeResourcePath string, expectedFlags []string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		resourceState, ok := state.RootModule().Resources[resourcePath]
		if !ok {
			return fmt.Errorf("Failed to find resourceState %s in state", resourcePath)
		}

		wrapupCodeResourceState, ok := state.RootModule().Resources[wrapupCodeResourcePath]
		if !ok {
			return fmt.Errorf("Failed to find resourceState %s in state", resourcePath)
		}
		expectedWrapupCodeId := wrapupCodeResourceState.Primary.ID

		_, ok = resourceState.Primary.Attributes["mappings.#"]
		if !ok {
			return fmt.Errorf("No mappings found for %s in state", resourceState.Primary.ID)
		}

		// Since we're dealing with a set, we'll keep track of the special index
		// in the state that correspond to mappings attribute
		mapSetIndices := make([]string, 0)
		for stateKey := range resourceState.Primary.Attributes {
			stateKeyParts := strings.Split(stateKey, ".")
			if len(stateKeyParts) < 3 || stateKeyParts[0] != "mappings" {
				continue
			}
			mapSetIndices = append(mapSetIndices, stateKeyParts[1])
		}

		for _, msi := range mapSetIndices {
			wucId := resourceState.Primary.Attributes["mappings."+msi+".wrapup_code_id"]
			flagsList := attributeFlagsToList("mappings."+msi+".flags", resourceState.Primary.Attributes)
			if wucId == expectedWrapupCodeId {
				if !lists.AreEquivalent(flagsList, expectedFlags) {
					return fmt.Errorf("mismatch for field %v. Expected: %v, got: %v\n", fmt.Sprintf("mappings.%v.flags", msi), expectedFlags, flagsList)
				}
				return nil
			}
		}
		return fmt.Errorf("Could not find wrapupcode with id %s in state", expectedWrapupCodeId)
	}
}

// attributeFlagsToList return a list of the mapping flags from the tf state attributes
func attributeFlagsToList(flagPrefix string, attr map[string]string) []string {
	const maxFlags = 4
	ret := make([]string, 0)
	for i := 0; i < maxFlags; i++ {
		if flagVal, ok := attr[flagPrefix+"."+strconv.Itoa(i)]; ok {
			ret = append(ret, flagVal)
		}
	}
	return ret
}
