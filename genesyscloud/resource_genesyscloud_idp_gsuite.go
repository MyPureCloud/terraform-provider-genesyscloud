package genesyscloud

import (
	"context"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v48/platformclientv2"
)

func getAllIdpGsuite(ctx context.Context, clientConfig *platformclientv2.Configuration) (ResourceIDMetaMap, diag.Diagnostics) {
	idpAPI := platformclientv2.NewIdentityProviderApiWithConfig(clientConfig)
	resources := make(ResourceIDMetaMap)

	_, resp, getErr := idpAPI.GetIdentityprovidersGsuite()
	if getErr != nil {
		if resp != nil && resp.StatusCode == 404 {
			// Don't export if config doesn't exist
			return resources, nil
		}
		return nil, diag.Errorf("Failed to get IDP GSuite: %v", getErr)
	}

	resources["0"] = &ResourceMeta{Name: "gsuite"}
	return resources, nil
}

func idpGsuiteExporter() *ResourceExporter {
	return &ResourceExporter{
		GetResourcesFunc: getAllWithPooledClient(getAllIdpGsuite),
		RefAttrs:         map[string]*RefAttrSettings{}, // No references
	}
}

func resourceIdpGsuite() *schema.Resource {
	return &schema.Resource{
		Description: "Genesys Cloud Single Sign-on GSuite Identity Provider. See this page for detailed configuration instructions: https://help.mypurecloud.com/articles/add-google-g-suite-single-sign-provider/",

		CreateContext: createWithPooledClient(createIdpGsuite),
		ReadContext:   readWithPooledClient(readIdpGsuite),
		UpdateContext: updateWithPooledClient(updateIdpGsuite),
		DeleteContext: deleteWithPooledClient(deleteIdpGsuite),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			"certificates": {
				Description: "PEM or DER encoded public X.509 certificates for SAML signature validation.",
				Type:        schema.TypeSet,
				Required:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"issuer_uri": {
				Description: "Issuer URI provided by GSuite.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"target_uri": {
				Description: "Target URI provided by GSuite.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"relying_party_identifier": {
				Description: "String used to identify Genesys Cloud to GSuite.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"disabled": {
				Description: "True if GSuite is disabled.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
			},
		},
	}
}

func createIdpGsuite(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Printf("Creating IDP GSuite")
	d.SetId("gsuite")
	return updateIdpGsuite(ctx, d, meta)
}

func readIdpGsuite(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*providerMeta).ClientConfig
	idpAPI := platformclientv2.NewIdentityProviderApiWithConfig(sdkConfig)

	log.Printf("Reading IDP GSuite")
	gsuite, resp, getErr := idpAPI.GetIdentityprovidersGsuite()
	if getErr != nil {
		if resp != nil && resp.StatusCode == 404 {
			d.SetId("")
			return nil
		}
		return diag.Errorf("Failed to read IDP GSuite: %s", getErr)
	}

	if gsuite.Certificate != nil {
		d.Set("certificates", stringListToSet([]string{*gsuite.Certificate}))
	} else if gsuite.Certificates != nil {
		d.Set("certificates", stringListToSet(*gsuite.Certificates))
	} else {
		d.Set("certificates", nil)
	}

	if gsuite.IssuerURI != nil {
		d.Set("issuer_uri", *gsuite.IssuerURI)
	} else {
		d.Set("issuer_uri", nil)
	}

	if gsuite.SsoTargetURI != nil {
		d.Set("target_uri", *gsuite.SsoTargetURI)
	} else {
		d.Set("target_uri", nil)
	}

	if gsuite.RelyingPartyIdentifier != nil {
		d.Set("relying_party_identifier", *gsuite.RelyingPartyIdentifier)
	} else {
		d.Set("relying_party_identifier", nil)
	}

	if gsuite.Disabled != nil {
		d.Set("disabled", *gsuite.Disabled)
	} else {
		d.Set("disabled", nil)
	}

	log.Printf("Read IDP GSuite")
	return nil
}

func updateIdpGsuite(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	issuerUri := d.Get("issuer_uri").(string)
	targetUri := d.Get("target_uri").(string)
	relyingPartyID := d.Get("relying_party_identifier").(string)
	disabled := d.Get("disabled").(bool)

	sdkConfig := meta.(*providerMeta).ClientConfig
	idpAPI := platformclientv2.NewIdentityProviderApiWithConfig(sdkConfig)

	log.Printf("Updating IDP GSuite")
	update := platformclientv2.Gsuite{
		IssuerURI:              &issuerUri,
		SsoTargetURI:           &targetUri,
		RelyingPartyIdentifier: &relyingPartyID,
		Disabled:               &disabled,
	}

	certificates := buildSdkStringList(d, "certificates")
	if certificates != nil {
		if len(*certificates) == 1 {
			update.Certificate = &(*certificates)[0]
		} else {
			update.Certificates = certificates
		}
	}

	_, _, err := idpAPI.PutIdentityprovidersGsuite(update)
	if err != nil {
		return diag.Errorf("Failed to update IDP GSuite: %s", err)
	}

	log.Printf("Updated IDP GSuite")
	time.Sleep(2 * time.Second)
	return readIdpGsuite(ctx, d, meta)
}

func deleteIdpGsuite(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*providerMeta).ClientConfig
	idpAPI := platformclientv2.NewIdentityProviderApiWithConfig(sdkConfig)

	log.Printf("Deleting IDP GSuite")
	_, _, err := idpAPI.DeleteIdentityprovidersGsuite()
	if err != nil {
		return diag.Errorf("Failed to delete IDP GSuite: %s", err)
	}
	log.Printf("Deleted IDP GSuite")
	return nil
}
