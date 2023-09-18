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

func ChunkBy[T any](items []T, chunkSize int) (chunks [][]T) {
	for chunkSize < len(items) {
		items, chunks = items[chunkSize:], append(chunks, items[0:chunkSize:chunkSize])
	}
	return append(chunks, items)
}

func mapItems[T, P any](items []T, mapBuilder func(T) P) []P {
	return u.Map(items, func(item T) P {
		return mapBuilder(item)
	})
}

func ChunkItems[T, P any](items []T, mapBuilder func(T) P, chunkSize int) (chunks [][]P) {
	mappedItems := mapItems(items, mapBuilder)
	return ChunkBy(mappedItems, chunkSize)
}

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
