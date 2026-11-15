package tfexporter

import (
	"context"
	"sync"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	telephonySite "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/telephony_providers_edges_site"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func managedSiteState() *terraform.InstanceState {
	return &terraform.InstanceState{
		ID: "site-1",
		Attributes: map[string]string{
			"id":          "site-1",
			"name":        "PureCloud Voice - AWS",
			"managed":     "true",
			"media_model": "Cloud",
		},
	}
}

// The data source copy carries the data source's attributes, and the original state is untouched.
func TestUnitDataSourceInstanceStateDoesNotMutateTheInput(t *testing.T) {
	state := managedSiteState()
	original := map[string]string{}
	for k, v := range state.Attributes {
		original[k] = v
	}

	copied := dataSourceInstanceState(telephonySite.SiteExporter(), telephonySite.DataSourceSite(), state)

	assert.NotSame(t, state, copied, "a new state must be returned")
	assert.Equal(t, original, state.Attributes, "the input state must keep the instance's own attributes")
	assert.Equal(t, "site-1", copied.ID)
	assert.Equal(t, map[string]string{"name": "PureCloud Voice - AWS"}, copied.Attributes,
		"the copy carries only the data source's attributes")
}

// The incident: Mr Mo reads the ResourceData after the export, and the exporter's rewrite of the
// shared state removed "managed", so the exporter's own decision could not be repeated from it.
func TestUnitMrMoResourceDataKeepsItsAttributesAfterTheDataSourceCopy(t *testing.T) {
	site := telephonySite.ResourceSite()
	exporter := telephonySite.SiteExporter()
	state := managedSiteState()

	// What getResourceState hands to Mr Mo, before anything has read it.
	resourceData := site.Data(state)
	// What the exporter then does for an entity exported as a data source.
	_ = dataSourceInstanceState(exporter, telephonySite.DataSourceSite(), state)

	attrs := resourceData.State().Attributes
	assert.Equal(t, "true", attrs["managed"], "the instance's own attributes must still be readable")
	assert.Equal(t, "Cloud", attrs["media_model"])

	exportAsData, err := exporter.ExportAsDataFunc(context.Background(), nil, attrs)
	require.NoError(t, err, "the decision must be repeatable from the ResourceData")
	assert.True(t, exportAsData)
}

func TestUnitRecordMrMoExportedAsData(t *testing.T) {
	g := &GenesysCloudResourceExporter{}

	assert.NotNil(t, g.mrMoExportedAsData("genesyscloud_location"), "never nil, even before anything is recorded")
	assert.Empty(t, g.mrMoExportedAsData("genesyscloud_location"))

	g.recordMrMoExportedAsData("genesyscloud_location", "loc-1")
	g.recordMrMoExportedAsData("genesyscloud_telephony_providers_edges_site", "site-1")

	assert.Equal(t, map[string]bool{"loc-1": true}, g.mrMoExportedAsData("genesyscloud_location"))
	assert.Equal(t, map[string]bool{"site-1": true}, g.mrMoExportedAsData("genesyscloud_telephony_providers_edges_site"),
		"ids are kept per resource type")

	// The result is a copy.
	got := g.mrMoExportedAsData("genesyscloud_location")
	got["other"] = true
	assert.Equal(t, map[string]bool{"loc-1": true}, g.mrMoExportedAsData("genesyscloud_location"))
}

func TestUnitMrMoResourceExportedAsData(t *testing.T) {
	g := &GenesysCloudResourceExporter{}
	g.recordMrMoExportedAsData("genesyscloud_location", "loc-1")

	flagged := &schema.ResourceData{}
	flagged.SetId("loc-1")
	ordinary := &schema.ResourceData{}
	ordinary.SetId("loc-2")

	assert.True(t, g.mrMoResourceExportedAsData("genesyscloud_location", flagged))
	assert.False(t, g.mrMoResourceExportedAsData("genesyscloud_location", ordinary))
	assert.False(t, g.mrMoResourceExportedAsData("genesyscloud_location", nil))
	assert.False(t, g.mrMoResourceExportedAsData("genesyscloud_group", flagged), "another type does not match")
}

func TestUnitRecordMrMoExportedAsDataIsSafeConcurrently(t *testing.T) {
	g := &GenesysCloudResourceExporter{}
	const workers = 50

	var wg sync.WaitGroup
	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go func(i int) {
			defer wg.Done()
			g.recordMrMoExportedAsData("genesyscloud_telephony_providers_edges_site", string(rune('a'+i%26))+string(rune('a'+i/26)))
			_ = g.mrMoExportedAsData("genesyscloud_telephony_providers_edges_site")
		}(i)
	}
	wg.Wait()

	assert.Len(t, g.mrMoExportedAsData("genesyscloud_telephony_providers_edges_site"), workers)
}
