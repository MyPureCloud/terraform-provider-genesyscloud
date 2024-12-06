package recording_media_retention_policy

import (
	"context"
	"fmt"
	"log"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"terraform-provider-genesyscloud/genesyscloud/util/constants"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"terraform-provider-genesyscloud/genesyscloud/consistency_checker"
	"terraform-provider-genesyscloud/genesyscloud/util/resourcedata"

	gcloud "terraform-provider-genesyscloud/genesyscloud"
	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v146/platformclientv2"
)

/*
The resource_genesyscloud_recording_media_retention_policy.go contains all of the methods that perform the core logic for a resource.
In general a resource should have a approximately 5 methods in it:

1.  A getAll.... function that the CX as Code exporter will use during the process of exporting Genesys Cloud.
2.  A create.... function that the resource will use to create a Genesys Cloud object (e.g. genesyscloud_recording_media_retention_policy)
3.  A read.... function that looks up a single resource.
4.  An update... function that updates a single resource.
5.  A delete.... function that deletes a single resource.

Two things to note:

 1. All code in these methods should be focused on getting data in and out of Terraform.  All code that is used for interacting
    with a Genesys API should be encapsulated into a proxy class contained within the package.

 2. In general, to keep this file somewhat manageable, if you find yourself with a number of helper functions move them to a

utils function in the package.  This will keep the code manageable and easy to work through.
*/

// getAllMediaRetentionPolicies retrieves all of the recording media retention policies via Terraform in the Genesys Cloud and is used for the exporter
func getAllMediaRetentionPolicies(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	resources := make(resourceExporter.ResourceIDMetaMap)
	pp := getPolicyProxy(clientConfig)

	retentionPolicies, resp, err := pp.getAllPolicies(ctx)
	if err != nil {
		return nil, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to get page of media retention policies error: %s", err), resp)
	}

	for _, retentionPolicy := range *retentionPolicies {
		resources[*retentionPolicy.Id] = &resourceExporter.ResourceMeta{BlockLabel: *retentionPolicy.Name}
	}
	return resources, nil
}

// createMediaRetentionPolicy is used by the recording media retention policy resource to create Genesyscloud a media retention policy
func createMediaRetentionPolicy(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	pp := getPolicyProxy(sdkConfig)

	name := d.Get("name").(string)
	order := d.Get("order").(int)
	description := d.Get("description").(string)
	enabled := d.Get("enabled").(bool)
	err, mediaPolicies := buildMediaPolicies(d, pp, ctx)

	if err != nil {
		util.BuildDiagnosticError(ResourceType, "error while calling buildMediaPolicie()in createMediaRetention", err)
	}

	conditions := buildConditions(d)
	err, actions := buildPolicyActionsFromResource(d, pp, ctx)
	if err != nil {
		util.BuildDiagnosticError(ResourceType, "error while calling buildPolicyActionsFromResource()", err)
	}

	policyErrors := buildPolicyErrors(d)

	reqBody := platformclientv2.Policycreate{
		Name:          &name,
		Order:         &order,
		Description:   &description,
		Enabled:       &enabled,
		MediaPolicies: mediaPolicies,
		Conditions:    conditions,
		Actions:       actions,
		PolicyErrors:  policyErrors,
	}

	log.Printf("Creating media retention policy %s", name)
	policy, resp, err := pp.createPolicy(ctx, &reqBody)
	log.Printf("Media retention policy creation status %#v", resp.Status)
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to create media retention policy %s error: %s", name, err), resp)
	}

	// Make sure form is properly created
	policyId := policy.Id
	d.SetId(*policyId)
	log.Printf("Created media retention policy %s %s", name, *policy.Id)
	return readMediaRetentionPolicy(ctx, d, meta)
}

