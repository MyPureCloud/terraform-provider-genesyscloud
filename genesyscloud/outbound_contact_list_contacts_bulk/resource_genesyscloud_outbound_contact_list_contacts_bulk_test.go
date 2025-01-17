package outbound_contact_list_contacts_bulk

import (
	"testing"

	contactList "terraform-provider-genesyscloud/genesyscloud/outbound_contact_list"
	"terraform-provider-genesyscloud/genesyscloud/util"
)

func TestAccResourceOutboundContactListContactsBulkBasic(t *testing.T) {
	var (
		bulkResourceLabel = "contactListBulk"
		contactListLabel  = "contactList"
		contactListName = "TestContactList" + uuid.NewString()
		filepath      = util.GetTestDataPath("resource", ResourceType, "contacts_bulk.csv")

	)

	contactListMockResource := contactList.GenerateOutboundContactList(
		contactListLabel,
		contactListName,
		util.NullValue,
		util.NullValue,
		[]string{},
		[]string{strconv.Quote(column1), strconv.Quote(column2)},
		util.FalseValue,
		util.NullValue,
		util.NullValue,
	)

	contactListBulkContactsMockResource := GenerateOutboundContactListBulkContacts(
			bulkResourceLabel,
			contactListLabel,
			filepath,
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config:
					util.NullValue,
					util.NullValue,
					util.NullValue,

}
