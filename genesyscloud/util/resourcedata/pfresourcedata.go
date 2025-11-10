package resourcedata

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/mypurecloud/platform-client-sdk-go/v165/platformclientv2"
)

// Plugin Framework equivalents for resourcedata utility functions

// SetPFNillableValueString sets a *string value to a types.String field
// PF equivalent of SetPFNillableValue for string types
// Usage: SetPFNillableValueString(&state.Name, currentUser.Name)
func SetPFNillableValueString(target *types.String, value *string) {
	if value != nil {
		*target = types.StringValue(*value)
	} else {
		*target = types.StringNull()
	}
}

// SetPFNillableValueInt64 sets a *int value to a types.Int64 field
// PF equivalent of SetPFNillableValue for int types
// Usage: SetPFNillableValueInt64(&state.Age, currentUser.Age)
func SetPFNillableValueInt64(target *types.Int64, value *int) {
	if value != nil {
		*target = types.Int64Value(int64(*value))
	} else {
		*target = types.Int64Null()
	}
}

// SetPFNillableValueBool sets a *bool value to a types.Bool field
// PF equivalent of SetPFNillableValue for bool types
// Usage: SetPFNillableValueBool(&state.Active, currentUser.Active)
func SetPFNillableValueBool(target *types.Bool, value *bool) {
	if value != nil {
		*target = types.BoolValue(*value)
	} else {
		*target = types.BoolNull()
	}
}

// SetPFNillableValueFloat64 sets a *float64 value to a types.Float64 field
// PF equivalent of SetPFNillableValue for float64 types
// Usage: SetPFNillableValueFloat64(&state.Score, currentUser.Score)
func SetPFNillableValueFloat64(target *types.Float64, value *float64) {
	if value != nil {
		*target = types.Float64Value(*value)
	} else {
		*target = types.Float64Null()
	}
}

// SetPFNillableReference sets a Domainentityref ID to a types.String field
// PF equivalent of SetPFNillableReference for Domainentityref types
// Usage: SetPFNillableReference(&state.DivisionId, currentUser.Division)
func SetPFNillableReference(target *types.String, value *platformclientv2.Domainentityref) {
	if value != nil && value.Id != nil {
		*target = types.StringValue(*value.Id)
	} else {
		*target = types.StringNull()
	}
}

// SetPFNillableReferenceWritableDivision sets a Writabledivision ID to a types.String field
// PF equivalent of SetPFNillableReferenceWritableDivision
// Usage: SetPFNillableReferenceWritableDivision(&state.DivisionId, division)
func SetPFNillableReferenceWritableDivision(target *types.String, value *platformclientv2.Writabledivision) {
	if value != nil && value.Id != nil {
		*target = types.StringValue(*value.Id)
	} else {
		*target = types.StringNull()
	}
}

// SetPFNillableReferenceDivision sets a Division ID to a types.String field
// PF equivalent of SetPFNillableReferenceDivision
// Usage: SetPFNillableReferenceDivision(&state.DivisionId, division)
func SetPFNillableReferenceDivision(target *types.String, value *platformclientv2.Division) {
	if value != nil && value.Id != nil {
		*target = types.StringValue(*value.Id)
	} else {
		*target = types.StringNull()
	}
}
