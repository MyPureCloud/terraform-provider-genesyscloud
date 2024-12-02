package routing_skill_group

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"terraform-provider-genesyscloud/genesyscloud/util/constants"
	"terraform-provider-genesyscloud/genesyscloud/util/resourcedata"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"terraform-provider-genesyscloud/genesyscloud/consistency_checker"

	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	lists "terraform-provider-genesyscloud/genesyscloud/util/lists"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v146/platformclientv2"
)

func getAllRoutingSkillGroups(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	proxy := getRoutingSkillGroupsProxy(clientConfig)
	resources := make(resourceExporter.ResourceIDMetaMap)

	allSkillGroups, resp, err := proxy.getAllRoutingSkillGroups(ctx, "")
	if err != nil {
		return nil, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to get routing skill groups: %v", err), resp)
	}

	for _, skillGroup := range *allSkillGroups {
		resources[*skillGroup.Id] = &resourceExporter.ResourceMeta{BlockLabel: *skillGroup.Name}
	}

	return resources, nil
}

func createSkillGroups(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getRoutingSkillGroupsProxy(sdkConfig)
	name := d.Get("name").(string)
	description := d.Get("description").(string)

	createRequestBody := platformclientv2.Skillgroupwithmemberdivisions{
		Name:        &name,
		Description: &description,
	}

	if divisionID := d.Get("division_id").(string); divisionID != "" {
		createRequestBody.Division = &platformclientv2.Writabledivision{Id: &divisionID}
	}

	if skillConditions := d.Get("skill_conditions").(string); skillConditions != "" {
		if err := json.Unmarshal([]byte(skillConditions), &createRequestBody.SkillConditions); err != nil {
			return util.BuildDiagnosticError(ResourceType, fmt.Sprintf("Failed to unmarshal the JSON skill conditions while creating the skills group %v", &createRequestBody.Name), err)
		}
	}

	group, response, err := proxy.createRoutingSkillGroups(ctx, &createRequestBody)
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to create skill groups %v error: %s", &createRequestBody.Name, err), response)
	}

	d.SetId(*group.Id)
	log.Printf("Created skill group %v %v", &createRequestBody.Name, group.Id)

	// Update member division IDs
	if diagErr := assignMemberDivisionIds(ctx, d, meta, true); diagErr != nil {
		return diagErr
	}

	return readSkillGroups(ctx, d, meta)
}

func readSkillGroups(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getRoutingSkillGroupsProxy(sdkConfig)
	cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceRoutingSkillGroup(), constants.ConsistencyChecks(), ResourceType)

	log.Printf("Reading skills group %s", d.Id())

	return util.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		skillGroup, resp, err := proxy.getRoutingSkillGroupsById(ctx, d.Id())
		if err != nil {
			if util.IsStatus404(resp) {
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read skill groups %s | error: %s", d.Id(), err), resp))
			}
			return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read skill groups %s | error: %s", d.Id(), err), resp))
		}

		resourcedata.SetNillableValue(d, "name", skillGroup.Name)
		resourcedata.SetNillableValue(d, "description", skillGroup.Description)
		resourcedata.SetNillableReferenceWritableDivision(d, "division_id", skillGroup.Division)

		skillConditionsBytes, _ := json.Marshal(skillGroup.SkillConditions)
		skillConditions := string(skillConditionsBytes)
		if skillConditions != "" {
			_ = d.Set("skill_conditions", skillConditions)
		} else {
			_ = d.Set("skill_conditions", nil)
		}

		// Set member_division_ids avoiding plan not empty error
		memberDivIds, diagErr := readSkillGroupMemberDivisions(ctx, d, meta)
		if diagErr != nil {
			return retry.NonRetryableError(fmt.Errorf("%v", diagErr))
		}

		var schemaMemberDivisionIds []string
		if divIds, ok := d.Get("member_division_ids").([]interface{}); ok {
			schemaMemberDivisionIds = lists.InterfaceListToStrings(divIds)
		}

		memberDivisionIds := organizeMemberDivisionIdsForRead(schemaMemberDivisionIds, memberDivIds, *skillGroup.Division.Id)
		_ = d.Set("member_division_ids", memberDivisionIds)

		log.Printf("Read skill groups name  %s %s", d.Id(), *skillGroup.Name)
		return cc.CheckState(d)
	})
}

func updateSkillGroups(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getRoutingSkillGroupsProxy(sdkConfig)
	name := d.Get("name").(string)
	description := d.Get("description").(string)

	updateRequestBody := platformclientv2.Skillgroup{
		Name:        &name,
		Description: &description,
	}

	if divisionID := d.Get("division_id").(string); divisionID != "" {
		updateRequestBody.Division = &platformclientv2.Writabledivision{Id: &divisionID}
	}

	if SkillConditions := d.Get("skill_conditions").(string); SkillConditions != "" {
		if err := json.Unmarshal([]byte(SkillConditions), &updateRequestBody.SkillConditions); err != nil {
			return util.BuildDiagnosticError(ResourceType, fmt.Sprintf("Failed to unmarshal the JSON skill conditions while updating the skills group %v", &updateRequestBody.Name), err)
		}
	}

	group, resp, err := proxy.updateRoutingSkillGroups(ctx, d.Id(), &updateRequestBody)
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to update skill groups %v error: %s", &updateRequestBody.Name, err), resp)
	}

	log.Printf("Updated skill group %v", &group.Name)

	// Update member division IDs
	if diagErr := assignMemberDivisionIds(ctx, d, meta, false); diagErr != nil {
		return diagErr
	}

	return readSkillGroups(ctx, d, meta)
}

