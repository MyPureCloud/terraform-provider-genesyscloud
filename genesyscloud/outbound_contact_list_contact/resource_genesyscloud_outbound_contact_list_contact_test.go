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
		resourceId     = "contact"
		fullResourceId = fmt.Sprintf("%s.%s", resourceName, resourceId)

		cellColumnKey        = "Cell"
		dataCellValue        = "+000000"
		dataCellValueUpdated = "+111111"

		homeColumnKey        = "Home"
		dataHomeValue        = "+22222222"
		dataHomeValueUpdated = "+33333333"

		emailColumnKey        = "Email"
		dataEmailValue        = "email@fake.com"
		dataEmailValueUpdated = "fake@email.cmo"

		contactListResourceId     = "contact_list"
		contactListFullResourceId = "genesyscloud_outbound_contact_list." + contactListResourceId
		contactListName           = "tf test contact list " + uuid.NewString()
		columnNames               = []string{
			strconv.Quote(cellColumnKey),
			strconv.Quote(homeColumnKey),
			strconv.Quote(emailColumnKey),
		}
	)

	const (
		emailMediaType = "Email"
		voiceMediaType = "Voice"
	)

	contactListResource := outboundContactList.GenerateOutboundContactList(
		contactListResourceId,
		contactListName,
		util.NullValue,
		strconv.Quote(cellColumnKey),
		[]string{strconv.Quote(cellColumnKey)},
		columnNames,
		util.FalseValue,
		util.NullValue,
		util.NullValue,
		outboundContactList.GenerateEmailColumnsBlock(
			emailColumnKey,
			"work",
			util.NullValue,
		),
		outboundContactList.GeneratePhoneColumnsBlock(
			cellColumnKey,
			"cell",
			strconv.Quote(cellColumnKey),
		),
		outboundContactList.GeneratePhoneColumnsBlock(
			homeColumnKey,
			"home",
			strconv.Quote(cellColumnKey),
		),
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, nil),
		Steps: []resource.TestStep{
			{
				Config: contactListResource + generateOutboundContactListContact(
					resourceId,
					contactListFullResourceId+".id",
					util.TrueValue,
					util.GenerateMapAttrWithMapProperties(
						"data",
						map[string]string{
							cellColumnKey:  strconv.Quote(dataCellValue),
							homeColumnKey:  strconv.Quote(dataHomeValue),
							emailColumnKey: strconv.Quote(dataEmailValue),
						},
					),
					generatePhoneNumberStatus(cellColumnKey, util.FalseValue),
					generatePhoneNumberStatus(homeColumnKey, util.TrueValue),
					generateContactableStatus(
						voiceMediaType,
						util.FalseValue, // contactable
						generateColumnStatus(cellColumnKey, util.FalseValue),
					),
					generateContactableStatus(
						emailMediaType,
						util.TrueValue, // contactable
						generateColumnStatus(emailColumnKey, util.TrueValue),
					),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fullResourceId, "callable", util.TrueValue),
					resource.TestCheckResourceAttrPair(fullResourceId, "contact_list_id", contactListFullResourceId, "id"),
					resource.TestCheckResourceAttr(fullResourceId, "data."+cellColumnKey, dataCellValue),
					resource.TestCheckResourceAttr(fullResourceId, "data."+homeColumnKey, dataHomeValue),
					resource.TestCheckResourceAttr(fullResourceId, "data."+emailColumnKey, dataEmailValue),
					resource.TestCheckResourceAttr(fullResourceId, "phone_number_status.#", "2"),
					resource.TestCheckResourceAttr(fullResourceId, "phone_number_status.0.key", cellColumnKey),
					resource.TestCheckResourceAttr(fullResourceId, "phone_number_status.0.callable", util.FalseValue),
					resource.TestCheckResourceAttr(fullResourceId, "phone_number_status.1.key", homeColumnKey),
					resource.TestCheckResourceAttr(fullResourceId, "phone_number_status.1.callable", util.TrueValue),
					resource.TestCheckResourceAttr(fullResourceId, "contactable_status.#", "2"),
					resource.TestCheckResourceAttr(fullResourceId, "contactable_status.0.media_type", voiceMediaType),
					resource.TestCheckResourceAttr(fullResourceId, "contactable_status.0.contactable", util.FalseValue),
					resource.TestCheckResourceAttr(fullResourceId, "contactable_status.0.column_status.0.column", cellColumnKey),
					resource.TestCheckResourceAttr(fullResourceId, "contactable_status.0.column_status.0.contactable", util.FalseValue),
					resource.TestCheckResourceAttr(fullResourceId, "contactable_status.1.media_type", emailMediaType),
					resource.TestCheckResourceAttr(fullResourceId, "contactable_status.1.contactable", util.TrueValue),
					resource.TestCheckResourceAttr(fullResourceId, "contactable_status.1.column_status.0.column", emailColumnKey),
					resource.TestCheckResourceAttr(fullResourceId, "contactable_status.1.column_status.0.contactable", util.TrueValue),
				),
			},
			{
				// Update
				Config: contactListResource + generateOutboundContactListContact(
					resourceId,
					contactListFullResourceId+".id",
					util.FalseValue,
					util.GenerateMapAttrWithMapProperties(
						"data",
						map[string]string{
							cellColumnKey:  strconv.Quote(dataCellValueUpdated),
							homeColumnKey:  strconv.Quote(dataHomeValueUpdated),
							emailColumnKey: strconv.Quote(dataEmailValueUpdated),
						},
					),
					generatePhoneNumberStatus(cellColumnKey, util.FalseValue),
					generatePhoneNumberStatus(homeColumnKey, util.TrueValue),
					generateContactableStatus(
						voiceMediaType,
						util.FalseValue, // contactable
						generateColumnStatus(cellColumnKey, util.FalseValue),
					),
					generateContactableStatus(
						emailMediaType,
						util.TrueValue, // contactable
						generateColumnStatus(emailColumnKey, util.TrueValue),
					),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fullResourceId, "callable", util.FalseValue),
					resource.TestCheckResourceAttrPair(fullResourceId, "contact_list_id", contactListFullResourceId, "id"),
					resource.TestCheckResourceAttr(fullResourceId, "data."+cellColumnKey, dataCellValueUpdated),
					resource.TestCheckResourceAttr(fullResourceId, "data."+homeColumnKey, dataHomeValueUpdated),
					resource.TestCheckResourceAttr(fullResourceId, "data."+emailColumnKey, dataEmailValueUpdated),
					resource.TestCheckResourceAttr(fullResourceId, "phone_number_status.#", "2"),
					resource.TestCheckResourceAttr(fullResourceId, "phone_number_status.0.key", cellColumnKey),
					resource.TestCheckResourceAttr(fullResourceId, "phone_number_status.0.callable", util.FalseValue),
					resource.TestCheckResourceAttr(fullResourceId, "phone_number_status.1.key", homeColumnKey),
					resource.TestCheckResourceAttr(fullResourceId, "phone_number_status.1.callable", util.TrueValue),
					resource.TestCheckResourceAttr(fullResourceId, "contactable_status.#", "2"),
					resource.TestCheckResourceAttr(fullResourceId, "contactable_status.0.media_type", voiceMediaType),
					resource.TestCheckResourceAttr(fullResourceId, "contactable_status.0.contactable", util.FalseValue),
					resource.TestCheckResourceAttr(fullResourceId, "contactable_status.0.column_status.0.column", cellColumnKey),
					resource.TestCheckResourceAttr(fullResourceId, "contactable_status.0.column_status.0.contactable", util.FalseValue),
					resource.TestCheckResourceAttr(fullResourceId, "contactable_status.1.media_type", emailMediaType),
					resource.TestCheckResourceAttr(fullResourceId, "contactable_status.1.contactable", util.TrueValue),
					resource.TestCheckResourceAttr(fullResourceId, "contactable_status.1.column_status.0.column", emailColumnKey),
					resource.TestCheckResourceAttr(fullResourceId, "contactable_status.1.column_status.0.contactable", util.TrueValue),
				),
			},
		},
	})
}

func generateOutboundContactListContact(
	resourceId,
	contactListId,
	callable,
	data string,
	nestedBlocks ...string,
) string {
	return fmt.Sprintf(`resource "%s" "%s" {
    contact_list_id = %s
    callable        = %s
    %s
    %s
}`, resourceName, resourceId, contactListId, callable, data, strings.Join(nestedBlocks, "\n"))
}

func generatePhoneNumberStatus(key, callable string) string {
	return fmt.Sprintf(`
	phone_number_status {
		key      = "%s"
        callable = %s
	}`, key, callable)
}

func generateContactableStatus(mediaType, contactable string, nestedBlocks ...string) string {
	return fmt.Sprintf(`
	contactable_status {
		media_type  = "%s"
		contactable = %s
		%s
	}`, mediaType, contactable, strings.Join(nestedBlocks, "\n"))
}

func generateColumnStatus(column, contactable string) string {
	return fmt.Sprintf(`
		column_status {
			column      = "%s"
			contactable = %s
		}`, column, contactable)
}
