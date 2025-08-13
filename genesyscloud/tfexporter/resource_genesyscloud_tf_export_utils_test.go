package tfexporter

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v165/platformclientv2"
	architectFlow "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/architect_flow"
	authDivision "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/auth_division"
	obContactList "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/outbound_contact_list"
	qualityFormsEvaluation "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/quality_forms_evaluation"
	routingQueue "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/routing_queue"
	routingWrapupcode "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/routing_wrapupcode"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/user"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/lists"
)

type UserExport struct {
	Email                 string `json:"email"`
	ExportedLabel         string `json:"name"`
	State                 string `json:"state"`
	OriginalResourceLabel string ``
}

type QueueExport struct {
	AcwTimeoutMs          int    `json:"acw_timeout_ms"`
	Description           string `json:"description"`
	ExportedLabel         string `json:"name"`
	OriginalResourceLabel string ``
}

type WrapupcodeExport struct {
	Name                  string `json:"name"`
	OriginalResourceLabel string ``
}

func testUserExport(filePath, resourceType, resourceLabel string, expectedUser *UserExport) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		raw, err := getResourceDefinition(filePath, resourceType)
		if err != nil {
			return err
		}

		if raw[resourceLabel] == nil {
			return fmt.Errorf("expected a resource name for the resource type %s", resourceType)
		}

		var r *json.RawMessage
		if err := json.Unmarshal(*raw[resourceLabel], &r); err != nil {
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

// testQueueExportEqual  Checks to see if the queues passed match the expected value
func testQueueExportEqual(filePath, resourceType, resourceLabel string, expectedQueue QueueExport) resource.TestCheckFunc {
	mccMutex.Lock()
	defer mccMutex.Unlock()
	return func(state *terraform.State) error {
		expectedQueue.OriginalResourceLabel = "" //Setting the resource label to be empty because it is not needed
		raw, err := getResourceDefinition(filePath, resourceType)
		if err != nil {
			return err
		}

		// Check if the raw data or label is nil
		if raw == nil {
			return fmt.Errorf("raw data is nil")
		}
		if _, ok := raw[resourceLabel]; !ok {
			return fmt.Errorf("resource label not found in raw data")
		}

		if _, ok := raw[resourceLabel]; !ok {
			return fmt.Errorf("failed to find resource %s in resource definition", resourceLabel)
		}

		var r *json.RawMessage
		if err := json.Unmarshal(*raw[resourceLabel], &r); err != nil {
			return err
		}

		// Check if r is nil
		if r == nil {
			return fmt.Errorf("unmarshaled raw message is nil")
		}

		exportedQueue := &QueueExport{}
		if err := json.Unmarshal(*r, exportedQueue); err != nil {
			return err
		}

		// Check if exportedQueue is nil
		if exportedQueue == nil {
			return fmt.Errorf("exportedQueue is nil after unmarshaling")
		}

		if *exportedQueue != expectedQueue {
			return fmt.Errorf("objects are not equal. Expected: %v. Got: %v", expectedQueue, *exportedQueue)
		}

		return nil
	}
}

// testDependentContactList tests to see if the dependent contactListResource for the flow is exported.
func testDependentContactList(filePath, resourceType, resourceLabel string) resource.TestCheckFunc {
	return func(state *terraform.State) error {

		_, err := os.ReadFile(filePath)
		if err != nil {
			fmt.Println("Error reading file:", err)
			return err
		}

		raw, err := getResourceDefinition(filePath, resourceType)
		if err != nil {
			return err
		}

		var r *json.RawMessage
		if err := json.Unmarshal(*raw[resourceLabel], &r); err != nil {
			return err
		}

		return nil
	}
}

// testQueueExportMatchesRegEx tests to see if all of the queues retrieved in the export match the regex passed into it.
func testQueueExportMatchesRegEx(filePath, resourceType, regEx string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		tfExport, err := os.ReadFile(filePath)
		if err != nil {
			return err
		}

		var tfExportRaw map[string]*json.RawMessage
		if err := json.Unmarshal(tfExport, &tfExportRaw); err != nil {
			return err
		}

		var resourceRaw map[string]*json.RawMessage
		if err := json.Unmarshal(*tfExportRaw["resource"], &resourceRaw); err != nil {
			return err
		}

		var resources map[string]interface{}
		if json.Unmarshal(*resourceRaw[resourceType], &resources); err != nil {
			return err
		}

		for k := range resources {
			regEx := regexp.MustCompile(regEx)

			if !regEx.MatchString(k) {
				return fmt.Errorf("Resource %s::%s was found in the config file when it did not match the include regex: %s", resourceType, k, regEx)
			}
		}

		return nil
	}
}

// testWrapupcodeExportEqual  Checks to see if the wrapupcodes passed match the expected value
func testWrapupcodeExportEqual(filePath, resourceType, resourceLabel string, expectedWrapupcode WrapupcodeExport) resource.TestCheckFunc {
	mccMutex.Lock()
	defer mccMutex.Unlock()
	return func(state *terraform.State) error {
		expectedWrapupcode.OriginalResourceLabel = ""
		raw, err := getResourceDefinition(filePath, resourceType)
		if err != nil {
			return err
		}

		// Check if the raw data or name is nil
		if raw == nil {
			return fmt.Errorf("raw data is nil")
		}
		if _, ok := raw[resourceLabel]; !ok {
			return fmt.Errorf("resource name not found in raw data")
		}

		var r *json.RawMessage
		if err := json.Unmarshal(*raw[resourceLabel], &r); err != nil {
			return err
		}

		// Check if r is nil
		if r == nil {
			return fmt.Errorf("unmarshaled raw message is nil")
		}

		exportedWrapupcode := &WrapupcodeExport{}
		if err := json.Unmarshal(*r, exportedWrapupcode); err != nil {
			return err
		}

		// Check if exportedWrapupcode is nil
		if exportedWrapupcode == nil {
			return fmt.Errorf("exportedWrapupcode is nil after unmarshaling")
		}

		if *exportedWrapupcode != expectedWrapupcode {
			return fmt.Errorf("objects are not equal. Expected: %v. Got: %v", expectedWrapupcode, *exportedWrapupcode)
		}

		return nil
	}
}

