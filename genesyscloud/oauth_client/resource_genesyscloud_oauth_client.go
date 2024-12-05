package oauth_client

import (
	"context"
	"fmt"
	"log"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"terraform-provider-genesyscloud/genesyscloud/util/constants"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"terraform-provider-genesyscloud/genesyscloud/consistency_checker"
	"terraform-provider-genesyscloud/genesyscloud/util/resourcedata"

	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	lists "terraform-provider-genesyscloud/genesyscloud/util/lists"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v146/platformclientv2"
)

func getAllOAuthClients(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	resources := make(resourceExporter.ResourceIDMetaMap)

	oauthClientProxy := GetOAuthClientProxy(clientConfig)
	clients, resp, getErr := oauthClientProxy.getAllOAuthClients(ctx)

	if getErr != nil {
		return nil, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to get page of oauth clients error: %s", getErr), resp)
	}

	for _, client := range *clients {
		if client.State != nil && *client.State == "disabled" {
			// Don't include clients disabled by support
			continue
		}
		resources[*client.Id] = &resourceExporter.ResourceMeta{BlockLabel: *client.Name}
	}
	return resources, nil
}

func createOAuthClient(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)
	description := d.Get("description").(string)
	tokenSeconds := d.Get("access_token_validity_seconds").(int)
	grantType := d.Get("authorized_grant_type").(string)
	state := d.Get("state").(string)

	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	oauthClientProxy := GetOAuthClientProxy(sdkConfig)

	roles, diagErr := buildOAuthRoles(d)
	if diagErr != nil {
		return diagErr
	}

	//Before we create the oauth client we need to take any roles that are assigned to this oauth client and assign them to the oauth client running this script
	diagErr = updateTerraformUserWithRole(ctx, sdkConfig, roles)
	if diagErr != nil {
		return diagErr
	}

	log.Printf("Creating oauth client %s", name)
	oauthRequest := &platformclientv2.Oauthclientrequest{
		Name:                       &name,
		Description:                &description,
		AccessTokenValiditySeconds: &tokenSeconds,
		AuthorizedGrantType:        &grantType,
		State:                      &state,
		RegisteredRedirectUri:      buildOAuthRedirectURIs(d),
		Scope:                      buildOAuthScopes(d),
		RoleDivisions:              roles,
	}

	client, resp, err := oauthClientProxy.createOAuthClient(ctx, *oauthRequest)
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to create oauth client %s error: %s", name, err), resp)
	}

	if client == nil || client.Id == nil {
		return util.BuildAPIDiagnosticError(ResourceType, "failed to retrieve client ID from createOAuthClient.", resp)
	}

	createCredential(ctx, d, client, oauthClientProxy)

	d.SetId(*client.Id)
	log.Printf("Created oauth client %s %s", name, *client.Id)
	return readOAuthClient(ctx, d, meta)
}

func readOAuthClient(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	oAuthProxy := GetOAuthClientProxy(sdkConfig)
	cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceOAuthClient(), constants.ConsistencyChecks(), ResourceType)

	log.Printf("Reading oauth client %s", d.Id())

	return util.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		client, resp, getErr := oAuthProxy.getOAuthClient(ctx, d.Id())
		if getErr != nil {
			if util.IsStatus404(resp) {
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read oauth client %s | error: %s", d.Id(), getErr), resp))
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read oauth client %s | error: %s", d.Id(), getErr), resp))
		}

		resourcedata.SetNillableValue(d, "name", client.Name)
		resourcedata.SetNillableValue(d, "description", client.Description)
		resourcedata.SetNillableValue(d, "access_token_validity_seconds", client.AccessTokenValiditySeconds)
		resourcedata.SetNillableValue(d, "authorized_grant_type", client.AuthorizedGrantType)
		resourcedata.SetNillableValue(d, "state", client.State)

		if client.RegisteredRedirectUri != nil {
			_ = d.Set("registered_redirect_uris", lists.StringListToSet(*client.RegisteredRedirectUri))
		} else {
			_ = d.Set("registered_redirect_uris", nil)
		}

		if client.Scope != nil {
			_ = d.Set("scopes", lists.StringListToSet(*client.Scope))
		} else {
			_ = d.Set("scopes", nil)
		}

		if client.RoleDivisions != nil {
			_ = d.Set("roles", flattenOAuthRoles(*client.RoleDivisions))
		} else {
			_ = d.Set("roles", nil)
		}

		log.Printf("Read oauth client %s", d.Id())
		return cc.CheckState(d)
	})
}

func updateOAuthClient(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return cascadeUpdateOAuthClient(ctx, d, meta, true)
}

