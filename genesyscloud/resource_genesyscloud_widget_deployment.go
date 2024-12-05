package genesyscloud

import (
	"context"
	"fmt"
	"log"
	"strings"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"terraform-provider-genesyscloud/genesyscloud/util/constants"
	"terraform-provider-genesyscloud/genesyscloud/util/resourcedata"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"terraform-provider-genesyscloud/genesyscloud/consistency_checker"

	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	"terraform-provider-genesyscloud/genesyscloud/util/lists"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/mypurecloud/platform-client-sdk-go/v146/platformclientv2"
)

const (
	V1         = "v1"
	V1HTTP     = "v1-http"
	V2         = "v2"
	THIRDPARTY = "third-party"
)

var (
	validClientTypes           = []string{V1, V1HTTP, V2, THIRDPARTY}
	clientConfigSchemaResource = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"webchat_skin": {
				Description: "Skin for the webchat user. (basic, modern-caret-skin)",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"authentication_url": {
				Description: "Url endpoint to perform_authentication",
				Type:        schema.TypeString,
				Optional:    true,
			},
		},
	}
)

func getAllWidgetDeployments(_ context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	resources := make(resourceExporter.ResourceIDMetaMap)
	widgetsAPI := platformclientv2.NewWidgetsApiWithConfig(clientConfig)
	widgetDeployments, resp, getErr := widgetsAPI.GetWidgetsDeployments()
	if getErr != nil {
		return nil, util.BuildAPIDiagnosticError("genesyscloud_widget_deployment", fmt.Sprintf("Failed to get page of widget deployment error: %s", getErr), resp)
	}

	for _, widgetDeployment := range *widgetDeployments.Entities {
		resources[*widgetDeployment.Id] = &resourceExporter.ResourceMeta{BlockLabel: *widgetDeployment.Name}
	}

	return resources, nil
}

func WidgetDeploymentExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: provider.GetAllWithPooledClient(getAllWidgetDeployments),
		RefAttrs: map[string]*resourceExporter.RefAttrSettings{
			"flow_id": {RefType: "genesyscloud_flow"},
		},
	}
}

func ResourceWidgetDeployment() *schema.Resource {
	return &schema.Resource{
		Description: "Genesys Cloud Widget Deployment",

		CreateContext: provider.CreateWithPooledClient(createWidgetDeployment),
		ReadContext:   provider.ReadWithPooledClient(readWidgetDeployment),
		UpdateContext: provider.UpdateWithPooledClient(updateWidgetDeployment),
		DeleteContext: provider.DeleteWithPooledClient(deleteWidgetDeployment),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "Name of the Widget Deployment.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"description": {
				Description: "Widget Deployment description.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"authentication_required": {
				Description: "When true, the customer members starting a chat must be authenticated by supplying their JWT to the create operation.",
				Type:        schema.TypeBool,
				Required:    true,
			},
			"disabled": {
				Description: "When true, all create chat operations using this Deployment will be rejected.",
				Type:        schema.TypeBool,
				Required:    true,
			},
			"flow_id": {
				Description: "The Inbound Chat Flow to run when new chats are initiated under this Deployment",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"allowed_domains": {
				Description: "The list of domains that are approved to use this Deployment; the list will be added to CORS headers for ease of web use",
				Type:        schema.TypeList,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Optional:    true,
			},
			"client_type": {
				Description:  "The type of display widget for which this Deployment is configured, which controls the administrator settings shown. Valid values: " + strings.Join(validClientTypes, ", "),
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice(validClientTypes, false),
			},
			"client_config": {
				Description: "The V1 and V1-http client configuration options that should be made available to the clients of this Deployment.",
				Type:        schema.TypeSet,
				MaxItems:    1,
				Optional:    true,
				// when this field is removed, V1 and V1HTTP should also be removed from validClientTypes list
				Deprecated: "This field is inactive and will be removed entirely in a later version. Please use `v2_client_config` or `third_party_client_config` instead.",
				Elem:       clientConfigSchemaResource,
			},
			"v2_client_config": {
				Description: "The v2 client configuration options that should be made available to the clients of this Deployment.",
				Type:        schema.TypeMap,
				Optional:    true,
			},
			"third_party_client_config": {
				Description:   "The third party client configuration options that should be made available to the clients of this Deployment.",
				Type:          schema.TypeMap,
				Optional:      true,
				ConflictsWith: []string{"v2_client_config"},
			},
		},
	}
}