func testQueueExportExcludesRegEx(filePath, resourceType, regEx string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		tfExport, err := os.ReadFile(filePath)
		if err != nil {
			return err
		}

		var tfExportRaw map[string]*json.RawMessage
		if err := json.Unmarshal(tfExport, &tfExportRaw); err != nil {
			return err
		}

		var resourceRaw map[string]*json.RawMessage
		if err := json.Unmarshal(*tfExportRaw["resource"], &resourceRaw); err != nil {
			return err
		}

		var resources map[string]interface{}
		if json.Unmarshal(*resourceRaw[resourceType], &resources); err != nil {
			return err
		}

		for k := range resources {
			regEx := regexp.MustCompile(regEx)

			if regEx.MatchString(k) {
				return fmt.Errorf("Resource %s::%s was found in the config file when it should have been excluded by the regex: %s", resourceType, k, regEx)
			}
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
	tfExport, err := os.ReadFile(filePath)
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

func generateExportResourceIncludeFilterWithEnableDepRes(
	resourceLabel,
	directory,
	includeStateFile,
	exportFormat,
	enableDepRes string,
	includeResources,
	dependsOn []string,
) string {
	return fmt.Sprintf(`
resource "genesyscloud_tf_export" "%s" {
	directory                    = "%s"
  	include_state_file           = %s
  	export_format                = %s
  	enable_dependency_resolution = %s
  	include_filter_resources     = [%s]
    depends_on = [%s]
}
`, resourceLabel, directory, includeStateFile, exportFormat, enableDepRes, strings.Join(includeResources, ", "), strings.Join(dependsOn, ", "))
}

func addContactsToContactList(state *terraform.State) error {
	outboundAPI := platformclientv2.NewOutboundApi()
	contactListResource := state.RootModule().Resources["genesyscloud_outbound_contact_list.contact_list"]
	if contactListResource == nil {
		return fmt.Errorf("genesyscloud_outbound_contact_list.contact_list contactListResource not found in state")
	}

	contactList, _, err := outboundAPI.GetOutboundContactlist(contactListResource.Primary.ID, false, false)
	if err != nil {
		return fmt.Errorf("genesyscloud_outbound_contact_list (%s) not available", contactListResource.Primary.ID)
	}
	contactsJSON := `[{
			"data": {
			  "First Name": "Asa",
			  "Last Name": "Acosta",
			  "Cell": "+13335554",
			  "Home": "3335552345"
			},
			"callable": true,
			"phoneNumberStatus": {}
		  },
		  {
			"data": {
			  "First Name": "Leonidas",
			  "Last Name": "Acosta",
			  "Cell": "4445551234",
			  "Home": "4445552345"
			},
			"callable": true,
			"phoneNumberStatus": {}
		  }]`
	var contacts []platformclientv2.Writabledialercontact
	err = json.Unmarshal([]byte(contactsJSON), &contacts)
	if err != nil {
		return fmt.Errorf("could not unmarshall JSON contacts to add to contact list")
	}
	_, _, err = outboundAPI.PostOutboundContactlistContacts(*contactList.Id, contacts, false, false, false)
	if err != nil {
		return fmt.Errorf("could not post contacts to contact list")
	}
	return nil
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
		if util.StrArrayEquals(ignored, cycle) {
			return true
		}
	}
	return false
}

func resNodeIndex(resourceType string, resourceTypes []string) int64 {
	for i, resType := range resourceTypes {
		if resourceType == resType {
			return int64(i)
		}
	}
	return -1
}

func generateTfExportResource(
	resourceLabel string,
	directory string,
	includeState string,
	excludedAttributes string) string {
	return fmt.Sprintf(`resource "genesyscloud_tf_export" "%s" {
		directory = "%s"
		include_state_file = %s
		include_filter_resources = [
			"genesyscloud_architect_datatable",
			"genesyscloud_architect_datatable_row",
			//"genesyscloud_flow",
			"genesyscloud_flow_milestone",
			//"genesyscloud_flow_outcome",
			"genesyscloud_architect_ivr",
			"genesyscloud_architect_schedules",
			"genesyscloud_architect_schedulegroups",
			"genesyscloud_architect_user_prompt",
			"genesyscloud_auth_division",
			"genesyscloud_auth_role",
			"genesyscloud_employeeperformance_externalmetrics_definitions",
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
			"genesyscloud_outbound_settings",
			"genesyscloud_responsemanagement_library",
			"genesyscloud_routing_email_domain",
			"genesyscloud_routing_email_route",
			"genesyscloud_routing_language",
			//"genesyscloud_routing_queue",
			"genesyscloud_routing_settings",
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
			//"genesyscloud_user_roles",
			"genesyscloud_webdeployments_configuration",
			"genesyscloud_webdeployments_deployment",
			"genesyscloud_knowledge_knowledgebase"
		]
		exclude_attributes = [%s]
	}
	`, resourceLabel, directory, includeState, excludedAttributes)
}

// generateTfExportResourceForCompress creates a resource to test compressed exported results
func generateTfExportResourceForCompress(
	resourceLabel string,
	directory string,
	includeState string,
	compressFlag string,
	includeResourcesFilter []string,
	excludedAttributes string) string {
	return fmt.Sprintf(`resource "genesyscloud_tf_export" "%s" {
		directory = "%s"
		include_state_file = %s
		compress=%s
		include_filter_resources = [%s]
		exclude_attributes = [%s]
	}
	`, resourceLabel, directory, includeState, compressFlag, strings.Join(includeResourcesFilter, ","), excludedAttributes)
}

func generateTfExportResourceMin(
	resourceLabel string,
	directory string,
	includeState string,
	excludedAttributes string) string {
	return fmt.Sprintf(`resource "genesyscloud_tf_export" "%s" {
		directory = "%s"
		include_state_file = %s
		resource_types = [
			"genesyscloud_routing_language",
			"genesyscloud_routing_settings",
			"genesyscloud_routing_skill",
			"genesyscloud_routing_utilization",
			"genesyscloud_routing_wrapupcode",
		]
		exclude_attributes = [%s]
	}
	`, resourceLabel, directory, includeState, excludedAttributes)
}

func generateTfExportByFilter(
	resourceLabel string,
	directory string,
	includeState string,
	resourceTypesFilter []string,
	excludedAttributes string,
	exportFormat string,
	logErrors string,
	dependencies []string) string {
	return fmt.Sprintf(`resource "genesyscloud_tf_export" "%s" {
		directory = "%s"
		include_state_file = %s
		resource_types = [%s]
		exclude_attributes = [%s]
		export_format = %s
		log_permission_errors = %s
		depends_on=[%s]
	}
	`, resourceLabel, directory, includeState, strings.Join(resourceTypesFilter, ","), excludedAttributes, exportFormat, logErrors, strings.Join(dependencies, ","))
}

func generateTfExportByIncludeFilterResources(
	resourceLabel string,
	directory string,
	includeState string,
	includeFilterResources []string,
	exportFormat string,
	splitByResource string,
	dependencies []string,
) string {
	return fmt.Sprintf(`resource "genesyscloud_tf_export" "%s" {
		directory = "%s"
		include_state_file = %s
		include_filter_resources = [%s]
		export_format = %s
		split_files_by_resource = %s
		depends_on = [%s]
	}
	`, resourceLabel, directory, includeState, strings.Join(includeFilterResources, ","), exportFormat, splitByResource, strings.Join(dependencies, ","))
}

func generateTfExportByFlowDependsOnResources(
	resourceLabel string,
	directory string,
	includeState string,
	includeFilterResources []string,
	exportFormat string,
	splitByResource string,
	dependsOn string,
) string {
	return fmt.Sprintf(`resource "genesyscloud_tf_export" "%s" {
		directory = "%s"
		include_state_file = %s
		include_filter_resources = [%s]
		export_format = %s
		split_files_by_resource = %s
		enable_dependency_resolution = %s
		depends_on = [time_sleep.wait_10_seconds]
	}
	`, resourceLabel, directory, includeState, strings.Join(includeFilterResources, ","), exportFormat, splitByResource, dependsOn)
}

func generateTfExportByExcludeFilterResources(
	resourceLabel string,
	directory string,
	includeState string,
	excludeFilterResources []string,
	exportFormat string,
	splitByResource string,
	dependencies []string,
) string {
	return fmt.Sprintf(`resource "genesyscloud_tf_export" "%s" {
		directory = "%s"
		include_state_file = %s
		exclude_filter_resources = [%s]
		log_permission_errors=true
		export_format = %s
		split_files_by_resource = %s
		depends_on=[%s]
	}
	`, resourceLabel, directory, includeState, strings.Join(excludeFilterResources, ","), exportFormat, splitByResource, strings.Join(dependencies, ","))
}

func generateTFExportResourceCustom(
	resourceLabel,
	directory,
	includeStateFile,
	exportFormat,
	useLegacyFlowExporter string,
	includeResources []string,
) string {
	return fmt.Sprintf(`
resource "%s" "%s" {
	directory                          = "%s"
	include_state_file                 = %s
	export_format                      = %s
	use_legacy_architect_flow_exporter = %s

	include_filter_resources = [%s]
}
`, ResourceType, resourceLabel, directory, includeStateFile, exportFormat, useLegacyFlowExporter, strings.Join(includeResources, "\n"))
}

func getExportedFileContents(filename string, result *string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		d, err := os.ReadFile(filename)
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
			return fmt.Errorf("failed to find file '%s'. Error: %w", filename, err)
		}
		return nil
	}
}

