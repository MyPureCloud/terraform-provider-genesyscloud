package genesyscloud

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"terraform-provider-genesyscloud/genesyscloud/consistency_checker"

	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	lists "terraform-provider-genesyscloud/genesyscloud/util/lists"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/mypurecloud/platform-client-sdk-go/v115/platformclientv2"
)

const (
	V1            = "v1"
	V1HTTP        = "v1-http"
	V2            = "v2"
	THIRDPARTY    = "third-party"
	HTTPSPROTOCOL = "https"
	WEBSKINBASIC  = "basic"
	WEBSKINMODERN = "modern-caret-skin"
)

var (
	clientConfigSchemaResource = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"webchat_skin": {
				Description:  "Skin for the webchat user. (basic, modern-caret-skin)",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{WEBSKINBASIC, WEBSKINMODERN}, false),
			},
			"authentication_url": {
				Description:      "Url endpoint to perform_authentication",
				Type:             schema.TypeString,
				Required:         false,
				Optional:         true,
				ValidateDiagFunc: validateAuthURL,
			},
		},
	}
)

func getAllWidgetDeployments(_ context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	resources := make(resourceExporter.ResourceIDMetaMap)
	widgetsAPI := platformclientv2.NewWidgetsApiWithConfig(clientConfig)
	widgetDeployments, _, getErr := widgetsAPI.GetWidgetsDeployments()
	if getErr != nil {
		return nil, diag.Errorf("Failed to get page of widget deployments: %v", getErr)
	}

	for _, widgetDeployment := range *widgetDeployments.Entities {
		resources[*widgetDeployment.Id] = &resourceExporter.ResourceMeta{Name: *widgetDeployment.Name}
	}

	return resources, nil
}

func buildSdkAllowedDomains(d *schema.ResourceData) *[]string {
	allowed_domains := []string{}
	if domains, ok := d.GetOk("allowed_domains"); ok {
		allowed_domains = lists.InterfaceListToStrings(domains.([]interface{}))
	}
	return &allowed_domains
}

func parseSdkClientConfigData(d *schema.ResourceData) (webchatSkin *string, authenticationUrl *string) {
	clientConfigSet := d.Get("client_config").(*schema.Set)

	if clientConfigSet != nil && len(clientConfigSet.List()) > 0 {
		clientConfig := clientConfigSet.List()[0].(map[string]interface{})
		fields := make(map[string]string)

		for k, v := range clientConfig {
			fields[k] = v.(string)
		}

		webchatSkin := fields["webchat_skin"]
		authUrl := fields["authentication_url"]
		return &webchatSkin, &authUrl
	}
	return nil, nil
}

func validateAuthURL(authUrl interface{}, _ cty.Path) diag.Diagnostics {
	authUrlString := authUrl.(string)
	u, err := url.Parse(authUrlString)
	if err != nil {
		return diag.Errorf("Authorization url %s provided is not a valid URL", authUrlString)
	}

	if u.Scheme == "" || u.Host == "" {
		log.Printf("Scheme: %s", u.Scheme)
		log.Printf("Host: %s", u.Host)
		return diag.Errorf("Authorization url %s provided is not valid url", authUrlString)
	}

	if u.Scheme != HTTPSPROTOCOL {
		return diag.Errorf("Authorization url %s provided must begin with https", authUrlString)
	}

	return nil
}

func buildSDKClientConfig(clientType string, d *schema.ResourceData) (*platformclientv2.Widgetclientconfig, error) {
	widgetClientConfig := &platformclientv2.Widgetclientconfig{}
	clientConfig := d.Get("client_config").(*schema.Set)

	if (clientType == V1 || clientType == V1HTTP) && clientConfig.Len() == 0 {
		return nil, fmt.Errorf("V1 and v1-http widget configurations must have a client_config defined. ")
	}

	if clientType == V1 {
		v1Client := &platformclientv2.Widgetclientconfigv1{}

		v1Client.WebChatSkin, v1Client.AuthenticationUrl = parseSdkClientConfigData(d)

		widgetClientConfig.V1 = v1Client
	}

	if clientType == V1HTTP {
		v1HttpClient := &platformclientv2.Widgetclientconfigv1http{}
		v1HttpClient.WebChatSkin, v1HttpClient.AuthenticationUrl = parseSdkClientConfigData(d)
		widgetClientConfig.V1Http = v1HttpClient
	}

	return widgetClientConfig, nil
}

func WidgetDeploymentExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: GetAllWithPooledClient(getAllWidgetDeployments),
		RefAttrs: map[string]*resourceExporter.RefAttrSettings{
			"flow_id": {RefType: "genesyscloud_flow"},
		},
	}
}

