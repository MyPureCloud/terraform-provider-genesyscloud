package outbound_wrapupcode_mappings

import (
	"fmt"
	"strconv"
	"testing"

	gcloud "terraform-provider-genesyscloud/genesyscloud"
	lists "terraform-provider-genesyscloud/genesyscloud/util/lists"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccResourceOutboundWrapupCodeMapping(t *testing.T) {

	t.Parallel()
	var (
		resourceId            = "wrapupcodemappings"
		wrapupCode1ResourceId = "wrapupcode1"
		wrapupCode1Name       = "tf test wrapupcode" + uuid.NewString()
		wrapupCode2ResourceId = "wrapupcode2"
		wrapupCode2Name       = "tf test wrapupcode" + uuid.NewString()
		wrapupCode3ResourceId = "wrapupcode3"
		wrapupCode3Name       = "tf test wrapupcode" + uuid.NewString()
	)
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { gcloud.TestAccPreCheck(t) },
		ProviderFactories: gcloud.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: gcloud.GenerateRoutingWrapupcodeResource(wrapupCode1ResourceId, wrapupCode1Name) +
					fmt.Sprintf(`
resource "genesyscloud_outbound_wrapupcodemappings"	"%s" {	
	default_set = ["Number_UnCallable", "Contact_UnCallable"]
	mappings {
		wrapup_code_id = genesyscloud_routing_wrapupcode.%s.id
		flags          = ["Contact_UnCallable"]
	}
}		
`, resourceId, wrapupCode1ResourceId),
				Check: resource.ComposeTestCheckFunc(
					gcloud.ValidateStringInArray("genesyscloud_outbound_wrapupcodemappings."+resourceId, "default_set", "Contact_UnCallable"),
					gcloud.ValidateStringInArray("genesyscloud_outbound_wrapupcodemappings."+resourceId, "default_set", "Number_UnCallable"),
					verifyWrapupCodeMappingsMappingValues("genesyscloud_outbound_wrapupcodemappings."+resourceId,
						"genesyscloud_routing_wrapupcode."+wrapupCode1ResourceId, []string{"Contact_UnCallable"}),
				),
			},
			// Update
			{
				Config: gcloud.GenerateRoutingWrapupcodeResource(wrapupCode1ResourceId, wrapupCode1Name) +
					gcloud.GenerateRoutingWrapupcodeResource(wrapupCode2ResourceId, wrapupCode2Name) +
					fmt.Sprintf(`
resource "genesyscloud_outbound_wrapupcodemappings"	"%s" {
	default_set = ["Right_Party_Contact", "Contact_UnCallable"]
	mappings {
		wrapup_code_id = genesyscloud_routing_wrapupcode.%s.id
		flags          = ["Contact_UnCallable"]
	}
	mappings {
		wrapup_code_id = genesyscloud_routing_wrapupcode.%s.id
		flags          = ["Number_UnCallable", "Right_Party_Contact"]
	}
}		
`, resourceId, wrapupCode1ResourceId, wrapupCode2ResourceId),
				Check: resource.ComposeTestCheckFunc(
					gcloud.ValidateStringInArray("genesyscloud_outbound_wrapupcodemappings."+resourceId, "default_set", "Contact_UnCallable"),
					gcloud.ValidateStringInArray("genesyscloud_outbound_wrapupcodemappings."+resourceId, "default_set", "Right_Party_Contact"),
					verifyWrapupCodeMappingsMappingValues("genesyscloud_outbound_wrapupcodemappings."+resourceId,
						"genesyscloud_routing_wrapupcode."+wrapupCode1ResourceId, []string{"Contact_UnCallable"}),
					verifyWrapupCodeMappingsMappingValues("genesyscloud_outbound_wrapupcodemappings."+resourceId,
						"genesyscloud_routing_wrapupcode."+wrapupCode2ResourceId, []string{"Number_UnCallable", "Right_Party_Contact"}),
				),
			},
			// Update
			{
				Config: gcloud.GenerateRoutingWrapupcodeResource(wrapupCode1ResourceId, wrapupCode1Name) +
					gcloud.GenerateRoutingWrapupcodeResource(wrapupCode2ResourceId, wrapupCode2Name) +
					gcloud.GenerateRoutingWrapupcodeResource(wrapupCode3ResourceId, wrapupCode3Name) +
					fmt.Sprintf(`
resource "genesyscloud_outbound_wrapupcodemappings"	"%s" {
	default_set = ["Right_Party_Contact", "Number_UnCallable", "Contact_UnCallable"]
	mappings {
		wrapup_code_id = genesyscloud_routing_wrapupcode.%s.id
		flags          = ["Contact_UnCallable"]
	}
	mappings {
		wrapup_code_id = genesyscloud_routing_wrapupcode.%s.id
		flags          = ["Number_UnCallable", "Right_Party_Contact"]
	}
	mappings {
		wrapup_code_id = genesyscloud_routing_wrapupcode.%s.id
		flags          = ["Number_UnCallable", "Contact_UnCallable", "Right_Party_Contact"]
	}
}		
`, resourceId, wrapupCode1ResourceId, wrapupCode2ResourceId, wrapupCode3ResourceId),
				Check: resource.ComposeTestCheckFunc(
					gcloud.ValidateStringInArray("genesyscloud_outbound_wrapupcodemappings."+resourceId, "default_set", "Contact_UnCallable"),
					gcloud.ValidateStringInArray("genesyscloud_outbound_wrapupcodemappings."+resourceId, "default_set", "Number_UnCallable"),
					gcloud.ValidateStringInArray("genesyscloud_outbound_wrapupcodemappings."+resourceId, "default_set", "Right_Party_Contact"),
					verifyWrapupCodeMappingsMappingValues("genesyscloud_outbound_wrapupcodemappings."+resourceId,
						"genesyscloud_routing_wrapupcode."+wrapupCode1ResourceId, []string{"Contact_UnCallable"}),
					verifyWrapupCodeMappingsMappingValues("genesyscloud_outbound_wrapupcodemappings."+resourceId,
						"genesyscloud_routing_wrapupcode."+wrapupCode2ResourceId, []string{"Number_UnCallable", "Right_Party_Contact"}),
					verifyWrapupCodeMappingsMappingValues("genesyscloud_outbound_wrapupcodemappings."+resourceId,
						"genesyscloud_routing_wrapupcode."+wrapupCode3ResourceId, []string{"Number_UnCallable", "Right_Party_Contact", "Contact_UnCallable"}),
				),
			},
			{
				ResourceName:            "genesyscloud_outbound_wrapupcodemappings." + resourceId,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"default_set", "mappings", "placeholder"},
			},
		},
	})
}

