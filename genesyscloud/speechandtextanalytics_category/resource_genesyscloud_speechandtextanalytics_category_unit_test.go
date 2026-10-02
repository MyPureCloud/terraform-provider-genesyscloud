package speechandtextanalytics_category

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v195/platformclientv2"
	"github.com/stretchr/testify/assert"
)

// testCriteriaJSON mirrors the structure the API requires: the top two operand
// levels must be "OperandGroup", with "Term"/"Topic" leaves below them.
const testCriteriaJSON = `{"type":"OperandGroup","operands":[{"type":"OperandGroup","operands":[{"type":"Term","term":{"word":"refund","participantType":"External"}}]}]}`

// buildTestCriteriaOperand parses testCriteriaJSON into the SDK Operand so that the
// value returned by the read stub flattens back to an equivalent JSON string. This keeps
// the consistency checker in the read path happy (otherwise it retries until timeout).
func buildTestCriteriaOperand(t *testing.T) *platformclientv2.Operand {
	t.Helper()
	var criteria platformclientv2.Operand
	if err := json.Unmarshal([]byte(testCriteriaJSON), &criteria); err != nil {
		t.Fatalf("failed to build test criteria operand: %s", err)
	}
	return &criteria
}

func buildCategoryResourceMap(id, name, description, interactionType, criteria string) map[string]interface{} {
	return map[string]interface{}{
		"name":             name,
		"description":      description,
		"interaction_type": interactionType,
		"criteria":         criteria,
	}
}

func TestUnitCategoryCreate(t *testing.T) {
	var (
		id              = uuid.NewString()
		name            = "Unit Test Category"
		description     = "Test description"
		interactionType = "Voice"
	)

	proxy := &categoryProxy{}

	proxy.createCategoryAttr = func(ctx context.Context, p *categoryProxy, category *platformclientv2.Categoryrequest) (*platformclientv2.Stacategory, *platformclientv2.APIResponse, error) {
		assert.Equal(t, name, *category.Name, "category.Name check failed in create")
		assert.Equal(t, description, *category.Description, "category.Description check failed in create")
		assert.Equal(t, interactionType, *category.InteractionType, "category.InteractionType check failed in create")
		assert.NotNil(t, category.Criteria, "category.Criteria should be parsed from JSON in create")
		assert.Equal(t, "OperandGroup", *category.Criteria.VarType, "category.Criteria.type check failed in create")
		assert.NotNil(t, category.Criteria.Operands, "category.Criteria.operands should be populated in create")

		created := &platformclientv2.Stacategory{
			Id:              &id,
			Name:            &name,
			Description:     &description,
			InteractionType: &interactionType,
			Criteria:        category.Criteria,
		}
		return created, &platformclientv2.APIResponse{StatusCode: http.StatusOK}, nil
	}

	proxy.getCategoryByIdAttr = func(ctx context.Context, p *categoryProxy, gid string) (*platformclientv2.Stacategory, *platformclientv2.APIResponse, error) {
		assert.Equal(t, id, gid)
		category := &platformclientv2.Stacategory{
			Id:              &id,
			Name:            &name,
			Description:     &description,
			InteractionType: &interactionType,
			Criteria:        buildTestCriteriaOperand(t),
		}
		return category, &platformclientv2.APIResponse{StatusCode: http.StatusOK}, nil
	}

	internalProxy = proxy
	defer func() { internalProxy = nil }()

	ctx := context.Background()
	gCloud := &provider.ProviderMeta{ClientConfig: &platformclientv2.Configuration{}}
	d := schema.TestResourceDataRaw(t, ResourceCategory().Schema, buildCategoryResourceMap(id, name, description, interactionType, testCriteriaJSON))

	diag := createCategory(ctx, d, gCloud)
	assert.Equal(t, false, diag.HasError())
	assert.Equal(t, id, d.Id())
}

