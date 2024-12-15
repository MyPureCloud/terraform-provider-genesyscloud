package conversations_messaging_integrations_whatsapp

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v146/platformclientv2"
	"log"
	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	"time"

	"terraform-provider-genesyscloud/genesyscloud/consistency_checker"

	gcloud "terraform-provider-genesyscloud/genesyscloud"

	"terraform-provider-genesyscloud/genesyscloud/util/resourcedata"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
)

/*
The resource_genesyscloud_conversations_messaging_integrations_whatsapp.go contains all of the methods that perform the core logic for a resource.
*/

// getAllAuthConversationsMessagingIntegrationsWhatsapp retrieves all of the conversations messaging integrations whatsapp via Terraform in the Genesys Cloud and is used for the exporter
func getAllAuthConversationsMessagingIntegrationsWhatsapps(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	proxy := newConversationsMessagingIntegrationsWhatsappProxy(clientConfig)
	resources := make(resourceExporter.ResourceIDMetaMap)

	whatsAppEmbeddedSignupIntegrationRequests, err := proxy.getAllConversationsMessagingIntegrationsWhatsapp(ctx)
	if err != nil {
		return nil, diag.Errorf("Failed to get conversations messaging integrations whatsapp: %v", err)
	}

	for _, whatsAppEmbeddedSignupIntegrationRequest := range *whatsAppEmbeddedSignupIntegrationRequests {
		resources[*whatsAppEmbeddedSignupIntegrationRequest.Id] = &resourceExporter.ResourceMeta{Name: *whatsAppEmbeddedSignupIntegrationRequest.Name}
	}

	return resources, nil
}

// createConversationsMessagingIntegrationsWhatsapp is used by the conversations_messaging_integrations_whatsapp resource to create Genesys cloud conversations messaging integrations whatsapp
func createConversationsMessagingIntegrationsWhatsapp(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*gcloud.ProviderMeta).ClientConfig
	proxy := getConversationsMessagingIntegrationsWhatsappProxy(sdkConfig)

	conversationsMessagingIntegrationsWhatsapp := getConversationsMessagingIntegrationsWhatsappFromResourceData(d)

	log.Printf("Creating conversations messaging integrations whatsapp %s", *conversationsMessagingIntegrationsWhatsapp.Name)
	whatsAppEmbeddedSignupIntegrationRequest, err := proxy.createConversationsMessagingIntegrationsWhatsapp(ctx, &conversationsMessagingIntegrationsWhatsapp)
	if err != nil {
		return diag.Errorf("Failed to create conversations messaging integrations whatsapp: %s", err)
	}

	d.SetId(*whatsAppEmbeddedSignupIntegrationRequest.Id)
	log.Printf("Created conversations messaging integrations whatsapp %s", *whatsAppEmbeddedSignupIntegrationRequest.Id)
	return readConversationsMessagingIntegrationsWhatsapp(ctx, d, meta)
}

// readConversationsMessagingIntegrationsWhatsapp is used by the conversations_messaging_integrations_whatsapp resource to read an conversations messaging integrations whatsapp from genesys cloud
func readConversationsMessagingIntegrationsWhatsapp(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*gcloud.ProviderMeta).ClientConfig
	proxy := getConversationsMessagingIntegrationsWhatsappProxy(sdkConfig)

	log.Printf("Reading conversations messaging integrations whatsapp %s", d.Id())

	return gcloud.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		whatsAppEmbeddedSignupIntegrationRequest, respCode, getErr := proxy.getConversationsMessagingIntegrationsWhatsappById(ctx, d.Id())
		if getErr != nil {
			if gcloud.IsStatus404ByInt(respCode) {
				return retry.RetryableError(fmt.Errorf("Failed to read conversations messaging integrations whatsapp %s: %s", d.Id(), getErr))
			}
			return retry.NonRetryableError(fmt.Errorf("Failed to read conversations messaging integrations whatsapp %s: %s", d.Id(), getErr))
		}

		cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceConversationsMessagingIntegrationsWhatsapp())

		resourcedata.SetNillableValue(d, "name", whatsAppEmbeddedSignupIntegrationRequest.Name)
		resourcedata.SetNillableValueWithInterfaceArrayWithFunc(d, "supported_content", whatsAppEmbeddedSignupIntegrationRequest.SupportedContent, flattenSupportedContentReference)
		resourcedata.SetNillableValueWithInterfaceArrayWithFunc(d, "messaging_setting", whatsAppEmbeddedSignupIntegrationRequest.MessagingSetting, flattenMessagingSettingRequestReference)
		resourcedata.SetNillableValue(d, "embedded_signup_access_token", whatsAppEmbeddedSignupIntegrationRequest.EmbeddedSignupAccessToken)

		log.Printf("Read conversations messaging integrations whatsapp %s %s", d.Id(), *whatsAppEmbeddedSignupIntegrationRequest.Name)
		return cc.CheckState()
	})
}

// updateConversationsMessagingIntegrationsWhatsapp is used by the conversations_messaging_integrations_whatsapp resource to update an conversations messaging integrations whatsapp in Genesys Cloud
func updateConversationsMessagingIntegrationsWhatsapp(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*gcloud.ProviderMeta).ClientConfig
	proxy := getConversationsMessagingIntegrationsWhatsappProxy(sdkConfig)

	conversationsMessagingIntegrationsWhatsapp := getConversationsMessagingIntegrationsWhatsappFromResourceData(d)

	log.Printf("Updating conversations messaging integrations whatsapp %s", *conversationsMessagingIntegrationsWhatsapp.Name)
	whatsAppEmbeddedSignupIntegrationRequest, err := proxy.updateConversationsMessagingIntegrationsWhatsapp(ctx, d.Id(), &conversationsMessagingIntegrationsWhatsapp)
	if err != nil {
		return diag.Errorf("Failed to update conversations messaging integrations whatsapp: %s", err)
	}

	log.Printf("Updated conversations messaging integrations whatsapp %s", *whatsAppEmbeddedSignupIntegrationRequest.Id)
	return readConversationsMessagingIntegrationsWhatsapp(ctx, d, meta)
}

// deleteConversationsMessagingIntegrationsWhatsapp is used by the conversations_messaging_integrations_whatsapp resource to delete an conversations messaging integrations whatsapp from Genesys cloud
func deleteConversationsMessagingIntegrationsWhatsapp(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*gcloud.ProviderMeta).ClientConfig
	proxy := getConversationsMessagingIntegrationsWhatsappProxy(sdkConfig)

	_, err := proxy.deleteConversationsMessagingIntegrationsWhatsapp(ctx, d.Id())
	if err != nil {
		return diag.Errorf("Failed to delete conversations messaging integrations whatsapp %s: %s", d.Id(), err)
	}

	return gcloud.WithRetries(ctx, 180*time.Second, func() *retry.RetryError {
		_, respCode, err := proxy.getConversationsMessagingIntegrationsWhatsappById(ctx, d.Id())

		if err != nil {
			if gcloud.IsStatus404ByInt(respCode) {
				log.Printf("Deleted conversations messaging integrations whatsapp %s", d.Id())
				return nil
			}
			return retry.NonRetryableError(fmt.Errorf("Error deleting conversations messaging integrations whatsapp %s: %s", d.Id(), err))
		}

		return retry.RetryableError(fmt.Errorf("conversations messaging integrations whatsapp %s still exists", d.Id()))
	})
}
