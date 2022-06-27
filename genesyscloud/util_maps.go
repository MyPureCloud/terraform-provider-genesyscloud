package genesyscloud

func getNullableMapValue[T any](m map[string]interface{}, key string) *T {
	value, ok := m[key]
	if ok {
		v := value.(T)
		return &v
	}
	return nil
}

func setMapValueIfNotNil[T any](m map[string]interface{}, key string, value *T) {
	if value != nil {
		m[key] = *value
	}
}
