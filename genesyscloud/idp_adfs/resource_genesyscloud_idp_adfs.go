package idp_adfs

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
The resource_genesyscloud_idp_adfs.go contains all of the methods that perform the core logic for a resource.
*/

// getAllAuthIdpAdfss retrieves all of the idp adfs via Terraform in the Genesys Cloud and is used for the exporter
func getAllAuthIdpAdfss(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	// Although this resource typically has only a single instance,
	// we are attempting to fetch the data from the API in order to
	// verify the user's permission to access this resource's API endpoint(s).

	proxy := getIdpAdfsProxy(clientConfig)
	resources := make(resourceExporter.ResourceIDMetaMap)

	_, resp, err := proxy.getIdpAdfs(ctx)
	if err != nil {
		if util.IsStatus404(resp) {
			// Don't export if config doesn't exist
			return resources, nil
		}
		return nil, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to get IDP ADFS error: %s", err), resp)
	}
	resources["0"] = &resourceExporter.ResourceMeta{BlockLabel: "adfs"}
	return resources, nil
}

// createIdpAdfs is used by the idp_adfs resource to create Genesys cloud idp adfs
func createIdpAdfs(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Printf("Creating IDP ADFS")
	d.SetId("adfs")
	return updateIdpAdfs(ctx, d, meta)
}

// readIdpAdfs is used by the idp_adfs resource to read an idp adfs from genesys cloud
func readIdpAdfs(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getIdpAdfsProxy(sdkConfig)

	log.Printf("Reading idp adfs %s", d.Id())

	cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceIdpAdfs(), constants.ConsistencyChecks(), ResourceType)

	return util.WithRetriesForReadCustomTimeout(ctx, d.Timeout(schema.TimeoutRead), d, func() *retry.RetryError {
		aDFS, resp, getErr := proxy.getIdpAdfs(ctx)
		if getErr != nil {
			if util.IsStatus404(resp) {
				createIdpAdfs(ctx, d, meta)
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read IDP ADFS: %s", getErr), resp))
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read IDP ADFS: %s", getErr), resp))
		}

		resourcedata.SetNillableValue(d, "name", aDFS.Name)
		resourcedata.SetNillableValue(d, "disabled", aDFS.Disabled)
		resourcedata.SetNillableValue(d, "issuer_uri", aDFS.IssuerURI)
		resourcedata.SetNillableValue(d, "target_uri", aDFS.SsoTargetURI)
		resourcedata.SetNillableValue(d, "relying_party_identifier", aDFS.RelyingPartyIdentifier)
		resourcedata.SetNillableValue(d, "slo_uri", aDFS.SloURI)
		resourcedata.SetNillableValue(d, "slo_binding", aDFS.SloBinding)

		if aDFS.Certificate != nil {
			d.Set("certificates", lists.StringListToInterfaceList([]string{*aDFS.Certificate}))
		} else if aDFS.Certificates != nil {
			d.Set("certificates", lists.StringListToInterfaceList(*aDFS.Certificates))
		} else {
			d.Set("certificates", nil)
		}

		log.Printf("Read idp adfs")
		return cc.CheckState(d)
	})
}

// updateIdpAdfs is used by the idp_adfs resource to update an idp adfs in Genesys Cloud
func updateIdpAdfs(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getIdpAdfsProxy(sdkConfig)

	idpAdfs := getIdpAdfsFromResourceData(d)
	certificates := lists.BuildSdkStringListFromInterfaceArray(d, "certificates")
	if certificates != nil {
		if len(*certificates) == 1 {
			idpAdfs.Certificate = &(*certificates)[0]
		}
		idpAdfs.Certificates = certificates
	}
	log.Printf("Updating idp adfs")
	resp, err := proxy.updateIdpAdfs(ctx, d.Id(), &idpAdfs)
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to update IDP ADFS %s error: %s", d.Id(), err), resp)
	}

	log.Printf("Updated idp adfs")
	return readIdpAdfs(ctx, d, meta)
}

// deleteIdpAdfs is used by the idp_adfs resource to delete an idp adfs from Genesys cloud
func deleteIdpAdfs(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getIdpAdfsProxy(sdkConfig)

	resp, err := proxy.deleteIdpAdfs(ctx, d.Id())
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to delete idp adfs %s error: %s", d.Id(), err), resp)
	}

	return util.WithRetries(ctx, 180*time.Second, func() *retry.RetryError {
		_, resp, err := proxy.getIdpAdfs(ctx)
		if err != nil {
			if util.IsStatus404(resp) {
				// IDP ADFS deleted
				log.Printf("Deleted IDP ADFS")
				return nil
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Error deleting IDP ADFS: %s", err), resp))
		}
		return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("IDP ADFS still exists"), resp))
	})
}

// getIdpAdfsFromResourceData maps data from schema ResourceData object to a platformclientv2.Adfs
func getIdpAdfsFromResourceData(d *schema.ResourceData) platformclientv2.Adfs {
	return platformclientv2.Adfs{
		Name:                   platformclientv2.String(d.Get("name").(string)),
		Disabled:               platformclientv2.Bool(d.Get("disabled").(bool)),
		IssuerURI:              platformclientv2.String(d.Get("issuer_uri").(string)),
		SsoTargetURI:           platformclientv2.String(d.Get("target_uri").(string)),
		RelyingPartyIdentifier: platformclientv2.String(d.Get("relying_party_identifier").(string)),
		SloURI:                 platformclientv2.String(d.Get("slo_uri").(string)),
		SloBinding:             platformclientv2.String(d.Get("slo_binding").(string)),
	}
}
