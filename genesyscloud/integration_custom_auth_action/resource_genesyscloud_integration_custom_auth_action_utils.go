package integration_custom_auth_action

import (
	"context"
	"fmt"
	integrationAction "terraform-provider-genesyscloud/genesyscloud/integration_action"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"
)

/*
The resource_genesyscloud_integration_action_utils.go file contains various helper methods to marshal
and unmarshal data into formats consumable by Terraform and/or Genesys Cloud.

Note:  Look for opportunities to minimize boilerplate code using functions and Generics
*/

const (
	customAuthIdPrefix        = "customAuth" // Custom Auth Data Action IDs start with this
	customAuthCredentialType  = "userDefinedOAuth"
	customRestIntegrationType = "custom-rest-actions"

	reqTemplateFileName     = "requesttemplate.vm"
	successTemplateFileName = "successtemplate.vm"
)

// BuildSdkCustomAuthActionConfig takes the resource data and builds the SDK platformclientv2.Actionconfig from it
// This is a stripped version of the integrationAction.BuildSdkActionConfig because 'timeoutSeconds'
// is invalid for Custom Auth Actions
func BuildSdkCustomAuthActionConfig(d *schema.ResourceData) *platformclientv2.Actionconfig {
	ActionConfig := &platformclientv2.Actionconfig{
		Request:  integrationAction.BuildSdkActionConfigRequest(d),
		Response: integrationAction.BuildSdkActionConfigResponse(d),
	}
	return ActionConfig
}

// isIntegrationAndCredTypesCorrect checks if the integration is of type Web Services Data Action ("custom-rest-actions")
// and checks that the credential is configured as "userDefinedOAuth" which are the requirements
// for a Custom Auth Data Action(Genesys Cloud managed) to exist for the integration.
func isIntegrationAndCredTypesCorrect(ctx context.Context, cap *customAuthActionsProxy, integrationId string) (bool, error) {
	// Check that the integration is the correct type
	integType, resp, err := cap.getIntegrationType(ctx, integrationId)
	if err != nil {
		return false, fmt.Errorf("cannot identify integration type of integration %s: %v %v", integrationId, err, resp)
	}
	if integType != customRestIntegrationType {
		return false, fmt.Errorf("integration should be of type %v to use custom auth action. Actual: %v", customRestIntegrationType, integType)
	}

	// Check credentials
	credType, resp, err := cap.getIntegrationCredentialsType(ctx, integrationId)
	if err != nil {
		return false, err
	}
	if credType != customAuthCredentialType {
		return false, fmt.Errorf("credentials type of integration %s should be %s %v", integrationId, customAuthCredentialType, resp)
	}
	return true, nil
}

// getCustomAuthIdFromIntegration gets the expected custom auth action ID from the integration ID.
func getCustomAuthIdFromIntegration(integrationId string) string {
	return fmt.Sprintf("%s_-_%s", customAuthIdPrefix, integrationId)
}
