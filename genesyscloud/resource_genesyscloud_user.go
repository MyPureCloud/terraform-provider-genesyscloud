package genesyscloud

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"terraform-provider-genesyscloud/genesyscloud/consistency_checker"

	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	chunksProcess "terraform-provider-genesyscloud/genesyscloud/util/chunks"
	lists "terraform-provider-genesyscloud/genesyscloud/util/lists"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/mypurecloud/platform-client-sdk-go/v115/platformclientv2"
	"github.com/nyaruka/phonenumbers"
)

var (
	contactTypeEmail = "EMAIL"

	phoneNumberResource = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"number": {
				Description:      "Phone number. Phone number must be in an E.164 number format.",
				Type:             schema.TypeString,
				Optional:         true,
				ValidateDiagFunc: ValidatePhoneNumber,
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

func getAllUsers(ctx context.Context, sdkConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	resources := make(resourceExporter.ResourceIDMetaMap)
	usersAPI := platformclientv2.NewUsersApiWithConfig(sdkConfig)

	// Newly created resources often aren't returned unless there's a delay
	time.Sleep(5 * time.Second)

	errorChan := make(chan error)
	wgDone := make(chan bool)
	usersChan := make(chan platformclientv2.User, 200)
	defer close(usersChan)

	// Cancel remaining goroutines if an error occurs
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup

	getUsersByStatus := func(userStatus string) {
		defer wg.Done()

		const pageSize = 100
		users, _, err := usersAPI.GetUsers(pageSize, 1, nil, nil, "", nil, "", userStatus)
		if err != nil {
			errorChan <- err
			cancel()
			return
		}
		for _, user := range *users.Entities {
			usersChan <- user
		}
		select {
		case <-ctx.Done():
			return
		default:
		}

		for pageNum := 2; pageNum <= *users.PageCount; pageNum++ {
			users, _, err := usersAPI.GetUsers(pageSize, pageNum, nil, nil, "", nil, "", userStatus)
			if err != nil {
				errorChan <- err
				cancel()
				return
			}

			for _, user := range *users.Entities {
				usersChan <- user
			}

			select {
			case <-ctx.Done():
				return
			default:
			}
		}
	}

	wg.Add(2)
	go getUsersByStatus("active")
	go getUsersByStatus("inactive")

	go func() {
		wg.Wait()

		// Make sure the buffer channel is emptied out
		for {
			if len(usersChan) == 0 {
				wgDone <- true
			}
		}
	}()

	go func() {
		for user := range usersChan {
			resources[*user.Id] = &resourceExporter.ResourceMeta{Name: *user.Email}
		}
	}()

	// Wait until either WaitGroup is done or an error is received
	select {
	case <-wgDone:
		return resources, nil
	case err := <-errorChan:
		return nil, diag.Errorf("Failed to get all users: %v", err)
	}
}

func UserExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: GetAllWithPooledClient(getAllUsers),
		RefAttrs: map[string]*resourceExporter.RefAttrSettings{
			"manager":                       {RefType: "genesyscloud_user"},
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

func ResourceUser() *schema.Resource {
	return &schema.Resource{
		Description: "Genesys Cloud User",

		CreateContext: CreateWithPooledClient(createUser),
		ReadContext:   ReadWithPooledClient(readUser),
		UpdateContext: UpdateWithPooledClient(updateUser),
		DeleteContext: DeleteWithPooledClient(deleteUser),
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
							ValidateDiagFunc: validateDate,
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
					},
				},
			},
		},
	}
}

