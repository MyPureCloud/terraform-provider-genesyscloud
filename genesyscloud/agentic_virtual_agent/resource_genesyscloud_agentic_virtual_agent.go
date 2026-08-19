package agentic_virtual_agent

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v195/platformclientv2"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/consistency_checker"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/constants"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/resourcedata"
)

/*
   resource_genesyscloud_agentic_virtual_agent.go contains the CRUD logic for the
   genesyscloud_agentic_virtual_agent resource.
*/

// getAllAgenticVirtualAgents is used by the exporter to retrieve all agents in the org.
func getAllAgenticVirtualAgents(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	resources := make(resourceExporter.ResourceIDMetaMap)
	proxy := getAgenticVirtualAgentProxy(clientConfig)

	log.Printf("Retrieving all Agentic Virtual Agents")

	agents, resp, err := proxy.getAllAgenticVirtualAgents(ctx, "")
	if err != nil {
		return nil, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to get agentic virtual agents: %s", err), resp)
	}

	if agents == nil {
		return resources, nil
	}

	for _, agent := range *agents {
		resources[*agent.Id] = &resourceExporter.ResourceMeta{BlockLabel: *agent.Name}
	}

	log.Printf("Successfully retrieved all Agentic Virtual Agents")
	return resources, nil
}

// createAgenticVirtualAgent creates a new agentic virtual agent.
func createAgenticVirtualAgent(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getAgenticVirtualAgentProxy(sdkConfig)

	name := d.Get("name").(string)

	agentReq := &AgenticVirtualAgentCreate{
		Name: name,
	}

	if v, ok := d.GetOk("image_uri"); ok {
		imageUri := v.(string)
		agentReq.ImageUri = &imageUri
	}

	log.Printf("Creating Agentic Virtual Agent: %s", name)

	agent, resp, err := proxy.createAgenticVirtualAgent(ctx, agentReq)
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to create agentic virtual agent %s: %s", name, err), resp)
	}

	d.SetId(*agent.Id)
	log.Printf("Created Agentic Virtual Agent: %s with ID: %s", name, *agent.Id)

	return readAgenticVirtualAgent(ctx, d, meta)
}

// readAgenticVirtualAgent reads an agentic virtual agent by ID.
func readAgenticVirtualAgent(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getAgenticVirtualAgentProxy(sdkConfig)
	cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceAgenticVirtualAgent(), constants.ConsistencyChecks(), ResourceType)

	log.Printf("Reading Agentic Virtual Agent: %s", d.Id())

	return util.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		agent, resp, err := proxy.getAgenticVirtualAgentById(ctx, d.Id())
		if err != nil {
			if util.IsStatus404(resp) {
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read agentic virtual agent %s: %s", d.Id(), err), resp))
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read agentic virtual agent %s: %s", d.Id(), err), resp))
		}

		resourcedata.SetNillableValue(d, "name", agent.Name)
		resourcedata.SetNillableValue(d, "image_uri", agent.ImageUri)
		resourcedata.SetNillableValue(d, "status", agent.Status)

		// Set latest_saved_version from the nested object
		if agent.LatestSavedVersion != nil && agent.LatestSavedVersion.Version != nil {
			_ = d.Set("latest_saved_version", *agent.LatestSavedVersion.Version)
		} else {
			_ = d.Set("latest_saved_version", "")
		}

		// Set latest_production_ready_version from the nested object
		if agent.LatestProductionReadyVersion != nil && agent.LatestProductionReadyVersion.Version != nil {
			_ = d.Set("latest_production_ready_version", *agent.LatestProductionReadyVersion.Version)
		} else {
			_ = d.Set("latest_production_ready_version", "")
		}

		d.SetId(*agent.Id)

		log.Printf("Read Agentic Virtual Agent: %s", d.Id())
		return cc.CheckState(d)
	})
}

// updateAgenticVirtualAgent updates an existing agentic virtual agent.
func updateAgenticVirtualAgent(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getAgenticVirtualAgentProxy(sdkConfig)

	name := d.Get("name").(string)

	agentReq := &AgenticVirtualAgentUpdate{
		Name: name,
	}

	if v, ok := d.GetOk("image_uri"); ok {
		imageUri := v.(string)
		agentReq.ImageUri = &imageUri
	}

	log.Printf("Updating Agentic Virtual Agent: %s", d.Id())

	_, resp, err := proxy.updateAgenticVirtualAgent(ctx, d.Id(), agentReq)
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to update agentic virtual agent %s: %s", d.Id(), err), resp)
	}

	log.Printf("Updated Agentic Virtual Agent: %s", d.Id())

	return readAgenticVirtualAgent(ctx, d, meta)
}

// deleteAgenticVirtualAgentResource deletes an agentic virtual agent using async job.
func deleteAgenticVirtualAgentResource(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getAgenticVirtualAgentProxy(sdkConfig)

	log.Printf("Deleting Agentic Virtual Agent: %s", d.Id())

	job, resp, err := proxy.deleteAgenticVirtualAgent(ctx, d.Id())
	if err != nil {
		if util.IsStatus404(resp) {
			log.Printf("Agentic Virtual Agent %s already deleted", d.Id())
			return nil
		}
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to delete agentic virtual agent %s: %s", d.Id(), err), resp)
	}

	log.Printf("Delete job for Agentic Virtual Agent %s started with job ID: %s", d.Id(), job.Id)

	// Poll the delete job until it completes
	return util.WithRetries(ctx, 2*time.Minute, func() *retry.RetryError {
		jobStatus, jobResp, jobErr := proxy.getDeleteJobStatus(ctx, d.Id(), job.Id)
		if jobErr != nil {
			if util.IsStatus404(jobResp) {
				// Agent already gone
				log.Printf("Deleted Agentic Virtual Agent: %s", d.Id())
				return nil
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Error checking delete job status for agentic virtual agent %s: %s", d.Id(), jobErr), jobResp))
		}

		switch jobStatus.Status {
		case "InProgress":
			return retry.RetryableError(fmt.Errorf("delete job for agentic virtual agent %s still in progress", d.Id()))
		case "Succeeded":
			log.Printf("Deleted Agentic Virtual Agent: %s", d.Id())
			return nil
		case "Failed":
			if len(jobStatus.Errors) > 0 && jobStatus.Errors[0].Code == "ava.dependency.exists" {
				return retry.NonRetryableError(fmt.Errorf("cannot delete agentic virtual agent %s: it is referenced by a published Architect bot flow. Remove the agent from Architect flows first", d.Id()))
			}
			if len(jobStatus.Errors) > 0 && jobStatus.Errors[0].Message != "" {
				return retry.NonRetryableError(fmt.Errorf("delete job failed for agentic virtual agent %s: %s", d.Id(), jobStatus.Errors[0].Message))
			}
			return retry.NonRetryableError(fmt.Errorf("delete job failed for agentic virtual agent %s", d.Id()))
		default:
			return retry.RetryableError(fmt.Errorf("unexpected job status for agentic virtual agent %s: %s", d.Id(), jobStatus.Status))
		}
	})
}
