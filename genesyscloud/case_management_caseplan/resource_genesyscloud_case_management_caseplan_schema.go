package case_management_caseplan

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

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

const (
	activityTypeNone     = "None"
	activityTypeWorkitem = "Workitem"
	maxStageplans        = 5
)

// SetRegistrar registers all of the resources, datasources and exporters in the package
func SetRegistrar(regInstance registrar.Registrar) {
	regInstance.RegisterResource(ResourceType, ResourceCaseManagementCaseplan())
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

	workitemSettingsResource := &schema.Resource{
		Schema: map[string]*schema.Schema{
			"worktype_id": {
				Description: `Worktype id. The worktype's schema must be the caseplan's data schema, and its division must match the caseplan division (unless the caseplan division is "*").`,
				Required:    true,
				Type:        schema.TypeString,
			},
		},
	}

	stepplanResource := &schema.Resource{
		Schema: map[string]*schema.Schema{
			"id": {
				Description: `Stepplan id.`,
				Computed:    true,
				Type:        schema.TypeString,
			},
			"name": {
				Description:  `The name of the stepplan.`,
				Required:     true,
				Type:         schema.TypeString,
				ValidateFunc: validation.StringLenBetween(3, 256),
			},
			"description": {
				Description:  `The description of the stepplan.`,
				Optional:     true,
				Type:         schema.TypeString,
				ValidateFunc: validation.StringLenBetween(0, 512),
			},
			"activity_type": {
				Description:  `"None" completes the step as soon as it activates. "Workitem" creates a workitem from workitem_settings.worktype_id and completes when the workitem completes.`,
				Optional:     true,
				Default:      activityTypeNone,
				Type:         schema.TypeString,
				ValidateFunc: validation.StringInSlice([]string{activityTypeNone, activityTypeWorkitem}, false),
			},
			"workitem_settings": {
				Description: `Workitem settings. Required when activity_type is "Workitem", and not allowed otherwise.`,
				Optional:    true,
				Type:        schema.TypeList,
				MaxItems:    1,
				Elem:        workitemSettingsResource,
			},
		},
	}

	stageplanResource := &schema.Resource{
		Schema: map[string]*schema.Schema{
			"id": {
				Description: `Stageplan id. Stable across caseplan versions.`,
				Computed:    true,
				Type:        schema.TypeString,
			},
			"name": {
				Description:  `The name of the stageplan. Must be unique within the caseplan; stageplans are matched to existing ones by name, then by position.`,
				Required:     true,
				Type:         schema.TypeString,
				ValidateFunc: validation.StringLenBetween(3, 256),
			},
			"description": {
				Description:  `The description of the stageplan.`,
				Optional:     true,
				Type:         schema.TypeString,
				ValidateFunc: validation.StringLenBetween(0, 512),
			},
			"stepplan": {
				Description: `The stepplan of this stageplan. The API supports exactly one stepplan per stageplan.`,
				Required:    true,
				Type:        schema.TypeList,
				MinItems:    1,
				MaxItems:    1,
				Elem:        stepplanResource,
			},
		},
	}

	return &schema.Resource{
		Description: "Genesys Cloud case management caseplan, including its stageplans and stepplans.\n\n" +
			"Terraform manages the **published** caseplan. Every apply that changes a versioned field (anything other than `name` and `description`) " +
			"creates a draft (or reuses an existing one), writes the configuration over it, and publishes it. Changes made to a draft outside Terraform are overwritten by the next such apply. " +
			"`name` and `description` are not versioned and are updated in place without publishing.\n\n" +
			"`division_id`, `customer_intent`, `reference_prefix` and `data_schema` cannot change once the caseplan is published, which happens on the first apply.\n\n" +
			"Adding, removing or reordering `stageplan` blocks requires the add/delete stageplans feature in the org.",

		CreateContext: provider.CreateWithPooledClient(createCaseManagementCaseplan),
		ReadContext:   provider.ReadWithPooledClient(readCaseManagementCaseplan),
		UpdateContext: provider.UpdateWithPooledClient(updateCaseManagementCaseplan),
		DeleteContext: provider.DeleteWithPooledClient(deleteCaseManagementCaseplan),
		Importer: &schema.ResourceImporter{
			StateContext: importCaseManagementCaseplan,
		},
		CustomizeDiff: customizeCaseManagementCaseplanDiff,
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			`name`: {
				Description: `The name of the Caseplan.`,
				Required:    true,
				Type:        schema.TypeString,
			},
			`division_id`: {
				Description: `The division to which this entity belongs. Cannot be changed after the caseplan has been published.`,
				Optional:    true,
				Computed:    true,
				Type:        schema.TypeString,
			},
			`description`: {
				Description: `The description of the Caseplan.`,
				Optional:    true,
				Type:        schema.TypeString,
			},
			`reference_prefix`: {
				Description: `The prefix used when creating the reference for Cases from the Caseplan. Cannot be changed after the caseplan has been published.`,
				Optional:    true,
				Type:        schema.TypeString,
			},
			`default_due_duration_in_seconds`: {
				Description: `The default due duration in seconds for Cases created from the Caseplan.`,
				Optional:    true,
				Computed:    true,
				Type:        schema.TypeInt,
			},
			`default_ttl_seconds`: {
				Description: `The default TTL in seconds for Cases created from the Caseplan.`,
				Optional:    true,
				Computed:    true,
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
				Description: `The customer intent for the Cases created from the caseplan. Cannot be changed after the caseplan has been published.`,
				Optional:    true,
				Type:        schema.TypeList,
				MaxItems:    1,
				Elem:        customerIntentReferenceResource,
			},
			`data_schema`: {
				Description: `Task management workitem schema bound to case data for this caseplan. Publishing requires a data schema. Cannot be changed after the caseplan has been published.`,
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
					},
				},
			},
			`intake_settings`: {
				Description: `Intake field configuration when collecting case data. Up to 10 entries.`,
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
			`stageplan`: {
				Description: "Ordered stageplans; the block order is the process-flow order. A new caseplan starts with 3 stageplans, which the configured blocks are mapped onto by position. " +
					"Stageplans are matched to existing ones by name, then by position; renaming and moving the same stageplan in one apply replaces it with a new id. " +
					"When no blocks are configured, stageplans are not managed.",
				Optional: true,
				Type:     schema.TypeList,
				MinItems: 1,
				MaxItems: maxStageplans,
				Elem:     stageplanResource,
			},
			`published_version`: {
				Description: `The published version number. 0 when the caseplan has never been published.`,
				Computed:    true,
				Type:        schema.TypeInt,
			},
			`has_draft`: {
				Description: `Whether an unpublished draft exists. A draft created outside Terraform is overwritten and published by the next apply that changes a versioned field.`,
				Computed:    true,
				Type:        schema.TypeBool,
			},
		},
	}
}

// CaseManagementCaseplanExporter returns the resourceExporter object used to hold the genesyscloud_case_management_caseplan exporter's config
func CaseManagementCaseplanExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: provider.GetAllWithPooledClient(getAllAuthCaseManagementCaseplans),
		RefAttrs: map[string]*resourceExporter.RefAttrSettings{
			"division_id":           {RefType: "genesyscloud_auth_division", AltValues: []string{"*"}},
			"default_case_owner.id": {RefType: "genesyscloud_user"},
			"customer_intent.id":    {RefType: "genesyscloud_intents_customerintents"},
			"data_schema.id":        {RefType: "genesyscloud_task_management_workitem_schema"},
			"stageplan.stepplan.workitem_settings.worktype_id": {RefType: "genesyscloud_task_management_worktype"},
		},
		ExcludedAttributes: []string{
			"published_version",
			"has_draft",
			"stageplan.id",
			"stageplan.stepplan.id",
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
