package business_rules_decision_table

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	platformclientv2 "github.com/mypurecloud/platform-client-sdk-go/v195/platformclientv2"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/files"
)

const (
	rowIdColumnName        = "rowId"
	importModeReplace      = "Replace"
	importStatusUploading  = "Uploading"
	importStatusProcessing = "Processing"
	importStatusComplete   = "Complete"
	importStatusFailed     = "Failed"
	importStatusCancelled  = "Cancelled"
	exportTypePopulated    = "Populated"
	exportFormatCsv        = "Csv"
	exportStatusComplete   = "Complete"
	exportStatusFailed     = "Failed"
	csvImportPollInterval  = 5 * time.Second
	csvExportPollInterval  = 5 * time.Second
)

// S3Enabled mirrors outbound_contact_list so ValidatePath / HashFileContent accept s3:// paths.
const S3Enabled = true

const csvInputHeaderSep = "::"

// purgeRowIdColumn removes the rowId column (column 0 when header is "rowId") from an exported CSV.
// Populated export already uses correct column headers (inputs as key::Comparator).
// On-disk CSV omits rowId; Replace import needs that header with empty cells — reinjectEmptyRowIdColumn adds it on upload.
func purgeRowIdColumn(data []byte) ([]byte, error) {
	records, err := readCSVRecords(data)
	if err != nil {
		return nil, err
	}
	if len(records) == 0 {
		return data, nil
	}
	if records[0][0] != rowIdColumnName {
		return data, nil
	}
	out := make([][]string, len(records))
	for i, row := range records {
		if len(row) == 0 {
			out[i] = row
			continue
		}
		out[i] = row[1:]
	}
	return writeCSVRecords(out)
}

// reinjectEmptyRowIdColumn prepends rowId with empty cell values.
// Replace import requires the rowId header; values must be empty (CX as Code always uses Replace).
// On-disk CSV headers must already match platform export (inputs key::Comparator, outputs bare key).
func reinjectEmptyRowIdColumn(data []byte) ([]byte, error) {
	records, err := readCSVRecords(data)
	if err != nil {
		return nil, err
	}
	if len(records) == 0 {
		return nil, fmt.Errorf("CSV is empty")
	}
	if len(records[0]) > 0 && records[0][0] == rowIdColumnName {
		return data, nil
	}
	out := make([][]string, len(records))
	for i, row := range records {
		if i == 0 {
			out[i] = append([]string{rowIdColumnName}, row...)
			continue
		}
		out[i] = append([]string{""}, row...)
	}
	return writeCSVRecords(out)
}

func readCSVRecords(data []byte) ([][]string, error) {
	r := csv.NewReader(bytes.NewReader(data))
	r.FieldsPerRecord = -1
	records, err := r.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to parse CSV: %w", err)
	}
	return records, nil
}

