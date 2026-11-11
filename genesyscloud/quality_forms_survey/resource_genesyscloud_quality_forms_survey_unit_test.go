package quality_forms_survey

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v200/platformclientv2"
	"github.com/stretchr/testify/assert"
)

func TestUnitSurveyFormNewTopLevelFieldsSchema(t *testing.T) {
	s := ResourceQualityFormsSurvey().Schema

	// Writable field that the API also returns
	assert.True(t, s["division_id"].Optional, "division_id should be optional")
	assert.True(t, s["division_id"].Computed, "division_id should be computed")
}

func TestUnitSurveyFormExporterConfig(t *testing.T) {
	exporter := QualityFormsSurveyExporter()

	assert.NotContains(t, exporter.ExcludedAttributes, "division_id")

	ref, ok := exporter.RefAttrs["division_id"]
	assert.True(t, ok, "division_id should be exported as a reference")
	assert.Equal(t, "genesyscloud_auth_division", ref.RefType)
}

func TestUnitBuildSdkDivision(t *testing.T) {
	res := ResourceQualityFormsSurvey()

	withDivision := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{"division_id": "division-123"})
	division := buildSdkDivision(withDivision)
	if assert.NotNil(t, division) {
		assert.Equal(t, "division-123", *division.Id)
	}

	withoutDivision := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{})
	assert.Nil(t, buildSdkDivision(withoutDivision), "division must not be sent when division_id is not set")
}

func TestUnitFlattenDivisionId(t *testing.T) {
	res := ResourceQualityFormsSurvey()
	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{})

	divisionId := "division-456"
	flattenDivisionId(d, &platformclientv2.Writablestarrabledivision{Id: &divisionId})
	assert.Equal(t, divisionId, d.Get("division_id"))

	// A nil division in the API response must leave the existing value untouched
	flattenDivisionId(d, nil)
	assert.Equal(t, divisionId, d.Get("division_id"))
}

func TestUnitGenerateSurveyFormDivisionId(t *testing.T) {
	assert.Equal(t, "", generateSurveyFormDivisionId(""))
	assert.Equal(t, "division_id = genesyscloud_auth_division.d.id", generateSurveyFormDivisionId("genesyscloud_auth_division.d.id"))
}
