package idp_gsuite

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
The resource_genesyscloud_idp_gsuite.go contains all of the methods that perform the core logic for a resource.
*/

// getAllAuthIdpGsuite retrieves all of the idp gsuite via Terraform in the Genesys Cloud and is used for the exporter
func getAllAuthIdpGsuites(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	// Although this resource typically has only a single instance,
	// we are attempting to fetch the data from the API in order to
	// verify the user's permission to access this resource's API endpoint(s).

	proxy := getIdpGsuiteProxy(clientConfig)
	resources := make(resourceExporter.ResourceIDMetaMap)

	_, resp, err := proxy.getIdpGsuite(ctx)
	if err != nil {
		if util.IsStatus404(resp) {
			// Don't export if config doesn't exist
			return resources, nil
		}
		return nil, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to get IDP GSuite error: %s", err), resp)
	}

	resources["0"] = &resourceExporter.ResourceMeta{BlockLabel: "gsuite"}
	return resources, nil
}

// createIdpGsuite is used by the idp_gsuite resource to create Genesys cloud idp gsuite
func createIdpGsuite(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Printf("Creating IDP GSuite")
	d.SetId("gsuite")
	return updateIdpGsuite(ctx, d, meta)
}

// readIdpGsuite is used by the idp_gsuite resource to read an idp gsuite from genesys cloud
func readIdpGsuite(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getIdpGsuiteProxy(sdkConfig)

	log.Printf("Reading idp gsuite")

	cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceIdpGsuite(), constants.ConsistencyChecks(), ResourceType)

	return util.WithRetriesForReadCustomTimeout(ctx, d.Timeout(schema.TimeoutRead), d, func() *retry.RetryError {
		gSuite, resp, getErr := proxy.getIdpGsuite(ctx)
		if getErr != nil {
			if util.IsStatus404(resp) {
				createIdpGsuite(ctx, d, meta)
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read IDP GSuite: %s", getErr), resp))
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read IDP GSuite: %s", getErr), resp))
		}

		resourcedata.SetNillableValue(d, "name", gSuite.Name)
		resourcedata.SetNillableValue(d, "disabled", gSuite.Disabled)
		resourcedata.SetNillableValue(d, "issuer_uri", gSuite.IssuerURI)
		resourcedata.SetNillableValue(d, "target_uri", gSuite.SsoTargetURI)
		resourcedata.SetNillableValue(d, "relying_party_identifier", gSuite.RelyingPartyIdentifier)
		resourcedata.SetNillableValue(d, "slo_uri", gSuite.SloURI)
		resourcedata.SetNillableValue(d, "slo_binding", gSuite.SloBinding)

		if gSuite.Certificate != nil {
			d.Set("certificates", lists.StringListToInterfaceList([]string{*gSuite.Certificate}))
		} else if gSuite.Certificates != nil {
			d.Set("certificates", lists.StringListToInterfaceList(*gSuite.Certificates))
		} else {
			d.Set("certificates", nil)
		}

		log.Printf("Read idp gsuite")
		return cc.CheckState(d)
	})
}

// updateIdpGsuite is used by the idp_gsuite resource to update an idp gsuite in Genesys Cloud
func updateIdpGsuite(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getIdpGsuiteProxy(sdkConfig)

	idpGsuite := getIdpGsuiteFromResourceData(d)

	log.Printf("Updating idp gsuite")

	certificates := lists.BuildSdkStringListFromInterfaceArray(d, "certificates")
	if certificates != nil {
		if len(*certificates) == 1 {
			idpGsuite.Certificate = &(*certificates)[0]
		}
		idpGsuite.Certificates = certificates
	}

	_, resp, err := proxy.updateIdpGsuite(ctx, d.Id(), &idpGsuite)
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to update IDP GSuite %s error: %s", d.Id(), err), resp)
	}

	log.Printf("Updated idp gsuite")
	return readIdpGsuite(ctx, d, meta)
}

// deleteIdpGsuite is used by the idp_gsuite resource to delete an idp gsuite from Genesys cloud
func deleteIdpGsuite(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getIdpGsuiteProxy(sdkConfig)

	resp, err := proxy.deleteIdpGsuite(ctx, d.Id())
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to delete IDP GSuite %s error: %s", d.Id(), err), resp)
	}

	return util.WithRetries(ctx, 60*time.Second, func() *retry.RetryError {
		_, resp, err := proxy.getIdpGsuite(ctx)

		if err != nil {
			if util.IsStatus404(resp) {
				// IDP GSuite deleted
				log.Printf("Deleted IDP GSuite")
				return nil
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Error deleting IDP GSuite: %s", err), resp))
		}
		return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("IDP GSuite still exists"), resp))
	})
}

// getIdpGsuiteFromResourceData maps data from schema ResourceData object to a platformclientv2.Gsuite
func getIdpGsuiteFromResourceData(d *schema.ResourceData) platformclientv2.Gsuite {
	return platformclientv2.Gsuite{
		Name:                   platformclientv2.String(d.Get("name").(string)),
		Disabled:               platformclientv2.Bool(d.Get("disabled").(bool)),
		IssuerURI:              platformclientv2.String(d.Get("issuer_uri").(string)),
		SsoTargetURI:           platformclientv2.String(d.Get("target_uri").(string)),
		RelyingPartyIdentifier: platformclientv2.String(d.Get("relying_party_identifier").(string)),
		SloURI:                 platformclientv2.String(d.Get("slo_uri").(string)),
		SloBinding:             platformclientv2.String(d.Get("slo_binding").(string)),
	}
}
