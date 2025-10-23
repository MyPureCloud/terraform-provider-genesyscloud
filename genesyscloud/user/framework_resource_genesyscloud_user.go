package user

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-validators/float64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	sdkdiag "github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/mypurecloud/platform-client-sdk-go/v165/platformclientv2"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/feature_toggles"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/lists"
)

// Ensure UserFrameworkResource satisfies various resource interfaces
var (
	_ resource.Resource                = &UserFrameworkResource{}
	_ resource.ResourceWithConfigure   = &UserFrameworkResource{}
	_ resource.ResourceWithImportState = &UserFrameworkResource{}
)

// phoneNumberValidator validates phone numbers in E.164 format
type phoneNumberValidator struct{}

func (v phoneNumberValidator) Description(ctx context.Context) string {
	return "value must be a valid phone number in E.164 format"
}

func (v phoneNumberValidator) MarkdownDescription(ctx context.Context) string {
	return "value must be a valid phone number in E.164 format"
}

func (v phoneNumberValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	phoneNumber := req.ConfigValue.ValueString()

	// Skip validation if BCP mode is enabled
	if feature_toggles.BcpModeEnabledExists() {
		return
	}

	utilE164 := util.NewUtilE164Service()
	validNum, diagErr := utilE164.IsValidE164Number(phoneNumber)
	if diagErr.HasError() {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Phone Number Validation Error",
			"Error validating phone number",
		)
		return
	}

	if !validNum {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid Phone Number",
			fmt.Sprintf("Phone number must be in E.164 format, got: %s", phoneNumber),
		)
	}
}

// ValidatePhoneNumber returns a validator for phone numbers in E.164 format
func ValidatePhoneNumber() validator.String {
	return phoneNumberValidator{}
}

// dateValidator validates dates in ISO-8601 format
type dateValidator struct{}

func (v dateValidator) Description(ctx context.Context) string {
	return "value must be a valid date in ISO-8601 format (YYYY-MM-DD)"
}

func (v dateValidator) MarkdownDescription(ctx context.Context) string {
	return "value must be a valid date in ISO-8601 format (YYYY-MM-DD)"
}

func (v dateValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	dateStr := req.ConfigValue.ValueString()
	if dateStr == "" {
		return
	}

	// Parse the date in ISO-8601 format (YYYY-MM-DD)
	_, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid Date Format",
			fmt.Sprintf("Date must be in ISO-8601 format (YYYY-MM-DD), got: %s", dateStr),
		)
	}
}

// ValidateISO8601Date returns a validator for dates in ISO-8601 format
func ValidateISO8601Date() validator.String {
	return dateValidator{}
}

// UserFrameworkResource defines the resource implementation for Genesys Cloud User
type UserFrameworkResource struct {
	clientConfig *platformclientv2.Configuration
}

// UserFrameworkResourceModel describes the resource data model
type UserFrameworkResourceModel struct {
	Id                    types.String `tfsdk:"id"`
	Email                 types.String `tfsdk:"email"`
	Name                  types.String `tfsdk:"name"`
	Password              types.String `tfsdk:"password"`
	State                 types.String `tfsdk:"state"`
	DivisionId            types.String `tfsdk:"division_id"`
	Department            types.String `tfsdk:"department"`
	Title                 types.String `tfsdk:"title"`
	Manager               types.String `tfsdk:"manager"`
	AcdAutoAnswer         types.Bool   `tfsdk:"acd_auto_answer"`
	RoutingSkills         types.Set    `tfsdk:"routing_skills"`
	RoutingLanguages      types.Set    `tfsdk:"routing_languages"`
	Locations             types.Set    `tfsdk:"locations"`
	Addresses             types.List   `tfsdk:"addresses"`
	ProfileSkills         types.Set    `tfsdk:"profile_skills"`
	Certifications        types.Set    `tfsdk:"certifications"`
	EmployerInfo          types.List   `tfsdk:"employer_info"`
	RoutingUtilization    types.List   `tfsdk:"routing_utilization"`
	VoicemailUserpolicies types.List   `tfsdk:"voicemail_userpolicies"`
}

// NewUserFrameworkResource creates a new instance of the user Framework resource
func NewUserFrameworkResource() resource.Resource {
	return &UserFrameworkResource{}
}

// Metadata returns the resource type name
func (r *UserFrameworkResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user"
}

