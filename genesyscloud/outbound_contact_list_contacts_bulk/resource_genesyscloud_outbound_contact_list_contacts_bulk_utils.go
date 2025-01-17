package outbound_contact_list_contacts_bulk

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// TODO: FIX
func GenerateOutboundContactListContactsBulk(
	resourceLabel,
	contactListId,
	contactId,
	callable,
	data string,
	nestedBlocks ...string,
) string {
	return fmt.Sprintf(`resource "%s" "%s" {
		contact_list_id = %s
		contact_id = "%s"
    callable        = %s
    %s
    %s
}`, ResourceType, resourceLabel, contactListId, contactId, callable, data, strings.Join(nestedBlocks, "\n"))
}

func fileContentHashChanged(ctx context.Context, d *schema.ResourceDiff, meta interface{}) bool {
	filepath := d.Get("filepath").(string)

	newHash, err := fileContentHashReader(filepath)
	if err != nil {
		log.Printf("Error calculating file content hash: %v", err)
		return false
	}

	// Get the current hash value
	oldHash := d.Get("file_content_hash").(string)

	// Return true if the hashes are different
	return oldHash != newHash
}

func fileContentHashReader(filepath string) (string, error) {
	// Read file content
	content, err := os.ReadFile(filepath)
	if err != nil {
		log.Printf("Error reading file content: %v", err)
		return "", err
	}

	// Calculate SHA256 hash of file content
	hasher := sha256.New()
	hasher.Write(content)
	hash := hex.EncodeToString(hasher.Sum(nil))

	return hash, nil
}

// func importTemplateAttributesSchemaLogic() schema.CustomizeDiffFunc {
// 	return func(ctx context.Context, diff *schema.ResourceDiff, i interface{}) error {
// 		templateID := diff.Get("contact_list_template_id").(string)
// 		contactListID := diff.Get("contact_list_id").(string)
// 		listNamePrefix := diff.Get("list_name_prefix").(string)
// 		divisionID := diff.Get("division_id_for_target_contact_lists").(string)

// 		if templateID != "" {
// 			// If template ID is set, list_name_prefix must be set
// 			if listNamePrefix == "" {
// 				return fmt.Errorf("list_name_prefix is required when contact_list_template_id is set")
// 			}
// 			// contact_list_id should not be set
// 			if contactListID != "" {
// 				return fmt.Errorf("contact_list_id cannot be set when using contact_list_template_id")
// 			}
// 		} else if contactListID != "" {
// 			// If contact_list_id is set, template-related attributes should not be set
// 			if listNamePrefix != "" {
// 				return fmt.Errorf("list_name_prefix cannot be set when using contact_list_id")
// 			}
// 			if divisionID != "" {
// 				return fmt.Errorf("division_id_for_target_contact_lists cannot be set when using contact_list_id")
// 			}
// 		}

// 		return nil
// 	}
// }

// We add the extra suffix to the id in order to prevent conflicts with actual contact lists
func buildBulkContactId(contactListId string) string {
	return fmt.Sprintf("%s_contacts_bulk", contactListId)
}
