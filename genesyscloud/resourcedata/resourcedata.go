package resourcedata

import (
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/leekchan/timeutil"
)

func SetNillableValue[T any](d *schema.ResourceData, key string, value *T) {
	if value != nil {
		d.Set(key, *value)
	} else {
		d.Set(key, nil)
	}
}

func SetNillableTime(d *schema.ResourceData, key string, value *time.Time) {
	var timeValue *string = nil
	if value != nil {
		timeAsString := timeutil.Strftime(value, "%Y-%m-%dT%H:%M:%S.%f")
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
		timeValue, err := time.Parse("2006-01-02T15:04:05.000000", *stringValue)
		if err != nil {
			return nil
		}
		return &timeValue
	}
	return nil
}

func BuildSdkListFirstElement[T interface{}](d *schema.ResourceData, attrName string, elementBuilder func(map[string]interface{}) *T) *T {
	list := d.Get(attrName).(*schema.Set).List()
	if len(list) > 0 {
		return elementBuilder(list[0].(map[string]interface{}))
	}
	return elementBuilder(nil)
}

func BuildSdkList[T interface{}](d *schema.ResourceData, attrName string, elementBuilder func(map[string]interface{}) *T) *[]T {
	list := d.Get(attrName).(*schema.Set).List()
	if len(list) > 0 {
		sdkList := make([]T, len(list))
		for i, element := range list {
			sdkList[i] = *elementBuilder(element.(map[string]interface{}))
		}
		return &sdkList
	}
	return nil
}
