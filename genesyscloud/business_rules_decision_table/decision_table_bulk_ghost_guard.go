package business_rules_decision_table

import (
	"encoding/json"
	"net/http"

	"github.com/mypurecloud/platform-client-sdk-go/v193/platformclientv2"
)

// decisionTableDuplicateRowCode is the API error code returned when a row being
// added duplicates a row that already exists on the table version.
const decisionTableDuplicateRowCode = "decision.table.duplicate.row"

func expectedChunkIndexRange(baseRowCount, chunkStart, chunkSize int) (first, last int) {
	return baseRowCount + chunkStart + 1, baseRowCount + chunkStart + chunkSize
}

// isGhostChunkDuplicate reports whether resp is a decision.table.duplicate.row
// error whose reported duplicate index falls within the 1-based index range that
// the current bulk chunk would occupy if appended after baseRowCount existing
// rows at offset chunkStart within the rows being added.
//
// This indicates the chunk was already created server-side by a prior attempt -
// typically a bulk POST that returned 504 at the gateway but succeeded on the
// backend, whose non-idempotent retry then collides with those ghost rows.
//
// Genuine config duplicates report the index of an earlier (pre-existing) row,
// which lies outside this range. RULE_PERSISTENCE_CONFLICT and other 409s are
// never skipped (keyed off error code, not status alone).
func isGhostChunkDuplicate(resp *platformclientv2.APIResponse, baseRowCount, chunkStart, chunkSize int) bool {
	if resp == nil || resp.StatusCode != http.StatusConflict || len(resp.RawBody) == 0 {
		return false
	}

	var apiErr struct {
		Code          string `json:"code"`
		MessageParams struct {
			DuplicateRows string `json:"duplicateRows"`
		} `json:"messageParams"`
	}
	if err := json.Unmarshal(resp.RawBody, &apiErr); err != nil {
		return false
	}
	if apiErr.Code != decisionTableDuplicateRowCode {
		return false
	}

	first, last := expectedChunkIndexRange(baseRowCount, chunkStart, chunkSize)

	var dups []struct {
		Id    string `json:"id"`
		Index int    `json:"index"`
	}
	if err := json.Unmarshal([]byte(apiErr.MessageParams.DuplicateRows), &dups); err != nil {
		return false
	}
	for _, d := range dups {
		if d.Index >= first && d.Index <= last {
			return true
		}
	}
	return false
}
