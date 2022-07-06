package genesyscloud

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/consistency_checker"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/mypurecloud/platform-client-sdk-go/v80/platformclientv2"
)

func getAllOutboundDncLists(_ context.Context, clientConfig *platformclientv2.Configuration) (ResourceIDMetaMap, diag.Diagnostics) {
	resources := make(ResourceIDMetaMap)
	outboundAPI := platformclientv2.NewOutboundApiWithConfig(clientConfig)

	for pageNum := 1; ; pageNum++ {
		const pageSize = 100
		dncListConfigs, _, getErr := outboundAPI.GetOutboundDnclists(false, false, pageSize, pageNum, true, "", "", "", []string{}, "", "")
		if getErr != nil {
			return nil, diag.Errorf("Failed to get page of dnc list configs: %v", getErr)
		}
		if dncListConfigs.Entities == nil || len(*dncListConfigs.Entities) == 0 {
			break
		}
		for _, dncListConfig := range *dncListConfigs.Entities {
			resources[*dncListConfig.Id] = &ResourceMeta{Name: *dncListConfig.Name}
		}
	}

	return resources, nil
}

func outboundDncListExporter() *ResourceExporter {
	return &ResourceExporter{
		GetResourcesFunc: getAllWithPooledClient(getAllOutboundDncLists),
		RefAttrs: map[string]*RefAttrSettings{
			"division_id": {RefType: "genesyscloud_auth_division"},
		},
	}
}

func resourceOutboundDncList() *schema.Resource {
	return &schema.Resource{
		Description: `Genesys Cloud Outbound DNC List`,

		CreateContext: createWithPooledClient(createOutboundDncList),
		ReadContext:   readWithPooledClient(readOutboundDncList),
		UpdateContext: updateWithPooledClient(updateOutboundDncList),
		DeleteContext: deleteWithPooledClient(deleteOutboundDncList),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			`name`: {
				Description: `The name of the DncList.`,
				Required:    true,
				Type:        schema.TypeString,
			},
			`contact_method`: {
				Description:  `The contact method. Required if dncSourceType is rds.`,
				Optional:     true,
				Type:         schema.TypeString,
				ValidateFunc: validation.StringInSlice([]string{`Email`, `Phone`}, false),
			},
			`login_id`: {
				Description: `A dnc.com loginId. Required if the dncSourceType is dnc.com.`,
				Optional:    true,
				Type:        schema.TypeString,
			},
			`campaign_id`: {
				Description: `A dnc.com campaignId. Optional if the dncSourceType is dnc.com.`,
				Optional:    true,
				Type:        schema.TypeString,
			},
			`dnc_codes`: {
				Description: `The list of dnc.com codes to be treated as DNC. Required if the dncSourceType is dnc.com.`,
				Optional:    true,
				Type:        schema.TypeList,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringInSlice([]string{`B`, `C`, `D`, `E`, `F`, `G`, `H`, `I`, `L`, `M`, `O`, `P`, `R`, `S`, `T`, `V`, `W`, `X`, `Y`}, false),
				},
			},
			`license_id`: {
				Description: `A gryphon license number. Required if the dncSourceType is gryphon.`,
				Optional:    true,
				Type:        schema.TypeString,
			},
			`division_id`: {
				Description: `The division this DNC List belongs to.`,
				Optional:    true,
				Computed:    true,
				Type:        schema.TypeString,
			},
			`dnc_source_type`: {
				Description:  `The type of the DNC List.`,
				Required:     true,
				ForceNew:     true,
				Type:         schema.TypeString,
				ValidateFunc: validation.StringInSlice([]string{`rds`, `dnc.com`, `gryphon`}, false),
			},
			`entries`: {
				Description: `Rows to add to the DNC list. To emulate removing phone numbers, you can set expiration_date to a date in the past.`,
				Optional:    true,
				Type:        schema.TypeList,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						`expiration_date`: {
							Description:      `Expiration date for DNC phone numbers in yyyy-MM-ddTHH:mmZ format.`,
							Optional:         true,
							Type:             schema.TypeString,
							ValidateDiagFunc: validateDateTime,
						},
						`phone_numbers`: {
							Description: `Phone numbers to add to a DNC list. Only possible if the dncSourceType is rds.`,
							Optional:    true,
							Type:        schema.TypeList,
							Elem: &schema.Schema{
								Type:             schema.TypeString,
								ValidateDiagFunc: validatePhoneNumber,
							},
						},
					},
				},
			},
		},
	}
}