func validateFileNotCreated(filename string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		_, err := os.Stat(filename)
		if err == nil {
			return fmt.Errorf("expected '%s' to not exist", filename)
		}
		if !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("unexpected error while verifying file '%s' does not exist: %w", filename, err)
		}
		return nil
	}
}

func validateCompressedCreated(filename string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		_, err := filepath.Glob(filename)
		if err != nil {
			return fmt.Errorf("Failed to find file")
		}
		return nil
	}
}

func deleteTestCompressedZip(exportPath string, zipFileName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		dir, err := os.ReadDir(exportPath)
		if err != nil {
			return fmt.Errorf("Failed to read compressed zip %s", exportPath)
		}
		for _, d := range dir {
			os.RemoveAll(filepath.Join(exportPath, d.Name()))
		}
		files, err := filepath.Glob(zipFileName)

		if err != nil {
			return fmt.Errorf("Failed to get zip: %s", err)
		}
		for _, f := range files {
			if err := os.Remove(f); err != nil {
				return fmt.Errorf("Failed to delete: %s", err)
			}
		}

		return nil
	}
}

func testVerifyExportsDestroyedFunc(exportTestDir string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
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
}

func validateEvaluationFormAttributes(resourceLabel string, form qualityFormsEvaluation.EvaluationFormStruct) resource.TestCheckFunc {
	return resource.ComposeTestCheckFunc(
		resource.TestCheckResourceAttr("genesyscloud_quality_forms_evaluation."+resourceLabel, "name", resourceLabel),
		resource.TestCheckResourceAttr("genesyscloud_quality_forms_evaluation."+resourceLabel, "published", util.FalseValue),
		resource.TestCheckResourceAttr("genesyscloud_quality_forms_evaluation."+resourceLabel, "question_groups.0.name", form.QuestionGroups[0].Name),
		resource.TestCheckResourceAttr("genesyscloud_quality_forms_evaluation."+resourceLabel, "question_groups.0.weight", fmt.Sprintf("%v", form.QuestionGroups[0].Weight)),
		resource.TestCheckResourceAttr("genesyscloud_quality_forms_evaluation."+resourceLabel, "question_groups.0.questions.1.text", form.QuestionGroups[0].Questions[1].Text),
		resource.TestCheckResourceAttr("genesyscloud_quality_forms_evaluation."+resourceLabel, "question_groups.1.questions.0.answer_options.0.text", form.QuestionGroups[1].Questions[0].AnswerOptions[0].Text),
		resource.TestCheckResourceAttr("genesyscloud_quality_forms_evaluation."+resourceLabel, "question_groups.1.questions.0.answer_options.1.value", fmt.Sprintf("%v", form.QuestionGroups[1].Questions[0].AnswerOptions[1].Value)),
		resource.TestCheckResourceAttr("genesyscloud_quality_forms_evaluation."+resourceLabel, "question_groups.0.questions.1.visibility_condition.0.combining_operation", form.QuestionGroups[0].Questions[1].VisibilityCondition.CombiningOperation),
		resource.TestCheckResourceAttr("genesyscloud_quality_forms_evaluation."+resourceLabel, "question_groups.0.questions.1.visibility_condition.0.predicates.0", form.QuestionGroups[0].Questions[1].VisibilityCondition.Predicates[0]),
		resource.TestCheckResourceAttr("genesyscloud_quality_forms_evaluation."+resourceLabel, "question_groups.0.questions.1.visibility_condition.0.predicates.1", form.QuestionGroups[0].Questions[1].VisibilityCondition.Predicates[1]),
	)
}

