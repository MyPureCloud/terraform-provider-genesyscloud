package util

import (
	"fmt"
	"log"
	"regexp"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/nyaruka/phonenumbers"
)

type UtilE164Service struct {
	GetDefaultCountryCodeFunc func() string
}

func NewUtilE164Service() *UtilE164Service {
	return &UtilE164Service{GetDefaultCountryCodeFunc: provider.GetOrgDefaultCountryCode}
}

// Validates a string as a valid E.164 international standard telephone format, with any viable country code accepted
// Use this when data is coming from the user so we can appropriately error.
func (m *UtilE164Service) IsValidE164Number(number string) (bool, diag.Diagnostics) {

	defaultRegion := m.GetDefaultCountryCodeFunc()
	if defaultRegion == "" {
		defaultRegion = "ZZ" // Unknown region
	}

	STARTING_PLUS_CHARS_REGEX := regexp.MustCompile("^[+\uFF0B]")

	matchValidStartingNumber := STARTING_PLUS_CHARS_REGEX.MatchString(number)
	if !matchValidStartingNumber {
		return false, diag.Errorf("Phone number must start with a '+'")
	}

	phoneNumber, err := phonenumbers.Parse(number, defaultRegion)

	if err != nil {
		return false, diag.Errorf("Failed to format phone number %s: %s", number, err)
	}

	return phonenumbers.IsPossibleNumber(phoneNumber), nil
}

// Formats string as a valid E.164 international standard telephone format, parsing the number with
// a default region that matches the default country code set on the GC organization.
// Use this function when data is being returned back from the API already in expected format
func (m *UtilE164Service) FormatAsCalculatedE164Number(number string) string {
	if number == "" {
		return ""
	}
	defaultLang := m.GetDefaultCountryCodeFunc()
	if defaultLang == "" {
		defaultLang = "ZZ"
	}
	log.Printf("Default language is %s", defaultLang)
	phoneNumber, _ := phonenumbers.Parse(number, defaultLang)
	formattedNum := phonenumbers.Format(phoneNumber, phonenumbers.E164)
	// +00 means the number was unable to be formatted. Since it's coming from the API,
	// we will just return the value directly, ensuring there is a leading + sign.
	if formattedNum == "+00" {
		if number[0:1] != "+" {
			formattedNum = fmt.Sprintf("+%s", number)
		} else {
			formattedNum = number
		}
	}
	return formattedNum
}
