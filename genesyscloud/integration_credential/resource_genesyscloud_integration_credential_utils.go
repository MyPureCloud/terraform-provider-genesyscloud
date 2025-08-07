package integration_credential

import (
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v165/platformclientv2"
	oauth "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/oauth_client"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
)

/*
The resource_genesyscloud_integration_credential_utils.go file contains various helper methods to marshal
and unmarshal data into formats consumable by Terraform and/or Genesys Cloud.

Note:  Look for opportunities to minimize boilerplate code using functions and Generics

//If if is a Genesys Cloud OAuth Client and the user has not provided a secret field we should look for the
//item in the cache DEVTOOLING-448
*/

func buildCredentialFields(d *schema.ResourceData, sdkConfig *platformclientv2.Configuration) map[string]string {
	if d == nil || sdkConfig == nil {
		log.Println("ResourceData or SDK config is nil")
		return nil
	}

	fieldsMap := getFieldsFromResource(d)

	updatedFields, err := fetchValueFromFields(d, fieldsMap, sdkConfig)
	if err != nil {
		log.Printf("Failed to fetch value from fields: %v", err)
	}

	return updatedFields
}

func getFieldsFromResource(d *schema.ResourceData) map[string]string {
	results := make(map[string]string)

	if fields, ok := d.GetOk("fields"); ok {
		if fieldMap, ok := fields.(map[string]interface{}); ok {
			for k, v := range fieldMap {
				results[k] = v.(string)
			}
		}
	}

	return results
}

// fetchValueFromFields builds a map of credential fields from the resource
func fetchValueFromFields(d *schema.ResourceData, fields map[string]string, sdkConfig *platformclientv2.Configuration) (map[string]string, error) {
	if fields == nil {
		return nil, fmt.Errorf("invalid input parameters: one or more required parameters are nil")
	}

	credType, ok := d.Get("credential_type_name").(string)
	if !ok {
		return nil, fmt.Errorf("invalid or missing credential_type_name")
	}

	result := make(map[string]string, len(fields))
	for k, v := range fields {
		result[k] = v
	}

	if credType == "pureCloudOAuthClient" && !isFieldNonEmpty(result, "clientSecret") {
		oauth.RetrieveCachedOauthClientSecret(sdkConfig, result)

		// Retry with metadata cache if still missing
		if !isFieldNonEmpty(result, "clientSecret") {
			if err := oauth.FetchFieldsFromMetaDataCache(result, oauth.CacheFile); err != nil {
				log.Printf("failed to update fields from cache: %v", err)
			}
		}
	}
	return result, nil
}

func isFieldNonEmpty(fields map[string]string, fieldName string) bool {
	value, exists := fields[fieldName]
	return exists && value != ""
}

// GenerateCredentialResource generates the terraform string for creating genesyscloud_integration_credential resource. Used for testing.
func GenerateCredentialResource(resourceLabel string, name string, credentialType string, fields string) string {
	return fmt.Sprintf(`resource "genesyscloud_integration_credential" "%s" {
        name = %s
        credential_type_name = %s
        %s
	}
	`, resourceLabel, name, credentialType, fields)
}

// GenerateCredentialFields builds a terraform string for multiple credential fields
func GenerateCredentialFields(fields map[string]string) string {
	return util.GenerateMapAttrWithMapProperties("fields", fields)
}
