package conversations_messaging_integrations_whatsapp

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v165/platformclientv2"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/consistency_checker"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/constants"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/resourcedata"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
)

/*
The resource_genesyscloud_conversations_messaging_integrations_whatsapp.go contains all of the methods that perform the core logic for a resource.
*/

// getAllAuthConversationsMessagingIntegrationsWhatsapp retrieves all of the conversations messaging integrations whatsapp via Terraform in the Genesys Cloud and is used for the exporter
func getAllAuthConversationsMessagingIntegrationsWhatsapps(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	proxy := getConversationsMessagingIntegrationsWhatsappProxy(clientConfig)
	resources := make(resourceExporter.ResourceIDMetaMap)

	whatsAppEmbeddedSignupIntegrationRequests, resp, err := proxy.getAllConversationsMessagingIntegrationsWhatsapp(ctx)
	if err != nil {
		return nil, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to get conversations messaging integrations whatsapp: %v", err), resp)
	}

	for _, whatsAppEmbeddedSignupIntegrationRequest := range *whatsAppEmbeddedSignupIntegrationRequests {
		resources[*whatsAppEmbeddedSignupIntegrationRequest.Id] = &resourceExporter.ResourceMeta{BlockLabel: *whatsAppEmbeddedSignupIntegrationRequest.Name}
	}

	return resources, nil
}

// createConversationsMessagingIntegrationsWhatsapp is used by the conversations_messaging_integrations_whatsapp resource to create Genesys cloud conversations messaging integrations whatsapp
func createConversationsMessagingIntegrationsWhatsapp(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getConversationsMessagingIntegrationsWhatsappProxy(sdkConfig)

	conversationsMessagingIntegrationsWhatsapp := getConversationsMessagingIntegrationsWhatsappFromResourceData(d)

	log.Printf("Creating conversations messaging integrations whatsapp %s", *conversationsMessagingIntegrationsWhatsapp.Name)
	whatsAppEmbeddedSignupIntegrationRequest, resp, err := proxy.createConversationsMessagingIntegrationsWhatsapp(ctx, &conversationsMessagingIntegrationsWhatsapp)
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to create conversations messaging integrations whatsapp %s: %s", *conversationsMessagingIntegrationsWhatsapp.Name, err), resp)
	}

	d.SetId(*whatsAppEmbeddedSignupIntegrationRequest.Id)
	log.Printf("Created conversations messaging integrations whatsapp %s", *whatsAppEmbeddedSignupIntegrationRequest.Id)

	//check if user wants to activate the whatsapp integration
	if activateWhatsapp := d.Get("activate_whatsapp").(*schema.Set); activateWhatsapp != nil {
		if activateWhatsapp.Len() > 0 {
			return activateConversationsMessagingIntegrationsWhatsapp(ctx, d, meta)
		}
	}
	return readConversationsMessagingIntegrationsWhatsapp(ctx, d, meta)
}

// readConversationsMessagingIntegrationsWhatsapp is used by the conversations_messaging_integrations_whatsapp resource to read an conversations messaging integrations whatsapp from genesys cloud
func readConversationsMessagingIntegrationsWhatsapp(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getConversationsMessagingIntegrationsWhatsappProxy(sdkConfig)
	cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceConversationsMessagingIntegrationsWhatsapp(), constants.ConsistencyChecks(), ResourceType)
	log.Printf("Reading conversations messaging integrations whatsapp %s", d.Id())

	return util.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		whatsAppEmbeddedSignupIntegrationRequest, resp, err := proxy.getConversationsMessagingIntegrationsWhatsappById(ctx, d.Id())
		if err != nil {
			if util.IsStatus404(resp) {
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read conversations messaging integrations whatsapp %s: %s", d.Id(), err), resp))
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read conversations messaging integrations whatsapp %s: %s", d.Id(), err), resp))
		}

		resourcedata.SetNillableValue(d, "name", whatsAppEmbeddedSignupIntegrationRequest.Name)

		if whatsAppEmbeddedSignupIntegrationRequest.SupportedContent != nil && whatsAppEmbeddedSignupIntegrationRequest.SupportedContent.Id != nil {
			_ = d.Set("supported_content_id", *whatsAppEmbeddedSignupIntegrationRequest.SupportedContent.Id)
		}

		if whatsAppEmbeddedSignupIntegrationRequest.MessagingSetting != nil && whatsAppEmbeddedSignupIntegrationRequest.MessagingSetting.Id != nil {
			_ = d.Set("messaging_setting_id", *whatsAppEmbeddedSignupIntegrationRequest.MessagingSetting.Id)
		}

		log.Printf("Read conversations messaging integrations whatsapp %s %s", d.Id(), *whatsAppEmbeddedSignupIntegrationRequest.Name)
		return cc.CheckState(d)
	})
}

