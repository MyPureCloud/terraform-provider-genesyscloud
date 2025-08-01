package telephony_providers_edges_phone

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/consistency_checker"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/constants"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/resourcedata"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	resourceExporter "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_exporter"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v162/platformclientv2"
)

func getAllPhones(ctx context.Context, sdkConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	resources := make(resourceExporter.ResourceIDMetaMap)
	pp := getPhoneProxy(sdkConfig)

	phones, resp, err := pp.getAllPhones(ctx)
	if err != nil {
		return nil, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to get page of phones error: %s", err), resp)
	}

	for _, phone := range *phones {
		resources[*phone.Id] = &resourceExporter.ResourceMeta{BlockLabel: *phone.Name}
	}
	return resources, nil
}

func createPhone(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	pp := getPhoneProxy(sdkConfig)

	phoneConfig, err := getPhoneFromResourceData(ctx, pp, d)
	if err != nil {
		return util.BuildDiagnosticError(ResourceType, fmt.Sprintf("failed to create phone %v", *phoneConfig.Name), err)
	}

	log.Printf("Creating phone %s", *phoneConfig.Name)

	diagErr := util.RetryWhen(util.IsStatus404, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
		phone, resp, err := pp.createPhone(ctx, phoneConfig)
		if err != nil {
			return resp, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to create phone %s error: %s", *phoneConfig.Name, err), resp)
		}
		log.Printf("Completed call to create phone name %s with status code %d, correlation id %s", *phoneConfig.Name, resp.StatusCode, resp.CorrelationID)

		d.SetId(*phone.Id)

		webRtcUserId := d.Get("web_rtc_user_id")
		if webRtcUserId != "" {
			diagErr := assignUserToWebRtcPhone(ctx, pp, webRtcUserId.(string), *phone.Id)
			if diagErr != nil {
				return resp, diagErr
			}
		}

		log.Printf("Created phone %s", *phone.Id)
		return nil, nil
	})

	if diagErr != nil {
		return diagErr
	}

	return readPhone(ctx, d, meta)
}

func readPhone(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	pp := getPhoneProxy(sdkConfig)
	cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourcePhone(), constants.ConsistencyChecks(), ResourceType)

	log.Printf("Reading phone %s", d.Id())
	return util.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		currentPhone, resp, getErr := pp.getPhoneById(ctx, d.Id())
		if getErr != nil {
			if util.IsStatus404(resp) {
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("failed to read phone %s | error: %s", d.Id(), getErr), resp))
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("failed to read phone %s | error: %s", d.Id(), getErr), resp))
		}

		if currentPhone.Site != nil && currentPhone.Site.Id != nil {
			_ = d.Set("site_id", *currentPhone.Site.Id)
			log.Printf("Phone ID = %s and the site_id = %s", d.Id(), *currentPhone.Site.Id)
		} else {
			log.Printf("Phone ID = %s and the site_id is nil", d.Id())
		}

		if currentPhone.Name != nil {
			_ = d.Set("name", *currentPhone.Name)
		}

		if currentPhone.PhoneBaseSettings != nil && currentPhone.PhoneBaseSettings.Id != nil {
			_ = d.Set("phone_base_settings_id", *currentPhone.PhoneBaseSettings.Id)
		}

		if currentPhone.State != nil {
			_ = d.Set("state", *currentPhone.State)
		}

		if currentPhone.LineBaseSettings != nil {
			_ = d.Set("line_base_settings_id", *currentPhone.LineBaseSettings.Id)
		}

		if currentPhone.PhoneMetaBase != nil {
			_ = d.Set("phone_meta_base_id", *currentPhone.PhoneMetaBase.Id)
		}

		if currentPhone.WebRtcUser != nil {
			_ = d.Set("web_rtc_user_id", *currentPhone.WebRtcUser.Id)
		}

		if currentPhone.Lines != nil {
			resourcedata.SetNillableValueWithInterfaceArrayWithFunc(d, "line_properties", currentPhone.Lines, flattenLines)
		}

		_ = d.Set("properties", nil)
		if currentPhone.Properties != nil {
			properties, err := util.FlattenTelephonyProperties(currentPhone.Properties)
			if err != nil {
				return retry.NonRetryableError(fmt.Errorf("%v", err))
			}
			_ = d.Set("properties", properties)
		}

		resourcedata.SetNillableValueWithInterfaceArrayWithFunc(d, "capabilities", currentPhone.Capabilities, flattenPhoneCapabilities)

		log.Printf("Read phone %s %s", d.Id(), *currentPhone.Name)
		return cc.CheckState(d)
	})
}

func updatePhone(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	pp := getPhoneProxy(sdkConfig)

	phoneConfig, err := getPhoneFromResourceData(ctx, pp, d)
	if err != nil {
		return util.BuildDiagnosticError(ResourceType, fmt.Sprintf("failed to updated phone %v", *phoneConfig.Name), err)
	}
	log.Printf("Updating phone %s", *phoneConfig.Name)
	phone, resp, err := pp.updatePhone(ctx, d.Id(), phoneConfig)
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to update phone %s error: %s", *phoneConfig.Name, err), resp)
	}

	log.Printf("Updated phone %s", *phone.Id)

	webRtcUserId := d.Get("web_rtc_user_id")
	if webRtcUserId != "" {
		diagErr := assignUserToWebRtcPhone(ctx, pp, webRtcUserId.(string), *phone.Id)
		if diagErr != nil {
			return diagErr
		}
	}

	return readPhone(ctx, d, meta)
}

func deletePhone(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	pp := getPhoneProxy(sdkConfig)

	log.Printf("Deleting Phone")
	resp, err := pp.deletePhone(ctx, d.Id())

	/*
	  Adding a small sleep because when a phone is deleted, the station associated with the phone and the site
	  objects need time to disassociate from the phone. This eventual consistency problem was discovered during
	  building the GCX Now project.  Adding the sleep gives the platform time to settle down.
	*/
	time.Sleep(5 * time.Second)
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to delete phone %s error: %s", d.Id(), err), resp)
	}

	return util.WithRetries(ctx, 30*time.Second, func() *retry.RetryError {
		phone, resp, err := pp.getPhoneById(ctx, d.Id())
		if err != nil {
			if util.IsStatus404(resp) {
				// Phone deleted
				log.Printf("Deleted Phone %s", d.Id())
				return nil
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("error deleting Phone %s | error: %s", d.Id(), err), resp))
		}

		if phone.State != nil && *phone.State == "deleted" {
			// phone deleted
			log.Printf("Deleted Phone %s", d.Id())
			return nil
		}

		return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("phone %s still exists", d.Id()), resp))
	})
}
