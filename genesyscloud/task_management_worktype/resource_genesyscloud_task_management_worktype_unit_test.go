package task_management_worktype

// build
import (
	"context"
	"fmt"

	"net/http"
	gcloud "terraform-provider-genesyscloud/genesyscloud"
	"terraform-provider-genesyscloud/genesyscloud/util/lists"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v121/platformclientv2"
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

		statuses: []worktypeStatusConfig{
			{
				id:                           uuid.NewString(),
				name:                         "Open Status",
				description:                  "Description of open status. Updated",
				defaultDestinationStatusName: "WIP",
				destinationStatusNames:       []string{"WIP", "Waiting Status"},
				transitionDelay:              120,
				statusTransitionTime:         "12:11:10",
				category:                     "Open",
			},
			{
				id:          uuid.NewString(),
				name:        "WIP",
				description: "Description of in progress status. Updated",
				category:    "InProgress",
			},
			{
				id:          uuid.NewString(),
				name:        "Waiting Status",
				description: "Description of waiting status. Updated",
				category:    "Waiting",
			},
			{
				id:          uuid.NewString(),
				name:        "Close Status",
				description: "Description of close status. Updated",
				category:    "Closed",
			},
		},
		defaultStatusName: "Open Status",

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

	taskProxy := &taskManagementWorktypeProxy{}

	taskProxy.createTaskManagementWorktypeAttr = func(ctx context.Context, p *taskManagementWorktypeProxy, create *platformclientv2.Worktypecreate) (*platformclientv2.Worktype, error) {
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
		}, nil
	}

	taskProxy.createTaskManagementWorktypeStatusAttr = func(ctx context.Context, p *taskManagementWorktypeProxy, worktypeId string, status *platformclientv2.Workitemstatuscreate) (*platformclientv2.Workitemstatus, error) {
		return &platformclientv2.Workitemstatus{
			Id:                           &worktypeId,
			Name:                         status.Name,
			Category:                     status.Category,
			Description:                  status.Description,
			StatusTransitionDelaySeconds: status.StatusTransitionDelaySeconds,
			StatusTransitionTime:         status.StatusTransitionTime,
		}, nil
	}

	taskProxy.updateTaskManagementWorktypeStatusAttr = func(ctx context.Context, p *taskManagementWorktypeProxy, worktypeId string, statusId string, statusUpdate *platformclientv2.Workitemstatusupdate) (*platformclientv2.Workitemstatus, error) {
		return &platformclientv2.Workitemstatus{
			Id:                  &statusId,
			Name:                statusUpdate.Name,
			DestinationStatuses: buildDestinationStatuses(statusUpdate.DestinationStatusIds),
			Description:         statusUpdate.Description,
			DefaultDestinationStatus: &platformclientv2.Workitemstatusreference{
				Id: statusUpdate.DefaultDestinationStatusId,
			},
			StatusTransitionDelaySeconds: statusUpdate.StatusTransitionDelaySeconds,
			StatusTransitionTime:         statusUpdate.StatusTransitionTime,
		}, nil
	}

	taskProxy.getTaskManagementWorktypeByIdAttr = func(ctx context.Context, p *taskManagementWorktypeProxy, id string) (*platformclientv2.Worktype, int, error) {
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

			Statuses: &[]platformclientv2.Workitemstatus{
				{
					Id:          &wt.statuses[0].id,
					Name:        &wt.statuses[0].name,
					Description: &wt.statuses[0].description,
					DefaultDestinationStatus: &platformclientv2.Workitemstatusreference{
						Id: wt.getStatusIdFromName(wt.statuses[0].defaultDestinationStatusName),
					},
					DestinationStatuses: &[]platformclientv2.Workitemstatusreference{
						{
							Id: wt.getStatusIdFromName(wt.statuses[0].destinationStatusNames[0]),
						},
						{
							Id: wt.getStatusIdFromName(wt.statuses[0].destinationStatusNames[1]),
						},
					},
					StatusTransitionDelaySeconds: &wt.statuses[0].transitionDelay,
					StatusTransitionTime:         &wt.statuses[0].statusTransitionTime,
					Category:                     &wt.statuses[0].category,
				},
				{
					Id:          &wt.statuses[1].id,
					Name:        &wt.statuses[1].name,
					Description: &wt.statuses[1].description,
					Category:    &wt.statuses[1].category,
				},
				{
					Id:          &wt.statuses[2].id,
					Name:        &wt.statuses[2].name,
					Description: &wt.statuses[2].description,
					Category:    &wt.statuses[2].category,
				},
				{
					Id:          &wt.statuses[3].id,
					Name:        &wt.statuses[3].name,
					Description: &wt.statuses[3].description,
					Category:    &wt.statuses[3].category,
				},
			},

			DefaultStatus: &platformclientv2.Workitemstatusreference{
				Id: wt.getStatusIdFromName(wt.defaultStatusName),
			},
		}

		apiResponse := &platformclientv2.APIResponse{StatusCode: http.StatusOK}

		return wt, apiResponse.StatusCode, nil
	}

	internalProxy = taskProxy
	defer func() { internalProxy = nil }()

	ctx := context.Background()
	gcloud := &gcloud.ProviderMeta{ClientConfig: &platformclientv2.Configuration{}}

	//Grab our defined schema
	resourceSchema := ResourceTaskManagementWorktype().Schema

	//Setup a map of values
	resourceDataMap := buildWorktypeResourceMap(tId, wt)

	//Found this gem TestResourceDataRaw here: https://github.com/hashicorp/terraform/issues/890
	d := schema.TestResourceDataRaw(t, resourceSchema, resourceDataMap)
	d.SetId(tId)

	diag := createTaskManagementWorktype(ctx, d, gcloud)
	assert.Equal(t, false, diag.HasError())
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

		statuses: []worktypeStatusConfig{
			{
				id:                           uuid.NewString(),
				name:                         "Open Status",
				description:                  "Description of open status. Updated",
				defaultDestinationStatusName: "WIP",
				destinationStatusNames:       []string{"WIP", "Waiting Status"},
				transitionDelay:              120,
				statusTransitionTime:         "12:11:10",
				category:                     "Open",
			},
			{
				id:          uuid.NewString(),
				name:        "WIP",
				description: "Description of in progress status. Updated",
				category:    "InProgress",
			},
			{
				id:          uuid.NewString(),
				name:        "Waiting Status",
				description: "Description of waiting status. Updated",
				category:    "Waiting",
			},
			{
				id:          uuid.NewString(),
				name:        "Close Status",
				description: "Description of close status. Updated",
				category:    "Closed",
			},
		},
		defaultStatusName: "Open Status",

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

	taskProxy := &taskManagementWorktypeProxy{}

	taskProxy.getTaskManagementWorktypeByIdAttr = func(ctx context.Context, p *taskManagementWorktypeProxy, id string) (*platformclientv2.Worktype, int, error) {
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

			Statuses: &[]platformclientv2.Workitemstatus{
				{
					Id:          &wt.statuses[0].id,
					Name:        &wt.statuses[0].name,
					Description: &wt.statuses[0].description,
					DefaultDestinationStatus: &platformclientv2.Workitemstatusreference{
						Id: wt.getStatusIdFromName(wt.statuses[0].defaultDestinationStatusName),
					},
					DestinationStatuses: &[]platformclientv2.Workitemstatusreference{
						{
							Id: wt.getStatusIdFromName(wt.statuses[0].destinationStatusNames[0]),
						},
						{
							Id: wt.getStatusIdFromName(wt.statuses[0].destinationStatusNames[1]),
						},
					},
					StatusTransitionDelaySeconds: &wt.statuses[0].transitionDelay,
					StatusTransitionTime:         &wt.statuses[0].statusTransitionTime,
					Category:                     &wt.statuses[0].category,
				},
				{
					Id:          &wt.statuses[1].id,
					Name:        &wt.statuses[1].name,
					Description: &wt.statuses[1].description,
					Category:    &wt.statuses[1].category,
				},
				{
					Id:          &wt.statuses[2].id,
					Name:        &wt.statuses[2].name,
					Description: &wt.statuses[2].description,
					Category:    &wt.statuses[2].category,
				},
				{
					Id:          &wt.statuses[3].id,
					Name:        &wt.statuses[3].name,
					Description: &wt.statuses[3].description,
					Category:    &wt.statuses[3].category,
				},
			},

			DefaultStatus: &platformclientv2.Workitemstatusreference{
				Id: wt.getStatusIdFromName(wt.defaultStatusName),
			},
		}

		apiResponse := &platformclientv2.APIResponse{StatusCode: http.StatusOK}

		return wt, apiResponse.StatusCode, nil
	}
	internalProxy = taskProxy
	defer func() { internalProxy = nil }()

	ctx := context.Background()
	gcloud := &gcloud.ProviderMeta{ClientConfig: &platformclientv2.Configuration{}}

	//Grab our defined schema
	resourceSchema := ResourceTaskManagementWorktype().Schema

	//Setup a map of values
	resourceDataMap := buildWorktypeResourceMap(tId, wt)

	//Found this gem TestResourceDataRaw here: https://github.com/hashicorp/terraform/issues/890
	d := schema.TestResourceDataRaw(t, resourceSchema, resourceDataMap)
	d.SetId(tId)

	diag := readTaskManagementWorktype(ctx, d, gcloud)

	assert.Equal(t, false, diag.HasError())
	assert.Equal(t, tId, d.Id())
	assert.Equal(t, wt.name, d.Get("name").(string))
	assert.Equal(t, wt.description, d.Get("description").(string))
	assert.Equal(t, wt.divisionId, d.Get("division_id").(string))
	assert.Equal(t, wt.defaultWorkbinId, d.Get("default_workbin_id").(string))
	assert.Equal(t, len(wt.statuses), d.Get("statuses").(*schema.Set).Len())
	assert.Equal(t, wt.defaultStatusName, d.Get("default_status_name").(string))
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

		statuses: []worktypeStatusConfig{
			{
				id:                           uuid.NewString(),
				name:                         "Open Status",
				description:                  "Description of open status. Updated",
				defaultDestinationStatusName: "WIP",
				destinationStatusNames:       []string{"WIP", "Waiting Status"},
				transitionDelay:              120,
				statusTransitionTime:         "12:11:10",
				category:                     "Open",
			},
			{
				id:          uuid.NewString(),
				name:        "WIP",
				description: "Description of in progress status. Updated",
				category:    "InProgress",
			},
			{
				id:          uuid.NewString(),
				name:        "Waiting Status",
				description: "Description of waiting status. Updated",
				category:    "Waiting",
			},
			{
				id:          uuid.NewString(),
				name:        "Close Status",
				description: "Description of close status. Updated",
				category:    "Closed",
			},
		},
		defaultStatusName: "Open Status",

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

	taskProxy := &taskManagementWorktypeProxy{}

	taskProxy.updateTaskManagementWorktypeAttr = func(ctx context.Context, p *taskManagementWorktypeProxy, id string, update *platformclientv2.Worktypeupdate) (*platformclientv2.Worktype, error) {
		assert.Equal(t, tId, id)
		assert.Equal(t, wt.name, *update.Name, "wt.Name check failed in create createTaskManagementWorktypeAttr")
		assert.Equal(t, wt.description, *update.Description, "wt.Description check failed in create createTaskManagementWorktypeAttr")
		assert.Equal(t, wt.defaultWorkbinId, *update.DefaultWorkbinId, "wt.defaultWorkbinId check failed in create createTaskManagementWorktypeAttr")
		assert.Equal(t, wt.defaultDurationS, *update.DefaultDurationSeconds, "wt.defaultDurationS check failed in create createTaskManagementWorktypeAttr")
		assert.Equal(t, wt.defaultExpirationS, *update.DefaultExpirationSeconds, "wt.defaultExpirationS check failed in create createTaskManagementWorktypeAttr")
		assert.Equal(t, wt.defaultDueDurationS, *update.DefaultDueDurationSeconds, "wt.defaultDueDurationS check failed in create createTaskManagementWorktypeAttr")
		assert.Equal(t, wt.defaultPriority, *update.DefaultPriority, "wt.defaultPriority check failed in create createTaskManagementWorktypeAttr")
		assert.Equal(t, wt.defaultTtlS, *update.DefaultTtlSeconds, "wt.defaultTtlS check failed in create createTaskManagementWorktypeAttr")
		assert.Equal(t, wt.defaultLanguageId, *update.DefaultLanguageId, "wt.defaultLanguageId check failed in create createTaskManagementWorktypeAttr")
		assert.Equal(t, wt.defaultQueueId, *update.DefaultQueueId, "wt.defaultQueueId check failed in create createTaskManagementWorktypeAttr")
		assert.ElementsMatch(t, wt.defaultSkillIds, *update.DefaultSkillIds, "wt.defaultSkillIds check failed in create createTaskManagementWorktypeAttr")
		assert.Equal(t, wt.assignmentEnabled, *update.AssignmentEnabled, "wt.assignmentEnabled check failed in create createTaskManagementWorktypeAttr")
		assert.Equal(t, wt.schemaVersion, *update.SchemaVersion, "wt.schemaVersion check failed in create createTaskManagementWorktypeAttr")

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
		}, nil
	}

	taskProxy.createTaskManagementWorktypeStatusAttr = func(ctx context.Context, p *taskManagementWorktypeProxy, worktypeId string, status *platformclientv2.Workitemstatuscreate) (*platformclientv2.Workitemstatus, error) {
		return &platformclientv2.Workitemstatus{
			Id:                           &worktypeId,
			Name:                         status.Name,
			Category:                     status.Category,
			Description:                  status.Description,
			StatusTransitionDelaySeconds: status.StatusTransitionDelaySeconds,
			StatusTransitionTime:         status.StatusTransitionTime,
		}, nil
	}

	taskProxy.updateTaskManagementWorktypeStatusAttr = func(ctx context.Context, p *taskManagementWorktypeProxy, worktypeId string, statusId string, statusUpdate *platformclientv2.Workitemstatusupdate) (*platformclientv2.Workitemstatus, error) {
		return &platformclientv2.Workitemstatus{
			Id:                  &statusId,
			Name:                statusUpdate.Name,
			DestinationStatuses: buildDestinationStatuses(statusUpdate.DestinationStatusIds),
			Description:         statusUpdate.Description,
			DefaultDestinationStatus: &platformclientv2.Workitemstatusreference{
				Id: statusUpdate.DefaultDestinationStatusId,
			},
			StatusTransitionDelaySeconds: statusUpdate.StatusTransitionDelaySeconds,
			StatusTransitionTime:         statusUpdate.StatusTransitionTime,
		}, nil
	}

	// The final complete worktype for read
	// This is where we'll be asserting the statuses
	taskProxy.getTaskManagementWorktypeByIdAttr = func(ctx context.Context, p *taskManagementWorktypeProxy, id string) (*platformclientv2.Worktype, int, error) {
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

			Statuses: &[]platformclientv2.Workitemstatus{
				{
					Id:          &wt.statuses[0].id,
					Name:        &wt.statuses[0].name,
					Description: &wt.statuses[0].description,
					DefaultDestinationStatus: &platformclientv2.Workitemstatusreference{
						Id: wt.getStatusIdFromName(wt.statuses[0].defaultDestinationStatusName),
					},
					DestinationStatuses: &[]platformclientv2.Workitemstatusreference{
						{
							Id: wt.getStatusIdFromName(wt.statuses[0].destinationStatusNames[0]),
						},
						{
							Id: wt.getStatusIdFromName(wt.statuses[0].destinationStatusNames[1]),
						},
					},
					StatusTransitionDelaySeconds: &wt.statuses[0].transitionDelay,
					StatusTransitionTime:         &wt.statuses[0].statusTransitionTime,
					Category:                     &wt.statuses[0].category,
				},
				{
					Id:          &wt.statuses[1].id,
					Name:        &wt.statuses[1].name,
					Description: &wt.statuses[1].description,
					Category:    &wt.statuses[1].category,
				},
				{
					Id:          &wt.statuses[2].id,
					Name:        &wt.statuses[2].name,
					Description: &wt.statuses[2].description,
					Category:    &wt.statuses[2].category,
				},
				{
					Id:          &wt.statuses[3].id,
					Name:        &wt.statuses[3].name,
					Description: &wt.statuses[3].description,
					Category:    &wt.statuses[3].category,
				},
			},

			DefaultStatus: &platformclientv2.Workitemstatusreference{
				Id: wt.getStatusIdFromName(wt.defaultStatusName),
			},
		}

		apiResponse := &platformclientv2.APIResponse{StatusCode: http.StatusOK}

		return wt, apiResponse.StatusCode, nil
	}

	internalProxy = taskProxy
	defer func() { internalProxy = nil }()

	ctx := context.Background()
	gcloud := &gcloud.ProviderMeta{ClientConfig: &platformclientv2.Configuration{}}

	//Grab our defined schema
	resourceSchema := ResourceTaskManagementWorktype().Schema

	//Setup a map of values
	resourceDataMap := buildWorktypeResourceMap(tId, wt)

	//Found this gem TestResourceDataRaw here: https://github.com/hashicorp/terraform/issues/890
	d := schema.TestResourceDataRaw(t, resourceSchema, resourceDataMap)
	d.SetId(tId)

	diag := updateTaskManagementWorktype(ctx, d, gcloud)
	assert.Equal(t, false, diag.HasError())
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

	taskProxy := &taskManagementWorktypeProxy{}

	taskProxy.deleteTaskManagementWorktypeAttr = func(ctx context.Context, p *taskManagementWorktypeProxy, id string) (int, error) {
		assert.Equal(t, tId, id)

		apiResponse := &platformclientv2.APIResponse{StatusCode: http.StatusNoContent}
		return apiResponse.StatusCode, nil
	}

	taskProxy.getTaskManagementWorktypeByIdAttr = func(ctx context.Context, p *taskManagementWorktypeProxy, id string) (worktype *platformclientv2.Worktype, responseCode int, err error) {
		assert.Equal(t, tId, id)

		apiResponse := &platformclientv2.APIResponse{
			StatusCode: http.StatusNotFound,
			Error: &platformclientv2.APIError{
				Status: 404,
			},
		}
		return nil, apiResponse.StatusCode, fmt.Errorf("not found")
	}

	internalProxy = taskProxy
	defer func() { internalProxy = nil }()

	ctx := context.Background()
	gcloud := &gcloud.ProviderMeta{ClientConfig: &platformclientv2.Configuration{}}

	//Grab our defined schema
	resourceSchema := ResourceTaskManagementWorktype().Schema

	//Setup a map of values
	resourceDataMap := buildWorktypeResourceMap(tId, wt)

	//Found this gem TestResourceDataRaw here: https://github.com/hashicorp/terraform/issues/890
	d := schema.TestResourceDataRaw(t, resourceSchema, resourceDataMap)
	d.SetId(tId)

	diag := deleteTaskManagementWorktype(ctx, d, gcloud)
	assert.Nil(t, diag)
	assert.Equal(t, tId, d.Id())
}

