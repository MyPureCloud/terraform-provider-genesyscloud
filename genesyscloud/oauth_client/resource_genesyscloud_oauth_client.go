package oauth_client

import (
	"context"
	"fmt"
	"log"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"terraform-provider-genesyscloud/genesyscloud/consistency_checker"
	"terraform-provider-genesyscloud/genesyscloud/util/resourcedata"

	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	lists "terraform-provider-genesyscloud/genesyscloud/util/lists"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v125/platformclientv2"
)

func getAllOAuthClients(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	resources := make(resourceExporter.ResourceIDMetaMap)

	oauthClientProxy := getOAuthClientProxy(clientConfig)
	clients, resp, getErr := oauthClientProxy.getAllOAuthClients(ctx)

	if getErr != nil {
		return nil, diag.Errorf("Failed to get page of oauth clients: %v %v", getErr, resp)
	}

	for _, client := range *clients {
		if client.State != nil && *client.State == "disabled" {
			// Don't include clients disabled by support
			continue
		}
		resources[*client.Id] = &resourceExporter.ResourceMeta{Name: *client.Name}
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
	oauthClientProxy := getOAuthClientProxy(sdkConfig)

	roles, diagErr := buildOAuthRoles(d)
	if diagErr != nil {
		return diagErr
	}

	log.Printf("Creating oauth client %s", name)
	client, resp, err := oauthClientProxy.createOAuthClient(ctx, platformclientv2.Oauthclientrequest{
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
		return diag.Errorf("Failed to create oauth client %s: %s %v", name, err, resp)
	}

	credentialName := resourcedata.GetNillableValue[string](d, "integration_credential_name")
	if credentialName != nil {

		cred_type := "pureCloudOAuthClient"
		results := make(map[string]string)
		results["clientId"] = *client.Id
		results["clientSecret"] = *client.Secret

		createCredential := platformclientv2.Credential{
			Name: credentialName,
			VarType: &platformclientv2.Credentialtype{
				Name: &cred_type,
			},
			CredentialFields: &results,
		}

		credential, resp, err := oauthClientProxy.createIntegrationClient(ctx, createCredential)

		if err != nil {
			return diag.Errorf("Failed to create credential %s : %s %v", name, err, resp)
		}

		d.Set("integration_credential_id", *credential.Id)
		d.Set("integration_credential_name", *credential.Name)
	}

	d.SetId(*client.Id)
	log.Printf("Created oauth client %s %s", name, *client.Id)
	return readOAuthClient(ctx, d, meta)
}

func readOAuthClient(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	oAuthProxy := getOAuthClientProxy(sdkConfig)

	log.Printf("Reading oauth client %s", d.Id())

	return util.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		client, resp, getErr := oAuthProxy.getOAuthClient(ctx, d.Id())
		if getErr != nil {
			if util.IsStatus404(resp) {
				return retry.RetryableError(fmt.Errorf("Failed to read oauth client %s: %s", d.Id(), getErr))
			}
			return retry.NonRetryableError(fmt.Errorf("Failed to read oauth client %s: %s", d.Id(), getErr))
		}

		cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceOAuthClient())
		d.Set("name", *client.Name)

		resourcedata.SetNillableValue(d, "description", client.Description)
		resourcedata.SetNillableValue(d, "access_token_validity_seconds", client.AccessTokenValiditySeconds)
		resourcedata.SetNillableValue(d, "authorized_grant_type", client.AuthorizedGrantType)
		resourcedata.SetNillableValue(d, "state", client.State)

		if client.RegisteredRedirectUri != nil {
			d.Set("registered_redirect_uris", lists.StringListToSet(*client.RegisteredRedirectUri))
		} else {
			d.Set("registered_redirect_uris", nil)
		}

		if client.Scope != nil {
			d.Set("scopes", lists.StringListToSet(*client.Scope))
		} else {
			d.Set("scopes", nil)
		}

		if client.RoleDivisions != nil {
			d.Set("roles", flattenOAuthRoles(*client.RoleDivisions))
		} else {
			d.Set("roles", nil)
		}

		log.Printf("Read oauth client %s %s", d.Id(), *client.Name)
		return cc.CheckState()
	})
}

func updateOAuthClient(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)
	description := d.Get("description").(string)
	tokenSeconds := d.Get("access_token_validity_seconds").(int)
	grantType := d.Get("authorized_grant_type").(string)
	state := d.Get("state").(string)

	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	oauthClientProxy := getOAuthClientProxy(sdkConfig)

	roles, diagErr := buildOAuthRoles(d)
	if diagErr != nil {
		return diagErr
	}

	log.Printf("Updating oauth client %s", name)
	_, resp, err := oauthClientProxy.updateOAuthClient(ctx, d.Id(), platformclientv2.Oauthclientrequest{
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
		return diag.Errorf("Failed to update oauth client %s: %s %v", name, err, resp)
	}

	log.Printf("Updated oauth client %s", name)
	return readOAuthClient(ctx, d, meta)
}

func deleteOAuthClient(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	oauthClientProxy := getOAuthClientProxy(sdkConfig)

	// check if there is a integration credential to delete
	credentialId := resourcedata.GetNillableValue[string](d, "integration_credential_id")
	if credentialId != nil {
		currentCredential, resp, getErr := oauthClientProxy.getIntegrationCredential(ctx, d.Id())
		if getErr == nil {
			_, err := oauthClientProxy.deleteIntegrationCredential(ctx, d.Id())
			return diag.Errorf("failed to delete integration credential %s (%s): %s %v", *currentCredential.Id, *currentCredential.Name, err, resp)
		}
	}

	name := d.Get("name").(string)

	log.Printf("Deleting oauth client %s", name)

	// The client state must be set to inactive before deleting
	d.Set("state", "inactive")
	diagErr := updateOAuthClient(ctx, d, meta)
	if diagErr != nil {
		return diagErr
	}

	resp, err := oauthClientProxy.deleteOAuthClient(ctx, d.Id())
	if err != nil {
		return diag.Errorf("Failed to delete oauth client %s: %s %v", name, err, resp)
	}

	return util.WithRetries(ctx, 30*time.Second, func() *retry.RetryError {
		oauthClient, resp, err := oauthClientProxy.getOAuthClient(ctx, d.Id())
		if err != nil {
			if util.IsStatus404(resp) {
				// OAuth client deleted
				log.Printf("Deleted OAuth client %s", d.Id())
				return nil
			}
			return retry.NonRetryableError(fmt.Errorf("Error deleting OAuth client %s: %s", d.Id(), err))
		}

		if oauthClient.State != nil && *oauthClient.State == "deleted" {
			// OAuth client deleted
			log.Printf("Deleted OAuth client %s", d.Id())
			return nil
		}
		return retry.RetryableError(fmt.Errorf("OAuth client %s still exists", d.Id()))
	})
}
