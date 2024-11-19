package integration_credential

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"terraform-provider-genesyscloud/genesyscloud/util"
)

/*
The resource_genesyscloud_integration_credential_utils.go file contains various helper methods to marshal
and unmarshal data into formats consumable by Terraform and/or Genesys Cloud.

Note:  Look for opportunities to minimize boilerplate code using functions and Generics
*/

// buildCredentialFields builds a map of credential fields from the resource
func buildCredentialFields(d *schema.ResourceData) map[string]string {
	results := make(map[string]string)
	if fields, ok := d.GetOk("fields"); ok {
		fieldMap := fields.(map[string]interface{})
		for k, v := range fieldMap {
			results[k] = v.(string)
		}
		return results
	}
	return results
}

// GenerateCredentialResource generates the terraform string for creating genesyscloud_integration_credential resource. Used for testing.
func GenerateCredentialResource(resourceID string, name string, credentialType string, fields string) string {
	return fmt.Sprintf(`resource "genesyscloud_integration_credential" "%s" {
        name = %s
        credential_type_name = %s
        %s
	}
	`, resourceID, name, credentialType, fields)
}

// GenerateCredentialFields builds a terraform string for multiple credential fields
func GenerateCredentialFields(fields map[string]string) string {
	return util.GenerateMapAttrWithMapProperties("fields", fields)
}
