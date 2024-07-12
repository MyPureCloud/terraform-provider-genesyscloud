package outbound_messagingcampaign

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
	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"

	"terraform-provider-genesyscloud/genesyscloud/consistency_checker"

	"terraform-provider-genesyscloud/genesyscloud/util/constants"
	"terraform-provider-genesyscloud/genesyscloud/util/resourcedata"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
)

/*
The resource_genesyscloud_outbound_messagingcampaign.go contains all of the methods that perform the core logic for a resource.
*/

// getAllAuthOutboundMessagingcampaign retrieves all of the outbound messagingcampaign via Terraform in the Genesys Cloud and is used for the exporter
func getAllAuthOutboundMessagingcampaigns(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	proxy := getOutboundMessagingcampaignProxy(clientConfig)
	resources := make(
		resourceExporter.ResourceIDMetaMap)

	messagingCampaigns, resp, err := proxy.getAllOutboundMessagingcampaign(ctx)
	if err != nil {
		return nil, util.BuildAPIDiagnosticError(resourceName, fmt.Sprintf("Error requesting page of Outbound Messagingcampaign error: %s", err), resp)
	}

	for _, messagingCampaign := range *messagingCampaigns {
		resources[*messagingCampaign.Id] = &resourceExporter.ResourceMeta{Name: *messagingCampaign.Name}
	}

	return resources, nil
}

// createOutboundMessagingcampaign is used by the outbound_messagingcampaign resource to create Genesys cloud outbound messagingcampaign
func createOutboundMessagingcampaign(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getOutboundMessagingcampaignProxy(sdkConfig)

	outboundMessagingcampaign := getOutboundMessagingcampaignFromResourceData(d)

	msg, valid := validateSmsconfig(d.Get("sms_config").(*schema.Set))
	if !valid {
		return util.BuildDiagnosticError(resourceName, "Configuration error", errors.New(msg))
	}

	log.Printf("Creating outbound messagingcampaign %s", *outboundMessagingcampaign.Name)
	messagingCampaign, resp, err := proxy.createOutboundMessagingcampaign(ctx, &outboundMessagingcampaign)
	if err != nil {
		return util.BuildAPIDiagnosticError(resourceName, fmt.Sprintf("Failed to create outbound messagingcampaign %s error: %s", *outboundMessagingcampaign.Name, err), resp)
	}

	d.SetId(*messagingCampaign.Id)
	log.Printf("Created outbound messagingcampaign %s", *messagingCampaign.Id)
	return readOutboundMessagingcampaign(ctx, d, meta)
}

// readOutboundMessagingcampaign is used by the outbound_messagingcampaign resource to read an outbound messagingcampaign from genesys cloud
func readOutboundMessagingcampaign(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getOutboundMessagingcampaignProxy(sdkConfig)
	cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceOutboundMessagingcampaign(), constants.DefaultConsistencyChecks, resourceName)

	log.Printf("Reading outbound messagingcampaign %s", d.Id())

	return util.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		messagingCampaign, resp, getErr := proxy.getOutboundMessagingcampaignById(ctx, d.Id())
		if getErr != nil {
			if util.IsStatus404(resp) {
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(resourceName, fmt.Sprintf("Failed to read Outbound Messagingcampaign %s | error: %s", d.Id(), getErr), resp))
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(resourceName, fmt.Sprintf("Failed to read Outbound Messagingcampaign %s | error: %s", d.Id(), getErr), resp))
		}

		resourcedata.SetNillableValue(d, "name", messagingCampaign.Name)
		resourcedata.SetNillableReference(d, "division_id", messagingCampaign.Division)
		resourcedata.SetNillableValue(d, "campaign_status", messagingCampaign.CampaignStatus)
		resourcedata.SetNillableReference(d, "callable_time_set_id", messagingCampaign.CallableTimeSet)
		resourcedata.SetNillableReference(d, "contact_list_id", messagingCampaign.ContactList)
		if messagingCampaign.DncLists != nil {
			_ = d.Set("dnc_list_ids", util.SdkDomainEntityRefArrToList(*messagingCampaign.DncLists))
		}
		resourcedata.SetNillableValue(d, "always_running", messagingCampaign.AlwaysRunning)
		resourcedata.SetNillableValueWithInterfaceArrayWithFunc(d, "contact_sorts", messagingCampaign.ContactSorts, flattenContactSorts)
		resourcedata.SetNillableValue(d, "messages_per_minute", messagingCampaign.MessagesPerMinute)
		if messagingCampaign.RuleSets != nil {
			_ = d.Set("rule_set_ids", util.SdkDomainEntityRefArrToList(*messagingCampaign.RuleSets))
		}
		if messagingCampaign.ContactListFilters != nil {
			_ = d.Set("contact_list_filter_ids", util.SdkDomainEntityRefArrToList(*messagingCampaign.ContactListFilters))
		}
		resourcedata.SetNillableValueWithInterfaceArrayWithFunc(d, "errors", messagingCampaign.Errors, flattenRestErrorDetails)
		resourcedata.SetNillableValueWithInterfaceArrayWithFunc(d, "dynamic_contact_queueing_settings", messagingCampaign.DynamicContactQueueingSettings, flattenDynamicContactQueueingSettingss)
		//TODO: add email configs in future
		// resourcedata.SetNillableValueWithInterfaceArrayWithFunc(d, "email_config", messagingCampaign.EmailConfig, flattenEmailConfigs)
		d.Set("sms_config", flattenSmsConfigs(messagingCampaign.SmsConfig))

		log.Printf("Read outbound messagingcampaign %s %s", d.Id(), *messagingCampaign.Name)
		return cc.CheckState(d)
	})
}

