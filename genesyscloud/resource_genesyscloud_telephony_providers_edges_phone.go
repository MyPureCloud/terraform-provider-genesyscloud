package genesyscloud

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"terraform-provider-genesyscloud/genesyscloud/consistency_checker"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	lists "terraform-provider-genesyscloud/genesyscloud/util/lists"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/mypurecloud/platform-client-sdk-go/v105/platformclientv2"
)

var (
	phoneCapabilities = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"provisions": {
				Description: "Provisions",
				Type:        schema.TypeBool,
				Optional:    true,
			},
			"registers": {
				Description: "Registers",
				Type:        schema.TypeBool,
				Optional:    true,
			},
			"dual_registers": {
				Description: "Dual Registers",
				Type:        schema.TypeBool,
				Optional:    true,
			},
			"hardware_id_type": {
				Description: "HardwareId Type",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"allow_reboot": {
				Description: "Allow Reboot",
				Type:        schema.TypeBool,
				Optional:    true,
			},
			"no_rebalance": {
				Description: "No Rebalance",
				Type:        schema.TypeBool,
				Optional:    true,
			},
			"no_cloud_provisioning": {
				Description: "No Cloud Provisioning",
				Type:        schema.TypeBool,
				Optional:    true,
			},
			"media_codecs": {
				Description: "Media Codecs",
				Type:        schema.TypeList,
				Optional:    true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringInSlice([]string{"audio/opus", "audio/pcmu", "audio/pcma", "audio/g729", "audio/g722"}, false),
				},
			},
			"cdm": {
				Description: "CDM",
				Type:        schema.TypeBool,
				Optional:    true,
			},
		},
	}
)

func ResourcePhone() *schema.Resource {
	return &schema.Resource{
		Description: "Genesys Cloud Phone",

		CreateContext: CreateWithPooledClient(createPhone),
		ReadContext:   ReadWithPooledClient(readPhone),
		UpdateContext: UpdateWithPooledClient(updatePhone),
		DeleteContext: DeleteWithPooledClient(deletePhone),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "The name of the entity.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"state": {
				Description:  "Indicates if the resource is active, inactive, or deleted. Valid values: active, inactive, deleted.",
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "active",
				ValidateFunc: validation.StringInSlice([]string{"active", "inactive", "deleted"}, false),
			},
			"site_id": {
				Description: "The site ID associated to the phone.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"phone_base_settings_id": {
				Description: "Phone Base Settings ID.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"line_base_settings_id": {
				Description: "Line Base Settings ID.",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
			},
			"phone_meta_base_id": {
				Description: "Phone Meta Base ID.",
				Type:        schema.TypeString,
				Computed:    true,
				Optional:    true,
			},
			"web_rtc_user_id": {
				Description: "Web RTC User ID. This is necessary when creating a Web RTC phone. This user will be assigned to the phone after it is created.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"line_addresses": {
				Description: "Ordered list of Line DIDs for standalone phones.  Each phone number must be in an E.164 phone number format.",
				Type:        schema.TypeList,
				Optional:    true,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString, ValidateDiagFunc: ValidatePhoneNumber},
			},
			"capabilities": {
				Description: "Phone Capabilities.",
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Computed:    true,
				Elem:        phoneCapabilities,
			},
		},
	}
}

func getLineBaseSettingsID(api *platformclientv2.TelephonyProvidersEdgeApi, phoneBaseSettingsId string) (string, error) {
	phoneBase, _, err := api.GetTelephonyProvidersEdgesPhonebasesetting(phoneBaseSettingsId)
	if err != nil {
		return "", err
	}
	if len(*phoneBase.Lines) == 0 {
		return "", nil
	}
	return *(*phoneBase.Lines)[0].Id, nil
}

