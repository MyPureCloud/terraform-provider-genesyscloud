package genesyscloud

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

const (
	nullValue  = "null"
	trueValue  = "true"
	falseValue = "false"
	testCert1  = "MIIDazCCAlKgAwIBAgIBADANBgkqhkiG9w0BAQsFADBOMQswCQYDVQQGEwJ1czEXMBUGA1UECAwOTm9ydGggQ2Fyb2xpbmExEDAOBgNVBAoMB0dlbmVzeXMxFDASBgNVBAMMC215cHVyZWNsb3VkMCAXDTIyMDUxNzEzNDUzM1oYDzIxMjIwNDIzMTM0NTMzWjBOMQswCQYDVQQGEwJ1czEXMBUGA1UECAwOTm9ydGggQ2Fyb2xpbmExEDAOBgNVBAoMB0dlbmVzeXMxFDASBgNVBAMMC215cHVyZWNsb3VkMIIBIzANBgkqhkiG9w0BAQEFAAOCARAAMIIBCwKCAQIAuicPlCgrmmzIuu/Hh0HBqmGOvO7lLeKq4ZryZxd11XmcVE4T4mhdI+u1rgv8GBnn9JmFkXGU793l1PuUmrZuUInkuvVhvOjcl/95WzGE5++bkvQ/AhROn4onAWQIrQvpUq+xKv3vZ4z7JncqbkBRsJ1BKsCxtL3nKLlUBD2z8/KrrbKjENEDCIlhdua5KPfl/d+IwW8iOmTsLQYNsSv8ZvovwK/WwvcFsjtQIdBSdJfPguAzKiQIaihzya6dzXLFlxYsBsbA39MEcNTeOpy+b1xNEo0WCvVW0qctVV+z3qHMHqcjkikT4PUzBkeceZe5dnqfm+P1TFTk1OO8b0xmkgECAwEAAaNQME4wHQYDVR0OBBYEFCuD7HIc4V8HNEAftG5w+nFFl5JVMB8GA1UdIwQYMBaAFCuD7HIc4V8HNEAftG5w+nFFl5JVMAwGA1UdEwQFMAMBAf8wDQYJKoZIhvcNAQELBQADggECAEUmWVt01Kh1Be4U+CrI8Vdz6Hls3RJmto/x0WQUARjUO3+0SiFUxFAgRGGkFJTdtH+J93OntLsK8Av+G3U+ZNCODbRBubXqcnljbXnaeXDp4saUWuRs4G6zYFPM0rCvSz46XK6G5dyANeEJFgdO7wKkHO/eyy4PkIgjBE59DAx97sbXW877DTdvSfbmsEKiuEB0an+kdPYZHbTLdM910Y8YyeEQBkzp1Kjz3u5fwpAKFULOhsBmXYtXTReMqtWHjG4czsRZr04wHIng45WD8weMdw1UsCpr8fJ4CYMJsKgwJkKOc8fw6Fmj7mqrXIlUMMpeyDNpqEMaNIryiG/UsZma"
	testCert2  = "MIIDazCCAlKgAwIBAgIBADANBgkqhkiG9w0BAQsFADBOMQswCQYDVQQGEwJ1czEXMBUGA1UECAwOTm9ydGggQ2Fyb2xpbmExEDAOBgNVBAoMB0dlbmVzeXMxFDASBgNVBAMMC215cHVyZWNsb3VkMCAXDTIyMDUxNzEzNDY0N1oYDzIxMjIwNDIzMTM0NjQ3WjBOMQswCQYDVQQGEwJ1czEXMBUGA1UECAwOTm9ydGggQ2Fyb2xpbmExEDAOBgNVBAoMB0dlbmVzeXMxFDASBgNVBAMMC215cHVyZWNsb3VkMIIBIzANBgkqhkiG9w0BAQEFAAOCARAAMIIBCwKCAQIAzWc4XQthXrGexwsH2urKc1dFPhZMoWhUVjXrb1bc1IdCH63KklnhYiBAB2YakRJVSzoat5iY0X2kNjSIyCtHCxPycpplP4P6BfIEM9jm0s8NmYW3S/8JZW1MiNs/2XTibfyoXmQiHh76BzKCDgniulj2qOxpNHi5M1Az0QxV+GSgVE+mcPA6041idt7n1HpG3gQ7/MrZEd5OdBhyVUa6JPDyTAF7UE9P9v7mIbGoe6R7Y9qQEIbJ8ihoSM+w65fhyDafl9dWjfLmqkI65cYCJ82cGqyseeiHYOXgyfkcC1njrLr5g92DHnOVqVoHZCTzwV+kciyAntuQqyJtHGCGnskCAwEAAaNQME4wHQYDVR0OBBYEFDNbxsJcQMKJVSIHT/3BM1Osb+JOMB8GA1UdIwQYMBaAFDNbxsJcQMKJVSIHT/3BM1Osb+JOMAwGA1UdEwQFMAMBAf8wDQYJKoZIhvcNAQELBQADggECAGuzz8i3w3YrFGeGgxRzwEWUKiH53Sf4w7KIxGeK6oW2BOnhXMYJfuqIAiGaAVQ3uHbTcKwByHLK9/2oWmQAsYsbA3wZpcZXyXk84iCc3aqYkWjeUl0A5wECjNIKkFvS56DCtENLMlc2VI8NGzPoFMaC7Z3nMOlogqsf6KNNydUMgqyosLQqYoRdDbBMXShbn7fvibK4jzhYxuoXCyTwKDg/lr69i5zsVNBMjTu8W3DnmBPbTVBQ9Kd9/nAJoXCbHfx1QW4UEx3mLFDVNhRRdGqran7DIEjCo8BcGilXvHCVCAKwXF1MyqiyLEm8/W7FYzdBBkkVnxOBhMIVjlPGpwLS"

	dataSourceTestType = "data_source"
	resourceTestType   = "resource"
	testObjectIdPrefix = "terraform_test_"
)

