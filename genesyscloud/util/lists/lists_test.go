package lists

import (
	"fmt"
	"testing"
)

func TestChunkSliceMethod(t *testing.T) {
	slice := []string{"1", "2", "3", "4", "5", "6", "7", "8", "9"}
	chunkSize := 4

	chunks := ChunkStringSlice(slice, chunkSize)

	if len(chunks) != 3 {
		t.Errorf("Expected to receive %v chunks, got %v", 3, len(chunks))
	}

	chunk1Str := fmt.Sprintf("%v", chunks[0])
	expectedStr := fmt.Sprintf("%v", []string{"1", "2", "3", "4"})
	if chunk1Str != expectedStr {
		t.Errorf("Expected first chunk to look like %v, got %v", expectedStr, chunk1Str)
	}

	chunk2Str := fmt.Sprintf("%v", chunks[1])
	expectedStr = fmt.Sprintf("%v", []string{"5", "6", "7", "8"})
	if chunk2Str != expectedStr {
		t.Errorf("Expected first chunk to look like %v, got %v", expectedStr, chunk2Str)
	}

	chunk3Str := fmt.Sprintf("%v", chunks[2])
	expectedStr = fmt.Sprintf("%v", []string{"9"})
	if chunk3Str != expectedStr {
		t.Errorf("Expected first chunk to look like %v, got %v", expectedStr, chunk3Str)
	}
}
