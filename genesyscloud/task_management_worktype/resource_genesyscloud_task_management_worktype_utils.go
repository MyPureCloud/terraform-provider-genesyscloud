package task_management_worktype

import (
	"fmt"
	"terraform-provider-genesyscloud/genesyscloud/util/lists"
	"terraform-provider-genesyscloud/genesyscloud/util/resourcedata"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"
)

/*
The resource_genesyscloud_task_management_worktype_utils.go file contains various helper methods to marshal
and unmarshal data into formats consumable by Terraform and/or Genesys Cloud.
*/

type worktypeConfig struct {
	resID            string
	name             string
	description      string
	divisionId       string
	defaultWorkbinId string

	defaultDurationS    int
	defaultExpirationS  int
	defaultDueDurationS int
	defaultPriority     int
	defaultTtlS         int

	defaultLanguageId string
	defaultQueueId    string
	defaultSkillIds   []string
	assignmentEnabled bool

	schemaId      string
	schemaVersion int
}

// getWorktypeCreateFromResourceData maps data from schema ResourceData object to a platformclientv2.Worktypecreate
func getWorktypecreateFromResourceData(d *schema.ResourceData) platformclientv2.Worktypecreate {
	worktype := platformclientv2.Worktypecreate{
		Name:                         platformclientv2.String(d.Get("name").(string)),
		DivisionId:                   platformclientv2.String(d.Get("division_id").(string)),
		Description:                  platformclientv2.String(d.Get("description").(string)),
		DisableDefaultStatusCreation: platformclientv2.Bool(false),
		DefaultWorkbinId:             platformclientv2.String(d.Get("default_workbin_id").(string)),
		SchemaId:                     platformclientv2.String(d.Get("schema_id").(string)),
		SchemaVersion:                resourcedata.GetNillableValue[int](d, "schema_version"),

		DefaultPriority: platformclientv2.Int(d.Get("default_priority").(int)),

		DefaultLanguageId: resourcedata.GetNillableValue[string](d, "default_language_id"),
		DefaultQueueId:    resourcedata.GetNillableValue[string](d, "default_queue_id"),
		DefaultSkillIds:   lists.BuildSdkStringListFromInterfaceArray(d, "default_skills_ids"),
		AssignmentEnabled: platformclientv2.Bool(d.Get("assignment_enabled").(bool)),

		DefaultDurationSeconds:    resourcedata.GetNillableValue[int](d, "default_duration_seconds"),
		DefaultExpirationSeconds:  resourcedata.GetNillableValue[int](d, "default_expiration_seconds"),
		DefaultDueDurationSeconds: resourcedata.GetNillableValue[int](d, "default_due_duration_seconds"),
		DefaultTtlSeconds:         resourcedata.GetNillableValue[int](d, "default_ttl_seconds"),
	}

	return worktype
}

// getWorktypeupdateFromResourceData maps data from schema ResourceData object to a platformclientv2.Worktypeupdate
func getWorktypeupdateFromResourceData(d *schema.ResourceData) platformclientv2.Worktypeupdate {
	worktype := platformclientv2.Worktypeupdate{}
	worktype.SetField("Name", platformclientv2.String(d.Get("name").(string)))
	if d.HasChange("description") {
		worktype.SetField("Description", platformclientv2.String(d.Get("description").(string)))
	}
	if d.HasChange("default_workbin_id") {
		worktype.SetField("DefaultWorkbinId", platformclientv2.String(d.Get("default_workbin_id").(string)))
	}

	if d.HasChange("default_priority") {
		worktype.SetField("DefaultPriority", platformclientv2.Int(d.Get("default_priority").(int)))
	}

	if d.HasChange("schema_id") {
		worktype.SetField("SchemaId", platformclientv2.String(d.Get("schema_id").(string)))
	}

	if d.HasChange("default_language_id") {
		worktype.SetField("DefaultLanguageId", resourcedata.GetNillableValue[string](d, "default_language_id"))
	}

	if d.HasChange("default_queue_id") {
		worktype.SetField("DefaultQueueId", resourcedata.GetNillableValue[string](d, "default_queue_id"))
	}

	if d.HasChange("default_skills_ids") {
		worktype.SetField("DefaultSkillIds", lists.BuildSdkStringListFromInterfaceArray(d, "default_skills_ids"))
	}

	if d.HasChange("assignment_enabled") {
		worktype.SetField("AssignmentEnabled", platformclientv2.Bool(d.Get("assignment_enabled").(bool)))
	}

	if d.HasChange("schema_version") {
		worktype.SetField("SchemaVersion", resourcedata.GetNillableValue[int](d, "schema_version"))
	}

	if d.HasChange("default_duration_seconds") {
		worktype.SetField("DefaultDurationSeconds", resourcedata.GetNillableValue[int](d, "default_duration_seconds"))
	}
	if d.HasChange("default_expiration_seconds") {
		worktype.SetField("DefaultExpirationSeconds", resourcedata.GetNillableValue[int](d, "default_duration_seconds"))
	}
	if d.HasChange("default_due_duration_seconds") {
		worktype.SetField("DefaultDueDurationSeconds", resourcedata.GetNillableValue[int](d, "default_due_duration_seconds"))
	}
	if d.HasChange("default_ttl_seconds") {
		worktype.SetField("DefaultTtlSeconds", resourcedata.GetNillableValue[int](d, "default_ttl_seconds"))
	}

	return worktype
}

// flattenRoutingSkillReferences maps a Genesys Cloud *[]platformclientv2.Routingskillreference into a []interface{}
func flattenRoutingSkillReferences(routingSkillReferences *[]platformclientv2.Routingskillreference) []interface{} {
	if len(*routingSkillReferences) == 0 {
		return nil
	}

	var routingSkillReferenceList []interface{}
	for _, routingSkillReference := range *routingSkillReferences {
		routingSkillReferenceList = append(routingSkillReferenceList, *routingSkillReference.Id)
	}

	return routingSkillReferenceList
}

// GenerateWorktypeResourceBasic generates a terraform config string for a basic worktype
func GenerateWorktypeResourceBasic(resId, name, description, workbinResourceId, schemaResourceId, attrs string) string {
	return fmt.Sprintf(`resource "%s" "%s" {
		name = "%s"
		description = "%s"
		default_workbin_id = %s
		schema_id = %s
		%s
	}
	`, resourceName, resId, name, description, workbinResourceId, schemaResourceId, attrs)
}