func deleteOAuthClient(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	oauthClientProxy := GetOAuthClientProxy(sdkConfig)

	// check if there is a integration credential to delete
	deleteCredential(ctx, d, oauthClientProxy)

	name := d.Get("name").(string)

	log.Printf("Deleting oauth client %s", name)

	// The client state must be set to inactive before deleting
	_ = d.Set("state", "inactive")
	diagErr := cascadeUpdateOAuthClient(ctx, d, meta, false)
	if diagErr != nil {
		return diagErr
	}

	resp, err := oauthClientProxy.deleteOAuthClient(ctx, d.Id())
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to delete oauth client %s error: %s", name, err), resp)
	}

	return util.WithRetries(ctx, 30*time.Second, func() *retry.RetryError {
		oauthClient, resp, err := oauthClientProxy.getOAuthClient(ctx, d.Id())
		if err != nil {
			if util.IsStatus404(resp) {
				// OAuth client deleted
				log.Printf("Deleted OAuth client %s", d.Id())
				return nil
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Error deleting OAuth client %s | error: %s", d.Id(), err), resp))
		}

		if oauthClient.State != nil && *oauthClient.State == "deleted" {
			// OAuth client deleted
			log.Printf("Deleted OAuth client %s", d.Id())
			return nil
		}
		return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("OAuth client %s still exists", d.Id()), resp))
	})
}

func updateTerraformUserWithRole(ctx context.Context, sdkConfig *platformclientv2.Configuration, addedRoles *[]platformclientv2.Roledivision) diag.Diagnostics {
	op := GetOAuthClientProxy(sdkConfig)

	//Step #1 Retrieve the parent oauth client from the token API and check to make sure it is not a client credential grant
	tokenInfo, resp, err := op.getParentOAuthClientToken(ctx)
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Error trying to retrieve the token for the OAuth client running our CX as Code provider %s", err), resp)
	}

	if *tokenInfo.OAuthClient.Organization.Id != "purecloud-builtin" {
		log.Printf("This terraform client is being run with an OAuth Client Credential Grant.  You might get an error in your terraform scripts if you try to create a role in CX as Code and try to assign it to the oauth client.")
		return nil
	}

	//Step #2: Look up the user who is running the user
	log.Printf("The OAuth Client being used is purecloud-builtin. Retrieving the user running the terraform client and assigning the target role to them.")
	terraformUser, resp, err := op.GetTerraformUser(ctx)
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to retrieved the terraform user running this terraform code %s", err), resp)
	}

	if terraformUser == nil || terraformUser.Id == nil {
		return util.BuildAPIDiagnosticError(ResourceType, "Failed to retrieve the terraform user ID with the function GetTerraformUser", resp)
	}

	//Step #3: Lookup the users addedRoles
	userRoles, resp, err := op.GetTerraformUserRoles(ctx, *terraformUser.Id)
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to retrieve the terraform user addedRoles running this terraform code %s", err), resp)
	}

	var totalRoles []string
	//Step #4  - Concat the addedRoles
	if addedRoles != nil {
		for _, role := range *addedRoles {
			if role.RoleId == nil {
				continue
			}
			totalRoles = append(totalRoles, *role.RoleId)
		}
	}

	if userRoles != nil && userRoles.Roles != nil {
		for _, role := range *userRoles.Roles {
			if role.Id == nil {
				continue
			}
			totalRoles = append(totalRoles, *role.Id)
		}
	}

	//Step #5 - Update addedRoles
	_, resp, err = op.UpdateTerraformUserRoles(ctx, *terraformUser.Id, totalRoles)
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to update the terraform user addedRoles running this terraform code %s", err), resp)
	}

	//Do not remove this sleep.  The auth service is a mishmash of caches and eventually consistency.  After we perform an update we need
	//to sleep approximately 10 seconds for the item to be written across multiple databases.  Originally, I tried to do a retry loop to
	//wait until the retry happens but the act of the first call immediately happen could cause bad data to cache.  After talking with the auth
	//team we put a sleep in here.
	time.Sleep(10 * time.Second)
	return nil
}