func createPhone(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)
	state := d.Get("state").(string)
	site := BuildSdkDomainEntityRef(d, "site_id")
	phoneBaseSettings := buildSdkPhoneBaseSettings(d, "phone_base_settings_id")

	capabilities := buildSdkCapabilities(d)
	webRtcUserId := d.Get("web_rtc_user_id")

	sdkConfig := meta.(*ProviderMeta).ClientConfig
	edgesAPI := platformclientv2.NewTelephonyProvidersEdgeApiWithConfig(sdkConfig)

	var err error
	lineBaseSettingsID := d.Get("line_base_settings_id").(string)
	if lineBaseSettingsID == "" {
		lineBaseSettingsID, err = getLineBaseSettingsID(edgesAPI, *phoneBaseSettings.Id)
		if err != nil {
			return diag.Errorf("Failed to get line base settings for %s: %s", name, err)
		}
	}

	lineBaseSettings := &platformclientv2.Domainentityref{Id: &lineBaseSettingsID}
	lines, isStandalone := buildSdkLines(d, lineBaseSettings, edgesAPI)

	phoneMetaBaseId, err := getPhoneMetaBaseId(edgesAPI, *phoneBaseSettings.Id)
	if err != nil {
		return diag.Errorf("Failed to get phone meta base for %s: %s", name, err)
	}

	phoneMetaBase := &platformclientv2.Domainentityref{
		Id: &phoneMetaBaseId,
	}

	//Have to create a phonebasesettings object now as of version v90.  Used to be a domain ref but the engineering team changed the type in the swagger def
	phoneSettings := &platformclientv2.Phonebasesettings{
		Id: phoneBaseSettings.Id,
	}

	createPhone := &platformclientv2.Phone{
		Name:              &name,
		State:             &state,
		Site:              site,
		PhoneBaseSettings: phoneSettings,
		LineBaseSettings:  lineBaseSettings,
		PhoneMetaBase:     phoneMetaBase,
		Lines:             lines,
		Capabilities:      capabilities,
	}

	if isStandalone {
		createPhone.Properties = &map[string]interface{}{
			"phone_standalone": &map[string]interface{}{
				"value": &map[string]interface{}{
					"instance": true,
				},
			},
		}
	}

	if webRtcUserId != "" {
		createPhone.WebRtcUser = BuildSdkDomainEntityRef(d, "web_rtc_user_id")
	}

	log.Printf("Creating phone %s", name)
	diagErr := RetryWhen(IsStatus400, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
		phone, resp, err := edgesAPI.PostTelephonyProvidersEdgesPhones(*createPhone)
		if err != nil {
			return resp, diag.Errorf("Failed to create phone %s: %s", name, err)
		}

		d.SetId(*phone.Id)

		if webRtcUserId != "" {
			diagErr := assignUserToWebRtcPhone(ctx, sdkConfig, webRtcUserId.(string))
			if diagErr != nil {
				return resp, diagErr
			}
		}

		log.Printf("Created phone %s", *phone.Id)
		return nil, nil
	})
	if diagErr != nil {
		return diagErr
	}

	return readPhone(ctx, d, meta)
}

func readPhone(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*ProviderMeta).ClientConfig
	edgesAPI := platformclientv2.NewTelephonyProvidersEdgeApiWithConfig(sdkConfig)

	log.Printf("Reading phone %s", d.Id())
	return WithRetriesForRead(ctx, d, func() *retry.RetryError {
		currentPhone, resp, getErr := edgesAPI.GetTelephonyProvidersEdgesPhone(d.Id())
		if getErr != nil {
			if IsStatus404(resp) {
				return retry.RetryableError(fmt.Errorf("Failed to read phone %s: %s", d.Id(), getErr))
			}
			return retry.NonRetryableError(fmt.Errorf("Failed to read phone %s: %s", d.Id(), getErr))
		}

		cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourcePhone())
		d.Set("name", *currentPhone.Name)
		d.Set("state", *currentPhone.State)
		d.Set("site_id", *currentPhone.Site.Id)
		d.Set("phone_base_settings_id", *currentPhone.PhoneBaseSettings.Id)
		d.Set("line_base_settings_id", *currentPhone.LineBaseSettings.Id)
		if currentPhone.PhoneMetaBase != nil {
			d.Set("phone_meta_base_id", *currentPhone.PhoneMetaBase.Id)
		}

		if currentPhone.WebRtcUser != nil {
			d.Set("web_rtc_user_id", *currentPhone.WebRtcUser.Id)
		}

		if currentPhone.Lines != nil {
			d.Set("line_addresses", flattenPhoneLines(currentPhone.Lines))
		}

		if currentPhone.Capabilities != nil {
			d.Set("capabilities", flattenPhoneCapabilities(currentPhone.Capabilities))
		}

		log.Printf("Read phone %s %s", d.Id(), *currentPhone.Name)
		return cc.CheckState()
	})
}

