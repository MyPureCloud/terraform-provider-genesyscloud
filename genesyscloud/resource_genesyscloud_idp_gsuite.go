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

func getAllIdpGsuite(_ context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	idpAPI := platformclientv2.NewIdentityProviderApiWithConfig(clientConfig)
	resources := make(resourceExporter.ResourceIDMetaMap)

	_, resp, getErr := idpAPI.GetIdentityprovidersGsuite()
	if getErr != nil {
		if IsStatus404(resp) {
			// Don't export if config doesn't exist
			return resources, nil
		}
		return nil, diag.Errorf("Failed to get IDP GSuite: %v", getErr)
	}

	resources["0"] = &resourceExporter.ResourceMeta{Name: "gsuite"}
	return resources, nil
}

func IdpGsuiteExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: GetAllWithPooledClient(getAllIdpGsuite),
		RefAttrs:         map[string]*resourceExporter.RefAttrSettings{}, // No references
	}
}

func ResourceIdpGsuite() *schema.Resource {
	return &schema.Resource{
		Description: "Genesys Cloud Single Sign-on GSuite Identity Provider. See this page for detailed configuration instructions: https://help.mypurecloud.com/articles/add-google-g-suite-single-sign-provider/",

		CreateContext: CreateWithPooledClient(createIdpGsuite),
		ReadContext:   ReadWithPooledClient(readIdpGsuite),
		UpdateContext: UpdateWithPooledClient(updateIdpGsuite),
		DeleteContext: DeleteWithPooledClient(deleteIdpGsuite),
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
	sdkConfig := meta.(*ProviderMeta).ClientConfig
	idpAPI := platformclientv2.NewIdentityProviderApiWithConfig(sdkConfig)

	log.Printf("Reading IDP GSuite")

	return WithRetriesForReadCustomTimeout(ctx, d.Timeout(schema.TimeoutRead), d, func() *retry.RetryError {
		gsuite, resp, getErr := idpAPI.GetIdentityprovidersGsuite()
		if getErr != nil {
			if IsStatus404(resp) {
				createIdpGsuite(ctx, d, meta)
				return retry.RetryableError(fmt.Errorf("Failed to read IDP GSuite: %s", getErr))
			}
			return retry.NonRetryableError(fmt.Errorf("Failed to read IDP GSuite: %s", getErr))
		}

		cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceIdpGsuite())
		if gsuite.Certificate != nil {
			d.Set("certificates", lists.StringListToInterfaceList([]string{*gsuite.Certificate}))
		} else if gsuite.Certificates != nil {
			d.Set("certificates", lists.StringListToInterfaceList(*gsuite.Certificates))
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
		return cc.CheckState()
	})
}

func updateIdpGsuite(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	issuerUri := d.Get("issuer_uri").(string)
	targetUri := d.Get("target_uri").(string)
	relyingPartyID := d.Get("relying_party_identifier").(string)
	disabled := d.Get("disabled").(bool)

	sdkConfig := meta.(*ProviderMeta).ClientConfig
	idpAPI := platformclientv2.NewIdentityProviderApiWithConfig(sdkConfig)

	log.Printf("Updating IDP GSuite")
	update := platformclientv2.Gsuite{
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

	_, _, err := idpAPI.PutIdentityprovidersGsuite(update)
	if err != nil {
		return diag.Errorf("Failed to update IDP GSuite: %s", err)
	}

	log.Printf("Updated IDP GSuite")
	return readIdpGsuite(ctx, d, meta)
}

func deleteIdpGsuite(ctx context.Context, _ *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*ProviderMeta).ClientConfig
	idpAPI := platformclientv2.NewIdentityProviderApiWithConfig(sdkConfig)

	log.Printf("Deleting IDP GSuite")
	_, _, err := idpAPI.DeleteIdentityprovidersGsuite()
	if err != nil {
		return diag.Errorf("Failed to delete IDP GSuite: %s", err)
	}

	return WithRetries(ctx, 60*time.Second, func() *retry.RetryError {
		_, resp, err := idpAPI.GetIdentityprovidersGsuite()
		if err != nil {
			if IsStatus404(resp) {
				// IDP GSuite deleted
				log.Printf("Deleted IDP GSuite")
				return nil
			}
			return retry.NonRetryableError(fmt.Errorf("Error deleting IDP GSuite: %s", err))
		}
		return retry.RetryableError(fmt.Errorf("IDP GSuite still exists"))
	})
}