// validateCompressedFile unzips and validates the exported resulted in the compressed folder
func validateCompressedFile(path string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		files, err := filepath.Glob(path)
		if err != nil {
			return err
		}
		for _, f := range files {
			reader, err := zip.OpenReader(f)
			if err != nil {
				return err
			}
			for _, file := range reader.File {
				err = validateCompressedConfigFiles(f, file)
				if err != nil {
					return err
				}
			}
		}
		return nil
	}
}

// validateCompressedConfigFiles validates the data inside the compressed json file
func validateCompressedConfigFiles(dirName string, file *zip.File) error {

	if file.FileInfo().Name() == defaultTfJSONFile {
		rc, _ := file.Open()

		buf := new(bytes.Buffer)
		buf.ReadFrom(rc)
		var data map[string]interface{}

		if err := json.Unmarshal(buf.Bytes(), &data); err != nil {
			return fmt.Errorf("failed to unmarshal json exportData to map variable: %v", err)
		}

		if _, ok := data["resource"]; !ok {
			return fmt.Errorf("config file missing resource attribute")
		}

		if _, ok := data["terraform"]; !ok {
			return fmt.Errorf("config file missing terraform attribute")
		}
		rc.Close()
		return nil
	}
	return nil
}

func validateConfigFile(path string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		result, err := loadJsonFileToMap(path)
		if err != nil {
			return err
		}

		if _, ok := result["resource"]; !ok {
			return fmt.Errorf("config file missing resource attribute")
		}

		if _, ok := result["terraform"]; !ok {
			return fmt.Errorf("config file missing terraform attribute")
		}
		return nil
	}
}

func validateHCLConfigFile(filename string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		parser := hclparse.NewParser()
		_, diag := parser.ParseHCLFile(filename)
		if diag.HasErrors() {
			return fmt.Errorf("Invalid HCL format: %v", diag)
		}
		return nil
	}
}

func validateMediaSettings(resourceLabel string, settingsAttr string, alertingTimeout string, slPercent string, slDurationMs string) resource.TestCheckFunc {
	return resource.ComposeAggregateTestCheckFunc(
		resource.TestCheckResourceAttr("genesyscloud_routing_queue."+resourceLabel, settingsAttr+".0.alerting_timeout_sec", alertingTimeout),
		resource.TestCheckResourceAttr("genesyscloud_routing_queue."+resourceLabel, settingsAttr+".0.service_level_percentage", slPercent),
		resource.TestCheckResourceAttr("genesyscloud_routing_queue."+resourceLabel, settingsAttr+".0.service_level_duration_ms", slDurationMs),
	)
}

func validateRoutingRules(resourceLabel string, ringNum int, operator string, threshold string, waitSec string) resource.TestCheckFunc {
	ringNumStr := strconv.Itoa(ringNum)
	return resource.ComposeAggregateTestCheckFunc(
		resource.TestCheckResourceAttr("genesyscloud_routing_queue."+resourceLabel, "routing_rules."+ringNumStr+".operator", operator),
		resource.TestCheckResourceAttr("genesyscloud_routing_queue."+resourceLabel, "routing_rules."+ringNumStr+".threshold", threshold),
		resource.TestCheckResourceAttr("genesyscloud_routing_queue."+resourceLabel, "routing_rules."+ringNumStr+".wait_seconds", waitSec),
	)
}

func buildQueueResources(queueExports []QueueExport) string {
	queueResourceDefinitions := ""
	for _, queueExport := range queueExports {
		queueResourceDefinitions = queueResourceDefinitions + routingQueue.GenerateRoutingQueueResource(
			queueExport.OriginalResourceLabel,
			queueExport.ExportedLabel,
			queueExport.Description,
			util.NullValue,                              // MANDATORY_TIMEOUT
			fmt.Sprintf("%v", queueExport.AcwTimeoutMs), // acw_timeout
			util.NullValue,                              // ALL
			util.NullValue,                              // auto_answer_only true
			util.NullValue,                              // No calling party name
			util.NullValue,                              // No calling party number
			util.NullValue,                              // enable_audio_monitoring false
			util.NullValue,                              // enable_manual_assignment false
			util.NullValue,                              //suppressCall_record_false
			util.NullValue,                              // enable_transcription false
			strconv.Quote("TimestampAndPriority"),
			util.NullValue,
			util.NullValue,
			util.NullValue,
		)
	}

	return queueResourceDefinitions
}

func buildUserResources(userExports []UserExport) string {
	userResourceDefinitions := ""
	for _, userExport := range userExports {
		userResourceDefinitions = userResourceDefinitions + user.GenerateBasicUserResource(
			userExport.OriginalResourceLabel,
			userExport.Email,
			userExport.ExportedLabel,
		)
	}

	return userResourceDefinitions
}

