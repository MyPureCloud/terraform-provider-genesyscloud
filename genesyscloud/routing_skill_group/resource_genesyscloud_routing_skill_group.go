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
	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"
)

func getAllRoutingSkillGroups(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	proxy := getRoutingSkillGroupProxy(clientConfig)
	resources := make(resourceExporter.ResourceIDMetaMap)

	skillGroupWithMemberDivisions, resp, err := proxy.getAllRoutingSkillGroups(ctx, "")
	if err != nil {
		return nil, util.BuildAPIDiagnosticError(resourceName, fmt.Sprintf("Failed to get routing skill groups: %v", err), resp)
	}

	for _, skillGroupWithMemberDivisions := range *skillGroupWithMemberDivisions {
		resources[*skillGroupWithMemberDivisions.Id] = &resourceExporter.ResourceMeta{Name: *skillGroupWithMemberDivisions.Name}
	}

	return resources, nil
}

func createSkillGroups(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig

	skillGroupsRequest := &SkillGroupsRequest{
		Name:        d.Get("name").(string),
		Description: d.Get("description").(string),
	}
	divisionId := d.Get("division_id").(string)
	if divisionId != "" {
		skillGroupsRequest.Division.ID = divisionId
	}

	finalSkillGroupsJson, err := mergeSkillConditionsIntoSkillGroups(d, skillGroupsRequest)
	if err != nil {
		return util.BuildDiagnosticError(resourceName, fmt.Sprintf("Failed to read the before skills groups request before: %s", skillGroupsRequest.Name), err)
	}

	routingAPI := platformclientv2.NewRoutingApiWithConfig(sdkConfig)
	apiClient := &routingAPI.Configuration.APIClient
	path := routingAPI.Configuration.BasePath + "/api/v2/routing/skillgroups"
	headerParams := BuildHeaderParams(routingAPI)

	var skillGroupsPayload map[string]interface{}
	err = json.Unmarshal([]byte(finalSkillGroupsJson), &skillGroupsPayload)
	if err != nil {
		return util.BuildDiagnosticError(resourceName, fmt.Sprintf("Failed to unmarshal the JSON payload while creating/updating the skills group %s", skillGroupsRequest.Name), err)
	}

	response, err := apiClient.CallAPI(path, "POST", skillGroupsPayload, headerParams, nil, nil, "", nil)
	if err != nil {
		return util.BuildAPIDiagnosticError(resourceName, fmt.Sprintf("Failed to create/update skill groups %s error: %s", skillGroupsRequest.Name, err), response)
	}

	//Get the results and pull out the id
	skillGroupPayload := make(map[string]interface{})
	err = json.Unmarshal(response.RawBody, &skillGroupPayload)
	if err != nil {
		return util.BuildDiagnosticError(resourceName, fmt.Sprintf("Failed to unmarshal skill groups"), err)
	}

	id := skillGroupPayload["id"].(string)
	d.SetId(id)
	log.Printf("Created skill group %s %s", skillGroupsRequest.Name, id)

	// Update member division IDs
	apiSkillGroupMemberDivisionIds, diagErr := readSkillGroupMemberDivisions(ctx, d, meta)
	if diagErr != nil {
		return diagErr
	}

	diagErr = createRoutingSkillGroupsMemberDivisions(ctx, d, meta, apiSkillGroupMemberDivisionIds, true)
	if diagErr != nil {
		return diagErr
	}

	return readSkillGroups(ctx, d, meta)
}

