package telephony_providers_edges_phonebasesettings

import (
	"context"
	"net/http"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v150/platformclientv2"
	"github.com/stretchr/testify/assert"
)

func TestUnitResourceTelephonyProvidersEdgesPhoneBaseSettingsCreate(t *testing.T) {

	tPhoneBaseSettingsId := uuid.NewString()
	tLineId := uuid.NewString()

	tSettingsName := "Polycom VVX 500 settings name"
	tSettingsState := "active"
	tSettingsMetaBaseId := "polycom_vvx_500.json"
	tSettingsMetaBaseName := "Polycom VVX 500"

	tTemplateLineMetaBaseId := "polycom_vvx.json"
	tTemplateLineMetaBaseName := "Polycom VVX line appearances"

	functionCalls := make([]string, 0)

	pbsProxy := &phoneBaseProxy{}

	tlineMetaBase := platformclientv2.Domainentityref{
		Id:   &tTemplateLineMetaBaseId,
		Name: &tTemplateLineMetaBaseName,
	}
	pbsProxy.getPhoneBaseSettingTemplateAttr = func(ctx context.Context, p *phoneBaseProxy, phoneBaseSettingsId string) (*platformclientv2.Phonebase, *platformclientv2.APIResponse, error) {
		functionCalls = append(functionCalls, "getPhoneBaseSettingTemplateAttr")
		apiResponse := platformclientv2.APIResponse{
			StatusCode: http.StatusOK,
		}

		lineBase := []platformclientv2.Linebase{
			{
				LineMetaBase: &tlineMetaBase,
			},
		}

		phoneBaseSettings := platformclientv2.Phonebase{
			State: &tSettingsState,
			Lines: &lineBase,
			Name:  &tSettingsName,
			PhoneMetaBase: &platformclientv2.Domainentityref{
				Id:   &tSettingsMetaBaseId,
				Name: &tSettingsMetaBaseName,
			},
		}
		return &phoneBaseSettings, &apiResponse, nil
	}

	pbsProxy.getPhoneBaseSettingAttr = func(ctx context.Context, p *phoneBaseProxy, phoneBaseSettingsId string) (*platformclientv2.Phonebase, *platformclientv2.APIResponse, error) {
		functionCalls = append(functionCalls, "getPhoneBaseSettingAttr")
		apiResponse := platformclientv2.APIResponse{
			StatusCode: http.StatusOK,
		}

		lineBase := []platformclientv2.Linebase{
			{
				Id:           &tLineId,
				State:        &tSettingsState,
				LineMetaBase: &tlineMetaBase,
			},
		}

		phoneBaseSettings := platformclientv2.Phonebase{
			State: &tSettingsState,
			Lines: &lineBase,
			PhoneMetaBase: &platformclientv2.Domainentityref{
				Id:   &tSettingsMetaBaseId,
				Name: &tSettingsMetaBaseName,
			},
			Name: &tSettingsName,
		}
		return &phoneBaseSettings, &apiResponse, nil
	}

	pbsProxy.postPhoneBaseSettingAttr = func(ctx context.Context, p *phoneBaseProxy, body platformclientv2.Phonebase) (*platformclientv2.Phonebase, *platformclientv2.APIResponse, error) {
		functionCalls = append(functionCalls, "postPhoneBaseSettingAttr")
		assert.Equal(t, tSettingsName, *body.Name)
		assert.Equal(t, tSettingsMetaBaseId, *body.PhoneMetaBase.Id)

		body.Id = &tPhoneBaseSettingsId
		apiResponse := platformclientv2.APIResponse{
			StatusCode: http.StatusOK,
		}

		return &body, &apiResponse, nil
	}

	internalProxy = pbsProxy
	defer func() {
		internalProxy = nil
	}()

	ctx := context.Background()
	gcloud := &provider.ProviderMeta{ClientConfig: &platformclientv2.Configuration{}}

	// Grab our defined schema
	resourceSchema := ResourcePhoneBaseSettings().Schema

	// Setup a map of values
	resourceDataMap := buildPhoneBaseSettingsResource(
		tSettingsName, tSettingsMetaBaseId,
	)

	d := schema.TestResourceDataRaw(t, resourceSchema, resourceDataMap)
	d.SetId(tPhoneBaseSettingsId)

	diag := createPhoneBaseSettings(ctx, d, gcloud)
	assert.Equal(t, false, diag.HasError(), "createPhoneBaseSettings returned error")
	assert.Equal(t, []string{
		"getPhoneBaseSettingTemplateAttr",
		"postPhoneBaseSettingAttr",
		"getPhoneBaseSettingAttr", // This is called twice, once for the read function and once for the custom diff function
		"getPhoneBaseSettingAttr",
	}, functionCalls)
	assert.Equal(t, tPhoneBaseSettingsId, d.Id())
	assert.Equal(t, tSettingsName, d.Get("name").(string))
	assert.Equal(t, tSettingsMetaBaseId, d.Get("phone_meta_base_id").(string))
}

