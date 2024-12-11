package util

import (
	"hash/fnv"
	"regexp"
	"strconv"
	"strings"

	"github.com/google/uuid"
)

var matchFirstCap = regexp.MustCompile("(.)([A-Z][a-z]+)")
var matchAllCap = regexp.MustCompile("([a-z0-9])([A-Z])")
var matchUnderscore = regexp.MustCompile("_")

func ToSnakeCase(str string) string {
	snake := matchFirstCap.ReplaceAllString(str, "${1}_${2}")
	snake = matchAllCap.ReplaceAllString(snake, "${1}_${2}")
	return strings.ToLower(snake)
}

func ToCamelCase(str string) string {
	terms := matchUnderscore.Split(str, -1)
	camel := ""
	for i, term := range terms {
		if i == 0 {
			camel += term
		} else {
			camel += strings.Title(term)
		}
	}
	return camel
}

func StringExists(target string, slice []string) bool {
	for _, str := range slice {
		if str == target {
			return true
		}
	}
	return false
}

func GetUniqueString() string {
	hasher := fnv.New32()
	hasher.Write([]byte(uuid.NewString()))
	return strconv.FormatUint(uint64(hasher.Sum32()), 10)
}

func StringOrNil(s *string) string {
	if s == nil {
		return "nil"
	}
	return *s
}
