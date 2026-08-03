package team

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/consistency_checker"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/constants"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/lists"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v193/platformclientv2"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/resourcedata"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
)

/*
The resource_genesyscloud_team.go contains all of the methods that perform the core logic for a resource.
*/

// getAllAuthTeam retrieves all of the team via Terraform in the Genesys Cloud and is used for the exporter
func getAllAuthTeams(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	proxy := getTeamProxy(clientConfig)
	resources := make(resourceExporter.ResourceIDMetaMap)
	teams, resp, err := proxy.getAllTeam(ctx, "")
	if err != nil {
		return nil, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to get team error: %s", err), resp)
	}
	for _, team := range *teams {
		resources[*team.Id] = &resourceExporter.ResourceMeta{BlockLabel: *team.Name}
	}
	return resources, nil
}

// createTeam is used by the team resource to create Genesys cloud team
func createTeam(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getTeamProxy(sdkConfig)
	team := getTeamFromResourceData(d)

	log.Printf("Creating team %s", *team.Name)
	teamObj, resp, err := proxy.createTeam(ctx, &team)
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to create team %s error: %s", *team.Name, err), resp)
	}
	d.SetId(*teamObj.Id)

	// adding members to the team
	diagErr := updateTeamMembers(ctx, d, sdkConfig)
	if diagErr != nil {
		// Preserve prior state on apply error (SDK #476).
		// Partial causes State() to read from prior state instead of proposed config.
		d.Partial(true)
		return diagErr
	}

	log.Printf("Created team %s", *teamObj.Id)
	diags := readTeam(ctx, d, meta)
	if diags.HasError() {
		d.Partial(true)
	}
	return diags
}

// readTeam is used by the team resource to read a team from genesys cloud
func readTeam(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getTeamProxy(sdkConfig)
	cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceTeam(), constants.ConsistencyChecks(), ResourceType)

	log.Printf("Reading team %s", d.Id())

	return util.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		team, resp, getErr := proxy.getTeamById(ctx, d.Id())
		if getErr != nil {
			if util.IsStatus404(resp) {
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("failed to read team %s | error: %s", d.Id(), getErr), resp))
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("failed to read team %s | error: %s", d.Id(), getErr), resp))
		}

		resourcedata.SetNillableValue(d, "name", team.Name)
		resourcedata.SetNillableReferenceWritableDivision(d, "division_id", team.Division)
		resourcedata.SetNillableValue(d, "description", team.Description)

		var schemaMemberIds []string
		if raw, ok := d.Get("member_ids").([]interface{}); ok {
			schemaMemberIds = lists.InterfaceListToStrings(raw)
		}

		members := []string{}
		if team.MemberCount != nil && *team.MemberCount > 0 {
			apiMembers, err := readTeamMembers(ctx, d.Id(), sdkConfig)
			if err != nil {
				return retry.NonRetryableError(fmt.Errorf("%v", err))
			}
			members = apiMembers
		}
		_ = d.Set("member_ids", organizeMemberIdsForRead(schemaMemberIds, members))

		log.Printf("Read team %s %s", d.Id(), *team.Name)
		return cc.CheckState(d)
	})
}

// updateTeam is used by the team resource to update a team in Genesys Cloud
func updateTeam(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getTeamProxy(sdkConfig)
	team := getTeamFromResourceData(d)

	log.Printf("Updating team %s", *team.Name)
	teamObj, resp, err := proxy.updateTeam(ctx, d.Id(), &team)
	if err != nil {
		// Preserve prior state on apply error (SDK #476).
		d.Partial(true)
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to update team %s error: %s", *team.Name, err), resp)
	}

	diagErr := updateTeamMembers(ctx, d, sdkConfig)
	if diagErr != nil {
		d.Partial(true)
		return diagErr
	}

	log.Printf("Updated team %s", *teamObj.Id)
	diags := readTeam(ctx, d, meta)
	if diags.HasError() {
		d.Partial(true)
	}
	return diags
}

// deleteTeam is used by the team resource to delete a team from Genesys cloud
func deleteTeam(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getTeamProxy(sdkConfig)

	log.Printf("Deleting team %s", d.Id())
	resp, err := proxy.deleteTeam(ctx, d.Id())
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to delete team %s error: %s", d.Id(), err), resp)
	}
	return util.WithRetries(ctx, 180*time.Second, func() *retry.RetryError {
		_, resp, err := proxy.getTeamById(ctx, d.Id())
		if err != nil {
			if util.IsStatus404(resp) {
				log.Printf("Deleted team %s", d.Id())
				return nil
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("error deleting team %s | error: %s", d.Id(), err), resp))
		}
		return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("team %s still exists", d.Id()), resp))
	})
}
