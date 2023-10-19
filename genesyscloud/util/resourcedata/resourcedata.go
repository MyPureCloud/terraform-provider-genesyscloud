package resourcedata

import (
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/leekchan/timeutil"
	"github.com/mypurecloud/platform-client-sdk-go/v115/platformclientv2"
)

const (
	TimeWriteFormat = "%Y-%m-%dT%H:%M:%S.%f"
	TimeParseFormat = "2006-01-02T15:04:05.000000"
	DateParseFormat = "2006-01-02"
)

// Use these functions to read properties from the schema and set it on in a map in build function

// BuildSDKStringValueIfNotNil will read a map and set the string property on an object if the value exists
func BuildSDKStringValueIfNotNil(field **string, targetMap map[string]interface{}, key string) {
	if value := targetMap[key].(string); value != "" {
		*field = &value
	}
}

// BuildSDKInterfaceArrayValueIfNotNil will read a map and use the provided function to read the nested values if the value exists
func BuildSDKInterfaceArrayValueIfNotNil[T any](field **T, targetMap map[string]interface{}, key string, f func([]interface{}) *T) {
	if values := targetMap[key]; values != nil {
		*field = f(values.([]interface{}))
	}
}

// BuildSDKStringArrayValueIfNotNil will read a map and set the string[] property on an object if the value exists
func BuildSDKStringArrayValueIfNotNil(field **[]string, targetMap map[string]interface{}, key string) {
	array := make([]string, 0)
	for _, v := range targetMap[key].([]interface{}) {
		array = append(array, v.(string))
	}
	*field = &array
}

// BuildSDKStringMapValueIfNotNil will read a map and set the map[string][string] property on an object if the value exists
func BuildSDKStringMapValueIfNotNil(field **map[string]string, targetMap map[string]interface{}, key string) {
	if values := targetMap[key].(map[string]interface{}); values != nil {
		valueMap := map[string]string{}
		for k, v := range values {
			valueMap[k] = v.(string)
		}
		*field = &valueMap
	}
}

// Use these functions to read properties of objects inside flatten functions

// SetMapStringArrayValueIfNotNil will read the value of a string array property and set it in a map
func SetMapStringArrayValueIfNotNil(targetMap map[string]interface{}, key string, valueList *[]string) {
	if valueList != nil {
		array := make([]string, 0)
		array = append(array, *valueList...)
		targetMap[key] = array
	}
}

// SetMapStringMapValueIfNotNil will read the value of a map property and set it in a map
func SetMapStringMapValueIfNotNil(targetMap map[string]interface{}, key string, valueList *map[string]string) {
	if valueList != nil {
		results := make(map[string]interface{})
		for k, v := range *valueList {
			results[k] = v
		}
		targetMap[key] = results
	}
}

// SetMapReferenceValueIfNotNil will read the value of a reference property and set it in a map
func SetMapReferenceValueIfNotNil(targetMap map[string]interface{}, key string, value *platformclientv2.Domainentityref) {
	if value != nil && value.Id != nil {
		targetMap[key] = value.Id
	}
}

// SetMapValueIfNotNil will read the value of a basic type property and set it in a map
func SetMapValueIfNotNil[T any](targetMap map[string]interface{}, key string, value *T) {
	if value != nil {
		targetMap[key] = *value
	}
}

// SetMapInterfaceArrayWithFuncIfNotNil will read the values in a nested resource using the provided function and set it in a map
func SetMapInterfaceArrayWithFuncIfNotNil[T any](targetMap map[string]interface{}, key string, value *T, f func(*T) []interface{}) {
	if value != nil {
		targetMap[key] = f(value)
	}
}

// Use these functions to read values for an object and set them on the schema

// SetNillableReference will read the value of a reference property and set it on the schema
func SetNillableReference(d *schema.ResourceData, key string, value *platformclientv2.Domainentityref) {
	if value != nil && value.Id != nil {
		d.Set(key, value.Id)
	} else {
		d.Set(key, nil)
	}
}

// SetNillableReferenceWritableDivision functions the same as SetNillableReference, but for fields that are type Writabledivision, not Domainentityref
func SetNillableReferenceWritableDivision(d *schema.ResourceData, key string, value *platformclientv2.Writabledivision) {
	if value != nil && value.Id != nil {
		_ = d.Set(key, *value.Id)
	} else {
		_ = d.Set(key, nil)
	}
}

// SetNillableValue will read a basic type and set it on the schema
func SetNillableValue[T any](d *schema.ResourceData, key string, value *T) {
	if value != nil {
		d.Set(key, *value)
	} else {
		d.Set(key, nil)
	}
}

// SetNillableValueWithInterfaceArrayWithFunc will read the values in a nested resource using the provided function and set it on the schema
func SetNillableValueWithInterfaceArrayWithFunc[T any](d *schema.ResourceData, key string, value *T, f func(*T) []interface{}) {
	if value != nil {
		d.Set(key, f(value))
	} else {
		d.Set(key, nil)
	}
}

func SetNillableTime(d *schema.ResourceData, key string, value *time.Time) {
	var timeValue *string = nil
	if value != nil {
		timeAsString := timeutil.Strftime(value, TimeWriteFormat)
		timeValue = &timeAsString
	}
	SetNillableValue(d, key, timeValue)
}

func GetNillableValue[T any](d *schema.ResourceData, key string) *T {
	value, ok := d.GetOk(key)
	if ok {
		v := value.(T)
		return &v
	}
	return nil
}

// More info about using deprecated GetOkExists: https://github.com/hashicorp/terraform-plugin-sdk/issues/817
func GetNillableBool(d *schema.ResourceData, key string) *bool {
	value, ok := d.GetOkExists(key)
	if ok {
		v := value.(bool)
		return &v
	}
	return nil
}

func GetNillableTime(d *schema.ResourceData, key string) *time.Time {
	stringValue := GetNillableValue[string](d, key)
	if stringValue != nil {
		timeValue, err := time.Parse(TimeParseFormat, *stringValue)
		if err != nil {
			log.Printf("GetNillableTime failed for %s. Required format: %s", *stringValue, TimeParseFormat)
			return nil
		}
		return &timeValue
	}
	return nil
}

func BuildSdkListFirstElement[T interface{}](d *schema.ResourceData, key string, elementBuilder func(map[string]interface{}) *T, nilForEmpty bool) *T {
	list := d.Get(key).(*schema.Set).List()
	if len(list) > 0 {
		return elementBuilder(list[0].(map[string]interface{}))
	}
	if nilForEmpty {
		return nil
	}
	return elementBuilder(nil)
}

func BuildSdkList[T interface{}](d *schema.ResourceData, key string, elementBuilder func(map[string]interface{}) *T) *[]T {
	list := d.Get(key).(*schema.Set).List()
	if len(list) > 0 {
		sdkList := make([]T, len(list))
		for i, element := range list {
			sdkList[i] = *elementBuilder(element.(map[string]interface{}))
		}
		return &sdkList
	}
	return nil
}
