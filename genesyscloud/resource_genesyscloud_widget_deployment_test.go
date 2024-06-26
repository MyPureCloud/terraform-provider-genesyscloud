package genesyscloud

import (
	"fmt"
	"strconv"
	"strings"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"
)

type widgetDeploymentConfig struct {
	resourceID             string
	name                   string
	description            string
	authenticationRequired string
	disabled               string
	flowID                 string
	allowedDomains         []string
	clientType             string
	v2ClientConfig         map[string]string
	thirdPartyClientConfig map[string]string
}

func deleteWidgetDeploymentWithName(name string) {
	widgetsAPI := platformclientv2.NewWidgetsApiWithConfig(sdkConfig)
	widgetDeployments, _, getErr := widgetsAPI.GetWidgetsDeployments()
	if getErr != nil {
		return
	}

	for _, widgetDeployment := range *widgetDeployments.Entities {
		if strings.Contains(*widgetDeployment.Name, name) {
			_, _ = widgetsAPI.DeleteWidgetsDeployment(*widgetDeployment.Id)
		}
	}
}

func generateWidgetDeploymentResource(wdConfig *widgetDeploymentConfig) string {
	var (
		v2ClientConfigStr   string
		thirdPartyConfigStr string
	)
	if wdConfig.v2ClientConfig != nil {
		v2ClientConfigStr = util.GenerateMapAttrWithMapProperties("v2_client_config", wdConfig.v2ClientConfig)
	}
	if wdConfig.thirdPartyClientConfig != nil {
		thirdPartyConfigStr = util.GenerateMapAttrWithMapProperties("third_party_client_config", wdConfig.thirdPartyClientConfig)
	}
	return fmt.Sprintf(`resource "genesyscloud_widget_deployment" "%s" {
		name                    = "%s"
		description             =  %s
		flow_id                 =  %s
		client_type             = "%s"
		authentication_required = %s
		disabled                = %s
		%s
        %s
	}
	`, wdConfig.resourceID,
		wdConfig.name,
		wdConfig.description,
		wdConfig.flowID,
		wdConfig.clientType,
		wdConfig.authenticationRequired,
		wdConfig.disabled,
		v2ClientConfigStr,
		thirdPartyConfigStr,
	)
}

func copyWidgetDeploymentConfig(original widgetDeploymentConfig) *widgetDeploymentConfig {
	var widgetCopy widgetDeploymentConfig
	widgetCopy = widgetDeploymentConfig{
		resourceID:             original.resourceID,
		name:                   original.name,
		description:            original.description,
		authenticationRequired: original.authenticationRequired,
		disabled:               original.disabled,
		flowID:                 original.flowID,
		allowedDomains:         append(widgetCopy.allowedDomains, original.allowedDomains...),
		clientType:             original.clientType,
		v2ClientConfig:         original.v2ClientConfig,
		thirdPartyClientConfig: original.thirdPartyClientConfig,
	}
	return &widgetCopy
}

func TestAccResourceWidgetDeploymentThirdPartyWidget(t *testing.T) {
	t.Parallel()
	name := "My Test Third Party Widget" + uuid.NewString()
	description := "This is a test description"
	flowId := uuid.NewString()
	widgetDeployment := &widgetDeploymentConfig{
		resourceID:             "third_party_widget",
		name:                   name,
		description:            strconv.Quote(description),
		flowID:                 strconv.Quote(flowId),
		clientType:             V2,
		authenticationRequired: util.FalseValue,
		disabled:               util.TrueValue,
		thirdPartyClientConfig: map[string]string{
			"foo": strconv.Quote("bar"),
		},
	}

	updatedDescription := "New description"
	widgetDeploymentUpdate := copyWidgetDeploymentConfig(*widgetDeployment)
	widgetDeploymentUpdate.description = strconv.Quote(updatedDescription)
	widgetDeploymentUpdate.thirdPartyClientConfig = map[string]string{
		"foo": strconv.Quote("bar2"),
	}

	deleteWidgetDeploymentWithName(name)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),

		Steps: []resource.TestStep{
			{
				//create
				Config: generateWidgetDeploymentResource(widgetDeployment),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_widget_deployment."+widgetDeployment.resourceID, "name", name),
					resource.TestCheckResourceAttr("genesyscloud_widget_deployment."+widgetDeployment.resourceID, "description", description),
					resource.TestCheckResourceAttr("genesyscloud_widget_deployment."+widgetDeployment.resourceID, "flow_id", flowId),
					resource.TestCheckResourceAttr("genesyscloud_widget_deployment."+widgetDeployment.resourceID, "client_type", widgetDeployment.clientType),
					resource.TestCheckResourceAttr("genesyscloud_widget_deployment."+widgetDeployment.resourceID, "authentication_required", widgetDeployment.authenticationRequired),
					resource.TestCheckResourceAttr("genesyscloud_widget_deployment."+widgetDeployment.resourceID, "disabled", widgetDeployment.disabled),
					resource.TestCheckResourceAttr("genesyscloud_widget_deployment."+widgetDeployment.resourceID, "third_party_client_config.foo", "bar"),
				),
			},
			{
				//update
				Config: generateWidgetDeploymentResource(widgetDeploymentUpdate),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_widget_deployment."+widgetDeployment.resourceID, "name", name),
					resource.TestCheckResourceAttr("genesyscloud_widget_deployment."+widgetDeployment.resourceID, "description", updatedDescription),
					resource.TestCheckResourceAttr("genesyscloud_widget_deployment."+widgetDeployment.resourceID, "flow_id", flowId),
					resource.TestCheckResourceAttr("genesyscloud_widget_deployment."+widgetDeployment.resourceID, "client_type", widgetDeployment.clientType),
					resource.TestCheckResourceAttr("genesyscloud_widget_deployment."+widgetDeployment.resourceID, "authentication_required", widgetDeployment.authenticationRequired),
					resource.TestCheckResourceAttr("genesyscloud_widget_deployment."+widgetDeployment.resourceID, "disabled", widgetDeployment.disabled),
					resource.TestCheckResourceAttr("genesyscloud_widget_deployment."+widgetDeployment.resourceID, "third_party_client_config.foo", "bar2"),
				),
			},
			{
				// Import/Read
				ResourceName:      "genesyscloud_widget_deployment." + widgetDeployment.resourceID,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
		CheckDestroy: testVerifyWidgetDeploymentDestroyed,
	})
}

func testVerifyWidgetDeploymentDestroyed(state *terraform.State) error {
	widgetAPI := platformclientv2.NewWidgetsApi()

	for _, rs := range state.RootModule().Resources {
		if rs.Type != "genesyscloud_widget_deployment" {
			continue
		}

		widgetDeployment, resp, err := widgetAPI.GetWidgetsDeployment(rs.Primary.ID)

		if widgetDeployment != nil {
			return fmt.Errorf("Widget deployment (%s) still exists", rs.Primary.ID)
		}

		if util.IsStatus404(resp) {
			// Widget deployment does not exits keep going
			continue
		}

		// Unexpected error
		return fmt.Errorf("Unexpected error: %s", err)
	}
	// Success. Widget Deployment
	return nil
}
