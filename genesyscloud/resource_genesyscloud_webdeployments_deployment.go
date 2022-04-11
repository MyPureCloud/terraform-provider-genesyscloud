package genesyscloud

import (
	"context"
	"errors"
	"fmt"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/consistency_checker"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/mypurecloud/platform-client-sdk-go/v67/platformclientv2"
)

func getAllWebDeployments(ctx context.Context, clientConfig *platformclientv2.Configuration) (ResourceIDMetaMap, diag.Diagnostics) {
	resources := make(ResourceIDMetaMap)
	webDeploymentsAPI := platformclientv2.NewWebDeploymentsApiWithConfig(clientConfig)

	deployments, _, getErr := webDeploymentsAPI.GetWebdeploymentsDeployments()
	if getErr != nil {
		return nil, diag.Errorf("Failed to get web deployments: %v", getErr)
	}

	for _, deployment := range *deployments.Entities {
		resources[*deployment.Id] = &ResourceMeta{Name: *deployment.Name}
	}

	return resources, nil
}

func webDeploymentExporter() *ResourceExporter {
	return &ResourceExporter{
		GetResourcesFunc: getAllWithPooledClient(getAllWebDeployments),
		RefAttrs: map[string]*RefAttrSettings{
			"flow_id":          {RefType: "genesyscloud_flow"},
			"configuration.id": {RefType: "genesyscloud_webdeployments_configuration"},
		},
	}
}

func resourceWebDeployment() *schema.Resource {
	return &schema.Resource{
		Description: "Genesys Cloud Web Deployment",

		CreateContext: createWithPooledClient(createWebDeployment),
		ReadContext:   readWithPooledClient(readWebDeployment),
		UpdateContext: updateWithPooledClient(updateWebDeployment),
		DeleteContext: deleteWithPooledClient(deleteWebDeployment),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "Deployment name",
				Type:        schema.TypeString,
				Required:    true,
			},
			"description": {
				Description: "Deployment description",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"allow_all_domains": {
				Description: "Whether all domains are allowed or not. allowedDomains must be empty when this is true.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
			},
			"allowed_domains": {
				Description: "The list of domains that are approved to use this deployment; the list will be added to CORS headers for ease of web use.",
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"flow_id": {
				Description: "A reference to the inboundshortmessage flow used by this deployment.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"status": {
				Description: "The current status of the deployment. Valid values: Pending, Active, Inactive, Error, Deleting.",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				ValidateFunc: validation.StringInSlice([]string{
					"Pending",
					"Active",
					"Inactive",
					"Error",
					"Deleting",
				}, false),
				DiffSuppressFunc: validateDeploymentStatusChange,
			},
			"configuration": {
				Description: "The published configuration version used by this deployment",
				Type:        schema.TypeList,
				Required:    true,
				MaxItems:    1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Required: true,
						},
						"version": {
							Type:             schema.TypeString,
							Optional:         true,
							Computed:         true,
							DiffSuppressFunc: alwaysDifferent, // The newly-computed configuration version is not available when computing the diff so we assume it will be different
						},
					},
				},
			},
		},
	}
}

func alwaysDifferent(k, old, new string, d *schema.ResourceData) bool {
	return false
}

func validateDeploymentStatusChange(k, old, new string, d *schema.ResourceData) bool {
	// Deployments will begin in a pending status and may or may not make it to active (or error) by the time we retrieve their state,
	// so allow the status to change from pending to a less ephemeral status
	return old == "Pending"
}

