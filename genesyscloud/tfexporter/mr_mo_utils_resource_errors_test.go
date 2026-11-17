package tfexporter

import (
	"context"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestUnitRecordResourceErrors_NilMapOnMrMoPath reproduces the panic seen in the
// reconciler's source export: the MrMo entrypoints never call Export(), which is the
// only place that used to initialize resourceErrors, so the first failed instance of
// a resource type wrote to a nil map and killed the process.
func TestUnitRecordResourceErrors_NilMapOnMrMoPath(t *testing.T) {
	g := &GenesysCloudResourceExporter{ctx: context.Background()}
	require.Nil(t, g.resourceErrors, "precondition: the MrMo path starts with a nil map")

	errs := []ResourceErrorInfo{{
		ResourceID:   "8eedd8b9-ae19-4b3d-abf0-008c20b85bb7",
		ResourceType: "genesyscloud_user",
		ErrorMessage: "Failed after 3 retries: API Error: 429",
		IsTimeout:    false,
	}}

	require.NotPanics(t, func() {
		g.recordResourceErrors("genesyscloud_user", errs)
	})

	assert.Equal(t, errs, g.resourceErrorsForType("genesyscloud_user"))
}

func TestUnitResourceErrorsForType_NilMapReturnsNothing(t *testing.T) {
	g := &GenesysCloudResourceExporter{ctx: context.Background()}

	var got []ResourceErrorInfo
	require.NotPanics(t, func() {
		got = g.resourceErrorsForType("genesyscloud_user")
	})
	assert.Empty(t, got)
}

// TestUnitRecordResourceErrors_ConcurrentTypes covers the reason the write is
// mutex-guarded: getResourcesForType runs per resource type and the MrMo export
// worker pool drives several concurrently in one process.
func TestUnitRecordResourceErrors_ConcurrentTypes(t *testing.T) {
	g := &GenesysCloudResourceExporter{ctx: context.Background()}

	resTypes := []string{
		"genesyscloud_user",
		"genesyscloud_routing_queue",
		"genesyscloud_auth_division",
		"genesyscloud_location",
	}

	var wg sync.WaitGroup
	for _, resType := range resTypes {
		wg.Add(1)
		go func(resType string) {
			defer wg.Done()
			g.recordResourceErrors(resType, []ResourceErrorInfo{{ResourceType: resType}})
		}(resType)
	}
	wg.Wait()

	for _, resType := range resTypes {
		recorded := g.resourceErrorsForType(resType)
		require.Len(t, recorded, 1, "expected recorded errors for %s", resType)
		assert.Equal(t, resType, recorded[0].ResourceType)
	}
}

// TestUnitExportByTypeForMrMo_SurfacesResourceErrors proves the errors now leave the
// provider. Export()'s step #9 reporting (export_errors.json plus a warning
// diagnostic) never runs on the MrMo path, and tflog output is dropped when the
// exporter runs standalone, so the response field is the caller's only signal that
// ResourceDataList is incomplete.
func TestUnitExportByTypeForMrMo_SurfacesResourceErrors(t *testing.T) {
	g := &GenesysCloudResourceExporter{ctx: context.Background()}
	g.recordResourceErrors("genesyscloud_user", []ResourceErrorInfo{
		{ResourceID: "id-1", ResourceType: "genesyscloud_user", ErrorMessage: "boom"},
	})

	response := &MrMoExportByTypeResponse{
		ResourceErrors: g.resourceErrorsForType("genesyscloud_user"),
	}

	require.Len(t, response.ResourceErrors, 1)
	assert.Equal(t, "id-1", response.ResourceErrors[0].ResourceID)
	assert.Equal(t, "boom", response.ResourceErrors[0].ErrorMessage)
}
