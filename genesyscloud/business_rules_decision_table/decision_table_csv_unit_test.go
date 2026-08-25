package business_rules_decision_table

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v195/platformclientv2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_exporter"
)

func TestUnitPurgeAndReinjectRowId(t *testing.T) {
	userCSV := []byte("col_a::Equals,col_b\n1,2\n3,4\n")
	withRowId, err := reinjectEmptyRowIdColumn(userCSV)
	require.NoError(t, err)
	assert.Contains(t, string(withRowId), "rowId,col_a::Equals,col_b")

	purged, err := purgeRowIdColumn(withRowId)
	require.NoError(t, err)
	records, err := readCSVRecords(purged)
	require.NoError(t, err)
	require.Len(t, records, 3)
	assert.Equal(t, []string{"col_a::Equals", "col_b"}, records[0])
	assert.Equal(t, []string{"1", "2"}, records[1])

	exportedCSV := []byte("rowId,col_a::Equals,col_b\nuuid-1,a,b\n")
	purgedExport, err := purgeRowIdColumn(exportedCSV)
	require.NoError(t, err)
	assert.NotContains(t, string(purgedExport), "rowId")
	assert.Contains(t, string(purgedExport), "col_a::Equals,col_b")
}

func TestUnitImportDecisionTableRowsFromCSV_Success(t *testing.T) {
	dir := t.TempDir()
	csvPath := filepath.Join(dir, "rows.csv")
	require.NoError(t, os.WriteFile(csvPath, []byte("a::Equals,b\n1,2\n"), 0644))

	resource := ResourceBusinessRulesDecisionTable()
	d := schema.TestResourceDataRaw(t, resource.Schema, map[string]interface{}{
		"name":              "t",
		"division_id":       "d",
		"schema_id":         "s",
		"rows_csv_filepath": csvPath,
		"columns": []interface{}{
			map[string]interface{}{
				"inputs": []interface{}{
					map[string]interface{}{
						"expression": []interface{}{
							map[string]interface{}{
								"contractual": []interface{}{
									map[string]interface{}{"schema_property_key": "a"},
								},
								"comparator": "Equals",
							},
						},
						"defaults_to": []interface{}{
							map[string]interface{}{"value": "x"},
						},
					},
				},
				"outputs": []interface{}{
					map[string]interface{}{
						"value": []interface{}{
							map[string]interface{}{"schema_property_key": "b"},
						},
						"defaults_to": []interface{}{
							map[string]interface{}{"value": "y"},
						},
					},
				},
			},
		},
	})

	createCalls := 0
	publishCalls := 0
	var uploaded []byte
	proxy := &BusinessRulesDecisionTableProxy{
		createDecisionTableImportJobAttr: func(ctx context.Context, p *BusinessRulesDecisionTableProxy, tableId string, request *platformclientv2.Createdecisiontableimportjobrequest) (*platformclientv2.Decisiontableimportjob, *platformclientv2.APIResponse, error) {
			createCalls++
			id := "job-1"
			url := "https://example.com/upload"
			status := importStatusUploading
			headers := map[string]string{"x-test": "1"}
			return &platformclientv2.Decisiontableimportjob{
				Id:            &id,
				UploadUrl:     &url,
				UploadHeaders: &headers,
				Status:        &status,
			}, nil, nil
		},
		uploadDecisionTableImportFileAttr: func(ctx context.Context, p *BusinessRulesDecisionTableProxy, uploadUrl string, headers map[string]string, body []byte) error {
			uploaded = append([]byte(nil), body...)
			return nil
		},
		getDecisionTableImportJobAttr: func(ctx context.Context, p *BusinessRulesDecisionTableProxy, tableId string, importJobId string) (*platformclientv2.Decisiontableimportjob, *platformclientv2.APIResponse, error) {
			status := importStatusComplete
			return &platformclientv2.Decisiontableimportjob{Status: &status}, nil, nil
		},
		getBusinessRulesDecisionTableAttr: func(ctx context.Context, p *BusinessRulesDecisionTableProxy, tableId string) (*platformclientv2.Decisiontable, *platformclientv2.APIResponse, error) {
			v := 2
			return &platformclientv2.Decisiontable{
				Latest: &platformclientv2.Decisiontableversionentity{Version: &v},
			}, nil, nil
		},
		getBusinessRulesDecisionTableVersionAttr: func(ctx context.Context, p *BusinessRulesDecisionTableProxy, tableId string, versionNumber int) (*platformclientv2.Decisiontableversion, *platformclientv2.APIResponse, error) {
			status := "Draft"
			v := versionNumber
			return &platformclientv2.Decisiontableversion{Version: &v, Status: &status}, nil, nil
		},
		publishDecisionTableVersionAttr: func(ctx context.Context, p *BusinessRulesDecisionTableProxy, tableId string, version int) (*platformclientv2.APIResponse, error) {
			publishCalls++
			assert.Equal(t, 2, version)
			return &platformclientv2.APIResponse{StatusCode: 200}, nil
		},
	}

	require.NoError(t, importDecisionTableRowsFromCSV(context.Background(), d, proxy, "table-1", false))
	assert.Equal(t, 1, createCalls)
	assert.Equal(t, 1, publishCalls)
	uploadRecs, err := readCSVRecords(uploaded)
	require.NoError(t, err)
	assert.Equal(t, []string{"rowId", "a::Equals", "b"}, uploadRecs[0])
	hash := d.Get("rows_csv_content_hash").(string)
	require.NotEmpty(t, hash)
	assert.Equal(t, 1, d.Get("rows_record_count"))
}

