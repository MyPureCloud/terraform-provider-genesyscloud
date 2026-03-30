package business_rules_decision_table

import (
	"context"
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/mypurecloud/platform-client-sdk-go/v179/platformclientv2"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/chunks"
)

// TestUnitBulkDecisionTableRowChunkShapeMatchesChunksPackage verifies that the API bulk
// limits (15 / 15 / 49) partition row slices the same way as genesyscloud/util/chunks.ChunkBy,
// matching the style of chunk tests elsewhere in the repo (small chunk sizes with fixed slices).
func TestUnitBulkDecisionTableRowChunkShapeMatchesChunksPackage(t *testing.T) {
	t.Parallel()

	type tc struct {
		nRows  int
		limit  int
		wantN  int
		name   string
	}
	cases := []tc{
		{1, 15, 1, "single under add limit"},
		{15, 15, 1, "exactly add limit"},
		{16, 15, 2, "one over add limit"},
		{30, 15, 2, "two full add chunks"},
		{31, 15, 3, "two full and one partial"},
		{1, 49, 1, "remove single"},
		{49, 49, 1, "exactly remove limit"},
		{50, 49, 2, "one over remove limit"},
	}
	for _, c := range cases {
		rows := make([]string, c.nRows)
		for i := range rows {
			rows[i] = "x"
		}
		got := chunks.ChunkBy(rows, c.limit)
		if len(got) != c.wantN {
			t.Fatalf("%s: n=%d limit=%d want %d chunks got %d", c.name, c.nRows, c.limit, c.wantN, len(got))
		}
	}
}