// updateOutboundMessagingcampaign is used by the outbound_messagingcampaign resource to update an outbound messagingcampaign in Genesys Cloud
func updateOutboundMessagingcampaign(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getOutboundMessagingcampaignProxy(sdkConfig)

	outboundMessagingcampaign := getOutboundMessagingcampaignFromResourceData(d)

	msg, valid := validateSmsconfig(d.Get("sms_config").(*schema.Set))

	if !valid {
		return util.BuildDiagnosticError(resourceName, "Configuration error", errors.New(msg))
	}

	log.Printf("Updating outbound messagingcampaign %s", *outboundMessagingcampaign.Name)
	diagErr := util.RetryWhen(util.IsVersionMismatch, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
		// Get current Outbound Messagingcampaign version
		outboundMessagingcampaignById, resp, getErr := proxy.getOutboundMessagingcampaignById(ctx, d.Id())
		if getErr != nil {
			return resp, util.BuildAPIDiagnosticError(resourceName, fmt.Sprintf("Failed to read Outbound Messagingcampaign %s error: %s", *outboundMessagingcampaign.Name, getErr), resp)
		}
		outboundMessagingcampaign.Version = outboundMessagingcampaignById.Version
		messagingCampaign, resp, updateErr := proxy.updateOutboundMessagingcampaign(ctx, d.Id(), &outboundMessagingcampaign)
		if updateErr != nil {
			return resp, util.BuildAPIDiagnosticError(resourceName, fmt.Sprintf("Failed to update Outbound Messagingcampaign %s error: %s", *messagingCampaign.Name, updateErr), resp)
		}
		return nil, nil
	})
	if diagErr != nil {
		return diagErr
	}

	log.Printf("Updated outbound messagingcampaign")
	return readOutboundMessagingcampaign(ctx, d, meta)
}

// deleteOutboundMessagingcampaign is used by the outbound_messagingcampaign resource to delete an outbound messagingcampaign from Genesys cloud
func deleteOutboundMessagingcampaign(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getOutboundMessagingcampaignProxy(sdkConfig)

	diagErr := util.RetryWhen(util.IsStatus400, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
		log.Printf("Deleting Outbound Messagingcampaign")
		_, resp, err := proxy.deleteOutboundMessagingcampaign(ctx, d.Id())
		if err != nil {
			return resp, util.BuildAPIDiagnosticError(resourceName, fmt.Sprintf("Failed to delete outbound Messagingcampaign %s error: %s", d.Id(), err), resp)
		}
		return resp, nil
	})
	if diagErr != nil {
		return diagErr
	}

	return util.WithRetries(ctx, 180*time.Second, func() *retry.RetryError {
		_, resp, err := proxy.getOutboundMessagingcampaignById(ctx, d.Id())

		if err != nil {
			if util.IsStatus404(resp) {
				log.Printf("Deleted outbound messagingcampaign %s", d.Id())
				return nil
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(resourceName, fmt.Sprintf("Error deleting Outbound Messagingcampaign %s | error: %s", d.Id(), err), resp))
		}

		return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(resourceName, fmt.Sprintf("Outbound Messagingcampaign %s still exists", d.Id()), resp))
	})
}
