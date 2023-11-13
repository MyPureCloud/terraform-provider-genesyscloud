package genesyscloud

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"terraform-provider-genesyscloud/genesyscloud/consistency_checker"

	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	lists "terraform-provider-genesyscloud/genesyscloud/util/lists"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v115/platformclientv2"
)

type SkillGroupsRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Division    struct {
		ID   string `json:"id,omitempty"`
		Name string `json:"name,omitempty"`
	} `json:"division,omitempty"`
	SkillConditions struct{} `json:"skillConditions"` //Keep this here.  Even though we do not use this field in the struct The generated attributed is used as a placeholder
}

type AllSkillGroups struct {
	Entities []struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}
	NextURI     string `json:"nextUri"`
	SelfURI     string `json:"selfUri"`
	PreviousURI string `json:"previousUri"`
}

func getAllSkillGroups(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	resources := make(resourceExporter.ResourceIDMetaMap)
	routingAPI := platformclientv2.NewRoutingApiWithConfig(clientConfig)
	apiClient := &routingAPI.Configuration.APIClient
	route := "/api/v2/routing/skillgroups"

	headerParams := buildHeaderParams(routingAPI)

	for {
		path := routingAPI.Configuration.BasePath + route
		skillGroupPayload := &AllSkillGroups{}

		response, err := apiClient.CallAPI(path, "GET", nil, headerParams, nil, nil, "", nil)

		if err != nil {
			return nil, diag.Errorf("Failed to get page of skill groups: %s", err)

		}

		err = json.Unmarshal(response.RawBody, &skillGroupPayload)
		if err != nil {
			return nil, diag.Errorf("Failed to unmarshal skill groups. %s", err)
		}

		if skillGroupPayload.Entities == nil || len(skillGroupPayload.Entities) == 0 {
			break
		}

		for _, skillGroup := range skillGroupPayload.Entities {

			resources[skillGroup.ID] = &resourceExporter.ResourceMeta{Name: skillGroup.Name}
		}

		if route == skillGroupPayload.NextURI || skillGroupPayload.NextURI == "" {
			break
		} else {
			route = skillGroupPayload.NextURI
		}

	}

	return resources, nil
}

func ResourceSkillGroupExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: GetAllWithPooledClient(getAllSkillGroups),
		RefAttrs: map[string]*resourceExporter.RefAttrSettings{
			"division_id":         {RefType: "genesyscloud_auth_division"},
			"member_division_ids": {RefType: "genesyscloud_auth_division"},
		},
		RemoveIfMissing: map[string][]string{
			"division_id": {"division_id"},
		},
		JsonEncodeAttributes: []string{"skill_conditions"},
	}
}

