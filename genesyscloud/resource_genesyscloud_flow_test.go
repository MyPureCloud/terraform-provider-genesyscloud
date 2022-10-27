package genesyscloud

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v80/platformclientv2"
)

// lockFlow will search for a specific flow and then lock it.  This is to specifically test the force_unlock flag where I want to create a flow,  simulate some one locking it and then attempt to
// do another CX as Code deploy.
func lockFlow(flowName string, flowType string) {
	archAPI := platformclientv2.NewArchitectApi()
	ctx := context.Background()
	withRetries(ctx, 5*time.Second, func() *resource.RetryError {
		const pageSize = 100
		for pageNum := 1; ; pageNum++ {
			flows, _, getErr := archAPI.GetFlows(nil, pageNum, pageSize, "", "", nil, flowName, "", "", "", "", "", "", "", false, false, "", "", nil)
			if getErr != nil {
				return resource.NonRetryableError(fmt.Errorf("Error requesting flow %s: %s", flowName, getErr))
			}

			if flows.Entities == nil || len(*flows.Entities) == 0 {
				return resource.RetryableError(fmt.Errorf("No flows found with name %s", flowName))
			}

			for _, entity := range *flows.Entities {
				if *entity.Name == flowName && *entity.VarType == flowType {
					flow, response, err := archAPI.PostFlowsActionsCheckout(*entity.Id)

					if err != nil || response.Error != nil {
						return resource.NonRetryableError(fmt.Errorf("Error requesting flow %s: %s", flowName, getErr))
					}

					log.Printf("Flow (%s) with FlowName: %s has been locked Flow resource after checkout: %v\n", *flow.Id, flowName, *flow.LockedClient.Name)

					return nil
				}
			}
		}
	})
}

// Tests the force_unlock functionality.
func TestAccResourceArchFlowForceUnlock(t *testing.T) {
	//myDir, _ := os.Getwd()
	var (
		flowResource = "test_force_unlock_flow1"
		flowName     = "Terraform Flow Test ForceUnlock-" + uuid.NewString()
		flowType     = "INBOUNDCALL"
		filePath     = "../examples/resources/genesyscloud_flow/inboundcall_flow_example.yaml"

		inboundcallConfig1 = fmt.Sprintf("inboundCall:\n  name: %s\n  defaultLanguage: en-us\n  startUpRef: ./menus/menu[mainMenu]\n  initialGreeting:\n    tts: Archy says hi!!!\n  menus:\n    - menu:\n        name: Main Menu\n        audio:\n          tts: You are at the Main Menu, press 9 to disconnect.\n        refId: mainMenu\n        choices:\n          - menuDisconnect:\n              name: Disconnect\n              dtmf: digit_9", flowName)
		inboundcallConfig2 = fmt.Sprintf("inboundCall:\n  name: %s\n  defaultLanguage: en-us\n  startUpRef: ./menus/menu[mainMenu]\n  initialGreeting:\n    tts: Archy says hi again!!!\n  menus:\n    - menu:\n        name: Main Menu\n        audio:\n          tts: You are at the Main Menu, press 9 to disconnect.\n        refId: mainMenu\n        choices:\n          - menuDisconnect:\n              name: Disconnect\n              dtmf: digit_9", flowName)
	)

	//Create an anonymous function that closes around the flow name and flow Type
	var flowLocFunc = func() {
		lockFlow(flowName, flowType)
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				// Create flow
				Config: generateFlowResource(
					flowResource,
					filePath,
					inboundcallConfig1,
					false,
				),
				Check: resource.ComposeTestCheckFunc(
					validateFlow("genesyscloud_flow."+flowResource, flowName, flowType),
				),
			},
			{
				//Lock the flow, do a deploy and check to make sure the flow is locked b
				PreConfig: flowLocFunc, //This will lock the flow.
				Config: generateFlowResource(
					flowResource,
					filePath,
					inboundcallConfig2,
					true,
				),
				Check: resource.ComposeTestCheckFunc(
					validateFlowUnlocked("genesyscloud_flow."+flowResource),
					validateFlow("genesyscloud_flow."+flowResource, flowName, flowType),
				),
			},
			{
				// Import/Read
				ResourceName:            "genesyscloud_flow." + flowResource,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"filepath", "force_unlock", "file_content_hash"},
			},
		},
		CheckDestroy: testVerifyFlowDestroyed,
	})
}

