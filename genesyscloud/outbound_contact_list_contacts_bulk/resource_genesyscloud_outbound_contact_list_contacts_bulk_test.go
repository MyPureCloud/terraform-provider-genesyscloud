package outbound_contact_list_contacts_bulk

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccResourceOutboundContactListContactsBulk(t *testing.T) {
	var (
		resourceLabel    = "contact_list_contacts_bulk"
		contactListLabel = "contact_list" + uuid.NewString()[:8]

		csvContent = `id,firstName,lastName,phone,email
100,John,Doe,+13175555555,john.doe@example.com
101,Jane,Smith,+13175555556,jane.smith@example.com`

		// New CSV content with different data
		updatedCsvContent = `id,firstName,lastName,phone,email
100,John,Doe,+13175555555,john.doe@example.com
101,Jane,Smith,+13175555556,jane.smith@example.com
102,Bob,Johnson,+13175555557,bob.johnson@example.com`
	)

	// Create temporary CSV file
	tempDir := t.TempDir()
	csvFile := filepath.Join(tempDir, "contacts.csv")

	err := os.WriteFile(csvFile, []byte(csvContent), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// Calculate initial file hash
	hash := sha256.Sum256([]byte(csvContent))
	expectedHash := hex.EncodeToString(hash[:])

	// Calculate updated file hash
	updatedHash := sha256.Sum256([]byte(updatedCsvContent))
	expectedUpdatedHash := hex.EncodeToString(updatedHash[:])

	outboundContactListResourceConfigs := generateOutboundContactList(contactListLabel)
	outboundContactListContactsBulkResourceConfigs := generateOutboundContactListContactsBulk(
		resourceLabel,
		contactListLabel,
		csvFile,
		"id",
	)
	generatedResourceConfigs := outboundContactListResourceConfigs + outboundContactListContactsBulkResourceConfigs

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			util.TestAccPreCheck(t)
		},
		ProviderFactories: provider.GetProviderFactories(providerResources, nil),
		Steps: []resource.TestStep{
			{
				// Initial creation
				Config: generatedResourceConfigs,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"genesyscloud_outbound_contact_list_contacts_bulk."+resourceLabel,
						"contact_list_id",
					),
					resource.TestCheckResourceAttr(
						"genesyscloud_outbound_contact_list_contacts_bulk."+resourceLabel,
						"filepath",
						csvFile,
					),
					resource.TestCheckResourceAttr(
						"genesyscloud_outbound_contact_list_contacts_bulk."+resourceLabel,
						"file_content_hash",
						expectedHash,
					),
					resource.TestCheckResourceAttr(
						"genesyscloud_outbound_contact_list_contacts_bulk."+resourceLabel,
						"record_count",
						"2",
					),
				),
			},
			{
				// Update with new file content
				PreConfig: func() {
					// Update the CSV file content
					err := os.WriteFile(csvFile, []byte(updatedCsvContent), 0644)
					if err != nil {
						t.Fatal(err)
					}
				},
				Config: generatedResourceConfigs,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"genesyscloud_outbound_contact_list_contacts_bulk."+resourceLabel,
						"file_content_hash",
						expectedUpdatedHash,
					),
					resource.TestCheckResourceAttr(
						"genesyscloud_outbound_contact_list_contacts_bulk."+resourceLabel,
						"record_count",
						"3",
					),
				),
			},
			{
				Config: outboundContactListResourceConfigs,
				Check: resource.TestCheckResourceAttrWith(
					"genesyscloud_outbound_contact_list."+contactListLabel,
					"id",
					// Check to ensure that the records have been cleared out when resource is destroyed
					func(contactListId string) error {
						ctx := context.Background()
						contactListBulkProxy := &contactsBulkProxy{}
						providerMeta := provider.GetProviderMeta()
						contactListBulkProxy = newBulkContactProxy(providerMeta.ClientConfig)
						_, recordCount, _, err := contactListBulkProxy.readContactListAndRecordLengthById(ctx, contactListId)
						if err != nil {
							return err
						}
						if recordCount != 0 {
							return fmt.Errorf("Expected record count to be 0 (cleared), got %d", recordCount)
						}
						return nil
					}),
			},
			{
				Config:       "",
				RefreshState: true,
				Check:        nil,
			},
		},
	})
}

func generateOutboundContactList(resourceLabel string) string {
	return fmt.Sprintf(`
resource "genesyscloud_outbound_contact_list" "%s" {
    name = "Test Contact List %s"
    column_names = ["id", "firstName", "lastName", "phone", "email"]
    phone_columns {
        column_name = "phone"
        type       = "home"
    }
}
`, resourceLabel, resourceLabel)

}

func generateOutboundContactListContactsBulk(resourceLabel, contactListLabel, csvFilePath, contactIdName string) string {
	return fmt.Sprintf(`
resource "genesyscloud_outbound_contact_list_contacts_bulk" "%s" {
    contact_list_id = genesyscloud_outbound_contact_list.%s.id
    filepath        = "%s"
    contact_id_name = "%s"
}
`, resourceLabel, contactListLabel, csvFilePath, contactIdName)
}