func buildWorktypeResourceMap(tId string, wt *worktypeConfig) map[string]interface{} {
	resourceDataMap := map[string]interface{}{
		"id":                           tId,
		"name":                         wt.name,
		"description":                  wt.description,
		"division_id":                  wt.divisionId,
		"statuses":                     buildWorktypeStatusesList(&wt.statuses),
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

func buildWorktypeStatusesList(statuses *[]worktypeStatusConfig) []interface{} {
	statusList := []interface{}{}
	for _, s := range *statuses {
		statusList = append(statusList, buildWorktypeStatusResourceMap(&s))
	}

	return statusList
}

func buildWorktypeStatusResourceMap(status *worktypeStatusConfig) map[string]interface{} {
	return map[string]interface{}{
		"id":                              status.id,
		"name":                            status.name,
		"description":                     status.description,
		"category":                        status.category,
		"destination_status_names":        lists.StringListToInterfaceList(status.destinationStatusNames),
		"default_destination_status_name": status.defaultDestinationStatusName,
		"status_transition_delay_seconds": status.transitionDelay,
		"status_transition_time":          status.statusTransitionTime,
	}
}

func buildDestinationStatuses(statusIds *[]string) *[]platformclientv2.Workitemstatusreference {
	ret := []platformclientv2.Workitemstatusreference{}

	for _, sid := range *statusIds {
		ret = append(ret, platformclientv2.Workitemstatusreference{
			Id: &sid,
		})
	}

	return &ret
}
