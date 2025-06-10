package guide_jobs

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v157/platformclientv2"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/consistency_checker"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/constants"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/resourcedata"
	"log"
	"time"
)

func getAllGuideJobs(_ context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	return nil, nil
}

func createGuideJob(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getGuideJobsProxy(sdkConfig)

	log.Printf("Creating Guide Job")
	guideJobRequest := buildGuideJobFromResourceData(d)

	job, resp, err := proxy.createGuideJob(ctx, &guideJobRequest)
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to create guide job error: %s", err), resp)
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

		// TODO: Set Guide, The Addressable Entity Ref
		//resourcedata.SetNillableValueWithSchemaSetWithFunc(d, "guide_id", job.Guide, flattenAddressableEntityRefs)

		log.Printf("Read Guide Job: %s", d.Id())
		return cc.CheckState(d)
	})
}

func deleteGuideJob(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getGuideJobsProxy(sdkConfig)

	log.Printf("Deleting Guide Job: %s", d.Id())
	resp, err := proxy.deleteGuideJob(ctx, d.Id())
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to delete guide job | error: %s", err), resp)
	}

	return util.WithRetries(ctx, 180*time.Second, func() *retry.RetryError {
		_, resp, err := proxy.getGuideJobById(ctx, d.Id())

		if err != nil {
			if util.IsStatus404(resp) {
				log.Printf("Deleted Guide Job: %s", d.Id())
				return nil
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Error deleting Guide Job %s | error: %s", d.Id(), err), resp))
		}
		return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Guide Job %s still exists", d.Id()), resp))
	})
}
