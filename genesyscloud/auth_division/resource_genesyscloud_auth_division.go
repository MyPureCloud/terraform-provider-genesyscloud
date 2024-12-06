package auth_division

import (
	"context"
	"fmt"
	"log"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"terraform-provider-genesyscloud/genesyscloud/util/constants"
	"terraform-provider-genesyscloud/genesyscloud/util/resourcedata"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"terraform-provider-genesyscloud/genesyscloud/consistency_checker"

	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v146/platformclientv2"
)

func getAllAuthDivisions(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	resources := make(resourceExporter.ResourceIDMetaMap)
	proxy := getAuthDivisionProxy(clientConfig)

	divisions, resp, getErr := proxy.getAllAuthDivision(ctx, "")
	if getErr != nil {
		return nil, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to get divisions | error: %s", getErr), resp)
	}

	for _, division := range *divisions {
		resources[*division.Id] = &resourceExporter.ResourceMeta{BlockLabel: *division.Name}
	}

	return resources, nil
}

func createAuthDivision(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getAuthDivisionProxy(sdkConfig)

	name := d.Get("name").(string)
	description := d.Get("description").(string)
	home := d.Get("home").(bool)

	if home {
		// Home division must already exist, or it cannot be modified
		id, diagErr := util.GetHomeDivisionID()
		if diagErr != nil {
			return diagErr
		}
		d.SetId(id)
		return updateAuthDivision(ctx, d, meta)
	}

	log.Printf("Creating division %s", name)

	division, resp, err := proxy.createAuthDivision(ctx, &platformclientv2.Authzdivision{
		Name:        &name,
		Description: &description,
	})
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to create division %s error: %s", name, err), resp)
	}

	d.SetId(*division.Id)
	log.Printf("Created division %s %s", name, *division.Id)
	return readAuthDivision(ctx, d, meta)
}

func readAuthDivision(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getAuthDivisionProxy(sdkConfig)
	cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceAuthDivision(), constants.ConsistencyChecks(), ResourceType)

	log.Printf("Reading division %s", d.Id())

	return util.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		division, resp, getErr := proxy.getAuthDivisionById(ctx, d.Id(), false, true)
		if getErr != nil {
			if util.IsStatus404(resp) {
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read division %s | error: %s", d.Id(), getErr), resp))
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read division %s | error: %s", d.Id(), getErr), resp))
		}

		_ = d.Set("name", *division.Name)
		resourcedata.SetNillableValue(d, "description", division.Description)
		resourcedata.SetNillableValue(d, "home", division.HomeDivision)

		log.Printf("Read division %s %s", d.Id(), *division.Name)
		return cc.CheckState(d)
	})
}

func updateAuthDivision(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getAuthDivisionProxy(sdkConfig)

	name := d.Get("name").(string)
	description := d.Get("description").(string)

	log.Printf("Updating division %s", name)

	_, resp, err := proxy.updateAuthDivision(ctx, d.Id(), &platformclientv2.Authzdivision{
		Name:        &name,
		Description: &description,
	})
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to update division %s error: %s", name, err), resp)
	}

	log.Printf("Updated division %s", name)
	return readAuthDivision(ctx, d, meta)
}

func deleteAuthDivision(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getAuthDivisionProxy(sdkConfig)

	name := d.Get("name").(string)
	home := d.Get("home").(bool)

	if home {
		// Do not delete home division
		log.Printf("Not deleting home division %s", name)
		return nil
	}

	// Sometimes a division with resources in it priorly still thinks it is attached to those resources during a destroy run.
	// We're retrying again as those resources should detach completely eventually.
	diagErr := util.RetryWhen(util.IsStatus400, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
		log.Printf("Deleting division %s", name)
		resp, err := proxy.deleteAuthDivision(ctx, d.Id(), false)
		if err != nil {
			return resp, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to delete Division %s error: %s", d.Id(), err), resp)
		}
		return resp, nil
	})
	if diagErr != nil {
		return diagErr
	}

	return util.WithRetries(ctx, 180*time.Second, func() *retry.RetryError {
		_, resp, err := proxy.getAuthDivisionById(ctx, d.Id(), false, false)
		if err != nil {
			if util.IsStatus404(resp) {
				// Division deleted
				log.Printf("Deleted division %s", name)
				return nil
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Error deleting division %s | error:: %s", name, err), resp))
		}
		return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Division %s still exists", name), resp))
	})
}

func GenerateAuthDivisionBasic(resourceLabel string, name string) string {
	return GenerateAuthDivisionResource(resourceLabel, name, util.NullValue, util.FalseValue)
}

func GenerateAuthDivisionResource(
	resourceLabel string,
	name string,
	description string,
	home string) string {
	return fmt.Sprintf(`resource "genesyscloud_auth_division" "%s" {
		name = "%s"
		description = %s
		home = %s
	}
	`, resourceLabel, name, description, home)
}