func deleteSkillGroups(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getRoutingSkillGroupsProxy(sdkConfig)

	log.Printf("Deleting skill group %s", d.Id())
	resp, err := proxy.deleteRoutingSkillGroups(ctx, d.Id())
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to delete skill group %s error: %s", d.Id(), err), resp)
	}

	return util.WithRetries(ctx, 30*time.Second, func() *retry.RetryError {
		_, resp, err := proxy.getRoutingSkillGroupsById(ctx, d.Id())
		if err != nil {
			if util.IsStatus404(resp) {
				log.Printf("Deleted skills group %s", d.Id())
				return nil
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Error deleting skill group %s | error: %s", d.Id(), err), resp))
		}
		return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Skill group %s still exists", d.Id()), resp))
	})
}

func createRoutingSkillGroupsMemberDivisions(ctx context.Context, d *schema.ResourceData, meta interface{}, skillGroupDivisionIds []string, create bool) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getRoutingSkillGroupsProxy(sdkConfig)
	memberDivIds := d.Get("member_division_ids").([]interface{})
	var reqBody platformclientv2.Skillgroupmemberdivisions

	if memberDivIds == nil {
		return readSkillGroups(ctx, d, meta)
	}
	schemaDivisionIds := lists.InterfaceListToStrings(memberDivIds)

	toAdd, toRemove, diagErr := createListsForSkillgroupsMembersDivisions(schemaDivisionIds, skillGroupDivisionIds, create, meta)
	if diagErr != nil {
		return diagErr
	}

	toRemove, diagErr = removeSkillGroupDivisionID(d, toRemove)
	if diagErr != nil {
		return diagErr
	}

	if len(toAdd) < 1 && len(toRemove) < 1 {
		return readSkillGroups(ctx, d, meta)
	}

	log.Printf("Updating skill group %s member divisions", d.Id())

	if len(toRemove) > 0 {
		reqBody.RemoveDivisionIds = &toRemove
	}
	if len(toAdd) > 0 {
		reqBody.AddDivisionIds = &toAdd
	}

	resp, err := proxy.createRoutingSkillGroupsMemberDivision(ctx, d.Id(), reqBody)
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to update skill group %s member divisions error: %s", d.Id(), err), resp)
	}

	log.Printf("Updated skill group %s member divisions", d.Id())
	return nil
}

func readSkillGroupMemberDivisions(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]string, diag.Diagnostics) {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getRoutingSkillGroupsProxy(sdkConfig)

	log.Printf("Reading skill group %s member divisions", d.Get("name").(string))

	memberDivisions, resp, err := proxy.getRoutingSkillGroupsMemberDivison(ctx, d.Id())
	if err != nil {
		return nil, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to get member divisions for skill group %s error: %s", d.Id(), err), resp)
	}

	skillGroupMemberDivisionIds := make([]string, 0)
	for _, division := range *memberDivisions.Entities {
		skillGroupMemberDivisionIds = append(skillGroupMemberDivisionIds, *division.Id)
	}

	log.Printf("Read skill group %s member divisions", d.Get("name").(string))

	return skillGroupMemberDivisionIds, nil
}

func GenerateRoutingSkillGroupResourceBasic(
	resourceLabel string,
	name string,
	description string) string {
	return fmt.Sprintf(`resource "%s" "%s" {
		name = "%s"
		description="%s"
	}
	`, ResourceType, resourceLabel, name, description)
}

// Todo: remove once auth divisions is refactored into its own package

func getAllAuthDivisionIds(meta interface{}) ([]string, diag.Diagnostics) {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	allIds := make([]string, 0)

	divisionResourcesMap, err := getAllAuthDivisions(nil, sdkConfig)
	if err != nil {
		return nil, err
	}

	for key, _ := range divisionResourcesMap {
		allIds = append(allIds, key)
	}

	return allIds, nil
}

func getAllAuthDivisions(_ context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	resources := make(resourceExporter.ResourceIDMetaMap)
	authAPI := platformclientv2.NewAuthorizationApiWithConfig(clientConfig)

	for pageNum := 1; ; pageNum++ {
		const pageSize = 100
		divisions, resp, getErr := authAPI.GetAuthorizationDivisions(pageSize, pageNum, "", nil, "", "", false, nil, "")
		if getErr != nil {
			return nil, util.BuildAPIDiagnosticError("genesyscloud_auth_division", fmt.Sprintf("Failed to get page of divisions error: %s", getErr), resp)
		}

		if divisions.Entities == nil || len(*divisions.Entities) == 0 {
			break
		}

		for _, division := range *divisions.Entities {
			resources[*division.Id] = &resourceExporter.ResourceMeta{BlockLabel: *division.Name}
		}
	}

	return resources, nil
}
