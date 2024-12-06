package idp_onelogin

import (
	"context"
	"fmt"
	"log"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v146/platformclientv2"

	"terraform-provider-genesyscloud/genesyscloud/consistency_checker"

	"terraform-provider-genesyscloud/genesyscloud/util/constants"
	"terraform-provider-genesyscloud/genesyscloud/util/lists"
	"terraform-provider-genesyscloud/genesyscloud/util/resourcedata"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
)

/*
The resource_genesyscloud_idp_onelogin.go contains all of the methods that perform the core logic for a resource.
*/

// getAllAuthIdpOnelogin retrieves all of the idp onelogin via Terraform in the Genesys Cloud and is used for the exporter
func getAllAuthIdpOnelogins(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	// Although this resource typically has only a single instance,
	// we are attempting to fetch the data from the API in order to
	// verify the user's permission to access this resource's API endpoint(s).

	proxy := getIdpOneloginProxy(clientConfig)
	resources := make(resourceExporter.ResourceIDMetaMap)

	_, resp, err := proxy.getIdpOnelogin(ctx)
	if err != nil {
		if util.IsStatus404(resp) {
			// Don't export if config doesn't exist
			return resources, nil
		}
		return nil, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to get IDP Onelogin error: %s", err), resp)
	}

	resources["0"] = &resourceExporter.ResourceMeta{BlockLabel: "onelogin"}
	return resources, nil
}

// createIdpOnelogin is used by the idp_onelogin resource to create Genesys cloud idp onelogin
func createIdpOnelogin(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Printf("Creating IDP Onelogin")
	d.SetId("onelogin")
	return updateIdpOnelogin(ctx, d, meta)
}

// readIdpOnelogin is used by the idp_onelogin resource to read an idp onelogin from genesys cloud
func readIdpOnelogin(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getIdpOneloginProxy(sdkConfig)

	log.Printf("Reading idp onelogin")

	return util.WithRetriesForReadCustomTimeout(ctx, d.Timeout(schema.TimeoutRead), d, func() *retry.RetryError {
		oneLogin, resp, getErr := proxy.getIdpOnelogin(ctx)
		if getErr != nil {
			if util.IsStatus404(resp) {
				createIdpOnelogin(ctx, d, meta)
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read IDP Onelogin: %s", getErr), resp))
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read IDP Onelogin: %s", getErr), resp))
		}

		cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceIdpOnelogin(), constants.ConsistencyChecks(), ResourceType)

		if oneLogin.Certificate != nil {
			d.Set("certificates", lists.StringListToInterfaceList([]string{*oneLogin.Certificate}))
		} else if oneLogin.Certificates != nil {
			d.Set("certificates", lists.StringListToInterfaceList(*oneLogin.Certificates))
		} else {
			d.Set("certificates", nil)
		}

		resourcedata.SetNillableValue(d, "name", oneLogin.Name)
		resourcedata.SetNillableValue(d, "disabled", oneLogin.Disabled)
		resourcedata.SetNillableValue(d, "issuer_uri", oneLogin.IssuerURI)
		resourcedata.SetNillableValue(d, "target_uri", oneLogin.SsoTargetURI)
		resourcedata.SetNillableValue(d, "slo_uri", oneLogin.SloURI)
		resourcedata.SetNillableValue(d, "slo_binding", oneLogin.SloBinding)
		resourcedata.SetNillableValue(d, "relying_party_identifier", oneLogin.RelyingPartyIdentifier)

		log.Printf("Read idp onelogin")
		return cc.CheckState(d)
	})
}

// updateIdpOnelogin is used by the idp_onelogin resource to update an idp onelogin in Genesys Cloud
func updateIdpOnelogin(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getIdpOneloginProxy(sdkConfig)

	idpOnelogin := getIdpOneloginFromResourceData(d)

	certificates := lists.BuildSdkStringListFromInterfaceArray(d, "certificates")
	if certificates != nil {
		if len(*certificates) == 1 {
			idpOnelogin.Certificate = &(*certificates)[0]
		}
		idpOnelogin.Certificates = certificates
	}

	log.Printf("Updating idp onelogin")
	_, resp, err := proxy.updateIdpOnelogin(ctx, d.Id(), &idpOnelogin)
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to update IDP Onelogin %s error: %s", d.Id(), err), resp)
	}

	log.Printf("Updated idp onelogin")
	return readIdpOnelogin(ctx, d, meta)
}

// deleteIdpOnelogin is used by the idp_onelogin resource to delete an idp onelogin from Genesys cloud
func deleteIdpOnelogin(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getIdpOneloginProxy(sdkConfig)

	log.Printf("Deleting IDP Onelogin")

	resp, err := proxy.deleteIdpOnelogin(ctx, d.Id())
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to updadeletete IDP Onelogin %s error: %s", d.Id(), err), resp)
	}

	return util.WithRetries(ctx, 60*time.Second, func() *retry.RetryError {
		_, resp, err := proxy.getIdpOnelogin(ctx)

		if err != nil {
			if util.IsStatus404(resp) {
				// IDP Onelogin deleted
				log.Printf("Deleted IDP Onelogin")
				return nil
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Error deleting IDP Onelogin: %s", err), resp))
		}
		return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("IDP Onelogin still exists"), resp))
	})
}

// getIdpOneloginFromResourceData maps data from schema ResourceData object to a platformclientv2.Onelogin
func getIdpOneloginFromResourceData(d *schema.ResourceData) platformclientv2.Onelogin {
	return platformclientv2.Onelogin{
		Name:                   platformclientv2.String(d.Get("name").(string)),
		Disabled:               platformclientv2.Bool(d.Get("disabled").(bool)),
		IssuerURI:              platformclientv2.String(d.Get("issuer_uri").(string)),
		SsoTargetURI:           platformclientv2.String(d.Get("target_uri").(string)),
		SloURI:                 platformclientv2.String(d.Get("slo_uri").(string)),
		SloBinding:             platformclientv2.String(d.Get("slo_binding").(string)),
		RelyingPartyIdentifier: platformclientv2.String(d.Get("relying_party_identifier").(string)),
	}
}
