package external_contacts

import (
	"fmt"
	"terraform-provider-genesyscloud/genesyscloud/util/resourcedata"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"
	"github.com/nyaruka/phonenumbers"
)

/*
The resource_genesyscloud_externalcontacts_contacts_utils.go file contains various helper methods to marshal
and unmarshal data into formats consumable by Terraform and/or Genesys Cloud.

Note:  Look for opportunities to minimize boilerplate code using functions and Generics
*/

// getExternalContactFromResourceData maps data from schema ResourceData object to a platformclientv2.Externalcontact
func getExternalContactFromResourceData(d *schema.ResourceData) platformclientv2.Externalcontact {
	firstName := d.Get("first_name").(string)
	middleName := d.Get("middle_name").(string)
	lastName := d.Get("last_name").(string)
	salutation := d.Get("salutation").(string)
	title := d.Get("title").(string)
	workEmail := d.Get("work_email").(string)
	personalEmail := d.Get("personal_email").(string)
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
		PersonalEmail:     &personalEmail,
		OtherEmail:        &otherEmail,
		Address:           buildSdkAddress(d, "address"),
		TwitterId:         buildSdkTwitterId(d, "twitter_id"),
		LineId:            buildSdkLineId(d, "line_id"),
		WhatsAppId:        buildSdkWhatsAppId(d, "whatsapp_id"),
		FacebookId:        buildSdkFacebookId(d, "facebook_id"),
		SurveyOptOut:      &surveyOptOut,
		ExternalSystemUrl: &externalSystemUrl,
	}
}

// buildPhonenumberFromData is a helper method to map phone data to the GenesysCloud platformclientv2.PhoneNumber
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

// buildSdkPhoneNumber is a helper method to build a Genesys Cloud SDK PhoneNumber
func buildSdkPhoneNumber(d *schema.ResourceData, key string) *platformclientv2.Phonenumber {
	if d.Get(key) != nil {
		phoneData := d.Get(key).([]interface{})

		if len(phoneData) > 0 {
			return buildPhonenumberFromData(phoneData)
		}
	}
	return nil
}

// flattenPhoneNumber converts a platformclientv2.Phonenumber into a map and then into array for consumption by Terraform
func flattenPhoneNumber(phonenumber *platformclientv2.Phonenumber) []interface{} {
	phonenumberInterface := make(map[string]interface{})
	resourcedata.SetMapValueIfNotNil(phonenumberInterface, "display", phonenumber.Display)
	resourcedata.SetMapValueIfNotNil(phonenumberInterface, "extension", phonenumber.Extension)
	resourcedata.SetMapValueIfNotNil(phonenumberInterface, "accepts_sms", phonenumber.AcceptsSMS)
	resourcedata.SetMapValueIfNotNil(phonenumberInterface, "e164", phonenumber.E164)
	resourcedata.SetMapValueIfNotNil(phonenumberInterface, "country_code", phonenumber.CountryCode)
	return []interface{}{phonenumberInterface}
}

