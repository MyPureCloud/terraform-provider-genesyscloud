package outbound_contact_list_contact

import (
	"fmt"
	"strconv"
	outboundContactList "terraform-provider-genesyscloud/genesyscloud/outbound_contact_list"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccResourceOutboundContactListContact(t *testing.T) {
	var (
		resourceLabel     = "contact"
		fullResourceLabel = fmt.Sprintf("%s.%s", ResourceType, resourceLabel)

		cellColumnKey        = "Cell"
		dataCellValue        = "+000000"
		dataCellValueUpdated = "+111111"

		homeColumnKey        = "Home"
		dataHomeValue        = "+22222222"
		dataHomeValueUpdated = "+33333333"

		emailColumnKey        = "Email"
		dataEmailValue        = "email@fake.com"
		dataEmailValueUpdated = "fake@email.cmo"

		contactListResourceLabel     = "contact_list"
		contactListFullResourceLabel = "genesyscloud_outbound_contact_list." + contactListResourceLabel
		contactListName              = "tf test contact list " + uuid.NewString()
		columnNames                  = []string{
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
		contactListResourceLabel,
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
				Config: contactListResource + GenerateOutboundContactListContact(
					resourceLabel,
					contactListFullResourceLabel+".id",
					util.NullValue,
					util.TrueValue,
					util.GenerateMapAttrWithMapProperties(
						"data",
						map[string]string{
							cellColumnKey:  strconv.Quote(dataCellValue),
							homeColumnKey:  strconv.Quote(dataHomeValue),
							emailColumnKey: strconv.Quote(dataEmailValue),
						},
					),
					GeneratePhoneNumberStatus(cellColumnKey, util.FalseValue),
					GeneratePhoneNumberStatus(homeColumnKey, util.TrueValue),
					GenerateContactableStatus(
						voiceMediaType,
						util.FalseValue, // contactable
						GenerateColumnStatus(cellColumnKey, util.FalseValue),
					),
					GenerateContactableStatus(
						emailMediaType,
						util.TrueValue, // contactable
						GenerateColumnStatus(emailColumnKey, util.TrueValue),
					),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fullResourceLabel, "callable", util.TrueValue),
					resource.TestCheckResourceAttrPair(fullResourceLabel, "contact_list_id", contactListFullResourceLabel, "id"),
					resource.TestCheckResourceAttrSet(fullResourceLabel, "contact_id"),
					resource.TestCheckResourceAttr(fullResourceLabel, "data."+cellColumnKey, dataCellValue),
					resource.TestCheckResourceAttr(fullResourceLabel, "data."+homeColumnKey, dataHomeValue),
					resource.TestCheckResourceAttr(fullResourceLabel, "data."+emailColumnKey, dataEmailValue),
					resource.TestCheckResourceAttr(fullResourceLabel, "phone_number_status.#", "2"),
					resource.TestCheckResourceAttr(fullResourceLabel, "phone_number_status.0.key", cellColumnKey),
					resource.TestCheckResourceAttr(fullResourceLabel, "phone_number_status.0.callable", util.FalseValue),
					resource.TestCheckResourceAttr(fullResourceLabel, "phone_number_status.1.key", homeColumnKey),
					resource.TestCheckResourceAttr(fullResourceLabel, "phone_number_status.1.callable", util.TrueValue),
					resource.TestCheckResourceAttr(fullResourceLabel, "contactable_status.#", "2"),
					resource.TestCheckResourceAttr(fullResourceLabel, "contactable_status.0.media_type", voiceMediaType),
					resource.TestCheckResourceAttr(fullResourceLabel, "contactable_status.0.contactable", util.FalseValue),
					resource.TestCheckResourceAttr(fullResourceLabel, "contactable_status.0.column_status.0.column", cellColumnKey),
					resource.TestCheckResourceAttr(fullResourceLabel, "contactable_status.0.column_status.0.contactable", util.FalseValue),
					resource.TestCheckResourceAttr(fullResourceLabel, "contactable_status.1.media_type", emailMediaType),
					resource.TestCheckResourceAttr(fullResourceLabel, "contactable_status.1.contactable", util.TrueValue),
					resource.TestCheckResourceAttr(fullResourceLabel, "contactable_status.1.column_status.0.column", emailColumnKey),
					resource.TestCheckResourceAttr(fullResourceLabel, "contactable_status.1.column_status.0.contactable", util.TrueValue),
				),
			},
			{
				// Update
				Config: contactListResource + GenerateOutboundContactListContact(
					resourceLabel,
					contactListFullResourceLabel+".id",
					util.NullValue,
					util.FalseValue,
					util.GenerateMapAttrWithMapProperties(
						"data",
						map[string]string{
							cellColumnKey:  strconv.Quote(dataCellValueUpdated),
							homeColumnKey:  strconv.Quote(dataHomeValueUpdated),
							emailColumnKey: strconv.Quote(dataEmailValueUpdated),
						},
					),
					GeneratePhoneNumberStatus(cellColumnKey, util.FalseValue),
					GeneratePhoneNumberStatus(homeColumnKey, util.TrueValue),
					GenerateContactableStatus(
						voiceMediaType,
						util.FalseValue, // contactable
						GenerateColumnStatus(cellColumnKey, util.FalseValue),
					),
					GenerateContactableStatus(
						emailMediaType,
						util.TrueValue, // contactable
						GenerateColumnStatus(emailColumnKey, util.TrueValue),
					),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fullResourceLabel, "callable", util.FalseValue),
					resource.TestCheckResourceAttrPair(fullResourceLabel, "contact_list_id", contactListFullResourceLabel, "id"),
					resource.TestCheckResourceAttrSet(fullResourceLabel, "contact_id"),
					resource.TestCheckResourceAttr(fullResourceLabel, "data."+cellColumnKey, dataCellValueUpdated),
					resource.TestCheckResourceAttr(fullResourceLabel, "data."+homeColumnKey, dataHomeValueUpdated),
					resource.TestCheckResourceAttr(fullResourceLabel, "data."+emailColumnKey, dataEmailValueUpdated),
					resource.TestCheckResourceAttr(fullResourceLabel, "phone_number_status.#", "2"),
					resource.TestCheckResourceAttr(fullResourceLabel, "phone_number_status.0.key", cellColumnKey),
					resource.TestCheckResourceAttr(fullResourceLabel, "phone_number_status.0.callable", util.FalseValue),
					resource.TestCheckResourceAttr(fullResourceLabel, "phone_number_status.1.key", homeColumnKey),
					resource.TestCheckResourceAttr(fullResourceLabel, "phone_number_status.1.callable", util.TrueValue),
					resource.TestCheckResourceAttr(fullResourceLabel, "contactable_status.#", "2"),
					resource.TestCheckResourceAttr(fullResourceLabel, "contactable_status.0.media_type", voiceMediaType),
					resource.TestCheckResourceAttr(fullResourceLabel, "contactable_status.0.contactable", util.FalseValue),
					resource.TestCheckResourceAttr(fullResourceLabel, "contactable_status.0.column_status.0.column", cellColumnKey),
					resource.TestCheckResourceAttr(fullResourceLabel, "contactable_status.0.column_status.0.contactable", util.FalseValue),
					resource.TestCheckResourceAttr(fullResourceLabel, "contactable_status.1.media_type", emailMediaType),
					resource.TestCheckResourceAttr(fullResourceLabel, "contactable_status.1.contactable", util.TrueValue),
					resource.TestCheckResourceAttr(fullResourceLabel, "contactable_status.1.column_status.0.column", emailColumnKey),
					resource.TestCheckResourceAttr(fullResourceLabel, "contactable_status.1.column_status.0.contactable", util.TrueValue),
				),
			},
		},
	})
}

