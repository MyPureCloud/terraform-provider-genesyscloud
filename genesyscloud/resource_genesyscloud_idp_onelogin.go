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

func getAllIdpOnelogin(_ context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	idpAPI := platformclientv2.NewIdentityProviderApiWithConfig(clientConfig)
	resources := make(resourceExporter.ResourceIDMetaMap)

	_, resp, getErr := idpAPI.GetIdentityprovidersOnelogin()
	if getErr != nil {
		if IsStatus404(resp) {
			// Don't export if config doesn't exist
			return resources, nil
		}
		return nil, diag.Errorf("Failed to get IDP Onelogin: %v", getErr)
	}

	resources["0"] = &resourceExporter.ResourceMeta{Name: "onelogin"}
	return resources, nil
}

func IdpOneloginExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: GetAllWithPooledClient(getAllIdpOnelogin),
		RefAttrs:         map[string]*resourceExporter.RefAttrSettings{}, // No references
	}
}

func ResourceIdpOnelogin() *schema.Resource {
	return &schema.Resource{
		Description: "Genesys Cloud Single Sign-on OneLogin Identity Provider. See this page for detailed configuration instructions: https://help.mypurecloud.com/articles/add-onelogin-as-single-sign-on-provider/",

		CreateContext: CreateWithPooledClient(createIdpOnelogin),
		ReadContext:   ReadWithPooledClient(readIdpOnelogin),
		UpdateContext: UpdateWithPooledClient(updateIdpOnelogin),
		DeleteContext: DeleteWithPooledClient(deleteIdpOnelogin),
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
				Description: "Issuer URI provided by OneLogin.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"target_uri": {
				Description: "Target URI provided by OneLogin.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"disabled": {
				Description: "True if OneLogin is disabled.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
			},
		},
	}
}

func createIdpOnelogin(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Printf("Creating IDP Onelogin")
	d.SetId("onelogin")
	return updateIdpOnelogin(ctx, d, meta)
}

func readIdpOnelogin(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*ProviderMeta).ClientConfig
	idpAPI := platformclientv2.NewIdentityProviderApiWithConfig(sdkConfig)

	log.Printf("Reading IDP Onelogin")

	return WithRetriesForReadCustomTimeout(ctx, d.Timeout(schema.TimeoutRead), d, func() *retry.RetryError {
		onelogin, resp, getErr := idpAPI.GetIdentityprovidersOnelogin()
		if getErr != nil {
			if IsStatus404(resp) {
				createIdpOkta(ctx, d, meta)
				return retry.RetryableError(fmt.Errorf("Failed to read IDP Onelogin: %s", getErr))
			}
			return retry.NonRetryableError(fmt.Errorf("Failed to read IDP Onelogin: %s", getErr))
		}

		cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceIdpOnelogin())
		if onelogin.Certificate != nil {
			d.Set("certificates", lists.StringListToInterfaceList([]string{*onelogin.Certificate}))
		} else if onelogin.Certificates != nil {
			d.Set("certificates", lists.StringListToInterfaceList(*onelogin.Certificates))
		} else {
			d.Set("certificates", nil)
		}

		if onelogin.IssuerURI != nil {
			d.Set("issuer_uri", *onelogin.IssuerURI)
		} else {
			d.Set("issuer_uri", nil)
		}

		if onelogin.SsoTargetURI != nil {
			d.Set("target_uri", *onelogin.SsoTargetURI)
		} else {
			d.Set("target_uri", nil)
		}

		if onelogin.Disabled != nil {
			d.Set("disabled", *onelogin.Disabled)
		} else {
			d.Set("disabled", nil)
		}

		log.Printf("Read IDP Onelogin")
		return cc.CheckState()
	})
}

func updateIdpOnelogin(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	issuerUri := d.Get("issuer_uri").(string)
	targetUri := d.Get("target_uri").(string)
	disabled := d.Get("disabled").(bool)

	sdkConfig := meta.(*ProviderMeta).ClientConfig
	idpAPI := platformclientv2.NewIdentityProviderApiWithConfig(sdkConfig)

	log.Printf("Updating IDP Onelogin")
	update := platformclientv2.Onelogin{
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

	_, _, err := idpAPI.PutIdentityprovidersOnelogin(update)
	if err != nil {
		return diag.Errorf("Failed to update IDP Onelogin: %s", err)
	}

	log.Printf("Updated IDP Onelogin")
	return readIdpOnelogin(ctx, d, meta)
}

func deleteIdpOnelogin(ctx context.Context, _ *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*ProviderMeta).ClientConfig
	idpAPI := platformclientv2.NewIdentityProviderApiWithConfig(sdkConfig)

	log.Printf("Deleting IDP Onelogin")
	_, _, err := idpAPI.DeleteIdentityprovidersOnelogin()
	if err != nil {
		return diag.Errorf("Failed to delete IDP Onelogin: %s", err)
	}

	return WithRetries(ctx, 60*time.Second, func() *retry.RetryError {
		_, resp, err := idpAPI.GetIdentityprovidersOnelogin()
		if err != nil {
			if IsStatus404(resp) {
				// IDP Onelogin deleted
				log.Printf("Deleted IDP Onelogin")
				return nil
			}
			return retry.NonRetryableError(fmt.Errorf("Error deleting IDP Onelogin: %s", err))
		}
		return retry.RetryableError(fmt.Errorf("IDP Onelogin still exists"))
	})
}