func ResourceRoutingSkillGroup() *schema.Resource {
	return &schema.Resource{
		Description: `Genesys Cloud Skill Group`,

		CreateContext: CreateWithPooledClient(createSkillGroups),
		ReadContext:   ReadWithPooledClient(readSkillGroups),
		UpdateContext: UpdateWithPooledClient(updateSkillGroups),
		DeleteContext: DeleteWithPooledClient(deleteSkillGroups),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "The group name",
				Type:        schema.TypeString,
				Required:    true,
			},
			"description": {
				Description: "Description of the skill group",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
			},
			"division_id": {
				Description: "The division to which this entity belongs",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
			},
			"skill_conditions": {
				Description:      "JSON encoded array of rules that will be used to determine group membership.",
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				DiffSuppressFunc: SuppressEquivalentJsonDiffs,
			},
			"member_division_ids": {
				Description: "The IDs of member divisions to add or remove for this skill group. An empty array means all divisions will be removed, \"*\" means all divisions will be added.",
				Type:        schema.TypeList,
				MaxItems:    50,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func createSkillGroups(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return createOrUpdateSkillGroups(ctx, d, meta, "/api/v2/routing/skillgroups", true)
}

func createOrUpdateSkillGroups(ctx context.Context, d *schema.ResourceData, meta interface{}, route string, create bool) diag.Diagnostics {
	if create == true {
		log.Printf("Creating a skill group using")
	} else {
		log.Printf("Updating a skill group using")
	}
	// TODO: After public API endpoint is published and exposed to public, change to SDK method instead of direct invocation
	skillGroupsRequest := &SkillGroupsRequest{
		Name:        d.Get("name").(string),
		Description: d.Get("description").(string),
	}

	//Get the division information
	divisionId := d.Get("division_id").(string)
	if divisionId != "" {
		skillGroupsRequest.Division.ID = divisionId
	}

	//Merge in skill conditions
	finalSkillGroupsJson, err := mergeSkillConditionsIntoSkillGroups(d, skillGroupsRequest)
	if err != nil {
		return diag.Errorf("Failed to read the before skills groups request before: %s: %s", skillGroupsRequest.Name, err)
	}

	sdkConfig := meta.(*ProviderMeta).ClientConfig
	routingAPI := platformclientv2.NewRoutingApiWithConfig(sdkConfig)
	apiClient := &routingAPI.Configuration.APIClient
	path := routingAPI.Configuration.BasePath + route
	headerParams := buildHeaderParams(routingAPI)

	/*
	   Since API Client expects either a struct or map of maps (of json), convert the JSON string to a map
	   and then pass it into API client
	*/
	var skillGroupsPayload map[string]interface{}
	err = json.Unmarshal([]byte(finalSkillGroupsJson), &skillGroupsPayload)
	if err != nil {
		return diag.Errorf("Failed to unmarshal the JSON payload while creating/updating the skills group %s: %s", skillGroupsRequest.Name, err)
	}

	httpMethod := "POST"
	if create == false {
		httpMethod = "PATCH"
	}

	response, err := apiClient.CallAPI(path, httpMethod, skillGroupsPayload, headerParams, nil, nil, "", nil)
	if err != nil {
		return diag.Errorf("Failed to create/update skill groups %s: %s", skillGroupsRequest.Name, err)
	}

	//Get the results and pull out the id
	skillGroupPayload := make(map[string]interface{})
	err = json.Unmarshal(response.RawBody, &skillGroupPayload)
	if err != nil {
		return diag.Errorf("Failed to unmarshal skill groups. %s", err)
	}

	if create == true {
		id := skillGroupPayload["id"].(string)
		d.SetId(id)
		log.Printf("Created skill group %s %s", skillGroupsRequest.Name, id)
	} else {
		log.Printf("Updated skill group %s", skillGroupsRequest.Name)
	}

	// Update member division IDs
	apiSkillGroupMemberDivisionIds, diagErr := readSkillGroupMemberDivisionIds(d, routingAPI)
	if diagErr != nil {
		return diagErr
	}

	diagErr = postSkillGroupMemberDivisions(ctx, d, meta, routingAPI, apiSkillGroupMemberDivisionIds, create)
	if diagErr != nil {
		return diagErr
	}

	return readSkillGroups(ctx, d, meta)
}

func postSkillGroupMemberDivisions(ctx context.Context, d *schema.ResourceData, meta interface{}, routingAPI *platformclientv2.RoutingApi, apiSkillGroupMemberDivisionIds []string, create bool) diag.Diagnostics {
	name := d.Get("name").(string)
	memberDivisionIds := d.Get("member_division_ids").([]interface{})

	if memberDivisionIds == nil {
		return readSkillGroups(ctx, d, meta)
	}
	schemaDivisionIds := lists.InterfaceListToStrings(memberDivisionIds)

	toAdd, toRemove, diagErr := createListsForSkillgroupsMembersDivisionsPost(schemaDivisionIds, apiSkillGroupMemberDivisionIds, create, meta)
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

	skillGroupsMemberDivisionIdsPayload := make(map[string][]string, 0)
	if len(toRemove) > 0 {
		skillGroupsMemberDivisionIdsPayload["removeDivisionIds"] = toRemove
	}
	if len(toAdd) > 0 {
		skillGroupsMemberDivisionIdsPayload["addDivisionIds"] = toAdd
	}

	headerParams := buildHeaderParams(routingAPI)
	apiClient := &routingAPI.Configuration.APIClient
	path := fmt.Sprintf("%s/api/v2/routing/skillgroups/%s/members/divisions", routingAPI.Configuration.BasePath, d.Id())
	response, err := apiClient.CallAPI(path, "POST", skillGroupsMemberDivisionIdsPayload, headerParams, nil, nil, "", nil)
	if err != nil || response.Error != nil {
		return diag.Errorf("Failed to create/update skill group %s member divisions: %s", name, err)
	}

	log.Printf("Updated skill group %s member divisions", name)
	return nil
}

func getAllAuthDivisionIds(meta interface{}) ([]string, diag.Diagnostics) {
	sdkConfig := meta.(*ProviderMeta).ClientConfig
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

func createListsForSkillgroupsMembersDivisionsPost(schemaMemberDivisionIds []string, apiMemberDivisionIds []string,
	create bool, meta interface{}) ([]string, []string, diag.Diagnostics) {
	toAdd := make([]string, 0)
	toRemove := make([]string, 0)

	if allMemberDivisionsSpecified(schemaMemberDivisionIds) {
		if len(schemaMemberDivisionIds) > 1 {
			return nil, nil, diag.Errorf(`member_division_ids should not contain more than one item when the value of an item is "*"`)
		}
		toAdd, err := getAllAuthDivisionIds(meta)
		return toAdd, nil, err
	}

	if len(schemaMemberDivisionIds) > 0 {
		if create == true {
			return schemaMemberDivisionIds, nil, nil
		}
		toAdd, toRemove = organizeMemberDivisionIdsForUpdate(schemaMemberDivisionIds, apiMemberDivisionIds)
		return toAdd, toRemove, nil
	}

	// Empty array - remove all
	for _, id := range apiMemberDivisionIds {
		toRemove = append(toRemove, id)
	}
	return nil, toRemove, nil
}

/*
Sometimes you just need to get ugly.  skillConditions has a recursive function that is super ugly to manage to a static Golang
Struct.  So our struct always has a placeholder "skillConditions": {} field. So what I do is convert the struct to JSON and then
check to see if skill_conditions on the Terraform resource data.  I then do a string replace on the skillConditions json attribute
and replace the empty stringConditions string with the contents of skill_conditions.

Not the most eloquent code, but these are uncivilized times.
*/
func mergeSkillConditionsIntoSkillGroups(d *schema.ResourceData, skillGroupsRequest *SkillGroupsRequest) (string, error) {
	skillsConditionsJsonString := fmt.Sprintf(`"skillConditions": %s`, d.Get("skill_conditions").(string))

	//Get the before image of the JSON.  Note this a byte array
	skillGroupsRequestBefore, err := json.Marshal(skillGroupsRequest)
	if err != nil {
		return "", err
	}

	skillGroupsRequestAfter := ""

	//Skill conditions are present, replace skill conditions with the content of the string
	if d.Get("skill_conditions").(string) != "" {
		skillGroupsRequestAfter = strings.Replace(string(skillGroupsRequestBefore), `"skillConditions":{}`, skillsConditionsJsonString, 1)
	} else {
		//Skill conditions are not present, get rid of skill conditions.
		skillGroupsRequestAfter = strings.Replace(string(skillGroupsRequestBefore), `,"skillConditions":{}`, "", 1)
	}

	return skillGroupsRequestAfter, nil
}

func readSkillGroups(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*ProviderMeta).ClientConfig
	routingAPI := platformclientv2.NewRoutingApiWithConfig(sdkConfig)

	// TODO: After public API endpoint is published and exposed to public, change to SDK method instead of direct invocation
	apiClient := &routingAPI.Configuration.APIClient
	path := routingAPI.Configuration.BasePath + "/api/v2/routing/skillgroups/" + d.Id()

	// add default headers if any
	headerParams := buildHeaderParams(routingAPI)

	log.Printf("Reading skills group %s", d.Id())

	return WithRetriesForRead(ctx, d, func() *retry.RetryError {

		skillGroupPayload := make(map[string]interface{})
		response, err := apiClient.CallAPI(path, "GET", nil, headerParams, nil, nil, "", nil)

		if err != nil {
			return retry.NonRetryableError(fmt.Errorf("Failed to retrieve skill groups %s", err))
		}

		if err == nil && response.Error != nil && response.StatusCode != http.StatusNotFound {
			return retry.NonRetryableError(fmt.Errorf("Failed to retrieve skill groups. %s", err))
		}

		err = json.Unmarshal(response.RawBody, &skillGroupPayload)
		if err != nil {
			return retry.NonRetryableError(fmt.Errorf("Failed to unmarshal skill groups. %s", err))
		}

		if err == nil && IsStatus404(response) {
			return retry.RetryableError(fmt.Errorf("Failed to read skill groups %s: %s", d.Id(), err))
		}

		cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceRoutingSkillGroup())

		name := skillGroupPayload["name"]
		divisionId := skillGroupPayload["division"].(map[string]interface{})["id"]
		description := skillGroupPayload["description"]
		skillConditionsBytes, err := json.Marshal(skillGroupPayload["skillConditions"])

		if err != nil {
			return retry.NonRetryableError(fmt.Errorf("Failed to unmarshal skill groups conditions. %s", err))
		}

		skillConditions := string(skillConditionsBytes)

		if name != "" {
			d.Set("name", name)
		} else {
			d.Set("name", nil)
		}

		if divisionId != nil && divisionId != "" {
			d.Set("division_id", divisionId)
		} else {
			d.Set("division_id", nil)
		}

		if description != "" {
			d.Set("description", description)
		} else {
			d.Set("description", nil)
		}

		if skillConditions != "" {
			d.Set("skill_conditions", skillConditions)
		} else {
			d.Set("skill_conditions", nil)
		}

		apiMemberDivisionIds, diagErr := readSkillGroupMemberDivisionIds(d, routingAPI)
		if diagErr != nil {
			return retry.NonRetryableError(fmt.Errorf("%v", diagErr))
		}

		var schemaMemberDivisionIds []string
		if divIds, ok := d.Get("member_division_ids").([]interface{}); ok {
			schemaMemberDivisionIds = lists.InterfaceListToStrings(divIds)
		}

		memberDivisionIds := organizeMemberDivisionIdsForRead(schemaMemberDivisionIds, apiMemberDivisionIds, divisionId.(string))
		_ = d.Set("member_division_ids", memberDivisionIds)

		log.Printf("Read skill groups name  %s %s", d.Id(), name)
		return cc.CheckState()
	})
}

func buildHeaderParams(routingAPI *platformclientv2.RoutingApi) map[string]string {
	headerParams := make(map[string]string)

	for key := range routingAPI.Configuration.DefaultHeader {
		headerParams[key] = routingAPI.Configuration.DefaultHeader[key]
	}

	headerParams["Authorization"] = "Bearer " + routingAPI.Configuration.AccessToken
	headerParams["Content-Type"] = "application/json"
	headerParams["Accept"] = "application/json"

	return headerParams
}

func updateSkillGroups(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	route := "/api/v2/routing/skillgroups/" + d.Id()
	return createOrUpdateSkillGroups(ctx, d, meta, route, false)
}

func deleteSkillGroups(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)

	sdkConfig := meta.(*ProviderMeta).ClientConfig
	routingAPI := platformclientv2.NewRoutingApiWithConfig(sdkConfig)

	// TODO: After public API endpoint is published and exposed to public, change to SDK method instead of direct invocation
	apiClient := &routingAPI.Configuration.APIClient
	path := routingAPI.Configuration.BasePath + "/api/v2/routing/skillgroups/" + d.Id()

	// add default headers if any
	headerParams := buildHeaderParams(routingAPI)

	log.Printf("Deleting skills group %s", name)
	response, err := apiClient.CallAPI(path, "DELETE", nil, headerParams, nil, nil, "", nil)

	if err != nil {
		if IsStatus404(response) {
			//Skills Group already deleted
			log.Printf("Skills Group was already deleted %s", d.Id())
			return nil
		}
		return diag.Errorf("Failed to delete skills group %s: %s", d.Id(), err)
	}

	return WithRetries(ctx, 30*time.Second, func() *retry.RetryError {
		log.Printf("Deleting skills group %s", name)
		response, err := apiClient.CallAPI(path, "DELETE", nil, headerParams, nil, nil, "", nil)

		if err != nil {
			if IsStatus404(response) {
				// Skills Group Deleted
				log.Printf("Deleted skills group %s", d.Id())
				return nil
			}
			return retry.NonRetryableError(fmt.Errorf("Error deleting skill group %s: %s", d.Id(), err))
		}
		return retry.RetryableError(fmt.Errorf("Skill group %s still exists", d.Id()))
	})
}

