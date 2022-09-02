package genesyscloud

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/mypurecloud/platform-client-sdk-go/v75/platformclientv2"
	"log"
	"time"
)

var (
	outboundcallabletimesetcallabletimeResource = &schema.Resource{
		Schema: map[string]*schema.Schema{
			`time_slots`: {
				Description: `The time intervals for which it is acceptable to place outbound calls.`,
				Required:    true,
				Type:        schema.TypeSet,
				Elem:        outboundcallabletimesetcampaigntimeslotResource,
			},
			`time_zone_id`: {
				Description: `The time zone for the time slots; for example, Africa/Abidjan`,
				Required:    true,
				Type:        schema.TypeString,
			},
		},
	}
	outboundcallabletimesetcampaigntimeslotResource = &schema.Resource{
		Schema: map[string]*schema.Schema{
			`start_time`: {
				Description:      `The start time of the interval as an ISO-8601 string, i.e. HH:mm:ss`,
				Required:         true,
				ValidateDiagFunc: validateTime,
				Type:             schema.TypeString,
			},
			`stop_time`: {
				Description:      `The end time of the interval as an ISO-8601 string, i.e. HH:mm:ss`,
				Required:         true,
				ValidateDiagFunc: validateTime,
				Type:             schema.TypeString,
			},
			`day`: {
				Description:  `The day of the interval. Valid values: [1-7], representing Monday through Sunday`,
				Required:     true,
				ValidateFunc: validation.IntInSlice([]int{1, 2, 3, 4, 5, 6, 7}),
				Type:         schema.TypeInt,
			},
		},
	}
)

func resourceOutboundCallabletimeset() *schema.Resource {
	return &schema.Resource{
		Description: `Genesys Cloud outbound callabletimeset`,

		CreateContext: createWithPooledClient(createOutboundCallabletimeset),
		ReadContext:   readWithPooledClient(readOutboundCallabletimeset),
		UpdateContext: updateWithPooledClient(updateOutboundCallabletimeset),
		DeleteContext: deleteWithPooledClient(deleteOutboundCallabletimeset),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			`name`: {
				Description: `The name of the CallableTimeSet.`,
				Required:    true,
				Type:        schema.TypeString,
			},
			`callable_times`: {
				Description: `The list of CallableTimes for which it is acceptable to place outbound calls.`,
				Required:    true,
				Type:        schema.TypeSet,
				Elem:        outboundcallabletimesetcallabletimeResource,
			},
		},
	}
}

func outboundCallableTimesetExporter() *ResourceExporter {
	return &ResourceExporter{
		GetResourcesFunc: getAllWithPooledClient(getAllOutboundCallableTimesets),
	}
}

func getAllOutboundCallableTimesets(_ context.Context, clientConfig *platformclientv2.Configuration) (ResourceIDMetaMap, diag.Diagnostics) {
	resource := make(ResourceIDMetaMap)
	outboundAPI := platformclientv2.NewOutboundApiWithConfig(clientConfig)

	for pageNum := 1; ; pageNum++ {
		const pageSize = 100
		callableTimesetConfigs, _, getErr := outboundAPI.GetOutboundCallabletimesets(pageSize, pageNum, false, "", "", "", "")
		if getErr != nil {
			return nil, diag.Errorf("Failed to get page of callable timeset configs: %v", getErr)
		}

		if callableTimesetConfigs.Entities == nil || len(*callableTimesetConfigs.Entities) == 0 {
			break
		}

		for _, callableTimesetConfig := range *callableTimesetConfigs.Entities {
			resource[*callableTimesetConfig.Id] = &ResourceMeta{Name: *callableTimesetConfig.Name}
		}

	}
	return resource, nil
}

func createOutboundCallabletimeset(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)

	sdkConfig := meta.(*providerMeta).ClientConfig
	outboundApi := platformclientv2.NewOutboundApiWithConfig(sdkConfig)

	sdkcallabletimeset := platformclientv2.Callabletimeset{
		CallableTimes: buildSdkoutboundcallabletimesetCallabletimeSlice(d.Get("callable_times").(*schema.Set)),
	}

	if name != "" {
		sdkcallabletimeset.Name = &name
	}

	log.Printf("Creating Outbound Callabletimeset %s", name)
	outboundCallabletimeset, _, err := outboundApi.PostOutboundCallabletimesets(sdkcallabletimeset)
	if err != nil {
		return diag.Errorf("Failed to create Outbound Callabletimeset %s: %s", name, err)
	}

	d.SetId(*outboundCallabletimeset.Id)
	fmt.Printf("Before trim: %s\n", outboundCallabletimeset)
	// Remove the milliseconds added to start_time and stop_time by the API
	trimTime(outboundCallabletimeset.CallableTimes)
	if sdkcallabletimeset.CallableTimes != nil {
		d.Set("callable_times", flattenSdkoutboundcallabletimesetCallabletimeSlice(*outboundCallabletimeset.CallableTimes))
	}
	fmt.Printf("After trim: %s\n", outboundCallabletimeset)
	log.Printf("Created Outbound Callabletimeset %s %s", name, *outboundCallabletimeset.Id)
	return readOutboundCallabletimeset(ctx, d, meta)
}