func createOutboundDncList(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)
	contactMethod := d.Get("contact_method").(string)
	loginId := d.Get("login_id").(string)
	//campaignId := d.Get("campaign_id").(string)
	licenseId := d.Get("license_id").(string)
	dncSourceType := d.Get("dnc_source_type").(string)
	dncCodes := interfaceListToStrings(d.Get("dnc_codes").([]interface{}))
	entries := d.Get("entries").([]interface{})

	sdkConfig := meta.(*providerMeta).ClientConfig
	outboundApi := platformclientv2.NewOutboundApiWithConfig(sdkConfig)

	sdkDncListCreate := platformclientv2.Dnclistcreate{
		DncCodes: &dncCodes,
		Division: buildSdkDomainEntityRef(d, "division_id"),
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
	if licenseId != "" {
		sdkDncListCreate.LicenseId = &licenseId
	}
	if dncSourceType != "" {
		sdkDncListCreate.DncSourceType = &dncSourceType
	}

	log.Printf("Creating Outbound DNC list %s", name)
	outboundDncList, _, err := outboundApi.PostOutboundDnclists(sdkDncListCreate)
	if err != nil {
		return diag.Errorf("Failed to create Outbound DNC list %s: %s", name, err)
	}

	d.SetId(*outboundDncList.Id)

	if len(entries) > 0 {
		if *sdkDncListCreate.DncSourceType == "rds" {
			for _, entry := range entries {
				_, err := uploadPhoneEntriesToDncList(outboundApi, outboundDncList, entry)
				if err != nil {
					return err
				}
			}
		} else {
			return diag.Errorf("Phone numbers can only be uploaded to internal DNC lists.")
		}
	}

	log.Printf("Created Outbound DNC list %s %s", name, *outboundDncList.Id)
	return readOutboundDncList(ctx, d, meta)
}

