package architect_datatable

import (
	"context"
	"fmt"
	"log"
	"terraform-provider-genesyscloud/genesyscloud"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"terraform-provider-genesyscloud/genesyscloud/consistency_checker"

	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v119/platformclientv2"
)

type Datatableproperty struct {
	Id           *string      `json:"$id,omitempty"`
	VarType      *string      `json:"type,omitempty"`
	Title        *string      `json:"title,omitempty"`
	Default      *interface{} `json:"default,omitempty"`
	DisplayOrder *int         `json:"displayOrder,omitempty"`
}

// Overriding the SDK Datatable document as it does not allow setting additionalProperties to 'false' as required by the API
type Jsonschemadocument struct {
	Schema               *string                       `json:"$schema,omitempty"`
	VarType              *string                       `json:"type,omitempty"`
	Required             *[]string                     `json:"required,omitempty"`
	Properties           *map[string]Datatableproperty `json:"properties,omitempty"`
	AdditionalProperties *interface{}                  `json:"additionalProperties,omitempty"`
}

type Datatable struct {
	Id          *string                            `json:"id,omitempty"`
	Name        *string                            `json:"name,omitempty"`
	Description *string                            `json:"description,omitempty"`
	Division    *platformclientv2.Writabledivision `json:"division,omitempty"`
	Schema      *Jsonschemadocument                `json:"schema,omitempty"`
}

func getAllArchitectDatatables(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	resources := make(resourceExporter.ResourceIDMetaMap)

	archProxy := getArchitectDatatableProxy(clientConfig)
	tables, err := archProxy.getAllArchitectDatatable(ctx)

	if err != nil {
		return resources, diag.Errorf("Error encountered while calling getAllArchitectDatattables %s.\n", err)
	}

	for _, table := range *tables {
		resources[*table.Id] = &resourceExporter.ResourceMeta{Name: *table.Name}
	}

	return resources, nil
}

func createArchitectDatatable(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)
	divisionID := d.Get("division_id").(string)
	description := d.Get("description").(string)

	sdkConfig := meta.(*genesyscloud.ProviderMeta).ClientConfig
	archProxy := getArchitectDatatableProxy(sdkConfig)

	log.Printf("Creating architect_datatable %s", name)

	datatableSchema, diagErr := buildSdkDatatableSchema(d)
	if diagErr != nil {
		return diagErr
	}

	datatable := &Datatable{
		Name:   &name,
		Schema: datatableSchema,
	}
	// Optional
	if divisionID != "" {
		datatable.Division = &platformclientv2.Writabledivision{Id: &divisionID}
	}

	if description != "" {
		datatable.Description = &description
	}

	table, _, err := archProxy.createArchitectDatatable(ctx, datatable)
	if err != nil {
		return diag.Errorf("Failed to create architect_datatable %s: %s", name, err)
	}

	d.SetId(*table.Id)

	log.Printf("Created architect_datatable %s %s", name, *table.Id)
	return readArchitectDatatable(ctx, d, meta)
}

func readArchitectDatatable(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*genesyscloud.ProviderMeta).ClientConfig
	archProxy := getArchitectDatatableProxy(sdkConfig)

	log.Printf("Reading architect_datatable %s", d.Id())

	return genesyscloud.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		datatable, resp, getErr := archProxy.getArchitectDatatable(ctx, d.Id(), "schema")
		if getErr != nil {
			if genesyscloud.IsStatus404(resp) {
				return retry.RetryableError(fmt.Errorf("Failed to read architect_datatable %s: %s", d.Id(), getErr))
			}
			return retry.NonRetryableError(fmt.Errorf("Failed to read architect_datatable %s: %s", d.Id(), getErr))
		}
		cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceArchitectDatatable())
		d.Set("name", *datatable.Name)
		d.Set("division_id", *datatable.Division.Id)

		if datatable.Description != nil {
			d.Set("description", *datatable.Description)
		} else {
			d.Set("description", nil)
		}

		if datatable.Schema != nil && datatable.Schema.Properties != nil {
			d.Set("properties", flattenDatatableProperties(*datatable.Schema.Properties))
		} else {
			d.Set("properties", nil)
		}

		log.Printf("Read architect_datatable %s %s", d.Id(), *datatable.Name)

		return cc.CheckState()
	})
}

func updateArchitectDatatable(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	id := d.Id()
	name := d.Get("name").(string)
	divisionID := d.Get("division_id").(string)
	description := d.Get("description").(string)

	sdkConfig := meta.(*genesyscloud.ProviderMeta).ClientConfig
	archProxy := getArchitectDatatableProxy(sdkConfig)

	log.Printf("Updating architect_datatable %s", name)

	datatableSchema, diagErr := buildSdkDatatableSchema(d)
	if diagErr != nil {
		return diagErr
	}

	datatable := &Datatable{
		Id:     &id,
		Name:   &name,
		Schema: datatableSchema,
	}
	// Optional
	if divisionID != "" {
		datatable.Division = &platformclientv2.Writabledivision{Id: &divisionID}
	}

	if description != "" {
		datatable.Description = &description
	}

	_, _, err := archProxy.updateArchitectDatatable(ctx, datatable)
	if err != nil {
		return diag.Errorf("Failed to update architect_datatable %s: %s", name, err)
	}

	log.Printf("Updated architect_datatable %s", name)
	return readArchitectDatatable(ctx, d, meta)
}

func deleteArchitectDatatable(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)

	sdkConfig := meta.(*genesyscloud.ProviderMeta).ClientConfig
	archProxy := getArchitectDatatableProxy(sdkConfig)

	log.Printf("Deleting architect_datatable %s", name)
	_, err := archProxy.deleteArchitectDatatable(ctx, d.Id())
	if err != nil {
		return diag.Errorf("Failed to delete architect_datatable %s: %s", name, err)
	}

	return genesyscloud.WithRetries(ctx, 30*time.Second, func() *retry.RetryError {
		//might neeed to add expand with the "" as the expand
		_, resp, err := archProxy.getArchitectDatatable(ctx, d.Id(), "")
		if err != nil {
			if genesyscloud.IsStatus404(resp) {
				// Datatable row deleted
				log.Printf("Deleted architect_datatable row %s", name)
				return nil
			}
			return retry.NonRetryableError(fmt.Errorf("Error deleting architect_datatable row %s: %s", name, err))
		}
		return retry.RetryableError(fmt.Errorf("Datatable row %s still exists", name))
	})
}
