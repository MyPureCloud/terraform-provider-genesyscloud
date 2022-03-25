package consistency_checker

import (
	"context"
	"fmt"
	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"unsafe"
)

var (
	mcc      map[string]*consistencyCheck
	mccMutex sync.RWMutex
)

func init() {
	mcc = make(map[string]*consistencyCheck)
	mccMutex = sync.RWMutex{}
}

type consistencyCheck struct {
	ctx           context.Context
	d             *schema.ResourceData
	r             *schema.Resource
	originalState map[string]interface{}
	meta          interface{}
	isEmptyState  *bool
}

type consistencyError struct {
	key      string
	oldValue interface{}
	newValue interface{}
}

func (e *consistencyError) Error() string {
	return fmt.Sprintf(`mismatch on attribute %s:
expected value: %v
actual value:   %v`, e.key, e.oldValue, e.newValue)
}

func NewConsistencyCheck(ctx context.Context, d *schema.ResourceData, meta interface{}, r *schema.Resource) *consistencyCheck {
	emptyState := isEmptyState(d)
	if *emptyState {
		return &consistencyCheck{isEmptyState: emptyState}
	}
	var cc *consistencyCheck

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

	cc = &consistencyCheck{
		ctx:           ctx,
		d:             d,
		r:             r,
		originalState: originalState,
		meta:          meta,
		isEmptyState:  emptyState,
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

func compareValues(oldValue, newValue interface{}, index, index2 int, key string) bool {
	switch oldValueType := oldValue.(type) {
	case []interface{}:
		if index >= len(oldValueType) {
			return false
		}
		ov := oldValueType[index]
		switch t := ov.(type) {
		case map[string]interface{}:
			return compareValues(t[key], newValue, index2, 0, "")
		default:
			return cmp.Equal(ov, newValue)
		}
	case string:
		return cmp.Equal(oldValue, newValue)
	}

	return true
}

func isNilOrEmpty(value interface{}) bool {
	if value == nil {
		return true
	}

	switch valueType := value.(type) {
	case int:
		return valueType == 0
	case string:
		return valueType == ""
	case []interface{}:
		return len(valueType) == 0
	case *schema.Set:
		if valueType == nil {
			return true
		}
		return valueType.Len() == 0
	default:
		// An interface holding a nil value is not nil. Need to use reflection to see if it's nil
		kind := reflect.ValueOf(value).Kind()
		if kind == reflect.Map {
			return reflect.ValueOf(value).IsNil()
		}
	}

	return false
}

func (c *consistencyCheck) isComputed(key string) bool {
	schemaInterface := getUnexportedField(reflect.ValueOf(c.d).Elem().FieldByName("schema"))
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

func (c *consistencyCheck) CheckState() *resource.RetryError {
	if c.isEmptyState == nil {
		panic("consistencyCheck must be initialized with NewConsistencyCheck")
	}

	if *c.isEmptyState {
		return nil
	}

	originalState := filterMap(c.originalState)
	resourceConfig := &terraform.ResourceConfig{
		ComputedKeys: []string{},
		Config:       originalState,
		Raw:          originalState,
	}
	diff, _ := c.r.SimpleDiff(c.ctx, c.d.State(), resourceConfig, c.meta)
	if diff != nil {
		//fmt.Println(diff)
		for k, v := range diff.Attributes {
			if strings.HasSuffix(k, "#") {
				continue
			}

			//fmt.Println("!c.d.HasChange(k)", !c.d.HasChange(k))
			//fmt.Println("isNilOrEmpty(c.d.Get(k))", isNilOrEmpty(c.d.Get(k)))
			//fmt.Println("isNilOrEmpty(c.originalState[k])", isNilOrEmpty(c.originalState[k]))
			//fmt.Println("c.isComputed(k)", c.isComputed(k))
			//fmt.Println("key", k)
			if !c.d.HasChange(k) /*&& (isNilOrEmpty(c.d.Get(k)) || isNilOrEmpty(c.originalState[k]))*/ && c.isComputed(k) {
				//fmt.Println("continue")
				continue
			}

			//if c.d.HasChange(k) {
			//fmt.Println("has change", k)
			if strings.Contains(k, ".") {
				//fmt.Println("contains")
				parts := strings.Split(k, ".")
				i, _ := strconv.Atoi(parts[1])
				index2 := 0
				key := ""
				if len(parts) >= 3 {
					key = parts[2]
				}
				if len(parts) == 4 {
					index2, _ = strconv.Atoi(parts[3])
				}
				if !compareValues(c.originalState[parts[0]], v.Old, i, index2, key) {
					return resource.RetryableError(&consistencyError{
						key:      k,
						oldValue: c.originalState[k],
						newValue: c.d.Get(k),
					})
				}
			} else {
				return resource.RetryableError(&consistencyError{
					key:      k,
					oldValue: c.originalState[k],
					newValue: c.d.Get(k),
				})
			}
			//}
		}
	}

	DeleteConsistencyCheck(c.d.Id())

	return nil
}
