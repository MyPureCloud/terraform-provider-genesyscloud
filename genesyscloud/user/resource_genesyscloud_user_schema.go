package user

import (
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/float64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	listvalidator "github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	datasourceschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	registrar "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_register"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/phoneplan"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/validators"
)

const ResourceType = "genesyscloud_user"

// SetRegistrar registers all the resources and exporters in the package
func SetRegistrar(l registrar.Registrar) {
	l.RegisterFrameworkDataSource(ResourceType, NewUserFrameworkDataSource)
	l.RegisterFrameworkResource(ResourceType, NewUserFrameworkResource)
	l.RegisterExporter(ResourceType, UserExporter())
}

// UserDataSourceSchema returns the schema for the user data source following SDK DataSourceUser() pattern
func UserDataSourceSchema() datasourceschema.Schema {
	return datasourceschema.Schema{
		Description: "Data source for Genesys Cloud Users. Select a user by email or name. If both email & name are specified, the name won't be used for user lookup",
		Attributes: map[string]datasourceschema.Attribute{
			"id": datasourceschema.StringAttribute{
				Description: "The ID of the user.",
				Computed:    true,
			},
			"email": datasourceschema.StringAttribute{
				Description: "User email.",
				Optional:    true,
			},
			"name": datasourceschema.StringAttribute{
				Description: "User name.",
				Optional:    true,
			},
		},
	}
}