func createWebDeployment(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)
	description := d.Get("description").(string)
	allowAllDomains := d.Get("allow_all_domains").(bool)
	allowedDomains := interfaceListToStrings(d.Get("allowed_domains").([]interface{}))

	err := validAllowedDomainsSettings(d)
	if err != nil {
		return diag.Errorf("Failed to create web deployment %s: %s", name, err)
	}

	sdkConfig := meta.(*providerMeta).ClientConfig
	api := platformclientv2.NewWebDeploymentsApiWithConfig(sdkConfig)

	log.Printf("Creating web deployment %s", name)

	configId := d.Get("configuration.0.id").(string)
	configVersion := d.Get("configuration.0.version").(string)

	flow := buildSdkDomainEntityRef(d, "flow_id")

	inputDeployment := platformclientv2.Webdeployment{
		Name: &name,
		Configuration: &platformclientv2.Webdeploymentconfigurationversionentityref{
			Id:      &configId,
			Version: &configVersion,
		},
		AllowAllDomains: &allowAllDomains,
		AllowedDomains:  &allowedDomains,
	}

	if description != "" {
		inputDeployment.Description = &description
	}

	if flow != nil {
		inputDeployment.Flow = flow
	}

	diagErr := withRetries(ctx, 30*time.Second, func() *resource.RetryError {
		deployment, resp, err := api.PostWebdeploymentsDeployments(inputDeployment)
		if err != nil {
			if isStatus400(resp) {
				return resource.RetryableError(fmt.Errorf("Failed to create web deployment %s: %s", name, err))
			}
			return resource.NonRetryableError(fmt.Errorf("Failed to create web deployment %s: %s", name, err))
		}

		d.SetId(*deployment.Id)

		log.Printf("Created web deployment %s %s", name, *deployment.Id)

		return nil
	})
	if diagErr != nil {
		return diagErr
	}

	activeError := waitForDeploymentToBeActive(ctx, api, d.Id())
	if activeError != nil {
		return diag.Errorf("Web deployment %s did not become active and could not be created", name)
	}

	return readWebDeployment(ctx, d, meta)
}

func waitForDeploymentToBeActive(ctx context.Context, api *platformclientv2.WebDeploymentsApi, id string) diag.Diagnostics {
	return withRetries(ctx, 60*time.Second, func() *resource.RetryError {
		deployment, resp, err := api.GetWebdeploymentsDeployment(id)
		if err != nil {
			if isStatus404(resp) {
				return resource.RetryableError(fmt.Errorf("Error verifying active status for new web deployment %s: %s", id, err))
			}
			return resource.NonRetryableError(fmt.Errorf("Error verifying active status for new web deployment %s: %s", id, err))
		}

		if *deployment.Status == "Active" {
			return nil
		}

		return resource.RetryableError(fmt.Errorf("Web deployment %s not active yet. Status: %s", id, *deployment.Status))
	})
}

func readWebDeployment(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*providerMeta).ClientConfig
	api := platformclientv2.NewWebDeploymentsApiWithConfig(sdkConfig)

	log.Printf("Reading web deployment %s", d.Id())
	return withRetriesForRead(ctx, d, func() *resource.RetryError {
		deployment, resp, getErr := api.GetWebdeploymentsDeployment(d.Id())
		if getErr != nil {
			if isStatus404(resp) {
				return resource.RetryableError(fmt.Errorf("Failed to read web deployment %s: %s", d.Id(), getErr))
			}
			return resource.NonRetryableError(fmt.Errorf("Failed to read web deployment %s: %s", d.Id(), getErr))
		}

		cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, resourceWebDeployment())
		d.Set("name", *deployment.Name)
		if deployment.Description != nil {
			d.Set("description", *deployment.Description)
		}
		d.Set("allow_all_domains", *deployment.AllowAllDomains)
		d.Set("configuration", flattenConfiguration(deployment.Configuration))
		d.Set("allowed_domains", *deployment.AllowedDomains)
		if deployment.AllowedDomains != nil && len(*deployment.AllowedDomains) > 0 {
			d.Set("allowed_domains", *deployment.AllowedDomains)
		}
		if deployment.Flow != nil {
			d.Set("flow_id", *deployment.Flow.Id)
		}
		d.Set("status", *deployment.Status)

		log.Printf("Read web deployment %s %s", d.Id(), *deployment.Name)
		return cc.CheckState()
	})
}

