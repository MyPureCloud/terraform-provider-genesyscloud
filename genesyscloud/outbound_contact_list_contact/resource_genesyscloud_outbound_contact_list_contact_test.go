package outbound_contact_list_contact

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"strconv"
	"strings"
	outboundContactList "terraform-provider-genesyscloud/genesyscloud/outbound_contact_list"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"testing"
)

func TestAccResourceOutboundContactListContact(t *testing.T) {
	var (
		resourceId = "contact"

		contactListResourceId = "contact_list"
		contactListName       = "tf test contact list " + uuid.NewString()
		columnNames           = []string{strconv.Quote("Cell"), strconv.Quote("Home")}
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, nil),
		Steps: []resource.TestStep{
			{
				Config: outboundContactList.GenerateOutboundContactList(
					contactListResourceId,
					contactListName,
					util.NullValue,
					util.NullValue,
					[]string{},
					columnNames,
					util.FalseValue,
					util.NullValue,
					util.NullValue,
				) + generateOutboundContactListContact(
					resourceId,
					util.NullValue,
					"genesyscloud_outbound_contact_list."+contactListResourceId+".id",
					util.GenerateMapAttr(
						"data",
						util.GenerateMapProperty(
							"Cell",
							strconv.Quote("+11111111"),
						),
						util.GenerateMapProperty(
							"Home",
							strconv.Quote("+22222222"),
						),
					),
				),
			},
		},
	})
}

func generateOutboundContactListContact(
	resourceId,
	id,
	contactListId,
	data string,
	nestedBlocks ...string,
) string {
	return fmt.Sprintf(`
resource "%s" "%s" {
	id              = %s
    contact_list_id = %s
    %s
    %s
}
`, resourceName, resourceId, id, contactListId, data, strings.Join(nestedBlocks, "\n"))
}

func generatePhoneNumberStatus(key, callable string) string {
	return fmt.Sprintf(`
	phone_number_status {
		key      = "%s"
        callable = %s
	}
`, key, callable)
}
