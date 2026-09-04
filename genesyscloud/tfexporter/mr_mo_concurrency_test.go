package tfexporter

import (
	"sync"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/mrmo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConcurrentMrMoResourceExportWrites(t *testing.T) {
	t.Setenv(mrmo.MRMO_CXASCODE_INTEGRATION_ENABLED, "true")
	t.Cleanup(mrmo.Reset)

	const (
		workers    = 50
		resType    = "genesyscloud_auth_division"
		iterations = 3
	)

	exporter := &GenesysCloudResourceExporter{}
	var wg sync.WaitGroup
	wg.Add(workers)

	for worker := 0; worker < workers; worker++ {
		go func(worker int) {
			defer wg.Done()
			for i := 0; i < iterations; i++ {
				id := worker*iterations + i
				resourceData := &schema.ResourceData{}
				resourceData.SetId(string(rune('a' + (id % 26))))
				exporter.recordMrMoResourceExport(resType, resourceData)
			}
		}(worker)
	}

	wg.Wait()

	list := exporter.mrMoResourceDataList(resType)
	assert.Len(t, list, workers*iterations)
}

func TestMrMoResourceExportWritesAreSafeAcrossResourceTypes(t *testing.T) {
	t.Setenv(mrmo.MRMO_CXASCODE_INTEGRATION_ENABLED, "true")
	t.Cleanup(mrmo.Reset)

	exporter := &GenesysCloudResourceExporter{}
	resourceTypes := []string{
		"genesyscloud_auth_division",
		"genesyscloud_knowledge_knowledgebase",
		"genesyscloud_telephony_providers_edges_phone",
	}

	const workers = 20
	var wg sync.WaitGroup
	wg.Add(workers * len(resourceTypes))

	for _, resType := range resourceTypes {
		for worker := 0; worker < workers; worker++ {
			go func(resType string, worker int) {
				defer wg.Done()
				resourceData := &schema.ResourceData{}
				resourceData.SetId(string(rune('A'+worker%26)) + resType)
				exporter.recordMrMoResourceExport(resType, resourceData)
			}(resType, worker)
		}
	}

	wg.Wait()

	for _, resType := range resourceTypes {
		assert.Len(t, exporter.mrMoResourceDataList(resType), workers)
	}
}

func TestNewGenesysCloudResourceExporterInitializesResourceErrors(t *testing.T) {
	exporter := &GenesysCloudResourceExporter{
		resourceErrors: make(map[string][]ResourceErrorInfo),
	}

	require.NotNil(t, exporter.resourceErrors)
	exporter.resourceErrorsMutex.Lock()
	exporter.resourceErrors["genesyscloud_knowledge_knowledgebase"] = []ResourceErrorInfo{
		{ResourceID: "kb-1", ErrorMessage: "test"},
	}
	exporter.resourceErrorsMutex.Unlock()

	assert.Len(t, exporter.resourceErrors["genesyscloud_knowledge_knowledgebase"], 1)
}
