package consistency_checker

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
)

// this is a test resource to avoid importing an existing resource
// which would cause an import cycle
func resourcePerson() *schema.Resource {
	return &schema.Resource{
		Description: "A test resource for unit tests",
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"age": {
				Type:     schema.TypeInt,
				Required: true,
			},
			"siblings": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"sibling_id": {
							Type:     schema.TypeString,
							Required: true,
						},
						"sibling_names": {
							Type:     schema.TypeSet,
							Required: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},
		},
	}
}

type siblingStruct struct {
	siblingId    string
	siblingNames []string
}

// TestUnitConsistencyCheckerBlockBasic will test the consistency checkers ability to handle
// when properties changes unexpectedly
func TestUnitConsistencyCheckerBlockBasic(t *testing.T) {
	// Create a sample resource to use to test the consistency checker
	tId := uuid.NewString()
	tName := "Sample name"
	tAge := 20
	tSibling1 := siblingStruct{
		siblingId:    uuid.NewString(),
		siblingNames: []string{"Bob", "Tod", "Jr"},
	}
	tSibling2 := siblingStruct{
		siblingId:    uuid.NewString(),
		siblingNames: []string{"Mary", "Beth", "Smith"},
	}

	ctx := context.Background()

	resourceSchema := resourcePerson().Schema
	resourceDataMap := buildPersonResourceMap(tId, tName, tAge, []siblingStruct{tSibling1, tSibling2})

	d := schema.TestResourceDataRaw(t, resourceSchema, resourceDataMap)
	d.SetId(tId)

	cc := NewConsistencyCheck(ctx, d, nil, resourcePerson(), 5, "person")

	// Change an attribute value and run the consistency checker
	tNameNew := "new name"
	_ = d.Set("name", tNameNew)
	err := retry.RetryContext(ctx, time.Second*5, func() *retry.RetryError {
		return cc.CheckState(d) // Consistency checker should catch the unexpected change and return an error
	})

	// Match sure the consistency checker gave an error
	assert.NotNil(t, err, "Error is nil. Consistency checker did not catch unexpected change")

	// Check for expected error
	expectedErr := "mismatch on attribute name:\nexpected value: Sample name\nactual value: new name"	
	assert.Contains(t, err, expectedErr, fmt.Sprintf("Incorrect error:\nExpect Error: %s\nActual Error: %s", strings.ReplaceAll(expectedErr, "\n", " "), strings.ReplaceAll(err.Error(), "\n", " ")))
}

// TestUnitConsistencyCheckerBlocksReorder will test the consistency checkers ability to handle
// nested blocks that change unexpectedly
func TestUnitConsistencyCheckerNestedBlocks(t *testing.T) {
	// Create a sample resource to use to test the consistency checker
	tId := uuid.NewString()
	tName := "Sample name"
	tAge := 20
	tSibling1 := siblingStruct{
		siblingId:    "01234",
		siblingNames: []string{"Bob", "Tod", "Jr"},
	}
	tSibling2 := siblingStruct{
		siblingId:    "56789",
		siblingNames: []string{"Mary", "Beth", "Smith"},
	}

	ctx := context.Background()

	resourceSchema := resourcePerson().Schema
	resourceDataMap := buildPersonResourceMap(tId, tName, tAge, []siblingStruct{tSibling1, tSibling2})

	d := schema.TestResourceDataRaw(t, resourceSchema, resourceDataMap)
	d.SetId(tId)

	cc := NewConsistencyCheck(ctx, d, nil, resourcePerson(), 5, "person")

	// Reverse the order of the sibling blocks and run consistency checker
	_ = d.Set("siblings", flattenSiblings([]siblingStruct{tSibling2, tSibling1}))
	err := retry.RetryContext(ctx, time.Second*5, func() *retry.RetryError {
		return cc.CheckState(d) // Consistency checker should handle the re-order and not return an error
	})

	assert.Nilf(t, err, "%s", err)

	// Remove a sibling block and run consistency checker

	// Add a sibling block and run consistency checker
}

func buildPersonResourceMap(tId string, tName string, tAge int, tSiblings []siblingStruct) map[string]interface{} {
	resourceDataMap := map[string]interface{}{
		"id":       tId,
		"name":     tName,
		"age":      tAge,
		"siblings": flattenSiblings(tSiblings),
	}
	return resourceDataMap
}

func flattenSiblings(siblings []siblingStruct) []interface{} {
	var siblingsList []interface{}

	for _, sibling := range siblings {
		siblingMap := make(map[string]interface{})

		siblingMap["sibling_id"] = sibling.siblingId
		var names []interface{}
		for _, name := range sibling.siblingNames {
			names = append(names, name)
		}
		siblingMap["sibling_names"] = names

		siblingsList = append(siblingsList, siblingMap)
	}

	return siblingsList
}
