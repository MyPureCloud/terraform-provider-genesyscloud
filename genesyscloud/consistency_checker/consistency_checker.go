package consistency_checker

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"
	"sync"
	featureToggles "terraform-provider-genesyscloud/genesyscloud/util/feature_toggles"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var (
	mcc      map[string]*ConsistencyCheck
	mccMutex sync.Mutex
)

func init() {
	mcc = make(map[string]*ConsistencyCheck)
	mccMutex = sync.Mutex{}
}

type ConsistencyCheck struct {
	ctx                 context.Context
	resource            *schema.Resource
	originalState       *schema.ResourceData
	originalStateMap    map[string]interface{}
	originalStateValues map[string]string
	meta                interface{}
	isEmptyState        *bool
	checks              int
	maxStateChecks      int
	resourceType        string
}

type consistencyError struct {
	key      string
	oldValue interface{}
	newValue interface{}
}

type consistencyErrorJson struct {
	ResourceType     string `json:"resourceType"`
	ResourceId       string `json:"resourceId"`
	GCloudObjectName string `json:"GCloudObjectName"`
	ErrorMessage     string `json:"errorMessage"`
}

func (e *consistencyError) Error() string {
	return fmt.Sprintf(`mismatch on attribute %s:
expected value: %v
actual value: %v`, e.key, e.oldValue, e.newValue)
}

func NewConsistencyCheck(ctx context.Context, d *schema.ResourceData, meta interface{}, r *schema.Resource, maxStateChecks int, resourceType string) *ConsistencyCheck {
	emptyState := isEmptyState(d)
	if *emptyState {
		return &ConsistencyCheck{isEmptyState: emptyState}
	}
	var cc *ConsistencyCheck

	mccMutex.Lock()
	cc = mcc[d.Id()]
	mccMutex.Unlock()

	if cc != nil {
		return cc
	}

	cc = &ConsistencyCheck{
		ctx:                 ctx,
		resource:            r,
		originalState:       d,
		originalStateMap:    resourceDataToMap(d),
		originalStateValues: d.State().Attributes,
		meta:                meta,
		isEmptyState:        emptyState,
		maxStateChecks:      maxStateChecks,
		resourceType:        resourceType,
	}

	mccMutex.Lock()
	defer mccMutex.Unlock()
	mcc[d.Id()] = cc

	return cc
}

func DeleteConsistencyCheck(id string) {
	mccMutex.Lock()
	defer mccMutex.Unlock()
	delete(mcc, id)
}

// CheckState will compare the current state of a resource with the original state
func (cc *ConsistencyCheck) CheckState(currentState *schema.ResourceData) *retry.RetryError {
	if cc.isEmptyState == nil {
		panic("consistencyCheck must be initialized with NewConsistencyCheck")
	}

	if cc.resource == nil {
		return nil
	}

	resourceConfig := &terraform.ResourceConfig{
		ComputedKeys: []string{},
		Config:       cc.originalStateMap,
		Raw:          cc.originalStateMap,
	}

	diff, _ := cc.resource.SimpleDiff(cc.ctx, currentState.State(), resourceConfig, cc.meta)
	if diff != nil && len(diff.Attributes) > 0 {
		currentStateMap := resourceDataToMap(currentState)
		currentStateValues := currentState.State().Attributes

		// Sort attributes. This ensure we check top level attributes and number of blocks before nested blocks
		var attributesSorted []string
		for attribute := range diff.Attributes {
			attributesSorted = append(attributesSorted, attribute)
		}
		sort.Slice(attributesSorted, func(i, j int) bool {
			containsSpecialChar := func(s string) bool {
				return !strings.Contains(s, ".") || strings.Contains(s, "#")
			}

			// Both strings contain special characters or both don't
			if containsSpecialChar(attributesSorted[i]) == containsSpecialChar(attributesSorted[j]) {
				return attributesSorted[i] < attributesSorted[j] // Regular alphabetical order
			}

			return containsSpecialChar(attributesSorted[i])
		})

		for _, attribute := range attributesSorted {
			// Handle top level attributes and check we have the same number of nested blocks
			if !strings.Contains(attribute, ".") || strings.HasSuffix(attribute, "#") {
				if cc.originalStateValues[attribute] != currentStateValues[attribute] {
					return cc.handleError(attribute, cc.originalStateValues[attribute], currentStateValues[attribute])
				}
			} else {
				// Handled nested blocks
				if !compareStateMaps(cc.originalStateMap, currentStateMap) {
					return cc.handleError(attribute, cc.originalStateValues[attribute], currentStateValues[attribute])
				}
			}
		}
	}

	DeleteConsistencyCheck(currentState.Id())
	return nil
}

// handleError will create the error message the consistency checker will throw and check if we should return it or write to a file
func (cc *ConsistencyCheck) handleError(attribute string, originalValue string, currentValue string) *retry.RetryError {
	err := retry.RetryableError(&consistencyError{
		key:      attribute,
		oldValue: originalValue,
		newValue: currentValue,
	})

	if exists := featureToggles.CCToggleExists(); cc.checks >= cc.maxStateChecks && exists {
		log.Printf("%s is set, writing consistency errors to consistency-errors.log.json", featureToggles.CCToggleName())

		cc.writeConsistencyErrorToFile(err)
		return nil
	}

	cc.checks++
	return err
}

// writeConsistencyErrorToFile will create a JSON error and write it to a file
func (cc *ConsistencyCheck) writeConsistencyErrorToFile(consistencyError *retry.RetryError) {
	const filePath = "consistency-errors.log.json"
	errorJson := consistencyErrorJson{
		ResourceType: cc.resourceType,
		ResourceId:   cc.originalState.Id(),
		ErrorMessage: consistencyError.Err.Error(),
	}

	if name, _ := cc.originalState.Get("name").(string); name != "" {
		errorJson.GCloudObjectName = name
	}

	jsonData, err := json.Marshal(errorJson)
	if err != nil {
		fmt.Println("Error marshaling JSON:", err)
		return
	}

	err = os.WriteFile(filePath, jsonData, os.ModePerm)
	if err != nil {
		log.Printf("Error writing file %s: %v", filePath, err)
	}
}
