package members

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v116/platformclientv2"

	"terraform-provider-genesyscloud/genesyscloud/consistency_checker"

	gcloud "terraform-provider-genesyscloud/genesyscloud"

	"terraform-provider-genesyscloud/genesyscloud/util/resourcedata"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
)

/*
The resource_genesyscloud_members.go contains all of the methods that perform the core logic for a resource.
*/

// createMembers is used by the members resource to create Genesys cloud members
func createMembers(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*gcloud.ProviderMeta).ClientConfig
	proxy := getMembersProxy(sdkConfig)

	members := getMembersFromResourceData(d)

	log.Printf("Creating members %s", *members.Name)
	teamMemberResponse, err := proxy.createMembers(ctx, teamId, members)
	if err != nil {
		return diag.Errorf("Failed to create members: %s", err)
	}

	d.SetId(*teamMemberResponse.)
	//log.Printf("Created members %s", *teamMemberResponse.Id)
	return readMembers(ctx, d, meta)
}

// readMembers is used by the members resource to read a members from genesys cloud
func readMembers(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*gcloud.ProviderMeta).ClientConfig
	proxy := getMembersProxy(sdkConfig)

	log.Printf("Reading members %s", d.Id())

	return gcloud.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		teamMemberEntityListing, respCode, getErr := proxy.getMembersById(ctx, d.Id())
		if getErr != nil {
			if gcloud.IsStatus404ByInt(respCode) {
				return retry.RetryableError(fmt.Errorf("Failed to read members %s: %s", d.Id(), getErr))
			}
			return retry.NonRetryableError(fmt.Errorf("Failed to read members %s: %s", d.Id(), getErr))
		}

		cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceMembers())

		resourcedata.SetNillableValue(d, "name", userReferenceWithName.Name)

		log.Printf("Read members %s %s", d.Id(), *userReferenceWithName.Name)
		return cc.CheckState()
	})
}

// deleteMembers is used by the members resource to delete a members from Genesys cloud
func deleteMembers(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*gcloud.ProviderMeta).ClientConfig
	proxy := getMembersProxy(sdkConfig)

	_, err := proxy.deleteMembers(ctx, d.Id())
	if err != nil {
		return diag.Errorf("Failed to delete members %s: %s", d.Id(), err)
	}

	return gcloud.WithRetries(ctx, 180*time.Second, func() *retry.RetryError {
		_, respCode, err := proxy.getMembersById(ctx, d.Id())

		if err != nil {
			if gcloud.IsStatus404ByInt(respCode) {
				log.Printf("Deleted members %s", d.Id())
				return nil
			}
			return retry.NonRetryableError(fmt.Errorf("Error deleting members %s: %s", d.Id(), err))
		}

		return retry.RetryableError(fmt.Errorf("members %s still exists", d.Id()))
	})
}

// getMembersFromResourceData maps data from schema ResourceData object to a platformclientv2.Userreferencewithname
func getMembersFromResourceData(d *schema.ResourceData) platformclientv2.Userreferencewithname {
	return platformclientv2.Userreferencewithname{
		Name: platformclientv2.String(d.Get("name").(string)),
	}
}
