package user

import (
	"fmt"
	"strings"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	registrar "terraform-provider-genesyscloud/genesyscloud/resource_register"
	"terraform-provider-genesyscloud/genesyscloud/validators"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

const ResourceType = "genesyscloud_user"

// SetRegistrar registers all the resources and exporters in the package
func SetRegistrar(l registrar.Registrar) {
	l.RegisterDataSource(ResourceType, DataSourceUser())
	l.RegisterResource(ResourceType, ResourceUser())
	l.RegisterExporter(ResourceType, UserExporter())
}

var (
	contactTypeEmail = "EMAIL"

	phoneNumberResource = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"number": {
				Description:      "Phone number. Phone number must be in an E.164 number format.",
				Type:             schema.TypeString,
				Optional:         true,
				ValidateDiagFunc: validators.ValidatePhoneNumber,
			},
			"media_type": {
				Description:  "Media type of phone number (SMS | PHONE).",
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "PHONE",
				ValidateFunc: validation.StringInSlice([]string{"PHONE", "SMS"}, false),
			},
			"type": {
				Description:  "Type of number (WORK | WORK2 | WORK3 | WORK4 | HOME | MOBILE | OTHER).",
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "WORK",
				ValidateFunc: validation.StringInSlice([]string{"WORK", "WORK2", "WORK3", "WORK4", "HOME", "MOBILE", "OTHER"}, false),
			},
			"extension": {
				Description: "Phone number extension",
				Type:        schema.TypeString,
				Optional:    true,
			},
		},
	}

	utilizationSettingsResource = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"maximum_capacity": {
				Description:  "Maximum capacity of conversations of this media type. Value must be between 0 and 25.",
				Type:         schema.TypeInt,
				Required:     true,
				ValidateFunc: validation.IntBetween(0, 25),
			},
			"interruptible_media_types": {
				Description: fmt.Sprintf("Set of other media types that can interrupt this media type (%s).", strings.Join(getSdkUtilizationTypes(), " | ")),
				Type:        schema.TypeSet,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"include_non_acd": {
				Description: "Block this media type when on a non-ACD conversation.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
			},
		},
	}

	utilizationLabelResource = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"label_id": {
				Description: "Id of the label being configured.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"maximum_capacity": {
				Description:  "Maximum capacity of conversations with this label. Value must be between 0 and 25.",
				Type:         schema.TypeInt,
				Required:     true,
				ValidateFunc: validation.IntBetween(0, 25),
			},
			"interrupting_label_ids": {
				Description: "Set of other labels that can interrupt this label.",
				Type:        schema.TypeSet,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
		},
	}

	otherEmailResource = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"address": {
				Description: "Email address.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"type": {
				Description:  "Type of email address (WORK | HOME).",
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "WORK",
				ValidateFunc: validation.StringInSlice([]string{"WORK", "HOME"}, false),
			},
		},
	}
	userSkillResource = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"skill_id": {
				Description: "ID of routing skill.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"proficiency": {
				Description:  "Rating from 0.0 to 5.0 on how competent an agent is for a particular skill. It is used when a queue is set to 'Best available skills' mode to allow acd interactions to target agents with higher proficiency ratings.",
				Type:         schema.TypeFloat,
				Required:     true,
				ValidateFunc: validation.FloatBetween(0, 5),
			},
		},
	}
	userLanguageResource = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"language_id": {
				Description: "ID of routing language.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"proficiency": {
				Description:  "Proficiency is a rating from 0 to 5 on how competent an agent is for a particular language. It is used when a queue is set to 'Best available language' mode to allow acd interactions to target agents with higher proficiency ratings.",
				Type:         schema.TypeInt, // The API accepts a float, but the backend rounds to the nearest int
				Required:     true,
				ValidateFunc: validation.IntBetween(0, 5),
			},
		},
	}
	userLocationResource = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"location_id": {
				Description: "ID of location.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"notes": {
				Description: "Optional description on the user's location.",
				Type:        schema.TypeString,
				Optional:    true,
			},
		},
	}
)

