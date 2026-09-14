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

// TestUnitOrderUpdateWavesSplitsChain verifies a VIP->Standard / Standard->Premium
// chain is two bulk waves: the vacating row in the first request, the claiming row
// in a later request. Putting both in one bulk update 409s because the API is not
// atomic with respect to uniqueness (RULES-1907).
func TestUnitOrderUpdateWavesSplitsChain(t *testing.T) {
	const (
		mover   = "ab6f9490" // commercial -> passenger (claims "passenger")
		vacater = "9f6602b4" // passenger -> watercraft (frees "passenger")
	)

	oldRows := []interface{}{
		unitTestRowWithInput(mover, "commercial"),
		unitTestRowWithInput(vacater, "passenger"),
	}
	changes := RowChange{
		updates: []map[string]interface{}{
			unitTestRowWithInput(mover, "passenger"),
			unitTestRowWithInput(vacater, "watercraft"),
		},
	}

	waves, err := orderUpdateWavesForApply(changes, oldRows, unitTestOrderInputCols, unitTestOrderOutputCols)
	assert.NoError(t, err)
	assert.Len(t, waves, 2)
	assert.Len(t, waves[0], 1)
	assert.Equal(t, vacater, waves[0][0]["row_id"])
	assert.Len(t, waves[1], 1)
	assert.Equal(t, mover, waves[1][0]["row_id"])
}

// TestUnitOrderUpdatesReordersChain reproduces RULES-1907: one row moves onto the
// input value another row is moving away from. Applied in configuration order the
// updates would briefly duplicate "passenger" and fail with a 409; the ordering
// must apply the row moving away first, no matter what order the updates arrive in.
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
// end up with identical inputs - a real duplicate in the configuration, not a
// temporary one.
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

// unitTestRowIO builds a row with an explicit input and output value, for tests
// that need to vary the output independently of the input.
func unitTestRowIO(rowID, inputValue, outputValue string) map[string]interface{} {
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
					map[string]interface{}{"value": outputValue, "type": "string"},
				},
			},
		},
	}
}

// TestUnitOrderUpdatesMultiRowChain verifies a longer chain (3 rows) is ordered
// so each row's target value is free when it is applied, regardless of the order
// the updates arrive in.
func TestUnitOrderUpdatesMultiRowChain(t *testing.T) {
	oldRows := []interface{}{
		unitTestRowWithInput("row-a", "a"),
		unitTestRowWithInput("row-b", "b"),
		unitTestRowWithInput("row-c", "c"),
	}
	// row-a moves to a fresh value (frees "a"); row-b takes "a"; row-c takes "b".
	// Only safe order is a -> b -> c. Pass them in reverse to prove reordering.
	changes := RowChange{
		updates: []map[string]interface{}{
			unitTestRowWithInput("row-c", "b"),
			unitTestRowWithInput("row-b", "a"),
			unitTestRowWithInput("row-a", "z"),
		},
	}

	ordered, err := orderUpdatesForApply(changes, oldRows, unitTestOrderInputCols, unitTestOrderOutputCols)
	assert.NoError(t, err)
	assert.Len(t, ordered, 3)
	got := []string{ordered[0]["row_id"].(string), ordered[1]["row_id"].(string), ordered[2]["row_id"].(string)}
	assert.Equal(t, []string{"row-a", "row-b", "row-c"}, got, "each row must be applied after the row holding its target value moves")
}

// TestUnitOrderUpdatesDeleteFreesTuple verifies that a value vacated by a delete
// is treated as available: an update claiming it has no dependency (deletes are
// applied before updates by the caller) and does not error as a duplicate.
func TestUnitOrderUpdatesDeleteFreesTuple(t *testing.T) {
	oldRows := []interface{}{
		unitTestRowWithInput("row-a", "a"),
		unitTestRowWithInput("row-b", "b"),
	}
	// row-b is deleted (frees "b"); row-a moves onto "b".
	changes := RowChange{
		deletes: []string{"row-b"},
		updates: []map[string]interface{}{
			unitTestRowWithInput("row-a", "b"),
		},
	}

	ordered, err := orderUpdatesForApply(changes, oldRows, unitTestOrderInputCols, unitTestOrderOutputCols)
	assert.NoError(t, err)
	assert.Len(t, ordered, 1)
	assert.Equal(t, "row-a", ordered[0]["row_id"])
}

// TestUnitOrderUpdatesOutputOnlyChange verifies an update that changes only the
// output (input values unchanged) is not flagged as a self-collision or cycle.
func TestUnitOrderUpdatesOutputOnlyChange(t *testing.T) {
	oldRows := []interface{}{
		unitTestRowIO("row-a", "x", "out-1"),
		unitTestRowIO("row-b", "y", "out-1"),
	}
	// row-a keeps input "x", only its output changes.
	changes := RowChange{
		updates: []map[string]interface{}{
			unitTestRowIO("row-a", "x", "out-2"),
		},
	}

	ordered, err := orderUpdatesForApply(changes, oldRows, unitTestOrderInputCols, unitTestOrderOutputCols)
	assert.NoError(t, err)
	assert.Len(t, ordered, 1)
	assert.Equal(t, "row-a", ordered[0]["row_id"])
}