func updateOutboundCallabletimeset(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)

	sdkConfig := meta.(*providerMeta).ClientConfig
	outboundApi := platformclientv2.NewOutboundApiWithConfig(sdkConfig)

	sdkcallabletimeset := platformclientv2.Callabletimeset{
		CallableTimes: buildSdkoutboundcallabletimesetCallabletimeSlice(d.Get("callable_times").(*schema.Set)),
	}

	if name != "" {
		sdkcallabletimeset.Name = &name
	}

	log.Printf("Updating Outbound Callabletimeset %s", name)
	diagErr := retryWhen(isVersionMismatch, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
		// Get current Outbound Callabletimeset version
		outboundCallabletimeset, resp, getErr := outboundApi.GetOutboundCallabletimeset(d.Id())
		if getErr != nil {
			return resp, diag.Errorf("Failed to read Outbound Callabletimeset %s: %s", d.Id(), getErr)
		}
		sdkcallabletimeset.Version = outboundCallabletimeset.Version
		outboundCallabletimeset, _, updateErr := outboundApi.PutOutboundCallabletimeset(d.Id(), sdkcallabletimeset)
		if updateErr != nil {
			return resp, diag.Errorf("Failed to update Outbound Callabletimeset %s: %s", name, updateErr)
		}

		trimTime(outboundCallabletimeset.CallableTimes)
		if sdkcallabletimeset.CallableTimes != nil {
			d.Set("callable_times", flattenSdkoutboundcallabletimesetCallabletimeSlice(*outboundCallabletimeset.CallableTimes))
		}
		return nil, nil
	})
	if diagErr != nil {
		return diagErr
	}

	log.Printf("Updated Outbound Callabletimeset %s", name)
	return readOutboundCallabletimeset(ctx, d, meta)
}

func readOutboundCallabletimeset(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*providerMeta).ClientConfig
	outboundApi := platformclientv2.NewOutboundApiWithConfig(sdkConfig)

	log.Printf("Reading Outbound Callabletimeset %s", d.Id())

	return withRetriesForRead(ctx, d, func() *resource.RetryError {
		sdkcallabletimeset, resp, getErr := outboundApi.GetOutboundCallabletimeset(d.Id())
		if getErr != nil {
			if isStatus404(resp) {
				return resource.RetryableError(fmt.Errorf("Failed to read Outbound Callabletimeset %s: %s", d.Id(), getErr))
			}
			return resource.NonRetryableError(fmt.Errorf("Failed to read Outbound Callabletimeset %s: %s", d.Id(), getErr))
		}

		//cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, resourceOutboundCallabletimeset())

		if sdkcallabletimeset.Name != nil {
			d.Set("name", *sdkcallabletimeset.Name)
		}
		if sdkcallabletimeset.CallableTimes != nil {
			d.Set("callable_times", flattenSdkoutboundcallabletimesetCallabletimeSlice(*sdkcallabletimeset.CallableTimes))
		}

		log.Printf("Read Outbound Callabletimeset %s %s", d.Id(), *sdkcallabletimeset.Name)
		//return cc.CheckState()
		return nil
	})
}

func deleteOutboundCallabletimeset(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*providerMeta).ClientConfig
	outboundApi := platformclientv2.NewOutboundApiWithConfig(sdkConfig)

	diagErr := retryWhen(isStatus400, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
		log.Printf("Deleting Outbound Callabletimeset")
		resp, err := outboundApi.DeleteOutboundCallabletimeset(d.Id())
		if err != nil {
			return resp, diag.Errorf("Failed to delete Outbound Callabletimeset: %s", err)
		}
		return resp, nil
	})
	if diagErr != nil {
		return diagErr
	}

	return withRetries(ctx, 30*time.Second, func() *resource.RetryError {
		_, resp, err := outboundApi.GetOutboundCallabletimeset(d.Id())
		if err != nil {
			if isStatus404(resp) {
				// Outbound Callabletimeset deleted
				log.Printf("Deleted Outbound Callabletimeset %s", d.Id())
				return nil
			}
			return resource.NonRetryableError(fmt.Errorf("Error deleting Outbound Callabletimeset %s: %s", d.Id(), err))
		}

		return resource.RetryableError(fmt.Errorf("Outbound Callabletimeset %s still exists", d.Id()))
	})
}

