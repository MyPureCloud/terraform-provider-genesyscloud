package genesyscloud

import (
	"context"
	"fmt"
	"log"
	"strings"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"terraform-provider-genesyscloud/genesyscloud/util/constants"
	"terraform-provider-genesyscloud/genesyscloud/validators"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"terraform-provider-genesyscloud/genesyscloud/consistency_checker"

	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	lists "terraform-provider-genesyscloud/genesyscloud/util/lists"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"
	"github.com/nyaruka/phonenumbers"
)

func getAllLocations(_ context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	resources := make(resourceExporter.ResourceIDMetaMap)
	locationsAPI := platformclientv2.NewLocationsApiWithConfig(clientConfig)

	for pageNum := 1; ; pageNum++ {
		const pageSize = 100
		locations, resp, getErr := locationsAPI.GetLocations(pageSize, pageNum, nil, "")
		if getErr != nil {
			return nil, util.BuildAPIDiagnosticError("genesyscloud_location", fmt.Sprintf("Failed to get page of locations error: %s", getErr), resp)
		}

		if locations.Entities == nil || len(*locations.Entities) == 0 {
			break
		}

		for _, location := range *locations.Entities {
			resources[*location.Id] = &resourceExporter.ResourceMeta{Name: *location.Name}
		}
	}

	return resources, nil
}

func LocationExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: provider.GetAllWithPooledClient(getAllLocations),
		RefAttrs: map[string]*resourceExporter.RefAttrSettings{
			"path": {RefType: "genesyscloud_location"},
		},
		CustomValidateExports: map[string][]string{
			"E164": {"emergency_number.number"},
		},
	}
}

func ResourceLocation() *schema.Resource {
	return &schema.Resource{
		Description: "Genesys Cloud Location",

		CreateContext: provider.CreateWithPooledClient(createLocation),
		ReadContext:   provider.ReadWithPooledClient(readLocation),
		UpdateContext: provider.UpdateWithPooledClient(updateLocation),
		DeleteContext: provider.DeleteWithPooledClient(deleteLocation),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "Location name.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"path": {
				Description: "A list of ancestor location IDs. This can be used to create sublocations.",
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"notes": {
				Description: "Notes for this location.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"emergency_number": {
				Description: "Emergency phone number for this location.",
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"number": {
							Description:      "Emergency phone number.  Must be in an E.164 number format.",
							Type:             schema.TypeString,
							Required:         true,
							ValidateDiagFunc: validators.ValidatePhoneNumber,
							DiffSuppressFunc: comparePhoneNumbers,
						},
						"type": {
							Description:  "Type of emergency number (default | elin).",
							Type:         schema.TypeString,
							Optional:     true,
							Default:      "default",
							ValidateFunc: validation.StringInSlice([]string{"default", "elin"}, false),
						},
					},
				},
			},
			"address": {
				Description: "Address for this location. This cannot be changed while an emergency number is assigned.",
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"city": {
							Description: "Location city.",
							Type:        schema.TypeString,
							Required:    true,
						},
						"country": {
							Description: "Country abbreviation.",
							Type:        schema.TypeString,
							Required:    true,
						},
						"state": {
							Description: "Location state. Required for countries with states.",
							Type:        schema.TypeString,
							Optional:    true,
						},
						"street1": {
							Description: "Street address 1.",
							Type:        schema.TypeString,
							Required:    true,
						},
						"street2": {
							Description: "Street address 2.",
							Type:        schema.TypeString,
							Optional:    true,
						},
						"zip_code": {
							Description: "Location zip code.",
							Type:        schema.TypeString,
							Required:    true,
						},
					},
				},
			},
		},
	}
}

func createLocation(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)
	notes := d.Get("notes").(string)

	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	locationsAPI := platformclientv2.NewLocationsApiWithConfig(sdkConfig)

	create := platformclientv2.Locationcreatedefinition{
		Name:            &name,
		Path:            buildSdkLocationPath(d),
		EmergencyNumber: buildSdkLocationEmergencyNumber(d),
		Address:         buildSdkLocationAddress(d),
	}
	if notes != "" {
		// API does not let allow empty string for notes on create
		create.Notes = &notes
	}

	log.Printf("Creating location %s", name)
	location, resp, err := locationsAPI.PostLocations(create)
	if err != nil {
		return util.BuildAPIDiagnosticError("genesyscloud_location", fmt.Sprintf("Failed to create location %s error: %s", name, err), resp)
	}

	d.SetId(*location.Id)

	log.Printf("Created location %s %s", name, *location.Id)
	return readLocation(ctx, d, meta)
}

