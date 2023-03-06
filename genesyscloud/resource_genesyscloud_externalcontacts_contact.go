package genesyscloud

import (
	"context"
	"fmt"
	"log"
	"time"

	"terraform-provider-genesyscloud/genesyscloud/consistency_checker"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v92/platformclientv2"
)

var (
	phoneNumber = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"display": {
				Description: "Display string of the phone number.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"extension": {
				Description: "Phone extension.",
				Type:        schema.TypeInt,
				Optional:    true,
			},
			"accepts_sms": {
				Description: "If contact accept SMS.",
				Type:        schema.TypeBool,
				Optional:    true,
			},
			"e164": {
				Description: "Phone number in e164 format.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"country_code": {
				Description: "Phone number country code.",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
			},
		},
	}

	address = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"address1": {
				Description: "Contact address 1.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"address2": {
				Description: "Contact address 2.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"city": {
				Description: "Contact address city.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"state": {
				Description: "Contact address state.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"postal_code": {
				Description: "Contact address postal code.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"country_code": {
				Description: "Contact address country code.",
				Type:        schema.TypeString,
				Optional:    true,
			},
		},
	}

	twitterId = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"id": {
				Description: "Contact twitter id.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"name": {
				Description: "Contact twitter name.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"screen_name": {
				Description: "Contact twitter screen name.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"profile_url": {
				Description: "Contact twitter account url.",
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
	}

	lineIds = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"user_id": {
				Description: "Contact line id.",
				Type:        schema.TypeString,
				Optional:    true,
			},
		},
	}

	lineId = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"ids": {
				Description: "Contact line id.",
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        lineIds,
			},
			"display_name": {
				Description: "Contact line display name.",
				Type:        schema.TypeString,
				Optional:    true,
			},
		},
	}

	whatsappId = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"phone_number": {
				Description: "Contact whatsapp phone number.",
				Type:        schema.TypeList,
				Required:    true,
				Elem:        phoneNumber,
			},
			"display_name": {
				Description: "Contact whatsapp display name.",
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}

	facebookIds = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"scoped_id": {
				Description: "Contact facebook scoped id.",
				Type:        schema.TypeString,
				Optional:    true,
			},
		},
	}

	facebookId = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"ids": {
				Description: "Contact facebook scoped id.",
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        facebookIds,
			},
			"display_name": {
				Description: "Contact whatsapp display name.",
				Type:        schema.TypeString,
				Optional:    true,
			},
		},
	}
)

func getAllAuthExternalContacts(_ context.Context, clientConfig *platformclientv2.Configuration) (ResourceIDMetaMap, diag.Diagnostics) {
	resources := make(ResourceIDMetaMap)
	externalAPI := platformclientv2.NewExternalContactsApiWithConfig(clientConfig)

	for pageNum := 1; ; pageNum++ {
		const pageSize = 100
		externalContacts, _, getErr := externalAPI.GetExternalcontactsContacts(pageSize, pageNum, "", "", nil)
		if getErr != nil {
			return nil, diag.Errorf("Failed to get external contacts: %v", getErr)
		}

		if externalContacts.Entities == nil || len(*externalContacts.Entities) == 0 {
			break
		}

		for _, externalContact := range *externalContacts.Entities {
			log.Printf("Dealing with external concat id : %s", *externalContact.Id)
			resources[*externalContact.Id] = &ResourceMeta{Name: *externalContact.Id}
		}
	}

	return resources, nil
}

func externalContactExporter() *ResourceExporter {
	return &ResourceExporter{
		GetResourcesFunc: getAllWithPooledClient(getAllAuthExternalContacts),
		RefAttrs: map[string]*RefAttrSettings{
			"external_organization": {}, //Need to add this when we external orgs implemented
		},
	}
}

