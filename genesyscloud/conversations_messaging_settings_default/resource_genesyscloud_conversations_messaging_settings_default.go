package conversations_messaging_settings_default

import (
	"context"
	"fmt"
	"log"
	"terraform-provider-genesyscloud/genesyscloud/consistency_checker"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"terraform-provider-genesyscloud/genesyscloud/util/constants"
	"terraform-provider-genesyscloud/genesyscloud/util/resourcedata"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v146/platformclientv2"
)

func createConversationsMessagingSettingsDefault(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	d.SetId("messaging_settings_default")
	return updateConversationsMessagingSettingsDefault(ctx, d, meta)
}

func readConversationsMessagingSettingsDefault(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getConversationsMessagingSettingsDefaultProxy(sdkConfig)
	cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceConversationsMessagingSettingsDefault(), constants.ConsistencyChecks(), ResourceType)

	log.Printf("Reading conversations messaging settings default %s", d.Id())

	return util.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		messagingSettingDefault, resp, getErr := proxy.getConversationsMessagingSettingsDefault(ctx)
		if getErr != nil {
			if util.IsStatus404(resp) {
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read conversations messaging settings default %s: %s", d.Id(), getErr), resp))
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read conversations messaging settings default %s: %s", d.Id(), getErr), resp))
		}

		resourcedata.SetNillableValue(d, "setting_id", messagingSettingDefault.Id)

		log.Printf("Read conversations messaging settings default %s %s", d.Id(), *messagingSettingDefault.Id)
		return cc.CheckState(d)
	})
}

// updateConversationsMessagingSettingsDefault is used by the conversations_messaging_settings_default resource to update an conversations messaging settings default in Genesys Cloud
func updateConversationsMessagingSettingsDefault(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getConversationsMessagingSettingsDefaultProxy(sdkConfig)
	settingId := d.Get("setting_id").(string)

	updateRequest := platformclientv2.Messagingsettingdefaultrequest{
		SettingId: &settingId,
	}

	log.Printf("Updating conversations messaging settings default %s", settingId)
	messagingSettingDefault, resp, err := proxy.updateConversationsMessagingSettingsDefault(ctx, &updateRequest)
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to update conversations messaging settings default %s: %s", d.Id(), err), resp)
	}

	log.Printf("Updated conversations messaging settings default %s", *messagingSettingDefault.Id)
	return readConversationsMessagingSettingsDefault(ctx, d, meta)
}

func deleteConversationsMessagingSettingsDefault(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getConversationsMessagingSettingsDefaultProxy(sdkConfig)

	log.Printf("Deleting conversations messaging settings default")
	resp, err := proxy.deleteConversationsMessagingSettingsDefault(ctx)
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to delete conversations messaging settings default %s: %s", d.Id(), err), resp)
	}

	return util.WithRetries(ctx, 180*time.Second, func() *retry.RetryError {
		defaultSetting, resp, err := proxy.getConversationsMessagingSettingsDefault(ctx)
		if err != nil {
			if util.IsStatus404(resp) {
				return nil
			}
		}
		if defaultSetting == nil {
			log.Printf("Deleted conversations messaging settings default %s", d.Id())
			return nil
		}
		return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("conversations messaging settings default %s still exists", d.Id()), resp))
	})
}