func readLocation(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	locationsAPI := platformclientv2.NewLocationsApiWithConfig(sdkConfig)
	cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceLocation(), constants.DefaultConsistencyChecks, "genesyscloud_location")

	log.Printf("Reading location %s", d.Id())
	return util.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		location, resp, getErr := locationsAPI.GetLocation(d.Id(), nil)
		if getErr != nil {
			if util.IsStatus404(resp) {
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError("genesyscloud_location", fmt.Sprintf("Failed to read location %s | error: %s", d.Id(), getErr), resp))
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError("genesyscloud_location", fmt.Sprintf("Failed to read location %s | error: %s", d.Id(), getErr), resp))
		}

		if location.State != nil && *location.State == "deleted" {
			d.SetId("")
			return nil
		}

		d.Set("name", *location.Name)

		if location.Notes != nil {
			d.Set("notes", *location.Notes)
		} else {
			d.Set("notes", nil)
		}

		if location.Path != nil {
			d.Set("path", *location.Path)
		} else {
			d.Set("path", nil)
		}

		d.Set("emergency_number", flattenLocationEmergencyNumber(location.EmergencyNumber))
		d.Set("address", flattenLocationAddress(location.Address))

		log.Printf("Read location %s %s", d.Id(), *location.Name)
		return cc.CheckState(d)
	})
}

func updateLocation(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)
	notes := d.Get("notes").(string)

	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	locationsAPI := platformclientv2.NewLocationsApiWithConfig(sdkConfig)

	log.Printf("Updating location %s", name)
	diagErr := util.RetryWhen(util.IsVersionMismatch, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
		// Get current location version
		location, resp, getErr := locationsAPI.GetLocation(d.Id(), nil)
		if getErr != nil {
			return resp, util.BuildAPIDiagnosticError("genesyscloud_location", fmt.Sprintf("Failed to read location %s error: %s", name, getErr), resp)
		}

		update := platformclientv2.Locationupdatedefinition{
			Version:         location.Version,
			Name:            &name,
			Path:            buildSdkLocationPath(d),
			EmergencyNumber: buildSdkLocationEmergencyNumber(d),
		}
		if d.HasChange("address") {
			// Even if address is the same, the API does not allow it in the patch request if a number is assigned
			update.Address = buildSdkLocationAddress(d)
		}
		if notes != "" {
			update.Notes = &notes
		} else {
			// nil will result in no change occurring, and an empty string is invalid for this field
			filler := " "
			update.Notes = &filler
			err := d.Set("notes", filler)
			if err != nil {
				return nil, util.BuildDiagnosticError("genesyscloud_location", fmt.Sprintf("error setting the value of 'notes' attribute"), err)
			}
		}

		log.Printf("Updating location %s", name)
		_, resp, putErr := locationsAPI.PatchLocation(d.Id(), update)
		if putErr != nil {
			return resp, util.BuildAPIDiagnosticError("genesyscloud_location", fmt.Sprintf("Failed to update location %s error: %s", name, putErr), resp)
		}
		return resp, nil
	})
	if diagErr != nil {
		return diagErr
	}

	log.Printf("Updated location %s %s", name, d.Id())
	return readLocation(ctx, d, meta)
}

func deleteLocation(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)

	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	locationsAPI := platformclientv2.NewLocationsApiWithConfig(sdkConfig)

	log.Printf("Deleting location %s", name)
	diagErr := util.RetryWhen(util.IsVersionMismatch, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
		// Directory occasionally returns version errors on deletes if an object was updated at the same time.
		resp, err := locationsAPI.DeleteLocation(d.Id())
		if err != nil {
			return resp, util.BuildAPIDiagnosticError("genesyscloud_location", fmt.Sprintf("Failed to delete location %s error: %s", name, err), resp)
		}
		return nil, nil
	})
	if diagErr != nil {
		return diagErr
	}

	return util.WithRetries(ctx, 30*time.Second, func() *retry.RetryError {
		location, resp, err := locationsAPI.GetLocation(d.Id(), nil)
		if err != nil {
			if util.IsStatus404(resp) {
				// Location deleted
				log.Printf("Deleted location %s", d.Id())
				return nil
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError("genesyscloud_location", fmt.Sprintf("Error deleting location %s | error: %s", d.Id(), err), resp))
		}

		if location.State != nil && *location.State == "deleted" {
			// Location deleted
			log.Printf("Deleted location %s", d.Id())
			return nil
		}

		return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError("genesyscloud_location", fmt.Sprintf("Location %s still exists", d.Id()), resp))
	})
}

