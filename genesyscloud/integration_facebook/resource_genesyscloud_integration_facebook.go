package integration_facebook

import (
	"context"
	"errors"
	"fmt"
	"log"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v146/platformclientv2"

	"terraform-provider-genesyscloud/genesyscloud/consistency_checker"

	"terraform-provider-genesyscloud/genesyscloud/util/constants"
	"terraform-provider-genesyscloud/genesyscloud/util/resourcedata"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
)

/*
The resource_genesyscloud_integration_facebook.go contains all of the methods that perform the core logic for a resource.
*/

// getAllAuthIntegrationFacebook retrieves all of the integration facebook via Terraform in the Genesys Cloud and is used for the exporter
func getAllAuthIntegrationFacebooks(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	proxy := getIntegrationFacebookProxy(clientConfig)
	resources := make(resourceExporter.ResourceIDMetaMap)

	facebookIntegrationRequests, resp, err := proxy.getAllIntegrationFacebook(ctx)
	if err != nil {
		return nil, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to get integration facebook: %v", err), resp)
	}

	for _, facebookIntegrationRequest := range *facebookIntegrationRequests {
		resources[*facebookIntegrationRequest.Id] = &resourceExporter.ResourceMeta{BlockLabel: *facebookIntegrationRequest.Name}
	}

	return resources, nil
}

// createIntegrationFacebook is used by the integration_facebook resource to create Genesys cloud integration facebook
func createIntegrationFacebook(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getIntegrationFacebookProxy(sdkConfig)

	integrationFacebook := getIntegrationFacebookFromResourceData(d)

	// if PageAccessToken is provided, no need to provide PageId and UserAccessToken
	if *integrationFacebook.PageAccessToken != "" && (*integrationFacebook.PageId != "" || *integrationFacebook.UserAccessToken != "") {
		return util.BuildDiagnosticError(ResourceType, "Configuration Error", errors.New("the pageId and userAccessToken should not be set if specifying the pageAccessToken"))
	}

	if (*integrationFacebook.AppId == "" && *integrationFacebook.AppSecret != "") || (*integrationFacebook.AppId != "" && *integrationFacebook.AppSecret == "") {
		return util.BuildDiagnosticError(ResourceType, "Configuration Error", errors.New("the appSecret is required when appId is provided"))
	}

	log.Printf("Creating integration facebook %s", *integrationFacebook.Name)
	facebookIntegrationRequest, resp, err := proxy.createIntegrationFacebook(ctx, &integrationFacebook)
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to create integration facebook: %s", err), resp)
	}

	d.SetId(*facebookIntegrationRequest.Id)
	log.Printf("Created integration facebook %s", *facebookIntegrationRequest.Id)
	return readIntegrationFacebook(ctx, d, meta)
}

// readIntegrationFacebook is used by the integration_facebook resource to read an integration facebook from genesys cloud
func readIntegrationFacebook(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getIntegrationFacebookProxy(sdkConfig)

	log.Printf("Reading integration facebook ")

	cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceIntegrationFacebook(), constants.ConsistencyChecks(), ResourceType)

	return util.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		facebookIntegrationRequest, resp, getErr := proxy.getIntegrationFacebookById(ctx, d.Id())
		if getErr != nil {
			if util.IsStatus404(resp) {
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read integration facebook %s: %s", d.Id(), getErr), resp))
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read integration facebook %s: %s", d.Id(), getErr), resp))
		}

		resourcedata.SetNillableValue(d, "name", facebookIntegrationRequest.Name)

		if facebookIntegrationRequest.SupportedContent != nil && facebookIntegrationRequest.SupportedContent.Id != nil {
			_ = d.Set("supported_content_id", *facebookIntegrationRequest.SupportedContent.Id)
		}

		if facebookIntegrationRequest.MessagingSetting != nil && facebookIntegrationRequest.MessagingSetting.Id != nil {
			_ = d.Set("messaging_setting_id", *facebookIntegrationRequest.MessagingSetting.Id)
		}

		resourcedata.SetNillableValue(d, "page_id", facebookIntegrationRequest.PageId)
		resourcedata.SetNillableValue(d, "app_id", facebookIntegrationRequest.AppId)

		log.Printf("Read integration facebook %s", *facebookIntegrationRequest.Name)
		return cc.CheckState(d)
	})
}

// updateIntegrationFacebook is used by the integration_facebook resource to update an integration facebook in Genesys Cloud
func updateIntegrationFacebook(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getIntegrationFacebookProxy(sdkConfig)

	supportedContentId := d.Get("supported_content_id").(string)
	messagingContentId := d.Get("messaging_setting_id").(string)
	pageAccessToken := d.Get("page_access_token").(string)
	userAccessToken := d.Get("user_access_token").(string)

	integrationFacebook := platformclientv2.Facebookintegrationupdaterequest{
		Name:             platformclientv2.String(d.Get("name").(string)),
		SupportedContent: &platformclientv2.Supportedcontentreference{Id: &supportedContentId},
		MessagingSetting: &platformclientv2.Messagingsettingrequestreference{Id: &messagingContentId},
		PageAccessToken:  &pageAccessToken,
		UserAccessToken:  &userAccessToken,
	}

	log.Printf("Updating integration facebook %s", *integrationFacebook.Name)
	facebookIntegrationRequest, resp, err := proxy.updateIntegrationFacebook(ctx, d.Id(), &integrationFacebook)
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to update integration facebook: %s", err), resp)
	}

	log.Printf("Updated integration facebook %s", *facebookIntegrationRequest.Id)
	return readIntegrationFacebook(ctx, d, meta)
}

// deleteIntegrationFacebook is used by the integration_facebook resource to delete an integration facebook from Genesys cloud
func deleteIntegrationFacebook(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getIntegrationFacebookProxy(sdkConfig)

	_, err := proxy.deleteIntegrationFacebook(ctx, d.Id())
	if err != nil {
		return diag.Errorf("Failed to delete integration facebook %s: %s", d.Id(), err)
	}

	return util.WithRetries(ctx, 180*time.Second, func() *retry.RetryError {
		_, resp, err := proxy.getIntegrationFacebookById(ctx, d.Id())

		if err != nil {
			if util.IsStatus404(resp) {
				log.Printf("Deleted integration facebook %s", d.Id())
				return nil
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Error deleting integration facebook %s: %s", d.Id(), err), resp))
		}

		return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("integration facebook %s still exists", d.Id()), resp))
	})
}
