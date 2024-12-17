package external_contacts_organization

import (
	"encoding/json"
	"fmt"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"terraform-provider-genesyscloud/genesyscloud/util/lists"
	"terraform-provider-genesyscloud/genesyscloud/util/resourcedata"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v149/platformclientv2"
	"github.com/nyaruka/phonenumbers"
)

/*
The resource_genesyscloud_external_contacts_organization_utils.go file contains various helper methods to marshal
and unmarshal data into formats consumable by Terraform and/or Genesys Cloud.
*/

// getExternalContactsOrganizationFromResourceData maps data from schema ResourceData object to a platformclientv2.Externalorganization
func getExternalContactsOrganizationFromResourceData(d *schema.ResourceData) (platformclientv2.Externalorganization, error) {
	externalOrganization := platformclientv2.Externalorganization{
		Name:                platformclientv2.String(d.Get("name").(string)),
		CompanyType:         platformclientv2.String(d.Get("company_type").(string)),
		Industry:            platformclientv2.String(d.Get("industry").(string)),
		Address:             buildSdkAddress(d, "address"),
		PhoneNumber:         buildSdkPhoneNumber(d, "phone_number"),
		FaxNumber:           buildSdkPhoneNumber(d, "fax_number"),
		EmployeeCount:       platformclientv2.Int(d.Get("employee_count").(int)),
		Revenue:             platformclientv2.Int(d.Get("revenue").(int)),
		Tickers:             buildTickers(d.Get("tickers").([]interface{})),
		TwitterId:           buildSdkTwitterId(d, "twitter"),
		ExternalSystemUrl:   platformclientv2.String(d.Get("external_system_url").(string)),
		Trustor:             buildTrustor(d.Get("trustor").([]interface{})),
		ExternalDataSources: buildExternalDataSources(d.Get("external_data_sources").([]interface{})),
	}
	tags := lists.InterfaceListToStrings(d.Get("tags").([]interface{}))
	websites := lists.InterfaceListToStrings(d.Get("websites").([]interface{}))
	externalOrganization.Tags = &tags
	externalOrganization.Websites = &websites
	if d.Get("primary_contact_id").(string) != "" {
		externalOrganization.PrimaryContactId = platformclientv2.String(d.Get("primary_contact_id").(string))
	}
	schema, err := BuildOrganizationSchema(d.Get("schema").([]interface{}))
	if err != nil {
		return externalOrganization, err
	}
	externalOrganization.Schema = schema
	customFields, err := buildCustomFieldsNillable(d.Get("custom_fields").(string))
	if err != nil {
		return externalOrganization, err

	}
	externalOrganization.CustomFields = customFields

	return externalOrganization, nil
}

// buildPhonenumberFromData is a helper method to map phone data to the GenesysCloud platformclientv2.PhoneNumber
func buildPhonenumberFromData(phoneData []interface{}) *platformclientv2.Phonenumber {

	phoneMap, ok := phoneData[0].(map[string]interface{})
	if !ok {
		return nil
	}
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
		if len(addressData) == 0 {
			return nil
		}
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
		if len(twitterData) == 0 {
			return nil
		}
		twitterMap := twitterData[0].(map[string]interface{})
		id := twitterMap["twitter_id"].(string)
		name := twitterMap["name"].(string)
		screenname := twitterMap["screen_name"].(string)

		return &platformclientv2.Twitterid{
			Id:         &id,
			Name:       &name,
			ScreenName: &screenname,
		}

	}
	return nil
}

