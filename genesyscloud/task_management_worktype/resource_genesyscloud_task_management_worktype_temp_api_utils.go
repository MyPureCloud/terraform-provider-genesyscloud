package task_management_worktype

import (
	"github.com/mypurecloud/platform-client-sdk-go/v119/platformclientv2"
)

/*
This file contains temporary structs and methods to represent the following models from the Genesys Cloud API:
- Worktype
- Workitemstatus
- Workitemstatuscreate
- Workitemstatusupdate

TODO: Once https://inindca.atlassian.net/browse/WORKITEMS-2193 is fixed, then refactor
the resource implementation to use official SDK structs and methods.
This file should be deleted and none of the items in here should be needed at that time.

NOTE: The only difference between these and the official models is the status' `StatusTransitionTime`
property and some of the status properties with no 'omitempty's. These allows them to be passed on the API call
with their JSON null values. This is a reminder for refactorting that these properties need to be set
using the SDK model's SetField so it can retain empty values.
*/

type Worktype struct {
	Id                        *string                                   `json:"id,omitempty"`
	Name                      *string                                   `json:"name,omitempty"`
	Division                  *platformclientv2.Division                `json:"division,omitempty"`
	Description               *string                                   `json:"description,omitempty"`
	DefaultWorkbin            *platformclientv2.Workbinreference        `json:"defaultWorkbin,omitempty"`
	DefaultStatus             *platformclientv2.Workitemstatusreference `json:"defaultStatus,omitempty"`
	Statuses                  *[]Workitemstatus                         `json:"statuses,omitempty"`
	DefaultDurationSeconds    *int                                      `json:"defaultDurationSeconds,omitempty"`
	DefaultExpirationSeconds  *int                                      `json:"defaultExpirationSeconds,omitempty"`
	DefaultDueDurationSeconds *int                                      `json:"defaultDueDurationSeconds,omitempty"`
	DefaultPriority           *int                                      `json:"defaultPriority,omitempty"`
	DefaultLanguage           *platformclientv2.Languagereference       `json:"defaultLanguage,omitempty"`
	DefaultTtlSeconds         *int                                      `json:"defaultTtlSeconds,omitempty"`
	ModifiedBy                *platformclientv2.Userreference           `json:"modifiedBy,omitempty"`
	DefaultQueue              *platformclientv2.Queuereference          `json:"defaultQueue,omitempty"`
	DefaultSkills             *[]platformclientv2.Routingskillreference `json:"defaultSkills,omitempty"`
	AssignmentEnabled         *bool                                     `json:"assignmentEnabled,omitempty"`
	Schema                    *platformclientv2.Workitemschema          `json:"schema,omitempty"`
}

type Workitemstatus struct {
	Id                           *string                                     `json:"id,omitempty"`
	Name                         *string                                     `json:"name,omitempty"`
	Category                     *string                                     `json:"category,omitempty"`
	DestinationStatuses          *[]platformclientv2.Workitemstatusreference `json:"destinationStatuses"`
	Description                  *string                                     `json:"description,omitempty"`
	DefaultDestinationStatus     *platformclientv2.Workitemstatusreference   `json:"defaultDestinationStatus"`
	StatusTransitionDelaySeconds *int                                        `json:"statusTransitionDelaySeconds"`
	StatusTransitionTime         *string                                     `json:"statusTransitionTime"`
	Worktype                     *platformclientv2.Worktypereference         `json:"worktype,omitempty"`
}

type Workitemstatuscreate struct {
	Name                         *string   `json:"name,omitempty"`
	Category                     *string   `json:"category,omitempty"`
	DestinationStatusIds         *[]string `json:"destinationStatusIds"`
	Description                  *string   `json:"description,omitempty"`
	DefaultDestinationStatusId   *string   `json:"defaultDestinationStatusId"`
	StatusTransitionDelaySeconds *int      `json:"statusTransitionDelaySeconds"`
	StatusTransitionTime         *string   `json:"statusTransitionTime"`
}

type Workitemstatusupdate struct {
	Name                         *string   `json:"name,omitempty"`
	DestinationStatusIds         *[]string `json:"destinationStatusIds"`
	Description                  *string   `json:"description,omitempty"`
	DefaultDestinationStatusId   *string   `json:"defaultDestinationStatusId"`
	StatusTransitionDelaySeconds *int      `json:"statusTransitionDelaySeconds"`
	StatusTransitionTime         *string   `json:"statusTransitionTime"`
}