func flattenConfiguration(configuration *platformclientv2.Webdeploymentconfigurationversionentityref) []interface{} {
	return []interface{}{map[string]interface{}{
		"id":      *configuration.Id,
		"version": *configuration.Version,
	}}
}

func updateWebDeployment(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)
	description := d.Get("description").(string)
	allowAllDomains := d.Get("allow_all_domains").(bool)
	allowedDomains := interfaceListToStrings(d.Get("allowed_domains").([]interface{}))
	status := d.Get("status").(string)

	err := validAllowedDomainsSettings(d)
	if err != nil {
		return diag.Errorf("Failed to update web deployment %s: %s", name, err)
	}

	sdkConfig := meta.(*providerMeta).ClientConfig
	api := platformclientv2.NewWebDeploymentsApiWithConfig(sdkConfig)

	log.Printf("Updating web deployment %s", name)

	configId := d.Get("configuration.0.id").(string)
	configVersion := d.Get("configuration.0.version").(string)

	flow := buildSdkDomainEntityRef(d, "flow_id")

	inputDeployment := platformclientv2.Webdeployment{
		Name: &name,
		Configuration: &platformclientv2.Webdeploymentconfigurationversionentityref{
			Id:      &configId,
			Version: &configVersion,
		},
		AllowAllDomains: &allowAllDomains,
		AllowedDomains:  &allowedDomains,
		Status:          &status,
	}

	if description != "" {
		inputDeployment.Description = &description
	}

	if flow != nil {
		inputDeployment.Flow = flow
	}

	diagErr := withRetries(ctx, 30*time.Second, func() *resource.RetryError {
		_, resp, err := api.PutWebdeploymentsDeployment(d.Id(), inputDeployment)
		if err != nil {
			if isStatus400(resp) {
				return resource.RetryableError(fmt.Errorf("Error updating web deployment %s: %s", name, err))
			}
			return resource.NonRetryableError(fmt.Errorf("Error updating web deployment %s: %s", name, err))
		}

		return nil
	})
	if diagErr != nil {
		return diagErr
	}

	activeError := waitForDeploymentToBeActive(ctx, api, d.Id())
	if activeError != nil {
		return diag.Errorf("Web deployment %s did not become active and could not be created", name)
	}

	log.Printf("Finished updating web deployment %s", name)
	return readWebDeployment(ctx, d, meta)
}

func deleteWebDeployment(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)

	sdkConfig := meta.(*providerMeta).ClientConfig
	api := platformclientv2.NewWebDeploymentsApiWithConfig(sdkConfig)

	log.Printf("Deleting web deployment %s", name)
	_, err := api.DeleteWebdeploymentsDeployment(d.Id())

	if err != nil {
		return diag.Errorf("Failed to delete web deployment %s: %s", name, err)
	}

	return withRetries(ctx, 30*time.Second, func() *resource.RetryError {
		_, resp, err := api.GetWebdeploymentsDeployment(d.Id())
		if err != nil {
			if isStatus404(resp) {
				log.Printf("Deleted web deployment %s", d.Id())
				return nil
			}
			return resource.NonRetryableError(fmt.Errorf("Error deleting web deployment %s: %s", d.Id(), err))
		}

		return resource.RetryableError(fmt.Errorf("Web deployment %s still exists", d.Id()))
	})
}

func validAllowedDomainsSettings(d *schema.ResourceData) error {
	allowAllDomains := d.Get("allow_all_domains").(bool)
	_, allowedDomainsSet := d.GetOk("allowed_domains")

	if allowAllDomains && allowedDomainsSet {
		return errors.New("Allowed domains cannot be specified when all domains are allowed")
	}

	if !allowAllDomains && !allowedDomainsSet {
		return errors.New("Either allowed domains must be specified or all domains must be allowed")
	}

	return nil
}
