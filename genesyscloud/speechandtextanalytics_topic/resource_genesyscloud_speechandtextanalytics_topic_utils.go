package speechandtextanalytics_topic

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v179/platformclientv2"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
)

func buildTopicRequest(d *schema.ResourceData) *platformclientv2.Topicrequest {
	name := d.Get("name").(string)
	dialect := d.Get("dialect").(string)
	description := d.Get("description").(string)
	strictness := d.Get("strictness").(string)
	participants := d.Get("participants").(string)

	req := &platformclientv2.Topicrequest{
		Name:         &name,
		Dialect:      &dialect,
		Strictness:   &strictness,
		Participants: &participants,
	}

	if description != "" {
		req.Description = &description
	}

	if programIds, ok := d.Get("program_ids").(*schema.Set); ok && programIds.Len() > 0 {
		ids := make([]string, 0, programIds.Len())
		for _, v := range programIds.List() {
			ids = append(ids, v.(string))
		}
		req.ProgramIds = &ids
	}

	if tags, ok := d.Get("tags").(*schema.Set); ok && tags.Len() > 0 {
		t := make([]string, 0, tags.Len())
		for _, v := range tags.List() {
			t = append(t, v.(string))
		}
		req.Tags = &t
	}

	if phrases, ok := d.Get("phrases").([]interface{}); ok && len(phrases) > 0 {
		ps := make([]platformclientv2.Phrase, 0, len(phrases))
		for _, p := range phrases {
			m := p.(map[string]interface{})
			text := m["text"].(string)
			phrase := platformclientv2.Phrase{Text: &text}
			if v, ok := m["strictness"].(string); ok && v != "" {
				phrase.Strictness = &v
			}
			if v, ok := m["sentiment"].(string); ok && v != "" {
				phrase.Sentiment = &v
			}
			ps = append(ps, phrase)
		}
		req.Phrases = &ps
	}

	return req
}

func flattenTopicToResourceData(d *schema.ResourceData, topic *platformclientv2.Topic) error {
	if topic == nil {
		return nil
	}

	if topic.Name != nil {
		_ = d.Set("name", *topic.Name)
	}
	if topic.Dialect != nil {
		_ = d.Set("dialect", *topic.Dialect)
	}
	if topic.Description != nil {
		_ = d.Set("description", *topic.Description)
	} else {
		_ = d.Set("description", "")
	}
	if topic.Strictness != nil {
		_ = d.Set("strictness", *topic.Strictness)
	}
	if topic.Participants != nil {
		_ = d.Set("participants", *topic.Participants)
	}
	if topic.Tags != nil {
		_ = d.Set("tags", *topic.Tags)
	} else {
		_ = d.Set("tags", []string{})
	}

	// Programs are returned as objects, but request wants IDs. We do not set program_ids from read.
	// This avoids drift unless user explicitly manages program IDs.

	if topic.Phrases != nil {
		phrases := make([]interface{}, 0, len(*topic.Phrases))
		for _, p := range *topic.Phrases {
			pm := make(map[string]interface{})
			if p.Text != nil {
				pm["text"] = *p.Text
			}
			if p.Strictness != nil {
				pm["strictness"] = *p.Strictness
			}
			if p.Sentiment != nil {
				pm["sentiment"] = *p.Sentiment
			}
			phrases = append(phrases, pm)
		}
		_ = d.Set("phrases", phrases)
	}

	if topic.Published != nil {
		_ = d.Set("published", *topic.Published)
	}

	return nil
}

func waitForPublishJob(ctx context.Context, proxy *sttTopicProxy, jobId string, timeout time.Duration) diag.Diagnostics {
	return util.WithRetries(ctx, timeout, func() *retry.RetryError {
		job, resp, err := proxy.getPublishJob(ctx, jobId)
		if err != nil {
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to get topics publish job %s: %s", jobId, err), resp))
		}

		if job == nil || job.State == nil {
			return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Topics publish job %s state not available yet", jobId), resp))
		}

		switch *job.State {
		case "Completed":
			return nil
		case "Failed":
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Topics publish job %s failed", jobId), resp))
		default:
			return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Topics publish job %s not completed yet (state: %s)", jobId, *job.State), resp))
		}
	})
}
