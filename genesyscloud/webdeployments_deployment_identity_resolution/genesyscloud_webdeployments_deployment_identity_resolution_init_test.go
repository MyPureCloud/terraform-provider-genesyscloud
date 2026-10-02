package webdeployments_deployment_identity_resolution

import (
	"sync"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v195/platformclientv2"
	gcloud "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	webDeployConfig "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/webdeployments_configuration"
	webDeployDeploy "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/webdeployments_deployment"
)

var (
	sdkConfig           *platformclientv2.Configuration
	providerDataSources map[string]*schema.Resource
	providerResources   map[string]*schema.Resource
)

type registerTestInstance struct {
	resourceMapMutex   sync.RWMutex
	datasourceMapMutex sync.RWMutex
}

func (r *registerTestInstance) registerTestResources() {
	r.resourceMapMutex.Lock()
	defer r.resourceMapMutex.Unlock()

	providerResources[ResourceType] = ResourceWebDeploymentIdentityResolution()
	providerResources[webDeployDeploy.ResourceType] = webDeployDeploy.ResourceWebDeployment()
	providerResources[webDeployConfig.ResourceType] = webDeployConfig.ResourceWebDeploymentConfiguration()
}

func (r *registerTestInstance) registerTestDataSources() {
	r.datasourceMapMutex.Lock()
	defer r.datasourceMapMutex.Unlock()

	providerDataSources["genesyscloud_auth_division_home"] = gcloud.DataSourceAuthDivisionHome()
}

func initTestResources() {
	sdkConfig = provider.SdkConfigurationForTests()

	providerDataSources = make(map[string]*schema.Resource)
	providerResources = make(map[string]*schema.Resource)

	regInstance := &registerTestInstance{}
	regInstance.registerTestResources()
	regInstance.registerTestDataSources()
}

func TestMain(m *testing.M) {
	initTestResources()
	m.Run()
}
