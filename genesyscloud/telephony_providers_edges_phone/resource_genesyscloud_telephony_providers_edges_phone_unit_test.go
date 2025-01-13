package telephony_providers_edges_phone

import (
	"reflect"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

/* Tests the GetLineProperties function to ensure that the NIL values are checked*/
func TestUnitGetLineProperties(t *testing.T) {
	tests := []struct {
		name           string
		resourceData   *schema.ResourceData
		wantLineAddr   *[]interface{}
		wantRemoteAddr *[]interface{}
	}{

		{
			name:           "empty_resource_data",
			resourceData:   schema.TestResourceDataRaw(t, map[string]*schema.Schema{}, map[string]interface{}{}),
			wantLineAddr:   &[]interface{}{},
			wantRemoteAddr: &[]interface{}{},
		},
		{
			name: "valid_line_properties",
			resourceData: schema.TestResourceDataRaw(t, map[string]*schema.Schema{
				"line_properties": {
					Type:     schema.TypeList,
					Optional: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"line_address": {
								Type:     schema.TypeList,
								Optional: true,
								Elem:     &schema.Schema{Type: schema.TypeString},
							},
							"remote_address": {
								Type:     schema.TypeList,
								Optional: true,
								Elem:     &schema.Schema{Type: schema.TypeString},
							},
						},
					},
				},
			}, map[string]interface{}{
				"line_properties": []interface{}{
					map[string]interface{}{
						"line_address":   []interface{}{"192.168.1.1"},
						"remote_address": []interface{}{"10.0.0.1"},
					},
				},
			}),
			wantLineAddr:   &[]interface{}{"192.168.1.1"},
			wantRemoteAddr: &[]interface{}{"10.0.0.1"},
		},
		{
			name: "empty_line_properties",
			resourceData: schema.TestResourceDataRaw(t, map[string]*schema.Schema{
				"line_properties": {
					Type:     schema.TypeList,
					Optional: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"line_address": {
								Type:     schema.TypeList,
								Optional: true,
								Elem:     &schema.Schema{Type: schema.TypeString},
							},
							"remote_address": {
								Type:     schema.TypeList,
								Optional: true,
								Elem:     &schema.Schema{Type: schema.TypeString},
							},
						},
					},
				},
			}, map[string]interface{}{
				"line_properties": []interface{}{
					map[string]interface{}{},
				},
			}),
			wantLineAddr:   &[]interface{}{},
			wantRemoteAddr: &[]interface{}{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotLineAddr, gotRemoteAddr := getLineProperties(tt.resourceData)

			// If not nil, compare the actual values
			if gotLineAddr != nil {
				if !reflect.DeepEqual(*gotLineAddr, *tt.wantLineAddr) {
					t.Errorf("getLineProperties() gotLineAddr = %v, want %v", *gotLineAddr, *tt.wantLineAddr)
				}
			}
			if gotRemoteAddr != nil {
				if !reflect.DeepEqual(*gotRemoteAddr, *tt.wantRemoteAddr) {
					t.Errorf("getLineProperties() gotRemoteAddr = %v, want %v", *gotRemoteAddr, *tt.wantRemoteAddr)
				}
			}
		})
	}
}