func resourceExternalContact() *schema.Resource {
	return &schema.Resource{
		Description: "Genesys Cloud External Contact",

		CreateContext: createWithPooledClient(createExternalContact),
		ReadContext:   readWithPooledClient(readExternalContact),
		UpdateContext: updateWithPooledClient(updateExternalContact),
		DeleteContext: deleteWithPooledClient(deleteExternalContact),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			"first_name": {
				Description: "The first name of the contact.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"middle_name": {
				Description: "The middle name of the contact.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"last_name": {
				Description: "The middle name of the contact.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"salutation": {
				Description: "The salutation of the contact.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"title": {
				Description: "The title of the contact.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"work_phone": {
				Description: "Contact work phone settings.",
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Computed:    true,
				Elem:        phoneNumber,
			},
			"cell_phone": {
				Description: "Contact call phone settings.",
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Elem:        phoneNumber,
			},
			"home_phone": {
				Description: "Contact home phone settings.",
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Elem:        phoneNumber,
			},
			"other_phone": {
				Description: "Contact other phone settings.",
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Elem:        phoneNumber,
			},
			"work_email": {
				Description: "Contact work email.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"personnal_email": {
				Description: "Contact personnal email.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"other_email": {
				Description: "Contact other email.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"address": {
				Description: "Contact address.",
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Elem:        address,
			},
			"twitter_id": {
				Description: "Contact twitter account informations.",
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Elem:        twitterId,
			},
			"line_id": {
				Description: "Contact line account informations.",
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Elem:        lineId,
			},
			"whatsapp_id": {
				Description: "Contact whatsapp account informations.",
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Computed:    true,
				Elem:        whatsappId,
			},
			"facebook_id": {
				Description: "Contact facebook account informations.",
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Elem:        facebookId,
			},
			"external_organization": {
				Description: "Contact survey opt out preference.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"survey_opt_out": {
				Description: "Contact survey opt out preference.",
				Type:        schema.TypeBool,
				Optional:    true,
			},
			"external_system_url": {
				Description: "Contact external system url.",
				Type:        schema.TypeString,
				Optional:    true,
			},
		},
	}
}

func createExternalContact(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*ProviderMeta).ClientConfig
	externalAPI := platformclientv2.NewExternalContactsApiWithConfig(sdkConfig)

	externalContact := getExternalContactFromResourceData(d)

	contact, _, err := externalAPI.PostExternalcontactsContacts(externalContact)
	if err != nil {
		return diag.Errorf("Failed to create external contact: %s", err)
	}

	d.SetId(*contact.Id)
	log.Printf("Created external contact %s", *contact.Id)
	return readExternalContact(ctx, d, meta)
}

func getExternalContactFromResourceData(d *schema.ResourceData) platformclientv2.Externalcontact {
	firstName := d.Get("first_name").(string)
	middleName := d.Get("middle_name").(string)
	lastName := d.Get("last_name").(string)
	salutation := d.Get("salutation").(string)
	title := d.Get("title").(string)
	workEmail := d.Get("work_email").(string)
	personnalEmail := d.Get("personnal_email").(string)
	otherEmail := d.Get("other_email").(string)
	surveyOptOut := d.Get("survey_opt_out").(bool)
	externalSystemUrl := d.Get("external_system_url").(string)

	return platformclientv2.Externalcontact{
		FirstName:         &firstName,
		MiddleName:        &middleName,
		LastName:          &lastName,
		Salutation:        &salutation,
		Title:             &title,
		WorkPhone:         buildSdkPhoneNumber(d, "work_phone"),
		CellPhone:         buildSdkPhoneNumber(d, "cell_phone"),
		HomePhone:         buildSdkPhoneNumber(d, "home_phone"),
		OtherPhone:        buildSdkPhoneNumber(d, "other_phone"),
		WorkEmail:         &workEmail,
		PersonalEmail:     &personnalEmail,
		OtherEmail:        &otherEmail,
		Address:           buildSdkAddresse(d, "address"),
		TwitterId:         buildSdkTwitterId(d, "twitter_id"),
		LineId:            buildSdkLineId(d, "line_id"),
		WhatsAppId:        buildSdkWhatsAppId(d, "whatsapp_id"),
		FacebookId:        buildSdkFacebookId(d, "facebook_id"),
		SurveyOptOut:      &surveyOptOut,
		ExternalSystemUrl: &externalSystemUrl,
	}
}

func buildPhonenumberFromData(phoneData []interface{}) *platformclientv2.Phonenumber {
	phoneMap := phoneData[0].(map[string]interface{})

	display := phoneMap["display"].(string)
	extension := phoneMap["extension"].(int)
	acceptSMS := phoneMap["accepts_sms"].(bool)
	e164 := phoneMap["e164"].(string)
	countryCode := phoneMap["country_code"].(string)

	return &platformclientv2.Phonenumber{
		Display:     &display,
		Extension:   &extension,
		AcceptsSMS:  &acceptSMS,
		E164:        &e164,
		CountryCode: &countryCode,
	}
}

