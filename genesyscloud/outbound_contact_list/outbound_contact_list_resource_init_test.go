package outbound_contact_list

import (
	gcloud "terraform-provider-genesyscloud/genesyscloud"
	obAttemptLimit "terraform-provider-genesyscloud/genesyscloud/outbound_attempt_limit"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var providerDataSources map[string]*schema.Resource
var providerResources map[string]*schema.Resource

func initTestResources() {
	providerDataSources = make(map[string]*schema.Resource)
	providerResources = make(map[string]*schema.Resource)

	regInstance := &registerTestInstance{}

	regInstance.registerTestResources()
	regInstance.registerTestDataSources()
}

type registerTestInstance struct {
}

func (r *registerTestInstance) registerTestResources() {
	providerResources["genesyscloud_outbound_contact_list"] = ResourceOutboundContactList()
	providerResources["genesyscloud_outbound_attempt_limit"] = obAttemptLimit.ResourceOutboundAttemptLimit()
}

func (r *registerTestInstance) registerTestDataSources() {
	providerDataSources["genesyscloud_outbound_contact_list"] = DataSourceOutboundContactList()
	providerDataSources["genesyscloud_outbound_attempt_limit"] = obAttemptLimit.DataSourceOutboundAttemptLimit()
	providerDataSources["genesyscloud_auth_division_home"] = gcloud.DataSourceAuthDivisionHome()
	providerDataSources["genesyscloud_auth_division_home"] = gcloud.DataSourceAuthDivisionHome()
}

func TestMain(m *testing.M) {
	// Run setup function before starting the test suite
	initTestResources()

	// Run the test suite for outbound ruleset
	m.Run()
}
