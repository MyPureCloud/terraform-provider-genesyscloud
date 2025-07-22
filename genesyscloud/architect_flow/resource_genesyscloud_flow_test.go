package architect_flow

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
	utilAws "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/aws"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/testrunner"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v162/platformclientv2"
)

// lockFlow will search for a specific flow and then lock it.  This is to specifically test the force_unlock flag where I want to create a flow,  simulate some one locking it and then attempt to
// do another CX as Code deploy.
func lockFlow(flowName string, flowType string) {
	archAPI := platformclientv2.NewArchitectApi()
	ctx := context.Background()
	util.WithRetries(ctx, 5*time.Second, func() *retry.RetryError {
		const pageSize = 100
		for pageNum := 1; ; pageNum++ {
			flows, resp, getErr := archAPI.GetFlows(nil, pageNum, pageSize, "", "", nil, flowName, "", "", "", "", "", "", "", false, false, false, "", "", nil)
			if getErr != nil {
				return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("error requesting flow %s | error: %s", flowName, getErr), resp))
			}

			if flows.Entities == nil || len(*flows.Entities) == 0 {
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("no flows found with name %s", flowName), resp))
			}

			for _, entity := range *flows.Entities {
				if *entity.Name == flowName && *entity.VarType == flowType {
					flow, response, err := archAPI.PostFlowsActionsCheckout(*entity.Id)

					if err != nil || response.Error != nil {
						return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("error requesting flow %s | error: %s", flowName, getErr), resp))
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
	var (
		flowResourceLabel = "test_force_unlock_flow1"
		flowName          = "Terraform Flow Test ForceUnlock-" + uuid.NewString()
		flowType          = "INBOUNDCALL"
		filePath          = filepath.Join(testrunner.RootDir, "examples/resources/genesyscloud_flow/inboundcall_flow_example.yaml")

		inboundcallConfig1 = fmt.Sprintf("inboundCall:\n  name: %s\n  defaultLanguage: en-us\n  startUpRef: ./menus/menu[mainMenu]\n  initialGreeting:\n    tts: Archy says hi!!!\n  menus:\n    - menu:\n        name: Main Menu\n        audio:\n          tts: You are at the Main Menu, press 9 to disconnect.\n        refId: mainMenu\n        choices:\n          - menuDisconnect:\n              name: Disconnect\n              dtmf: digit_9", flowName)
		inboundcallConfig2 = fmt.Sprintf("inboundCall:\n  name: %s\n  defaultLanguage: en-us\n  startUpRef: ./menus/menu[mainMenu]\n  initialGreeting:\n    tts: Archy says hi again!!!\n  menus:\n    - menu:\n        name: Main Menu\n        audio:\n          tts: You are at the Main Menu, press 9 to disconnect.\n        refId: mainMenu\n        choices:\n          - menuDisconnect:\n              name: Disconnect\n              dtmf: digit_9", flowName)
	)

	//Create an anonymous function that closes around the flow name and flow Type
	var flowLocFunc = func() {
		lockFlow(flowName, flowType)
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create flow
				Config: GenerateFlowResource(
					flowResourceLabel,
					filePath,
					inboundcallConfig1,
					false,
				),
				Check: resource.ComposeTestCheckFunc(
					validateFlow("genesyscloud_flow."+flowResourceLabel, flowName, "", flowType),
				),
			},
			{
				//Lock the flow, deploy, and check to make sure the flow is locked
				PreConfig: flowLocFunc, //This will lock the flow.
				Config: GenerateFlowResource(
					flowResourceLabel,
					filePath,
					inboundcallConfig2,
					true,
				),
				Check: resource.ComposeTestCheckFunc(
					validateFlowUnlocked("genesyscloud_flow."+flowResourceLabel),
					validateFlow("genesyscloud_flow."+flowResourceLabel, flowName, "", flowType),
				),
			},
			{
				// Import/Read
				ResourceName:            "genesyscloud_flow." + flowResourceLabel,
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
		flowResourceLabel1 = "test_flow1"
		flowResourceLabel2 = "test_flow2"
		flowName           = "Terraform Flow Test-" + uuid.NewString()
		flowDescription1   = "test description 1"
		flowDescription2   = "test description 2"
		flowType1          = "INBOUNDCALL"
		flowType2          = "INBOUNDEMAIL"
		filePath1          = filepath.Join(testrunner.RootDir, "examples/resources/genesyscloud_flow/inboundcall_flow_example.yaml")
		filePath2          = filepath.Join(testrunner.RootDir, "examples/resources/genesyscloud_flow/inboundcall_flow_example2.yaml")
		filePath3          = filepath.Join(testrunner.RootDir, "examples/resources/genesyscloud_flow/inboundcall_flow_example3.yaml")

		inboundcallConfig1 = fmt.Sprintf("inboundCall:\n  name: %s\n  description: %s\n  defaultLanguage: en-us\n  startUpRef: ./menus/menu[mainMenu]\n  initialGreeting:\n    tts: Archy says hi!!!\n  menus:\n    - menu:\n        name: Main Menu\n        audio:\n          tts: You are at the Main Menu, press 9 to disconnect.\n        refId: mainMenu\n        choices:\n          - menuDisconnect:\n              name: Disconnect\n              dtmf: digit_9", flowName, flowDescription1)
		inboundcallConfig2 = fmt.Sprintf("inboundCall:\n  name: %s\n  description: %s\n  defaultLanguage: en-us\n  startUpRef: ./menus/menu[mainMenu]\n  initialGreeting:\n    tts: Archy says hi!!!!!\n  menus:\n    - menu:\n        name: Main Menu\n        audio:\n          tts: You are at the Main Menu, press 9 to disconnect.\n        refId: mainMenu\n        choices:\n          - menuDisconnect:\n              name: Disconnect\n              dtmf: digit_9", flowName, flowDescription2)
	)

	inboundemailConfig1 := fmt.Sprintf(`inboundEmail:
    name: %s
    description: %s
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
`, flowName, flowDescription1)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create flow
				Config: GenerateFlowResource(
					flowResourceLabel1,
					filePath1,
					inboundcallConfig1,
					false,
				),
				Check: resource.ComposeTestCheckFunc(
					validateFlow("genesyscloud_flow."+flowResourceLabel1, flowName, flowDescription1, flowType1),
				),
			},
			{
				// Update flow description
				Config: GenerateFlowResource(
					flowResourceLabel1,
					filePath2,
					inboundcallConfig2,
					false,
				),
				Check: resource.ComposeTestCheckFunc(
					validateFlow("genesyscloud_flow."+flowResourceLabel1, flowName, flowDescription2, flowType1),
				),
			},
			{
				// Import/Read
				ResourceName:            "genesyscloud_flow." + flowResourceLabel1,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"filepath", "force_unlock", "file_content_hash"},
			},
			{
				// Create inboundemail flow
				Config: GenerateFlowResource(
					flowResourceLabel2,
					filePath3,
					inboundemailConfig1,
					false,
				),
				Check: resource.ComposeTestCheckFunc(
					validateFlow("genesyscloud_flow."+flowResourceLabel2, flowName, flowDescription1, flowType2),
				),
			},
			{
				// Import/Read
				ResourceName:            "genesyscloud_flow." + flowResourceLabel2,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"filepath", "force_unlock", "file_content_hash"},
			},
		},
		CheckDestroy: testVerifyFlowDestroyed,
	})
}

