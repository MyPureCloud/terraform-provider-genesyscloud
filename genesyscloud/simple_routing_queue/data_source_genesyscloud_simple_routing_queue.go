package simple_routing_queue

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"log"
	gcloud "terraform-provider-genesyscloud/genesyscloud"
	"time"
)

/*
   The data_source_genesyscloud_simple_routing_queue.go contains the data source implementation
   for the resource.
*/

// dataSourceSimpleRoutingQueueRead retrieves by search term the id in question
func dataSourceSimpleRoutingQueueRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// Get an instance of our proxy

	// Grab our queue name from the schema.ResourceData object
	name := d.Get("name").(string)

	log.Printf("Finding queue by name '%s'", name)
	return gcloud.WithRetries(ctx, 15*time.Second, func() *resource.RetryError {

		// Call to the proxy function getRoutingQueueIdByName(context.Context, string)
		// This function returns values in the following order: queueId (string), retryable (bool), err (error)

		// If the error is not nil, and retryable equals false, return a resource.NonRetryableError
		// letting the user know that an error occurred

		// If retryable equals true, return a resource.RetryableError and let them know the queue could not be found
		// with that name

		// If we made it this far, we can set the queue ID in the schema.ResourceData object, and return nil
		return nil
	})
}
