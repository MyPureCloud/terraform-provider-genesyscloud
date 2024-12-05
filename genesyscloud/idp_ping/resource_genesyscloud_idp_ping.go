package idp_ping

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
The resource_genesyscloud_idp_ping.go contains all of the methods that perform the core logic for a resource.
*/

// getAllAuthIdpPing retrieves all of the idp ping via Terraform in the Genesys Cloud and is used for the exporter
func getAllAuthIdpPings(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	// Although this resource typically has only a single instance,
	// we are attempting to fetch the data from the API in order to
	// verify the user's permission to access this resource's API endpoint(s).

	proxy := getIdpPingProxy(clientConfig)
	resources := make(resourceExporter.ResourceIDMetaMap)

	_, resp, err := proxy.getIdpPing(ctx)
	if err != nil {
		if util.IsStatus404(resp) {
			// Don't export if config doesn't exist
			return resources, nil
		}
		return nil, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to get IDP Ping error: %s", err), resp)
	}

	resources["0"] = &resourceExporter.ResourceMeta{BlockLabel: "ping"}
	return resources, nil
}

// createIdpPing is used by the idp_ping resource to create Genesys cloud idp ping
func createIdpPing(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Printf("Creating IDP Ping")
	d.SetId("ping")
	return updateIdpPing(ctx, d, meta)
}

// readIdpPing is used by the idp_ping resource to read an idp ping from genesys cloud
func readIdpPing(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getIdpPingProxy(sdkConfig)

	log.Printf("Reading idp ping")

	cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceIdpPing(), constants.ConsistencyChecks(), ResourceType)

	return util.WithRetriesForReadCustomTimeout(ctx, d.Timeout(schema.TimeoutRead), d, func() *retry.RetryError {
		pingIdentity, resp, getErr := proxy.getIdpPing(ctx)
		if getErr != nil {
			if util.IsStatus404(resp) {
				createIdpPing(ctx, d, meta)
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read IDP Ping: %s", getErr), resp))
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read IDP Ping: %s", getErr), resp))
		}

		resourcedata.SetNillableValue(d, "name", pingIdentity.Name)
		resourcedata.SetNillableValue(d, "disabled", pingIdentity.Disabled)
		resourcedata.SetNillableValue(d, "issuer_uri", pingIdentity.IssuerURI)
		resourcedata.SetNillableValue(d, "target_uri", pingIdentity.SsoTargetURI)
		resourcedata.SetNillableValue(d, "relying_party_identifier", pingIdentity.RelyingPartyIdentifier)
		resourcedata.SetNillableValue(d, "slo_uri", pingIdentity.SloURI)
		resourcedata.SetNillableValue(d, "slo_binding", pingIdentity.SloBinding)

		if pingIdentity.Certificate != nil {
			d.Set("certificates", lists.StringListToInterfaceList([]string{*pingIdentity.Certificate}))
		} else if pingIdentity.Certificates != nil {
			d.Set("certificates", lists.StringListToInterfaceList(*pingIdentity.Certificates))
		} else {
			d.Set("certificates", nil)
		}

		log.Printf("Read idp ping")
		return cc.CheckState(d)
	})
}

// updateIdpPing is used by the idp_ping resource to update an idp ping in Genesys Cloud
func updateIdpPing(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getIdpPingProxy(sdkConfig)

	idpPing := getIdpPingFromResourceData(d)

	certificates := lists.BuildSdkStringListFromInterfaceArray(d, "certificates")
	if certificates != nil {
		if len(*certificates) == 1 {
			idpPing.Certificate = &(*certificates)[0]
		}
		idpPing.Certificates = certificates
	}

	log.Printf("Updating idp ping")
	_, resp, err := proxy.updateIdpPing(ctx, d.Id(), &idpPing)
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to update IDP Ping %s error: %s", d.Id(), err), resp)
	}

	log.Printf("Updated idp ping")
	return readIdpPing(ctx, d, meta)
}

// deleteIdpPing is used by the idp_ping resource to delete an idp ping from Genesys cloud
func deleteIdpPing(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getIdpPingProxy(sdkConfig)

	resp, err := proxy.deleteIdpPing(ctx, d.Id())
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to delete IDP Ping %s error: %s", d.Id(), err), resp)
	}

	return util.WithRetries(ctx, 60*time.Second, func() *retry.RetryError {
		_, resp, err := proxy.getIdpPing(ctx)

		if err != nil {
			if util.IsStatus404(resp) {
				// IDP Ping deleted
				log.Printf("Deleted IDP Ping")
				return nil
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Error deleting IDP Ping: %s", err), resp))
		}
		return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("IDP Ping still exists"), resp))
	})
}

// getIdpPingFromResourceData maps data from schema ResourceData object to a platformclientv2.Pingidentity
func getIdpPingFromResourceData(d *schema.ResourceData) platformclientv2.Pingidentity {
	return platformclientv2.Pingidentity{
		Name:                   platformclientv2.String(d.Get("name").(string)),
		Disabled:               platformclientv2.Bool(d.Get("disabled").(bool)),
		IssuerURI:              platformclientv2.String(d.Get("issuer_uri").(string)),
		SsoTargetURI:           platformclientv2.String(d.Get("target_uri").(string)),
		RelyingPartyIdentifier: platformclientv2.String(d.Get("relying_party_identifier").(string)),
		SloURI:                 platformclientv2.String(d.Get("slo_uri").(string)),
		SloBinding:             platformclientv2.String(d.Get("slo_binding").(string)),
	}
}
