package task_management_workitem

import (
	"context"
	"fmt"
	"net/http"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"

	lists "terraform-provider-genesyscloud/genesyscloud/util/lists"
	"terraform-provider-genesyscloud/genesyscloud/util/resourcedata"

	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"
)

var (
	utWorkitemConfig = &workitemConfig{
		name:                   "tf-workitem" + uuid.NewString(),
		worktype_id:            "tf-worktype" + uuid.NewString(),
		description:            "test workitem created by CX as Code",
		language_id:            "tf-language" + uuid.NewString(),
		priority:               42,
		date_due:               time.Now().Add(time.Hour * 10).Format(resourcedata.TimeParseFormat),
		date_expires:           time.Now().Add(time.Hour * 20).Format(resourcedata.TimeParseFormat),
		duration_seconds:       99999,
		ttl:                    int(time.Now().Add(time.Hour * 24 * 30 * 6).Unix()),
		status_id:              "tf-status" + uuid.NewString(),
		workbin_id:             "tf-workbin" + uuid.NewString(),
		assignee_id:            "tf-user" + uuid.NewString(),
		external_contact_id:    "tf-external-contact" + uuid.NewString(),
		external_tag:           "external tag",
		queue_id:               "tf-queue" + uuid.NewString(),
		skills_ids:             []string{"tf-skill" + uuid.NewString(), "tf-skill" + uuid.NewString()},
		preferred_agents_ids:   []string{"tf-user" + uuid.NewString(), "tf-user" + uuid.NewString()},
		auto_status_transition: false,
		custom_fields:          `{"customField1": "customValue1", "customField2": "customValue2" }`,
		scored_agents: []scoredAgentConfig{{
			agent_id: "tf-user" + uuid.NewString(),
			score:    42,
		}},
	}
)