func buildSdkPhoneNumber(d *schema.ResourceData, key string) *platformclientv2.Phonenumber {
	if d.Get(key) != nil {
		phoneData := d.Get(key).([]interface{})

		if len(phoneData) > 0 {
			return buildPhonenumberFromData(phoneData)
		}
	}
	return nil
}

func flattenPhoneNumber(phonenumber *platformclientv2.Phonenumber) []interface{} {
	phonenumberInterface := make(map[string]interface{})
	if phonenumber.Display != nil {
		phonenumberInterface["display"] = *phonenumber.Display
	}
	if phonenumber.Extension != nil {
		phonenumberInterface["extension"] = *phonenumber.Extension
	}
	if phonenumber.AcceptsSMS != nil {
		phonenumberInterface["accepts_sms"] = *phonenumber.AcceptsSMS
	}
	if phonenumber.E164 != nil {
		phonenumberInterface["e164"] = *phonenumber.E164
	}
	if phonenumber.CountryCode != nil {
		phonenumberInterface["country_code"] = *phonenumber.CountryCode
	}
	return []interface{}{phonenumberInterface}
}

func buildSdkAddresse(d *schema.ResourceData, key string) *platformclientv2.Contactaddress {
	if d.Get(key) != nil {
		addressData := d.Get(key).([]interface{})
		if len(addressData) > 0 {
			addressMap := addressData[0].(map[string]interface{})
			address1 := addressMap["address1"].(string)
			address2 := addressMap["address2"].(string)
			city := addressMap["city"].(string)
			state := addressMap["state"].(string)
			postalcode := addressMap["postal_code"].(string)
			countrycode := addressMap["country_code"].(string)

			return &platformclientv2.Contactaddress{
				Address1:    &address1,
				Address2:    &address2,
				City:        &city,
				State:       &state,
				PostalCode:  &postalcode,
				CountryCode: &countrycode,
			}
		}
	}
	return nil
}

func flattenSdkAddress(address platformclientv2.Contactaddress) []interface{} {
	addressInterface := make(map[string]interface{})
	if address.Address1 != nil {
		addressInterface["address1"] = address.Address1
	}
	if address.Address2 != nil {
		addressInterface["address2"] = address.Address2
	}
	if address.City != nil {
		addressInterface["city"] = address.City
	}
	if address.State != nil {
		addressInterface["state"] = address.State
	}
	if address.PostalCode != nil {
		addressInterface["postal_code"] = address.PostalCode
	}
	if address.CountryCode != nil {
		addressInterface["country_code"] = address.CountryCode
	}
	return []interface{}{addressInterface}
}

func buildSdkTwitterId(d *schema.ResourceData, key string) *platformclientv2.Twitterid {
	if d.Get(key) != nil {
		twitterData := d.Get(key).([]interface{})
		if len(twitterData) > 0 {
			twitterMap := twitterData[0].(map[string]interface{})
			id := twitterMap["id"].(string)
			name := twitterMap["name"].(string)
			screenname := twitterMap["screen_name"].(string)
			profileurl := twitterMap["profile_url"].(string)

			return &platformclientv2.Twitterid{
				Id:         &id,
				Name:       &name,
				ScreenName: &screenname,
				ProfileUrl: &profileurl,
			}
		}
	}
	return nil
}

func flattenSdkTwitterId(twitterId platformclientv2.Twitterid) []interface{} {
	twitterInterface := make(map[string]interface{})
	if twitterId.Id != nil {
		twitterInterface["id"] = twitterId.Id
	}
	if twitterId.Name != nil {
		twitterInterface["name"] = twitterId.Name
	}
	if twitterId.ScreenName != nil {
		url := "https://www.twitter.com/" + *twitterId.ScreenName
		twitterInterface["screen_name"] = twitterId.ScreenName
		twitterInterface["profile_url"] = &url
	}
	if twitterId.ProfileUrl != nil {
		twitterInterface["profile_url"] = twitterId.ProfileUrl
	}
	return []interface{}{twitterInterface}
}

func buildSdkLineId(d *schema.ResourceData, key string) *platformclientv2.Lineid {
	if d.Get(key) != nil {
		lineData := d.Get(key).([]interface{})
		if len(lineData) > 0 {
			lineMap := lineData[0].(map[string]interface{})
			displayname := lineMap["display_name"].(string)
			userId := lineMap["ids"].([]interface{})[0].(map[string]interface{})["user_id"].(string)

			ids := []platformclientv2.Lineuserid{
				{
					UserId: &userId,
				},
			}
			lineId := platformclientv2.Lineid{
				DisplayName: &displayname,
				Ids:         &ids,
			}
			return &lineId
		}
	}
	return nil
}

