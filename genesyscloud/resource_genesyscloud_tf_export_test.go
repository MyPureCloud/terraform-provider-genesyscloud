package genesyscloud

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"gonum.org/v1/gonum/graph/simple"
	"gonum.org/v1/gonum/graph/topo"
)

const exportTestDir = "../.terraform"

type UserExport struct {
	Email string `json:"email"`
	Name  string `json:"name"`
	State string `json:"state"`
}

type QueueExport struct {
	AcwTimeoutMs int    `json:"acw_timeout_ms"`
	Description  string `json:"description"`
	Name         string `json:"name"`
}

func TestAccResourceTfExport(t *testing.T) {
	t.Parallel()
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

func TestAccResourceTfExportByName(t *testing.T) {
	var (
		exportResource1 = "test-export1"

		userResource1 = "test-user1"
		userEmail1    = "terraform-" + uuid.NewString() + "@example.com"
		userName1     = "John Data-" + uuid.NewString()

		userResource2 = "test-user2"
		userEmail2    = "terraform-" + uuid.NewString() + "@example.com"
		userName2     = "John Data-" + uuid.NewString()

		queueResource   = "test-queue"
		queueName       = "Terraform Test Queue-" + uuid.NewString()
		queueDesc       = "This is a test"
		queueAcwTimeout = 200000
	)

	testUser1 := &UserExport{
		Name:  userName1,
		Email: userEmail1,
		State: "active",
	}

	testUser2 := &UserExport{
		Name:  userName2,
		Email: userEmail2,
		State: "active",
	}

	testQueue := &QueueExport{
		Name:         queueName,
		Description:  queueDesc,
		AcwTimeoutMs: queueAcwTimeout,
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				// Generate a user and export it
				Config: generateBasicUserResource(
					userResource1,
					userEmail1,
					userName1,
				),
			},
			{
				// Generate a user and export it
				Config: generateBasicUserResource(
					userResource1,
					userEmail1,
					userName1,
				) + generateTfExportByName(
					exportResource1,
					exportTestDir,
					trueValue,
					[]string{strconv.Quote("genesyscloud_user::" + userEmail1)},
					"",
					falseValue,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_tf_export."+exportResource1,
						"resource_types.0", "genesyscloud_user::"+userEmail1),
					testUserExport(exportTestDir+"/"+defaultTfJSONFile, "genesyscloud_user", sanitizeResourceName(userEmail1), testUser1),
				),
			},
			{
				// Generate a queue as well and export it
				Config: generateBasicUserResource(
					userResource1,
					userEmail1,
					userName1,
				) + generateRoutingQueueResource(
					queueResource,
					queueName,
					queueDesc,
					nullValue,                          // MANDATORY_TIMEOUT
					fmt.Sprintf("%v", queueAcwTimeout), // acw_timeout
					nullValue,                          // ALL
					nullValue,                          // auto_answer_only true
					nullValue,                          // No calling party name
					nullValue,                          // No calling party number
					nullValue,                          // enable_manual_assignment false
					nullValue,                          // enable_transcription false
				) + generateTfExportByName(
					exportResource1,
					exportTestDir,
					trueValue,
					[]string{
						strconv.Quote("genesyscloud_user::" + userEmail1),
						strconv.Quote("genesyscloud_routing_queue::" + queueName),
					},
					"",
					falseValue,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_tf_export."+exportResource1,
						"resource_types.0", "genesyscloud_user::"+userEmail1),
					resource.TestCheckResourceAttr("genesyscloud_tf_export."+exportResource1,
						"resource_types.1", "genesyscloud_routing_queue::"+queueName),
					testUserExport(exportTestDir+"/"+defaultTfJSONFile, "genesyscloud_user", sanitizeResourceName(userEmail1), testUser1),
					testQueueExport(exportTestDir+"/"+defaultTfJSONFile, "genesyscloud_routing_queue", sanitizeResourceName(queueName), testQueue),
				),
			},
			{
				// Export all trunk base settings as well
				Config: generateBasicUserResource(
					userResource1,
					userEmail1,
					userName1,
				) + generateRoutingQueueResource(
					queueResource,
					queueName,
					queueDesc,
					nullValue,                          // MANDATORY_TIMEOUT
					fmt.Sprintf("%v", queueAcwTimeout), // acw_timeout
					nullValue,                          // ALL
					nullValue,                          // auto_answer_only true
					nullValue,                          // No calling party name
					nullValue,                          // No calling party number
					nullValue,                          // enable_manual_assignment false
					nullValue,                          // enable_transcription false
				) + generateTfExportByName(
					exportResource1,
					exportTestDir,
					trueValue,
					[]string{
						strconv.Quote("genesyscloud_user::" + userEmail1),
						strconv.Quote("genesyscloud_routing_queue::" + queueName),
						strconv.Quote("genesyscloud_telephony_providers_edges_trunkbasesettings"),
					},
					"",
					falseValue,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"genesyscloud_tf_export."+exportResource1, "resource_types.0",
						"genesyscloud_user::"+userEmail1),
					resource.TestCheckResourceAttr(
						"genesyscloud_tf_export."+exportResource1, "resource_types.1",
						"genesyscloud_routing_queue::"+queueName),
					resource.TestCheckResourceAttr(
						"genesyscloud_tf_export."+exportResource1, "resource_types.2",
						"genesyscloud_telephony_providers_edges_trunkbasesettings"),
					testUserExport(exportTestDir+"/"+defaultTfJSONFile, "genesyscloud_user", sanitizeResourceName(userEmail1), testUser1),
					testQueueExport(exportTestDir+"/"+defaultTfJSONFile, "genesyscloud_routing_queue", sanitizeResourceName(queueName), testQueue),
					testTrunkBaseSettingsExport(exportTestDir+"/"+defaultTfJSONFile, "genesyscloud_telephony_providers_edges_trunkbasesettings"),
				),
			},
			{
				// Export all trunk base settings as well
				Config: generateBasicUserResource(
					userResource1,
					userEmail1,
					userName1,
				) + generateBasicUserResource(
					userResource2,
					userEmail2,
					userName2,
				) + generateRoutingQueueResource(
					queueResource,
					queueName,
					queueDesc,
					nullValue,                          // MANDATORY_TIMEOUT
					fmt.Sprintf("%v", queueAcwTimeout), // acw_timeout
					nullValue,                          // ALL
					nullValue,                          // auto_answer_only true
					nullValue,                          // No calling party name
					nullValue,                          // No calling party number
					nullValue,                          // enable_manual_assignment false
					nullValue,                          // enable_transcription false
				) + generateTfExportByName(
					exportResource1,
					exportTestDir,
					trueValue,
					[]string{
						strconv.Quote("genesyscloud_user::" + userEmail1),
						strconv.Quote("genesyscloud_user::" + userEmail2),
						strconv.Quote("genesyscloud_routing_queue::" + queueName),
						strconv.Quote("genesyscloud_telephony_providers_edges_trunkbasesettings"),
					},
					"",
					falseValue,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"genesyscloud_tf_export."+exportResource1, "resource_types.0",
						"genesyscloud_user::"+userEmail1),
					resource.TestCheckResourceAttr(
						"genesyscloud_tf_export."+exportResource1, "resource_types.1",
						"genesyscloud_user::"+userEmail2),
					resource.TestCheckResourceAttr(
						"genesyscloud_tf_export."+exportResource1, "resource_types.2",
						"genesyscloud_routing_queue::"+queueName),
					resource.TestCheckResourceAttr(
						"genesyscloud_tf_export."+exportResource1, "resource_types.3",
						"genesyscloud_telephony_providers_edges_trunkbasesettings"),
					testUserExport(exportTestDir+"/"+defaultTfJSONFile, "genesyscloud_user", sanitizeResourceName(userEmail1), testUser1),
					testUserExport(exportTestDir+"/"+defaultTfJSONFile, "genesyscloud_user", sanitizeResourceName(userEmail2), testUser2),
					testQueueExport(exportTestDir+"/"+defaultTfJSONFile, "genesyscloud_routing_queue", sanitizeResourceName(queueName), testQueue),
					testTrunkBaseSettingsExport(exportTestDir+"/"+defaultTfJSONFile, "genesyscloud_telephony_providers_edges_trunkbasesettings"),
				),
			},
		},
		CheckDestroy: testVerifyExportsDestroyed,
	})
}

