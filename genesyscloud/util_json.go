package genesyscloud

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// SuppressDiffFunc for properties that will accept JSON strings
// Any null property in the 'new' json is removed(deeply) prior to deep comparison
func SuppressEquivalentJsonDiffs(k, old, new string, d *schema.ResourceData) bool {
	if old == new {
		return true
	}

	ob := bytes.NewBufferString("")
	if err := json.Compact(ob, []byte(old)); err != nil {
		return false
	}

	nb := bytes.NewBufferString("")
	if err := json.Compact(nb, []byte(new)); err != nil {
		return false
	}

	var nbi interface{}
	if err := json.Unmarshal(nb.Bytes(), &nbi); err != nil {
		return false
	}
	jsonDeepDelNullProperties(nbi)

	nbClean, err := json.Marshal(nbi)
	if err != nil {
		return false
	}

	return jsonBytesEqual(ob.Bytes(), nbClean)
}

// Recursively go through decoded JSON map and remove any property that is null.
// Parameter can also be an array(slice) like in JSON but will only traverse for
// further map/slice elements. ie null elements in slices are not removed.
func jsonDeepDelNullProperties(o interface{}) {
	ov := reflect.ValueOf(o)
	switch ov.Kind() {
	case reflect.Slice:
		for _, n := range o.([]any) {
			jsonDeepDelNullProperties(n)
		}
	case reflect.Map:
		for key, value := range o.(map[string]interface{}) {
			if value == nil {
				delete(o.(map[string]interface{}), key)
			}
			v1 := reflect.ValueOf(value)
			if v1.Kind() == reflect.Map {
				jsonDeepDelNullProperties(value)
			}
		}
	}
}

func jsonBytesEqual(b1, b2 []byte) bool {
	var o1 interface{}
	if err := json.Unmarshal(b1, &o1); err != nil {
		return false
	}

	var o2 interface{}
	if err := json.Unmarshal(b2, &o2); err != nil {
		return false
	}

	return reflect.DeepEqual(o1, o2)
}

func interfaceToString(val interface{}) string {
	return fmt.Sprintf("%v", val)
}

func interfaceToJson(val interface{}) (string, error) {
	j, err := json.Marshal(val)
	if err != nil {
		return "", fmt.Errorf("failed to marshal %v: %v", val, err)
	}
	return string(j), nil
}

func JsonStringToInterface(jsonStr string) (interface{}, error) {
	var obj interface{}
	err := json.Unmarshal([]byte(jsonStr), &obj)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal %s: %v", jsonStr, err)
	}
	return obj, nil
}
