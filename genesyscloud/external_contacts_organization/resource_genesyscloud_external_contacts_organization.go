package external_contacts_organization

import (
	"context"
	"fmt"
	"log"
	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v146/platformclientv2"

	"terraform-provider-genesyscloud/genesyscloud/consistency_checker"

	"terraform-provider-genesyscloud/genesyscloud/provider"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
)

/*
The resource_genesyscloud_external_contacts_organization.go contains all of the methods that perform the core logic for a resource.
*/

// getAllAuthExternalContactsOrganization retrieves all of the external contacts organization via Terraform in the Genesys Cloud and is used for the exporter
func getAllAuthExternalContactsOrganizations(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	proxy := newExternalContactsOrganizationProxy(clientConfig)
	resources := make(resourceExporter.ResourceIDMetaMap)

	externalOrganizations, err := proxy.getAllExternalContactsOrganization(ctx)
	if err != nil {
		return nil, diag.Errorf("Failed to get external contacts organization: %v", err)
	}

	for _, externalOrganization := range *externalOrganizations {
		resources[*externalOrganization.Id] = &resourceExporter.ResourceMeta{Name: *externalOrganization.Id}
	}

	return resources, nil
}

// createExternalContactsOrganization is used by the external_contacts_organization resource to create Genesys cloud external contacts organization
func createExternalContactsOrganization(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getExternalContactsOrganizationProxy(sdkConfig)

	externalContactsOrganization := getExternalContactsOrganizationFromResourceData(d)

	log.Printf("Creating external contacts organization %s", *externalContactsOrganization.Name)
	externalOrganization, err := proxy.createExternalContactsOrganization(ctx, &externalContactsOrganization)
	if err != nil {
		return diag.Errorf("Failed to create external contacts organization: %s", err)
	}

	d.SetId(*externalOrganization.Id)
	log.Printf("Created external contacts organization %s", *externalOrganization.Id)
	return readExternalContactsOrganization(ctx, d, meta)
}

// readExternalContactsOrganization is used by the external_contacts_organization resource to read an external contacts organization from genesys cloud
func readExternalContactsOrganization(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getExternalContactsOrganizationProxy(sdkConfig)

	log.Printf("Reading external contacts organization %s", d.Id())

	return util.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		externalOrganization, respCode, getErr := proxy.getExternalContactsOrganizationById(ctx, d.Id())
		if getErr != nil {
			if util.IsStatus404ByInt(respCode) {
				return retry.RetryableError(fmt.Errorf("Failed to read external contacts organization %s: %s", d.Id(), getErr))
			}
			return retry.NonRetryableError(fmt.Errorf("Failed to read external contacts organization %s: %s", d.Id(), getErr))
		}

		cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceExternalContactsOrganization(), 0, "")
		/*
			resourcedata.SetNillableValue(d, "name", externalOrganization.Name)
			resourcedata.SetNillableValue(d, "company_type", externalOrganization.CompanyType)
			resourcedata.SetNillableValue(d, "industry", externalOrganization.Industry)
			resourcedata.SetNillableValue(d, "primary_contact_id", externalOrganization.PrimaryContactId)
			resourcedata.SetNillableValueWithInterfaceArrayWithFunc(d, "address", externalOrganization.Address, flattenContactAddress)
			resourcedata.SetNillableValueWithInterfaceArrayWithFunc(d, "phone_number", externalOrganization.PhoneNumber, flattenPhoneNumber)
			resourcedata.SetNillableValueWithInterfaceArrayWithFunc(d, "fax_number", externalOrganization.FaxNumber, flattenPhoneNumber)
			resourcedata.SetNillableValue(d, "employee_count", externalOrganization.EmployeeCount)
			resourcedata.SetNillableValue(d, "revenue", externalOrganization.Revenue)
			resourcedata.SetNillableValue(d, "tags", externalOrganization.Tags)
			resourcedata.SetNillableValue(d, "websites", externalOrganization.Websites)
			resourcedata.SetNillableValueWithInterfaceArrayWithFunc(d, "tickers", externalOrganization.Tickers, flattenTickers)
			resourcedata.SetNillableValueWithInterfaceArrayWithFunc(d, "twitter_id", externalOrganization.TwitterId, flattenTwitterId)
			resourcedata.SetNillableValue(d, "external_system_url", externalOrganization.ExternalSystemUrl)
			resourcedata.SetNillableValue(d, "modify_date", externalOrganization.ModifyDate)
			resourcedata.SetNillableValue(d, "create_date", externalOrganization.CreateDate)
			resourcedata.SetNillableValueWithInterfaceArrayWithFunc(d, "trustor", externalOrganization.Trustor, flattenTrustor)
			resourcedata.SetNillableValueWithInterfaceArrayWithFunc(d, "schema", externalOrganization.Schema, flattenDataSchema)
			// TODO: Handle custom_fields property
			resourcedata.SetNillableValueWithInterfaceArrayWithFunc(d, "external_data_sources", externalOrganization.ExternalDataSources, flattenExternalDataSources)
		*/
		log.Printf("Read external contacts organization %s %s", d.Id(), *externalOrganization.Name)
		return cc.CheckState(d)
	})
}

// updateExternalContactsOrganization is used by the external_contacts_organization resource to update an external contacts organization in Genesys Cloud
func updateExternalContactsOrganization(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getExternalContactsOrganizationProxy(sdkConfig)

	externalContactsOrganization := getExternalContactsOrganizationFromResourceData(d)

	log.Printf("Updating external contacts organization %s", *externalContactsOrganization.Name)
	externalOrganization, err := proxy.updateExternalContactsOrganization(ctx, d.Id(), &externalContactsOrganization)
	if err != nil {
		return diag.Errorf("Failed to update external contacts organization: %s", err)
	}

	log.Printf("Updated external contacts organization %s", *externalOrganization.Id)
	return readExternalContactsOrganization(ctx, d, meta)
}

// deleteExternalContactsOrganization is used by the external_contacts_organization resource to delete an external contacts organization from Genesys cloud
func deleteExternalContactsOrganization(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getExternalContactsOrganizationProxy(sdkConfig)

	_, err := proxy.deleteExternalContactsOrganization(ctx, d.Id())
	if err != nil {
		return diag.Errorf("Failed to delete external contacts organization %s: %s", d.Id(), err)
	}

	return util.WithRetries(ctx, 180*time.Second, func() *retry.RetryError {
		_, respCode, err := proxy.getExternalContactsOrganizationById(ctx, d.Id())

		if err != nil {
			if util.IsStatus404ByInt(respCode) {
				log.Printf("Deleted external contacts organization %s", d.Id())
				return nil
			}
			return retry.NonRetryableError(fmt.Errorf("Error deleting external contacts organization %s: %s", d.Id(), err))
		}

		return retry.RetryableError(fmt.Errorf("external contacts organization %s still exists", d.Id()))
	})
}
