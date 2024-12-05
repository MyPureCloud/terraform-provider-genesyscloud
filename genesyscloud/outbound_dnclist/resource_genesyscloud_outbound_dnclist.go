package outbound_dnclist

import (
	"context"
	"fmt"
	"log"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"terraform-provider-genesyscloud/genesyscloud/util/constants"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"terraform-provider-genesyscloud/genesyscloud/consistency_checker"

	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	lists "terraform-provider-genesyscloud/genesyscloud/util/lists"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v146/platformclientv2"
)

func getAllOutboundDncLists(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	resources := make(resourceExporter.ResourceIDMetaMap)
	proxy := getOutboundDnclistProxy(clientConfig)

	dnclists, resp, err := proxy.getAllOutboundDnclist(ctx)
	if err != nil {
		return nil, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to get dnclists error: %s", err), resp)
	}
	for _, dncListConfig := range *dnclists {
		resources[*dncListConfig.Id] = &resourceExporter.ResourceMeta{BlockLabel: *dncListConfig.Name}
	}
	return resources, nil
}

func createOutboundDncList(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)
	contactMethod := d.Get("contact_method").(string)
	loginId := d.Get("login_id").(string)
	campaignId := d.Get("campaign_id").(string)
	licenseId := d.Get("license_id").(string)
	dncSourceType := d.Get("dnc_source_type").(string)
	dncCodes := lists.InterfaceListToStrings(d.Get("dnc_codes").([]interface{}))
	entries := d.Get("entries").([]interface{})

	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getOutboundDnclistProxy(sdkConfig)

	sdkDncListCreate := platformclientv2.Dnclistcreate{
		DncCodes: &dncCodes,
		Division: util.BuildSdkDomainEntityRef(d, "division_id"),
	}

	if name != "" {
		sdkDncListCreate.Name = &name
	}
	if contactMethod != "" {
		sdkDncListCreate.ContactMethod = &contactMethod
	}
	if loginId != "" {
		sdkDncListCreate.LoginId = &loginId
	}
	if campaignId != "" {
		sdkDncListCreate.CampaignId = &campaignId
	}
	if licenseId != "" {
		sdkDncListCreate.LicenseId = &licenseId
	}
	if dncSourceType != "" {
		sdkDncListCreate.DncSourceType = &dncSourceType
	}

	log.Printf("Creating Outbound DNC list %s", name)
	outboundDncList, resp, err := proxy.createOutboundDnclist(ctx, &sdkDncListCreate)
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to create Outbound DNC list %s error: %s", name, err), resp)
	}

	d.SetId(*outboundDncList.Id)

	if len(entries) > 0 {
		if *sdkDncListCreate.DncSourceType == "rds" {
			for _, entry := range entries {
				resp, err := proxy.uploadPhoneEntriesToDncList(outboundDncList, entry)
				if err != nil {
					return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to create Outbound DNC list %s error: %v", name, err), resp)
				}
			}
		} else {
			return util.BuildDiagnosticError(ResourceType, "Phone numbers can only be uploaded to internal DNC lists.", fmt.Errorf("phone numbers can only be uploaded to internal DNC Lists"))
		}
	}
	log.Printf("Created Outbound DNC list %s %s", name, *outboundDncList.Id)
	return readOutboundDncList(ctx, d, meta)
}

func updateOutboundDncList(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)
	contactMethod := d.Get("contact_method").(string)
	loginId := d.Get("login_id").(string)
	campaignId := d.Get("campaign_id").(string)
	dncCodes := lists.InterfaceListToStrings(d.Get("dnc_codes").([]interface{}))
	licenseId := d.Get("license_id").(string)
	dncSourceType := d.Get("dnc_source_type").(string)
	entries := d.Get("entries").([]interface{})

	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getOutboundDnclistProxy(sdkConfig)

	sdkDncList := platformclientv2.Dnclist{
		DncCodes: &dncCodes,
		Division: util.BuildSdkDomainEntityRef(d, "division_id"),
	}

	if name != "" {
		sdkDncList.Name = &name
	}
	if contactMethod != "" {
		sdkDncList.ContactMethod = &contactMethod
	}
	if loginId != "" {
		sdkDncList.LoginId = &loginId
	}
	if campaignId != "" {
		sdkDncList.CampaignId = &campaignId
	}
	if licenseId != "" {
		sdkDncList.LicenseId = &licenseId
	}
	if dncSourceType != "" {
		sdkDncList.DncSourceType = &dncSourceType
	}
	log.Printf("Updating Outbound DNC list %s", name)
	diagErr := util.RetryWhen(util.IsVersionMismatch, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
		// Get current Outbound DNC list version
		outboundDncList, resp, getErr := proxy.getOutboundDnclistById(ctx, d.Id())
		if getErr != nil {
			return resp, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to read Outbound DNC list %s error: %s", name, getErr), resp)
		}
		sdkDncList.Version = outboundDncList.Version
		outboundDncList, response, updateErr := proxy.updateOutboundDnclist(ctx, d.Id(), &sdkDncList)
		if updateErr != nil {
			return resp, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to update Outbound DNC list %s error: %s", name, updateErr), response)
		}
		if len(entries) > 0 {
			if *sdkDncList.DncSourceType == "rds" {
				for _, entry := range entries {
					response, err := proxy.uploadPhoneEntriesToDncList(outboundDncList, entry)
					if err != nil {
						return resp, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to update Outbound DNC list %s error: %v", name, err), response)
					}
				}
			} else {
				return nil, util.BuildDiagnosticError(ResourceType, "Phone numbers can only be uploaded to internal DNC lists", fmt.Errorf("phone numbers can only be uploaded to internal DNC lists"))
			}
		}
		return nil, nil
	})
	if diagErr != nil {
		return diagErr
	}

	log.Printf("Updated Outbound DNC list %s", name)
	return readOutboundDncList(ctx, d, meta)
}

