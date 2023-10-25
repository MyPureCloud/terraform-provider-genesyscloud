package telephony_providers_edges_phone

import (
	"context"
	"fmt"
	"log"
	"terraform-provider-genesyscloud/genesyscloud/consistency_checker"
	"terraform-provider-genesyscloud/genesyscloud/util/resourcedata"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	gcloud "terraform-provider-genesyscloud/genesyscloud"
	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v115/platformclientv2"
)

func getAllPhones(ctx context.Context, sdkConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	resources := make(resourceExporter.ResourceIDMetaMap)
	pp := getPhoneProxy(sdkConfig)

	phones, err := pp.getAllPhones(ctx)
	if err != nil {
		return nil, diag.Errorf("Failed to get page of phones: %v", err)
	}

	for _, phone := range *phones {
		resources[*phone.Id] = &resourceExporter.ResourceMeta{Name: *phone.Name}
	}

	return resources, nil
}

func createPhone(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*gcloud.ProviderMeta).ClientConfig
	pp := getPhoneProxy(sdkConfig)

	phoneConfig, err := getPhoneFromResourceData(ctx, pp, d)
	if err != nil {
		return diag.Errorf("failed to create phone %v: %v", *phoneConfig.Name, err)
	}

	log.Printf("Creating phone %s", *phoneConfig.Name)
	diagErr := gcloud.RetryWhen(gcloud.IsStatus400, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
		phone, resp, err := pp.createPhone(ctx, phoneConfig)
		if err != nil {
			return resp, diag.Errorf("failed to create phone %s: %s", *phoneConfig.Name, err)
		}

		d.SetId(*phone.Id)

		webRtcUserId := d.Get("web_rtc_user_id")
		if webRtcUserId != "" {
			diagErr := assignUserToWebRtcPhone(ctx, pp, webRtcUserId.(string))
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
	sdkConfig := meta.(*gcloud.ProviderMeta).ClientConfig
	pp := getPhoneProxy(sdkConfig)

	log.Printf("Reading phone %s", d.Id())
	return gcloud.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		currentPhone, resp, getErr := pp.getPhoneById(ctx, d.Id())
		if getErr != nil {
			if gcloud.IsStatus404(resp) {
				return retry.RetryableError(fmt.Errorf("failed to read phone %s: %s", d.Id(), getErr))
			}
			return retry.NonRetryableError(fmt.Errorf("failed to read phone %s: %s", d.Id(), getErr))
		}

		cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourcePhone())
		d.Set("name", *currentPhone.Name)
		d.Set("state", *currentPhone.State)
		d.Set("site_id", *currentPhone.Site.Id)
		d.Set("phone_base_settings_id", *currentPhone.PhoneBaseSettings.Id)
		d.Set("line_base_settings_id", *currentPhone.LineBaseSettings.Id)

		if currentPhone.PhoneMetaBase != nil {
			d.Set("phone_meta_base_id", *currentPhone.PhoneMetaBase.Id)
		}

		if currentPhone.WebRtcUser != nil {
			d.Set("web_rtc_user_id", *currentPhone.WebRtcUser.Id)
		}

		if currentPhone.Lines != nil {
			d.Set("line_addresses", flattenPhoneLines(currentPhone.Lines))
		}

		resourcedata.SetNillableValueWithInterfaceArrayWithFunc(d, "capabilities", currentPhone.Capabilities, flattenPhoneCapabilities)

		log.Printf("Read phone %s %s", d.Id(), *currentPhone.Name)
		return cc.CheckState()
	})
}

func updatePhone(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*gcloud.ProviderMeta).ClientConfig
	pp := getPhoneProxy(sdkConfig)

	phoneConfig, err := getPhoneFromResourceData(ctx, pp, d)
	if err != nil {
		return diag.Errorf("failed to updated phone %v: %v", *phoneConfig.Name, err)
	}

	log.Printf("Updating phone %s", *phoneConfig.Name)
	phone, err := pp.updatePhone(ctx, d.Id(), phoneConfig)
	if err != nil {
		return diag.Errorf("failed to update phone %s: %s", *phoneConfig.Name, err)
	}

	log.Printf("Updated phone %s", *phone.Id)

	webRtcUserId := d.Get("web_rtc_user_id")
	if webRtcUserId != "" {
		if d.HasChange("web_rtc_user_id") {
			diagErr := assignUserToWebRtcPhone(ctx, pp, webRtcUserId.(string))
			if diagErr != nil {
				return diagErr
			}
		}
	}

	return readPhone(ctx, d, meta)
}

func deletePhone(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*gcloud.ProviderMeta).ClientConfig
	pp := getPhoneProxy(sdkConfig)

	log.Printf("Deleting Phone")
	_, err := pp.deletePhone(ctx, d.Id())

	/*
	  Adding a small sleep because when a phone is deleted, the station associated with the phone and the site
	  objects need time to disassociate from the phone. This eventual consistency problem was discovered during
	  building the GCX Now project.  Adding the sleep gives the platform time to settle down.
	*/
	time.Sleep(5 * time.Second)
	if err != nil {
		return diag.Errorf("failed to delete phone: %s", err)
	}

	return gcloud.WithRetries(ctx, 30*time.Second, func() *retry.RetryError {
		phone, resp, err := pp.getPhoneById(ctx, d.Id())
		if err != nil {
			if gcloud.IsStatus404(resp) {
				// Phone deleted
				log.Printf("Deleted Phone %s", d.Id())
				return nil
			}
			return retry.NonRetryableError(fmt.Errorf("error deleting Phone %s: %s", d.Id(), err))
		}

		if phone.State != nil && *phone.State == "deleted" {
			// phone deleted
			log.Printf("Deleted Phone %s", d.Id())
			return nil
		}

		return retry.RetryableError(fmt.Errorf("phone %s still exists", d.Id()))
	})
}
