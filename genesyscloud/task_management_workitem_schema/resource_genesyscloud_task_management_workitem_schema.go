package task_management_workitem_schema

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v119/platformclientv2"

	"terraform-provider-genesyscloud/genesyscloud/consistency_checker"

	gcloud "terraform-provider-genesyscloud/genesyscloud"

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

	schemas, err := proxy.getAllTaskManagementWorkitemSchema(ctx)
	if err != nil {
		return nil, diag.Errorf("failed to get all workitem schemas: %v", err)
	}

	for _, schema := range *schemas {
		log.Printf("Dealing with task management workitem schema id: %s", *schema.Id)
		resources[*schema.Id] = &resourceExporter.ResourceMeta{Name: *schema.Name}
	}

	return resources, nil
}

// createTaskManagementWorkitemSchema is used by the task_management_workitem_schema resource to create Genesys cloud task management workitem schemas
func createTaskManagementWorkitemSchema(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*gcloud.ProviderMeta).ClientConfig
	proxy := getTaskManagementProxy(sdkConfig)

	dataSchema, err := BuildSdkWorkitemSchema(d, nil)
	if err != nil {
		return diag.Errorf("create: failed to build task management workitem schema: %s", err)
	}

	log.Printf("Creating task management workitem schema")
	schema, err := proxy.createTaskManagementWorkitemSchema(ctx, dataSchema)
	if err != nil {
		return diag.Errorf("failed to create task management workitem schema: %s", err)
	}

	d.SetId(*schema.Id)

	// If enabled is set to 'false' do an update call to the schema
	if enabled, ok := d.Get("enabled").(bool); ok && !enabled {
		log.Printf("Updating task management workitem schema: %s, to set 'enabled' to 'false'", *schema.Name)
		dataSchema.Version = platformclientv2.Int(1)
		_, err := proxy.updateTaskManagementWorkitemSchema(ctx, *schema.Id, dataSchema)
		if err != nil {
			return diag.Errorf("failed to update task management workitem schema: %s", err)
		}
		log.Printf("Updated newly created workitem schema: %s. 'enabled' set to to 'false'", *schema.Name)
	}

	log.Printf("Created task management workitem schema %s: %s", *schema.Name, *schema.Id)
	return readTaskManagementWorkitemSchema(ctx, d, meta)
}

// readTaskManagementWorkitemSchema is used by the task_management_workitem_schema resource to read a task management workitem schema from genesys cloud
func readTaskManagementWorkitemSchema(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*gcloud.ProviderMeta).ClientConfig
	proxy := getTaskManagementProxy(sdkConfig)

	log.Printf("Reading task management workitem schema %s", d.Id())

	return gcloud.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		schema, respCode, getErr := proxy.getTaskManagementWorkitemSchemaById(ctx, d.Id())
		if getErr != nil {
			if gcloud.IsStatus404ByInt(respCode) {
				return retry.RetryableError(fmt.Errorf("failed to read task management workitem schema %s: %s", d.Id(), getErr))
			}
			return retry.NonRetryableError(fmt.Errorf("failed to read task management workitem schema %s: %s", d.Id(), getErr))
		}

		cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceTaskManagementWorkitemSchema())

		schemaProps, err := json.Marshal(schema.JsonSchema.Properties)
		if err != nil {
			return retry.NonRetryableError(fmt.Errorf("error in reading json schema properties of %s: %v", *schema.Name, err))
		}
		var schemaPropsPtr *string
		if string(schemaProps) != gcloud.NullValue {
			schemaPropsStr := string(schemaProps)
			schemaPropsPtr = &schemaPropsStr
		}

		resourcedata.SetNillableValue(d, "name", schema.Name)
		resourcedata.SetNillableValue(d, "description", schema.JsonSchema.Description)
		resourcedata.SetNillableValue(d, "properties", schemaPropsPtr)
		resourcedata.SetNillableValue(d, "enabled", schema.Enabled)

		log.Printf("Read task management workitem schema %s %s", d.Id(), *schema.Name)
		return cc.CheckState()
	})
}

// updateTaskManagementWorkitemSchema is used by the task_management_workitem_schema resource to update a task management workitem schema in Genesys Cloud
func updateTaskManagementWorkitemSchema(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*gcloud.ProviderMeta).ClientConfig
	proxy := getTaskManagementProxy(sdkConfig)

	log.Printf("Getting version of workitem schema")
	curSchema, _, err := proxy.getTaskManagementWorkitemSchemaById(ctx, d.Id())
	if err != nil {
		return diag.Errorf("failed to update task management workitem schema: %s", err)
	}

	dataSchema, err := BuildSdkWorkitemSchema(d, curSchema.Version)
	if err != nil {
		return diag.Errorf("update: failed to build task management workitem schema: %s", err)
	}

	log.Printf("Updating task management workitem schema")
	updatedSchema, err := proxy.updateTaskManagementWorkitemSchema(ctx, d.Id(), dataSchema)
	if err != nil {
		return diag.Errorf("failed to update task management workitem schema: %s", err)
	}

	log.Printf("Updated task management workitem schema %s", *updatedSchema.Id)
	return readTaskManagementWorkitemSchema(ctx, d, meta)
}

// deleteTaskManagementWorkitemSchema is used by the task_management_workitem_schema resource to delete a task management workitem schema from Genesys cloud
func deleteTaskManagementWorkitemSchema(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*gcloud.ProviderMeta).ClientConfig
	proxy := getTaskManagementProxy(sdkConfig)

	_, err := proxy.deleteTaskManagementWorkitemSchema(ctx, d.Id())
	if err != nil {
		return diag.Errorf("failed to delete task management workitem schema %s: %s", d.Id(), err)
	}

	return gcloud.WithRetries(ctx, 180*time.Second, func() *retry.RetryError {
		isDeleted, respCode, err := proxy.getTaskManagementWorkitemSchemaDeletedStatus(ctx, d.Id())
		if err != nil {
			if gcloud.IsStatus404ByInt(respCode) {
				log.Printf("Deleted task management workitem schema %s", d.Id())
				return nil
			}
			return retry.NonRetryableError(fmt.Errorf("error deleting task management workitem schema %s: %s", d.Id(), err))
		}

		if isDeleted {
			log.Printf("Deleted task management workitem schema %s", d.Id())
			return nil
		}

		return retry.RetryableError(fmt.Errorf("task management workitem schema %s still exists", d.Id()))
	})
}
