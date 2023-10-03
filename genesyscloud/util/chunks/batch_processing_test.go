package chunks

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
)

func TestChunkByMethod(t *testing.T) {
	slice := []string{"1", "2", "3", "4", "5", "6", "7", "8", "9"}
	chunkSize := 4

	chunks := ChunkBy(slice, chunkSize)

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

func TestChunkItems(t *testing.T) {
	items := []int{1, 2, 3, 4, 5, 6, 7, 8, 9}
	chunkSize := 3

	mapBuilder := func(item int) int {
		return item * 2
	}

	expectedChunks := [][]int{
		{2, 4, 6},
		{8, 10, 12},
		{14, 16, 18},
	}
	chunks := ChunkItems(items, mapBuilder, chunkSize)

	if !reflect.DeepEqual(chunks, expectedChunks) {
		t.Errorf("Chunks do not match. Expected: %v, Got: %v", expectedChunks, chunks)
	}
}

func TestProcessChunks(t *testing.T) {
	chunks := []int{1, 2, 3, 4, 5}

	// Define a mock chunkProcessor function
	mockChunkProcessor := func(chunk int) diag.Diagnostics {
		if chunk == 3 {
			return diag.Errorf("Error processing chunk %d", chunk)
		}
		return nil
	}

	err := ProcessChunks(chunks, mockChunkProcessor)

	// Check if an error occurred
	if err == nil {
		t.Error("Expected an error, but got nil")
	}

	// Check if the error message is correct
	expectedErrMsg := "Error processing chunk 3"
	if err[0].Summary != expectedErrMsg {
		t.Errorf("Expected error message '%s', but got '%s'", expectedErrMsg, err[0].Summary)
	}
}
