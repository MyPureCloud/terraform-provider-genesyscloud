package task_management_workbin

import (
	"context"
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
The resource_genesyscloud_task_management_workbin.go contains all of the methods that perform the core logic for a resource.
*/

// getAllAuthTaskManagementWorkbin retrieves all of the task management workbin via Terraform in the Genesys Cloud and is used for the exporter
func getAllAuthTaskManagementWorkbins(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	proxy := getTaskManagementWorkbinProxy(clientConfig)
	resources := make(resourceExporter.ResourceIDMetaMap)

	workbins, err := proxy.getAllTaskManagementWorkbin(ctx)
	if err != nil {
		return nil, diag.Errorf("failed to get all workbins: %v", err)
	}

	for _, workbin := range *workbins {
		log.Printf("Dealing with task management workbin id: %s", *workbin.Id)
		resources[*workbin.Id] = &resourceExporter.ResourceMeta{Name: *workbin.Name}
	}

	return resources, nil
}

// createTaskManagementWorkbin is used by the task_management_workbin resource to create Genesys cloud task management workbin
func createTaskManagementWorkbin(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*gcloud.ProviderMeta).ClientConfig
	proxy := getTaskManagementWorkbinProxy(sdkConfig)

	taskManagementWorkbin := platformclientv2.Workbincreate{
		Name:        platformclientv2.String(d.Get("name").(string)),
		DivisionId:  platformclientv2.String(d.Get("division_id").(string)),
		Description: platformclientv2.String(d.Get("description").(string)),
	}

	log.Printf("Creating task management workbin %s", *taskManagementWorkbin.Name)
	workbin, err := proxy.createTaskManagementWorkbin(ctx, &taskManagementWorkbin)
	if err != nil {
		return diag.Errorf("failed to create task management workbin: %s", err)
	}

	d.SetId(*workbin.Id)
	log.Printf("Created task management workbin %s", *workbin.Id)
	return readTaskManagementWorkbin(ctx, d, meta)
}

// readTaskManagementWorkbin is used by the task_management_workbin resource to read an task management workbin from genesys cloud
func readTaskManagementWorkbin(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*gcloud.ProviderMeta).ClientConfig
	proxy := getTaskManagementWorkbinProxy(sdkConfig)

	log.Printf("Reading task management workbin %s", d.Id())

	return gcloud.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		workbin, respCode, getErr := proxy.getTaskManagementWorkbinById(ctx, d.Id())
		if getErr != nil {
			if gcloud.IsStatus404ByInt(respCode) {
				return retry.RetryableError(fmt.Errorf("failed to read task management workbin %s: %s", d.Id(), getErr))
			}
			return retry.NonRetryableError(fmt.Errorf("failed to read task management workbin %s: %s", d.Id(), getErr))
		}

		cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceTaskManagementWorkbin())

		resourcedata.SetNillableValue(d, "name", workbin.Name)
		resourcedata.SetNillableReferenceDivision(d, "division_id", workbin.Division)
		resourcedata.SetNillableValue(d, "description", workbin.Description)

		log.Printf("Read task management workbin %s %s", d.Id(), *workbin.Name)
		return cc.CheckState()
	})
}

// updateTaskManagementWorkbin is used by the task_management_workbin resource to update an task management workbin in Genesys Cloud
func updateTaskManagementWorkbin(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*gcloud.ProviderMeta).ClientConfig
	proxy := getTaskManagementWorkbinProxy(sdkConfig)

	taskManagementWorkbin := platformclientv2.Workbinupdate{
		Name:        platformclientv2.String(d.Get("name").(string)),
		Description: platformclientv2.String(d.Get("description").(string)),
	}

	log.Printf("Updating task management workbin %s", *taskManagementWorkbin.Name)
	workbin, err := proxy.updateTaskManagementWorkbin(ctx, d.Id(), &taskManagementWorkbin)
	if err != nil {
		return diag.Errorf("failed to update task management workbin: %s", err)
	}

	log.Printf("Updated task management workbin %s", *workbin.Id)
	return readTaskManagementWorkbin(ctx, d, meta)
}

// deleteTaskManagementWorkbin is used by the task_management_workbin resource to delete an task management workbin from Genesys cloud
func deleteTaskManagementWorkbin(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*gcloud.ProviderMeta).ClientConfig
	proxy := getTaskManagementWorkbinProxy(sdkConfig)

	_, err := proxy.deleteTaskManagementWorkbin(ctx, d.Id())
	if err != nil {
		return diag.Errorf("failed to delete task management workbin %s: %s", d.Id(), err)
	}

	return gcloud.WithRetries(ctx, 180*time.Second, func() *retry.RetryError {
		_, respCode, err := proxy.getTaskManagementWorkbinById(ctx, d.Id())

		if err != nil {
			if gcloud.IsStatus404ByInt(respCode) {
				log.Printf("Deleted task management workbin %s", d.Id())
				return nil
			}
			return retry.NonRetryableError(fmt.Errorf("error deleting task management workbin %s: %s", d.Id(), err))
		}

		return retry.RetryableError(fmt.Errorf("task management workbin %s still exists", d.Id()))
	})
}
