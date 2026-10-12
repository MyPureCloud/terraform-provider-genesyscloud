package integration

import (
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

/*
Test Class for the Webhook Integrations Data Source

Note: The webhookId and invocationUrl are extracted from the integration's attributes field,
not from the config properties. When a webhook integration is created, these attributes
are automatically populated by the Genesys Cloud system.
*/
func TestAccDataSourceIntegrationWebhook(t *testing.T) {

	var (
		inteResourceLabel1 = "test_webhook_integration1"
		inteResourceLabel2 = "test_webhook_integration2"
		inteName1          = "Terraform Webhook Integration Test-" + uuid.NewString()
		typeID             = "webhook"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Step 1: Create the webhook integration only.
				// The webhookId/invocationUrl attributes are populated asynchronously by
				// Genesys Cloud after the integration is created. Creating the integration
				// in its own step gives those attributes time to be generated before the
				// data source (which has a short internal retry window) reads them.
				Config: GenerateIntegrationResource(
					inteResourceLabel1,
					util.NullValue, //Empty intended_state, default value is "DISABLED"
					strconv.Quote(typeID),
					GenerateIntegrationConfig(
						strconv.Quote(inteName1),
						util.NullValue, //Empty notes
						"",             //Empty credential ID
						util.NullValue, //Empty properties
						util.NullValue, //Empty advanced JSON
					),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("genesyscloud_integration."+inteResourceLabel1, "id"),
				),
			},
			{
				// Step 2: Read the webhook integration through the data source after the
				// asynchronously-populated attributes have had time to appear. The PreConfig
				// sleep runs before this apply, so the webhookId is available by the time the
				// data source's internal retry loop runs.
				PreConfig: func() { time.Sleep(30 * time.Second) },
				Config: GenerateIntegrationResource(
					inteResourceLabel1,
					util.NullValue, //Empty intended_state, default value is "DISABLED"
					strconv.Quote(typeID),
					GenerateIntegrationConfig(
						strconv.Quote(inteName1),
						util.NullValue, //Empty notes
						"",             //Empty credential ID
						util.NullValue, //Empty properties
						util.NullValue, //Empty advanced JSON
					),
				) + generateIntegrationWebhookDataSource(inteResourceLabel2,
					inteName1,
					"genesyscloud_integration."+inteResourceLabel1),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.genesyscloud_integration_webhook."+inteResourceLabel2, "id", "genesyscloud_integration."+inteResourceLabel1, "id"),
					// Note: web_hook_id and invocation_url will be populated from the integration's attributes
					// by the Genesys Cloud system when the webhook integration is created
					resource.TestCheckResourceAttrSet("data.genesyscloud_integration_webhook."+inteResourceLabel2, "web_hook_id"),
					resource.TestCheckResourceAttrSet("data.genesyscloud_integration_webhook."+inteResourceLabel2, "invocation_url"),
				),
			},
		},
	})

}

func generateIntegrationWebhookDataSource(
	resourceLabel string,
	name string,
	// Must explicitly use depends_on in terraform v0.13 when a data source references a resource
	// Fixed in v0.14 https://github.com/hashicorp/terraform/pull/26284
	dependsOnResource string) string {
	return fmt.Sprintf(`data "genesyscloud_integration_webhook" "%s" {
		name = "%s"
		depends_on=[%s]
	}
	`, resourceLabel, name, dependsOnResource)
}
