package recording_media_retention_policy

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"terraform-provider-genesyscloud/genesyscloud/consistency_checker"
	"terraform-provider-genesyscloud/genesyscloud/util/resourcedata"

	gcloud "terraform-provider-genesyscloud/genesyscloud"
	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v115/platformclientv2"
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

	retentionPolicies, err := pp.getAllPolicies(ctx)
	if err != nil {
		return nil, diag.Errorf("Failed to get page of media retention policies %v", err)
	}

	for _, retentionPolicy := range *retentionPolicies {
		resources[*retentionPolicy.Id] = &resourceExporter.ResourceMeta{Name: *retentionPolicy.Name}
	}

	return resources, nil
}

// createMediaRetentionPolicy is used by the recording media retention policy resource to create Genesyscloud a media retention policy
func createMediaRetentionPolicy(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*gcloud.ProviderMeta).ClientConfig
	pp := getPolicyProxy(sdkConfig)

	name := d.Get("name").(string)
	order := d.Get("order").(int)
	description := d.Get("description").(string)
	enabled := d.Get("enabled").(bool)
	mediaPolicies := buildMediaPolicies(d, pp, ctx)
	conditions := buildConditions(d)
	actions := buildPolicyActionsFromResource(d, pp, ctx)
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
		return diag.Errorf("Failed to create media retention policy %s: %s", name, err)
	}

	// Make sure form is properly created
	policyId := policy.Id
	d.SetId(*policyId)
	log.Printf("Created media retention policy %s %s", name, *policy.Id)
	return readMediaRetentionPolicy(ctx, d, meta)
}

// readMediaRetentionPolicy is used by the recording media retention policy resource to read a media retention policy from genesys cloud.
func readMediaRetentionPolicy(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*gcloud.ProviderMeta).ClientConfig
	pp := getPolicyProxy(sdkConfig)

	log.Printf("Reading media retention policy %s", d.Id())

	return gcloud.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		retentionPolicy, resp, err := pp.getPolicyById(ctx, d.Id())
		if err != nil {
			if gcloud.IsStatus404(resp) {
				return retry.RetryableError(fmt.Errorf("failed to read media retention policy %s: %s", d.Id(), err))
			}
			return retry.NonRetryableError(fmt.Errorf("failed to read media retention policy %s: %s", d.Id(), err))
		}

		cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, gcloud.ResourceSurveyForm())

		resourcedata.SetNillableValue(d, "name", retentionPolicy.Name)
		resourcedata.SetNillableValue(d, "order", retentionPolicy.Order)
		resourcedata.SetNillableValue(d, "description", retentionPolicy.Description)
		resourcedata.SetNillableValue(d, "enabled", retentionPolicy.Enabled)
		resourcedata.SetNillableValueWithInterfaceArrayWithFunc(d, "conditions", retentionPolicy.Conditions, flattenConditions)
		resourcedata.SetNillableValueWithInterfaceArrayWithFunc(d, "policy_errors", retentionPolicy.PolicyErrors, flattenPolicyErrors)

		if retentionPolicy.MediaPolicies != nil {
			d.Set("media_policies", flattenMediaPolicies(retentionPolicy.MediaPolicies, pp, ctx))
		}
		if retentionPolicy.Actions != nil {
			d.Set("actions", flattenPolicyActions(retentionPolicy.Actions, pp, ctx))
		}

		return cc.CheckState()
	})
}

// updateMediaRetentionPolicy is used by the recording media retention policy resource to update a media retention policy in Genesys Cloud
func updateMediaRetentionPolicy(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*gcloud.ProviderMeta).ClientConfig
	pp := getPolicyProxy(sdkConfig)

	name := d.Get("name").(string)
	order := d.Get("order").(int)
	description := d.Get("description").(string)
	enabled := d.Get("enabled").(bool)

	mediaPolicies := buildMediaPolicies(d, pp, ctx)
	conditions := buildConditions(d)
	actions := buildPolicyActionsFromResource(d, pp, ctx)
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
	policy, err := pp.updatePolicy(ctx, d.Id(), &reqBody)
	if err != nil {
		return diag.Errorf("Failed to update media retention policy %s: %s", name, err)
	}

	log.Printf("Updated media retention policy %s %s", name, *policy.Id)
	return readMediaRetentionPolicy(ctx, d, meta)
}

// deleteMediaRetentionPolicy is used by the recording media retention policy resource to delete a media retention policy from Genesys cloud.
func deleteMediaRetentionPolicy(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)

	sdkConfig := meta.(*gcloud.ProviderMeta).ClientConfig
	pp := getPolicyProxy(sdkConfig)

	log.Printf("Deleting media retention policy %s", name)
	_, err := pp.deletePolicy(ctx, d.Id())
	if err != nil {
		return diag.Errorf("Failed to delete media retention policy %s: %s", name, err)
	}

	return gcloud.WithRetries(ctx, 30*time.Second, func() *retry.RetryError {
		_, resp, err := pp.getPolicyById(ctx, d.Id())
		if err != nil {
			if gcloud.IsStatus404(resp) {
				// media retention policy deleted
				log.Printf("Deleted media retention policy %s", d.Id())
				return nil
			}
			return retry.NonRetryableError(fmt.Errorf("error deleting media retention policy %s: %s", d.Id(), err))
		}

		return retry.RetryableError(fmt.Errorf("media retention policy %s still exists", d.Id()))
	})
}