func writeCSVRecords(records [][]string) ([]byte, error) {
	var buf bytes.Buffer
	w := csv.NewWriter(&buf)
	if err := w.WriteAll(records); err != nil {
		return nil, fmt.Errorf("failed to write CSV: %w", err)
	}
	w.Flush()
	if err := w.Error(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// validateDecisionTableRowsCSV is a CustomizeDiff hook: ≥1 data row, no rowId col, headers match columns.
func validateDecisionTableRowsCSV(_ context.Context, diff *schema.ResourceDiff, _ interface{}) error {
	path, ok := diff.Get("rows_csv_filepath").(string)
	if !ok || path == "" {
		return nil
	}

	reader, file, err := files.DownloadOrOpenFile(context.Background(), path, S3Enabled)
	if err != nil {
		return fmt.Errorf("rows_csv_filepath: %w", err)
	}
	if file != nil {
		defer file.Close()
	}
	data, err := io.ReadAll(reader)
	if err != nil {
		return fmt.Errorf("rows_csv_filepath: failed to read CSV: %w", err)
	}

	records, err := readCSVRecords(data)
	if err != nil {
		return err
	}
	if len(records) < 2 {
		return fmt.Errorf("rows_csv_filepath: CSV must contain a header and at least one data row")
	}
	header := records[0]
	if len(header) > 0 && header[0] == rowIdColumnName {
		return fmt.Errorf("rows_csv_filepath: do not include a %s column in the on-disk CSV; Replace import requires empty %s cells and the provider reinjects that column on upload", rowIdColumnName, rowIdColumnName)
	}

	expected, err := expectedCSVHeadersFromDiff(diff)
	if err != nil {
		return err
	}
	if len(expected) == 0 {
		return nil
	}
	if len(header) != len(expected) {
		return fmt.Errorf("rows_csv_filepath: CSV has %d columns, expected %d (%s)", len(header), len(expected), strings.Join(expected, ", "))
	}
	for i := range expected {
		if header[i] != expected[i] {
			return fmt.Errorf("rows_csv_filepath: CSV header[%d]=%q, expected %q (platform export shape: inputs as schema_property_key::Comparator, then outputs as schema_property_key)", i, header[i], expected[i])
		}
	}
	return nil
}

// expectedCSVHeadersFromDiff returns CSV headers matching platform export: inputs key::Comparator, outputs bare key.
func expectedCSVHeadersFromDiff(diff *schema.ResourceDiff) ([]string, error) {
	columnsRaw, ok := diff.Get("columns").([]interface{})
	if !ok || len(columnsRaw) == 0 {
		return nil, nil
	}
	return csvHeadersFromColumnsRaw(columnsRaw)
}

func csvHeadersFromColumnsRaw(columnsRaw []interface{}) ([]string, error) {
	if len(columnsRaw) == 0 {
		return nil, nil
	}
	columnsMap, ok := columnsRaw[0].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("columns block is invalid")
	}
	var headers []string
	if inputs, ok := columnsMap["inputs"].([]interface{}); ok {
		for _, in := range inputs {
			key, comparator, err := inputColumnKeyAndComparator(in)
			if err != nil {
				return nil, err
			}
			headers = append(headers, key+csvInputHeaderSep+comparator)
		}
	}
	if outputs, ok := columnsMap["outputs"].([]interface{}); ok {
		for _, out := range outputs {
			key, err := outputColumnSchemaPropertyKey(out)
			if err != nil {
				return nil, err
			}
			headers = append(headers, key)
		}
	}
	return headers, nil
}

func inputColumnKeyAndComparator(col interface{}) (key string, comparator string, err error) {
	m, ok := col.(map[string]interface{})
	if !ok {
		return "", "", fmt.Errorf("input column is invalid")
	}
	exprList, ok := m["expression"].([]interface{})
	if !ok || len(exprList) == 0 {
		return "", "", fmt.Errorf("input column missing expression")
	}
	expr, ok := exprList[0].(map[string]interface{})
	if !ok {
		return "", "", fmt.Errorf("input column expression is invalid")
	}
	comparator, _ = expr["comparator"].(string)
	if comparator == "" {
		return "", "", fmt.Errorf("input column missing comparator")
	}
	contractualList, ok := expr["contractual"].([]interface{})
	if !ok || len(contractualList) == 0 {
		return "", "", fmt.Errorf("input column missing contractual")
	}
	key, err = nestedSchemaPropertyKey(contractualList[0])
	if err != nil {
		return "", "", err
	}
	return key, comparator, nil
}

func outputColumnSchemaPropertyKey(col interface{}) (string, error) {
	m, ok := col.(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("output column is invalid")
	}
	valueList, ok := m["value"].([]interface{})
	if !ok || len(valueList) == 0 {
		return "", fmt.Errorf("output column missing value")
	}
	return nestedSchemaPropertyKey(valueList[0])
}

func nestedSchemaPropertyKey(node interface{}) (string, error) {
	m, ok := node.(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("schema property node is invalid")
	}
	key, _ := m["schema_property_key"].(string)
	if key == "" {
		return "", fmt.Errorf("schema_property_key is required")
	}
	return key, nil
}

// importDecisionTableRowsFromCSV uploads the on-disk CSV via Replace import, waits for the
// resulting draft version, then publishes it — same publish step as nested rows.
//
// cleanupDraftOnFailure (update path) mirrors updateDecisionTableRows: on failure, delete the
// draft version so the prior published version remains. Create path leaves this false and
// deletes the whole table via rollbackDecisionTable instead.
func importDecisionTableRowsFromCSV(ctx context.Context, d *schema.ResourceData, proxy *BusinessRulesDecisionTableProxy, tableId string, cleanupDraftOnFailure bool) error {
	path, ok := d.Get("rows_csv_filepath").(string)
	if !ok || path == "" {
		return fmt.Errorf("rows_csv_filepath is required for CSV import")
	}

	hash, err := files.HashFileContent(ctx, path, S3Enabled)
	if err != nil {
		return fmt.Errorf("failed to hash rows CSV: %w", err)
	}

	publishedBefore := 0
	if cleanupDraftOnFailure {
		if v, ok, err := publishedDecisionTableVersion(ctx, proxy, tableId); err != nil {
			return err
		} else if ok {
			publishedBefore = v
		}
	}

	reader, file, err := files.DownloadOrOpenFile(ctx, path, S3Enabled)
	if err != nil {
		return fmt.Errorf("failed to open rows CSV: %w", err)
	}
	if file != nil {
		defer file.Close()
	}
	raw, err := io.ReadAll(reader)
	if err != nil {
		return fmt.Errorf("failed to read rows CSV: %w", err)
	}

	uploadBody, err := reinjectEmptyRowIdColumn(raw)
	if err != nil {
		return err
	}

	fileName := filepath.Base(path)
	if !strings.HasSuffix(strings.ToLower(fileName), ".csv") {
		fileName = fileName + ".csv"
	}
	importMode := importModeReplace
	job, resp, err := proxy.createDecisionTableImportJob(ctx, tableId, &platformclientv2.Createdecisiontableimportjobrequest{
		ImportMode: &importMode,
		FileName:   &fileName,
	})
	if err != nil {
		return fmt.Errorf("failed to create decision table import job: %w", err)
	}
	if job == nil || job.Id == nil {
		return fmt.Errorf("import job response missing id (resp=%v)", resp)
	}
	if job.UploadUrl == nil || *job.UploadUrl == "" {
		return fmt.Errorf("import job %s missing uploadUrl", *job.Id)
	}

	headers := map[string]string{}
	if job.UploadHeaders != nil {
		headers = *job.UploadHeaders
	}
	if err := proxy.uploadImportFile(ctx, *job.UploadUrl, headers, uploadBody); err != nil {
		return err
	}

	// Platform starts Processing after the CSV lands on the upload URL.
	// PATCH .../imports/{id} only accepts status=Cancelled (cancel), not Processing.
	if err := pollImportJob(ctx, proxy, tableId, *job.Id); err != nil {
		return err
	}

	draftVersion, err := latestDecisionTableVersion(ctx, proxy, tableId)
	if err != nil {
		return err
	}

	// Same as nested updateDecisionTableRows: clean up the draft version on failure.
	versionCreated := cleanupDraftOnFailure && draftVersion > publishedBefore
	if versionCreated {
		defer func() {
			if versionCreated {
				log.Printf("Cleaning up version %d due to failure", draftVersion)
				if cleanupErr := cleanupVersion(ctx, proxy, tableId, draftVersion); cleanupErr != nil {
					log.Printf("Warning: Failed to cleanup version %d: %s", draftVersion, cleanupErr)
				}
			}
		}()
	}

	log.Printf("Waiting for version %d to reach draft status", draftVersion)
	if err := waitForVersionDraftStatus(ctx, proxy, tableId, draftVersion); err != nil {
		return fmt.Errorf("failed to wait for version to reach draft status: %s", err)
	}
	log.Printf("Version %d is now in draft status", draftVersion)

	log.Printf("Publishing version %d", draftVersion)
	if err := publishDecisionTableVersion(ctx, proxy, tableId, draftVersion); err != nil {
		return fmt.Errorf("failed to publish version %d: %s", draftVersion, err)
	}
	versionCreated = false

	recordCount, err := countCSVDataRows(raw)
	if err != nil {
		return err
	}
	_ = d.Set("rows_csv_content_hash", hash)
	_ = d.Set("rows_record_count", recordCount)
	log.Printf("Successfully updated rows for decision table %s version %d (CSV import)", tableId, draftVersion)
	return nil
}

func publishedDecisionTableVersion(ctx context.Context, proxy *BusinessRulesDecisionTableProxy, tableId string) (version int, ok bool, err error) {
	table, _, err := proxy.getBusinessRulesDecisionTable(ctx, tableId)
	if err != nil {
		return 0, false, fmt.Errorf("failed to get decision table %s: %w", tableId, err)
	}
	if table == nil || table.Published == nil || table.Published.Version == nil {
		return 0, false, nil
	}
	return int(*table.Published.Version), true, nil
}

func latestDecisionTableVersion(ctx context.Context, proxy *BusinessRulesDecisionTableProxy, tableId string) (int, error) {
	table, _, err := proxy.getBusinessRulesDecisionTable(ctx, tableId)
	if err != nil {
		return 0, fmt.Errorf("failed to get decision table %s after import: %w", tableId, err)
	}
	if table == nil || table.Latest == nil || table.Latest.Version == nil {
		return 0, fmt.Errorf("decision table %s missing latest version after import", tableId)
	}
	return int(*table.Latest.Version), nil
}

func countCSVDataRows(data []byte) (int, error) {
	records, err := readCSVRecords(data)
	if err != nil {
		return 0, err
	}
	if len(records) < 2 {
		return 0, fmt.Errorf("CSV must contain at least one data row")
	}
	return len(records) - 1, nil
}

func pollImportJob(ctx context.Context, proxy *BusinessRulesDecisionTableProxy, tableId, jobId string) error {
	for {
		if err := ctx.Err(); err != nil {
			return fmt.Errorf("timed out waiting for import job %s: %w", jobId, err)
		}
		job, _, err := proxy.getDecisionTableImportJob(ctx, tableId, jobId)
		if err != nil {
			return fmt.Errorf("failed to get import job %s: %w", jobId, err)
		}
		if job == nil || job.Status == nil {
			return fmt.Errorf("import job %s returned empty status", jobId)
		}
		switch *job.Status {
		case importStatusComplete:
			return nil
		case importStatusFailed, importStatusCancelled:
			msg := *job.Status
			if job.VarError != nil {
				msg = fmt.Sprintf("%s: %+v", msg, job.VarError)
			}
			return fmt.Errorf("import job %s ended with status %s", jobId, msg)
		}
		select {
		case <-ctx.Done():
			return fmt.Errorf("timed out waiting for import job %s: %w", jobId, ctx.Err())
		case <-time.After(csvImportPollInterval):
		}
	}
}

func pollExportJob(ctx context.Context, proxy *BusinessRulesDecisionTableProxy, tableId, jobId string) (*platformclientv2.Decisiontableexportjob, error) {
	for {
		if err := ctx.Err(); err != nil {
			return nil, fmt.Errorf("timed out waiting for export job %s: %w", jobId, err)
		}
		job, _, err := proxy.getDecisionTableExportJob(ctx, tableId, jobId)
		if err != nil {
			return nil, fmt.Errorf("failed to get export job %s: %w", jobId, err)
		}
		if job == nil || job.Status == nil {
			return nil, fmt.Errorf("export job %s returned empty status", jobId)
		}
		switch *job.Status {
		case exportStatusComplete:
			return job, nil
		case exportStatusFailed:
			msg := *job.Status
			if job.VarError != nil {
				msg = fmt.Sprintf("%s: %+v", msg, job.VarError)
			}
			return nil, fmt.Errorf("export job %s ended with status %s", jobId, msg)
		}
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("timed out waiting for export job %s: %w", jobId, ctx.Err())
		case <-time.After(csvExportPollInterval):
		}
	}
}

