package conversations_messaging_settings

import (
	"context"
	"fmt"
	"log"
	"terraform-provider-genesyscloud/genesyscloud/consistency_checker"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"terraform-provider-genesyscloud/genesyscloud/util/constants"
	"terraform-provider-genesyscloud/genesyscloud/util/resourcedata"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"
)

func getAllAuthConversationMessagingSettingss(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	proxy := getConversationMessagingSettingsProxy(clientConfig)
	resources := make(resourceExporter.ResourceIDMetaMap)

	messagingSettings, resp, err := proxy.getAllConversationMessagingSettings(ctx)
	if err != nil {
		return nil, util.BuildAPIDiagnosticError(resourceName, fmt.Sprintf("Failed to get Conversation Messaging Settings: %s", err), resp)
	}

	for _, messagingSetting := range *messagingSettings {
		log.Printf("Dealing with messaging setting id : %s", *messagingSetting.Id)
		resources[*messagingSetting.Id] = &resourceExporter.ResourceMeta{Name: *messagingSetting.Name}
	}

	return resources, nil
}

func createConversationMessagingSettings(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getConversationMessagingSettingsProxy(sdkConfig)

	conversationMessagingSettingsReq := getConversationMessagingSettingsFromResourceData(d)

	log.Printf("Creating conversation messaging settings %s", *conversationMessagingSettingsReq.Name)
	messagingSetting, resp, err := proxy.createConversationMessagingSettings(ctx, &conversationMessagingSettingsReq)
	if err != nil {
		return util.BuildAPIDiagnosticError(resourceName, fmt.Sprintf("Failed to create conversation messaging setting %s error: %s", *conversationMessagingSettingsReq.Name, err), resp)
	}

	d.SetId(*messagingSetting.Id)
	log.Printf("Created conversation messaging settings %s", *messagingSetting.Id)
	log.Println("Hit1: ", messagingSetting.String())
	return readConversationMessagingSettings(ctx, d, meta)
}

func readConversationMessagingSettings(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getConversationMessagingSettingsProxy(sdkConfig)
	cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceConversationMessagingSettings(), constants.DefaultConsistencyChecks, resourceName)

	log.Printf("Reading conversation messaging settings %s", d.Id())

	return util.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		messagingSetting, resp, err := proxy.getConversationMessagingSettingsById(ctx, d.Id())
		if err != nil {
			if util.IsStatus404(resp) {
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(resourceName, fmt.Sprintf("Failed to read conversation messaging settings %s | error: %s", d.Id(), err), resp))
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(resourceName, fmt.Sprintf("Failed to read conversation messaging settings %s | error: %s", d.Id(), err), resp))
		}

		resourcedata.SetNillableValue(d, "name", messagingSetting.Name)
		resourcedata.SetNillableValueWithInterfaceArrayWithFunc(d, "content", messagingSetting.Content, flattenContentSettings)
		resourcedata.SetNillableValueWithInterfaceArrayWithFunc(d, "event", messagingSetting.Event, flattenEventSettings)

		log.Printf("Read conversation messaging settings %s", d.Id())
		log.Println("Hit2: ", messagingSetting.String())
		log.Println("\nHit2.5: ", d.State().String())
		return cc.CheckState(d)
	})
}

func updateConversationMessagingSettings(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getConversationMessagingSettingsProxy(sdkConfig)

	name := d.Get("name").(string)
	content := d.Get("content").([]interface{})
	event := d.Get("event").([]interface{})

	var conversationMessagingSettings *platformclientv2.Messagingsettingpatchrequest

	if name != "" {
		conversationMessagingSettings.Name = &name
	}
	if content != nil {
		conversationMessagingSettings.Content = buildContentSettings(content)
	}
	if event != nil {
		conversationMessagingSettings.Event = buildEventSetting(event)
	}

	_, resp, err := proxy.updateConversationMessagingSettings(ctx, d.Id(), conversationMessagingSettings)
	if err != nil {
		return util.BuildAPIDiagnosticError(resourceName, fmt.Sprintf("Failed to update conversation messaging settings %s error: %s", d.Id(), err), resp)
	}

	log.Printf("Updated conversation messaging settings %s", d.Id())
	return readConversationMessagingSettings(ctx, d, meta)
}

func deleteConversationMessagingSettings(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getConversationMessagingSettingsProxy(sdkConfig)

	resp, err := proxy.deleteConversationMessagingSettings(ctx, d.Id())
	if err != nil {
		return util.BuildAPIDiagnosticError(resourceName, fmt.Sprintf("Failed to delete conversation messaging setting %s error: %s", d.Id(), err), resp)
	}

	return util.WithRetries(ctx, 180*time.Second, func() *retry.RetryError {
		_, resp, err := proxy.getConversationMessagingSettingsById(ctx, d.Id())
		if err != nil {
			if util.IsStatus404(resp) {
				log.Printf("Deleted Conversation Messaging Setting")
				return nil
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(resourceName, fmt.Sprintf("Error deleting Conversation Messaging Setting: %s | error: %s", d.Id(), err), resp))
		}
		return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(resourceName, fmt.Sprintf("Conversation Messaging Setting %s still exists", d.Id()), resp))
	})
}
