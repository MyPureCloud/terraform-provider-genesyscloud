package external_contacts_organization

import (
	"terraform-provider-genesyscloud/genesyscloud/util"
	"terraform-provider-genesyscloud/genesyscloud/util/resourcedata"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v115/platformclientv2"
	"github.com/nyaruka/phonenumbers"
)

/*
The resource_genesyscloud_external_contacts_organization_utils.go file contains various helper methods to marshal
and unmarshal data into formats consumable by Terraform and/or Genesys Cloud.
*/

// getExternalContactsOrganizationFromResourceData maps data from schema ResourceData object to a platformclientv2.Externalorganization
func getExternalContactsOrganizationFromResourceData(d *schema.ResourceData) platformclientv2.Externalorganization {
	return platformclientv2.Externalorganization{
		Name:             platformclientv2.String(d.Get("name").(string)),
		CompanyType:      platformclientv2.String(d.Get("company_type").(string)),
		Industry:         platformclientv2.String(d.Get("industry").(string)),
		PrimaryContactId: platformclientv2.String(d.Get("primary_contact_id").(string)),
		Address:          buildContactAddress(d.Get("address").([]interface{})),
		PhoneNumber:      buildPhoneNumber(d.Get("phone_number").([]interface{})),
		FaxNumber:        buildPhoneNumber(d.Get("fax_number").([]interface{})),
		EmployeeCount:    platformclientv2.Int(d.Get("employee_count").(int)),
		Revenue:          platformclientv2.Int(d.Get("revenue").(int)),
		// TODO: Handle tags property
		// TODO: Handle websites property
		Tickers:           buildTickers(d.Get("tickers").([]interface{})),
		TwitterId:         buildTwitterId(d.Get("twitter_id").([]interface{})),
		ExternalSystemUrl: platformclientv2.String(d.Get("external_system_url").(string)),
		Trustor:           buildTrustor(d.Get("trustor").([]interface{})),
		Schema:            buildDataSchema(d.Get("schema").([]interface{})),
		// TODO: Handle custom_fields property
		ExternalDataSources: buildExternalDataSources(d.Get("external_data_sources").([]interface{})),
	}
}

