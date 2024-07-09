package telephony_providers_edges_phone

import (
	"context"
	"fmt"
	"hash/fnv"
	"log"
	"strconv"
	"strings"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	"terraform-provider-genesyscloud/genesyscloud/util/resourcedata"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"
)

type PhoneConfig struct {
	PhoneRes            string
	Name                string
	State               string
	SiteId              string
	PhoneBaseSettingsId string
	WebRtcUserId        string
	DependsOn           string
}

func getPhoneFromResourceData(ctx context.Context, pp *phoneProxy, d *schema.ResourceData) (*platformclientv2.Phone, error) {
	phoneConfig := &platformclientv2.Phone{
		Name:       platformclientv2.String(d.Get("name").(string)),
		State:      platformclientv2.String(d.Get("state").(string)),
		Site:       util.BuildSdkDomainEntityRef(d, "site_id"),
		Properties: util.BuildTelephonyProperties(d),
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
			return phoneConfig, fmt.Errorf("failed to get line base settings for %s: %s", *phoneConfig.Name, err)
		}
	}
	lineBaseSettings := &platformclientv2.Domainentityref{Id: &lineBaseSettingsID}
	lines, isStandalone, lineError := buildSdkLines(ctx, pp, d, lineBaseSettings)
	if lineError != nil {
		return phoneConfig, fmt.Errorf("failed to create lines for %s: %s", *phoneConfig.Name, lineError)
	}
	phoneConfig.LineBaseSettings = lineBaseSettings
	phoneConfig.Lines = lines

	// phone meta base
	phoneMetaBaseId, err := getPhoneMetaBaseId(ctx, pp, *phoneConfig.PhoneBaseSettings.Id)
	if err != nil {
		return phoneConfig, fmt.Errorf("failed to get phone meta base for %s: %s", *phoneConfig.Name, err)
	}
	phoneMetaBase := &platformclientv2.Domainentityref{
		Id: &phoneMetaBaseId,
	}
	phoneConfig.PhoneMetaBase = phoneMetaBase

	if isStandalone {
		if phoneConfig.Properties == nil {
			phoneConfig.Properties = &map[string]interface{}{}
		}
		phoneStandalone := map[string]interface{}{
			"value": &map[string]interface{}{
				"instance": true,
			},
		}
		(*phoneConfig.Properties)["phone_standalone"] = phoneStandalone
	}

	webRtcUserId := d.Get("web_rtc_user_id")
	if webRtcUserId != "" {
		phoneConfig.WebRtcUser = util.BuildSdkDomainEntityRef(d, "web_rtc_user_id")
	}

	return phoneConfig, nil
}

