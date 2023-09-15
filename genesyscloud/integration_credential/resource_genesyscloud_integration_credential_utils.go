package integration_credential

import (
	"fmt"

	gcloud "terraform-provider-genesyscloud/genesyscloud"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func buildCredentialFields(d *schema.ResourceData) *map[string]string {
	results := make(map[string]string)
	if fields, ok := d.GetOk("fields"); ok {
		fieldMap := fields.(map[string]interface{})
		for k, v := range fieldMap {
			results[k] = v.(string)
		}
		return &results
	}
	return &results
}

func GenerateCredentialResource(resourceID string, name string, credentialType string, fields string) string {
	return fmt.Sprintf(`resource "genesyscloud_integration_credential" "%s" {
        name = %s
        credential_type_name = %s
        %s
	}
	`, resourceID, name, credentialType, fields)
}

func GenerateCredentialFields(fields ...string) string {
	return gcloud.GenerateMapAttr("fields", fields...)
}
