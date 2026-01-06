package routing_wrapupcode

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/mypurecloud/platform-client-sdk-go/v165/platformclientv2"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	registrar "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_register"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
)

const ResourceType = "genesyscloud_routing_wrapupcode"

// SetRegistrar registers all of the resources, datasources and exporters in the package
func SetRegistrar(regInstance registrar.Registrar) {
	regInstance.RegisterFrameworkResource(ResourceType, NewRoutingWrapupcodeFrameworkResource)
	regInstance.RegisterFrameworkDataSource(ResourceType, NewRoutingWrapupcodeFrameworkDataSource)
	regInstance.RegisterExporter(ResourceType, RoutingWrapupcodeExporter())
}

func RoutingWrapupcodeExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: provider.GetAllWithPooledClient(GetAllRoutingWrapupcodes),
		RefAttrs: map[string]*resourceExporter.RefAttrSettings{
			"division_id": {RefType: "genesyscloud_auth_division"},
		},
	}
}

// GetAllRoutingWrapupcodes retrieves all routing wrapupcodes for export using the proxy
func GetAllRoutingWrapupcodes(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	proxy := getRoutingWrapupcodeProxy(clientConfig)
	wrapupcodes, _, err := proxy.getAllRoutingWrapupcode(ctx)
	if err != nil {
		return nil, diag.Errorf("Failed to get routing wrapupcodes for export: %v", err)
	}

	if wrapupcodes == nil {
		return resourceExporter.ResourceIDMetaMap{}, nil
	}

	exportMap := make(resourceExporter.ResourceIDMetaMap)
	for _, wrapupcode := range *wrapupcodes {
		exportMap[*wrapupcode.Id] = &resourceExporter.ResourceMeta{
			BlockLabel: *wrapupcode.Name,
		}
	}
	return exportMap, nil
}

// GenerateRoutingWrapupcodeResource generates a routing wrapupcode resource for cross-package testing
// This function is used by other packages that need to create routing wrapupcode resources in their tests
func GenerateRoutingWrapupcodeResource(resourceLabel string, name string, divisionId string, description string) string {
	divisionIdAttr := ""
	if divisionId != util.NullValue {
		divisionIdAttr = fmt.Sprintf(`
		division_id = %s`, divisionId)
	}

	descriptionAttr := ""
	if description != "" {
		descriptionAttr = fmt.Sprintf(`
		description = "%s"`, description)
	}

	return fmt.Sprintf(`resource "genesyscloud_routing_wrapupcode" "%s" {
		name = "%s"%s%s
	}
	`, resourceLabel, name, divisionIdAttr, descriptionAttr)
}
