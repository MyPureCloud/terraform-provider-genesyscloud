package business_rules_schema

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/constants"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v165/platformclientv2"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/consistency_checker"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/resourcedata"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
)

/*
The resource_genesyscloud_business_rules_schema.go contains all of the methods that perform the core logic for a resource.
*/

// getAllBusinessRulesSchemas retrieves all of the business rules schemas via Terraform in the Genesys Cloud and is used for the exporter
func getAllBusinessRulesSchemas(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	proxy := getBusinessRulesSchemaProxy(clientConfig)
	resources := make(resourceExporter.ResourceIDMetaMap)

	schemas, resp, err := proxy.getAllBusinessRulesSchema(ctx)
	if err != nil {
		return nil, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to get all business rules schemas error: %s", err), resp)
	}

	for _, schema := range *schemas {
		log.Printf("Dealing with business rules schema id: %s", *schema.Id)
		resources[*schema.Id] = &resourceExporter.ResourceMeta{BlockLabel: *schema.Name}
	}
	return resources, nil
}

// createBusinessRulesSchema is used by the business_rules_schema resource to create Genesys cloud business rules schemas
func createBusinessRulesSchema(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getBusinessRulesSchemaProxy(sdkConfig)

	dataSchema, err := BuildSdkBusinessRulesSchema(d, nil)
	if err != nil {
		return util.BuildDiagnosticError(ResourceType, "create: failed to build business rules schema", err)
	}

	log.Printf("Creating business rules schema")
	schema, resp, err := proxy.createBusinessRulesSchema(ctx, dataSchema)
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to create business rules schema %s error: %s", *dataSchema.Name, err), resp)
	}

	d.SetId(*schema.Id)

	// If enabled is set to 'false' do an update call to the schema
	if enabled, ok := d.Get("enabled").(bool); ok && !enabled {
		log.Printf("Updating business rules schema: %s, to set 'enabled' to 'false'", *schema.Name)
		dataSchema.Version = platformclientv2.Int(1)
		_, resp, err := proxy.updateBusinessRulesSchema(ctx, *schema.Id, dataSchema)
		if err != nil {
			return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to update business rules schema %s error: %s", d.Id(), err), resp)
		}
		log.Printf("Updated newly created business rules schema: %s. 'enabled' set to to 'false'", *schema.Name)
	}

	log.Printf("Created business rules schema %s: %s", *schema.Name, *schema.Id)
	return readBusinessRulesSchema(ctx, d, meta)
}

// readBusinessRulesSchema is used by the business_rules_schema resource to read a business rules schema from genesys cloud
func readBusinessRulesSchema(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getBusinessRulesSchemaProxy(sdkConfig)
	cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceBusinessRulesSchema(), constants.ConsistencyChecks(), ResourceType)

	log.Printf("Reading business rules schema %s", d.Id())

	return util.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		schema, resp, getErr := proxy.getBusinessRulesSchemaById(ctx, d.Id())
		if getErr != nil {
			if util.IsStatus404(resp) {
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("failed to read business rules schema %s | error: %s", d.Id(), getErr), resp))
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("failed to read business rules schema %s | error: %s", d.Id(), getErr), resp))
		}

		var schemaPropsPtr *string
		if schema.JsonSchema != nil {
			schemaProps, err := json.Marshal(schema.JsonSchema.Properties)
			if err != nil {
				return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("error in reading json schema properties of %s | error: %v", *schema.Name, err), resp))
			}
			if string(schemaProps) != util.NullValue {
				schemaPropsStr := string(schemaProps)
				schemaPropsPtr = &schemaPropsStr
			}
		} else {
			schemaPropsPtr = nil
		}

		resourcedata.SetNillableValue(d, "name", schema.Name)
		resourcedata.SetNillableValue(d, "description", schema.JsonSchema.Description)
		resourcedata.SetNillableValue(d, "properties", schemaPropsPtr)
		resourcedata.SetNillableValue(d, "enabled", schema.Enabled)
		resourcedata.SetNillableValue(d, "version", schema.Version)

		log.Printf("Read business rules schema %s %s", d.Id(), *schema.Name)
		return cc.CheckState(d)
	})
}

// updateBusinessRulesSchema is used by the business_rules_schema resource to update a business rules schema in Genesys Cloud
func updateBusinessRulesSchema(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getBusinessRulesSchemaProxy(sdkConfig)

	log.Printf("Getting version of business rules schema")
	curSchema, resp, err := proxy.getBusinessRulesSchemaById(ctx, d.Id())
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to get business rules schema By id %s error: %s", d.Id(), err), resp)
	}

	dataSchema, err := BuildSdkBusinessRulesSchema(d, curSchema.Version)
	if err != nil {
		return util.BuildDiagnosticError(ResourceType, "update: failed to build business rules schema", err)
	}

	log.Printf("Updating business rules schema")
	updatedSchema, resp, err := proxy.updateBusinessRulesSchema(ctx, d.Id(), dataSchema)
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to update business rules schema %s error: %s", d.Id(), err), resp)
	}

	log.Printf("Updated business rules schema %s", *updatedSchema.Id)
	return readBusinessRulesSchema(ctx, d, meta)
}

// deleteBusinessRulesSchema is used by the business_rules_schema resource to delete a business rules schema from Genesys cloud
func deleteBusinessRulesSchema(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getBusinessRulesSchemaProxy(sdkConfig)

	resp, err := proxy.deleteBusinessRulesSchema(ctx, d.Id())
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to delete business rules schema %s error: %s", d.Id(), err), resp)
	}

	return util.WithRetries(ctx, 180*time.Second, func() *retry.RetryError {
		isDeleted, resp, err := proxy.getBusinessRulesSchemaDeletedStatus(ctx, d.Id())
		if err != nil {
			if util.IsStatus404(resp) {
				log.Printf("Deleted business rules schema %s", d.Id())
				return nil
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("error deleting business rules schema %s | error: %s", d.Id(), err), resp))
		}

		if isDeleted {
			log.Printf("Deleted business rules schema %s", d.Id())
			return nil
		}
		return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("business rules schema %s still exists", d.Id()), resp))
	})
}
