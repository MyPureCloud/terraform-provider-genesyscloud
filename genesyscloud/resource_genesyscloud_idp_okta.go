package genesyscloud

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v56/platformclientv2"
)

func getAllIdpOkta(_ context.Context, clientConfig *platformclientv2.Configuration) (ResourceIDMetaMap, diag.Diagnostics) {
	idpAPI := platformclientv2.NewIdentityProviderApiWithConfig(clientConfig)
	resources := make(ResourceIDMetaMap)

	_, resp, getErr := idpAPI.GetIdentityprovidersOkta()
	if getErr != nil {
		if isStatus404(resp) {
			// Don't export if config doesn't exist
			return resources, nil
		}
		return nil, diag.Errorf("Failed to get IDP Okta: %v", getErr)
	}

	resources["0"] = &ResourceMeta{Name: "okta"}
	return resources, nil
}

func idpOktaExporter() *ResourceExporter {
	return &ResourceExporter{
		GetResourcesFunc: getAllWithPooledClient(getAllIdpOkta),
		RefAttrs:         map[string]*RefAttrSettings{}, // No references
	}
}

func resourceIdpOkta() *schema.Resource {
	return &schema.Resource{
		Description: "Genesys Cloud Single Sign-on Okta Identity Provider. See this page for detailed configuration instructions: https://help.mypurecloud.com/articles/add-okta-as-a-single-sign-on-provider/",

		CreateContext: createWithPooledClient(createIdpOkta),
		ReadContext:   readWithPooledClient(readIdpOkta),
		UpdateContext: updateWithPooledClient(updateIdpOkta),
		DeleteContext: deleteWithPooledClient(deleteIdpOkta),
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
				Type:        schema.TypeSet,
				Required:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"issuer_uri": {
				Description: "Issuer URI provided by Okta.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"target_uri": {
				Description: "Target URI provided by Okta.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"disabled": {
				Description: "True if Okta is disabled.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
			},
		},
	}
}

func createIdpOkta(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Printf("Creating IDP Okta")
	d.SetId("okta")
	return updateIdpOkta(ctx, d, meta)
}

func readIdpOkta(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*providerMeta).ClientConfig
	idpAPI := platformclientv2.NewIdentityProviderApiWithConfig(sdkConfig)

	log.Printf("Reading IDP Okta")

	return withRetriesForRead(ctx, d.Timeout(schema.TimeoutRead), d, func() *resource.RetryError {
		okta, resp, getErr := idpAPI.GetIdentityprovidersOkta()
		if getErr != nil {
			if isStatus404(resp) {
				return resource.RetryableError(fmt.Errorf("Failed to read IDP Okta: %s", getErr))
			}
			return resource.NonRetryableError(fmt.Errorf("Failed to read IDP Okta: %s", getErr))
		}

		if okta.Certificate != nil {
			d.Set("certificates", stringListToSet([]string{*okta.Certificate}))
		} else if okta.Certificates != nil {
			d.Set("certificates", stringListToSet(*okta.Certificates))
		} else {
			d.Set("certificates", nil)
		}

		if okta.IssuerURI != nil {
			d.Set("issuer_uri", *okta.IssuerURI)
		} else {
			d.Set("issuer_uri", nil)
		}

		if okta.SsoTargetURI != nil {
			d.Set("target_uri", *okta.SsoTargetURI)
		} else {
			d.Set("target_uri", nil)
		}

		if okta.Disabled != nil {
			d.Set("disabled", *okta.Disabled)
		} else {
			d.Set("disabled", nil)
		}

		log.Printf("Read IDP Okta")
		return nil
	})
}

func updateIdpOkta(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	issuerUri := d.Get("issuer_uri").(string)
	targetUri := d.Get("target_uri").(string)
	disabled := d.Get("disabled").(bool)

	sdkConfig := meta.(*providerMeta).ClientConfig
	idpAPI := platformclientv2.NewIdentityProviderApiWithConfig(sdkConfig)

	log.Printf("Updating IDP Okta")
	update := platformclientv2.Okta{
		IssuerURI:    &issuerUri,
		SsoTargetURI: &targetUri,
		Disabled:     &disabled,
	}

	certificates := buildSdkStringList(d, "certificates")
	if certificates != nil {
		if len(*certificates) == 1 {
			update.Certificate = &(*certificates)[0]
		} else {
			update.Certificates = certificates
		}
	}

	_, _, err := idpAPI.PutIdentityprovidersOkta(update)
	if err != nil {
		return diag.Errorf("Failed to update IDP Okta: %s", err)
	}

	log.Printf("Updated IDP Okta")
	// Give time for public API caches to update
	// It takes a very very long time with idp resources
	time.Sleep(d.Timeout(schema.TimeoutUpdate))
	return readIdpOkta(ctx, d, meta)
}

func deleteIdpOkta(ctx context.Context, _ *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*providerMeta).ClientConfig
	idpAPI := platformclientv2.NewIdentityProviderApiWithConfig(sdkConfig)

	log.Printf("Deleting IDP Okta")
	_, _, err := idpAPI.DeleteIdentityprovidersOkta()
	if err != nil {
		return diag.Errorf("Failed to delete IDP Okta: %s", err)
	}

	return withRetries(ctx, 60*time.Second, func() *resource.RetryError {
		_, resp, err := idpAPI.GetIdentityprovidersOkta()
		if err != nil {
			if isStatus404(resp) {
				// IDP Okta deleted
				log.Printf("Deleted IDP Okta")
				return nil
			}
			return resource.NonRetryableError(fmt.Errorf("Error deleting IDP Okta: %s", err))
		}
		return resource.RetryableError(fmt.Errorf("IDP Okta still exists"))
	})
}