func updateSkillGroups(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig

	skillGroupsRequest := &SkillGroupsRequest{
		Name:        d.Get("name").(string),
		Description: d.Get("description").(string),
	}
	divisionId := d.Get("division_id").(string)
	if divisionId != "" {
		skillGroupsRequest.Division.ID = divisionId
	}

	finalSkillGroupsJson, err := mergeSkillConditionsIntoSkillGroups(d, skillGroupsRequest)
	if err != nil {
		return util.BuildDiagnosticError(resourceName, fmt.Sprintf("Failed to read the before skills groups request before: %s", skillGroupsRequest.Name), err)
	}

	routingAPI := platformclientv2.NewRoutingApiWithConfig(sdkConfig)
	apiClient := &routingAPI.Configuration.APIClient
	path := routingAPI.Configuration.BasePath + "/api/v2/routing/skillgroups/" + d.Id()
	headerParams := BuildHeaderParams(routingAPI)

	var skillGroupsPayload map[string]interface{}
	err = json.Unmarshal([]byte(finalSkillGroupsJson), &skillGroupsPayload)
	if err != nil {
		return util.BuildDiagnosticError(resourceName, fmt.Sprintf("Failed to unmarshal the JSON payload while creating/updating the skills group %s", skillGroupsRequest.Name), err)
	}

	response, err := apiClient.CallAPI(path, "PATCH", skillGroupsPayload, headerParams, nil, nil, "", nil)
	if err != nil {
		return util.BuildAPIDiagnosticError(resourceName, fmt.Sprintf("Failed to create/update skill groups %s error: %s", skillGroupsRequest.Name, err), response)
	}

	//Get the results and pull out the id
	skillGroupPayload := make(map[string]interface{})
	err = json.Unmarshal(response.RawBody, &skillGroupPayload)
	if err != nil {
		return util.BuildDiagnosticError(resourceName, fmt.Sprintf("Failed to unmarshal skill groups"), err)
	}

	log.Printf("Updated skill group %s", skillGroupsRequest.Name)

	// Update member division IDs
	apiSkillGroupMemberDivisionIds, diagErr := readSkillGroupMemberDivisions(ctx, d, meta)
	if diagErr != nil {
		return diagErr
	}

	diagErr = createRoutingSkillGroupsMemberDivisions(ctx, d, meta, apiSkillGroupMemberDivisionIds, false)
	if diagErr != nil {
		return diagErr
	}

	return readSkillGroups(ctx, d, meta)
}

func readSkillGroups(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getRoutingSkillGroupProxy(sdkConfig)
	cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceRoutingSkillGroup(), constants.DefaultConsistencyChecks, resourceName)

	log.Printf("Reading skills group %s", d.Id())

	return util.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		skillGroup, resp, err := proxy.getRoutingSkillGroupsById(ctx, d.Id())
		if err != nil {
			if util.IsStatus404(resp) {
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(resourceName, fmt.Sprintf("Failed to read skill groups %s | error: %s", d.Id(), err), resp))
			}
			return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(resourceName, fmt.Sprintf("Failed to read skill groups %s | error: %s", d.Id(), err), resp))
		}

		resourcedata.SetNillableValue(d, "name", skillGroup.Name)
		resourcedata.SetNillableValue(d, "description", skillGroup.Description)
		resourcedata.SetNillableValue(d, "division_id", skillGroup.Division.Id)

		// TODO: Handle Skill conditions
		skillGroupPayload := make(map[string]interface{})
		err = json.Unmarshal(resp.RawBody, &skillGroupPayload)
		if err != nil {
			return retry.NonRetryableError(fmt.Errorf("failed to unmarshal skill groups. %s", err))
		}
		skillConditionsBytes, _ := json.Marshal(skillGroupPayload["skillConditions"])
		skillConditions := string(skillConditionsBytes)
		if skillConditions != "" {
			d.Set("skill_conditions", skillConditions)
		} else {
			d.Set("skill_conditions", nil)
		}

		// Prepare member_division_ids list to avoid an unnecessary plan not empty error
		apiMemberDivisionIds, diagErr := readSkillGroupMemberDivisions(ctx, d, meta)
		if diagErr != nil {
			return retry.NonRetryableError(fmt.Errorf("%v", diagErr))
		}

		var schemaMemberDivisionIds []string
		if divIds, ok := d.Get("member_division_ids").([]interface{}); ok {
			schemaMemberDivisionIds = lists.InterfaceListToStrings(divIds)
		}

		memberDivisionIds := organizeMemberDivisionIdsForRead(schemaMemberDivisionIds, apiMemberDivisionIds, *skillGroup.Division.Id)
		_ = d.Set("member_division_ids", memberDivisionIds)

		log.Printf("Read skill groups name  %s %s", d.Id(), *skillGroup.Name)
		return cc.CheckState(d)
	})
}

func deleteSkillGroups(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getRoutingSkillGroupProxy(sdkConfig)

	log.Printf("Deleting skill group %s", d.Id())
	resp, err := proxy.deleteRoutingSkillGroup(ctx, d.Id())
	if err != nil {
		return util.BuildAPIDiagnosticError(resourceName, fmt.Sprintf("Failed to delete skill group %s error: %s", d.Id(), err), resp)
	}

	return util.WithRetries(ctx, 30*time.Second, func() *retry.RetryError {
		_, resp, err := proxy.getRoutingSkillGroupsById(ctx, d.Id())
		if err != nil {
			if util.IsStatus404(resp) {
				log.Printf("Deleted skills group %s", d.Id())
				return nil
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(resourceName, fmt.Sprintf("Error deleting skill group %s | error: %s", d.Id(), err), resp))
		}
		return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(resourceName, fmt.Sprintf("Skill group %s still exists", d.Id()), resp))
	})
}

