package speechandtextanalytics_program

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v199/platformclientv2"
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
