package tfexporter

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	recordingSettings "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/recording_settings"
	sttCategory "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/speechandtextanalytics_category"
	sttProgram "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/speechandtextanalytics_program"
	sttSentimentFeedback "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/speechandtextanalytics_sentimentfeedback"
	sttSettings "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/speechandtextanalytics_settings"
	sttTopic "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/speechandtextanalytics_topic"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/testrunner"
)

func TestUnitSttRecordingExportersAreRegistered(t *testing.T) {
	for _, resourceType := range []string{
		sttTopic.ResourceType,
		sttProgram.ResourceType,
		sttCategory.ResourceType,
		sttSentimentFeedback.ResourceType,
		sttSettings.ResourceType,
		recordingSettings.ResourceType,
	} {
		if resourceExporters[resourceType] == nil {
			t.Errorf("no exporter registered for %s", resourceType)
		}
		if providerResources[resourceType] == nil {
			t.Errorf("no resource registered for %s", resourceType)
		}
	}
}

func TestUnitSttProgramExporterResolvesTopicReferences(t *testing.T) {
	ref := sttProgram.SpeechAndTextAnalyticsProgramExporter().RefAttrs["topic_ids"]
	if ref == nil {
		t.Fatal("topic_ids is not a reference attribute; exported programs would hard-code topic GUIDs")
	}
	if ref.RefType != sttTopic.ResourceType {
		t.Errorf("topic_ids RefType = %q, want %q", ref.RefType, sttTopic.ResourceType)
	}
}

func TestUnitSttCategoryExporterJsonEncodesCriteria(t *testing.T) {
	exporter := sttCategory.CategoryExporter()
	for _, attr := range exporter.JsonEncodeAttributes {
		if attr == "criteria" {
			return
		}
	}
	t.Errorf("criteria is not in JsonEncodeAttributes (%v); exported category criteria would not be valid HCL", exporter.JsonEncodeAttributes)
}

func TestUnitSttRecordingSingletonExporters(t *testing.T) {
	cases := map[string]struct {
		singleton bool
		exportId  string
	}{
		sttSettings.ResourceType: {
			singleton: sttSettings.SpeechAndTextAnalyticsSettingsExporter().IsSingleton,
			exportId:  sttSettings.SpeechAndTextAnalyticsSettingsExporter().ExportId,
		},
		recordingSettings.ResourceType: {
			singleton: recordingSettings.RecordingSettingsExporter().IsSingleton,
			exportId:  recordingSettings.RecordingSettingsExporter().ExportId,
		},
	}
	for resourceType, c := range cases {
		if !c.singleton {
			t.Errorf("%s exporter must be a singleton", resourceType)
		}
		if c.exportId != resourceType {
			t.Errorf("%s exporter ExportId = %q, want the resource type", resourceType, c.exportId)
		}
	}
}

func TestAccResourceTfExportSttProgramTopicReference(t *testing.T) {
	testSetup(t)

	var (
		exportDir   = testrunner.GetTestTempPath(".terraform" + uuid.NewString())
		topicName   = "tfacc-export-topic-" + uuid.NewString()
		programName = "tfacc-export-program-" + uuid.NewString()
		resources   = fmt.Sprintf(`
resource "genesyscloud_speechandtextanalytics_topic" "topic" {
  name        = %q
  dialect     = "en-US"
  description = "tf export acceptance test"
}

resource "genesyscloud_speechandtextanalytics_program" "program" {
  name      = %q
  topic_ids = [genesyscloud_speechandtextanalytics_topic.topic.id]
}
`, topicName, programName)
	)
	defer removeExportDir(exportDir)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: resources,
			},
			{
				Config: resources + generateSttExport(exportDir,
					[]string{sttTopic.ResourceType, sttProgram.ResourceType},
					"genesyscloud_speechandtextanalytics_program.program", "genesyscloud_speechandtextanalytics_topic.topic"),
				Check: resource.ComposeTestCheckFunc(
					assertSttExportedBlock(exportDir, sttTopic.ResourceType, "name", topicName, nil),
					assertSttExportedBlock(exportDir, sttProgram.ResourceType, "name", programName, func(attrs map[string]interface{}) error {
						ids, _ := attrs["topic_ids"].([]interface{})
						if len(ids) != 1 {
							return fmt.Errorf("expected 1 exported topic_ids entry, got %v", attrs["topic_ids"])
						}
						ref, _ := ids[0].(string)
						if !strings.HasPrefix(ref, "${"+sttTopic.ResourceType+".") {
							return fmt.Errorf("topic_ids[0] = %q, expected a reference to a %s resource", ref, sttTopic.ResourceType)
						}
						return nil
					}),
				),
			},
		},
		CheckDestroy: testVerifyExportsDestroyedFunc(exportDir),
	})
}

