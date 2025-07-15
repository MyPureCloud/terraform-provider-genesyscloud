package provider

import (
	"fmt"
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

func CombineProviderFactories(providers ...map[string]func() (*schema.Provider, error)) map[string]func() (*schema.Provider, error) {
	combined := map[string]func() (*schema.Provider, error){}
	for _, provider := range providers {
		for k, v := range provider {
			combined[k] = v
		}
	}
	return combined
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

// Ensure the Meta (with ClientCredentials) is accessible throughout the provider, especially
// within acceptance testing
var (
	providerMeta   *ProviderMeta
	mutex          sync.RWMutex
	providerConfig *schema.ResourceData
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

func GetProviderConfig() *schema.ResourceData {
	mutex.RLock()
	defer mutex.RUnlock()
	return providerConfig
}

func setProviderConfig(p *schema.ResourceData) {
	mutex.Lock()
	defer mutex.Unlock()
	providerConfig = p
}

func GetOrgDefaultCountryCode() string {
	meta := GetProviderMeta()
	if meta == nil {
		return ""
	}
	return meta.DefaultCountryCode
}
