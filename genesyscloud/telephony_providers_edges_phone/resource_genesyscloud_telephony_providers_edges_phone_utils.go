package telephony_providers_edges_phone

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	gcloud "terraform-provider-genesyscloud/genesyscloud"
	lists "terraform-provider-genesyscloud/genesyscloud/util/lists"
	"terraform-provider-genesyscloud/genesyscloud/util/resourcedata"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v115/platformclientv2"
)

type PhoneConfig struct {
	PhoneRes            string
	Name                string
	State               string
	SiteId              string
	PhoneBaseSettingsId string
	LineAddresses       []string
	WebRtcUserId        string
	Depends_on          string
}

func getPhoneFromResourceData(ctx context.Context, pp *phoneProxy, d *schema.ResourceData) (*platformclientv2.Phone, error) {
	phoneConfig := &platformclientv2.Phone{
		Name:  platformclientv2.String(d.Get("name").(string)),
		State: platformclientv2.String(d.Get("state").(string)),
		Site:  gcloud.BuildSdkDomainEntityRef(d, "site_id"),
		PhoneBaseSettings: &platformclientv2.Phonebasesettings{
			Id: buildSdkPhoneBaseSettings(d, "phone_base_settings_id").Id,
		},
		Capabilities: buildSdkCapabilities(d),
	}

	// Line base settings and lines
	var err error
	lineBaseSettingsID := d.Get("line_base_settings_id").(string)
	if lineBaseSettingsID == "" {
		lineBaseSettingsID, err = getLineBaseSettingsID(ctx, pp, *phoneConfig.PhoneBaseSettings.Id)
		if err != nil {
			return nil, fmt.Errorf("ailed to get line base settings for %s: %s", *phoneConfig.Name, err)
		}
	}
	lineBaseSettings := &platformclientv2.Domainentityref{Id: &lineBaseSettingsID}
	lines, isStandalone := buildSdkLines(ctx, pp, d, lineBaseSettings)
	phoneConfig.LineBaseSettings = lineBaseSettings
	phoneConfig.Lines = lines

	// phone meta base
	phoneMetaBaseId, err := getPhoneMetaBaseId(ctx, pp, *phoneConfig.PhoneBaseSettings.Id)
	if err != nil {
		return nil, fmt.Errorf("failed to get phone meta base for %s: %s", *phoneConfig.Name, err)
	}
	phoneMetaBase := &platformclientv2.Domainentityref{
		Id: &phoneMetaBaseId,
	}
	phoneConfig.PhoneMetaBase = phoneMetaBase

	if isStandalone {
		phoneConfig.Properties = &map[string]interface{}{
			"phone_standalone": &map[string]interface{}{
				"value": &map[string]interface{}{
					"instance": true,
				},
			},
		}
	}

	webRtcUserId := d.Get("web_rtc_user_id")
	if webRtcUserId != "" {
		phoneConfig.WebRtcUser = gcloud.BuildSdkDomainEntityRef(d, "web_rtc_user_id")
	}

	return phoneConfig, nil
}

func getLineBaseSettingsID(ctx context.Context, pp *phoneProxy, phoneBaseSettingsId string) (string, error) {
	phoneBase, err := pp.getPhoneBaseSetting(ctx, phoneBaseSettingsId)
	if err != nil {
		return "", err
	}
	if len(*phoneBase.Lines) == 0 {
		return "", nil
	}
	return *(*phoneBase.Lines)[0].Id, nil
}

func assignUserToWebRtcPhone(ctx context.Context, pp *phoneProxy, userId string) diag.Diagnostics {
	stationId := ""
	stationIsAssociated := false

	retryErr := gcloud.WithRetries(ctx, 60*time.Second, func() *retry.RetryError {
		station, retryable, err := pp.getStationOfUser(ctx, userId)
		if err != nil && !retryable {
			return retry.NonRetryableError(fmt.Errorf("error requesting stations: %s", err))
		}
		if retryable {
			return retry.RetryableError(fmt.Errorf("no stations found with userID %v", userId))
		}

		stationId = *station.Id
		stationIsAssociated = *station.Status == "ASSOCIATED"

		return nil
	})
	if retryErr != nil {
		return retryErr
	}

	diagErr := gcloud.RetryWhen(gcloud.IsStatus400, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
		if stationIsAssociated {
			log.Printf("Disassociating user from phone station %s", stationId)
			if resp, err := pp.unassignUserFromStation(ctx, stationId); err != nil {
				return resp, diag.Errorf("Error unassigning user from station %s: %v", stationId, err)
			}
		}

		resp, putErr := pp.assignUserToStation(ctx, userId, stationId)
		if putErr != nil {
			return resp, diag.Errorf("Failed to assign user %v to the station %s: %s", userId, stationId, putErr)
		}
		return resp, nil
	})
	if diagErr != nil {
		return diagErr
	}

	return nil
}