func TestAccResourceArchFlowStandard(t *testing.T) {
	var (
		flowResource1 = "test_flow1"
		flowResource2 = "test_flow2"
		flowName1     = "Terraform Flow Test-" + uuid.NewString()
		flowName2     = "Terraform Flow Test-" + uuid.NewString()
		flowType1     = "INBOUNDCALL"
		flowType2     = "INBOUNDEMAIL"
		filePath1     = "../examples/resources/genesyscloud_flow/inboundcall_flow_example.yaml" //Have to use an explicit path because the filesha function gets screwy on relative class names
		filePath2     = "../examples/resources/genesyscloud_flow/inboundcall_flow_example2.yaml"
		filePath3     = "../examples/resources/genesyscloud_flow/inboundcall_flow_example3.yaml"

		inboundcallConfig1 = fmt.Sprintf("inboundCall:\n  name: %s\n  defaultLanguage: en-us\n  startUpRef: ./menus/menu[mainMenu]\n  initialGreeting:\n    tts: Archy says hi!!!\n  menus:\n    - menu:\n        name: Main Menu\n        audio:\n          tts: You are at the Main Menu, press 9 to disconnect.\n        refId: mainMenu\n        choices:\n          - menuDisconnect:\n              name: Disconnect\n              dtmf: digit_9", flowName1)
		inboundcallConfig2 = fmt.Sprintf("inboundCall:\n  name: %s\n  defaultLanguage: en-us\n  startUpRef: ./menus/menu[mainMenu]\n  initialGreeting:\n    tts: Archy says hi!!!!!\n  menus:\n    - menu:\n        name: Main Menu\n        audio:\n          tts: You are at the Main Menu, press 9 to disconnect.\n        refId: mainMenu\n        choices:\n          - menuDisconnect:\n              name: Disconnect\n              dtmf: digit_9", flowName2)
	)

	var homeDivisionName string
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: "data \"genesyscloud_auth_division_home\" \"home\" {}",
				Check: resource.ComposeTestCheckFunc(
					getHomeDivisionName("data.genesyscloud_auth_division_home.home", &homeDivisionName),
				),
			},
		},
	})

	inboundemailConfig1 := fmt.Sprintf(`inboundEmail:
    name: %s
    division: %s
    startUpRef: "/inboundEmail/states/state[Initial State_10]"
    defaultLanguage: en-us
    supportedLanguages:
        en-us:
            defaultLanguageSkill:
                noValue: true
    settingsInboundEmailHandling:
        emailHandling:
            disconnect:
                none: true
    settingsErrorHandling:
        errorHandling:
            disconnect:
                none: true
    states:
        - state:
            name: Initial State
            refId: Initial State_10
            actions:
                - disconnect:
                    name: Disconnect
`, flowName1, homeDivisionName)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				// Create flow
				Config: generateFlowResource(
					flowResource1,
					filePath1,
					inboundcallConfig1,
					false,
				),
				Check: resource.ComposeTestCheckFunc(
					validateFlow("genesyscloud_flow."+flowResource1, flowName1, flowType1),
				),
			},
			{
				// Update flow with name
				Config: generateFlowResource(
					flowResource1,
					filePath2,
					inboundcallConfig2,
					false,
				),
				Check: resource.ComposeTestCheckFunc(
					validateFlow("genesyscloud_flow."+flowResource1, flowName2, flowType1),
				),
			},
			{
				// Import/Read
				ResourceName:            "genesyscloud_flow." + flowResource1,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"filepath", "force_unlock", "file_content_hash"},
			},
			{
				// Create inboundemail flow
				Config: generateFlowResource(
					flowResource2,
					filePath3,
					inboundemailConfig1,
					false,
				),
				Check: resource.ComposeTestCheckFunc(
					validateFlow("genesyscloud_flow."+flowResource2, flowName1, flowType2),
				),
			},
			{
				// Update inboundemail flow to inboundcall
				Config: generateFlowResource(
					flowResource2,
					filePath2,
					inboundcallConfig2,
					false,
				),
				Check: resource.ComposeTestCheckFunc(
					validateFlow("genesyscloud_flow."+flowResource2, flowName2, flowType1),
				),
			},
			{
				// Import/Read
				ResourceName:            "genesyscloud_flow." + flowResource2,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"filepath", "force_unlock", "file_content_hash"},
			},
		},
		CheckDestroy: testVerifyFlowDestroyed,
	})
}

