package idp_salesforce

import (
	"context"
	"fmt"
	"log"
	"terraform-provider-genesyscloud/genesyscloud/consistency_checker"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"terraform-provider-genesyscloud/genesyscloud/util/constants"
	"terraform-provider-genesyscloud/genesyscloud/util/lists"
	"terraform-provider-genesyscloud/genesyscloud/util/resourcedata"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v146/platformclientv2"
)

/*
The resource_genesyscloud_idp_salesforce.go contains all of the methods that perform the core logic for a resource.
*/

func getAllIdpSalesforce(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	// Although this resource typically has only a single instance,
	// we are attempting to fetch the data from the API in order to
	// verify the user's permission to access this resource's API endpoint(s).

	proxy := getIdpSalesforceProxy(clientConfig)
	resources := make(resourceExporter.ResourceIDMetaMap)

	_, resp, getErr := proxy.getIdpSalesforce(ctx)
	if getErr != nil {
		if util.IsStatus404(resp) {
			// Don't export if config doesn't exist
			return resources, nil
		}
		return nil, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to get IDP Salesforce error: %s", getErr), resp)
	}

	resources["0"] = &resourceExporter.ResourceMeta{BlockLabel: "salesforce"}
	return resources, nil
}

func createIdpSalesforce(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Printf("Creating IDP Salesforce")
	d.SetId("salesforce")
	return updateIdpSalesforce(ctx, d, meta)
}

func readIdpSalesforce(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getIdpSalesforceProxy(sdkConfig)
	cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceIdpSalesforce(), constants.ConsistencyChecks(), ResourceType)

	log.Printf("Reading IDP Salesforce")

	return util.WithRetriesForReadCustomTimeout(ctx, d.Timeout(schema.TimeoutRead), d, func() *retry.RetryError {
		salesforce, resp, getErr := proxy.getIdpSalesforce(ctx)
		if getErr != nil {
			if util.IsStatus404(resp) {
				createIdpSalesforce(ctx, d, meta)
				return retry.RetryableError(fmt.Errorf("failed to read IDP Salesforce: %s", getErr))
			}
			return retry.NonRetryableError(fmt.Errorf("failed to read IDP Salesforce: %s", getErr))
		}

		if salesforce.Certificate != nil {
			_ = d.Set("certificates", lists.StringListToInterfaceList([]string{*salesforce.Certificate}))
		} else if salesforce.Certificates != nil {
			_ = d.Set("certificates", lists.StringListToInterfaceList(*salesforce.Certificates))
		} else {
			_ = d.Set("certificates", nil)
		}

		resourcedata.SetNillableValue(d, "issuer_uri", salesforce.IssuerURI)
		resourcedata.SetNillableValue(d, "target_uri", salesforce.SsoTargetURI)
		resourcedata.SetNillableValue(d, "disabled", salesforce.Disabled)
		resourcedata.SetNillableValue(d, "name", salesforce.Name)
		resourcedata.SetNillableValue(d, "slo_uri", salesforce.SloURI)
		resourcedata.SetNillableValue(d, "slo_binding", salesforce.SloBinding)
		resourcedata.SetNillableValue(d, "relying_party_identifier", salesforce.RelyingPartyIdentifier)

		log.Printf("Read IDP Salesforce")
		return cc.CheckState(d)
	})
}

func updateIdpSalesforce(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getIdpSalesforceProxy(sdkConfig)

	log.Printf("Updating IDP Salesforce")
	update := platformclientv2.Salesforce{
		IssuerURI: platformclientv2.String(d.Get("issuer_uri").(string)),
		Disabled:  platformclientv2.Bool(d.Get("disabled").(bool)),
	}

	if targetUri := d.Get("target_uri").(string); targetUri != "" {
		update.SsoTargetURI = &targetUri
	}

	if name, _ := d.Get("name").(string); name != "" {
		update.Name = &name
	}

	if sloUri, _ := d.Get("slo_uri").(string); sloUri != "" {
		update.SloURI = &sloUri
	}

	if sloBinding, _ := d.Get("slo_binding").(string); sloBinding != "" {
		update.SloBinding = &sloBinding
	}

	if rpId, _ := d.Get("relying_party_identifier").(string); rpId != "" {
		update.RelyingPartyIdentifier = &rpId
	}

	certificates := lists.BuildSdkStringListFromInterfaceArray(d, "certificates")
	if certificates != nil {
		if len(*certificates) == 1 {
			update.Certificate = &(*certificates)[0]
		}
		update.Certificates = certificates
	}

	_, resp, err := proxy.updateIdpSalesforce(ctx, &update)
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to update IDP Salesforce %s error: %s", d.Id(), err), resp)
	}

	log.Printf("Updated IDP Salesforce")
	return readIdpSalesforce(ctx, d, meta)
}

func deleteIdpSalesforce(ctx context.Context, _ *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getIdpSalesforceProxy(sdkConfig)

	log.Printf("Deleting IDP Salesforce")
	resp, err := proxy.deleteIdpSalesforce(ctx)
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to delete IDP Salesforce error: %s", err), resp)
	}

	return util.WithRetries(ctx, 180*time.Second, func() *retry.RetryError {
		_, resp, err := proxy.getIdpSalesforce(ctx)
		if err != nil {
			if util.IsStatus404(resp) {
				// IDP Salesforce deleted
				log.Printf("Deleted Salesforce Ping")
				return nil
			}
			return retry.NonRetryableError(fmt.Errorf("Error deleting IDP Salesforce: %s", err))
		}
		return retry.RetryableError(fmt.Errorf("IDP Salesforce still exists"))
	})
}
