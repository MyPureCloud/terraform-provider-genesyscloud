package authorization_product

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var providerDataSources map[string]*schema.Resource
var providerResources map[string]*schema.Resource

func initTestResources() {
	providerDataSources = make(map[string]*schema.Resource)
	providerResources = make(map[string]*schema.Resource)

	regInstance := &registerTestInstance{}
	regInstance.registerTestDataSources()
}

type registerTestInstance struct {
}

func (r *registerTestInstance) registerTestDataSources() {
	providerDataSources["genesyscloud_authorization_product"] = DataSourceAuthorizationProduct()
}

func TestMain(m *testing.M) {
	// Run setup function before starting the test suite
	initTestResources()

	// Run the test suite for outbound ruleset
	m.Run()
}
