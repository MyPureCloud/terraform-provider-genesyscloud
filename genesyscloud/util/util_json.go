package util

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"reflect"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// EquivalentJsons checks if two jsons are equivalent but has the special behavior
// where any null property in the 'incoming' json is removed(deeply) prior to deep comparison.
// Used to compare a json string from the terraform config and a json response from the API.
func EquivalentJsons(original, incoming string) bool {
	if original == incoming {
		return true
	}

	ob := bytes.NewBufferString("")
	if err := json.Compact(ob, []byte(original)); err != nil {
		log.Printf("error while comparing jsons: %v", err)
		return false
	}

	nb := bytes.NewBufferString("")
	if err := json.Compact(nb, []byte(incoming)); err != nil {
		log.Printf("error while comparing jsons: %v", err)
		return false
	}

	var nbi interface{}
	if err := json.Unmarshal(nb.Bytes(), &nbi); err != nil {
		log.Printf("error while comparing jsons: %v", err)
		return false
	}
	jsonDeepDelNullProperties(nbi)

	nbClean, err := json.Marshal(nbi)
	if err != nil {
		log.Printf("error while comparing jsons: %v", err)
		return false
	}

	return jsonBytesEqual(ob.Bytes(), nbClean)
}

// SuppressDiffFunc for properties that will accept JSON strings
func SuppressEquivalentJsonDiffs(k, old, new string, d *schema.ResourceData) bool {
	return EquivalentJsons(old, new)
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

func InterfaceToString(val interface{}) string {
	return fmt.Sprintf("%v", val)
}

func InterfaceToJson(val interface{}) (string, error) {
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

func MapToJson(m *map[string]interface{}) (string, error) {
	j, err := json.Marshal(*m)
	if err != nil {
		return "", fmt.Errorf("failed to marshal %v: %v", *m, err)
	}
	return string(j), nil
}
