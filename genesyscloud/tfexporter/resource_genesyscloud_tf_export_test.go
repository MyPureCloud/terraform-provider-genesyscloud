package tfexporter

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	integrationAction "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/integration_action"

	architectFlow "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/architect_flow"
	authDivision "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/auth_division"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/platform"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	qualityFormsEvaluation "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/quality_forms_evaluation"
	resourceExporter "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	routingQueue "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/routing_queue"
	telephonyProvidersEdgesSite "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/telephony_providers_edges_site"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/user"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/lists"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/testrunner"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"github.com/google/uuid"

	userPrompt "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/architect_user_prompt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"gonum.org/v1/gonum/graph/simple"
	"gonum.org/v1/gonum/graph/topo"
)

var (
	mccMutex sync.RWMutex
)

func init() {
	mccMutex = sync.RWMutex{}
}

// testSetup resets the client pool before each test to prevent client pool exhaustion
func testSetup(t *testing.T) {
	provider.ResetSDKClientPool()
}

// TestAccResourceTfExportIncludeFilterResourcesByRegEx will create 4 queues (three ending with -prod and then one watching with -test).  The
// The code will use a regex to include all queues that have a label that match a regular expression.  (e.g. -prod).  The test checks to see if any -test
// queues are exported.
func TestAccResourceTfExportIncludeFilterResourcesByRegEx(t *testing.T) {
	testSetup(t)
	var (
		exportTestDir       = testrunner.GetTestTempPath(".terraformregex" + uuid.NewString())
		exportResourceLabel = "test-export3"
		uniquePostfix       = randString(7)
		queueResources      = []QueueExport{
			{OriginalResourceLabel: "test-queue-prod-1", ExportedLabel: "test-queue-" + uuid.NewString() + "-prod-" + uniquePostfix, Description: "This is a test prod queue 1", AcwTimeoutMs: 200000},
			{OriginalResourceLabel: "test-queue-prod-2", ExportedLabel: "test-queue-" + uuid.NewString() + "-prod-" + uniquePostfix, Description: "This is a test prod queue 2", AcwTimeoutMs: 200000},
			{OriginalResourceLabel: "test-queue-prod-3", ExportedLabel: "test-queue-" + uuid.NewString() + "-prod-" + uniquePostfix, Description: "This is a test prod queue 3", AcwTimeoutMs: 200000},
			{OriginalResourceLabel: "test-queue-test-4", ExportedLabel: "test-queue-" + uuid.NewString() + "-prod-" + uniquePostfix, Description: "This is a test prod queue 4", AcwTimeoutMs: 200000},
		}
	)
	defer os.RemoveAll(exportTestDir)

	queueResourceDef := buildQueueResources(queueResources)
	baseConfig := queueResourceDef
	configWithExporter := baseConfig + generateTfExportByIncludeFilterResources(
		exportResourceLabel,
		exportTestDir,
		util.TrueValue,
		[]string{
			strconv.Quote("genesyscloud_routing_queue::.*-prod"),
		},
		strconv.Quote("json"),
		util.FalseValue,
		[]string{
			strconv.Quote("genesyscloud_routing_queue." + queueResources[0].OriginalResourceLabel),
			strconv.Quote("genesyscloud_routing_queue." + queueResources[1].OriginalResourceLabel),
			strconv.Quote("genesyscloud_routing_queue." + queueResources[2].OriginalResourceLabel),
		},
	)

	sanitizer := resourceExporter.NewSanitizerProvider()

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Generate a queue
				Config: baseConfig,
			},
			{
				// Now, export it
				Config: configWithExporter,
				Check: resource.ComposeTestCheckFunc(
					testQueueExportEqual(exportTestDir+"/"+defaultTfJSONFile, "genesyscloud_routing_queue", sanitizer.S.SanitizeResourceBlockLabel(queueResources[0].ExportedLabel), queueResources[0]),
					testQueueExportEqual(exportTestDir+"/"+defaultTfJSONFile, "genesyscloud_routing_queue", sanitizer.S.SanitizeResourceBlockLabel(queueResources[1].ExportedLabel), queueResources[1]),
					testQueueExportEqual(exportTestDir+"/"+defaultTfJSONFile, "genesyscloud_routing_queue", sanitizer.S.SanitizeResourceBlockLabel(queueResources[2].ExportedLabel), queueResources[2]),
					testQueueExportMatchesRegEx(exportTestDir+"/"+defaultTfJSONFile, "genesyscloud_routing_queue", "-prod"), //We should not find any "test" queues here because we only wanted to include queues that ended with a -prod
				),
			},
		},
		CheckDestroy: testVerifyExportsDestroyedFunc(exportTestDir),
	})
}

// TestAccResourceTfExportIncludeFilterResourcesByRegExAndSanitizedLabels will create 3 queues (twoc with foo bar, one to be excluded).
// The test ensures that resources can be exported directly by their actual label or their sanitized label.
func TestAccResourceTfExportIncludeFilterResourcesByRegExAndSanitizedLabels(t *testing.T) {
	testSetup(t)
	var (
		exportTestDir       = testrunner.GetTestTempPath(".terraformregex" + uuid.NewString())
		exportResourceLabel = "test-export3_1"
		uniquePostfix       = randString(7)
		queueResources      = []QueueExport{
			{OriginalResourceLabel: "test-queue-test-1", ExportedLabel: "include filter test - exclude me" + uuid.NewString() + uniquePostfix, Description: "This is an excluded bar test resource", AcwTimeoutMs: 200000},
			{OriginalResourceLabel: "test-queue-test-2", ExportedLabel: "include filter test - foo - bar me" + uuid.NewString() + uniquePostfix, Description: "This is a foo bar test resource", AcwTimeoutMs: 200000},
			{OriginalResourceLabel: "test-queue-test-3", ExportedLabel: "include filter test - fu - barre you" + uuid.NewString() + uniquePostfix, Description: "This is a foo bar test resource", AcwTimeoutMs: 200000},
		}
	)
	defer os.RemoveAll(exportTestDir)

	queueResourceDef := buildQueueResources(queueResources)

	baseConfig := queueResourceDef
	configWithExporter := baseConfig + generateTfExportByIncludeFilterResources(
		exportResourceLabel,
		exportTestDir,
		util.TrueValue,
		[]string{
			strconv.Quote("genesyscloud_routing_queue::include filter test - foo - bar me"),   // Unsanitized Label Resource
			strconv.Quote("genesyscloud_routing_queue::include_filter_test_-_fu_-_barre_you"), // Sanitized Label Resource
		},
		strconv.Quote("json"),
		util.FalseValue,
		[]string{
			strconv.Quote("genesyscloud_routing_queue." + queueResources[0].OriginalResourceLabel),
			strconv.Quote("genesyscloud_routing_queue." + queueResources[1].OriginalResourceLabel),
			strconv.Quote("genesyscloud_routing_queue." + queueResources[2].OriginalResourceLabel),
		},
	)

	sanitizer := resourceExporter.NewSanitizerProvider()

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Generate a queue
				Config: baseConfig,
			},
			{
				// Export the queue
				Config: configWithExporter,
				Check: resource.ComposeTestCheckFunc(
					testQueueExportEqual(exportTestDir+"/"+defaultTfJSONFile, "genesyscloud_routing_queue", sanitizer.S.SanitizeResourceBlockLabel(queueResources[1].ExportedLabel), queueResources[1]),
					testQueueExportEqual(exportTestDir+"/"+defaultTfJSONFile, "genesyscloud_routing_queue", sanitizer.S.SanitizeResourceBlockLabel(queueResources[2].ExportedLabel), queueResources[2]),
					testQueueExportExcludesRegEx(exportTestDir+"/"+defaultTfJSONFile, "genesyscloud_routing_queue", ".*exclude me.*"), //We should not find any "test" queues here because we only wanted to include queues that ended with a -prod
				),
			},
		},
		CheckDestroy: testVerifyExportsDestroyedFunc(exportTestDir),
	})
}

// TestAccResourceTfExportIncludeFilterResourcesByRegExExclusiveToResource
// will create two queues (one with a -prod suffix and one with a -test suffix)
// and two wrap up codes (one with a -prod suffix and one with a -test suffix)
// The code will use a regex to include queue resources with -prod suffix
// but wrap up code resources will not have any regex filter.
// eg. queue ending with prod should be exported but all wrap up codes should be
// exported as well
func TestAccResourceTfExportIncludeFilterResourcesByRegExExclusiveToResource(t *testing.T) {
	testSetup(t)
	var (
		exportTestDir       = testrunner.GetTestTempPath(".terraformInclude" + uuid.NewString())
		exportResourceLabel = "test-export4"
		uniquePostfix       = randString(7)

		queueResources = []QueueExport{
			{OriginalResourceLabel: "test-queue-prod", ExportedLabel: "test-queue-" + uuid.NewString() + "-prod-" + uniquePostfix, Description: "This is the prod queue", AcwTimeoutMs: 200000},
			{OriginalResourceLabel: "test-queue-test", ExportedLabel: "test-queue-" + uuid.NewString() + "-test-" + uniquePostfix, Description: "This is the test queue", AcwTimeoutMs: 200000},
		}

		wrapupCodeResources = []WrapupcodeExport{
			{OriginalResourceLabel: "test-wrapupcode-prod", Name: "test-wrapupcode-" + uuid.NewString() + "-prod-" + uniquePostfix},
			{OriginalResourceLabel: "test-wrapupcode-test", Name: "test-wrapupcode-" + uuid.NewString() + "-test-" + uniquePostfix},
		}
		divResourceLabel = "test-division"
		description      = "Terraform wrapup code description"
		divName          = "terraform-" + uuid.NewString()
	)
	defer os.RemoveAll(exportTestDir)

	queueResourceDef := buildQueueResources(queueResources)
	wrapupcodeResourceDef := buildWrapupcodeResources(wrapupCodeResources, "genesyscloud_auth_division."+divResourceLabel+".id", description)
	baseConfig := queueResourceDef + authDivision.GenerateAuthDivisionBasic(divResourceLabel, divName) + wrapupcodeResourceDef
	configWithExporter := baseConfig + generateTfExportByIncludeFilterResources(
		exportResourceLabel,
		exportTestDir,
		util.TrueValue,
		[]string{
			strconv.Quote("genesyscloud_routing_queue::.*-prod"),
			strconv.Quote("genesyscloud_routing_wrapupcode"),
		},
		strconv.Quote("json"),
		util.FalseValue,
		[]string{
			strconv.Quote("genesyscloud_routing_queue." + queueResources[0].OriginalResourceLabel),
			strconv.Quote("genesyscloud_routing_queue." + queueResources[1].OriginalResourceLabel),
			strconv.Quote("genesyscloud_routing_wrapupcode." + wrapupCodeResources[0].OriginalResourceLabel),
			strconv.Quote("genesyscloud_routing_wrapupcode." + wrapupCodeResources[1].OriginalResourceLabel),
		},
	)

	sanitizer := resourceExporter.NewSanitizerProvider()

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Generate a queue
				Config: baseConfig,
			},
			{
				// Export the queue
				Config: configWithExporter,
				Check: resource.ComposeTestCheckFunc(
					testQueueExportEqual(exportTestDir+"/"+defaultTfJSONFile, "genesyscloud_routing_queue", sanitizer.S.SanitizeResourceBlockLabel(queueResources[0].ExportedLabel), queueResources[0]),
					testWrapupcodeExportEqual(exportTestDir+"/"+defaultTfJSONFile, "genesyscloud_routing_wrapupcode", sanitizer.S.SanitizeResourceBlockLabel(wrapupCodeResources[0].Name), wrapupCodeResources[0]),
					testWrapupcodeExportEqual(exportTestDir+"/"+defaultTfJSONFile, "genesyscloud_routing_wrapupcode", sanitizer.S.SanitizeResourceBlockLabel(wrapupCodeResources[1].Name), wrapupCodeResources[1]),
					testQueueExportMatchesRegEx(exportTestDir+"/"+defaultTfJSONFile, "genesyscloud_routing_queue", ".*-prod"), //We should not find any "test" queues here because we only wanted to include queues that ended with a -prod
				),
			},
		},
		CheckDestroy: testVerifyExportsDestroyedFunc(exportTestDir),
	})
}

