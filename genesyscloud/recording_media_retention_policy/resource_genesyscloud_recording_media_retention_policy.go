package recording_media_retention_policy

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"terraform-provider-genesyscloud/genesyscloud/consistency_checker"

	gcloud "terraform-provider-genesyscloud/genesyscloud"
	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v105/platformclientv2"
)

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
		if retentionPolicy.Name != nil {
			d.Set("name", *retentionPolicy.Name)
		}
		if retentionPolicy.Order != nil {
			d.Set("order", *retentionPolicy.Order)
		}
		if retentionPolicy.Description != nil {
			d.Set("description", *retentionPolicy.Description)
		}
		if retentionPolicy.Enabled != nil {
			d.Set("enabled", *retentionPolicy.Enabled)
		}
		if retentionPolicy.MediaPolicies != nil {
			d.Set("media_policies", flattenMediaPolicies(retentionPolicy.MediaPolicies, pp, ctx))
		}
		if retentionPolicy.Conditions != nil {
			d.Set("conditions", flattenConditions(retentionPolicy.Conditions))
		}
		if retentionPolicy.Actions != nil {
			d.Set("actions", flattenPolicyActions(retentionPolicy.Actions, pp, ctx))
		}
		if retentionPolicy.PolicyErrors != nil {
			d.Set("policy_errors", flattenPolicyErrors(retentionPolicy.PolicyErrors))
		}

		return cc.CheckState()
	})
}

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
