package teams_resource

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v105/platformclientv2"
	"log"
	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	"time"

	"terraform-provider-genesyscloud/genesyscloud/consistency_checker"

	gcloud "terraform-provider-genesyscloud/genesyscloud"

	"terraform-provider-genesyscloud/genesyscloud/util/resourcedata"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
)

/*
The resource_genesyscloud_teams_resource.go contains all of the methods that perform the core logic for a resource.
*/

// getAllAuthTeamsResource retrieves all of the teams resource via Terraform in the Genesys Cloud and is used for the exporter
func getAllAuthTeamsResources(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	proxy := newTeamsResourceProxy(clientConfig)
	resources := make(resourceExporter.ResourceIDMetaMap)

	teams, err := proxy.getAllTeamsResource(ctx)
	if err != nil {
		return nil, diag.Errorf("Failed to get ruleset: %v", err)
	}

	for _, team := range *teams {
		log.Printf("Dealing with teams resource id : %s", *team.Id)
		resources[*team.Id] = &resourceExporter.ResourceMeta{Name: *team.Id}
	}

	return resources, nil
}

// createTeamsResource is used by the teams_resource resource to create Genesys cloud teams resource
func createTeamsResource(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*gcloud.ProviderMeta).ClientConfig
	proxy := newTeamsResourceProxy(sdkConfig)

	teamsResource := getTeamsResourceFromResourceData(d)

	log.Printf("Creating teams resource %s", *teamsResource.Name)
	team, err := proxy.createTeamsResource(ctx, &teamsResource)
	if err != nil {
		return diag.Errorf("Failed to create teams resource: %s", err)
	}

	d.SetId(*team.Id)
	log.Printf("Created teams resource %s", *team.Id)
	return readTeamsResource(ctx, d, meta)
}

// readTeamsResource is used by the teams_resource resource to read an teams resource from genesys cloud
func readTeamsResource(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*gcloud.ProviderMeta).ClientConfig
	proxy := newTeamsResourceProxy(sdkConfig)

	log.Printf("Reading teams resource %s", d.Id())

	return gcloud.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		team, respCode, getErr := proxy.getTeamsResourceById(ctx, d.Id())
		if getErr != nil {
			if gcloud.IsStatus404ByInt(respCode) {
				return retry.RetryableError(fmt.Errorf("Failed to read teams resource %s: %s", d.Id(), getErr))
			}
			return retry.NonRetryableError(fmt.Errorf("Failed to read teams resource %s: %s", d.Id(), getErr))
		}

		cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceTeamsResource())

		resourcedata.SetNillableValue(d, "name", team.Name)
		resourcedata.SetNillableValue(d, "division_id", team.DivisionId)
		resourcedata.SetNillableValue(d, "description", team.Description)
		resourcedata.SetNillableValue(d, "member_count", team.MemberCount)

		log.Printf("Read teams resource %s %s", d.Id(), *team.Name)
		return cc.CheckState()
	})
}

// updateTeamsResource is used by the teams_resource resource to update an teams resource in Genesys Cloud
func updateTeamsResource(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*gcloud.ProviderMeta).ClientConfig
	proxy := newTeamsResourceProxy(sdkConfig)

	teamsResource := getTeamsResourceFromResourceData(d)

	log.Printf("Updating teams resource %s", *teamsResource.Name)
	team, err := proxy.updateTeamsResource(ctx, d.Id(), &teamsResource)
	if err != nil {
		return diag.Errorf("Failed to update teams resource: %s", err)
	}

	log.Printf("Updated teams resource %s", *team.Id)
	return readTeamsResource(ctx, d, meta)
}

// deleteTeamsResource is used by the teams_resource resource to delete an teams resource from Genesys cloud
func deleteTeamsResource(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*gcloud.ProviderMeta).ClientConfig
	proxy := newTeamsResourceProxy(sdkConfig)

	_, err := proxy.deleteTeamsResource(ctx, d.Id())
	if err != nil {
		return diag.Errorf("Failed to delete teams resource %s: %s", d.Id(), err)
	}

	return gcloud.WithRetries(ctx, 180*time.Second, func() *retry.RetryError {
		_, respCode, err := proxy.getTeamsResourceById(ctx, d.Id())

		if err == nil {
			if gcloud.IsStatus404ByInt(respCode) {
				log.Printf("Deleted teams resource %s", d.Id())
				return nil
			}
			return retry.NonRetryableError(fmt.Errorf("Error deleting teams resource %s: %s", d.Id(), err))
		}

		return retry.RetryableError(fmt.Errorf("teams resource %s still exists", d.Id()))
	})
}

// getTeamsResourceFromResourceData maps data from schema ResourceData object to a platformclientv2.Team
func getTeamsResourceFromResourceData(d *schema.ResourceData) platformclientv2.Team {
	return platformclientv2.Team{
		Name:        platformclientv2.String(d.Get("name").(string)),
		DivisionId:  platformclientv2.String(d.Get("division_id").(string)),
		Description: platformclientv2.String(d.Get("description").(string)),
		MemberCount: platformclientv2.Int(d.Get("member_count").(int)),
	}
}
