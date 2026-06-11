package case_management_caseplan

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v191/platformclientv2"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/resourcedata"
)

// getCaseManagementCaseplanCreateFromResourceData maps ResourceData to Caseplancreate for POST /caseplans.
func getCaseManagementCaseplanCreateFromResourceData(d *schema.ResourceData) platformclientv2.Caseplancreate {
	c := platformclientv2.Caseplancreate{}
	c.Name = platformclientv2.String(d.Get("name").(string))
	if v, ok := d.GetOk("division_id"); ok && v.(string) != "" {
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
	if intake := expandCaseplanIntakeSettingsForCreate(d); intake != nil {
		c.IntakeSettings = intake
	}
	return c
}

// buildCaseplanPatchFromResourceData builds Caseplanupdate for PATCH /caseplans/{id}. Only fields with HasChange are set (SDK JSON uses SetFieldNames).
func buildCaseplanPatchFromResourceData(d *schema.ResourceData) (*platformclientv2.Caseplanupdate, bool) {
	patch := &platformclientv2.Caseplanupdate{}
	has := false

	if d.HasChange("name") {
		patch.SetField("Name", platformclientv2.String(d.Get("name").(string)))
		has = true
	}
	if d.HasChange("division_id") {
		div := d.Get("division_id").(string)
		if div == "" {
			patch.SetField("DivisionId", platformclientv2.String("*"))
		} else {
			patch.SetField("DivisionId", platformclientv2.String(div))
		}
		has = true
	}
	if d.HasChange("description") {
		patch.SetField("Description", platformclientv2.String(d.Get("description").(string)))
		has = true
	}
	if d.HasChange("reference_prefix") {
		patch.SetField("ReferencePrefix", platformclientv2.String(d.Get("reference_prefix").(string)))
		has = true
	}
	if d.HasChange("default_due_duration_in_seconds") {
		patch.SetField("DefaultDueDurationInSeconds", platformclientv2.Int(d.Get("default_due_duration_in_seconds").(int)))
		has = true
	}
	if d.HasChange("default_ttl_seconds") {
		patch.SetField("DefaultTtlSeconds", platformclientv2.Int(d.Get("default_ttl_seconds").(int)))
		has = true
	}
	if d.HasChange("default_case_owner") {
		uid := firstMapString(d.Get("default_case_owner").([]interface{}), "id")
		if uid != "" {
			patch.SetField("DefaultCaseOwnerId", platformclientv2.String(uid))
		} else {
			patch.SetField("DefaultCaseOwnerId", nil)
		}
		has = true
	}
	if d.HasChange("customer_intent") {
		cid := firstMapString(d.Get("customer_intent").([]interface{}), "id")
		if cid != "" {
			patch.SetField("CustomerIntentId", platformclientv2.String(cid))
		} else {
			patch.SetField("CustomerIntentId", nil)
		}
		has = true
	}

	if !has {
		return nil, false
	}
	return patch, true
}

func expandCaseplanIntakeSettingsForCreate(d *schema.ResourceData) *[]platformclientv2.Intakesetting {
	raw := d.Get("intake_settings").([]interface{})
	if len(raw) == 0 {
		return nil
	}
	out := expandCaseplanIntakeSettingsSlice(raw)
	return &out
}

// expandCaseplanIntakeSettingsForPut builds the slice for PUT .../intakesettings (empty list clears settings).
func expandCaseplanIntakeSettingsForPut(d *schema.ResourceData) *[]platformclientv2.Intakesetting {
	raw := d.Get("intake_settings").([]interface{})
	out := expandCaseplanIntakeSettingsSlice(raw)
	return &out
}

func expandCaseplanIntakeSettingsSlice(raw []interface{}) []platformclientv2.Intakesetting {
	out := make([]platformclientv2.Intakesetting, 0, len(raw))
	for _, item := range raw {
		m, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		var row platformclientv2.Intakesetting
		if prop, ok := m["property"].(string); ok && prop != "" {
			row.Property = platformclientv2.String(prop)
		}
		if v, ok := m["required"].(bool); ok {
			row.Required = platformclientv2.Bool(v)
		}
		if v, ok := m["display_order"].(int); ok {
			row.DisplayOrder = platformclientv2.Int(v)
		}
		out = append(out, row)
	}
	return out
}

func flattenCaseplanIntakeSettings(entities *[]platformclientv2.Intakesetting) []interface{} {
	if entities == nil || len(*entities) == 0 {
		return []interface{}{}
	}
	out := make([]interface{}, 0, len(*entities))
	for i := range *entities {
		s := &(*entities)[i]
		m := make(map[string]interface{})
		if s.Property != nil {
			m["property"] = *s.Property
		}
		// Always set required/display_order so state matches config and the consistency checker
		// does not see schema defaults (false/0) when the API omits nil pointers.
		required := false
		if s.Required != nil {
			required = *s.Required
		}
		m["required"] = required
		displayOrder := 0
		if s.DisplayOrder != nil {
			displayOrder = *s.DisplayOrder
		}
		m["display_order"] = displayOrder
		out = append(out, m)
	}
	return out
}

func expandCaseplanDataSchemas(d *schema.ResourceData) *[]platformclientv2.Caseplandataschema {
	raw := d.Get("data_schema").([]interface{})
	out := caseplanDataSchemasFromResourceList(raw)
	if len(out) == 0 {
		return nil
	}
	return &out
}

func caseplanDataSchemaIDSetFromRaw(raw []interface{}) map[string]struct{} {
	out := make(map[string]struct{})
	for _, row := range caseplanDataSchemasFromResourceList(raw) {
		if row.Id != nil && *row.Id != "" {
			out[*row.Id] = struct{}{}
		}
	}
	return out
}

func caseplanDataSchemasFromResourceList(raw []interface{}) []platformclientv2.Caseplandataschema {
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
		out = append(out, row)
	}
	return out
}

// caseplanDataSchemaSyncPlanFromState returns schema IDs removed from config (deleteIDs) and rows to bind (puts).
// Only id is tracked: new ids get a POST after delete; re-binding the same id after a delete uses the fallback in execCaseplanDataSchemaSync.
func caseplanDataSchemaSyncPlanFromState(oldRaw, newRaw []interface{}) (deleteIDs []string, puts []platformclientv2.Caseplandataschema) {
	oldIDs := caseplanDataSchemaIDSetFromRaw(oldRaw)
	newRows := caseplanDataSchemasFromResourceList(newRaw)
	newIDSet := make(map[string]struct{})
	for _, row := range newRows {
		if row.Id != nil && *row.Id != "" {
			newIDSet[*row.Id] = struct{}{}
		}
	}
	for id := range oldIDs {
		if _, ok := newIDSet[id]; !ok {
			deleteIDs = append(deleteIDs, id)
		}
	}
	for _, row := range newRows {
		if row.Id == nil || *row.Id == "" {
			continue
		}
		id := *row.Id
		if _, had := oldIDs[id]; !had {
			puts = append(puts, platformclientv2.Caseplandataschema{Id: platformclientv2.String(id)})
		}
	}
	return deleteIDs, puts
}

func flattenCaseplanDataSchemas(schemas *[]platformclientv2.Caseplandataschema) []interface{} {
	if schemas == nil || len(*schemas) == 0 {
		return nil
	}
	out := make([]interface{}, 0, len(*schemas))
	for i := range *schemas {
		s := &(*schemas)[i]
		m := make(map[string]interface{})
		if s.Id != nil {
			m["id"] = *s.Id
		}
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
