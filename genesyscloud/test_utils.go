package genesyscloud

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

const (
	nullValue  = "null"
	trueValue  = "true"
	falseValue = "false"
	testCert1  = "MIIDgjCCAmoCCQCY7/3Fvy+CmDANBgkqhkiG9w0BAQsFADCBgjELMAkGA1UEBhMCVVMxCzAJBgNVBAgMAklOMQ8wDQYDVQQHDAZDYXJtZWwxEDAOBgNVBAoMB0dlbmVzeXMxDDAKBgNVBAsMA0RldjEUMBIGA1UEAwwLZ2VuZXN5cy5jb20xHzAdBgkqhkiG9w0BCQEWEHRlc3RAZ2VuZXN5cy5jb20wHhcNMjEwMzI5MTgwOTE0WhcNMjIwMzI5MTgwOTE0WjCBgjELMAkGA1UEBhMCVVMxCzAJBgNVBAgMAklOMQ8wDQYDVQQHDAZDYXJtZWwxEDAOBgNVBAoMB0dlbmVzeXMxDDAKBgNVBAsMA0RldjEUMBIGA1UEAwwLZ2VuZXN5cy5jb20xHzAdBgkqhkiG9w0BCQEWEHRlc3RAZ2VuZXN5cy5jb20wggEiMA0GCSqGSIb3DQEBAQUAA4IBDwAwggEKAoIBAQDUyNj9Z/OEdvvnXD/5F9RO30Nc0/+Ay3TUP/UWzGX2A7xdC7ixMOjm0y1D29gsi5Q48Ury9q0z8cRB8hvixQzs4dH+7kHgIiESkFB/P6N1EfZB2YyHs0cIWzoDEe9m71Lt5M+FGUqFexVQ1c1nA/sJsBvp9P394C1N7+G9DPuhAnto1z8Q67FOvJ1seOJO4y7X7N5dyi5SIqtkHNxn+O+WGvUtpEAaduB9q9QLZPlqQpHs3tyz1D3TOW5Oou6KMhiulQQtd4lkIcBR1vJ9e6N4gXs305F8Bi/0fMgro43rRQYL/dSF8z5nzQENObNlkHicjRLbydkpLAQQu5D9/knNAgMBAAEwDQYJKoZIhvcNAQELBQADggEBADyTiBs6qD76HAtLnsFlrMWen+yXnYL3TPkYGzFH7L7PAkS6zk1w9rMOl4kD3bLUzcv5ndK3YRy2LziBghCgKCKN3QPB+i/z9hSGeg0KVYw5pKniy0QOGZLWXVPO1xpNyWZX6TUX6QQdCkxN6QNbgMQRpQeC121TxrG0Br3wB53ASUub37SwuLCUKmKQIMG9rrUkLjuC6D09+K+zw35CW2PLaG/0tjH1EdV16OJQ2HerNgzjinX95Xadgq6ClCR6M5HpZydipzrzn/gVD+zHmqlecxOQn7P1midH+Bb9k44y9Y+GuivMUQMeQuDbiWcuj/73fLXyYYRFr4dcTcc5ZnY="
	testCert2  = "MIIDnjCCAoYCCQD9X0RdADwPozANBgkqhkiG9w0BAQsFADCBkDELMAkGA1UEBhMCVVMxCzAJBgNVBAgMAkNBMRIwEAYDVQQHDAlEYWx5IENpdHkxEDAOBgNVBAoMB0dlbmVzeXMxEDAOBgNVBAsMB1Byb2R1Y3QxGDAWBgNVBAMMD215cHVyZWNsb3VkLmNvbTEiMCAGCSqGSIb3DQEJARYTbm9yZXBseUBnZW5lc3lzLmNvbTAeFw0yMTAzMzAxMzM5NDJaFw0yMjAzMzAxMzM5NDJaMIGQMQswCQYDVQQGEwJVUzELMAkGA1UECAwCQ0ExEjAQBgNVBAcMCURhbHkgQ2l0eTEQMA4GA1UECgwHR2VuZXN5czEQMA4GA1UECwwHUHJvZHVjdDEYMBYGA1UEAwwPbXlwdXJlY2xvdWQuY29tMSIwIAYJKoZIhvcNAQkBFhNub3JlcGx5QGdlbmVzeXMuY29tMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA6q37OAiuVFCNDejcxv3W3D9iDFUiZc/AtvRzfApH+QPLWyYfCgH5p7n5rOiezs3eY6Do6rvSk/Y9D0LZtafBQ/0TdYTakyc5+Q5rEJoP40DByJht3D9dK7ww8Z6avWYUvbRfNZCHtuykbcUC7RxTZuDKZlf2XV2DzzXYUTqojBKS5HuLkLREU2UhR47a1FEwErqQbNLD7FLsr2AYiP3EtlZDjwluGnRied/eOhVQuVSQ69rSewj2vK1QzMAUGyyaYKbK4xU7AA/gTAiYwGqFj0CPCC1g8NllfB6BDxmYrKD8ypTToJZbTWtOKFH1Wjw72Yi8NM5shXCg3wrsU1842wIDAQABMA0GCSqGSIb3DQEBCwUAA4IBAQC53LaV+RX4cgNUKJxLXybTxiXpY4RTDjX1Y2SPzY6hiqP4sNTiwKiPNCGtF4ySQpCh8QUonPS+a2g3zMZuq5JOtuQhDrebRSEyhy0YnUBPBMmzlBOBpgfXEgK8279bUznRg0MKwFb+67yWqXfoGYQJ3Sep4s94Y7bUJ04/+/P+fK0NUC03Oj5bejKzS9B+PWjJr47+IWzEVijAC8dsax7UUK7RNxGgc/dagWCWo4GNlIuBz946AD32Rx+XoGtIscI/OUsaNld7uLTSD2tygksedsBhrQ/0Sukom1mEAcPyEoYyeGs4izBZh0JdPJBXQ9ZDuj6Z7gNQFizyGK+oZP7p"
)

// Verify default division is home division
func testDefaultHomeDivision(resource string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		homeDivID, err := getHomeDivisionID()
		if err != nil {
			return fmt.Errorf("Failed to query home division: %v", err)
		}

		r := state.RootModule().Resources[resource]
		if r == nil {
			return fmt.Errorf("%s not found in state", resource)
		}

		a := r.Primary.Attributes

		if a["division_id"] != homeDivID {
			return fmt.Errorf("expected division to be home division %s", homeDivID)
		}

		return nil
	}
}

func generateStringArray(vals ...string) string {
	return fmt.Sprintf("[%s]", strings.Join(vals, ","))
}

func validateStringInArray(resourceName string, attrName string, value string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		resource, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Failed to find resource %s in state", resourceName)
		}
		resourceID := resource.Primary.ID

		numAttr, ok := resource.Primary.Attributes[attrName+".#"]
		if !ok {
			return fmt.Errorf("No %s found for %s in state", attrName, resourceID)
		}

		numValues, _ := strconv.Atoi(numAttr)
		for i := 0; i < numValues; i++ {
			if resource.Primary.Attributes[attrName+"."+strconv.Itoa(i)] == value {
				// Found value
				return nil
			}
		}

		return fmt.Errorf("%s %s not found for group %s in state", attrName, value, resourceID)
	}
}
