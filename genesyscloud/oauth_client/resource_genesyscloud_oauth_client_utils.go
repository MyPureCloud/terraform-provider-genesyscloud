package oauth_client

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/files"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/lists"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v165/platformclientv2"
)

func buildOAuthRedirectURIs(d *schema.ResourceData) *[]string {
	if config, ok := d.GetOk("registered_redirect_uris"); ok {
		return lists.SetToStringList(config.(*schema.Set))
	}
	return nil
}

func buildOAuthScopes(d *schema.ResourceData) *[]string {
	if config, ok := d.GetOk("scopes"); ok {
		return lists.SetToStringList(config.(*schema.Set))
	}
	return nil
}

func buildOAuthRoles(d *schema.ResourceData) (*[]platformclientv2.Roledivision, diag.Diagnostics) {
	if config, ok := d.GetOk("roles"); ok {
		var sdkRoles []platformclientv2.Roledivision
		roleConfig := config.(*schema.Set).List()
		for _, role := range roleConfig {
			roleMap := role.(map[string]interface{})
			roleId := roleMap["role_id"].(string)

			var divisionId string
			if divConfig, ok := roleMap["division_id"]; ok {
				divisionId = divConfig.(string)
			}

			if divisionId == "" {
				// Set to home division if not set
				var diagErr diag.Diagnostics
				divisionId, diagErr = util.GetHomeDivisionID()
				if diagErr != nil {
					return nil, diagErr
				}
			}

			roleDiv := platformclientv2.Roledivision{
				RoleId:     &roleId,
				DivisionId: &divisionId,
			}
			sdkRoles = append(sdkRoles, roleDiv)
		}
		return &sdkRoles, nil
	}
	return nil, nil
}

func flattenOAuthRoles(sdkRoles []platformclientv2.Roledivision) *schema.Set {
	roleSet := schema.NewSet(schema.HashResource(oauthClientRoleDivResource), []interface{}{})
	for _, roleDiv := range sdkRoles {
		role := make(map[string]interface{})
		if roleDiv.RoleId != nil {
			role["role_id"] = *roleDiv.RoleId
		}
		if roleDiv.DivisionId != nil {
			role["division_id"] = *roleDiv.DivisionId
		}
		roleSet.Add(role)
	}
	return roleSet
}

func updateMetaCache(client *platformclientv2.Oauthclient, cacheFile string) {
	integrationMeta := &provider.IntegrationMeta{
		ClientSecret: *client.Secret,
		ClientId:     *client.Id,
	}

	data, err := json.Marshal(integrationMeta)
	if err != nil {
		log.Printf("failed to convert to Json: %v", err)
	} else {
		cachePath := filepath.Join(os.TempDir(), cacheFile)
		err = os.WriteFile(cachePath, data, 0644)
		if err != nil {
			log.Printf("failed to write to cache file: %v", err)
		}
	}
}

func updateSecretToDirectory(directory string, oauthClientResult *platformclientv2.Oauthclient) {
	if directory != "" {
		path, err := files.GetDirPath(directory)
		if err != nil {
			log.Printf("Unable to fetch directory specified. %v", err)
		}
		filePath := filepath.Join(path, *oauthClientResult.Id)
		err = files.WriteToFile([]byte(*oauthClientResult.Secret), filePath)
		if err != nil {
			log.Printf("Unable to write to directory specified. %v", err)
		}
	}
}

func fetchOauthClientSecret(sdkConfig *platformclientv2.Configuration, id string) map[string]string {
	fields := make(map[string]string)
	fields["client_id"] = id
	RetrieveCachedOauthClientSecret(sdkConfig, fields)
	if _, exists := fields["client_secret"]; !exists {
		err := FetchFieldsFromMetaDataCache(fields, CacheFile)
		if err != nil {
			log.Printf("Unable to fetch from provider Cache %v", err)
		}
	}
	return fields
}

func RetrieveCachedOauthClientSecret(sdkConfig *platformclientv2.Configuration, fields map[string]string) {
	op := GetOAuthClientProxy(sdkConfig)
	log.Printf("fields %v", fields)
	if clientId, ok := fields["clientId"]; ok {
		oAuthClient := op.GetCachedOAuthClient(clientId)
		if oAuthClient != nil {
			fields["clientSecret"] = *oAuthClient.Secret
			log.Printf("Successfully matched with OAuth Client Credential id %s", clientId)
			log.Printf("Successfully matched with OAuth Client Credential id fields %v", fields)
		}
	}
}

func FetchFieldsFromMetaDataCache(fields map[string]string, cacheFile string) error {
	metaData, err := readMetaDataFromProviderCache(cacheFile)
	if err != nil {
		return err
	}
	if metaData == nil {
		return fmt.Errorf("metadata is nil")
	}

	if metaData.ClientSecret != "" {
		fields["clientSecret"] = metaData.ClientSecret
	}
	if metaData.ClientId != "" {
		fields["clientId"] = metaData.ClientId
	}

	return nil
}

func readMetaDataFromProviderCache(cacheFile string) (*provider.IntegrationMeta, error) {
	cachePath := filepath.Join(os.TempDir(), cacheFile)
	data, err := os.ReadFile(cachePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read cache file: %v", err)
	}
	if len(data) == 0 {
		return nil, fmt.Errorf("cache file is empty")
	}

	var metaData provider.IntegrationMeta
	if err := json.Unmarshal(data, &metaData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal cache data: %v", err)
	}

	return &metaData, nil
}
