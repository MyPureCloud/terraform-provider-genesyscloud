package resource_exporter

import (
	"context"
	"errors"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/stretchr/testify/assert"
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
func TestUnitSanitizeResourceOriginal(t *testing.T) {
	randNumSuffixFn := func(amount string) string {
		return fmt.Sprintf(`[a-z0-9_]{%s}`, amount)
	}

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
			output: "wrapupcode___mappings" + randNumSuffixFn("0,21"),
			name:   "ascii chars",
		},
		{
			input:  metaMap["4"].BlockLabel,
			output: "wrapupcode___mappings" + randNumSuffixFn("0,21"),
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

	// assert no two block labels are the same
	labelsOnlyAppearOnceInSanitizedMap(t, metaMap)
}

// Tests the optimized sanitizing algorithm
func TestUnitSanitizeResourceLabelOriginal(t *testing.T) {
	simpleString := "foobar"
	intString := "1234"
	underscore := "_"
	dash := "-"
	unsafeUnicode := "Ⱥ®ÊƩ"
	unsafeAscii := "#%$^@&"
	unsafeUnicodeTransliteration := "A_r_ESH"

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

func TestUnitIdsForResourceType(t *testing.T) {
	filter := []string{"t::u1", "t::u2", "t::u1", "t_roles::r1", "q::q1", "t::"}
	assert.Equal(t, []string{"u1", "u2"}, IdsForResourceType("t", filter))
	assert.Empty(t, IdsForResourceType("missing", filter))
}

func TestUnitGetResourcesByIds(t *testing.T) {
	res, diags := GetResourcesByIds(context.Background(), []string{"a", "gone", "b"}, func(_ context.Context, id string) (*ResourceMeta, diag.Diagnostics) {
		if id == "gone" {
			return nil, nil
		}
		return &ResourceMeta{BlockLabel: id}, nil
	})
	assert.False(t, diags.HasError())
	assert.Len(t, res, 2)

	res, diags = GetResourcesByIds(context.Background(), []string{"a", "bad"}, func(_ context.Context, id string) (*ResourceMeta, diag.Diagnostics) {
		if id == "bad" {
			return nil, diag.FromErr(errors.New("boom"))
		}
		return &ResourceMeta{BlockLabel: id}, nil
	})
	assert.True(t, diags.HasError())
	assert.Nil(t, res)
}

func TestUnitLoadSanitizedResourceMapByIds(t *testing.T) {
	const rt = "t"
	byId := func(result ResourceIDMetaMap, resourceType string, filters []string) ResourceIDMetaMap {
		out := make(ResourceIDMetaMap)
		for _, id := range IdsForResourceType(resourceType, filters) {
			if v, ok := result[id]; ok {
				out[id] = v
			}
		}
		return out
	}
	newExporter := func(listCalls, idCalls *int) *ResourceExporter {
		all := func() ResourceIDMetaMap {
			return ResourceIDMetaMap{"1": {BlockLabel: "one"}, "2": {BlockLabel: "two"}, "3": {BlockLabel: "three"}}
		}
		return &ResourceExporter{
			FilterResource: byId,
			GetResourcesFunc: func(context.Context) (ResourceIDMetaMap, diag.Diagnostics) {
				*listCalls++
				return all(), nil
			},
			GetResourcesByIdsFunc: func(_ context.Context, ids []string) (ResourceIDMetaMap, diag.Diagnostics) {
				*idCalls++
				return byId(all(), rt, []string{rt + "::" + ids[0], rt + "::" + ids[len(ids)-1]}), nil
			},
		}
	}
	filter := []string{rt + "::1", rt + "::3"}

	var listed, fetched int
	byIds := newExporter(&listed, &fetched)
	byIds.FilterByID = true
	assert.False(t, byIds.LoadSanitizedResourceMap(context.Background(), rt, filter).HasError())
	assert.Equal(t, []int{0, 1}, []int{listed, fetched}, "must fetch by ID without listing")

	var listed2, fetched2 int
	listing := newExporter(&listed2, &fetched2)
	assert.False(t, listing.LoadSanitizedResourceMap(context.Background(), rt, filter).HasError())
	assert.Equal(t, []int{1, 0}, []int{listed2, fetched2}, "regex/name filters must keep listing")

	a, b := byIds.GetSanitizedResourceMap(), listing.GetSanitizedResourceMap()
	assert.Len(t, a, 2)
	for id, meta := range b {
		assert.Equal(t, meta.BlockLabel, a[id].BlockLabel)
	}
}