/** Unit Test **/
func TestUnitResourceWorkitemCreate(t *testing.T) {
	tId := uuid.NewString()
	wi := utWorkitemConfig

	taskProxy := &taskManagementWorkitemProxy{}
	taskProxy.getTaskManagementWorkitemByIdAttr = func(ctx context.Context, p *taskManagementWorkitemProxy, id string) (*platformclientv2.Workitem, *platformclientv2.APIResponse, error) {
		assert.Equal(t, tId, id)

		dateDueTime, err := time.Parse(resourcedata.TimeParseFormat, wi.date_due)
		if err != nil {
			assert.Fail(t, "Failed to parse date_due")
		}
		dateExpTime, err := time.Parse(resourcedata.TimeParseFormat, wi.date_expires)
		if err != nil {
			assert.Fail(t, "Failed to parse date_expires")
		}

		workitem := &platformclientv2.Workitem{
			Id:          &tId,
			Name:        &wi.name,
			Description: &wi.description,
			VarType: &platformclientv2.Worktypereference{
				Id: &wi.worktype_id,
			},
			Language: &platformclientv2.Languagereference{
				Id: &wi.language_id,
			},
			Priority:        &wi.priority,
			DateCreated:     timePtr(time.Now()),
			DateModified:    timePtr(time.Now()),
			DateDue:         timePtr(dateDueTime),
			DateExpires:     timePtr(dateExpTime),
			DurationSeconds: &wi.duration_seconds,
			Ttl:             &wi.ttl,
			Status: &platformclientv2.Workitemstatusreference{
				Id: &wi.status_id,
			},
			Workbin: &platformclientv2.Workbinreference{
				Id: &wi.workbin_id,
			},
			Assignee: &platformclientv2.Userreferencewithname{
				Id: &wi.assignee_id,
			},
			ExternalContact: &platformclientv2.Externalcontactreference{
				Id: &wi.external_contact_id,
			},
			ExternalTag: &wi.external_tag,
			Queue: &platformclientv2.Workitemqueuereference{
				Id: &wi.queue_id,
			},
			Skills: &[]platformclientv2.Routingskillreference{
				{
					Id: &wi.skills_ids[0],
				},
				{
					Id: &wi.skills_ids[1],
				},
			},
			PreferredAgents: &[]platformclientv2.Userreference{
				{
					Id: &wi.preferred_agents_ids[0],
				},
				{
					Id: &wi.preferred_agents_ids[1],
				},
			},
			AutoStatusTransition: &wi.auto_status_transition,
			CustomFields: &map[string]interface{}{
				"customField1": "customValue1",
				"customField2": "customValue2",
			},
			ScoredAgents: &[]platformclientv2.Workitemscoredagent{
				{
					Agent: &platformclientv2.Userreference{
						Id: &wi.scored_agents[0].agent_id,
					},
					Score: &wi.scored_agents[0].score,
				},
			},
		}

		return workitem, nil, nil
	}

	taskProxy.createTaskManagementWorkitemAttr = func(ctx context.Context, p *taskManagementWorkitemProxy, workitem *platformclientv2.Workitemcreate) (*platformclientv2.Workitem, *platformclientv2.APIResponse, error) {
		assert.Equal(t, wi.name, *workitem.Name, "Name check failed in create createTaskManagementWorkitemAttr")
		assert.Equal(t, wi.worktype_id, *workitem.TypeId, "TypeId check failed in create createTaskManagementWorkitemAttr")
		assert.Equal(t, wi.description, *workitem.Description, "Description check failed in create createTaskManagementWorkitemAttr")
		assert.Equal(t, wi.language_id, *workitem.LanguageId, "LanguageId check failed in create createTaskManagementWorkitemAttr")
		assert.Equal(t, wi.priority, *workitem.Priority, "Priority check failed in create createTaskManagementWorkitemAttr")
		assert.Equal(t, wi.date_due, (*workitem.DateDue).Format(resourcedata.TimeParseFormat), "DateDue check failed in create createTaskManagementWorkitemAttr")
		assert.Equal(t, wi.date_expires, (*workitem.DateExpires).Format(resourcedata.TimeParseFormat), "DateExpires check failed in create createTaskManagementWorkitemAttr")
		assert.Equal(t, wi.duration_seconds, *workitem.DurationSeconds, "DurationSeconds check failed in create createTaskManagementWorkitemAttr")
		assert.Equal(t, wi.ttl, *workitem.Ttl, "Ttl check failed in create createTaskManagementWorkitemAttr")
		assert.Equal(t, wi.status_id, *workitem.StatusId, "StatusId check failed in create createTaskManagementWorkitemAttr")
		assert.Equal(t, wi.workbin_id, *workitem.WorkbinId, "WorkbinId check failed in create createTaskManagementWorkitemAttr")
		assert.Equal(t, wi.assignee_id, *workitem.AssigneeId, "AssigneeId check failed in create createTaskManagementWorkitemAttr")
		assert.Equal(t, wi.external_contact_id, *workitem.ExternalContactId, "ExternalContactId check failed in create createTaskManagementWorkitemAttr")
		assert.Equal(t, wi.external_tag, *workitem.ExternalTag, "ExternalTag check failed in create createTaskManagementWorkitemAttr")
		assert.Equal(t, wi.queue_id, *workitem.QueueId, "QueueId check failed in create createTaskManagementWorkitemAttr")
		assert.ElementsMatch(t, wi.skills_ids, *workitem.SkillIds, "SkillIds check failed in create createTaskManagementWorkitemAttr")
		assert.ElementsMatch(t, wi.preferred_agents_ids, *workitem.PreferredAgentIds, "PreferredAgentIds check failed in create createTaskManagementWorkitemAttr")
		assert.Equal(t, wi.auto_status_transition, *workitem.AutoStatusTransition, "AutoStatusTransition check failed in create createTaskManagementWorkitemAttr")
		assert.ElementsMatch(t, wi.scored_agents, apiScoredAgentReqToScoredAgentConfig(workitem.ScoredAgents), "ScoredAgents check failed in create createTaskManagementWorkitemAttr")

		cfjson, err := util.MapToJson(workitem.CustomFields)
		if err != nil {
			assert.Fail(t, "Failed to parse CustomFields: %v", err)
		}
		if !equivalentJsons(wi.custom_fields, cfjson) {
			assert.Fail(t, "ScoredAgents check failed in create createTaskManagementWorkitemAttr")
		}

		return &platformclientv2.Workitem{
			Id: &tId,
		}, nil, nil
	}

	internalProxy = taskProxy
	defer func() { internalProxy = nil }()

	ctx := context.Background()
	gcloud := &provider.ProviderMeta{ClientConfig: &platformclientv2.Configuration{}}

	//Grab our defined schema
	resourceSchema := ResourceTaskManagementWorkitem().Schema

	//Setup a map of values
	resourceDataMap := buildWorkitemResourceMap(tId, wi)

	//Found this gem TestResourceDataRaw here: https://github.com/hashicorp/terraform/issues/890
	d := schema.TestResourceDataRaw(t, resourceSchema, resourceDataMap)
	d.SetId(tId)

	diag := createTaskManagementWorkitem(ctx, d, gcloud)
	assert.Equal(t, false, diag.HasError())
	assert.Equal(t, tId, d.Id())
}