// TestUnitBulkApplyRowChangesMultiChunkWithSmallLimits uses bulkChunkLimitsOverride with small
// add/remove sizes (like util/chunks tests using chunkSize 3–4) so we assert multi-chunk
// behavior without building 16+ real Terraform rows.
func TestUnitBulkApplyRowChangesMultiChunkWithSmallLimits(t *testing.T) {
	tId := uuid.NewString()
	tSchemaId := uuid.NewString()

	tColumns := &platformclientv2.Decisiontablecolumns{
		Inputs: &[]platformclientv2.Decisiontableinputcolumn{
			{
				Id: platformclientv2.String("input-column-id-1"),
				DefaultsTo: &platformclientv2.Decisiontablecolumndefaultrowvalue{
					Value: platformclientv2.String("Standard"),
				},
				Expression: &platformclientv2.Decisiontableinputcolumnexpression{
					Contractual: func() **platformclientv2.Contractual {
						contractual := &platformclientv2.Contractual{
							SchemaPropertyKey: platformclientv2.String("customer_type"),
						}
						return &contractual
					}(),
					Comparator: platformclientv2.String("Equals"),
				},
			},
		},
		Outputs: &[]platformclientv2.Decisiontableoutputcolumn{
			{
				Id: platformclientv2.String("output-column-id-1"),
				DefaultsTo: &platformclientv2.Decisiontablecolumndefaultrowvalue{
					Value: platformclientv2.String("genesyscloud_routing_queue.output_queue.id"),
				},
				Value: &platformclientv2.Outputvalue{
					SchemaPropertyKey: platformclientv2.String("transfer_queue"),
					Properties: &[]platformclientv2.Outputvalue{
						{
							SchemaPropertyKey: platformclientv2.String("queue"),
							Properties: &[]platformclientv2.Outputvalue{
								{SchemaPropertyKey: platformclientv2.String("id")},
							},
						},
					},
				},
			},
			{
				Id: platformclientv2.String("output-column-id-2"),
				DefaultsTo: &platformclientv2.Decisiontablecolumndefaultrowvalue{
					Value: platformclientv2.String("VIP"),
				},
				Value: &platformclientv2.Outputvalue{
					SchemaPropertyKey: platformclientv2.String("escalation_level"),
				},
			},
		},
	}

	minRow := func(suffix string) map[string]interface{} {
		return map[string]interface{}{
			"inputs": []interface{}{
				map[string]interface{}{
					"schema_property_key": "customer_type",
					"comparator":          "Equals",
					"literal": []interface{}{
						map[string]interface{}{"value": "queue-" + suffix, "type": "string"},
					},
				},
			},
			"outputs": []interface{}{
				map[string]interface{}{
					"schema_property_key": "transfer_queue",
					"literal": []interface{}{
						map[string]interface{}{"value": "tq-" + suffix, "type": "string"},
					},
				},
				map[string]interface{}{
					"schema_property_key": "escalation_level",
					"literal": []interface{}{
						map[string]interface{}{"value": "lvl-" + suffix, "type": "string"},
					},
				},
			},
		}
	}

	var bulkAddLens []int
	var bulkRemoveLens []int

	proxy := &BusinessRulesDecisionTableProxy{}
	proxy.getBusinessRulesDecisionTableVersionAttr = func(ctx context.Context, p *BusinessRulesDecisionTableProxy, tableId string, versionNumber int) (*platformclientv2.Decisiontableversion, *platformclientv2.APIResponse, error) {
		return &platformclientv2.Decisiontableversion{
			Id:      &tId,
			Version: platformclientv2.Int(versionNumber),
			Contract: &platformclientv2.Decisiontablecontract{
				ParentSchema: &platformclientv2.Domainentityref{Id: &tSchemaId},
			},
			Columns: tColumns,
		}, &platformclientv2.APIResponse{StatusCode: http.StatusOK}, nil
	}
	proxy.bulkAddDecisionTableRowsAttr = func(ctx context.Context, p *BusinessRulesDecisionTableProxy, tableId string, version int, rows []platformclientv2.Createdecisiontablerowrequest) (*platformclientv2.APIResponse, error) {
		bulkAddLens = append(bulkAddLens, len(rows))
		return &platformclientv2.APIResponse{StatusCode: http.StatusOK}, nil
	}
	proxy.bulkRemoveDecisionTableRowsAttr = func(ctx context.Context, p *BusinessRulesDecisionTableProxy, tableId string, version int, rowIds []string) (*platformclientv2.APIResponse, error) {
		bulkRemoveLens = append(bulkRemoveLens, len(rowIds))
		return &platformclientv2.APIResponse{StatusCode: http.StatusNoContent}, nil
	}
	proxy.bulkUpdateDecisionTableRowsAttr = func(ctx context.Context, p *BusinessRulesDecisionTableProxy, tableId string, version int, rows []bulkUpdateDecisionTableRowBody) (*platformclientv2.APIResponse, error) {
		t.Fatal("bulk update should not be called in this test")
		return nil, nil
	}

	// Small limits: 5 adds → expect chunk sizes 2,2,1 ; 7 deletes → 3,3,1
	bulkChunkLimitsOverride = &struct {
		Add, Update, Remove int
	}{Add: 2, Update: 15, Remove: 3}
	defer func() { bulkChunkLimitsOverride = nil }()

	adds := []map[string]interface{}{
		minRow("1"), minRow("2"), minRow("3"), minRow("4"), minRow("5"),
	}
	deletes := []string{"d1", "d2", "d3", "d4", "d5", "d6", "d7"}

	err := applyRowChanges(context.Background(), proxy, tId, 2, RowChange{
		deletes: deletes,
		adds:    adds,
		updates: nil,
	})
	if err != nil {
		t.Fatalf("applyRowChanges: %v", err)
	}

	wantAddChunks := [][]int{{2}, {2}, {1}}
	if len(bulkAddLens) != len(wantAddChunks) {
		t.Fatalf("bulk add calls: got %v (len=%d), want %d calls", bulkAddLens, len(bulkAddLens), len(wantAddChunks))
	}
	for i, w := range wantAddChunks {
		if bulkAddLens[i] != w[0] {
			t.Fatalf("bulk add call %d: got len %d want %d", i, bulkAddLens[i], w[0])
		}
	}

	wantRemove := []int{3, 3, 1}
	if len(bulkRemoveLens) != len(wantRemove) {
		t.Fatalf("bulk remove lens: got %v want %v", bulkRemoveLens, wantRemove)
	}
	for i, w := range wantRemove {
		if bulkRemoveLens[i] != w {
			t.Fatalf("bulk remove call %d: got %d want %d", i, bulkRemoveLens[i], w)
		}
	}
}