func buildSdkPhoneBaseSettings(d *schema.ResourceData, idAttr string) *platformclientv2.Phonebasesettings {
	idVal := d.Get(idAttr).(string)
	if idVal == "" {
		return nil
	}
	return &platformclientv2.Phonebasesettings{Id: &idVal}
}

func getPhoneMetaBaseId(ctx context.Context, pp *phoneProxy, phoneBaseSettingsId string) (string, error) {
	phoneBase, err := pp.getPhoneBaseSetting(ctx, phoneBaseSettingsId)
	if err != nil {
		return "", err
	}

	return *phoneBase.PhoneMetaBase.Id, nil
}

func flattenPhoneLines(lines *[]platformclientv2.Line) []string {
	if lines == nil {
		return nil
	}

	lineAddressList := []string{}
	for i := 0; i < len(*lines); i++ {
		line := (*lines)[i]
		did := ""
		if k := (*line.Properties)["station_identity_address"]; k != nil {
			didI := k.(map[string]interface{})["value"].(map[string]interface{})["instance"]
			if didI != nil {
				did = didI.(string)
			}
		}

		if len(did) == 0 {
			continue
		}
		lineAddressList = append(lineAddressList, did)
	}

	return lineAddressList
}

func flattenPhoneCapabilities(capabilities *platformclientv2.Phonecapabilities) []interface{} {
	if capabilities == nil {
		return nil
	}

	capabilitiesMap := make(map[string]interface{})
	resourcedata.SetMapValueIfNotNil(capabilitiesMap, "provisions", capabilities.Provisions)
	resourcedata.SetMapValueIfNotNil(capabilitiesMap, "registers", capabilities.Registers)
	resourcedata.SetMapValueIfNotNil(capabilitiesMap, "dual_registers", capabilities.DualRegisters)
	resourcedata.SetMapValueIfNotNil(capabilitiesMap, "hardware_id_type", capabilities.HardwareIdType)
	resourcedata.SetMapValueIfNotNil(capabilitiesMap, "allow_reboot", capabilities.AllowReboot)
	resourcedata.SetMapValueIfNotNil(capabilitiesMap, "no_rebalance", capabilities.NoRebalance)
	resourcedata.SetMapValueIfNotNil(capabilitiesMap, "no_cloud_provisioning", capabilities.NoCloudProvisioning)
	resourcedata.SetMapValueIfNotNil(capabilitiesMap, "media_codecs", capabilities.MediaCodecs)
	resourcedata.SetMapValueIfNotNil(capabilitiesMap, "cdm", capabilities.Cdm)

	return []interface{}{capabilitiesMap}
}

func buildSdkLines(ctx context.Context, pp *phoneProxy, d *schema.ResourceData, lineBaseSettings *platformclientv2.Domainentityref) (linesPtr *[]platformclientv2.Line, isStandAlone bool) {
	lines := []platformclientv2.Line{}
	isStandAlone = false

	lineAddresses, ok := d.GetOk("line_addresses")
	lineStringList := lists.InterfaceListToStrings(lineAddresses.([]interface{}))

	// If line_addresses is not provided, phone is not standalone
	if !ok || len(lineStringList) == 0 {
		lineName := "line_" + *lineBaseSettings.Id
		line := platformclientv2.Line{
			Name:             &lineName,
			LineBaseSettings: lineBaseSettings,
		}

		lineId, err := getLineIdByPhoneId(ctx, pp, d.Id())
		if err != nil {
			log.Printf("Failed to retrieve ID for phone %s: %v", d.Id(), err)
		} else {
			line.Id = &lineId
		}

		lines = append(lines, line)

		linesPtr = &lines
		return
	}

	for i := 0; i < len(lineStringList); i++ {
		lineName := "line_" + *lineBaseSettings.Id + "_" + strconv.Itoa(i+1)
		properties := map[string]interface{}{
			"station_identity_address": &map[string]interface{}{
				"value": &map[string]interface{}{
					"instance": (lineStringList)[i],
				},
			},
		}
		lines = append(lines, platformclientv2.Line{
			Name:             &lineName,
			LineBaseSettings: lineBaseSettings,
			Properties:       &properties,
		})
	}

	linesPtr = &lines
	isStandAlone = true

	return
}

