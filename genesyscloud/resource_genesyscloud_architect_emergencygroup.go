package genesyscloud

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"terraform-provider-genesyscloud/genesyscloud/consistency_checker"

	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v115/platformclientv2"
)

func getAllEmergencyGroups(_ context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	resources := make(resourceExporter.ResourceIDMetaMap)
	architectAPI := platformclientv2.NewArchitectApiWithConfig(clientConfig)

	for pageNum := 1; ; pageNum++ {
		const pageSize = 100
		emergencyGroupConfigs, _, getErr := architectAPI.GetArchitectEmergencygroups(pageNum, pageSize, "", "", "")
		if getErr != nil {
			return nil, diag.Errorf("Failed to get page of emergency group configs: %v", getErr)
		}

		if emergencyGroupConfigs.Entities == nil || len(*emergencyGroupConfigs.Entities) == 0 {
			break
		}

		for _, emergencyGroupConfig := range *emergencyGroupConfigs.Entities {
			if emergencyGroupConfig.State != nil && *emergencyGroupConfig.State != "deleted" {
				resources[*emergencyGroupConfig.Id] = &resourceExporter.ResourceMeta{Name: *emergencyGroupConfig.Name}
			}
		}
	}

	return resources, nil
}

func ArchitectEmergencyGroupExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: GetAllWithPooledClient(getAllEmergencyGroups),
		RefAttrs: map[string]*resourceExporter.RefAttrSettings{
			"division_id":                            {RefType: "genesyscloud_auth_division"},
			"emergency_call_flows.emergency_flow_id": {RefType: "genesyscloud_flow"},
			"emergency_call_flows.ivr_ids":           {RefType: "genesyscloud_architect_ivr"},
		},
	}
}

func ResourceArchitectEmergencyGroup() *schema.Resource {
	return &schema.Resource{
		Description: "Genesys Cloud Architect Emergency Group",

		CreateContext: CreateWithPooledClient(createEmergencyGroup),
		ReadContext:   ReadWithPooledClient(readEmergencyGroup),
		UpdateContext: UpdateWithPooledClient(updateEmergencyGroup),
		DeleteContext: DeleteWithPooledClient(deleteEmergencyGroup),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "Name of the emergency group. Note:  If the name is changed, the emergency group is dropped and recreated with a new ID. This can cause an Architect flow to be invalid if it references the old emergency group",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"division_id": {
				Description: "The division to which this emergency group will belong. If not set, the home division will be used.",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
			},
			"description": {
				Description: "Description of the emergency group.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"enabled": {
				Description: "The state of the emergency group. Defaults to false/inactive.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
			},
			"emergency_call_flows": {
				Description: "The emergency call flows for this emergency group.",
				Type:        schema.TypeList,
				Optional:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"emergency_flow_id": {
							Description: "The ID of the connected call flow.",
							Type:        schema.TypeString,
							Required:    true,
						},
						"ivr_ids": {
							Description: "The IDs of the connected IVRs.",
							Type:        schema.TypeSet,
							Required:    true,
							Elem:        &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},
		},
	}
}

func createEmergencyGroup(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)
	description := d.Get("description").(string)
	divisionId := d.Get("division_id").(string)
	enabled := d.Get("enabled").(bool)

	sdkConfig := meta.(*ProviderMeta).ClientConfig
	architectAPI := platformclientv2.NewArchitectApiWithConfig(sdkConfig)

	emergencyGroup := platformclientv2.Emergencygroup{
		Name:               &name,
		Enabled:            &enabled,
		EmergencyCallFlows: buildSdkEmergencyGroupCallFlows(d),
	}

	// Optional attributes
	if description != "" {
		emergencyGroup.Description = &description
	}

	if divisionId != "" {
		emergencyGroup.Division = &platformclientv2.Writabledivision{Id: &divisionId}
	}

	log.Printf("Creating emergency group %s", name)
	eGroup, _, err := architectAPI.PostArchitectEmergencygroups(emergencyGroup)
	if err != nil {
		return diag.Errorf("Failed to create emergency group %s: %s", name, err)
	}

	d.SetId(*eGroup.Id)

	log.Printf("Created emergency group %s %s", name, *eGroup.Id)

	return readEmergencyGroup(ctx, d, meta)
}

func readEmergencyGroup(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*ProviderMeta).ClientConfig
	architectApi := platformclientv2.NewArchitectApiWithConfig(sdkConfig)

	log.Printf("Reading emergency group %s", d.Id())
	return WithRetriesForRead(ctx, d, func() *retry.RetryError {
		emergencyGroup, resp, getErr := architectApi.GetArchitectEmergencygroup(d.Id())
		if getErr != nil {
			if IsStatus404(resp) {
				return retry.RetryableError(fmt.Errorf("Failed to read emergency group %s: %s", d.Id(), getErr))
			}
			return retry.NonRetryableError(fmt.Errorf("Failed to read emergency group %s: %s", d.Id(), getErr))
		}

		cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceArchitectEmergencyGroup())

		if emergencyGroup.State != nil && *emergencyGroup.State == "deleted" {
			d.SetId("")
			return nil
		}

		d.Set("name", *emergencyGroup.Name)
		d.Set("division_id", *emergencyGroup.Division.Id)

		if emergencyGroup.Description != nil {
			d.Set("description", *emergencyGroup.Description)
		} else {
			d.Set("description", nil)
		}

		if emergencyGroup.Enabled != nil {
			d.Set("enabled", *emergencyGroup.Enabled)
		} else {
			d.Set("enabled", nil)
		}

		if emergencyGroup.EmergencyCallFlows != nil && len(*emergencyGroup.EmergencyCallFlows) > 0 {
			d.Set("emergency_call_flows", flattenEmergencyCallFlows(*emergencyGroup.EmergencyCallFlows))
		} else {
			d.Set("emergency_call_flows", nil)
		}

		log.Printf("Read emergency group %s %s", d.Id(), *emergencyGroup.Name)
		return cc.CheckState()
	})
}

