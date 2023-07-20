package genesyscloud

import (
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceRoutingQueueBasic(t *testing.T) {
	var (
		queueResource = "test-queue"
		queueName     = "Terraform Test Queue-" + uuid.NewString()
		queueDesc     = "This is a test"

		queueDataSource = "test-queue-ds"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { TestAccPreCheck(t) },
		ProviderFactories: GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create
				Config: GenerateRoutingQueueResource(
					queueResource,
					queueName,
					queueDesc,
					nullValue, // MANDATORY_TIMEOUT
					"200000",  // acw_timeout
					nullValue, // ALL
					nullValue, // auto_answer_only true
					nullValue, // No calling party name
					nullValue, // No calling party number
					nullValue, // enable_manual_assignment false
					nullValue, // enable_transcription false
				) + generateRoutingQueueDataSource(
					queueDataSource,
					"genesyscloud_routing_queue."+queueResource+".name",
					"genesyscloud_routing_queue."+queueResource,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.genesyscloud_routing_queue."+queueDataSource,
						"id", "genesyscloud_routing_queue."+queueResource, "id",
					),
				),
			},
		},
	})
}

func generateRoutingQueueDataSource(
	resourceID string,
	name string,
	// Must explicitly use depends_on in terraform v0.13 when a data source references a resource
	// Fixed in v0.14 https://github.com/hashicorp/terraform/pull/26284
	dependsOnResource string) string {
	return fmt.Sprintf(`data "genesyscloud_routing_queue" "%s" {
		name = %s
		depends_on=[%s]
	}
	`, resourceID, name, dependsOnResource)
}
