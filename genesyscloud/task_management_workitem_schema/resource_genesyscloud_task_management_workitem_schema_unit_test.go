package task_management_workitem_schema

// build
import (
	"context"
	"encoding/json"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"

	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"
	"github.com/stretchr/testify/assert"
)

/** Unit Test **/
func TestUnitResourceWorkitemSchemaCreate(t *testing.T) {
	tId := uuid.NewString()
	tName := "Unit Test Schema"
	tDescription := "CX as Code Unit Test Workitem Schema"
	tEnabled := true
	tJsonSchema := platformclientv2.Jsonschemadocument{
		Title:       &tName,
		Description: &tDescription,
		Properties: &map[string]interface{}{
			"custom_attribute_text": map[string]interface{}{
				"allOf": []interface{}{
					map[string]interface{}{
						"$ref": "#/definitions/text",
					},
				},
				"title":       "custom_attribute",
				"description": "Custom attribute for text",
				"minLength":   float64(0),
				"maxLength":   float64(50),
			},
		},
	}

	taskProxy := &taskManagementProxy{}

	taskProxy.getTaskManagementWorkitemSchemaByIdAttr = func(ctx context.Context, p *taskManagementProxy, id string) (*platformclientv2.Dataschema, *platformclientv2.APIResponse, error) {
		assert.Equal(t, tId, id)
		schema := &platformclientv2.Dataschema{
			Name:       &tName,
			Enabled:    &tEnabled,
			JsonSchema: &tJsonSchema,
		}

		apiResponse := &platformclientv2.APIResponse{StatusCode: http.StatusOK}

		return schema, apiResponse, nil
	}

	taskProxy.createTaskManagementWorkitemSchemaAttr = func(ctx context.Context, p *taskManagementProxy, schemaCreate *platformclientv2.Dataschema) (*platformclientv2.Dataschema, *platformclientv2.APIResponse, error) {
		schema := platformclientv2.Dataschema{}

		assert.Equal(t, tName, *schemaCreate.Name, "schema.Name check failed in create createTaskManagementWorkitemSchemaAttr")
		assert.Equal(t, tDescription, *schemaCreate.JsonSchema.Description, "schema.JsonSchema.Description check failed in create createTaskManagementWorkitemSchemaAttr")
		assert.Equal(t, tName, *schemaCreate.JsonSchema.Title, "schema.JsonSchema.Title check failed in create createTaskManagementWorkitemSchemaAttr")
		assert.Equal(t, tEnabled, *schemaCreate.Enabled, "schema.Enabled check failed in create createTaskManagementWorkitemSchemaAttr")
		assert.Equal(t, *tJsonSchema.Properties, *schemaCreate.JsonSchema.Properties, "schema.JsonSchema check failed in create createTaskManagementWorkitemSchemaAttr")

		schema.Id = &tId
		schema.Name = &tName

		return &schema, nil, nil
	}

	internalProxy = taskProxy
	defer func() { internalProxy = nil }()

	ctx := context.Background()
	gcloud := &provider.ProviderMeta{ClientConfig: &platformclientv2.Configuration{}}

	//Grab our defined schema
	resourceSchema := ResourceTaskManagementWorkitemSchema().Schema

	//Setup a map of values
	tProperties, err := json.Marshal(*tJsonSchema.Properties)
	if err != nil {
		t.Errorf("failed to build properties for resource map: %v", err)
	}
	resourceDataMap := buildWorkitemSchemaResourceMap(tId, tName, tDescription, tEnabled, string(tProperties))

	//Found this gem TestResourceDataRaw here: https://github.com/hashicorp/terraform/issues/890
	d := schema.TestResourceDataRaw(t, resourceSchema, resourceDataMap)
	d.SetId(tId)

	diag := createTaskManagementWorkitemSchema(ctx, d, gcloud)
	assert.Equal(t, false, diag.HasError())
	assert.Equal(t, tId, d.Id())
}

