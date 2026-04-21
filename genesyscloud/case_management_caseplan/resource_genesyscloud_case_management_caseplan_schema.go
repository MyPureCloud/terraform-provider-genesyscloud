package case_management_caseplan

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	registrar "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_register"
)

/*
resource_genesycloud_case_management_caseplan_schema.go holds four functions within it:

1.  The registration code that registers the Datasource, Resource and Exporter for the package.
2.  The resource schema definitions for the case_management_caseplan resource.
3.  The datasource schema definitions for the case_management_caseplan datasource.
4.  The resource exporter configuration for the case_management_caseplan exporter.
*/
const resourceName = "genesyscloud_case_management_caseplan"

// ResourceType is the Terraform type name for this resource.
const ResourceType = "genesyscloud_case_management_caseplan"

// SetRegistrar registers all of the resources, datasources and exporters in the package
func SetRegistrar(regInstance registrar.Registrar) {
	regInstance.RegisterResource(ResourceType, ResourceCaseManagementCaseplan())
	regInstance.RegisterResource(PublishResourceType, ResourceCaseManagementCaseplanPublish())
	regInstance.RegisterResource(CreateVersionResourceType, ResourceCaseManagementCaseplanCreateVersion())
	regInstance.RegisterDataSource(ResourceType, DataSourceCaseManagementCaseplan())
	regInstance.RegisterExporter(ResourceType, CaseManagementCaseplanExporter())
}

// ResourceCaseManagementCaseplan registers the genesyscloud_case_management_caseplan resource with Terraform
func ResourceCaseManagementCaseplan() *schema.Resource {
	userReferenceResource := &schema.Resource{
		Schema: map[string]*schema.Schema{
			"id": {
				Description: `User id for default case owner (maps to defaultCaseOwnerId on create).`,
				Optional:    true,
				Type:        schema.TypeString,
			},
		},
	}

	customerIntentReferenceResource := &schema.Resource{
		Schema: map[string]*schema.Schema{
			"id": {
				Description: `Customer intent id (maps to customerIntentId on create).`,
				Optional:    true,
				Type:        schema.TypeString,
			},
		},
	}

	return &schema.Resource{
		Description: `Genesys Cloud case management caseplan`,

		CreateContext: provider.CreateWithPooledClient(createCaseManagementCaseplan),
		ReadContext:   provider.ReadWithPooledClient(readCaseManagementCaseplan),
		UpdateContext: provider.UpdateWithPooledClient(updateCaseManagementCaseplan),
		DeleteContext: provider.DeleteWithPooledClient(deleteCaseManagementCaseplan),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			`name`: {
				Description: `The name of the Caseplan.`,
				Optional:    true,
				Type:        schema.TypeString,
			},
			`division_id`: {
				Description: `The division to which this entity belongs. Cannot be changed after the caseplan has been published at least once.`,
				Optional:    true,
				Type:        schema.TypeString,
			},
			`description`: {
				Description: `The description of the Caseplan.`,
				Optional:    true,
				Type:        schema.TypeString,
			},
			`reference_prefix`: {
				Description: `The prefix used when creating the reference for Cases from the Caseplan. Cannot be changed after the caseplan has been published at least once.`,
				Optional:    true,
				Type:        schema.TypeString,
			},
			`default_due_duration_in_seconds`: {
				Description: `The default due duration in seconds for Cases created from the Caseplan.`,
				Optional:    true,
				Type:        schema.TypeInt,
			},
			`default_ttl_seconds`: {
				Description: `The default TTL in seconds for Cases created from the Caseplan.`,
				Optional:    true,
				Type:        schema.TypeInt,
			},
			`default_case_owner`: {
				Description: `The default case owner for Cases created from the Caseplan.`,
				Optional:    true,
				Type:        schema.TypeList,
				MaxItems:    1,
				Elem:        userReferenceResource,
			},
			`customer_intent`: {
				Description: `The customer intent for the Cases created from the caseplan. Cannot be changed after the caseplan has been published at least once.`,
				Optional:    true,
				Type:        schema.TypeList,
				MaxItems:    1,
				Elem:        customerIntentReferenceResource,
			},
			`data_schema`: {
				Description: `Task management workitem schema bound to case data for this caseplan. Updates use PUT /caseplans/{id}/dataschemas/default (only key "default" exists today). IDs must be task-management workitem schemas. Cannot be changed after the caseplan has been published at least once.`,
				Required:    true,
				Type:        schema.TypeList,
				MinItems:    1,
				MaxItems:    1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						`id`: {
							Description: `Workitem schema id.`,
							Type:        schema.TypeString,
							Required:    true,
						},
						`version`: {
							Description: `Workitem schema version number.`,
							Type:        schema.TypeInt,
							Required:    true,
						},
					},
				},
			},
			`intake_settings`: {
				Description: `Intake field configuration when collecting case data (maps to API intakeSettings). Up to 10 entries. Read uses GET .../caseplans/{id}/versions/{version}/intakesettings because the caseplan GET response does not include this field. Cannot be changed after the caseplan has been published at least once.`,
				Optional:    true,
				Type:        schema.TypeList,
				MaxItems:    10,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						`property`: {
							Description: `Property name from the bound workitem schema (data schema).`,
							Type:        schema.TypeString,
							Required:    true,
						},
						`required`: {
							Description: `Whether this property is required at intake.`,
							Optional:    true,
							Default:     false,
							Type:        schema.TypeBool,
						},
						`display_order`: {
							Description: `Display order for this property in the intake UI.`,
							Optional:    true,
							Default:     0,
							Type:        schema.TypeInt,
						},
					},
				},
			},
		},
	}
}

// CaseManagementCaseplanExporter returns the resourceExporter object used to hold the genesyscloud_case_management_caseplan exporter's config
func CaseManagementCaseplanExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: provider.GetAllWithPooledClient(getAllAuthCaseManagementCaseplans),
		RefAttrs:         map[string]*resourceExporter.RefAttrSettings{
			// TODO: Add any reference attributes here
		},
	}
}

// DataSourceCaseManagementCaseplan registers the genesyscloud_case_management_caseplan data source
func DataSourceCaseManagementCaseplan() *schema.Resource {
	return &schema.Resource{
		Description: `Genesys Cloud case management caseplan data source. Select an case management caseplan by name`,
		ReadContext: provider.ReadWithPooledClient(dataSourceCaseManagementCaseplanRead),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Description: `case management caseplan name`,
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}
}
