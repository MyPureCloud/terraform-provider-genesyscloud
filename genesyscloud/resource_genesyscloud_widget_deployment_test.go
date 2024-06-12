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
	"github.com/mypurecloud/platform-client-sdk-go/v130/platformclientv2"
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

func generateWidgetDeployV2(wdConfig *widgetDeploymentConfig) string {
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
		name = "%s"
		description = "%s"
		flow_id = "%s"
		client_type = "%s"
		authentication_required = %s
		disabled = %s
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

func TestAccResourceWidgetDeploymentV2Widget(t *testing.T) {
	t.Parallel()
	name := "My Test V2 Widget"
	widgetDeployV2 := &widgetDeploymentConfig{
		resourceID:             "myTestV2Widget",
		name:                   name + uuid.NewString(),
		description:            "This is a test description",
		flowID:                 uuid.NewString(),
		clientType:             V2,
		authenticationRequired: util.FalseValue,
		disabled:               util.TrueValue,
		thirdPartyClientConfig: map[string]string{
			"foo": strconv.Quote("bar"),
		},
	}

	updatedDescription := "New description"
	widgetDeployV2Update := widgetDeployV2
	widgetDeployV2Update.description = updatedDescription

	deleteWidgetDeploymentWithName(name)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),

		Steps: []resource.TestStep{
			{
				//create
				Config: generateWidgetDeployV2(widgetDeployV2),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_widget_deployment."+widgetDeployV2.resourceID, "name", widgetDeployV2.name),
					resource.TestCheckResourceAttr("genesyscloud_widget_deployment."+widgetDeployV2.resourceID, "description", widgetDeployV2.description),
					resource.TestCheckResourceAttr("genesyscloud_widget_deployment."+widgetDeployV2.resourceID, "flow_id", widgetDeployV2.flowID),
					resource.TestCheckResourceAttr("genesyscloud_widget_deployment."+widgetDeployV2.resourceID, "client_type", widgetDeployV2.clientType),
					resource.TestCheckResourceAttr("genesyscloud_widget_deployment."+widgetDeployV2.resourceID, "authentication_required", widgetDeployV2.authenticationRequired),
					resource.TestCheckResourceAttr("genesyscloud_widget_deployment."+widgetDeployV2.resourceID, "disabled", widgetDeployV2.disabled),
					resource.TestCheckResourceAttr("genesyscloud_widget_deployment."+widgetDeployV2.resourceID, "third_party_client_config.foo", "bar"),
				),
			},
			{
				//update
				Config: generateWidgetDeployV2(widgetDeployV2Update),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_widget_deployment."+widgetDeployV2.resourceID, "name", widgetDeployV2.name),
					resource.TestCheckResourceAttr("genesyscloud_widget_deployment."+widgetDeployV2.resourceID, "description", updatedDescription),
					resource.TestCheckResourceAttr("genesyscloud_widget_deployment."+widgetDeployV2.resourceID, "flow_id", widgetDeployV2.flowID),
					resource.TestCheckResourceAttr("genesyscloud_widget_deployment."+widgetDeployV2.resourceID, "client_type", widgetDeployV2.clientType),
					resource.TestCheckResourceAttr("genesyscloud_widget_deployment."+widgetDeployV2.resourceID, "authentication_required", widgetDeployV2.authenticationRequired),
					resource.TestCheckResourceAttr("genesyscloud_widget_deployment."+widgetDeployV2.resourceID, "disabled", widgetDeployV2.disabled),
					resource.TestCheckResourceAttr("genesyscloud_widget_deployment."+widgetDeployV2.resourceID, "third_party_client_config.foo", "bar"),
				),
			},
			{
				// Import/Read
				ResourceName:      "genesyscloud_widget_deployment." + widgetDeployV2.resourceID,
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
