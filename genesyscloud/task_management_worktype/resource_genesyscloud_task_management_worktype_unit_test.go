package task_management_worktype

import (
	"context"
	"fmt"
	"terraform-provider-genesyscloud/genesyscloud/provider"

	"net/http"
	"terraform-provider-genesyscloud/genesyscloud/util/lists"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"
	"github.com/stretchr/testify/assert"
)

/** Unit Test **/
func TestUnitResourceWorktypeCreate(t *testing.T) {
	tId := uuid.NewString()

	// The complete configuration for the worktype
	wt := &worktypeConfig{
		name:             "tf_worktype_" + uuid.NewString(),
		description:      "worktype created for CX as Code test case",
		divisionId:       uuid.NewString(),
		defaultWorkbinId: uuid.NewString(),

		defaultDurationS:    86400,
		defaultExpirationS:  86400,
		defaultDueDurationS: 86400,
		defaultPriority:     100,
		defaultTtlS:         86400,

		defaultLanguageId: uuid.NewString(),
		defaultQueueId:    uuid.NewString(),
		defaultSkillIds:   []string{uuid.NewString(), uuid.NewString()},
		assignmentEnabled: false,

		schemaId:      uuid.NewString(),
		schemaVersion: 1,
	}

	taskProxy := &TaskManagementWorktypeProxy{}

	taskProxy.createTaskManagementWorktypeAttr = func(ctx context.Context, p *TaskManagementWorktypeProxy, create *platformclientv2.Worktypecreate) (*platformclientv2.Worktype, *platformclientv2.APIResponse, error) {
		assert.Equal(t, wt.name, *create.Name, "wt.Name check failed in create createTaskManagementWorktypeAttr")
		assert.Equal(t, wt.description, *create.Description, "wt.Description check failed in create createTaskManagementWorktypeAttr")
		assert.Equal(t, wt.divisionId, *create.DivisionId, "wt.divisionId check failed in create createTaskManagementWorktypeAttr")
		assert.Equal(t, wt.defaultWorkbinId, *create.DefaultWorkbinId, "wt.defaultWorkbinId check failed in create createTaskManagementWorktypeAttr")
		assert.Equal(t, wt.defaultDurationS, *create.DefaultDurationSeconds, "wt.defaultDurationS check failed in create createTaskManagementWorktypeAttr")
		assert.Equal(t, wt.defaultExpirationS, *create.DefaultExpirationSeconds, "wt.defaultExpirationS check failed in create createTaskManagementWorktypeAttr")
		assert.Equal(t, wt.defaultDueDurationS, *create.DefaultDueDurationSeconds, "wt.defaultDueDurationS check failed in create createTaskManagementWorktypeAttr")
		assert.Equal(t, wt.defaultPriority, *create.DefaultPriority, "wt.defaultPriority check failed in create createTaskManagementWorktypeAttr")
		assert.Equal(t, wt.defaultTtlS, *create.DefaultTtlSeconds, "wt.defaultTtlS check failed in create createTaskManagementWorktypeAttr")
		assert.Equal(t, wt.defaultLanguageId, *create.DefaultLanguageId, "wt.defaultLanguageId check failed in create createTaskManagementWorktypeAttr")
		assert.Equal(t, wt.defaultQueueId, *create.DefaultQueueId, "wt.defaultQueueId check failed in create createTaskManagementWorktypeAttr")
		assert.ElementsMatch(t, wt.defaultSkillIds, *create.DefaultSkillIds, "wt.defaultSkillIds check failed in create createTaskManagementWorktypeAttr")
		assert.Equal(t, wt.assignmentEnabled, *create.AssignmentEnabled, "wt.assignmentEnabled check failed in create createTaskManagementWorktypeAttr")
		assert.Equal(t, wt.schemaId, *create.SchemaId, "wt.schemaId check failed in create createTaskManagementWorktypeAttr")
		assert.Equal(t, wt.schemaVersion, *create.SchemaVersion, "wt.schemaVersion check failed in create createTaskManagementWorktypeAttr")

		return &platformclientv2.Worktype{
			Id:          &tId,
			Name:        &wt.name,
			Description: &wt.description,
			Division: &platformclientv2.Division{
				Id: &wt.divisionId,
			},

			DefaultDurationSeconds:    &wt.defaultDueDurationS,
			DefaultExpirationSeconds:  &wt.defaultExpirationS,
			DefaultDueDurationSeconds: &wt.defaultDueDurationS,
			DefaultPriority:           &wt.defaultPriority,
			DefaultTtlSeconds:         &wt.defaultTtlS,

			DefaultLanguage: &platformclientv2.Languagereference{
				Id: &wt.defaultLanguageId,
			},
			DefaultQueue: &platformclientv2.Workitemqueuereference{
				Id: &wt.defaultQueueId,
			},
			AssignmentEnabled: &wt.assignmentEnabled,
			Schema: &platformclientv2.Workitemschema{
				Id:      &wt.schemaId,
				Version: &wt.schemaVersion,
			},
		}, nil, nil
	}

	taskProxy.getTaskManagementWorktypeByIdAttr = func(ctx context.Context, p *TaskManagementWorktypeProxy, id string) (*platformclientv2.Worktype, *platformclientv2.APIResponse, error) {
		assert.Equal(t, tId, id)

		// The expected final form of the worktype
		wt := &platformclientv2.Worktype{
			Id:          &tId,
			Name:        &wt.name,
			Description: &wt.description,
			Division: &platformclientv2.Division{
				Id: &wt.divisionId,
			},

			DefaultDurationSeconds:    &wt.defaultDueDurationS,
			DefaultExpirationSeconds:  &wt.defaultExpirationS,
			DefaultDueDurationSeconds: &wt.defaultDueDurationS,
			DefaultPriority:           &wt.defaultPriority,
			DefaultTtlSeconds:         &wt.defaultTtlS,

			DefaultLanguage: &platformclientv2.Languagereference{
				Id: &wt.defaultLanguageId,
			},
			DefaultQueue: &platformclientv2.Workitemqueuereference{
				Id: &wt.defaultQueueId,
			},
			DefaultSkills: &[]platformclientv2.Routingskillreference{
				{
					Id: &wt.defaultSkillIds[0],
				},
				{
					Id: &wt.defaultSkillIds[1],
				},
			},
			AssignmentEnabled: &wt.assignmentEnabled,
			Schema: &platformclientv2.Workitemschema{
				Id:      &wt.schemaId,
				Version: &wt.schemaVersion,
			},
		}

		apiResponse := &platformclientv2.APIResponse{StatusCode: http.StatusOK}

		return wt, apiResponse, nil
	}

	internalProxy = taskProxy
	defer func() { internalProxy = nil }()

	ctx := context.Background()
	gcloud := &provider.ProviderMeta{ClientConfig: &platformclientv2.Configuration{}}

	//Grab our defined schema
	resourceSchema := ResourceTaskManagementWorktype().Schema

	//Setup a map of values
	resourceDataMap := buildWorktypeResourceMap(tId, wt)

	//Found this gem TestResourceDataRaw here: https://github.com/hashicorp/terraform/issues/890
	d := schema.TestResourceDataRaw(t, resourceSchema, resourceDataMap)
	d.SetId(tId)

	diag := createTaskManagementWorktype(ctx, d, gcloud)
	assert.Equal(t, false, diag.HasError(), diag)
	assert.Equal(t, tId, d.Id())
}