func TestUnitResourceWorkitemRead(t *testing.T) {
	tId := uuid.NewString()
	wi := utWorkitemConfig

	taskProxy := &taskManagementWorkitemProxy{}

	taskProxy.getTaskManagementWorkitemByIdAttr = func(ctx context.Context, p *taskManagementWorkitemProxy, id string) (*platformclientv2.Workitem, *platformclientv2.APIResponse, error) {
		assert.Equal(t, tId, id)

		dateDueTime, err := time.Parse(resourcedata.TimeParseFormat, wi.date_due)
		if err != nil {
			assert.Fail(t, "Failed to parse date_due")
		}
		dateExpTime, err := time.Parse(resourcedata.TimeParseFormat, wi.date_expires)
		if err != nil {
			assert.Fail(t, "Failed to parse date_expires")
		}

		workitem := &platformclientv2.Workitem{
			Id:          &tId,
			Name:        &wi.name,
			Description: &wi.description,
			VarType: &platformclientv2.Worktypereference{
				Id: &wi.worktype_id,
			},
			Language: &platformclientv2.Languagereference{
				Id: &wi.language_id,
			},
			Priority:        &wi.priority,
			DateCreated:     timePtr(time.Now()),
			DateModified:    timePtr(time.Now()),
			DateDue:         timePtr(dateDueTime),
			DateExpires:     timePtr(dateExpTime),
			DurationSeconds: &wi.duration_seconds,
			Ttl:             &wi.ttl,
			Status: &platformclientv2.Workitemstatusreference{
				Id: &wi.status_id,
			},
			Workbin: &platformclientv2.Workbinreference{
				Id: &wi.workbin_id,
			},
			Assignee: &platformclientv2.Userreferencewithname{
				Id: &wi.assignee_id,
			},
			ExternalContact: &platformclientv2.Externalcontactreference{
				Id: &wi.external_contact_id,
			},
			ExternalTag: &wi.external_tag,
			Queue: &platformclientv2.Workitemqueuereference{
				Id: &wi.queue_id,
			},
			Skills: &[]platformclientv2.Routingskillreference{
				{
					Id: &wi.skills_ids[0],
				},
				{
					Id: &wi.skills_ids[1],
				},
			},
			PreferredAgents: &[]platformclientv2.Userreference{
				{
					Id: &wi.preferred_agents_ids[0],
				},
				{
					Id: &wi.preferred_agents_ids[1],
				},
			},
			AutoStatusTransition: &wi.auto_status_transition,
			CustomFields: &map[string]interface{}{
				"customField1": "customValue1",
				"customField2": "customValue2",
			},
			ScoredAgents: &[]platformclientv2.Workitemscoredagent{
				{
					Agent: &platformclientv2.Userreference{
						Id: &wi.scored_agents[0].agent_id,
					},
					Score: &wi.scored_agents[0].score,
				},
			},
		}

		return workitem, nil, nil
	}

	internalProxy = taskProxy
	defer func() { internalProxy = nil }()

	ctx := context.Background()
	gcloud := &provider.ProviderMeta{ClientConfig: &platformclientv2.Configuration{}}

	//Grab our defined schema
	resourceSchema := ResourceTaskManagementWorkitem().Schema

	//Setup a map of values
	resourceDataMap := buildWorkitemResourceMap(tId, wi)

	//Found this gem TestResourceDataRaw here: https://github.com/hashicorp/terraform/issues/890
	d := schema.TestResourceDataRaw(t, resourceSchema, resourceDataMap)
	d.SetId(tId)

	diag := readTaskManagementWorkitem(ctx, d, gcloud)

	assert.Equal(t, false, diag.HasError())
	assert.Equal(t, tId, d.Id())
	assert.Equal(t, wi.name, d.Get("name").(string))
	assert.Equal(t, wi.description, d.Get("description").(string))
	assert.Equal(t, wi.worktype_id, d.Get("worktype_id").(string))
	assert.Equal(t, wi.language_id, d.Get("language_id").(string))
	assert.Equal(t, wi.priority, d.Get("priority").(int))
	assert.Equal(t, wi.date_due, d.Get("date_due").(string))
	assert.Equal(t, wi.date_expires, d.Get("date_expires").(string))
	assert.Equal(t, wi.duration_seconds, d.Get("duration_seconds").(int))
	assert.Equal(t, wi.ttl, d.Get("ttl").(int))
	assert.Equal(t, wi.status_id, d.Get("status_id").(string))
	assert.Equal(t, wi.workbin_id, d.Get("workbin_id").(string))
	assert.Equal(t, wi.assignee_id, d.Get("assignee_id").(string))
	assert.Equal(t, wi.external_contact_id, d.Get("external_contact_id").(string))
	assert.Equal(t, wi.external_tag, d.Get("external_tag").(string))
	assert.Equal(t, wi.queue_id, d.Get("queue_id").(string))
	assert.ElementsMatch(t, wi.skills_ids, d.Get("skills_ids").([]interface{}))
	assert.ElementsMatch(t, wi.preferred_agents_ids, d.Get("preferred_agents_ids").([]interface{}))
	assert.Equal(t, wi.auto_status_transition, d.Get("auto_status_transition").(bool))
	if !equivalentJsons(wi.custom_fields, d.Get("custom_fields").(string)) {
		assert.Fail(t, "custom_fields do not match")
	}
	assert.ElementsMatch(t, wi.scored_agents, *scoredAgentInterfaceToConfig(d.Get("scored_agents").([]interface{})))
}