// DecisionTableRowsExporterResolver writes a purged Populated CSV and sets rows_csv_filepath.
func DecisionTableRowsExporterResolver(resourceId, exportDirectory, subDirectory string, configMap map[string]interface{}, meta interface{}, resource resourceExporter.ResourceInfo) error {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getBusinessRulesDecisionTableProxy(sdkConfig)

	tableId := resource.State.Attributes["id"]
	if tableId == "" {
		tableId = resourceId
	}
	exportFileName := fmt.Sprintf("%s.csv", resource.BlockLabel)
	fullDirectoryPath := filepath.Join(exportDirectory, subDirectory)
	if err := os.MkdirAll(fullDirectoryPath, os.ModePerm); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", fullDirectoryPath, err)
	}

	ctx := context.Background()
	exportType := exportTypePopulated
	format := exportFormatCsv
	req := &platformclientv2.Decisiontableexportjobrequest{
		ExportType: &exportType,
		Format:     &format,
	}
	if vStr, ok := resource.State.Attributes["version"]; ok && vStr != "" {
		if v, err := strconv.Atoi(vStr); err == nil {
			req.TableVersion = &v
		}
	}

	job, _, err := proxy.createDecisionTableExportJob(ctx, tableId, req)
	if err != nil {
		return fmt.Errorf("failed to create decision table export job: %w", err)
	}
	if job == nil || job.Id == nil {
		return fmt.Errorf("export job response missing id")
	}

	completed, err := pollExportJob(ctx, proxy, tableId, *job.Id)
	if err != nil {
		return err
	}
	if completed.Download == nil || completed.Download.SelfUri == nil || *completed.Download.SelfUri == "" {
		return fmt.Errorf("export job %s completed without download SelfUri", *job.Id)
	}

	raw, err := proxy.downloadExportFile(ctx, *completed.Download.SelfUri)
	if err != nil {
		return err
	}
	purged, err := purgeRowIdColumn(raw)
	if err != nil {
		return err
	}

	fullCurrentPath := filepath.Join(fullDirectoryPath, exportFileName)
	if err := os.WriteFile(fullCurrentPath, purged, 0644); err != nil {
		return fmt.Errorf("failed to write exported CSV %s: %w", fullCurrentPath, err)
	}

	fullRelativePath := filepath.Join(subDirectory, exportFileName)
	configMap["rows_csv_filepath"] = fullRelativePath
	delete(configMap, "rows")
	delete(configMap, "rows_csv_content_hash")
	delete(configMap, "rows_record_count")

	hash, err := files.HashFileContent(ctx, fullCurrentPath, S3Enabled)
	if err != nil {
		return fmt.Errorf("failed to hash exported CSV: %w", err)
	}
	recordCount, err := countCSVDataRows(purged)
	if err != nil {
		return err
	}

	resource.State.Attributes["rows_csv_filepath"] = fullRelativePath
	resource.State.Attributes["rows_csv_content_hash"] = hash
	resource.State.Attributes["rows_record_count"] = strconv.Itoa(recordCount)
	return nil
}
