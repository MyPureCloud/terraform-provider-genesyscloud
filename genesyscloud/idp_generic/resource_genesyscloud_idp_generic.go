package idp_generic

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
The resource_genesyscloud_idp_generic.go contains all of the methods that perform the core logic for a resource.
*/

// getAllAuthIdpGeneric retrieves all of the idp generic via Terraform in the Genesys Cloud and is used for the exporter
func getAllAuthIdpGenerics(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	// Although this resource typically has only a single instance,
	// we are attempting to fetch the data from the API in order to
	// verify the user's permission to access this resource's API endpoint(s).

	proxy := getIdpGenericProxy(clientConfig)
	resources := make(resourceExporter.ResourceIDMetaMap)

	_, resp, getErr := proxy.getIdpGeneric(ctx)
	if getErr != nil {
		if util.IsStatus404(resp) {
			// Don't export if config doesn't exist
			return resources, nil
		}
		return nil, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to get IDP Generic error: %s", getErr), resp)
	}

	resources["0"] = &resourceExporter.ResourceMeta{BlockLabel: "generic"}
	return resources, nil
}

// createIdpGeneric is used by the idp_generic resource to create Genesys cloud idp generic
func createIdpGeneric(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Printf("Creating IDP Generic")
	d.SetId("generic")
	return updateIdpGeneric(ctx, d, meta)
}

// readIdpGeneric is used by the idp_generic resource to read an idp generic from genesys cloud
func readIdpGeneric(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getIdpGenericProxy(sdkConfig)

	log.Printf("Reading idp generic %s", d.Id())

	cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceIdpGeneric(), constants.ConsistencyChecks(), ResourceType)

	return util.WithRetriesForReadCustomTimeout(ctx, d.Timeout(schema.TimeoutRead), d, func() *retry.RetryError {
		genericSAML, resp, getErr := proxy.getIdpGeneric(ctx)
		if getErr != nil {
			if util.IsStatus404(resp) {
				createIdpGeneric(ctx, d, meta)
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read IDP Generic: %s", getErr), resp))
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read IDP Generic: %s", getErr), resp))
		}

		resourcedata.SetNillableValue(d, "name", genericSAML.Name)
		resourcedata.SetNillableValue(d, "disabled", genericSAML.Disabled)
		resourcedata.SetNillableValue(d, "issuer_uri", genericSAML.IssuerURI)
		resourcedata.SetNillableValue(d, "target_uri", genericSAML.SsoTargetURI)
		resourcedata.SetNillableValue(d, "slo_uri", genericSAML.SloURI)
		resourcedata.SetNillableValue(d, "slo_binding", genericSAML.SloBinding)
		resourcedata.SetNillableValue(d, "relying_party_identifier", genericSAML.RelyingPartyIdentifier)
		resourcedata.SetNillableValue(d, "logo_image_data", genericSAML.LogoImageData)
		resourcedata.SetNillableValue(d, "endpoint_compression", genericSAML.EndpointCompression)
		resourcedata.SetNillableValue(d, "name_identifier_format", genericSAML.NameIdentifierFormat)

		if genericSAML.Certificate != nil {
			d.Set("certificates", lists.StringListToInterfaceList([]string{*genericSAML.Certificate}))
		} else if genericSAML.Certificates != nil {
			d.Set("certificates", lists.StringListToInterfaceList(*genericSAML.Certificates))
		} else {
			d.Set("certificates", nil)
		}

		log.Printf("Read idp generic %s %s", d.Id(), *genericSAML.Name)
		return cc.CheckState(d)
	})
}

// updateIdpGeneric is used by the idp_generic resource to update an idp generic in Genesys Cloud
func updateIdpGeneric(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getIdpGenericProxy(sdkConfig)

	idpGeneric := getIdpGenericFromResourceData(d)

	log.Printf("Updating idp generic %s", *idpGeneric.Name)

	certificates := lists.BuildSdkStringListFromInterfaceArray(d, "certificates")
	if certificates != nil {
		if len(*certificates) == 1 {
			idpGeneric.Certificate = &(*certificates)[0]
		}
		idpGeneric.Certificates = certificates
	}

	_, resp, err := proxy.updateIdpGeneric(ctx, d.Id(), &idpGeneric)
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to update IDP Generic %s error: %s", d.Id(), err), resp)
	}

	log.Printf("Updated idp generic")
	return readIdpGeneric(ctx, d, meta)
}

// deleteIdpGeneric is used by the idp_generic resource to delete an idp generic from Genesys cloud
func deleteIdpGeneric(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getIdpGenericProxy(sdkConfig)

	resp, err := proxy.deleteIdpGeneric(ctx, d.Id())
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to delete IDP Generic %s error: %s", d.Id(), err), resp)
	}

	return util.WithRetries(ctx, 60*time.Second, func() *retry.RetryError {
		_, resp, err := proxy.getIdpGeneric(ctx)

		if err != nil {
			if util.IsStatus404(resp) {
				// IDP Generic deleted
				log.Printf("Deleted IDP Generic")
				return nil
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Error deleting IDP Generic: %s", err), resp))
		}
		return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("IDP Generic still exists"), resp))
	})
}

// getIdpGenericFromResourceData maps data from schema ResourceData object to a platformclientv2.Genericsaml
func getIdpGenericFromResourceData(d *schema.ResourceData) platformclientv2.Genericsaml {
	return platformclientv2.Genericsaml{
		Name:                   platformclientv2.String(d.Get("name").(string)),
		Disabled:               platformclientv2.Bool(d.Get("disabled").(bool)),
		IssuerURI:              platformclientv2.String(d.Get("issuer_uri").(string)),
		SsoTargetURI:           platformclientv2.String(d.Get("target_uri").(string)),
		RelyingPartyIdentifier: platformclientv2.String(d.Get("relying_party_identifier").(string)),
		LogoImageData:          platformclientv2.String(d.Get("logo_image_data").(string)),
		EndpointCompression:    platformclientv2.Bool(d.Get("endpoint_compression").(bool)),
		NameIdentifierFormat:   platformclientv2.String(d.Get("name_identifier_format").(string)),
		SloURI:                 platformclientv2.String(d.Get("slo_uri").(string)),
		SloBinding:             platformclientv2.String(d.Get("slo_binding").(string)),
	}
}