func TestAccResourceTfExportFormAsHCL(t *testing.T) {
	var (
		exportedContents string
		pathToHclFile    = exportTestDir + "/" + defaultTfHCLFile
		formName         = "terraform_form_evaluations_" + uuid.NewString()
		formResourceName = formName

		// Complete evaluation form
		evaluationForm1 = evaluationFormStruct{
			name:      formName,
			published: false,
			questionGroups: []evaluationFormQuestionGroupStruct{
				{
					name:                    "Test Question Group 1",
					defaultAnswersToHighest: true,
					defaultAnswersToNA:      true,
					naEnabled:               true,
					weight:                  1,
					manualWeight:            true,
					questions: []evaluationFormQuestionStruct{
						{
							text: "Did the agent perform the opening spiel?",
							answerOptions: []answerOptionStruct{
								{
									text:  "Yes",
									value: 1,
								},
								{
									text:  "No",
									value: 0,
								},
							},
						},
						{
							text:             "Did the agent greet the customer?",
							helpText:         "Help text here",
							naEnabled:        true,
							commentsRequired: true,
							isKill:           true,
							isCritical:       true,
							visibilityCondition: visibilityConditionStruct{
								combiningOperation: "AND",
								predicates:         []string{"/form/questionGroup/0/question/0/answer/0", "/form/questionGroup/0/question/0/answer/1"},
							},
							answerOptions: []answerOptionStruct{
								{
									text:  "Yes",
									value: 1,
								},
								{
									text:  "No",
									value: 0,
								},
							},
						},
					},
				},
				{
					name:   "Test Question Group 2",
					weight: 2,
					questions: []evaluationFormQuestionStruct{
						{
							text: "Did the agent offer to sell product?",
							answerOptions: []answerOptionStruct{
								{
									text:  "Yes",
									value: 1,
								},
								{
									text:  "No",
									value: 0,
								},
							},
						},
					},
					visibilityCondition: visibilityConditionStruct{
						combiningOperation: "AND",
						predicates:         []string{"/form/questionGroup/0/question/0/answer/1"},
					},
				},
			},
		}
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: generateEvaluationFormResource(formResourceName, &evaluationForm1),
				Check: resource.ComposeTestCheckFunc(
					validateEvaluationFormAttributes(formResourceName, evaluationForm1),
				),
			},
			{
				Config: generateEvaluationFormResource(formResourceName, &evaluationForm1) + generateTfExportByName(
					formResourceName,
					exportTestDir,
					trueValue,
					[]string{strconv.Quote("genesyscloud_quality_forms_evaluation::" + formName)},
					"",
					trueValue,
				),
				Check: resource.ComposeTestCheckFunc(
					getExportedFileContents(pathToHclFile, &exportedContents),
				),
			},
		},
		CheckDestroy: testVerifyExportsDestroyed,
	})

	exportedContents = removeTfConfigBlock(exportedContents)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: exportedContents,
				Check: resource.ComposeTestCheckFunc(
					validateEvaluationFormAttributes(formResourceName, evaluationForm1),
				),
			},
		},
		CheckDestroy: testVerifyExportsDestroyed,
	})
}

