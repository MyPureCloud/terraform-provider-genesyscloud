package tfexporter

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v200/platformclientv2"
	ivrIdentityResolution "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/architect_ivr_identity_resolution"
	appleIdentityResolution "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/conversations_messaging_integrations_apple_identity_resolution"
	instagramIdentityResolution "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/conversations_messaging_integrations_instagram_identity_resolution"
	openIdentityResolution "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/conversations_messaging_integrations_open_identity_resolution"
	whatsappIdentityResolution "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/conversations_messaging_integrations_whatsapp_identity_resolution"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	routingEmailDomain "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/routing_email_domain"
	routingEmailRoute "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/routing_email_route"
	emailRouteIdentityResolution "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/routing_email_route_identity_resolution"
	queueIdentityResolution "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/routing_queue_identity_resolution"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/testrunner"
	webDeploymentIdentityResolution "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/webdeployments_deployment_identity_resolution"
)

func TestUnitIdentityResolutionExportersAreRegistered(t *testing.T) {
	for _, resourceType := range []string{
		ivrIdentityResolution.ResourceType,
		appleIdentityResolution.ResourceType,
		instagramIdentityResolution.ResourceType,
		openIdentityResolution.ResourceType,
		whatsappIdentityResolution.ResourceType,
		emailRouteIdentityResolution.ResourceType,
		webDeploymentIdentityResolution.ResourceType,
		queueIdentityResolution.ResourceType,
	} {
		if resourceExporters[resourceType] == nil {
			t.Errorf("no exporter registered for %s", resourceType)
		}
		if providerResources[resourceType] == nil {
			t.Errorf("no resource registered for %s", resourceType)
		}
	}
}

// resolve_identities=true with the unassigned division is the platform default, which the exporter skips
func TestAccResourceTfExportEmailRouteIdentityResolution(t *testing.T) {
	testSetup(t)

	var (
		exportDir    = testrunner.GetTestTempPath(".terraform" + uuid.NewString())
		domainId     = fmt.Sprintf("terraformexport.%s.com", strings.ReplaceAll(uuid.NewString(), "-", ""))
		routePattern = "tf" + strings.ReplaceAll(uuid.NewString(), "-", "")[:8]
		domainLabel  = "export-ir-domain"
		routeLabel   = "export-ir-route"
		irLabel      = "export-ir"
		domainRef    = routingEmailDomain.ResourceType + "." + domainLabel
		routeRef     = routingEmailRoute.ResourceType + "." + routeLabel
		irRef        = emailRouteIdentityResolution.ResourceType + "." + irLabel
		parentConfig = routingEmailDomain.GenerateRoutingEmailDomainResource(domainLabel, domainId, util.FalseValue, util.NullValue) +
			routingEmailRoute.GenerateRoutingEmailRouteResource(routeLabel, domainRef+".id", routePattern, "TF Export IR", fmt.Sprintf(`from_email = "tf@%s"`, domainId))
		irConfig = fmt.Sprintf(`
resource "%s" "%s" {
  domain_name        = %s.id
  route_id           = %s.id
  resolve_identities = false
}
`, emailRouteIdentityResolution.ResourceType, irLabel, domainRef, routeRef)
	)
	defer removeExportDir(exportDir)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: parentConfig + irConfig,
			},
			{
				PreConfig: func() { waitForEmailRouteIdentityResolution(t, domainId, routePattern) },
				Config: parentConfig + irConfig + generateSttExport(exportDir,
					[]string{emailRouteIdentityResolution.ResourceType, routingEmailRoute.ResourceType, routingEmailDomain.ResourceType},
					irRef, routeRef, domainRef),
				Check: assertExportedEmailRouteIdentityResolution(exportDir),
			},
		},
		CheckDestroy: testVerifyExportsDestroyedFunc(exportDir),
	})
}

func assertExportedEmailRouteIdentityResolution(exportDir string) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		blocks, err := getResourceDefinition(filepath.Join(exportDir, defaultTfJSONFile), emailRouteIdentityResolution.ResourceType)
		if err != nil {
			return fmt.Errorf("%w (exported resource types: %s)", err, exportedResourceTypes(filepath.Join(exportDir, defaultTfJSONFile)))
		}
		if len(blocks) != 1 {
			return fmt.Errorf("expected 1 exported %s block, got %d", emailRouteIdentityResolution.ResourceType, len(blocks))
		}
		for _, raw := range blocks {
			var attrs map[string]interface{}
			if err := json.Unmarshal(*raw, &attrs); err != nil {
				return fmt.Errorf("failed to unmarshal %s block: %w", emailRouteIdentityResolution.ResourceType, err)
			}
			if attrs["resolve_identities"] != false {
				return fmt.Errorf("resolve_identities = %v, expected false", attrs["resolve_identities"])
			}
			routeId, _ := attrs["route_id"].(string)
			if !strings.HasPrefix(routeId, "${"+routingEmailRoute.ResourceType+".") {
				return fmt.Errorf("route_id = %q, expected a reference to a %s resource", routeId, routingEmailRoute.ResourceType)
			}
			domainName, _ := attrs["domain_name"].(string)
			if !strings.HasPrefix(domainName, "${"+routingEmailDomain.ResourceType+".") {
				return fmt.Errorf("domain_name = %q, expected a reference to a %s resource", domainName, routingEmailDomain.ResourceType)
			}
		}
		return nil
	}
}

// the exporter lists routes through the list API, which can lag behind a just-created route
func waitForEmailRouteIdentityResolution(t *testing.T, domainId, routePattern string) {
	t.Helper()
	config, err := provider.AuthorizeSdk()
	if err != nil {
		t.Fatalf("failed to authorize SDK: %v", err)
	}
	routingApi := platformclientv2.NewRoutingApiWithConfig(config)
	lastState := "route not listed yet"
	for deadline := time.Now().Add(2 * time.Minute); time.Now().Before(deadline); time.Sleep(3 * time.Second) {
		routes, _, err := routingApi.GetRoutingEmailDomainRoutes(domainId, 100, 1, "", nil)
		if err != nil || routes == nil || routes.Entities == nil {
			lastState = fmt.Sprintf("listing routes failed: %v", err)
			continue
		}
		for _, route := range *routes.Entities {
			if route.Id == nil || route.Pattern == nil || *route.Pattern != routePattern {
				continue
			}
			ir, _, irErr := routingApi.GetRoutingEmailDomainRouteIdentityresolution(domainId, *route.Id)
			if irErr == nil && ir != nil && ir.ResolveIdentities != nil && !*ir.ResolveIdentities {
				return
			}
			lastState = fmt.Sprintf("route %s listed, identity resolution not applied yet (err: %v)", *route.Id, irErr)
		}
	}
	t.Fatalf("route %s/%s did not show the expected identity resolution within 2 minutes: %s", domainId, routePattern, lastState)
}

func exportedResourceTypes(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return "export file unreadable: " + err.Error()
	}
	var top map[string]json.RawMessage
	if err := json.Unmarshal(data, &top); err != nil || top["resource"] == nil {
		return "no resource section"
	}
	var types map[string]json.RawMessage
	if err := json.Unmarshal(top["resource"], &types); err != nil {
		return "unreadable resource section"
	}
	names := make([]string, 0, len(types))
	for name := range types {
		names = append(names, name)
	}
	return strings.Join(names, ", ")
}
