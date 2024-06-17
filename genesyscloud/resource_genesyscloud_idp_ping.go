package genesyscloud

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

	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	lists "terraform-provider-genesyscloud/genesyscloud/util/lists"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v131/platformclientv2"
)

func getAllIdpPing(_ context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	idpAPI := platformclientv2.NewIdentityProviderApiWithConfig(clientConfig)
	resources := make(resourceExporter.ResourceIDMetaMap)

	_, resp, getErr := idpAPI.GetIdentityprovidersPing()
	if getErr != nil {
		if util.IsStatus404(resp) {
			// Don't export if config doesn't exist
			return resources, nil
		}
		return nil, util.BuildAPIDiagnosticError("genesyscloud_idp_ping", fmt.Sprintf("Failed to get IDP Ping error: %s", getErr), resp)
	}

	resources["0"] = &resourceExporter.ResourceMeta{Name: "ping"}
	return resources, nil
}

func IdpPingExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: provider.GetAllWithPooledClient(getAllIdpPing),
		RefAttrs:         map[string]*resourceExporter.RefAttrSettings{}, // No references
	}
}

func ResourceIdpPing() *schema.Resource {
	return &schema.Resource{
		Description: "Genesys Cloud Single Sign-on Ping Identity Provider. See this page for detailed configuration instructions: https://help.mypurecloud.com/articles/add-ping-identity-single-sign-provider/",

		CreateContext: provider.CreateWithPooledClient(createIdpPing),
		ReadContext:   provider.ReadWithPooledClient(readIdpPing),
		UpdateContext: provider.UpdateWithPooledClient(updateIdpPing),
		DeleteContext: provider.DeleteWithPooledClient(deleteIdpPing),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Timeouts: &schema.ResourceTimeout{
			Update: schema.DefaultTimeout(8 * time.Minute),
			Read:   schema.DefaultTimeout(8 * time.Minute),
		},
		Schema: map[string]*schema.Schema{
			"certificates": {
				Description: "PEM or DER encoded public X.509 certificates for SAML signature validation.",
				Type:        schema.TypeList,
				Required:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"issuer_uri": {
				Description: "Issuer URI provided by Ping.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"target_uri": {
				Description: "Target URI provided by Ping.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"relying_party_identifier": {
				Description: "String used to identify Genesys Cloud to Ping.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"disabled": {
				Description: "True if Ping is disabled.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
			},
		},
	}
}

func createIdpPing(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Printf("Creating IDP Ping")
	d.SetId("ping")
	return updateIdpPing(ctx, d, meta)
}

func readIdpPing(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	idpAPI := platformclientv2.NewIdentityProviderApiWithConfig(sdkConfig)
	cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceIdpPing(), constants.DefaultConsistencyChecks, "genesyscloud_idp_ping")

	log.Printf("Reading IDP Ping")

	return util.WithRetriesForReadCustomTimeout(ctx, d.Timeout(schema.TimeoutRead), d, func() *retry.RetryError {
		ping, resp, getErr := idpAPI.GetIdentityprovidersPing()
		if getErr != nil {
			if util.IsStatus404(resp) {
				createIdpPing(ctx, d, meta)
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError("genesyscloud_idp_ping", fmt.Sprintf("Failed to read IDP Ping: %s", getErr), resp))
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError("genesyscloud_idp_ping", fmt.Sprintf("Failed to read IDP Ping: %s", getErr), resp))
		}

		if ping.Certificate != nil {
			d.Set("certificates", lists.StringListToInterfaceList([]string{*ping.Certificate}))
		} else if ping.Certificates != nil {
			d.Set("certificates", lists.StringListToInterfaceList(*ping.Certificates))
		} else {
			d.Set("certificates", nil)
		}

		if ping.IssuerURI != nil {
			d.Set("issuer_uri", *ping.IssuerURI)
		} else {
			d.Set("issuer_uri", nil)
		}

		if ping.SsoTargetURI != nil {
			d.Set("target_uri", *ping.SsoTargetURI)
		} else {
			d.Set("target_uri", nil)
		}

		if ping.RelyingPartyIdentifier != nil {
			d.Set("relying_party_identifier", *ping.RelyingPartyIdentifier)
		} else {
			d.Set("relying_party_identifier", nil)
		}

		if ping.Disabled != nil {
			d.Set("disabled", *ping.Disabled)
		} else {
			d.Set("disabled", nil)
		}

		log.Printf("Read IDP Ping")
		return cc.CheckState(d)
	})
}

func updateIdpPing(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	issuerUri := d.Get("issuer_uri").(string)
	targetUri := d.Get("target_uri").(string)
	relyingPartyID := d.Get("relying_party_identifier").(string)
	disabled := d.Get("disabled").(bool)

	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	idpAPI := platformclientv2.NewIdentityProviderApiWithConfig(sdkConfig)

	log.Printf("Updating IDP Ping")
	update := platformclientv2.Pingidentity{
		IssuerURI:              &issuerUri,
		SsoTargetURI:           &targetUri,
		RelyingPartyIdentifier: &relyingPartyID,
		Disabled:               &disabled,
	}

	certificates := lists.BuildSdkStringListFromInterfaceArray(d, "certificates")
	if certificates != nil {
		if len(*certificates) == 1 {
			update.Certificate = &(*certificates)[0]
		}
		update.Certificates = certificates
	}

	_, resp, err := idpAPI.PutIdentityprovidersPing(update)
	if err != nil {
		return util.BuildAPIDiagnosticError("genesyscloud_idp_ping", fmt.Sprintf("Failed to update IDP Ping %s error: %s", d.Id(), err), resp)
	}

	log.Printf("Updated IDP Ping")
	return readIdpPing(ctx, d, meta)
}

func deleteIdpPing(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	idpAPI := platformclientv2.NewIdentityProviderApiWithConfig(sdkConfig)

	log.Printf("Deleting IDP Ping")
	_, resp, err := idpAPI.DeleteIdentityprovidersPing()
	if err != nil {
		return util.BuildAPIDiagnosticError("genesyscloud_idp_ping", fmt.Sprintf("Failed to delete IDP Ping %s error: %s", d.Id(), err), resp)
	}

	return util.WithRetries(ctx, 60*time.Second, func() *retry.RetryError {
		_, resp, err := idpAPI.GetIdentityprovidersPing()
		if err != nil {
			if util.IsStatus404(resp) {
				// IDP Ping deleted
				log.Printf("Deleted IDP Ping")
				return nil
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError("genesyscloud_idp_ping", fmt.Sprintf("Error deleting IDP Ping: %s", err), resp))
		}
		return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError("genesyscloud_idp_ping", fmt.Sprintf("IDP Ping still exists"), resp))
	})
}
