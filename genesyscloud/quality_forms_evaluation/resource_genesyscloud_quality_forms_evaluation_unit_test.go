package quality_forms_evaluation

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v200/platformclientv2"
	"github.com/stretchr/testify/assert"
)

func TestUnitEvaluationFormNewTopLevelFieldsSchema(t *testing.T) {
	s := ResourceEvaluationForm().Schema

	// Writable field
	assert.True(t, s["division_id"].Optional, "division_id should be optional")
	assert.True(t, s["division_id"].Computed, "division_id should be computed")

	// Read-only fields must be Computed only
	for _, key := range []string{"weight_mode", "latest_version_form_name"} {
		assert.True(t, s[key].Computed, "%s should be computed", key)
		assert.False(t, s[key].Optional, "%s must not be settable", key)
		assert.False(t, s[key].Required, "%s must not be settable", key)
	}
}

func TestUnitEvaluationFormExporterConfig(t *testing.T) {
	exporter := EvaluationFormExporter()

	assert.Contains(t, exporter.ExcludedAttributes, "weight_mode")
	assert.Contains(t, exporter.ExcludedAttributes, "latest_version_form_name")
	assert.NotContains(t, exporter.ExcludedAttributes, "division_id")

	ref, ok := exporter.RefAttrs["division_id"]
	assert.True(t, ok, "division_id should be exported as a reference")
	assert.Equal(t, "genesyscloud_auth_division", ref.RefType)
}

func TestUnitBuildSdkDivision(t *testing.T) {
	res := ResourceEvaluationForm()

	withDivision := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{"division_id": "division-123"})
	division := buildSdkDivision(withDivision)
	if assert.NotNil(t, division) {
		assert.Equal(t, "division-123", *division.Id)
	}

	withoutDivision := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{})
	assert.Nil(t, buildSdkDivision(withoutDivision), "division must not be sent when division_id is not set")
}

func TestUnitFlattenDivisionId(t *testing.T) {
	res := ResourceEvaluationForm()
	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{})

	divisionId := "division-456"
	flattenDivisionId(d, &platformclientv2.Writablestarrabledivision{Id: &divisionId})
	assert.Equal(t, divisionId, d.Get("division_id"))

	// A nil division in the API response must leave the existing value untouched
	flattenDivisionId(d, nil)
	assert.Equal(t, divisionId, d.Get("division_id"))
}

func TestUnitGenerateFormDivisionId(t *testing.T) {
	assert.Equal(t, "", generateFormDivisionId(""))
	assert.Equal(t, "division_id = genesyscloud_auth_division.d.id", generateFormDivisionId("genesyscloud_auth_division.d.id"))
}