func getLineIdByPhoneId(ctx context.Context, pp *phoneProxy, phoneId string) (string, error) {
	phone, _, err := pp.getPhoneById(ctx, phoneId)
	if err != nil {
		return "", err
	}
	if phone.Lines != nil && len(*phone.Lines) > 0 {
		return *(*phone.Lines)[0].Id, nil
	}
	return "", fmt.Errorf("could not access line ID for phone %s", phoneId)
}

func buildSdkCapabilities(d *schema.ResourceData) *platformclientv2.Phonecapabilities {
	if capabilities := d.Get("capabilities").([]interface{}); capabilities != nil {
		sdkPhoneCapabilities := platformclientv2.Phonecapabilities{}
		if len(capabilities) > 0 {
			if _, ok := capabilities[0].(map[string]interface{}); !ok {
				return nil
			}
			capabilitiesMap := capabilities[0].(map[string]interface{})

			sdkPhoneCapabilities = platformclientv2.Phonecapabilities{
				Provisions:          platformclientv2.Bool(capabilitiesMap["provisions"].(bool)),
				Registers:           platformclientv2.Bool(capabilitiesMap["registers"].(bool)),
				DualRegisters:       platformclientv2.Bool(capabilitiesMap["dual_registers"].(bool)),
				AllowReboot:         platformclientv2.Bool(capabilitiesMap["allow_reboot"].(bool)),
				NoRebalance:         platformclientv2.Bool(capabilitiesMap["no_rebalance"].(bool)),
				NoCloudProvisioning: platformclientv2.Bool(capabilitiesMap["no_cloud_provisioning"].(bool)),
				Cdm:                 platformclientv2.Bool(capabilitiesMap["cdm"].(bool)),
			}

			// Hardware ID type
			if checkHardwareIdType := capabilitiesMap["hardware_id_type"].(string); len(checkHardwareIdType) > 0 {
				sdkPhoneCapabilities.HardwareIdType = &checkHardwareIdType
			}

			// Media codecs
			mediaCodecs := make([]string, 0)
			if checkMediaCodecs := capabilitiesMap["media_codecs"].([]interface{}); len(checkMediaCodecs) > 0 {
				for _, codec := range checkMediaCodecs {
					mediaCodecs = append(mediaCodecs, fmt.Sprintf("%v", codec))
				}
			}

			sdkPhoneCapabilities.MediaCodecs = &mediaCodecs
		}
		return &sdkPhoneCapabilities
	}
	return nil
}

func GeneratePhoneResourceWithCustomAttrs(config *PhoneConfig, otherAttrs ...string) string {
	lineStrs := make([]string, len(config.LineAddresses))
	for i, val := range config.LineAddresses {
		lineStrs[i] = fmt.Sprintf("\"%s\"", val)
	}

	webRtcUser := ""
	if len(config.WebRtcUserId) != 0 {
		webRtcUser = fmt.Sprintf(`web_rtc_user_id = %s`, config.WebRtcUserId)
	}

	finalConfig := fmt.Sprintf(`resource "genesyscloud_telephony_providers_edges_phone" "%s" {
		name = "%s"
		state = "%s"
		site_id = %s
		phone_base_settings_id = %s
		line_addresses = [%s]
		depends_on=[%s]
		%s
		%s
	}
	`, config.PhoneRes,
		config.Name,
		config.State,
		config.SiteId,
		config.PhoneBaseSettingsId,
		strings.Join(lineStrs, ","),
		config.Depends_on,
		webRtcUser,
		strings.Join(otherAttrs, "\n"),
	)

	return finalConfig
}

func TestVerifyWebRtcPhoneDestroyed(state *terraform.State) error {
	edgesAPI := platformclientv2.NewTelephonyProvidersEdgeApi()
	for _, rs := range state.RootModule().Resources {
		if rs.Type != "genesyscloud_telephony_providers_edges_phone" {
			continue
		}

		phone, resp, err := edgesAPI.GetTelephonyProvidersEdgesPhone(rs.Primary.ID)
		if phone != nil {
			return fmt.Errorf("phone (%s) still exists", rs.Primary.ID)
		} else if gcloud.IsStatus404(resp) {
			// Phone not found as expected
			continue
		} else {
			// Unexpected error
			return fmt.Errorf("unexpected error: %s", err)
		}
	}
	//Success. Phone destroyed
	return nil
}