// updateConversationsMessagingIntegrationsWhatsapp is used by the conversations_messaging_integrations_whatsapp resource to update an conversations messaging integrations whatsapp in Genesys Cloud
func updateConversationsMessagingIntegrationsWhatsapp(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getConversationsMessagingIntegrationsWhatsappProxy(sdkConfig)

	// Activate WhatsApp integration if requested, otherwise proceed with update
	if d.HasChange("activate_whatsapp") {
		if activateWhatsapp := d.Get("activate_whatsapp").(*schema.Set); activateWhatsapp != nil {
			if activateWhatsapp.Len() > 0 {
				return activateConversationsMessagingIntegrationsWhatsapp(ctx, d, meta)
			}
		}
	}

	supportedContentId := d.Get("supported_content_id").(string)
	messagingSettingId := d.Get("messaging_setting_id").(string)

	conversationsMessagingIntegrationsWhatsapp := platformclientv2.Whatsappintegrationupdaterequest{
		Name:             platformclientv2.String(d.Get("name").(string)),
		SupportedContent: &platformclientv2.Supportedcontentreference{Id: &supportedContentId},
		MessagingSetting: &platformclientv2.Messagingsettingrequestreference{Id: &messagingSettingId},
	}

	log.Printf("Updating conversations messaging integrations whatsapp %s", *conversationsMessagingIntegrationsWhatsapp.Name)
	_, resp, err := proxy.updateConversationsMessagingIntegrationsWhatsapp(ctx, d.Id(), &conversationsMessagingIntegrationsWhatsapp)
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to update conversations messaging integrations whatsapp %s: %s", *conversationsMessagingIntegrationsWhatsapp.Name, err), resp)
	}

	log.Printf("Updated conversations messaging integrations whatsapp %s", *conversationsMessagingIntegrationsWhatsapp.Name)
	return readConversationsMessagingIntegrationsWhatsapp(ctx, d, meta)
}

// activateConversationsMessagingIntegrationsWhatsapp is used by the conversations_messaging_integrations_whatsapp resource to activate a WhatsApp integration by submitting the phone number and pin for verification in Genesys Cloud
func activateConversationsMessagingIntegrationsWhatsapp(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	activateWhatsapp := d.Get("activate_whatsapp").(*schema.Set)

	// Extract phone number and pin from the activation data
	activateWhatsappMap := activateWhatsapp.List()[0].(map[string]interface{})
	phoneNumber := activateWhatsappMap["phone_number"].(string)
	pin := activateWhatsappMap["pin"].(string)

	// Return if required fields are empty
	if phoneNumber == "" || pin == "" {
		return nil
	}

	// Get SDK configuration and proxy
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getConversationsMessagingIntegrationsWhatsappProxy(sdkConfig)

	// Construct activation request
	activationRequest := platformclientv2.Whatsappembeddedsignupintegrationactivationrequest{
		PhoneNumber: &phoneNumber,
		Pin:         &pin,
	}

	// Call API for activating the WhatsApp integration
	log.Printf("Activating conversations messaging integrations whatsapp %s", d.Id())
	_, resp, err := proxy.updateConversationsMessagingIntegrationsWhatsappEmbeddedSignup(ctx, d.Id(), &activationRequest)
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to activate conversations messaging integrations whatsapp %s: %s", d.Id(), err), resp)
	}

	log.Printf("Activated conversations messaging integrations whatsapp %s", d.Id())
	// Read back the resource state after activation
	return readConversationsMessagingIntegrationsWhatsapp(ctx, d, meta)
}

// deleteConversationsMessagingIntegrationsWhatsapp is used by the conversations_messaging_integrations_whatsapp resource to delete an conversations messaging integrations whatsapp from Genesys cloud
func deleteConversationsMessagingIntegrationsWhatsapp(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getConversationsMessagingIntegrationsWhatsappProxy(sdkConfig)

	resp, err := proxy.deleteConversationsMessagingIntegrationsWhatsapp(ctx, d.Id())
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to delete conversations messaging integrations whatsapp %s: %s", d.Id(), err), resp)
	}

	return util.WithRetries(ctx, 180*time.Second, func() *retry.RetryError {
		_, resp, err := proxy.getConversationsMessagingIntegrationsWhatsappById(ctx, d.Id())

		if err != nil {
			if util.IsStatus404(resp) {
				log.Printf("Deleted conversations messaging integrations whatsapp %s", d.Id())
				return nil
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Error deleting conversations messaging integrations whatsapp %s: %s", d.Id(), err), resp))
		}

		return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("conversations messaging integrations whatsapp %s still exists", d.Id()), resp))
	})
}
