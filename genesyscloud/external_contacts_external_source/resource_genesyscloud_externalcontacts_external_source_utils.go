package external_contacts_external_source

import (
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/resourcedata"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v165/platformclientv2"
)

/*
The resource_genesyscloud_external_contacts_external_source_utils.go file contains various helper methods to marshal
and unmarshal data into formats consumable by Terraform and/or Genesys Cloud.
*/

// getExternalContactsExternalSourceFromResourceData maps data from schema ResourceData object to a platformclientv2.Externalsource
func getExternalContactsExternalSourceFromResourceData(d *schema.ResourceData) (platformclientv2.Externalsource, error) {
	externalSource := platformclientv2.Externalsource{
		Name:              platformclientv2.String(d.Get("name").(string)),
		Active:            platformclientv2.Bool(d.Get("active").(bool)),
		LinkConfiguration: buildLinkConfiguration(d, "link_configuration"),
	}

	return externalSource, nil
}

// buildLinkConfiguration constructs a platformclientv2.Linkconfiguration structure
func buildLinkConfiguration(d *schema.ResourceData, key string) *platformclientv2.Linkconfiguration {
	if d.Get(key) != nil {
		linkConfigurationData := d.Get(key).([]interface{})
		if len(linkConfigurationData) == 0 {
			return nil
		}
		linkConfigurationMap := linkConfigurationData[0].(map[string]interface{})
		uriTemplate := linkConfigurationMap["uri_template"].(string)

		return &platformclientv2.Linkconfiguration{
			UriTemplate: &uriTemplate,
		}

	}
	return nil
}

// flattenLinkConfiguration converts a *platformclientv2.LinkConfiguration into a map and then into array for consumption by Terraform
func flattenLinkConfiguration(linkConfiguration *platformclientv2.Linkconfiguration) []interface{} {
	linkConfigurationInterface := make(map[string]interface{})

	resourcedata.SetMapValueIfNotNil(linkConfigurationInterface, "uri_template", linkConfiguration.UriTemplate)

	return []interface{}{linkConfigurationInterface}
}