func TestAccResourceArchFlowSubstitutions(t *testing.T) {
	//myDir, _ := os.Getwd()
	var (
		flowResource1 = "test_flow1"
		flowName1     = "Terraform Flow Test-" + uuid.NewString()
		flowName2     = "Terraform Flow Test-" + uuid.NewString()
		filePath1     = "../examples/resources/genesyscloud_flow/inboundcall_flow_example_substitutions.yaml"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				// Create flow
				Config: generateFlowResource(
					flowResource1,
					filePath1,
					"",
					false,
					generateFlowSubstitutions(map[string]string{
						"flow_name":            flowName1,
						"default_language":     "en-us",
						"greeting":             "Archy says hi!!!",
						"menu_disconnect_name": "Disconnect",
					}),
				),
				Check: resource.ComposeTestCheckFunc(
					validateFlow("genesyscloud_flow."+flowResource1, flowName1, "INBOUNDCALL"),
				),
			},
			{
				// Update
				Config: generateFlowResource(
					flowResource1,
					filePath1,
					"",
					false,
					generateFlowSubstitutions(map[string]string{
						"flow_name":            flowName2,
						"default_language":     "en-us",
						"greeting":             "Archy says hi!!!",
						"menu_disconnect_name": "Disconnect",
					}),
				),
				Check: resource.ComposeTestCheckFunc(
					validateFlow("genesyscloud_flow."+flowResource1, flowName2, "INBOUNDCALL"),
				),
			},
		},
		CheckDestroy: testVerifyFlowDestroyed,
	})
}

func copyFile(src string, dest string) {
	bytesRead, err := ioutil.ReadFile(src)

	if err != nil {
		log.Fatal(err)
	}

	err = ioutil.WriteFile(dest, bytesRead, 0644)

	if err != nil {
		log.Fatal(err)
	}
}

func removeFile(fileName string) {
	err := os.Remove(fileName)
	if err != nil {
		log.Fatal(err)
	}
}

func transformFile(fileName string) {
	input, err := ioutil.ReadFile(fileName)
	if err != nil {
		log.Fatalln(err)
	}

	lines := strings.Split(string(input), "\n")

	for i, line := range lines {
		lines[i] = strings.Replace(line, "You are at the Main Menu, press 9 to disconnect.", "Hi you are at the Main Menu, press 9 to disconnect.", 1)
	}

	output := strings.Join(lines, "\n")
	err = ioutil.WriteFile(fileName, []byte(output), 0644)
	if err != nil {
		log.Fatalln(err)
	}
}

/*
This test case was put out here to test for the problem described in: DEVENGAGE-1472.  Basically the bug manifested
itself when you deploy a flow and then modify the yaml file so that the hash changes.  This bug had two manifestations.
One the new values in a substitution would not be picked up.  Two, if the flow file changed, a flow would be deployed
even if the user was only doing a plan or destroy.

This test exercises this bug by first deploying a flow file with a substitution.  Then modifying the flow file and rerunning
the flow with a substitution.
*/
func TestAccResourceArchFlowSubstitutionsWithMultipleTouch(t *testing.T) {
	var (
		flowResource1 = "test_flow1"
		flowName1     = "Terraform Flow Test-" + uuid.NewString()
		flowName2     = "Terraform Flow Test-" + uuid.NewString()
		srcFile       = "../examples/resources/genesyscloud_flow/inboundcall_flow_example_substitutions.yaml"
		destFile      = "../examples/resources/genesyscloud_flow/inboundcall_flow_example_holder.yaml"
	)

	//Copy the example substitution file over to a temp file that can be manipulated and modified
	copyFile(srcFile, destFile)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				// Create flow
				Config: generateFlowResource(
					flowResource1,
					destFile,
					"",
					false,
					generateFlowSubstitutions(map[string]string{
						"flow_name":            flowName1,
						"default_language":     "en-us",
						"greeting":             "Archy says hi!!!",
						"menu_disconnect_name": "Disconnect",
					}),
				),
				Check: resource.ComposeTestCheckFunc(
					validateFlow("genesyscloud_flow."+flowResource1, flowName1, "INBOUNDCALL"),
				),
			},
			{ // Update the flow, but make sure that we touch the YAML file and change something int
				PreConfig: func() { transformFile(destFile) },
				Config: generateFlowResource(
					flowResource1,
					destFile,
					"",
					false,
					generateFlowSubstitutions(map[string]string{
						"flow_name":            flowName2,
						"default_language":     "en-us",
						"greeting":             "Archy says hi!!!",
						"menu_disconnect_name": "Disconnect",
					}),
				),
				Check: resource.ComposeTestCheckFunc(
					validateFlow("genesyscloud_flow."+flowResource1, flowName2, "INBOUNDCALL"),
				),
			},
		},
		CheckDestroy: testVerifyFlowDestroyed,
	})

	//Clean up the temporary file
	removeFile(destFile)
}

