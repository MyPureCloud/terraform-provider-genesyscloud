package speechandtextanalytics_program

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v200/platformclientv2"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
)

// buildProgramRequest converts Terraform resource data into a Programrequest.
// Only writable fields are included; read-only fields (published, topics, audit fields) are never sent.
func buildProgramRequest(d *schema.ResourceData) *platformclientv2.Programrequest {
	name := d.Get("name").(string)
	description := d.Get("description").(string)

	req := &platformclientv2.Programrequest{
		Name: &name,
	}

	if description != "" {
		req.Description = &description
	}

	if topicIds, ok := d.Get("topic_ids").(*schema.Set); ok && topicIds.Len() > 0 {
		ids := make([]string, 0, topicIds.Len())
		for _, v := range topicIds.List() {
			ids = append(ids, v.(string))
		}
		req.TopicIds = &ids
	}

	if tags, ok := d.Get("tags").(*schema.Set); ok && tags.Len() > 0 {
		t := make([]string, 0, tags.Len())
		for _, v := range tags.List() {
			t = append(t, v.(string))
		}
		req.Tags = &t
	}

	return req
}

// flattenProgramToResourceData maps a Program response onto Terraform resource data.
func flattenProgramToResourceData(d *schema.ResourceData, program *platformclientv2.Program) {
	if program == nil {
		return
	}

	if program.Name != nil {
		_ = d.Set("name", *program.Name)
	}
	if program.Description != nil {
		_ = d.Set("description", *program.Description)
	} else {
		_ = d.Set("description", "")
	}

	// The API accepts topicIds on write but returns full topic entities on read.
	// Map the returned topics back to their IDs so the set round-trips cleanly.
	if program.Topics != nil {
		ids := make([]string, 0, len(*program.Topics))
		for _, t := range *program.Topics {
			if t.Id != nil {
				ids = append(ids, *t.Id)
			}
		}
		_ = d.Set("topic_ids", ids)
	} else {
		_ = d.Set("topic_ids", []string{})
	}

	if program.Tags != nil {
		_ = d.Set("tags", *program.Tags)
	} else {
		_ = d.Set("tags", []string{})
	}

	if program.Published != nil {
		_ = d.Set("published", *program.Published)
	}
}

// waitForPublishJob polls the programs publish job until it completes, fails, or the timeout elapses.
func waitForPublishJob(ctx context.Context, proxy *sttProgramProxy, jobId string, timeout time.Duration) diag.Diagnostics {
	return util.WithRetries(ctx, timeout, func() *retry.RetryError {
		job, resp, err := proxy.getPublishJob(ctx, jobId)
		if err != nil {
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to get programs publish job %s: %s", jobId, err), resp))
		}

		if job == nil || job.State == nil {
			return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Programs publish job %s state not available yet", jobId), resp))
		}

		switch *job.State {
		case "Completed":
			return nil
		case "Failed":
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Programs publish job %s failed", jobId), resp))
		default:
			return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Programs publish job %s not completed yet (state: %s)", jobId, *job.State), resp))
		}
	})
}
