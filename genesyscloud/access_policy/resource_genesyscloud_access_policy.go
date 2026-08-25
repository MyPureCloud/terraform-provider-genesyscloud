package access_policy

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
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/resourcedata"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v195/platformclientv2"
)

/*
The resource_genesyscloud_access_policy.go contains all of the methods that perform the core logic for a resource.
*/

// getAllAccessPolicies retrieves all of the access policies via Terraform in the Genesys Cloud and is used for the exporter
func getAllAccessPolicies(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	proxy := newAccessPolicyProxy(clientConfig)
	resources := make(resourceExporter.ResourceIDMetaMap)

	policies, resp, err := proxy.getAllAccessPolicy(ctx)
	if err != nil {
		return nil, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to get access policies: %s", err), resp)
	}

	for _, policy := range *policies {
		if policy.Id != nil && policy.Name != nil {
			resources[*policy.Id] = &resourceExporter.ResourceMeta{BlockLabel: *policy.Name}
		}
	}
	return resources, nil
}

// createAccessPolicy is used by the access_policy resource to create a Genesys Cloud access policy
func createAccessPolicy(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getAccessPolicyProxy(sdkConfig)

	policy, err := buildAccessPolicyFromResourceData(d)
	if err != nil {
		return diag.Errorf("Failed to build access policy from resource data: %s", err)
	}

	targetResource := d.Get("target_resource").(string)

	createdPolicy, resp, createErr := proxy.createAccessPolicy(ctx, targetResource, policy)
	if createErr != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to create access policy %s: %s", *policy.Name, createErr), resp)
	}

	d.SetId(*createdPolicy.Id)
	log.Printf("Created access policy %s %s", *createdPolicy.Name, *createdPolicy.Id)
	return readAccessPolicy(ctx, d, meta)
}

// readAccessPolicy is used by the access_policy resource to read an access policy from Genesys Cloud
func readAccessPolicy(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getAccessPolicyProxy(sdkConfig)
	cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceAccessPolicy(), constants.ConsistencyChecks(), ResourceType)

	log.Printf("Reading access policy %s", d.Id())

	return util.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		policy, resp, getErr := proxy.getAccessPolicyById(ctx, d.Id())
		if getErr != nil {
			if util.IsStatus404(resp) {
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read access policy %s | error: %s", d.Id(), getErr), resp))
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read access policy %s | error: %s", d.Id(), getErr), resp))
		}

		resourcedata.SetNillableValue(d, "name", policy.Name)
		resourcedata.SetNillableValue(d, "description", policy.Description)
		resourcedata.SetNillableValue(d, "target_resource", policy.TargetResource)
		resourcedata.SetNillableValue(d, "effect", policy.Effect)
		resourcedata.SetNillableValue(d, "enabled", policy.Active)
		resourcedata.SetNillableValue(d, "apply_to_clients", policy.ApplyToClients)

		// Flatten subject
		if policy.Subject != nil {
			if policy.Subject.VarType != nil {
				_ = d.Set("subject_type", *policy.Subject.VarType)
			}
			if policy.Subject.Id != nil {
				_ = d.Set("subject_id", *policy.Subject.Id)
			}
		}

		// Flatten condition JSON
		if conditionJSON := flattenConditionToJSON(policy.Condition); conditionJSON != "" {
			_ = d.Set("condition_json", conditionJSON)
		}

		// Flatten preset attributes JSON
		if presetJSON := flattenPresetAttributesToJSON(policy.PresetAttributes); presetJSON != "" {
			_ = d.Set("preset_attributes_json", presetJSON)
		}

		log.Printf("Read access policy %s %s", d.Id(), *policy.Name)
		return cc.CheckState(d)
	})
}

// updateAccessPolicy is used by the access_policy resource to update an access policy in Genesys Cloud
func updateAccessPolicy(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getAccessPolicyProxy(sdkConfig)

	policy, err := buildAccessPolicyFromResourceData(d)
	if err != nil {
		return diag.Errorf("Failed to build access policy from resource data: %s", err)
	}

	log.Printf("Updating access policy %s", *policy.Name)
	_, resp, updateErr := proxy.updateAccessPolicy(ctx, d.Id(), policy)
	if updateErr != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to update access policy %s: %s", d.Id(), updateErr), resp)
	}

	log.Printf("Updated access policy %s", d.Id())
	return readAccessPolicy(ctx, d, meta)
}

// deleteAccessPolicy is used by the access_policy resource to delete an access policy from Genesys Cloud
func deleteAccessPolicy(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getAccessPolicyProxy(sdkConfig)

	targetResource := d.Get("target_resource").(string)

	// Use subject_id from state (which may have been computed by the API, e.g., "all")
	subjectId := d.Get("subject_id").(string)
	if subjectId == "" {
		// Fallback: use subject_type as the identifier (e.g., "ALL")
		subjectId = d.Get("subject_type").(string)
	}

	log.Printf("Deleting access policy %s", d.Id())
	resp, deleteErr := proxy.deleteAccessPolicy(ctx, d.Id(), targetResource, subjectId)
	if deleteErr != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to delete access policy %s: %s", d.Id(), deleteErr), resp)
	}

	return util.WithRetries(ctx, 60*time.Second, func() *retry.RetryError {
		_, resp, err := proxy.getAccessPolicyById(ctx, d.Id())
		if err != nil {
			if util.IsStatus404(resp) {
				log.Printf("Deleted access policy %s", d.Id())
				return nil
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Error deleting access policy %s | error: %s", d.Id(), err), resp))
		}
		return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Access policy %s still exists", d.Id()), resp))
	})
}