// TestAccResourceTfExportExcludeFilterResourcesByRegExExclusiveToResource will exclude any test resources that match a
// regular expression provided for the resource. In this test we expect queues ending with -test or -dev to be excluded.
// Wrap up codes will be created with -prod, -dev and -test suffixes but they should not be affected by the regex filter
// for the queue and should all be exported
func TestAccResourceTfExportExcludeFilterResourcesByRegExExclusiveToResource(t *testing.T) {
	testSetup(t)
	var (
		exportTestDir       = testrunner.GetTestTempPath(".terraformExclude" + uuid.NewString())
		exportResourceLabel = "test-export6"
		uniquePostfix       = randString(7)
		queueResources      = []QueueExport{
			{OriginalResourceLabel: "test-queue-prod", ExportedLabel: "test-queue-" + uuid.NewString() + "-prod-" + uniquePostfix, Description: "This is a test prod queue", AcwTimeoutMs: 200000},
			{OriginalResourceLabel: "test-queue-test", ExportedLabel: "test-queue-" + uuid.NewString() + "-test-" + uniquePostfix, Description: "This is a test queue", AcwTimeoutMs: 200000},
			{OriginalResourceLabel: "test-queue-dev", ExportedLabel: "test-queue-" + uuid.NewString() + "-dev-" + uniquePostfix, Description: "This is a dev queue", AcwTimeoutMs: 200000},
		}

		wrapupCodeResources = []WrapupcodeExport{
			{OriginalResourceLabel: "test-wrapupcode-prod", Name: "test-wrapupcode-" + uuid.NewString() + "-prod-" + uniquePostfix},
			{OriginalResourceLabel: "test-wrapupcode-test", Name: "test-wrapupcode-" + uuid.NewString() + "-test-" + uniquePostfix},
			{OriginalResourceLabel: "test-wrapupcode-dev", Name: "test-wrapupcode-" + uuid.NewString() + "-dev-" + uniquePostfix},
		}
		divResourceLabel = "test-division"
		divName          = "terraform-" + uuid.NewString()
		description      = "Terraform wrapup code description"
	)
	defer os.RemoveAll(exportTestDir)

	fullListOfResourceTypes := resourceExporter.GetAvailableExporterTypes()
	fullListOfResourceTypes = lists.RemoveStringFromSlice("genesyscloud_routing_wrapupcode", fullListOfResourceTypes)
	fullListOfResourceTypes = lists.RemoveStringFromSlice("genesyscloud_routing_queue", fullListOfResourceTypes)
	fullListOfResourceTypes = lists.Map(fullListOfResourceTypes, func(str string) string {
		return strconv.Quote(str)
	})
	fullListOfResourceTypes = append(fullListOfResourceTypes, strconv.Quote("genesyscloud_routing_queue::.*-(dev|test)"))

	queueResourceDef := buildQueueResources(queueResources)
	wrapupcodeResourceDef := buildWrapupcodeResources(wrapupCodeResources, "genesyscloud_auth_division."+divResourceLabel+".id", description)
	baseConfig := queueResourceDef + authDivision.GenerateAuthDivisionBasic(divResourceLabel, divName) + wrapupcodeResourceDef
	configWithExporter := baseConfig + generateTfExportByExcludeFilterResources(
		exportResourceLabel,
		exportTestDir,
		util.TrueValue,
		fullListOfResourceTypes,
		strconv.Quote("json"),
		util.FalseValue,
		[]string{
			strconv.Quote("genesyscloud_routing_queue." + queueResources[0].OriginalResourceLabel),
			strconv.Quote("genesyscloud_routing_queue." + queueResources[1].OriginalResourceLabel),
			strconv.Quote("genesyscloud_routing_queue." + queueResources[2].OriginalResourceLabel),
			strconv.Quote("genesyscloud_routing_wrapupcode." + wrapupCodeResources[0].OriginalResourceLabel),
			strconv.Quote("genesyscloud_routing_wrapupcode." + wrapupCodeResources[1].OriginalResourceLabel),
			strconv.Quote("genesyscloud_routing_wrapupcode." + wrapupCodeResources[2].OriginalResourceLabel),
		},
	)

	sanitizer := resourceExporter.NewSanitizerProvider()
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Generate a queue with wrapup codes
				Config: baseConfig,
			},
			{
				// Generate a queue with wrapup codes and exporter
				Config: configWithExporter,
				Check: resource.ComposeTestCheckFunc(
					testQueueExportEqual(exportTestDir+"/"+defaultTfJSONFile, "genesyscloud_routing_queue", sanitizer.S.SanitizeResourceBlockLabel(queueResources[0].ExportedLabel), queueResources[0]),
					testWrapupcodeExportEqual(exportTestDir+"/"+defaultTfJSONFile, "genesyscloud_routing_wrapupcode", sanitizer.S.SanitizeResourceBlockLabel(wrapupCodeResources[0].Name), wrapupCodeResources[0]),
					testWrapupcodeExportEqual(exportTestDir+"/"+defaultTfJSONFile, "genesyscloud_routing_wrapupcode", sanitizer.S.SanitizeResourceBlockLabel(wrapupCodeResources[1].Name), wrapupCodeResources[1]),
					testWrapupcodeExportEqual(exportTestDir+"/"+defaultTfJSONFile, "genesyscloud_routing_wrapupcode", sanitizer.S.SanitizeResourceBlockLabel(wrapupCodeResources[2].Name), wrapupCodeResources[2]),
					testQueueExportExcludesRegEx(exportTestDir+"/"+defaultTfJSONFile, "genesyscloud_routing_queue", ".*-(dev|test)"),
				),
			},
		},
		CheckDestroy: testVerifyExportsDestroyedFunc(exportTestDir),
	})
}

// TestAccResourceTfExportSplitFilesAsJSON will create 2 queues, 2 wrap up codes, and 2 users.
// The exporter will be run in split mode so 3 resource tf.jsons should be created as well as a provider.tf.json
func TestAccResourceTfExportSplitFilesAsJSON(t *testing.T) {
	testSetup(t)
	var (
		exportTestDir       = testrunner.GetTestTempPath(".terraform" + uuid.NewString())
		exportResourceLabel = "test-export-split"
		uniquePostfix       = randString(7)
		expectedFilesPath   = []string{
			filepath.Join(exportTestDir, "genesyscloud_routing_queue.tf.json"),
			filepath.Join(exportTestDir, "genesyscloud_user.tf.json"),
			filepath.Join(exportTestDir, "genesyscloud_routing_wrapupcode.tf.json"),
			filepath.Join(exportTestDir, "provider.tf.json"),
		}

		queueResources = []QueueExport{
			{OriginalResourceLabel: "test-queue-1", ExportedLabel: "test-queue-1-" + uuid.NewString() + uniquePostfix, Description: "This is a test queue", AcwTimeoutMs: 200000},
			{OriginalResourceLabel: "test-queue-2", ExportedLabel: "test-queue-1-" + uuid.NewString() + uniquePostfix, Description: "This is a test queue too", AcwTimeoutMs: 200000},
		}

		userResources = []UserExport{
			{OriginalResourceLabel: "test-user-1", ExportedLabel: "test-user-1" + uuid.NewString() + uniquePostfix, Email: "test-user-1" + uuid.NewString() + "@test.com" + uniquePostfix, State: "active"},
			{OriginalResourceLabel: "test-user-2", ExportedLabel: "test-user-2" + uuid.NewString() + uniquePostfix, Email: "test-user-2" + uuid.NewString() + "@test.com" + uniquePostfix, State: "active"},
		}

		wrapupCodeResources = []WrapupcodeExport{
			{OriginalResourceLabel: "test-wrapupcode-1", Name: "test-wrapupcode-1-" + uuid.NewString() + uniquePostfix},
			{OriginalResourceLabel: "test-wrapupcode-2", Name: "test-wrapupcode-2-" + uuid.NewString() + uniquePostfix},
		}

		divResourceLabel = "test-division"
		divName          = "terraform-" + uuid.NewString()
		description      = "Terraform wrapup code description"
	)
	defer os.RemoveAll(exportTestDir)

	queueResourceDef := buildQueueResources(queueResources)
	userResourcesDef := buildUserResources(userResources)
	wrapupcodeResourceDef := buildWrapupcodeResources(wrapupCodeResources, "genesyscloud_auth_division."+divResourceLabel+".id", description)
	baseConfig := queueResourceDef + authDivision.GenerateAuthDivisionBasic(divResourceLabel, divName) + wrapupcodeResourceDef + userResourcesDef
	configWithExporter := baseConfig + generateTfExportByIncludeFilterResources(
		exportResourceLabel,
		exportTestDir,
		util.TrueValue,
		[]string{
			strconv.Quote("genesyscloud_routing_queue::" + uniquePostfix + "$"),
			strconv.Quote("genesyscloud_user::" + uniquePostfix + "$"),
			strconv.Quote("genesyscloud_routing_wrapupcode::" + uniquePostfix + "$"),
		},
		strconv.Quote("json"),
		util.TrueValue,
		[]string{
			strconv.Quote("genesyscloud_routing_queue." + queueResources[0].OriginalResourceLabel),
			strconv.Quote("genesyscloud_routing_queue." + queueResources[1].OriginalResourceLabel),
			strconv.Quote("genesyscloud_user." + userResources[0].OriginalResourceLabel),
			strconv.Quote("genesyscloud_user." + userResources[1].OriginalResourceLabel),
			strconv.Quote("genesyscloud_routing_wrapupcode." + wrapupCodeResources[0].OriginalResourceLabel),
			strconv.Quote("genesyscloud_routing_wrapupcode." + wrapupCodeResources[1].OriginalResourceLabel),
		},
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: baseConfig,
			},
			{
				Config: configWithExporter,
				Check: resource.ComposeTestCheckFunc(
					validateFileCreated(expectedFilesPath[0]),
					validateFileCreated(expectedFilesPath[1]),
					validateFileCreated(expectedFilesPath[2]),
					validateFileCreated(expectedFilesPath[3]),
				),
			},
		},
		CheckDestroy: testVerifyExportsDestroyedFunc(exportTestDir),
	})
}