func TestUnitResourceWorktypeRead(t *testing.T) {
	tId := uuid.NewString()

	// The complete configuration for the worktype
	wt := &worktypeConfig{
		name:             "tf_worktype_" + uuid.NewString(),
		description:      "worktype created for CX as Code test case",
		divisionId:       uuid.NewString(),
		defaultWorkbinId: uuid.NewString(),

		defaultDurationS:    86400,
		defaultExpirationS:  86400,
		defaultDueDurationS: 86400,
		defaultPriority:     100,
		defaultTtlS:         86400,

		defaultLanguageId: uuid.NewString(),
		defaultQueueId:    uuid.NewString(),
		defaultSkillIds:   []string{uuid.NewString(), uuid.NewString()},
		assignmentEnabled: false,

		schemaId:      uuid.NewString(),
		schemaVersion: 1,
	}

	taskProxy := &TaskManagementWorktypeProxy{}

	taskProxy.getTaskManagementWorktypeByIdAttr = func(ctx context.Context, p *TaskManagementWorktypeProxy, id string) (*platformclientv2.Worktype, *platformclientv2.APIResponse, error) {
		assert.Equal(t, tId, id)

		// The expected final form of the worktype
		wt := &platformclientv2.Worktype{
			Id:          &tId,
			Name:        &wt.name,
			Description: &wt.description,
			Division: &platformclientv2.Division{
				Id: &wt.divisionId,
			},

			DefaultDurationSeconds:    &wt.defaultDueDurationS,
			DefaultExpirationSeconds:  &wt.defaultExpirationS,
			DefaultDueDurationSeconds: &wt.defaultDueDurationS,
			DefaultPriority:           &wt.defaultPriority,
			DefaultTtlSeconds:         &wt.defaultTtlS,

			DefaultLanguage: &platformclientv2.Languagereference{
				Id: &wt.defaultLanguageId,
			},
			DefaultQueue: &platformclientv2.Workitemqueuereference{
				Id: &wt.defaultQueueId,
			},
			DefaultSkills: &[]platformclientv2.Routingskillreference{
				{
					Id: &wt.defaultSkillIds[0],
				},
				{
					Id: &wt.defaultSkillIds[1],
				},
			},
			AssignmentEnabled: &wt.assignmentEnabled,
			Schema: &platformclientv2.Workitemschema{
				Id:      &wt.schemaId,
				Version: &wt.schemaVersion,
			},
		}

		apiResponse := &platformclientv2.APIResponse{StatusCode: http.StatusOK}

		return wt, apiResponse, nil
	}
	internalProxy = taskProxy
	defer func() { internalProxy = nil }()

	ctx := context.Background()
	gcloud := &provider.ProviderMeta{ClientConfig: &platformclientv2.Configuration{}}

	//Grab our defined schema
	resourceSchema := ResourceTaskManagementWorktype().Schema

	//Setup a map of values
	resourceDataMap := buildWorktypeResourceMap(tId, wt)

	//Found this gem TestResourceDataRaw here: https://github.com/hashicorp/terraform/issues/890
	d := schema.TestResourceDataRaw(t, resourceSchema, resourceDataMap)
	d.SetId(tId)

	diag := readTaskManagementWorktype(ctx, d, gcloud)

	assert.Equal(t, false, diag.HasError(), diag)
	assert.Equal(t, tId, d.Id())
	assert.Equal(t, wt.name, d.Get("name").(string))
	assert.Equal(t, wt.description, d.Get("description").(string))
	assert.Equal(t, wt.divisionId, d.Get("division_id").(string))
	assert.Equal(t, wt.defaultWorkbinId, d.Get("default_workbin_id").(string))
	assert.Equal(t, wt.defaultDurationS, d.Get("default_duration_seconds").(int))
	assert.Equal(t, wt.defaultExpirationS, d.Get("default_expiration_seconds").(int))
	assert.Equal(t, wt.defaultDueDurationS, d.Get("default_due_duration_seconds").(int))
	assert.Equal(t, wt.defaultPriority, d.Get("default_priority").(int))
	assert.Equal(t, wt.defaultTtlS, d.Get("default_ttl_seconds").(int))
	assert.Equal(t, wt.defaultLanguageId, d.Get("default_language_id").(string))
	assert.Equal(t, wt.defaultQueueId, d.Get("default_queue_id").(string))
	assert.ElementsMatch(t, wt.defaultSkillIds, d.Get("default_skills_ids").([]interface{}))
	assert.Equal(t, wt.assignmentEnabled, d.Get("assignment_enabled").(bool))
	assert.Equal(t, wt.schemaId, d.Get("schema_id").(string))
	assert.Equal(t, wt.schemaVersion, d.Get("schema_version").(int))
}

