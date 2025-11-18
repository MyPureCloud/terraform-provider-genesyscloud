package util

import (
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func SuppressCertificateDiff(k, old, new string, d *schema.ResourceData) bool {
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
