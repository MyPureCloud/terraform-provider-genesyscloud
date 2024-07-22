package util

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/nyaruka/phonenumbers"
)

// Formats string as a valid E.164 international standard telephone format.
// Use this when data is coming from the user so we can appropriately error.
func FormatAsValidE164Number(number string) (string, diag.Diagnostics) {
	phoneNumber, err := phonenumbers.Parse(number, "US")
	if err != nil {
		return "", diag.Errorf("Failed to format phone number %s: %s", number, err)
	}
	formattedNum := phonenumbers.Format(phoneNumber, phonenumbers.E164)
	return formattedNum, nil
}

// Formats string as a valid E.164 international standard telephone format.
// Use this function when data is being returned back from the API already in expected format
func FormatAsCalculatedE164Number(number string) string {
	phoneNumber, _ := phonenumbers.Parse(number, "US")
	formattedNum := phonenumbers.Format(phoneNumber, phonenumbers.E164)
	return formattedNum
}