func ResourceUser() *schema.Resource {
	return &schema.Resource{
		Description: "Genesys Cloud User",

		CreateContext: provider.CreateWithPooledClient(createUser),
		ReadContext:   provider.ReadWithPooledClient(readUser),
		UpdateContext: provider.UpdateWithPooledClient(updateUser),
		DeleteContext: provider.DeleteWithPooledClient(deleteUser),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			"email": {
				Description: "User's primary email and username.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"name": {
				Description: "User's full name.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"password": {
				Description: "User's password. If specified, this is only set on user create.",
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   true,
			},
			"state": {
				Description:  "User's state (active | inactive). Default is 'active'.",
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "active",
				ValidateFunc: validation.StringInSlice([]string{"active", "inactive"}, false),
			},
			"division_id": {
				Description: "The division to which this user will belong. If not set, the home division will be used.",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
			},
			"department": {
				Description: "User's department.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"title": {
				Description: "User's title.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"manager": {
				Description: "User ID of this user's manager.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"acd_auto_answer": {
				Description: "Enable ACD auto-answer.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
			},
			"routing_skills": {
				Description: "Skills and proficiencies for this user. If not set, this resource will not manage user skills.",
				Type:        schema.TypeSet,
				Optional:    true,
				Computed:    true,
				ConfigMode:  schema.SchemaConfigModeAttr,
				Elem:        userSkillResource,
			},
			"routing_languages": {
				Description: "Languages and proficiencies for this user. If not set, this resource will not manage user languages.",
				Type:        schema.TypeSet,
				Optional:    true,
				Computed:    true,
				ConfigMode:  schema.SchemaConfigModeAttr,
				Elem:        userLanguageResource,
			},
			"locations": {
				Description: "The user placement at each site location. If not set, this resource will not manage user locations.",
				Type:        schema.TypeSet,
				Optional:    true,
				Computed:    true,
				ConfigMode:  schema.SchemaConfigModeAttr,
				Elem:        userLocationResource,
			},
			"addresses": {
				Description: "The address settings for this user. If not set, this resource will not manage addresses.",
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Computed:    true,
				ConfigMode:  schema.SchemaConfigModeAttr,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"other_emails": {
							Description: "Other Email addresses for this user.",
							Type:        schema.TypeSet,
							Optional:    true,
							Elem:        otherEmailResource,
							ConfigMode:  schema.SchemaConfigModeAttr,
						},
						"phone_numbers": {
							Description: "Phone number addresses for this user.",
							Type:        schema.TypeSet,
							Optional:    true,
							Set:         phoneNumberHash,
							Elem:        phoneNumberResource,
							ConfigMode:  schema.SchemaConfigModeAttr,
						},
					},
				},
			},
			"profile_skills": {
				Description: "Profile skills for this user. If not set, this resource will not manage profile skills.",
				Type:        schema.TypeSet,
				Optional:    true,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"certifications": {
				Description: "Certifications for this user. If not set, this resource will not manage certifications.",
				Type:        schema.TypeSet,
				Optional:    true,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"employer_info": {
				Description: "The employer info for this user. If not set, this resource will not manage employer info.",
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Computed:    true,
				ConfigMode:  schema.SchemaConfigModeAttr,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"official_name": {
							Description: "User's official name.",
							Type:        schema.TypeString,
							Optional:    true,
						},
						"employee_id": {
							Description: "Employee ID.",
							Type:        schema.TypeString,
							Optional:    true,
						},
						"employee_type": {
							Description:  "Employee type (Full-time | Part-time | Contractor).",
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringInSlice([]string{"Full-time", "Part-time", "Contractor"}, false),
						},
						"date_hire": {
							Description:      "Hiring date. Dates must be an ISO-8601 string. For example: yyyy-MM-dd.",
							Type:             schema.TypeString,
							Optional:         true,
							ValidateDiagFunc: validators.ValidateDate,
						},
					},
				},
			},
			"routing_utilization": {
				Description: "The routing utilization settings for this user. If empty list, the org default settings are used. If not set, this resource will not manage the users's utilization settings.",
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Computed:    true,
				ConfigMode:  schema.SchemaConfigModeAttr,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"call": {
							Description: "Call media settings. If not set, this reverts to the default media type settings.",
							Type:        schema.TypeList,
							MaxItems:    1,
							Optional:    true,
							Computed:    true,
							ConfigMode:  schema.SchemaConfigModeAttr,
							Elem:        utilizationSettingsResource,
						},
						"callback": {
							Description: "Callback media settings. If not set, this reverts to the default media type settings.",
							Type:        schema.TypeList,
							MaxItems:    1,
							Optional:    true,
							Computed:    true,
							ConfigMode:  schema.SchemaConfigModeAttr,
							Elem:        utilizationSettingsResource,
						},
						"message": {
							Description: "Message media settings. If not set, this reverts to the default media type settings.",
							Type:        schema.TypeList,
							MaxItems:    1,
							Optional:    true,
							Computed:    true,
							ConfigMode:  schema.SchemaConfigModeAttr,
							Elem:        utilizationSettingsResource,
						},
						"email": {
							Description: "Email media settings. If not set, this reverts to the default media type settings.",
							Type:        schema.TypeList,
							MaxItems:    1,
							Optional:    true,
							Computed:    true,
							ConfigMode:  schema.SchemaConfigModeAttr,
							Elem:        utilizationSettingsResource,
						},
						"chat": {
							Description: "Chat media settings. If not set, this reverts to the default media type settings.",
							Type:        schema.TypeList,
							MaxItems:    1,
							Optional:    true,
							Computed:    true,
							ConfigMode:  schema.SchemaConfigModeAttr,
							Elem:        utilizationSettingsResource,
						},
						"label_utilizations": {
							Description: "Label utilization settings. If not set, default label settings will be applied. This is in PREVIEW and should not be used unless the feature is available to your organization.",
							Type:        schema.TypeList,
							Optional:    true,
							Computed:    true,
							ConfigMode:  schema.SchemaConfigModeAttr,
							Elem:        utilizationLabelResource,
						},
					},
				},
			},
		},
	}
}

func DataSourceUser() *schema.Resource {
	return &schema.Resource{
		Description:        "Data source for Genesys Cloud Users. Select a user by email or name. If both email & name are specified, the name won't be used for user lookup",
		ReadWithoutTimeout: provider.ReadWithPooledClient(DataSourceUserRead),
		Schema: map[string]*schema.Schema{
			"email": {
				Description: "User email.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"name": {
				Description: "User name.",
				Type:        schema.TypeString,
				Optional:    true,
			},
		},
	}
}

func UserExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: provider.GetAllWithPooledClient(GetAllUsers),
		RefAttrs: map[string]*resourceExporter.RefAttrSettings{
			"manager":                       {RefType: ResourceType},
			"division_id":                   {RefType: "genesyscloud_auth_division"},
			"routing_skills.skill_id":       {RefType: "genesyscloud_routing_skill"},
			"routing_languages.language_id": {RefType: "genesyscloud_routing_language"},
			"locations.location_id":         {RefType: "genesyscloud_location"},
		},
		RemoveIfMissing: map[string][]string{
			"routing_skills":    {"skill_id"},
			"routing_languages": {"language_id"},
			"locations":         {"location_id"},
		},
		AllowZeroValues: []string{"routing_skills.proficiency", "routing_languages.proficiency"},
	}
}
