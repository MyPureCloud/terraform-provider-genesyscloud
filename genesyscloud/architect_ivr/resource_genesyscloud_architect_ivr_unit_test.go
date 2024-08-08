package architect_ivr

// build
import (
	"context"
	"fmt"
	"terraform-provider-genesyscloud/genesyscloud/provider"

	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"
	"github.com/stretchr/testify/assert"
)

/** Unit Test **/
func TestUnitResourceArchitectRead(t *testing.T) {
	tId := uuid.NewString()
	tName := "My Unit Test IVR"
	tDescription := "My Unit Test IVR"
	tDnis := []string{"+920-555-2902", "+920-321-5463"}
	tIDnis := make([]interface{}, len(tDnis))
	for i, v := range tDnis {
		tIDnis[i] = v
	}
	tOpenHoursFlowId := uuid.NewString()
	tClosedHoursFlowId := uuid.NewString()
	tHolidayHoursFlowId := uuid.NewString()
	tScheduleGroupId := uuid.NewString()
	tDivisionId := uuid.NewString()

	archProxy := &architectIvrProxy{}

	archProxy.getArchitectIvrAttr = func(ctx context.Context, a *architectIvrProxy, id string) (*platformclientv2.Ivr, *platformclientv2.APIResponse, error) {
		assert.Equal(t, tId, id)
		ivr := &platformclientv2.Ivr{
			Name:             &tName,
			Description:      &tDescription,
			Dnis:             &tDnis,
			OpenHoursFlow:    &platformclientv2.Domainentityref{Id: &tOpenHoursFlowId},
			ClosedHoursFlow:  &platformclientv2.Domainentityref{Id: &tClosedHoursFlowId},
			HolidayHoursFlow: &platformclientv2.Domainentityref{Id: &tHolidayHoursFlowId},
			ScheduleGroup:    &platformclientv2.Domainentityref{Id: &tScheduleGroupId},
			Division:         &platformclientv2.Writabledivision{Id: &tDivisionId},
		}

		apiResponse := &platformclientv2.APIResponse{StatusCode: http.StatusOK}
		return ivr, apiResponse, nil
	}
	internalProxy = archProxy
	defer func() { internalProxy = nil }()

	ctx := context.Background()
	gcloud := &provider.ProviderMeta{ClientConfig: &platformclientv2.Configuration{}}

	//Grab our defined schema
	resourceSchema := ResourceArchitectIvrConfig().Schema

	//Setup a map of values
	resourceDataMap := buildIvrResourceMap(tId, tName, tDescription, tIDnis, tOpenHoursFlowId, tClosedHoursFlowId, tHolidayHoursFlowId, tScheduleGroupId, tDivisionId)

	//Found this gem TestResourceDataRaw here: https://github.com/hashicorp/terraform/issues/890
	d := schema.TestResourceDataRaw(t, resourceSchema, resourceDataMap)
	d.SetId(tId)

	diag := readIvrConfig(ctx, d, gcloud)
	assert.Equal(t, false, diag.HasError())
	assert.Equal(t, tId, d.Id())
	assert.Equal(t, tName, d.Get("name").(string))
	assert.Equal(t, tName, d.Get("description").(string))
	assert.Equal(t, tOpenHoursFlowId, d.Get("open_hours_flow_id").(string))
	assert.Equal(t, tClosedHoursFlowId, d.Get("closed_hours_flow_id").(string))
	assert.Equal(t, tHolidayHoursFlowId, d.Get("holiday_hours_flow_id").(string))
	assert.Equal(t, tScheduleGroupId, d.Get("schedule_group_id").(string))
	assert.Equal(t, tDivisionId, d.Get("division_id").(string))

}

