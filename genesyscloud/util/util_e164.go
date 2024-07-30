package util

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/nyaruka/phonenumbers"
)

func FormatAsE164Number(number string) (string, diag.Diagnostics) {
	phoneNumber, err := phonenumbers.Parse(number, "US")
	if err != nil {
		return "", diag.Errorf("Failed to format phone number %s: %s", number, err)
	}
	formattedNum := phonenumbers.Format(phoneNumber, phonenumbers.E164)
	return formattedNum, nil
}