func TestUnitResourceWorkitemSchemaRead(t *testing.T) {
	tId := uuid.NewString()
	tName := "Unit Test Schema"
	tDescription := "CX as Code Unit Test Workitem Schema"
	tEnabled := true
	tJsonSchema := platformclientv2.Jsonschemadocument{
		Title:       &tName,
		Description: &tDescription,
		Properties: &map[string]interface{}{
			"custom_attribute_text": map[string]interface{}{
				"allOf": []interface{}{
					map[string]string{
						"$ref": "#/definitions/text",
					},
				},
				"title":       "custom_attribute",
				"description": "Custom attribute for text",
				"minLength":   0,
				"maxLength":   50,
			},
		},
	}

	taskProxy := &taskManagementProxy{}

	taskProxy.getTaskManagementWorkitemSchemaByIdAttr = func(ctx context.Context, p *taskManagementProxy, id string) (*platformclientv2.Dataschema, *platformclientv2.APIResponse, error) {
		assert.Equal(t, tId, id)
		schema := &platformclientv2.Dataschema{
			Name:       &tName,
			Enabled:    &tEnabled,
			JsonSchema: &tJsonSchema,
		}

		apiResponse := &platformclientv2.APIResponse{StatusCode: http.StatusOK}
		return schema, apiResponse, nil
	}
	internalProxy = taskProxy
	defer func() { internalProxy = nil }()

	ctx := context.Background()
	gcloud := &provider.ProviderMeta{ClientConfig: &platformclientv2.Configuration{}}

	//Grab our defined schema
	resourceSchema := ResourceTaskManagementWorkitemSchema().Schema

	//Setup a map of values
	tProperties, err := json.Marshal(*tJsonSchema.Properties)
	if err != nil {
		t.Errorf("failed to build properties for resource map: %v", err)
	}
	resourceDataMap := buildWorkitemSchemaResourceMap(tId, tName, tDescription, tEnabled, string(tProperties))

	//Found this gem TestResourceDataRaw here: https://github.com/hashicorp/terraform/issues/890
	d := schema.TestResourceDataRaw(t, resourceSchema, resourceDataMap)
	d.SetId(tId)

	diag := readTaskManagementWorkitemSchema(ctx, d, gcloud)
	assert.Equal(t, false, diag.HasError())
	assert.Equal(t, tId, d.Id())
	assert.Equal(t, tName, d.Get("name").(string))
	assert.Equal(t, tDescription, d.Get("description").(string))
	assert.Equal(t, tEnabled, d.Get("enabled").(bool))
	assert.True(t, equivalentJsons(string(tProperties), d.Get("properties").(string)))
}

func TestUnitResourceWorkitemSchemaDelete(t *testing.T) {
	tId := uuid.NewString()
	tName := "Unit Test Schema"
	tDescription := "CX as Code Unit Test Workitem Schema"
	tEnabled := true
	tJsonSchema := platformclientv2.Jsonschemadocument{
		Title:       &tName,
		Description: &tDescription,
		Properties: &map[string]interface{}{
			"custom_attribute_text": map[string]interface{}{
				"allOf": []interface{}{
					map[string]string{
						"$ref": "#/definitions/text",
					},
				},
				"title":       "custom_attribute",
				"description": "Custom attribute for text",
				"minLength":   0,
				"maxLength":   50,
			},
		},
	}

	taskProxy := &taskManagementProxy{}

	taskProxy.deleteTaskManagementWorkitemSchemaAttr = func(ctx context.Context, p *taskManagementProxy, id string) (*platformclientv2.APIResponse, error) {
		assert.Equal(t, tId, id)

		apiResponse := &platformclientv2.APIResponse{StatusCode: http.StatusNoContent}
		return apiResponse, nil
	}

	taskProxy.getTaskManagementWorkitemSchemaDeletedStatusAttr = func(ctx context.Context, p *taskManagementProxy, schemaId string) (isDeleted bool, resp *platformclientv2.APIResponse, err error) {
		assert.Equal(t, tId, schemaId)

		apiResponse := &platformclientv2.APIResponse{StatusCode: http.StatusOK}
		return true, apiResponse, nil
	}

	internalProxy = taskProxy
	defer func() { internalProxy = nil }()

	ctx := context.Background()
	gcloud := &provider.ProviderMeta{ClientConfig: &platformclientv2.Configuration{}}

	//Grab our defined schema
	resourceSchema := ResourceTaskManagementWorkitemSchema().Schema

	//Setup a map of values
	tProperties, err := json.Marshal(*tJsonSchema.Properties)
	if err != nil {
		t.Errorf("failed to build properties for resource map: %v", err)
	}
	resourceDataMap := buildWorkitemSchemaResourceMap(tId, tName, tDescription, tEnabled, string(tProperties))

	//Found this gem TestResourceDataRaw here: https://github.com/hashicorp/terraform/issues/890
	d := schema.TestResourceDataRaw(t, resourceSchema, resourceDataMap)
	d.SetId(tId)

	diag := deleteTaskManagementWorkitemSchema(ctx, d, gcloud)
	assert.Nil(t, diag)
	assert.Equal(t, tId, d.Id())
}