// Verify default division is home division
func testDefaultHomeDivision(resource string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		homeDivID, err := getHomeDivisionID()
		if err != nil {
			return fmt.Errorf("Failed to query home division: %v", err)
		}

		r := state.RootModule().Resources[resource]
		if r == nil {
			return fmt.Errorf("%s not found in state", resource)
		}

		a := r.Primary.Attributes

		if a["division_id"] != homeDivID {
			return fmt.Errorf("expected division to be home division %s", homeDivID)
		}

		return nil
	}
}

func generateStringArray(vals ...string) string {
	return fmt.Sprintf("[%s]", strings.Join(vals, ","))
}

func validateStringInArray(resourceName string, attrName string, value string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		resourceState, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Failed to find resourceState %s in state", resourceName)
		}
		resourceID := resourceState.Primary.ID

		numAttr, ok := resourceState.Primary.Attributes[attrName+".#"]
		if !ok {
			return fmt.Errorf("No %s found for %s in state", attrName, resourceID)
		}

		numValues, _ := strconv.Atoi(numAttr)
		for i := 0; i < numValues; i++ {
			if resourceState.Primary.Attributes[attrName+"."+strconv.Itoa(i)] == value {
				// Found value
				return nil
			}
		}

		return fmt.Errorf("%s %s not found for group %s in state", attrName, value, resourceID)
	}
}

func strArrayEquals(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}

func validateValueInJsonAttr(resourceName string, attrName string, jsonProp string, jsonValue string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		resourceState, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Failed to find resource %s in state", resourceName)
		}
		resourceID := resourceState.Primary.ID

		jsonAttr, ok := resourceState.Primary.Attributes[attrName]
		if !ok {
			return fmt.Errorf("No %s found for %s in state", attrName, resourceID)
		}

		var jsonMap map[string]interface{}
		if err := json.Unmarshal([]byte(jsonAttr), &jsonMap); err != nil {
			return fmt.Errorf("Error parsing JSON for %s in state: %v", resourceID, err)
		}

		propPath := strings.Split(jsonProp, ".")
		if val, ok := jsonMap[propPath[0]]; ok {
			for i := 1; i < len(propPath); i++ {
				switch obj := val.(type) {
				case map[string]interface{}:
					val = obj[propPath[i]]
				case []interface{}:
					val = obj
				default:
					return fmt.Errorf("JSON property %s not found for %s in state", jsonProp, resourceID)
				}
			}
			if arr, ok := val.([]interface{}); ok {
				// Property is an array. Check if string value exists in array.
				if stringInSlice(jsonValue, interfaceListToStrings(arr)) {
					return nil
				}
				return fmt.Errorf("JSON array property for resourceState %s.%s does not contain expected %s", resourceName, jsonProp, jsonValue)
			} else {
				strVal := interfaceToString(val)
				if strVal != jsonValue {
					return fmt.Errorf("JSON property for resource %s %s=%s does not match expected %s", resourceName, jsonProp, strVal, jsonValue)
				}
			}
		} else {
			return fmt.Errorf("JSON property %s not found for %s in state", jsonProp, resourceID)
		}
		return nil
	}
}

