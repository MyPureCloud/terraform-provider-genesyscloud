package outbound_contact_list_contact

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
)

func createOutboundContactListContact(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	cp := getContactProxy(sdkConfig)

	contactListId := d.Get("contact_list_id").(string)
	priority := d.Get("priority").(bool)
	clearSystemData := d.Get("clear_system_data").(bool)
	doNotQueue := d.Get("do_not_queue").(bool)

	contactRequest := buildWritableContactFromResourceData(d)

	contactResponse, resp, err := cp.createContact(ctx, contactListId, contactRequest, priority, clearSystemData, doNotQueue)
	if err != nil {
		return util.BuildAPIDiagnosticError(resourceName, fmt.Sprintf("failed to create contact '%s' for contact list '%s': %v", contactRequest.Id, contactListId, err), resp)
	}

	fmt.Println(contactResponse)

	return nil
}

func readOutboundContactListContact(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	return nil
}

func updateOutboundContactListContact(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	return nil
}

func deleteOutboundContactListContact(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	return nil
}