func TestUnitCategoryRead(t *testing.T) {
	var (
		id              = uuid.NewString()
		name            = "Unit Test Category"
		description     = "Test description"
		interactionType = "Digital"
	)

	proxy := &categoryProxy{}

	proxy.getCategoryByIdAttr = func(ctx context.Context, p *categoryProxy, gid string) (*platformclientv2.Stacategory, *platformclientv2.APIResponse, error) {
		assert.Equal(t, id, gid)
		category := &platformclientv2.Stacategory{
			Id:              &id,
			Name:            &name,
			Description:     &description,
			InteractionType: &interactionType,
			Criteria:        buildTestCriteriaOperand(t),
		}
		return category, &platformclientv2.APIResponse{StatusCode: http.StatusOK}, nil
	}

	internalProxy = proxy
	defer func() { internalProxy = nil }()

	ctx := context.Background()
	gCloud := &provider.ProviderMeta{ClientConfig: &platformclientv2.Configuration{}}
	d := schema.TestResourceDataRaw(t, ResourceCategory().Schema, buildCategoryResourceMap(id, name, description, interactionType, testCriteriaJSON))
	d.SetId(id)

	diag := readCategory(ctx, d, gCloud)
	assert.Equal(t, false, diag.HasError())
	assert.Equal(t, id, d.Id())
	assert.Equal(t, name, d.Get("name").(string))
	assert.Equal(t, description, d.Get("description").(string))
	assert.Equal(t, interactionType, d.Get("interaction_type").(string))
	assert.NotEmpty(t, d.Get("criteria").(string))
}

func TestUnitCategoryUpdate(t *testing.T) {
	var (
		id              = uuid.NewString()
		name            = "Updated Category"
		description     = "Updated description"
		interactionType = "Voice"
	)

	var capturedPutBody *platformclientv2.Categoryrequest

	proxy := &categoryProxy{}

	proxy.updateCategoryAttr = func(ctx context.Context, p *categoryProxy, uid string, category *platformclientv2.Categoryrequest) (*platformclientv2.Stacategory, *platformclientv2.APIResponse, error) {
		assert.Equal(t, id, uid)
		capturedPutBody = category
		updated := &platformclientv2.Stacategory{
			Id:              &id,
			Name:            category.Name,
			Description:     category.Description,
			InteractionType: category.InteractionType,
			Criteria:        category.Criteria,
		}
		return updated, &platformclientv2.APIResponse{StatusCode: http.StatusOK}, nil
	}

	proxy.getCategoryByIdAttr = func(ctx context.Context, p *categoryProxy, gid string) (*platformclientv2.Stacategory, *platformclientv2.APIResponse, error) {
		category := &platformclientv2.Stacategory{
			Id:              &id,
			Name:            &name,
			Description:     &description,
			InteractionType: &interactionType,
			Criteria:        buildTestCriteriaOperand(t),
		}
		return category, &platformclientv2.APIResponse{StatusCode: http.StatusOK}, nil
	}

	internalProxy = proxy
	defer func() { internalProxy = nil }()

	ctx := context.Background()
	gCloud := &provider.ProviderMeta{ClientConfig: &platformclientv2.Configuration{}}
	d := schema.TestResourceDataRaw(t, ResourceCategory().Schema, buildCategoryResourceMap(id, name, description, interactionType, testCriteriaJSON))
	d.SetId(id)

	diag := updateCategory(ctx, d, gCloud)
	assert.Equal(t, false, diag.HasError())
	assert.Equal(t, id, d.Id())
	assert.NotNil(t, capturedPutBody)
	assert.Equal(t, name, *capturedPutBody.Name)
}

func TestUnitCategoryDelete(t *testing.T) {
	id := uuid.NewString()

	proxy := &categoryProxy{}

	proxy.deleteCategoryAttr = func(ctx context.Context, p *categoryProxy, did string) (*platformclientv2.APIResponse, error) {
		assert.Equal(t, id, did)
		return &platformclientv2.APIResponse{StatusCode: http.StatusOK}, nil
	}

	proxy.getCategoryByIdAttr = func(ctx context.Context, p *categoryProxy, gid string) (*platformclientv2.Stacategory, *platformclientv2.APIResponse, error) {
		return nil, &platformclientv2.APIResponse{StatusCode: http.StatusNotFound}, assert.AnError
	}

	internalProxy = proxy
	defer func() { internalProxy = nil }()

	ctx := context.Background()
	gCloud := &provider.ProviderMeta{ClientConfig: &platformclientv2.Configuration{}}
	d := schema.TestResourceDataRaw(t, ResourceCategory().Schema, buildCategoryResourceMap(id, "n", "d", "Voice", testCriteriaJSON))
	d.SetId(id)

	diag := deleteCategory(ctx, d, gCloud)
	assert.Equal(t, false, diag.HasError())
}

