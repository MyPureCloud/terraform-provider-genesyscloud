package genesyscloud

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"terraform-provider-genesyscloud/genesyscloud/consistency_checker"
	"terraform-provider-genesyscloud/genesyscloud/util/resourcedata"

	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	lists "terraform-provider-genesyscloud/genesyscloud/util/lists"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/mypurecloud/platform-client-sdk-go/v115/platformclientv2"
)

var (
	oauthClientRoleDivResource = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"role_id": {
				Description: "Role to be associated with the given division which forms a grant.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"division_id": {
				Description: "Division associated with the given role which forms a grant. If not set, the home division will be used. '*' may be set for all divisions.",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
			},
		},
	}
)

func getAllOAuthClients(_ context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	resources := make(resourceExporter.ResourceIDMetaMap)
	oauthAPI := platformclientv2.NewOAuthApiWithConfig(clientConfig)

	clients, _, getErr := oauthAPI.GetOauthClients()
	if getErr != nil {
		return nil, diag.Errorf("Failed to get page of oauth clients: %v", getErr)
	}

	if clients.Entities == nil || len(*clients.Entities) == 0 {
		return resources, nil
	}

	for _, client := range *clients.Entities {
		if client.State != nil && *client.State == "disabled" {
			// Don't include clients disabled by support
			continue
		}
		resources[*client.Id] = &resourceExporter.ResourceMeta{Name: *client.Name}
	}

	return resources, nil
}

func OauthClientExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: GetAllWithPooledClient(getAllOAuthClients),
		RefAttrs: map[string]*resourceExporter.RefAttrSettings{
			"roles.role_id":             {RefType: "genesyscloud_auth_role"},
			"roles.division_id":         {RefType: "genesyscloud_auth_division", AltValues: []string{"*"}},
			"integration_credential_id": {RefType: "genesyscloud_integration_credential"},
		},
		RemoveIfMissing: map[string][]string{
			"roles":                       {"role_id"},
			"integration_credential_id":   {"integration_credential_id"},
			"integration_credential_name": {"integration_credential_name"},
		},
	}
}

func ResourceOAuthClient() *schema.Resource {
	return &schema.Resource{
		Description: "Genesys Cloud OAuth Clients. See this page for detailed configuration information: https://help.mypurecloud.com/articles/create-an-oauth-client/",

		CreateContext: CreateWithPooledClient(createOAuthClient),
		ReadContext:   ReadWithPooledClient(readOAuthClient),
		UpdateContext: UpdateWithPooledClient(updateOAuthClient),
		DeleteContext: DeleteWithPooledClient(deleteOAuthClient),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "The name of the OAuth client.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"description": {
				Description: "The description of the OAuth client.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"access_token_validity_seconds": {
				Description:  "The number of seconds, between 5mins and 48hrs, until tokens created with this client expire. Only clients using Genesys Cloud SCIM (Identity Management) can have a maximum duration of 38880000secs/450 days.",
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntBetween(300, 38880000),
				Default:      86400,
			},
			"registered_redirect_uris": {
				Description: "List of allowed callbacks for this client. For example: https://myapp.example.com/auth/callback.",
				Type:        schema.TypeSet,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Optional:    true,
			},
			"authorized_grant_type": {
				Description:  "The OAuth Grant/Client type supported by this client (CODE | TOKEN | SAML2BEARER | PASSWORD | CLIENT-CREDENTIALS).",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{"CODE", "TOKEN", "SAML2BEARER", "PASSWORD", "CLIENT-CREDENTIALS"}, false),
			},
			"scopes": {
				Description: "The scopes requested by this client. Scopes must be set for clients not using the CLIENT-CREDENTIALS grant.",
				Type:        schema.TypeSet,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Optional:    true,
			},
			"roles": {
				Description: "Set of roles and their corresponding divisions associated with this client. Roles must be set for clients using the CLIENT-CREDENTIALS grant. The roles must also already be assigned to the OAuth Client used by Terraform.",
				Type:        schema.TypeSet,
				Elem:        oauthClientRoleDivResource,
				Optional:    true,
			},
			"state": {
				Description:  "The state of the OAuth client (active | inactive). Access tokens cannot be created with inactive clients.",
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice([]string{"active", "inactive"}, false),
				Default:      "active",
			},
			"integration_credential_id": {
				Description: "The Id of the created Integration Credential using this new OAuth Client.",
				Type:        schema.TypeString,
				Optional:    false,
				Required:    false,
				Computed:    true, //If Required and Optional are both false, the attribute will be considered
				// "read only" for the practitioner, with only the provider able to set its value.
			},
			"integration_credential_name": {
				Description: "Optionally, a Name of a Integration Credential (with credential type pureCloudOAuthClient) to be created using this new OAuth Client.",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
			},
		},
	}
}