// UserResourceSchema returns the schema for the user resource following SDK ResourceUser() pattern
func UserResourceSchema() schema.Schema {
	return schema.Schema{
		Description: `Genesys Cloud User.

Export block label: "{email}"`,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The ID of the user.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"email": schema.StringAttribute{
				Description: "User's primary email and username.",
				Required:    true,
			},
			"name": schema.StringAttribute{
				Description: "User's full name.",
				Required:    true,
			},
			"password": schema.StringAttribute{
				Description: "User's password. If specified, this is only set on user create.",
				Optional:    true,
				Sensitive:   true,
			},
			"state": schema.StringAttribute{
				Description: "User's state (active | inactive). Default is 'active'.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("active"),
				Validators: []validator.String{
					stringvalidator.OneOf("active", "inactive"),
				},
			},
			"division_id": schema.StringAttribute{
				Description: "The division to which this user will belong. If not set, the home division will be used.",
				Optional:    true,
				Computed:    true,
			},
			"department": schema.StringAttribute{
				Description: "User's department.",
				Optional:    true,
			},
			"title": schema.StringAttribute{
				Description: "User's title.",
				Optional:    true,
			},
			"manager": schema.StringAttribute{
				Description: "User ID of this user's manager.",
				Optional:    true,
			},
			"acd_auto_answer": schema.BoolAttribute{
				Description: "Enable ACD auto-answer.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"profile_skills": schema.SetAttribute{
				Description: "Profile skills for this user. If not set, this resource will not manage profile skills.",
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
			},
			"certifications": schema.SetAttribute{
				Description: "Certifications for this user. If not set, this resource will not manage certifications.",
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
			},
		},
		Blocks: map[string]schema.Block{
			//NOTE:
			// In the Terraform Plugin Framework, SetNestedBlock (and ListNestedBlock)
			// does NOT support the `Optional` or `Required` fields like attributes do.
			// To emulate SDKv2 behavior where `routing_languages` was both Optional and Computed,
			// we use a SetNestedBlock WITHOUT `Optional`, and apply a plan modifier:
			//   setplanmodifier.UseStateForUnknown()
			// This causes Terraform to retain prior state if the block is omitted in config,
			// which mirrors how Computed: true worked in the SDKv2 TypeSet.
			// Inner attributes should remain Required: true to enforce valid entries.
			"routing_skills": schema.SetNestedBlock{
				Description: "Skills and proficiencies for this user. If not set, this resource will not manage user skills.",
				// Emulate SDKv2 Computed at the container level:
				// - When config omits routing_skills, keep prior state (acts like Computed)
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.UseStateForUnknown(),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"skill_id": schema.StringAttribute{
							Description: "ID of routing skill.",
							Required:    true,
						},
						"proficiency": schema.Float64Attribute{
							Description: "Rating from 0.0 to 5.0 on how competent an agent is for a particular skill. It is used when a queue is set to 'Best available skills' mode to allow acd interactions to target agents with higher proficiency ratings.",
							Required:    true,
							Validators: []validator.Float64{
								float64validator.Between(0, 5),
							},
						},
					},
				},
			},
			"routing_languages": schema.SetNestedBlock{
				Description: "Languages and proficiencies for this user. If not set, this resource will not manage user languages.",
				// Emulate SDKv2 Computed at the container level:
				// - When config omits routing_languages, keep prior state (acts like Computed)
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.UseStateForUnknown(),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"language_id": schema.StringAttribute{
							Description: "ID of routing language.",
							Required:    true,
						},
						"proficiency": schema.Int64Attribute{
							Description: "Proficiency is a rating from 0 to 5 on how competent an agent is for a particular language. It is used when a queue is set to 'Best available language' mode to allow acd interactions to target agents with higher proficiency ratings.",
							Required:    true,
							Validators: []validator.Int64{
								int64validator.Between(0, 5),
							},
						},
					},
				},
			},
			"locations": schema.SetNestedBlock{
				Description: "The user placement at each site location. If not set, this resource will not manage user locations.",
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.UseStateForUnknown(),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"location_id": schema.StringAttribute{
							Description: "ID of location.",
							Required:    true,
						},
						"notes": schema.StringAttribute{
							Description: "Optional description on the user's location.",
							Optional:    true,
						},
					},
				},
			},
			"addresses": schema.ListNestedBlock{
				Description: "The address settings for this user. If not set, this resource will not manage addresses.",
				PlanModifiers: []planmodifier.List{
					phoneplan.ClearOnOmitList{},
				},
				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						"other_emails": schema.SetNestedBlock{
							Description: "Other Email addresses for this user.",
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"address": schema.StringAttribute{
										Description: "Email address.",
										Required:    true,
									},
									"type": schema.StringAttribute{
										Description: "Type of email address (WORK | HOME).",
										Optional:    true,
										// PF rule: Attributes with Default must be Computed. We keep Optional+Computed+Default to mirror SDK v2
										// behavior where omitting this field yields "WORK", while still allowing user overrides.
										Computed: true,
										Default:  stringdefault.StaticString("WORK"),
										Validators: []validator.String{
											stringvalidator.OneOf("WORK", "HOME"),
										},
									},
								},
							},
						},
						"phone_numbers": schema.SetNestedBlock{
							Description: "Phone number addresses for this user.",
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"number": schema.StringAttribute{
										Description: "Phone number. Phone number must be in an E.164 number format.",
										Optional:    true,
										Validators:  []validator.String{validators.FWValidatePhoneNumber()},
										// Safe E.164 canonicalization: noop if null/unknown/empty or parse fails
										PlanModifiers: []planmodifier.String{phoneplan.E164{DefaultRegion: "US"}},
									},
									"media_type": schema.StringAttribute{
										Description: "Media type of phone number (SMS | PHONE).",
										Optional:    true,
										Computed:    true,
										Default:     stringdefault.StaticString("PHONE"),
										Validators: []validator.String{
											stringvalidator.OneOf("PHONE", "SMS"),
										},
									},
									"type": schema.StringAttribute{
										Description: "Type of number (WORK | WORK2 | WORK3 | WORK4 | HOME | MOBILE | OTHER).",
										Optional:    true,
										Computed:    true,
										Default:     stringdefault.StaticString("WORK"),
										Validators: []validator.String{
											stringvalidator.OneOf("WORK", "WORK2", "WORK3", "WORK4", "HOME", "MOBILE", "OTHER"),
										},
									},
									"extension": schema.StringAttribute{
										Description: "Phone number extension",
										Optional:    true,
									},
									"extension_pool_id": schema.StringAttribute{
										Description:   "Id of the extension pool which contains this extension.",
										Optional:      true,
										PlanModifiers: []planmodifier.String{phoneplan.NullIfEmpty{}},
									},
								},
							},
						},
					},
				},
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
			},
			"employer_info": schema.ListNestedBlock{
				Description: "The employer info for this user. If not set, this resource will not manage employer info.",
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"official_name": schema.StringAttribute{
							Description: "User's official name.",
							Optional:    true,
						},
						"employee_id": schema.StringAttribute{
							Description: "Employee ID.",
							Optional:    true,
						},
						"employee_type": schema.StringAttribute{
							Description: "Employee type (Full-time | Part-time | Contractor).",
							Optional:    true,
							Validators: []validator.String{
								stringvalidator.OneOf("Full-time", "Part-time", "Contractor"),
							},
						},
						"date_hire": schema.StringAttribute{
							Description: "Hiring date. Dates must be an ISO-8601 string. For example: yyyy-MM-dd.",
							Optional:    true,
							Validators: []validator.String{
								validators.FWValidateDate(),
							},
						},
					},
				},
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
			},
			"routing_utilization": schema.ListNestedBlock{
				Description: "The routing utilization settings for this user. If empty list, the org default settings are used. If not set, this resource will not manage the users's utilization settings.",
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						"call": schema.ListNestedBlock{
							Description: "Call media settings. If not set, this reverts to the default media type settings.",
							PlanModifiers: []planmodifier.List{
								listplanmodifier.UseStateForUnknown(),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"maximum_capacity": schema.Int64Attribute{
										Description: "Maximum capacity of conversations of this media type. Value must be between 0 and 25.",
										Required:    true,
										Validators: []validator.Int64{
											int64validator.Between(0, 25),
										},
									},
									"interruptible_media_types": schema.SetAttribute{
										Description: fmt.Sprintf("Set of other media types that can interrupt this media type (%s).", strings.Join(getSdkUtilizationTypes(), " | ")),
										Optional:    true,
										ElementType: types.StringType,
									},
									"include_non_acd": schema.BoolAttribute{
										Description: "Block this media type when on a non-ACD conversation.",
										Optional:    true,
										Computed:    true,
										Default:     booldefault.StaticBool(false),
									},
								},
							},
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
						},
						"callback": schema.ListNestedBlock{
							Description: "Callback media settings. If not set, this reverts to the default media type settings.",
							PlanModifiers: []planmodifier.List{
								listplanmodifier.UseStateForUnknown(),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"maximum_capacity": schema.Int64Attribute{
										Description: "Maximum capacity of conversations of this media type. Value must be between 0 and 25.",
										Required:    true,
										Validators: []validator.Int64{
											int64validator.Between(0, 25),
										},
									},
									"interruptible_media_types": schema.SetAttribute{
										Description: "Set of other media types that can interrupt this media type.",
										Optional:    true,
										ElementType: types.StringType,
									},
									"include_non_acd": schema.BoolAttribute{
										Description: "Block this media type when on a non-ACD conversation.",
										Optional:    true,
										Computed:    true,
										Default:     booldefault.StaticBool(false),
									},
								},
							},
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
						},
						"message": schema.ListNestedBlock{
							Description: "Message media settings. If not set, this reverts to the default media type settings.",
							PlanModifiers: []planmodifier.List{
								listplanmodifier.UseStateForUnknown(),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"maximum_capacity": schema.Int64Attribute{
										Description: "Maximum capacity of conversations of this media type. Value must be between 0 and 25.",
										Required:    true,
										Validators: []validator.Int64{
											int64validator.Between(0, 25),
										},
									},
									"interruptible_media_types": schema.SetAttribute{
										Description: "Set of other media types that can interrupt this media type.",
										Optional:    true,
										ElementType: types.StringType,
									},
									"include_non_acd": schema.BoolAttribute{
										Description: "Block this media type when on a non-ACD conversation.",
										Optional:    true,
										Computed:    true,
										Default:     booldefault.StaticBool(false),
									},
								},
							},
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
						},
						"email": schema.ListNestedBlock{
							Description: "Email media settings. If not set, this reverts to the default media type settings.",
							PlanModifiers: []planmodifier.List{
								listplanmodifier.UseStateForUnknown(),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"maximum_capacity": schema.Int64Attribute{
										Description: "Maximum capacity of conversations of this media type. Value must be between 0 and 25.",
										Required:    true,
										Validators: []validator.Int64{
											int64validator.Between(0, 25),
										},
									},
									"interruptible_media_types": schema.SetAttribute{
										Description: "Set of other media types that can interrupt this media type.",
										Optional:    true,
										ElementType: types.StringType,
									},
									"include_non_acd": schema.BoolAttribute{
										Description: "Block this media type when on a non-ACD conversation.",
										Optional:    true,
										Computed:    true,
										Default:     booldefault.StaticBool(false),
									},
								},
							},
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
						},
						"chat": schema.ListNestedBlock{
							Description: "Chat media settings. If not set, this reverts to the default media type settings.",
							PlanModifiers: []planmodifier.List{
								listplanmodifier.UseStateForUnknown(),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"maximum_capacity": schema.Int64Attribute{
										Description: "Maximum capacity of conversations of this media type. Value must be between 0 and 25.",
										Required:    true,
										Validators: []validator.Int64{
											int64validator.Between(0, 25),
										},
									},
									"interruptible_media_types": schema.SetAttribute{
										Description: "Set of other media types that can interrupt this media type.",
										Optional:    true,
										ElementType: types.StringType,
									},
									"include_non_acd": schema.BoolAttribute{
										Description: "Block this media type when on a non-ACD conversation.",
										Optional:    true,
										Computed:    true,
										Default:     booldefault.StaticBool(false),
									},
								},
							},
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
						},
						"label_utilizations": schema.ListNestedBlock{
							Description: "Label utilization settings. If not set, default label settings will be applied. This is in PREVIEW and should not be used unless the feature is available to your organization.",
							PlanModifiers: []planmodifier.List{
								listplanmodifier.UseStateForUnknown(),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"label_id": schema.StringAttribute{
										Description: "Id of the label being configured.",
										Required:    true,
									},
									"maximum_capacity": schema.Int64Attribute{
										Description: "Maximum capacity of conversations with this label. Value must be between 0 and 25.",
										Required:    true,
										Validators: []validator.Int64{
											int64validator.Between(0, 25),
										},
									},
									"interrupting_label_ids": schema.SetAttribute{
										Description: "Set of other labels that can interrupt this label.",
										Optional:    true,
										ElementType: types.StringType,
									},
								},
							},
						},
					},
				},
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
			},
			"voicemail_userpolicies": schema.ListNestedBlock{
				Description: "User's voicemail policies. If not set, default user policies will be applied.",
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"alert_timeout_seconds": schema.Int64Attribute{
							Description: "The number of seconds to ring the user's phone before a call is transferred to voicemail.",
							Optional:    true,
						},
						"send_email_notifications": schema.BoolAttribute{
							Description: "Whether email notifications are sent to the user when a new voicemail is received.",
							Optional:    true,
							Computed:    true,
						},
					},
				},
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
			},
		},
	}
}

func UserExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: provider.GetAllWithPooledClient(GetAllUsersSDK),
		RefAttrs: map[string]*resourceExporter.RefAttrSettings{
			"manager":                                   {RefType: ResourceType},
			"division_id":                               {RefType: "genesyscloud_auth_division"},
			"routing_skills.skill_id":                   {RefType: "genesyscloud_routing_skill"},
			"routing_languages.language_id":             {RefType: "genesyscloud_routing_language"},
			"locations.location_id":                     {RefType: "genesyscloud_location"},
			"addresses.phone_numbers.extension_pool_id": {RefType: "genesyscloud_telephony_providers_edges_extension_pool"},
		},
		RemoveIfMissing: map[string][]string{
			"routing_skills":         {"skill_id"},
			"routing_languages":      {"language_id"},
			"locations":              {"location_id"},
			"voicemail_userpolicies": {"alert_timeout_seconds"},
		},
		AllowEmptyArrays: []string{"routing_skills", "routing_languages"},
		AllowZeroValues: []string{
			"routing_skills.proficiency",
			"routing_languages.proficiency",
			//added for PF migration
			"routing_utilization.call.maximum_capacity",
			"routing_utilization.callback.maximum_capacity",
			"routing_utilization.chat.maximum_capacity",
			"routing_utilization.email.maximum_capacity",
			"routing_utilization.message.maximum_capacity",
		},
	}
}