func TestAccResourceTfExportQueueAsHCL(t *testing.T) {
	var (
		exportContents string
		pathToHclFile  = exportTestDir + "/" + defaultTfHCLFile
	)

	// routing queue attributes
	var (
		queueName      = fmt.Sprintf("Charlie_Test_Queue_%v", uuid.NewString())
		queueID        = queueName
		description    = "This is a test queue"
		autoAnswerOnly = "true"

		alertTimeoutSec = "30"
		slPercentage    = "0.7"
		slDurationMs    = "10000"

		rrOperator    = "MEETS_THRESHOLD"
		rrThreshold   = "9"
		rrWaitSeconds = "300"

		chatScriptID  = "81ddba00-9fad-11e7-9a00-3137c42c4ae9"
		emailScriptID = "153fcff5-597e-4f17-94e5-17eac456a0b2"
	)

	routingQueue := generateRoutingQueueResource(
		queueID,
		queueName,
		description,
		strconv.Quote("MANDATORY_TIMEOUT"),
		"300000",
		strconv.Quote("BEST"),
		autoAnswerOnly,
		strconv.Quote("Example Inc."),
		nullValue,
		"true",
		"true",
		generateMediaSettings("media_settings_call", alertTimeoutSec, slPercentage, slDurationMs),
		generateRoutingRules(rrOperator, rrThreshold, rrWaitSeconds),
		generateDefaultScriptIDs(chatScriptID, emailScriptID),
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: routingQueue,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_routing_queue."+queueID, "name", queueName),
					resource.TestCheckResourceAttr("genesyscloud_routing_queue."+queueID, "description", description),
					resource.TestCheckResourceAttr("genesyscloud_routing_queue."+queueID, "auto_answer_only", "true"),
					resource.TestCheckResourceAttr("genesyscloud_routing_queue."+queueID, "default_script_ids.CHAT", chatScriptID),
					resource.TestCheckResourceAttr("genesyscloud_routing_queue."+queueID, "default_script_ids.EMAIL", emailScriptID),
					validateMediaSettings(queueID, "media_settings_call", alertTimeoutSec, slPercentage, slDurationMs),
					validateRoutingRules(queueID, 0, rrOperator, rrThreshold, rrWaitSeconds),
				),
			},
			{
				Config: routingQueue + generateTfExportByName("export",
					exportTestDir,
					trueValue,
					[]string{strconv.Quote("genesyscloud_routing_queue::" + queueName)},
					"",
					trueValue,
				),
				Check: resource.ComposeTestCheckFunc(
					getExportedFileContents(pathToHclFile, &exportContents),
				),
			},
		},
		CheckDestroy: testVerifyExportsDestroyed,
	})

	exportContents = removeTfConfigBlock(exportContents)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: exportContents,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_routing_queue."+queueID, "name", queueName),
					resource.TestCheckResourceAttr("genesyscloud_routing_queue."+queueID, "description", description),
					resource.TestCheckResourceAttr("genesyscloud_routing_queue."+queueID, "auto_answer_only", "true"),
					resource.TestCheckResourceAttr("genesyscloud_routing_queue."+queueID, "default_script_ids.CHAT", chatScriptID),
					resource.TestCheckResourceAttr("genesyscloud_routing_queue."+queueID, "default_script_ids.EMAIL", emailScriptID),
					validateMediaSettings(queueID, "media_settings_call", alertTimeoutSec, slPercentage, slDurationMs),
					validateRoutingRules(queueID, 0, rrOperator, rrThreshold, rrWaitSeconds),
				),
			},
		},
	})
}