func TestUnitResourceWorkitemSchemaUpdate(t *testing.T) {
	tId := uuid.NewString()
	tName := "Unit Test Schema"
	tDescription := "Updated description. CX as Code Unit Test Workitem Schema"
	tEnabled := true
	tJsonSchema := platformclientv2.Jsonschemadocument{
		Title:       &tName,
		Description: &tDescription,
		Properties: &map[string]interface{}{
			"custom_attribute_text": map[string]interface{}{
				"allOf": []interface{}{
					map[string]interface{}{
						"$ref": "#/definitions/text",
					},
				},
				"title":       "custom_attribute",
				"description": "Custom attribute for text",
				"minLength":   float64(0),
				"maxLength":   float64(50),
			},
		},
	}

	taskProxy := &taskManagementProxy{}

	taskProxy.getTaskManagementWorkitemSchemaByIdAttr = func(ctx context.Context, p *taskManagementProxy, id string) (*platformclientv2.Dataschema, *platformclientv2.APIResponse, error) {
		assert.Equal(t, tId, id)
		schema := &platformclientv2.Dataschema{
			Name:       &tName,
			Enabled:    &tEnabled,
			JsonSchema: &tJsonSchema,
		}

		apiResponse := &platformclientv2.APIResponse{StatusCode: http.StatusOK}

		return schema, apiResponse, nil
	}

	taskProxy.updateTaskManagementWorkitemSchemaAttr = func(ctx context.Context, p *taskManagementProxy, schemaId string, schemaCreate *platformclientv2.Dataschema) (*platformclientv2.Dataschema, *platformclientv2.APIResponse, error) {
		schema := platformclientv2.Dataschema{}

		assert.Equal(t, tName, *schemaCreate.Name, "schema.Name check failed in create createTaskManagementWorkitemSchemaAttr")
		assert.Equal(t, tDescription, *schemaCreate.JsonSchema.Description, "schema.JsonSchema.Description check failed in create createTaskManagementWorkitemSchemaAttr")
		assert.Equal(t, tName, *schemaCreate.JsonSchema.Title, "schema.JsonSchema.Title check failed in create createTaskManagementWorkitemSchemaAttr")
		assert.Equal(t, tEnabled, *schemaCreate.Enabled, "schema.Enabled check failed in create createTaskManagementWorkitemSchemaAttr")
		assert.Equal(t, *tJsonSchema.Properties, *schemaCreate.JsonSchema.Properties, "schema.JsonSchema check failed in create createTaskManagementWorkitemSchemaAttr")

		schema.Id = &tId
		schema.Name = &tName

		return &schema, nil, nil
	}

	internalProxy = taskProxy
	defer func() { internalProxy = nil }()

	ctx := context.Background()
	gcloud := &provider.ProviderMeta{ClientConfig: &platformclientv2.Configuration{}}

	//Grab our defined schema
	resourceSchema := ResourceTaskManagementWorkitemSchema().Schema

	//Setup a map of values
	tProperties, err := json.Marshal(*tJsonSchema.Properties)
	if err != nil {
		t.Errorf("failed to build properties for resource map: %v", err)
	}
	resourceDataMap := buildWorkitemSchemaResourceMap(tId, tName, tDescription, tEnabled, string(tProperties))

	//Found this gem TestResourceDataRaw here: https://github.com/hashicorp/terraform/issues/890
	d := schema.TestResourceDataRaw(t, resourceSchema, resourceDataMap)
	d.SetId(tId)

	diag := updateTaskManagementWorkitemSchema(ctx, d, gcloud)
	assert.Equal(t, false, diag.HasError())
	assert.Equal(t, tId, d.Id())
	assert.Equal(t, tDescription, d.Get("description").(string))
}

func buildWorkitemSchemaResourceMap(tId string, tName string, tDescription string, tEnabled bool, tProperties string) map[string]interface{} {
	resourceDataMap := map[string]interface{}{
		"id":          tId,
		"name":        tName,
		"description": tDescription,
		"enabled":     tEnabled,
		"properties":  tProperties,
	}

	return resourceDataMap
}

func equivalentJsons(json1, json2 string) bool {
	return util.EquivalentJsons(json1, json2)
}
