package genesyscloud

import (
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/leekchan/timeutil"
)

func setNillableValue[T any](d *schema.ResourceData, key string, value *T) {
	if value != nil {
		d.Set(key, *value)
	} else {
		d.Set(key, nil)
	}
}

func setNillableTimeValue(d *schema.ResourceData, key string, value *time.Time) {
	var timeValue *string = nil
	if value != nil {
		timeAsString := timeutil.Strftime(value, "%Y-%m-%dT%H:%M:%S.%f")
		timeValue = &timeAsString
	}
	setNillableValue(d, key, timeValue)
}

func getNillableValue[T any](d *schema.ResourceData, key string) *T {
	value, ok := d.GetOk(key)
	if ok {
		v := value.(T)
		return &v
	}
	return nil
}

// More info about using deprecated GetOkExists: https://github.com/hashicorp/terraform-plugin-sdk/issues/817
func getNillableBool(d *schema.ResourceData, key string) *bool {
	value, ok := d.GetOkExists(key)
	if ok {
		v := value.(bool)
		return &v
	}
	return nil
}

func getNillableTime(d *schema.ResourceData, key string) *time.Time {
	stringValue := getNillableValue[string](d, key)
	if stringValue != nil {
		timeValue, err := time.Parse("2006-01-02T15:04:05.000000", *stringValue)
		if err != nil {
			return nil
		}
		return &timeValue
	}
	return nil
}