func TestAccResourceTfExportSttCategoryCriteria(t *testing.T) {
	testSetup(t)

	var (
		exportDir    = testrunner.GetTestTempPath(".terraform" + uuid.NewString())
		categoryName = "tfacc-export-category-" + uuid.NewString()
		word         = "refund"
		resources    = fmt.Sprintf(`
resource "genesyscloud_speechandtextanalytics_category" "category" {
  name             = %q
  interaction_type = "Voice"
  criteria = jsonencode({
    type       = "OperandGroup"
    inverted   = false
    occurrence = 1
    operands = [
      {
        type       = "OperandGroup"
        inverted   = false
        occurrence = 1
        operands = [
          {
            type       = "Term"
            inverted   = false
            occurrence = 1
            term = {
              word            = %q
              participantType = "External"
            }
          }
        ]
      }
    ]
  })
}
`, categoryName, word)
	)
	defer removeExportDir(exportDir)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: resources,
			},
			{
				Config: resources + generateSttExport(exportDir, []string{sttCategory.ResourceType},
					"genesyscloud_speechandtextanalytics_category.category"),
				Check: assertSttExportedBlock(exportDir, sttCategory.ResourceType, "name", categoryName, func(attrs map[string]interface{}) error {
					criteria, ok := attrs["criteria"]
					if !ok {
						return fmt.Errorf("criteria missing from exported category")
					}
					raw, err := json.Marshal(criteria)
					if err != nil {
						return err
					}
					if !strings.Contains(string(raw), word) {
						return fmt.Errorf("exported criteria does not contain the term %q: %s", word, raw)
					}
					return nil
				}),
			},
		},
		CheckDestroy: testVerifyExportsDestroyedFunc(exportDir),
	})
}

func TestAccResourceTfExportSttSentimentFeedback(t *testing.T) {
	testSetup(t)

	var (
		exportDir = testrunner.GetTestTempPath(".terraform" + uuid.NewString())
		phrase    = "tfacc export phrase " + strings.ReplaceAll(uuid.NewString(), "-", "")[:12]
		resources = fmt.Sprintf(`
resource "genesyscloud_speechandtextanalytics_sentimentfeedback" "feedback" {
  phrase         = %q
  dialect        = "en-US"
  feedback_value = "Positive"
}
`, phrase)
	)
	defer removeExportDir(exportDir)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: resources,
			},
			{
				Config: resources + generateSttExport(exportDir, []string{sttSentimentFeedback.ResourceType},
					"genesyscloud_speechandtextanalytics_sentimentfeedback.feedback"),
				Check: assertSttExportedBlock(exportDir, sttSentimentFeedback.ResourceType, "phrase", phrase, func(attrs map[string]interface{}) error {
					if attrs["dialect"] != "en-US" || attrs["feedback_value"] != "Positive" {
						return fmt.Errorf("unexpected exported attributes: %v", attrs)
					}
					return nil
				}),
			},
		},
		CheckDestroy: testVerifyExportsDestroyedFunc(exportDir),
	})
}

// Read-only: exports the org's existing settings and changes nothing.
func TestAccResourceTfExportSttRecordingSingletons(t *testing.T) {
	testSetup(t)

	exportDir := testrunner.GetTestTempPath(".terraform" + uuid.NewString())
	defer removeExportDir(exportDir)

	for _, resourceType := range []string{sttSettings.ResourceType, recordingSettings.ResourceType} {
		resourceType := resourceType
		t.Run(resourceType, func(t *testing.T) {
			dir := filepath.Join(exportDir, resourceType)
			resource.Test(t, resource.TestCase{
				PreCheck:          func() { util.TestAccPreCheck(t) },
				ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
				Steps: []resource.TestStep{
					{
						Config: generateSttExport(dir, []string{resourceType}),
						Check: func(_ *terraform.State) error {
							blocks, err := getResourceDefinition(filepath.Join(dir, defaultTfJSONFile), resourceType)
							if err != nil {
								return err
							}
							if len(blocks) != 1 {
								return fmt.Errorf("expected exactly 1 exported %s block, got %d", resourceType, len(blocks))
							}
							return nil
						},
					},
				},
				CheckDestroy: testVerifyExportsDestroyedFunc(dir),
			})
		})
	}
}

func generateSttExport(dir string, resourceTypes []string, dependsOn ...string) string {
	quoted := make([]string, len(resourceTypes))
	for i, rt := range resourceTypes {
		quoted[i] = strconv.Quote(rt)
	}
	dep := ""
	if len(dependsOn) > 0 {
		dep = fmt.Sprintf("\n  depends_on               = [%s]", strings.Join(dependsOn, ", "))
	}
	return fmt.Sprintf(`
resource "genesyscloud_tf_export" "export" {
  directory                = %q
  include_state_file       = false
  export_format            = "json"
  include_filter_resources = [%s]%s
}
`, dir, strings.Join(quoted, ", "), dep)
}

func assertSttExportedBlock(exportDir, resourceType, matchAttr, matchValue string, check func(map[string]interface{}) error) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		blocks, err := getResourceDefinition(filepath.Join(exportDir, defaultTfJSONFile), resourceType)
		if err != nil {
			return err
		}
		for _, raw := range blocks {
			var attrs map[string]interface{}
			if err := json.Unmarshal(*raw, &attrs); err != nil {
				return fmt.Errorf("failed to unmarshal %s block: %w", resourceType, err)
			}
			if attrs[matchAttr] != matchValue {
				continue
			}
			if check == nil {
				return nil
			}
			return check(attrs)
		}
		return fmt.Errorf("no exported %s block with %s = %q in %s", resourceType, matchAttr, matchValue, exportDir)
	}
}

func removeExportDir(path string) {
	if err := os.RemoveAll(path); err != nil {
		log.Printf("An error occurred while removing directory '%s': %s", path, err)
	}
}