// TestAccResourceTfExportExcludeFilterResourcesByRegExExclusiveToResourceAndSanitizedLabels will exclude any test resources that match a
// regular expression provided for the resource. In this test we check against both sanitized and unsanitized labels.
func TestAccResourceTfExportExcludeFilterResourcesByRegExExclusiveToResourceAndSanitizedLabels(t *testing.T) {
	testSetup(t)
	var (
		exportTestDir       = testrunner.GetTestTempPath(".terraformExclude" + uuid.NewString())
		exportResourceLabel = "test-export6_1"
		uniquePostfix       = randString(7)

		queueResources = []QueueExport{
			{OriginalResourceLabel: "test-queue-test-1", ExportedLabel: "exclude filter - exclude me" + uuid.NewString() + uniquePostfix, Description: "This is an excluded bar test resource", AcwTimeoutMs: 200000},
			{OriginalResourceLabel: "test-queue-test-2", ExportedLabel: "exclude filter - foo - bar me" + uuid.NewString() + uniquePostfix, Description: "This is a foo bar test resource", AcwTimeoutMs: 200000},
			{OriginalResourceLabel: "test-queue-test-3", ExportedLabel: "exclude filter - fu - barre you" + uuid.NewString() + uniquePostfix, Description: "This is a foo bar test resource", AcwTimeoutMs: 200000},
		}

		wrapupCodeResources = []WrapupcodeExport{
			{OriginalResourceLabel: "test-wrapupcode-prod", Name: "exclude me" + uuid.NewString() + uniquePostfix},
			{OriginalResourceLabel: "test-wrapupcode-test", Name: "foo + bar me" + uuid.NewString() + uniquePostfix},
			{OriginalResourceLabel: "test-wrapupcode-dev", Name: "fu - barre you" + uuid.NewString() + uniquePostfix},
		}
		divResourceLabel = "test-division"
		divName          = "terraform-" + uuid.NewString()
		description      = "Terraform wrapup code description"
	)
	cleanupFunc := func() {
		if provider.SdkClientPool != nil {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			_ = provider.SdkClientPool.Close(ctx)
		}
		if err := os.RemoveAll(exportTestDir); err != nil {
			t.Logf("Error while cleaning up %v", err)
		}
	}
	t.Cleanup(cleanupFunc)

	fullListOfResourceTypes := resourceExporter.GetAvailableExporterTypes()
	fullListOfResourceTypes = lists.RemoveStringFromSlice("genesyscloud_routing_wrapupcode", fullListOfResourceTypes)
	fullListOfResourceTypes = lists.RemoveStringFromSlice("genesyscloud_routing_queue", fullListOfResourceTypes)
	fullListOfResourceTypes = lists.Map(fullListOfResourceTypes, func(str string) string {
		return strconv.Quote(str)
	})
	fullListOfResourceTypes = append(fullListOfResourceTypes, strconv.Quote("genesyscloud_routing_queue::exclude filter - foo - bar me"))
	fullListOfResourceTypes = append(fullListOfResourceTypes, strconv.Quote("genesyscloud_routing_queue::exclude_filter_-_fu_-_barre_you"))

	queueResourceDef := buildQueueResources(queueResources)
	wrapupcodeResourceDef := buildWrapupcodeResources(wrapupCodeResources, "genesyscloud_auth_division."+divResourceLabel+".id", description)
	baseConfig := queueResourceDef + authDivision.GenerateAuthDivisionBasic(divResourceLabel, divName) + wrapupcodeResourceDef
	configWithExporter := baseConfig + generateTfExportByExcludeFilterResources(
		exportResourceLabel,
		exportTestDir,
		util.TrueValue,
		fullListOfResourceTypes,
		strconv.Quote("json"),
		util.FalseValue,
		[]string{
			strconv.Quote("genesyscloud_routing_queue." + queueResources[0].OriginalResourceLabel),
			strconv.Quote("genesyscloud_routing_queue." + queueResources[1].OriginalResourceLabel),
			strconv.Quote("genesyscloud_routing_queue." + queueResources[2].OriginalResourceLabel),
			strconv.Quote("genesyscloud_routing_wrapupcode." + wrapupCodeResources[0].OriginalResourceLabel),
			strconv.Quote("genesyscloud_routing_wrapupcode." + wrapupCodeResources[1].OriginalResourceLabel),
			strconv.Quote("genesyscloud_routing_wrapupcode." + wrapupCodeResources[2].OriginalResourceLabel),
		},
	)

	sanitizer := resourceExporter.NewSanitizerProvider()
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Generate a queue
				Config: baseConfig,
			},
			{
				// Now export it
				Config: configWithExporter,
				Check: resource.ComposeTestCheckFunc(
					testQueueExportEqual(exportTestDir+"/"+defaultTfJSONFile, "genesyscloud_routing_queue", sanitizer.S.SanitizeResourceBlockLabel(queueResources[0].ExportedLabel), queueResources[0]),
					testWrapupcodeExportEqual(exportTestDir+"/"+defaultTfJSONFile, "genesyscloud_routing_wrapupcode", sanitizer.S.SanitizeResourceBlockLabel(wrapupCodeResources[0].Name), wrapupCodeResources[0]),
					testWrapupcodeExportEqual(exportTestDir+"/"+defaultTfJSONFile, "genesyscloud_routing_wrapupcode", sanitizer.S.SanitizeResourceBlockLabel(wrapupCodeResources[1].Name), wrapupCodeResources[1]),
					testWrapupcodeExportEqual(exportTestDir+"/"+defaultTfJSONFile, "genesyscloud_routing_wrapupcode", sanitizer.S.SanitizeResourceBlockLabel(wrapupCodeResources[2].Name), wrapupCodeResources[2]),
					testQueueExportExcludesRegEx(exportTestDir+"/"+defaultTfJSONFile, "genesyscloud_routing_queue", "exclude[ _]filter[ _]-[ _](foo|fu).*"),
				),
			},
		},
		CheckDestroy: testVerifyExportsDestroyedFunc(exportTestDir),
	})
}

// TestAccResourceTfExportForCompress does a basic test check to make sure the compressed file is created.
func TestAccResourceTfExportForCompress(t *testing.T) {
	testSetup(t)
	var (
		exportTestDir        = testrunner.GetTestTempPath(".terraform" + uuid.NewString())
		exportResourceLabel1 = "test-export1"
		zipFileName          = filepath.Join(exportTestDir, "..", "archive_genesyscloud_tf_export*")
		divResourceLabel     = "test-division"
		divName              = "terraform-" + uuid.NewString()
	)

	baseConfig := authDivision.GenerateAuthDivisionBasic(divResourceLabel, divName)
	defer os.RemoveAll(exportTestDir)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			// Create basic object
			{
				Config: baseConfig,
			},
			{
				// Run export without state file
				Config: baseConfig + generateTfExportResourceForCompress(
					exportResourceLabel1,
					exportTestDir,
					util.TrueValue,
					util.TrueValue,
					[]string{
						strconv.Quote(authDivision.ResourceType),
					},
					"",
				),
				Check: resource.ComposeTestCheckFunc(
					validateCompressedCreated(zipFileName),
					validateCompressedFile(zipFileName),
				),
			},
		},
		CheckDestroy: deleteTestCompressedZip(exportTestDir, zipFileName),
	})
}

// TestAccResourceTfExport does a basic test check to make sure the export file is created.
func TestAccResourceTfExport(t *testing.T) {
	testSetup(t)
	var (
		exportTestDir        = testrunner.GetTestTempPath(".terraform" + uuid.NewString())
		exportResourceLabel1 = "test-export1"
		configPath           = filepath.Join(exportTestDir, defaultTfJSONFile)
		statePath            = filepath.Join(exportTestDir, defaultTfStateFile)
	)

	defer os.RemoveAll(exportTestDir)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Run export without state file
				Config: generateTfExportResource(
					exportResourceLabel1,
					exportTestDir,
					util.TrueValue,
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
					exportResourceLabel1,
					exportTestDir,
					util.TrueValue,
					strconv.Quote("genesyscloud_auth_role.permission_policies.conditions"),
				),
				Check: resource.ComposeTestCheckFunc(
					validateFileCreated(configPath),
					validateConfigFile(configPath),
					validateFileCreated(statePath),
				),
			},
			{
				// Run export with state file and excluded attribute with regex
				Config: generateTfExportResourceMin(
					exportResourceLabel1,
					exportTestDir,
					util.TrueValue,
					strconv.Quote("g*.name"),
				),
				Check: resource.ComposeTestCheckFunc(
					validateFileCreated(configPath),
					validateConfigFile(configPath),
					validateFileCreated(statePath),
				),
			},
		},
		CheckDestroy: testVerifyExportsDestroyedFunc(exportTestDir),
	})
}

func TestAccResourceTfExportByLabel(t *testing.T) {
	testSetup(t)
	var (
		exportTestDir        = testrunner.GetTestTempPath(".terraform" + uuid.NewString())
		exportResourceLabel1 = "test-export1"

		userResourceLabel1 = "test-user1"
		userEmail1         = "terraform-" + uuid.NewString() + "@example.com"
		userName1          = "John Data-" + uuid.NewString()

		userResourceLabel2 = "test-user2"
		userEmail2         = "terraform-" + uuid.NewString() + "@example.com"
		userName2          = "John Data-" + uuid.NewString()

		queueResourceLabel = "test-queue"
		queueName          = "Terraform Test Queue-" + uuid.NewString()
		queueDesc          = "This is a test"
		queueAcwTimeout    = 200000
	)

	defer os.RemoveAll(exportTestDir)

	testUser1 := &UserExport{
		ExportedLabel: userName1,
		Email:         userEmail1,
		State:         "active",
	}

	testUser2 := &UserExport{
		ExportedLabel: userName2,
		Email:         userEmail2,
		State:         "active",
	}

	testQueue := &QueueExport{
		ExportedLabel: queueName,
		Description:   queueDesc,
		AcwTimeoutMs:  queueAcwTimeout,
	}

	sanitizer := resourceExporter.NewSanitizerProvider()

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Generate a user
				Config: user.GenerateBasicUserResource(
					userResourceLabel1,
					userEmail1,
					userName1,
				),
			},
			{
				// Now export it
				Config: user.GenerateBasicUserResource(
					userResourceLabel1,
					userEmail1,
					userName1,
				) + generateTfExportByFilter(
					exportResourceLabel1,
					exportTestDir,
					util.TrueValue,
					[]string{strconv.Quote("genesyscloud_user::" + userEmail1)},
					"",
					strconv.Quote("json"),
					util.FalseValue,
					[]string{strconv.Quote("genesyscloud_user." + userResourceLabel1)},
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_tf_export."+exportResourceLabel1,
						"resource_types.0", "genesyscloud_user::"+userEmail1),
					testUserExport(exportTestDir+"/"+defaultTfJSONFile, "genesyscloud_user", sanitizer.S.SanitizeResourceBlockLabel(userEmail1), testUser1),
				),
			},
			{
				// Generate a queue as well and export it
				Config: user.GenerateBasicUserResource(
					userResourceLabel1,
					userEmail1,
					userName1,
				) + routingQueue.GenerateRoutingQueueResource(
					queueResourceLabel,
					queueName,
					queueDesc,
					util.NullValue,                     // MANDATORY_TIMEOUT
					fmt.Sprintf("%v", queueAcwTimeout), // acw_timeout
					util.NullValue,                     // ALL
					util.NullValue,                     // auto_answer_only true
					util.NullValue,                     // No calling party name
					util.NullValue,                     // No calling party number
					util.NullValue,                     // enable_audio_monitoring false
					util.NullValue,                     // enable_manual_assignment false
					util.FalseValue,                    // suppressCall_record_false
					util.NullValue,                     // enable_transcription false
					strconv.Quote("TimestampAndPriority"),
					util.NullValue,
					util.NullValue,
					util.NullValue,
				) + generateTfExportByFilter(
					exportResourceLabel1,
					exportTestDir,
					util.TrueValue,
					[]string{
						strconv.Quote("genesyscloud_user::" + userEmail1),
						strconv.Quote("genesyscloud_routing_queue::" + queueName),
					},
					"",
					strconv.Quote("json"),
					util.FalseValue,
					[]string{strconv.Quote("genesyscloud_user." + userResourceLabel1)},
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_tf_export."+exportResourceLabel1,
						"resource_types.0", "genesyscloud_user::"+userEmail1),
					resource.TestCheckResourceAttr("genesyscloud_tf_export."+exportResourceLabel1,
						"resource_types.1", "genesyscloud_routing_queue::"+queueName),
					testUserExport(exportTestDir+"/"+defaultTfJSONFile, "genesyscloud_user", sanitizer.S.SanitizeResourceBlockLabel(userEmail1), testUser1),
					testQueueExportEqual(exportTestDir+"/"+defaultTfJSONFile, "genesyscloud_routing_queue", sanitizer.S.SanitizeResourceBlockLabel(queueName), *testQueue),
				),
			},
			{
				// Export all trunk base settings as well
				Config: user.GenerateBasicUserResource(
					userResourceLabel1,
					userEmail1,
					userName1,
				) + routingQueue.GenerateRoutingQueueResource(
					queueResourceLabel,
					queueName,
					queueDesc,
					util.NullValue,                     // MANDATORY_TIMEOUT
					fmt.Sprintf("%v", queueAcwTimeout), // acw_timeout
					util.NullValue,                     // ALL
					util.NullValue,                     // auto_answer_only true
					util.NullValue,                     // No calling party name
					util.NullValue,                     // No calling party number
					util.NullValue,                     // enable_audio_monitoring false
					util.NullValue,                     // enable_manual_assignment false
					util.FalseValue,                    // suppressCall_record_false
					util.NullValue,                     // enable_transcription false
					strconv.Quote("TimestampAndPriority"),
					util.NullValue,
					util.NullValue,
					util.NullValue,
				) + generateTfExportByFilter(
					exportResourceLabel1,
					exportTestDir,
					util.TrueValue,
					[]string{
						strconv.Quote("genesyscloud_user::" + userEmail1),
						strconv.Quote("genesyscloud_routing_queue::" + queueName),
						strconv.Quote("genesyscloud_telephony_providers_edges_trunkbasesettings"),
					},
					"",
					strconv.Quote("json"),
					util.FalseValue,
					[]string{
						strconv.Quote("genesyscloud_routing_queue." + queueResourceLabel),
						strconv.Quote("genesyscloud_user." + userResourceLabel1)},
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"genesyscloud_tf_export."+exportResourceLabel1, "resource_types.0",
						"genesyscloud_user::"+userEmail1),
					resource.TestCheckResourceAttr(
						"genesyscloud_tf_export."+exportResourceLabel1, "resource_types.1",
						"genesyscloud_routing_queue::"+queueName),
					resource.TestCheckResourceAttr(
						"genesyscloud_tf_export."+exportResourceLabel1, "resource_types.2",
						"genesyscloud_telephony_providers_edges_trunkbasesettings"),
					testUserExport(exportTestDir+"/"+defaultTfJSONFile, "genesyscloud_user", sanitizer.S.SanitizeResourceBlockLabel(userEmail1), testUser1),
					testQueueExportEqual(exportTestDir+"/"+defaultTfJSONFile, "genesyscloud_routing_queue", sanitizer.S.SanitizeResourceBlockLabel(queueName), *testQueue),
					testTrunkBaseSettingsExport(exportTestDir+"/"+defaultTfJSONFile, "genesyscloud_telephony_providers_edges_trunkbasesettings"),
				),
			},
			{
				// Export all trunk base settings as well
				Config: user.GenerateBasicUserResource(
					userResourceLabel1,
					userEmail1,
					userName1,
				) + user.GenerateBasicUserResource(
					userResourceLabel2,
					userEmail2,
					userName2,
				) + routingQueue.GenerateRoutingQueueResource(
					queueResourceLabel,
					queueName,
					queueDesc,
					util.NullValue,                     // MANDATORY_TIMEOUT
					fmt.Sprintf("%v", queueAcwTimeout), // acw_timeout
					util.NullValue,                     // ALL
					util.NullValue,                     // auto_answer_only true
					util.NullValue,                     // No calling party name
					util.NullValue,                     // No calling party number
					util.NullValue,                     // enable_audio_monitoring false
					util.NullValue,                     // enable_manual_assignment false
					util.FalseValue,                    // suppressCall_record_false
					util.NullValue,                     // enable_transcription false
					strconv.Quote("TimestampAndPriority"),
					util.NullValue,
					util.NullValue,
					util.NullValue,
				) + generateTfExportByFilter(
					exportResourceLabel1,
					exportTestDir,
					util.TrueValue,
					[]string{
						strconv.Quote("genesyscloud_user::" + userEmail1),
						strconv.Quote("genesyscloud_user::" + userEmail2),
						strconv.Quote("genesyscloud_routing_queue::" + queueName),
						strconv.Quote("genesyscloud_telephony_providers_edges_trunkbasesettings"),
					},
					"",
					strconv.Quote("json"),
					util.FalseValue,
					[]string{
						strconv.Quote("genesyscloud_routing_queue." + queueResourceLabel),
						strconv.Quote("genesyscloud_user." + userResourceLabel1),
						strconv.Quote("genesyscloud_user." + userResourceLabel2)},
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"genesyscloud_tf_export."+exportResourceLabel1, "resource_types.0",
						"genesyscloud_user::"+userEmail1),
					resource.TestCheckResourceAttr(
						"genesyscloud_tf_export."+exportResourceLabel1, "resource_types.1",
						"genesyscloud_user::"+userEmail2),
					resource.TestCheckResourceAttr(
						"genesyscloud_tf_export."+exportResourceLabel1, "resource_types.2",
						"genesyscloud_routing_queue::"+queueName),
					resource.TestCheckResourceAttr(
						"genesyscloud_tf_export."+exportResourceLabel1, "resource_types.3",
						"genesyscloud_telephony_providers_edges_trunkbasesettings"),
					testUserExport(exportTestDir+"/"+defaultTfJSONFile, "genesyscloud_user", sanitizer.S.SanitizeResourceBlockLabel(userEmail1), testUser1),
					testUserExport(exportTestDir+"/"+defaultTfJSONFile, "genesyscloud_user", sanitizer.S.SanitizeResourceBlockLabel(userEmail2), testUser2),
					testQueueExportEqual(exportTestDir+"/"+defaultTfJSONFile, "genesyscloud_routing_queue", sanitizer.S.SanitizeResourceBlockLabel(queueName), *testQueue),
					testTrunkBaseSettingsExport(exportTestDir+"/"+defaultTfJSONFile, "genesyscloud_telephony_providers_edges_trunkbasesettings"),
				),
			},
		},
		CheckDestroy: testVerifyExportsDestroyedFunc(exportTestDir),
	})
}