func getLineBaseSettingsID(ctx context.Context, pp *phoneProxy, phoneBaseSettingsId string) (string, error) {
	phoneBase, _, err := pp.getPhoneBaseSetting(ctx, phoneBaseSettingsId)
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

	retryErr := util.WithRetries(ctx, 60*time.Second, func() *retry.RetryError {
		station, retryable, resp, err := pp.getStationOfUser(ctx, userId)
		if err != nil && !retryable {
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(resourceName, fmt.Sprintf("error requesting stations: %s", err), resp))
		}
		if retryable {
			return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(resourceName, fmt.Sprintf("no stations found with userID %v", userId), resp))
		}

		stationId = *station.Id
		stationIsAssociated = *station.Status == "ASSOCIATED"

		return nil
	})
	if retryErr != nil {
		return retryErr
	}

	diagErr := util.RetryWhen(util.IsStatus400, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
		if stationIsAssociated {
			log.Printf("Disassociating user from phone station %s", stationId)
			if resp, err := pp.unassignUserFromStation(ctx, stationId); err != nil {
				return resp, util.BuildAPIDiagnosticError(resourceName, fmt.Sprintf("Error unassigning user from station %s: %v", stationId, err), resp)
			}
		}

		resp, putErr := pp.assignUserToStation(ctx, userId, stationId)
		if putErr != nil {
			return resp, util.BuildAPIDiagnosticError(resourceName, fmt.Sprintf("Failed to assign user %v to the station %s: %s", userId, stationId, putErr), resp)
		}

		resp, putErr = pp.assignStationAsDefault(ctx, userId, stationId)
		if putErr != nil {
			return resp, util.BuildAPIDiagnosticError(resourceName, fmt.Sprintf("Failed to assign Station %v as the default station for user %s: %s", stationId, userId, putErr), resp)
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
	phoneBase, _, err := pp.getPhoneBaseSetting(ctx, phoneBaseSettingsId)
	if err != nil {
		return "", err
	}

	return *phoneBase.PhoneMetaBase.Id, nil
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

func buildSdkLines(ctx context.Context, pp *phoneProxy, d *schema.ResourceData, lineBaseSettings *platformclientv2.Domainentityref) (linesPtr *[]platformclientv2.Line, isStandAlone bool, err error) {
	lines := []platformclientv2.Line{}
	lineAddress, remoteAddress := getLineProperties(d)
	if len(*lineAddress) > 0 && len(*remoteAddress) > 0 {
		return linesPtr, false, fmt.Errorf("remote stations cannot be standalone phones, line_address and remote_address cannot exist at the same time")
	}
	if len(*lineAddress) > 0 {
		linesPtr = createStandalonePhoneLines(*lineAddress, &lines, lineBaseSettings)
		isStandAlone = true
		return linesPtr, isStandAlone, nil
	}
	if len(*remoteAddress) > 0 {
		linesPtr = createNonStandalonePhoneLine(*remoteAddress, &lines, lineBaseSettings)
		isStandAlone = false
		return linesPtr, isStandAlone, nil
	}
	// If line_addresses is not provided, phone is not standalone
	hasher := fnv.New32()
	hasher.Write([]byte(d.Get("name").(string)))
	lineName := "line_" + *lineBaseSettings.Id + fmt.Sprintf("%x", hasher.Sum32())
	line := platformclientv2.Line{
		Name:             &lineName,
		LineBaseSettings: lineBaseSettings,
	}
	// If this function is invoked on a phone create, the ID won't exist yet
	if d.Id() != "" {
		lineId, err := getLineIdByPhoneId(ctx, pp, d.Id())
		if err != nil {
			log.Printf("Failed to retrieve ID for phone %s: %v", d.Id(), err)
		} else {
			line.Id = &lineId
		}
	}
	lines = append(lines, line)
	linesPtr = &lines
	isStandAlone = false
	return linesPtr, isStandAlone, nil
}

func getLineProperties(d *schema.ResourceData) (*[]interface{}, *[]interface{}) {
	lineAddress := make([]interface{}, 0)
	remoteAddress := make([]interface{}, 0)
	linePropertiesMap := make(map[string]interface{})
	if linePropertiesObject, ok := d.Get("line_properties").([]interface{}); ok && len(linePropertiesObject) > 0 {
		linePropertiesMap = linePropertiesObject[0].(map[string]interface{})
	}
	if lineAddressObject, ok := linePropertiesMap["line_address"].([]interface{}); ok {
		lineAddress = lineAddressObject
	}
	if remoteAddressObject, ok := linePropertiesMap["remote_address"].([]interface{}); ok {
		remoteAddress = remoteAddressObject
	}
	return &lineAddress, &remoteAddress
}

func generatePhoneProperties(hardware_id string) string {
	// A random selection of properties
	return "properties = " + util.GenerateJsonEncodedProperties(
		util.GenerateJsonProperty(
			"phone_hardwareId", util.GenerateJsonObject(
				util.GenerateJsonProperty(
					"value", util.GenerateJsonObject(
						util.GenerateJsonProperty("instance", strconv.Quote(hardware_id)),
					)))),
	)
}
func createNonStandalonePhoneLine(remoteAddress []interface{}, linesPtr *[]platformclientv2.Line, lineBaseSettings *platformclientv2.Domainentityref) *[]platformclientv2.Line {
	lines := *linesPtr
	for i, eachAddress := range remoteAddress {
		lineName := "line_" + *lineBaseSettings.Id + "_" + strconv.Itoa(i+1)
		properties := map[string]interface{}{
			"station_remote_address": &map[string]interface{}{
				"value": &map[string]interface{}{
					"instance": eachAddress.(string),
				},
			},
		}
		lines = append(lines, platformclientv2.Line{
			Name:             &lineName,
			LineBaseSettings: lineBaseSettings,
			Properties:       &properties,
		})
	}
	return &lines
}
func createStandalonePhoneLines(lineAddress []interface{}, linesPtr *[]platformclientv2.Line, lineBaseSettings *platformclientv2.Domainentityref) *[]platformclientv2.Line {
	lines := *linesPtr
	for i, eachLineAddress := range lineAddress {
		lineName := "line_" + *lineBaseSettings.Id + "_" + strconv.Itoa(i+1)
		properties := map[string]interface{}{
			"station_identity_address": &map[string]interface{}{
				"value": &map[string]interface{}{
					"instance": eachLineAddress.(string),
				},
			},
		}
		lines = append(lines, platformclientv2.Line{
			Name:             &lineName,
			LineBaseSettings: lineBaseSettings,
			Properties:       &properties,
		})
	}
	return &lines
}

func flattenLines(phoneLines *[]platformclientv2.Line) []interface{} {
	if len(*phoneLines) == 0 {
		return nil
	}
	lineAddressList := []string{}
	remoteAddressList := []string{}
	linePropertiesMap := make(map[string]interface{})

	for _, phoneLine := range *phoneLines {
		if phoneLine.Properties == nil {
			continue
		}
		if idAddressKey := (*phoneLine.Properties)["station_identity_address"]; idAddressKey != nil {
			didI := idAddressKey.(map[string]interface{})["value"].(map[string]interface{})["instance"]
			if didI != nil && len(didI.(string)) > 0 {
				did := didI.(string)
				lineAddressList = append(lineAddressList, did)

			}
		}
		if remoteAddressKey := (*phoneLine.Properties)["station_remote_address"]; remoteAddressKey != nil {
			remoteAddress := remoteAddressKey.(map[string]interface{})["value"].(map[string]interface{})["instance"]
			if remoteAddress != nil && len(remoteAddress.(string)) > 0 {
				remoteAddressStr := remoteAddress.(string)
				remoteAddressList = append(remoteAddressList, remoteAddressStr)
			}
		}

	}

	if len(lineAddressList) > 0 {
		resourcedata.SetMapValueIfNotNil(linePropertiesMap, "line_address", &lineAddressList)
	}
	if len(remoteAddressList) > 0 {
		resourcedata.SetMapValueIfNotNil(linePropertiesMap, "remote_address", &remoteAddressList)
	}
	if len(linePropertiesMap) > 0 {
		return []interface{}{linePropertiesMap}

	}
	return nil
}

func generateLineProperties(lineAddress string, remoteAddress string) string {
	if lineAddress == "" {
		return fmt.Sprintf(`
		line_properties {
			remote_address = [%s]
		}
	`, remoteAddress)
	}

	if remoteAddress == "" {
		return fmt.Sprintf(`
		line_properties {
			line_address = [%s]
		}
	`, lineAddress)
	}

	return fmt.Sprintf(`
	line_properties {
		line_address = [%s]
		remote_address = [%s]
	}
`, lineAddress, remoteAddress)
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

	webRtcUser := ""
	if len(config.WebRtcUserId) != 0 {
		webRtcUser = fmt.Sprintf(`web_rtc_user_id = %s`, config.WebRtcUserId)
	}

	finalConfig := fmt.Sprintf(`resource "genesyscloud_telephony_providers_edges_phone" "%s" {
		name = "%s"
		state = "%s"
		site_id = %s
		phone_base_settings_id = %s
		depends_on=[%s]
		%s
		%s
	}
	`, config.PhoneRes,
		config.Name,
		config.State,
		config.SiteId,
		config.PhoneBaseSettingsId,
		config.DependsOn,
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
		} else if util.IsStatus404(resp) {
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