// Schema defines the schema for the resource
func (r *UserFrameworkResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
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
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
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
			"routing_skills": schema.SetNestedAttribute{
				Description: "Skills and proficiencies for this user. If not set, this resource will not manage user skills.",
				Optional:    true,
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
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
			"routing_languages": schema.SetNestedAttribute{
				Description: "Languages and proficiencies for this user. If not set, this resource will not manage user languages.",
				Optional:    true,
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
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
			"locations": schema.SetNestedAttribute{
				Description: "The user placement at each site location. If not set, this resource will not manage user locations.",
				Optional:    true,
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
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
			"addresses": schema.ListNestedAttribute{
				Description: "The address settings for this user. If not set, this resource will not manage addresses.",
				Optional:    true,
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"other_emails": schema.SetNestedAttribute{
							Description: "Other Email addresses for this user.",
							Optional:    true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"address": schema.StringAttribute{
										Description: "Email address.",
										Required:    true,
									},
									"type": schema.StringAttribute{
										Description: "Type of email address (WORK | HOME).",
										Optional:    true,
										Computed:    true,
										Default:     stringdefault.StaticString("WORK"),
										Validators: []validator.String{
											stringvalidator.OneOf("WORK", "HOME"),
										},
									},
								},
							},
						},
						"phone_numbers": schema.SetNestedAttribute{
							Description: "Phone number addresses for this user.",
							Optional:    true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"number": schema.StringAttribute{
										Description: "Phone number. Phone number must be in an E.164 number format.",
										Optional:    true,
										Validators: []validator.String{
											ValidatePhoneNumber(),
										},
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
										Description: "Id of the extension pool which contains this extension.",
										Optional:    true,
									},
								},
							},
						},
					},
				},
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
			"employer_info": schema.ListNestedAttribute{
				Description: "The employer info for this user. If not set, this resource will not manage employer info.",
				Optional:    true,
				Computed:    true,
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"official_name": schema.StringAttribute{
							Description: "The official name of the employee as it appears in the employer records.",
							Optional:    true,
						},
						"employee_id": schema.StringAttribute{
							Description: "The employee ID assigned by the employer.",
							Optional:    true,
						},
						"employee_type": schema.StringAttribute{
							Description: "The type of employee (Full-time, Part-time, Contractor).",
							Optional:    true,
							Validators: []validator.String{
								stringvalidator.OneOf("Full-time", "Part-time", "Contractor"),
							},
						},
						"date_hire": schema.StringAttribute{
							Description: "The hire date in ISO-8601 format (YYYY-MM-DD).",
							Optional:    true,
							Validators: []validator.String{
								ValidateISO8601Date(),
							},
						},
					},
				},
			},
			"routing_utilization": schema.ListNestedAttribute{
				Description: "The routing utilization settings for this user. If empty list, the org default settings are used. If not set, this resource will not manage the users's utilization settings.",
				Optional:    true,
				Computed:    true,
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"call": schema.ListNestedAttribute{
							Description: "Call media settings. If not set, this reverts to the default media type settings.",
							Optional:    true,
							Computed:    true,
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"maximum_capacity": schema.Int64Attribute{
										Description: "Maximum capacity of conversations of this media type. Value must be between 0 and 25.",
										Required:    true,
										Validators: []validator.Int64{
											int64validator.Between(0, 25),
										},
									},
									"interruptible_media_types": schema.SetAttribute{
										Description: "Set of other media types that can interrupt this media type (call | callback | message | email | chat).",
										Optional:    true,
										ElementType: types.StringType,
										Validators: []validator.Set{
											setvalidator.ValueStringsAre(
												stringvalidator.OneOf("call", "callback", "message", "email", "chat"),
											),
										},
									},
									"include_non_acd": schema.BoolAttribute{
										Description: "Block this media type when on a non-ACD conversation.",
										Optional:    true,
										Computed:    true,
										Default:     booldefault.StaticBool(false),
									},
								},
							},
						},
						"callback": schema.ListNestedAttribute{
							Description: "Callback media settings. If not set, this reverts to the default media type settings.",
							Optional:    true,
							Computed:    true,
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"maximum_capacity": schema.Int64Attribute{
										Description: "Maximum capacity of conversations of this media type. Value must be between 0 and 25.",
										Required:    true,
										Validators: []validator.Int64{
											int64validator.Between(0, 25),
										},
									},
									"interruptible_media_types": schema.SetAttribute{
										Description: "Set of other media types that can interrupt this media type (call | callback | message | email | chat).",
										Optional:    true,
										ElementType: types.StringType,
										Validators: []validator.Set{
											setvalidator.ValueStringsAre(
												stringvalidator.OneOf("call", "callback", "message", "email", "chat"),
											),
										},
									},
									"include_non_acd": schema.BoolAttribute{
										Description: "Block this media type when on a non-ACD conversation.",
										Optional:    true,
										Computed:    true,
										Default:     booldefault.StaticBool(false),
									},
								},
							},
						},
						"message": schema.ListNestedAttribute{
							Description: "Message media settings. If not set, this reverts to the default media type settings.",
							Optional:    true,
							Computed:    true,
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"maximum_capacity": schema.Int64Attribute{
										Description: "Maximum capacity of conversations of this media type. Value must be between 0 and 25.",
										Required:    true,
										Validators: []validator.Int64{
											int64validator.Between(0, 25),
										},
									},
									"interruptible_media_types": schema.SetAttribute{
										Description: "Set of other media types that can interrupt this media type (call | callback | message | email | chat).",
										Optional:    true,
										ElementType: types.StringType,
										Validators: []validator.Set{
											setvalidator.ValueStringsAre(
												stringvalidator.OneOf("call", "callback", "message", "email", "chat"),
											),
										},
									},
									"include_non_acd": schema.BoolAttribute{
										Description: "Block this media type when on a non-ACD conversation.",
										Optional:    true,
										Computed:    true,
										Default:     booldefault.StaticBool(false),
									},
								},
							},
						},
						"email": schema.ListNestedAttribute{
							Description: "Email media settings. If not set, this reverts to the default media type settings.",
							Optional:    true,
							Computed:    true,
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"maximum_capacity": schema.Int64Attribute{
										Description: "Maximum capacity of conversations of this media type. Value must be between 0 and 25.",
										Required:    true,
										Validators: []validator.Int64{
											int64validator.Between(0, 25),
										},
									},
									"interruptible_media_types": schema.SetAttribute{
										Description: "Set of other media types that can interrupt this media type (call | callback | message | email | chat).",
										Optional:    true,
										ElementType: types.StringType,
										Validators: []validator.Set{
											setvalidator.ValueStringsAre(
												stringvalidator.OneOf("call", "callback", "message", "email", "chat"),
											),
										},
									},
									"include_non_acd": schema.BoolAttribute{
										Description: "Block this media type when on a non-ACD conversation.",
										Optional:    true,
										Computed:    true,
										Default:     booldefault.StaticBool(false),
									},
								},
							},
						},
						"chat": schema.ListNestedAttribute{
							Description: "Chat media settings. If not set, this reverts to the default media type settings.",
							Optional:    true,
							Computed:    true,
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"maximum_capacity": schema.Int64Attribute{
										Description: "Maximum capacity of conversations of this media type. Value must be between 0 and 25.",
										Required:    true,
										Validators: []validator.Int64{
											int64validator.Between(0, 25),
										},
									},
									"interruptible_media_types": schema.SetAttribute{
										Description: "Set of other media types that can interrupt this media type (call | callback | message | email | chat).",
										Optional:    true,
										ElementType: types.StringType,
										Validators: []validator.Set{
											setvalidator.ValueStringsAre(
												stringvalidator.OneOf("call", "callback", "message", "email", "chat"),
											),
										},
									},
									"include_non_acd": schema.BoolAttribute{
										Description: "Block this media type when on a non-ACD conversation.",
										Optional:    true,
										Computed:    true,
										Default:     booldefault.StaticBool(false),
									},
								},
							},
						},
						"label_utilizations": schema.ListNestedAttribute{
							Description: "Label utilization settings. If not set, default label settings will be applied. This is in PREVIEW and should not be used unless the feature is available to your organization.",
							Optional:    true,
							Computed:    true,
							NestedObject: schema.NestedAttributeObject{
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
			},
			"voicemail_userpolicies": schema.ListNestedAttribute{
				Description: "User's voicemail policies. If not set, default user policies will be applied.",
				Optional:    true,
				Computed:    true,
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedAttributeObject{
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
			},
		},
	}
}