func readSkillGroupMemberDivisionIds(d *schema.ResourceData, routingAPI *platformclientv2.RoutingApi) ([]string, diag.Diagnostics) {
	headers := buildHeaderParams(routingAPI)
	apiClient := &routingAPI.Configuration.APIClient
	path := fmt.Sprintf("%s/api/v2/routing/skillgroups/%s/members/divisions", routingAPI.Configuration.BasePath, d.Id())

	log.Printf("Reading skill group %s member divisions", d.Get("name").(string))

	response, err := apiClient.CallAPI(path, "GET", nil, headers, nil, nil, "", nil)
	if err != nil || response.Error != nil {
		return nil, diag.Errorf("Failed to get member divisions for skill group %s: %v", d.Id(), err)
	}

	memberDivisionsPayload := make(map[string]interface{}, 0)
	err = json.Unmarshal(response.RawBody, &memberDivisionsPayload)
	if err != nil {
		return nil, diag.Errorf("Failed to unmarshal member divisions. %s", err)
	}

	apiSkillGroupMemberDivisionIds := make([]string, 0)
	entities := memberDivisionsPayload["entities"].([]interface{})
	for _, entity := range entities {
		if entityMap, ok := entity.(map[string]interface{}); ok {
			apiSkillGroupMemberDivisionIds = append(apiSkillGroupMemberDivisionIds, entityMap["id"].(string))
		}
	}

	log.Printf("Read skill group %s member divisions", d.Get("name").(string))

	return apiSkillGroupMemberDivisionIds, nil
}