func TestUnitDataSourceCategoryRead(t *testing.T) {
	var (
		id   = uuid.NewString()
		name = "Unit Test Category DS"
	)

	proxy := &categoryProxy{}

	proxy.getCategoryIdByNameAttr = func(ctx context.Context, p *categoryProxy, searchName string) (string, bool, *platformclientv2.APIResponse, error) {
		assert.Equal(t, name, searchName)
		return id, false, &platformclientv2.APIResponse{StatusCode: http.StatusOK}, nil
	}

	internalProxy = proxy
	defer func() { internalProxy = nil }()

	ctx := context.Background()
	gCloud := &provider.ProviderMeta{ClientConfig: &platformclientv2.Configuration{}}
	d := schema.TestResourceDataRaw(t, DataSourceCategory().Schema, map[string]interface{}{"name": name})

	diag := dataSourceCategoryRead(ctx, d, gCloud)
	assert.Equal(t, false, diag.HasError())
	assert.Equal(t, id, d.Id())
}

func TestUnitFlattenCriteriaToJSON(t *testing.T) {
	t.Run("nil criteria returns empty string", func(t *testing.T) {
		assert.Equal(t, "", flattenCriteriaToJSON(nil))
	})

	t.Run("valid criteria returns JSON string", func(t *testing.T) {
		varType := "Term"
		criteria := &platformclientv2.Operand{VarType: &varType}
		result := flattenCriteriaToJSON(criteria)
		assert.NotEmpty(t, result)
		assert.Contains(t, result, "Term")
	})

	t.Run("nested operands round-trip preserves structure", func(t *testing.T) {
		// An OperandGroup containing two Term operands (the shape the API expects).
		nested := `{"type":"OperandGroup","operands":[{"type":"Term","term":{"word":"refund"}},{"type":"Term","term":{"word":"cancel"}}]}`

		var parsed platformclientv2.Operand
		if err := json.Unmarshal([]byte(nested), &parsed); err != nil {
			t.Fatalf("failed to parse nested criteria: %s", err)
		}

		flattened := flattenCriteriaToJSON(&parsed)
		assert.NotEmpty(t, flattened)

		// Re-parse both and compare canonical JSON to ensure no data was lost.
		var original, roundTripped interface{}
		_ = json.Unmarshal([]byte(nested), &original)
		if err := json.Unmarshal([]byte(flattened), &roundTripped); err != nil {
			t.Fatalf("failed to re-parse flattened criteria: %s", err)
		}
		originalBytes, _ := json.Marshal(original)
		roundTrippedBytes, _ := json.Marshal(roundTripped)
		assert.JSONEq(t, string(originalBytes), string(roundTrippedBytes))
	})
}