func TestUnitResourceWorktypeUpdate(t *testing.T) {
	tId := uuid.NewString()

	// The complete configuration for the worktype
	wt := &worktypeConfig{
		name:             "tf_worktype_" + uuid.NewString(),
		description:      "worktype created for CX as Code test case",
		divisionId:       uuid.NewString(),
		defaultWorkbinId: uuid.NewString(),

		defaultDurationS:    86400,
		defaultExpirationS:  86400,
		defaultDueDurationS: 86400,
		defaultPriority:     100,
		defaultTtlS:         86400,

		defaultLanguageId: uuid.NewString(),
		defaultQueueId:    uuid.NewString(),
		defaultSkillIds:   []string{uuid.NewString(), uuid.NewString()},
		assignmentEnabled: false,

		schemaId:      uuid.NewString(),
		schemaVersion: 1,
	}

	taskProxy := &TaskManagementWorktypeProxy{}

	taskProxy.updateTaskManagementWorktypeAttr = func(ctx context.Context, p *TaskManagementWorktypeProxy, id string, update *platformclientv2.Worktypeupdate) (*platformclientv2.Worktype, *platformclientv2.APIResponse, error) {
		assert.Equal(t, tId, id)
		assert.Equal(t, wt.name, *update.Name, "wt.Name check failed in create createTaskManagementWorktypeAttr")

		return &platformclientv2.Worktype{
			Id:          &tId,
			Name:        &wt.name,
			Description: &wt.description,
			Division: &platformclientv2.Division{
				Id: &wt.divisionId,
			},

			DefaultDurationSeconds:    &wt.defaultDueDurationS,
			DefaultExpirationSeconds:  &wt.defaultExpirationS,
			DefaultDueDurationSeconds: &wt.defaultDueDurationS,
			DefaultPriority:           &wt.defaultPriority,
			DefaultTtlSeconds:         &wt.defaultTtlS,

			DefaultLanguage: &platformclientv2.Languagereference{
				Id: &wt.defaultLanguageId,
			},
			DefaultQueue: &platformclientv2.Workitemqueuereference{
				Id: &wt.defaultQueueId,
			},
			AssignmentEnabled: &wt.assignmentEnabled,
			Schema: &platformclientv2.Workitemschema{
				Id:      &wt.schemaId,
				Version: &wt.schemaVersion,
			},
		}, nil, nil
	}

	// The final complete worktype for read
	taskProxy.getTaskManagementWorktypeByIdAttr = func(ctx context.Context, p *TaskManagementWorktypeProxy, id string) (*platformclientv2.Worktype, *platformclientv2.APIResponse, error) {
		assert.Equal(t, tId, id)

		// The expected final form of the worktype
		wt := &platformclientv2.Worktype{
			Id:          &tId,
			Name:        &wt.name,
			Description: &wt.description,
			Division: &platformclientv2.Division{
				Id: &wt.divisionId,
			},

			DefaultDurationSeconds:    &wt.defaultDueDurationS,
			DefaultExpirationSeconds:  &wt.defaultExpirationS,
			DefaultDueDurationSeconds: &wt.defaultDueDurationS,
			DefaultPriority:           &wt.defaultPriority,
			DefaultTtlSeconds:         &wt.defaultTtlS,

			DefaultLanguage: &platformclientv2.Languagereference{
				Id: &wt.defaultLanguageId,
			},
			DefaultQueue: &platformclientv2.Workitemqueuereference{
				Id: &wt.defaultQueueId,
			},
			DefaultSkills: &[]platformclientv2.Routingskillreference{
				{
					Id: &wt.defaultSkillIds[0],
				},
				{
					Id: &wt.defaultSkillIds[1],
				},
			},
			AssignmentEnabled: &wt.assignmentEnabled,
			Schema: &platformclientv2.Workitemschema{
				Id:      &wt.schemaId,
				Version: &wt.schemaVersion,
			},
		}

		apiResponse := &platformclientv2.APIResponse{StatusCode: http.StatusOK}

		return wt, apiResponse, nil
	}

	internalProxy = taskProxy
	defer func() { internalProxy = nil }()

	ctx := context.Background()
	gcloud := &provider.ProviderMeta{ClientConfig: &platformclientv2.Configuration{}}

	//Grab our defined schema
	resourceSchema := ResourceTaskManagementWorktype().Schema

	//Setup a map of values
	resourceDataMap := buildWorktypeResourceMap(tId, wt)

	//Found this gem TestResourceDataRaw here: https://github.com/hashicorp/terraform/issues/890
	d := schema.TestResourceDataRaw(t, resourceSchema, resourceDataMap)
	d.SetId(tId)

	diag := updateTaskManagementWorktype(ctx, d, gcloud)
	assert.Equal(t, false, diag.HasError(), diag)
	assert.Equal(t, tId, d.Id())
}