func TestUnitResourceWorkitemUpdate(t *testing.T) {
	tId := uuid.NewString()

	// The complete configuration for the worktype
	wi := utWorkitemConfig

	taskProxy := &taskManagementWorkitemProxy{}

	taskProxy.updateTaskManagementWorkitemAttr = func(ctx context.Context, p *taskManagementWorkitemProxy, id string, workitem *platformclientv2.Workitemupdate) (*platformclientv2.Workitem, *platformclientv2.APIResponse, error) {
		assert.Equal(t, wi.name, *workitem.Name, "Name check failed in create updateTaskManagementWorktypeAttr")
		assert.Equal(t, wi.description, *workitem.Description, "Description check failed in create updateTaskManagementWorktypeAttr")
		assert.Equal(t, wi.language_id, *workitem.LanguageId, "LanguageId check failed in create updateTaskManagementWorktypeAttr")
		assert.Equal(t, wi.priority, *workitem.Priority, "Priority check failed in create updateTaskManagementWorktypeAttr")
		assert.Equal(t, wi.date_due, (*workitem.DateDue).Format(resourcedata.TimeParseFormat), "DateDue check failed in create updateTaskManagementWorktypeAttr")
		assert.Equal(t, wi.date_expires, (*workitem.DateExpires).Format(resourcedata.TimeParseFormat), "DateExpires check failed in create updateTaskManagementWorktypeAttr")
		assert.Equal(t, wi.duration_seconds, *workitem.DurationSeconds, "DurationSeconds check failed in create updateTaskManagementWorktypeAttr")
		assert.Equal(t, wi.ttl, *workitem.Ttl, "Ttl check failed in create updateTaskManagementWorktypeAttr")
		assert.Equal(t, wi.status_id, *workitem.StatusId, "StatusId check failed in create updateTaskManagementWorktypeAttr")
		assert.Equal(t, wi.workbin_id, *workitem.WorkbinId, "WorkbinId check failed in create updateTaskManagementWorktypeAttr")
		assert.Equal(t, wi.assignee_id, *workitem.AssigneeId, "AssigneeId check failed in create updateTaskManagementWorktypeAttr")
		assert.Equal(t, wi.external_contact_id, *workitem.ExternalContactId, "ExternalContactId check failed in create updateTaskManagementWorktypeAttr")
		assert.Equal(t, wi.external_tag, *workitem.ExternalTag, "ExternalTag check failed in create updateTaskManagementWorktypeAttr")
		assert.Equal(t, wi.queue_id, *workitem.QueueId, "QueueId check failed in create updateTaskManagementWorktypeAttr")
		assert.ElementsMatch(t, wi.skills_ids, *workitem.SkillIds, "SkillIds check failed in create updateTaskManagementWorktypeAttr")
		assert.ElementsMatch(t, wi.preferred_agents_ids, *workitem.PreferredAgentIds, "PreferredAgentIds check failed in create updateTaskManagementWorktypeAttr")
		assert.Equal(t, wi.auto_status_transition, *workitem.AutoStatusTransition, "AutoStatusTransition check failed in create updateTaskManagementWorktypeAttr")
		assert.ElementsMatch(t, wi.scored_agents, apiScoredAgentReqToScoredAgentConfig(workitem.ScoredAgents), "ScoredAgents check failed in create updateTaskManagementWorktypeAttr")

		cfjson, err := util.MapToJson(workitem.CustomFields)
		if err != nil {
			assert.Fail(t, "Failed to parse CustomFields: %v", err)
		}
		if !equivalentJsons(wi.custom_fields, cfjson) {
			assert.Fail(t, "ScoredAgents check failed in create createTaskManagementWorkitemAttr")
		}

		return &platformclientv2.Workitem{
			Id: &tId,
		}, nil, nil
	}

	taskProxy.getTaskManagementWorkitemByIdAttr = func(ctx context.Context, p *taskManagementWorkitemProxy, id string) (*platformclientv2.Workitem, *platformclientv2.APIResponse, error) {
		assert.Equal(t, tId, id)

		dateDueTime, err := time.Parse(resourcedata.TimeParseFormat, wi.date_due)
		if err != nil {
			assert.Fail(t, "Failed to parse date_due")
		}
		dateExpTime, err := time.Parse(resourcedata.TimeParseFormat, wi.date_expires)
		if err != nil {
			assert.Fail(t, "Failed to parse date_expires")
		}

		workitem := &platformclientv2.Workitem{
			Id:          &tId,
			Name:        &wi.name,
			Description: &wi.description,
			VarType: &platformclientv2.Worktypereference{
				Id: &wi.worktype_id,
			},
			Language: &platformclientv2.Languagereference{
				Id: &wi.language_id,
			},
			Priority:        &wi.priority,
			DateCreated:     timePtr(time.Now()),
			DateModified:    timePtr(time.Now()),
			DateDue:         timePtr(dateDueTime),
			DateExpires:     timePtr(dateExpTime),
			DurationSeconds: &wi.duration_seconds,
			Ttl:             &wi.ttl,
			Status: &platformclientv2.Workitemstatusreference{
				Id: &wi.status_id,
			},
			Workbin: &platformclientv2.Workbinreference{
				Id: &wi.workbin_id,
			},
			Assignee: &platformclientv2.Userreferencewithname{
				Id: &wi.assignee_id,
			},
			ExternalContact: &platformclientv2.Externalcontactreference{
				Id: &wi.external_contact_id,
			},
			ExternalTag: &wi.external_tag,
			Queue: &platformclientv2.Workitemqueuereference{
				Id: &wi.queue_id,
			},
			Skills: &[]platformclientv2.Routingskillreference{
				{
					Id: &wi.skills_ids[0],
				},
				{
					Id: &wi.skills_ids[1],
				},
			},
			PreferredAgents: &[]platformclientv2.Userreference{
				{
					Id: &wi.preferred_agents_ids[0],
				},
				{
					Id: &wi.preferred_agents_ids[1],
				},
			},
			AutoStatusTransition: &wi.auto_status_transition,
			CustomFields: &map[string]interface{}{
				"customField1": "customValue1",
				"customField2": "customValue2",
			},
			ScoredAgents: &[]platformclientv2.Workitemscoredagent{
				{
					Agent: &platformclientv2.Userreference{
						Id: &wi.scored_agents[0].agent_id,
					},
					Score: &wi.scored_agents[0].score,
				},
			},
		}

		return workitem, nil, nil
	}

	internalProxy = taskProxy
	defer func() { internalProxy = nil }()

	ctx := context.Background()
	gcloud := &provider.ProviderMeta{ClientConfig: &platformclientv2.Configuration{}}

	//Grab our defined schema
	resourceSchema := ResourceTaskManagementWorkitem().Schema

	//Setup a map of values
	resourceDataMap := buildWorkitemResourceMap(tId, wi)

	//Found this gem TestResourceDataRaw here: https://github.com/hashicorp/terraform/issues/890
	d := schema.TestResourceDataRaw(t, resourceSchema, resourceDataMap)
	d.SetId(tId)

	diag := updateTaskManagementWorkitem(ctx, d, gcloud)
	assert.Equal(t, false, diag.HasError())
	assert.Equal(t, tId, d.Id())
}

