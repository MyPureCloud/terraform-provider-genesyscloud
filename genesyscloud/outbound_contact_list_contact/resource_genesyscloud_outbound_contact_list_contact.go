package outbound_contact_list_contact

import (
	"context"
	"fmt"
	"log"
	"terraform-provider-genesyscloud/genesyscloud/consistency_checker"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"terraform-provider-genesyscloud/genesyscloud/util/constants"
	"terraform-provider-genesyscloud/genesyscloud/util/resourcedata"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v146/platformclientv2"
)

func getAllContacts(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	resources := make(resourceExporter.ResourceIDMetaMap)
	cp := getContactProxy(clientConfig)

	contacts, resp, err := cp.getAllContacts(ctx)
	if err != nil {
		msg := fmt.Sprintf("Failed to read all contact list contacts. Error: %v", err)
		if resp != nil {
			return nil, util.BuildAPIDiagnosticError(ResourceType, msg, resp)
		}
		return nil, util.BuildDiagnosticError(ResourceType, msg, err)
	}

	for _, contact := range contacts {
		id := createComplexContact(*contact.ContactListId, *contact.Id)
		// We construct this to adhere to Terraform's Block Label requirements
		name := "_" + createComplexContactWithDelimiter(*contact.ContactListId, *contact.Id, "_")
		resources[id] = &resourceExporter.ResourceMeta{BlockLabel: name}
	}

	return resources, nil
}

func createOutboundContactListContact(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	cp := getContactProxy(sdkConfig)

	contactListId := d.Get("contact_list_id").(string)
	priority := d.Get("priority").(bool)
	clearSystemData := d.Get("clear_system_data").(bool)
	doNotQueue := d.Get("do_not_queue").(bool)

	contactRequestBody := buildWritableContactFromResourceData(d)

	log.Printf("Creating contact in contact list '%s'", contactListId)
	contactResponseBody, resp, err := cp.createContact(ctx, contactListId, contactRequestBody, priority, clearSystemData, doNotQueue)
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("failed to create contact for contact list '%s': %v", contactListId, err), resp)
	}

	if len(contactResponseBody) != 1 {
		msg := fmt.Sprintf("expected to receive one dialer contact object in contact creation response. Received %v", len(contactResponseBody))
		return util.BuildDiagnosticError(ResourceType, msg, fmt.Errorf("%v", msg))
	}
	contactId := *contactResponseBody[0].Id
	_ = d.Set("contact_id", contactId)
	id := createComplexContact(contactListId, contactId)
	d.SetId(id)
	log.Printf("Finished creating contact '%s' in contact list '%s'", contactId, contactListId)
	return readOutboundContactListContact(ctx, d, meta)
}

func readOutboundContactListContact(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var (
		resp *platformclientv2.APIResponse
		err  error

		sdkConfig = meta.(*provider.ProviderMeta).ClientConfig
		cp        = getContactProxy(sdkConfig)
	)

	contactListId, contactId := splitComplexContact(d.Id())
	if contactListId == "" {
		contactListId = d.Get("contact_list_id").(string)
	}
	if contactId == "" {
		contactId = d.Get("contact_id").(string)
	}

	cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceOutboundContactListContact(), constants.ConsistencyChecks(), ResourceType)

	retryErr := util.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		var contactResponseBody *platformclientv2.Dialercontact

		log.Printf("Reading contact '%s' in contact list '%s'", contactId, contactListId)
		contactResponseBody, resp, err = cp.readContactById(ctx, contactListId, contactId)
		if err != nil {
			if util.IsStatus404(resp) {
				return retry.RetryableError(err)
			}
			return retry.NonRetryableError(err)
		}

		_ = d.Set("contact_list_id", *contactResponseBody.ContactListId)
		_ = d.Set("contact_id", *contactResponseBody.Id)
		resourcedata.SetNillableValue(d, "callable", contactResponseBody.Callable)
		resourcedata.SetNillableValue(d, "data", contactResponseBody.Data)
		resourcedata.SetNillableValueWithSchemaSetWithFunc(d, "phone_number_status", contactResponseBody.PhoneNumberStatus, flattenPhoneNumberStatus)
		resourcedata.SetNillableValueWithSchemaSetWithFunc(d, "contactable_status", contactResponseBody.ContactableStatus, flattenContactableStatus)

		return cc.CheckState(d)
	})
	if retryErr != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("failed to read contact by ID '%s' from contact list '%s'. Error: %v", contactId, contactListId, retryErr), resp)
	}
	log.Printf("Done reading contact '%s' in contact list '%s'", contactId, contactListId)
	return nil
}

func updateOutboundContactListContact(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	cp := getContactProxy(sdkConfig)

	contactRequestBody := buildDialerContactFromResourceData(d)
	contactListId, contactId := splitComplexContact(d.Id())
	if contactListId == "" {
		contactListId = d.Get("contact_list_id").(string)
	}
	if contactId == "" {
		contactId = d.Get("contact_id").(string)
	}

	log.Printf("Updating contact '%s' in contact list '%s'", contactId, contactListId)
	_, resp, err := cp.updateContact(ctx, contactListId, contactId, contactRequestBody)
	if err != nil {
		msg := fmt.Sprintf("failed to update contact '%s' for contact list '%s'. Error: %v", contactId, contactListId, err)
		return util.BuildAPIDiagnosticError(ResourceType, msg, resp)
	}

	log.Printf("Finished updating contact '%s' in contact list '%s'", contactId, contactListId)
	return readOutboundContactListContact(ctx, d, meta)
}

func deleteOutboundContactListContact(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	cp := getContactProxy(sdkConfig)

	contactListId, contactId := splitComplexContact(d.Id())
	if contactListId == "" {
		contactListId = d.Get("contact_list_id").(string)
	}
	if contactId == "" {
		contactId = d.Get("contact_id").(string)
	}

	log.Printf("Deleting contact '%s' from contact list '%s'", contactId, contactListId)
	resp, err := cp.deleteContact(ctx, contactListId, contactId)
	if err != nil {
		msg := fmt.Sprintf("failed to delete contact '%s' from contact list '%s'. Error: %v", contactId, contactListId, err)
		return util.BuildAPIDiagnosticError(ResourceType, msg, resp)
	}

	return util.WithRetries(ctx, 60*time.Second, func() *retry.RetryError {
		log.Printf("Reading contact '%s'", d.Id())
		_, resp, err := cp.readContactById(ctx, contactListId, contactId)
		if err != nil {
			if util.IsStatus404(resp) {
				log.Printf("Contact '%s' deleted", contactId)
				return nil
			}
			msg := fmt.Sprintf("failed to delete contact '%s'. Error: %v", contactId, err)
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, msg, resp))
		}
		msg := fmt.Sprintf("contact '%s' still exists in contact list '%s'", contactId, contactListId)
		return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, msg, resp))
	})
}