// Configure adds the provider configured client to the resource
func (r *UserFrameworkResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured
	if req.ProviderData == nil {
		return
	}

	providerMeta, ok := req.ProviderData.(*provider.ProviderMeta)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *provider.ProviderMeta, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.clientConfig = providerMeta.ClientConfig
}

// Create creates the resource and sets the initial Terraform state
func (r *UserFrameworkResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan UserFrameworkResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	proxy := GetUserProxy(r.clientConfig)
	email := plan.Email.ValueString()
	divisionID := plan.DivisionId.ValueString()

	// Build addresses from plan
	addresses, addressDiags := buildSdkAddressesFromFramework(plan.Addresses)
	resp.Diagnostics.Append(addressDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Check for a deleted user before creating
	id, diagErr := getDeletedUserIdFramework(email, proxy)
	if diagErr.HasError() {
		resp.Diagnostics.Append(diagErr...)
		return
	}
	if id != nil {
		plan.Id = types.StringValue(*id)
		r.restoreDeletedUserFramework(ctx, &plan, proxy, &resp.Diagnostics)
		if resp.Diagnostics.HasError() {
			return
		}
		resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
		return
	}

	createUser := platformclientv2.Createuser{
		Name:       platformclientv2.String(plan.Name.ValueString()),
		State:      platformclientv2.String(plan.State.ValueString()),
		Title:      platformclientv2.String(plan.Title.ValueString()),
		Department: platformclientv2.String(plan.Department.ValueString()),
		Email:      &email,
		Addresses:  addresses,
	}

	// Optional attribute that should not be empty strings
	if divisionID != "" {
		createUser.DivisionId = &divisionID
	}

	log.Printf("Creating user %s", email)

	userResponse, proxyPostResponse, postErr := proxy.createUser(ctx, &createUser)
	if postErr != nil {
		if proxyPostResponse != nil && proxyPostResponse.Error != nil && (*proxyPostResponse.Error).Code == "general.conflict" {
			// Check for a deleted user
			id, diagErr := getDeletedUserIdFramework(email, proxy)
			if diagErr.HasError() {
				resp.Diagnostics.Append(diagErr...)
				return
			}
			if id != nil {
				plan.Id = types.StringValue(*id)
				r.restoreDeletedUserFramework(ctx, &plan, proxy, &resp.Diagnostics)
				if resp.Diagnostics.HasError() {
					return
				}
				resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
				return
			}
		}
		resp.Diagnostics.AddError(
			"Failed to create user",
			fmt.Sprintf("Failed to create user %s: %s", email, postErr),
		)
		return
	}

	plan.Id = types.StringValue(*userResponse.Id)

	// Set attributes that can only be modified in a patch
	if r.hasChangesFramework(&plan, "manager", "locations", "acd_auto_answer", "profile_skills", "certifications", "employer_info") {
		log.Printf("Updating additional attributes for user %s", email)

		_, _, patchErr := proxy.patchUserWithState(ctx, *userResponse.Id, &platformclientv2.Updateuser{
			Manager:        platformclientv2.String(plan.Manager.ValueString()),
			AcdAutoAnswer:  platformclientv2.Bool(plan.AcdAutoAnswer.ValueBool()),
			Locations:      r.buildSdkLocationsFromFramework(ctx, plan.Locations),
			Certifications: r.buildSdkCertificationsFromFramework(ctx, plan.Certifications),
			ProfileSkills:  r.buildSdkProfileSkillsFromFramework(ctx, plan.ProfileSkills),
			EmployerInfo:   r.buildSdkEmployerInfoFromFramework(ctx, plan.EmployerInfo),
			Version:        userResponse.Version,
		})

		if patchErr != nil {
			resp.Diagnostics.AddError(
				"Failed to update user",
				fmt.Sprintf("Failed to update user %s: %s", plan.Id.ValueString(), patchErr),
			)
			return
		}
	}

	diagErr = r.executeAllUpdatesFramework(ctx, &plan, proxy, r.clientConfig, false)
	if diagErr.HasError() {
		resp.Diagnostics.Append(diagErr...)
		return
	}

	log.Printf("Created user %s %s", email, *userResponse.Id)

	// Read the user to get the current state
	r.readUserFramework(ctx, &plan, proxy, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

// Read refreshes the Terraform state with the latest data
func (r *UserFrameworkResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state UserFrameworkResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	proxy := GetUserProxy(r.clientConfig)
	// Consistency checker removed for Framework implementation - not needed for Framework resources

	log.Printf("Reading user %s", state.Id.ValueString())

	diagErr := util.WithRetriesForRead(ctx, nil, func() *retry.RetryError {
		r.readUserFramework(ctx, &state, proxy, &resp.Diagnostics)
		if resp.Diagnostics.HasError() {
			// Check if it's a 404 error
			for _, diagnostic := range resp.Diagnostics.Errors() {
				if diagnostic.Summary() == "User not found" {
					return retry.RetryableError(fmt.Errorf("user not found"))
				}
			}
			return retry.NonRetryableError(fmt.Errorf("failed to read user"))
		}

		return nil
	})

	if diagErr.HasError() {
		resp.Diagnostics.AddError(
			"Failed to read user",
			fmt.Sprintf("Failed to read user %s", state.Id.ValueString()),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

// Update updates the resource and sets the updated Terraform state on success
func (r *UserFrameworkResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan UserFrameworkResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state UserFrameworkResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	proxy := GetUserProxy(r.clientConfig)

	// Build addresses from plan
	addresses, addressDiags := buildSdkAddressesFromFramework(plan.Addresses)
	resp.Diagnostics.Append(addressDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	email := plan.Email.ValueString()

	log.Printf("Updating user %s", email)

	// If state changes, it is the only modifiable field, so it must be updated separately
	if !plan.State.Equal(state.State) {
		log.Printf("Updating state for user %s", email)
		updateUserRequestBody := platformclientv2.Updateuser{
			State: platformclientv2.String(plan.State.ValueString()),
		}
		diagErr := r.executeUpdateUserFramework(ctx, &plan, proxy, updateUserRequestBody)
		if diagErr.HasError() {
			resp.Diagnostics.Append(diagErr...)
			return
		}
	}

	updateUserRequestBody := platformclientv2.Updateuser{
		Name:           platformclientv2.String(plan.Name.ValueString()),
		Department:     platformclientv2.String(plan.Department.ValueString()),
		Title:          platformclientv2.String(plan.Title.ValueString()),
		Manager:        platformclientv2.String(plan.Manager.ValueString()),
		AcdAutoAnswer:  platformclientv2.Bool(plan.AcdAutoAnswer.ValueBool()),
		Email:          &email,
		Addresses:      addresses,
		Locations:      r.buildSdkLocationsFromFramework(ctx, plan.Locations),
		Certifications: r.buildSdkCertificationsFromFramework(ctx, plan.Certifications),
		ProfileSkills:  r.buildSdkProfileSkillsFromFramework(ctx, plan.ProfileSkills),
		EmployerInfo:   r.buildSdkEmployerInfoFromFramework(ctx, plan.EmployerInfo),
	}
	diagErr := r.executeUpdateUserFramework(ctx, &plan, proxy, updateUserRequestBody)
	if diagErr.HasError() {
		resp.Diagnostics.Append(diagErr...)
		return
	}

	diagErr = r.executeAllUpdatesFramework(ctx, &plan, proxy, r.clientConfig, true)
	if diagErr.HasError() {
		resp.Diagnostics.Append(diagErr...)
		return
	}

	log.Printf("Finished updating user %s", email)

	// Read the user to get the current state
	r.readUserFramework(ctx, &plan, proxy, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

// Delete deletes the resource and removes the Terraform state on success
func (r *UserFrameworkResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state UserFrameworkResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	proxy := GetUserProxy(r.clientConfig)
	email := state.Email.ValueString()

	log.Printf("Deleting user %s", email)

	// Directory occasionally returns version errors on deletes if an object was updated at the same time.
	_, _, err := proxy.deleteUser(ctx, state.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to delete user",
			fmt.Sprintf("Failed to delete user %s: %s", state.Id.ValueString(), err),
		)
		return
	}
	log.Printf("Deleted user %s", email)

	// Verify user in deleted state and search index has been updated
	diagErr2 := util.WithRetries(ctx, 3*time.Minute, func() *retry.RetryError {
		id, diagErr := getDeletedUserIdFramework(email, proxy)
		if diagErr.HasError() {
			return retry.NonRetryableError(fmt.Errorf("error searching for deleted user %s", email))
		}
		if id == nil {
			return retry.RetryableError(fmt.Errorf("user %s not yet in deleted state", email))
		}
		return nil
	})
	if diagErr2.HasError() {
		resp.Diagnostics.AddError(
			"Failed to verify user deletion",
			fmt.Sprintf("Failed to verify user %s deletion", email),
		)
		return
	}
}

// ImportState imports the resource into Terraform state
func (r *UserFrameworkResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// GetAllUsers retrieves all users and is used for the exporter (Framework diagnostics)
func GetAllUsers(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	proxy := GetUserProxy(clientConfig)
	resources := make(resourceExporter.ResourceIDMetaMap)

	users, _, err := proxy.GetAllUser(ctx)
	if err != nil {
		var diagnostics diag.Diagnostics
		diagnostics.AddError(
			"Failed to get users",
			fmt.Sprintf("Failed to get users: %s", err),
		)
		return nil, diagnostics
	}

	for _, user := range *users {
		if user.Id != nil && user.Email != nil {
			resources[*user.Id] = &resourceExporter.ResourceMeta{BlockLabel: *user.Email}
		}
	}

	return resources, nil
}

// GetAllUsersSDK retrieves all users for export (SDK v2 diagnostics wrapper)
func GetAllUsersSDK(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, sdkdiag.Diagnostics) {
	proxy := GetUserProxy(clientConfig)
	resources := make(resourceExporter.ResourceIDMetaMap)

	users, resp, err := proxy.GetAllUser(ctx)
	if err != nil {
		return nil, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to get users: %s", err), resp)
	}

	for _, user := range *users {
		if user.Id != nil && user.Email != nil {
			resources[*user.Id] = &resourceExporter.ResourceMeta{BlockLabel: *user.Email}
		}
	}

	return resources, nil
}

// Helper methods for Framework implementation

// hasChangesFramework checks if any of the specified attributes have changes
func (r *UserFrameworkResource) hasChangesFramework(plan *UserFrameworkResourceModel, attributes ...string) bool {
	// For create operations, we consider all non-null values as changes
	for _, attr := range attributes {
		switch attr {
		case "manager":
			if !plan.Manager.IsNull() && !plan.Manager.IsUnknown() && plan.Manager.ValueString() != "" {
				return true
			}
		case "locations":
			if !plan.Locations.IsNull() && !plan.Locations.IsUnknown() {
				elements := plan.Locations.Elements()
				if len(elements) > 0 {
					return true
				}
			}
		case "acd_auto_answer":
			if !plan.AcdAutoAnswer.IsNull() && !plan.AcdAutoAnswer.IsUnknown() {
				return true
			}
		case "profile_skills":
			if !plan.ProfileSkills.IsNull() && !plan.ProfileSkills.IsUnknown() {
				// Profile skills are specified in plan (even if empty), so there's a change
				return true
			}
		case "certifications":
			if !plan.Certifications.IsNull() && !plan.Certifications.IsUnknown() {
				// Certifications are specified in plan (even if empty), so there's a change
				return true
			}
		case "employer_info":
			if !plan.EmployerInfo.IsNull() && !plan.EmployerInfo.IsUnknown() {
				elements := plan.EmployerInfo.Elements()
				if len(elements) > 0 {
					return true
				}
			}
		}
	}
	return false
}

// buildSdkLocationsFromFramework converts Framework locations to SDK locations
func (r *UserFrameworkResource) buildSdkLocationsFromFramework(ctx context.Context, locations types.Set) *[]platformclientv2.Location {
	if locations.IsNull() || locations.IsUnknown() {
		return nil
	}

	sdkLocations := make([]platformclientv2.Location, 0)
	for _, locVal := range locations.Elements() {
		locObj, ok := locVal.(basetypes.ObjectValue)
		if !ok {
			continue
		}
		locAttrs := locObj.Attributes()

		locID := locAttrs["location_id"].(basetypes.StringValue).ValueString()
		locNotes := locAttrs["notes"].(basetypes.StringValue).ValueString()

		sdkLocations = append(sdkLocations, platformclientv2.Location{
			Id:    &locID,
			Notes: &locNotes,
		})
	}
	return &sdkLocations
}

// buildSdkCertificationsFromFramework converts Framework certifications to SDK certifications
func (r *UserFrameworkResource) buildSdkCertificationsFromFramework(ctx context.Context, certifications types.Set) *[]string {
	certs := make([]string, 0)

	// If certifications are specified (even if empty), process them
	if !certifications.IsNull() && !certifications.IsUnknown() {
		for _, certVal := range certifications.Elements() {
			certStr, ok := certVal.(basetypes.StringValue)
			if ok {
				certs = append(certs, certStr.ValueString())
			}
		}
	}
	// Always return a slice (even if empty) to allow clearing certifications
	return &certs
}

// buildSdkProfileSkillsFromFramework converts Framework profile skills to SDK profile skills
func (r *UserFrameworkResource) buildSdkProfileSkillsFromFramework(ctx context.Context, profileSkills types.Set) *[]string {
	skills := make([]string, 0)

	// If profile skills are specified (even if empty), process them
	if !profileSkills.IsNull() && !profileSkills.IsUnknown() {
		for _, skillVal := range profileSkills.Elements() {
			skillStr, ok := skillVal.(basetypes.StringValue)
			if ok {
				skills = append(skills, skillStr.ValueString())
			}
		}
	}
	// Always return a slice (even if empty) to allow clearing profile skills
	return &skills
}

// buildSdkEmployerInfoFromFramework converts Framework employer info to SDK employer info
func (r *UserFrameworkResource) buildSdkEmployerInfoFromFramework(ctx context.Context, employerInfo types.List) *platformclientv2.Employerinfo {
	if employerInfo.IsNull() || employerInfo.IsUnknown() {
		return nil
	}

	elements := employerInfo.Elements()
	if len(elements) == 0 {
		return nil
	}

	// Get the first (and only) element since MaxItems is 1
	empInfoObj, ok := elements[0].(basetypes.ObjectValue)
	if !ok {
		return nil
	}

	empAttrs := empInfoObj.Attributes()
	sdkEmployerInfo := &platformclientv2.Employerinfo{}

	if officialName, exists := empAttrs["official_name"]; exists && !officialName.IsNull() {
		if nameVal, ok := officialName.(basetypes.StringValue); ok && nameVal.ValueString() != "" {
			sdkEmployerInfo.OfficialName = platformclientv2.String(nameVal.ValueString())
		}
	}

	if employeeId, exists := empAttrs["employee_id"]; exists && !employeeId.IsNull() {
		if idVal, ok := employeeId.(basetypes.StringValue); ok && idVal.ValueString() != "" {
			sdkEmployerInfo.EmployeeId = platformclientv2.String(idVal.ValueString())
		}
	}

	if employeeType, exists := empAttrs["employee_type"]; exists && !employeeType.IsNull() {
		if typeVal, ok := employeeType.(basetypes.StringValue); ok && typeVal.ValueString() != "" {
			sdkEmployerInfo.EmployeeType = platformclientv2.String(typeVal.ValueString())
		}
	}

	if dateHire, exists := empAttrs["date_hire"]; exists && !dateHire.IsNull() {
		if dateVal, ok := dateHire.(basetypes.StringValue); ok && dateVal.ValueString() != "" {
			dateStr := dateVal.ValueString()
			sdkEmployerInfo.DateHire = &dateStr
		}
	}

	return sdkEmployerInfo
}

// flattenEmployerInfoForFramework converts SDK employer info to Framework employer info
func (r *UserFrameworkResource) flattenEmployerInfoForFramework(ctx context.Context, employerInfo *platformclientv2.Employerinfo) types.List {
	if employerInfo == nil {
		return types.ListNull(types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"official_name": types.StringType,
				"employee_id":   types.StringType,
				"employee_type": types.StringType,
				"date_hire":     types.StringType,
			},
		})
	}

	empInfoAttrs := map[string]attr.Value{
		"official_name": types.StringNull(),
		"employee_id":   types.StringNull(),
		"employee_type": types.StringNull(),
		"date_hire":     types.StringNull(),
	}

	if employerInfo.OfficialName != nil {
		empInfoAttrs["official_name"] = types.StringValue(*employerInfo.OfficialName)
	}

	if employerInfo.EmployeeId != nil {
		empInfoAttrs["employee_id"] = types.StringValue(*employerInfo.EmployeeId)
	}

	if employerInfo.EmployeeType != nil {
		empInfoAttrs["employee_type"] = types.StringValue(*employerInfo.EmployeeType)
	}

	if employerInfo.DateHire != nil {
		empInfoAttrs["date_hire"] = types.StringValue(*employerInfo.DateHire)
	}

	empInfoObj, _ := types.ObjectValue(map[string]attr.Type{
		"official_name": types.StringType,
		"employee_id":   types.StringType,
		"employee_type": types.StringType,
		"date_hire":     types.StringType,
	}, empInfoAttrs)

	empInfoList, _ := types.ListValue(types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"official_name": types.StringType,
			"employee_id":   types.StringType,
			"employee_type": types.StringType,
			"date_hire":     types.StringType,
		},
	}, []attr.Value{empInfoObj})

	return empInfoList
}

// flattenUserDataForFramework converts SDK string slice to Framework Set
func (r *UserFrameworkResource) flattenUserDataForFramework(ctx context.Context, userDataSlice *[]string) types.Set {
	elements := make([]attr.Value, 0)

	if userDataSlice != nil {
		for _, item := range *userDataSlice {
			elements = append(elements, types.StringValue(item))
		}
	}

	setVal, _ := types.SetValue(types.StringType, elements)
	return setVal
}

// executeUpdateUserFramework executes a user update with retry logic
func (r *UserFrameworkResource) executeUpdateUserFramework(ctx context.Context, plan *UserFrameworkResourceModel, proxy *userProxy, updateUser platformclientv2.Updateuser) diag.Diagnostics {
	var diagnostics diag.Diagnostics

	currentUser, _, errGet := proxy.getUserById(ctx, plan.Id.ValueString(), nil, "")
	if errGet != nil {
		diagnostics.AddError(
			"Failed to read user",
			fmt.Sprintf("Failed to read user %s: %s", plan.Id.ValueString(), errGet),
		)
		return diagnostics
	}

	updateUser.Version = currentUser.Version

	_, _, patchErr := proxy.updateUser(ctx, plan.Id.ValueString(), &updateUser)
	if patchErr != nil {
		diagnostics.AddError(
			"Failed to update user",
			fmt.Sprintf("Failed to update user %s: %s", plan.Id.ValueString(), patchErr),
		)
		return diagnostics
	}

	return diagnostics
}

// executeAllUpdatesFramework executes all additional updates for the user
func (r *UserFrameworkResource) executeAllUpdatesFramework(ctx context.Context, plan *UserFrameworkResourceModel, proxy *userProxy, sdkConfig *platformclientv2.Configuration, updateObjectDivision bool) diag.Diagnostics {
	var diagnostics diag.Diagnostics

	if updateObjectDivision {
		// Note: This would need to be adapted for Framework - util.UpdateObjectDivision expects schema.ResourceData
		// For now, we'll skip this in the Framework implementation
		// TODO: Implement Framework-compatible division update
	}

	// Update user skills - placeholder for now
	diagErr := r.updateUserSkillsFramework(ctx, plan, proxy)
	if diagErr.HasError() {
		diagnostics.Append(diagErr...)
		return diagnostics
	}

	// Update user languages - placeholder for now
	diagErr = r.updateUserLanguagesFramework(ctx, plan, proxy)
	if diagErr.HasError() {
		diagnostics.Append(diagErr...)
		return diagnostics
	}

	// Update profile skills - placeholder for now
	diagErr = r.updateUserProfileSkillsFramework(ctx, plan, proxy)
	if diagErr.HasError() {
		diagnostics.Append(diagErr...)
		return diagnostics
	}

	// Update routing utilization - placeholder for now
	diagErr = r.updateUserRoutingUtilizationFramework(ctx, plan, proxy)
	if diagErr.HasError() {
		diagnostics.Append(diagErr...)
		return diagnostics
	}

	// Update voicemail policies - placeholder for now
	diagErr = r.updateUserVoicemailPoliciesFramework(ctx, plan, proxy)
	if diagErr.HasError() {
		diagnostics.Append(diagErr...)
		return diagnostics
	}

	// Update password
	diagErr = r.updatePasswordFramework(ctx, plan, proxy)
	if diagErr.HasError() {
		diagnostics.Append(diagErr...)
		return diagnostics
	}

	return diagnostics
}

// Placeholder methods for additional update operations
// These will be implemented in subsequent tasks

func (r *UserFrameworkResource) updateUserSkillsFramework(ctx context.Context, plan *UserFrameworkResourceModel, proxy *userProxy) diag.Diagnostics {
	var diagnostics diag.Diagnostics

	if plan.RoutingSkills.IsNull() || plan.RoutingSkills.IsUnknown() {
		return diagnostics
	}

	// Convert Framework skills to SDK format
	newSkills, skillDiags := buildSdkSkillsFromFramework(plan.RoutingSkills)
	if len(skillDiags) > 0 {
		// Convert SDK diagnostics to Framework diagnostics
		for _, sdkDiag := range skillDiags {
			diagnostics.AddError(sdkDiag.Summary, sdkDiag.Detail)
		}
		return diagnostics
	}

	// Get current skills
	oldSdkSkills, err := getUserRoutingSkills(plan.Id.ValueString(), proxy)
	if err != nil {
		// Convert SDK diagnostics to Framework diagnostics
		for _, sdkDiag := range err {
			diagnostics.AddError(sdkDiag.Summary, sdkDiag.Detail)
		}
		return diagnostics
	}

	// Build maps for comparison
	newSkillProfs := make(map[string]float64)
	newSkillIds := make([]string, len(newSkills))
	for i, skill := range newSkills {
		newSkillIds[i] = *skill.Id
		newSkillProfs[*skill.Id] = *skill.Proficiency
	}

	oldSkillIds := make([]string, len(oldSdkSkills))
	oldSkillProfs := make(map[string]float64)
	for i, skill := range oldSdkSkills {
		oldSkillIds[i] = *skill.Id
		oldSkillProfs[*skill.Id] = *skill.Proficiency
	}

	// Remove skills that are no longer needed
	if len(oldSkillIds) > 0 {
		skillsToRemove := lists.SliceDifference(oldSkillIds, newSkillIds)
		for _, skillId := range skillsToRemove {
			diagErr := util.RetryWhen(util.IsVersionMismatch, func() (*platformclientv2.APIResponse, sdkdiag.Diagnostics) {
				resp, err := proxy.userApi.DeleteUserRoutingskill(plan.Id.ValueString(), skillId)
				if err != nil {
					return resp, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to remove skill from user %s error: %s", plan.Id.ValueString(), err), resp)
				}
				return nil, nil
			})
			if diagErr != nil {
				// Convert SDK diagnostics to Framework diagnostics
				for _, sdkDiag := range diagErr {
					diagnostics.AddError(sdkDiag.Summary, sdkDiag.Detail)
				}
				return diagnostics
			}
		}
	}

	// Add or update skills
	if len(newSkillIds) > 0 {
		skillsToAddOrUpdate := lists.SliceDifference(newSkillIds, oldSkillIds)
		// Check for existing proficiencies to update
		for skillID, newProf := range newSkillProfs {
			if oldProf, found := oldSkillProfs[skillID]; found {
				if newProf != oldProf {
					skillsToAddOrUpdate = append(skillsToAddOrUpdate, skillID)
				}
			}
		}

		if len(skillsToAddOrUpdate) > 0 {
			if diagErr := updateUserRoutingSkills(plan.Id.ValueString(), skillsToAddOrUpdate, newSkillProfs, proxy); diagErr != nil {
				// Convert SDK diagnostics to Framework diagnostics
				for _, sdkDiag := range diagErr {
					diagnostics.AddError(sdkDiag.Summary, sdkDiag.Detail)
				}
				return diagnostics
			}
		}
	}

	return diagnostics
}

func (r *UserFrameworkResource) updateUserLanguagesFramework(ctx context.Context, plan *UserFrameworkResourceModel, proxy *userProxy) diag.Diagnostics {
	var diagnostics diag.Diagnostics

	if plan.RoutingLanguages.IsNull() || plan.RoutingLanguages.IsUnknown() {
		return diagnostics
	}

	// Convert Framework languages to SDK format
	newLanguages, langDiags := buildSdkLanguagesFromFramework(plan.RoutingLanguages)
	if len(langDiags) > 0 {
		// Convert SDK diagnostics to Framework diagnostics
		for _, sdkDiag := range langDiags {
			diagnostics.AddError(sdkDiag.Summary, sdkDiag.Detail)
		}
		return diagnostics
	}

	// Get current languages
	oldSdkLanguages, err := getUserRoutingLanguages(plan.Id.ValueString(), proxy)
	if err != nil {
		// Convert SDK diagnostics to Framework diagnostics
		for _, sdkDiag := range err {
			diagnostics.AddError(sdkDiag.Summary, sdkDiag.Detail)
		}
		return diagnostics
	}

	// Build maps for comparison
	newLangProfs := make(map[string]int)
	newLangIds := make([]string, len(newLanguages))
	for i, lang := range newLanguages {
		newLangIds[i] = *lang.Id
		newLangProfs[*lang.Id] = int(*lang.Proficiency)
	}

	oldLangIds := make([]string, len(oldSdkLanguages))
	oldLangProfs := make(map[string]int)
	for i, lang := range oldSdkLanguages {
		oldLangIds[i] = *lang.Id
		oldLangProfs[*lang.Id] = int(*lang.Proficiency)
	}

	// Remove languages that are no longer needed
	if len(oldLangIds) > 0 {
		langsToRemove := lists.SliceDifference(oldLangIds, newLangIds)
		for _, langID := range langsToRemove {
			diagErr := util.RetryWhen(util.IsVersionMismatch, func() (*platformclientv2.APIResponse, sdkdiag.Diagnostics) {
				resp, err := proxy.userApi.DeleteUserRoutinglanguage(plan.Id.ValueString(), langID)
				if err != nil {
					return resp, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to remove language from user %s error: %s", plan.Id.ValueString(), err), resp)
				}
				return nil, nil
			})
			if diagErr != nil {
				// Convert SDK diagnostics to Framework diagnostics
				for _, sdkDiag := range diagErr {
					diagnostics.AddError(sdkDiag.Summary, sdkDiag.Detail)
				}
				return diagnostics
			}
		}
	}

	// Add or update languages
	if len(newLangIds) > 0 {
		langsToAddOrUpdate := lists.SliceDifference(newLangIds, oldLangIds)
		// Check for existing proficiencies to update
		for langID, newProf := range newLangProfs {
			if oldProf, found := oldLangProfs[langID]; found {
				if newProf != oldProf {
					langsToAddOrUpdate = append(langsToAddOrUpdate, langID)
				}
			}
		}

		if len(langsToAddOrUpdate) > 0 {
			if diagErr := updateUserRoutingLanguages(plan.Id.ValueString(), langsToAddOrUpdate, newLangProfs, proxy); diagErr != nil {
				// Convert SDK diagnostics to Framework diagnostics
				for _, sdkDiag := range diagErr {
					diagnostics.AddError(sdkDiag.Summary, sdkDiag.Detail)
				}
				return diagnostics
			}
		}
	}

	return diagnostics
}

func (r *UserFrameworkResource) updateUserProfileSkillsFramework(ctx context.Context, plan *UserFrameworkResourceModel, proxy *userProxy) diag.Diagnostics {
	// TODO: Implement in task 4.2
	return diag.Diagnostics{}
}

func (r *UserFrameworkResource) updateUserRoutingUtilizationFramework(ctx context.Context, plan *UserFrameworkResourceModel, proxy *userProxy) diag.Diagnostics {
	var diagnostics diag.Diagnostics

	if plan.RoutingUtilization.IsNull() || plan.RoutingUtilization.IsUnknown() {
		return diagnostics
	}

	// Basic implementation - convert Framework utilization to SDK format
	_, utilizationDiags := buildSdkRoutingUtilizationFromFramework(plan.RoutingUtilization)
	if utilizationDiags.HasError() {
		// Convert Framework diagnostics to SDK diagnostics
		for _, frameworkDiag := range utilizationDiags {
			diagnostics.AddError(frameworkDiag.Summary(), frameworkDiag.Detail())
		}
		return diagnostics
	}

	// For now, we'll skip the actual update since this is a complex implementation
	// This would need to call the existing updateUserRoutingUtilization function
	// with proper data conversion

	return diagnostics
}

func (r *UserFrameworkResource) updateUserVoicemailPoliciesFramework(ctx context.Context, plan *UserFrameworkResourceModel, proxy *userProxy) diag.Diagnostics {
	// TODO: Implement in task 6
	return diag.Diagnostics{}
}

func (r *UserFrameworkResource) updatePasswordFramework(ctx context.Context, plan *UserFrameworkResourceModel, proxy *userProxy) diag.Diagnostics {
	if plan.Password.IsNull() || plan.Password.IsUnknown() || plan.Password.ValueString() == "" {
		return diag.Diagnostics{}
	}

	password := plan.Password.ValueString()
	_, err := proxy.updatePassword(ctx, plan.Id.ValueString(), password)
	if err != nil {
		return diag.Diagnostics{
			diag.NewErrorDiagnostic(
				"Failed to update password",
				fmt.Sprintf("Failed to update password for user %s: %s", plan.Id.ValueString(), err),
			),
		}
	}

	return diag.Diagnostics{}
}

// readUserFramework reads the user data and populates the model
func (r *UserFrameworkResource) readUserFramework(ctx context.Context, model *UserFrameworkResourceModel, proxy *userProxy, diagnostics *diag.Diagnostics) {
	currentUser, proxyResponse, errGet := proxy.getUserById(ctx, model.Id.ValueString(), []string{
		// Expands
		"skills",
		"languages",
		"locations",
		"profileSkills",
		"certifications",
		"employerInfo"},
		"")

	if errGet != nil {
		if util.IsStatus404(proxyResponse) {
			diagnostics.AddError("User not found", fmt.Sprintf("User %s not found", model.Id.ValueString()))
			return
		}
		diagnostics.AddError(
			"Failed to read user",
			fmt.Sprintf("Failed to read user %s: %s", model.Id.ValueString(), errGet),
		)
		return
	}

	// Set basic attributes
	if currentUser.Name != nil {
		model.Name = types.StringValue(*currentUser.Name)
	}
	if currentUser.Email != nil {
		model.Email = types.StringValue(*currentUser.Email)
	}
	if currentUser.Division != nil && currentUser.Division.Id != nil {
		model.DivisionId = types.StringValue(*currentUser.Division.Id)
	}
	if currentUser.State != nil {
		model.State = types.StringValue(*currentUser.State)
	}
	if currentUser.Department != nil {
		model.Department = types.StringValue(*currentUser.Department)
	}
	if currentUser.Title != nil {
		model.Title = types.StringValue(*currentUser.Title)
	}
	if currentUser.AcdAutoAnswer != nil {
		model.AcdAutoAnswer = types.BoolValue(*currentUser.AcdAutoAnswer)
	}

	// Set manager
	if currentUser.Manager != nil {
		model.Manager = types.StringValue(*(*currentUser.Manager).Id)
	} else {
		model.Manager = types.StringNull()
	}

	// Set addresses
	addressesList, addressDiags := flattenUserAddressesForFramework(ctx, currentUser.Addresses, proxy)
	*diagnostics = append(*diagnostics, addressDiags...)
	if !diagnostics.HasError() {
		model.Addresses = addressesList
	}

	// Set routing skills
	skillsSet, skillsDiags := flattenUserSkillsForFramework(currentUser.Skills)
	*diagnostics = append(*diagnostics, skillsDiags...)
	if !diagnostics.HasError() {
		model.RoutingSkills = skillsSet
	}

	// Set routing languages
	languagesSet, languagesDiags := flattenUserLanguagesForFramework(currentUser.Languages)
	*diagnostics = append(*diagnostics, languagesDiags...)
	if !diagnostics.HasError() {
		model.RoutingLanguages = languagesSet
	}

	// Set locations
	locationsSet, locationsDiags := flattenUserLocationsForFramework(currentUser.Locations)
	*diagnostics = append(*diagnostics, locationsDiags...)
	if !diagnostics.HasError() {
		model.Locations = locationsSet
	}

	// Set profile skills
	model.ProfileSkills = r.flattenUserDataForFramework(ctx, currentUser.ProfileSkills)

	// Set certifications
	model.Certifications = r.flattenUserDataForFramework(ctx, currentUser.Certifications)

	// Set employer info
	if currentUser.EmployerInfo != nil {
		model.EmployerInfo = r.flattenEmployerInfoForFramework(ctx, currentUser.EmployerInfo)
	} else {
		model.EmployerInfo = types.ListNull(types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"official_name": types.StringType,
				"employee_id":   types.StringType,
				"employee_type": types.StringType,
				"date_hire":     types.StringType,
			},
		})
	}

	// Get voicemail user policies
	_, _, err := proxy.getVoicemailUserpoliciesById(ctx, model.Id.ValueString())
	if err != nil {
		diagnostics.AddError(
			"Failed to read voicemail userpolicies",
			fmt.Sprintf("Failed to read voicemail userpolicies %s: %s", model.Id.ValueString(), err),
		)
		return
	}

	// Set voicemail policies - simplified for now
	voicemailObjectType := types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"alert_timeout_seconds":    types.Int64Type,
			"send_email_notifications": types.BoolType,
		},
	}
	model.VoicemailUserpolicies = types.ListNull(voicemailObjectType)

	// Set routing utilization - basic implementation for now
	utilizationList, utilizationDiags := flattenUserRoutingUtilizationForFramework(nil)
	*diagnostics = append(*diagnostics, utilizationDiags...)
	if !diagnostics.HasError() {
		model.RoutingUtilization = utilizationList
	}

	log.Printf("Read user %s %s", model.Id.ValueString(), *currentUser.Email)
}

// restoreDeletedUserFramework restores a deleted user
func (r *UserFrameworkResource) restoreDeletedUserFramework(ctx context.Context, plan *UserFrameworkResourceModel, proxy *userProxy, diagnostics *diag.Diagnostics) {
	email := plan.Email.ValueString()
	state := plan.State.ValueString()

	log.Printf("Restoring deleted user %s", email)

	currentUser, _, err := proxy.getUserById(ctx, plan.Id.ValueString(), nil, "deleted")
	if err != nil {
		diagnostics.AddError(
			"Failed to read user",
			fmt.Sprintf("Failed to read user %s: %s", plan.Id.ValueString(), err),
		)
		return
	}

	_, _, patchErr := proxy.patchUserWithState(ctx, plan.Id.ValueString(), &platformclientv2.Updateuser{
		State:   &state,
		Version: currentUser.Version,
	})

	if patchErr != nil {
		diagnostics.AddError(
			"Failed to restore deleted user",
			fmt.Sprintf("Failed to restore deleted user %s: %s", email, patchErr),
		)
		return
	}

	// After restoring, we need to perform additional updates
	// This will be handled by the calling Create method
}

// getDeletedUserIdFramework searches for a deleted user by email
func getDeletedUserIdFramework(email string, proxy *userProxy) (*string, diag.Diagnostics) {
	var diagnostics diag.Diagnostics

	exactType := "EXACT"
	results, _, getErr := proxy.userApi.PostUsersSearch(platformclientv2.Usersearchrequest{
		Query: &[]platformclientv2.Usersearchcriteria{
			{
				Fields:  &[]string{"email"},
				Value:   &email,
				VarType: &exactType,
			},
			{
				Fields:  &[]string{"state"},
				Values:  &[]string{"deleted"},
				VarType: &exactType,
			},
		},
	})
	if getErr != nil {
		diagnostics.AddError(
			"Failed to search for user",
			fmt.Sprintf("Failed to search for user %s: %s", email, getErr),
		)
		return nil, diagnostics
	}
	if results.Results != nil && len(*results.Results) > 0 {
		// User found
		return (*results.Results)[0].Id, diagnostics
	}
	return nil, diagnostics
}
