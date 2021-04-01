package genesyscloud

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"gonum.org/v1/gonum/graph/simple"
	"gonum.org/v1/gonum/graph/topo"
)

const exportTestDir = "../.terraform"

func TestAccResourceTfExport(t *testing.T) {
	var (
		exportResource1 = "test-export1"
		configPath      = filepath.Join(exportTestDir, defaultTfJSONFile)
		statePath       = filepath.Join(exportTestDir, defaultTfStateFile)
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				// Run export without state file
				Config: generateTfExportResource(
					exportResource1,
					exportTestDir,
					falseValue,
					"",
				),
				Check: resource.ComposeTestCheckFunc(
					validateFileCreated(configPath),
					validateConfigFile(configPath),
				),
			},
			{
				// Run export with state file and excluded attribute
				Config: generateTfExportResource(
					exportResource1,
					exportTestDir,
					trueValue,
					strconv.Quote("genesyscloud_auth_role.permission_policies.conditions"),
				),
				Check: resource.ComposeTestCheckFunc(
					validateFileCreated(configPath),
					validateConfigFile(configPath),
					validateFileCreated(statePath),
				),
			},
		},
		CheckDestroy: testVerifyExportsDestroyed,
	})
}

// Create a directed graph of exported resources to their references. Report any potential graph cycles in this test.
// Reference cycles can sometimes be broken by exporting a separate resource to update membership after the member
// and container resources are created/updated (see genesyscloud_user_roles).
func TestForExportCycles(t *testing.T) {

	// Assumes exporting all resource types
	exporters := getResourceExporters(nil)

	graph := simple.NewDirectedGraph()

	var resNames []string
	for resName := range exporters {
		graph.AddNode(simple.Node(len(resNames)))
		resNames = append(resNames, resName)
	}

	for i, resName := range resNames {
		currNode := simple.Node(i)
		for attrName, refSettings := range exporters[resName].RefAttrs {
			if refSettings.RefType == resName {
				// Resources that can reference themselves are ignored
				// Cycles caused by self-refs are likely a misconfiguration
				// (e.g. two users that are each other's manager)
				continue
			}
			if exporters[resName].isAttributeExcluded(attrName) {
				// This reference attribute will be excluded from the export
				continue
			}
			graph.SetEdge(simple.Edge{F: currNode, T: simple.Node(resNodeIndex(refSettings.RefType, resNames))})
		}
	}

	cycles := topo.DirectedCyclesIn(graph)
	if len(cycles) > 0 {
		cycleResources := make([][]string, len(cycles))
		for i, cycle := range cycles {
			cycleResources[i] = make([]string, len(cycle))
			for j, cycleNode := range cycle {
				cycleResources[i][j] = resNames[cycleNode.ID()]
			}
		}
		t.Fatalf("Found the following potential reference cycles:\n %s", cycleResources)
	}
}

func resNodeIndex(resName string, resNames []string) int64 {
	for i, name := range resNames {
		if resName == name {
			return int64(i)
		}
	}
	return -1
}

func generateTfExportResource(
	resourceID string,
	directory string,
	includeState string,
	excludedAttributes string) string {
	return fmt.Sprintf(`resource "genesyscloud_tf_export" "%s" {
		directory = "%s"
        include_state_file = %s
		resource_types = [
			"genesyscloud_auth_role",
			"genesyscloud_auth_division",
			"genesyscloud_group",
			"genesyscloud_group_roles",
			"genesyscloud_idp_adfs",
			"genesyscloud_idp_generic",
			"genesyscloud_idp_gsuite",
			"genesyscloud_idp_okta",
			"genesyscloud_idp_onelogin",
			"genesyscloud_idp_ping",
			"genesyscloud_idp_salesforce",
			"genesyscloud_location",
			"genesyscloud_routing_language",
			"genesyscloud_routing_queue",
			"genesyscloud_routing_skill"
		]
		exclude_attributes = [%s]
    }
	`, resourceID, directory, includeState, excludedAttributes)
}

func validateFileCreated(filename string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		_, err := os.Stat(filename)
		if err != nil {
			return fmt.Errorf("Failed to find file %s", filename)
		}
		return nil
	}
}

func testVerifyExportsDestroyed(state *terraform.State) error {
	// Check config file deleted
	jsonConfigPath := filepath.Join(exportTestDir, defaultTfJSONFile)
	_, err := os.Stat(jsonConfigPath)
	if !os.IsNotExist(err) {
		return fmt.Errorf("Failed to delete JSON config file %s", jsonConfigPath)
	}

	// Check state file deleted
	statePath := filepath.Join(exportTestDir, defaultTfStateFile)
	_, err = os.Stat(statePath)
	if !os.IsNotExist(err) {
		return fmt.Errorf("Failed to delete state file %s", statePath)
	}
	return nil
}

func validateConfigFile(path string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		jsonFile, err := os.Open(path)
		if err != nil {
			return err
		}
		defer jsonFile.Close()

		byteValue, _ := ioutil.ReadAll(jsonFile)

		var result map[string]interface{}
		err = json.Unmarshal([]byte(byteValue), &result)
		if err != nil {
			return err
		}

		if _, ok := result["resource"]; !ok {
			return fmt.Errorf("Config file missing resource attribute.")
		}

		if _, ok := result["terraform"]; !ok {
			return fmt.Errorf("Config file missing terraform attribute.")
		}
		return nil
	}
}