func flattenSdkLineId(lineId platformclientv2.Lineid) []interface{} {
	lineInterface := make(map[string]interface{})
	flattenUserid := flattenSdkLineUserId(lineId.Ids)
	lineInterface["display_name"] = *lineId.DisplayName
	lineInterface["ids"] = &flattenUserid
	return []interface{}{lineInterface}
}

func flattenSdkLineUserId(lineUserdid *[]platformclientv2.Lineuserid) []interface{} {
	lineUseridInterface := make(map[string]interface{})
	if (*lineUserdid)[0].UserId != nil {
		lineUseridInterface["user_id"] = (*lineUserdid)[0].UserId
	}
	return []interface{}{lineUseridInterface}
}

func buildSdkWhatsAppId(d *schema.ResourceData, key string) *platformclientv2.Whatsappid {
	if d.Get(key) != nil {
		whatsappData := d.Get(key).([]interface{})
		if len(whatsappData) > 0 {
			whatsappMap := whatsappData[0].(map[string]interface{})
			displayName := whatsappMap["display_name"].(string)

			return &platformclientv2.Whatsappid{
				DisplayName: &displayName,
				PhoneNumber: buildPhonenumberFromData(whatsappMap["phone_number"].([]interface{})),
			}
		}
	}
	return nil
}

func flattenSdkWhatsAppId(whatsappId platformclientv2.Whatsappid) []interface{} {
	whatsappInterface := make(map[string]interface{})
	flattenPhonenumber := flattenPhoneNumber(whatsappId.PhoneNumber)
	whatsappInterface["display_name"] = *whatsappId.DisplayName
	whatsappInterface["phone_number"] = &flattenPhonenumber
	return []interface{}{whatsappInterface}
}

func buildSdkFacebookId(d *schema.ResourceData, key string) *platformclientv2.Facebookid {
	if d.Get(key) != nil {
		facebookData := d.Get(key).([]interface{})
		if len(facebookData) > 0 {
			facebookMap := facebookData[0].(map[string]interface{})
			displayname := facebookMap["display_name"].(string)
			scopedId := facebookMap["ids"].([]interface{})[0].(map[string]interface{})["scoped_id"].(string)

			facebookIds := []platformclientv2.Facebookscopedid{
				{
					ScopedId: &scopedId,
				},
			}
			facebookId := platformclientv2.Facebookid{
				DisplayName: &displayname,
				Ids:         &facebookIds,
			}
			return &facebookId
		}
	}
	return nil
}

func flattenSdkFacebookId(facebookid platformclientv2.Facebookid) []interface{} {
	whatsappInterface := make(map[string]interface{})
	flattenScopedid := flattenSdkFacebookScopedId(facebookid.Ids)
	whatsappInterface["display_name"] = *facebookid.DisplayName
	whatsappInterface["ids"] = &flattenScopedid
	return []interface{}{whatsappInterface}
}

func flattenSdkFacebookScopedId(facebookScopedid *[]platformclientv2.Facebookscopedid) []interface{} {
	facebookScopedidInterface := make(map[string]interface{})
	if (*facebookScopedid)[0].ScopedId != nil {
		facebookScopedidInterface["scoped_id"] = (*facebookScopedid)[0].ScopedId
	}
	return []interface{}{facebookScopedidInterface}
}