func TestUnitResourceWorkitemDelete(t *testing.T) {
	tId := uuid.NewString()
	wi := utWorkitemConfig

	taskProxy := &taskManagementWorkitemProxy{}

	taskProxy.deleteTaskManagementWorkitemAttr = func(ctx context.Context, p *taskManagementWorkitemProxy, id string) (resp *platformclientv2.APIResponse, err error) {
		assert.Equal(t, tId, id)

		apiResponse := &platformclientv2.APIResponse{StatusCode: http.StatusNoContent}
		return apiResponse, nil
	}

	taskProxy.getTaskManagementWorkitemByIdAttr = func(ctx context.Context, p *taskManagementWorkitemProxy, id string) (workitem *platformclientv2.Workitem, resp *platformclientv2.APIResponse, err error) {
		assert.Equal(t, tId, id)
		apiResponse := &platformclientv2.APIResponse{StatusCode: http.StatusNotFound}

		return nil, apiResponse, fmt.Errorf("not found")
	}

	internalProxy = taskProxy
	defer func() { internalProxy = nil }()

	ctx := context.Background()
	gcloud := &provider.ProviderMeta{ClientConfig: &platformclientv2.Configuration{}}

	//Grab our defined schema
	resourceSchema := ResourceTaskManagementWorkitem().Schema

	//Setup a map of values
	resourceDataMap := buildWorkitemResourceMap(tId, wi)

	//Found this gem TestResourceDataRaw here: https://github.com/hashicorp/terraform/issues/890
	d := schema.TestResourceDataRaw(t, resourceSchema, resourceDataMap)
	d.SetId(tId)

	diag := deleteTaskManagementWorkitem(ctx, d, gcloud)
	assert.Nil(t, diag)
	assert.Equal(t, tId, d.Id())
}

