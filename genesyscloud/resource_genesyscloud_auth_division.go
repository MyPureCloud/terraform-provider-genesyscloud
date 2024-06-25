package genesyscloud

import (
	"context"
	"fmt"
	"log"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"terraform-provider-genesyscloud/genesyscloud/util/constants"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"terraform-provider-genesyscloud/genesyscloud/consistency_checker"

	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"
)

func getAllAuthDivisions(_ context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	resources := make(resourceExporter.ResourceIDMetaMap)
	authAPI := platformclientv2.NewAuthorizationApiWithConfig(clientConfig)

	for pageNum := 1; ; pageNum++ {
		const pageSize = 100
		divisions, resp, getErr := authAPI.GetAuthorizationDivisions(pageSize, pageNum, "", nil, "", "", false, nil, "")
		if getErr != nil {
			return nil, util.BuildAPIDiagnosticError("genesyscloud_auth_division", fmt.Sprintf("Failed to get page of divisions error: %s", getErr), resp)
		}

		if divisions.Entities == nil || len(*divisions.Entities) == 0 {
			break
		}

		for _, division := range *divisions.Entities {
			resources[*division.Id] = &resourceExporter.ResourceMeta{Name: *division.Name}
		}
	}

	return resources, nil
}

func AuthDivisionExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: provider.GetAllWithPooledClient(getAllAuthDivisions),
		RefAttrs:         map[string]*resourceExporter.RefAttrSettings{}, // No references
	}
}

func ResourceAuthDivision() *schema.Resource {
	return &schema.Resource{
		Description: "Genesys Cloud Authorization Division",

		CreateContext: provider.CreateWithPooledClient(createAuthDivision),
		ReadContext:   provider.ReadWithPooledClient(readAuthDivision),
		UpdateContext: provider.UpdateWithPooledClient(updateAuthDivision),
		DeleteContext: provider.DeleteWithPooledClient(deleteAuthDivision),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "Division name.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"description": {
				Description: "Division description.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"home": {
				Description: "True if this is the home division. This can be set to manage the pre-existing home division.  Note: If name attribute is changed, this will cause the auth_division to be dropped and recreated. This will generate a new ID the division.  Existing objects with the old division will not be migrated to the new division",
				Type:        schema.TypeBool,
				Optional:    true,
				ForceNew:    true,
			},
		},
	}
}

func createAuthDivision(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)
	description := d.Get("description").(string)
	home := d.Get("home").(bool)

	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	authAPI := platformclientv2.NewAuthorizationApiWithConfig(sdkConfig)

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
	division, resp, err := authAPI.PostAuthorizationDivisions(platformclientv2.Authzdivision{
		Name:        &name,
		Description: &description,
	})
	if err != nil {
		return util.BuildAPIDiagnosticError("genesyscloud_auth_division", fmt.Sprintf("Failed to create division %s error: %s", name, err), resp)
	}

	d.SetId(*division.Id)
	log.Printf("Created division %s %s", name, *division.Id)
	return readAuthDivision(ctx, d, meta)
}

func readAuthDivision(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	authAPI := platformclientv2.NewAuthorizationApiWithConfig(sdkConfig)
	cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceAuthDivision(), constants.DefaultConsistencyChecks, "genesyscloud_auth_division")

	log.Printf("Reading division %s", d.Id())

	return util.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		division, resp, getErr := authAPI.GetAuthorizationDivision(d.Id(), false)
		if getErr != nil {
			if util.IsStatus404(resp) {
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError("genesyscloud_auth_division", fmt.Sprintf("Failed to read division %s | error: %s", d.Id(), getErr), resp))
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError("genesyscloud_auth_division", fmt.Sprintf("Failed to read division %s | error: %s", d.Id(), getErr), resp))
		}

		d.Set("name", *division.Name)

		if division.Description != nil {
			d.Set("description", *division.Description)
		} else {
			d.Set("description", nil)
		}

		if division.HomeDivision != nil {
			d.Set("home", *division.HomeDivision)
		} else {
			d.Set("home", nil)
		}

		log.Printf("Read division %s %s", d.Id(), *division.Name)
		return cc.CheckState(d)
	})
}

func updateAuthDivision(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)
	description := d.Get("description").(string)

	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	authAPI := platformclientv2.NewAuthorizationApiWithConfig(sdkConfig)

	log.Printf("Updating division %s", name)
	_, resp, err := authAPI.PutAuthorizationDivision(d.Id(), platformclientv2.Authzdivision{
		Name:        &name,
		Description: &description,
	})
	if err != nil {
		return util.BuildAPIDiagnosticError("genesyscloud_auth_division", fmt.Sprintf("Failed to update division %s error: %s", name, err), resp)
	}

	log.Printf("Updated division %s", name)

	return readAuthDivision(ctx, d, meta)
}

func deleteAuthDivision(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)
	home := d.Get("home").(bool)

	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	authAPI := platformclientv2.NewAuthorizationApiWithConfig(sdkConfig)

	if home {
		// Do not delete home division
		log.Printf("Not deleting home division %s", name)
		return nil
	}

	// Sometimes a division with resources in it priorly still thinks it is attached to those resources during a destroy run.
	// We're retrying again as those resources should detach completely eventually.
	diagErr := util.RetryWhen(util.IsStatus400, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
		log.Printf("Deleting division %s", name)
		resp, err := authAPI.DeleteAuthorizationDivision(d.Id(), false)
		if err != nil {
			return resp, util.BuildAPIDiagnosticError("genesyscloud_auth_division", fmt.Sprintf("Failed to delete Division %s error: %s", d.Id(), err), resp)
		}
		return resp, nil
	})
	if diagErr != nil {
		return diagErr
	}

	return util.WithRetries(ctx, 180*time.Second, func() *retry.RetryError {
		_, resp, err := authAPI.GetAuthorizationDivision(d.Id(), false)
		if err != nil {
			if util.IsStatus404(resp) {
				// Division deleted
				log.Printf("Deleted division %s", name)
				return nil
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError("genesyscloud_auth_division", fmt.Sprintf("Error deleting division %s | error:: %s", name, err), resp))
		}
		return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError("genesyscloud_auth_division", fmt.Sprintf("Division %s still exists", name), resp))
	})
}

func GenerateAuthDivisionBasic(resourceID string, name string) string {
	return GenerateAuthDivisionResource(resourceID, name, util.NullValue, util.FalseValue)
}

func GenerateAuthDivisionResource(
	resourceID string,
	name string,
	description string,
	home string) string {
	return fmt.Sprintf(`resource "genesyscloud_auth_division" "%s" {
		name = "%s"
		description = %s
		home = %s
	}
	`, resourceID, name, description, home)
}