func cascadeUpdateOAuthClient(ctx context.Context, d *schema.ResourceData, meta interface{}, dealIntegrationFlag bool) diag.Diagnostics {
	name := d.Get("name").(string)
	description := d.Get("description").(string)
	tokenSeconds := d.Get("access_token_validity_seconds").(int)
	grantType := d.Get("authorized_grant_type").(string)
	state := d.Get("state").(string)

	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	oauthClientProxy := GetOAuthClientProxy(sdkConfig)

	roles, diagErr := buildOAuthRoles(d)
	if diagErr != nil {
		return diagErr
	}

	log.Printf("Updating oauth client %s", name)
	client, resp, err := oauthClientProxy.updateOAuthClient(ctx, d.Id(), platformclientv2.Oauthclientrequest{
		Name:                       &name,
		Description:                &description,
		AccessTokenValiditySeconds: &tokenSeconds,
		AuthorizedGrantType:        &grantType,
		State:                      &state,
		RegisteredRedirectUri:      buildOAuthRedirectURIs(d),
		Scope:                      buildOAuthScopes(d),
		RoleDivisions:              roles,
	})
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to update oauth client %s error: %s", d.Id(), err), resp)
	}

	log.Printf("Updated oauth client %s", name)

	// check if there is a integration credential to update/create
	if dealIntegrationFlag {
		credentialId := resourcedata.GetNillableValue[string](d, "integration_credential_id")
		if credentialId != nil {
			currentCredential, resp, getErr := oauthClientProxy.getIntegrationCredential(ctx, *credentialId)
			if getErr != nil {
				return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to get integration credential %s %s", *credentialId, getErr), resp)
			}

			if currentCredential != nil {
				credentialName := resourcedata.GetNillableValue[string](d, "integration_credential_name")
				if credentialName != nil {
					updateCredential(ctx, d, client, oauthClientProxy)
				} else {
					deleteCredential(ctx, d, oauthClientProxy)
				}
			}
		} else {
			createCredential(ctx, d, client, oauthClientProxy)
		}
	}

	return readOAuthClient(ctx, d, meta)
}

func createCredential(ctx context.Context, d *schema.ResourceData, client *platformclientv2.Oauthclient, oauthClientProxy *oauthClientProxy) diag.Diagnostics {
	credentialName := resourcedata.GetNillableValue[string](d, "integration_credential_name")
	if credentialName != nil {

		credType := "pureCloudOAuthClient"
		results := make(map[string]string)
		results["clientId"] = *client.Id
		results["clientSecret"] = *client.Secret

		createCredential := platformclientv2.Credential{
			Name: credentialName,
			VarType: &platformclientv2.Credentialtype{
				Name: &credType,
			},
			CredentialFields: &results,
		}

		log.Printf("Creating Integration Credential client %s", *credentialName)
		credential, resp, err := oauthClientProxy.createIntegrationClient(ctx, createCredential)
		if err != nil {
			return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to create credential %s error: %s", *credentialName, err), resp)
		}

		resourcedata.SetNillableValue(d, "integration_credential_id", credential.Id)
		resourcedata.SetNillableValue(d, "integration_credential_name", credential.Name)

		log.Printf("Created Integration Credential client %s", *credentialName)
	}
	return nil
}

func updateCredential(ctx context.Context, d *schema.ResourceData, client *platformclientv2.Oauthclient, oauthClientProxy *oauthClientProxy) diag.Diagnostics {
	credentialId := resourcedata.GetNillableValue[string](d, "integration_credential_id")
	credentialName := resourcedata.GetNillableValue[string](d, "integration_credential_name")
	if credentialName != nil {
		credType := "pureCloudOAuthClient"
		results := make(map[string]string)
		results["clientId"] = *client.Id
		results["clientSecret"] = *client.Secret

		updateCred := platformclientv2.Credential{
			Name: credentialName,
			VarType: &platformclientv2.Credentialtype{
				Name: &credType,
			},
			CredentialFields: &results,
		}
		log.Printf("Updating Integration Credential client %s", *credentialName)
		credential, resp, err := oauthClientProxy.updateIntegrationClient(ctx, *credentialId, updateCred)

		if err != nil {
			return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to update credential %s error: %s", *credentialName, err), resp)
		}

		resourcedata.SetNillableValue(d, "integration_credential_id", credential.Id)
		resourcedata.SetNillableValue(d, "integration_credential_name", credential.Name)

		log.Printf("Updated Integration Credential client %s", *credentialName)
	}
	return nil
}

func deleteCredential(ctx context.Context, d *schema.ResourceData, oauthClientProxy *oauthClientProxy) diag.Diagnostics {
	credentialId := resourcedata.GetNillableValue[string](d, "integration_credential_id")

	if credentialId != nil {
		log.Printf("Deleting integration credential %s", *credentialId)
		_, resp, getErr := oauthClientProxy.getIntegrationCredential(ctx, *credentialId)
		if getErr != nil {
			return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to get integration credential %s %s", *credentialId, getErr), resp)
		}
		_, err := oauthClientProxy.deleteIntegrationCredential(ctx, *credentialId)
		if err != nil {
			return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to delete integration credential %s %s", *credentialId, err), resp)
		}
		log.Printf("Deleted Integration Credential client %s", *credentialId)

		_ = d.Set("integration_credential_id", nil)
		_ = d.Set("integration_credential_name", nil)
	}
	return nil
}
