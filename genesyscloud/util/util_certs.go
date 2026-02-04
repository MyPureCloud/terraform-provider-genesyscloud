package util

import (
	"log"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func SuppressCertificateDiff(k, old, new string, d *schema.ResourceData) bool {
	lastDotIndex := strings.LastIndex(k, ".")
	if lastDotIndex != -1 {
		k = string(k[:lastDotIndex])
	}
	o, n := d.GetChange(k)
	log.Printf("%s, %s", o, n)

	return normalizeCertificate(old) == normalizeCertificate(new)
}

func normalizeCertificate(cert string) string {
	// Remove PEM headers/footers and whitespace for comparison
	cert = strings.ReplaceAll(cert, "-----BEGIN CERTIFICATE-----", "")
	cert = strings.ReplaceAll(cert, "-----END CERTIFICATE-----", "")
	cert = strings.ReplaceAll(cert, "\n", "")
	cert = strings.ReplaceAll(cert, "\r", "")
	cert = strings.ReplaceAll(cert, " ", "")
	return cert
}
