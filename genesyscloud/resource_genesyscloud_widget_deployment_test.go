package genesyscloud

import (
	"fmt"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v115/platformclientv2"
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
	webChatSkin            string
	authenticationUrl      string
}

func deleteWidgetDeploymentWithName(name string) {
	widgetsAPI := platformclientv2.NewWidgetsApiWithConfig(sdkConfig)
	widgetDeployments, _, getErr := widgetsAPI.GetWidgetsDeployments()
	if getErr != nil {
		return
	}

	for _, widgetDeployment := range *widgetDeployments.Entities {
		if strings.Contains(*widgetDeployment.Name, name) {
			widgetsAPI.DeleteWidgetsDeployment(*widgetDeployment.Id)
		}
	}
}

func generateWidgetDeployV2(widgetDeploymentConfig *widgetDeploymentConfig) string {
	return fmt.Sprintf(`resource "genesyscloud_widget_deployment" "%s" {
		name = "%s"
		description = "%s"
		flow_id = "%s"
		client_type = "%s"
		authentication_required = %s
		disabled = %s
	}
	`, widgetDeploymentConfig.resourceID,
		widgetDeploymentConfig.name,
		widgetDeploymentConfig.description,
		widgetDeploymentConfig.flowID,
		widgetDeploymentConfig.clientType,
		widgetDeploymentConfig.authenticationRequired,
		widgetDeploymentConfig.disabled)
}

func generateWidgetDeployV1(widgetDeploymentConfig *widgetDeploymentConfig) string {
	return fmt.Sprintf(`resource "genesyscloud_widget_deployment" "%s" {
		name = "%s"
		description = "%s"
		flow_id = "%s"
		client_type = "%s"
		authentication_required = %s
		disabled = %s
		client_config {
    		authentication_url = "%s"
    		webchat_skin       = "%s"
  		}
	}
	`, widgetDeploymentConfig.resourceID,
		widgetDeploymentConfig.name,
		widgetDeploymentConfig.description,
		widgetDeploymentConfig.flowID,
		widgetDeploymentConfig.clientType,
		widgetDeploymentConfig.authenticationRequired,
		widgetDeploymentConfig.disabled,
		widgetDeploymentConfig.authenticationUrl,
		widgetDeploymentConfig.webChatSkin)
}

