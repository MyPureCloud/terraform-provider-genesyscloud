package consistency_checker

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
)

var friendBlock = &schema.Resource{
	Schema: map[string]*schema.Schema{
		"name": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"computed_value": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Used to test a nested computed value",
		},
	},
}

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
			"computed_value": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Used the test a computed value at the top level",
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
			"friends": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     friendBlock,
			},
		},
	}
}

type siblingStruct struct {
	siblingId    string
	siblingNames []string
}

type friendStruct struct {
	friendName    string
	computedValue string
}

// TestUnitConsistencyCheckerBlockBasic will test the consistency checkers ability to handle
// when a property changes unexpectedly
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
	friend := friendStruct{
		friendName: "Tom",
	}

	ctx := context.Background()

	resourceSchema := resourcePerson().Schema
	resourceDataMap := buildPersonResourceMap(tId, tName, tAge, []siblingStruct{tSibling1, tSibling2}, []friendStruct{friend})

	d := schema.TestResourceDataRaw(t, resourceSchema, resourceDataMap)
	d.SetId(tId)

	cc := NewConsistencyCheck(ctx, d, nil, resourcePerson(), 5, "person")

	// Change an attribute value and run the consistency checker
	tNameNew := "new name"
	_ = d.Set("name", tNameNew)
	err := retry.RetryContext(ctx, time.Second*5, func() *retry.RetryError {
		return cc.CheckState(d) // Consistency checker should catch the unexpected change and return an error
	})

	// Make sure the consistency checker gave an error
	assert.NotNil(t, err, "Error is nil. Consistency checker did not catch unexpected change")

	// Check for expected error
	expectedErr := "mismatch on attribute name:\nexpected value: Sample name\nactual value: new name"
	assert.Equal(t, err.Error(), expectedErr, fmt.Sprintf("Incorrect error:\nExpect Error: %s\nActual Error: %s", expectedErr, err))
}

// TestUnitConsistencyCheckerBlockBasic will test the consistency checkers ability to handle
// computed values at any level (top level or nested)
func TestUnitConsistencyCheckerComputed(t *testing.T) {
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
	friend := friendStruct{
		friendName: "Tom",
	}

	ctx := context.Background()

	resourceSchema := resourcePerson().Schema
	resourceDataMap := buildPersonResourceMap(tId, tName, tAge, []siblingStruct{tSibling1, tSibling2}, []friendStruct{friend})

	d := schema.TestResourceDataRaw(t, resourceSchema, resourceDataMap)
	d.SetId(tId)

	cc := NewConsistencyCheck(ctx, d, nil, resourcePerson(), 5, "person")

	// Set the computed values and run the consistency checker
	_ = d.Set("computed_value", "012345")
	err := retry.RetryContext(ctx, time.Second*5, func() *retry.RetryError {
		return cc.CheckState(d) // Consistency checker should handle the computed value and not return an error
	})

	// Check for error
	assert.Nil(t, err, "Consistency error returned: %s", err)
}

// TestUnitConsistencyCheckerBlocks will test the consistency checkers ability to handle
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
	tSibling3 := siblingStruct{
		siblingId:    "13579",
		siblingNames: []string{"John", "Paul", "Riley"},
	}
	friend := friendStruct{
		friendName: "Tom",
	}

	ctx := context.Background()

	resourceSchema := resourcePerson().Schema
	resourceDataMap := buildPersonResourceMap(tId, tName, tAge, []siblingStruct{tSibling1, tSibling2}, []friendStruct{friend})

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
	_ = d.Set("siblings", flattenSiblings([]siblingStruct{tSibling1}))
	err = retry.RetryContext(ctx, time.Second*5, func() *retry.RetryError {
		return cc.CheckState(d)
	})

	expectedErr := "mismatch on attribute siblings.#:\nexpected value: 2\nactual value: 1"
	assert.Equal(t, err.Error(), expectedErr, fmt.Sprintf("Incorrect error:\nExpect Error: %s\nActual Error: %s", expectedErr, err))

	// Add a sibling block and run consistency checker
	_ = d.Set("siblings", flattenSiblings([]siblingStruct{tSibling1, tSibling2, tSibling3}))
	err = retry.RetryContext(ctx, time.Second*5, func() *retry.RetryError {
		return cc.CheckState(d)
	})

	expectedErr = "mismatch on attribute siblings.#:\nexpected value: 2\nactual value: 3"
	assert.Equal(t, err.Error(), expectedErr, fmt.Sprintf("Incorrect error:\nExpect Error: %s\nActual Error: %s", expectedErr, err))

	//Change a sibling block and run consistency checker
	_ = d.Set("siblings", flattenSiblings([]siblingStruct{tSibling1, tSibling3}))
	err = retry.RetryContext(ctx, time.Second*5, func() *retry.RetryError {
		return cc.CheckState(d)
	})
	assert.ErrorContains(t, err, "mismatch on attribute") // Can't guarantee what attribute will be checked first we can't check for specific error message
}

func buildPersonResourceMap(tId string, tName string, tAge int, tSiblings []siblingStruct, tFriends []friendStruct) map[string]interface{} {
	resourceDataMap := map[string]interface{}{
		"id":       tId,
		"name":     tName,
		"age":      tAge,
		"siblings": flattenSiblings(tSiblings),
		"friends":  flattenFriends(tFriends),
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

func flattenFriends(friends []friendStruct) *schema.Set {
	friendSet := schema.NewSet(schema.HashResource(friendBlock), []interface{}{})

	for _, friend := range friends {
		var friendMap = make(map[string]interface{})

		friendMap["name"] = friend.friendName
		if friend.computedValue != "" {
			friendMap["computed_value"] = friend.computedValue
		}

		friendSet.Add(friendMap)
	}

	return friendSet
}