func allMemberDivisionsSpecified(schemaSkillGroupMemberDivisionIds []string) bool {
	return lists.ItemInSlice("*", schemaSkillGroupMemberDivisionIds)
}

func organizeMemberDivisionIdsForUpdate(schemaIds, apiIds []string) ([]string, []string) {
	toAdd := make([]string, 0)
	toRemove := make([]string, 0)
	// items that are in hcl and not in api-returned list - add
	for _, id := range schemaIds {
		if !lists.ItemInSlice(id, apiIds) {
			toAdd = append(toAdd, id)
		}
	}
	// items that are not in hcl and are in api-returned list - remove
	for _, id := range apiIds {
		if !lists.ItemInSlice(id, schemaIds) {
			toRemove = append(toRemove, id)
		}
	}
	return toAdd, toRemove
}

// Prepare member_division_ids list to avoid an unnecessary plan not empty error
func organizeMemberDivisionIdsForRead(schemaList, apiList []string, divisionId string) []string {
	if !lists.ItemInSlice(divisionId, schemaList) {
		apiList = lists.RemoveStringFromSlice(divisionId, apiList)
	}
	if len(schemaList) == 1 && schemaList[0] == "*" {
		return schemaList
	} else {
		// if hcl & api lists are the same but with different ordering - set with original ordering
		if lists.AreEquivalent(schemaList, apiList) {
			return schemaList
		} else {
			return apiList
		}
	}
}

// Remove the value of division_id, or if this field was left blank; the home division ID
func removeSkillGroupDivisionID(d *schema.ResourceData, list []string) ([]string, diag.Diagnostics) {
	if len(list) == 0 || list == nil {
		return list, nil
	}
	divisionId := d.Get("division_id").(string)
	if divisionId == "" {
		id, diagErr := getHomeDivisionID()
		if diagErr != nil {
			return nil, diagErr
		}
		divisionId = id
	}
	if lists.ItemInSlice(divisionId, list) {
		list = lists.RemoveStringFromSlice(divisionId, list)
	}
	return list, nil
}