func buildWrapupcodeResources(wrapupcodeExports []WrapupcodeExport, divisionId string, description string) string {
	wrapupcodeesourceDefinitions := ""
	for _, wrapupcodeExport := range wrapupcodeExports {
		wrapupcodeesourceDefinitions = wrapupcodeesourceDefinitions + routingWrapupcode.GenerateRoutingWrapupcodeResource(
			wrapupcodeExport.OriginalResourceLabel,
			wrapupcodeExport.Name,
			divisionId,
			description,
		)
	}

	return wrapupcodeesourceDefinitions
}

// Returns random string. Helpful for regex export testing as unique prefix or postfix
func randString(length int) string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	letters := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

	s := make([]rune, length)
	for i := range s {
		s[i] = letters[r.Intn(len(letters))]
	}

	return string(s)
}

func GenerateOutboundCampaignBasicforFlowExport(
	contactListResourceLabel string,
	outboundFlowFilePath string,
	flowResourceLabel string,
	flowName string,
	divisionName,
	wrapupcodeResourceLabel string,
	contactListName string) string {
	referencedResources := GenerateReferencedResourcesForOutboundCampaignTests(
		contactListResourceLabel,
		outboundFlowFilePath,
		flowResourceLabel,
		flowName,
		divisionName,
		wrapupcodeResourceLabel,
		contactListName,
	)
	return fmt.Sprintf(`
%s
`, referencedResources)
}

func GenerateReferencedResourcesForOutboundCampaignTests(
	contactListResourceLabel string,
	outboundFlowFilePath string,
	flowResourceLabel string,
	flowName string,
	divisionName string,
	wrapUpCodeResourceLabel string,
	contactListName string,
) string {
	var (
		contactList             string
		callAnalysisResponseSet string
		divResourceLabel        = "test-division"
		divName                 = "terraform-" + uuid.NewString()
		description             = "Terraform wrapup code description"
	)
	if contactListResourceLabel != "" {
		contactList = obContactList.GenerateOutboundContactList(
			contactListResourceLabel,
			contactListName,
			util.NullValue,
			strconv.Quote("Cell"),
			[]string{strconv.Quote("Cell")},
			[]string{strconv.Quote("Cell"), strconv.Quote("Home"), strconv.Quote("zipcode")},
			util.FalseValue,
			util.NullValue,
			util.NullValue,
			obContactList.GeneratePhoneColumnsBlock("Cell", "cell", strconv.Quote("Cell")),
			obContactList.GeneratePhoneColumnsBlock("Home", "home", strconv.Quote("Home")))
	}

	callAnalysisResponseSet = authDivision.GenerateAuthDivisionBasic(divResourceLabel, divName) + routingWrapupcode.GenerateRoutingWrapupcodeResource(
		wrapUpCodeResourceLabel,
		"wrapupcode "+uuid.NewString(),
		"genesyscloud_auth_division."+divResourceLabel+".id",
		description,
	) + architectFlow.GenerateFlowResource(
		flowResourceLabel,
		outboundFlowFilePath,
		"",
		false,
		util.GenerateSubstitutionsMap(map[string]string{
			"flow_name":          flowName,
			"home_division_name": divisionName,
			"contact_list_name":  "${genesyscloud_outbound_contact_list." + contactListResourceLabel + ".name}",
			"wrapup_code_name":   "${genesyscloud_routing_wrapupcode." + wrapUpCodeResourceLabel + ".name}",
		}),
	)

	return fmt.Sprintf(`
			%s
			%s
		`, contactList, callAnalysisResponseSet)
}

// Check if flow is published, then check if flow name and type are correct
func validateFlow(flowResourcePath, flowName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		flowResource, ok := state.RootModule().Resources[flowResourcePath]
		if !ok {
			return fmt.Errorf("failed to find flow %s in state", flowResourcePath)
		}
		flowID := flowResource.Primary.ID
		architectAPI := platformclientv2.NewArchitectApi()

		log.Printf("Reading flow %s", flowID)
		flow, _, err := architectAPI.GetFlow(flowID, false)
		if err != nil {
			return fmt.Errorf("unexpected error: %s", err)
		}

		if flow == nil {
			return fmt.Errorf("Flow (%s) not found. ", flowID)
		}

		if *flow.Name != flowName {
			return fmt.Errorf("returned flow (%s) has incorrect name. Expect: %s, Actual: %s", flowID, flowName, *flow.Name)
		}

		return nil
	}
}

func generateTfExportResourceExportFormat(
	exportResourceLabel1 string,
	exportFormat string,
	includeResources []string,
	path string) string {

	includeResourceStr := "null"
	if includeResources != nil {
		includeResourceStr = "[" + strings.Join(formatStringArray(includeResources), ",") + "]"
	}

	return fmt.Sprintf(`resource "genesyscloud_tf_export" "%s" {
        export_format = %s
        include_filter_resources = %s
        directory = "%s"
    }
    `, exportResourceLabel1, exportFormat, includeResourceStr, path)
}

func formatStringArray(arr []string) []string {
	quotedStrings := make([]string, len(arr))
	for i, s := range arr {
		quotedStrings[i] = fmt.Sprintf(`"%s"`, s)
	}
	return quotedStrings
}