func assignUserToWebRtcPhone(ctx context.Context, sdkConfig *platformclientv2.Configuration, userId string) diag.Diagnostics {
	stationsAPI := platformclientv2.NewStationsApiWithConfig(sdkConfig)
	stationId := ""
	stationIsAssociated := false

	retryErr := WithRetries(ctx, 60*time.Second, func() *retry.RetryError {
		const pageSize = 100
		const pageNum = 1
		stations, _, getErr := stationsAPI.GetStations(pageSize, pageNum, "", "", "", userId, "", "")
		if getErr != nil {
			return retry.NonRetryableError(fmt.Errorf("Error requesting stations: %s", getErr))
		}

		if stations.Entities == nil || len(*stations.Entities) == 0 {
			return retry.RetryableError(fmt.Errorf("No stations found with userID %v", userId))
		}

		stationId = *(*stations.Entities)[0].Id
		stationIsAssociated = *(*stations.Entities)[0].Status == "ASSOCIATED"

		return nil
	})
	if retryErr != nil {
		return retryErr
	}

	usersAPI := platformclientv2.NewUsersApiWithConfig(sdkConfig)

	diagErr := RetryWhen(IsStatus400, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
		if stationIsAssociated {
			log.Printf("Disassociating user from phone station %s", stationId)
			if resp, err := stationsAPI.DeleteStationAssociateduser(stationId); err != nil {
				return resp, diag.Errorf("Error unassigning user from station %s: %v", stationId, err)
			}
		}

		resp, putErr := usersAPI.PutUserStationAssociatedstationStationId(userId, stationId)
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

func updatePhone(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)
	site := BuildSdkDomainEntityRef(d, "site_id")
	phoneBaseSettings := buildSdkPhoneBaseSettings(d, "phone_base_settings_id")
	phoneMetaBase := BuildSdkDomainEntityRef(d, "phone_meta_base_id")
	webRtcUserId := d.Get("web_rtc_user_id")

	sdkConfig := meta.(*ProviderMeta).ClientConfig
	edgesAPI := platformclientv2.NewTelephonyProvidersEdgeApiWithConfig(sdkConfig)

	var err error
	lineBaseSettingsID := d.Get("line_base_settings_id").(string)
	if lineBaseSettingsID == "" {
		lineBaseSettingsID, err = getLineBaseSettingsID(edgesAPI, *phoneBaseSettings.Id)
		if err != nil {
			return diag.Errorf("Failed to get line base settings for %s: %s", name, err)
		}
	}

	lineBaseSettings := &platformclientv2.Domainentityref{Id: &lineBaseSettingsID}
	lines, isStandalone := buildSdkLines(d, lineBaseSettings, edgesAPI)

	//Have to create a phonebasesettings object now as of version v90.  Used to be a domain ref but the engineering team changed the type in the swagger def
	phoneSettings := &platformclientv2.Phonebasesettings{
		Id: phoneBaseSettings.Id,
	}

	updatePhoneBody := &platformclientv2.Phone{
		Name:              &name,
		Site:              site,
		PhoneBaseSettings: phoneSettings,
		PhoneMetaBase:     phoneMetaBase,
		LineBaseSettings:  lineBaseSettings,
		Lines:             lines,
	}

	if isStandalone {
		updatePhoneBody.Properties = &map[string]interface{}{
			"phone_standalone": &map[string]interface{}{
				"value": &map[string]interface{}{
					"instance": true,
				},
			},
		}
	}

	if webRtcUserId != "" {
		updatePhoneBody.WebRtcUser = BuildSdkDomainEntityRef(d, "web_rtc_user_id")
	}

	log.Printf("Updating phone %s", name)

	phone, _, err := edgesAPI.PutTelephonyProvidersEdgesPhone(d.Id(), *updatePhoneBody)
	if err != nil {
		return diag.Errorf("Failed to update phone %s: %s", name, err)
	}

	log.Printf("Updated phone %s", *phone.Id)

	if webRtcUserId != "" {
		if d.HasChange("web_rtc_user_id") {
			diagErr := assignUserToWebRtcPhone(ctx, sdkConfig, webRtcUserId.(string))
			if diagErr != nil {
				return diagErr
			}
		}
	}

	return readPhone(ctx, d, meta)
}

func deletePhone(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*ProviderMeta).ClientConfig
	edgesAPI := platformclientv2.NewTelephonyProvidersEdgeApiWithConfig(sdkConfig)

	log.Printf("Deleting Phone")
	_, err := edgesAPI.DeleteTelephonyProvidersEdgesPhone(d.Id())

	/*
	  Adding a small sleep because when a phone is deleted, the station associated with the phone and the site
	  objects need time to disassociate from the phone. This eventual consistency problem was discovered during
	  building the GCX Now project.  Adding the sleep gives the platform time to settle down.
	*/
	time.Sleep(5 * time.Second)
	if err != nil {
		return diag.Errorf("Failed to delete phone: %s", err)
	}

	return WithRetries(ctx, 30*time.Second, func() *retry.RetryError {
		phone, resp, err := edgesAPI.GetTelephonyProvidersEdgesPhone(d.Id())
		if err != nil {
			if IsStatus404(resp) {
				// Phone deleted
				log.Printf("Deleted Phone %s", d.Id())
				return nil
			}
			return retry.NonRetryableError(fmt.Errorf("Error deleting Phone %s: %s", d.Id(), err))
		}

		if phone.State != nil && *phone.State == "deleted" {

			// phone deleted
			log.Printf("Deleted Phone %s", d.Id())
			return nil
		}

		return retry.RetryableError(fmt.Errorf("Phone %s still exists", d.Id()))
	})
}