// buildPhonenumberFromData is a helper method to map phone data to the GenesysCloud platformclientv2.PhoneNumber
func buildPhonenumberFromData(phoneData []interface{}) *platformclientv2.Phonenumber {
	phoneMap := phoneData[0].(map[string]interface{})

	display := phoneMap["display"].(string)
	acceptSMS := phoneMap["accepts_sms"].(bool)
	e164 := phoneMap["e164"].(string)
	countryCode := phoneMap["country_code"].(string)
	phoneNumber := &platformclientv2.Phonenumber{
		Display:     &display,
		AcceptsSMS:  &acceptSMS,
		E164:        &e164,
		CountryCode: &countryCode,
	}
	extension := phoneMap["extension"].(int)
	if extension != 0 {
		phoneNumber.Extension = &extension
	}
	return phoneNumber

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
	if phonenumber == nil {
		return nil
	}

	phonenumberInterface := make(map[string]interface{})
	resourcedata.SetMapValueIfNotNil(phonenumberInterface, "display", phonenumber.Display)
	resourcedata.SetMapValueIfNotNil(phonenumberInterface, "extension", phonenumber.Extension)
	resourcedata.SetMapValueIfNotNil(phonenumberInterface, "accepts_sms", phonenumber.AcceptsSMS)
	if phonenumber.E164 != nil && *phonenumber.E164 != "" {
		var phoneNumberE164 string
		utilE164 := util.NewUtilE164Service()
		phoneNumberE164 = utilE164.FormatAsCalculatedE164Number(*phonenumber.E164)
		phonenumberInterface["e164"] = phoneNumberE164
	}
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

// buildTickers maps an []interface{} into a Genesys Cloud *[]platformclientv2.Ticker
func buildTickers(tickers []interface{}) *[]platformclientv2.Ticker {
	tickersSlice := make([]platformclientv2.Ticker, 0)
	for _, ticker := range tickers {
		var sdkTicker platformclientv2.Ticker
		tickersMap, ok := ticker.(map[string]interface{})
		if !ok {
			continue
		}

		resourcedata.BuildSDKStringValueIfNotNil(&sdkTicker.Symbol, tickersMap, "symbol")
		resourcedata.BuildSDKStringValueIfNotNil(&sdkTicker.Exchange, tickersMap, "exchange")

		tickersSlice = append(tickersSlice, sdkTicker)
	}

	return &tickersSlice
}

// buildOrganizations maps an []interface{} into a Genesys Cloud *[]platformclientv2.Organization
func buildOrganizations(organizations []interface{}) *[]platformclientv2.Organization {
	organizationsSlice := make([]platformclientv2.Organization, 0)
	for _, organization := range organizations {
		var sdkOrganization platformclientv2.Organization
		organizationsMap, ok := organization.(map[string]interface{})
		if !ok {
			continue
		}

		resourcedata.BuildSDKStringValueIfNotNil(&sdkOrganization.Name, organizationsMap, "name")
		resourcedata.BuildSDKStringValueIfNotNil(&sdkOrganization.DefaultLanguage, organizationsMap, "default_language")
		resourcedata.BuildSDKStringValueIfNotNil(&sdkOrganization.DefaultCountryCode, organizationsMap, "default_country_code")
		resourcedata.BuildSDKStringValueIfNotNil(&sdkOrganization.ThirdPartyOrgName, organizationsMap, "third_party_org_name")
		resourcedata.BuildSDKStringValueIfNotNil(&sdkOrganization.ThirdPartyURI, organizationsMap, "third_party_u_r_i")
		resourcedata.BuildSDKStringValueIfNotNil(&sdkOrganization.Domain, organizationsMap, "domain")
		resourcedata.BuildSDKStringValueIfNotNil(&sdkOrganization.State, organizationsMap, "state")
		resourcedata.BuildSDKStringValueIfNotNil(&sdkOrganization.DefaultSiteId, organizationsMap, "default_site_id")
		resourcedata.BuildSDKStringValueIfNotNil(&sdkOrganization.SupportURI, organizationsMap, "support_u_r_i")
		sdkOrganization.VoicemailEnabled = platformclientv2.Bool(organizationsMap["voicemail_enabled"].(bool))
		resourcedata.BuildSDKStringValueIfNotNil(&sdkOrganization.ProductPlatform, organizationsMap, "product_platform")
		// TODO: Handle features property

		organizationsSlice = append(organizationsSlice, sdkOrganization)
	}

	return &organizationsSlice
}

// buildTrusteeAuthorizations maps an []interface{} into a Genesys Cloud *[]platformclientv2.Trusteeauthorization
func buildTrusteeAuthorizations(trusteeAuthorizations []interface{}) *[]platformclientv2.Trusteeauthorization {
	trusteeAuthorizationsSlice := make([]platformclientv2.Trusteeauthorization, 0)
	for _, trusteeAuthorization := range trusteeAuthorizations {
		var sdkTrusteeAuthorization platformclientv2.Trusteeauthorization
		trusteeAuthorizationsMap, ok := trusteeAuthorization.(map[string]interface{})
		if !ok {
			continue
		}

		resourcedata.BuildSDKStringArrayValueIfNotNil(&sdkTrusteeAuthorization.Permissions, trusteeAuthorizationsMap, "permissions")

		trusteeAuthorizationsSlice = append(trusteeAuthorizationsSlice, sdkTrusteeAuthorization)
	}

	return &trusteeAuthorizationsSlice
}

// buildTrustors maps an []interface{} into a Genesys Cloud *[]platformclientv2.Trustor
func buildTrustors(trustors []interface{}) *[]platformclientv2.Trustor {
	trustorsSlice := make([]platformclientv2.Trustor, 0)
	for _, trustor := range trustors {
		var sdkTrustor platformclientv2.Trustor
		trustorsMap, ok := trustor.(map[string]interface{})
		if !ok {
			continue
		}

		sdkTrustor.Enabled = platformclientv2.Bool(trustorsMap["enabled"].(bool))
		resourcedata.BuildSDKInterfaceArrayValueIfNotNil(&sdkTrustor.Organization, trustorsMap, "organization", buildOrganization)
		resourcedata.BuildSDKInterfaceArrayValueIfNotNil(&sdkTrustor.Authorization, trustorsMap, "authorization", buildTrusteeAuthorization)

		trustorsSlice = append(trustorsSlice, sdkTrustor)
	}

	return &trustorsSlice
}

// buildJsonSchemaDocuments maps an []interface{} into a Genesys Cloud *[]platformclientv2.Jsonschemadocument
func buildJsonSchemaDocuments(jsonSchemaDocuments []interface{}) *[]platformclientv2.Jsonschemadocument {
	jsonSchemaDocumentsSlice := make([]platformclientv2.Jsonschemadocument, 0)
	for _, jsonSchemaDocument := range jsonSchemaDocuments {
		var sdkJsonSchemaDocument platformclientv2.Jsonschemadocument
		jsonSchemaDocumentsMap, ok := jsonSchemaDocument.(map[string]interface{})
		if !ok {
			continue
		}

		// resourcedata.BuildSDKStringValueIfNotNil(&sdkJsonSchemaDocument.$schema, jsonSchemaDocumentsMap, "$schema")
		resourcedata.BuildSDKStringValueIfNotNil(&sdkJsonSchemaDocument.Title, jsonSchemaDocumentsMap, "title")
		resourcedata.BuildSDKStringValueIfNotNil(&sdkJsonSchemaDocument.Description, jsonSchemaDocumentsMap, "description")
		resourcedata.BuildSDKStringValueIfNotNil(&sdkJsonSchemaDocument.Type, jsonSchemaDocumentsMap, "type")
		resourcedata.BuildSDKStringArrayValueIfNotNil(&sdkJsonSchemaDocument.Required, jsonSchemaDocumentsMap, "required")
		// TODO: Handle properties property
		// TODO: Handle additional_properties property

		jsonSchemaDocumentsSlice = append(jsonSchemaDocumentsSlice, sdkJsonSchemaDocument)
	}

	return &jsonSchemaDocumentsSlice
}

// buildDataSchemas maps an []interface{} into a Genesys Cloud *[]platformclientv2.Dataschema
func buildDataSchemas(dataSchemas []interface{}) *[]platformclientv2.Dataschema {
	dataSchemasSlice := make([]platformclientv2.Dataschema, 0)
	for _, dataSchema := range dataSchemas {
		var sdkDataSchema platformclientv2.Dataschema
		dataSchemasMap, ok := dataSchema.(map[string]interface{})
		if !ok {
			continue
		}

		resourcedata.BuildSDKStringValueIfNotNil(&sdkDataSchema.Name, dataSchemasMap, "name")
		resourcedata.BuildSDKStringArrayValueIfNotNil(&sdkDataSchema.AppliesTo, dataSchemasMap, "applies_to")
		sdkDataSchema.Enabled = platformclientv2.Bool(dataSchemasMap["enabled"].(bool))
		resourcedata.BuildSDKInterfaceArrayValueIfNotNil(&sdkDataSchema.JsonSchema, dataSchemasMap, "json_schema", buildJsonSchemaDocument)

		dataSchemasSlice = append(dataSchemasSlice, sdkDataSchema)
	}

	return &dataSchemasSlice
}

// buildExternalDataSources maps an []interface{} into a Genesys Cloud *[]platformclientv2.Externaldatasource
func buildExternalDataSources(externalDataSources []interface{}) *[]platformclientv2.Externaldatasource {
	externalDataSourcesSlice := make([]platformclientv2.Externaldatasource, 0)
	for _, externalDataSource := range externalDataSources {
		var sdkExternalDataSource platformclientv2.Externaldatasource
		externalDataSourcesMap, ok := externalDataSource.(map[string]interface{})
		if !ok {
			continue
		}

		resourcedata.BuildSDKStringValueIfNotNil(&sdkExternalDataSource.Platform, externalDataSourcesMap, "platform")
		resourcedata.BuildSDKStringValueIfNotNil(&sdkExternalDataSource.Url, externalDataSourcesMap, "url")

		externalDataSourcesSlice = append(externalDataSourcesSlice, sdkExternalDataSource)
	}

	return &externalDataSourcesSlice
}

// flattenContactAddresss maps a Genesys Cloud *[]platformclientv2.Contactaddress into a []interface{}
func flattenContactAddresss(contactAddresss *[]platformclientv2.Contactaddress) []interface{} {
	if len(*contactAddresss) == 0 {
		return nil
	}

	var contactAddressList []interface{}
	for _, contactAddress := range *contactAddresss {
		contactAddressMap := make(map[string]interface{})

		resourcedata.SetMapValueIfNotNil(contactAddressMap, "address1", contactAddress.Address1)
		resourcedata.SetMapValueIfNotNil(contactAddressMap, "address2", contactAddress.Address2)
		resourcedata.SetMapValueIfNotNil(contactAddressMap, "city", contactAddress.City)
		resourcedata.SetMapValueIfNotNil(contactAddressMap, "state", contactAddress.State)
		resourcedata.SetMapValueIfNotNil(contactAddressMap, "postal_code", contactAddress.PostalCode)
		resourcedata.SetMapValueIfNotNil(contactAddressMap, "country_code", contactAddress.CountryCode)

		contactAddressList = append(contactAddressList, contactAddressMap)
	}

	return contactAddressList
}

// flattenPhoneNumbers maps a Genesys Cloud *[]platformclientv2.Phonenumber into a []interface{}
func flattenPhoneNumbers(phoneNumbers *[]platformclientv2.Phonenumber) []interface{} {
	if len(*phoneNumbers) == 0 {
		return nil
	}

	var phoneNumberList []interface{}
	for _, phoneNumber := range *phoneNumbers {
		phoneNumberMap := make(map[string]interface{})

		resourcedata.SetMapValueIfNotNil(phoneNumberMap, "display", phoneNumber.Display)
		resourcedata.SetMapValueIfNotNil(phoneNumberMap, "extension", phoneNumber.Extension)
		resourcedata.SetMapValueIfNotNil(phoneNumberMap, "accepts_s_m_s", phoneNumber.AcceptsSMS)
		resourcedata.SetMapValueIfNotNil(phoneNumberMap, "normalization_country_code", phoneNumber.NormalizationCountryCode)
		resourcedata.SetMapValueIfNotNil(phoneNumberMap, "user_input", phoneNumber.UserInput)
		resourcedata.SetMapValueIfNotNil(phoneNumberMap, "e164", phoneNumber.E164)
		resourcedata.SetMapValueIfNotNil(phoneNumberMap, "country_code", phoneNumber.CountryCode)

		phoneNumberList = append(phoneNumberList, phoneNumberMap)
	}

	return phoneNumberList
}

// flattenTickers maps a Genesys Cloud *[]platformclientv2.Ticker into a []interface{}
func flattenTickers(tickers *[]platformclientv2.Ticker) []interface{} {
	if len(*tickers) == 0 {
		return nil
	}

	var tickerList []interface{}
	for _, ticker := range *tickers {
		tickerMap := make(map[string]interface{})

		resourcedata.SetMapValueIfNotNil(tickerMap, "symbol", ticker.Symbol)
		resourcedata.SetMapValueIfNotNil(tickerMap, "exchange", ticker.Exchange)

		tickerList = append(tickerList, tickerMap)
	}

	return tickerList
}

// flattenTwitterIds maps a Genesys Cloud *[]platformclientv2.Twitterid into a []interface{}
func flattenTwitterIds(twitterIds *[]platformclientv2.Twitterid) []interface{} {
	if len(*twitterIds) == 0 {
		return nil
	}

	var twitterIdList []interface{}
	for _, twitterId := range *twitterIds {
		twitterIdMap := make(map[string]interface{})

		resourcedata.SetMapValueIfNotNil(twitterIdMap, "name", twitterId.Name)
		resourcedata.SetMapValueIfNotNil(twitterIdMap, "screen_name", twitterId.ScreenName)
		resourcedata.SetMapValueIfNotNil(twitterIdMap, "verified", twitterId.Verified)
		resourcedata.SetMapValueIfNotNil(twitterIdMap, "profile_url", twitterId.ProfileUrl)

		twitterIdList = append(twitterIdList, twitterIdMap)
	}

	return twitterIdList
}

// flattenOrganizations maps a Genesys Cloud *[]platformclientv2.Organization into a []interface{}
func flattenOrganizations(organizations *[]platformclientv2.Organization) []interface{} {
	if len(*organizations) == 0 {
		return nil
	}

	var organizationList []interface{}
	for _, organization := range *organizations {
		organizationMap := make(map[string]interface{})

		resourcedata.SetMapValueIfNotNil(organizationMap, "name", organization.Name)
		resourcedata.SetMapValueIfNotNil(organizationMap, "default_language", organization.DefaultLanguage)
		resourcedata.SetMapValueIfNotNil(organizationMap, "default_country_code", organization.DefaultCountryCode)
		resourcedata.SetMapValueIfNotNil(organizationMap, "third_party_org_name", organization.ThirdPartyOrgName)
		resourcedata.SetMapValueIfNotNil(organizationMap, "third_party_u_r_i", organization.ThirdPartyURI)
		resourcedata.SetMapValueIfNotNil(organizationMap, "domain", organization.Domain)
		resourcedata.SetMapValueIfNotNil(organizationMap, "state", organization.State)
		resourcedata.SetMapValueIfNotNil(organizationMap, "default_site_id", organization.DefaultSiteId)
		resourcedata.SetMapValueIfNotNil(organizationMap, "support_u_r_i", organization.SupportURI)
		resourcedata.SetMapValueIfNotNil(organizationMap, "voicemail_enabled", organization.VoicemailEnabled)
		resourcedata.SetMapValueIfNotNil(organizationMap, "product_platform", organization.ProductPlatform)
		// TODO: Handle features property

		organizationList = append(organizationList, organizationMap)
	}

	return organizationList
}

// flattenTrusteeAuthorizations maps a Genesys Cloud *[]platformclientv2.Trusteeauthorization into a []interface{}
func flattenTrusteeAuthorizations(trusteeAuthorizations *[]platformclientv2.Trusteeauthorization) []interface{} {
	if len(*trusteeAuthorizations) == 0 {
		return nil
	}

	var trusteeAuthorizationList []interface{}
	for _, trusteeAuthorization := range *trusteeAuthorizations {
		trusteeAuthorizationMap := make(map[string]interface{})

		resourcedata.SetMapStringArrayValueIfNotNil(trusteeAuthorizationMap, "permissions", trusteeAuthorization.Permissions)

		trusteeAuthorizationList = append(trusteeAuthorizationList, trusteeAuthorizationMap)
	}

	return trusteeAuthorizationList
}

// flattenTrustors maps a Genesys Cloud *[]platformclientv2.Trustor into a []interface{}
func flattenTrustors(trustors *[]platformclientv2.Trustor) []interface{} {
	if len(*trustors) == 0 {
		return nil
	}

	var trustorList []interface{}
	for _, trustor := range *trustors {
		trustorMap := make(map[string]interface{})

		resourcedata.SetMapValueIfNotNil(trustorMap, "enabled", trustor.Enabled)
		resourcedata.SetMapInterfaceArrayWithFuncIfNotNil(trustorMap, "organization", trustor.Organization, flattenOrganization)
		resourcedata.SetMapInterfaceArrayWithFuncIfNotNil(trustorMap, "authorization", trustor.Authorization, flattenTrusteeAuthorization)

		trustorList = append(trustorList, trustorMap)
	}

	return trustorList
}

// flattenJsonSchemaDocuments maps a Genesys Cloud *[]platformclientv2.Jsonschemadocument into a []interface{}
func flattenJsonSchemaDocuments(jsonSchemaDocuments *[]platformclientv2.Jsonschemadocument) []interface{} {
	if len(*jsonSchemaDocuments) == 0 {
		return nil
	}

	var jsonSchemaDocumentList []interface{}
	for _, jsonSchemaDocument := range *jsonSchemaDocuments {
		jsonSchemaDocumentMap := make(map[string]interface{})

		// resourcedata.SetMapValueIfNotNil(jsonSchemaDocumentMap, "$schema", jsonSchemaDocument.$schema)
		resourcedata.SetMapValueIfNotNil(jsonSchemaDocumentMap, "title", jsonSchemaDocument.Title)
		resourcedata.SetMapValueIfNotNil(jsonSchemaDocumentMap, "description", jsonSchemaDocument.Description)
		resourcedata.SetMapValueIfNotNil(jsonSchemaDocumentMap, "type", jsonSchemaDocument.Type)
		resourcedata.SetMapStringArrayValueIfNotNil(jsonSchemaDocumentMap, "required", jsonSchemaDocument.Required)
		// TODO: Handle properties property
		// TODO: Handle additional_properties property

		jsonSchemaDocumentList = append(jsonSchemaDocumentList, jsonSchemaDocumentMap)
	}

	return jsonSchemaDocumentList
}

// flattenDataSchemas maps a Genesys Cloud *[]platformclientv2.Dataschema into a []interface{}
func flattenDataSchemas(dataSchemas *[]platformclientv2.Dataschema) []interface{} {
	if len(*dataSchemas) == 0 {
		return nil
	}

	var dataSchemaList []interface{}
	for _, dataSchema := range *dataSchemas {
		dataSchemaMap := make(map[string]interface{})

		resourcedata.SetMapValueIfNotNil(dataSchemaMap, "name", dataSchema.Name)
		resourcedata.SetMapStringArrayValueIfNotNil(dataSchemaMap, "applies_to", dataSchema.AppliesTo)
		resourcedata.SetMapValueIfNotNil(dataSchemaMap, "enabled", dataSchema.Enabled)
		resourcedata.SetMapInterfaceArrayWithFuncIfNotNil(dataSchemaMap, "json_schema", dataSchema.JsonSchema, flattenJsonSchemaDocument)

		dataSchemaList = append(dataSchemaList, dataSchemaMap)
	}

	return dataSchemaList
}

// flattenExternalDataSources maps a Genesys Cloud *[]platformclientv2.Externaldatasource into a []interface{}
func flattenExternalDataSources(externalDataSources *[]platformclientv2.Externaldatasource) []interface{} {
	if len(*externalDataSources) == 0 {
		return nil
	}

	var externalDataSourceList []interface{}
	for _, externalDataSource := range *externalDataSources {
		externalDataSourceMap := make(map[string]interface{})

		resourcedata.SetMapValueIfNotNil(externalDataSourceMap, "platform", externalDataSource.Platform)
		resourcedata.SetMapValueIfNotNil(externalDataSourceMap, "url", externalDataSource.Url)

		externalDataSourceList = append(externalDataSourceList, externalDataSourceMap)
	}

	return externalDataSourceList
}

func hashFormattedPhoneNumber(val string) int {
	formattedNumber := ""

	number, err := phonenumbers.Parse(val, "US")
	if err == nil {
		formattedNumber = phonenumbers.Format(number, phonenumbers.E164)
	}

	return schema.HashString(formattedNumber)
}
