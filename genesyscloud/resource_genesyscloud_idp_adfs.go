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
	"github.com/mypurecloud/platform-client-sdk-go/v129/platformclientv2"
)

func getAllIdpAdfs(_ context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	idpAPI := platformclientv2.NewIdentityProviderApiWithConfig(clientConfig)
	resources := make(resourceExporter.ResourceIDMetaMap)

	_, resp, getErr := idpAPI.GetIdentityprovidersAdfs()
	if getErr != nil {
		if util.IsStatus404(resp) {
			// Don't export if config doesn't exist
			return resources, nil
		}
		return nil, util.BuildAPIDiagnosticError("genesyscloud_idp_adfs", fmt.Sprintf("Failed to get IDP ADFS error: %s", getErr), resp)
	}

	resources["0"] = &resourceExporter.ResourceMeta{Name: "adfs"}
	return resources, nil
}

func IdpAdfsExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: provider.GetAllWithPooledClient(getAllIdpAdfs),
		RefAttrs:         map[string]*resourceExporter.RefAttrSettings{}, // No references
	}
}

func ResourceIdpAdfs() *schema.Resource {
	return &schema.Resource{
		Description: "Genesys Cloud Single Sign-on ADFS Identity Provider. See this page for detailed configuration instructions: https://help.mypurecloud.com/articles/add-microsoft-adfs-single-sign-provider/",

		CreateContext: provider.CreateWithPooledClient(createIdpAdfs),
		ReadContext:   provider.ReadWithPooledClient(readIdpAdfs),
		UpdateContext: provider.UpdateWithPooledClient(updateIdpAdfs),
		DeleteContext: provider.DeleteWithPooledClient(deleteIdpAdfs),
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
				Description: "Issuer URI provided by ADFS.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"target_uri": {
				Description: "Target URI provided by ADFS.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"relying_party_identifier": {
				Description: "String used to identify Genesys Cloud to ADFS.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"disabled": {
				Description: "True if ADFS is disabled.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
			},
		},
	}
}

func createIdpAdfs(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Printf("Creating IDP ADFS")
	d.SetId("adfs")
	return updateIdpAdfs(ctx, d, meta)
}

func readIdpAdfs(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	idpAPI := platformclientv2.NewIdentityProviderApiWithConfig(sdkConfig)
	cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceIdpAdfs(), constants.DefaultConsistencyChecks, "genesyscloud_idp_adfs")

	log.Printf("Reading IDP ADFS")
	return util.WithRetriesForReadCustomTimeout(ctx, d.Timeout(schema.TimeoutRead), d, func() *retry.RetryError {
		adfs, resp, getErr := idpAPI.GetIdentityprovidersAdfs()
		if getErr != nil {
			if util.IsStatus404(resp) {
				createIdpAdfs(ctx, d, meta)
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError("genesyscloud_idp_adfs", fmt.Sprintf("Failed to read IDP ADFS: %s", getErr), resp))
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError("genesyscloud_idp_adfs", fmt.Sprintf("Failed to read IDP ADFS: %s", getErr), resp))
		}

		if adfs.Certificate != nil {
			d.Set("certificates", lists.StringListToInterfaceList([]string{*adfs.Certificate}))
		} else if adfs.Certificates != nil {
			d.Set("certificates", lists.StringListToInterfaceList(*adfs.Certificates))
		} else {
			d.Set("certificates", nil)
		}

		if adfs.IssuerURI != nil {
			d.Set("issuer_uri", *adfs.IssuerURI)
		} else {
			d.Set("issuer_uri", nil)
		}

		if adfs.SsoTargetURI != nil {
			d.Set("target_uri", *adfs.SsoTargetURI)
		} else {
			d.Set("target_uri", nil)
		}

		if adfs.RelyingPartyIdentifier != nil {
			d.Set("relying_party_identifier", *adfs.RelyingPartyIdentifier)
		} else {
			d.Set("relying_party_identifier", nil)
		}

		if adfs.Disabled != nil {
			d.Set("disabled", *adfs.Disabled)
		} else {
			d.Set("disabled", nil)
		}

		log.Printf("Read IDP ADFS")
		return cc.CheckState(d)
	})
}

func updateIdpAdfs(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	issuerUri := d.Get("issuer_uri").(string)
	targetUri := d.Get("target_uri").(string)
	relyingPartyID := d.Get("relying_party_identifier").(string)
	disabled := d.Get("disabled").(bool)

	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	idpAPI := platformclientv2.NewIdentityProviderApiWithConfig(sdkConfig)

	log.Printf("Updating IDP ADFS")
	update := platformclientv2.Adfs{
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

	_, resp, err := idpAPI.PutIdentityprovidersAdfs(update)
	if err != nil {
		return util.BuildAPIDiagnosticError("genesyscloud_idp_adfs", fmt.Sprintf("Failed to update IDP ADFS %s error: %s", d.Id(), err), resp)
	}

	log.Printf("Updated IDP ADFS")
	return readIdpAdfs(ctx, d, meta)
}

func deleteIdpAdfs(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	idpAPI := platformclientv2.NewIdentityProviderApiWithConfig(sdkConfig)

	log.Printf("Deleting IDP ADFS")
	_, resp, err := idpAPI.DeleteIdentityprovidersAdfs()
	if err != nil {
		return util.BuildAPIDiagnosticError("genesyscloud_idp_adfs", fmt.Sprintf("Failed to delete IDP ADFS %s error: %s", d.Id(), err), resp)
	}

	return util.WithRetries(ctx, 180*time.Second, func() *retry.RetryError {
		_, resp, err := idpAPI.GetIdentityprovidersAdfs()
		if err != nil {
			if util.IsStatus404(resp) {
				// IDP ADFS deleted
				log.Printf("Deleted IDP ADFS")
				return nil
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError("genesyscloud_idp_adfs", fmt.Sprintf("Error deleting IDP ADFS: %s", err), resp))
		}
		return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError("genesyscloud_idp_adfs", fmt.Sprintf("IDP ADFS still exists"), resp))
	})
}