func readWidgetDeployment(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	widgetsAPI := platformclientv2.NewWidgetsApiWithConfig(sdkConfig)
	cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceWidgetDeployment(), constants.ConsistencyChecks(), "genesyscloud_widget_deployment")

	log.Printf("Reading widget deployment %s", d.Id())
	return util.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		currentWidget, resp, getErr := widgetsAPI.GetWidgetsDeployment(d.Id())

		if getErr != nil {
			if util.IsStatus404(resp) {
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError("genesyscloud_widget_deployment", fmt.Sprintf("Failed to read widget deployment %s | error: %s", d.Id(), getErr), resp))
			}

			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError("genesyscloud_widget_deployment", fmt.Sprintf("Failed to read widget deployment %s | error: %s", d.Id(), getErr), resp))
		}

		_ = d.Set("name", *currentWidget.Name)
		resourcedata.SetNillableValue(d, "description", currentWidget.Description)
		resourcedata.SetNillableValue(d, "authentication_required", currentWidget.AuthenticationRequired)
		resourcedata.SetNillableValue(d, "disabled", currentWidget.Disabled)
		resourcedata.SetNillableValue(d, "allowed_domains", currentWidget.AllowedDomains)
		resourcedata.SetNillableValue(d, "client_type", currentWidget.ClientType)
		resourcedata.SetNillableReference(d, "flow_id", currentWidget.Flow)
		if currentWidget.ClientConfig != nil {
			flattenClientConfig(d, *currentWidget.ClientConfig)
		}

		return cc.CheckState(d)
	})
}

func createWidgetDeployment(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)

	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	widgetsAPI := platformclientv2.NewWidgetsApiWithConfig(sdkConfig)

	createWidget := platformclientv2.Widgetdeployment{
		Name:                   &name,
		Description:            platformclientv2.String(d.Get("description").(string)),
		AuthenticationRequired: platformclientv2.Bool(d.Get("authentication_required").(bool)),
		Disabled:               platformclientv2.Bool(d.Get("disabled").(bool)),
		Flow:                   util.BuildSdkDomainEntityRef(d, "flow_id"),
		AllowedDomains:         buildSdkAllowedDomains(d),
		ClientType:             platformclientv2.String(d.Get("client_type").(string)),
		ClientConfig:           buildSDKClientConfig(d),
	}

	log.Printf("Creating widgets deployment %s", name)

	// Get all existing deployments
	resourceIDMetaMap, _ := getAllWidgetDeployments(ctx, sdkConfig)

	widget, resp, err := widgetsAPI.PostWidgetsDeployments(createWidget)
	if err != nil {
		return util.BuildAPIDiagnosticError("genesyscloud_widget_deployment", fmt.Sprintf("Failed to create widget deployment %s error: %s", name, err), resp)
	}
	log.Printf("Widget created %s with id %s", name, *widget.Id)
	d.SetId(*widget.Id)

	time.Sleep(2 * time.Second)
	// Get all new deployments
	newResourceIDMetaMap, _ := getAllWidgetDeployments(ctx, sdkConfig)
	// Delete potential duplicates
	deletePotentialDuplicateDeployments(widgetsAPI, name, *widget.Id, resourceIDMetaMap, newResourceIDMetaMap)

	return readWidgetDeployment(ctx, d, meta)
}

func deleteWidgetDeployment(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)

	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	widgetAPI := platformclientv2.NewWidgetsApiWithConfig(sdkConfig)

	log.Printf("Deleting widget deployment %s", name)
	resp, err := widgetAPI.DeleteWidgetsDeployment(d.Id())
	if err != nil {
		return util.BuildAPIDiagnosticError("genesyscloud_widget_deployment", fmt.Sprintf("Failed to delete widget deployment %s error: %s", name, err), resp)
	}

	return util.WithRetries(ctx, 30*time.Second, func() *retry.RetryError {
		_, resp, err := widgetAPI.GetWidgetsDeployment(d.Id())
		if err != nil {
			if util.IsStatus404(resp) {
				log.Printf("Widget deployment %s deleted", name)
				return nil
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError("genesyscloud_widget_deployment", fmt.Sprintf("Error deleting widget deployment %s | error: %s", d.Id(), err), resp))
		}
		return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError("genesyscloud_widget_deployment", fmt.Sprintf("Widget deployment %s still exists", d.Id()), resp))
	})
}

