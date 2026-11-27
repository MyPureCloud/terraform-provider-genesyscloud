package responsemanagement_response

import (
	"bytes"
	"log"
	"os"
	"strings"
	"testing"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v200/platformclientv2"
)

func stringSet(values ...string) *schema.Set {
	items := make([]interface{}, len(values))
	for i, v := range values {
		items[i] = v
	}
	return schema.NewSet(schema.HashString, items)
}

func assetUsagesById(assets *[]platformclientv2.Rmsassetaddressableref) map[string]*string {
	result := map[string]*string{}
	for _, a := range *assets {
		result[*a.Id] = a.AssetUsage
	}
	return result
}

func TestUnitBuildResponseAssetsInlineOnlyKeepsExistingPayload(t *testing.T) {
	assets := buildResponseAssets(stringSet("img-1", "img-2"), stringSet())

	usages := assetUsagesById(assets)
	if len(usages) != 2 {
		t.Fatalf("expected 2 assets, got %d", len(usages))
	}
	for id, usage := range usages {
		if usage != nil {
			t.Errorf("inline asset %s must be sent without assetUsage, got %q", id, *usage)
		}
	}
}

func TestUnitBuildResponseAssetsEmptyReturnsSameAsBefore(t *testing.T) {
	assets := buildResponseAssets(stringSet(), stringSet())
	if assets == nil || len(*assets) != 0 {
		t.Fatalf("expected an empty assets list so updates can clear assets, got %v", assets)
	}
}

func TestUnitBuildResponseAssetsMixed(t *testing.T) {
	assets := buildResponseAssets(stringSet("img-1"), stringSet("pdf-1", "pdf-2"))

	usages := assetUsagesById(assets)
	if len(usages) != 3 {
		t.Fatalf("expected 3 assets, got %d", len(usages))
	}
	if usages["img-1"] != nil {
		t.Errorf("img-1 must be sent without assetUsage, got %q", *usages["img-1"])
	}
	for _, id := range []string{"pdf-1", "pdf-2"} {
		if usages[id] == nil || *usages[id] != assetUsageFileAttachments {
			t.Errorf("%s must be sent with assetUsage %q, got %v", id, assetUsageFileAttachments, usages[id])
		}
	}
}

func TestUnitBuildResponseAssetsFileAttachmentsOnly(t *testing.T) {
	assets := buildResponseAssets(stringSet(), stringSet("pdf-1"))

	usages := assetUsagesById(assets)
	if len(usages) != 1 || usages["pdf-1"] == nil || *usages["pdf-1"] != assetUsageFileAttachments {
		t.Fatalf("expected only pdf-1 with FileAttachments, got %v", usages)
	}
}

func ctyStringSet(values ...cty.Value) cty.Value {
	if len(values) == 0 {
		return cty.SetValEmpty(cty.String)
	}
	return cty.SetVal(values)
}

func TestUnitValidateResponseAssetIds(t *testing.T) {
	img, pdf := cty.StringVal("img-1"), cty.StringVal("pdf-1")

	if err := validateResponseAssetIds(ctyStringSet(img), ctyStringSet(pdf)); err != nil {
		t.Errorf("expected no error for disjoint sets, got %v", err)
	}
	if err := validateResponseAssetIds(cty.NullVal(cty.Set(cty.String)), ctyStringSet(pdf)); err != nil {
		t.Errorf("expected no error when asset_ids is not set, got %v", err)
	}

	err := validateResponseAssetIds(
		ctyStringSet(cty.StringVal("shared-b"), img, cty.StringVal("shared-a")),
		ctyStringSet(cty.StringVal("shared-a"), cty.StringVal("shared-b"), pdf),
	)
	if err == nil {
		t.Fatal("expected an error when an asset is in both sets")
	}
	if !strings.Contains(err.Error(), "shared-a, shared-b") {
		t.Errorf("expected sorted overlapping IDs in error, got %q", err.Error())
	}
}

func TestUnitValidateResponseAssetIdsSkipsUnknownValues(t *testing.T) {
	// Assets created in the same apply have unknown IDs at plan time; two unknowns must not count as a match.
	unknownInline := cty.SetVal([]cty.Value{cty.UnknownVal(cty.String)})
	unknownAttachments := cty.SetVal([]cty.Value{cty.UnknownVal(cty.String), cty.StringVal("pdf-1")})

	if err := validateResponseAssetIds(unknownInline, unknownAttachments); err != nil {
		t.Errorf("expected unknown IDs to be skipped, got %v", err)
	}
	if err := validateResponseAssetIds(cty.UnknownVal(cty.Set(cty.String)), ctyStringSet(cty.StringVal("pdf-1"))); err != nil {
		t.Errorf("expected a wholly unknown set to be skipped, got %v", err)
	}
}