func createUser(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	email := d.Get("email").(string)
	name := d.Get("name").(string)
	password := d.Get("password").(string)
	state := d.Get("state").(string)
	divisionID := d.Get("division_id").(string)
	department := d.Get("department").(string)
	title := d.Get("title").(string)
	manager := d.Get("manager").(string)
	acdAutoAnswer := d.Get("acd_auto_answer").(bool)

	sdkConfig := meta.(*ProviderMeta).ClientConfig
	usersAPI := platformclientv2.NewUsersApiWithConfig(sdkConfig)

	addresses, addrErr := buildSdkAddresses(d)
	if addrErr != nil {
		return addrErr
	}

	// Check for a deleted user before creating
	id, _ := getDeletedUserId(email, usersAPI)
	if id != nil {
		d.SetId(*id)
		return restoreDeletedUser(ctx, d, meta, usersAPI)
	}

	createUser := platformclientv2.Createuser{
		Email:      &email,
		Name:       &name,
		State:      &state,
		Addresses:  addresses,
		Department: &department,
		Title:      &title,
	}

	// Optional attributes that should not be empty strings
	if password != "" {
		createUser.Password = &password
	}
	if divisionID != "" {
		createUser.DivisionId = &divisionID
	}

	log.Printf("Creating user %s", email)

	user, resp, err := usersAPI.PostUsers(createUser)
	if err != nil {
		if resp != nil && resp.Error != nil && (*resp.Error).Code == "general.conflict" {
			// Check for a deleted user
			id, diagErr := getDeletedUserId(email, usersAPI)
			if diagErr != nil {
				return diagErr
			}
			if id != nil {
				d.SetId(*id)
				return restoreDeletedUser(ctx, d, meta, usersAPI)
			}
		}
		return diag.Errorf("Failed to create user %s: %s", email, err)
	}

	d.SetId(*user.Id)

	// Set attributes that can only be modified in a patch
	if d.HasChanges(
		"manager",
		"locations",
		"acd_auto_answer",
		"profile_skills",
		"certifications",
		"employer_info") {
		log.Printf("Updating additional attributes for user %s", email)
		_, _, patchErr := usersAPI.PatchUser(d.Id(), platformclientv2.Updateuser{
			Manager:        &manager,
			Locations:      buildSdkLocations(d),
			AcdAutoAnswer:  &acdAutoAnswer,
			Certifications: buildSdkCertifications(d),
			EmployerInfo:   buildSdkEmployerInfo(d),
			Version:        user.Version,
		})
		if patchErr != nil {
			return diag.Errorf("Failed to update user %s: %v", d.Id(), patchErr)
		}
	}

	diagErr := updateUserSkills(d, usersAPI)
	if diagErr != nil {
		return diagErr
	}

	diagErr = updateUserLanguages(d, usersAPI)
	if diagErr != nil {
		return diagErr
	}

	diagErr = updateUserProfileSkills(d, usersAPI)
	if diagErr != nil {
		return diagErr
	}

	diagErr = updateUserRoutingUtilization(d, usersAPI)
	if diagErr != nil {
		return diagErr
	}

	log.Printf("Created user %s %s", email, *user.Id)
	return readUser(ctx, d, meta)
}

func readUser(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*ProviderMeta).ClientConfig
	usersAPI := platformclientv2.NewUsersApiWithConfig(sdkConfig)

	log.Printf("Reading user %s", d.Id())
	return WithRetriesForRead(ctx, d, func() *retry.RetryError {
		currentUser, resp, getErr := usersAPI.GetUser(d.Id(), []string{
			// Expands
			"skills",
			"languages",
			"locations",
			"profileSkills",
			"certifications",
			"employerInfo",
		}, "", "")

		if getErr != nil {
			if IsStatus404(resp) {
				return retry.RetryableError(fmt.Errorf("Failed to read user %s: %s", d.Id(), getErr))
			}
			return retry.NonRetryableError(fmt.Errorf("Failed to read user %s: %s", d.Id(), getErr))
		}

		cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceUser())

		// Required attributes
		d.Set("name", *currentUser.Name)
		d.Set("email", *currentUser.Email)
		d.Set("division_id", *currentUser.Division.Id)
		d.Set("state", *currentUser.State)

		if currentUser.Department != nil {
			d.Set("department", *currentUser.Department)
		} else {
			d.Set("department", nil)
		}

		if currentUser.Title != nil {
			d.Set("title", *currentUser.Title)
		} else {
			d.Set("title", nil)
		}

		if currentUser.Manager != nil {
			d.Set("manager", *(*currentUser.Manager).Id)
		} else {
			d.Set("manager", nil)
		}

		if currentUser.AcdAutoAnswer != nil {
			d.Set("acd_auto_answer", *currentUser.AcdAutoAnswer)
		} else {
			d.Set("acd_auto_answer", nil)
		}

		d.Set("addresses", flattenUserAddresses(d, currentUser.Addresses))
		d.Set("routing_skills", flattenUserSkills(currentUser.Skills))
		d.Set("routing_languages", flattenUserLanguages(currentUser.Languages))
		d.Set("locations", flattenUserLocations(currentUser.Locations))
		d.Set("profile_skills", flattenUserProfileSkills(currentUser.ProfileSkills))
		d.Set("certifications", flattenUserCertifications(currentUser.Certifications))
		d.Set("employer_info", flattenUserEmployerInfo(currentUser.EmployerInfo))

		if diagErr := readUserRoutingUtilization(d, usersAPI); diagErr != nil {
			return retry.NonRetryableError(fmt.Errorf("%v", diagErr))
		}

		log.Printf("Read user %s %s", d.Id(), *currentUser.Email)
		return cc.CheckState()
	})
}

