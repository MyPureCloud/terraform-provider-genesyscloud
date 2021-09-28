package genesyscloud

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v53/platformclientv2"
)

func getAllIdpAdfs(ctx context.Context, clientConfig *platformclientv2.Configuration) (ResourceIDMetaMap, diag.Diagnostics) {
	idpAPI := platformclientv2.NewIdentityProviderApiWithConfig(clientConfig)
	resources := make(ResourceIDMetaMap)

	_, resp, getErr := idpAPI.GetIdentityprovidersAdfs()
	if getErr != nil {
		if resp != nil && resp.StatusCode == 404 {
			// Don't export if config doesn't exist
			return resources, nil
		}
		return nil, diag.Errorf("Failed to get IDP ADFS: %v", getErr)
	}

	resources["0"] = &ResourceMeta{Name: "adfs"}
	return resources, nil
}

func idpAdfsExporter() *ResourceExporter {
	return &ResourceExporter{
		GetResourcesFunc: getAllWithPooledClient(getAllIdpAdfs),
		RefAttrs:         map[string]*RefAttrSettings{}, // No references
	}
}

func resourceIdpAdfs() *schema.Resource {
	return &schema.Resource{
		Description: "Genesys Cloud Single Sign-on ADFS Identity Provider. See this page for detailed configuration instructions: https://help.mypurecloud.com/articles/add-microsoft-adfs-single-sign-provider/",

		CreateContext: createWithPooledClient(createIdpAdfs),
		ReadContext:   readWithPooledClient(readIdpAdfs),
		UpdateContext: updateWithPooledClient(updateIdpAdfs),
		DeleteContext: deleteWithPooledClient(deleteIdpAdfs),
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
	sdkConfig := meta.(*providerMeta).ClientConfig
	idpAPI := platformclientv2.NewIdentityProviderApiWithConfig(sdkConfig)

	log.Printf("Reading IDP ADFS")
	adfs, resp, getErr := idpAPI.GetIdentityprovidersAdfs()
	if getErr != nil {
		if resp != nil && resp.StatusCode == 404 {
			d.SetId("")
			return nil
		}
		return diag.Errorf("Failed to read IDP ADFS: %s", getErr)
	}

	if adfs.Certificate != nil {
		d.Set("certificates", stringListToSet([]string{*adfs.Certificate}))
	} else if adfs.Certificates != nil {
		d.Set("certificates", stringListToSet(*adfs.Certificates))
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
	return nil
}

func updateIdpAdfs(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	issuerUri := d.Get("issuer_uri").(string)
	targetUri := d.Get("target_uri").(string)
	relyingPartyID := d.Get("relying_party_identifier").(string)
	disabled := d.Get("disabled").(bool)

	sdkConfig := meta.(*providerMeta).ClientConfig
	idpAPI := platformclientv2.NewIdentityProviderApiWithConfig(sdkConfig)

	log.Printf("Updating IDP ADFS")
	update := platformclientv2.Adfs{
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

	_, _, err := idpAPI.PutIdentityprovidersAdfs(update)
	if err != nil {
		return diag.Errorf("Failed to update IDP ADFS: %s", err)
	}

	log.Printf("Updated IDP ADFS")
	time.Sleep(2 * time.Second)
	return readIdpAdfs(ctx, d, meta)
}

func deleteIdpAdfs(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*providerMeta).ClientConfig
	idpAPI := platformclientv2.NewIdentityProviderApiWithConfig(sdkConfig)

	log.Printf("Deleting IDP ADFS")
	_, _, err := idpAPI.DeleteIdentityprovidersAdfs()
	if err != nil {
		return diag.Errorf("Failed to delete IDP ADFS: %s", err)
	}

	return withRetries(ctx, 30*time.Second, func() *resource.RetryError {
		_, resp, err := idpAPI.GetIdentityprovidersAdfs()
		if err != nil {
			if resp != nil && resp.StatusCode == 404 {
				// IDP ADFS deleted
				log.Printf("Deleted IDP ADFS")
				return nil
			}
			return resource.NonRetryableError(fmt.Errorf("Error deleting IDP ADFS: %s", err))
		}
		return resource.RetryableError(fmt.Errorf("IDP ADFS still exists"))
	})
}
