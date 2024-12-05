package external_contacts

import (
	"context"
	"fmt"
	"log"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"terraform-provider-genesyscloud/genesyscloud/util/constants"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"terraform-provider-genesyscloud/genesyscloud/consistency_checker"
	"terraform-provider-genesyscloud/genesyscloud/util/resourcedata"

	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v146/platformclientv2"
)

/*
The resource_genesyscloud_externalcontacts_contact.go contains all of the methods that perform the core logic for a resource.
In general a resource should have a approximately 5 methods in it:

1.  A getAll.... function that the CX as Code exporter will use during the process of exporting Genesys Cloud.
2.  A create.... function that the resource will use to create a Genesys Cloud object (e.g. genesycloud_externalcontacts_contacts)
3.  A read.... function that looks up a single resource.
4.  An update... function that updates a single resource.
5.  A delete.... function that deletes a single resource.

Two things to note:

1.  All code in these methods should be focused on getting data in and out of Terraform.  All code that is used for interacting
    with a Genesys API should be encapsulated into a proxy class contained within the package.

2.  In general, to keep this file somewhat manageable, if you find yourself with a number of helper functions move them to a
utils function in the package.  This will keep the code manageable and easy to work through.
*/
// getAllAuthExternalContacts retrieves all of the external contacts via Terraform in the Genesys Cloud and is used for the exporter
func getAllAuthExternalContacts(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	ep := getExternalContactsContactsProxy(clientConfig)
	resources := make(resourceExporter.ResourceIDMetaMap)

	externalContacts, resp, err := ep.getAllExternalContacts(ctx)
	if err != nil {
		return nil, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to get External Contacts error: %s", err), resp)
	}

	for _, externalContact := range *externalContacts {
		if externalContact.Id == nil {
			continue
		}
		log.Printf("Dealing with external contact id : %s", *externalContact.Id)
		resources[*externalContact.Id] = &resourceExporter.ResourceMeta{BlockLabel: *externalContact.Id}
	}
	return resources, nil
}

// createExternalContact is used by the externalcontacts_contacts resource to create Genesyscloud external_contacts
func createExternalContact(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	ep := getExternalContactsContactsProxy(sdkConfig)

	log.Printf("Creating external contact")
	externalContact := getExternalContactFromResourceData(d)
	contact, resp, err := ep.createExternalContact(ctx, externalContact)
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to create external contact error: %s", err), resp)
	}

	if contact == nil || contact.Id == nil {
		msg := "No contact ID was returned on the response object from createExternalContact"
		return util.BuildDiagnosticError(ResourceType, msg, fmt.Errorf("%s", msg))
	}

	d.SetId(*contact.Id)
	log.Printf("Created external contact %s", *contact.Id)
	return readExternalContact(ctx, d, meta)
}

// readExternalContacts is used by the externalcontacts_contact resource to read an external contact from genesys cloud.
func readExternalContact(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	ep := getExternalContactsContactsProxy(sdkConfig)
	cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceExternalContact(), constants.ConsistencyChecks(), ResourceType)

	log.Printf("Reading external contact %s", d.Id())

	return util.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		externalContact, resp, getErr := ep.getExternalContactById(ctx, d.Id())
		if getErr != nil {
			if util.IsStatus404(resp) {
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read external contact %s | error: %s", d.Id(), getErr), resp))
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read external contact %s | error: %s", d.Id(), getErr), resp))
		}

		resourcedata.SetNillableValue(d, "first_name", externalContact.FirstName)
		resourcedata.SetNillableValue(d, "middle_name", externalContact.MiddleName)
		resourcedata.SetNillableValue(d, "last_name", externalContact.LastName)
		resourcedata.SetNillableValue(d, "salutation", externalContact.Salutation)
		resourcedata.SetNillableValue(d, "title", externalContact.Title)
		resourcedata.SetNillableValueWithInterfaceArrayWithFunc(d, "work_phone", externalContact.WorkPhone, flattenPhoneNumber)
		resourcedata.SetNillableValueWithInterfaceArrayWithFunc(d, "cell_phone", externalContact.CellPhone, flattenPhoneNumber)
		resourcedata.SetNillableValueWithInterfaceArrayWithFunc(d, "home_phone", externalContact.HomePhone, flattenPhoneNumber)
		resourcedata.SetNillableValueWithInterfaceArrayWithFunc(d, "other_phone", externalContact.OtherPhone, flattenPhoneNumber)
		resourcedata.SetNillableValue(d, "work_email", externalContact.WorkEmail)
		resourcedata.SetNillableValue(d, "personal_email", externalContact.PersonalEmail)
		resourcedata.SetNillableValue(d, "other_email", externalContact.OtherEmail)
		resourcedata.SetNillableValueWithInterfaceArrayWithFunc(d, "address", externalContact.Address, flattenSdkAddress)
		resourcedata.SetNillableValueWithInterfaceArrayWithFunc(d, "twitter_id", externalContact.TwitterId, flattenSdkTwitterId)
		resourcedata.SetNillableValueWithInterfaceArrayWithFunc(d, "line_id", externalContact.LineId, flattenSdkLineId)
		resourcedata.SetNillableValueWithInterfaceArrayWithFunc(d, "whatsapp_id", externalContact.WhatsAppId, flattenSdkWhatsAppId)
		resourcedata.SetNillableValueWithInterfaceArrayWithFunc(d, "facebook_id", externalContact.FacebookId, flattenSdkFacebookId)
		resourcedata.SetNillableValue(d, "survey_opt_out", externalContact.SurveyOptOut)
		resourcedata.SetNillableValue(d, "external_system_url", externalContact.ExternalSystemUrl)

		log.Printf("Read external contact %s", d.Id())
		return cc.CheckState(d)
	})
}

// updateExternalContacts is used by the externalcontacts_contacts resource to update an external contact in Genesys Cloud
func updateExternalContact(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	ep := getExternalContactsContactsProxy(sdkConfig)

	log.Printf("Updating external contact %s", d.Id())
	externalContact := getExternalContactFromResourceData(d)
	_, resp, err := ep.updateExternalContact(ctx, d.Id(), externalContact)
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to update external contact %s error: %s", d.Id(), err), resp)
	}

	log.Printf("Updated external contact %s", d.Id())
	return readExternalContact(ctx, d, meta)
}

// deleteExternalContacts is used by the externalcontacts_contacts resource to delete an external contact from Genesys cloud.
func deleteExternalContact(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	ep := getExternalContactsContactsProxy(sdkConfig)

	log.Printf("Deleting external contact %s", d.Id())
	resp, err := ep.deleteExternalContactId(ctx, d.Id())
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to delete external contact %s error: %s", d.Id(), err), resp)
	}

	return util.WithRetries(ctx, 180*time.Second, func() *retry.RetryError {
		_, resp, err := ep.getExternalContactById(ctx, d.Id())

		if err == nil {
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Error deleting external contact %s | error: %s", d.Id(), err), resp))
		}
		if util.IsStatus404(resp) {
			// Success  : External contact deleted
			log.Printf("Deleted external contact %s", d.Id())
			return nil
		}
		return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("External contact %s still exists", d.Id()), resp))
	})
}
