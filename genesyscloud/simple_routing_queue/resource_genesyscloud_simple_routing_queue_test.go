package simple_routing_queue

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"strconv"
	gcloud "terraform-provider-genesyscloud/genesyscloud"
	"testing"
)

func TestAccResourceSimpleRoutingQueue(t *testing.T) {
	var (
		resourceId          = "queue"
		name                = "Create 2023 Queue " + uuid.NewString()
		callingPartyName    = "Example Inc."
		enableTranscription = "true"

		fullResourcePath = resourceName + "." + resourceId
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { gcloud.TestAccPreCheck(t) },
		ProviderFactories: gcloud.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: generateSimpleRoutingQueueResource(
					resourceId,
					name,
					strconv.Quote(callingPartyName),
					enableTranscription,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fullResourcePath, "name", name),
					resource.TestCheckResourceAttr(fullResourcePath, "calling_party_name", callingPartyName),
					resource.TestCheckResourceAttr(fullResourcePath, "enable_transcription", "true"),
				),
			},
		},
	})
}

func generateSimpleRoutingQueueResource(resourceId, name, callingPartyName, enableTranscription string) string {
	return fmt.Sprintf(`
resource "genesyscloud_simple_routing_queue" "%s" {
	name                 = "%s"
    calling_party_name   = %s
	enable_transcription = %s
}
`, resourceId, name, callingPartyName, enableTranscription)
}
