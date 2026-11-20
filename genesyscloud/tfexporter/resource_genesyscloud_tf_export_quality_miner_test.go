package tfexporter

import (
	"fmt"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	qualityMiner "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/quality_miner"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/testrunner"
)

func TestUnitQualityMinerExporterRegistered(t *testing.T) {
	if resourceExporters[qualityMiner.ResourceType] == nil {
		t.Fatalf("no exporter registered for %s", qualityMiner.ResourceType)
	}
	if providerResources[qualityMiner.ResourceType] == nil {
		t.Fatalf("no resource registered for %s", qualityMiner.ResourceType)
	}
}

func TestAccResourceTfExportQualityMiner(t *testing.T) {
	testSetup(t)

	var (
		exportDir     = testrunner.GetTestTempPath(".terraform" + uuid.NewString())
		minerName     = "tf-export-miner-" + strings.ReplaceAll(uuid.NewString(), "-", "")[:8]
		minerLabel    = "export-miner"
		minerResource = qualityMiner.GenerateQualityMinerResource(minerLabel, minerName, "Topic", `language = "en-US"`)
	)
	defer removeExportDir(exportDir)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: minerResource,
			},
			{
				Config: minerResource + generateSttExport(exportDir, []string{qualityMiner.ResourceType},
					qualityMiner.ResourceType+"."+minerLabel),
				Check: assertSttExportedBlock(exportDir, qualityMiner.ResourceType, "name", minerName, func(attrs map[string]interface{}) error {
					if attrs["miner_type"] != "Topic" {
						return fmt.Errorf("miner_type = %v, expected Topic", attrs["miner_type"])
					}
					if language, _ := attrs["language"].(string); !strings.EqualFold(language, "en-US") {
						return fmt.Errorf("language = %v, expected en-US", attrs["language"])
					}
					return nil
				}),
			},
		},
		CheckDestroy: testVerifyExportsDestroyedFunc(exportDir),
	})
}
