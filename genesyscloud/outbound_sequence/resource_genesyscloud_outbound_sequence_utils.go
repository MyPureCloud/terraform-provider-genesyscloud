package outbound_sequence

import (
	"fmt"
	"strings"
	"terraform-provider-genesyscloud/genesyscloud/util"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"
)

/*
The resource_genesyscloud_outbound_sequence.go contains all of the methods that perform the core logic for a resource.
*/

// getOutboundSequenceFromResourceData maps data from schema ResourceData object to a platformclientv2.Campaignsequence
func getOutboundSequenceFromResourceData(d *schema.ResourceData) platformclientv2.Campaignsequence {
	return platformclientv2.Campaignsequence{
		Name:      platformclientv2.String(d.Get("name").(string)),
		Campaigns: util.BuildSdkDomainEntityRefArr(d, "campaign_ids"),
		Status:    platformclientv2.String("off"), // This will be updated separately
		Repeat:    platformclientv2.Bool(d.Get("repeat").(bool)),
	}
}

func GenerateOutboundSequence(
	resourceId string,
	name string,
	campaignIds []string,
	status string,
	repeat string) string {
	return fmt.Sprintf(`
		resource "genesyscloud_outbound_sequence" "%s" {
			name = "%s"
			campaign_ids = [%s]
			status = %s
			repeat = %s
		}
	`, resourceId, name, strings.Join(campaignIds, ", "), status, repeat)
}
