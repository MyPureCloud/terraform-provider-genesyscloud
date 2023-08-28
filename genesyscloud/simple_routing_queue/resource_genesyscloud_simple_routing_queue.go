package simple_routing_queue

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"log"
	gcloud "terraform-provider-genesyscloud/genesyscloud"
	"terraform-provider-genesyscloud/genesyscloud/consistency_checker"
	"time"
)

/*
The resource_genesyscloud_simple_routing_queue.go contains all of the methods that perform the core logic for a resource.
In general a resource should have a approximately 5 methods in it:

1.  A create.... function that the resource will use to create a Genesys Cloud object (e.g. genesyscloud_simple_routing_queue)
2.  A read.... function that looks up a single resource.
3.  An update... function that updates a single resource.
4.  A delete.... function that deletes a single resource.

Two things to note:

1.  All code in these methods should be focused on getting data in and out of Terraform.  All code that is used for interacting
    with a Genesys API should be encapsulated into a proxy class contained within the package.

2.  In general, to keep this file somewhat manageable, if you find yourself with a number of helper functions move them to a
utils file in the package.  This will keep the code manageable and easy to work through.
*/

// createSimpleRoutingQueue is used by the genesyscloud_simple_routing_queue resource to create a simple queue in Genesys cloud.
func createSimpleRoutingQueue(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// Get an instance of the proxy (example can be found in the delete method below)

	// Create variables for each field in our schema.ResourceData object
	// Example:
	name := d.Get("name").(string)

	log.Printf("Creating simple queue %s", name)

	// Create a queue struct using the Genesys Cloud platform go sdk

	// Call the proxy function to create our queue
	// Note: We won't need the response object returned. Also, don't forget about error handling!

	// Set ID in the schema.ResourceData object

	return readSimpleRoutingQueue(ctx, d, meta)
}

// readSimpleRoutingQueue is used by the genesyscloud_simple_routing_queue resource to read a simple queue from Genesys cloud.
func readSimpleRoutingQueue(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// Get an instance of the proxy

	log.Printf("Reading simple queue %s", d.Id())
	return gcloud.WithRetriesForRead(ctx, d, func() *resource.RetryError {
		// Call the read queue function to find our queue, passing in the ID from the resource data ( d.Id() )
		// The returned value are: Queue, Status Code (int), error
		// If the error is not nil, we should pass the status code to the function gcloud.IsStatus404ByInt(int)
		// If the status code is 404, return a resource.RetryableError. Otherwise, it should be a NonRetryableError

		// Define consistency checker
		cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceSimpleRoutingQueue())

		// Set our values in the schema resource data, based on the values
		// in the Queue object returned from the API

		return cc.CheckState()
	})
}

// updateSimpleRoutingQueue is used by the genesyscloud_simple_routing_queue resource to update a simple queue in Genesys cloud.
func updateSimpleRoutingQueue(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Printf("Updating simple queue %s", d.Id())
	// Get an instance of the proxy

	// Create a queue struct using the Genesys Cloud platform go sdk

	// Call the proxy function to update our queue, passing in the struct and the queue ID

	return readSimpleRoutingQueue(ctx, d, meta)
}

// deleteSimpleRoutingQueue is used by the genesyscloud_simple_routing_queue resource to delete a simple queue from Genesys cloud.
func deleteSimpleRoutingQueue(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// Get an instance of the proxy
	sdkConfig := meta.(*gcloud.ProviderMeta).ClientConfig
	proxy := getSimpleRoutingQueueProxy(sdkConfig)

	// Call the delete queue proxy function, passing in our queue ID from the schema.ResourceData object
	// Again, we won't be needing the returned response object

	log.Printf("Deleting simple queue %s", d.Id())
	// Check that queue has been deleted by trying to get it from the API
	return gcloud.WithRetries(ctx, 30*time.Second, func() *resource.RetryError {
		_, respCode, err := proxy.getRoutingQueue(ctx, d.Id())

		if err == nil {
			return resource.NonRetryableError(fmt.Errorf("error deleting routing queue %s: %s", d.Id(), err))
		}
		if gcloud.IsStatus404ByInt(respCode) {
			// Success: Routing Queue deleted
			log.Printf("Deleted routing queue %s", d.Id())
			return nil
		}

		return resource.RetryableError(fmt.Errorf("routing queue %s still exists", d.Id()))
	})
}
