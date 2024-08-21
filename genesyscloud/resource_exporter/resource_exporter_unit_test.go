package resource_exporter

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"testing"
)

type TestAssertion struct {
	input  string
	output string
	name   string
}

// Tests the original sanitizing algorithm
func TestUnitSanitizeResourceNameOriginal(t *testing.T) {
	simpleString := "foobar"
	intString := "1234"
	underscore := "_"
	dash := "-"
	unsafeUnicode := "Ⱥ®ÊƩ"
	unsafeAscii := "#%@&"

	randNumSuffix := "_[0-9]+$"

	envVarName := "GENESYS_SANITIZER_LEGACY"
	envVarValue := "1"
	os.Setenv(envVarName, envVarValue)
	defer func() { os.Unsetenv(envVarName) }()

	assertions := [14]TestAssertion{
		{
			name:   "First character",
			input:  string(simpleString[0]),
			output: string(simpleString[0]),
		},
		{
			name:   "Safe String",
			input:  simpleString,
			output: simpleString,
		},
		{
			name:   "Single Integer",
			input:  string(intString[0]),
			output: underscore + string(intString[0]),
		},
		{
			name:   "Single Underscore",
			input:  underscore,
			output: underscore,
		},
		{
			name:   "Single Dash",
			input:  dash,
			output: underscore + dash,
		},
		{
			name:   "Single Unsafe Ascii Character",
			input:  string(unsafeAscii[0]),
			output: underscore + randNumSuffix,
		},
		{
			name:   "Single Unsafe Unicode Character",
			input:  string(unsafeUnicode[0]),
			output: underscore + randNumSuffix,
		},
		{
			name:   "String beginning with Integer",
			input:  intString + simpleString,
			output: underscore + intString + simpleString,
		},
		{
			name:   "String beginning with Underscore",
			input:  underscore + simpleString,
			output: underscore + simpleString,
		},
		{
			name:   "String beginning with Dash",
			input:  dash + simpleString,
			output: underscore + dash + simpleString,
		},
		{
			name:   "String beginning with multiple dashes",
			input:  dash + dash + dash + dash + simpleString + dash + dash + dash + dash,
			output: underscore + dash + dash + dash + dash + simpleString + dash + dash + dash + dash,
		},
		{
			name:   "String beginning with Unsafe Ascii Character",
			input:  unsafeAscii + simpleString,
			output: strings.Repeat(underscore, len(unsafeAscii)) + simpleString + randNumSuffix,
		},
		{
			name:   "String beginning with Unicode",
			input:  unsafeUnicode + simpleString,
			output: strings.Repeat(underscore, len(unsafeAscii)) + simpleString + randNumSuffix,
		},
		{
			name:   "String with everything",
			input:  simpleString + unsafeAscii + underscore + intString + dash + unsafeUnicode + simpleString,
			output: simpleString + strings.Repeat(underscore, len(unsafeAscii)) + underscore + intString + dash + strings.Repeat(underscore, len([]rune(unsafeUnicode))) + simpleString,
		},
	}

	for _, assertion := range assertions {
		sanitizer := NewSanitizerProvider()
		output := sanitizer.S.SanitizeResourceName(assertion.input)

		assertionOutputRegex := regexp.MustCompile(assertion.output)
		if !assertionOutputRegex.MatchString(output) {
			t.Errorf("%s did not sanitize correctly!\nExpected Output: %v\nActual Output: %v", assertion.name, assertion.output, output)
		}
	}
}

