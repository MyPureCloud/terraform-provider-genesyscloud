package provider

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

// ProviderFactories are used to instantiate a provider during acceptance testing.
// The factory function will be invoked for every Terraform CLI command executed
// to create a provider server to which the CLI can reattach.

func GetProviderFactories(providerResources map[string]*schema.Resource, providerDataSources map[string]*schema.Resource) map[string]func() (*schema.Provider, error) {
	return map[string]func() (*schema.Provider, error){
		"genesyscloud": func() (*schema.Provider, error) {
			provider := New("0.1.0", providerResources, providerDataSources)()
			return provider, nil
		},
	}
}

// TestDefaultHomeDivision Verify default division is home division
func TestDefaultHomeDivision(resource string) resource.TestCheckFunc {
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

// validateLogFilePath validates that a log file path is not empty, does
// not contain any whitespaces, and that it ends with ".log"
// (Keeping this inside validators causes import cycle)
func validateLogFilePath(filepath any, _ cty.Path) (err diag.Diagnostics) {
	defer func() {
		if err != nil {
			err = diag.Errorf("validateLogFilePath failed: %v", err)
		}
	}()

	val, ok := filepath.(string)
	if !ok {
		return diag.Errorf("expected type of %v to be string, got %T", filepath, filepath)
	}

	// Check if the string is empty or contains any whitespace
	if val == "" || strings.ContainsAny(val, " \t\n\r") {
		return diag.Errorf("filepath must not be empty or contain whitespace, got: %s", val)
	}

	// Check if the file ends with .log
	if !strings.HasSuffix(val, ".log") {
		return diag.Errorf("%s must end with .log extension", val)
	}

	return err
}

// determineTokenPoolSize returns the token pool size based on a precedence order:
// 1. Value from ResourceData if set
// 2. Value from environment variable GENESYSCLOUD_TOKEN_POOL_SIZE if set and valid
// 3. Default value (tokenPoolSizeDefault)
//
// The token pool size determines the number of OAuth tokens that can be cached
// simultaneously for API authentication.
//
// Parameters:
//   - d: *schema.ResourceData containing the provider configuration
//
// Returns:
//   - int: The determined token pool size
func determineTokenPoolSize(d *schema.ResourceData) int {
	var tokenPoolSize = int(tokenPoolSizeDefault)

	if tps, ok := d.GetOk("token_pool_size"); ok {
		tokenPoolSize = tps.(int)
	} else {
		tpsStr, ok := os.LookupEnv(tokenPoolSizeEnvVar)
		if !ok {
			return tokenPoolSize
		}
		// Convert environment variable string to int
		parsedSize, err := strconv.ParseInt(tpsStr, 10, 64)
		if err != nil {
			log.Printf("failed to parse env var %s: %s", tokenPoolSizeEnvVar, err.Error())
			return tokenPoolSize
		}
		tokenPoolSize = int(parsedSize)
	}

	return tokenPoolSize
}

func determineProviderStringAttribute(d *schema.ResourceData, attrName, envVar string) string {
	if schemaVal, ok := d.GetOk(attrName); ok {
		return schemaVal.(string)
	}
	return os.Getenv(envVar)
}

func determineProviderStringAttributeWithDefaultFallback(d *schema.ResourceData, attrName, envVar, defaultValue string) string {
	if schemaVal, ok := d.GetOk(attrName); ok {
		return schemaVal.(string)
	}
	envVal, ok := os.LookupEnv(envVar)
	if ok {
		return envVal
	}
	return defaultValue
}

// Ensure the Meta (with ClientCredentials) is accessible throughout the provider, especially
// within acceptance testing
var (
	providerMeta *ProviderMeta
	mutex        sync.RWMutex
)

func GetProviderMeta() *ProviderMeta {
	mutex.RLock()
	defer mutex.RUnlock()
	return providerMeta
}

func setProviderMeta(p *ProviderMeta) {
	mutex.Lock()
	defer mutex.Unlock()
	providerMeta = p
}

func GetOrgDefaultCountryCode() string {
	meta := GetProviderMeta()
	if meta == nil {
		return ""
	}
	return meta.DefaultCountryCode
}
