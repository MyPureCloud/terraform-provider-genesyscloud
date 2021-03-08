package genesyscloud

import (
	"context"
	"log"
	"strings"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/mypurecloud/platform-client-sdk-go/platformclientv2"
	"github.com/nyaruka/phonenumbers"
)

var (
	contactTypeEmail = "EMAIL"

	phoneNumberResource = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"number": {
				Description:      "Phone number. Defaults to US country code.",
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: validatePhoneNumber,
			},
			"media_type": {
				Description:  "Media type of phone number (SMS | PHONE).",
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "PHONE",
				ValidateFunc: validation.StringInSlice([]string{"PHONE", "SMS"}, false),
			},
			"type": {
				Description:  "Type of number (WORK | WORK2 | WORK3 | WORK4 | HOME | MOBILE).",
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "WORK",
				ValidateFunc: validation.StringInSlice([]string{"WORK", "WORK2", "WORK3", "WORK4", "HOME", "MOBILE"}, false),
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
	userRoleResource = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"role_id": {
				Description: "Role ID.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"division_ids": {
				Description: "Divisions applied to this role. If not set, the home division will be used. '*' may be set for all divisions.",
				Type:        schema.TypeSet,
				Optional:    true,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
		},
	}
)

func getAllUsers(ctx context.Context, sdkConfig *platformclientv2.Configuration) (ResourceIDNameMap, diag.Diagnostics) {
	resources := make(map[string]string)
	usersAPI := platformclientv2.NewUsersApiWithConfig(sdkConfig)

	for pageNum := 1; ; pageNum++ {
		users, _, getErr := usersAPI.GetUsers(100, pageNum, nil, nil, "", nil, "", "")
		if getErr != nil {
			return nil, diag.Errorf("Failed to get page of users: %v", getErr)
		}

		if users.Entities == nil || len(*users.Entities) == 0 {
			break
		}

		for _, user := range *users.Entities {
			resources[*user.Id] = *user.Email
		}
	}

	return resources, nil
}

func userExporter() *ResourceExporter {
	return &ResourceExporter{
		GetResourcesFunc: getAllWithPooledClient(getAllUsers),
		ResourceDef:      resourceUser(),
		RefAttrs: map[string]*RefAttrSettings{
			"manager":                 {RefType: "genesyscloud_user"},
			"division_id":             {},
			"routing_skills.skill_id": {RefType: "genesyscloud_routing_skill", RemoveOuterItem: true},
			"roles.role_id":           {RefType: "genesyscloud_auth_role", RemoveOuterItem: true},
			"roles.division_ids":      {AltValues: []string{"*"}},
		},
		AllowZeroValues: []string{"routing_skills.proficiency"},
	}
}