func TestAccResourceTfExportIncludeFilterResourcesByType(t *testing.T) {
	testSetup(t)
	var (
		exportTestDir       = testrunner.GetTestTempPath(".terraform" + uuid.NewString())
		exportResourceLabel = "test-export2"
		uniquePostfix       = randString(7)
	)

	queueResources := []QueueExport{
		{OriginalResourceLabel: "test-queue-prod-1", ExportedLabel: "test-queue-" + uuid.NewString() + "-prod" + uniquePostfix, Description: "This is a test prod queue 1", AcwTimeoutMs: 200000},
		{OriginalResourceLabel: "test-queue-prod-2", ExportedLabel: "test-queue-" + uuid.NewString() + "-prod" + uniquePostfix, Description: "This is a test prod queue 2", AcwTimeoutMs: 200000},
		{OriginalResourceLabel: "test-queue-prod-3", ExportedLabel: "test-queue-" + uuid.NewString() + "-prod" + uniquePostfix, Description: "This is a test prod queue 3", AcwTimeoutMs: 200000},
		{OriginalResourceLabel: "test-queue-test-1", ExportedLabel: "test-queue-" + uuid.NewString() + "-test" + uniquePostfix, Description: "This is a test prod queue 4", AcwTimeoutMs: 200000},
	}

	defer os.RemoveAll(exportTestDir)

	queueResourceDef := buildQueueResources(queueResources)
	baseConfig := queueResourceDef
	configWithExporter := baseConfig + generateTfExportByIncludeFilterResources(
		exportResourceLabel,
		exportTestDir,
		util.TrueValue,
		[]string{
			strconv.Quote("genesyscloud_routing_queue"),
		},
		strconv.Quote("json"),
		util.FalseValue,
		[]string{
			strconv.Quote("genesyscloud_routing_queue." + queueResources[0].OriginalResourceLabel),
			strconv.Quote("genesyscloud_routing_queue." + queueResources[1].OriginalResourceLabel),
			strconv.Quote("genesyscloud_routing_queue." + queueResources[2].OriginalResourceLabel),
			strconv.Quote("genesyscloud_routing_queue." + queueResources[3].OriginalResourceLabel),
		},
	)

	sanitizer := resourceExporter.NewSanitizerProvider()
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Generate a queues
				Config: baseConfig,
			},
			{
				// Now export them
				Config: configWithExporter,
				Check: resource.ComposeTestCheckFunc(
					testQueueExportEqual(exportTestDir+"/"+defaultTfJSONFile, "genesyscloud_routing_queue", sanitizer.S.SanitizeResourceBlockLabel(queueResources[0].ExportedLabel), queueResources[0]),
					testQueueExportEqual(exportTestDir+"/"+defaultTfJSONFile, "genesyscloud_routing_queue", sanitizer.S.SanitizeResourceBlockLabel(queueResources[1].ExportedLabel), queueResources[1]),
					testQueueExportEqual(exportTestDir+"/"+defaultTfJSONFile, "genesyscloud_routing_queue", sanitizer.S.SanitizeResourceBlockLabel(queueResources[2].ExportedLabel), queueResources[2]),
					testQueueExportEqual(exportTestDir+"/"+defaultTfJSONFile, "genesyscloud_routing_queue", sanitizer.S.SanitizeResourceBlockLabel(queueResources[3].ExportedLabel), queueResources[3]),
				),
			},
		},
		CheckDestroy: testVerifyExportsDestroyedFunc(exportTestDir),
	})
}

// TestAccResourceTfExportExcludeFilterResourcesByRegEx will exclude any test resources that match a regular expression provided.  In our test case we exclude
// all routing queues that have a regex with -(dev|test)$ in it.  We then check to see if there are any prod queues present.
func TestAccResourceTfExportExcludeFilterResourcesByRegEx(t *testing.T) {
	testSetup(t)
	var (
		exportTestDir       = testrunner.GetTestTempPath(".terraform" + uuid.NewString())
		exportResourceLabel = "test-export5"
		uniquePostfix       = randString(7)

		queueResources = []QueueExport{
			{OriginalResourceLabel: "test-queue-prod-1", ExportedLabel: "test-queue-" + uuid.NewString() + "-prod-" + uniquePostfix, Description: "This is a test prod queue 1", AcwTimeoutMs: 200000},
			{OriginalResourceLabel: "test-queue-prod-2", ExportedLabel: "test-queue-" + uuid.NewString() + "-prod-" + uniquePostfix, Description: "This is a test prod queue 2", AcwTimeoutMs: 200000},
			{OriginalResourceLabel: "test-queue-prod-3", ExportedLabel: "test-queue-" + uuid.NewString() + "-prod-" + uniquePostfix, Description: "This is a test prod queue 3", AcwTimeoutMs: 200000},
			{OriginalResourceLabel: "test-queue-test-4", ExportedLabel: "test-queue-" + uuid.NewString() + "-test-" + uniquePostfix, Description: "This is a test queue 4", AcwTimeoutMs: 200000},
			{OriginalResourceLabel: "test-queue-dev-1", ExportedLabel: "test-queue-" + uuid.NewString() + "-dev-" + uniquePostfix, Description: "This is a dev queue 5", AcwTimeoutMs: 200000},
		}
	)
	defer os.RemoveAll(exportTestDir)

	queueResourceDef := buildQueueResources(queueResources)
	baseConfig := queueResourceDef
	configWithExporter := baseConfig + generateTfExportByExcludeFilterResources(
		exportResourceLabel,
		exportTestDir,
		util.TrueValue,
		[]string{
			strconv.Quote("genesyscloud_routing_queue::.*-(dev|test)"),
			strconv.Quote("genesyscloud_outbound_ruleset"),
			strconv.Quote("genesyscloud_user"),
			strconv.Quote("genesyscloud_user_roles"),
			strconv.Quote("genesyscloud_flow"),
		},
		strconv.Quote("json"),
		util.FalseValue,
		[]string{
			strconv.Quote("genesyscloud_routing_queue." + queueResources[0].OriginalResourceLabel),
			strconv.Quote("genesyscloud_routing_queue." + queueResources[1].OriginalResourceLabel),
			strconv.Quote("genesyscloud_routing_queue." + queueResources[2].OriginalResourceLabel),
			strconv.Quote("genesyscloud_routing_queue." + queueResources[3].OriginalResourceLabel),
			strconv.Quote("genesyscloud_routing_queue." + queueResources[4].OriginalResourceLabel),
		},
	)

	sanitizer := resourceExporter.NewSanitizerProvider()

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Generate a queue
				Config: baseConfig,
			},
			{
				// Now export them all
				Config: configWithExporter,
				Check: resource.ComposeTestCheckFunc(
					testQueueExportEqual(exportTestDir+"/"+defaultTfJSONFile, "genesyscloud_routing_queue", sanitizer.S.SanitizeResourceBlockLabel(queueResources[0].ExportedLabel), queueResources[0]), //Want to make sure the prod queues are queue is there
					testQueueExportEqual(exportTestDir+"/"+defaultTfJSONFile, "genesyscloud_routing_queue", sanitizer.S.SanitizeResourceBlockLabel(queueResources[1].ExportedLabel), queueResources[1]), //Want to make sure the prod queues are queue is there
					testQueueExportEqual(exportTestDir+"/"+defaultTfJSONFile, "genesyscloud_routing_queue", sanitizer.S.SanitizeResourceBlockLabel(queueResources[2].ExportedLabel), queueResources[2]), //Want to make sure the prod queues are queue is there
					testQueueExportExcludesRegEx(exportTestDir+"/"+defaultTfJSONFile, "genesyscloud_routing_queue", ".*-(dev|test)"),                                                                    //We should not find any "dev or test" queues here because we only wanted to include queues that ended with a -prod
				),
			},
		},
		CheckDestroy: testVerifyExportsDestroyedFunc(exportTestDir),
	})
}

