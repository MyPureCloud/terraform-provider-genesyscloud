package chunks

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	u "github.com/rjNemo/underscore"
)

func seqGen() func() int {
	seq := -1 // Initialize sequence

	return func() int {
		seq++
		return seq
	}
}

// Generic function to chunk given Items based on the size provided.
func ChunkBy[T any](items []T, chunkSize int) (chunks [][]T) {
	for chunkSize < len(items) {
		items, chunks = items[chunkSize:], append(chunks, items[0:chunkSize:chunkSize])
	}
	return append(chunks, items)
}

// Generic function to Map each Item in Items based on a transform/builder function
func mapItems[T, P any](items []T, mapBuilder func(T) P) []P {
	return u.Map(items, func(item T) P {
		return mapBuilder(item)
	})
}

// Generic function that Chunks the Items and then Maps
func ChunkItems[T, P any](items []T, mapBuilder func(T) P, chunkSize int) (chunks [][]P) {
	mappedItems := mapItems(items, mapBuilder)
	return ChunkBy(mappedItems, chunkSize)
}

// Generic function that takes array of Chunks and then Processes each Chunk with the defined Funcion.
// Typically Processor Function would be a REST API call

func ProcessChunks[T any](chunks []T, chunkProcessor func(T) diag.Diagnostics) diag.Diagnostics {
	var err diag.Diagnostics
	u.Map(chunks, func(chunk T) diag.Diagnostics {
		if err != nil {
			return err
		}
		err = chunkProcessor(chunk)
		return err
	})

	return err
}