func resourceUser() *schema.Resource {
	return &schema.Resource{
		Description: "Genesys Cloud User",

		CreateContext: createWithPooledClient(createUser),
		ReadContext:   readWithPooledClient(readUser),
		UpdateContext: updateWithPooledClient(updateUser),
		DeleteContext: deleteWithPooledClient(deleteUser),
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
			"addresses": {
				Description: "The address settings for this user. If not set, addresses will not be managed by this resource.",
				Type:        schema.TypeList,
				MaxItems:    1,
				Computed:    true,
				Optional:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"other_emails": {
							Description: "Other Email addresses for this user.",
							Type:        schema.TypeSet,
							Optional:    true,
							Elem:        otherEmailResource,
						},
						"phone_numbers": {
							Description: "Phone number addresses for this user.",
							Type:        schema.TypeSet,
							Optional:    true,
							Set:         phoneNumberHash,
							Elem:        phoneNumberResource,
						},
					},
				},
			},
			"routing_skills": {
				Description: "Skills and proficiencies for this user. If not set, skills will not be managed by this resource.",
				Type:        schema.TypeSet,
				Optional:    true,
				Computed:    true,
				Elem:        userSkillResource,
			},
			"roles": {
				Description: "Roles and their divisions assigned to this user. If not set, roles will not be managed by this resource.",
				Type:        schema.TypeSet,
				Optional:    true,
				Computed:    true,
				Elem:        userRoleResource,
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

	sdkConfig := meta.(*providerMeta).ClientConfig
	usersAPI := platformclientv2.NewUsersApiWithConfig(sdkConfig)
	authAPI := platformclientv2.NewAuthorizationApiWithConfig(sdkConfig)

	addresses, addrErr := buildSdkAddresses(d)
	if addrErr != nil {
		return addrErr
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
	user, _, err := usersAPI.PostUsers(createUser)
	if err != nil {
		// TODO: Handle restoring previously deleted users with the same email
		return diag.Errorf("Failed to create user %s: %s", email, err)
	}

	d.SetId(*user.Id)

	// Set attributes that can only be modified in a patch
	if d.HasChanges("manager") {
		log.Printf("Updating additional attributes for user %s", email)
		_, _, patchErr := usersAPI.PatchUser(d.Id(), platformclientv2.Updateuser{
			Manager: &manager,
			Version: user.Version,
		})
		if patchErr != nil {
			return diag.Errorf("Failed to update state for user %s: %s", email, patchErr)
		}
	}

	diagErr := updateUserSkills(d, usersAPI)
	if diagErr != nil {
		return diagErr
	}

	diagErr = updateUserRoles(d, authAPI)
	if diagErr != nil {
		return diagErr
	}
	log.Printf("Created user %s %s", email, *user.Id)
	return readUser(ctx, d, meta)
}

func readUser(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*providerMeta).ClientConfig
	usersAPI := platformclientv2.NewUsersApiWithConfig(sdkConfig)
	authAPI := platformclientv2.NewAuthorizationApiWithConfig(sdkConfig)

	log.Printf("Reading user %s", d.Id())
	currentUser, resp, getErr := usersAPI.GetUser(d.Id(), []string{"skills"}, "", "")
	if getErr != nil {
		if resp != nil && resp.StatusCode == 404 {
			d.SetId("")
			return nil
		}
		return diag.Errorf("Failed to read user %s: %s", d.Id(), getErr)
	}

	// Required attributes
	d.Set("name", *currentUser.Name)
	d.Set("email", *currentUser.Email)
	d.Set("division_id", *currentUser.Division.Id)
	d.Set("state", *currentUser.State)

	if currentUser.Department != nil {
		d.Set("department", *currentUser.Department)
	}

	if currentUser.Title != nil {
		d.Set("title", *currentUser.Title)
	}

	if currentUser.Manager != nil {
		d.Set("manager", *(*currentUser.Manager).Id)
	}

	if currentUser.Addresses != nil {
		d.Set("addresses", flattenUserAddresses(*currentUser.Addresses))
	}

	if currentUser.Skills != nil {
		d.Set("routing_skills", flattenUserSkills(*currentUser.Skills))
	}

	roles, err := readUserRoles(d.Id(), authAPI)
	if err != nil {
		return err
	}
	d.Set("roles", roles)

	log.Printf("Read user %s %s", d.Id(), *currentUser.Email)
	return nil
}

func updateUser(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)
	email := d.Get("email").(string)
	state := d.Get("state").(string)
	divisionID := d.Get("division_id").(string)
	department := d.Get("department").(string)
	title := d.Get("title").(string)
	manager := d.Get("manager").(string)

	sdkConfig := meta.(*providerMeta).ClientConfig
	usersAPI := platformclientv2.NewUsersApiWithConfig(sdkConfig)
	authAPI := platformclientv2.NewAuthorizationApiWithConfig(sdkConfig)

	log.Printf("Updating user %s", email)

	// Need the current user version for patches
	currentUser, _, getErr := usersAPI.GetUser(d.Id(), nil, "", "")
	if getErr != nil {
		return diag.Errorf("Failed to read user %s: %s", d.Id(), getErr)
	}

	addresses, err := buildSdkAddresses(d)
	if err != nil {
		return err
	}

	updateUser := platformclientv2.Updateuser{
		Name:       &name,
		Email:      &email,
		Department: &department,
		Title:      &title,
		Manager:    &manager,
		Version:    currentUser.Version,
		Addresses:  addresses,
	}

	// If state changes, it is the only modifiable field, so it must be updated separately
	if d.HasChange("state") {
		log.Printf("Updating state for user %s", email)
		updatedStateUser, _, patchErr := usersAPI.PatchUser(d.Id(), platformclientv2.Updateuser{
			State:   &state,
			Version: currentUser.Version,
		})
		if patchErr != nil {
			return diag.Errorf("Failed to update state for user %s: %s", email, patchErr)
		}
		// Update version for next patch
		updateUser.Version = updatedStateUser.Version
	}

	_, _, patchErr := usersAPI.PatchUser(d.Id(), updateUser)
	if patchErr != nil {
		return diag.Errorf("Failed to update user %s: %s", email, patchErr)
	}

	if d.HasChange("division_id") {
		// Default to home division
		if divisionID == "" {
			homeDivision, diagErr := getHomeDivisionID()
			if diagErr != nil {
				return diagErr
			}
			divisionID = homeDivision
		}

		log.Printf("Updating division for user %s to %s", email, divisionID)
		_, divErr := authAPI.PostAuthorizationDivisionObject(divisionID, "USER", []string{d.Id()})
		if divErr != nil {
			return diag.Errorf("Failed to update division for user %s: %s", email, divErr)
		}
	}

	diagErr := updateUserSkills(d, usersAPI)
	if diagErr != nil {
		return diagErr
	}

	diagErr = updateUserRoles(d, authAPI)
	if diagErr != nil {
		return diagErr
	}

	log.Printf("Finished updating user %s", email)
	return readUser(ctx, d, meta)
}