func ResourceWidgetDeployment() *schema.Resource {
	return &schema.Resource{
		Description: "Genesys Cloud Widget Deployment",

		CreateContext: CreateWithPooledClient(createWidgetDeployment),
		ReadContext:   ReadWithPooledClient(readWidgetDeployment),
		UpdateContext: UpdateWithPooledClient(updateWidgetDeployment),
		DeleteContext: DeleteWithPooledClient(deleteWidgetDeployment),
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
				Required:    false,
				Optional:    true,
			},
			"client_type": {
				Description:  "The type of display widget for which this Deployment is configured, which controls the administrator settings shown.Valid values: v1, v2, v1-http, third-party.",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{V1, V2, V1HTTP, THIRDPARTY}, false),
			},
			"client_config": {
				Description: " The V1 and V1-http client configuration options that should be made available to the clients of this Deployment.",
				Type:        schema.TypeSet,
				MaxItems:    1,
				Optional:    true,
				Elem:        clientConfigSchemaResource,
			},
		},
	}
}

func flattenClientConfig(clientType string, clientConfig platformclientv2.Widgetclientconfig) *schema.Set {
	clientConfigSet := schema.NewSet(schema.HashResource(clientConfigSchemaResource), []interface{}{})

	clientConfigMap := make(map[string]interface{})

	if clientType == V1 {
		if clientConfig.V1.WebChatSkin != nil {
			clientConfigMap["webchat_skin"] = *clientConfig.V1.WebChatSkin
		}

		if clientConfig.V1.AuthenticationUrl != nil {
			clientConfigMap["authentication_url"] = *clientConfig.V1.AuthenticationUrl
		}
	}

	if clientType == V1HTTP {
		if clientConfig.V1Http.WebChatSkin != nil {
			clientConfigMap["webchat_skin"] = *clientConfig.V1Http.WebChatSkin
		}

		if clientConfig.V1Http.AuthenticationUrl != nil {
			clientConfigMap["authentication_url"] = *clientConfig.V1Http.AuthenticationUrl
		}
	}

	clientConfigSet.Add(clientConfigMap)

	return clientConfigSet
}

func readWidgetDeployment(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*ProviderMeta).ClientConfig
	widgetsAPI := platformclientv2.NewWidgetsApiWithConfig(sdkConfig)

	log.Printf("Reading widget deployment %s", d.Id())
	return WithRetriesForRead(ctx, d, func() *retry.RetryError {
		currentWidget, resp, getErr := widgetsAPI.GetWidgetsDeployment(d.Id())

		if getErr != nil {
			if IsStatus404(resp) {
				return retry.RetryableError(fmt.Errorf("Failed to read widget deployment %s: %s", d.Id(), getErr))
			}

			return retry.NonRetryableError(fmt.Errorf("Failed to read widget deployment %s: %s", d.Id(), getErr))
		}

		cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceWidgetDeployment())
		d.Set("name", *currentWidget.Name)
		if currentWidget.Description != nil {
			d.Set("description", *currentWidget.Description)
		} else {
			d.Set("description", nil)
		}

		if currentWidget.AuthenticationRequired != nil {
			d.Set("authentication_required", *currentWidget.AuthenticationRequired)
		} else {
			d.Set("authentication_required", nil)
		}

		if currentWidget.Disabled != nil {
			d.Set("disabled", *currentWidget.Disabled)
		} else {
			d.Set("disabled", nil)
		}

		if currentWidget.Flow != nil {
			d.Set("flow_id", *currentWidget.Flow.Id)
		} else {
			d.Set("flow_id", nil)
		}

		if currentWidget.AllowedDomains != nil {
			d.Set("allowed_domains", *currentWidget.AllowedDomains)
		} else {
			d.Set("allowed_domains", nil)
		}

		if currentWidget.ClientType != nil {
			d.Set("client_type", *currentWidget.ClientType)
		} else {
			d.Set("client_type", nil)
		}

		if currentWidget.ClientConfig != nil {
			d.Set("client_config", flattenClientConfig(*currentWidget.ClientType, *currentWidget.ClientConfig))
		} else {
			d.Set("client_config", nil)
		}

		return cc.CheckState()
	})
}

