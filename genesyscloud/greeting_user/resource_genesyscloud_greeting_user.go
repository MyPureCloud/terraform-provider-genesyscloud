package greeting_user

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v176/platformclientv2"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/consistency_checker"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/constants"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/resourcedata"

	resourceExporter "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_exporter"
)

func getAllGreetings(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	resources := make(resourceExporter.ResourceIDMetaMap)
	proxy := getGreeetingProxy(clientConfig)

	greetings, resp, getErr := proxy.getAllGreetings(ctx)
	if getErr != nil {
		return nil, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to get page of greetings error: %s", getErr), resp)
	}
	if greetings != nil {
		for _, greeting := range *greetings {
			if greeting.Id != nil && greeting.Name != nil {
				resources[*greeting.Id] = &resourceExporter.ResourceMeta{BlockLabel: *greeting.Name}
			}
		}
	}
	return resources, nil
}

func createUserGreeting(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getGreeetingProxy(sdkConfig)

	greetingReq := getUserGreetingFromResourceData(d)

	log.Printf("Creating greeting")
	greeting, resp, err := proxy.createUserGreeting(ctx, &greetingReq)
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to create greeting error: %s", err), resp)
	}

	d.SetId(*greeting.Id)

	log.Printf("Created greeting %s", *greeting.Id)
	return readUserGreeting(ctx, d, meta)
}

func readUserGreeting(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getGreeetingProxy(sdkConfig)
	cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceGreeting(), constants.ConsistencyChecks(), ResourceType)

	log.Printf("Reading greeting %s", d.Id())
	return util.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		greeting, resp, getErr := proxy.getUserGreetingById(ctx, d.Id())
		if getErr != nil {
			if util.IsStatus404(resp) {
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read greeting %s | error: %s", d.Id(), getErr), resp))
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read greeting %s | error: %s", d.Id(), getErr), resp))
		}

		resourcedata.SetNillableValue(d, "name", greeting.Name)
		resourcedata.SetNillableValue(d, "type", greeting.VarType)
		resourcedata.SetNillableValue(d, "owner_type", greeting.OwnerType)
		resourcedata.SetNillableValue(d, "audio_tts", greeting.AudioTTS)
		resourcedata.SetNillableValue(d, "user_id", greeting.Owner.Id)
		resourcedata.SetNillableValue(d, "audio_file", greeting.AudioFile)

		log.Printf("Read greeting %s", d.Id())
		return cc.CheckState(d)
	})
}

func updateUserGreeting(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getGreeetingProxy(sdkConfig)

	greetingReq := getUserGreetingFromResourceData(d)

	log.Printf("Updating greeting")
	greeting, resp, err := proxy.updateUserGreeting(ctx, d.Id(), &greetingReq)
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to update greeting error: %s", err), resp)
	}

	if greeting != nil && greeting.Id != nil {
		log.Printf("Updated greeting %s", *greeting.Id)
	}
	return readUserGreeting(ctx, d, meta)
}

func deleteUserGreeting(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getGreeetingProxy(sdkConfig)

	log.Printf("Deleting greeting %s", d.Id())

	resp, err := proxy.deleteUserGreeting(ctx, d.Id())
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to delete Greeting %s error: %s", d.Id(), err), resp)
	}

	return util.WithRetries(ctx, 30*time.Second, func() *retry.RetryError {
		_, resp, err := proxy.getUserGreetingById(ctx, d.Id())
		if err != nil {
			if util.IsStatus404(resp) {
				// Greeting deleted
				log.Printf("Deleted greeting %s", d.Id())
				return nil
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Error deleting greeting %s | error: %s", d.Id(), err), resp))
		}

		return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Greeting %s still exists", d.Id()), resp))
	})
}
