package consistency_checker

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
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

// This unit test will test the consistency checkers ability to handle
// nested blocks that reorder unexpectedly
func TestUnitConsistencyChecker(t *testing.T) {
	// Create a sample resource to use to test the consistency checker
	tId := uuid.NewString()
	tName := "Sample name"
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
	resourceDataMap := buildPersonResourceMap(tId, tName, []siblingStruct{tSibling1, tSibling2})

	d := schema.TestResourceDataRaw(t, resourceSchema, resourceDataMap)
	d.SetId(tId)

	fmt.Println(d.State().String())
	cc := NewConsistencyCheck(ctx, d, nil, resourcePerson(), 5, "person")

	// Reverse the order of the siblings
	_ = d.Set("siblings", flattenSiblings([]siblingStruct{tSibling2, tSibling1}))
	fmt.Println(d.State().String())
	err := retry.RetryContext(ctx, time.Second*5, func() *retry.RetryError {
		return cc.CheckState(d) // Consistency checker should handle the re-order
	})

	assert.Nilf(t, err, "%s", err.Error())
}

func buildPersonResourceMap(tId string, tName string, tSiblings []siblingStruct) map[string]interface{} {
	resourceDataMap := map[string]interface{}{
		"id":       tId,
		"name":     tName,
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