func updateUser(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)
	email := d.Get("email").(string)
	state := d.Get("state").(string)
	department := d.Get("department").(string)
	title := d.Get("title").(string)
	manager := d.Get("manager").(string)
	acdAutoAnswer := d.Get("acd_auto_answer").(bool)

	sdkConfig := meta.(*ProviderMeta).ClientConfig
	usersAPI := platformclientv2.NewUsersApiWithConfig(sdkConfig)

	addresses, err := buildSdkAddresses(d)
	if err != nil {
		return err
	}

	log.Printf("Updating user %s", email)

	// If state changes, it is the only modifiable field, so it must be updated separately
	if d.HasChange("state") {
		log.Printf("Updating state for user %s", email)
		patchErr := patchUser(d.Id(), platformclientv2.Updateuser{
			State: &state,
		}, usersAPI)
		if patchErr != nil {
			return patchErr
		}
	}

	patchErr := patchUser(d.Id(), platformclientv2.Updateuser{
		Name:           &name,
		Email:          &email,
		Department:     &department,
		Title:          &title,
		Manager:        &manager,
		Addresses:      addresses,
		Locations:      buildSdkLocations(d),
		AcdAutoAnswer:  &acdAutoAnswer,
		Certifications: buildSdkCertifications(d),
		EmployerInfo:   buildSdkEmployerInfo(d),
	}, usersAPI)
	if patchErr != nil {
		return patchErr
	}

	diagErr := updateObjectDivision(d, "USER", sdkConfig)
	if diagErr != nil {
		return diagErr
	}

	diagErr = updateUserSkills(d, usersAPI)
	if diagErr != nil {
		return diagErr
	}

	diagErr = updateUserLanguages(d, usersAPI)
	if diagErr != nil {
		return diagErr
	}

	diagErr = updateUserProfileSkills(d, usersAPI)
	if diagErr != nil {
		return diagErr
	}

	diagErr = updateUserRoutingUtilization(d, usersAPI)
	if diagErr != nil {
		return diagErr
	}

	log.Printf("Finished updating user %s", email)
	return readUser(ctx, d, meta)
}

func deleteUser(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	email := d.Get("email").(string)

	sdkConfig := meta.(*ProviderMeta).ClientConfig
	usersAPI := platformclientv2.NewUsersApiWithConfig(sdkConfig)

	log.Printf("Deleting user %s", email)
	err := RetryWhen(IsVersionMismatch, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
		// Directory occasionally returns version errors on deletes if an object was updated at the same time.
		_, resp, err := usersAPI.DeleteUser(d.Id())
		if err != nil {
			time.Sleep(5 * time.Second)
			return resp, diag.Errorf("Failed to delete user %s: %s", email, err)
		}
		log.Printf("Deleted user %s", email)
		return nil, nil
	})
	if err != nil {
		return err
	}

	// Verify user in deleted state and search index has been updated
	return WithRetries(ctx, 180*time.Second, func() *retry.RetryError {
		id, err := getDeletedUserId(email, usersAPI)
		if err != nil {
			return retry.NonRetryableError(fmt.Errorf("Error searching for deleted user %s: %v", email, err))
		}
		if id == nil {
			return retry.RetryableError(fmt.Errorf("User %s not yet in deleted state", email))
		}
		return nil
	})
}

func patchUser(id string, update platformclientv2.Updateuser, usersAPI *platformclientv2.UsersApi) diag.Diagnostics {
	return patchUserWithState(id, "", update, usersAPI)
}

func patchUserWithState(id string, state string, update platformclientv2.Updateuser, usersAPI *platformclientv2.UsersApi) diag.Diagnostics {
	return RetryWhen(IsVersionMismatch, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
		currentUser, _, getErr := usersAPI.GetUser(id, nil, "", state)
		if getErr != nil {
			return nil, diag.Errorf("Failed to read user %s: %s", id, getErr)
		}

		update.Version = currentUser.Version
		_, resp, patchErr := usersAPI.PatchUser(id, update)
		if patchErr != nil {
			return resp, diag.Errorf("Failed to update user %s: %v", id, patchErr)
		}
		return nil, nil
	})
}

