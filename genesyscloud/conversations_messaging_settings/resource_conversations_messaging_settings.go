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

func getAllAuthConversationsMessagingSettings(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	proxy := getConversationsMessagingSettingsProxy(clientConfig)
	resources := make(resourceExporter.ResourceIDMetaMap)

	messagingSettings, resp, err := proxy.getAllConversationsMessagingSettings(ctx)
	if err != nil {
		return nil, util.BuildAPIDiagnosticError(resourceName, fmt.Sprintf("Failed to get Conversations messaging Settings: %s", err), resp)
	}

	for _, messagingSetting := range *messagingSettings {
		resources[*messagingSetting.Id] = &resourceExporter.ResourceMeta{Name: *messagingSetting.Name}
	}

	return resources, nil
}

func createConversationsMessagingSettings(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getConversationsMessagingSettingsProxy(sdkConfig)

	conversationsMessagingSettingsReq := getConversationsMessagingSettingsFromResourceData(d)

	log.Printf("Creating conversations messaging settings %s", *conversationsMessagingSettingsReq.Name)
	messagingSetting, resp, err := proxy.createConversationsMessagingSettings(ctx, &conversationsMessagingSettingsReq)
	if err != nil {
		return util.BuildAPIDiagnosticError(resourceName, fmt.Sprintf("Failed to create conversations messaging setting %s error: %s", *conversationsMessagingSettingsReq.Name, err), resp)
	}

	d.SetId(*messagingSetting.Id)
	log.Printf("Created conversations messaging settings %s", *messagingSetting.Id)
	return readConversationsMessagingSettings(ctx, d, meta)
}

func readConversationsMessagingSettings(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getConversationsMessagingSettingsProxy(sdkConfig)
	cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceConversationsMessagingSettings(), constants.DefaultConsistencyChecks, resourceName)

	log.Printf("Reading conversations messaging settings %s", d.Id())

	return util.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		messagingSetting, resp, err := proxy.getConversationsMessagingSettingsById(ctx, d.Id())
		if err != nil {
			if util.IsStatus404(resp) {
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(resourceName, fmt.Sprintf("Failed to read conversations messaging settings %s | error: %s", d.Id(), err), resp))
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(resourceName, fmt.Sprintf("Failed to read conversations messaging settings %s | error: %s", d.Id(), err), resp))
		}

		resourcedata.SetNillableValue(d, "name", messagingSetting.Name)
		resourcedata.SetNillableValueWithInterfaceArrayWithFunc(d, "content", messagingSetting.Content, flattenContentSettings)
		resourcedata.SetNillableValueWithInterfaceArrayWithFunc(d, "event", messagingSetting.Event, flattenEventSettings)

		log.Printf("Read conversations messaging settings %s", d.Id())
		return cc.CheckState(d)
	})
}

func updateConversationsMessagingSettings(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getConversationsMessagingSettingsProxy(sdkConfig)

	name := d.Get("name").(string)
	content := d.Get("content").([]interface{})
	event := d.Get("event").([]interface{})

	var conversationsMessagingSettings platformclientv2.Messagingsettingpatchrequest

	if name != "" {
		conversationsMessagingSettings.Name = &name
	}
	if content != nil {
		conversationsMessagingSettings.Content = buildContentSettings(content)
	}
	if event != nil {
		conversationsMessagingSettings.Event = buildEventSetting(event)
	}

	_, resp, err := proxy.updateConversationsMessagingSettings(ctx, d.Id(), &conversationsMessagingSettings)
	if err != nil {
		return util.BuildAPIDiagnosticError(resourceName, fmt.Sprintf("Failed to update conversations messaging settings %s error: %s", d.Id(), err), resp)
	}

	log.Printf("Updated conversations messaging settings %s", d.Id())
	return readConversationsMessagingSettings(ctx, d, meta)
}

func deleteConversationsMessagingSettings(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getConversationsMessagingSettingsProxy(sdkConfig)

	resp, err := proxy.deleteConversationsMessagingSettings(ctx, d.Id())
	if err != nil {
		return util.BuildAPIDiagnosticError(resourceName, fmt.Sprintf("Failed to delete conversations messaging setting %s error: %s", d.Id(), err), resp)
	}

	return util.WithRetries(ctx, 180*time.Second, func() *retry.RetryError {
		_, resp, err := proxy.getConversationsMessagingSettingsById(ctx, d.Id())
		if err != nil {
			if util.IsStatus404(resp) {
				log.Printf("Deleted Conversations messaging Setting")
				return nil
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(resourceName, fmt.Sprintf("Error deleting Conversations messaging Setting: %s | error: %s", d.Id(), err), resp))
		}
		return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(resourceName, fmt.Sprintf("Conversations messaging Setting %s still exists", d.Id()), resp))
	})
}
