package genesyscloud

import (
	"context"
	"fmt"
	"log"
	"time"

	"terraform-provider-genesyscloud/genesyscloud/consistency_checker"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v92/platformclientv2"
)

func getAllAuthExternalContacts(_ context.Context, clientConfig *platformclientv2.Configuration) (ResourceIDMetaMap, diag.Diagnostics) {
	resources := make(ResourceIDMetaMap)
	externalAPI := platformclientv2.NewExternalContactsApiWithConfig(clientConfig)

	for pageNum := 1; ; pageNum++ {
		const pageSize = 100
		externalContacts, _, getErr := externalAPI.GetExternalcontactsContacts(pageSize, pageNum, "", "", nil)
		if getErr != nil {
			return nil, diag.Errorf("Failed to get external contacts: %v", getErr)
		}

		if externalContacts.Entities == nil || len(*externalContacts.Entities) == 0 {
			break
		}

		for _, externalContact := range *externalContacts.Entities {
			log.Printf("Dealing with external concat id : %s", *externalContact.Id)
			resources[*externalContact.Id] = &ResourceMeta{Name: *externalContact.Id}
		}
	}

	return resources, nil
}

func externalContactExporter() *ResourceExporter {
	return &ResourceExporter{
		GetResourcesFunc: getAllWithPooledClient(getAllAuthExternalContacts),
		RefAttrs:         map[string]*RefAttrSettings{}, // No references
	}
}

func resourceExternalContact() *schema.Resource {
	return &schema.Resource{
		Description: "Genesys Cloud External Contact",

		CreateContext: createWithPooledClient(createExternalContact),
		ReadContext:   readWithPooledClient(readExternalContact),
		UpdateContext: updateWithPooledClient(updateExternalContact),
		DeleteContext: deleteWithPooledClient(deleteExternalContact),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			"firstname": {
				Description: "The first name of the contact.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"middlename": {
				Description: "The middle name of the contact.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"lastname": {
				Description: "The middle name of the contact.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"title": {
				Description: "The title of the contact.",
				Type:        schema.TypeString,
				Optional:    true,
			},
		},
	}
}

func createExternalContact(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	firstName := d.Get("firstname").(string)
	middleName := d.Get("middlename").(string)
	lastName := d.Get("lastname").(string)
	title := d.Get("title").(string)

	sdkConfig := meta.(*ProviderMeta).ClientConfig
	externalAPI := platformclientv2.NewExternalContactsApiWithConfig(sdkConfig)

	log.Printf("Creating External Contact %s", title)
	externalContact, _, err := externalAPI.PostExternalcontactsContacts(platformclientv2.Externalcontact{
		FirstName:  &firstName,
		MiddleName: &middleName,
		LastName:   &lastName,
		Title:      &title,
	})
	if err != nil {
		return diag.Errorf("Failed to create external contact %s: %s", title, err)
	}

	d.SetId(*externalContact.Id)
	log.Printf("Created external contact %s %s", title, *externalContact.Id)
	return readExternalContact(ctx, d, meta)
}

func readExternalContact(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*ProviderMeta).ClientConfig
	externalAPI := platformclientv2.NewExternalContactsApiWithConfig(sdkConfig)

	log.Printf("Reading contact %s", d.Id())

	return withRetriesForRead(ctx, d, func() *resource.RetryError {
		externalContact, resp, getErr := externalAPI.GetExternalcontactsContact(d.Id(), nil)
		if getErr != nil {
			if isStatus404(resp) {
				return resource.RetryableError(fmt.Errorf("Failed to read external contact %s: %s", d.Id(), getErr))
			}
			return resource.NonRetryableError(fmt.Errorf("Failed to read external contact %s: %s", d.Id(), getErr))
		}

		cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, resourceAuthDivision())

		if externalContact.FirstName != nil {
			d.Set("firstname", *externalContact.FirstName)
		} else {
			d.Set("firstname", nil)
		}

		if externalContact.MiddleName != nil {
			d.Set("middlename", *externalContact.MiddleName)
		} else {
			d.Set("middlename", nil)
		}

		if externalContact.LastName != nil {
			d.Set("lastname", *externalContact.LastName)
		} else {
			d.Set("lastname", nil)
		}

		if externalContact.Title != nil {
			d.Set("title", *externalContact.Title)
		} else {
			d.Set("title", nil)
		}

		log.Printf("Read external contact %s", d.Id())
		return cc.CheckState()
	})
}

func updateExternalContact(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	firstName := d.Get("firstname").(string)
	middleName := d.Get("middlename").(string)
	lastName := d.Get("lastname").(string)
	title := d.Get("title").(string)

	sdkConfig := meta.(*ProviderMeta).ClientConfig
	externalAPI := platformclientv2.NewExternalContactsApiWithConfig(sdkConfig)

	log.Printf("Updating external contact %s", title)
	_, _, err := externalAPI.PutExternalcontactsContact(d.Id(), platformclientv2.Externalcontact{
		FirstName:  &firstName,
		MiddleName: &middleName,
		LastName:   &lastName,
		Title:      &title,
	})
	if err != nil {
		return diag.Errorf("Failed to update external contact %s: %s", title, err)
	}

	log.Printf("Updated external contact %s", title)

	return readExternalContact(ctx, d, meta)
}

func deleteExternalContact(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	title := d.Get("title").(string)

	sdkConfig := meta.(*ProviderMeta).ClientConfig
	externalAPI := platformclientv2.NewExternalContactsApiWithConfig(sdkConfig)

	_, _, err := externalAPI.DeleteExternalcontactsContact(d.Id())
	if err != nil {
		return diag.Errorf("Failed to delete external contact %s: %s", title, err)
	}

	return withRetries(ctx, 180*time.Second, func() *resource.RetryError {
		_, resp, err := externalAPI.GetExternalcontactsContact(title, nil)

		if err == nil {
			return resource.NonRetryableError(fmt.Errorf("Error deleting external contact %s: %s", title, err))
		}
		if isStatus404(resp) {
			// Success  : External contact deleted
			log.Printf("Deleted external contact %s", title)
			return nil
		}

		return resource.RetryableError(fmt.Errorf("External contact %s still exists", title))
	})
}
