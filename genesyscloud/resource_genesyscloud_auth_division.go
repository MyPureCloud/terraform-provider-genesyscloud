package genesyscloud

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/platformclientv2"
)

func getAllAuthDivisions(ctx context.Context, clientConfig *platformclientv2.Configuration) (ResourceIDMetaMap, diag.Diagnostics) {
	resources := make(ResourceIDMetaMap)
	authAPI := platformclientv2.NewAuthorizationApiWithConfig(clientConfig)

	for pageNum := 1; ; pageNum++ {
		divisions, _, getErr := authAPI.GetAuthorizationDivisions(100, pageNum, "", nil, "", "", false, nil, "")
		if getErr != nil {
			return nil, diag.Errorf("Failed to get page of divisions: %v", getErr)
		}

		if divisions.Entities == nil || len(*divisions.Entities) == 0 {
			break
		}

		for _, division := range *divisions.Entities {
			resources[*division.Id] = &ResourceMeta{Name: *division.Name}
		}
	}

	return resources, nil
}

func authDivisionExporter() *ResourceExporter {
	return &ResourceExporter{
		GetResourcesFunc: getAllWithPooledClient(getAllAuthDivisions),
		RefAttrs:         map[string]*RefAttrSettings{}, // No references
	}
}

func resourceAuthDivision() *schema.Resource {
	return &schema.Resource{
		Description: "Genesys Cloud Authorization Division",

		CreateContext: createWithPooledClient(createAuthDivision),
		ReadContext:   readWithPooledClient(readAuthDivision),
		UpdateContext: updateWithPooledClient(updateAuthDivision),
		DeleteContext: deleteWithPooledClient(deleteAuthDivision),
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
				Description: "True if this is the home division. This can be set to manage the pre-existing home division.",
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

	sdkConfig := meta.(*providerMeta).ClientConfig
	authAPI := platformclientv2.NewAuthorizationApiWithConfig(sdkConfig)

	if home {
		// Home division must already exist, or it cannot be modified
		id, diagErr := getHomeDivisionID()
		if diagErr != nil {
			return diagErr
		}
		d.SetId(id)
		return updateAuthDivision(ctx, d, meta)
	}

	log.Printf("Creating division %s", name)
	division, _, err := authAPI.PostAuthorizationDivisions(platformclientv2.Authzdivision{
		Name:        &name,
		Description: &description,
	})
	if err != nil {
		return diag.Errorf("Failed to create division %s: %s", name, err)
	}

	// Give auth service's indexes time to update
	time.Sleep(5 * time.Second)

	d.SetId(*division.Id)
	log.Printf("Created division %s %s", name, *division.Id)
	return readAuthDivision(ctx, d, meta)
}

func readAuthDivision(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*providerMeta).ClientConfig
	authAPI := platformclientv2.NewAuthorizationApiWithConfig(sdkConfig)

	log.Printf("Reading division %s", d.Id())

	division, resp, getErr := authAPI.GetAuthorizationDivision(d.Id(), false)
	if getErr != nil {
		if resp != nil && resp.StatusCode == 404 {
			d.SetId("")
			return nil
		}
		return diag.Errorf("Failed to read division %s: %s", d.Id(), getErr)
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
	return nil
}

func updateAuthDivision(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)
	description := d.Get("description").(string)

	sdkConfig := meta.(*providerMeta).ClientConfig
	authAPI := platformclientv2.NewAuthorizationApiWithConfig(sdkConfig)

	log.Printf("Updating division %s", name)
	_, _, err := authAPI.PutAuthorizationDivision(d.Id(), platformclientv2.Authzdivision{
		Name:        &name,
		Description: &description,
	})
	if err != nil {
		return diag.Errorf("Failed to update division %s: %s", name, err)
	}

	log.Printf("Updated division %s", name)

	// Give time for public API caches to update
	time.Sleep(5 * time.Second)
	return readAuthDivision(ctx, d, meta)
}

func deleteAuthDivision(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)
	home := d.Get("home").(bool)

	sdkConfig := meta.(*providerMeta).ClientConfig
	authAPI := platformclientv2.NewAuthorizationApiWithConfig(sdkConfig)

	if home {
		// Do not delete home division
		log.Printf("Not deleting home division %s", name)
		return nil
	}

	log.Printf("Deleting division %s", name)
	_, err := authAPI.DeleteAuthorizationDivision(d.Id(), true)
	if err != nil {
		return diag.Errorf("Failed to delete division %s: %s", name, err)
	}

	// Give public API caches time to expire
	time.Sleep(5 * time.Second)
	return withRetries(ctx, 30*time.Second, func() *resource.RetryError {
		_, resp, err := authAPI.GetAuthorizationDivision(d.Id(), false)
		if err != nil {
			if resp != nil && resp.StatusCode == 404 {
				// Division deleted
				log.Printf("Deleted division %s", name)
				return nil
			}
			return resource.NonRetryableError(fmt.Errorf("Error deleting division %s: %s", name, err))
		}
		return resource.RetryableError(fmt.Errorf("Division %s still exists", name))
	})
}
