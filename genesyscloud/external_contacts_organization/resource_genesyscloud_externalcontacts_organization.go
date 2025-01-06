package external_contacts_organization

import (
	"context"
	"fmt"
	"log"
	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"terraform-provider-genesyscloud/genesyscloud/util/resourcedata"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v150/platformclientv2"

	"terraform-provider-genesyscloud/genesyscloud/consistency_checker"

	"terraform-provider-genesyscloud/genesyscloud/provider"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
)

/*
The resource_genesyscloud_external_contacts_organization.go contains all of the methods that perform the core logic for a resource.
*/

// getAllAuthExternalContactsOrganization retrieves all of the external contacts organization via Terraform in the Genesys Cloud and is used for the exporter
func getAllAuthExternalContactsOrganizations(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	log.Println(ResourceType + " resources cannot be exported due to an API limitation.")
	return nil, nil

	// TODO uncomment once DEVTOOLING-977 has been resolved
	//proxy := newExternalContactsOrganizationProxy(clientConfig)
	//resources := make(resourceExporter.ResourceIDMetaMap)

	//externalOrganizations, response, err := proxy.getAllExternalContactsOrganization(ctx, "")
	//if err != nil {
	//	return nil, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("failed to get external organization error: %s", err), response)
	//}
	//
	//for _, externalOrganization := range *externalOrganizations {
	//	resources[*externalOrganization.Id] = &resourceExporter.ResourceMeta{BlockLabel: *externalOrganization.Id}
	//}
	//
	//return resources, nil
}

// createExternalContactsOrganization is used by the external_contacts_organization resource to create Genesys cloud external contacts organization
func createExternalContactsOrganization(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getExternalContactsOrganizationProxy(sdkConfig)

	externalContactsOrganization, err := getExternalContactsOrganizationFromResourceData(d)
	if err != nil {
		return util.BuildDiagnosticError(ResourceType, "failed to build external organization request body", err)
	}

	log.Printf("Creating external contacts organization %s", *externalContactsOrganization.Name)
	externalOrganization, response, err := proxy.createExternalContactsOrganization(ctx, &externalContactsOrganization)
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("failed to get external organization error: %s", err), response)
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
		externalOrganization, response, err := proxy.getExternalContactsOrganizationById(ctx, d.Id())
		if err != nil {
			if util.IsStatus404ByInt(response.StatusCode) {
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read organization contact %s | error: %s", d.Id(), err), response))
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read organization contact %s | error: %s", d.Id(), err), response))
		}

		cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceExternalContactsOrganization(), 0, "")

		resourcedata.SetNillableValue(d, "name", externalOrganization.Name)
		resourcedata.SetNillableValue(d, "company_type", externalOrganization.CompanyType)
		resourcedata.SetNillableValue(d, "industry", externalOrganization.Industry)
		resourcedata.SetNillableValue(d, "primary_contact_id", externalOrganization.PrimaryContactId)
		resourcedata.SetNillableValueWithInterfaceArrayWithFunc(d, "address", externalOrganization.Address, flattenSdkAddress)
		resourcedata.SetNillableValueWithInterfaceArrayWithFunc(d, "phone_number", externalOrganization.PhoneNumber, flattenPhoneNumber)
		resourcedata.SetNillableValueWithInterfaceArrayWithFunc(d, "fax_number", externalOrganization.FaxNumber, flattenPhoneNumber)
		resourcedata.SetNillableValue(d, "employee_count", externalOrganization.EmployeeCount)
		resourcedata.SetNillableValue(d, "revenue", externalOrganization.Revenue)
		resourcedata.SetNillableValue(d, "tags", externalOrganization.Tags)
		resourcedata.SetNillableValue(d, "websites", externalOrganization.Websites)
		resourcedata.SetNillableValueWithInterfaceArrayWithFunc(d, "tickers", externalOrganization.Tickers, flattenTickers)
		resourcedata.SetNillableValueWithInterfaceArrayWithFunc(d, "twitter", externalOrganization.TwitterId, flattenSdkTwitterId)
		resourcedata.SetNillableValue(d, "external_system_url", externalOrganization.ExternalSystemUrl)
		resourcedata.SetNillableValueWithInterfaceArrayWithFunc(d, "trustor", externalOrganization.Trustor, flattenTrustor)
		resourcedata.SetNillableValueWithInterfaceArrayWithFunc(d, "external_data_sources", externalOrganization.ExternalDataSources, flattenExternalDataSources)
		dataSchema, err := flattenDataSchema(externalOrganization.Schema)
		if err != nil {
			return retry.NonRetryableError(fmt.Errorf("failed to flatten Data Schema for resource %s | error: %s", d.Id(), err))
		}
		_ = d.Set("schema", dataSchema)
		cf, err := flattenCustomFields(externalOrganization.CustomFields)
		if err != nil {
			return retry.NonRetryableError(fmt.Errorf("failed to flatten Custom Schema for resource %s | error: %s", d.Id(), err))
		}
		_ = d.Set("custom_fields", cf)

		log.Printf("Read external contacts organization %s %s", d.Id(), *externalOrganization.Name)
		return cc.CheckState(d)
	})
}

// updateExternalContactsOrganization is used by the external_contacts_organization resource to update an external contacts organization in Genesys Cloud
func updateExternalContactsOrganization(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getExternalContactsOrganizationProxy(sdkConfig)

	externalContactsOrganization, err := getExternalContactsOrganizationFromResourceData(d)

	if err != nil {
		return util.BuildDiagnosticError(ResourceType, fmt.Sprintf("failed to build external organization error: %s", err), nil)
	}

	log.Printf("Updating external contacts organization %s", *externalContactsOrganization.Name)
	externalOrganization, response, err := proxy.updateExternalContactsOrganization(ctx, d.Id(), &externalContactsOrganization)
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("failed to update external organization error: %s", err), response)
	}

	log.Printf("Updated external contacts organization %s", *externalOrganization.Id)
	return readExternalContactsOrganization(ctx, d, meta)
}

// deleteExternalContactsOrganization is used by the external_contacts_organization resource to delete an external contacts organization from Genesys cloud
func deleteExternalContactsOrganization(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getExternalContactsOrganizationProxy(sdkConfig)

	response, err := proxy.deleteExternalContactsOrganization(ctx, d.Id())
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("failed to delete external organization %s: %s", d.Id(), err), response)
	}

	return util.WithRetries(ctx, 180*time.Second, func() *retry.RetryError {
		_, response, err := proxy.getExternalContactsOrganizationById(ctx, d.Id())

		if err != nil {
			if util.IsStatus404ByInt(response.StatusCode) {
				log.Printf("Deleted external organization %s", d.Id())
				return nil
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("error deleting external organization %s: %s", d.Id(), err), response))
		}

		return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("external organization still exists %s:", d.Id()), response))
	})
}
