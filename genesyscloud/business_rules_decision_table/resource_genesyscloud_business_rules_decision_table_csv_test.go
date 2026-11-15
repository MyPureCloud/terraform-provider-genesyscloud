package business_rules_decision_table

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"testing"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

// TestAccResourceBusinessRulesDecisionTableCSVHappyPath creates with rows_csv_filepath only
// (stringList columns + '||' delimiter), asserts hash/count/version, PlanOnly no-op, then
// rewrites CSV on disk and expects a new published version without recreate.
//
// Version after CSV create may be 1 or 2 depending on platform (empty draft + Replace import).
// Capture actual version; do not hardcode.
func TestAccResourceBusinessRulesDecisionTableCSVHappyPath(t *testing.T) {
	t.Parallel()

	var (
		tableResourceLabel  = "test-dt-csv-happy"
		schemaResourceLabel = "test-schema-csv-happy"
		tableName           = "TF Test DT CSV Happy-" + uuid.NewString()[:8]
		tableDesc           = "CSV happy path ACC"
		schemaName          = "TF Test Schema CSV Happy-" + uuid.NewString()[:8]
		schemaDescription   = "Schema for CSV happy path ACC"
		tableID             string
		createVersion       int
	)

	csvPath := filepath.Join(t.TempDir(), "happy_rows.csv")
	csvV1 := "customer_type::Equals,customer_tags::ContainsAny,skill,assigned_skills\nVIP,vip||premium,Premium Support,premium_support||escalation\n"
	csvV2 := "customer_type::Equals,customer_tags::ContainsAny,skill,assigned_skills\nVIP,vip||premium,Premium Support Updated,premium_support||escalation\n"
	if err := os.WriteFile(csvPath, []byte(csvV1), 0644); err != nil {
		t.Fatal(err)
	}
	csvPathHCL := filepath.ToSlash(csvPath)

	config := generateBusinessRulesSchemaResource(schemaResourceLabel, schemaName, schemaDescription) +
		generateHomeDivisionReference() +
		generateBusinessRulesDecisionTableResourceCSV(
			tableResourceLabel,
			tableName,
			tableDesc,
			"data.genesyscloud_auth_division_home.home.id",
			"genesyscloud_business_rules_schema."+schemaResourceLabel+".id",
			generateColumnsForCSVStringList(),
			csvPathHCL,
		)

	resourceAddr := "genesyscloud_business_rules_decision_table." + tableResourceLabel

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceAddr, "name", tableName),
					resource.TestCheckResourceAttr(resourceAddr, "rows_csv_filepath", csvPathHCL),
					resource.TestCheckResourceAttr(resourceAddr, "rows_record_count", "1"),
					resource.TestCheckResourceAttrSet(resourceAddr, "rows_csv_content_hash"),
					resource.TestCheckResourceAttrSet(resourceAddr, "version"),
					resource.TestCheckResourceAttrSet(resourceAddr, "id"),
					func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources[resourceAddr]
						if !ok {
							return fmt.Errorf("not found: %s", resourceAddr)
						}
						tableID = rs.Primary.ID
						if tableID == "" {
							return fmt.Errorf("empty id for %s", resourceAddr)
						}
						v, err := strconv.Atoi(rs.Primary.Attributes["version"])
						if err != nil || v < 1 {
							return fmt.Errorf("expected published version >= 1, got %q", rs.Primary.Attributes["version"])
						}
						createVersion = v
						if rowsCount, ok := rs.Primary.Attributes["rows.#"]; ok && rowsCount != "0" {
							return fmt.Errorf("expected nested rows absent/cleared in CSV mode, got rows.#=%s", rowsCount)
						}
						return nil
					},
				),
			},
			{
				Config:             config,
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
			{
				PreConfig: func() {
					if err := os.WriteFile(csvPath, []byte(csvV2), 0644); err != nil {
						t.Fatal(err)
					}
				},
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceAddr, "rows_record_count", "1"),
					resource.TestCheckResourceAttrSet(resourceAddr, "rows_csv_content_hash"),
					func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources[resourceAddr]
						if !ok {
							return fmt.Errorf("not found: %s", resourceAddr)
						}
						if rs.Primary.ID != tableID {
							return fmt.Errorf("resource recreated on CSV content update: was %s, now %s", tableID, rs.Primary.ID)
						}
						v, err := strconv.Atoi(rs.Primary.Attributes["version"])
						if err != nil {
							return fmt.Errorf("invalid version after CSV update: %q", rs.Primary.Attributes["version"])
						}
						if v <= createVersion {
							return fmt.Errorf("expected version to increment after CSV rewrite: was %d, now %d", createVersion, v)
						}
						return nil
					},
				),
			},
		},
		CheckDestroy: testVerifyBusinessRulesDecisionTablesDestroyed,
	})
}