func TestUnitResourceTelephonyProvidersEdgesPhoneBaseSettingsUpdate(t *testing.T) {

	tPhoneBaseSettingsId := uuid.NewString()
	tLineId := uuid.NewString()

	tSettingsName := "Polycom VVX 500 settings name"
	tSettingsState := "active"
	tSettingsMetaBaseId := "polycom_vvx_500.json"
	tSettingsMetaBaseName := "Polycom VVX 500"

	tTemplateLineMetaBaseId := "polycom_vvx.json"
	tTemplateLineMetaBaseName := "Polycom VVX line appearances"
	tlineMetaBase := platformclientv2.Domainentityref{
		Id:   &tTemplateLineMetaBaseId,
		Name: &tTemplateLineMetaBaseName,
	}
	functionCalls := make([]string, 0)

	pbsProxy := &phoneBaseProxy{}

	pbsProxy.getPhoneBaseSettingTemplateAttr = func(ctx context.Context, p *phoneBaseProxy, phoneBaseSettingsId string) (*platformclientv2.Phonebase, *platformclientv2.APIResponse, error) {
		functionCalls = append(functionCalls, "getPhoneBaseSettingTemplateAttr")
		apiResponse := platformclientv2.APIResponse{
			StatusCode: http.StatusOK,
		}

		lineBase := []platformclientv2.Linebase{
			{
				LineMetaBase: &tlineMetaBase,
			},
		}

		phoneBaseSettings := platformclientv2.Phonebase{
			State: &tSettingsState,
			Lines: &lineBase,
			Name:  &tSettingsName,
			PhoneMetaBase: &platformclientv2.Domainentityref{
				Id:   &tSettingsMetaBaseId,
				Name: &tSettingsMetaBaseName,
			},
		}
		return &phoneBaseSettings, &apiResponse, nil
	}

	pbsProxy.getPhoneBaseSettingAttr = func(ctx context.Context, p *phoneBaseProxy, phoneBaseSettingsId string) (*platformclientv2.Phonebase, *platformclientv2.APIResponse, error) {
		functionCalls = append(functionCalls, "getPhoneBaseSettingAttr")
		apiResponse := platformclientv2.APIResponse{
			StatusCode: http.StatusOK,
		}

		lineBase := []platformclientv2.Linebase{
			{
				Id:           &tLineId,
				State:        &tSettingsState,
				LineMetaBase: &tlineMetaBase,
			},
		}

		phoneBaseSettings := platformclientv2.Phonebase{
			State: &tSettingsState,
			Lines: &lineBase,
			PhoneMetaBase: &platformclientv2.Domainentityref{
				Id:   &tSettingsMetaBaseId,
				Name: &tSettingsMetaBaseName,
			},
			Name: &tSettingsName,
		}
		return &phoneBaseSettings, &apiResponse, nil
	}

	pbsProxy.putPhoneBaseSettingAttr = func(ctx context.Context, p *phoneBaseProxy, phoneBaseSettingsId string, body platformclientv2.Phonebase) (*platformclientv2.Phonebase, *platformclientv2.APIResponse, error) {
		functionCalls = append(functionCalls, "putPhoneBaseSettingAttr")
		assert.Equal(t, tSettingsName, *body.Name)
		assert.Equal(t, tSettingsMetaBaseId, *body.PhoneMetaBase.Id)

		// TODO fill out rest of assert commands

		body.Id = &tPhoneBaseSettingsId
		apiResponse := platformclientv2.APIResponse{
			StatusCode: http.StatusOK,
		}

		return &body, &apiResponse, nil
	}

	internalProxy = pbsProxy
	defer func() {
		internalProxy = nil
	}()

	ctx := context.Background()
	gcloud := &provider.ProviderMeta{ClientConfig: &platformclientv2.Configuration{}}

	// Grab our defined schema
	resourceSchema := ResourcePhoneBaseSettings().Schema

	// Setup a map of values
	resourceDataMap := buildPhoneBaseSettingsResource(
		tSettingsName, tSettingsMetaBaseId,
	)

	d := schema.TestResourceDataRaw(t, resourceSchema, resourceDataMap)
	d.SetId(tPhoneBaseSettingsId)

	diag := updatePhoneBaseSettings(ctx, d, gcloud)
	assert.Equal(t, false, diag.HasError(), "createPhoneBaseSettings returned error")
	assert.Equal(t, []string{
		"getPhoneBaseSettingAttr",
		"getPhoneBaseSettingTemplateAttr",
		"putPhoneBaseSettingAttr",
		"getPhoneBaseSettingAttr", // This is called twice, once for the read function and once for the custom diff function
		"getPhoneBaseSettingAttr",
	}, functionCalls)
	assert.Equal(t, tPhoneBaseSettingsId, d.Id())
	assert.Equal(t, tSettingsName, d.Get("name").(string))
	assert.Equal(t, tSettingsMetaBaseId, d.Get("phone_meta_base_id").(string))
}

func buildPhoneBaseSettingsResource(
	settingsName string,
	settingsMetaBaseId string,
) map[string]interface{} {
	resourceDataMap := map[string]interface{}{
		"name":               settingsName,
		"phone_meta_base_id": settingsMetaBaseId,
	}

	return resourceDataMap
}
