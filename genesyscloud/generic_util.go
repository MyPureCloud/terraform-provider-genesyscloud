package genesyscloud

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
)

func ProcessChunksInBatches[T any, K any, U any](c *chunkUploader[T, K, U]) diag.Diagnostics {
	if c.batchSize <= 0 {
		return diag.Errorf("batchSize must be greater than zero")
	}
	if len(c.items) > 0 {
		for i := 0; i < len(c.items); i += c.batchSize {
			end := i + c.batchSize
			if end > len(c.items) {
				end = len(c.items)
			}
			var updateChunk []K
			for j := i; j < end; j++ {
				updateChunk = c.chunkBuilderAttr(j, updateChunk, c)
			}
			if len(updateChunk) > 0 {
				return c.chunkProcessorAttr(updateChunk, c)
			}
		}
	} else {
		return c.chunkProcessorAttr([]K{}, c)
	}
	return nil
}

type chunkProcessorFunc[T any, K any, U any] func(updateChunk []K, c *chunkUploader[T, K, U]) diag.Diagnostics

type chunkBuilderFunc[T any, K any, U any] func(seq int, updateChunk []K, c *chunkUploader[T, K, U]) []K

type chunkUploader[T any, K any, U any] struct {
	id                 string
	items              []T
	remove             bool
	platformApi        *U
	chunk              []K
	batchSize          int
	chunkProcessorAttr chunkProcessorFunc[T, K, U]
	chunkBuilderAttr   chunkBuilderFunc[T, K, U]
}
