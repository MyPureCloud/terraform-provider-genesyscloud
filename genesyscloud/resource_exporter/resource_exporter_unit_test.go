package resource_exporter

import (
	"os"
	"regexp"
	"strings"
	lists "terraform-provider-genesyscloud/genesyscloud/util/lists"
	"testing"
)

type TestAssertion struct {
	input  string
	output string
	name   string
}

// Tests the original sanitizing algorithm
func TestUnitSanitizeResourceOriginal(t *testing.T) {
	randNumSuffix := "_[0-9]+"
	metaMap := make(ResourceIDMetaMap)
	metaMap["1"] = &ResourceMeta{BlockLabel: "wrapupcodemappings"}
	metaMap["2"] = &ResourceMeta{BlockLabel: "foobar"}
	metaMap["3"] = &ResourceMeta{BlockLabel: "wrapupcode$%^mappings"}
	metaMap["4"] = &ResourceMeta{BlockLabel: "wrapupcode*#@mappings"}
	metaMap["5"] = &ResourceMeta{BlockLabel: "-suuuuueeeey"}
	metaMap["6"] = &ResourceMeta{BlockLabel: "1-2bucklemyshoe"}
	metaMap["7"] = &ResourceMeta{BlockLabel: "unsafeUnicodeȺ®Here"}
	metaMap["8"] = &ResourceMeta{BlockLabel: "unsafeUnicodeÊƩHere"}
	metaMap["9"] = &ResourceMeta{BlockLabel: "unsafeUnicodeÊƩȺ®Here"}

	sanitizer := NewSanitizerProvider()

	sanitizer.S.Sanitize(metaMap)

	assertions := [9]TestAssertion{
		{
			input:  metaMap["1"].BlockLabel,
			output: "wrapupcodemappings",
			name:   "actual resource label",
		},
		{
			input:  metaMap["2"].BlockLabel,
			output: "foobar",
			name:   "any label",
		},
		{
			input:  metaMap["3"].BlockLabel,
			output: "wrapupcode___mappings" + randNumSuffix,
			name:   "ascii chars",
		},
		{
			input:  metaMap["4"].BlockLabel,
			output: "wrapupcode___mappings" + randNumSuffix,
			name:   "ascii chars with same structure different chars",
		},
		{
			input:  metaMap["5"].BlockLabel,
			output: "_-suuuuueeeey",
			name:   "starting dash",
		},
		{
			input:  metaMap["6"].BlockLabel,
			output: "_1-2bucklemyshoe",
			name:   "starting number",
		},
		{
			input:  metaMap["7"].BlockLabel,
			output: "unsafeUnicode__Here" + randNumSuffix,
			name:   "unsafe unicode",
		},
		{
			input:  metaMap["8"].BlockLabel,
			output: "unsafeUnicode__Here" + randNumSuffix,
			name:   "unsafe unicode matching pattern",
		},
		{
			input:  metaMap["9"].BlockLabel,
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
func TestUnitSanitizeResourceLabelOriginal(t *testing.T) {
	simpleString := "foobar"
	intString := "1234"
	underscore := "_"
	dash := "-"
	unsafeUnicode := "Ⱥ®ÊƩ"
	unsafeAscii := "#%$^@&"

	sanitizer := NewSanitizerProvider()

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
		output := sanitizer.S.SanitizeResourceBlockLabel(assertion.input)
		assertionOutputRegex := regexp.MustCompile("^" + assertion.output + "$")
		if !assertionOutputRegex.MatchString(output) {
			t.Errorf("%s did not sanitize correctly!\nExpected Output: %v\nActual Output: %v", assertion.name, assertion.output, output)
		}
	}
}

// Tests the optimized sanitizing algorithm
func TestUnitSanitizeResourceOptimized(t *testing.T) {
	randNumSuffix := "_[a-f0-9]+"
	metaMap := make(ResourceIDMetaMap)
	metaMap["1"] = &ResourceMeta{BlockLabel: "wrapupcodemappings"}
	metaMap["2"] = &ResourceMeta{BlockLabel: "foobar"}
	metaMap["3"] = &ResourceMeta{BlockLabel: "wrapupcode$%^mappings"}
	metaMap["4"] = &ResourceMeta{BlockLabel: "wrapupcode*#@mappings"}
	metaMap["5"] = &ResourceMeta{BlockLabel: "-suuuuueeeey"}
	metaMap["6"] = &ResourceMeta{BlockLabel: "1-2bucklemyshoe"}
	metaMap["7"] = &ResourceMeta{BlockLabel: "unsafeUnicodeȺ®Here"}
	metaMap["8"] = &ResourceMeta{BlockLabel: "unsafeUnicodeÊƩHere"}
	metaMap["9"] = &ResourceMeta{BlockLabel: "unsafeUnicodeÊƩȺ®Here"}

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

	sanitizer.S.Sanitize(metaMap)

	assertions := [9]TestAssertion{
		{
			input:  metaMap["1"].BlockLabel,
			output: "wrapupcodemappings",
			name:   "actual resource label",
		},
		{
			input:  metaMap["2"].BlockLabel,
			output: "foobar",
			name:   "any label",
		},
		{
			input:  strings.Join(lists.RemoveStringFromSlice("wrapupcode___mappings", []string{metaMap["3"].BlockLabel, metaMap["4"].BlockLabel}), ","),
			output: "wrapupcode___mappings" + randNumSuffix,
			name:   "ascii chars",
		},
		{
			input:  metaMap["5"].BlockLabel,
			output: "_-suuuuueeeey",
			name:   "starting dash",
		},
		{
			input:  metaMap["6"].BlockLabel,
			output: "_1-2bucklemyshoe",
			name:   "starting number",
		},
		{
			input:  metaMap["7"].BlockLabel,
			output: "unsafeUnicodeA_r_Here",
			name:   "unsafe unicode",
		},
		{
			input:  metaMap["8"].BlockLabel,
			output: "unsafeUnicodeESHHere",
			name:   "unsafe unicode matching pattern",
		},
		{
			input:  metaMap["9"].BlockLabel,
			output: "unsafeUnicodeESHA_r_Here",
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
func TestUnitSanitizeResourceTransliterationOptimized(t *testing.T) {
	simpleString := "foobar"
	intString := "1234"
	underscore := "_"
	dash := "-"
	unsafeUnicode := "Ⱥ®ÊƩ"
	unsafeUnicodeTransliteration := "A_r_ESH"
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
			output: "E",
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
			output: unsafeUnicodeTransliteration + simpleString,
		},
		{
			name:   "String with everything",
			input:  simpleString + unsafeAscii + underscore + intString + dash + unsafeUnicode + simpleString,
			output: simpleString + strings.Repeat(underscore, len(unsafeAscii)) + underscore + intString + dash + unsafeUnicodeTransliteration + simpleString,
		},
	}

	for _, assertion := range assertions {
		output := sanitizer.S.SanitizeResourceBlockLabel(assertion.input)
		assertionOutputRegex := regexp.MustCompile("^" + assertion.output + "$")
		if !assertionOutputRegex.MatchString(output) {
			t.Errorf("%s did not sanitize correctly!\nExpected Output: %v\nActual Output: %v", assertion.name, assertion.output, output)
		}
	}
}