func updateWidgetDeployment(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)

	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	widgetsAPI := platformclientv2.NewWidgetsApiWithConfig(sdkConfig)

	updateWidget := platformclientv2.Widgetdeployment{
		Name:                   &name,
		Description:            platformclientv2.String(d.Get("description").(string)),
		AuthenticationRequired: platformclientv2.Bool(d.Get("authentication_required").(bool)),
		Disabled:               platformclientv2.Bool(d.Get("disabled").(bool)),
		Flow:                   util.BuildSdkDomainEntityRef(d, "flow_id"),
		AllowedDomains:         buildSdkAllowedDomains(d),
		ClientType:             platformclientv2.String(d.Get("client_type").(string)),
		ClientConfig:           buildSDKClientConfig(d),
	}

	log.Printf("Updating widget deployment %s", name)
	widget, resp, err := widgetsAPI.PutWidgetsDeployment(d.Id(), updateWidget)
	if err != nil {
		return util.BuildAPIDiagnosticError("genesyscloud_widget_deployment", fmt.Sprintf("Failed to update widget deployment %s error: %s", name, err), resp)
	}
	d.SetId(*widget.Id)

	log.Printf("Finished updating widget deployment %s", name)
	return readWidgetDeployment(ctx, d, meta)
}

func buildSDKClientConfig(d *schema.ResourceData) *platformclientv2.Widgetclientconfig {
	v2ClientConfig := d.Get("v2_client_config")
	if m, ok := v2ClientConfig.(map[string]any); ok && len(m) > 0 {
		return &platformclientv2.Widgetclientconfig{V2: &v2ClientConfig}
	}
	thirdPartyClientConfig := d.Get("third_party_client_config")
	if m, ok := thirdPartyClientConfig.(map[string]any); ok && len(m) > 0 {
		return &platformclientv2.Widgetclientconfig{ThirdParty: &thirdPartyClientConfig}
	}
	return nil
}

func buildSdkAllowedDomains(d *schema.ResourceData) *[]string {
	if domains, ok := d.Get("allowed_domains").([]any); ok {
		allowedDomains := lists.InterfaceListToStrings(domains)
		return &allowedDomains
	}
	return nil
}

func flattenClientConfig(d *schema.ResourceData, config platformclientv2.Widgetclientconfig) {
	if config.ThirdParty != nil {
		thirdParty := *config.ThirdParty
		if thirdPartyMap, _ := thirdParty.(map[string]any); len(thirdPartyMap) > 0 {
			_ = d.Set("third_party_client_config", thirdPartyMap)
			return
		}
	}
	if config.V2 != nil {
		v2Config := *config.V2
		if v2ConfigMap, _ := v2Config.(map[string]any); len(v2ConfigMap) > 0 {
			_ = d.Set("v2_client_config", v2ConfigMap)
		}
	}
}

// deletePotentialDuplicateDeployments Sometimes the Widget API creates 2 deployments due to a bug. This function will delete any duplicates
func deletePotentialDuplicateDeployments(widgetAPI *platformclientv2.WidgetsApi, name, id string, existingResourceIDMetaMap, newResourceIDMetaMap resourceExporter.ResourceIDMetaMap) {
	for _, val := range existingResourceIDMetaMap {
		for key1, val1 := range newResourceIDMetaMap {
			if val.BlockLabel == val1.BlockLabel {
				delete(newResourceIDMetaMap, key1)
				break
			}
		}
	}

	for key, val := range newResourceIDMetaMap {
		if key != id && val.BlockLabel == name {
			log.Printf("Deleting duplicate widget deployment %s", name)
			_, err := widgetAPI.DeleteWidgetsDeployment(key)
			if err != nil {
				log.Printf("failed to delete widget deployment %s: %s", name, err)
			}
		}
	}
}