// readMediaRetentionPolicy is used by the recording media retention policy resource to read a media retention policy from genesys cloud.
func readMediaRetentionPolicy(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	pp := getPolicyProxy(sdkConfig)
	cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, gcloud.ResourceSurveyForm(), constants.ConsistencyChecks(), ResourceType)

	log.Printf("Reading media retention policy %s", d.Id())

	return util.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		retentionPolicy, resp, err := pp.getPolicyById(ctx, d.Id())
		if err != nil {
			if util.IsStatus404(resp) {
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("failed to read media retention policy %s | error: %s", d.Id(), err), resp))
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("failed to read media retention policy %s | error: %s", d.Id(), err), resp))
		}

		resourcedata.SetNillableValue(d, "name", retentionPolicy.Name)
		resourcedata.SetNillableValue(d, "order", retentionPolicy.Order)
		resourcedata.SetNillableValue(d, "description", retentionPolicy.Description)
		resourcedata.SetNillableValue(d, "enabled", retentionPolicy.Enabled)
		resourcedata.SetNillableValueWithInterfaceArrayWithFunc(d, "conditions", retentionPolicy.Conditions, flattenConditions)
		resourcedata.SetNillableValueWithInterfaceArrayWithFunc(d, "policy_errors", retentionPolicy.PolicyErrors, flattenPolicyErrors)

		err, mediaPolicies := flattenMediaPolicies(retentionPolicy.MediaPolicies, pp, ctx)
		if err != nil {
			return retry.NonRetryableError(fmt.Errorf("Unable to flatten media policies in readMediaRetentionPolicy() method: %s", err))
		}
		if retentionPolicy.MediaPolicies != nil {
			d.Set("media_policies", mediaPolicies)
		}

		err, actions := flattenPolicyActions(retentionPolicy.Actions, pp, ctx)
		if err != nil {
			return retry.NonRetryableError(fmt.Errorf("Unable to flatten actions in readMediaRetentionPolicy(): %s", err))
		}
		if retentionPolicy.Actions != nil {
			d.Set("actions", actions)
		}
		return cc.CheckState(d)
	})
}

// updateMediaRetentionPolicy is used by the recording media retention policy resource to update a media retention policy in Genesys Cloud
func updateMediaRetentionPolicy(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	pp := getPolicyProxy(sdkConfig)

	name := d.Get("name").(string)
	order := d.Get("order").(int)
	description := d.Get("description").(string)
	enabled := d.Get("enabled").(bool)
	err, mediaPolicies := buildMediaPolicies(d, pp, ctx)
	if err != nil {
		return util.BuildDiagnosticError(ResourceType, "Error while retrieving buildMediaPolicies() function in updateMediaRetentionPolicy() method)", err)
	}

	conditions := buildConditions(d)
	err, actions := buildPolicyActionsFromResource(d, pp, ctx)

	if err != nil {
		return util.BuildDiagnosticError(ResourceType, "Error while retrieving buildPolicyActionsFromResource() function in updateMediaRetentionPolicy() method)", err)
	}

	policyErrors := buildPolicyErrors(d)

	reqBody := platformclientv2.Policy{
		Name:          &name,
		Order:         &order,
		Description:   &description,
		Enabled:       &enabled,
		MediaPolicies: mediaPolicies,
		Conditions:    conditions,
		Actions:       actions,
		PolicyErrors:  policyErrors,
	}

	log.Printf("Updating media retention policy %s", name)
	policy, resp, err := pp.updatePolicy(ctx, d.Id(), &reqBody)
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to update media retention policy %s error: %s", name, err), resp)
	}

	log.Printf("Updated media retention policy %s %s", name, *policy.Id)
	return readMediaRetentionPolicy(ctx, d, meta)
}

// deleteMediaRetentionPolicy is used by the recording media retention policy resource to delete a media retention policy from Genesys cloud.
func deleteMediaRetentionPolicy(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)

	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	pp := getPolicyProxy(sdkConfig)

	log.Printf("Deleting media retention policy %s", name)
	resp, err := pp.deletePolicy(ctx, d.Id())
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to delete media retention policy %s error: %s", d.Id(), err), resp)
	}

	return util.WithRetries(ctx, 30*time.Second, func() *retry.RetryError {
		_, resp, err := pp.getPolicyById(ctx, d.Id())
		if err != nil {
			if util.IsStatus404(resp) {
				// media retention policy deleted
				log.Printf("Deleted media retention policy %s", d.Id())
				return nil
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("error deleting media retention policy %s | error: %s", d.Id(), err), resp))
		}
		return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("media retention policy %s still exists", d.Id()), resp))
	})
}