func getDeletedUserId(email string, usersAPI *platformclientv2.UsersApi) (*string, diag.Diagnostics) {
	exactType := "EXACT"
	results, _, getErr := usersAPI.PostUsersSearch(platformclientv2.Usersearchrequest{
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
		return nil, diag.Errorf("Failed to search for user %s: %s", email, getErr)
	}
	if results.Results != nil && len(*results.Results) > 0 {
		// User found
		return (*results.Results)[0].Id, nil
	}
	return nil, nil
}

func restoreDeletedUser(ctx context.Context, d *schema.ResourceData, meta interface{}, usersAPI *platformclientv2.UsersApi) diag.Diagnostics {
	email := d.Get("email").(string)
	state := d.Get("state").(string)

	log.Printf("Restoring deleted user %s", email)
	patchErr := patchUserWithState(d.Id(), "deleted", platformclientv2.Updateuser{
		State: &state,
	}, usersAPI)
	if patchErr != nil {
		return patchErr
	}
	return updateUser(ctx, d, meta)
}

func phoneNumberHash(val interface{}) int {
	// Copy map to avoid modifying state
	phoneMap := make(map[string]interface{})
	for k, v := range val.(map[string]interface{}) {
		phoneMap[k] = v
	}
	if num, ok := phoneMap["number"]; ok {
		// Attempt to format phone numbers before hashing
		number, err := phonenumbers.Parse(num.(string), "US")
		if err == nil {
			phoneMap["number"] = phonenumbers.Format(number, phonenumbers.E164)
		}
	}
	return schema.HashResource(phoneNumberResource)(phoneMap)
}

func buildSdkEmails(configEmails *schema.Set) []platformclientv2.Contact {
	emailSlice := configEmails.List()
	sdkContacts := make([]platformclientv2.Contact, len(emailSlice))
	for i, configEmail := range emailSlice {
		emailMap := configEmail.(map[string]interface{})
		emailAddress, _ := emailMap["address"].(string)
		emailType, _ := emailMap["type"].(string)

		sdkContacts[i] = platformclientv2.Contact{
			Address:   &emailAddress,
			MediaType: &contactTypeEmail,
			VarType:   &emailType,
		}
	}
	return sdkContacts
}

func buildSdkPhoneNumbers(configPhoneNumbers *schema.Set) ([]platformclientv2.Contact, diag.Diagnostics) {
	phoneNumberSlice := configPhoneNumbers.List()
	sdkContacts := make([]platformclientv2.Contact, len(phoneNumberSlice))
	for i, configPhone := range phoneNumberSlice {
		phoneMap := configPhone.(map[string]interface{})
		phoneMediaType := phoneMap["media_type"].(string)
		phoneType := phoneMap["type"].(string)

		contact := platformclientv2.Contact{
			MediaType: &phoneMediaType,
			VarType:   &phoneType,
		}

		if phoneNum, ok := phoneMap["number"].(string); ok && phoneNum != "" {
			contact.Address = &phoneNum
		}
		if phoneExt, ok := phoneMap["extension"].(string); ok && phoneExt != "" {
			contact.Extension = &phoneExt
		}

		sdkContacts[i] = contact
	}
	return sdkContacts, nil
}

func buildSdkAddresses(d *schema.ResourceData) (*[]platformclientv2.Contact, diag.Diagnostics) {
	if addresses := d.Get("addresses").([]interface{}); addresses != nil {
		sdkAddresses := make([]platformclientv2.Contact, 0)
		var otherEmails *schema.Set
		var phoneNumbers *schema.Set
		if len(addresses) > 0 {
			if addressMap, ok := addresses[0].(map[string]interface{}); ok {
				otherEmails = addressMap["other_emails"].(*schema.Set)
				phoneNumbers = addressMap["phone_numbers"].(*schema.Set)
			} else {
				return nil, nil
			}
		}

		if otherEmails != nil {
			sdkAddresses = append(sdkAddresses, buildSdkEmails(otherEmails)...)
		}
		if phoneNumbers != nil {
			sdkNums, err := buildSdkPhoneNumbers(phoneNumbers)
			if err != nil {
				return nil, err
			}
			sdkAddresses = append(sdkAddresses, sdkNums...)
		}
		return &sdkAddresses, nil
	}
	return nil, nil
}

func buildSdkLocations(d *schema.ResourceData) *[]platformclientv2.Location {
	if locationConfig := d.Get("locations"); locationConfig != nil {
		sdkLocations := make([]platformclientv2.Location, 0)
		locationList := locationConfig.(*schema.Set).List()
		for _, configLoc := range locationList {
			locMap := configLoc.(map[string]interface{})
			locID := locMap["location_id"].(string)
			locNotes := locMap["notes"].(string)

			sdkLocations = append(sdkLocations, platformclientv2.Location{
				Id:    &locID,
				Notes: &locNotes,
			})
		}
		return &sdkLocations
	}
	return nil
}

func buildSdkEmployerInfo(d *schema.ResourceData) *platformclientv2.Employerinfo {
	if configInfo := d.Get("employer_info").([]interface{}); configInfo != nil {
		var sdkInfo platformclientv2.Employerinfo
		if len(configInfo) > 0 {
			if _, ok := configInfo[0].(map[string]interface{}); !ok {
				return nil
			}
			infoMap := configInfo[0].(map[string]interface{})
			// Only set non-empty values.
			if offName := infoMap["official_name"].(string); len(offName) > 0 {
				sdkInfo.OfficialName = &offName
			}
			if empID := infoMap["employee_id"].(string); len(empID) > 0 {
				sdkInfo.EmployeeId = &empID
			}
			if empType := infoMap["employee_type"].(string); len(empType) > 0 {
				sdkInfo.EmployeeType = &empType
			}
			if dateHire := infoMap["date_hire"].(string); len(dateHire) > 0 {
				sdkInfo.DateHire = &dateHire
			}
		}
		return &sdkInfo
	}
	return nil
}

func buildSdkCertifications(d *schema.ResourceData) *[]string {
	if certs := d.Get("certifications"); certs != nil {
		return lists.SetToStringList(certs.(*schema.Set))
	}
	return nil
}

func getNumbers(d *schema.ResourceData, index int) (bool, bool) {
	isNumber := false
	isExtension := false

	if addresses1 := d.Get("addresses").([]interface{}); addresses1 != nil {
		var phoneNumbers *schema.Set
		if len(addresses1) > 0 {
			addressMap := addresses1[0].(map[string]interface{})
			phoneNumbers = addressMap["phone_numbers"].(*schema.Set)
		}

		if phoneNumbers != nil {
			phoneNumberSlice := phoneNumbers.List()
			for ii, configPhone := range phoneNumberSlice {
				if ii != index {
					continue
				}
				phoneMap := configPhone.(map[string]interface{})
				if phoneNum, ok := phoneMap["number"].(string); ok && phoneNum != "" {
					isNumber = true
				}
				if phoneExt, ok := phoneMap["extension"].(string); ok && phoneExt != "" {
					isExtension = true
				}
				break
			}
		}
	}
	return isNumber, isExtension
}

func flattenUserAddresses(d *schema.ResourceData, addresses *[]platformclientv2.Contact) []interface{} {
	if addresses == nil || len(*addresses) == 0 {
		return nil
	}

	emailSet := schema.NewSet(schema.HashResource(otherEmailResource), []interface{}{})
	phoneNumSet := schema.NewSet(phoneNumberHash, []interface{}{})

	for i, address := range *addresses {
		if address.MediaType != nil {
			if *address.MediaType == "SMS" || *address.MediaType == "PHONE" {
				phoneNumber := make(map[string]interface{})
				phoneNumber["media_type"] = *address.MediaType

				// Strip off any parentheses from phone numbers
				if address.Address != nil {
					phoneNumber["number"] = strings.Trim(*address.Address, "()")
				} else if address.Display != nil {
					// Some numbers are only returned in Display
					isNumber, isExtension := getNumbers(d, i)

					if isNumber && phoneNumber["number"] != "" {
						phoneNumber["number"] = strings.Trim(*address.Display, "()")
					}
					if isExtension {
						phoneNumber["extension"] = strings.Trim(*address.Display, "()")
					}

					if !isNumber && !isExtension {
						if address.Extension == nil {
							phoneNumber["extension"] = strings.Trim(*address.Display, "()")
						} else if phoneNumber["number"] != "" {
							phoneNumber["number"] = strings.Trim(*address.Display, "()")
						}
					}
				}

				if address.Extension != nil {
					phoneNumber["extension"] = *address.Extension
				}

				if address.VarType != nil {
					phoneNumber["type"] = *address.VarType
				}
				phoneNumSet.Add(phoneNumber)
			} else if *address.MediaType == "EMAIL" {
				email := make(map[string]interface{})
				email["type"] = *address.VarType
				email["address"] = *address.Address
				emailSet.Add(email)
			} else {
				log.Printf("Unknown address media type %s", *address.MediaType)
			}
		}
	}
	return []interface{}{map[string]interface{}{
		"other_emails":  emailSet,
		"phone_numbers": phoneNumSet,
	}}
}

func flattenUserEmployerInfo(empInfo *platformclientv2.Employerinfo) []interface{} {
	if empInfo == nil {
		return nil
	}
	var (
		offName  string
		empID    string
		empType  string
		dateHire string
	)

	if empInfo.OfficialName != nil {
		offName = *empInfo.OfficialName
	}
	if empInfo.EmployeeId != nil {
		empID = *empInfo.EmployeeId
	}
	if empInfo.EmployeeType != nil {
		empType = *empInfo.EmployeeType
	}
	if empInfo.DateHire != nil {
		dateHire = *empInfo.DateHire
	}

	return []interface{}{map[string]interface{}{
		"official_name": offName,
		"employee_id":   empID,
		"employee_type": empType,
		"date_hire":     dateHire,
	}}
}

func readUserRoutingUtilization(d *schema.ResourceData, usersAPI *platformclientv2.UsersApi) diag.Diagnostics {
	settings, resp, getErr := usersAPI.GetRoutingUserUtilization(d.Id())
	if getErr != nil {
		if IsStatus404(resp) {
			d.SetId("") // User doesn't exist
			return nil
		}
		return diag.Errorf("Failed to read Routing Utilization for user %s: %s", d.Id(), getErr)
	}

	if settings != nil && settings.Utilization != nil {
		// If the settings are org-wide, set to empty to indicate no settings on the user
		if settings.Level != nil && *settings.Level == "Organization" {
			d.Set("routing_utilization", []interface{}{})
		} else {
			allSettings := map[string]interface{}{}
			for sdkType, schemaType := range utilizationMediaTypes {
				if mediaSettings, ok := (*settings.Utilization)[sdkType]; ok {
					allSettings[schemaType] = flattenUtilizationSetting(mediaSettings)
				}
			}
			d.Set("routing_utilization", []interface{}{allSettings})
		}
	} else {
		d.Set("routing_utilization", nil)
	}

	return nil
}

func updateUserSkills(d *schema.ResourceData, usersAPI *platformclientv2.UsersApi) diag.Diagnostics {

	transformFunc := func(configSkill interface{}) platformclientv2.Userroutingskillpost {
		skillMap := configSkill.(map[string]interface{})
		skillID := skillMap["skill_id"].(string)
		skillProf := skillMap["proficiency"].(float64)

		return platformclientv2.Userroutingskillpost{
			Id:          &skillID,
			Proficiency: &skillProf,
		}
	}

	chunkProcessor := func(chunk []platformclientv2.Userroutingskillpost) diag.Diagnostics {
		diagErr := RetryWhen(IsVersionMismatch, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
			_, resp, err := usersAPI.PatchUserRoutingskillsBulk(d.Id(), chunk)
			if err != nil {
				return resp, diag.Errorf("Failed to update skills for user %s: %s", d.Id(), err)
			}
			return nil, nil
		})
		if diagErr != nil {
			return diagErr
		}
		return nil
	}

	if d.HasChange("routing_skills") {
		if skillsConfig := d.Get("routing_skills"); skillsConfig != nil {
			skillsList := skillsConfig.(*schema.Set).List()
			chunks := chunksProcess.ChunkItems(skillsList, transformFunc, 50)
			return chunksProcess.ProcessChunks(chunks, chunkProcessor)
		}
	}
	return nil
}