func validateStateFileHasPublishedAndUnpublished(filename string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		if _, err := os.Stat(filename); err != nil {
			return fmt.Errorf("failed to find file %s", filename)
		}

		stateData, err := loadJsonFileToMap(filename)
		if err != nil {
			return err
		}

		modules, ok := stateData["modules"].([]interface{})
		if !ok {
			return fmt.Errorf("unexpected structure for modules")
		}

		log.Println("Successfully loaded export config into map variable ")

		resources := make([]interface{}, 0, len(modules))
		for _, module := range modules {
			m, ok := module.(map[string]interface{})
			if !ok {
				continue
			}
			if r, ok := m["resources"].([]interface{}); ok {
				resources = append(resources, r...)
			}
		}

		log.Printf("checking that quality forms surveys exports published and unpublished in tf state")

		for _, r := range resources {
			out, ok := r.(map[string]interface{})
			if !ok {
				continue
			}

			instances, ok := out["instances"].([]interface{})
			if !ok {
				return fmt.Errorf("unexpected structure for form %s", filename)
			}

			res, ok := instances[0].(map[string]interface{})["attributes_flat"].(map[string]interface{})
			if !ok {
				return fmt.Errorf("unexpected structure attributes %s", filename)
			}

			name, ok := res["name"].(string)
			if !ok {
				return fmt.Errorf("unexpected name structure for form %s", filename)
			}

			published, ok := res["published"].(string)
			if !ok {
				return fmt.Errorf("unexpected published structure for form %s", filename)
			}

			if name == "test-published-form" {
				if published == "true" {
					log.Printf("Form with name '%s' is correctly exported as published\n", name)
				} else {
					return fmt.Errorf("Form with name '%s' is not correctly exported as published\n", name)
				}
			}

			if name == "test-unpublished-form" {
				if published == "false" {
					log.Printf("Form with name '%s' is correctly exported as unpublished\n", name)
				} else {
					return fmt.Errorf("Form with name '%s' is not correctly exported as unpublished\n", name)
				}
			}
		}

		return nil
	}
}

// validateStateFileAsData verifies that the default managed site 'PureCloud Voice - AWS' is exported as a data source
func validateStateFileAsData(filename, siteName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		_, err := os.Stat(filename)
		if err != nil {
			return fmt.Errorf("failed to find file %s", filename)
		}

		stateData, err := loadJsonFileToMap(filename)
		if err != nil {
			return err
		}
		log.Println("Successfully loaded export config into map variable ")

		// Check if data sources exist in the exported data
		if resources, ok := stateData["resources"].([]interface{}); ok {
			fmt.Printf("checking that managed site with name %s is exported as data source in tf state\n", siteName)

			// Validate each site's name
			for _, r := range resources {
				fmt.Printf("resource that managed site with name %v is exported as data source\n", r)

				res, ok := r.(map[string]interface{})
				if !ok {
					return fmt.Errorf("unexpected structure for site %s", siteName)
				}

				name, ok := res["name"].(string)
				if !ok {
					return fmt.Errorf("unexpected structure for site %s", siteName)
				}

				mode, ok := res["mode"].(string)
				if !ok {
					return fmt.Errorf("unexpected structure for site %s", siteName)
				}

				if name == strings.ReplaceAll(siteName, " ", "_") {
					if mode == "data" {
						log.Printf("Site with name '%s' is correctly exported as data source\n", siteName)
						return nil
					} else {
						log.Printf("Site with name '%s' is not correctly exported as data\n", siteName)
						return nil
					}
				}
			}
			return fmt.Errorf("no Resources '%s' was not exported as data source", siteName)
		} else {
			return fmt.Errorf("no data sources found in exported data")
		}
	}
}

func validatePublishedAndUnpublishedExported(configFile string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		_, err := os.ReadFile(configFile)
		if err != nil {
			return fmt.Errorf("failed to read state file %s: %v", configFile, err)
		}

		// Load the JSON content of the export file
		log.Println("Loading export config into map variable")
		exportData, err := loadJsonFileToMap(configFile)
		if err != nil {
			return err
		}

		if data, ok := exportData["resource"].(map[string]interface{}); ok {
			forms, ok := data["genesyscloud_quality_forms_survey"].(map[string]interface{})
			if !ok {
				return fmt.Errorf("no resources exported for genesyscloud_quality_forms_survey")
			}

			if publishedForm, ok := forms["test-published-form"].(map[string]interface{}); ok {
				if publishedForm["published"].(bool) != true {
					return fmt.Errorf("test-published-form is not published")
				}
			} else {
				return fmt.Errorf("test-published-form is not exported")
			}
			if unpublishedForm, ok := forms["test-unpublished-form"].(map[string]interface{}); ok {
				if unpublishedForm["published"].(bool) != false {
					return fmt.Errorf("test-unpublished-form is published")
				}
			} else {
				return fmt.Errorf("test-unpublished-form is not exported")
			}
		}
		return nil
	}
}

func validateExportedResourceAttributeValue[T comparable](configFile, resourcePath, attr string, value T) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		items := strings.Split(resourcePath, ".")
		resourceType := items[0]
		resourceLabel := items[1]

		_, err := os.ReadFile(configFile)
		if err != nil {
			return fmt.Errorf("failed to read file %s: %v", configFile, err)
		}

		// Load the JSON content of the export file
		log.Println("Loading export config into map variable")
		exportData, err := loadJsonFileToMap(configFile)
		if err != nil {
			return err
		}

		resourceMap, ok := exportData["resource"].(map[string]any)
		if !ok {
			return fmt.Errorf("could not find resource block in exported file %s", configFile)
		}

		resources, ok := resourceMap[resourceType].(map[string]any)
		if !ok {
			return fmt.Errorf("no %s resources exported", resourceType)
		}

		exportedResource, ok := resources[resourceLabel].(map[string]any)
		if !ok {
			return fmt.Errorf("no resource %s exported", resourcePath)
		}

		exportedValue, ok := exportedResource[attr].(T)
		if !ok {
			return fmt.Errorf("field %s not exported in resource %s config", attr, resourcePath)
		}

		if exportedValue != value {
			return fmt.Errorf("expected %s to equal %v, got %v", attr, value, exportedValue)
		}
		return nil
	}
}