func TestAccResourceTfExportFormAsHCL(t *testing.T) {
	testSetup(t)
	var (
		exportTestDir     = testrunner.GetTestTempPath(".terraform" + uuid.NewString())
		exportedContents  string
		pathToHclFile     = filepath.Join(exportTestDir, defaultTfHCLFile)
		formName          = "terraform_form_evaluations_" + uuid.NewString()
		formResourceLabel = formName

		// Complete evaluation form
		evaluationForm1 = qualityFormsEvaluation.EvaluationFormStruct{
			Name:      formName,
			Published: false,

			QuestionGroups: []qualityFormsEvaluation.EvaluationFormQuestionGroupStruct{
				{
					Name:                    "Test Question Group 1",
					DefaultAnswersToHighest: true,
					DefaultAnswersToNA:      true,
					NaEnabled:               true,
					Weight:                  1,
					ManualWeight:            true,
					Questions: []qualityFormsEvaluation.EvaluationFormQuestionStruct{
						{
							Text: "Did the agent perform the opening spiel?",
							AnswerOptions: []qualityFormsEvaluation.AnswerOptionStruct{
								{
									Text:  "Yes",
									Value: 1,
								},
								{
									Text:  "No",
									Value: 0,
								},
							},
						},
						{
							Text:             "Did the agent greet the customer?",
							HelpText:         "Help text here",
							NaEnabled:        true,
							CommentsRequired: true,
							IsKill:           true,
							IsCritical:       true,
							VisibilityCondition: qualityFormsEvaluation.VisibilityConditionStruct{
								CombiningOperation: "AND",
								Predicates:         []string{"/form/questionGroup/0/question/0/answer/0", "/form/questionGroup/0/question/0/answer/1"},
							},
							AnswerOptions: []qualityFormsEvaluation.AnswerOptionStruct{
								{
									Text:  "Yes",
									Value: 1,
								},
								{
									Text:  "No",
									Value: 0,
								},
							},
						},
					},
				},
				{
					Name:   "Test Question Group 2",
					Weight: 2,
					Questions: []qualityFormsEvaluation.EvaluationFormQuestionStruct{
						{
							Text: "Did the agent offer to sell product?",
							AnswerOptions: []qualityFormsEvaluation.AnswerOptionStruct{
								{
									Text:  "Yes",
									Value: 1,
								},
								{
									Text:  "No",
									Value: 0,
								},
							},
						},
					},
					VisibilityCondition: qualityFormsEvaluation.VisibilityConditionStruct{
						CombiningOperation: "AND",
						Predicates:         []string{"/form/questionGroup/0/question/0/answer/1"},
					},
				},
			},
		}
	)

	defer os.RemoveAll(exportTestDir)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: qualityFormsEvaluation.GenerateEvaluationFormResource(formResourceLabel, &evaluationForm1),
				Check: resource.ComposeTestCheckFunc(
					validateEvaluationFormAttributes(formResourceLabel, evaluationForm1),
				),
			},
			{
				Config: qualityFormsEvaluation.GenerateEvaluationFormResource(formResourceLabel, &evaluationForm1) + generateTfExportByFilter(
					formResourceLabel,
					exportTestDir,
					util.TrueValue,
					[]string{strconv.Quote("genesyscloud_quality_forms_evaluation::" + formName)},
					"",
					strconv.Quote("hcl"),
					util.FalseValue,
					[]string{
						strconv.Quote("genesyscloud_quality_forms_evaluation." + formResourceLabel),
					},
				),
				Check: resource.ComposeTestCheckFunc(
					getExportedFileContents(pathToHclFile, &exportedContents),
				),
			},
		},
		CheckDestroy: testVerifyExportsDestroyedFunc(exportTestDir),
	})

	exportedContents = removeTerraformProviderBlock(exportedContents)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: exportedContents,
				Check: resource.ComposeTestCheckFunc(
					validateEvaluationFormAttributes(formResourceLabel, evaluationForm1),
				),
			},
		},
		CheckDestroy: testVerifyExportsDestroyedFunc(exportTestDir),
	})
}

func TestAccResourceTfExportQueueAsHCL(t *testing.T) {
	testSetup(t)
	var (
		exportTestDir  = testrunner.GetTestTempPath(".terraform" + uuid.NewString())
		exportContents string
		pathToHclFile  = filepath.Join(exportTestDir, defaultTfHCLFile)
	)

	defer os.RemoveAll(exportTestDir)

	// routing queue attributes
	var (
		queueName      = fmt.Sprintf("queue_%v", uuid.NewString())
		queueLabel     = queueName
		description    = "This is a test queue"
		autoAnswerOnly = "true"

		alertTimeoutSec = "30"
		slPercentage    = "0.7"
		slDurationMs    = "10000"

		rrOperator    = "MEETS_THRESHOLD"
		rrThreshold   = "9"
		rrWaitSeconds = "300"

		chatScriptID  = uuid.NewString()
		emailScriptID = uuid.NewString()
	)

	routingQueue := routingQueue.GenerateRoutingQueueResource(
		queueLabel,
		queueName,
		description,
		strconv.Quote("MANDATORY_TIMEOUT"),
		"300000",
		strconv.Quote("BEST"),
		autoAnswerOnly,
		strconv.Quote("Example Inc."),
		util.NullValue,
		"true",
		"true",
		util.TrueValue,
		util.FalseValue,
		strconv.Quote("TimestampAndPriority"),
		util.NullValue,
		util.NullValue,
		util.NullValue,
		routingQueue.GenerateMediaSettings("media_settings_call", alertTimeoutSec, util.FalseValue, slPercentage, slDurationMs),
		routingQueue.GenerateRoutingRules(rrOperator, rrThreshold, rrWaitSeconds),
		routingQueue.GenerateDefaultScriptIDs(chatScriptID, emailScriptID),
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: routingQueue,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_routing_queue."+queueLabel, "name", queueName),
					resource.TestCheckResourceAttr("genesyscloud_routing_queue."+queueLabel, "description", description),
					resource.TestCheckResourceAttr("genesyscloud_routing_queue."+queueLabel, "auto_answer_only", "true"),
					resource.TestCheckResourceAttr("genesyscloud_routing_queue."+queueLabel, "default_script_ids.CHAT", chatScriptID),
					resource.TestCheckResourceAttr("genesyscloud_routing_queue."+queueLabel, "default_script_ids.EMAIL", emailScriptID),
					validateMediaSettings(queueLabel, "media_settings_call", alertTimeoutSec, slPercentage, slDurationMs),
					validateRoutingRules(queueLabel, 0, rrOperator, rrThreshold, rrWaitSeconds),
				),
			},
			{
				Config: routingQueue + generateTfExportByFilter("export",
					exportTestDir,
					util.TrueValue,
					[]string{strconv.Quote("genesyscloud_routing_queue::" + queueName)},
					"",
					strconv.Quote("hcl"),
					util.FalseValue,
					[]string{
						strconv.Quote("genesyscloud_routing_queue." + queueLabel),
					},
				),
				Check: resource.ComposeTestCheckFunc(
					getExportedFileContents(pathToHclFile, &exportContents),
				),
			},
		},
		CheckDestroy: testVerifyExportsDestroyedFunc(exportTestDir),
	})

	exportContents = removeTerraformProviderBlock(exportContents)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: exportContents,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_routing_queue."+queueLabel, "name", queueName),
					resource.TestCheckResourceAttr("genesyscloud_routing_queue."+queueLabel, "description", description),
					resource.TestCheckResourceAttr("genesyscloud_routing_queue."+queueLabel, "auto_answer_only", "true"),
					resource.TestCheckResourceAttr("genesyscloud_routing_queue."+queueLabel, "default_script_ids.CHAT", chatScriptID),
					resource.TestCheckResourceAttr("genesyscloud_routing_queue."+queueLabel, "default_script_ids.EMAIL", emailScriptID),
					validateMediaSettings(queueLabel, "media_settings_call", alertTimeoutSec, slPercentage, slDurationMs),
					validateRoutingRules(queueLabel, 0, rrOperator, rrThreshold, rrWaitSeconds),
				),
			},
		},
	})
}

func TestAccResourceTfExportLogMissingPermissions(t *testing.T) {
	testSetup(t)
	var (
		exportTestDir           = testrunner.GetTestTempPath(".terraform" + uuid.NewString())
		configPath              = filepath.Join(exportTestDir, defaultTfJSONFile)
		permissionsErrorMessage = "API Error 403 - Missing view permissions."
		otherErrorMessage       = "API Error 411 - Another type of error."

		err1    = diag.Diagnostic{Summary: permissionsErrorMessage}
		err2    = diag.Diagnostic{Summary: otherErrorMessage}
		errors1 = diag.Diagnostics{err1}
		errors2 = diag.Diagnostics{err1, err2}
	)

	defer os.RemoveAll(exportTestDir)

	mockError = errors1

	// Checking that the config file is created when the error is 403 & log_permission_errors = true
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: generateTfExportByFilter("test-export",
					exportTestDir,
					util.FalseValue,
					[]string{strconv.Quote("genesyscloud_quality_forms_evaluation")},
					"",
					strconv.Quote("json"),
					util.TrueValue,
					[]string{}),
				Check: resource.ComposeTestCheckFunc(
					validateFileCreated(configPath),
				),
			},
		},
		CheckDestroy: testVerifyExportsDestroyedFunc(exportTestDir),
	})

	// Check that the export fails when a non-403 error exists
	mockError = errors2

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: generateTfExportByFilter("test-export",
					exportTestDir,
					util.FalseValue,
					[]string{strconv.Quote("genesyscloud_quality_forms_evaluation")},
					"",
					strconv.Quote("json"),
					util.TrueValue,
					[]string{}),
				ExpectError: regexp.MustCompile(otherErrorMessage),
			},
		},
		CheckDestroy: testVerifyExportsDestroyedFunc(exportTestDir),
	})

	// Check that info about attr exists in error summary when 403 is found & log_permission_errors = false
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: generateTfExportByFilter("test-export",
					exportTestDir,
					util.FalseValue,
					[]string{strconv.Quote("genesyscloud_quality_forms_evaluation")},
					"",
					strconv.Quote("json"),
					util.FalseValue,
					[]string{}),
				ExpectError: regexp.MustCompile(logAttrInfo),
			},
		},
		CheckDestroy: testVerifyExportsDestroyedFunc(exportTestDir),
	})

	// Clean up
	mockError = nil
}

