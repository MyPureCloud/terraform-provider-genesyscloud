package guide

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

func createGuideJob(ctx context.Context, d *schema.ResourceData, meta interface{}, guideName string) (*GeneratedGuideContent, diag.Diagnostics) {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getGuideProxy(sdkConfig)

	prompt := d.Get("prompt").(string)
	url := d.Get("url").(string)

	if prompt == "" && url == "" {
		return nil, diag.Errorf("either prompt or url is required when source is set to Prompt")
	}

	jobReq := buildGuideJobRequest(prompt, url)

	job, resp, err := proxy.createGuideJob(ctx, jobReq)
	if err != nil {
		return nil, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("failed to create guide job for %s: %s", guideName, err), resp)
	}

	log.Printf("Created guide job with ID: %s, Status: %s", *job.Id, *job.Status)

	content, diagErr := readGuideJob(ctx, proxy, *job.Id, guideName)
	if diagErr != nil {
		return nil, diagErr
	}

	if content != nil {
		log.Printf("Guide job completed successfully, returning generated content for guide: %s", guideName)
		return content, nil
	}

	return nil, diag.Errorf("guide job completed but no content was generated for guide: %s", guideName)
}

func readGuideJob(ctx context.Context, proxy *guideProxy, jobId string, guideName string) (*GeneratedGuideContent, diag.Diagnostics) {
	log.Printf("Reading guide job %s for guide %s", jobId, guideName)

	var result *GeneratedGuideContent
	err := util.WithRetries(ctx, 2*time.Minute, func() *retry.RetryError {
		job, resp, err := proxy.getGuideJobById(ctx, jobId)
		if err != nil {
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("error checking guide job status for %s: %s", guideName, err), resp))
		}

		if job.Status == nil {
			return retry.NonRetryableError(fmt.Errorf("guide job %s has nil status for %s", jobId, guideName))
		}

		log.Printf("Guide job %s status: %s", jobId, *job.Status)

		switch *job.Status {
		case "InProgress":
			return retry.RetryableError(fmt.Errorf("guide job for %s still in progress: %s", guideName, *job.Status))
		case "Succeeded":
			log.Printf("Guide job %s completed successfully", jobId)
			result = job.GuideContent
			return nil
		case "Failed":
			if len(job.Errors) > 0 && job.Errors[0].Message != "" {
				return retry.NonRetryableError(fmt.Errorf("guide job failed for %s: %s", guideName, job.Errors[0].Message))
			}
			return retry.NonRetryableError(fmt.Errorf("guide job failed for %s | Status: %s", guideName, *job.Status))
		default:
			return retry.RetryableError(fmt.Errorf("unexpected job status for %s | Status: %s", guideName, *job.Status))
		}
	})

	if err != nil {
		return nil, err
	}
	return result, nil
}