func TestAccResourceArchFlowSubstitutions(t *testing.T) {
	var (
		flowResourceLabel1 = "test_flow1"
		flowName           = "Terraform Flow Test-" + uuid.NewString()
		flowDescription1   = "description 1"
		flowDescription2   = "description 2"
		filePath1          = filepath.Join(testrunner.RootDir, "/examples/resources/genesyscloud_flow/inboundcall_flow_example_substitutions.yaml")
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create flow
				Config: GenerateFlowResource(
					flowResourceLabel1,
					filePath1,
					"",
					false,
					util.GenerateSubstitutionsMap(map[string]string{
						"flow_name":            flowName,
						"description":          flowDescription1,
						"default_language":     "en-us",
						"greeting":             "Archy says hi!!!",
						"menu_disconnect_name": "Disconnect",
					}),
				),
				Check: resource.ComposeTestCheckFunc(
					validateFlow("genesyscloud_flow."+flowResourceLabel1, flowName, flowDescription1, "INBOUNDCALL"),
				),
			},
			{
				// Update
				Config: GenerateFlowResource(
					flowResourceLabel1,
					filePath1,
					"",
					false,
					util.GenerateSubstitutionsMap(map[string]string{
						"flow_name":            flowName,
						"description":          flowDescription2,
						"default_language":     "en-us",
						"greeting":             "Archy says hi!!!",
						"menu_disconnect_name": "Disconnect",
					}),
				),
				Check: resource.ComposeTestCheckFunc(
					validateFlow("genesyscloud_flow."+flowResourceLabel1, flowName, flowDescription2, "INBOUNDCALL"),
				),
			},
		},
		CheckDestroy: testVerifyFlowDestroyed,
	})
}

