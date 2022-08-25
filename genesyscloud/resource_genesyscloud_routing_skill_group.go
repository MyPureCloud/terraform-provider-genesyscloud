package genesyscloud

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v77/platformclientv2"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/consistency_checker"
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

func getAllSkillGroups(ctx context.Context, clientConfig *platformclientv2.Configuration) (ResourceIDMetaMap, diag.Diagnostics) {
	resources := make(ResourceIDMetaMap)
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

			resources[skillGroup.ID] = &ResourceMeta{Name: skillGroup.Name}
		}

		if route == skillGroupPayload.NextURI || skillGroupPayload.NextURI == "" {
			break
		} else {
			route = skillGroupPayload.NextURI
		}

	}

	return resources, nil
}

func resourceSkillGroupExporter() *ResourceExporter {
	return &ResourceExporter{
		GetResourcesFunc: getAllWithPooledClient(getAllSkillGroups),
		RefAttrs: map[string]*RefAttrSettings{
			"division_id": {RefType: "genesyscloud_auth_division"},
		},
		RemoveIfMissing: map[string][]string{
			"division_id": {"division_id"},
		},
		JsonEncodeAttributes: []string{"skill_conditions"},
	}
}

func resourceRoutingSkillGroup() *schema.Resource {
	return &schema.Resource{
		Description: `Genesys Cloud Skill Group`,

		CreateContext: createWithPooledClient(createSkillGroups),
		ReadContext:   readWithPooledClient(readSkillGroups),
		UpdateContext: updateWithPooledClient(updateSkillGroups),
		DeleteContext: deleteWithPooledClient(deleteSkillGroups),
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
				DiffSuppressFunc: suppressEquivalentJsonDiffs,
			},
		},
	}
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

	sdkConfig := meta.(*providerMeta).ClientConfig
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

	return readSkillGroups(ctx, d, meta)
}

func createSkillGroups(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return createOrUpdateSkillGroups(ctx, d, meta, "/api/v2/routing/skillgroups", true)
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
	sdkConfig := meta.(*providerMeta).ClientConfig
	routingAPI := platformclientv2.NewRoutingApiWithConfig(sdkConfig)

	// TODO: After public API endpoint is published and exposed to public, change to SDK method instead of direct invocation
	apiClient := &routingAPI.Configuration.APIClient
	path := routingAPI.Configuration.BasePath + "/api/v2/routing/skillgroups/" + d.Id()

	// add default headers if any
	headerParams := buildHeaderParams(routingAPI)

	log.Printf("Reading skills group %s", d.Id())

	return withRetriesForRead(ctx, d, func() *resource.RetryError {

		skillGroupPayload := make(map[string]interface{})
		response, err := apiClient.CallAPI(path, "GET", nil, headerParams, nil, nil, "", nil)

		if err != nil {
			return resource.NonRetryableError(fmt.Errorf("Failed to retrieve skill groups %s", err))
		}

		if err == nil && response.Error != nil && response.StatusCode != http.StatusNotFound {
			return resource.NonRetryableError(fmt.Errorf("Failed to retrieve skill groups. %s", err))
		}

		err = json.Unmarshal(response.RawBody, &skillGroupPayload)
		if err != nil {
			return resource.NonRetryableError(fmt.Errorf("Failed to unmarshal skill groups. %s", err))
		}

		if err != nil && isStatus404(response) {
			return resource.RetryableError(fmt.Errorf("Failed to read skill groups %s: %s", d.Id(), err))
		}

		cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, resourceRoutingSkillGroup())
		name := skillGroupPayload["name"]
		divisionId := skillGroupPayload["division"].(map[string]interface{})["id"]
		description := skillGroupPayload["description"]
		skillConditionsBytes, err := json.Marshal(skillGroupPayload["skillConditions"])

		if err != nil {
			return resource.NonRetryableError(fmt.Errorf("Failed to unmarshal skill groups conditions. %s", err))
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

	sdkConfig := meta.(*providerMeta).ClientConfig
	routingAPI := platformclientv2.NewRoutingApiWithConfig(sdkConfig)

	// TODO: After public API endpoint is published and exposed to public, change to SDK method instead of direct invocation
	apiClient := &routingAPI.Configuration.APIClient
	path := routingAPI.Configuration.BasePath + "/api/v2/routing/skillgroups/" + d.Id()

	// add default headers if any
	headerParams := buildHeaderParams(routingAPI)

	log.Printf("Deleting skills group %s", name)
	response, err := apiClient.CallAPI(path, "DELETE", nil, headerParams, nil, nil, "", nil)

	if err != nil {
		if isStatus404(response) {
			//Skills Group already deleted
			log.Printf("Skills Group was already deleted %s", d.Id())
			return nil
		}
		return diag.Errorf("Failed to delete skills group %s: %s", d.Id(), err)
	}

	return withRetries(ctx, 30*time.Second, func() *resource.RetryError {
		log.Printf("Deleting skills group %s", name)
		response, err := apiClient.CallAPI(path, "DELETE", nil, headerParams, nil, nil, "", nil)

		if err != nil {
			if isStatus404(response) {
				// Skills Group Deleted
				log.Printf("Deleted skills group %s", d.Id())
				return nil
			}
			return resource.NonRetryableError(fmt.Errorf("Error deleting skill group %s: %s", d.Id(), err))
		}
		return resource.RetryableError(fmt.Errorf("Skill group %s still exists", d.Id()))
	})
}