// TestAccResourceBusinessRulesDecisionTableCSVQueueName creates a decision table whose CSV
// transfer_queue cell uses the queue friendly name (not UUID).
func TestAccResourceBusinessRulesDecisionTableCSVQueueName(t *testing.T) {
	t.Parallel()

	var (
		tableResourceLabel  = "test-dt-csv-queue"
		schemaResourceLabel = "test-schema-csv-queue"
		queueResourceLabel  = "test-queue-csv"
		tableName           = "TF Test DT CSV Queue-" + uuid.NewString()[:8]
		tableDesc           = "CSV queue name ACC"
		schemaName          = "TF Test Schema CSV Queue-" + uuid.NewString()[:8]
		schemaDescription   = "Schema for CSV queue name ACC"
		queueName           = "TF Test Queue CSV-" + uuid.NewString()[:8]
	)

	csvPath := filepath.Join(t.TempDir(), "queue_rows.csv")
	csvContent := fmt.Sprintf("customer_type::Equals,transfer_queue\nVIP,%s\n", queueName)
	if err := os.WriteFile(csvPath, []byte(csvContent), 0644); err != nil {
		t.Fatal(err)
	}
	csvPathHCL := filepath.ToSlash(csvPath)

	resourceAddr := "genesyscloud_business_rules_decision_table." + tableResourceLabel

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: generateBusinessRulesSchemaResource(schemaResourceLabel, schemaName, schemaDescription) +
					generateHomeDivisionReference() +
					generateRoutingQueueResource(queueResourceLabel, queueName) +
					generateBusinessRulesDecisionTableResourceCSV(
						tableResourceLabel,
						tableName,
						tableDesc,
						"data.genesyscloud_auth_division_home.home.id",
						"genesyscloud_business_rules_schema."+schemaResourceLabel+".id",
						generateColumnsForCSVQueue(queueResourceLabel),
						csvPathHCL,
					),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceAddr, "name", tableName),
					resource.TestCheckResourceAttr(resourceAddr, "rows_csv_filepath", csvPathHCL),
					resource.TestCheckResourceAttr(resourceAddr, "rows_record_count", "1"),
					resource.TestCheckResourceAttrSet(resourceAddr, "rows_csv_content_hash"),
					resource.TestCheckResourceAttrSet(resourceAddr, "id"),
					resource.TestCheckResourceAttrSet(resourceAddr, "version"),
				),
			},
		},
		CheckDestroy: testVerifyBusinessRulesDecisionTablesDestroyed,
	})
}

// TestAccResourceBusinessRulesDecisionTableCSVValidation covers CustomizeDiff / ExactlyOneOf
// failures for rows_csv_filepath (rowId column, wrong headers, header-only, both rows+CSV).
func TestAccResourceBusinessRulesDecisionTableCSVValidation(t *testing.T) {
	t.Parallel()

	var (
		schemaResourceLabel = "test-schema-csv-val"
		schemaName          = "tf_schema_csv_val_" + uuid.NewString()[:8]
		schemaDescription   = "Schema for CSV validation ACC"
	)

	tmpDir := t.TempDir()
	writeCSV := func(name, content string) string {
		p := filepath.Join(tmpDir, name)
		if err := os.WriteFile(p, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
		return filepath.ToSlash(p)
	}

	rowIdCSV := writeCSV("rowid.csv", "rowId,customer_type::Equals,skill\n,VIP,Premium Support\n")
	wrongHeaderCSV := writeCSV("wrong_headers.csv", "wrong_col,skill\nVIP,Premium Support\n")
	headerOnlyCSV := writeCSV("header_only.csv", "customer_type::Equals,skill\n")
	validCSV := writeCSV("valid.csv", "customer_type::Equals,skill\nVIP,Premium Support\n")

	baseDeps := generateBusinessRulesSchemaResource(schemaResourceLabel, schemaName, schemaDescription) +
		generateHomeDivisionReference()

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: baseDeps + generateBusinessRulesDecisionTableResourceCSV(
					"test-dt-csv-rowid",
					"tf_dt_csv_rowid_"+uuid.NewString()[:8],
					"CSV with rowId column",
					"data.genesyscloud_auth_division_home.home.id",
					"genesyscloud_business_rules_schema."+schemaResourceLabel+".id",
					generateMinimalColumnsForCSVMigration(),
					rowIdCSV,
				),
				ExpectError: regexp.MustCompile(`(?i)do not include a rowId`),
			},
			{
				Config: baseDeps + generateBusinessRulesDecisionTableResourceCSV(
					"test-dt-csv-headers",
					"tf_dt_csv_headers_"+uuid.NewString()[:8],
					"CSV with wrong headers",
					"data.genesyscloud_auth_division_home.home.id",
					"genesyscloud_business_rules_schema."+schemaResourceLabel+".id",
					generateMinimalColumnsForCSVMigration(),
					wrongHeaderCSV,
				),
				ExpectError: regexp.MustCompile(`(?i)(CSV header|expected)`),
			},
			{
				Config: baseDeps + generateBusinessRulesDecisionTableResourceCSV(
					"test-dt-csv-empty",
					"tf_dt_csv_empty_"+uuid.NewString()[:8],
					"CSV with header only",
					"data.genesyscloud_auth_division_home.home.id",
					"genesyscloud_business_rules_schema."+schemaResourceLabel+".id",
					generateMinimalColumnsForCSVMigration(),
					headerOnlyCSV,
				),
				ExpectError: regexp.MustCompile(`(?i)at least one data row`),
			},
			{
				Config: baseDeps + generateBusinessRulesDecisionTableResourceBothRowsAndCSV(
					"test-dt-csv-both",
					"tf_dt_csv_both_"+uuid.NewString()[:8],
					"Both nested rows and CSV filepath",
					"data.genesyscloud_auth_division_home.home.id",
					"genesyscloud_business_rules_schema."+schemaResourceLabel+".id",
					generateMinimalColumnsForCSVMigration(),
					generateMinimalRowsForCSVMigration(),
					validCSV,
				),
				ExpectError: regexp.MustCompile(`(?i)(Invalid combination|Exactly one of|"rows".*rows_csv_filepath|rows_csv_filepath.*"rows")`),
			},
		},
	})
}
