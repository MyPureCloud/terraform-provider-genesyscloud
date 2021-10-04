package genesyscloud

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/mypurecloud/platform-client-sdk-go/v55/platformclientv2"
	"github.com/nyaruka/phonenumbers"
)

func getAllLocations(ctx context.Context, clientConfig *platformclientv2.Configuration) (ResourceIDMetaMap, diag.Diagnostics) {
	resources := make(ResourceIDMetaMap)
	locationsAPI := platformclientv2.NewLocationsApiWithConfig(clientConfig)

	for pageNum := 1; ; pageNum++ {
		locations, _, getErr := locationsAPI.GetLocations(100, pageNum, nil, "")
		if getErr != nil {
			return nil, diag.Errorf("Failed to get page of locations: %v", getErr)
		}

		if locations.Entities == nil || len(*locations.Entities) == 0 {
			break
		}

		for _, location := range *locations.Entities {
			resources[*location.Id] = &ResourceMeta{Name: *location.Name}
		}
	}

	return resources, nil
}

func locationExporter() *ResourceExporter {
	return &ResourceExporter{
		GetResourcesFunc: getAllWithPooledClient(getAllLocations),
		RefAttrs: map[string]*RefAttrSettings{
			"path": {RefType: "genesyscloud_location"},
		},
	}
}

func resourceLocation() *schema.Resource {
	return &schema.Resource{
		Description: "Genesys Cloud Location",

		CreateContext: createWithPooledClient(createLocation),
		ReadContext:   readWithPooledClient(readLocation),
		UpdateContext: updateWithPooledClient(updateLocation),
		DeleteContext: deleteWithPooledClient(deleteLocation),
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
							Description:      "Emergency phone number.",
							Type:             schema.TypeString,
							Required:         true,
							ValidateDiagFunc: validatePhoneNumber,
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

	sdkConfig := meta.(*providerMeta).ClientConfig
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
	location, _, err := locationsAPI.PostLocations(create)
	if err != nil {
		return diag.Errorf("Failed to create location %s: %s", name, err)
	}

	d.SetId(*location.Id)

	log.Printf("Created location %s %s", name, *location.Id)
	return readLocation(ctx, d, meta)
}

func readLocation(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*providerMeta).ClientConfig
	locationsAPI := platformclientv2.NewLocationsApiWithConfig(sdkConfig)

	log.Printf("Reading location %s", d.Id())
	location, resp, getErr := locationsAPI.GetLocation(d.Id(), nil)
	if getErr != nil {
		if resp != nil && resp.StatusCode == 404 {
			d.SetId("")
			return nil
		}
		return diag.Errorf("Failed to read location %s: %s", d.Id(), getErr)
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
	return nil
}

func updateLocation(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)
	notes := d.Get("notes").(string)

	sdkConfig := meta.(*providerMeta).ClientConfig
	locationsAPI := platformclientv2.NewLocationsApiWithConfig(sdkConfig)

	log.Printf("Updating location %s", name)
	diagErr := retryWhen(isVersionMismatch, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
		// Get current location version
		location, resp, getErr := locationsAPI.GetLocation(d.Id(), nil)
		if getErr != nil {
			return resp, diag.Errorf("Failed to read location %s: %s", d.Id(), getErr)
		}

		update := platformclientv2.Locationupdatedefinition{
			Version:         location.Version,
			Name:            &name,
			Notes:           &notes,
			Path:            buildSdkLocationPath(d),
			EmergencyNumber: buildSdkLocationEmergencyNumber(d),
		}
		if d.HasChange("address") {
			// Even if address is the same, the API does not allow it in the patch request if a number is assigned
			update.Address = buildSdkLocationAddress(d)
		}

		log.Printf("Updating location %s", name)
		_, resp, putErr := locationsAPI.PatchLocation(d.Id(), update)
		if putErr != nil {
			return resp, diag.Errorf("Failed to update location %s: %s", d.Id(), putErr)
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

	sdkConfig := meta.(*providerMeta).ClientConfig
	locationsAPI := platformclientv2.NewLocationsApiWithConfig(sdkConfig)

	log.Printf("Deleting location %s", name)
	retryWhen(isVersionMismatch, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
		// Directory occasionally returns version errors on deletes if an object was updated at the same time.
		resp, err := locationsAPI.DeleteLocation(d.Id())
		if err != nil {
			return resp, diag.Errorf("Failed to delete location %s: %s", name, err)
		}
		return nil, nil
	})

	return withRetries(ctx, 30*time.Second, func() *resource.RetryError {
		location, resp, err := locationsAPI.GetLocation(d.Id(), nil)
		if err != nil {
			if resp != nil && resp.StatusCode == 404 {
				// Location deleted
				log.Printf("Deleted location %s", d.Id())
				return nil
			}
			return resource.NonRetryableError(fmt.Errorf("Error deleting location %s: %s", d.Id(), err))
		}

		if *location.State == "deleted" {
			// Location deleted
			log.Printf("Deleted location %s", d.Id())
			return nil
		}

		return resource.RetryableError(fmt.Errorf("Location %s still exists", d.Id()))
	})
}

func buildSdkLocationPath(d *schema.ResourceData) *[]string {
	path := []string{}
	if pathConfig, ok := d.GetOk("path"); ok {
		path = interfaceListToStrings(pathConfig.([]interface{}))
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
		numberSettings["number"] = *numberConfig.Number
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

func comparePhoneNumbers(k, old, new string, d *schema.ResourceData) bool {
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