func readOutboundDncList(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getOutboundDnclistProxy(sdkConfig)
	cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceOutboundDncList(), constants.ConsistencyChecks(), ResourceType)

	log.Printf("Reading Outbound DNC list %s", d.Id())

	return util.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		sdkDncList, resp, getErr := proxy.getOutboundDnclistById(ctx, d.Id())
		if getErr != nil {
			if util.IsStatus404(resp) {
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("failed to read Outbound DNC list %s | error: %s", d.Id(), getErr), resp))
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("failed to read Outbound DNC list %s | error: %s", d.Id(), getErr), resp))
		}

		if sdkDncList.Name != nil {
			_ = d.Set("name", *sdkDncList.Name)
		}
		if sdkDncList.ContactMethod != nil {
			_ = d.Set("contact_method", *sdkDncList.ContactMethod)
		}
		if sdkDncList.LoginId != nil {
			_ = d.Set("login_id", *sdkDncList.LoginId)
		}
		if sdkDncList.CampaignId != nil {
			_ = d.Set("campaign_id", *sdkDncList.CampaignId)
		}
		if sdkDncList.DncCodes != nil {
			schemaCodes := lists.InterfaceListToStrings(d.Get("dnc_codes").([]interface{}))
			// preserve ordering and avoid a plan not empty error
			if lists.AreEquivalent(schemaCodes, *sdkDncList.DncCodes) {
				_ = d.Set("dnc_codes", schemaCodes)
			} else {
				_ = d.Set("dnc_codes", lists.StringListToInterfaceList(*sdkDncList.DncCodes))
			}
		}
		if sdkDncList.DncSourceType != nil {
			_ = d.Set("dnc_source_type", *sdkDncList.DncSourceType)
		}
		if sdkDncList.LicenseId != nil {
			_ = d.Set("license_id", *sdkDncList.LicenseId)
		}
		if sdkDncList.Division != nil && sdkDncList.Division.Id != nil {
			_ = d.Set("division_id", *sdkDncList.Division.Id)
		}
		log.Printf("Read Outbound DNC list %s %s", d.Id(), *sdkDncList.Name)
		return cc.CheckState(d)
	})
}

func deleteOutboundDncList(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getOutboundDnclistProxy(sdkConfig)

	diagErr := util.RetryWhen(util.IsStatus400, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
		log.Printf("Deleting Outbound DNC list")
		resp, err := proxy.deleteOutboundDnclist(ctx, d.Id())
		if err != nil {
			return resp, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to delete Outbound DNC list %s error: %s", d.Id(), err), resp)
		}
		return resp, nil
	})
	if diagErr != nil {
		return diagErr
	}

	return util.WithRetries(ctx, 30*time.Second, func() *retry.RetryError {
		_, resp, err := proxy.getOutboundDnclistById(ctx, d.Id())
		if err != nil {
			if util.IsStatus404(resp) {
				// Outbound DNC list deleted
				log.Printf("Deleted Outbound DNC list %s", d.Id())
				return nil
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("error deleting Outbound DNC list %s | error: %s", d.Id(), err), resp))
		}
		return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Outbound DNC list %s still exists", d.Id()), resp))
	})
}

func GenerateOutboundDncListBasic(resourceLabel string, name string) string {
	return fmt.Sprintf(`
resource "genesyscloud_outbound_dnclist" "%s" {
	name            = "%s"
	dnc_source_type = "rds"
	contact_method  = "Phone"
}
`, resourceLabel, name)
}
