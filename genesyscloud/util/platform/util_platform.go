package platform

import (
	"reflect"
	"strings"
	"time"
)

/*
  These functions are mirrors of helper functions that are found within the platformclientv2's apiclient.go
*/

func Copy(data []byte, v interface{}) {
	dataS := string(data)

	reflect.ValueOf(v).Elem().Set(reflect.ValueOf(&dataS))
}

func Contains(source []string, containvalue string) bool {
	for _, a := range source {
		if strings.ToLower(a) == strings.ToLower(containvalue) {
			return true
		}
	}
	return false
}

// toTime is a helper function for models to prevent a static import of the "time" package duplicating dynamic imports determined by the codegen process
func ToTime(o interface{}) *time.Time {
	return o.(*time.Time)
}

func GetFieldName(t reflect.Type, field string) string {
	// Find JSON prop name
	structField, ok := t.Elem().FieldByName(field)
	if ok {
		tag := structField.Tag.Get("json")
		if tag != "" {
			tagParts := strings.Split(tag, ",")
			if len(tagParts) > 0 {
				return tagParts[0]
			}
		}
	}

	return field
}
