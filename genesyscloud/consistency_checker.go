package genesyscloud

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"reflect"
	"strings"
	"sync"
	"unsafe"
)

var (
	mcc map[string]*consistencyCheck
)

func init() {
	mcc = make(map[string]*consistencyCheck)
}

type consistencyCheck struct {
	originalState  map[string]interface{}
	originalSchema map[string]*schema.Schema
	d              *schema.ResourceData
	dMutex         sync.RWMutex
	isEmptyState   *bool
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

func NewConsistencyCheck(d *schema.ResourceData) *consistencyCheck {
	if mcc[d.Id()] == nil {
		schemaInterface := getUnexportedField(reflect.ValueOf(d).Elem().FieldByName("schema"))
		resourceSchema := schemaInterface.(map[string]*schema.Schema)

		originalState := make(map[string]interface{})
		originalSchema := make(map[string]*schema.Schema)
		for k, v := range resourceSchema {
			originalState[k] = d.Get(k)
			originalSchema[k] = v
		}

		mcc[d.Id()] = &consistencyCheck{
			d:              d,
			originalState:  originalState,
			originalSchema: originalSchema,
			isEmptyState:   isEmptyState(d),
			dMutex:         sync.RWMutex{},
		}
	}

	return mcc[d.Id()]
}

func getUnexportedField(field reflect.Value) interface{} {
	return reflect.NewAt(field.Type(), unsafe.Pointer(field.UnsafeAddr())).Elem().Interface()
}

func (c *consistencyCheck) compareInterfaceSlices(old, new []interface{}) bool {
	if len(new) == 0 && len(old) == 1 {
		m := old[0].(map[string]interface{})
		if len(m) == 0 {
			return true
		}
	}
	return c.sliceEqual(old, new)
}

func hasRequired(s *schema.Schema) bool {
	if s.Required {
		return true
	}

	//if s.Elem != nil {
	//	if subResource, ok := s.Elem.(*schema.Resource); ok {
	//		for _, v := range subResource.Schema {
	//			if hasRequired(v) {
	//				return true
	//			}
	//		}
	//	}
	//}

	return false
}

func shouldSkip(s *schema.Schema, oldValue, newValue interface{}) bool {
	if s == nil {
		return false
	}
	if s.Computed {
		if isNilOrEmpty(oldValue) && !isNilOrEmpty(newValue) {
			return true
		}
		//h := hasRequired(s)
		//i := isNilOrEmpty(newValue)
		if /*(h && i) ||*/ s.Elem != nil {
			// don't skip

			switch s.Elem.(type) {
			case *schema.Resource:
				//for k, v := range t.Schema {
					//fmt.Println(k, v)
					//s := v.(*schema.Schema)
					oldValueSlice, oOk := oldValue.([]interface{})
					newValueSlice, nOk := newValue.([]interface{})
					if oOk && nOk {
						if len(oldValueSlice) == 0 && len(newValueSlice) == 0 {
							return true
						}
						if len(oldValueSlice) != len(newValueSlice) {
							return false
						}
					}
				//}
			}
		} else {
			//fmt.Println("skipping", s, oldValue, newValue)
			return true
		}
	}

	if s.Set != nil {
		oldValueSet, oOk := oldValue.(*schema.Set)
		newValueSet, nOk := newValue.(*schema.Set)
		if oOk && nOk {
			oldValueList := oldValueSet.List()
			newValueList := newValueSet.List()
			if len(oldValueList) > 0 && len(newValueList) > 0 {
				return s.Set(oldValueList[0]) == s.Set(newValueList[0])
			}
		}
	}

	if s.DiffSuppressFunc != nil {
		oStr, oOk := oldValue.(string)
		nStr, nOk := newValue.(string)
		if oOk && nOk {
			return s.DiffSuppressFunc("", oStr, nStr, nil)
		}
	}

	if s.Elem != nil {
		if subResource, ok := s.Elem.(*schema.Resource); ok {
			if subResource != nil && len(subResource.Schema) > 0 {
				for k, v := range subResource.Schema {
					var ifc interface{}

					switch t := oldValue.(type) {
					case []interface{}:
						if ok && len(t) > 0 {
							if oldValueMap, ok := t[0].(map[string]interface{}); ok {
								ifc = oldValueMap[k]
								newValueSlice, ok := newValue.([]interface{})
								if ok && len(newValueSlice) > 0 {
									if newValueMap, ok := newValueSlice[0].(map[string]interface{}); ok {
										newValue = newValueMap[k]
									}
								}
							}
						}
					case *schema.Set:
						// TODO
						return true
						//continue
					}

					if shouldSkip(v, ifc, newValue) {
						//fmt.Println("calling", v, ifc, newValue)
						return true
					}
				}
			}
		}
	}

	return false
}

func hash(i interface{}) string {
	var b bytes.Buffer
	gob.NewEncoder(&b).Encode(i)
	return b.String()
}

// Equal compares slice a to slice b
func (c *consistencyCheck) sliceEqual(old, new []interface{}) bool {
	if len(old) != len(new) {
		return false
	}
	if len(old) == 0 {
		return true
	}

	// Can't reliably sort slices of map[string]interface{} so need to do a slow compare across both
	return sliceEqualImpl(old, new) && sliceEqualImpl(new, old)
}

func sliceEqualImpl(old, new []interface{}) bool {
	for i := 0; i < len(old); i++ {
		numEqual := 0
		for j := 0; j < len(old); j++ {
			if cmp.Equal(old[i], new[j]) {
				numEqual++
			}
		}
		if numEqual == 0 {
			return false
		}
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

func (c *consistencyCheck) getSchemaElem(key string) *schema.Schema {
	if c.d == nil {
		return nil
	}
	//if c.currentSchema == nil {
	//	schemaInterface := getUnexportedField(reflect.ValueOf(c.d).Elem().FieldByName("schema"))
	//	resourceSchema := schemaInterface.(map[string]*schema.Schema)
	//
	//	return resourceSchema[key]
	//}
	//
	//switch t := c.currentSchema.s.Elem.(type) {
	//case *schema.Resource:
	//	return t.Schema[key]
	//case *schema.Schema:
	//	//fmt.Println(t)
	//	return t
	//}
	//if isNilOrEmpty(c.currentSchema.s.Elem) {
	//	return nil
	//}
	//r := c.currentSchema.s.Elem.(*schema.Resource)
	//resourceSchema := r.Schema

	return nil //resourceSchema[key]
}

//func (c *consistencyCheck) Set(key string, value interface{}) {
//	if c.isEmptyState == nil {
//		panic("consistencyCheck must be initialized with NewConsistencyCheck")
//	}
//
//	c.dMutex.Lock()
//	c.d.Set(key, value)
//	c.dMutex.Unlock()
//}
//

func (c *consistencyCheck) Get(key string) interface{} {
	if c.isEmptyState == nil {
		panic("consistencyCheck must be initialized with NewConsistencyCheck")
	}

	var value interface{}
	c.dMutex.Lock()
	value = c.d.Get(key)
	c.dMutex.Unlock()

	return value
}

func (c *consistencyCheck) GetOk(key string) (interface{}, bool) {
	if c.isEmptyState == nil {
		panic("consistencyCheck must be initialized with NewConsistencyCheck")
	}

	var value interface{}
	var ok bool
	c.dMutex.Lock()
	value, ok = c.d.GetOk(key)
	c.dMutex.Unlock()

	return value, ok
}

func (c *consistencyCheck) compareValues(oldValue, value interface{}) bool {
	switch oldValueType := oldValue.(type) {
	case *schema.Set:
		newValueSet, _ := value.(*schema.Set)
		if !c.compareInterfaceSlices(oldValueType.List(), newValueSet.List()) {
			return false
		}
	case []interface{}:
		switch newValueType := value.(type) {
		case []string:
			oldValueListStr := make([]string, 0)
			for _, oldValueListItem := range oldValueType {
				oldValueListStr = append(oldValueListStr, oldValueListItem.(string))
			}
			if len(sliceDifference(oldValueListStr, newValueType)) != 0 {
				return false
			}
		case []interface{}:
			if !c.compareInterfaceSlices(oldValueType, newValueType) {
				return false
			}
		}
	case map[string]interface{}:
		newValueMap, _ := value.(map[string]interface{})
		for k, v := range newValueMap {
			if !cmp.Equal(v, oldValueType[k]) {
				return false
			}
		}
	default:
		return oldValue == value
	}

	return true
}

func (c *consistencyCheck) SetId(id string) {
	c.dMutex.Lock()
	c.d.SetId(id)
	c.dMutex.Unlock()
}

func (c *consistencyCheck) Id() string {
	c.dMutex.Lock()
	id := c.d.Id()
	c.dMutex.Unlock()

	return id
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
	}
	return &isEmpty
}

func (c *consistencyCheck) CheckErr() *resource.RetryError {
	if c.isEmptyState == nil {
		panic("consistencyCheck must be initialized with NewConsistencyCheck")
	}

	schemaInterface := getUnexportedField(reflect.ValueOf(c.d).Elem().FieldByName("schema"))
	resourceSchema := schemaInterface.(map[string]*schema.Schema)

	//fmt.Println("=======")

	for key := range resourceSchema {
		newValue := c.Get(key)

		oldValue := c.originalState[key]

		//if key == "addresses" && !shouldSkip(c.originalSchema[key], oldValue, newValue) {
		//	fmt.Println("calling shouldSkip on addresses")
		//	shouldSkip(c.originalSchema[key], oldValue, newValue)
		//	shouldSkip(c.originalSchema[key], oldValue, newValue)
		//	shouldSkip(c.originalSchema[key], oldValue, newValue)
		//	shouldSkip(c.originalSchema[key], oldValue, newValue)
		//	shouldSkip(c.originalSchema[key], oldValue, newValue)
		//}

		//fmt.Println("key", key)

		//b := !c.d.HasChange(key)
		//_, okk := c.d.GetOk(key)
		//fmt.Println("b ok", b, okk)
		if *c.isEmptyState || shouldSkip(c.originalSchema[key], oldValue, newValue) {
			//fmt.Println("skipping", key, *c.isEmptyState, b, shouldSkip(c.originalSchema[key], oldValue, newValue))
			continue
		}

		if !c.compareValues(oldValue, newValue) {
			//fmt.Println("retrying because of ", key, oldValue, newValue)
			return resource.RetryableError(&consistencyError{
				key:      key,
				newValue: newValue,
				oldValue: oldValue,
			})
		}
	}

	delete(mcc, c.d.Id())

	return nil
}