func buildSdkPhoneBaseSettings(d *schema.ResourceData, idAttr string) *platformclientv2.Phonebasesettings {
	idVal := d.Get(idAttr).(string)
	if idVal == "" {
		return nil
	}
	return &platformclientv2.Phonebasesettings{Id: &idVal}
}

func getPhoneMetaBaseId(api *platformclientv2.TelephonyProvidersEdgeApi, phoneBaseSettingsId string) (string, error) {
	phoneBase, _, err := api.GetTelephonyProvidersEdgesPhonebasesetting(phoneBaseSettingsId)
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
	if capabilities.Provisions != nil {
		capabilitiesMap["provisions"] = *capabilities.Provisions
	}
	if capabilities.Registers != nil {
		capabilitiesMap["registers"] = *capabilities.Registers
	}
	if capabilities.DualRegisters != nil {
		capabilitiesMap["dual_registers"] = *capabilities.DualRegisters
	}
	if capabilities.HardwareIdType != nil {
		capabilitiesMap["hardware_id_type"] = *capabilities.HardwareIdType
	}
	if capabilities.AllowReboot != nil {
		capabilitiesMap["allow_reboot"] = *capabilities.AllowReboot
	}
	if capabilities.NoRebalance != nil {
		capabilitiesMap["no_rebalance"] = *capabilities.NoRebalance
	}
	if capabilities.NoCloudProvisioning != nil {
		capabilitiesMap["no_cloud_provisioning"] = *capabilities.NoCloudProvisioning
	}
	if capabilities.MediaCodecs != nil {
		capabilitiesMap["media_codecs"] = *capabilities.MediaCodecs
	}
	if capabilities.Cdm != nil {
		capabilitiesMap["cdm"] = *capabilities.Cdm
	}

	return []interface{}{capabilitiesMap}
}