func validateValueInJsonPropertiesAttr(resourceName string, attrName string, jsonProp string, jsonValue string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		resourceState, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Failed to find resourceState %s in state", resourceName)
		}
		resourceID := resourceState.Primary.ID

		jsonAttr, ok := resourceState.Primary.Attributes[attrName]
		if !ok {
			return fmt.Errorf("No %s found for %s in state", attrName, resourceID)
		}

		var jsonMap map[string]interface{}
		if err := json.Unmarshal([]byte(jsonAttr), &jsonMap); err != nil {
			return fmt.Errorf("Error parsing JSON for %s in state: %v", resourceID, err)
		}

		propPath := strings.Split(jsonProp, ".")
		if val, ok := jsonMap[propPath[0]]; ok {
			for i := 1; i < len(propPath); i++ {
				switch obj := val.(type) {
				case map[string]interface{}:
					val = obj[propPath[i]]
				case []interface{}:
					val = obj
				default:
					return fmt.Errorf("JSON property %s not found for %s in state", jsonProp, resourceID)
				}
			}

			valInstance := val.(map[string]interface{})["value"].(map[string]interface{})["instance"]
			if valInstanceString, ok := valInstance.(string); ok {
				if valInstanceString != jsonValue {
					return fmt.Errorf("JSON property for resource %s %s=%s does not match expected %s", resourceName, jsonProp, valInstanceString, jsonValue)
				}
			} else if valInstanceFloat, ok := valInstance.(float64); ok {
				intValue, err := strconv.Atoi(jsonValue)
				if err != nil {
					return err
				}
				if int(valInstanceFloat) != intValue {
					return fmt.Errorf("JSON property for resource %s %s=%v does not match expected %v", resourceName, jsonProp, valInstanceFloat, jsonValue)
				}
			} else if valInstanceBool, ok := valInstance.(bool); ok {
				boolValue, err := strconv.ParseBool(jsonValue)
				if err != nil {
					return err
				}
				if valInstanceBool != boolValue {
					return fmt.Errorf("JSON property for resource %s %s=%v does not match expected %v", resourceName, jsonProp, valInstanceBool, jsonValue)
				}
			} else if valInstanceSlice, ok := valInstance.([]interface{}); ok {
				if _, ok := valInstanceSlice[0].(float64); ok {
					ints := make([]string, 0)
					for _, i := range valInstanceSlice {
						ints = append(ints, strconv.Itoa(int(i.(float64))))
					}
					intsJoined := strings.Join(ints, ",")

					if intsJoined != jsonValue {
						return fmt.Errorf("JSON property for resource %s %s=%s does not match expected %s", resourceName, jsonProp, intsJoined, jsonValue)
					}
				} else if _, ok := valInstanceSlice[0].(string); ok {
					strs := make([]string, 0)
					for _, s := range valInstanceSlice {
						strs = append(strs, s.(string))
					}
					strsJoined := strings.Join(strs, ",")

					if strsJoined != jsonValue {
						return fmt.Errorf("JSON property for resource %s %s=%s does not match expected %s", resourceName, jsonProp, strsJoined, jsonValue)
					}
				}
			}
		} else {
			return fmt.Errorf("JSON property %s not found for %s in state", jsonProp, resourceID)
		}
		return nil
	}
}

func generateJsonEncodedProperties(properties ...string) string {
	return fmt.Sprintf(`jsonencode({
		%s
	})
	`, strings.Join(properties, "\n"))
}

func generateJsonProperty(propName string, propValue string) string {
	return fmt.Sprintf(`"%s" = %s`, propName, propValue)
}

func generateJsonArrayProperty(propName string, propValues ...string) string {
	return fmt.Sprintf(`"%s" = [%s]`, propName, strings.Join(propValues, ", "))
}

func generateJsonObject(properties ...string) string {
	return fmt.Sprintf(`{
		%s
	}`, strings.Join(properties, "\n"))
}

func generateMapProperty(propName string, propValue string) string {
	return fmt.Sprintf(`%s = %s`, propName, propValue)
}

func generateMapAttr(name string, properties ...string) string {
	return fmt.Sprintf(`%s = {
		%s
	}
	`, name, strings.Join(properties, "\n"))
}

func randString(length int) string {
	rand.Seed(time.Now().UnixNano())

	letters := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

	s := make([]rune, length)
	for i := range s {
		s[i] = letters[rand.Intn(len(letters))]
	}

	return string(s)
}

func generateTestSteps(testType string, resourceName string, testCaseName string, idPrefix string, checkFuncs []resource.TestCheckFunc) []resource.TestStep {
	var testSteps []resource.TestStep

	testCasePath := filepath.Join("..", "test", "data", testType, resourceName, testCaseName)
	testCaseFiles, _ := os.ReadDir(testCasePath)
	checkFuncIndex := 0
	for _, testCaseFile := range testCaseFiles {
		if !testCaseFile.IsDir() && strings.HasSuffix(testCaseFile.Name(), ".tf") {
			testCaseResource, _ := os.ReadFile(filepath.Join(testCasePath, testCaseFile.Name()))
			config := strings.ReplaceAll(string(testCaseResource), "-TEST-CASE-", testCaseName)
			var checkFunc resource.TestCheckFunc = nil
			if checkFuncs != nil && checkFuncIndex < len(checkFuncs) {
				checkFunc = checkFuncs[checkFuncIndex]
			}
			testSteps = append(testSteps, resource.TestStep{Config: config, Check: checkFunc})
			checkFuncIndex++
		}
	}
	log.Printf("Generated %d test steps for testcase: %s", len(testSteps), testCasePath)

	testSteps = append(testSteps, resource.TestStep{
		ResourceName:      resourceName + "." + idPrefix + testCaseName,
		ImportState:       true,
		ImportStateVerify: true,
	})

	return testSteps
}
