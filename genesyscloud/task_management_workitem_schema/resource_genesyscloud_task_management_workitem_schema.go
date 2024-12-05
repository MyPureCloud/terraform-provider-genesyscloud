package task_management_workitem_schema

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"terraform-provider-genesyscloud/genesyscloud/util/constants"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v146/platformclientv2"

	"terraform-provider-genesyscloud/genesyscloud/consistency_checker"

	"terraform-provider-genesyscloud/genesyscloud/util/resourcedata"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
)

/*
The resource_genesyscloud_task_management_workitem_schema.go contains all of the methods that perform the core logic for a resource.
*/

// getAllTaskManagementWorkitemSchemas retrieves all of the task management workitem schemas via Terraform in the Genesys Cloud and is used for the exporter
func getAllTaskManagementWorkitemSchemas(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	proxy := getTaskManagementProxy(clientConfig)
	resources := make(resourceExporter.ResourceIDMetaMap)

	schemas, resp, err := proxy.getAllTaskManagementWorkitemSchema(ctx)
	if err != nil {
		return nil, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to get all workitem schemas error: %s", err), resp)
	}

	for _, schema := range *schemas {
		log.Printf("Dealing with task management workitem schema id: %s", *schema.Id)
		resources[*schema.Id] = &resourceExporter.ResourceMeta{BlockLabel: *schema.Name}
	}
	return resources, nil
}

// createTaskManagementWorkitemSchema is used by the task_management_workitem_schema resource to create Genesys cloud task management workitem schemas
func createTaskManagementWorkitemSchema(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getTaskManagementProxy(sdkConfig)

	dataSchema, err := BuildSdkWorkitemSchema(d, nil)
	if err != nil {
		return util.BuildDiagnosticError(ResourceType, fmt.Sprintf("create: failed to build task management workitem schema"), err)
	}

	log.Printf("Creating task management workitem schema")
	schema, resp, err := proxy.createTaskManagementWorkitemSchema(ctx, dataSchema)
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to create task management workitem schema %s error: %s", *dataSchema.Name, err), resp)
	}

	d.SetId(*schema.Id)

	// If enabled is set to 'false' do an update call to the schema
	if enabled, ok := d.Get("enabled").(bool); ok && !enabled {
		log.Printf("Updating task management workitem schema: %s, to set 'enabled' to 'false'", *schema.Name)
		dataSchema.Version = platformclientv2.Int(1)
		_, resp, err := proxy.updateTaskManagementWorkitemSchema(ctx, *schema.Id, dataSchema)
		if err != nil {
			return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to update task management workitem schema %s error: %s", d.Id(), err), resp)
		}
		log.Printf("Updated newly created workitem schema: %s. 'enabled' set to to 'false'", *schema.Name)
	}

	log.Printf("Created task management workitem schema %s: %s", *schema.Name, *schema.Id)
	return readTaskManagementWorkitemSchema(ctx, d, meta)
}

// readTaskManagementWorkitemSchema is used by the task_management_workitem_schema resource to read a task management workitem schema from genesys cloud
func readTaskManagementWorkitemSchema(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getTaskManagementProxy(sdkConfig)
	cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceTaskManagementWorkitemSchema(), constants.ConsistencyChecks(), ResourceType)

	log.Printf("Reading task management workitem schema %s", d.Id())

	return util.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		schema, resp, getErr := proxy.getTaskManagementWorkitemSchemaById(ctx, d.Id())
		if getErr != nil {
			if util.IsStatus404(resp) {
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("failed to read task management workitem schema %s | error: %s", d.Id(), getErr), resp))
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("failed to read task management workitem schema %s | error: %s", d.Id(), getErr), resp))
		}

		schemaProps, err := json.Marshal(schema.JsonSchema.Properties)
		if err != nil {
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("error in reading json schema properties of %s | error: %v", *schema.Name, err), resp))
		}
		var schemaPropsPtr *string
		if string(schemaProps) != util.NullValue {
			schemaPropsStr := string(schemaProps)
			schemaPropsPtr = &schemaPropsStr
		}

		resourcedata.SetNillableValue(d, "name", schema.Name)
		resourcedata.SetNillableValue(d, "description", schema.JsonSchema.Description)
		resourcedata.SetNillableValue(d, "properties", schemaPropsPtr)
		resourcedata.SetNillableValue(d, "enabled", schema.Enabled)

		log.Printf("Read task management workitem schema %s %s", d.Id(), *schema.Name)
		return cc.CheckState(d)
	})
}

// updateTaskManagementWorkitemSchema is used by the task_management_workitem_schema resource to update a task management workitem schema in Genesys Cloud
func updateTaskManagementWorkitemSchema(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getTaskManagementProxy(sdkConfig)

	log.Printf("Getting version of workitem schema")
	curSchema, resp, err := proxy.getTaskManagementWorkitemSchemaById(ctx, d.Id())
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to get task management workitem schema By id %s error: %s", d.Id(), err), resp)
	}

	dataSchema, err := BuildSdkWorkitemSchema(d, curSchema.Version)
	if err != nil {
		return util.BuildDiagnosticError(ResourceType, fmt.Sprintf("update: failed to build task management workitem schema"), err)
	}

	log.Printf("Updating task management workitem schema")
	updatedSchema, resp, err := proxy.updateTaskManagementWorkitemSchema(ctx, d.Id(), dataSchema)
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to update task management workitem schema %s error: %s", d.Id(), err), resp)
	}

	log.Printf("Updated task management workitem schema %s", *updatedSchema.Id)
	return readTaskManagementWorkitemSchema(ctx, d, meta)
}

// deleteTaskManagementWorkitemSchema is used by the task_management_workitem_schema resource to delete a task management workitem schema from Genesys cloud
func deleteTaskManagementWorkitemSchema(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getTaskManagementProxy(sdkConfig)

	resp, err := proxy.deleteTaskManagementWorkitemSchema(ctx, d.Id())
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to delete task management workitem schema %s error: %s", d.Id(), err), resp)
	}

	return util.WithRetries(ctx, 180*time.Second, func() *retry.RetryError {
		isDeleted, resp, err := proxy.getTaskManagementWorkitemSchemaDeletedStatus(ctx, d.Id())
		if err != nil {
			if util.IsStatus404(resp) {
				log.Printf("Deleted task management workitem schema %s", d.Id())
				return nil
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("error deleting task management workitem schema %s | error: %s", d.Id(), err), resp))
		}

		if isDeleted {
			log.Printf("Deleted task management workitem schema %s", d.Id())
			return nil
		}
		return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("task management workitem schema %s still exists", d.Id()), resp))
	})
}
