package team

import (
	"context"
	"fmt"
	"log"
	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v119/platformclientv2"

	"terraform-provider-genesyscloud/genesyscloud/consistency_checker"

	gcloud "terraform-provider-genesyscloud/genesyscloud"

	"terraform-provider-genesyscloud/genesyscloud/util/resourcedata"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
)

/*
The resource_genesyscloud_team.go contains all of the methods that perform the core logic for a resource.
*/

// getAllAuthTeam retrieves all of the team via Terraform in the Genesys Cloud and is used for the exporter
func getAllAuthTeams(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	proxy := getTeamProxy(clientConfig)
	resources := make(resourceExporter.ResourceIDMetaMap)
	teams, err := proxy.getAllTeam(ctx, "")
	if err != nil {
		return nil, diag.Errorf("Failed to get team: %v", err)
	}

	for _, team := range *teams {
		resources[*team.Id] = &resourceExporter.ResourceMeta{Name: *team.Name}
	}

	return resources, nil
}

// createTeam is used by the team resource to create Genesys cloud team
func createTeam(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*gcloud.ProviderMeta).ClientConfig
	proxy := getTeamProxy(sdkConfig)

	team := getTeamFromResourceData(d)

	log.Printf("Creating team %s", *team.Name)
	teamObj, err := proxy.createTeam(ctx, &team)
	if err != nil {
		return diag.Errorf("Failed to create team: %s", err)
	}

	d.SetId(*teamObj.Id)
	log.Printf("Created team %s", *teamObj.Id)
	return readTeam(ctx, d, meta)
}

// readTeam is used by the team resource to read an team from genesys cloud
func readTeam(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*gcloud.ProviderMeta).ClientConfig
	proxy := getTeamProxy(sdkConfig)

	log.Printf("Reading team %s", d.Id())

	return gcloud.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		team, respCode, getErr := proxy.getTeamById(ctx, d.Id())
		if getErr != nil {
			if gcloud.IsStatus404ByInt(respCode) {
				return retry.RetryableError(fmt.Errorf("Failed to read team %s: %s", d.Id(), getErr))
			}
			return retry.NonRetryableError(fmt.Errorf("Failed to read team %s: %s", d.Id(), getErr))
		}

		cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceTeam())

		resourcedata.SetNillableValue(d, "name", team.Name)

		resourcedata.SetNillableReferenceWritableDivision(d, "division_id", team.Division)

		resourcedata.SetNillableValue(d, "description", team.Description)

		log.Printf("Read team %s %s", d.Id(), *team.Name)
		return cc.CheckState()
	})
}

// updateTeam is used by the team resource to update an team in Genesys Cloud
func updateTeam(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*gcloud.ProviderMeta).ClientConfig
	proxy := getTeamProxy(sdkConfig)

	team := getTeamFromResourceData(d)

	log.Printf("Updating team %s", *team.Name)
	teamObj, err := proxy.updateTeam(ctx, d.Id(), &team)
	if err != nil {
		return diag.Errorf("Failed to update team: %s", err)
	}

	log.Printf("Updated team %s", *teamObj.Id)
	return readTeam(ctx, d, meta)
}

// deleteTeam is used by the team resource to delete an team from Genesys cloud
func deleteTeam(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*gcloud.ProviderMeta).ClientConfig
	proxy := getTeamProxy(sdkConfig)

	_, err := proxy.deleteTeam(ctx, d.Id())
	if err != nil {
		return diag.Errorf("Failed to delete team %s: %s", d.Id(), err)
	}

	return gcloud.WithRetries(ctx, 180*time.Second, func() *retry.RetryError {
		_, respCode, err := proxy.getTeamById(ctx, d.Id())

		if err != nil {
			if gcloud.IsStatus404ByInt(respCode) {
				log.Printf("Deleted team %s", d.Id())
				return nil
			}
			return retry.NonRetryableError(fmt.Errorf("Error deleting team %s: %s", d.Id(), err))
		}

		return retry.RetryableError(fmt.Errorf("team %s still exists", d.Id()))
	})
}

// getTeamFromResourceData maps data from schema ResourceData object to a platformclientv2.Team
func getTeamFromResourceData(d *schema.ResourceData) platformclientv2.Team {

	name := d.Get("name").(string)
	division := d.Get("division_id").(string)

	return platformclientv2.Team{
		Name:        &name,
		Division:    &platformclientv2.Writabledivision{Id: &division},
		Description: platformclientv2.String(d.Get("description").(string)),
	}
}