func createRoutingSkillGroupsMemberDivisions(ctx context.Context, d *schema.ResourceData, meta interface{}, skillGroupDivisionIds []string, create bool) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getRoutingSkillGroupProxy(sdkConfig)
	name := d.Get("name").(string)
	memberDivisionIds := d.Get("member_division_ids").([]interface{})

	if memberDivisionIds == nil {
		return readSkillGroups(ctx, d, meta)
	}
	schemaDivisionIds := lists.InterfaceListToStrings(memberDivisionIds)

	toAdd, toRemove, diagErr := createListsForSkillgroupsMembersDivisionsPost(schemaDivisionIds, skillGroupDivisionIds, create, meta)
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

	log.Printf("Updating skill group %s member divisions", name)
	var reqBody platformclientv2.Skillgroupmemberdivisions

	if len(toRemove) > 0 {
		reqBody.RemoveDivisionIds = &toRemove
	}
	if len(toAdd) > 0 {
		reqBody.AddDivisionIds = &toAdd
	}

	resp, err := proxy.createRoutingSkillGroupsMemberDivision(ctx, d.Id(), reqBody)
	if err != nil {
		return util.BuildAPIDiagnosticError(resourceName, fmt.Sprintf("Failed to create/update skill group %s member divisions error: %s", d.Id(), err), resp)
	}

	log.Printf("Updated skill group %s member divisions", name)
	return nil
}

func readSkillGroupMemberDivisions(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]string, diag.Diagnostics) {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getRoutingSkillGroupProxy(sdkConfig)

	log.Printf("Reading skill group %s member divisions", d.Get("name").(string))

	divisions, resp, err := proxy.getRoutingSkillGroupsMemberDivison(ctx, d.Id())
	if err != nil {
		return nil, util.BuildAPIDiagnosticError(resourceName, fmt.Sprintf("Failed to  get member divisions for skill group %s error: %s", d.Id(), err), resp)
	}

	skillGroupMemberDivisionIds := make([]string, 0)
	for _, division := range *divisions.Entities {
		skillGroupMemberDivisionIds = append(skillGroupMemberDivisionIds, *division.Id)
	}

	log.Printf("Read skill group %s member divisions", d.Get("name").(string))

	return skillGroupMemberDivisionIds, nil
}

//
//
// Putting everything I have to refactor under this
//
//

type SkillGroupsRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Division    struct {
		ID   string `json:"id,omitempty"`
		Name string `json:"name,omitempty"`
	} `json:"division,omitempty"`
	SkillConditions struct{} `json:"skillConditions"` //Keep this here.  Even though we do not use this field in the struct The generated attributed is used as a placeholder
}

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
			resources[*division.Id] = &resourceExporter.ResourceMeta{Name: *division.Name}
		}
	}

	return resources, nil
}

func createListsForSkillgroupsMembersDivisionsPost(schemaMemberDivisionIds []string, apiMemberDivisionIds []string, create bool, meta interface{}) ([]string, []string, diag.Diagnostics) {
	toAdd := make([]string, 0)
	toRemove := make([]string, 0)

	if allMemberDivisionsSpecified(schemaMemberDivisionIds) {
		if len(schemaMemberDivisionIds) > 1 {
			return nil, nil, util.BuildDiagnosticError(resourceName, fmt.Sprintf(`member_division_ids should not contain more than one item when the value of an item is "*"`), fmt.Errorf(`member_division_ids should not contain more than one item when the value of an item is "*"`))
		}
		toAdd, err := getAllAuthDivisionIds(meta)
		return toAdd, nil, err
	}

	if len(schemaMemberDivisionIds) > 0 {
		if create {
			return schemaMemberDivisionIds, nil, nil
		}
		toAdd, toRemove = organizeMemberDivisionIdsForUpdate(schemaMemberDivisionIds, apiMemberDivisionIds)
		return toAdd, toRemove, nil
	}

	// Empty array - remove all
	toRemove = append(toRemove, apiMemberDivisionIds...)

	return nil, toRemove, nil
}

func GenerateRoutingSkillGroupResourceBasic(
	resourceID string,
	name string,
	description string) string {
	return fmt.Sprintf(`resource "%s" "%s" {
		name = "%s"
		description="%s"
	}
	`, resourceName, resourceID, name, description)
}