func copyFile(src string, dest string) {
	bytesRead, err := os.ReadFile(src)

	if err != nil {
		log.Fatal(err)
	}

	err = os.WriteFile(dest, bytesRead, 0644)

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
	input, err := os.ReadFile(fileName)
	if err != nil {
		log.Fatalln(err)
	}

	lines := strings.Split(string(input), "\n")

	for i, line := range lines {
		lines[i] = strings.Replace(line, "You are at the Main Menu, press 9 to disconnect.", "Hi you are at the Main Menu, press 9 to disconnect.", 1)
	}

	output := strings.Join(lines, "\n")
	err = os.WriteFile(fileName, []byte(output), 0644)
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
		flowResourceLabel1 = "test_flow1"
		flowName           = "Terraform Flow Test-" + uuid.NewString()
		flowDescription1   = "description 1"
		flowDescription2   = "description 2"
		srcFile            = filepath.Join(testrunner.RootDir, "examples/resources/genesyscloud_flow/inboundcall_flow_example_substitutions.yaml")
		destFile           = filepath.Join(testrunner.RootDir, "examples/resources/genesyscloud_flow/inboundcall_flow_example_holder.yaml")
	)

	//Copy the example substitution file over to a temp file that can be manipulated and modified
	copyFile(srcFile, destFile)

	//Clean up the temporary file
	defer removeFile(destFile)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create flow
				Config: GenerateFlowResource(
					flowResourceLabel1,
					destFile,
					"",
					false,
					util.GenerateSubstitutionsMap(map[string]string{
						"flow_name":            flowName,
						"description":          flowDescription1,
						"default_language":     "en-us",
						"greeting":             "Archy says hi!!!",
						"menu_disconnect_name": "Disconnect",
					}),
				),
				Check: resource.ComposeTestCheckFunc(
					validateFlow("genesyscloud_flow."+flowResourceLabel1, flowName, flowDescription1, "INBOUNDCALL"),
				),
			},
			{ // Update the flow, but make sure that we touch the YAML file and change something int
				PreConfig: func() { transformFile(destFile) },
				Config: GenerateFlowResource(
					flowResourceLabel1,
					destFile,
					"",
					false,
					util.GenerateSubstitutionsMap(map[string]string{
						"flow_name":            flowName,
						"description":          flowDescription2,
						"default_language":     "en-us",
						"greeting":             "Archy says hi!!!",
						"menu_disconnect_name": "Disconnect",
					}),
				),
				Check: resource.ComposeTestCheckFunc(
					validateFlow("genesyscloud_flow."+flowResourceLabel1, flowName, flowDescription2, "INBOUNDCALL"),
				),
			},
		},
		CheckDestroy: testVerifyFlowDestroyed,
	})
}

