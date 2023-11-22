package stringmap

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func GetNillableValue[T any](m map[string]interface{}, key string) *T {
	value, ok := m[key]
	if ok {
		v := value.(T)
		return &v
	}
	return nil
}

func GetNonDefaultValue[T comparable](m map[string]interface{}, key string) *T {
	value := GetNillableValue[T](m, key)
	if value != nil {
		defaultValue := new(T)
		if *value != *defaultValue {
			return value
		}
	}
	return nil
}

func SetValueIfNotNil[T any](m map[string]interface{}, key string, value *T) {
	if value != nil {
		m[key] = *value
	}
}

func BuildSdkStringList(m map[string]interface{}, key string) *[]string {
	return BuildSdkList[string](m, key, nil)
}

func BuildSdkListFirstElement[T interface{}](m map[string]interface{}, key string, elementBuilder func(map[string]interface{}) *T, nilForEmpty bool) *T {
	list := m[key].(*schema.Set).List()
	if len(list) > 0 {
		return elementBuilder(list[0].(map[string]interface{}))
	}
	if nilForEmpty {
		return nil
	}
	return elementBuilder(nil)
}

func BuildSdkList[T interface{}](m map[string]interface{}, key string, elementBuilder func(map[string]interface{}) *T) *[]T {
	child := m[key]
	if child != nil {
		list := child.(*schema.Set).List()
		sdkList := make([]T, len(list))
		for i, element := range list {
			switch element.(type) {
			case T:
				sdkList[i] = element.(T)
			case map[string]interface{}:
				sdkList[i] = *elementBuilder(element.(map[string]interface{}))
			}
		}
		return &sdkList
	}
	return nil
}

func MergeMaps[T, U comparable](m1, m2 map[T][]U) map[T][]U {
	result := make(map[T][]U)

	for key, value := range m1 {
		result[key] = value
	}

	for key, value := range m2 {
		result[key] = value
	}

	return result
}

func MergeSingularMaps[T, U comparable](m1, m2 map[T]U) map[T]U {
	result := make(map[T]U)

	for key, value := range m1 {
		result[key] = value
	}

	for key, value := range m2 {
		result[key] = value
	}

	return result
}