func removeTfConfigBlock(export string) string {
	return strings.Replace(export, terraformHCLBlock, "", -1)
}

func testUserExport(filePath, resourceType, resourceName string, expectedUser *UserExport) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		raw, err := getResourceDefinition(filePath, resourceType)
		if err != nil {
			return err
		}

		if raw[resourceName] == nil {
			return fmt.Errorf("expected a resource name for the resource type %s", resourceType)
		}

		var r *json.RawMessage
		if err := json.Unmarshal(*raw[resourceName], &r); err != nil {
			return err
		}

		exportedUser := &UserExport{}
		if err := json.Unmarshal(*r, exportedUser); err != nil {
			return err
		}

		if *exportedUser != *expectedUser {
			return fmt.Errorf("objects are not equal. Expected: %v. Got: %v", *expectedUser, *exportedUser)
		}

		return nil
	}
}

func testQueueExport(filePath, resourceType, resourceName string, expectedQueue *QueueExport) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		raw, err := getResourceDefinition(filePath, resourceType)
		if err != nil {
			return err
		}

		var r *json.RawMessage
		if err := json.Unmarshal(*raw[resourceName], &r); err != nil {
			return err
		}

		exportedQueue := &QueueExport{}
		if err := json.Unmarshal(*r, exportedQueue); err != nil {
			return err
		}

		if *exportedQueue != *expectedQueue {
			return fmt.Errorf("objects are not equal. Expected: %v. Got: %v", *expectedQueue, *exportedQueue)
		}

		return nil
	}
}