func TestAccResourceTfExportUserPromptExportAudioFile(t *testing.T) {
	testSetup(t)
	var (
		userPromptResourceLabel     = "test_prompt"
		userPromptName              = "TestPrompt" + strings.Replace(uuid.NewString(), "-", "", -1)
		userPromptDescription       = "Test description"
		userPromptResourceLanguage  = "en-us"
		userPromptResourceText      = "This is a test greeting!"
		userResourcePromptFilename1 = testrunner.GetTestDataPath("resource", userPrompt.ResourceType, "test-prompt-01.wav")
		userResourcePromptFilename2 = testrunner.GetTestDataPath("resource", userPrompt.ResourceType, "test-prompt-02.wav")

		userPromptResourceLanguage2 = "pt-br"
		userPromptResourceText2     = "This is a test greeting!!!"

		exportResourceLabel = "export"
		exportTestDir       = testrunner.GetTestTempPath(".terraform" + uuid.NewString())
	)

	userPromptAsset := userPrompt.UserPromptResourceStruct{
		Language:        userPromptResourceLanguage,
		Tts_string:      util.NullValue,
		Text:            strconv.Quote(userPromptResourceText),
		Filename:        strconv.Quote(userResourcePromptFilename1),
		FileContentHash: userResourcePromptFilename1,
	}

	userPromptAsset2 := userPrompt.UserPromptResourceStruct{
		Language:        userPromptResourceLanguage2,
		Tts_string:      util.NullValue,
		Text:            strconv.Quote(userPromptResourceText2),
		Filename:        strconv.Quote(userResourcePromptFilename2),
		FileContentHash: userResourcePromptFilename2,
	}

	userPromptResources := []*userPrompt.UserPromptResourceStruct{&userPromptAsset}
	userPromptResources2 := []*userPrompt.UserPromptResourceStruct{&userPromptAsset, &userPromptAsset2}

	defer os.RemoveAll(exportTestDir)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: userPrompt.GenerateUserPromptResource(&userPrompt.UserPromptStruct{
					ResourceLabel: userPromptResourceLabel,
					Name:          userPromptName,
					Description:   strconv.Quote(userPromptDescription),
					Resources:     userPromptResources,
				}),
			},
			{
				Config: generateTfExportByFilter(
					exportResourceLabel,
					exportTestDir,
					util.FalseValue,
					[]string{strconv.Quote("genesyscloud_architect_user_prompt::" + userPromptName)},
					"",
					strconv.Quote("json"),
					util.FalseValue,
					[]string{
						strconv.Quote("genesyscloud_architect_user_prompt." + userPromptResourceLabel),
					},
				) + userPrompt.GenerateUserPromptResource(&userPrompt.UserPromptStruct{
					ResourceLabel: userPromptResourceLabel,
					Name:          userPromptName,
					Description:   strconv.Quote(userPromptDescription),
					Resources:     userPromptResources,
				}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_architect_user_prompt."+userPromptResourceLabel, "name", userPromptName),
					resource.TestCheckResourceAttr("genesyscloud_architect_user_prompt."+userPromptResourceLabel, "description", userPromptDescription),
					resource.TestCheckResourceAttr("genesyscloud_architect_user_prompt."+userPromptResourceLabel, "resources.0.language", userPromptResourceLanguage),
					resource.TestCheckResourceAttr("genesyscloud_architect_user_prompt."+userPromptResourceLabel, "resources.0.text", userPromptResourceText),
					resource.TestCheckResourceAttr("genesyscloud_architect_user_prompt."+userPromptResourceLabel, "resources.0.filename", userResourcePromptFilename1),
					testUserPromptAudioFileExport(filepath.Join(exportTestDir, defaultTfJSONFile), "genesyscloud_architect_user_prompt", userPromptResourceLabel, exportTestDir, userPromptName),
				),
			},
			// Update to two resources with separate audio files
			{
				Config: userPrompt.GenerateUserPromptResource(&userPrompt.UserPromptStruct{
					ResourceLabel: userPromptResourceLabel,
					Name:          userPromptName,
					Description:   strconv.Quote(userPromptDescription),
					Resources:     userPromptResources2,
				}),
				ExpectNonEmptyPlan: true,
			},
			{
				Config: generateTfExportByFilter(
					exportResourceLabel,
					exportTestDir,
					util.FalseValue,
					[]string{strconv.Quote("genesyscloud_architect_user_prompt::" + userPromptName)},
					"",
					strconv.Quote("json"),
					util.FalseValue,
					[]string{
						strconv.Quote("genesyscloud_architect_user_prompt." + userPromptResourceLabel),
					},
				) + userPrompt.GenerateUserPromptResource(&userPrompt.UserPromptStruct{
					ResourceLabel: userPromptResourceLabel,
					Name:          userPromptName,
					Description:   strconv.Quote(userPromptDescription),
					Resources:     userPromptResources2,
				}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_architect_user_prompt."+userPromptResourceLabel, "name", userPromptName),
					resource.TestCheckResourceAttr("genesyscloud_architect_user_prompt."+userPromptResourceLabel, "description", userPromptDescription),
					resource.TestCheckResourceAttr("genesyscloud_architect_user_prompt."+userPromptResourceLabel, "resources.0.language", userPromptResourceLanguage),
					resource.TestCheckResourceAttr("genesyscloud_architect_user_prompt."+userPromptResourceLabel, "resources.0.text", userPromptResourceText),
					resource.TestCheckResourceAttr("genesyscloud_architect_user_prompt."+userPromptResourceLabel, "resources.0.filename", userResourcePromptFilename1),
					testUserPromptAudioFileExport(filepath.Join(exportTestDir, defaultTfJSONFile), "genesyscloud_architect_user_prompt", userPromptResourceLabel, exportTestDir, userPromptName),
				),
			},
			{
				ResourceName:            "genesyscloud_architect_user_prompt." + userPromptResourceLabel,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"resources"},
			},
		},
		CheckDestroy: testVerifyExportsDestroyedFunc(exportTestDir),
	})
}

func TestAccResourceSurveyFormsPublishedAndUnpublished(t *testing.T) {
	testSetup(t)
	var (
		exportTestDir = testrunner.GetTestTempPath(".terraformregex" + uuid.NewString())
		resourceLabel = "export"
		configPath    = filepath.Join(exportTestDir, defaultTfJSONFile)
		statePath     = filepath.Join(exportTestDir, defaultTfStateFile)
	)

	// Clean up
	defer func(path string) {
		if err := os.RemoveAll(path); err != nil {
			t.Logf("failed to remove dir %s: %s", path, err)
		}
	}(exportTestDir)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: generateTfExportByIncludeFilterResources(
					resourceLabel,
					exportTestDir,
					util.TrueValue, // include_state_file
					[]string{ // include_filter_resources
						strconv.Quote("genesyscloud_quality_forms_survey"),
					},
					strconv.Quote("json"), // export_as_hcl
					util.FalseValue,
					[]string{},
				),
				Check: resource.ComposeTestCheckFunc(
					validatePublishedAndUnpublishedExported(configPath),
					validateStateFileHasPublishedAndUnpublished(statePath),
				),
			},
		},
	})
}

// TestAccResourceExportManagedSitesAsData checks that during an export, managed sites are exported as data source
// Managed can't be set on sites, therefore the default managed site is checked during the test that it is exported as data
func TestAccResourceExportManagedSitesAsData(t *testing.T) {
	testSetup(t)
	var (
		exportTestDir = testrunner.GetTestTempPath(".terraform" + uuid.NewString())
		resourceLabel = "export"
		configPath    = filepath.Join(exportTestDir, defaultTfJSONFile)
		statePath     = filepath.Join(exportTestDir, defaultTfStateFile)
		siteName      = "PureCloud Voice - AWS"
	)

	if err := telephonyProvidersEdgesSite.CheckForDefaultSite(siteName); err != nil {
		t.Skipf("failed to get default site %v", err)
	}

	platform := platform.GetPlatform()
	if platform.IsDevelopmentPlatform() {
		t.Skip("Skipping test for development platform due to inability to properly convert statefile to v4 within development platform")
	}

	defer func(path string) {
		if err := os.RemoveAll(path); err != nil {
			t.Logf("failed to remove dir %s: %s", path, err)
		}
	}(exportTestDir)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: generateTfExportByIncludeFilterResources(
					resourceLabel,
					exportTestDir,
					util.TrueValue, // include_state_file
					[]string{ // include_filter_resources
						strconv.Quote("genesyscloud_telephony_providers_edges_site"),
					},
					strconv.Quote("json"), // export_format attribute
					util.FalseValue,
					[]string{},
				),
				Check: resource.ComposeTestCheckFunc(
					validateStateFileAsData(statePath, siteName),
					validateExportManagedSitesAsData(configPath, siteName),
				),
			},
		},
	})
}

// TestAccResourceTfExportSplitFilesAsHCL will create 2 queues, 2 wrap up codes, and 2 users.
// The exporter will be run in split mode so 3 resource tfs should be created as well as a provider.tf
func TestAccResourceTfExportSplitFilesAsHCL(t *testing.T) {
	testSetup(t)
	var (
		exportTestDir       = testrunner.GetTestTempPath(".terraform" + uuid.NewString())
		exportResourceLabel = "test-export-split"
		uniquePostfix       = randString(7)
		expectedFilesPath   = []string{
			filepath.Join(exportTestDir, "genesyscloud_routing_queue.tf"),
			filepath.Join(exportTestDir, "genesyscloud_user.tf"),
			filepath.Join(exportTestDir, "genesyscloud_routing_wrapupcode.tf"),
			filepath.Join(exportTestDir, "provider.tf"),
		}

		queueResources = []QueueExport{
			{OriginalResourceLabel: "test-queue-1", ExportedLabel: "test-queue-1-" + uuid.NewString() + uniquePostfix, Description: "This is a test queue", AcwTimeoutMs: 200000},
			{OriginalResourceLabel: "test-queue-2", ExportedLabel: "test-queue-1-" + uuid.NewString() + uniquePostfix, Description: "This is a test queue too", AcwTimeoutMs: 200000},
		}

		userResources = []UserExport{
			{OriginalResourceLabel: "test-user-1", ExportedLabel: "test-user-1", Email: "test-user-1" + uuid.NewString() + "@test.com" + uniquePostfix, State: "active"},
			{OriginalResourceLabel: "test-user-2", ExportedLabel: "test-user-2", Email: "test-user-2" + uuid.NewString() + "@test.com" + uniquePostfix, State: "active"},
		}

		wrapupCodeResources = []WrapupcodeExport{
			{OriginalResourceLabel: "test-wrapupcode-1", Name: "test-wrapupcode-1-" + uuid.NewString() + uniquePostfix},
			{OriginalResourceLabel: "test-wrapupcode-2", Name: "test-wrapupcode-2-" + uuid.NewString() + uniquePostfix},
		}

		divResourceLabel = "test-division"
		divName          = "terraform-" + uuid.NewString()
		description      = "Terraform wrapup code description"
	)
	defer os.RemoveAll(exportTestDir)

	queueResourceDef := buildQueueResources(queueResources)
	userResourcesDef := buildUserResources(userResources)

	wrapupcodeResourceDef := buildWrapupcodeResources(wrapupCodeResources, "genesyscloud_auth_division."+divResourceLabel+".id", description)
	baseConfig := queueResourceDef + authDivision.GenerateAuthDivisionBasic(divResourceLabel, divName) + wrapupcodeResourceDef + userResourcesDef
	configWithExporter := baseConfig + generateTfExportByIncludeFilterResources(
		exportResourceLabel,
		exportTestDir,
		util.TrueValue,
		[]string{
			strconv.Quote("genesyscloud_routing_queue::" + uniquePostfix + "$"),
			strconv.Quote("genesyscloud_user::" + uniquePostfix + "$"),
			strconv.Quote("genesyscloud_routing_wrapupcode::" + uniquePostfix + "$"),
		},
		strconv.Quote("hcl"),
		util.TrueValue,
		[]string{
			strconv.Quote("genesyscloud_routing_queue." + queueResources[0].OriginalResourceLabel),
			strconv.Quote("genesyscloud_routing_queue." + queueResources[1].OriginalResourceLabel),
			strconv.Quote("genesyscloud_user." + userResources[0].OriginalResourceLabel),
			strconv.Quote("genesyscloud_user." + userResources[1].OriginalResourceLabel),
			strconv.Quote("genesyscloud_routing_wrapupcode." + wrapupCodeResources[0].OriginalResourceLabel),
			strconv.Quote("genesyscloud_routing_wrapupcode." + wrapupCodeResources[1].OriginalResourceLabel),
		},
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: baseConfig,
			},
			{
				Config: configWithExporter,
				Check: resource.ComposeTestCheckFunc(
					validateFileCreated(expectedFilesPath[0]),
					validateFileCreated(expectedFilesPath[1]),
					validateFileCreated(expectedFilesPath[2]),
					validateFileCreated(expectedFilesPath[3]),
				),
			},
		},
		CheckDestroy: testVerifyExportsDestroyedFunc(exportTestDir),
	})
}