func generateFlowSubstitutions(substitutions map[string]string) string {
	var substitutionsStr string
	for k, v := range substitutions {
		substitutionsStr += fmt.Sprintf("\t%s = \"%s\"\n", k, v)
	}
	return fmt.Sprintf(`substitutions = {
%s}`, substitutionsStr)
}

func generateFlowResource(resourceID, srcFile, filecontent string, force_unlock bool, substitutions ...string) string {
	fullyQualifiedPath, _ := filepath.Abs(srcFile)

	if filecontent != "" {
		updateFile(srcFile, filecontent)
	}

	flowResourceStr := fmt.Sprintf(`resource "genesyscloud_flow" "%s" {
        filepath = %s
		file_content_hash =  filesha256(%s)
		force_unlock = %v
		%s
	}
	`, resourceID, strconv.Quote(srcFile), strconv.Quote(fullyQualifiedPath), force_unlock, strings.Join(substitutions, "\n"))

	return flowResourceStr
}

func updateFile(filepath, content string) {
	file, err := os.OpenFile(filepath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)

	if err != nil {
		log.Println(err)
		return
	}
	defer file.Close()

	file.WriteString(content)
}

// Check if flow is published, then check if flow name and type are correct
func validateFlow(flowResourceName, name, flowType string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		flowResource, ok := state.RootModule().Resources[flowResourceName]
		if !ok {
			return fmt.Errorf("Failed to find flow %s in state", flowResourceName)
		}
		flowID := flowResource.Primary.ID
		architectAPI := platformclientv2.NewArchitectApi()

		flow, _, err := architectAPI.GetFlow(flowID, false)

		if err != nil {
			return fmt.Errorf("Unexpected error: %s", err)
		}

		if flow == nil {
			return fmt.Errorf("Flow (%s) not found. ", flowID)
		}

		if *flow.Name != name {
			return fmt.Errorf("Returned flow (%s) has incorrect name. Expect: %s, Actual: %s", flowID, name, *flow.Name)
		}

		if *flow.VarType != flowType {
			return fmt.Errorf("Returned flow (%s) has incorrect type. Expect: %s, Actual: %s", flowID, flowType, *flow.VarType)
		}

		return nil
	}
}

// Will attempt to determine if a flow is unlocked. I check to see if a flow is locked, by attempting to check the flow again.  If the flow is locked the second checkout
// will fail with a 409 status code.  If the flow is unlocked, the status code will be a 200
func validateFlowUnlocked(flowResourceName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		flowResource, ok := state.RootModule().Resources[flowResourceName]
		if !ok {
			return fmt.Errorf("Failed to find flow %s in state", flowResourceName)
		}

		flowID := flowResource.Primary.ID
		architectAPI := platformclientv2.NewArchitectApi()

		flow, response, err := architectAPI.PostFlowsActionsCheckout(flowID)

		if err != nil && response == nil {
			return fmt.Errorf("Unexpected error: %s", err)
		}

		if err != nil && response.StatusCode == http.StatusConflict {
			return fmt.Errorf("Flow (%s) is supposed to be in an unlocked state and it is in a locked state. Tried to lock the flow to see if I could lock it and it failed.", flowID)
		}

		if flow == nil {
			return fmt.Errorf("Flow (%s) not found. ", flowID)
		}

		return nil
	}
}

func testVerifyFlowDestroyed(state *terraform.State) error {
	architectAPI := platformclientv2.NewArchitectApi()
	for _, rs := range state.RootModule().Resources {
		if rs.Type != "genesyscloud_flow" {
			continue
		}

		flow, resp, err := architectAPI.GetFlow(rs.Primary.ID, false)
		if flow != nil {
			return fmt.Errorf("Flow (%s) still exists", rs.Primary.ID)
		} else if resp != nil && resp.StatusCode == 410 {
			// Flow not found as expected
			log.Printf("Flow (%s) successfully deleted", rs.Primary.ID)
			continue
		} else {
			// Unexpected error
			return fmt.Errorf("Unexpected error: %s", err)
		}
	}
	// Success. All Flows destroyed
	return nil
}
