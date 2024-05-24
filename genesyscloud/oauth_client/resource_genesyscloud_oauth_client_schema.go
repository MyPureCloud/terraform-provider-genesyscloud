package oauth_client

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	registrar "terraform-provider-genesyscloud/genesyscloud/resource_register"
)

const (
	resourceName = "genesyscloud_oauth_client"
)

func SetRegistrar(l registrar.Registrar) {
	l.RegisterDataSource(resourceName, DataSourceOAuthClient())
	l.RegisterResource(resourceName, ResourceOAuthClient())
	l.RegisterExporter(resourceName, OauthClientExporter())
}

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

func ResourceOAuthClient() *schema.Resource {
	return &schema.Resource{
		Description: "Genesys Cloud OAuth Clients. See this page for detailed configuration information: https://help.mypurecloud.com/articles/create-an-oauth-client/",

		CreateContext: provider.CreateWithPooledClient(createOAuthClient),
		ReadContext:   provider.ReadWithPooledClient(readOAuthClient),
		UpdateContext: provider.UpdateWithPooledClient(updateOAuthClient),
		DeleteContext: provider.DeleteWithPooledClient(deleteOAuthClient),
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

func OauthClientExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: provider.GetAllWithPooledClient(getAllOAuthClients),
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

func DataSourceOAuthClient() *schema.Resource {
	return &schema.Resource{
		Description: "Data source for Genesys Cloud OAuth Clients. Select an OAuth Client by name.",
		ReadContext: provider.ReadWithPooledClient(dataSourceOAuthClientRead),
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "OAuth Client name.",
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}
}