func updateUserLanguages(d *schema.ResourceData, usersAPI *platformclientv2.UsersApi) diag.Diagnostics {
	if d.HasChange("routing_languages") {
		if languages := d.Get("routing_languages"); languages != nil {
			log.Printf("Updating languages for user %s", d.Get("email"))
			newLangProfs := make(map[string]int)
			langList := languages.(*schema.Set).List()
			newLangIds := make([]string, len(langList))
			for i, lang := range langList {
				langMap := lang.(map[string]interface{})
				newLangIds[i] = langMap["language_id"].(string)
				newLangProfs[newLangIds[i]] = langMap["proficiency"].(int)
			}

			oldSdkLangs, err := getUserRoutingLanguages(d.Id(), usersAPI)
			if err != nil {
				return err
			}

			oldLangIds := make([]string, len(oldSdkLangs))
			oldLangProfs := make(map[string]int)
			for i, lang := range oldSdkLangs {
				oldLangIds[i] = *lang.Id
				oldLangProfs[oldLangIds[i]] = int(*lang.Proficiency)
			}

			if len(oldLangIds) > 0 {
				langsToRemove := lists.SliceDifference(oldLangIds, newLangIds)
				for _, langID := range langsToRemove {
					diagErr := RetryWhen(IsVersionMismatch, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
						resp, err := usersAPI.DeleteUserRoutinglanguage(d.Id(), langID)
						if err != nil {
							return resp, diag.Errorf("Failed to remove language from user %s: %s", d.Id(), err)
						}
						return nil, nil
					})
					if diagErr != nil {
						return diagErr
					}
				}
			}

			if len(newLangIds) > 0 {
				// Languages to add
				langsToAddOrUpdate := lists.SliceDifference(newLangIds, oldLangIds)

				// Check for existing proficiencies to update which can be done with the same API
				for langID, newNum := range newLangProfs {
					if oldNum, found := oldLangProfs[langID]; found {
						if newNum != oldNum {
							langsToAddOrUpdate = append(langsToAddOrUpdate, langID)
						}
					}
				}
				if diagErr := updateUserRoutingLanguages(d.Id(), langsToAddOrUpdate, newLangProfs, usersAPI); diagErr != nil {
					return diagErr
				}
			}
			log.Printf("Languages updated for user %s", d.Get("email"))
		}
	}
	return nil
}