// Tests the optimized sanitizing algorithm
func TestUnitSanitizeResourceNamesOptimized(t *testing.T) {
	randNumSuffix := "_[0-9]+"
	metaMap := make(ResourceIDMetaMap)
	metaMap["1"] = &ResourceMeta{ObjectName: "wrapupcodemappings", BlockLabel: "wrapupcodemappings"}
	metaMap["2"] = &ResourceMeta{ObjectName: "foobar", BlockLabel: "foobar"}
	metaMap["3"] = &ResourceMeta{ObjectName: "wrapupcode$%^mappings", BlockLabel: "wrapupcode$%^mappings"}
	metaMap["4"] = &ResourceMeta{ObjectName: "wrapupcode*#@mappings", BlockLabel: "wrapupcode*#@mappings"}
	metaMap["5"] = &ResourceMeta{ObjectName: "-suuuuueeeey", BlockLabel: "-suuuuueeeey"}
	metaMap["6"] = &ResourceMeta{ObjectName: "1-2bucklemyshoe", BlockLabel: "1-2bucklemyshoe"}
	metaMap["7"] = &ResourceMeta{ObjectName: "unsafeUnicodeȺ®Here", BlockLabel: "unsafeUnicodeȺ®Here"}
	metaMap["8"] = &ResourceMeta{ObjectName: "unsafeUnicodeÊƩHere", BlockLabel: "unsafeUnicodeÊƩHere"}
	metaMap["9"] = &ResourceMeta{ObjectName: "unsafeUnicodeÊƩȺ®Here", BlockLabel: "unsafeUnicodeÊƩȺ®Here"}

	sanitizer := NewSanitizerProvider()

	sanitizer.S.Sanitize(metaMap)

	assertions := [9]TestAssertion{
		{
			input:  metaMap["1"].SanitizedBlockLabel,
			output: "wrapupcodemappings",
			name:   "actual resource name",
		},
		{
			input:  metaMap["2"].SanitizedBlockLabel,
			output: "foobar",
			name:   "any name",
		},
		{
			input:  metaMap["3"].SanitizedBlockLabel,
			output: "wrapupcode___mappings" + randNumSuffix,
			name:   "ascii chars",
		},
		{
			input:  metaMap["4"].SanitizedBlockLabel,
			output: "wrapupcode___mappings" + randNumSuffix,
			name:   "ascii chars with same structure different chars",
		},
		{
			input:  metaMap["5"].SanitizedBlockLabel,
			output: "_-suuuuueeeey",
			name:   "starting dash",
		},
		{
			input:  metaMap["6"].SanitizedBlockLabel,
			output: "_1-2bucklemyshoe",
			name:   "starting number",
		},
		{
			input:  metaMap["7"].SanitizedBlockLabel,
			output: "unsafeUnicode__Here" + randNumSuffix,
			name:   "unsafe unicode",
		},
		{
			input:  metaMap["8"].SanitizedBlockLabel,
			output: "unsafeUnicode__Here" + randNumSuffix,
			name:   "unsafe unicode matching pattern",
		},
		{
			input:  metaMap["9"].SanitizedBlockLabel,
			output: "unsafeUnicode____Here",
			name:   "unsafe unicode non-matching pattern, no added random suffix",
		},
	}

	for _, assertion := range assertions {
		assertionOutputRegex := regexp.MustCompile("^" + assertion.output + "$")
		if !assertionOutputRegex.MatchString(assertion.input) {
			t.Errorf("%s did not sanitize correctly!\nExpected Output: %v\nActual Output: %v", assertion.name, assertion.output, assertion.input)
		}
	}

}