func TestAccResourceOutboundContactListContactWithId(t *testing.T) {
	var (
		resourceLabel     = "contact"
		fullResourceLabel = fmt.Sprintf("%s.%s", ResourceType, resourceLabel)
		contactId         = uuid.NewString()

		cellColumnKey        = "Cell"
		dataCellValue        = "+000000"
		dataCellValueUpdated = "+111111"

		homeColumnKey        = "Home"
		dataHomeValue        = "+22222222"
		dataHomeValueUpdated = "+33333333"

		emailColumnKey        = "Email"
		dataEmailValue        = "email@fake.com"
		dataEmailValueUpdated = "fake@email.cmo"

		contactListResourceLabel     = "contact_list"
		contactListFullResourceLabel = "genesyscloud_outbound_contact_list." + contactListResourceLabel
		contactListName              = "tf test contact list " + uuid.NewString()
		columnNames                  = []string{
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
		contactListResourceLabel,
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
				Config: contactListResource + GenerateOutboundContactListContact(
					resourceLabel,
					contactListFullResourceLabel+".id",
					contactId,
					util.TrueValue,
					util.GenerateMapAttrWithMapProperties(
						"data",
						map[string]string{
							cellColumnKey:  strconv.Quote(dataCellValue),
							homeColumnKey:  strconv.Quote(dataHomeValue),
							emailColumnKey: strconv.Quote(dataEmailValue),
						},
					),
					GeneratePhoneNumberStatus(cellColumnKey, util.FalseValue),
					GeneratePhoneNumberStatus(homeColumnKey, util.TrueValue),
					GenerateContactableStatus(
						voiceMediaType,
						util.FalseValue, // contactable
						GenerateColumnStatus(cellColumnKey, util.FalseValue),
					),
					GenerateContactableStatus(
						emailMediaType,
						util.TrueValue, // contactable
						GenerateColumnStatus(emailColumnKey, util.TrueValue),
					),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fullResourceLabel, "callable", util.TrueValue),
					resource.TestCheckResourceAttrPair(fullResourceLabel, "contact_list_id", contactListFullResourceLabel, "id"),
					resource.TestCheckResourceAttr(fullResourceLabel, "contact_id", contactId),
					resource.TestCheckResourceAttr(fullResourceLabel, "data."+cellColumnKey, dataCellValue),
					resource.TestCheckResourceAttr(fullResourceLabel, "data."+homeColumnKey, dataHomeValue),
					resource.TestCheckResourceAttr(fullResourceLabel, "data."+emailColumnKey, dataEmailValue),
					resource.TestCheckResourceAttr(fullResourceLabel, "phone_number_status.#", "2"),
					resource.TestCheckResourceAttr(fullResourceLabel, "phone_number_status.0.key", cellColumnKey),
					resource.TestCheckResourceAttr(fullResourceLabel, "phone_number_status.0.callable", util.FalseValue),
					resource.TestCheckResourceAttr(fullResourceLabel, "phone_number_status.1.key", homeColumnKey),
					resource.TestCheckResourceAttr(fullResourceLabel, "phone_number_status.1.callable", util.TrueValue),
					resource.TestCheckResourceAttr(fullResourceLabel, "contactable_status.#", "2"),
					resource.TestCheckResourceAttr(fullResourceLabel, "contactable_status.0.media_type", voiceMediaType),
					resource.TestCheckResourceAttr(fullResourceLabel, "contactable_status.0.contactable", util.FalseValue),
					resource.TestCheckResourceAttr(fullResourceLabel, "contactable_status.0.column_status.0.column", cellColumnKey),
					resource.TestCheckResourceAttr(fullResourceLabel, "contactable_status.0.column_status.0.contactable", util.FalseValue),
					resource.TestCheckResourceAttr(fullResourceLabel, "contactable_status.1.media_type", emailMediaType),
					resource.TestCheckResourceAttr(fullResourceLabel, "contactable_status.1.contactable", util.TrueValue),
					resource.TestCheckResourceAttr(fullResourceLabel, "contactable_status.1.column_status.0.column", emailColumnKey),
					resource.TestCheckResourceAttr(fullResourceLabel, "contactable_status.1.column_status.0.contactable", util.TrueValue),
				),
			},
			{
				// Update
				Config: contactListResource + GenerateOutboundContactListContact(
					resourceLabel,
					contactListFullResourceLabel+".id",
					contactId,
					util.FalseValue,
					util.GenerateMapAttrWithMapProperties(
						"data",
						map[string]string{
							cellColumnKey:  strconv.Quote(dataCellValueUpdated),
							homeColumnKey:  strconv.Quote(dataHomeValueUpdated),
							emailColumnKey: strconv.Quote(dataEmailValueUpdated),
						},
					),
					GeneratePhoneNumberStatus(cellColumnKey, util.FalseValue),
					GeneratePhoneNumberStatus(homeColumnKey, util.TrueValue),
					GenerateContactableStatus(
						voiceMediaType,
						util.FalseValue, // contactable
						GenerateColumnStatus(cellColumnKey, util.FalseValue),
					),
					GenerateContactableStatus(
						emailMediaType,
						util.TrueValue, // contactable
						GenerateColumnStatus(emailColumnKey, util.TrueValue),
					),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fullResourceLabel, "callable", util.FalseValue),
					resource.TestCheckResourceAttrPair(fullResourceLabel, "contact_list_id", contactListFullResourceLabel, "id"),
					resource.TestCheckResourceAttr(fullResourceLabel, "contact_id", contactId),
					resource.TestCheckResourceAttr(fullResourceLabel, "data."+cellColumnKey, dataCellValueUpdated),
					resource.TestCheckResourceAttr(fullResourceLabel, "data."+homeColumnKey, dataHomeValueUpdated),
					resource.TestCheckResourceAttr(fullResourceLabel, "data."+emailColumnKey, dataEmailValueUpdated),
					resource.TestCheckResourceAttr(fullResourceLabel, "phone_number_status.#", "2"),
					resource.TestCheckResourceAttr(fullResourceLabel, "phone_number_status.0.key", cellColumnKey),
					resource.TestCheckResourceAttr(fullResourceLabel, "phone_number_status.0.callable", util.FalseValue),
					resource.TestCheckResourceAttr(fullResourceLabel, "phone_number_status.1.key", homeColumnKey),
					resource.TestCheckResourceAttr(fullResourceLabel, "phone_number_status.1.callable", util.TrueValue),
					resource.TestCheckResourceAttr(fullResourceLabel, "contactable_status.#", "2"),
					resource.TestCheckResourceAttr(fullResourceLabel, "contactable_status.0.media_type", voiceMediaType),
					resource.TestCheckResourceAttr(fullResourceLabel, "contactable_status.0.contactable", util.FalseValue),
					resource.TestCheckResourceAttr(fullResourceLabel, "contactable_status.0.column_status.0.column", cellColumnKey),
					resource.TestCheckResourceAttr(fullResourceLabel, "contactable_status.0.column_status.0.contactable", util.FalseValue),
					resource.TestCheckResourceAttr(fullResourceLabel, "contactable_status.1.media_type", emailMediaType),
					resource.TestCheckResourceAttr(fullResourceLabel, "contactable_status.1.contactable", util.TrueValue),
					resource.TestCheckResourceAttr(fullResourceLabel, "contactable_status.1.column_status.0.column", emailColumnKey),
					resource.TestCheckResourceAttr(fullResourceLabel, "contactable_status.1.column_status.0.contactable", util.TrueValue),
				),
			},
		},
	})
}