func TestUnitImportDecisionTableRowsFromCSV_FailedJob(t *testing.T) {
	dir := t.TempDir()
	csvPath := filepath.Join(dir, "rows.csv")
	require.NoError(t, os.WriteFile(csvPath, []byte("a::Equals,b\n1,2\n"), 0644))

	resource := ResourceBusinessRulesDecisionTable()
	d := schema.TestResourceDataRaw(t, resource.Schema, map[string]interface{}{
		"name":              "t",
		"division_id":       "d",
		"schema_id":         "s",
		"rows_csv_filepath": csvPath,
		"columns": []interface{}{
			map[string]interface{}{
				"inputs": []interface{}{
					map[string]interface{}{
						"expression": []interface{}{
							map[string]interface{}{
								"contractual": []interface{}{
									map[string]interface{}{"schema_property_key": "a"},
								},
								"comparator": "Equals",
							},
						},
						"defaults_to": []interface{}{
							map[string]interface{}{"value": "x"},
						},
					},
				},
				"outputs": []interface{}{
					map[string]interface{}{
						"value": []interface{}{
							map[string]interface{}{"schema_property_key": "b"},
						},
						"defaults_to": []interface{}{
							map[string]interface{}{"value": "y"},
						},
					},
				},
			},
		},
	})

	proxy := &BusinessRulesDecisionTableProxy{
		createDecisionTableImportJobAttr: func(ctx context.Context, p *BusinessRulesDecisionTableProxy, tableId string, request *platformclientv2.Createdecisiontableimportjobrequest) (*platformclientv2.Decisiontableimportjob, *platformclientv2.APIResponse, error) {
			id := "job-fail"
			url := "https://example.com/upload"
			status := importStatusProcessing
			return &platformclientv2.Decisiontableimportjob{Id: &id, UploadUrl: &url, Status: &status}, nil, nil
		},
		uploadDecisionTableImportFileAttr: func(ctx context.Context, p *BusinessRulesDecisionTableProxy, uploadUrl string, headers map[string]string, body []byte) error {
			return nil
		},
		getDecisionTableImportJobAttr: func(ctx context.Context, p *BusinessRulesDecisionTableProxy, tableId string, importJobId string) (*platformclientv2.Decisiontableimportjob, *platformclientv2.APIResponse, error) {
			status := importStatusFailed
			return &platformclientv2.Decisiontableimportjob{Status: &status}, nil, nil
		},
	}

	err := importDecisionTableRowsFromCSV(context.Background(), d, proxy, "table-1", false)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Failed")
}