func getUserRoutingLanguages(userID string, api *platformclientv2.UsersApi) ([]platformclientv2.Userroutinglanguage, diag.Diagnostics) {
	const maxPageSize = 50

	var sdkLanguages []platformclientv2.Userroutinglanguage
	for pageNum := 1; ; pageNum++ {
		langs, _, err := api.GetUserRoutinglanguages(userID, maxPageSize, pageNum, "")
		if err != nil {
			return nil, diag.Errorf("Failed to query languages for user %s: %s", userID, err)
		}
		if langs == nil || langs.Entities == nil || len(*langs.Entities) == 0 {
			return sdkLanguages, nil
		}
		for _, language := range *langs.Entities {
			sdkLanguages = append(sdkLanguages, language)
		}
	}
}

func updateUserRoutingLanguages(
	userID string,
	langsToUpdate []string,
	langProfs map[string]int,
	api *platformclientv2.UsersApi) diag.Diagnostics {
	// Bulk API restricts language adds to 50 per call
	const maxBatchSize = 50

	chunkBuild := func(val string) platformclientv2.Userroutinglanguagepost {
		newProf := float64(langProfs[val])
		return platformclientv2.Userroutinglanguagepost{
			Id:          &val,
			Proficiency: &newProf,
		}
	}

	// Generic call to prepare chunks for the Update. Takes in three args
	// 1. langsToUpdate 2. The Entity prepare func for the update 3. Chunk Size
	chunks := chunksProcess.ChunkItems(langsToUpdate, chunkBuild, maxBatchSize)
	// Closure to process the chunks

	chunkProcessor := func(chunk []platformclientv2.Userroutinglanguagepost) diag.Diagnostics {
		diagErr := RetryWhen(IsVersionMismatch, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
			_, resp, err := api.PatchUserRoutinglanguagesBulk(userID, chunk)
			if err != nil {
				return resp, diag.Errorf("Failed to update languages for user %s: %s", userID, err)
			}
			return nil, nil
		})
		if diagErr != nil {
			return diagErr
		}
		return nil
	}

	// Genric Function call which takes in the chunks and the processing function
	return chunksProcess.ProcessChunks(chunks, chunkProcessor)
}