// uploadTestFileToMinIO creates a MinIO S3 bucket and uploads a test file to it.
// This function is used by integration tests to set up S3-compatible storage
// for testing architect flow S3 integration functionality.
//
// Parameters:
//   - ctx: Context for the operation
//   - minioS3Client: MinIO S3 client for bucket and object operations
//   - bucketName: Name of the S3 bucket to create
//   - filePath: Local path to the file to upload
//
// The function performs the following operations:
// 1. Creates a new S3 bucket with the specified name
// 2. Extracts the filename from the file path
// 3. Uploads the file to the bucket using FPutObject
// 4. Logs the upload success with file size information
//
// Returns an error if any operation fails, nil on success.
func uploadTestFileToMinIO(ctx context.Context, minioS3Client *minio.Client, bucketName, filePath string) error {
	location := "us-east-1"

	err := minioS3Client.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{Region: location})
	if err != nil {
		// Check to see if we already own this bucket (which happens if you run this twice)
		exists, errBucketExists := minioS3Client.BucketExists(ctx, bucketName)
		if errBucketExists == nil && exists {
			log.Printf("We already own %s\n", bucketName)
		} else {
			return fmt.Errorf("failed to create bucket %s: %v", bucketName, err)
		}
	} else {
		log.Printf("Successfully created %s\n", bucketName)
	}

	fileName := filepath.Base(filePath)
	contentType := "application/yaml"

	// Upload the test file with FPutObject
	info, err := minioS3Client.FPutObject(ctx, bucketName, fileName, filePath, minio.PutObjectOptions{ContentType: contentType})
	if err != nil {
		return fmt.Errorf("failed to upload test file: %v", err)
	}

	log.Printf("Successfully uploaded %s of size %d\n", fileName, info.Size)
	return nil
}

// TestAccResourceArchFlowS3MinioIntegration tests the S3 integration functionality of the architect flow resource
// using MinIO as an S3-compatible service. This test validates the complete flow creation process
// from S3 file upload to flow deployment and verification.
//
// Test Flow:
// 1. Creates a temporary YAML file with flow configuration
// 2. Uploads the file to MinIO S3 bucket using FPutObject
// 3. Configures the architect flow resource with S3 filepath
// 4. Creates the flow in Genesys Cloud using the S3-hosted configuration
// 5. Verifies the flow was created with correct name and description
// 6. Cleans up by deleting the flow and temporary files
//
// This test ensures that the architect flow resource can successfully:
// - Read flow configurations from S3-compatible storage
// - Process YAML files stored in S3 buckets
// - Deploy flows using S3-hosted configuration files
// - Handle S3 authentication and file access
//
// Dependencies:
// - MinIO S3-compatible service (play.min.io)
// - AWS SDK for S3 operations
// - Genesys Cloud API for flow management
//
// Note: This is an integration test that requires external services and may take
// longer to execute than unit tests. It validates the complete S3 integration
// workflow rather than individual components.
func TestAccResourceArchFlowS3MinioIntegration(t *testing.T) {
	var (
		bucketName = "testbucket"

		flowName        = "A TF MinIO Test Flow " + uuid.NewString()
		flowDescription = "Example"
		fileName        = fmt.Sprintf("testfile-%s.yml", uuid.NewString())
		filePath        = "s3://" + bucketName + "/" + fileName

		ctx = context.Background()

		inboundcallConfig = fmt.Sprintf("inboundCall:\n  name: %s\n  description: %s\n  defaultLanguage: en-us\n  startUpRef: ./menus/menu[mainMenu]\n  initialGreeting:\n    tts: MinIO says hi!!!\n  menus:\n    - menu:\n        name: Main Menu\n        audio:\n          tts: You are at the Main Menu, press 9 to disconnect.\n        refId: mainMenu\n        choices:\n          - menuDisconnect:\n              name: Disconnect\n              dtmf: digit_9", flowName, flowDescription)
	)

	tempFilePath := filepath.Join(os.TempDir(), fileName)
	err := os.WriteFile(tempFilePath, []byte(inboundcallConfig), 0644)
	if err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}
	defer func() {
		os.Remove(tempFilePath)
	}()

	sdkConfig, err := provider.AuthorizeSdk()
	if err != nil {
		t.Fatalf("Failed to authorize sdk: %v", err)
	}

	t.Log("Creating minio client")
	minioClient, err := utilAws.NewMinIOS3Client("play.min.io")
	if err != nil {
		t.Fatalf("Failed to create minio client: %v", err)
	}

	t.Log("Uploading test file to minio")
	err = uploadTestFileToMinIO(ctx, minioClient.Client(), bucketName, tempFilePath)
	if err != nil {
		t.Fatalf("Failed to upload test file to minio: %v", err)
	}

	t.Log("Creating s3 client config")
	customS3Client := utilAws.NewS3ClientConfig().WithS3Client(minioClient)

	proxy := getArchitectFlowProxy(sdkConfig)
	proxy.s3Client = customS3Client

	flowConfig := map[string]any{
		"filepath": filePath,
	}

	resourceData := schema.TestResourceDataRaw(t, ResourceArchitectFlow().Schema, flowConfig)

	providerMeta := provider.ProviderMeta{
		ClientConfig: sdkConfig,
	}

	diags := createFlow(ctx, resourceData, &providerMeta)
	if diags.HasError() {
		t.Fatalf("Failed to create flow: %v", diags)
	}

	t.Logf("Flow created: %v", resourceData.Id())
	time.Sleep(2 * time.Second)

	t.Logf("Reading flow: %v", resourceData.Id())
	flow, _, err := proxy.GetFlow(ctx, resourceData.Id())
	if err != nil || flow == nil {
		t.Fatalf("Failed to read flow '%s': %v", resourceData.Id(), err)
	}

	if *flow.Name != flowName {
		t.Logf("Flow name mismatch: %s != %s", *flow.Name, flowName)
		t.Fail()
	}

	if *flow.Description != flowDescription {
		t.Logf("Flow description mismatch: %s != %s", *flow.Description, flowDescription)
		t.Fail()
	}

	t.Logf("Deleting flow: %v", resourceData.Id())
	resp, err := proxy.DeleteFlow(ctx, resourceData.Id())
	if err != nil {
		t.Logf("[WARN] Failed to delete flow: %v", err)
	} else {
		t.Logf("Flow deleted: %v", resp.StatusCode)
	}
}

