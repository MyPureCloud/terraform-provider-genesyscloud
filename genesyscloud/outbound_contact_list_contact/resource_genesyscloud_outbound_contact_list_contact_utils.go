package outbound_contact_list_contact

import (
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v130/platformclientv2"
)

func buildWritableContactFromResourceData(d *schema.ResourceData) *platformclientv2.Writabledialercontact {
	contactListId := d.Get("contact_list_id").(string)
	contactId, _ := d.Get("id").(string)
	if contactId == "" {
		contactId = uuid.NewString()
	}
	callable := d.Get("callable").(bool)

	// TODO - add the rest of the fields
	var contactRequest = &platformclientv2.Writabledialercontact{
		Id:            &contactId,
		ContactListId: &contactListId,
		Callable:      &callable,
	}

	return contactRequest
}
