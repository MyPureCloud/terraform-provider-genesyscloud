package external_contacts

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"terraform-provider-genesyscloud/genesyscloud/consistency_checker"
	"terraform-provider-genesyscloud/genesyscloud/util/resourcedata"

	gcloud "terraform-provider-genesyscloud/genesyscloud"
	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v105/platformclientv2"
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

	externalContacts, err := ep.getAllExternalContacts(ctx)
	if err != nil {
		return nil, diag.Errorf("Failed to get external contacts: %v", err)
	}

	for _, externalContact := range *externalContacts {
		log.Printf("Dealing with external contact id : %s", *externalContact.Id)
		resources[*externalContact.Id] = &resourceExporter.ResourceMeta{Name: *externalContact.Id}
	}

	return resources, nil
}

// createExternalContact is used by the externalcontacts_contacts resource to create Genesyscloud external_contacts
func createExternalContact(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*gcloud.ProviderMeta).ClientConfig
	ep := getExternalContactsContactsProxy(sdkConfig)

	externalContact := getExternalContactFromResourceData(d)

	contact, err := ep.createExternalContact(ctx, &externalContact)
	if err != nil {
		return diag.Errorf("Failed to create external contact: %s", err)
	}

	d.SetId(*contact.Id)
	log.Printf("Created external contact %s", *contact.Id)
	return readExternalContact(ctx, d, meta)
}

// readExternalContacts is used by the externalcontacts_contact resource to read an external contact from genesys cloud.
func readExternalContact(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*gcloud.ProviderMeta).ClientConfig
	ep := getExternalContactsContactsProxy(sdkConfig)

	log.Printf("Reading contact %s", d.Id())

	return gcloud.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		externalContact, respCode, getErr := ep.getExternalContactById(ctx, d.Id())
		if getErr != nil {
			if gcloud.IsStatus404ByInt(respCode) {
				return retry.RetryableError(fmt.Errorf("Failed to read external contact %s: %s", d.Id(), getErr))
			}
			return retry.NonRetryableError(fmt.Errorf("Failed to read external contact %s: %s", d.Id(), getErr))
		}

		cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceExternalContact())

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
		resourcedata.SetNillableValue(d, "personal_email", externalContact.WorkEmail)
		resourcedata.SetNillableValue(d, "other_email", externalContact.OtherEmail)
		resourcedata.SetNillableValueWithInterfaceArrayWithFunc(d, "address", externalContact.Address, flattenSdkAddress)
		resourcedata.SetNillableValueWithInterfaceArrayWithFunc(d, "twitter_id", externalContact.TwitterId, flattenSdkTwitterId)
		resourcedata.SetNillableValueWithInterfaceArrayWithFunc(d, "line_id", externalContact.LineId, flattenSdkLineId)
		resourcedata.SetNillableValueWithInterfaceArrayWithFunc(d, "whatsapp_id", externalContact.WhatsAppId, flattenSdkWhatsAppId)
		resourcedata.SetNillableValueWithInterfaceArrayWithFunc(d, "facebook_id", externalContact.FacebookId, flattenSdkFacebookId)
		resourcedata.SetNillableValue(d, "survey_opt_out", externalContact.SurveyOptOut)
		resourcedata.SetNillableValue(d, "external_system_url", externalContact.ExternalSystemUrl)

		log.Printf("Read external contact %s", d.Id())
		return cc.CheckState()
	})
}

// updateExternalContacts is used by the externalcontacts_contacts resource to update an external contact in Genesys Cloud
func updateExternalContact(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*gcloud.ProviderMeta).ClientConfig
	ep := getExternalContactsContactsProxy(sdkConfig)

	externalContact := getExternalContactFromResourceData(d)
	_, err := ep.updateExternalContact(ctx, d.Id(), &externalContact)
	if err != nil {
		return diag.Errorf("Failed to update external contact: %s", err)
	}

	log.Printf("Updated external contact")

	return readExternalContact(ctx, d, meta)
}

// deleteExternalContacts is used by the externalcontacts_contacts resource to delete an external contact from Genesys cloud.
func deleteExternalContact(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*gcloud.ProviderMeta).ClientConfig
	ep := getExternalContactsContactsProxy(sdkConfig)

	_, err := ep.deleteExternalContactId(ctx, d.Id())
	if err != nil {
		return diag.Errorf("Failed to delete external contact %s: %s", d.Id(), err)
	}

	return gcloud.WithRetries(ctx, 180*time.Second, func() *retry.RetryError {
		_, respCode, err := ep.getExternalContactById(ctx, d.Id())

		if err == nil {
			return retry.NonRetryableError(fmt.Errorf("Error deleting external contact %s: %s", d.Id(), err))
		}
		if gcloud.IsStatus404ByInt(respCode) {
			// Success  : External contact deleted
			log.Printf("Deleted external contact %s", d.Id())
			return nil
		}

		return retry.RetryableError(fmt.Errorf("External contact %s still exists", d.Id()))
	})
}