// buildSdkAddress constructs a platformclientv2.Contactaddress structure
func buildSdkAddress(d *schema.ResourceData, key string) *platformclientv2.Contactaddress {
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

// flattenflattenSdkAddress converts a *platformclientv2.Contactaddress into a map and then into array for consumption by Terraform
func flattenSdkAddress(address *platformclientv2.Contactaddress) []interface{} {
	addressInterface := make(map[string]interface{})

	resourcedata.SetMapValueIfNotNil(addressInterface, "address1", address.Address1)
	resourcedata.SetMapValueIfNotNil(addressInterface, "address2", address.Address2)
	resourcedata.SetMapValueIfNotNil(addressInterface, "city", address.City)
	resourcedata.SetMapValueIfNotNil(addressInterface, "state", address.State)
	resourcedata.SetMapValueIfNotNil(addressInterface, "postal_code", address.PostalCode)
	resourcedata.SetMapValueIfNotNil(addressInterface, "country_code", address.CountryCode)

	return []interface{}{addressInterface}
}

// buildSdkTwitterid maps data from a Terraform data object into a Genesys Cloud *platformclientv2.Twitterid
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

// flattenSdkTwitterId maps a Genesys Cloud platformclientv2.Twitterid into a []interface{}
func flattenSdkTwitterId(twitterId *platformclientv2.Twitterid) []interface{} {
	twitterInterface := make(map[string]interface{})
	resourcedata.SetMapValueIfNotNil(twitterInterface, "id", twitterId.Id)
	resourcedata.SetMapValueIfNotNil(twitterInterface, "name", twitterId.Name)
	if twitterId.ScreenName != nil {
		url := "https://www.twitter.com/" + *twitterId.ScreenName
		twitterInterface["screen_name"] = twitterId.ScreenName
		twitterInterface["profile_url"] = &url
	}

	resourcedata.SetMapValueIfNotNil(twitterInterface, "profile_url", twitterId.ProfileUrl)

	return []interface{}{twitterInterface}
}

// buildSdkLineId builds platformclientv2.Lineid struct from a Terramform resource data struct
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

// flattenSdkLineId maps platformclientv2.Lineid to a []interace{}
func flattenSdkLineId(lineId *platformclientv2.Lineid) []interface{} {
	lineInterface := make(map[string]interface{})
	flattenUserid := flattenSdkLineUserId(lineId.Ids)
	lineInterface["display_name"] = *lineId.DisplayName
	lineInterface["ids"] = &flattenUserid
	return []interface{}{lineInterface}
}

// flattenSdkLineUserId maps an []platformclientv2.Lineuserid to a []interface{}
func flattenSdkLineUserId(lineUserdid *[]platformclientv2.Lineuserid) []interface{} {
	lineUseridInterface := make(map[string]interface{})
	if (*lineUserdid)[0].UserId != nil {
		lineUseridInterface["user_id"] = (*lineUserdid)[0].UserId
	}
	return []interface{}{lineUseridInterface}
}

// buildSdkWhatsAppId maps a Terraform schema.ResourceData to a Genesys Cloud platformclientv2.Whatsappid
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

// flattenSdkWhatsAppId maps a Genesys Cloud platformclientv2.Whatsappid to a []interface{}
func flattenSdkWhatsAppId(whatsappId *platformclientv2.Whatsappid) []interface{} {
	whatsappInterface := make(map[string]interface{})
	flattenPhonenumber := flattenPhoneNumber(whatsappId.PhoneNumber)
	whatsappInterface["display_name"] = *whatsappId.DisplayName
	whatsappInterface["phone_number"] = &flattenPhonenumber
	return []interface{}{whatsappInterface}
}

// buildSdkFacebookId maps a Terraform schema.ResourceData struct to a Genesys Cloud platformclientv2.Facebookid
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

// flattenSdkFacebookId maps a Genesys Cloud platformclientv2.Facebookid object to a []interface{}
func flattenSdkFacebookId(facebookid *platformclientv2.Facebookid) []interface{} {
	whatsappInterface := make(map[string]interface{})
	flattenScopedid := flattenSdkFacebookScopedId(facebookid.Ids)
	whatsappInterface["display_name"] = *facebookid.DisplayName
	whatsappInterface["ids"] = &flattenScopedid
	return []interface{}{whatsappInterface}
}

// flattenSdkFacebookScopedId maps a Genesys Cloud platformclientv2.Facebookscopedid struct ot a []interface{}
func flattenSdkFacebookScopedId(facebookScopedid *[]platformclientv2.Facebookscopedid) []interface{} {
	facebookScopedidInterface := make(map[string]interface{})
	if (*facebookScopedid)[0].ScopedId != nil {
		facebookScopedidInterface["scoped_id"] = (*facebookScopedid)[0].ScopedId
	}
	return []interface{}{facebookScopedidInterface}
}

// formatPhoneNumber formats a given string to E164 format and hashes it for comparison
func hashFormattedPhoneNumber(val string) int {
	formattedNumber := ""

	number, err := phonenumbers.Parse(val, "US")
	if err == nil {
		formattedNumber = phonenumbers.Format(number, phonenumbers.E164)
	}

	return schema.HashString(formattedNumber)
}

func GenerateBasicExternalContactResource(resourceID string, title string) string {
	return fmt.Sprintf(`resource "genesyscloud_externalcontacts_contact" "%s" {
		title = "%s"
	}
	`, resourceID, title)
}