func updateUserProfileSkills(d *schema.ResourceData, usersAPI *platformclientv2.UsersApi) diag.Diagnostics {
	if d.HasChange("profile_skills") {
		if profileSkills := d.Get("profile_skills"); profileSkills != nil {
			profileSkills := lists.SetToStringList(profileSkills.(*schema.Set))
			diagErr := RetryWhen(IsVersionMismatch, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
				_, resp, err := usersAPI.PutUserProfileskills(d.Id(), *profileSkills)
				if err != nil {
					return resp, diag.Errorf("Failed to update profile skills for user %s: %s", d.Id(), err)
				}
				return nil, nil
			})
			if diagErr != nil {
				return diagErr
			}
		}
	}
	return nil
}

func updateUserRoutingUtilization(d *schema.ResourceData, usersAPI *platformclientv2.UsersApi) diag.Diagnostics {
	if d.HasChange("routing_utilization") {
		if utilConfig := d.Get("routing_utilization").([]interface{}); utilConfig != nil {
			if len(utilConfig) > 0 { // Specified but empty utilization list will reset to org-wide defaults
				sdkSettings := make(map[string]platformclientv2.Mediautilization)
				allSettings := utilConfig[0].(map[string]interface{})
				for sdkType, schemaType := range utilizationMediaTypes {
					if mediaSettings, ok := allSettings[schemaType]; ok && len(mediaSettings.([]interface{})) > 0 {
						sdkSettings[sdkType] = buildSdkMediaUtilization(mediaSettings.([]interface{}))
					}
				}
				// Update settings
				_, _, err := usersAPI.PutRoutingUserUtilization(d.Id(), platformclientv2.Utilization{
					Utilization: &sdkSettings,
				})
				if err != nil {
					return diag.Errorf("Failed to update Routing Utilization for user %s: %s", d.Id(), err)
				}
			} else {
				// Reset to org-wide defaults
				_, err := usersAPI.DeleteRoutingUserUtilization(d.Id())
				if err != nil {
					return diag.Errorf("Failed to delete Routing Utilization for user %s: %s", d.Id(), err)
				}
			}
		}
	}
	return nil
}