func trimTime(values *[]platformclientv2.Callabletime) {
	for _, value := range *values {
		for _, timeSlot := range *value.TimeSlots {
			startTime := *timeSlot.StartTime
			timeSlot.StartTime = platformclientv2.String(startTime[:8])
			//fmt.Printf("Time before trim: %s\nTime sfter trim: %s\n\n", startTime, *timeSlot.StartTime)

			stopTime := *timeSlot.StopTime
			timeSlot.StopTime = platformclientv2.String(stopTime[:8])
		}
	}
}

func buildSdkoutboundcallabletimesetCampaigntimeslotSlice(campaigntimeslot *schema.Set) *[]platformclientv2.Campaigntimeslot {
	if campaigntimeslot == nil {
		return nil
	}
	sdkCampaigntimeslotSlice := make([]platformclientv2.Campaigntimeslot, 0)
	campaigntimeslotList := campaigntimeslot.List()
	for _, configcampaigntimeslot := range campaigntimeslotList {
		var sdkCampaigntimeslot platformclientv2.Campaigntimeslot

		campaigntimeslotMap := configcampaigntimeslot.(map[string]interface{})
		if startTime := campaigntimeslotMap["start_time"].(string); startTime != "" {
			sdkCampaigntimeslot.StartTime = &startTime
		}
		if stopTime := campaigntimeslotMap["stop_time"].(string); stopTime != "" {
			sdkCampaigntimeslot.StopTime = &stopTime
		}
		sdkCampaigntimeslot.Day = platformclientv2.Int(campaigntimeslotMap["day"].(int))

		sdkCampaigntimeslotSlice = append(sdkCampaigntimeslotSlice, sdkCampaigntimeslot)
	}
	return &sdkCampaigntimeslotSlice
}

func buildSdkoutboundcallabletimesetCallabletimeSlice(callabletime *schema.Set) *[]platformclientv2.Callabletime {
	if callabletime == nil {
		return nil
	}
	sdkCallabletimeSlice := make([]platformclientv2.Callabletime, 0)
	callabletimeList := callabletime.List()
	for _, configcallabletime := range callabletimeList {
		var sdkCallabletime platformclientv2.Callabletime
		callabletimeMap := configcallabletime.(map[string]interface{})
		if timeSlots := callabletimeMap["time_slots"]; timeSlots != nil {
			sdkCallabletime.TimeSlots = buildSdkoutboundcallabletimesetCampaigntimeslotSlice(timeSlots.(*schema.Set))
		}
		if timeZoneId := callabletimeMap["time_zone_id"].(string); timeZoneId != "" {
			sdkCallabletime.TimeZoneId = &timeZoneId
		}

		sdkCallabletimeSlice = append(sdkCallabletimeSlice, sdkCallabletime)
	}
	return &sdkCallabletimeSlice
}

func flattenSdkoutboundcallabletimesetCampaigntimeslotSlice(campaigntimeslots []platformclientv2.Campaigntimeslot) *schema.Set {
	if len(campaigntimeslots) == 0 {
		return nil
	}

	campaigntimeslotSet := schema.NewSet(schema.HashResource(outboundcallabletimesetcampaigntimeslotResource), []interface{}{})
	for _, campaigntimeslot := range campaigntimeslots {
		campaigntimeslotMap := make(map[string]interface{})

		if campaigntimeslot.StartTime != nil {
			campaigntimeslotMap["start_time"] = *campaigntimeslot.StartTime
		}
		if campaigntimeslot.StopTime != nil {
			campaigntimeslotMap["stop_time"] = *campaigntimeslot.StopTime
		}
		if campaigntimeslot.Day != nil {
			campaigntimeslotMap["day"] = *campaigntimeslot.Day
		}

		campaigntimeslotSet.Add(campaigntimeslotMap)
	}

	return campaigntimeslotSet
}

func flattenSdkoutboundcallabletimesetCallabletimeSlice(callabletimes []platformclientv2.Callabletime) *schema.Set {
	if len(callabletimes) == 0 {
		return nil
	}

	callabletimeSet := schema.NewSet(schema.HashResource(outboundcallabletimesetcallabletimeResource), []interface{}{})
	for _, callabletime := range callabletimes {
		callabletimeMap := make(map[string]interface{})

		if callabletime.TimeSlots != nil {
			callabletimeMap["time_slots"] = flattenSdkoutboundcallabletimesetCampaigntimeslotSlice(*callabletime.TimeSlots)
		}
		if callabletime.TimeZoneId != nil {
			callabletimeMap["time_zone_id"] = *callabletime.TimeZoneId
		}

		callabletimeSet.Add(callabletimeMap)
	}

	return callabletimeSet
}
