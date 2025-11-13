package util

import (
	"context"
	"fmt"
	"log"
	"reflect"
	"sync"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/mypurecloud/platform-client-sdk-go/v165/platformclientv2"
)

// Attempt to get the home division once during a provider run (Plugin Framework)
var pfDivOnce sync.Once
var pfHomeDivID string
var pfHomeDivName string
var pfHomeDivErr error

// GetHomeDivisionIDPF returns the home division ID for Plugin Framework resources
func GetHomeDivisionIDPF(ctx context.Context) (string, diag.Diagnostics) {
	var diagErr diag.Diagnostics

	pfDivOnce.Do(func() {
		authAPI := platformclientv2.NewAuthorizationApi()
		homeDiv, _, err := authAPI.GetAuthorizationDivisionsHome()
		if err != nil {
			pfHomeDivErr = fmt.Errorf("failed to query home division: %w", err)
			return
		}
		if homeDiv.Id == nil {
			pfHomeDivErr = fmt.Errorf("home division ID is nil")
			return
		}
		pfHomeDivID = *homeDiv.Id
		if homeDiv.Name != nil {
			pfHomeDivName = *homeDiv.Name
		}
	})

	if pfHomeDivErr != nil {
		diagErr.AddError("Failed to get home division", pfHomeDivErr.Error())
		return "", diagErr
	}
	return pfHomeDivID, diagErr
}

// GetHomeDivisionNamePF returns the home division name for Plugin Framework resources
func GetHomeDivisionNamePF(ctx context.Context) (string, diag.Diagnostics) {
	var diagErr diag.Diagnostics

	pfDivOnce.Do(func() {
		authAPI := platformclientv2.NewAuthorizationApi()
		homeDiv, _, err := authAPI.GetAuthorizationDivisionsHome()
		if err != nil {
			pfHomeDivErr = fmt.Errorf("failed to query home division: %w", err)
			return
		}
		if homeDiv.Id == nil {
			pfHomeDivErr = fmt.Errorf("home division ID is nil")
			return
		}
		pfHomeDivID = *homeDiv.Id
		if homeDiv.Name != nil {
			pfHomeDivName = *homeDiv.Name
		}
	})

	if pfHomeDivErr != nil {
		diagErr.AddError("Failed to get home division", pfHomeDivErr.Error())
		return "", diagErr
	}
	return pfHomeDivName, diagErr
}

// UpdateObjectDivisionPF updates an object's division for Plugin Framework resources
// This is the Plugin Framework equivalent of UpdateObjectDivision from SDK v2
//
// Parameters:
//   - ctx: context for the operation
//   - plan: the plan model object (must have DivisionId and Id fields)
//   - state: the state model object (can be nil for create operations)
//   - objType: the object type string (e.g., "user", "queue")
//   - sdkConfig: SDK configuration
//
// Returns Framework diagnostics
func UpdateObjectDivisionPF(ctx context.Context, plan interface{}, state interface{}, objType string, sdkConfig *platformclientv2.Configuration) diag.Diagnostics {
	var diagErr diag.Diagnostics

	// Validate plan is not nil
	if plan == nil {
		diagErr.AddError("Invalid plan object", "Plan object cannot be nil")
		return diagErr
	}

	// Extract values from plan using reflection
	planValue := reflect.ValueOf(plan)
	if planValue.Kind() == reflect.Ptr {
		planValue = planValue.Elem()
	}

	// Extract new division ID from plan
	newDivisionIDField := planValue.FieldByName("DivisionId")
	if !newDivisionIDField.IsValid() {
		diagErr.AddError("Invalid plan object", "Plan object must have a DivisionId field")
		return diagErr
	}
	newDivisionID := newDivisionIDField.Interface().(types.String)

	// Extract object ID from plan
	objectIDField := planValue.FieldByName("Id")
	if !objectIDField.IsValid() {
		diagErr.AddError("Invalid plan object", "Plan object must have an Id field")
		return diagErr
	}
	objectID := objectIDField.Interface().(types.String).ValueString()

	// Extract old division ID from state (if provided)
	var oldDivisionID types.String
	if state != nil {
		stateValue := reflect.ValueOf(state)
		if stateValue.Kind() == reflect.Ptr {
			stateValue = stateValue.Elem()
		}

		oldDivisionIDField := stateValue.FieldByName("DivisionId")
		if oldDivisionIDField.IsValid() {
			oldDivisionID = oldDivisionIDField.Interface().(types.String)
		} else {
			oldDivisionID = types.StringNull()
		}
	} else {
		oldDivisionID = types.StringNull()
	}

	// Check if division_id has changed
	if oldDivisionID.Equal(newDivisionID) {
		return diagErr
	}

	authAPI := platformclientv2.NewAuthorizationApiWithConfig(sdkConfig)
	divisionID := newDivisionID.ValueString()

	// Default to home division if empty
	if divisionID == "" {
		homeDivision, homeDiagErr := GetHomeDivisionIDPF(ctx)
		if homeDiagErr.HasError() {
			return homeDiagErr
		}
		divisionID = homeDivision
	}

	log.Printf("Updating division for %s %s to %s", objType, objectID, divisionID)
	_, divErr := authAPI.PostAuthorizationDivisionObject(divisionID, objType, []string{objectID})
	if divErr != nil {
		diagErr.AddError(
			fmt.Sprintf("Failed to update division for %s %s", objType, objectID),
			divErr.Error(),
		)
		return diagErr
	}

	return diagErr
}
