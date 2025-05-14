package util

import (
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/nyaruka/phonenumbers"
)

type UtilE164Service struct {
	GetDefaultCountryCodeFunc func() string
}

func NewUtilE164Service() *UtilE164Service {
	return &UtilE164Service{GetDefaultCountryCodeFunc: provider.GetOrgDefaultCountryCode}
}

// Formats string as a valid E.164 international standard telephone format, parsing the number with
// a default region that matches the default country code set on the GC organization.
// Use this when data is coming from the user so we can appropriately error.
func (m *UtilE164Service) FormatAsValidE164Number(number string) (string, diag.Diagnostics) {
	defaultLang := m.GetDefaultCountryCodeFunc()
	if defaultLang == "" {
		defaultLang = "US"
	}
	log.Printf("Default language is %s", defaultLang)
	phoneNumber, err := phonenumbers.Parse(number, defaultLang)
	if err != nil {
		return "", diag.Errorf("Failed to format phone number %s: %s", number, err)
	}
	formattedNum := phonenumbers.Format(phoneNumber, phonenumbers.E164)
	return formattedNum, nil
}

// Formats string as a valid E.164 international standard telephone format, parsing the number with
// a default region that matches the default country code set on the GC organization.
// Use this function when data is being returned back from the API already in expected format
func (m *UtilE164Service) FormatAsCalculatedE164Number(number string) string {
	defaultLang := m.GetDefaultCountryCodeFunc()
	if defaultLang == "" {
		defaultLang = "US"
	}
	log.Printf("Default language is %s", defaultLang)
	phoneNumber, _ := phonenumbers.Parse(number, defaultLang)
	formattedNum := phonenumbers.Format(phoneNumber, phonenumbers.E164)
	return formattedNum
}