// TestAccResourceUserPromptsExported tests the new getAll functionality where it adds the name filter and makes a call per letter
// This will prevent the 10,000 return limit being hit on export and not returning everything
func TestAccResourceTfExportUserPromptsExported(t *testing.T) {
	testSetup(t)
	var (
		uniqueStr            = strings.Replace(uuid.NewString(), "-", "_", -1)
		exportTestDir        = testrunner.GetTestTempPath(".terraform" + uuid.NewString())
		resourceID           = "export"
		dPromptResourceLabel = "d_user_prompt"
		dPromptNameAttr      = "d_user_prompt_test" + uniqueStr
		hPromptResourceLabel = "h_user_prompt"
		hPromptNameAttr      = "h_user_prompt_test" + uniqueStr
		zPromptResourceLabel = "Z_user_prompt"
		zPromptNameAttr      = "Z_user_prompt_test" + uniqueStr

		allNames = []string{dPromptNameAttr, hPromptNameAttr, zPromptNameAttr}
	)

	promptConfig := fmt.Sprintf(`
resource "genesyscloud_architect_user_prompt" "%s" {
	name        = "%s"
	description = "Welcome greeting for all callers"
	resources {
		language   = "en-us"
		text       = "Good day. Thank you for calling."
		tts_string = "Good day. Thank you for calling."
	}
}

resource "genesyscloud_architect_user_prompt" "%s" {
	name        = "%s"
	description = "Welcome greeting for all callers"
	resources {
		language   = "en-us"
		text       = "Good day. Thank you for calling."
		tts_string = "Good day. Thank you for calling."
	}
}

resource "genesyscloud_architect_user_prompt" "%s" {
	name        = "%s"
	description = "Welcome greeting for all callers"
	resources {
		language   = "en-us"
		text       = "Good day. Thank you for calling."
		tts_string = "Good day. Thank you for calling."
	}
}
`, dPromptResourceLabel, dPromptNameAttr, hPromptResourceLabel, hPromptNameAttr, zPromptResourceLabel, zPromptNameAttr)

	defer func(path string) {
		if err := os.RemoveAll(path); err != nil {
			t.Logf("failed to remove dir %s: %s", path, err)
		}
	}(exportTestDir)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: promptConfig,
			},
			{
				Config: promptConfig + generateTfExportByIncludeFilterResources(
					resourceID,
					exportTestDir,
					util.TrueValue, // include_state_file
					[]string{ // include_filter_resources
						strconv.Quote("genesyscloud_architect_user_prompt::" + dPromptNameAttr),
						strconv.Quote("genesyscloud_architect_user_prompt::" + hPromptNameAttr),
						strconv.Quote("genesyscloud_architect_user_prompt::" + zPromptNameAttr),
					},
					strconv.Quote("json"), // export_format attribute
					util.FalseValue,
					[]string{
						strconv.Quote("genesyscloud_architect_user_prompt." + dPromptResourceLabel),
						strconv.Quote("genesyscloud_architect_user_prompt." + hPromptResourceLabel),
						strconv.Quote("genesyscloud_architect_user_prompt." + zPromptResourceLabel),
					},
				),
				Check: resource.ComposeTestCheckFunc(
					validatePromptsExported(exportTestDir, allNames),
				),
			},
		},
	})
}

// TestAccResourceTfExportCampaignScriptIdReferences exports two campaigns and ensures that the custom revolver OutboundCampaignAgentScriptResolver
// is working properly i.e. script_id should reference a data source pointing to the Default Outbound Script under particular circumstances
func TestAccResourceTfExportCampaignScriptIdReferences(t *testing.T) {
	testSetup(t)
	var (
		exportTestDir = testrunner.GetTestTempPath(".terraform" + uuid.NewString())
		resourceLabel = "export"

		campaignNameDefaultScript          = "tf test df campaign " + uuid.NewString()
		campaignResourceLabelDefaultScript = strings.Replace(campaignNameDefaultScript, " ", "_", -1)

		campaignNameCustomScript          = "tf test ct campaign " + uuid.NewString()
		campaignResourceLabelCustomScript = strings.Replace(campaignNameCustomScript, " ", "_", -1)

		contactListResourceLabel = "contact_list"
		contactListName          = "tf test contact list " + uuid.NewString()
		queueName                = "tf test queue " + uuid.NewString()

		scriptName            = "tf_test_script_" + uuid.NewString()
		pathToScriptFile      = testrunner.GetTestDataPath("resource", "genesyscloud_script", "test_script.json")
		fullyQualifiedPath, _ = filepath.Abs(pathToScriptFile)

		configPath = filepath.Join(exportTestDir, defaultTfJSONFile)
	)

	contactListConfig := fmt.Sprintf(`
resource "genesyscloud_outbound_contact_list" "%s" {
  name         = "%s"
  column_names = ["First Name", "Last Name", "Cell", "Home"]
  phone_columns {
    column_name = "Cell"
    type        = "cell"
  }
  phone_columns {
    column_name = "Home"
    type        = "home"
  }
}
`, contactListResourceLabel, contactListName)

	remainingConfig := fmt.Sprintf(`
data "genesyscloud_script" "default" {
	name = "Default Outbound Script"
}

resource "genesyscloud_outbound_campaign" "%s" {
  name            = "%s"
  queue_id        = "${genesyscloud_routing_queue.queue.id}"
  caller_address  = "+13335551234"
  contact_list_id = "${genesyscloud_outbound_contact_list.%s.id}"
  dialing_mode    = "preview"
  script_id       = "${data.genesyscloud_script.default.id}"
  caller_name     = "Callbacks Test Queue 1"
  division_id     = "${data.genesyscloud_auth_division_home.home.id}"
  campaign_status = "off"
  dynamic_contact_queueing_settings {
    sort = false
  }
  phone_columns {
    column_name = "Cell"
  }
}

resource "genesyscloud_outbound_campaign" "%s" {
  name            = "%s"
  queue_id        = "${genesyscloud_routing_queue.queue.id}"
  caller_address  = "+13335551234"
  contact_list_id = "${genesyscloud_outbound_contact_list.%s.id}"
  dialing_mode    = "preview"
  script_id       = "${genesyscloud_script.script.id}"
  caller_name     = "Callbacks Test Queue 1"
  division_id     = "${data.genesyscloud_auth_division_home.home.id}"
  campaign_status = "off"
  dynamic_contact_queueing_settings {
    sort = false
  }
  phone_columns {
    column_name = "Cell"
  }
}

resource "genesyscloud_script" "script" {
	script_name       = "%s"
	filepath          = "%s"
	file_content_hash = filesha256("%s")
}

data "genesyscloud_auth_division_home" "home" {}

resource "genesyscloud_routing_queue" "queue" {
  name = "%s"
}
`, campaignResourceLabelDefaultScript,
		campaignNameDefaultScript,
		contactListResourceLabel,
		campaignResourceLabelCustomScript,
		campaignNameCustomScript,
		contactListResourceLabel,
		scriptName,
		pathToScriptFile,
		fullyQualifiedPath,
		queueName)

	defer func(path string) {
		if err := os.RemoveAll(path); err != nil {
			t.Logf("failed to remove dir %s: %s", path, err)
		}
	}(exportTestDir)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: contactListConfig, // need to add contact list first to seed it with contacts before creating the campaigns
				Check:  addContactsToContactList,
			},
			{
				Config: remainingConfig + contactListConfig,
			},
			// Verify script_id fields are resolved properly when include_state_file = false and enable_dependency_resolution = true
			{
				Config: remainingConfig + contactListConfig + generateExportResourceIncludeFilterWithEnableDepRes(
					resourceLabel,
					exportTestDir,
					util.FalseValue,       // include_state_file
					strconv.Quote("json"), // export_format attribute
					util.TrueValue,        // enable_dependency_resolution
					[]string{ // include_filter_resources
						strconv.Quote("genesyscloud_outbound_campaign::" + campaignNameDefaultScript),
						strconv.Quote("genesyscloud_outbound_campaign::" + campaignNameCustomScript),
					},
					[]string{ // depends_on
						"genesyscloud_outbound_campaign." + campaignResourceLabelDefaultScript,
						"genesyscloud_outbound_campaign." + campaignResourceLabelCustomScript,
					},
				),
				Check: resource.ComposeTestCheckFunc(
					validateExportedCampaignScriptIds(
						configPath,
						campaignResourceLabelCustomScript,
						campaignResourceLabelDefaultScript,
						"${data.genesyscloud_script.Default_Outbound_Script.id}",
						fmt.Sprintf("${genesyscloud_script.%s.id}", scriptName),
						true,
					),
					validateNumberOfExportedDataSources(configPath),
				),
			},
			// Verify script_id fields are resolved properly when include_state_file = true and enable_dependency_resolution = true
			{
				Config: remainingConfig + contactListConfig + generateExportResourceIncludeFilterWithEnableDepRes(
					resourceLabel,
					exportTestDir,
					util.TrueValue,        // include_state_file
					strconv.Quote("json"), // export_format attribute
					util.TrueValue,        // enable_dependency_resolution
					[]string{ // include_filter_resources
						strconv.Quote("genesyscloud_outbound_campaign::" + campaignNameDefaultScript),
						strconv.Quote("genesyscloud_outbound_campaign::" + campaignNameCustomScript),
					},
					[]string{ // depends_on
						"genesyscloud_outbound_campaign." + campaignResourceLabelDefaultScript,
						"genesyscloud_outbound_campaign." + campaignResourceLabelCustomScript,
					},
				),
				Check: resource.ComposeTestCheckFunc(
					validateExportedCampaignScriptIds(
						configPath,
						campaignResourceLabelCustomScript,
						campaignResourceLabelDefaultScript,
						"${data.genesyscloud_script.Default_Outbound_Script.id}",
						fmt.Sprintf("${genesyscloud_script.%s.id}", scriptName),
						true,
					),
					validateNumberOfExportedDataSources(configPath),
				),
			},
			// Verify script_id fields are resolved properly when include_state_file = true and enable_dependency_resolution = false
			{
				Config: remainingConfig + contactListConfig + generateExportResourceIncludeFilterWithEnableDepRes(
					resourceLabel,
					exportTestDir,
					util.TrueValue,        // include_state_file
					strconv.Quote("json"), // export_format attribute
					util.FalseValue,       // enable_dependency_resolution
					[]string{ // include_filter_resources
						strconv.Quote("genesyscloud_outbound_campaign::" + campaignNameDefaultScript),
						strconv.Quote("genesyscloud_outbound_campaign::" + campaignNameCustomScript),
					},
					[]string{ // depends_on
						"genesyscloud_outbound_campaign." + campaignResourceLabelDefaultScript,
						"genesyscloud_outbound_campaign." + campaignResourceLabelCustomScript,
					},
				),
				Check: resource.ComposeTestCheckFunc(
					validateExportedCampaignScriptIds(
						configPath,
						campaignResourceLabelCustomScript,
						campaignResourceLabelDefaultScript,
						"${data.genesyscloud_script.Default_Outbound_Script.id}",
						"",
						false,
					),
					validateNumberOfExportedDataSources(configPath),
				),
			},
		},
		CheckDestroy: testVerifyExportsDestroyedFunc(exportTestDir),
	})
}

func TestAccResourceTfExportEnableDependsOn(t *testing.T) {
	testSetup(t)
	var (
		exportTestDir            = testrunner.GetTestTempPath(".terraform" + uuid.NewString())
		exportResourceLabel      = "test-export2"
		contactListResourceLabel = "contact_list" + uuid.NewString()
		contactListName          = "terraform contact list" + uuid.NewString()
		outboundFlowFilePath     = filepath.Join(testrunner.RootDir, "examples/resources/genesyscloud_flow/outboundcall_flow_example.yaml")
		flowName                 = "testflowcxcase"
		flowResourceLabel        = "flow"
		wrapupcodeResourceLabel  = "wrapupcode"
	)
	defer os.RemoveAll(exportTestDir)

	config := fmt.Sprintf(`
data "genesyscloud_auth_division_home" "home" {}
`+`
resource "time_sleep" "wait_10_seconds" {
create_duration = "100s"
}
`) + GenerateOutboundCampaignBasicforFlowExport(
		contactListResourceLabel,
		outboundFlowFilePath,
		flowResourceLabel,
		flowName,
		"${data.genesyscloud_auth_division_home.home.name}",
		wrapupcodeResourceLabel,
		contactListName,
	) +
		generateTfExportByFlowDependsOnResources(
			exportResourceLabel,
			exportTestDir,
			util.TrueValue,
			[]string{
				strconv.Quote("genesyscloud_flow::" + flowName),
			},
			strconv.Quote("json"),
			util.FalseValue,
			util.TrueValue,
		)

	sanitizer := resourceExporter.NewSanitizerProvider()

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		ExternalProviders: map[string]resource.ExternalProvider{
			"time": {
				VersionConstraint: "0.10.0",
				Source:            "hashicorp/time",
			},
		},
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					validateFlow("genesyscloud_flow."+flowResourceLabel, flowName),
					resource.TestCheckResourceAttr("genesyscloud_outbound_contact_list."+contactListResourceLabel, "name", contactListName),
					testDependentContactList(exportTestDir+"/"+defaultTfJSONFile, "genesyscloud_outbound_contact_list", sanitizer.S.SanitizeResourceBlockLabel(contactListName)),
				),
			},
		},
		CheckDestroy: testVerifyExportsDestroyedFunc(exportTestDir),
	})
}

