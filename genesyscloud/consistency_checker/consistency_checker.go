package consistency_checker

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"reflect"
	"strconv"
	"strings"
	"sync"
	featureToggles "terraform-provider-genesyscloud/genesyscloud/util/feature_toggles"
	"unsafe"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

var (
	mcc      map[string]*ConsistencyCheck
	mccMutex sync.RWMutex
)

func init() {
	mcc = make(map[string]*ConsistencyCheck)
	mccMutex = sync.RWMutex{}
}

type ConsistencyCheck struct {
	ctx            context.Context
	r              *schema.Resource
	originalState  map[string]interface{}
	meta           interface{}
	isEmptyState   *bool
	checks         int
	maxStateChecks int
	resourceType   string
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
actual value:   %v`, e.key, e.oldValue, e.newValue)
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

	schemaInterface := getUnexportedField(reflect.ValueOf(d).Elem().FieldByName("schema"))
	resourceSchema := schemaInterface.(map[string]*schema.Schema)

	originalState := make(map[string]interface{})
	for k := range resourceSchema {
		originalState[k] = d.Get(k)
	}

	cc = &ConsistencyCheck{
		ctx:            ctx,
		r:              r,
		originalState:  originalState,
		meta:           meta,
		isEmptyState:   emptyState,
		maxStateChecks: maxStateChecks,
		resourceType:   resourceType,
	}
	mccMutex.Lock()
	mcc[d.Id()] = cc
	mccMutex.Unlock()

	return cc
}

func DeleteConsistencyCheck(id string) {
	mccMutex.Lock()
	delete(mcc, id)
	mccMutex.Unlock()
}

func getUnexportedField(field reflect.Value) interface{} {
	return reflect.NewAt(field.Type(), unsafe.Pointer(field.UnsafeAddr())).Elem().Interface()
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

func (c *ConsistencyCheck) isComputed(d *schema.ResourceData, key string) bool {
	schemaInterface := getUnexportedField(reflect.ValueOf(d).Elem().FieldByName("schema"))
	resourceSchema := schemaInterface.(map[string]*schema.Schema)

	k := key
	if strings.Contains(key, ".") {
		k = strings.Split(key, ".")[0]
	}
	if resourceSchema[k] == nil {
		return false
	}

	return resourceSchema[k].Computed
}

func (c *ConsistencyCheck) CheckState(currentState *schema.ResourceData) *retry.RetryError {
	if c.isEmptyState == nil {
		panic("consistencyCheck must be initialized with NewConsistencyCheck")
	}

	if c.r == nil {
		return nil
	}

	if featureToggles.CCToggleExists() {
		log.Printf("%s is set, write consistency errors to consistency-errors.log.json", featureToggles.CCToggleName())
	} else {
		log.Printf("%s is not set, consistency checker behaving as default", featureToggles.CCToggleName())
	}

	originalState := filterMap(c.originalState)

	resourceConfig := &terraform.ResourceConfig{
		ComputedKeys: []string{},
		Config:       originalState,
		Raw:          originalState,
	}

	diff, _ := c.r.SimpleDiff(c.ctx, currentState.State(), resourceConfig, c.meta)
	if diff != nil && len(diff.Attributes) > 0 {
		for k, v := range diff.Attributes {
			if strings.HasSuffix(k, "#") {
				continue
			}
			vTemp := v.Old
			v.Old = v.New
			v.New = vTemp
			parts := strings.Split(k, ".")
			if strings.Contains(k, ".") {
				slice1Index, _ := strconv.Atoi(parts[1])
				slice2Index := 0
				key := ""
				if len(parts) >= 3 {
					key = parts[2]
					if len(parts) == 4 {
						slice2Index, _ = strconv.Atoi(parts[3])
					}
				}

				vv := v.New
				if currentState.HasChange(k) {
					if !compareValues(c.originalState[parts[0]], vv, slice1Index, slice2Index, key) {
						err := retry.RetryableError(&consistencyError{
							key:      k,
							oldValue: c.originalState[k],
							newValue: currentState.Get(k),
						})

						if exists := featureToggles.CCToggleExists(); c.checks >= c.maxStateChecks && exists {
							c.writeConsistencyErrorToFile(currentState, err)
							return nil
						}

						c.checks++
						return err
					}
				}
			} else {
				if currentState.HasChange(k) {
					err := retry.RetryableError(&consistencyError{
						key:      k,
						oldValue: c.originalState[k],
						newValue: currentState.Get(k),
					})

					if exists := featureToggles.CCToggleExists(); c.checks >= c.maxStateChecks && exists {
						c.writeConsistencyErrorToFile(currentState, err)
						return nil
					}

					c.checks++
					return err
				}
			}
		}
	}

	DeleteConsistencyCheck(currentState.Id())
	return nil
}

func (c *ConsistencyCheck) writeConsistencyErrorToFile(d *schema.ResourceData, consistencyError *retry.RetryError) {
	const filePath = "consistency-errors.log.json"
	errorJson := consistencyErrorJson{
		ResourceType: c.resourceType,
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
