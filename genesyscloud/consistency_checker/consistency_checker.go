package consistency_checker

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"log"
	"os"
	"reflect"
	"strings"
	"sync"
	featureToggles "terraform-provider-genesyscloud/genesyscloud/util/feature_toggles"
	"unsafe"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"github.com/google/go-cmp/cmp"
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
	ctx              context.Context
	resource         *schema.Resource
	originalState    *schema.ResourceData
	originalStateMap map[string]interface{}
	meta             interface{}
	isEmptyState     *bool
	checks           int
	maxStateChecks   int
	resourceType     string
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

	originalStateMap := filterMap(resourceDataToMap(d))

	cc = &ConsistencyCheck{
		ctx:              ctx,
		resource:         r,
		originalState:    d,
		originalStateMap: originalStateMap,
		meta:             meta,
		isEmptyState:     emptyState,
		maxStateChecks:   maxStateChecks,
		resourceType:     resourceType,
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

func getUnexportedField(field reflect.Value) interface{} {
	return reflect.NewAt(field.Type(), unsafe.Pointer(field.UnsafeAddr())).Elem().Interface()
}

func resourceDataToMap(d *schema.ResourceData) map[string]interface{} {
	schemaInterface := getUnexportedField(reflect.ValueOf(d).Elem().FieldByName("schema"))
	resourceSchema := schemaInterface.(map[string]*schema.Schema)

	originalState := make(map[string]interface{})
	for k := range resourceSchema {
		originalState[k] = d.Get(k)
	}

	return originalState
}

func isEmptyState(d *schema.ResourceData) *bool {
	stateString := strings.Split(d.State().String(), "\n")
	isEmpty := true
	for _, s := range stateString {
		if len(s) == 0 {
			continue
		}
		sSplit := strings.Split(s, " ")
		attribute := sSplit[0]
		if attribute == "ID" ||
			attribute == "Tainted" ||
			strings.HasSuffix(s, ".# = 0") ||
			strings.HasSuffix(attribute, "id") {
			continue
		}
		isEmpty = false
		break
	}
	return &isEmpty
}

func filterMap(m map[string]interface{}) map[string]interface{} {
	newM := make(map[string]interface{})
	for k, v := range m {
		switch t := v.(type) {
		case *schema.Set:
			newM[k] = filterMapSlice(t.List())
		case []interface{}:
			newM[k] = filterMapSlice(t)
		case map[string]interface{}:
			if len(t) > 0 {
				newM[k] = filterMap(t)
			}
		default:
			newM[k] = v
		}
	}

	return newM
}

func filterMapSlice(unfilteredSlice []interface{}) interface{} {
	if len(unfilteredSlice) == 0 {
		return unfilteredSlice
	}

	switch unfilteredSlice[0].(type) {
	case map[string]interface{}:
		filteredSlice := make([]interface{}, 0)
		for _, s := range unfilteredSlice {
			filteredSlice = append(filteredSlice, filterMap(s.(map[string]interface{})))
		}
		return filteredSlice
	}

	return unfilteredSlice
}

// CheckState will compare a state with the state that the consistency checker was initialized with
func (cc *ConsistencyCheck) CheckState(currentState *schema.ResourceData) *retry.RetryError {
	if cc.isEmptyState == nil {
		panic("consistencyCheck must be initialized with NewConsistencyCheck")
	}

	if cc.resource == nil {
		return nil
	}

	if featureToggles.CCToggleExists() {
		log.Printf("%s is set, write consistency errors to consistency-errors.log.json", featureToggles.CCToggleName())
	} else {
		log.Printf("%s is not set, consistency checker behaving as default", featureToggles.CCToggleName())
	}

	resourceConfig := &terraform.ResourceConfig{
		Config: cc.originalStateMap,
		Raw:    cc.originalStateMap,
	}

	diff, _ := cc.resource.Diff(cc.ctx, currentState.State(), resourceConfig, cc.meta)
	if diff != nil && len(diff.Attributes) > 0 {
		var nestedBlocksError bool // Used to indicate if there is a problem with any nested blocks
		for attribute := range diff.Attributes {
			// Make sure we have the same number of nested blocks
			if strings.HasSuffix(attribute, "#") {
				if cc.originalState.Get(attribute).(int) != currentState.Get(attribute).(int) {
					return cc.handleError(currentState, attribute)
				}
			}

			// Handle top level attributes
			if !strings.Contains(attribute, ".") {
				if currentState.HasChange(attribute) {
					return cc.handleError(currentState, attribute)
				}
			}

			// Handled nested blocks
			if strings.Contains(attribute, ".") && !nestedBlocksError {
				// We want to handle the nested blocks together, so we will do it out of the loop
				nestedBlocksError = true
			}
		}

		// If there's a problem with any nested blocks, check them manually
		if nestedBlocksError {
			fmt.Println(cc.originalStateMap)

			for attr, attrValue := range cc.originalStateMap {
				var err error
				switch attrValueType := attrValue.(type) {

				// Only check nested blocks, everything else has been checked above
				case []interface{}:
					err = cc.compareBlock(attr, attrValueType)
				case *schema.Set:
					err = cc.compareBlock(attr, attrValueType.List())
				}

				if err != nil {
					return cc.handleError(currentState, attr)
				}
			}
		}
	}

	DeleteConsistencyCheck(currentState.Id())
	return nil
}

func compareBlocks(resource map[string]interface{}) error {
	return nil
}

func (cc *ConsistencyCheck) compareBlock(blockName string, blockValues []interface{}) error {
	for _, blockValue := range blockValues {
		switch blockValue.(type) {
		case map[string]interface{}: // block is a nested resource
			fmt.Printf("Nested %s: %s\n", blockName, blockValue)
		default: // block is an array
			fmt.Printf("Array %s: %s\n", blockName, blockValue)
		}
	}

	return nil
}

func compareValues(oldValue, newValue interface{}, slice1Index, slice2Index int, key string) bool {
	switch oldValueType := oldValue.(type) {
	case []interface{}:
		if len(oldValueType) == 0 {
			return true
		}
		if slice1Index >= len(oldValueType) {
			for i := 0; i < len(oldValueType); i++ {
				if compareValues(oldValue, newValue, i, slice2Index, key) {
					return true
				}
			}
			return false
		}
		ov := oldValueType[slice1Index]
		switch t := ov.(type) {
		case map[string]interface{}:
			return compareValues(t[key], newValue, slice2Index, 0, "")
		default:
			return cmp.Equal(ov, newValue)
		}
	case *schema.Set:
		return compareValues(oldValueType.List(), newValue, slice1Index, slice2Index, key)
	case string:
		if oldValue != "" && newValue == "" {
			return true
		}
		return cmp.Equal(oldValue, newValue)
	default:
		return cmp.Equal(oldValue, newValue)
	}
}

func (cc *ConsistencyCheck) handleError(currentState *schema.ResourceData, attribute string) *retry.RetryError {
	err := retry.RetryableError(&consistencyError{
		key:      attribute,
		oldValue: cc.originalState.Get(attribute),
		newValue: currentState.Get(attribute),
	})

	return cc.writeOrReturnError(currentState, err)
}

func (cc *ConsistencyCheck) writeOrReturnError(currentState *schema.ResourceData, err *retry.RetryError) *retry.RetryError {
	if exists := featureToggles.CCToggleExists(); cc.checks >= cc.maxStateChecks && exists {
		cc.writeConsistencyErrorToFile(currentState, err)
		return nil
	}

	cc.checks++
	return err
}

func (cc *ConsistencyCheck) writeConsistencyErrorToFile(d *schema.ResourceData, consistencyError *retry.RetryError) {
	const filePath = "consistency-errors.log.json"
	errorJson := consistencyErrorJson{
		ResourceType: cc.resourceType,
		ResourceId:   d.Id(),
		ErrorMessage: consistencyError.Err.Error(),
	}

	if name, _ := d.Get("name").(string); name != "" {
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