func readExternalContact(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*ProviderMeta).ClientConfig
	externalAPI := platformclientv2.NewExternalContactsApiWithConfig(sdkConfig)

	log.Printf("Reading contact %s", d.Id())

	return withRetriesForRead(ctx, d, func() *resource.RetryError {
		externalContact, resp, getErr := externalAPI.GetExternalcontactsContact(d.Id(), nil)
		if getErr != nil {
			if isStatus404(resp) {
				return resource.RetryableError(fmt.Errorf("Failed to read external contact %s: %s", d.Id(), getErr))
			}
			return resource.NonRetryableError(fmt.Errorf("Failed to read external contact %s: %s", d.Id(), getErr))
		}

		cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, resourceExternalContact())

		if externalContact.FirstName != nil {
			d.Set("first_name", *externalContact.FirstName)
		} else {
			d.Set("first_name", nil)
		}

		if externalContact.MiddleName != nil {
			d.Set("middle_name", *externalContact.MiddleName)
		} else {
			d.Set("middle_name", nil)
		}

		if externalContact.LastName != nil {
			d.Set("last_name", *externalContact.LastName)
		} else {
			d.Set("last_name", nil)
		}

		if externalContact.Salutation != nil {
			d.Set("salutation", *externalContact.Salutation)
		} else {
			d.Set("salutation", nil)
		}

		if externalContact.Title != nil {
			d.Set("title", *externalContact.Title)
		} else {
			d.Set("title", nil)
		}

		if externalContact.WorkPhone != nil {
			d.Set("work_phone", flattenPhoneNumber(externalContact.WorkPhone))
		} else {
			d.Set("work_phone", nil)
		}

		if externalContact.CellPhone != nil {
			d.Set("cell_phone", flattenPhoneNumber(externalContact.CellPhone))
		} else {
			d.Set("cell_phone", nil)
		}

		if externalContact.HomePhone != nil {
			d.Set("home_phone", flattenPhoneNumber(externalContact.HomePhone))
		} else {
			d.Set("home_phone", nil)
		}

		if externalContact.OtherPhone != nil {
			d.Set("other_phone", flattenPhoneNumber(externalContact.OtherPhone))
		} else {
			d.Set("other_phone", nil)
		}

		if externalContact.WorkEmail != nil {
			d.Set("work_email", *externalContact.WorkEmail)
		} else {
			d.Set("work_email", nil)
		}

		if externalContact.PersonalEmail != nil {
			d.Set("personnal_email", *externalContact.PersonalEmail)
		} else {
			d.Set("personnal_email", nil)
		}

		if externalContact.OtherEmail != nil {
			d.Set("other_email", *externalContact.OtherEmail)
		} else {
			d.Set("other_email", nil)
		}

		if externalContact.Address != nil {
			d.Set("address", flattenSdkAddress(*externalContact.Address))
		} else {
			d.Set("address", nil)
		}

		if externalContact.TwitterId != nil {
			d.Set("twitter_id", flattenSdkTwitterId(*externalContact.TwitterId))
		} else {
			d.Set("twitter_id", nil)
		}

		if externalContact.LineId != nil {
			d.Set("line_id", flattenSdkLineId(*externalContact.LineId))
		} else {
			d.Set("line_id", nil)
		}

		if externalContact.WhatsAppId != nil {
			d.Set("whatsapp_id", flattenSdkWhatsAppId(*externalContact.WhatsAppId))
		} else {
			d.Set("whatsapp_id", nil)
		}

		if externalContact.FacebookId != nil {
			d.Set("facebook_id", flattenSdkFacebookId(*externalContact.FacebookId))
		} else {
			d.Set("facebook_id", nil)
		}

		if externalContact.SurveyOptOut != nil {
			d.Set("survey_opt_out", *externalContact.SurveyOptOut)
		} else {
			d.Set("survey_opt_out", nil)
		}

		if externalContact.ExternalSystemUrl != nil {
			d.Set("external_system_url", *externalContact.ExternalSystemUrl)
		} else {
			d.Set("external_system_url", nil)
		}

		log.Printf("Read external contact %s", d.Id())
		return cc.CheckState()
	})
}

func updateExternalContact(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*ProviderMeta).ClientConfig
	externalAPI := platformclientv2.NewExternalContactsApiWithConfig(sdkConfig)

	externalContact := getExternalContactFromResourceData(d)
	_, _, err := externalAPI.PutExternalcontactsContact(d.Id(), externalContact)
	if err != nil {
		return diag.Errorf("Failed to update external contact: %s", err)
	}

	log.Printf("Updated external contact")

	return readExternalContact(ctx, d, meta)
}

func deleteExternalContact(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*ProviderMeta).ClientConfig
	externalAPI := platformclientv2.NewExternalContactsApiWithConfig(sdkConfig)

	_, _, err := externalAPI.DeleteExternalcontactsContact(d.Id())
	if err != nil {
		return diag.Errorf("Failed to delete external contact %s: %s", d.Id(), err)
	}

	return withRetries(ctx, 180*time.Second, func() *resource.RetryError {
		_, resp, err := externalAPI.GetExternalcontactsContact(d.Id(), nil)

		if err == nil {
			return resource.NonRetryableError(fmt.Errorf("Error deleting external contact %s: %s", d.Id(), err))
		}
		if isStatus404(resp) {
			// Success  : External contact deleted
			log.Printf("Deleted external contact %s", d.Id())
			return nil
		}

		return resource.RetryableError(fmt.Errorf("External contact %s still exists", d.Id()))
	})
}
