package errors

import (
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
)

func ContainsPermissionsErrorOnly(err diag.Diagnostics) bool {
	foundPermissionsError := false
	for _, v := range err {
		if strings.Contains(v.Summary, "403") ||
			strings.Contains(v.Summary, "501") {
			foundPermissionsError = true
		} else {
			return false
		}
	}
	return foundPermissionsError
}