func buildWorkitemResourceMap(tId string, wt *workitemConfig) map[string]interface{} {
	return map[string]interface{}{
		"id":                     tId,
		"name":                   wt.name,
		"worktype_id":            wt.worktype_id,
		"description":            wt.description,
		"language_id":            wt.language_id,
		"priority":               wt.priority,
		"date_due":               wt.date_due,
		"date_expires":           wt.date_expires,
		"duration_seconds":       wt.duration_seconds,
		"ttl":                    wt.ttl,
		"status_id":              wt.status_id,
		"workbin_id":             wt.workbin_id,
		"assignee_id":            wt.assignee_id,
		"external_contact_id":    wt.external_contact_id,
		"external_tag":           wt.external_tag,
		"queue_id":               wt.queue_id,
		"skills_ids":             lists.StringListToInterfaceList(wt.skills_ids),
		"preferred_agents_ids":   lists.StringListToInterfaceList(wt.preferred_agents_ids),
		"auto_status_transition": wt.auto_status_transition,
		"custom_fields":          wt.custom_fields,
		"scored_agents":          buildScoredAgentsList(wt.scored_agents),
	}
}

func buildScoredAgentsList(scoredAgents []scoredAgentConfig) []interface{} {
	var scoredAgentsList []interface{}
	for _, scoredAgent := range scoredAgents {
		agentMap := map[string]interface{}{
			"agent_id": scoredAgent.agent_id,
			"score":    scoredAgent.score,
		}
		scoredAgentsList = append(scoredAgentsList, agentMap)
	}
	return scoredAgentsList
}

func apiScoredAgentReqToScoredAgentConfig(scoredAgents *[]platformclientv2.Workitemscoredagentrequest) []scoredAgentConfig {
	var scoredAgentConfigs []scoredAgentConfig
	for _, scoredAgent := range *scoredAgents {
		scoredAgentConfigs = append(scoredAgentConfigs, scoredAgentConfig{
			agent_id: *scoredAgent.Id,
			score:    *scoredAgent.Score,
		})
	}
	return scoredAgentConfigs
}

func scoredAgentInterfaceToConfig(scoredAgents []interface{}) *[]scoredAgentConfig {
	var scoredAgentConfigs []scoredAgentConfig
	for _, scoredAgent := range scoredAgents {
		agentMap := scoredAgent.(map[string]interface{})
		scoredAgentConfigs = append(scoredAgentConfigs, scoredAgentConfig{
			agent_id: agentMap["agent_id"].(string),
			score:    int(agentMap["score"].(int)),
		})
	}
	return &scoredAgentConfigs
}

func timePtr(t time.Time) *time.Time {
	return &t
}
func equivalentJsons(json1, json2 string) bool {
	return util.EquivalentJsons(json1, json2)
}
