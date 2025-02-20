package external_user

import (
	"fmt"
	"strings"
)

func createCompoundKey(userId, authorityName, externalKey string) string {
	return fmt.Sprintf("%s/%s/%s", userId, authorityName, externalKey)
}

func splitCompoundKey(compoundKey string) (string, string, string, error) {
	split := strings.Split(compoundKey, "/")
	if len(split) != 3 {
		return "", "", "", fmt.Errorf("invalid compound key: %s", compoundKey)
	}
	return split[0], split[1], split[2], nil
}
func generateExternalUserIdentity(resourceLabel, userId, authorityName, externalKey string) string {
	return fmt.Sprintf(`resource "genesyscloud_externalusers_identity" "%s" {
        user_id = %s
        authority_name = "%s"
        external_key = "%s"
	}
	`, resourceLabel, userId, authorityName, externalKey)
}
