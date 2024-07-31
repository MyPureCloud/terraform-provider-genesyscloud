package recording_media_retention_policy

import (
	"log"
	"sync"
	gcloud "terraform-provider-genesyscloud/genesyscloud"
	flow "terraform-provider-genesyscloud/genesyscloud/architect_flow"
	authRole "terraform-provider-genesyscloud/genesyscloud/auth_role"
	integration "terraform-provider-genesyscloud/genesyscloud/integration"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	routingQueue "terraform-provider-genesyscloud/genesyscloud/routing_queue"
	routingLanguage "terraform-provider-genesyscloud/genesyscloud/routing_language"
	userRoles "terraform-provider-genesyscloud/genesyscloud/user_roles"
	routingEmailDomain "terraform-provider-genesyscloud/genesyscloud/routing_email_domain"

	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"
)

/*
   The genesyscloud_recording_media_retention_policy_init_test.go file is used to initialize the data sources and resources
   used in testing the integration credential resource.

   Please make sure you register ALL resources and data sources your test cases will use.
*/

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
	providerResources["genesyscloud_routing_email_domain"] = routingEmailDomain.ResourceRoutingEmailDomain()
	providerResources["genesyscloud_routing_queue"] = routingQueue.ResourceRoutingQueue()
	providerResources["genesyscloud_auth_role"] = authRole.ResourceAuthRole()
	providerResources["genesyscloud_user_roles"] = userRoles.ResourceUserRoles()
	providerResources["genesyscloud_user"] = gcloud.ResourceUser()
	providerResources["genesyscloud_quality_forms_evaluation"] = gcloud.ResourceEvaluationForm()
	providerResources["genesyscloud_quality_forms_survey"] = gcloud.ResourceSurveyForm()
	providerResources["genesyscloud_integration"] = integration.ResourceIntegration()
	providerResources["genesyscloud_routing_language"] = routingLanguage.ResourceRoutingLanguage()
	providerResources["genesyscloud_routing_wrapupcode"] = gcloud.ResourceRoutingWrapupCode()
	providerResources["genesyscloud_flow"] = flow.ResourceArchitectFlow()
}

// registerTestDataSources registers all data sources used in the tests.
func (r *registerTestInstance) registerTestDataSources() {
	r.datasourceMapMutex.Lock()
	defer r.datasourceMapMutex.Unlock()

	providerDataSources[resourceName] = DataSourceRecordingMediaRetentionPolicy()
}

// initTestResources initializes all test resources and data sources.
func initTestResources() {
	if sdkConfig, err = provider.AuthorizeSdk(); err != nil {
		log.Fatal(err)
	}

	providerDataSources = make(map[string]*schema.Resource)
	providerResources = make(map[string]*schema.Resource)

	regInstance := &registerTestInstance{}

	regInstance.registerTestDataSources()
	regInstance.registerTestResources()
}

// TestMain is a "setup" function called by the testing framework when run the test
func TestMain(m *testing.M) {
	// Run setup function before starting the test suite for integration package
	initTestResources()

	// Run the test suite for the integration package
	m.Run()
}
