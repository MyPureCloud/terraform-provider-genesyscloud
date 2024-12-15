package conversations_messaging_integrations_whatsapp

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	gcloud "terraform-provider-genesyscloud/genesyscloud"

	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v146/platformclientv2"
)

/*
The resource_genesyscloud_conversations_messaging_integrations_whatsapp_test.go contains all of the test cases for running the resource
tests for conversations_messaging_integrations_whatsapp.
*/

func TestAccResourceConversationsMessagingIntegrationsWhatsapp(t *testing.T) {
	t.Parallel()
	var ()

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { gcloud.TestAccPreCheck(t) },
		ProviderFactories: gcloud.GetProviderFactories(providerResources, providerDataSources),
		Steps:             []resource.TestStep{},
		CheckDestroy:      testVerifyConversationsMessagingIntegrationsWhatsappDestroyed,
	})
}

func testVerifyConversationsMessagingIntegrationsWhatsappDestroyed(state *terraform.State) error {
	return nil
}
