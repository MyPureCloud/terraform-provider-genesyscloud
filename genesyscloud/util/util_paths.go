package util

import (
	"fmt"
	"net/url"
)

// GetQueryParamValueFromUri takes a url and a query parameter key, and returns the value assigned to that parameter.
// This function should not be used if the value associated with the param could be an array of string values.
func GetQueryParamValueFromUri(uri, param string) (string, error) {
	var value string
	u, err := url.Parse(uri)
	if err != nil {
		return "", fmt.Errorf("failed to parse url %s: %v", uri, err)
	}
	m, err := url.ParseQuery(u.RawQuery)
	if err != nil {
		return "", fmt.Errorf("failed to parse query parameters from url %s: %v", uri, err)
	}
	if paramSlice, ok := m[param]; ok && len(paramSlice) > 0 {
		value = paramSlice[0]
	}
	return value, nil
}