func TestUnitResourceArchitectDeleteStandard(t *testing.T) {
	tId := uuid.NewString()
	tName := "My Unit Test IVR"
	tDescription := "My Unit Test IVR"
	tDnis := []string{"+920-555-2902", "+920-321-5463"}
	tIDnis := make([]interface{}, len(tDnis))
	for i, v := range tDnis {
		tIDnis[i] = v
	}
	tOpenHoursFlowId := uuid.NewString()
	tClosedHoursFlowId := uuid.NewString()
	tHolidayHoursFlowId := uuid.NewString()
	tScheduleGroupId := uuid.NewString()
	tDivisionId := uuid.NewString()

	archProxy := &architectIvrProxy{}

	archProxy.deleteArchitectIvrAttr = func(ctx context.Context, a *architectIvrProxy, id string) (*platformclientv2.APIResponse, error) {
		assert.Equal(t, tId, id)

		apiResponse := &platformclientv2.APIResponse{StatusCode: http.StatusOK}
		return apiResponse, nil
	}

	archProxy.getArchitectIvrAttr = func(ctx context.Context, a *architectIvrProxy, id string) (*platformclientv2.Ivr, *platformclientv2.APIResponse, error) {
		assert.Equal(t, tId, id)

		apiResponse := &platformclientv2.APIResponse{StatusCode: http.StatusNotFound}
		err := fmt.Errorf("Unable to find targeted IVR: %s", id)
		return nil, apiResponse, err
	}

	internalProxy = archProxy
	defer func() { internalProxy = nil }()

	ctx := context.Background()
	gcloud := &provider.ProviderMeta{ClientConfig: &platformclientv2.Configuration{}}

	//Grab our defined schema
	resourceSchema := ResourceArchitectIvrConfig().Schema

	//Setup a map of values
	resourceDataMap := buildIvrResourceMap(tId, tName, tDescription, tIDnis, tOpenHoursFlowId, tClosedHoursFlowId, tHolidayHoursFlowId, tScheduleGroupId, tDivisionId)

	//Found this gem TestResourceDataRaw here: https://github.com/hashicorp/terraform/issues/890
	d := schema.TestResourceDataRaw(t, resourceSchema, resourceDataMap)
	d.SetId(tId)

	diag := deleteIvrConfig(ctx, d, gcloud)
	assert.Nil(t, diag)
	assert.Equal(t, tId, d.Id())
}

func TestUnitResourceArchitectDeleteSoftDelete(t *testing.T) {
	tId := uuid.NewString()
	tName := "My Unit Test IVR"
	tDescription := "My Unit Test IVR"
	tDnis := []string{"+920-555-2902", "+920-321-5463"}
	tIDnis := make([]interface{}, len(tDnis))
	for i, v := range tDnis {
		tIDnis[i] = v
	}
	tOpenHoursFlowId := uuid.NewString()
	tClosedHoursFlowId := uuid.NewString()
	tHolidayHoursFlowId := uuid.NewString()
	tScheduleGroupId := uuid.NewString()
	tDivisionId := uuid.NewString()

	archProxy := &architectIvrProxy{}

	archProxy.deleteArchitectIvrAttr = func(ctx context.Context, a *architectIvrProxy, id string) (*platformclientv2.APIResponse, error) {
		assert.Equal(t, tId, id)

		apiResponse := &platformclientv2.APIResponse{StatusCode: http.StatusOK}
		return apiResponse, nil
	}

	archProxy.getArchitectIvrAttr = func(ctx context.Context, a *architectIvrProxy, id string) (*platformclientv2.Ivr, *platformclientv2.APIResponse, error) {
		assert.Equal(t, tId, id)
		state := "deleted"
		ivr := &platformclientv2.Ivr{
			Name:             &tName,
			Description:      &tDescription,
			Dnis:             &tDnis,
			OpenHoursFlow:    &platformclientv2.Domainentityref{Id: &tOpenHoursFlowId},
			ClosedHoursFlow:  &platformclientv2.Domainentityref{Id: &tClosedHoursFlowId},
			HolidayHoursFlow: &platformclientv2.Domainentityref{Id: &tHolidayHoursFlowId},
			ScheduleGroup:    &platformclientv2.Domainentityref{Id: &tScheduleGroupId},
			Division:         &platformclientv2.Writabledivision{Id: &tDivisionId},
			State:            &state,
		}

		apiResponse := &platformclientv2.APIResponse{StatusCode: http.StatusOK}
		return ivr, apiResponse, nil
	}

	internalProxy = archProxy
	defer func() { internalProxy = nil }()
	ctx := context.Background()
	gcloud := &provider.ProviderMeta{ClientConfig: &platformclientv2.Configuration{}}

	//Grab our defined schema
	resourceSchema := ResourceArchitectIvrConfig().Schema

	//Setup a map of values
	resourceDataMap := buildIvrResourceMap(tId, tName, tDescription, tIDnis, tOpenHoursFlowId, tClosedHoursFlowId, tHolidayHoursFlowId, tScheduleGroupId, tDivisionId)

	//Found this gem TestResourceDataRaw here: https://github.com/hashicorp/terraform/issues/890
	d := schema.TestResourceDataRaw(t, resourceSchema, resourceDataMap)
	d.SetId(tId)

	diag := deleteIvrConfig(ctx, d, gcloud)
	assert.Nil(t, diag)
	assert.Equal(t, tId, d.Id())

}

