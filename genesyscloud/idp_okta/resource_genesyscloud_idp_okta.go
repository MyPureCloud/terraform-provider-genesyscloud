package idp_okta

import (
	"context"
	"fmt"
	"log"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"time"

	"terraform-provider-genesyscloud/genesyscloud/consistency_checker"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v146/platformclientv2"

	"terraform-provider-genesyscloud/genesyscloud/util/constants"
	"terraform-provider-genesyscloud/genesyscloud/util/lists"
	"terraform-provider-genesyscloud/genesyscloud/util/resourcedata"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
)

/*
The resource_genesyscloud_idp_okta.go contains all of the methods that perform the core logic for a resource.
*/

// getAllAuthIdpOkta retrieves all of the idp okta via Terraform in the Genesys Cloud and is used for the exporter
func getAllAuthIdpOktas(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	// Although this resource typically has only a single instance,
	// we are attempting to fetch the data from the API in order to
	// verify the user's permission to access this resource's API endpoint(s).

	proxy := getIdpOktaProxy(clientConfig)
	resources := make(resourceExporter.ResourceIDMetaMap)

	_, resp, err := proxy.getIdpOkta(ctx)
	if err != nil {
		if util.IsStatus404(resp) {
			// Don't export if config doesn't exist
			return resources, nil
		}
		return nil, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to get IDP okta error: %s", err), resp)
	}

	resources["0"] = &resourceExporter.ResourceMeta{BlockLabel: "okta"}
	return resources, nil
}

// createIdpOkta is used by the idp_okta resource to create Genesys cloud idp okta
func createIdpOkta(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Printf("Creating IDP Okta")
	d.SetId("okta")
	return updateIdpOkta(ctx, d, meta)
}

// readIdpOkta is used by the idp_okta resource to read an idp okta from genesys cloud
func readIdpOkta(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getIdpOktaProxy(sdkConfig)

	log.Printf("Reading idp okta %s", d.Id())

	return util.WithRetriesForReadCustomTimeout(ctx, d.Timeout(schema.TimeoutRead), d, func() *retry.RetryError {
		okta, resp, getErr := proxy.getIdpOkta(ctx)
		if getErr != nil {
			if util.IsStatus404(resp) {
				createIdpOkta(ctx, d, meta)
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read IDP Okta: %s", getErr), resp))
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read IDP Okta: %s", getErr), resp))
		}

		cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceIdpOkta(), constants.ConsistencyChecks(), "genesyscloud_idp_okta")

		resourcedata.SetNillableValue(d, "name", okta.Name)
		resourcedata.SetNillableValue(d, "disabled", okta.Disabled)
		resourcedata.SetNillableValue(d, "issuer_uri", okta.IssuerURI)
		resourcedata.SetNillableValue(d, "target_uri", okta.SsoTargetURI)
		resourcedata.SetNillableValue(d, "relying_party_identifier", okta.RelyingPartyIdentifier)
		resourcedata.SetNillableValue(d, "slo_uri", okta.SloURI)
		resourcedata.SetNillableValue(d, "slo_binding", okta.SloBinding)

		if okta.Certificate != nil {
			d.Set("certificates", lists.StringListToInterfaceList([]string{*okta.Certificate}))
		} else if okta.Certificates != nil {
			d.Set("certificates", lists.StringListToInterfaceList(*okta.Certificates))
		} else {
			d.Set("certificates", nil)
		}

		log.Printf("Read idp okta")
		return cc.CheckState(d)
	})
}

// updateIdpOkta is used by the idp_okta resource to update an idp okta in Genesys Cloud
func updateIdpOkta(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getIdpOktaProxy(sdkConfig)

	idpOkta := getIdpOktaFromResourceData(d)

	certificates := lists.BuildSdkStringListFromInterfaceArray(d, "certificates")
	if certificates != nil {
		if len(*certificates) == 1 {
			idpOkta.Certificate = &(*certificates)[0]
		}
		idpOkta.Certificates = certificates
	}

	log.Printf("Updating idp okta")
	_, resp, err := proxy.updateIdpOkta(ctx, d.Id(), &idpOkta)
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to update idp okta: %s", err), resp)
	}

	log.Printf("Updated idp okta")
	return readIdpOkta(ctx, d, meta)
}

// deleteIdpOkta is used by the idp_okta resource to delete an idp okta from Genesys cloud
func deleteIdpOkta(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getIdpOktaProxy(sdkConfig)

	resp, err := proxy.deleteIdpOkta(ctx, d.Id())
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to delete idp okta: %s", err), resp)
	}

	return util.WithRetries(ctx, 60*time.Second, func() *retry.RetryError {
		_, resp, err := proxy.getIdpOkta(ctx)

		if err != nil {
			if util.IsStatus404(resp) {
				// IDP Okta deleted
				log.Printf("Deleted IDP Okta")
				return nil
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Error deleting IDP Okta: %s", err), resp))
		}
		return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("IDP Okta still exists"), resp))
	})
}

// getIdpOktaFromResourceData maps data from schema ResourceData object to a platformclientv2.Okta
func getIdpOktaFromResourceData(d *schema.ResourceData) platformclientv2.Okta {
	return platformclientv2.Okta{
		Name:                   platformclientv2.String(d.Get("name").(string)),
		Disabled:               platformclientv2.Bool(d.Get("disabled").(bool)),
		IssuerURI:              platformclientv2.String(d.Get("issuer_uri").(string)),
		SsoTargetURI:           platformclientv2.String(d.Get("target_uri").(string)),
		RelyingPartyIdentifier: platformclientv2.String(d.Get("relying_party_identifier").(string)),
		SloURI:                 platformclientv2.String(d.Get("slo_uri").(string)),
		SloBinding:             platformclientv2.String(d.Get("slo_binding").(string)),
	}
}