func updateOutboundDncList(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)
	contactMethod := d.Get("contact_method").(string)
	loginId := d.Get("login_id").(string)
	//campaignId := d.Get("campaign_id").(string)
	dncCodes := interfaceListToStrings(d.Get("dnc_codes").([]interface{}))
	licenseId := d.Get("license_id").(string)
	dncSourceType := d.Get("dnc_source_type").(string)
	entries := d.Get("entries").([]interface{})

	sdkConfig := meta.(*providerMeta).ClientConfig
	outboundApi := platformclientv2.NewOutboundApiWithConfig(sdkConfig)

	sdkDncList := platformclientv2.Dnclist{
		DncCodes: &dncCodes,
		Division: buildSdkDomainEntityRef(d, "division_id"),
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
	if licenseId != "" {
		sdkDncList.LicenseId = &licenseId
	}
	if dncSourceType != "" {
		sdkDncList.DncSourceType = &dncSourceType
	}
	log.Printf("Updating Outbound DNC list %s", name)
	diagErr := retryWhen(isVersionMismatch, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
		// Get current Outbound DNC list version
		outboundDncList, resp, getErr := outboundApi.GetOutboundDnclist(d.Id(), false, false)
		if getErr != nil {
			return resp, diag.Errorf("Failed to read Outbound DNC list %s: %s", d.Id(), getErr)
		}
		sdkDncList.Version = outboundDncList.Version
		outboundDncList, _, updateErr := outboundApi.PutOutboundDnclist(d.Id(), sdkDncList)
		if updateErr != nil {
			return resp, diag.Errorf("Failed to update Outbound DNC list %s: %s", name, updateErr)
		}
		if len(entries) > 0 {
			if *sdkDncList.DncSourceType == "rds" {
				for _, entry := range entries {
					response, err := uploadPhoneEntriesToDncList(outboundApi, outboundDncList, entry)
					if err != nil {
						return response, err
					}
				}
			} else {
				return nil, diag.Errorf("Phone numbers can only be uploaded to internal DNC lists.")
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
	sdkConfig := meta.(*providerMeta).ClientConfig
	outboundApi := platformclientv2.NewOutboundApiWithConfig(sdkConfig)

	log.Printf("Reading Outbound DNC list %s", d.Id())

	return withRetriesForRead(ctx, d, func() *resource.RetryError {
		sdkDncList, resp, getErr := outboundApi.GetOutboundDnclist(d.Id(), false, false)
		if getErr != nil {
			if isStatus404(resp) {
				return resource.RetryableError(fmt.Errorf("failed to read Outbound DNC list %s: %s", d.Id(), getErr))
			}
			return resource.NonRetryableError(fmt.Errorf("failed to read Outbound DNC list %s: %s", d.Id(), getErr))
		}

		cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, resourceOutboundDncList())

		if sdkDncList.Name != nil {
			_ = d.Set("name", *sdkDncList.Name)
		}
		if sdkDncList.ContactMethod != nil {
			_ = d.Set("contact_method", *sdkDncList.ContactMethod)
		}
		if sdkDncList.LoginId != nil {
			_ = d.Set("login_id", *sdkDncList.LoginId)
		}
		if sdkDncList.DncCodes != nil {
			var dncCodes []string
			schemaCodes := interfaceListToStrings(d.Get("dnc_codes").([]interface{}))
			if len(*sdkDncList.DncCodes) == len(schemaCodes) {
				// Rearrange to the original ordering
				for _, schemaCode := range schemaCodes {
					for _, sdkCode := range *sdkDncList.DncCodes {
						if schemaCode == sdkCode {
							dncCodes = append(dncCodes, sdkCode)
						}
					}
				}
				if len(dncCodes) != len(*sdkDncList.DncCodes) {
					dncCodes = nil
				}
			}
			if len(dncCodes) == 0 {
				for _, v := range *sdkDncList.DncCodes {
					dncCodes = append(dncCodes, v)
				}
			}
			_ = d.Set("dnc_codes", dncCodes)
		}
		if sdkDncList.LicenseId != nil {
			_ = d.Set("license_id", *sdkDncList.LicenseId)
		}
		if sdkDncList.Division != nil && sdkDncList.Division.Id != nil {
			_ = d.Set("division_id", *sdkDncList.Division.Id)
		}
		log.Printf("Read Outbound DNC list %s %s", d.Id(), *sdkDncList.Name)
		return cc.CheckState()
	})
}

func deleteOutboundDncList(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*providerMeta).ClientConfig
	outboundApi := platformclientv2.NewOutboundApiWithConfig(sdkConfig)

	diagErr := retryWhen(isStatus400, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
		log.Printf("Deleting Outbound DNC list")
		resp, err := outboundApi.DeleteOutboundDnclist(d.Id())
		if err != nil {
			return resp, diag.Errorf("Failed to delete Outbound DNC list: %s", err)
		}
		return resp, nil
	})
	if diagErr != nil {
		return diagErr
	}

	return withRetries(ctx, 30*time.Second, func() *resource.RetryError {
		_, resp, err := outboundApi.GetOutboundDnclist(d.Id(), false, false)
		if err != nil {
			if isStatus404(resp) {
				// Outbound DNC list deleted
				log.Printf("Deleted Outbound DNC list %s", d.Id())
				return nil
			}
			return resource.NonRetryableError(fmt.Errorf("error deleting Outbound DNC list %s: %s", d.Id(), err))
		}

		return resource.RetryableError(fmt.Errorf("Outbound DNC list %s still exists", d.Id()))
	})
}

func uploadPhoneEntriesToDncList(api *platformclientv2.OutboundApi, dncList *platformclientv2.Dnclist, entry interface{}) (*platformclientv2.APIResponse, diag.Diagnostics) {
	var phoneNumbers []string
	if entryMap, ok := entry.(map[string]interface{}); ok && len(entryMap) > 0 {
		if phoneNumbersList := entryMap["phone_numbers"].([]interface{}); phoneNumbersList != nil {
			for _, number := range phoneNumbersList {
				phoneNumbers = append(phoneNumbers, number.(string))
			}
		}
		log.Printf("Uploading phone numbers to DNC list %s", *dncList.Name)
		// POST /api/v2/outbound/dnclists/{dncListId}/phonenumbers
		response, err := api.PostOutboundDnclistPhonenumbers(*dncList.Id, phoneNumbers, entryMap["expiration_date"].(string))
		if err != nil {
			return response, diag.Errorf("Failed to upload phone numbers to Outbound DNC list %s: %s", *dncList.Name, err)
		}
		log.Printf("Uploaded phone numbers to DNC list %s", *dncList.Name)
	}
	return nil, nil
}