func TestUnitBuildCategoryFromResourceData(t *testing.T) {
	t.Run("valid criteria JSON is parsed into Operand", func(t *testing.T) {
		d := schema.TestResourceDataRaw(t, ResourceCategory().Schema,
			buildCategoryResourceMap("", "cat", "desc", "Voice", testCriteriaJSON))

		category, err := buildCategoryFromResourceData(d)
		assert.NoError(t, err)
		assert.NotNil(t, category.Criteria)
		assert.Equal(t, "OperandGroup", *category.Criteria.VarType)
		assert.NotNil(t, category.Criteria.Operands)
		assert.Equal(t, "cat", *category.Name)
		assert.Equal(t, "Voice", *category.InteractionType)
	})

	t.Run("invalid criteria JSON returns an error", func(t *testing.T) {
		d := schema.TestResourceDataRaw(t, ResourceCategory().Schema,
			buildCategoryResourceMap("", "cat", "desc", "Voice", `{not valid json`))

		category, err := buildCategoryFromResourceData(d)
		assert.Error(t, err)
		assert.Nil(t, category)
	})

	t.Run("empty criteria string returns an error", func(t *testing.T) {
		// criteria is required; an empty string is not valid JSON and must be rejected.
		d := schema.TestResourceDataRaw(t, ResourceCategory().Schema,
			buildCategoryResourceMap("", "cat", "desc", "Voice", ""))

		category, err := buildCategoryFromResourceData(d)
		assert.Error(t, err)
		assert.Nil(t, category)
	})

	t.Run("required fields are always sent", func(t *testing.T) {
		d := schema.TestResourceDataRaw(t, ResourceCategory().Schema,
			buildCategoryResourceMap("", "cat", "desc", "All", testCriteriaJSON))

		category, err := buildCategoryFromResourceData(d)
		assert.NoError(t, err)
		assert.Equal(t, "cat", *category.Name)
		assert.Equal(t, "All", *category.InteractionType)
		assert.NotNil(t, category.Criteria)
		// description is optional but always sent so it can be cleared
		assert.NotNil(t, category.Description)
		assert.Equal(t, "desc", *category.Description)
	})
}

func TestUnitCategoryCreateError(t *testing.T) {
	proxy := &categoryProxy{}

	proxy.createCategoryAttr = func(ctx context.Context, p *categoryProxy, category *platformclientv2.Categoryrequest) (*platformclientv2.Stacategory, *platformclientv2.APIResponse, error) {
		return nil, &platformclientv2.APIResponse{StatusCode: http.StatusBadRequest}, assert.AnError
	}

	internalProxy = proxy
	defer func() { internalProxy = nil }()

	ctx := context.Background()
	gCloud := &provider.ProviderMeta{ClientConfig: &platformclientv2.Configuration{}}
	d := schema.TestResourceDataRaw(t, ResourceCategory().Schema,
		buildCategoryResourceMap("", "cat", "desc", "Voice", testCriteriaJSON))

	diag := createCategory(ctx, d, gCloud)
	assert.Equal(t, true, diag.HasError())
}

func TestUnitCategoryCreateInvalidCriteria(t *testing.T) {
	// createCategory should fail fast on invalid criteria JSON before any API call.
	proxy := &categoryProxy{}
	proxy.createCategoryAttr = func(ctx context.Context, p *categoryProxy, category *platformclientv2.Categoryrequest) (*platformclientv2.Stacategory, *platformclientv2.APIResponse, error) {
		t.Fatal("create API should not be called when criteria JSON is invalid")
		return nil, nil, nil
	}

	internalProxy = proxy
	defer func() { internalProxy = nil }()

	ctx := context.Background()
	gCloud := &provider.ProviderMeta{ClientConfig: &platformclientv2.Configuration{}}
	d := schema.TestResourceDataRaw(t, ResourceCategory().Schema,
		buildCategoryResourceMap("", "cat", "desc", "Voice", `{not valid json`))

	diag := createCategory(ctx, d, gCloud)
	assert.Equal(t, true, diag.HasError())
}

func TestUnitDataSourceCategoryReadNotFound(t *testing.T) {
	proxy := &categoryProxy{}

	// Non-retryable error path: the data source should surface an error, not spin.
	proxy.getCategoryIdByNameAttr = func(ctx context.Context, p *categoryProxy, searchName string) (string, bool, *platformclientv2.APIResponse, error) {
		return "", false, &platformclientv2.APIResponse{StatusCode: http.StatusInternalServerError}, assert.AnError
	}

	internalProxy = proxy
	defer func() { internalProxy = nil }()

	ctx := context.Background()
	gCloud := &provider.ProviderMeta{ClientConfig: &platformclientv2.Configuration{}}
	d := schema.TestResourceDataRaw(t, DataSourceCategory().Schema, map[string]interface{}{"name": "missing"})

	diag := dataSourceCategoryRead(ctx, d, gCloud)
	assert.Equal(t, true, diag.HasError())
}
