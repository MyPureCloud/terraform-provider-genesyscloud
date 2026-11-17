package integration_credential

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"sort"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v200/platformclientv2"
	oauth "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/oauth_client"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
)

const (
	// fieldsWoAttr is the write-only attribute holding the credential fields as a JSON object.
	// Its value is never stored in the plan or state, so it can only be read from the raw config.
	fieldsWoAttr = "fields_wo"
	// fieldsWoVersionAttr is the stored companion attribute that signals the write-only value changed.
	fieldsWoVersionAttr = "fields_wo_version"
)

/*
The resource_genesyscloud_integration_credential_utils.go file contains various helper methods to marshal
and unmarshal data into formats consumable by Terraform and/or Genesys Cloud.

Note:  Look for opportunities to minimize boilerplate code using functions and Generics

//If if is a Genesys Cloud OAuth Client and the user has not provided a secret field we should look for the
//item in the cache DEVTOOLING-448
*/

// buildCredentialFields returns the credential fields to send to Genesys Cloud. The source is the
// write-only fields_wo (read from the raw config, never from state) when it is set, otherwise the
// existing fields map. The two are mutually exclusive in the schema.
func buildCredentialFields(d *schema.ResourceData, sdkConfig *platformclientv2.Configuration) (map[string]string, diag.Diagnostics) {
	if d == nil || sdkConfig == nil {
		log.Println("ResourceData or SDK config is nil")
		return nil, nil
	}

	fieldsMap, woSet, err := getFieldsWoFromRawConfig(d)
	if err != nil {
		return nil, util.BuildDiagnosticError(ResourceType, "Invalid "+fieldsWoAttr, err)
	}
	if !woSet {
		fieldsMap = getFieldsFromResource(d)
	}

	updatedFields, err := fetchValueFromFields(d, fieldsMap, sdkConfig)
	if err != nil {
		log.Printf("Failed to fetch value from fields: %v", err)
	}

	return updatedFields, nil
}

// scrubStoredFieldsWhenWriteOnly clears any plaintext previously stored in `fields` once the credential's
// secrets are supplied through the write-only fields_wo. `fields` is Optional+Computed, so when a user
// migrates from `fields` to `fields_wo` and removes `fields` from their config, Terraform keeps the old
// value in state. That value is stale and still a plaintext secret, which would defeat the purpose of
// migrating, so it is cleared here.
func scrubStoredFieldsWhenWriteOnly(d *schema.ResourceData) {
	if _, set, err := getFieldsWoFromRawConfig(d); err == nil && set {
		_ = d.Set("fields", map[string]interface{}{})
	}
}

// ensureFieldsHasKnownValue records an explicit empty map in `fields` when nothing is stored there.
// `fields` is Optional+Computed and is never read back from the API (secrets are write-only on the
// server), so a credential created with the write-only fields_wo has no value for it in state. A null
// Computed value is planned as "known after apply" on every plan, which proposes a permanent in-place
// update (and would re-send the secret on every apply). An empty map is a known, non-secret value.
// Credentials that store their fields in `fields` are untouched.
func ensureFieldsHasKnownValue(d *schema.ResourceData) {
	if fields, ok := d.Get("fields").(map[string]interface{}); ok && len(fields) > 0 {
		return
	}
	_ = d.Set("fields", map[string]interface{}{})
}

// credentialChanged reports whether the credential must be sent to the API on update. fields_wo is
// write-only and never stored, so Terraform cannot see it change; fields_wo_version is the stored
// trigger. Changing it re-sends the credential, including the secret, to the API.
func credentialChanged(d *schema.ResourceData) bool {
	return d.HasChanges("name", "credential_type_name", "fields", fieldsWoVersionAttr)
}

// getFieldsWoFromRawConfig reads the write-only fields_wo attribute from the raw config. A write-only
// value is always null in state and plan, so d.Get cannot be used. The second return value reports
// whether fields_wo was set (known and non-null); when it is not, the caller falls back to fields.
// The raw config is null during operations that have no configuration (such as import).
func getFieldsWoFromRawConfig(d *schema.ResourceData) (map[string]string, bool, error) {
	raw := d.GetRawConfig()
	if raw.IsNull() || !raw.IsKnown() || !raw.Type().IsObjectType() || !raw.Type().HasAttribute(fieldsWoAttr) {
		return nil, false, nil
	}

	val := raw.GetAttr(fieldsWoAttr)
	if val.IsNull() || !val.IsKnown() || !val.Type().Equals(cty.String) {
		return nil, false, nil
	}

	fields, err := parseFieldsWoJSON(val.AsString())
	if err != nil {
		return nil, false, err
	}
	return fields, true, nil
}

// parseFieldsWoJSON parses the fields_wo JSON object into the map sent to the API. Values must be strings.
// An empty object is rejected: the update replaces the credential, so sending no fields would wipe it.
// Errors never include any part of the value, only offsets and key names, because the value is a secret.
func parseFieldsWoJSON(raw string) (map[string]string, error) {
	var generic map[string]interface{}
	if err := json.Unmarshal([]byte(raw), &generic); err != nil {
		var syntaxErr *json.SyntaxError
		if errors.As(err, &syntaxErr) {
			return nil, fmt.Errorf("%s must be a valid JSON object (syntax error at offset %d)", fieldsWoAttr, syntaxErr.Offset)
		}
		return nil, fmt.Errorf("%s must be a JSON object whose values are strings", fieldsWoAttr)
	}
	if len(generic) == 0 {
		return nil, fmt.Errorf("%s must contain at least one field", fieldsWoAttr)
	}

	keys := make([]string, 0, len(generic))
	for k := range generic {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	result := make(map[string]string, len(generic))
	for _, k := range keys {
		s, ok := generic[k].(string)
		if !ok {
			return nil, fmt.Errorf("%s: the value for key %q must be a string", fieldsWoAttr, k)
		}
		result[k] = s
	}
	return result, nil
}

// validateFieldsWoJSON is the plan-time validator for fields_wo. Unknown values are skipped by the SDK.
func validateFieldsWoJSON(v interface{}, _ cty.Path) diag.Diagnostics {
	s, ok := v.(string)
	if !ok {
		return diag.Errorf("%s must be a string", fieldsWoAttr)
	}
	if _, err := parseFieldsWoJSON(s); err != nil {
		return diag.FromErr(err)
	}
	return nil
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
