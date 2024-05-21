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
	"github.com/mypurecloud/platform-client-sdk-go/v129/platformclientv2"

	"terraform-provider-genesyscloud/genesyscloud/consistency_checker"

	"terraform-provider-genesyscloud/genesyscloud/util/constants"
	"terraform-provider-genesyscloud/genesyscloud/util/resourcedata"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
)

/*
The resource_genesyscloud_idp_adfs.go contains all of the methods that perform the core logic for a resource.
*/

// getAllAuthIdpAdfs retrieves all of the idp adfs via Terraform in the Genesys Cloud and is used for the exporter
func getAllAuthIdpAdfss(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	proxy := newIdpAdfsProxy(clientConfig)
	resources := make(resourceExporter.ResourceIDMetaMap)

	aDFS, err := proxy.getAllIdpAdfs(ctx)
	if err != nil {
		return nil, diag.Errorf("Failed to get idp adfs: %v", err)
	}
	resources[*aDFS.Id] = &resourceExporter.ResourceMeta{Name: *aDFS.Name}
	return resources, nil
}

// createIdpAdfs is used by the idp_adfs resource to create Genesys cloud idp adfs
func createIdpAdfs(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return nil
}

// readIdpAdfs is used by the idp_adfs resource to read an idp adfs from genesys cloud
func readIdpAdfs(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getIdpAdfsProxy(sdkConfig)

	log.Printf("Reading idp adfs %s", d.Id())

	return util.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		aDFS, getErr := proxy.getAllIdpAdfs(ctx)
		if getErr != nil {
			return retry.NonRetryableError(fmt.Errorf("Failed to read idp adfs %s: %s", d.Id(), getErr))
		}

		cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceIdpAdfs(), constants.DefaultConsistencyChecks, resourceName)

		resourcedata.SetNillableValue(d, "name", aDFS.Name)
		resourcedata.SetNillableValue(d, "disabled", aDFS.Disabled)
		resourcedata.SetNillableValue(d, "issuer_u_r_i", aDFS.IssuerURI)
		resourcedata.SetNillableValue(d, "sso_target_u_r_i", aDFS.SsoTargetURI)
		resourcedata.SetNillableValue(d, "slo_u_r_i", aDFS.SloURI)
		resourcedata.SetNillableValue(d, "slo_binding", aDFS.SloBinding)
		resourcedata.SetNillableValue(d, "relying_party_identifier", aDFS.RelyingPartyIdentifier)
		resourcedata.SetNillableValue(d, "certificate", aDFS.Certificate)
		resourcedata.SetNillableValue(d, "certificates", aDFS.Certificates)
		log.Printf("Read idp adfs %s %s", d.Id(), *aDFS.Name)
		return cc.CheckState(d)
	})
}

// updateIdpAdfs is used by the idp_adfs resource to update an idp adfs in Genesys Cloud
func updateIdpAdfs(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getIdpAdfsProxy(sdkConfig)

	idpAdfs := getIdpAdfsFromResourceData(d)

	log.Printf("Updating idp adfs %s", *idpAdfs.Name)
	_, err := proxy.updateIdpAdfs(ctx, d.Id(), &idpAdfs)
	if err != nil {
		return diag.Errorf("Failed to update idp adfs: %s", err)
	}

	log.Printf("Updated idp adfs %s", d.Id)
	return readIdpAdfs(ctx, d, meta)
}

// deleteIdpAdfs is used by the idp_adfs resource to delete an idp adfs from Genesys cloud
func deleteIdpAdfs(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getIdpAdfsProxy(sdkConfig)

	_, err := proxy.deleteIdpAdfs(ctx, d.Id())
	if err != nil {
		return diag.Errorf("Failed to delete idp adfs %s: %s", d.Id(), err)
	}

	return util.WithRetries(ctx, 180*time.Second, func() *retry.RetryError {
		_, err := proxy.getAllIdpAdfs(ctx)

		if err != nil {
			return retry.NonRetryableError(fmt.Errorf("Error deleting idp adfs %s: %s", d.Id(), err))
		}

		return retry.RetryableError(fmt.Errorf("idp adfs %s still exists", d.Id()))
	})
}

// getIdpAdfsFromResourceData maps data from schema ResourceData object to a platformclientv2.Adfs
func getIdpAdfsFromResourceData(d *schema.ResourceData) platformclientv2.Adfs {
	return platformclientv2.Adfs{
		Name:                   platformclientv2.String(d.Get("name").(string)),
		Disabled:               platformclientv2.Bool(d.Get("disabled").(bool)),
		IssuerURI:              platformclientv2.String(d.Get("issuer_u_r_i").(string)),
		SsoTargetURI:           platformclientv2.String(d.Get("sso_target_u_r_i").(string)),
		SloURI:                 platformclientv2.String(d.Get("slo_u_r_i").(string)),
		SloBinding:             platformclientv2.String(d.Get("slo_binding").(string)),
		RelyingPartyIdentifier: platformclientv2.String(d.Get("relying_party_identifier").(string)),
		Certificate:            platformclientv2.String(d.Get("certificate").(string)),
		// TODO: Handle certificates property

	}
}