func testTrunkBaseSettingsExport(filePath, resourceType string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		raw, err := getResourceDefinition(filePath, resourceType)
		if err != nil {
			return err
		}

		if len(raw) == 0 {
			return fmt.Errorf("expected several %v to be exported", resourceType)
		}

		return nil
	}
}

func getResourceDefinition(filePath, resourceType string) (map[string]*json.RawMessage, error) {
	tfExport, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var tfExportRaw map[string]*json.RawMessage
	if err := json.Unmarshal(tfExport, &tfExportRaw); err != nil {
		return nil, err
	}

	var resourceRaw map[string]*json.RawMessage
	if err := json.Unmarshal(*tfExportRaw["resource"], &resourceRaw); err != nil {
		return nil, err
	}

	if resourceRaw[resourceType] == nil {
		return nil, fmt.Errorf("%s not found in the config file", resourceType)
	}

	var r map[string]*json.RawMessage
	if err := json.Unmarshal(*resourceRaw[resourceType], &r); err != nil {
		return nil, err
	}

	return r, nil
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
		cycleResources := make([][]string, 0)
		for _, cycle := range cycles {
			cycleTemp := make([]string, len(cycle))
			for j, cycleNode := range cycle {
				if resNames != nil {
					cycleTemp[j] = resNames[cycleNode.ID()]
				}
			}
			if !isIgnoredReferenceCycle(cycleTemp) {
				cycleResources = append(cycleResources, cycleTemp)
			}
		}

		if len(cycleResources) > 0 {
			t.Fatalf("Found the following potential reference cycles:\n %s", cycleResources)
		}
	}
}

