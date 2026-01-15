package users_rules

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v176/platformclientv2"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/consistency_checker"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/constants"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/resourcedata"
)

/*
The resource_genesyscloud_users_rules.go contains all of the methods that perform the core logic for a resource.
*/

// getAllAuthUsersRules retrieves all of the users rules via Terraform in the Genesys Cloud and is used for the exporter
func getAllAuthUsersRules(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	proxy := newUsersRulesProxy(clientConfig)
	resources := make(resourceExporter.ResourceIDMetaMap)

	usersRules, resp, err := proxy.getAllUsersRules(ctx, "")
	if err != nil {
		return nil, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to get users rules error: %s", err), resp)
	}

	for _, usersRule := range *usersRules {
		resources[*usersRule.Id] = &resourceExporter.ResourceMeta{BlockLabel: *usersRule.Name}
	}

	return resources, nil
}

// createUsersRules is used by the users_rules resource to create Genesys cloud users rule
func createUsersRules(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getUsersRulesProxy(sdkConfig)

	var usersRuleRequest platformclientv2.Usersrulescreaterulerequest
	usersRuleRequest.Name = platformclientv2.String(d.Get("name").(string))
	description := d.Get("description").(string)
	if description != "" {
		usersRuleRequest.Description = &description
	}
	usersRuleRequest.VarType = platformclientv2.String(d.Get("type").(string))
	usersRuleRequest.Criteria = buildSdkUsersRulesCriteria(d.Get("criteria").([]interface{}))

	log.Printf("Creating users rule %s", *usersRuleRequest.Name)
	usersRule, resp, err := proxy.createUsersRules(ctx, &usersRuleRequest)
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to create users rule %s error: %s", *usersRule.Name, err), resp)
	}

	d.SetId(*usersRule.Id)
	log.Printf("Created users rule %s: %s", *usersRule.Name, *usersRule.Id)
	return readUsersRules(ctx, d, meta)
}

// readUsersRules is used by the users_rules resource to read a users rule from genesys cloud
func readUsersRules(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getUsersRulesProxy(sdkConfig)
	cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceUsersRules(), constants.ConsistencyChecks(), ResourceType)

	log.Printf("Reading users rule %s", d.Id())

	return util.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		usersRule, resp, getErr := proxy.getUsersRulesById(ctx, d.Id())
		if getErr != nil {
			if util.IsStatus404(resp) {
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read users rule %s | error: %s", d.Id(), getErr), resp))
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read users rule %s | error: %s", d.Id(), getErr), resp))
		}

		resourcedata.SetNillableValue(d, "name", usersRule.Name)
		resourcedata.SetNillableValue(d, "description", usersRule.Description)
		resourcedata.SetNillableValue(d, "type", usersRule.VarType)
		if usersRule.Criteria != nil {
			d.Set("criteria", flattenUsersRulesCriteria(usersRule.Criteria))
		}

		log.Printf("Read users rule %s %s", d.Id(), *usersRule.Name)
		return cc.CheckState(d)
	})
}

// updateUsersRules is used by the users_rules resource to update a users rule in Genesys Cloud
func updateUsersRules(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getUsersRulesProxy(sdkConfig)

	var usersRuleRequest platformclientv2.Usersrulesupdaterulerequest
	name := d.Get("name").(string)
	if name != "" {
		usersRuleRequest.Name = &name
	}
	description := d.Get("description").(string)
	if description != "" {
		usersRuleRequest.Description = &description
	}
	usersRuleRequest.Criteria = buildSdkUsersRulesCriteria(d.Get("criteria").([]interface{}))

	log.Printf("Updating users rule %s", d.Id())

	usersRule, resp, err := proxy.updateUsersRules(ctx, d.Id(), &usersRuleRequest)
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to update users rule %s error: %s", d.Id(), err), resp)
	}

	log.Printf("Updated users rule %s: %s", *usersRule.Name, d.Id())

	return readUsersRules(ctx, d, meta)
}

// deleteUsersRules is used by the users_rules resource to delete a users rule from Genesys cloud
func deleteUsersRules(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getUsersRulesProxy(sdkConfig)

	resp, err := proxy.deleteUsersRules(ctx, d.Id())
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to delete users rule %s error: %s", d.Id(), err), resp)
	}

	return util.WithRetries(ctx, 180*time.Second, func() *retry.RetryError {
		_, resp, err := proxy.getUsersRulesById(ctx, d.Id())

		if err != nil {
			if util.IsStatus404(resp) {
				log.Printf("Deleted users rule %s", d.Id())
				return nil
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Error deleting users rule %s | error: %s", d.Id(), err), resp))
		}
		return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("users rule %s still exists", d.Id()), resp))
	})
}
