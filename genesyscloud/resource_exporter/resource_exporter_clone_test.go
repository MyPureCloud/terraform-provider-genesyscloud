package resource_exporter

import (
	"context"
	"sync"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCloneResourceExporter_isolatesMutableState(t *testing.T) {
	t.Parallel()

	template := &ResourceExporter{
		RefAttrs: map[string]*RefAttrSettings{
			"division_id": {RefType: "genesyscloud_auth_division"},
		},
		SanitizedResourceMap: ResourceIDMetaMap{
			"id-1": {BlockLabel: "template-label"},
		},
		ExcludedAttributes: []string{"computed"},
		FilterResource: func(resourceIdMetaMap ResourceIDMetaMap, resourceType string, filter []string) ResourceIDMetaMap {
			return resourceIdMetaMap
		},
	}

	clone := CloneResourceExporter(template)
	require.NotNil(t, clone)
	assert.NotSame(t, template, clone)

	clone.SanitizedResourceMap = ResourceIDMetaMap{
		"id-2": {BlockLabel: "clone-label"},
	}
	clone.ExcludedAttributes = append(clone.ExcludedAttributes, "extra")
	clone.FilterResource = nil

	assert.Equal(t, "template-label", template.SanitizedResourceMap["id-1"].BlockLabel)
	assert.Len(t, template.ExcludedAttributes, 1)
	assert.NotNil(t, template.FilterResource)
	assert.Equal(t, template.RefAttrs, clone.RefAttrs)
}

func TestCloneResourceExporter_nilTemplate(t *testing.T) {
	t.Parallel()

	assert.Nil(t, CloneResourceExporter(nil))
}

func TestCloneResourceExporter_parallelLoadDoesNotShareMaps(t *testing.T) {
	t.Parallel()

	type loadIDKey struct{}

	template := &ResourceExporter{
		GetResourcesFunc: func(ctx context.Context) (ResourceIDMetaMap, diag.Diagnostics) {
			id := ctx.Value(loadIDKey{}).(string)
			return ResourceIDMetaMap{
				id: {BlockLabel: id},
			}, nil
		},
	}

	const workers = 8
	var wg sync.WaitGroup
	wg.Add(workers)

	clones := make([]*ResourceExporter, workers)
	for i := 0; i < workers; i++ {
		go func(idx int) {
			defer wg.Done()
			clone := CloneResourceExporter(template)
			id := string(rune('a' + idx))
			ctx := context.WithValue(context.Background(), loadIDKey{}, id)
			diags := clone.LoadSanitizedResourceMap(ctx, "test_type", nil)
			require.False(t, diags.HasError())
			clones[idx] = clone
		}(i)
	}
	wg.Wait()

	for i, clone := range clones {
		id := string(rune('a' + i))
		require.Len(t, clone.SanitizedResourceMap, 1)
		assert.Contains(t, clone.SanitizedResourceMap, id)
	}

	assert.Nil(t, template.SanitizedResourceMap)
}