func deleteUser(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	email := d.Get("email").(string)

	sdkConfig := meta.(*providerMeta).ClientConfig
	usersAPI := platformclientv2.NewUsersApiWithConfig(sdkConfig)

	log.Printf("Deleting user %s", email)
	_, _, err := usersAPI.DeleteUser(d.Id())
	if err != nil {
		return diag.Errorf("Failed to delete user %s: %s", email, err)
	}
	log.Printf("Deleted user %s", email)
	return nil
}

func validatePhoneNumber(number interface{}, _ cty.Path) diag.Diagnostics {
	numberStr := number.(string)
	_, err := phonenumbers.Parse(numberStr, "US")
	if err != nil {
		return diag.Errorf("Failed to validate phone number %s: %s", numberStr, err)
	}
	return nil
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

		if phoneNum, ok := phoneMap["number"].(string); ok {
			contact.Address = &phoneNum
		}
		if phoneExt, ok := phoneMap["extension"].(string); ok {
			contact.Extension = &phoneExt
		}

		sdkContacts[i] = contact
	}
	return sdkContacts, nil
}

func buildSdkAddresses(d *schema.ResourceData) (*[]platformclientv2.Contact, diag.Diagnostics) {
	addresses := d.Get("addresses").([]interface{})
	if addresses != nil {
		var otherEmails *schema.Set
		var phoneNumbers *schema.Set
		if len(addresses) > 0 {
			addressMap := addresses[0].(map[string]interface{})
			otherEmails = addressMap["other_emails"].(*schema.Set)
			phoneNumbers = addressMap["phone_numbers"].(*schema.Set)
		}

		sdkAddresses := make([]platformclientv2.Contact, 0)
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

func flattenUserAddresses(addresses []platformclientv2.Contact) []interface{} {
	emailSet := schema.NewSet(schema.HashResource(otherEmailResource), []interface{}{})
	phoneNumSet := schema.NewSet(phoneNumberHash, []interface{}{})

	for _, address := range addresses {
		if address.MediaType != nil {
			if *address.MediaType == "SMS" || *address.MediaType == "PHONE" {
				phoneNumber := make(map[string]interface{})
				phoneNumber["media_type"] = *address.MediaType

				// Strip off any parentheses from phone numbers
				if address.Address != nil {
					phoneNumber["number"] = strings.Trim(*address.Address, "()")
				} else if address.Display != nil {
					// Some numbers are only returned in Display
					phoneNumber["number"] = strings.Trim(*address.Display, "()")
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

func updateUserSkills(d *schema.ResourceData, usersAPI *platformclientv2.UsersApi) diag.Diagnostics {
	if skillsConfig, ok := d.GetOk("routing_skills"); ok && d.HasChange("routing_skills") {
		sdkSkills := make([]platformclientv2.Userroutingskillpost, 0)

		skillsList := skillsConfig.(*schema.Set).List()
		for _, configSkill := range skillsList {
			skillMap := configSkill.(map[string]interface{})
			skillID := skillMap["skill_id"].(string)
			skillProf := skillMap["proficiency"].(float64)

			sdkSkills = append(sdkSkills, platformclientv2.Userroutingskillpost{
				Id:          &skillID,
				Proficiency: &skillProf,
			})
		}

		_, _, err := usersAPI.PutUserRoutingskillsBulk(d.Id(), sdkSkills)
		if err != nil {
			return diag.Errorf("Failed to update skills for user %s: %s", d.Id(), err)
		}
	}
	return nil
}

func flattenUserSkills(skills []platformclientv2.Userroutingskill) *schema.Set {
	skillSet := schema.NewSet(schema.HashResource(userSkillResource), []interface{}{})
	for _, sdkSkill := range skills {
		skill := make(map[string]interface{})
		skill["skill_id"] = *sdkSkill.Id
		skill["proficiency"] = *sdkSkill.Proficiency
		skillSet.Add(skill)
	}
	return skillSet
}

func createRoleDivisionPair(roleID string, divisionID string) string {
	return roleID + ":" + divisionID
}

func roleDivPairsToGrants(grantPairs []string) platformclientv2.Roledivisiongrants {
	grants := make([]platformclientv2.Roledivisionpair, len(grantPairs))
	for i, pair := range grantPairs {
		roleDiv := strings.Split(pair, ":")
		grants[i] = platformclientv2.Roledivisionpair{
			RoleId:     &roleDiv[0],
			DivisionId: &roleDiv[1],
		}
	}
	return platformclientv2.Roledivisiongrants{
		Grants: &grants,
	}
}

func updateUserRoles(d *schema.ResourceData, authAPI *platformclientv2.AuthorizationApi) diag.Diagnostics {
	if d.HasChange("roles") {
		rolesConfig := d.Get("roles")
		if rolesConfig != nil {
			// Get existing roles/divisions
			subject, _, err := authAPI.GetAuthorizationSubject(d.Id())
			if err != nil {
				return diag.Errorf("Failed to get current roles for user %s: %s", d.Id(), err)
			}

			var existingGrants []string
			if subject != nil && subject.Grants != nil {
				for _, grant := range *subject.Grants {
					existingGrants = append(existingGrants, createRoleDivisionPair(*grant.Role.Id, *grant.Division.Id))
				}
			}

			homeDiv, diagErr := getHomeDivisionID()
			if diagErr != nil {
				return diagErr
			}

			var configGrants []string
			rolesList := rolesConfig.(*schema.Set).List()
			for _, configRole := range rolesList {
				roleMap := configRole.(map[string]interface{})
				roleID := roleMap["role_id"].(string)

				var divisionIDs []string
				if configDivs, ok := roleMap["division_ids"]; ok {
					divisionIDs = *setToStringList(configDivs.(*schema.Set))
				}

				if len(divisionIDs) == 0 {
					// No division set. Use the home division
					divisionIDs = []string{homeDiv}
				}

				for _, divID := range divisionIDs {
					configGrants = append(configGrants, createRoleDivisionPair(roleID, divID))
				}
			}

			grantsToRemove := sliceDifference(existingGrants, configGrants)
			if len(grantsToRemove) > 0 {
				_, err := authAPI.PostAuthorizationSubjectBulkremove(d.Id(), roleDivPairsToGrants(grantsToRemove))
				if err != nil {
					return diag.Errorf("Failed to remove role grants for user %s: %s", d.Id(), err)
				}
			}

			grantsToAdd := sliceDifference(configGrants, existingGrants)
			if len(grantsToAdd) > 0 {
				_, err := authAPI.PostAuthorizationSubjectBulkadd(d.Id(), roleDivPairsToGrants(grantsToAdd), "PC_USER")
				if err != nil {
					return diag.Errorf("Failed to add role grants for user %s: %s", d.Id(), err)
				}
			}
		}
	}
	return nil
}

func readUserRoles(userID string, authAPI *platformclientv2.AuthorizationApi) (*schema.Set, diag.Diagnostics) {
	subject, _, err := authAPI.GetAuthorizationSubject(userID)
	if err != nil {
		return nil, diag.Errorf("Failed to get current roles for user %s: %s", userID, err)
	}

	roleDivsMap := make(map[string]*schema.Set)
	if subject != nil && subject.Grants != nil {
		for _, grant := range *subject.Grants {
			if currentDivs, ok := roleDivsMap[*grant.Role.Id]; ok {
				currentDivs.Add(*grant.Division.Id)
			} else {
				roleDivsMap[*grant.Role.Id] = schema.NewSet(schema.HashString, []interface{}{*grant.Division.Id})
			}
		}
	}

	roleSet := schema.NewSet(schema.HashResource(userRoleResource), []interface{}{})
	for roleID, divs := range roleDivsMap {
		role := make(map[string]interface{})
		role["role_id"] = roleID
		role["division_ids"] = divs
		roleSet.Add(role)
	}
	return roleSet, nil
}