func TestUnitResourceArchitectCreate(t *testing.T) {
	tId := uuid.NewString()
	tName := "My Unit Test IVR"
	tDescription := "My Unit Test IVR"
	tDnis := []string{"+920-555-2902", "+920-321-5463"}
	tIDnis := make([]interface{}, len(tDnis))
	for i, v := range tDnis {
		tIDnis[i] = v
	}
	tOpenHoursFlowId := uuid.NewString()
	tClosedHoursFlowId := uuid.NewString()
	tHolidayHoursFlowId := uuid.NewString()
	tScheduleGroupId := uuid.NewString()
	tDivisionId := uuid.NewString()

	archProxy := &architectIvrProxy{}
	archProxy.getArchitectIvrAttr = func(ctx context.Context, a *architectIvrProxy, id string) (*platformclientv2.Ivr, *platformclientv2.APIResponse, error) {
		assert.Equal(t, tId, id)
		ivr := &platformclientv2.Ivr{
			Id:               &tId,
			Name:             &tName,
			Description:      &tDescription,
			Dnis:             &tDnis,
			OpenHoursFlow:    &platformclientv2.Domainentityref{Id: &tOpenHoursFlowId},
			ClosedHoursFlow:  &platformclientv2.Domainentityref{Id: &tClosedHoursFlowId},
			HolidayHoursFlow: &platformclientv2.Domainentityref{Id: &tHolidayHoursFlowId},
			ScheduleGroup:    &platformclientv2.Domainentityref{Id: &tScheduleGroupId},
			Division:         &platformclientv2.Writabledivision{Id: &tDivisionId},
		}

		apiResponse := &platformclientv2.APIResponse{StatusCode: http.StatusOK}
		return ivr, apiResponse, nil
	}

	archProxy.createArchitectIvrAttr = func(ctx context.Context, a *architectIvrProxy, ivr platformclientv2.Ivr) (*platformclientv2.Ivr, *platformclientv2.APIResponse, error) {
		assert.Equal(t, tName, *ivr.Name, "ivr.Name check failed in create createArchitectIvrAttr")
		assert.Equal(t, tDescription, *ivr.Description, "ivr.Description check failed in create createArchitectIvrAttr")
		assert.ElementsMatch(t, tDnis, *ivr.Dnis, "ivr.Dnis check failed in create createArchitectIvrAttr")
		assert.Equal(t, tOpenHoursFlowId, *ivr.OpenHoursFlow.Id, "ivr.OpenHoursFlow.Id check failed in create createArchitectIvrAttr")
		assert.Equal(t, tClosedHoursFlowId, *ivr.ClosedHoursFlow.Id, "ivr.ClosedHoursFlow.Id check failed in create createArchitectIvrAttr")
		assert.Equal(t, tHolidayHoursFlowId, *ivr.HolidayHoursFlow.Id, "ivr.HolidayHoursFlow.Id check failed in create createArchitectIvrAttr")
		assert.Equal(t, tScheduleGroupId, *ivr.ScheduleGroup.Id, "ivr.ScheduleGroup.Id check failed in create createArchitectIvrAttr")
		assert.Equal(t, tDivisionId, *ivr.Division.Id, "ivr.Division.Id check failed in create createArchitectIvrAttr")

		ivr.Id = &tId

		apiResponse := &platformclientv2.APIResponse{StatusCode: http.StatusOK}
		return &ivr, apiResponse, nil
	}

	internalProxy = archProxy
	defer func() { internalProxy = nil }()

	ctx := context.Background()
	gcloud := &provider.ProviderMeta{ClientConfig: &platformclientv2.Configuration{}}

	//Grab our defined schema
	resourceSchema := ResourceArchitectIvrConfig().Schema

	//Setup a map of values
	resourceDataMap := buildIvrResourceMap(tId, tName, tDescription, tIDnis, tOpenHoursFlowId, tClosedHoursFlowId, tHolidayHoursFlowId, tScheduleGroupId, tDivisionId)

	//Found this gem TestResourceDataRaw here: https://github.com/hashicorp/terraform/issues/890
	d := schema.TestResourceDataRaw(t, resourceSchema, resourceDataMap)
	d.SetId(tId)

	diag := createIvrConfig(ctx, d, gcloud)
	assert.Equal(t, false, diag.HasError())
	assert.Equal(t, tId, d.Id())
}