// validateExportManagedSitesAsData verifies that the default managed site 'PureCloud Voice - AWS' is exported as a data source
func validateExportManagedSitesAsData(filename, siteName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		// Check if the file exists
		_, err := os.Stat(filename)
		if err != nil {
			return fmt.Errorf("failed to find file %s", filename)
		}

		// Load the JSON content of the export file
		log.Println("Loading export config into map variable")
		exportData, err := loadJsonFileToMap(filename)
		if err != nil {
			return err
		}
		log.Println("Successfully loaded export config into map variable ")

		// Check if data sources exist in the exported data
		if data, ok := exportData["data"].(map[string]interface{}); ok {
			sites, ok := data["genesyscloud_telephony_providers_edges_site"].(map[string]interface{})
			if !ok {
				return fmt.Errorf("no data sources exported for genesyscloud_telephony_providers_edges_site")
			}

			log.Printf("checking that managed site with name %s is exported as data source\n", siteName)

			// Validate each site's name
			for siteID, site := range sites {
				siteAttributes, ok := site.(map[string]interface{})
				if !ok {
					return fmt.Errorf("unexpected structure for site %s", siteID)
				}

				name, _ := siteAttributes["name"].(string)
				if name == siteName {
					log.Printf("Site %s with name '%s' is correctly exported as data source", siteID, siteName)
					return nil
				}
			}
			return fmt.Errorf("site with name '%s' was not exported as data source", siteName)
		} else {
			return fmt.Errorf("no data sources found in exported data")
		}
	}
}

func validatePromptsExported(filename string, expectedPrompts []string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		_, err := os.Stat(filename + "/genesyscloud.tf.json")
		if err != nil {
			return fmt.Errorf("failed to find file %s", filename)
		}

		log.Println("Loading export config into map variable")
		exportData, err := loadJsonFileToMap(filename + "/genesyscloud.tf.json")
		if err != nil {
			return fmt.Errorf("failed to unmarshal JSON from file %s: %s", filename, err)
		}
		log.Println("Successfully loaded export config into map variable")

		resources, ok := exportData["resource"].(map[string]interface{})
		if !ok {
			return fmt.Errorf("no 'resource' section found in exported JSON")
		}
		prompts, ok := resources["genesyscloud_architect_user_prompt"].(map[string]interface{})
		if !ok {
			return fmt.Errorf("no resources exported for genesyscloud_architect_user_prompt")
		}

		for _, promptID := range expectedPrompts {
			if promptData, found := prompts[promptID]; found {
				promptAttributes, ok := promptData.(map[string]interface{})
				if !ok {
					return fmt.Errorf("unexpected structure for prompt %s", promptID)
				}

				// Check that the "name" attribute matches the expected prompt name
				name, _ := promptAttributes["name"].(string)
				log.Printf("Verifying prompt '%s' with name '%s'", promptID, name)
				if name != promptID {
					return fmt.Errorf("expected prompt name '%s' but found '%s'", promptID, name)
				}
			} else {
				return fmt.Errorf("expected prompt with ID '%s' not found in exported data", promptID)
			}
		}

		log.Println("All expected prompts are correctly exported")
		return nil
	}
}

func removeTerraformProviderBlock(export string) string {
	return strings.Replace(export, terraformHCLBlock, "", -1)
}

// validateExportedCampaignScriptIds loads the exported content and validates that the custom resolver function
// resolved the script_id attr to the Default Outbound Script data source, and did not affect the campaign with a custom-made script
func validateExportedCampaignScriptIds(
	filename,
	customCampaignResourceLabel,
	defaultCampaignResourceLabel,
	expectedValueForCampaignWithDefaultScript,
	expectedValueForCampaignWithCustomScript string,
	verifyCustomScriptIdValue bool,
) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		_, err := os.Stat(filename)
		if err != nil {
			return fmt.Errorf("failed to find file %s", filename)
		}

		log.Println("Loading export config into map variable")
		exportData, err := loadJsonFileToMap(filename)
		if err != nil {
			return err
		}
		log.Println("Successfully loaded export config into map variable")

		if resources, ok := exportData["resource"].(map[string]interface{}); ok {
			campaigns, ok := resources["genesyscloud_outbound_campaign"].(map[string]interface{})
			if !ok {
				return fmt.Errorf("no campaign resources exported")
			}

			log.Println("Checking that campaign script_id values were resolved as expected")

			customCampaign, ok := campaigns[customCampaignResourceLabel].(map[string]interface{})
			if !ok {
				return fmt.Errorf("campaign with custom script was not exported")
			}

			if verifyCustomScriptIdValue {
				customCampaignScriptId, _ := customCampaign["script_id"].(string)
				if customCampaignScriptId != expectedValueForCampaignWithCustomScript {
					return fmt.Errorf("expected script ID to be '%s' for campaign with custom script, got '%s'", expectedValueForCampaignWithCustomScript, customCampaignScriptId)
				}
			}

			defaultCampaign, ok := campaigns[defaultCampaignResourceLabel].(map[string]interface{})
			if !ok {
				return fmt.Errorf("campaign with Default Outbound Script was not exported")
			}
			defaultCampaignScriptId, _ := defaultCampaign["script_id"].(string)
			if defaultCampaignScriptId != expectedValueForCampaignWithDefaultScript {
				return fmt.Errorf("expected script ID to be '%s' for campaign with default script, got '%s'", expectedValueForCampaignWithDefaultScript, defaultCampaignScriptId)
			}

			log.Println("Successfully verified that campaign script_ids were resolved correctly.")
		}

		return nil
	}
}

// validateNumberOfExportedDataSources validates that exactly one script data source is exported
func validateNumberOfExportedDataSources(filename string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		jsonFile, err := os.Open(filename)
		if err != nil {
			return fmt.Errorf("failed to open export file at path %s: %v", filename, err)
		}
		defer func(jsonFile *os.File) {
			_ = jsonFile.Close()
		}(jsonFile)

		byteValue, err := io.ReadAll(jsonFile)
		if err != nil {
			return fmt.Errorf("failed to unmarshal json exportData to map variable: %v", err)
		}

		exportAsString := fmt.Sprintf("%s", byteValue)
		numberOfDataSourcesExported := strings.Count(exportAsString, "\"Default_Outbound_Script\"")
		if numberOfDataSourcesExported != 1 {
			return fmt.Errorf("expected to find \"Default_Outbound_Script\" once in the exported content (actual %v). It is possible the Default Outbound Script data source is being exported more or less than once", numberOfDataSourcesExported)
		}

		return nil
	}
}

