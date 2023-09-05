package process_automation_trigger

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	gcloud "terraform-provider-genesyscloud/genesyscloud"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v105/platformclientv2"
)

type ProcessAutomationTriggers struct {
	Entities *[]ProcessAutomationTrigger `json:"entities,omitempty"`
	NextUri  *string                     `json:"nextUri,omitempty"`
}

func dataSourceProcessAutomationTrigger() *schema.Resource {
	return &schema.Resource{
		Description: "Data source for Genesys Cloud process automation trigger. Select a trigger by name",
		ReadContext: gcloud.ReadWithPooledClient(dataSourceProcessAutomationTriggerRead),
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "The name of the trigger",
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}
}

func dataSourceProcessAutomationTriggerRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	sdkConfig := m.(*gcloud.ProviderMeta).ClientConfig
	integrationAPI := platformclientv2.NewIntegrationsApiWithConfig(sdkConfig)

	triggerName := d.Get("name").(string)

	return gcloud.WithRetries(ctx, 15*time.Second, func() *retry.RetryError {
		// create path
		path := integrationAPI.Configuration.BasePath + "/api/v2/processAutomation/triggers"

		for pageNum := 1; ; pageNum++ {
			processAutomationTriggers, _, getErr := getAllProcessAutomationTriggers(path, integrationAPI)

			if getErr != nil {
				return retry.NonRetryableError(fmt.Errorf("failed to get page of process automation triggers: %s", getErr))
			}

			if processAutomationTriggers.Entities == nil || len(*processAutomationTriggers.Entities) == 0 {
				return retry.RetryableError(fmt.Errorf("no process automation triggers found with name: %s", triggerName))
			}

			for _, trigger := range *processAutomationTriggers.Entities {
				if trigger.Name != nil && *trigger.Name == triggerName {
					d.SetId(*trigger.Id)
					return nil
				}
			}

			if processAutomationTriggers.NextUri == nil {
				return retry.NonRetryableError(fmt.Errorf("no process automation triggers found with name: %s", getErr))
			}

			path = integrationAPI.Configuration.BasePath + *processAutomationTriggers.NextUri
		}
	})
}

func getAllProcessAutomationTriggers(path string, api *platformclientv2.IntegrationsApi) (*ProcessAutomationTriggers, *platformclientv2.APIResponse, error) {
	apiClient := &api.Configuration.APIClient

	headerParams := make(map[string]string)

	// oauth required
	if api.Configuration.AccessToken != "" {
		headerParams["Authorization"] = "Bearer " + api.Configuration.AccessToken
	}
	// add default headers if any
	for key := range api.Configuration.DefaultHeader {
		headerParams[key] = api.Configuration.DefaultHeader[key]
	}

	headerParams["Content-Type"] = "application/json"
	headerParams["Accept"] = "application/json"

	var successPayload *ProcessAutomationTriggers
	response, err := apiClient.CallAPI(path, http.MethodGet, nil, headerParams, nil, nil, "", nil)
	if err != nil {
		// Nothing special to do here, but do avoid processing the response
	} else if response.Error != nil {
		err = errors.New(response.ErrorMessage)
	} else {
		err = json.Unmarshal([]byte(response.RawBody), &successPayload)
	}

	return successPayload, response, err
}