func getAllPhones(_ context.Context, sdkConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	resources := make(resourceExporter.ResourceIDMetaMap)

	edgesAPI := platformclientv2.NewTelephonyProvidersEdgeApiWithConfig(sdkConfig)

	for pageNum := 1; ; pageNum++ {
		const pageSize = 100
		phones, _, getErr := edgesAPI.GetTelephonyProvidersEdgesPhones(pageNum, pageSize, "", "", "", "", "", "", "", "", "", "", "", "", "", nil, nil)
		if getErr != nil {
			return nil, diag.Errorf("Failed to get page of phones: %v", getErr)
		}

		if phones.Entities == nil || len(*phones.Entities) == 0 {
			break
		}

		for _, phone := range *phones.Entities {
			if phone.State != nil && *phone.State != "deleted" {
				resources[*phone.Id] = &resourceExporter.ResourceMeta{Name: *phone.Name}
			}
		}
	}

	return resources, nil
}

func PhoneExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: GetAllWithPooledClient(getAllPhones),
		RefAttrs: map[string]*resourceExporter.RefAttrSettings{
			"web_rtc_user_id":        {RefType: "genesyscloud_user"},
			"site_id":                {RefType: "genesyscloud_telephony_providers_edges_site"},
			"phone_base_settings_id": {RefType: "genesyscloud_telephony_providers_edges_phonebasesettings"},
		},
	}
}

func buildSdkLines(d *schema.ResourceData, lineBaseSettings *platformclientv2.Domainentityref, api *platformclientv2.TelephonyProvidersEdgeApi) (linesPtr *[]platformclientv2.Line, isStandAlone bool) {
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

		lineId, err := getLineIdByPhoneId(d.Id(), api)
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

func getLineIdByPhoneId(phoneId string, api *platformclientv2.TelephonyProvidersEdgeApi) (string, error) {
	phone, _, err := api.GetTelephonyProvidersEdgesPhone(phoneId)
	if err != nil {
		return "", err
	}
	if phone.Lines != nil && len(*phone.Lines) > 0 {
		return *(*phone.Lines)[0].Id, nil
	}
	return "", fmt.Errorf("Could not access line ID for phone %s", phoneId)
}

func buildSdkCapabilities(d *schema.ResourceData) *platformclientv2.Phonecapabilities {
	if capabilities := d.Get("capabilities").([]interface{}); capabilities != nil {
		sdkPhoneCapabilities := platformclientv2.Phonecapabilities{}
		if len(capabilities) > 0 {
			if _, ok := capabilities[0].(map[string]interface{}); !ok {
				return nil
			}
			capabilitiesMap := capabilities[0].(map[string]interface{})

			// Only set non-empty values.
			provisions := capabilitiesMap["provisions"].(bool)
			registers := capabilitiesMap["registers"].(bool)
			dualRegisters := capabilitiesMap["dual_registers"].(bool)
			var hardwareIdType string
			if checkHardwareIdType := capabilitiesMap["hardware_id_type"].(string); len(checkHardwareIdType) > 0 {
				hardwareIdType = checkHardwareIdType
			}
			allowReboot := capabilitiesMap["allow_reboot"].(bool)
			noRebalance := capabilitiesMap["no_rebalance"].(bool)
			noCloudProvisioning := capabilitiesMap["no_cloud_provisioning"].(bool)
			mediaCodecs := make([]string, 0)
			if checkMediaCodecs := capabilitiesMap["media_codecs"].([]interface{}); len(checkMediaCodecs) > 0 {
				for _, codec := range checkMediaCodecs {
					mediaCodecs = append(mediaCodecs, fmt.Sprintf("%v", codec))
				}
			}
			cdm := capabilitiesMap["cdm"].(bool)

			sdkPhoneCapabilities = platformclientv2.Phonecapabilities{
				Provisions:          &provisions,
				Registers:           &registers,
				DualRegisters:       &dualRegisters,
				HardwareIdType:      &hardwareIdType,
				AllowReboot:         &allowReboot,
				NoRebalance:         &noRebalance,
				NoCloudProvisioning: &noCloudProvisioning,
				MediaCodecs:         &mediaCodecs,
				Cdm:                 &cdm,
			}
		}
		return &sdkPhoneCapabilities
	}
	return nil
}