func loadJsonFileToMap(filename string) (map[string]interface{}, error) {
	var data map[string]interface{}

	jsonFile, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open export file at path %s: %v", filename, err)
	}
	defer func(jsonFile *os.File) {
		_ = jsonFile.Close()
	}(jsonFile)

	byteValue, _ := io.ReadAll(jsonFile)
	if err := json.Unmarshal(byteValue, &data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal json exportData to map variable: %v", err)
	}

	return data, nil
}

func testUserPromptAudioFileExport(filePath, resourceType, resourceLabel, exportDir, nameAttrRef string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		raw, err := getResourceDefinition(filePath, resourceType)
		if err != nil {
			return err
		}
		var r *json.RawMessage
		if err := json.Unmarshal(*raw[nameAttrRef], &r); err != nil {
			return err
		}

		var obj interface{}
		if err := json.Unmarshal(*r, &obj); err != nil {
			return err
		}

		// Collect each filename from resources list
		var fileNames []string
		if objMap, ok := obj.(map[string]interface{}); ok {
			if resourcesList, ok := objMap["resources"].([]interface{}); ok {
				for _, r := range resourcesList {
					if rMap, ok := r.(map[string]interface{}); ok {
						if fileNameStr, ok := rMap["filename"].(string); ok && fileNameStr != "" {
							fileNames = append(fileNames, fileNameStr)
						}
					}
				}
			}
		}

		// Check that file exists in export directory
		for _, filename := range fileNames {
			pathToWavFile := filepath.Join(exportDir, filename)
			if _, err := os.Stat(pathToWavFile); err != nil {
				return err
			}
		}

		return nil
	}
}

// verifyLabelsExistInExportedStateFile parses the exported state file and verifies that the specified labels exist within it
func verifyLabelsExistInExportedStateFile(filename, relevantResourceType string, labels []string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		_, err := os.Stat(filename)
		if err != nil {
			return fmt.Errorf("failed to find file %s", filename)
		}

		stateData, err := loadJsonFileToMap(filename)
		if err != nil {
			return err
		}
		log.Println("Successfully loaded export config into map variable ")

		allLabelsInStateFile, err := collectAllLabelsFromTfStateFile(relevantResourceType, stateData)
		if err != nil {
			return err
		}

		return validateLabelsExistInExportedFile(labels, allLabelsInStateFile, filename)
	}
}

// verifyLabelsExistInExportedTfConfig parses the exported TF config file and verifies that the specified labels exist within it
func verifyLabelsExistInExportedTfConfig(filename, relevantResourceType string, labels []string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		_, err := os.Stat(filename)
		if err != nil {
			return fmt.Errorf("failed to find file %s", filename)
		}

		stateData, err := loadJsonFileToMap(filename)
		if err != nil {
			return err
		}
		log.Println("Successfully loaded export config into map variable")

		allLabelsInExportedConfigFile, err := collectAllLabelsFromTfConfigFile(relevantResourceType, stateData)
		if err != nil {
			return err
		}

		return validateLabelsExistInExportedFile(labels, allLabelsInExportedConfigFile, filename)
	}
}

// validateLabelsExistInExportedFile confirms that the expected labels and the labels collected from the exported file align
// i.e. all expected labels are present and non appear more than once.
//
// Parameters:
//   - expectedLabels: []string - slice of labels that should exist in the file
//   - allLabelsInFile: []string - slice of all labels found in the exported file
//   - filename: string - name of the file being validated
//
// Returns:
//   - error: returns nil if validation passes, otherwise returns an error with description
func validateLabelsExistInExportedFile(expectedLabels, allLabelsInFile []string, filename string) error {
	mapOfLabelsToValidateUniqueness := make(map[string]string)

	for _, label := range allLabelsInFile {
		if _, exists := mapOfLabelsToValidateUniqueness[label]; exists {
			return fmt.Errorf("label %s appeared more than once in file %s", strconv.Quote(label), strconv.Quote(filename))
		}
		mapOfLabelsToValidateUniqueness[label] = "*"
	}

	for _, label := range expectedLabels {
		if !lists.ItemInSlice(label, allLabelsInFile) {
			return fmt.Errorf("expected to find label %s in exported file %s", strconv.Quote(label), strconv.Quote(filename))
		}
	}

	return nil
}

func collectAllLabelsFromTfStateFile(relevantResourceType string, stateFileData map[string]any) ([]string, error) {
	modules, ok := stateFileData["modules"].([]any)
	if !ok || len(modules) == 0 {
		return nil, fmt.Errorf("no modules found in exported state file")
	}

	module, ok := modules[0].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("invalid module in exported state file")
	}

	resources, ok := module["resources"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("expected resources to map type map[string]any, got %T", module["resources"])
	}

	allLabelsInStateFile := make([]string, 0)
	for k := range resources {
		if !strings.HasPrefix(k, relevantResourceType) {
			continue
		}
		allLabelsInStateFile = append(allLabelsInStateFile, parseLabelFromResourceKey(k))
	}

	return allLabelsInStateFile, nil
}

func collectAllLabelsFromTfConfigFile(relevantResourceType string, config map[string]any) ([]string, error) {
	resources, ok := config["resource"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("expected 'resource' to be map[string]any, got %T", resources)
	}

	relevantResources, ok := resources[relevantResourceType].(map[string]any)
	if !ok || len(relevantResources) == 0 {
		return nil, fmt.Errorf("failed to find resources of type %s", relevantResourceType)
	}

	allLabelsInExportedConfigFile := make([]string, 0)

	for label := range relevantResources {
		allLabelsInExportedConfigFile = append(allLabelsInExportedConfigFile, label)
	}

	return allLabelsInExportedConfigFile, nil
}

func parseLabelFromResourceKey(key string) string {
	strs := strings.Split(key, ".")
	return strs[1]
}

// sanitizeResourceHash is used by TestAccResourceTfExportSanitizedDuplicateLabels and must perform
// the same actions as sanitizerOriginal.SanitizeResourceHash in the resource_exporter package or else the
// test case will fail
func sanitizeResourceHash(originalLabel string) string {
	h := sha256.New()
	h.Write([]byte(originalLabel))
	return hex.EncodeToString(h.Sum(nil)[:10]) // Use first 10 characters of hash
}
