package business_rules_decision_table

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// unitTestRowWithInput builds a single-input/single-output row map in provider
// form for the order/collision tests. The input literal value is what the API
// enforces uniqueness on, so it is what the ordering logic keys off of.
func unitTestRowWithInput(rowID, inputValue string) map[string]interface{} {
	return map[string]interface{}{
		"row_id": rowID,
		"inputs": []interface{}{
			map[string]interface{}{
				"literal": []interface{}{
					map[string]interface{}{"value": inputValue, "type": "string"},
				},
			},
		},
		"outputs": []interface{}{
			map[string]interface{}{
				"literal": []interface{}{
					map[string]interface{}{"value": "queue-1", "type": "string"},
				},
			},
		},
	}
}

var unitTestOrderInputCols = []string{"input-col-1"}
var unitTestOrderOutputCols = []string{"output-col-1"}

// TestUnitOrderUpdatesReordersChain reproduces RULES-1907: one row moves onto the
// input value another row is vacating. Applied in config order the update PUTs
// would transiently duplicate "passenger" and 409; the ordering must apply the
// vacating row first regardless of the order the updates arrive in.
func TestUnitOrderUpdatesReordersChain(t *testing.T) {
	const (
		mover   = "ab6f9490" // commercial -> passenger (claims "passenger")
		vacater = "9f6602b4" // passenger -> watercraft (frees "passenger")
	)

	oldRows := []interface{}{
		unitTestRowWithInput(mover, "commercial"),
		unitTestRowWithInput(vacater, "passenger"),
	}

	// Deliberately pass updates in the hazardous order (mover first).
	changes := RowChange{
		updates: []map[string]interface{}{
			unitTestRowWithInput(mover, "passenger"),
			unitTestRowWithInput(vacater, "watercraft"),
		},
	}

	ordered, err := orderUpdatesForApply(changes, oldRows, unitTestOrderInputCols, unitTestOrderOutputCols)
	assert.NoError(t, err)
	assert.Len(t, ordered, 2)
	assert.Equal(t, vacater, ordered[0]["row_id"], "row vacating 'passenger' must be updated first")
	assert.Equal(t, mover, ordered[1]["row_id"], "row claiming 'passenger' must be updated after it is freed")
}

// TestUnitOrderUpdatesIsDeterministic verifies independent updates come back in a
// stable (row_id-sorted) order so plans are reproducible run to run.
func TestUnitOrderUpdatesIsDeterministic(t *testing.T) {
	oldRows := []interface{}{
		unitTestRowWithInput("row-c", "c-old"),
		unitTestRowWithInput("row-a", "a-old"),
		unitTestRowWithInput("row-b", "b-old"),
	}
	changes := RowChange{
		updates: []map[string]interface{}{
			unitTestRowWithInput("row-c", "c-new"),
			unitTestRowWithInput("row-a", "a-new"),
			unitTestRowWithInput("row-b", "b-new"),
		},
	}

	ordered, err := orderUpdatesForApply(changes, oldRows, unitTestOrderInputCols, unitTestOrderOutputCols)
	assert.NoError(t, err)
	got := []string{ordered[0]["row_id"].(string), ordered[1]["row_id"].(string), ordered[2]["row_id"].(string)}
	assert.Equal(t, []string{"row-a", "row-b", "row-c"}, got)
}

// TestUnitOrderUpdatesGenuineDuplicateTargets errors when two updated rows would
// end up with identical inputs - a real config duplicate, not a transient one.
func TestUnitOrderUpdatesGenuineDuplicateTargets(t *testing.T) {
	oldRows := []interface{}{
		unitTestRowWithInput("row-a", "a-old"),
		unitTestRowWithInput("row-b", "b-old"),
	}
	changes := RowChange{
		updates: []map[string]interface{}{
			unitTestRowWithInput("row-a", "same"),
			unitTestRowWithInput("row-b", "same"),
		},
	}

	_, err := orderUpdatesForApply(changes, oldRows, unitTestOrderInputCols, unitTestOrderOutputCols)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "duplicate row in the desired configuration")
}

// TestUnitOrderUpdatesTargetsUnchangedRow errors when an updated row targets the
// input value held by a row that is not changing (permanent duplicate).
func TestUnitOrderUpdatesTargetsUnchangedRow(t *testing.T) {
	oldRows := []interface{}{
		unitTestRowWithInput("row-a", "a-old"),
		unitTestRowWithInput("row-unchanged", "taken"),
	}
	// Only row-a is updated; row-unchanged stays put holding "taken".
	changes := RowChange{
		updates: []map[string]interface{}{
			unitTestRowWithInput("row-a", "taken"),
		},
	}

	_, err := orderUpdatesForApply(changes, oldRows, unitTestOrderInputCols, unitTestOrderOutputCols)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unchanged row")
}

// TestUnitOrderUpdatesDetectsCycle errors on a true swap (A<->B exchange input
// values) which cannot be linearized with in-place single-row PUTs.
func TestUnitOrderUpdatesDetectsCycle(t *testing.T) {
	oldRows := []interface{}{
		unitTestRowWithInput("row-a", "p"),
		unitTestRowWithInput("row-b", "q"),
	}
	changes := RowChange{
		updates: []map[string]interface{}{
			unitTestRowWithInput("row-a", "q"),
			unitTestRowWithInput("row-b", "p"),
		},
	}

	_, err := orderUpdatesForApply(changes, oldRows, unitTestOrderInputCols, unitTestOrderOutputCols)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cyclic")
}
