package task_management_workitem_schema

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v115/platformclientv2"

	gcloud "terraform-provider-genesyscloud/genesyscloud"
)

// JsonToJsonSchemaDocument converts the json input configuration string to a Jsonschemadocument
// for use on the API
func jsonToJsonSchemaDocument(rawJson string) (*platformclientv2.Jsonschemadocument, error) {
	schema, err := gcloud.JsonStringToInterface(rawJson)
	if err != nil {
		return nil, fmt.Errorf("failed to convert to jsonschemadocument: %v", err)
	}

	schemaMap, ok := schema.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("schema is not type map[string]interface{}: %v", err)
	}

	return &platformclientv2.Jsonschemadocument{
		Schema:      mapValuePtrOrNil[string](schemaMap, "$schema"),
		Title:       mapValuePtrOrNil[string](schemaMap, "title"),
		Description: mapValuePtrOrNil[string](schemaMap, "description"),
		VarType:     mapValuePtrOrNil[string](schemaMap, "type"),
		Properties:  mapValuePtrOrNil[map[string]interface{}](schemaMap, "properties"),
	}, nil
}

// A DiffSuppressFunction to evaluate if the string JSON as configured in the terraform resource is equivalent
// to a GC Jsonschemadocument returned as the Workitem Schema.
func suppressEquivalentJsonSchemas(k, old, new string, d *schema.ResourceData) bool {
	return true
}

func mapValuePtrOrNil[T any](src map[string]interface{}, key string) *T {
	val, ok := src[key]
	if !ok {
		return nil
	}
	typedVal, ok := val.(T)
	if !ok {
		return nil
	}
	return &typedVal
}