func buildSdkLocationPath(d *schema.ResourceData) *[]string {
	path := []string{}
	if pathConfig, ok := d.GetOk("path"); ok {
		path = lists.InterfaceListToStrings(pathConfig.([]interface{}))
	}
	return &path
}

func buildSdkLocationEmergencyNumber(d *schema.ResourceData) *platformclientv2.Locationemergencynumber {
	if numberConfig := d.Get("emergency_number"); numberConfig != nil {
		if numberList := numberConfig.([]interface{}); len(numberList) > 0 {
			settingsMap := numberList[0].(map[string]interface{})

			number := settingsMap["number"].(string)
			typeStr := settingsMap["type"].(string)
			return &platformclientv2.Locationemergencynumber{
				Number:  &number,
				VarType: &typeStr,
			}
		}
	}
	return &platformclientv2.Locationemergencynumber{}
}

func buildSdkLocationAddress(d *schema.ResourceData) *platformclientv2.Locationaddress {
	if addressConfig := d.Get("address"); addressConfig != nil {
		if addrList := addressConfig.([]interface{}); len(addrList) > 0 {
			addrMap := addrList[0].(map[string]interface{})

			city := addrMap["city"].(string)
			country := addrMap["country"].(string)
			zip := addrMap["zip_code"].(string)
			street1 := addrMap["street1"].(string)
			address := platformclientv2.Locationaddress{
				City:    &city,
				Country: &country,
				Zipcode: &zip,
				Street1: &street1,
			}
			// Optional values
			if state, ok := addrMap["state"]; ok {
				stateStr := state.(string)
				address.State = &stateStr
			}
			if street2, ok := addrMap["street2"]; ok {
				street2Str := street2.(string)
				address.Street2 = &street2Str
			}
			return &address
		}
	}
	return &platformclientv2.Locationaddress{}
}

func flattenLocationEmergencyNumber(numberConfig *platformclientv2.Locationemergencynumber) []interface{} {
	if numberConfig == nil {
		return nil
	}
	numberSettings := make(map[string]interface{})
	if numberConfig.Number != nil {
		numberSettings["number"], _ = util.FormatAsE164Number(*numberConfig.Number)
	}
	if numberConfig.VarType != nil {
		numberSettings["type"] = *numberConfig.VarType
	}
	return []interface{}{numberSettings}
}

func flattenLocationAddress(addrConfig *platformclientv2.Locationaddress) []interface{} {
	if addrConfig == nil {
		return nil
	}
	addrSettings := make(map[string]interface{})
	if addrConfig.City != nil {
		addrSettings["city"] = *addrConfig.City
	}
	if addrConfig.Country != nil {
		addrSettings["country"] = *addrConfig.Country
	}
	if addrConfig.State != nil {
		addrSettings["state"] = *addrConfig.State
	}
	if addrConfig.Street1 != nil {
		addrSettings["street1"] = *addrConfig.Street1
	}
	if addrConfig.Street2 != nil {
		addrSettings["street2"] = *addrConfig.Street2
	}
	if addrConfig.Zipcode != nil {
		addrSettings["zip_code"] = *addrConfig.Zipcode
	}
	return []interface{}{addrSettings}
}

func comparePhoneNumbers(_, old, new string, _ *schema.ResourceData) bool {
	oldNum, err := phonenumbers.Parse(old, "US")
	if err != nil {
		return old == new
	}

	newNum, err := phonenumbers.Parse(new, "US")
	if err != nil {
		return old == new
	}
	return phonenumbers.IsNumberMatchWithNumbers(oldNum, newNum) == phonenumbers.EXACT_MATCH
}

func GenerateLocationResourceBasic(
	resourceID,
	name string,
	nestedBlocks ...string) string {
	return GenerateLocationResource(resourceID, name, "", []string{})
}

func GenerateLocationResource(
	resourceID,
	name,
	notes string,
	paths []string,
	nestedBlocks ...string) string {
	return fmt.Sprintf(`resource "genesyscloud_location" "%s" {
		name = "%s"
        notes = "%s"
        path = [%s]
        %s
	}
	`, resourceID, name, notes, strings.Join(paths, ","), strings.Join(nestedBlocks, "\n"))
}

func GenerateLocationEmergencyNum(number, typeStr string) string {
	return fmt.Sprintf(`emergency_number {
		number = "%s"
        type = %s
	}
	`, number, typeStr)
}

func GenerateLocationAddress(street1, city, state, country, zip string) string {
	return fmt.Sprintf(`address {
		street1  = "%s"
		city     = "%s"
		state    = "%s"
		country  = "%s"
		zip_code = "%s"
	}
	`, street1, city, state, country, zip)
}
