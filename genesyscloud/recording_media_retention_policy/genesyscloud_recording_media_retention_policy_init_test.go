package recording_media_retention_policy

import (
	"log"
	"sync"
	gcloud "terraform-provider-genesyscloud/genesyscloud"
	integration "terraform-provider-genesyscloud/genesyscloud/integration"

	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v105/platformclientv2"
)

const (
	trueValue  = "true"
	falseValue = "false"
	nullValue  = "null"
)

var sdkConfig *platformclientv2.Configuration
var err error

// providerDataSources holds a map of all registered datasources
var providerDataSources map[string]*schema.Resource

// providerResources holds a map of all registered resources
var providerResources map[string]*schema.Resource

type registerTestInstance struct {
	resourceMapMutex   sync.RWMutex
	datasourceMapMutex sync.RWMutex
}

// registerTestResources registers all resources used in the tests
func (r *registerTestInstance) registerTestResources() {
	r.resourceMapMutex.Lock()
	defer r.resourceMapMutex.Unlock()

	providerResources[resourceName] = ResourceMediaRetentionPolicy()
	providerResources["genesyscloud_routing_email_domain"] = gcloud.ResourceRoutingEmailDomain()
	providerResources["genesyscloud_routing_queue"] = gcloud.ResourceRoutingQueue()
	providerResources["genesyscloud_auth_role"] = gcloud.ResourceAuthRole()
	providerResources["genesyscloud_user_roles"] = gcloud.ResourceUserRoles()
	providerResources["genesyscloud_user"] = gcloud.ResourceUser()
	providerResources["genesyscloud_quality_forms_evaluation"] = gcloud.ResourceEvaluationForm()
	providerResources["genesyscloud_quality_forms_survey"] = gcloud.ResourceSurveyForm()
	providerResources["genesyscloud_integration"] = integration.ResourceIntegration()
	providerResources["genesyscloud_routing_language"] = gcloud.ResourceRoutingLanguage()
	providerResources["genesyscloud_routing_wrapupcode"] = gcloud.ResourceRoutingWrapupCode()
	providerResources["genesyscloud_flow"] = gcloud.ResourceFlow()
}

// registerTestDataSources registers all data sources used in the tests.
func (r *registerTestInstance) registerTestDataSources() {
	r.datasourceMapMutex.Lock()
	defer r.datasourceMapMutex.Unlock()

	providerDataSources[resourceName] = DataSourceRecordingMediaRetentionPolicy()
	providerDataSources["genesyscloud_routing_email_domain"] = gcloud.DataSourceRoutingEmailDomain()
	providerDataSources["genesyscloud_routing_queue"] = gcloud.DataSourceRoutingQueue()
	providerDataSources["genesyscloud_auth_role"] = gcloud.DataSourceAuthRole()
	providerDataSources["genesyscloud_user"] = gcloud.DataSourceUser()
	providerDataSources["genesyscloud_quality_forms_evaluation"] = gcloud.DataSourceQualityFormsEvaluations()
}

// initTestresources initializes all test resources and data sources.
func initTestresources() {
	if sdkConfig, err = gcloud.AuthorizeSdk(); err != nil {
		log.Fatal(err)
	}

	providerDataSources = make(map[string]*schema.Resource)
	providerResources = make(map[string]*schema.Resource)

	reg_instance := &registerTestInstance{}

	reg_instance.registerTestDataSources()
	reg_instance.registerTestResources()

}

// TestMain is a "setup" function called by the testing framework when run the test
func TestMain(m *testing.M) {
	// Run setup function before starting the test suite for integration package
	initTestresources()

	// Run the test suite for suite for the integration package
	m.Run()
}
