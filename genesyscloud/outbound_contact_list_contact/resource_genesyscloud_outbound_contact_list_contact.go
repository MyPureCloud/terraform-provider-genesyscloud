package outbound_contact_list_contact

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"terraform-provider-genesyscloud/genesyscloud/util/resourcedata"
)

func createOutboundContactListContact(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	cp := getContactProxy(sdkConfig)

	contactListId := d.Get("contact_list_id").(string)
	priority := d.Get("priority").(bool)
	clearSystemData := d.Get("clear_system_data").(bool)
	doNotQueue := d.Get("do_not_queue").(bool)

	contactRequestBody := buildWritableContactFromResourceData(d)

	contactResponseBody, resp, err := cp.createContact(ctx, contactListId, contactRequestBody, priority, clearSystemData, doNotQueue)
	if err != nil {
		return util.BuildAPIDiagnosticError(resourceName, fmt.Sprintf("failed to create contact '%s' for contact list '%s': %v", *contactRequestBody.Id, contactListId, err), resp)
	}

	if len(contactResponseBody) != 1 {
		msg := fmt.Sprintf("expected to receive one dialer contact object in contact creation response. Received %v", len(contactResponseBody))
		return util.BuildDiagnosticError(resourceName, msg, fmt.Errorf("%v", msg))
	}

	id := createContactId(contactListId, *contactResponseBody[0].Id)
	d.SetId(id)
	return readOutboundContactListContact(ctx, d, meta)
}

func readOutboundContactListContact(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	cp := getContactProxy(sdkConfig)

	contactListId, contactId, err := parseContactListIdAndContactId(d.Id())
	if err != nil {
		return util.BuildDiagnosticError(resourceName, "failed to parse contact list and contact ID", err)
	}

	contactResponseBody, resp, err := cp.readContactById(ctx, contactListId, contactId)
	if err != nil {
		return util.BuildAPIDiagnosticError(resourceName, fmt.Sprintf("failed to read contact by ID '%s' from contact list '%s'. Error: %v", contactId, contactListId, err), resp)
	}

	_ = d.Set("contact_list_id", *contactResponseBody.ContactListId)
	resourcedata.SetNillableValue(d, "callable", contactResponseBody.Callable)
	resourcedata.SetNillableValue(d, "data", contactResponseBody.Data)
	resourcedata.SetNillableValueWithSchemaSetWithFunc(d, "phone_number_status", contactResponseBody.PhoneNumberStatus, flattenPhoneNumberStatus)
	resourcedata.SetNillableValueWithSchemaSetWithFunc(d, "contactable_status", contactResponseBody.ContactableStatus, flattenContactableStatus)

	return nil
}

func updateOutboundContactListContact(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	return nil
}

func deleteOutboundContactListContact(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	return nil
}
