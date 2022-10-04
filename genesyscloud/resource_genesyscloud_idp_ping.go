package genesyscloud

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/consistency_checker"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v80/platformclientv2"
)

func getAllIdpPing(_ context.Context, clientConfig *platformclientv2.Configuration) (ResourceIDMetaMap, diag.Diagnostics) {
	idpAPI := platformclientv2.NewIdentityProviderApiWithConfig(clientConfig)
	resources := make(ResourceIDMetaMap)

	_, resp, getErr := idpAPI.GetIdentityprovidersPing()
	if getErr != nil {
		if isStatus404(resp) {
			// Don't export if config doesn't exist
			return resources, nil
		}
		return nil, diag.Errorf("Failed to get IDP Ping: %v", getErr)
	}

	resources["0"] = &ResourceMeta{Name: "ping"}
	return resources, nil
}

func idpPingExporter() *ResourceExporter {
	return &ResourceExporter{
		GetResourcesFunc: getAllWithPooledClient(getAllIdpPing),
		RefAttrs:         map[string]*RefAttrSettings{}, // No references
	}
}

func resourceIdpPing() *schema.Resource {
	return &schema.Resource{
		Description: "Genesys Cloud Single Sign-on Ping Identity Provider. See this page for detailed configuration instructions: https://help.mypurecloud.com/articles/add-ping-identity-single-sign-provider/",

		CreateContext: createWithPooledClient(createIdpPing),
		ReadContext:   readWithPooledClient(readIdpPing),
		UpdateContext: updateWithPooledClient(updateIdpPing),
		DeleteContext: deleteWithPooledClient(deleteIdpPing),
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
	sdkConfig := meta.(*providerMeta).ClientConfig
	idpAPI := platformclientv2.NewIdentityProviderApiWithConfig(sdkConfig)

	log.Printf("Reading IDP Ping")

	return withRetriesForReadCustomTimeout(ctx, d.Timeout(schema.TimeoutRead), d, func() *resource.RetryError {
		ping, resp, getErr := idpAPI.GetIdentityprovidersPing()
		if getErr != nil {
			if isStatus404(resp) {
				createIdpPing(ctx, d, meta)
				return resource.RetryableError(fmt.Errorf("Failed to read IDP Ping: %s", getErr))
			}
			return resource.NonRetryableError(fmt.Errorf("Failed to read IDP Ping: %s", getErr))
		}

		cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, resourceIdpPing())
		if ping.Certificate != nil {
			d.Set("certificates", stringListToInterfaceList([]string{*ping.Certificate}))
		} else if ping.Certificates != nil {
			d.Set("certificates", stringListToInterfaceList(*ping.Certificates))
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
		return cc.CheckState()
	})
}

func updateIdpPing(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	issuerUri := d.Get("issuer_uri").(string)
	targetUri := d.Get("target_uri").(string)
	relyingPartyID := d.Get("relying_party_identifier").(string)
	disabled := d.Get("disabled").(bool)

	sdkConfig := meta.(*providerMeta).ClientConfig
	idpAPI := platformclientv2.NewIdentityProviderApiWithConfig(sdkConfig)

	log.Printf("Updating IDP Ping")
	update := platformclientv2.Pingidentity{
		IssuerURI:              &issuerUri,
		SsoTargetURI:           &targetUri,
		RelyingPartyIdentifier: &relyingPartyID,
		Disabled:               &disabled,
	}

	certificates := buildSdkStringListFromInterfaceArray(d, "certificates")
	if certificates != nil {
		if len(*certificates) == 1 {
			update.Certificate = &(*certificates)[0]
		} else {
			update.Certificates = certificates
		}
	}

	_, _, err := idpAPI.PutIdentityprovidersPing(update)
	if err != nil {
		return diag.Errorf("Failed to update IDP Ping: %s", err)
	}

	log.Printf("Updated IDP Ping")
	return readIdpPing(ctx, d, meta)
}

func deleteIdpPing(ctx context.Context, _ *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*providerMeta).ClientConfig
	idpAPI := platformclientv2.NewIdentityProviderApiWithConfig(sdkConfig)

	log.Printf("Deleting IDP Ping")
	_, _, err := idpAPI.DeleteIdentityprovidersPing()
	if err != nil {
		return diag.Errorf("Failed to delete IDP Ping: %s", err)
	}

	return withRetries(ctx, 60*time.Second, func() *resource.RetryError {
		_, resp, err := idpAPI.GetIdentityprovidersPing()
		if err != nil {
			if isStatus404(resp) {
				// IDP Ping deleted
				log.Printf("Deleted IDP Ping")
				return nil
			}
			return resource.NonRetryableError(fmt.Errorf("Error deleting IDP Ping: %s", err))
		}
		return resource.RetryableError(fmt.Errorf("IDP Ping still exists"))
	})
}