// The mappings field is built using a map before being sent to the API, meaning the ordering of keys is changed,
// which makes it difficult to test values in the mappings list. This function loops through the state to find the correct
// item before testing attribute values
func verifyWrapupCodeMappingsMappingValues(resourceName string, wrapupCodeResourceName string, expectedFlags []string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		resourceState, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Failed to find resourceState %s in state", resourceName)
		}

		wrapupCodeResourceState, ok := state.RootModule().Resources[wrapupCodeResourceName]
		if !ok {
			return fmt.Errorf("Failed to find resourceState %s in state", resourceName)
		}
		expectedWrapupCodeId := wrapupCodeResourceState.Primary.ID

		numAttr, ok := resourceState.Primary.Attributes["mappings.#"]
		if !ok {
			return fmt.Errorf("No mappings found for %s in state", resourceState.Primary.ID)
		}

		numValues, _ := strconv.Atoi(numAttr)
		for i := 0; i < numValues; i++ {
			if resourceState.Primary.Attributes["mappings."+strconv.Itoa(i)+".wrapup_code_id"] == expectedWrapupCodeId {
				numFlagsStr := resourceState.Primary.Attributes["mappings."+strconv.Itoa(i)+".flags.#"]
				numFlags, _ := strconv.Atoi(numFlagsStr)
				flagsList := make([]string, 0)
				for j := 0; j < numFlags; j++ {
					flagsList = append(flagsList, resourceState.Primary.Attributes[fmt.Sprintf("mappings.%v.flags.%v", strconv.Itoa(i), strconv.Itoa(j))])
				}
				if !lists.AreEquivalent(flagsList, expectedFlags) {
					return fmt.Errorf("mismatch for field %v. Expected: %v, got: %v\n", fmt.Sprintf("mappings.%v.flags", strconv.Itoa(i)), expectedFlags, flagsList)
				}
				return nil
			}
		}
		return fmt.Errorf("Could not find wrapupcode with id %s in state", expectedWrapupCodeId)
	}
}
