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

func getAllIdpSalesforce(_ context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	idpAPI := platformclientv2.NewIdentityProviderApiWithConfig(clientConfig)
	resources := make(resourceExporter.ResourceIDMetaMap)

	_, resp, getErr := idpAPI.GetIdentityprovidersSalesforce()
	if getErr != nil {
		if IsStatus404(resp) {
			// Don't export if config doesn't exist
			return resources, nil
		}
		return nil, diag.Errorf("Failed to get IDP Salesforce: %v", getErr)
	}

	resources["0"] = &resourceExporter.ResourceMeta{Name: "salesforce"}
	return resources, nil
}

func IdpSalesforceExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: GetAllWithPooledClient(getAllIdpSalesforce),
		RefAttrs:         map[string]*resourceExporter.RefAttrSettings{}, // No references
	}
}

func ResourceIdpSalesforce() *schema.Resource {
	return &schema.Resource{
		Description: "Genesys Cloud Single Sign-on Salesforce Identity Provider. See this page for detailed configuration instructions: https://help.mypurecloud.com/articles/add-salesforce-as-a-single-sign-on-provider/",

		CreateContext: CreateWithPooledClient(createIdpSalesforce),
		ReadContext:   ReadWithPooledClient(readIdpSalesforce),
		UpdateContext: UpdateWithPooledClient(updateIdpSalesforce),
		DeleteContext: DeleteWithPooledClient(deleteIdpSalesforce),
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
				Description: "Issuer URI provided by Salesforce.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"target_uri": {
				Description: "Target URI provided by Salesforce.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"disabled": {
				Description: "True if Salesforce is disabled.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
			},
		},
	}
}

func createIdpSalesforce(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Printf("Creating IDP Salesforce")
	d.SetId("salesforce")
	return updateIdpSalesforce(ctx, d, meta)
}

func readIdpSalesforce(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*ProviderMeta).ClientConfig
	idpAPI := platformclientv2.NewIdentityProviderApiWithConfig(sdkConfig)

	log.Printf("Reading IDP Salesforce")

	return WithRetriesForReadCustomTimeout(ctx, d.Timeout(schema.TimeoutRead), d, func() *retry.RetryError {
		salesforce, resp, getErr := idpAPI.GetIdentityprovidersSalesforce()
		if getErr != nil {
			if IsStatus404(resp) {
				createIdpSalesforce(ctx, d, meta)
				return retry.RetryableError(fmt.Errorf("Failed to read IDP Salesforce: %s", getErr))
			}
			return retry.NonRetryableError(fmt.Errorf("Failed to read IDP Salesforce: %s", getErr))
		}

		cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceIdpSalesforce())
		if salesforce.Certificate != nil {
			d.Set("certificates", lists.StringListToInterfaceList([]string{*salesforce.Certificate}))
		} else if salesforce.Certificates != nil {
			d.Set("certificates", lists.StringListToInterfaceList(*salesforce.Certificates))
		} else {
			d.Set("certificates", nil)
		}

		if salesforce.IssuerURI != nil {
			d.Set("issuer_uri", *salesforce.IssuerURI)
		} else {
			d.Set("issuer_uri", nil)
		}

		if salesforce.SsoTargetURI != nil {
			d.Set("target_uri", *salesforce.SsoTargetURI)
		} else {
			d.Set("target_uri", nil)
		}

		if salesforce.Disabled != nil {
			d.Set("disabled", *salesforce.Disabled)
		} else {
			d.Set("disabled", nil)
		}

		log.Printf("Read IDP Salesforce")
		return cc.CheckState()
	})
}

func updateIdpSalesforce(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	issuerUri := d.Get("issuer_uri").(string)
	targetUri := d.Get("target_uri").(string)
	disabled := d.Get("disabled").(bool)

	sdkConfig := meta.(*ProviderMeta).ClientConfig
	idpAPI := platformclientv2.NewIdentityProviderApiWithConfig(sdkConfig)

	log.Printf("Updating IDP Salesforce")
	update := platformclientv2.Salesforce{
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

	_, _, err := idpAPI.PutIdentityprovidersSalesforce(update)
	if err != nil {
		return diag.Errorf("Failed to update IDP Salesforce: %s", err)
	}

	log.Printf("Updated IDP Salesforce")
	return readIdpSalesforce(ctx, d, meta)
}

func deleteIdpSalesforce(ctx context.Context, _ *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*ProviderMeta).ClientConfig
	idpAPI := platformclientv2.NewIdentityProviderApiWithConfig(sdkConfig)

	log.Printf("Deleting IDP Salesforce")
	_, _, err := idpAPI.DeleteIdentityprovidersSalesforce()
	if err != nil {
		return diag.Errorf("Failed to delete IDP Salesforce: %s", err)
	}

	return WithRetries(ctx, 180*time.Second, func() *retry.RetryError {
		_, resp, err := idpAPI.GetIdentityprovidersSalesforce()
		if err != nil {
			if IsStatus404(resp) {
				// IDP Salesforce deleted
				log.Printf("Deleted Salesforce Ping")
				return nil
			}
			return retry.NonRetryableError(fmt.Errorf("Error deleting IDP Salesforce: %s", err))
		}
		return retry.RetryableError(fmt.Errorf("IDP Salesforce still exists"))
	})
}
