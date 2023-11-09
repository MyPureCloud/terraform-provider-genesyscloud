package task_management_workitem_schema

import (
	"context"
	"fmt"
	"log"
	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v115/platformclientv2"

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
		resources[*schema.Id] = &resourceExporter.ResourceMeta{Name: *schema.Id}
	}

	return resources, nil
}

// createTaskManagementWorkitemSchema is used by the task_management_workitem_schema resource to create Genesys cloud task management workitem schemas
func createTaskManagementWorkitemSchema(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*gcloud.ProviderMeta).ClientConfig
	proxy := getTaskManagementProxy(sdkConfig)

	var jsonSchemaDoc platformclientv2.Jsonschemadocument
	jsonSchemaDoc.UnmarshalJSON(d.Get("json_schema").([]byte))

	log.Printf("Creating task management workitem schema")
	schema, err := proxy.createTaskManagementWorkitemSchema(ctx,
		&platformclientv2.Dataschema{
			JsonSchema: &jsonSchemaDoc,
			Enabled:    platformclientv2.Bool(d.Get("enabled").(bool)), // NOTE: At time of writing doesn't matter. Will be 'enabled' on creation.
		})
	if err != nil {
		return diag.Errorf("failed to create task management workitem schema: %s", err)
	}

	d.SetId(*schema.Id)

	// If enabled is set to 'false' do an update call to the schema
	if enabled, ok := d.Get("enabled").(bool); ok && !enabled {
		_, err := proxy.updateTaskManagementWorkitemSchema(ctx, *schema.Id,
			&platformclientv2.Dataschema{
				Version:    platformclientv2.Int(1),
				JsonSchema: &jsonSchemaDoc, // still required to pass
				Enabled:    platformclientv2.Bool(false),
			})
		if err != nil {
			return diag.Errorf("failed to update task management workitem schema: %s", err)
		}
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

		schemaStr := schema.JsonSchema.String()
		resourcedata.SetNillableValue(d, "json_schema", &schemaStr)
		resourcedata.SetNillableValue(d, "enable", schema.Enabled)

		log.Printf("Read task management workitem schema %s %s", d.Id(), *schema.Name)
		return cc.CheckState()
	})
}

// updateTaskManagementWorkitemSchema is used by the task_management_workitem_schema resource to update a task management workitem schema in Genesys Cloud
func updateTaskManagementWorkitemSchema(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*gcloud.ProviderMeta).ClientConfig
	proxy := getTaskManagementProxy(sdkConfig)

	var jsonSchemaDoc platformclientv2.Jsonschemadocument
	jsonSchemaDoc.UnmarshalJSON(d.Get("json_schema").([]byte))

	log.Printf("Getting version of workitem schema")
	curSchema, _, err := proxy.getTaskManagementWorkitemSchemaById(ctx, d.Id())
	if err != nil {
		return diag.Errorf("failed to update task management workitem schema: %s", err)
	}

	log.Printf("Updating task management workitem schema")
	updatedSchema, err := proxy.updateTaskManagementWorkitemSchema(ctx, d.Id(),
		&platformclientv2.Dataschema{
			Version:    platformclientv2.Int(*curSchema.Version),
			JsonSchema: &jsonSchemaDoc,
			Enabled:    platformclientv2.Bool(d.Get("enabled").(bool)),
		})
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
		_, respCode, err := proxy.getTaskManagementWorkitemSchemaById(ctx, d.Id())

		if err != nil {
			if gcloud.IsStatus404ByInt(respCode) {
				log.Printf("Deleted task management workitem schema %s", d.Id())
				return nil
			}
			return retry.NonRetryableError(fmt.Errorf("error deleting task management workitem schema %s: %s", d.Id(), err))
		}

		return retry.RetryableError(fmt.Errorf("task management workitem schema %s still exists", d.Id()))
	})
}