func flattenUserSkills(skills *[]platformclientv2.Userroutingskill) *schema.Set {
	if skills == nil {
		return nil
	}
	skillSet := schema.NewSet(schema.HashResource(userSkillResource), []interface{}{})
	for _, sdkSkill := range *skills {
		skill := make(map[string]interface{})
		skill["skill_id"] = *sdkSkill.Id
		skill["proficiency"] = *sdkSkill.Proficiency
		skillSet.Add(skill)
	}
	return skillSet
}

func flattenUserLanguages(languages *[]platformclientv2.Userroutinglanguage) *schema.Set {
	if languages == nil {
		return nil
	}
	languageSet := schema.NewSet(schema.HashResource(userLanguageResource), []interface{}{})
	for _, sdkLang := range *languages {
		language := make(map[string]interface{})
		language["language_id"] = *sdkLang.Id
		language["proficiency"] = int(*sdkLang.Proficiency)
		languageSet.Add(language)
	}
	return languageSet
}

func flattenUserLocations(locations *[]platformclientv2.Location) *schema.Set {
	if locations == nil {
		return nil
	}
	locSet := schema.NewSet(schema.HashResource(userLocationResource), []interface{}{})
	for _, sdkLoc := range *locations {
		if sdkLoc.LocationDefinition != nil {
			location := make(map[string]interface{})
			location["location_id"] = *sdkLoc.LocationDefinition.Id
			if sdkLoc.Notes != nil {
				location["notes"] = *sdkLoc.Notes
			}
			locSet.Add(location)
		}
	}
	return locSet
}

func flattenUserProfileSkills(skills *[]string) *schema.Set {
	if skills != nil {
		return lists.StringListToSet(*skills)
	}
	return nil
}

func flattenUserCertifications(certs *[]string) *schema.Set {
	if certs != nil {
		return lists.StringListToSet(*certs)
	}
	return nil
}

// Basic user with minimum required fields
func GenerateBasicUserResource(resourceID string, email string, name string) string {
	return GenerateUserResource(resourceID, email, name, nullValue, nullValue, nullValue, nullValue, nullValue, "", "")
}

func GenerateUserResource(
	resourceID string,
	email string,
	name string,
	state string,
	title string,
	department string,
	manager string,
	acdAutoAnswer string,
	profileSkills string,
	certifications string) string {
	return fmt.Sprintf(`resource "genesyscloud_user" "%s" {
		email = "%s"
		name = "%s"
		state = %s
		title = %s
		department = %s
		manager = %s
		acd_auto_answer = %s
		profile_skills = [%s]
		certifications = [%s]
	}
	`, resourceID, email, name, state, title, department, manager, acdAutoAnswer, profileSkills, certifications)
}

func GenerateUserWithCustomAttrs(resourceID string, email string, name string, attrs ...string) string {
	return fmt.Sprintf(`resource "genesyscloud_user" "%s" {
		email = "%s"
		name = "%s"
		%s
	}
	`, resourceID, email, name, strings.Join(attrs, "\n"))
}
