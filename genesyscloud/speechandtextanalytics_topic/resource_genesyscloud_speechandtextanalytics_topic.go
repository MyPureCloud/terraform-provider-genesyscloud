package speechandtextanalytics_topic

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
)

func createTopic(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getSttTopicProxy(sdkConfig)

	req := buildTopicRequest(d)
	log.Printf("Creating Speech & Text Analytics Topic %s", d.Get("name").(string))
	topic, resp, err := proxy.createTopic(ctx, req)
	if err != nil {
		input, _ := util.InterfaceToJson(*req)
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to create speech and text analytics topic: %s\n(input: %+v)", err, input), resp)
	}

	if topic == nil || topic.Id == nil {
		return diag.Errorf("API returned success but no topic ID for %s", d.Get("name").(string))
	}
	d.SetId(*topic.Id)

	desiredPublished := d.Get("published").(bool)
	if desiredPublished {
		job, resp, err := proxy.publishTopics(ctx, []string{d.Id()})
		if err != nil {
			return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to publish speech and text analytics topic %s: %s", d.Id(), err), resp)
		}
		if job != nil && job.Id != nil {
			if diagErr := waitForPublishJob(ctx, proxy, *job.Id, 10*time.Minute); diagErr != nil {
				return diagErr
			}
		}
	}

	return readTopic(ctx, d, meta)
}

func readTopic(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getSttTopicProxy(sdkConfig)

	log.Printf("Reading Speech & Text Analytics Topic %s", d.Id())
	return util.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		topic, resp, err := proxy.getTopic(ctx, d.Id())
		if err != nil {
			if util.IsStatus404(resp) {
				d.SetId("")
				return nil
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read speech and text analytics topic %s: %s", d.Id(), err), resp))
		}

		_ = flattenTopicToResourceData(d, topic)
		return nil
	})
}

func updateTopic(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getSttTopicProxy(sdkConfig)

	req := buildTopicRequest(d)
	log.Printf("Updating Speech & Text Analytics Topic %s", d.Id())
	topic, resp, err := proxy.updateTopic(ctx, d.Id(), req)
	if err != nil {
		input, _ := util.InterfaceToJson(*req)
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to update speech and text analytics topic %s: %s\n(input: %+v)", d.Id(), err, input), resp)
	}
	if topic == nil || topic.Id == nil {
		return diag.Errorf("API returned success but no topic ID for %s", d.Get("name").(string))
	}
	d.SetId(*topic.Id)

	desiredPublished := d.Get("published").(bool)
	if desiredPublished {
		// Only publish if not already published
		current, resp, err := proxy.getTopic(ctx, d.Id())
		if err != nil {
			return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to read topic %s before publishing: %s", d.Id(), err), resp)
		}
		if current == nil || current.Published == nil || !*current.Published {
			job, resp, err := proxy.publishTopics(ctx, []string{d.Id()})
			if err != nil {
				return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to publish speech and text analytics topic %s: %s", d.Id(), err), resp)
			}
			if job != nil && job.Id != nil {
				if diagErr := waitForPublishJob(ctx, proxy, *job.Id, 10*time.Minute); diagErr != nil {
					return diagErr
				}
			}
		}
	}

	return readTopic(ctx, d, meta)
}

func deleteTopic(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getSttTopicProxy(sdkConfig)

	log.Printf("Deleting Speech & Text Analytics Topic %s", d.Id())
	resp, err := proxy.deleteTopic(ctx, d.Id())
	if err != nil {
		if util.IsStatus404(resp) {
			return nil
		}
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to delete speech and text analytics topic %s: %s", d.Id(), err), resp)
	}

	return util.WithRetries(ctx, 180*time.Second, func() *retry.RetryError {
		_, resp, err := proxy.getTopic(ctx, d.Id())
		if err != nil {
			if util.IsStatus404(resp) {
				log.Printf("Deleted speech and text analytics topic %s", d.Id())
				return nil
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Error verifying deletion of topic %s: %s", d.Id(), err), resp))
		}
		return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("speech and text analytics topic %s still exists", d.Id()), resp))
	})
}