func TestUnitSplitResponseAssets(t *testing.T) {
	assets := []platformclientv2.Rmsassetaddressableref{
		{Id: platformclientv2.String("img-1")},
		{Id: platformclientv2.String("img-2"), AssetUsage: platformclientv2.String("")},
		{Id: platformclientv2.String("pdf-1"), AssetUsage: platformclientv2.String(assetUsageFileAttachments)},
		{AssetUsage: platformclientv2.String(assetUsageFileAttachments)},
	}

	inline, files := splitResponseAssets(&assets)

	if !schema.NewSet(schema.HashString, inline).Equal(stringSet("img-1", "img-2")) {
		t.Errorf("unexpected inline assets: %v", inline)
	}
	if !schema.NewSet(schema.HashString, files).Equal(stringSet("pdf-1")) {
		t.Errorf("unexpected file attachment assets: %v", files)
	}
}

func TestUnitSplitResponseAssetsNil(t *testing.T) {
	inline, files := splitResponseAssets(nil)
	if inline != nil || files != nil {
		t.Errorf("expected nil slices for nil assets, got %v / %v", inline, files)
	}
}

func TestUnitSplitResponseAssetsUnknownUsageFallsBackToInlineAndWarns(t *testing.T) {
	var logs bytes.Buffer
	log.SetOutput(&logs)
	defer log.SetOutput(os.Stderr)

	assets := []platformclientv2.Rmsassetaddressableref{
		{Id: platformclientv2.String("future-1"), AssetUsage: platformclientv2.String("SomeFutureUsage")},
	}

	inline, files := splitResponseAssets(&assets)

	if len(files) != 0 {
		t.Errorf("unknown usage must not be treated as a file attachment, got %v", files)
	}
	if len(inline) != 1 || inline[0] != "future-1" {
		t.Errorf("unknown usage asset must be kept in asset_ids, got %v", inline)
	}
	want := `[WARN] Response asset future-1 has unrecognised assetUsage "SomeFutureUsage"`
	if !strings.Contains(logs.String(), want) {
		t.Errorf("expected log to contain %q, got %q", want, logs.String())
	}
}

func TestUnitSplitResponseAssetsKnownUsagesDoNotWarn(t *testing.T) {
	var logs bytes.Buffer
	log.SetOutput(&logs)
	defer log.SetOutput(os.Stderr)

	assets := []platformclientv2.Rmsassetaddressableref{
		{Id: platformclientv2.String("img-1")},
		{Id: platformclientv2.String("pdf-1"), AssetUsage: platformclientv2.String(assetUsageFileAttachments)},
	}
	splitResponseAssets(&assets)

	if strings.Contains(logs.String(), "unrecognised assetUsage") {
		t.Errorf("expected no warning for known usages, got %q", logs.String())
	}
}

func TestUnitPreserveUnrecognisedAssetUsages(t *testing.T) {
	current := []platformclientv2.Rmsassetaddressableref{
		{Id: platformclientv2.String("future-1"), AssetUsage: platformclientv2.String("SomeFutureUsage")},
		{Id: platformclientv2.String("img-1")},
		{Id: platformclientv2.String("pdf-1"), AssetUsage: platformclientv2.String(assetUsageFileAttachments)},
	}
	outgoing := *buildResponseAssets(stringSet("future-1", "img-1", "pdf-1"), stringSet())

	preserveUnrecognisedAssetUsages(&outgoing, &current)

	usages := assetUsagesById(&outgoing)
	if usages["future-1"] == nil || *usages["future-1"] != "SomeFutureUsage" {
		t.Errorf("unrecognised usage must be kept on update, got %v", usages["future-1"])
	}
	if usages["img-1"] != nil {
		t.Errorf("inline image must stay without usage, got %q", *usages["img-1"])
	}
	if usages["pdf-1"] != nil {
		t.Errorf("FileAttachments is modelled by file_attachment_asset_ids; moving it to asset_ids must clear it, got %q", *usages["pdf-1"])
	}
}

func TestUnitPreserveUnrecognisedAssetUsagesDoesNotOverrideSetUsage(t *testing.T) {
	current := []platformclientv2.Rmsassetaddressableref{
		{Id: platformclientv2.String("future-1"), AssetUsage: platformclientv2.String("SomeFutureUsage")},
	}
	outgoing := *buildResponseAssets(stringSet(), stringSet("future-1"))

	preserveUnrecognisedAssetUsages(&outgoing, &current)

	if usage := assetUsagesById(&outgoing)["future-1"]; usage == nil || *usage != assetUsageFileAttachments {
		t.Errorf("an explicit file attachment must win over the current usage, got %v", usage)
	}
}

func TestUnitPreserveUnrecognisedAssetUsagesNil(t *testing.T) {
	preserveUnrecognisedAssetUsages(nil, nil)
	outgoing := *buildResponseAssets(stringSet("img-1"), stringSet())
	preserveUnrecognisedAssetUsages(&outgoing, nil)
	if assetUsagesById(&outgoing)["img-1"] != nil {
		t.Error("expected no change when the current response has no assets")
	}
}
