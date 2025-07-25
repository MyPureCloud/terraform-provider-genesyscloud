package recording_media_retention_policy

import (
	"log"
	"sync"
	"testing"

	flow "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/architect_flow"
	authDivision "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/auth_division"
	authRole "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/auth_role"
	integration "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/integration"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	qualityFormsEvaluation "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/quality_forms_evaluation"
	qualityFormsSurvey "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/quality_forms_survey"
	routingEmailDomain "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/routing_email_domain"
	routingLanguage "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/routing_language"
	routingQueue "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/routing_queue"
	routingWrapupcode "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/routing_wrapupcode"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/team"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/user"
	userRoles "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/user_roles"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v162/platformclientv2"
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

	providerResources[ResourceType] = ResourceMediaRetentionPolicy()
	providerResources[routingEmailDomain.ResourceType] = routingEmailDomain.ResourceRoutingEmailDomain()
	providerResources[routingQueue.ResourceType] = routingQueue.ResourceRoutingQueue()
	providerResources[authRole.ResourceType] = authRole.ResourceAuthRole()
	providerResources[userRoles.ResourceType] = userRoles.ResourceUserRoles()
	providerResources[user.ResourceType] = user.ResourceUser()
	providerResources[qualityFormsEvaluation.ResourceType] = qualityFormsEvaluation.ResourceEvaluationForm()
	providerResources[qualityFormsSurvey.ResourceType] = qualityFormsSurvey.ResourceQualityFormsSurvey()
	providerResources[integration.ResourceType] = integration.ResourceIntegration()
	providerResources[routingLanguage.ResourceType] = routingLanguage.ResourceRoutingLanguage()
	providerResources[routingWrapupcode.ResourceType] = routingWrapupcode.ResourceRoutingWrapupCode()
	providerResources[flow.ResourceType] = flow.ResourceArchitectFlow()
	providerResources[authDivision.ResourceType] = authDivision.ResourceAuthDivision()
	providerResources[team.ResourceType] = team.ResourceTeam()
}

// registerTestDataSources registers all data sources used in the tests.
func (r *registerTestInstance) registerTestDataSources() {
	r.datasourceMapMutex.Lock()
	defer r.datasourceMapMutex.Unlock()

	providerDataSources[ResourceType] = DataSourceRecordingMediaRetentionPolicy()
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
