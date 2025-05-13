package task_management_worktype_status_transition

import (
	"encoding/json"
	"reflect"
	"strconv"
	"strings"

	platformUtils "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/platform"

	"github.com/leekchan/timeutil"
)

type Workitemstatusupdate struct {
	// SetFieldNames defines the list of fields to use for controlled JSON serialization
	SetFieldNames map[string]bool `json:"-"`
	// Name - The name of the Status. Valid length between 3 and 256 characters.
	Name *string `json:"name,omitempty"`

	// DestinationStatusIds - A list of destination Statuses where a Workitem with this Status can transition to. If the list is empty Workitems with this Status can transition to all other Statuses defined on the Worktype. A Status can have a maximum of 24 destinations.
	DestinationStatusIds *[]string `json:"destinationStatusIds,omitempty"`

	// Description - The description of the Status. Maximum length of 512 characters.
	Description *string `json:"description,omitempty"`

	// DefaultDestinationStatusId - Default destination status to which this Status will transition to if auto status transition enabled.
	DefaultDestinationStatusId *string `json:"defaultDestinationStatusId"` // Intentionally does not include "omitempty" to enable sending a null value to the API to unset this field

	// StatusTransitionDelaySeconds - Delay in seconds for auto status transition. Required if defaultDestinationStatusId is provided.
	StatusTransitionDelaySeconds *int `json:"statusTransitionDelaySeconds,omitempty"`

	// StatusTransitionTime - Time is represented as an ISO-8601 string without a timezone. For example: HH:mm:ss.SSS
	StatusTransitionTime *string `json:"statusTransitionTime,omitempty"`

	// AutoTerminateWorkitem - Terminate workitem on selection of status. Applicable only for statuses in the Closed category.
	AutoTerminateWorkitem *bool `json:"autoTerminateWorkitem,omitempty"`
}

// SetField uses reflection to set a field on the model if the model has a property SetFieldNames, and triggers custom JSON serialization logic to only serialize properties that have been set using this function.
func (o *Workitemstatusupdate) SetField(field string, fieldValue interface{}) {
	// Get Value object for field
	target := reflect.ValueOf(o)
	targetField := reflect.Indirect(target).FieldByName(field)

	// Set value
	if fieldValue != nil {
		targetField.Set(reflect.ValueOf(fieldValue))
	} else {
		// Must create a new Value (creates **type) then get its element (*type), which will be nil pointer of the appropriate type
		x := reflect.Indirect(reflect.New(targetField.Type()))
		targetField.Set(x)
	}

	// Add field to set field names list
	if o.SetFieldNames == nil {
		o.SetFieldNames = make(map[string]bool)
	}
	o.SetFieldNames[field] = true
}

func (o Workitemstatusupdate) MarshalJSON() ([]byte, error) {
	// Special processing to dynamically construct object using only field names that have been set using SetField. This generates payloads suitable for use with PATCH API endpoints.
	if len(o.SetFieldNames) > 0 {
		// Get reflection Value
		val := reflect.ValueOf(o)

		// Known field names that require type overrides
		dateTimeFields := []string{}
		localDateTimeFields := []string{}
		dateFields := []string{}

		// Construct object
		newObj := make(map[string]interface{})
		for fieldName := range o.SetFieldNames {
			// Get initial field value
			fieldValue := val.FieldByName(fieldName).Interface()

			// Apply value formatting overrides
			if fieldValue == nil || reflect.ValueOf(fieldValue).IsNil() {
				// Do nothing. Just catching this case to avoid trying to custom serialize a nil value
			} else if platformUtils.Contains(dateTimeFields, fieldName) {
				fieldValue = timeutil.Strftime(platformUtils.ToTime(fieldValue), "%Y-%m-%dT%H:%M:%S.%fZ")
			} else if platformUtils.Contains(localDateTimeFields, fieldName) {
				fieldValue = timeutil.Strftime(platformUtils.ToTime(fieldValue), "%Y-%m-%dT%H:%M:%S.%f")
			} else if platformUtils.Contains(dateFields, fieldName) {
				fieldValue = timeutil.Strftime(platformUtils.ToTime(fieldValue), "%Y-%m-%d")
			}

			// Assign value to field using JSON tag name
			newObj[platformUtils.GetFieldName(reflect.TypeOf(&o), fieldName)] = fieldValue
		}

		// Marshal and return dynamically constructed interface
		return json.Marshal(newObj)
	}

	// Redundant initialization to avoid unused import errors for models with no Time values
	_ = timeutil.Timedelta{}
	type Alias Workitemstatusupdate

	return json.Marshal(&struct {
		Name *string `json:"name,omitempty"`

		DestinationStatusIds *[]string `json:"destinationStatusIds,omitempty"`

		Description *string `json:"description,omitempty"`

		DefaultDestinationStatusId *string `json:"defaultDestinationStatusId"` // Intentionally does not include "omitempty" to enable sending a null value to the API to unset this field

		StatusTransitionDelaySeconds *int `json:"statusTransitionDelaySeconds,omitempty"`

		StatusTransitionTime *string `json:"statusTransitionTime,omitempty"`

		AutoTerminateWorkitem *bool `json:"autoTerminateWorkitem,omitempty"`
		Alias
	}{
		Name: o.Name,

		DestinationStatusIds: o.DestinationStatusIds,

		Description: o.Description,

		DefaultDestinationStatusId: o.DefaultDestinationStatusId,

		StatusTransitionDelaySeconds: o.StatusTransitionDelaySeconds,

		StatusTransitionTime: o.StatusTransitionTime,

		AutoTerminateWorkitem: o.AutoTerminateWorkitem,
		Alias:                 (Alias)(o),
	})
}

func (o *Workitemstatusupdate) UnmarshalJSON(b []byte) error {
	var WorkitemstatusupdateMap map[string]interface{}
	err := json.Unmarshal(b, &WorkitemstatusupdateMap)
	if err != nil {
		return err
	}

	if Name, ok := WorkitemstatusupdateMap["name"].(string); ok {
		o.Name = &Name
	}

	if DestinationStatusIds, ok := WorkitemstatusupdateMap["destinationStatusIds"].([]interface{}); ok {
		DestinationStatusIdsString, _ := json.Marshal(DestinationStatusIds)
		json.Unmarshal(DestinationStatusIdsString, &o.DestinationStatusIds)
	}

	if Description, ok := WorkitemstatusupdateMap["description"].(string); ok {
		o.Description = &Description
	}

	if DefaultDestinationStatusId, ok := WorkitemstatusupdateMap["defaultDestinationStatusId"].(string); ok {
		o.DefaultDestinationStatusId = &DefaultDestinationStatusId
	}

	if StatusTransitionDelaySeconds, ok := WorkitemstatusupdateMap["statusTransitionDelaySeconds"].(float64); ok {
		StatusTransitionDelaySecondsInt := int(StatusTransitionDelaySeconds)
		o.StatusTransitionDelaySeconds = &StatusTransitionDelaySecondsInt
	}

	if StatusTransitionTime, ok := WorkitemstatusupdateMap["statusTransitionTime"].(string); ok {
		o.StatusTransitionTime = &StatusTransitionTime
	}

	if AutoTerminateWorkitem, ok := WorkitemstatusupdateMap["autoTerminateWorkitem"].(bool); ok {
		o.AutoTerminateWorkitem = &AutoTerminateWorkitem
	}

	return nil
}

// String returns a JSON representation of the model
func (o *Workitemstatusupdate) String() string {
	j, _ := json.Marshal(o)
	str, _ := strconv.Unquote(strings.Replace(strconv.Quote(string(j)), `\\u`, `\u`, -1))

	return str
}