func TestUnitSanitizeFlowName(t *testing.T) {
	// Define test cases with input and expected output
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "Single space",
			input:    "hello world",
			expected: "hello_world",
		},
		{
			name:     "Multiple spaces",
			input:    "hello   world",
			expected: "hello___world",
		},
		{
			name:     "Forward slashes",
			input:    "path/to/file",
			expected: "path_to_file",
		},
		{
			name:     "Back slashes",
			input:    "path\\to\\file",
			expected: "path_to_file",
		},
		{
			name:     "Mixed slashes and spaces",
			input:    "path/to\\file   name",
			expected: "path_to_file___name",
		},
		{
			name:     "Leading and trailing spaces",
			input:    " hello world  ",
			expected: "_hello_world__",
		},
		{
			name:     "Complex mixed case",
			input:    "  path/to\\file   name  with/\\spaces",
			expected: "__path_to_file___name__with__spaces",
		},
	}

	// Run all test cases
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizeFlowName(tt.input)
			if result != tt.expected {
				t.Errorf("sanitizeFlowName(%q) = %q, want %q",
					tt.input, result, tt.expected)
			}
		})
	}
}

// Check if flow is published, then check if flow name and type are correct
func validateFlow(flowResourcePath, name, description, flowType string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		flowResource, ok := state.RootModule().Resources[flowResourcePath]
		if !ok {
			return fmt.Errorf("Failed to find flow %s in state", flowResourcePath)
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

		if description != "" {
			if flow.Description == nil {
				return fmt.Errorf("returned flow (%s) has no description. Expected: '%s'", flowID, description)
			}
			if *flow.Description != description {
				return fmt.Errorf("returned flow (%s) has incorrect description. Expect: '%s', Actual: '%s'", flowID, description, *flow.Description)
			}
		}

		if *flow.VarType != flowType {
			return fmt.Errorf("Returned flow (%s) has incorrect type. Expect: %s, Actual: %s", flowID, flowType, *flow.VarType)
		}

		return nil
	}
}

// Will attempt to determine if a flow is unlocked. I check to see if a flow is locked, by attempting to check the flow again.  If the flow is locked the second checkout
// will fail with a 409 status code.  If the flow is unlocked, the status code will be a 200
func validateFlowUnlocked(flowResourcePath string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		flowResource, ok := state.RootModule().Resources[flowResourcePath]
		if !ok {
			return fmt.Errorf("Failed to find flow %s in state", flowResourcePath)
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
