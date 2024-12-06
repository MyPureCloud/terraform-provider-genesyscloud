package routing_queue

import (
	"fmt"
	"strconv"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceRoutingQueueBasic(t *testing.T) {
	var (
		queueResourceLabel = "test-queue"
		queueName          = "Terraform Test Queue-" + uuid.NewString()
		queueDesc          = "This is a test"

		queueDataSourceLabel = "test-queue-ds"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create
				Config: GenerateRoutingQueueResource(
					queueResourceLabel,
					queueName,
					queueDesc,
					util.NullValue, // MANDATORY_TIMEOUT
					"200000",       // acw_timeout
					util.NullValue, // ALL
					util.NullValue, // auto_answer_only true
					util.NullValue, // No calling party name
					util.NullValue, // No calling party number
					util.NullValue, // enable_audio_monitoring false
					util.NullValue, // enable_manual_assignment false
					util.NullValue, //suppressCall_record_false
					util.NullValue, // enable_transcription false
					strconv.Quote("TimestampAndPriority"),
					util.NullValue,
					util.NullValue,
				) + generateRoutingQueueDataSource(
					queueDataSourceLabel,
					"genesyscloud_routing_queue."+queueResourceLabel+".name",
					"genesyscloud_routing_queue."+queueResourceLabel,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.genesyscloud_routing_queue."+queueDataSourceLabel,
						"id", "genesyscloud_routing_queue."+queueResourceLabel, "id",
					),
				),
			},
		},
	})
}

func TestAccDataSourceRoutingQueueCaching(t *testing.T) {
	var (
		queue1ResourceLabel = "queue1"
		queueName1          = "terraform test queue " + uuid.NewString()
		queue2ResourceLabel = "queue2"
		queueName2          = "terraform test queue " + uuid.NewString()
		queue3ResourceLabel = "queue3"
		queueName3          = "terraform test queue " + uuid.NewString()

		dataSourceLabel1 = "data-1"
		dataSourceLabel2 = "data-2"
		dataSourceLabel3 = "data-3"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					time.Sleep(45 * time.Second)
				},
				Config: generateRoutingQueueResourceBasic( // queue resource
					queue1ResourceLabel,
					queueName1,
				) + generateRoutingQueueResourceBasic( // queue resource
					queue2ResourceLabel,
					queueName2,
				) + generateRoutingQueueResourceBasic( // queue resource
					queue3ResourceLabel,
					queueName3,
				) + generateRoutingQueueDataSource( // queue data source
					dataSourceLabel1,
					strconv.Quote(queueName1),
					"genesyscloud_routing_queue."+queue1ResourceLabel,
				) + generateRoutingQueueDataSource( // queue data source
					dataSourceLabel2,
					strconv.Quote(queueName2),
					"genesyscloud_routing_queue."+queue2ResourceLabel,
				) + generateRoutingQueueDataSource( // queue data source
					dataSourceLabel3,
					strconv.Quote(queueName3),
					"genesyscloud_routing_queue."+queue3ResourceLabel,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("genesyscloud_routing_queue."+queue1ResourceLabel, "id",
						"data.genesyscloud_routing_queue."+dataSourceLabel1, "id"),
					resource.TestCheckResourceAttrPair("genesyscloud_routing_queue."+queue2ResourceLabel, "id",
						"data.genesyscloud_routing_queue."+dataSourceLabel2, "id"),
					resource.TestCheckResourceAttrPair("genesyscloud_routing_queue."+queue3ResourceLabel, "id",
						"data.genesyscloud_routing_queue."+dataSourceLabel3, "id"),
				),
			},
		},
		CheckDestroy: testVerifyQueuesDestroyed,
	})
}

func generateRoutingQueueDataSource(
	resourceLabel string,
	name string,
	// Must explicitly use depends_on in terraform v0.13 when a data source references a resource
	// Fixed in v0.14 https://github.com/hashicorp/terraform/pull/26284
	dependsOnResource string) string {
	return fmt.Sprintf(`data "genesyscloud_routing_queue" "%s" {
		name = %s
		depends_on=[%s]
	}
	`, resourceLabel, name, dependsOnResource)
}