func TestAccResourceExporterFormat(t *testing.T) {
	testSetup(t)
	t.Parallel()
	var (
		exportTestDir        = testrunner.GetTestTempPath(".terraform" + uuid.NewString())
		exportResourceLabel1 = "exportFormatTest-export1"
		jsonConfigFilePath   = filepath.Join(exportTestDir, defaultTfJSONFile)
		hclConfigFilePath    = filepath.Join(exportTestDir, defaultTfHCLFile)
	)

	defer os.RemoveAll(exportTestDir)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		CheckDestroy:      testVerifyExportsDestroyedFunc(exportTestDir),
		Steps: []resource.TestStep{
			{
				Config: generateTfExportResourceExportFormat(
					exportResourceLabel1,
					strconv.Quote("hcl_json"),
					[]string{"genesyscloud_journey_segment"},
					exportTestDir,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_tf_export."+exportResourceLabel1, "export_format", "hcl_json"),
					validateFileCreated(jsonConfigFilePath),
					validateFileCreated(hclConfigFilePath),
					validateConfigFile(jsonConfigFilePath),
					validateHCLConfigFile(hclConfigFilePath),
				),
			},
		},
	})
}

// TestAccResourceTfExportArchitectFlowExporterLegacyAndNew Exports a flow using the legacy exporter (creates a tfvars file but does not export flow config files)
// and then exports using the new archy exporter by setting use_legacy_architect_flow_exporter to false
// Verifies that the appropriate files are/are not created when use_legacy_architect_flow_exporter is set to true/false
// Verifies that the appropriate filepath is set inside the exported resource config when use_legacy_architect_flow_exporter is set to true/false
func TestAccResourceTfExportArchitectFlowExporterLegacyAndNew(t *testing.T) {
	testSetup(t)
	const (
		systemFlowName = "Default Voicemail Flow"
		systemFlowType = "VOICEMAIL"
		systemFlowId   = "de4c63f0-0be1-11ec-9a03-0242ac130003"
	)

	var (
		systemFlowNameSanitized          = strings.Replace(systemFlowName, " ", "_", -1)
		exportedSystemFlowFileName       = architectFlow.BuildExportFileName(systemFlowName, systemFlowType, systemFlowId)
		exportResourceLabel              = "export"
		exportTestDir                    = testrunner.GetTestTempPath(".terraform" + uuid.NewString())
		exportFullPath                   = ResourceType + "." + exportResourceLabel
		pathToFolderHoldingExportedFlows = filepath.Join(exportTestDir, architectFlow.ExportSubDirectoryName)

		exportedFlowResourceLabel               = systemFlowType + "_" + systemFlowNameSanitized
		pathToExportedTerraformConfig           = filepath.Join(exportTestDir, defaultTfJSONFile)
		exportedFlowResourceFullPath            = architectFlow.ResourceType + "." + exportedFlowResourceLabel
		expectedFilepathValueWithLegacyExporter = fmt.Sprintf("${var.genesyscloud_flow_%s_%s_filepath}", systemFlowType, systemFlowNameSanitized)
		expectedFilepathValueWithNewExporter    = filepath.Join(architectFlow.ExportSubDirectoryName, fmt.Sprintf("%s-%s-%s.yaml", systemFlowNameSanitized, systemFlowType, systemFlowId))
	)

	defer func(path string) {
		if err := os.RemoveAll(path); err != nil {
			log.Printf("An error occured while removing directory '%s': %s", exportTestDir, err)
		}
	}(exportTestDir)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: generateTFExportResourceCustom(
					exportResourceLabel,
					exportTestDir,
					util.TrueValue,
					strconv.Quote("json"),
					util.NullValue, // use_legacy_architect_flow_exporter - should default to true
					[]string{
						strconv.Quote(architectFlow.ResourceType + "::" + systemFlowName),
					},
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(exportFullPath, "use_legacy_architect_flow_exporter", util.TrueValue),
					validateFileCreated(filepath.Join(exportTestDir, "terraform.tfvars")),
					validateFileNotCreated(filepath.Join(exportTestDir, architectFlow.ExportSubDirectoryName)),
					validateExportedResourceAttributeValue(pathToExportedTerraformConfig, exportedFlowResourceFullPath, "filepath", expectedFilepathValueWithLegacyExporter),
				),
			},
			{
				Config: generateTFExportResourceCustom(
					exportResourceLabel,
					exportTestDir,
					util.TrueValue,
					strconv.Quote("json"),
					util.FalseValue, // use_legacy_architect_flow_exporter
					[]string{
						strconv.Quote(architectFlow.ResourceType + "::" + systemFlowName),
					},
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(exportFullPath, "use_legacy_architect_flow_exporter", util.FalseValue),
					validateFileNotCreated(filepath.Join(exportTestDir, "terraform.tfvars")),
					validateFileCreated(pathToFolderHoldingExportedFlows),
					validateFileCreated(filepath.Join(pathToFolderHoldingExportedFlows, exportedSystemFlowFileName)),
					validateExportedResourceAttributeValue(pathToExportedTerraformConfig, exportedFlowResourceFullPath, "filepath", expectedFilepathValueWithNewExporter),
				),
			},
		},
		CheckDestroy: testVerifyExportsDestroyedFunc(exportTestDir),
	})
}

// TestUnitTestForExportCycles creates a directed graph of exported resources to their references. Report any potential graph cycles in this test.
// Reference cycles can sometimes be broken by exporting a separate resource to update membership after the member
// and container resources are created/updated (see genesyscloud_user_roles).
func TestUnitTestForExportCycles(t *testing.T) {

	// Assumes exporting all resource types
	exporters := resourceExporter.GetResourceExporters()

	graph := simple.NewDirectedGraph()

	var resTypes []string
	for resType := range exporters {
		graph.AddNode(simple.Node(len(resTypes)))
		resTypes = append(resTypes, resType)
	}

	for i, resType := range resTypes {
		currNode := simple.Node(i)
		for attrName, refSettings := range exporters[resType].RefAttrs {
			if refSettings.RefType == resType {
				// Resources that can reference themselves are ignored
				// Cycles caused by self-refs are likely a misconfiguration
				// (e.g. two users that are each other's manager)
				continue
			}
			if exporters[resType].IsAttributeExcluded(attrName) {
				// This reference attribute will be excluded from the export
				continue
			}
			graph.SetEdge(simple.Edge{F: currNode, T: simple.Node(resNodeIndex(refSettings.RefType, resTypes))})
		}
	}

	cycles := topo.DirectedCyclesIn(graph)
	if len(cycles) > 0 {
		cycleResources := make([][]string, 0)
		for _, cycle := range cycles {
			cycleTemp := make([]string, len(cycle))
			for j, cycleNode := range cycle {
				if resTypes != nil {
					cycleTemp[j] = resTypes[cycleNode.ID()]
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

// TestAccResourceTfExportSanitizedDuplicateLabels creates multiple data actions that will share the same sanitized labels and require
// hashes to be appended to them to guarantee uniqueness.
// The test will then export them, parse the exported tf state and tf configuration files, and validate that the labels are all present
// and that no labels appear more than once either file
func TestAccResourceTfExportSanitizedDuplicateLabels(t *testing.T) {
	testSetup(t)
	var (
		exportTestDir = testrunner.GetTestTempPath(".terraform" + uuid.NewString())

		integrationLabel = "integration"
		integrationName  = "tf test integration " + uuid.NewString()

		credentialLabel = "credential"
		credentialName  = "tf test credential " + uuid.NewString()

		dataActionLabel1 = "action_1"
		dataActionLabel2 = "action_2"
		dataActionLabel3 = "action_3"
		dataActionName   = "tf test data action " + uuid.NewString()

		stateFilePath = filepath.Join(exportTestDir, defaultTfStateFile)
		configPath    = filepath.Join(exportTestDir, defaultTfJSONFile)

		sanitizer      = resourceExporter.NewSanitizerProvider()
		sanitizedName  = sanitizer.S.SanitizeResourceBlockLabel(dataActionName)
		expectedLabels = []string{
			sanitizedName,
			sanitizedName + "_" + sanitizeResourceHash(dataActionName+"2"),
			sanitizedName + "_" + sanitizeResourceHash(dataActionName+"3"),
		}
	)

	defer func(path string) {
		if err := os.RemoveAll(path); err != nil {
			t.Logf("Failed to cleanup export directory: %s", err.Error())
		}
	}(exportTestDir)

	dataActionResource := func(label string) string {
		return fmt.Sprintf(`
resource "genesyscloud_integration_action" "%s" {
  name           = "%s"
  category       = "%s"
  integration_id = genesyscloud_integration.%s.id
  secure         = false
  config_request {
    request_template     = "$${input.rawRequest}"
    request_type         = "POST"
    request_url_template = "/api/v2/conversations/$${input.conversationId}/disconnect"
  }
  contract_output = jsonencode({
    "properties" : {},
    "type" : "object"
  })
  config_response {
    success_template = "$${rawResult}"
  }
  contract_input = jsonencode({
    "properties" : {
      "conversationId" : {
        "type" : "string"
      }
    },
    "type" : "object"
  })
}
`, label, dataActionName, integrationName, integrationLabel)
	}

	config := fmt.Sprintf(`
locals {
  shared_action_name = "%s"
  integration_name   = "%s"
}

resource "genesyscloud_tf_export" "export" {
  directory          = "%s"
  include_state_file = true
  export_format      = "json"
  include_filter_resources = [
    "genesyscloud_integration_action::${local.shared_action_name}",
    "genesyscloud_integration::${local.integration_name}"
  ]

  depends_on = [
    genesyscloud_integration_action.%s,
    genesyscloud_integration_action.%s,
    genesyscloud_integration_action.%s,
  ]
}

resource "genesyscloud_integration" "%s" {
  config {
    advanced = jsonencode({})
    credentials = {
      pureCloudOAuthClient = genesyscloud_integration_credential.%s.id
    }
    name       = local.integration_name
    properties = jsonencode({})
  }
  integration_type = "purecloud-data-actions"
  intended_state   = "ENABLED"
}

resource "genesyscloud_integration_credential" "%s" {
  name                 = "%s"
  credential_type_name = "pureCloudOAuthClient"
  fields = {
    clientId     = "someUserName"
    clientSecret = "$tr0ngP@s$w0rd"
  }
}

%s

%s

%s
`, dataActionName, integrationName, exportTestDir,
		dataActionLabel1,
		dataActionLabel2,
		dataActionLabel3,
		integrationLabel,
		credentialLabel,
		credentialLabel,
		credentialName,
		dataActionResource(dataActionLabel1),
		dataActionResource(dataActionLabel2),
		dataActionResource(dataActionLabel3),
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					verifyLabelsExistInExportedStateFile(stateFilePath, integrationAction.ResourceType, expectedLabels),
					verifyLabelsExistInExportedTfConfig(configPath, integrationAction.ResourceType, expectedLabels),
				),
			},
		},
		CheckDestroy: testVerifyExportsDestroyedFunc(exportTestDir),
	})
}