func createOAuthClient(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)
	description := d.Get("description").(string)
	tokenSeconds := d.Get("access_token_validity_seconds").(int)
	grantType := d.Get("authorized_grant_type").(string)
	state := d.Get("state").(string)

	sdkConfig := meta.(*ProviderMeta).ClientConfig
	oauthAPI := platformclientv2.NewOAuthApiWithConfig(sdkConfig)

	roles, diagErr := buildOAuthRoles(d)
	if diagErr != nil {
		return diagErr
	}

	log.Printf("Creating oauth client %s", name)
	client, _, err := oauthAPI.PostOauthClients(platformclientv2.Oauthclientrequest{
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
		return diag.Errorf("Failed to create oauth client %s: %s", name, err)
	}

	credentialName := resourcedata.GetNillableValue[string](d, "integration_credential_name")
	if credentialName != nil {
		integrationAPI := platformclientv2.NewIntegrationsApiWithConfig(sdkConfig)
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

		credential, _, err := integrationAPI.PostIntegrationsCredentials(createCredential)

		if err != nil {
			return diag.Errorf("Failed to create credential %s : %s", name, err)
		}

		d.Set("integration_credential_id", *credential.Id)
		d.Set("integration_credential_name", *credential.Name)
	}

	d.SetId(*client.Id)
	log.Printf("Created oauth client %s %s", name, *client.Id)
	return readOAuthClient(ctx, d, meta)
}

func readOAuthClient(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*ProviderMeta).ClientConfig
	oauthAPI := platformclientv2.NewOAuthApiWithConfig(sdkConfig)

	log.Printf("Reading oauth client %s", d.Id())

	return WithRetriesForRead(ctx, d, func() *retry.RetryError {
		client, resp, getErr := oauthAPI.GetOauthClient(d.Id())
		if getErr != nil {
			if IsStatus404(resp) {
				return retry.RetryableError(fmt.Errorf("Failed to read oauth client %s: %s", d.Id(), getErr))
			}
			return retry.NonRetryableError(fmt.Errorf("Failed to read oauth client %s: %s", d.Id(), getErr))
		}

		cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceOAuthClient())
		d.Set("name", *client.Name)

		if client.Description != nil {
			d.Set("description", *client.Description)
		} else {
			d.Set("description", nil)
		}

		if client.AccessTokenValiditySeconds != nil {
			d.Set("access_token_validity_seconds", *client.AccessTokenValiditySeconds)
		} else {
			d.Set("access_token_validity_seconds", nil)
		}

		if client.AuthorizedGrantType != nil {
			d.Set("authorized_grant_type", *client.AuthorizedGrantType)
		} else {
			d.Set("authorized_grant_type", nil)
		}

		if client.State != nil {
			d.Set("state", *client.State)
		} else {
			d.Set("state", nil)
		}

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

	sdkConfig := meta.(*ProviderMeta).ClientConfig
	oauthAPI := platformclientv2.NewOAuthApiWithConfig(sdkConfig)

	roles, diagErr := buildOAuthRoles(d)
	if diagErr != nil {
		return diagErr
	}

	log.Printf("Updating oauth client %s", name)
	_, _, err := oauthAPI.PutOauthClient(d.Id(), platformclientv2.Oauthclientrequest{
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
		return diag.Errorf("Failed to update oauth client %s: %s", name, err)
	}

	log.Printf("Updated oauth client %s", name)

	return readOAuthClient(ctx, d, meta)
}

func deleteOAuthClient(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*ProviderMeta).ClientConfig

	// check if there is a integration credential to delete
	credentialId := resourcedata.GetNillableValue[string](d, "integration_credential_id")
	if credentialId != nil {
		integrationAPI := platformclientv2.NewIntegrationsApiWithConfig(sdkConfig)
		currentCredential, _, getErr := integrationAPI.GetIntegrationsCredential(d.Id())
		if getErr == nil {
			_, err := integrationAPI.DeleteIntegrationsCredential(d.Id())
			diag.Errorf("failed to delete integration credential %s (%s): %s", *currentCredential.Id, *currentCredential.Name, err)
		}
	}

	name := d.Get("name").(string)

	oauthAPI := platformclientv2.NewOAuthApiWithConfig(sdkConfig)

	log.Printf("Deleting oauth client %s", name)

	// The client state must be set to inactive before deleting
	d.Set("state", "inactive")
	diagErr := updateOAuthClient(ctx, d, meta)
	if diagErr != nil {
		return diagErr
	}

	_, err := oauthAPI.DeleteOauthClient(d.Id())
	if err != nil {
		return diag.Errorf("Failed to delete oauth client %s: %s", name, err)
	}

	return WithRetries(ctx, 30*time.Second, func() *retry.RetryError {
		oauthClient, resp, err := oauthAPI.GetOauthClient(d.Id())
		if err != nil {
			if IsStatus404(resp) {
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
				divisionId, diagErr = getHomeDivisionID()
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
