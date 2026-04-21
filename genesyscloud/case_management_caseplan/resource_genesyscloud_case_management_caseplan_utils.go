package case_management_caseplan

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v186/platformclientv2"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/resourcedata"
)

// getCaseManagementCaseplanCreateFromResourceData maps ResourceData to Caseplancreate for POST /caseplans.
func getCaseManagementCaseplanCreateFromResourceData(d *schema.ResourceData) platformclientv2.Caseplancreate {
	c := platformclientv2.Caseplancreate{}
	if v, ok := d.GetOk("name"); ok {
		c.Name = platformclientv2.String(v.(string))
	}
	if v, ok := d.GetOk("division_id"); ok {
		c.DivisionId = platformclientv2.String(v.(string))
	}
	if v, ok := d.GetOk("description"); ok {
		c.Description = platformclientv2.String(v.(string))
	}
	if v, ok := d.GetOk("reference_prefix"); ok {
		c.ReferencePrefix = platformclientv2.String(v.(string))
	}
	if v, ok := d.GetOk("default_due_duration_in_seconds"); ok {
		c.DefaultDueDurationInSeconds = platformclientv2.Int(v.(int))
	}
	if v, ok := d.GetOk("default_ttl_seconds"); ok {
		c.DefaultTtlSeconds = platformclientv2.Int(v.(int))
	}
	if v, ok := d.GetOk("default_case_owner"); ok {
		if uid := firstMapString(v.([]interface{}), "id"); uid != "" {
			c.DefaultCaseOwnerId = platformclientv2.String(uid)
		}
	}
	if v, ok := d.GetOk("customer_intent"); ok {
		if cid := firstMapString(v.([]interface{}), "id"); cid != "" {
			c.CustomerIntentId = platformclientv2.String(cid)
		}
	}
	if schemas := expandCaseplanDataSchemas(d); schemas != nil {
		c.DataSchemas = schemas
	}
	return c
}

func expandCaseplanDataSchemas(d *schema.ResourceData) *[]platformclientv2.Caseplandataschema {
	raw := d.Get("data_schema").([]interface{})
	if len(raw) == 0 {
		return nil
	}
	out := make([]platformclientv2.Caseplandataschema, 0, len(raw))
	for _, item := range raw {
		m, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		var row platformclientv2.Caseplandataschema
		if id, ok := m["id"].(string); ok && id != "" {
			row.Id = platformclientv2.String(id)
		}
		if v, ok := m["version"].(int); ok {
			row.Version = platformclientv2.Int(v)
		}
		out = append(out, row)
	}
	return &out
}

func flattenCaseplanDataSchemas(schemas *[]platformclientv2.Caseplandataschema) []interface{} {
	if schemas == nil || len(*schemas) == 0 {
		return nil
	}
	out := make([]interface{}, 0, len(*schemas))
	for i := range *schemas {
		s := &(*schemas)[i]
		m := make(map[string]interface{})
		resourcedata.SetMapValueIfNotNil(m, "id", s.Id)
		resourcedata.SetMapValueIfNotNil(m, "version", s.Version)
		out = append(out, m)
	}
	return out
}

// caseplanVersionForDataschemaRead returns the caseplan version id string for GET .../versions/{versionId}/dataschemas.
func caseplanVersionForDataschemaRead(cp *platformclientv2.Caseplan) string {
	if cp == nil {
		return ""
	}
	if cp.Latest != nil {
		return fmt.Sprintf("%d", *cp.Latest)
	}
	if cp.Published != nil {
		return fmt.Sprintf("%d", *cp.Published)
	}
	return ""
}

func firstMapString(blocks []interface{}, key string) string {
	for _, raw := range blocks {
		m, ok := raw.(map[string]interface{})
		if !ok {
			continue
		}
		if v, ok := m[key].(string); ok && v != "" {
			return v
		}
	}
	return ""
}

func flattenUserReference(ref *platformclientv2.Userreference) []interface{} {
	if ref == nil {
		return nil
	}
	m := make(map[string]interface{})
	resourcedata.SetMapValueIfNotNil(m, "id", ref.Id)
	return []interface{}{m}
}

func flattenCustomerIntentReference(ref *platformclientv2.Customerintentreference) []interface{} {
	if ref == nil {
		return nil
	}
	m := make(map[string]interface{})
	resourcedata.SetMapValueIfNotNil(m, "id", ref.Id)
	return []interface{}{m}
}
