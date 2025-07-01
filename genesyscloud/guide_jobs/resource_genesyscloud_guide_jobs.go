package guide_jobs

import (
	"context"
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/consistency_checker"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/constants"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/resourcedata"
)

func createGuideJob(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getGuideJobsProxy(sdkConfig)

	log.Printf("Creating Guide Job")
	guideJobRequest := buildGuideJobFromResourceData(d)

	job, resp, err := proxy.createGuideJob(ctx, &guideJobRequest)
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to create guide job error: %s", err), resp)
	}

	switch job.Status {
	case "InProgress":
		log.Printf("Create job still in progress with status: %s", job.Status)
	case "Succeeded":
		log.Printf("Created successfully")
		return readGuideJob(ctx, d, meta)
	case "Failed":
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to create guide job, with error: %s", job.Errors[0].Message), resp)
	default:
		return util.BuildDiagnosticError(ResourceType, fmt.Sprintf("Unknown job status: %s", job.Status), nil)
	}

	d.SetId(job.Id)
	log.Printf("Created guide: %s", job.Id)
	return readGuideJob(ctx, d, meta)
}

func readGuideJob(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getGuideJobsProxy(sdkConfig)
	cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceGuideJobs(), constants.ConsistencyChecks(), ResourceType)

	log.Printf("Reading Guide Job: %s", d.Id())

	return util.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		job, resp, err := proxy.getGuideJobById(ctx, d.Id())
		if err != nil {
			if util.IsStatus404(resp) {
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read guide job: %s | Error: %s", d.Id(), err), resp))
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read guide job: %s | Error: %s", d.Id(), err), resp))
		}

		resourcedata.SetNillableValue(d, "status", &job.Status)

		log.Printf("Read Guide Job: %s", d.Id())
		return cc.CheckState(d)
	})
}

func deleteGuideJob(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return nil
}
