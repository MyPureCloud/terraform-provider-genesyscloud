package genesyscloud

func getNillableMapValue[T any](m map[string]interface{}, key string) *T {
	value, ok := m[key]
	if ok {
		v := value.(T)
		return &v
	}
	return nil
}

func getNonDefaultMapValue[T comparable](m map[string]interface{}, key string) *T {
	value := getNillableMapValue[T](m, key)
	if value != nil {
		defaultValue := new(T)
		if value != defaultValue {
			return value
		}
	}
	return nil
}

func setMapValueIfNotNil[T any](m map[string]interface{}, key string, value *T) {
	if value != nil {
		m[key] = *value
	}
}