func TestAccResourceWidgetDeploymentV2Widget(t *testing.T) {
	t.Parallel()
	name := "My Test V2 Widget"
	widgetDeployV2 := &widgetDeploymentConfig{
		resourceID:             "myTestV2Widget",
		name:                   name + uuid.NewString(),
		description:            "This is a test description",
		flowID:                 uuid.NewString(),
		clientType:             "v2",
		authenticationRequired: "false",
		disabled:               "true",
	}

	_, err := AuthorizeSdk()
	if err != nil {
		t.Fatal(err)
	}
	deleteWidgetDeploymentWithName(name)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { TestAccPreCheck(t) },
		ProviderFactories: GetProviderFactories(providerResources, providerDataSources),

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
				),
			},
			{
				//update
				Config: generateWidgetDeployV2(&widgetDeploymentConfig{
					resourceID:             widgetDeployV2.resourceID,
					name:                   widgetDeployV2.name,
					description:            "New test description",
					flowID:                 widgetDeployV2.flowID,
					clientType:             widgetDeployV2.clientType,
					authenticationRequired: widgetDeployV2.authenticationRequired,
					disabled:               widgetDeployV2.disabled,
				}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_widget_deployment."+widgetDeployV2.resourceID, "name", widgetDeployV2.name),
					resource.TestCheckResourceAttr("genesyscloud_widget_deployment."+widgetDeployV2.resourceID, "description", "New test description"),
					resource.TestCheckResourceAttr("genesyscloud_widget_deployment."+widgetDeployV2.resourceID, "flow_id", widgetDeployV2.flowID),
					resource.TestCheckResourceAttr("genesyscloud_widget_deployment."+widgetDeployV2.resourceID, "client_type", widgetDeployV2.clientType),
					resource.TestCheckResourceAttr("genesyscloud_widget_deployment."+widgetDeployV2.resourceID, "authentication_required", widgetDeployV2.authenticationRequired),
					resource.TestCheckResourceAttr("genesyscloud_widget_deployment."+widgetDeployV2.resourceID, "disabled", widgetDeployV2.disabled),
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

func TestAccResourceWidgetDeploymentV1Widget(t *testing.T) {
	t.Parallel()
	name := "My Text V1 Widget"
	widgetDeployV1 := &widgetDeploymentConfig{
		resourceID:             "myTestV1Widget",
		name:                   name + uuid.NewString(),
		description:            "This is a test description",
		flowID:                 uuid.NewString(),
		clientType:             "v1",
		authenticationRequired: "true",
		disabled:               "true",
		webChatSkin:            "basic",
		authenticationUrl:      "https://localhost",
	}

	_, err := AuthorizeSdk()
	if err != nil {
		t.Fatal(err)
	}
	deleteWidgetDeploymentWithName(name)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { TestAccPreCheck(t) },
		ProviderFactories: GetProviderFactories(providerResources, providerDataSources),

		Steps: []resource.TestStep{
			{
				//create
				Config: generateWidgetDeployV1(widgetDeployV1),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_widget_deployment."+widgetDeployV1.resourceID, "name", widgetDeployV1.name),
					resource.TestCheckResourceAttr("genesyscloud_widget_deployment."+widgetDeployV1.resourceID, "description", widgetDeployV1.description),
					resource.TestCheckResourceAttr("genesyscloud_widget_deployment."+widgetDeployV1.resourceID, "flow_id", widgetDeployV1.flowID),
					resource.TestCheckResourceAttr("genesyscloud_widget_deployment."+widgetDeployV1.resourceID, "client_type", widgetDeployV1.clientType),
					resource.TestCheckResourceAttr("genesyscloud_widget_deployment."+widgetDeployV1.resourceID, "authentication_required", widgetDeployV1.authenticationRequired),
					resource.TestCheckResourceAttr("genesyscloud_widget_deployment."+widgetDeployV1.resourceID, "disabled", widgetDeployV1.disabled),
					resource.TestCheckResourceAttr("genesyscloud_widget_deployment."+widgetDeployV1.resourceID, "client_config.0.authentication_url", widgetDeployV1.authenticationUrl),
					resource.TestCheckResourceAttr("genesyscloud_widget_deployment."+widgetDeployV1.resourceID, "client_config.0.webchat_skin", widgetDeployV1.webChatSkin),
				),
			},
			{
				//update
				Config: generateWidgetDeployV1(&widgetDeploymentConfig{
					resourceID:             widgetDeployV1.resourceID,
					name:                   widgetDeployV1.name,
					description:            "New test description",
					flowID:                 widgetDeployV1.flowID,
					clientType:             widgetDeployV1.clientType,
					authenticationRequired: widgetDeployV1.authenticationRequired,
					disabled:               widgetDeployV1.disabled,
					webChatSkin:            widgetDeployV1.webChatSkin,
					authenticationUrl:      widgetDeployV1.authenticationUrl,
				}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_widget_deployment."+widgetDeployV1.resourceID, "name", widgetDeployV1.name),
					resource.TestCheckResourceAttr("genesyscloud_widget_deployment."+widgetDeployV1.resourceID, "description", "New test description"),
					resource.TestCheckResourceAttr("genesyscloud_widget_deployment."+widgetDeployV1.resourceID, "flow_id", widgetDeployV1.flowID),
					resource.TestCheckResourceAttr("genesyscloud_widget_deployment."+widgetDeployV1.resourceID, "client_type", widgetDeployV1.clientType),
					resource.TestCheckResourceAttr("genesyscloud_widget_deployment."+widgetDeployV1.resourceID, "authentication_required", widgetDeployV1.authenticationRequired),
					resource.TestCheckResourceAttr("genesyscloud_widget_deployment."+widgetDeployV1.resourceID, "disabled", widgetDeployV1.disabled),
				),
			},
			{
				// Import/Read
				ResourceName:      "genesyscloud_widget_deployment." + widgetDeployV1.resourceID,
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

		if IsStatus404(resp) {
			// Widget deployment does not exits keep going
			continue
		}

		// Unexpected error
		return fmt.Errorf("Unexpected error: %s", err)
	}
	// Success. Widget Deployment
	return nil
}