func TestUnitImportDecisionTableRowsFromCSV_UpdateCleansDraftOnPublishFail(t *testing.T) {
	dir := t.TempDir()
	csvPath := filepath.Join(dir, "rows.csv")
	require.NoError(t, os.WriteFile(csvPath, []byte("a::Equals,b\n1,2\n"), 0644))

	resource := ResourceBusinessRulesDecisionTable()
	d := schema.TestResourceDataRaw(t, resource.Schema, map[string]interface{}{
		"name":              "t",
		"division_id":       "d",
		"schema_id":         "s",
		"rows_csv_filepath": csvPath,
		"columns": []interface{}{
			map[string]interface{}{
				"inputs": []interface{}{
					map[string]interface{}{
						"expression": []interface{}{
							map[string]interface{}{
								"contractual": []interface{}{
									map[string]interface{}{"schema_property_key": "a"},
								},
								"comparator": "Equals",
							},
						},
						"defaults_to": []interface{}{
							map[string]interface{}{"value": "x"},
						},
					},
				},
				"outputs": []interface{}{
					map[string]interface{}{
						"value": []interface{}{
							map[string]interface{}{"schema_property_key": "b"},
						},
						"defaults_to": []interface{}{
							map[string]interface{}{"value": "y"},
						},
					},
				},
			},
		},
	})

	cleanupCalled := false
	proxy := &BusinessRulesDecisionTableProxy{
		createDecisionTableImportJobAttr: func(ctx context.Context, p *BusinessRulesDecisionTableProxy, tableId string, request *platformclientv2.Createdecisiontableimportjobrequest) (*platformclientv2.Decisiontableimportjob, *platformclientv2.APIResponse, error) {
			id := "job-2"
			url := "https://example.com/upload"
			return &platformclientv2.Decisiontableimportjob{Id: &id, UploadUrl: &url}, nil, nil
		},
		uploadDecisionTableImportFileAttr: func(ctx context.Context, p *BusinessRulesDecisionTableProxy, uploadUrl string, headers map[string]string, body []byte) error {
			return nil
		},
		getDecisionTableImportJobAttr: func(ctx context.Context, p *BusinessRulesDecisionTableProxy, tableId string, importJobId string) (*platformclientv2.Decisiontableimportjob, *platformclientv2.APIResponse, error) {
			status := importStatusComplete
			return &platformclientv2.Decisiontableimportjob{Status: &status}, nil, nil
		},
		getBusinessRulesDecisionTableAttr: func(ctx context.Context, p *BusinessRulesDecisionTableProxy, tableId string) (*platformclientv2.Decisiontable, *platformclientv2.APIResponse, error) {
			pub := 2
			latest := 3
			return &platformclientv2.Decisiontable{
				Published: &platformclientv2.Decisiontableversionentity{Version: &pub},
				Latest:    &platformclientv2.Decisiontableversionentity{Version: &latest},
			}, nil, nil
		},
		getBusinessRulesDecisionTableVersionAttr: func(ctx context.Context, p *BusinessRulesDecisionTableProxy, tableId string, versionNumber int) (*platformclientv2.Decisiontableversion, *platformclientv2.APIResponse, error) {
			status := "Draft"
			v := versionNumber
			return &platformclientv2.Decisiontableversion{Version: &v, Status: &status}, nil, nil
		},
		publishDecisionTableVersionAttr: func(ctx context.Context, p *BusinessRulesDecisionTableProxy, tableId string, version int) (*platformclientv2.APIResponse, error) {
			return nil, fmt.Errorf("publish boom")
		},
		deleteDecisionTableVersionAttr: func(ctx context.Context, p *BusinessRulesDecisionTableProxy, tableId string, version int) (*platformclientv2.APIResponse, error) {
			cleanupCalled = true
			assert.Equal(t, 3, version)
			return &platformclientv2.APIResponse{StatusCode: 200}, nil
		},
	}

	err := importDecisionTableRowsFromCSV(context.Background(), d, proxy, "table-1", true)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "publish")
	assert.True(t, cleanupCalled, "update path must delete draft version after publish failure")
}

func TestUnitDecisionTableRowsExporterResolver(t *testing.T) {
	dir := t.TempDir()
	exportDir := filepath.Join(dir, "export")
	require.NoError(t, os.MkdirAll(exportDir, 0755))

	proxy := &BusinessRulesDecisionTableProxy{
		clientConfig: &platformclientv2.Configuration{BasePath: "https://api.example.com", AccessToken: "tok"},
		createDecisionTableExportJobAttr: func(ctx context.Context, p *BusinessRulesDecisionTableProxy, tableId string, request *platformclientv2.Decisiontableexportjobrequest) (*platformclientv2.Decisiontableexportjob, *platformclientv2.APIResponse, error) {
			id := "exp-1"
			return &platformclientv2.Decisiontableexportjob{Id: &id}, nil, nil
		},
		getDecisionTableExportJobAttr: func(ctx context.Context, p *BusinessRulesDecisionTableProxy, tableId string, exportJobId string) (*platformclientv2.Decisiontableexportjob, *platformclientv2.APIResponse, error) {
			status := exportStatusComplete
			selfUri := "/api/v2/downloads/abc"
			return &platformclientv2.Decisiontableexportjob{
				Status:   &status,
				Download: &platformclientv2.Addressableentityref{SelfUri: &selfUri},
			}, nil, nil
		},
		downloadDecisionTableExportAttr: func(ctx context.Context, p *BusinessRulesDecisionTableProxy, downloadUri string) ([]byte, error) {
			return []byte("rowId,col_a::Equals,col_b\nid-1,x,y\n"), nil
		},
	}
	internalProxy = proxy
	defer func() { internalProxy = nil }()

	configMap := map[string]interface{}{
		"rows": []interface{}{map[string]interface{}{"inputs": []interface{}{}}},
	}
	ri := resourceExporter.ResourceInfo{
		BlockLabel: "example_table",
		State: &terraform.InstanceState{
			Attributes: map[string]string{
				"id":      "table-1",
				"version": "2",
			},
		},
	}

	meta := &provider.ProviderMeta{ClientConfig: proxy.clientConfig}
	err := DecisionTableRowsExporterResolver("table-1", exportDir, "rows", configMap, meta, ri)
	require.NoError(t, err)

	assert.Equal(t, filepath.Join("rows", "example_table.csv"), configMap["rows_csv_filepath"])
	_, hasRows := configMap["rows"]
	assert.False(t, hasRows)
	outPath := filepath.Join(exportDir, "rows", "example_table.csv")
	data, err := os.ReadFile(outPath)
	require.NoError(t, err)
	assert.NotContains(t, string(data), "rowId")
	assert.Contains(t, string(data), "col_a::Equals,col_b")
	assert.NotEmpty(t, ri.State.Attributes["rows_csv_content_hash"])
}
