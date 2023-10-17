package genesyscloud

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"terraform-provider-genesyscloud/genesyscloud/consistency_checker"

	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	lists "terraform-provider-genesyscloud/genesyscloud/util/lists"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v115/platformclientv2"
)

func getAllIdpOkta(_ context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	idpAPI := platformclientv2.NewIdentityProviderApiWithConfig(clientConfig)
	resources := make(resourceExporter.ResourceIDMetaMap)

	_, resp, getErr := idpAPI.GetIdentityprovidersOkta()
	if getErr != nil {
		if IsStatus404(resp) {
			// Don't export if config doesn't exist
			return resources, nil
		}
		return nil, diag.Errorf("Failed to get IDP Okta: %v", getErr)
	}

	resources["0"] = &resourceExporter.ResourceMeta{Name: "okta"}
	return resources, nil
}

func IdpOktaExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: GetAllWithPooledClient(getAllIdpOkta),
		RefAttrs:         map[string]*resourceExporter.RefAttrSettings{}, // No references
	}
}

func ResourceIdpOkta() *schema.Resource {
	return &schema.Resource{
		Description: "Genesys Cloud Single Sign-on Okta Identity Provider. See this page for detailed configuration instructions: https://help.mypurecloud.com/articles/add-okta-as-a-single-sign-on-provider/",

		CreateContext: CreateWithPooledClient(createIdpOkta),
		ReadContext:   ReadWithPooledClient(readIdpOkta),
		UpdateContext: UpdateWithPooledClient(updateIdpOkta),
		DeleteContext: DeleteWithPooledClient(deleteIdpOkta),
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
	sdkConfig := meta.(*ProviderMeta).ClientConfig
	idpAPI := platformclientv2.NewIdentityProviderApiWithConfig(sdkConfig)

	log.Printf("Reading IDP Okta")

	return WithRetriesForReadCustomTimeout(ctx, d.Timeout(schema.TimeoutRead), d, func() *retry.RetryError {
		okta, resp, getErr := idpAPI.GetIdentityprovidersOkta()
		if getErr != nil {
			if IsStatus404(resp) {
				createIdpOkta(ctx, d, meta)
				return retry.RetryableError(fmt.Errorf("Failed to read IDP Okta: %s", getErr))
			}
			return retry.NonRetryableError(fmt.Errorf("Failed to read IDP Okta: %s", getErr))
		}

		cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceIdpOkta())
		if okta.Certificate != nil {
			d.Set("certificates", lists.StringListToInterfaceList([]string{*okta.Certificate}))
		} else if okta.Certificates != nil {
			d.Set("certificates", lists.StringListToInterfaceList(*okta.Certificates))
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
		return cc.CheckState()
	})
}

func updateIdpOkta(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	issuerUri := d.Get("issuer_uri").(string)
	targetUri := d.Get("target_uri").(string)
	disabled := d.Get("disabled").(bool)

	sdkConfig := meta.(*ProviderMeta).ClientConfig
	idpAPI := platformclientv2.NewIdentityProviderApiWithConfig(sdkConfig)

	log.Printf("Updating IDP Okta")
	update := platformclientv2.Okta{
		IssuerURI:    &issuerUri,
		SsoTargetURI: &targetUri,
		Disabled:     &disabled,
	}

	certificates := lists.BuildSdkStringListFromInterfaceArray(d, "certificates")
	if certificates != nil {
		if len(*certificates) == 1 {
			update.Certificate = &(*certificates)[0]
		}
		update.Certificates = certificates
	}

	_, _, err := idpAPI.PutIdentityprovidersOkta(update)
	if err != nil {
		return diag.Errorf("Failed to update IDP Okta: %s", err)
	}

	log.Printf("Updated IDP Okta")
	return readIdpOkta(ctx, d, meta)
}

func deleteIdpOkta(ctx context.Context, _ *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*ProviderMeta).ClientConfig
	idpAPI := platformclientv2.NewIdentityProviderApiWithConfig(sdkConfig)

	log.Printf("Deleting IDP Okta")
	_, _, err := idpAPI.DeleteIdentityprovidersOkta()
	if err != nil {
		return diag.Errorf("Failed to delete IDP Okta: %s", err)
	}

	return WithRetries(ctx, 60*time.Second, func() *retry.RetryError {
		_, resp, err := idpAPI.GetIdentityprovidersOkta()
		if err != nil {
			if IsStatus404(resp) {
				// IDP Okta deleted
				log.Printf("Deleted IDP Okta")
				return nil
			}
			return retry.NonRetryableError(fmt.Errorf("Error deleting IDP Okta: %s", err))
		}
		return retry.RetryableError(fmt.Errorf("IDP Okta still exists"))
	})
}
