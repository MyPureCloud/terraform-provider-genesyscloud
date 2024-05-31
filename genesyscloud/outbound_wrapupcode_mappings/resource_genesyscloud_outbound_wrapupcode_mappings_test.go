package outbound_wrapupcode_mappings

import (
	"fmt"
	"strconv"
	"strings"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"testing"
	"time"

	gcloud "terraform-provider-genesyscloud/genesyscloud"
	lists "terraform-provider-genesyscloud/genesyscloud/util/lists"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccResourceOutboundWrapupCodeMapping(t *testing.T) {
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
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					time.Sleep(30 * time.Second)
				},
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
					func(s *terraform.State) error {
						time.Sleep(30 * time.Second) // Wait for 45 seconds to get proper response
						return nil
					},
					util.ValidateStringInArray("genesyscloud_outbound_wrapupcodemappings."+resourceId, "default_set", "Contact_UnCallable"),
					util.ValidateStringInArray("genesyscloud_outbound_wrapupcodemappings."+resourceId, "default_set", "Number_UnCallable"),
					verifyWrapupCodeMappingsMappingValues("genesyscloud_outbound_wrapupcodemappings."+resourceId,
						"genesyscloud_routing_wrapupcode."+wrapupCode1ResourceId, []string{"Contact_UnCallable"}),
				),
			},
			// Update
			{
				PreConfig: func() {
					time.Sleep(30 * time.Second)
				},
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
					util.ValidateStringInArray("genesyscloud_outbound_wrapupcodemappings."+resourceId, "default_set", "Contact_UnCallable"),
					util.ValidateStringInArray("genesyscloud_outbound_wrapupcodemappings."+resourceId, "default_set", "Right_Party_Contact"),
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
					util.ValidateStringInArray("genesyscloud_outbound_wrapupcodemappings."+resourceId, "default_set", "Contact_UnCallable"),
					util.ValidateStringInArray("genesyscloud_outbound_wrapupcodemappings."+resourceId, "default_set", "Number_UnCallable"),
					util.ValidateStringInArray("genesyscloud_outbound_wrapupcodemappings."+resourceId, "default_set", "Right_Party_Contact"),
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

// verifyWrapupCodeMappingsMappingValues checks that the mapping attribute has the correct flags set.
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
	const maxFlags = 3
	ret := make([]string, 0)
	for i := 0; i < maxFlags; i++ {
		if flagVal, ok := attr[flagPrefix+"."+strconv.Itoa(i)]; ok {
			ret = append(ret, flagVal)
		}
	}
	return ret
}
