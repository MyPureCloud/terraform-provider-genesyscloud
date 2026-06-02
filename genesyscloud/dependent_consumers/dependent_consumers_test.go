package dependent_consumers

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSetDependentObjectMaps(t *testing.T) {
	result := SetDependentObjectMaps()

	assert.NotNil(t, result)
	assert.Greater(t, len(result), 0)

	// Test key mappings
	assert.Equal(t, "genesyscloud_routing_language", result["ACDLANGUAGE"])
	assert.Equal(t, "genesyscloud_routing_skill", result["ACDSKILL"])
	assert.Equal(t, "genesyscloud_flow", result["BOTFLOW"])
	assert.Equal(t, "genesyscloud_routing_queue", result["QUEUE"])
	assert.Equal(t, "genesyscloud_user", result["USER"])
}

func TestSetDependentObjectMaps_Idempotent(t *testing.T) {
	result1 := SetDependentObjectMaps()
	result2 := SetDependentObjectMaps()

	assert.Equal(t, result1, result2)
	assert.Equal(t, len(result1), len(result2))
}

func TestSetFlowTypeObjectMaps(t *testing.T) {
	result := SetFlowTypeObjectMaps()

	assert.NotNil(t, result)
	assert.Greater(t, len(result), 0)

	// Test key mappings
	assert.Equal(t, "BOTFLOW", result["BOT"])
	assert.Equal(t, "INBOUNDCALLFLOW", result["INBOUNDCALL"])
	assert.Equal(t, "INQUEUEEMAILFLOW", result["INQUEUEEMAIL"])
	assert.Equal(t, "VOICEFLOW", result["VOICE"])
	assert.Equal(t, "WORKITEMFLOW", result["WORKITEM"])
}

func TestSetFlowTypeObjectMaps_Idempotent(t *testing.T) {
	result1 := SetFlowTypeObjectMaps()
	result2 := SetFlowTypeObjectMaps()

	assert.Equal(t, result1, result2)
	assert.Equal(t, len(result1), len(result2))
}

func TestSetFlowTypeObjectMaps_CorrectInqueueEmailMapping(t *testing.T) {
	result := SetFlowTypeObjectMaps()

	// Verify the correct mapping
	assert.Equal(t, "INQUEUEEMAILFLOW", result["INQUEUEEMAIL"])
}

func TestDependentObjectMaps_ThreadSafe(t *testing.T) {
	// Run concurrent calls to verify sync.Once works
	done := make(chan bool)

	for i := 0; i < 10; i++ {
		go func() {
			result := SetDependentObjectMaps()
			assert.NotNil(t, result)
			done <- true
		}()
	}

	for i := 0; i < 10; i++ {
		<-done
	}
}

func TestFlowTypeObjectMaps_ThreadSafe(t *testing.T) {
	// Run concurrent calls to verify sync.Once works
	done := make(chan bool)

	for i := 0; i < 10; i++ {
		go func() {
			result := SetFlowTypeObjectMaps()
			assert.NotNil(t, result)
			done <- true
		}()
	}

	for i := 0; i < 10; i++ {
		<-done
	}
}