func isIgnoredReferenceCycle(cycle []string) bool {
	// Some cycles cannot be broken with a schema change and must be dealt with in the config
	// These cycles can be ignored by this test
	ignoredCycles := [][]string{
		// Email routes contain a ref to an inbound queue ID, and queues contain a ref to an outbound email route
		{"genesyscloud_routing_queue", "genesyscloud_routing_email_route", "genesyscloud_routing_queue"},
		{"genesyscloud_routing_email_route", "genesyscloud_routing_queue", "genesyscloud_routing_email_route"},
	}

	for _, ignored := range ignoredCycles {
		if strArrayEquals(ignored, cycle) {
			return true
		}
	}
	return false
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
			"genesyscloud_architect_datatable",
			"genesyscloud_architect_datatable_row",
			"genesyscloud_architect_ivr",
			"genesyscloud_architect_schedules",
			"genesyscloud_architect_schedulegroups",
			"genesyscloud_architect_user_prompt",
			"genesyscloud_auth_division",
			"genesyscloud_auth_role",
			"genesyscloud_group",
			"genesyscloud_group_roles",
			"genesyscloud_idp_adfs",
			"genesyscloud_idp_generic",
			"genesyscloud_idp_gsuite",
			"genesyscloud_idp_okta",
			"genesyscloud_idp_onelogin",
			"genesyscloud_idp_ping",
			"genesyscloud_idp_salesforce",
			"genesyscloud_integration",
			"genesyscloud_integration_action",
			"genesyscloud_integration_credential",
			"genesyscloud_location",
			"genesyscloud_oauth_client",
			"genesyscloud_routing_email_domain",
			"genesyscloud_routing_email_route",
			"genesyscloud_routing_language",
			"genesyscloud_routing_queue",
			"genesyscloud_routing_skill",
			"genesyscloud_routing_utilization",
			"genesyscloud_routing_wrapupcode",
			"genesyscloud_telephony_providers_edges_did_pool",
			"genesyscloud_telephony_providers_edges_edge_group",
			"genesyscloud_telephony_providers_edges_phone",
			"genesyscloud_telephony_providers_edges_site",
			"genesyscloud_telephony_providers_edges_phonebasesettings",
			"genesyscloud_telephony_providers_edges_trunkbasesettings",
			"genesyscloud_telephony_providers_edges_trunk",
			"genesyscloud_user",
			"genesyscloud_user_roles"
		]
		exclude_attributes = [%s]
    }
	`, resourceID, directory, includeState, excludedAttributes)
}

func generateTfExportByName(
	resourceID string,
	directory string,
	includeState string,
	items []string,
	excludedAttributes string,
	exportAsHCL string) string {
	return fmt.Sprintf(`resource "genesyscloud_tf_export" "%s" {
		directory = "%s"
		include_state_file = %s
		resource_types = [%s]
		exclude_attributes = [%s]
		export_as_hcl = %s
    }
	`, resourceID, directory, includeState, strings.Join(items, ","), excludedAttributes, exportAsHCL)
}

func getExportedFileContents(filename string, result *string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		d, err := ioutil.ReadFile(filename)
		if err != nil {
			return fmt.Errorf("error reading file: %v\n", err)
		}
		*result = string(d)
		return nil
	}
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

func validateEvaluationFormAttributes(resourceName string, form evaluationFormStruct) resource.TestCheckFunc {
	return resource.ComposeTestCheckFunc(
		resource.TestCheckResourceAttr("genesyscloud_quality_forms_evaluation."+resourceName, "name", resourceName),
		resource.TestCheckResourceAttr("genesyscloud_quality_forms_evaluation."+resourceName, "published", falseValue),
		resource.TestCheckResourceAttr("genesyscloud_quality_forms_evaluation."+resourceName, "question_groups.0.name", form.questionGroups[0].name),
		resource.TestCheckResourceAttr("genesyscloud_quality_forms_evaluation."+resourceName, "question_groups.0.weight", fmt.Sprintf("%v", form.questionGroups[0].weight)),
		resource.TestCheckResourceAttr("genesyscloud_quality_forms_evaluation."+resourceName, "question_groups.0.questions.1.text", form.questionGroups[0].questions[1].text),
		resource.TestCheckResourceAttr("genesyscloud_quality_forms_evaluation."+resourceName, "question_groups.1.questions.0.answer_options.0.text", form.questionGroups[1].questions[0].answerOptions[0].text),
		resource.TestCheckResourceAttr("genesyscloud_quality_forms_evaluation."+resourceName, "question_groups.1.questions.0.answer_options.1.value", fmt.Sprintf("%v", form.questionGroups[1].questions[0].answerOptions[1].value)),
		resource.TestCheckResourceAttr("genesyscloud_quality_forms_evaluation."+resourceName, "question_groups.0.questions.1.visibility_condition.0.combining_operation", form.questionGroups[0].questions[1].visibilityCondition.combiningOperation),
		resource.TestCheckResourceAttr("genesyscloud_quality_forms_evaluation."+resourceName, "question_groups.0.questions.1.visibility_condition.0.predicates.0", form.questionGroups[0].questions[1].visibilityCondition.predicates[0]),
		resource.TestCheckResourceAttr("genesyscloud_quality_forms_evaluation."+resourceName, "question_groups.0.questions.1.visibility_condition.0.predicates.1", form.questionGroups[0].questions[1].visibilityCondition.predicates[1]),
	)
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