// buildTickers maps an []interface{} into a Genesys Cloud *[]platformclientv2.Ticker
func buildTickers(tickers []interface{}) *[]platformclientv2.Ticker {
	if len(tickers) == 0 {
		return nil
	}
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

// buildTrustors maps an []interface{} into a Genesys Cloud *[]platformclientv2.Trustor
func buildTrustor(trustor []interface{}) *platformclientv2.Trustor {
	if len(trustor) == 0 {
		return nil
	}
	var sdkTrustor platformclientv2.Trustor

	trustorsMap := trustor[0].(map[string]interface{})
	sdkTrustor.Enabled = platformclientv2.Bool(trustorsMap["enabled"].(bool))

	return &sdkTrustor
}

func BuildOrganizationSchema(schemaList []interface{}) (*platformclientv2.Dataschema, error) {
	if len(schemaList) > 0 {
		schemaMap := schemaList[0].(map[string]interface{})
		dataSchema := &platformclientv2.Dataschema{}
		dataSchema.Name = platformclientv2.String(schemaMap["name"].(string))
		dataSchema.Enabled = platformclientv2.Bool(schemaMap["enabled"].(bool))
		version := platformclientv2.Int(schemaMap["version"].(int))
		if *version != 0 {
			dataSchema.Version = version
		}
		dataSchema.Id = platformclientv2.String(schemaMap["schema_id"].(string))
		jsonSchemaList := schemaMap["json_schema"].([]interface{})
		if len(jsonSchemaList) > 0 {
			jsonSchemMap := jsonSchemaList[0].(map[string]interface{})

			dataSchema.JsonSchema = &platformclientv2.Jsonschemadocument{}

			dataSchema.JsonSchema.Description = platformclientv2.String(jsonSchemMap["description"].(string))
			dataSchema.JsonSchema.Title = platformclientv2.String(jsonSchemMap["title"].(string))
			dataSchema.JsonSchema.Schema = platformclientv2.String("http://json-schema.org/draft-04/schema#")
			requiredList := lists.InterfaceListToStrings(jsonSchemMap["required"].([]interface{}))
			dataSchema.JsonSchema.Required = &requiredList

			if jsonSchemMap["properties"] != "" {
				var properties map[string]interface{}
				if err := json.Unmarshal([]byte(jsonSchemMap["properties"].(string)), &properties); err != nil {
					return nil, err
				}

				dataSchema.JsonSchema.Properties = &properties
			}

		}
		return dataSchema, nil

	}
	return nil, nil
}

// buildExternalDataSources maps an []interface{} into a Genesys Cloud *[]platformclientv2.Externaldatasource
func buildExternalDataSources(externalDataSources []interface{}) *[]platformclientv2.Externaldatasource {
	if len(externalDataSources) == 0 {
		return nil
	}
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
// flattenSdkTwitterId maps a Genesys Cloud platformclientv2.Twitterid into a []interface{}
func flattenSdkTwitterId(twitterId *platformclientv2.Twitterid) []interface{} {
	twitterMap := make(map[string]interface{})
	resourcedata.SetMapValueIfNotNil(twitterMap, "twitter_id", twitterId.Id)
	resourcedata.SetMapValueIfNotNil(twitterMap, "name", twitterId.Name)
	resourcedata.SetMapValueIfNotNil(twitterMap, "screen_name", twitterId.ScreenName)

	return []interface{}{twitterMap}
}

// flattenTrustors maps a Genesys Cloud *[]platformclientv2.Trustor into a []interface{}
func flattenTrustor(trustor *platformclientv2.Trustor) []interface{} {
	trustorMap := make(map[string]interface{})
	resourcedata.SetMapValueIfNotNil(trustorMap, "enabled", trustor.Enabled)
	return []interface{}{trustorMap}
}

// flattenJsonSchemaDocuments maps a Genesys Cloud *[]platformclientv2.Jsonschemadocument into a []interface{}
func flattenJsonSchemaDocuments(jsonSchemaDocuments *platformclientv2.Jsonschemadocument) ([]interface{}, error) {
	if jsonSchemaDocuments == nil {
		return nil, nil
	}
	jsonSchemaMap := make(map[string]interface{})
	schemaProps, err := json.Marshal(jsonSchemaDocuments.Properties)
	if err != nil {
		return nil, fmt.Errorf("error in reading json schema properties error: %v", err)
	}
	var schemaPropsPtr *string
	if string(schemaProps) != util.NullValue {
		schemaPropsStr := string(schemaProps)
		schemaPropsPtr = &schemaPropsStr
	}
	resourcedata.SetMapValueIfNotNil(jsonSchemaMap, "properties", schemaPropsPtr)
	resourcedata.SetMapValueIfNotNil(jsonSchemaMap, "description", jsonSchemaDocuments.Description)
	resourcedata.SetMapValueIfNotNil(jsonSchemaMap, "title", jsonSchemaDocuments.Title)
	jsonSchemaMap["required"] = lists.StringListToInterfaceList(*jsonSchemaDocuments.Required)
	return []interface{}{jsonSchemaMap}, nil
}

// flattenDataSchemas maps a Genesys Cloud *[]platformclientv2.Dataschema into a []interface{}
func flattenDataSchema(dataSchema *platformclientv2.Dataschema) ([]interface{}, error) {
	if dataSchema == nil {
		return nil, nil
	}
	dataSchemaMap := make(map[string]interface{})

	resourcedata.SetMapValueIfNotNil(dataSchemaMap, "name", dataSchema.Name)
	resourcedata.SetMapValueIfNotNil(dataSchemaMap, "version", dataSchema.Version)
	resourcedata.SetMapValueIfNotNil(dataSchemaMap, "enabled", dataSchema.Enabled)
	resourcedata.SetMapValueIfNotNil(dataSchemaMap, "schema_id", dataSchema.Id)
	jsonSchemaFlattened, err := flattenJsonSchemaDocuments(dataSchema.JsonSchema)
	if err != nil {
		return nil, err
	}
	dataSchemaMap["json_schema"] = &jsonSchemaFlattened
	return []interface{}{dataSchemaMap}, nil
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

func buildCustomFieldsNillable(fieldsJson string) (*map[string]interface{}, error) {
	if fieldsJson == "" {
		return nil, nil
	}

	fieldsInterface, err := util.JsonStringToInterface(fieldsJson)
	if err != nil {
		return nil, fmt.Errorf("failed to parse custom fields %s: %v", fieldsJson, err)
	}
	fieldsMap, ok := fieldsInterface.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("custom fields is not a JSON 'object': %v", fieldsJson)
	}

	return &fieldsMap, nil
}

// flattenCustomFields maps a Genesys Cloud custom fields *map[string]interface{} into a JSON string
func flattenCustomFields(customFields *map[string]interface{}) (string, error) {
	if customFields == nil {
		return "", nil
	}
	cfBytes, err := json.Marshal(customFields)
	if err != nil {
		return "", fmt.Errorf("error marshalling action contract %v: %v", customFields, err)
	}
	return string(cfBytes), nil
}
