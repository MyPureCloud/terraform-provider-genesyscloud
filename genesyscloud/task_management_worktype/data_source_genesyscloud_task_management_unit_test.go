package task_management_worktype

import (
	"context"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v131/platformclientv2"
	"github.com/stretchr/testify/assert"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"testing"
)

// TestUnitataSourceTaskManagementWorktypeStatus tests the retrieve of a status id.
// I am writing a unit test here as we already have the basic test coverage in another test.
func TestUnitDataSourceTaskManagementWorktypeStatus(t *testing.T) {
	name := "Insurance Claim-Test"
	defaultTtlSeconds := 2678400
	defaultPriority := 2000
	workTypeId := uuid.NewString()
	approvedStatusId := uuid.NewString()
	approvedName := "Approved"
	approvedCategory := "Closed"
	rejectedStatusId := uuid.NewString()
	rejectedName := "Rejected"
	rejectedCategory := "Closed"

	approvedStatus := &platformclientv2.Workitemstatus{
		Id:       &approvedStatusId,
		Name:     &approvedName,
		Category: &approvedCategory,
	}

	rejectedStatus := &platformclientv2.Workitemstatus{
		Id:       &rejectedStatusId,
		Name:     &rejectedName,
		Category: &rejectedCategory,
	}

	statuses := &[]platformclientv2.Workitemstatus{*approvedStatus, *rejectedStatus}
	workType := &platformclientv2.Worktype{
		Id:                &workTypeId,
		Name:              &name,
		DefaultTtlSeconds: &defaultTtlSeconds,
		DefaultPriority:   &defaultPriority,
		Statuses:          statuses,
	}

	workTypeProxy := &taskManagementWorktypeProxy{}
	workTypeProxy.getTaskManagementWorktypeByNameAttr = func(ctx context.Context, proxy *taskManagementWorktypeProxy, name string) (workItemType *platformclientv2.Worktype, retryable bool, resp *platformclientv2.APIResponse, err error) {
		return workType, false, resp, nil
	}

	internalProxy = workTypeProxy
	defer func() { internalProxy = nil }()

	ctx := context.Background()
	gcloud := &provider.ProviderMeta{ClientConfig: &platformclientv2.Configuration{}}

	//Grab our defined schema
	schemaDS := DataSourceTaskManagementWorktypeStatus().Schema

	//Setup a map of values
	dsDataMap := map[string]interface{}{
		"worktype_name":        *workType.Name,
		"worktype_status_name": *approvedStatus.Name,
	}

	d := schema.TestResourceDataRaw(t, schemaDS, dsDataMap)

	diag := dataSourceTaskManagementWorktypeStatusRead(ctx, d, gcloud)
	assert.Equal(t, false, diag.HasError())
	assert.Equal(t, approvedStatusId, d.Id())
}