// Tests the optimized sanitizing algorithm
func TestUnitSanitizeResourceNameOptimized(t *testing.T) {
	simpleString := "foobar"
	intString := "1234"
	underscore := "_"
	dash := "-"
	unsafeUnicode := "Ⱥ®ÊƩ"
	unsafeAscii := "#%$^@&"

	//We set the GENESYS_SANITIZER_OPTIMIZED environment variable to ensure the new optimized  is used
	envVarName := "GENESYS_SANITIZER_OPTIMIZED"
	envVarValue := "1"
	os.Setenv(envVarName, envVarValue)
	sanitizer := NewSanitizerProvider()

	//Make sure we unset the GENESYS_SANITIZER_OPTIMIZED environment variable after the test runs
	unsetEnv := func() {
		os.Unsetenv("GENESYS_SANITIZER_OPTIMIZED")
	}
	defer unsetEnv()

	assertions := [14]TestAssertion{
		{
			name:   "First character",
			input:  string(simpleString[0]),
			output: string(simpleString[0]),
		},
		{
			name:   "Safe String",
			input:  simpleString,
			output: simpleString,
		},
		{
			name:   "Single Integer",
			input:  string(intString[0]),
			output: underscore + string(intString[0]),
		},
		{
			name:   "Single Underscore",
			input:  underscore,
			output: underscore,
		},
		{
			name:   "Single Dash",
			input:  dash,
			output: underscore + dash,
		},
		{
			name:   "Single Unsafe Ascii Character",
			input:  string(unsafeAscii[0]),
			output: underscore,
		},
		{
			name:   "Single Unsafe Unicode Character",
			input:  string(unsafeUnicode[0]),
			output: underscore,
		},
		{
			name:   "String beginning with Integer",
			input:  intString + simpleString,
			output: underscore + intString + simpleString,
		},
		{
			name:   "String beginning with Underscore",
			input:  underscore + simpleString,
			output: underscore + simpleString,
		},
		{
			name:   "String beginning with Dash",
			input:  dash + simpleString,
			output: underscore + dash + simpleString,
		},
		{
			name:   "String beginning with multiple dashes",
			input:  dash + dash + dash + dash + simpleString + dash + dash + dash + dash,
			output: underscore + dash + dash + dash + dash + simpleString + dash + dash + dash + dash,
		},
		{
			name:   "String beginning with Unsafe Ascii Character",
			input:  unsafeAscii + simpleString,
			output: strings.Repeat(underscore, len(unsafeAscii)) + simpleString,
		},
		{
			name:   "String beginning with Unicode",
			input:  unsafeUnicode + simpleString,
			output: strings.Repeat(underscore, len([]rune(unsafeUnicode))) + simpleString,
		},
		{
			name:   "String with everything",
			input:  simpleString + unsafeAscii + underscore + intString + dash + unsafeUnicode + simpleString,
			output: simpleString + strings.Repeat(underscore, len(unsafeAscii)) + underscore + intString + dash + strings.Repeat(underscore, len([]rune(unsafeUnicode))) + simpleString,
		},
	}

	for _, assertion := range assertions {
		output := sanitizer.S.SanitizeResourceName(assertion.input)
		assertionOutputRegex := regexp.MustCompile("^" + assertion.output + "$")
		if !assertionOutputRegex.MatchString(output) {
			t.Errorf("%s did not sanitize correctly!\nExpected Output: %v\nActual Output: %v", assertion.name, assertion.output, output)
		}
	}
}

func TestUnitSanitizeOptimized(t *testing.T) {

	//We set the GENESYS_SANITIZER_OPTIMIZED environment variable to ensure the new optimized  is used
	envVarName := "GENESYS_SANITIZER_OPTIMIZED"
	envVarValue := "1"
	os.Setenv(envVarName, envVarValue)
	sanitizer := NewSanitizerProvider()

	//Make sure we unset the GENESYS_SANITIZER_OPTIMIZED environment variable after the test runs
	unsetEnv := func() {
		os.Unsetenv("GENESYS_SANITIZER_OPTIMIZED")
	}
	defer unsetEnv()

	meta1 := &ResourceMeta{
		BlockLabel: "Resource Name1",
	}
	meta2 := &ResourceMeta{
		BlockLabel: "+ResourceName2",
	}
	meta3 := &ResourceMeta{
		BlockLabel: "INQUEUECALL_NK_室町_InQueue",
	}
	meta4 := &ResourceMeta{
		BlockLabel: "INQUEUECALL_NK_鎌倉_InQueue",
	}

	// Create an instance of ResourceIDMetaMap and add the meta to it
	resources := ResourceIDMetaMap{
		"resource1": meta1,
		"resource2": meta2,
		"resource3": meta3,
		"resource4": meta4,
	}

	expected := map[string]string{
		"resource1": "Resource_Name1",
		"resource2": "_ResourceName2",
		"resource3": "INQUEUECALL_NK____InQueue_1977736383",
		"resource4": "INQUEUECALL_NK____InQueue_1369682474",
	}

	sanitizer.S.Sanitize(resources)

	for key, resource := range resources {
		sanitized := resource.SanitizedBlockLabel
		expectedOutput := expected[key]
		if sanitized != expectedOutput {
			t.Error(fmt.Printf("key %s did not sanitize ( %s ) as expected ( %s )", key, sanitized, expectedOutput))
		}
	}

}
