package external_user

import (
	"fmt"
	"net/url"
	"strings"
)

// createCompoundKey builds the Terraform resource id from user_id, authority_name, and external_key.
// Each segment is url.PathEscape'd so '/' and other reserved characters in authority_name or
// external_key do not break parsing: splitCompoundKey splits only on the two delimiter slashes
// between segments, not on slashes inside escaped values.
func createCompoundKey(userId, authorityName, externalKey string) string {
	return fmt.Sprintf("%s/%s/%s",
		url.PathEscape(userId),
		url.PathEscape(authorityName),
		url.PathEscape(externalKey))
}

func splitCompoundKey(compoundKey string) (string, string, string, error) {
	split := strings.Split(compoundKey, "/")
	if len(split) != 3 {
		return "", "", "", fmt.Errorf("invalid compound key: %s", compoundKey)
	}
	userId, err := url.PathUnescape(split[0])
	if err != nil {
		return "", "", "", fmt.Errorf("invalid compound key: %s", compoundKey)
	}
	authorityName, err := url.PathUnescape(split[1])
	if err != nil {
		return "", "", "", fmt.Errorf("invalid compound key: %s", compoundKey)
	}
	externalKey, err := url.PathUnescape(split[2])
	if err != nil {
		return "", "", "", fmt.Errorf("invalid compound key: %s", compoundKey)
	}
	return userId, authorityName, externalKey, nil
}

func generateExternalUserIdentity(resourceLabel, userId, authorityName, externalKey string) string {
	return fmt.Sprintf(`resource "genesyscloud_externalusers_identity" "%s" {
        user_id = %s
        authority_name = "%s"
        external_key = "%s"
	}
	`, resourceLabel, userId, authorityName, externalKey)
}