func TestUnitResourceWorktypeDelete(t *testing.T) {
	tId := uuid.NewString()
	wt := &worktypeConfig{
		name:             "tf_worktype_" + uuid.NewString(),
		description:      "worktype created for CX as Code test case",
		defaultWorkbinId: uuid.NewString(),
		schemaId:         uuid.NewString(),
	}

	taskProxy := &TaskManagementWorktypeProxy{}

	taskProxy.deleteTaskManagementWorktypeAttr = func(ctx context.Context, p *TaskManagementWorktypeProxy, id string) (*platformclientv2.APIResponse, error) {
		assert.Equal(t, tId, id)

		apiResponse := &platformclientv2.APIResponse{StatusCode: http.StatusNoContent}
		return apiResponse, nil
	}

	taskProxy.getTaskManagementWorktypeByIdAttr = func(ctx context.Context, p *TaskManagementWorktypeProxy, id string) (worktype *platformclientv2.Worktype, response *platformclientv2.APIResponse, err error) {
		assert.Equal(t, tId, id)

		apiResponse := &platformclientv2.APIResponse{
			StatusCode: http.StatusNotFound,
			Error: &platformclientv2.APIError{
				Status: 404,
			},
		}
		return nil, apiResponse, fmt.Errorf("not found")
	}

	internalProxy = taskProxy
	defer func() { internalProxy = nil }()

	ctx := context.Background()
	gcloud := &provider.ProviderMeta{ClientConfig: &platformclientv2.Configuration{}}

	//Grab our defined schema
	resourceSchema := ResourceTaskManagementWorktype().Schema

	//Setup a map of values
	resourceDataMap := buildWorktypeResourceMap(tId, wt)

	//Found this gem TestResourceDataRaw here: https://github.com/hashicorp/terraform/issues/890
	d := schema.TestResourceDataRaw(t, resourceSchema, resourceDataMap)
	d.SetId(tId)

	diag := deleteTaskManagementWorktype(ctx, d, gcloud)
	assert.Nil(t, diag, diag)
	assert.Equal(t, tId, d.Id())
}

func buildWorktypeResourceMap(tId string, wt *worktypeConfig) map[string]interface{} {
	resourceDataMap := map[string]interface{}{
		"id":                           tId,
		"name":                         wt.name,
		"description":                  wt.description,
		"division_id":                  wt.divisionId,
		"default_workbin_id":           wt.defaultWorkbinId,
		"default_duration_seconds":     wt.defaultDurationS,
		"default_expiration_seconds":   wt.defaultExpirationS,
		"default_due_duration_seconds": wt.defaultDueDurationS,
		"default_priority":             wt.defaultPriority,
		"default_ttl_seconds":          wt.defaultTtlS,
		"default_language_id":          wt.defaultLanguageId,
		"default_queue_id":             wt.defaultQueueId,
		"default_skills_ids":           lists.StringListToInterfaceList(wt.defaultSkillIds),
		"assignment_enabled":           wt.assignmentEnabled,
		"schema_id":                    wt.schemaId,
		"schema_version":               wt.schemaVersion,
	}

	return resourceDataMap
}