func updateEmergencyGroup(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)
	description := d.Get("description").(string)
	divisionId := d.Get("division_id").(string)
	enabled := d.Get("enabled").(bool)

	sdkConfig := meta.(*ProviderMeta).ClientConfig
	architectAPI := platformclientv2.NewArchitectApiWithConfig(sdkConfig)

	diagErr := RetryWhen(IsVersionMismatch, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
		// Get current emergency group version
		emergencyGroup, resp, getErr := architectAPI.GetArchitectEmergencygroup(d.Id())
		if getErr != nil {
			return resp, diag.Errorf("Failed to read emergency group %s: %s", d.Id(), getErr)
		}

		log.Printf("Updating emergency group %s", name)
		_, resp, putErr := architectAPI.PutArchitectEmergencygroup(d.Id(), platformclientv2.Emergencygroup{
			Name:               &name,
			Division:           &platformclientv2.Writabledivision{Id: &divisionId},
			Description:        &description,
			Version:            emergencyGroup.Version,
			State:              emergencyGroup.State,
			Enabled:            &enabled,
			EmergencyCallFlows: buildSdkEmergencyGroupCallFlows(d),
		})
		if putErr != nil {
			return resp, diag.Errorf("Failed to put emergency group %s: %s", d.Id(), putErr)
		}
		return resp, nil
	})

	if diagErr != nil {
		return diagErr
	}

	log.Printf("Finished updating emergency group %s", name)
	return readEmergencyGroup(ctx, d, meta)
}

func deleteEmergencyGroup(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*ProviderMeta).ClientConfig
	architectApi := platformclientv2.NewArchitectApiWithConfig(sdkConfig)

	log.Printf("Deleting emergency group %s", d.Id())
	_, err := architectApi.DeleteArchitectEmergencygroup(d.Id())
	if err != nil {
		return diag.Errorf("Failed to delete emergency group %s: %s", d.Id(), err)
	}
	return WithRetries(ctx, 30*time.Second, func() *retry.RetryError {
		emergencyGroup, resp, err := architectApi.GetArchitectEmergencygroup(d.Id())
		if err != nil {
			if IsStatus404(resp) {
				// group deleted
				log.Printf("Deleted emergency group %s", d.Id())
				return nil
			}
			return retry.NonRetryableError(fmt.Errorf("error deleting emergency group %s: %s", d.Id(), err))
		}

		if emergencyGroup.State != nil && *emergencyGroup.State == "deleted" {
			// group deleted
			log.Printf("Deleted emergency group %s", d.Id())
			return nil
		}

		return retry.RetryableError(fmt.Errorf("emergency group %s still exists", d.Id()))
	})
}

func buildSdkEmergencyGroupCallFlows(d *schema.ResourceData) *[]platformclientv2.Emergencycallflow {
	var allCallFlows []platformclientv2.Emergencycallflow
	if callFlows, ok := d.GetOk("emergency_call_flows"); ok {
		for _, callFlow := range callFlows.([]interface{}) {
			callFlowSettings := callFlow.(map[string]interface{})
			var currentCallFlow platformclientv2.Emergencycallflow

			if flowID, ok := callFlowSettings["emergency_flow_id"].(string); ok {
				currentCallFlow.EmergencyFlow = &platformclientv2.Domainentityref{Id: &flowID}
			}

			if ivrIds, ok := callFlowSettings["ivr_ids"]; ok {
				ids := ivrIds.(*schema.Set).List()
				if len(ids) > 0 {
					sdkIvrIds := make([]platformclientv2.Domainentityref, len(ids))
					for i, id := range ids {
						ivrID := id.(string)
						sdkIvrIds[i] = platformclientv2.Domainentityref{Id: &ivrID}
					}
					currentCallFlow.Ivrs = &sdkIvrIds
				}
			}
			allCallFlows = append(allCallFlows, currentCallFlow)
		}
	}
	return &allCallFlows
}

func flattenEmergencyCallFlows(emergencyCallFlows []platformclientv2.Emergencycallflow) []interface{} {
	callFlows := make([]interface{}, len(emergencyCallFlows))
	for i, callFlow := range emergencyCallFlows {
		callFlowSettings := make(map[string]interface{})
		if callFlow.EmergencyFlow != nil {
			callFlowSettings["emergency_flow_id"] = *callFlow.EmergencyFlow.Id
		}
		if callFlow.Ivrs != nil && len(*callFlow.Ivrs) > 0 {
			ivrIds := make([]interface{}, len(*callFlow.Ivrs))
			for k, id := range *callFlow.Ivrs {
				ivrIds[k] = *id.Id
			}
			callFlowSettings["ivr_ids"] = schema.NewSet(schema.HashString, ivrIds)
		}
		callFlows[i] = callFlowSettings
	}
	return callFlows
}