func TestUnitResourceArchitectUpdate(t *testing.T) {
	tId := uuid.NewString()
	tName := "My Unit Test IVR"
	tDescription := "My updated Unit Test IVR"
	tDnis := []string{"+920-555-2902", "+920-321-5463"}
	tIDnis := make([]interface{}, len(tDnis))
	for i, v := range tDnis {
		tIDnis[i] = v
	}
	tOpenHoursFlowId := uuid.NewString()
	tClosedHoursFlowId := uuid.NewString()
	tHolidayHoursFlowId := uuid.NewString()
	tScheduleGroupId := uuid.NewString()
	tDivisionId := uuid.NewString()

	archProxy := &architectIvrProxy{}
	archProxy.getArchitectIvrAttr = func(ctx context.Context, a *architectIvrProxy, id string) (*platformclientv2.Ivr, *platformclientv2.APIResponse, error) {
		assert.Equal(t, tId, id)
		ivr := &platformclientv2.Ivr{
			Id:               &tId,
			Name:             &tName,
			Description:      &tDescription,
			Dnis:             &tDnis,
			OpenHoursFlow:    &platformclientv2.Domainentityref{Id: &tOpenHoursFlowId},
			ClosedHoursFlow:  &platformclientv2.Domainentityref{Id: &tClosedHoursFlowId},
			HolidayHoursFlow: &platformclientv2.Domainentityref{Id: &tHolidayHoursFlowId},
			ScheduleGroup:    &platformclientv2.Domainentityref{Id: &tScheduleGroupId},
			Division:         &platformclientv2.Writabledivision{Id: &tDivisionId},
		}

		apiResponse := &platformclientv2.APIResponse{StatusCode: http.StatusOK}
		return ivr, apiResponse, nil
	}

	archProxy.updateArchitectIvrAttr = func(ctx context.Context, a *architectIvrProxy, id string, ivr platformclientv2.Ivr) (*platformclientv2.Ivr, *platformclientv2.APIResponse, error) {
		assert.Equal(t, tName, *ivr.Name, "ivr.Name check failed in create updateArchitectIvrAttr")
		assert.Equal(t, tDescription, *ivr.Description, "ivr.Description check failed in updateArchitectIvrAttr")
		assert.ElementsMatch(t, tDnis, *ivr.Dnis, "ivr.Dnis check failed in updateArchitectIvrAttr")
		assert.Equal(t, tOpenHoursFlowId, *ivr.OpenHoursFlow.Id, "ivr.OpenHoursFlow.Id check failed in updateArchitectIvrAttr")
		assert.Equal(t, tClosedHoursFlowId, *ivr.ClosedHoursFlow.Id, "ivr.ClosedHoursFlow.Id check failed in updateArchitectIvrAttr")
		assert.Equal(t, tHolidayHoursFlowId, *ivr.HolidayHoursFlow.Id, "ivr.HolidayHoursFlow.Id check failed in updateArchitectIvrAttr")
		assert.Equal(t, tScheduleGroupId, *ivr.ScheduleGroup.Id, "ivr.ScheduleGroup.Id check failed in updateArchitectIvrAttr")
		assert.Equal(t, tDivisionId, *ivr.Division.Id, "ivr.Division.Id check failed in updateArchitectIvrAttr")

		ivr.Id = &tId

		apiResponse := &platformclientv2.APIResponse{StatusCode: http.StatusOK}
		return &ivr, apiResponse, nil
	}

	internalProxy = archProxy
	defer func() { internalProxy = nil }()

	ctx := context.Background()
	gcloud := &provider.ProviderMeta{ClientConfig: &platformclientv2.Configuration{}}

	//Grab our defined schema
	resourceSchema := ResourceArchitectIvrConfig().Schema

	//Setup a map of values
	resourceDataMap := buildIvrResourceMap(tId, tName, tDescription, tIDnis, tOpenHoursFlowId, tClosedHoursFlowId, tHolidayHoursFlowId, tScheduleGroupId, tDivisionId)

	//Found this gem TestResourceDataRaw here: https://github.com/hashicorp/terraform/issues/890
	d := schema.TestResourceDataRaw(t, resourceSchema, resourceDataMap)
	d.SetId(tId)

	diag := updateIvrConfig(ctx, d, gcloud)
	assert.Equal(t, false, diag.HasError())
	assert.Equal(t, tId, d.Id())
	assert.Equal(t, tDescription, d.Get("description").(string))
}

func buildIvrResourceMap(tId string, tName string, tDescription string, tIDnis []interface{}, tOpenHoursFlowId string, tClosedHoursFlowId string, tHolidayHoursFlowId string, tScheduleGroupId string, tDivisionId string) map[string]interface{} {
	resourceDataMap := map[string]interface{}{
		"id":                    tId,
		"name":                  tName,
		"description":           tDescription,
		"dnis":                  tIDnis,
		"open_hours_flow_id":    tOpenHoursFlowId,
		"closed_hours_flow_id":  tClosedHoursFlowId,
		"holiday_hours_flow_id": tHolidayHoursFlowId,
		"schedule_group_id":     tScheduleGroupId,
		"division_id":           tDivisionId,
	}
	return resourceDataMap
}