func createWidgetDeployment(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)
	description := d.Get("description").(string)
	auth_required := d.Get("authentication_required").(bool)
	disabled := d.Get("disabled").(bool)
	flowId := BuildSdkDomainEntityRef(d, "flow_id")
	allowed_domains := buildSdkAllowedDomains(d) //Need to make this an array of strings.
	client_type := d.Get("client_type").(string)
	client_config, client_config_err := buildSDKClientConfig(client_type, d)

	if client_config_err != nil {
		return diag.Errorf("Failed to create widget deployment %s, %s", name, client_config_err)
	}

	sdkConfig := meta.(*ProviderMeta).ClientConfig
	widgetsAPI := platformclientv2.NewWidgetsApiWithConfig(sdkConfig)

	createWidget := platformclientv2.Widgetdeployment{
		Name:                   &name,
		Description:            &description,
		AuthenticationRequired: &auth_required,
		Disabled:               &disabled,
		Flow:                   flowId,
		AllowedDomains:         allowed_domains,
		ClientType:             &client_type,
		ClientConfig:           client_config,
	}

	log.Printf("Creating widgets deployment %s", name)

	// Get all existing deployments
	resourceIDMetaMap, _ := getAllWidgetDeployments(ctx, sdkConfig)

	widget, _, err := widgetsAPI.PostWidgetsDeployments(createWidget)
	if err != nil {
		return diag.Errorf("Failed to create widget deployment %s, %s", name, err)
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

// Sometimes the Widget API creates 2 deployments due to a bug. This function will delete any duplicates
func deletePotentialDuplicateDeployments(widgetAPI *platformclientv2.WidgetsApi, name, id string, existingResourceIDMetaMap, newResourceIDMetaMap resourceExporter.ResourceIDMetaMap) {
	for _, val := range existingResourceIDMetaMap {
		for key1, val1 := range newResourceIDMetaMap {
			if val.Name == val1.Name {
				delete(newResourceIDMetaMap, key1)
				break
			}
		}
	}

	for key, val := range newResourceIDMetaMap {
		if key != id && val.Name == name {
			log.Printf("Deleting duplicate widget deployment %s", name)
			_, err := widgetAPI.DeleteWidgetsDeployment(key)
			if err != nil {
				log.Printf("failed to delete widget deployment %s: %s", name, err)
			}
		}
	}
}

func deleteWidgetDeployment(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)

	sdkConfig := meta.(*ProviderMeta).ClientConfig
	widgetAPI := platformclientv2.NewWidgetsApiWithConfig(sdkConfig)

	log.Printf("Deleting widget deployment %s", name)
	_, err := widgetAPI.DeleteWidgetsDeployment(d.Id())
	if err != nil {
		return diag.Errorf("Failed to delete widget deployment %s: %s", name, err)
	}

	return WithRetries(ctx, 30*time.Second, func() *retry.RetryError {
		_, resp, err := widgetAPI.GetWidgetsDeployment(d.Id())
		if err != nil {
			if IsStatus404(resp) {
				log.Printf("Widget deployment %s deleted", name)
				return nil
			}
			return retry.NonRetryableError(fmt.Errorf("Error deleting widget deployment %s: %s", d.Id(), err))
		}
		return retry.RetryableError(fmt.Errorf("Widget deployment %s still exists", d.Id()))
	})
}

func updateWidgetDeployment(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)
	description := d.Get("description").(string)
	auth_required := d.Get("authentication_required").(bool)
	disabled := d.Get("disabled").(bool)
	flowId := BuildSdkDomainEntityRef(d, "flow_id")
	allowed_domains := buildSdkAllowedDomains(d) //Need to make this an array of strings.
	client_type := d.Get("client_type").(string)
	client_config, client_config_err := buildSDKClientConfig(client_type, d)

	if client_config_err != nil {
		return diag.Errorf("Failed updating widget deployment %s, %s", name, client_config_err)
	}

	sdkConfig := meta.(*ProviderMeta).ClientConfig
	widgetsAPI := platformclientv2.NewWidgetsApiWithConfig(sdkConfig)

	updateWidget := platformclientv2.Widgetdeployment{
		Name:                   &name,
		Description:            &description,
		AuthenticationRequired: &auth_required,
		Disabled:               &disabled,
		Flow:                   flowId,
		AllowedDomains:         allowed_domains,
		ClientType:             &client_type,
		ClientConfig:           client_config,
	}

	log.Printf("Updating widget deployment %s", name)
	widget, _, err := widgetsAPI.PutWidgetsDeployment(d.Id(), updateWidget)
	if err != nil {
		return diag.Errorf("Failed to update widget deployment %s, %s", name, err)
	}
	d.SetId(*widget.Id)

	log.Printf("Finished updating widget deployment %s", name)
	return readWidgetDeployment(ctx, d, meta)
}
