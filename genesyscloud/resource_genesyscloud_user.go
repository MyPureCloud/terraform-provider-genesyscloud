package genesyscloud

import (
	"context"
	"log"

	"github.com/MyPureCloud/platform-client-sdk-go/platformclientv2"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
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
				Description:  "Media type of phone number (SMS | PHONE). Defaults to PHONE",
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "PHONE",
				ValidateFunc: validation.StringInSlice([]string{"PHONE", "SMS"}, false),
			},
			"type": {
				Description:  "Type of number (WORK | WORK2 | WORK3 | WORK4 | HOME | MOBILE). Defaults to WORK",
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
				Description:  "Type of email address (WORK | HOME). Defaults to WORK.",
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "WORK",
				ValidateFunc: validation.StringInSlice([]string{"WORK", "HOME"}, false),
			},
		},
	}
)

func userResource() *schema.Resource {
	return &schema.Resource{
		Description: "Genesys Cloud User",

		CreateContext: createUser,
		ReadContext:   readUser,
		UpdateContext: updateUser,
		DeleteContext: deleteUser,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
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

	usersAPI := platformclientv2.NewUsersApi()

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

	return readUser(ctx, d, meta)
}

func readUser(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	usersAPI := platformclientv2.NewUsersApi()

	currentUser, _, getErr := usersAPI.GetUser(d.Id(), nil, "", "")
	if getErr != nil {
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
		emailSet, phoneNumSet := flattenUserAddresses(*currentUser.Addresses)
		d.Set("other_emails", emailSet)
		d.Set("phone_numbers", phoneNumSet)
	}

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

	usersAPI := platformclientv2.NewUsersApi()
	authAPI := platformclientv2.NewAuthorizationApi()

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

	log.Printf("Updating user %s", email)
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

	return readUser(ctx, d, meta)
}

func deleteUser(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	email := d.Get("email").(string)

	usersAPI := platformclientv2.NewUsersApi()

	log.Printf("Deleting user %s", email)
	_, _, err := usersAPI.DeleteUser(d.Id())
	if err != nil {
		return diag.Errorf("Failed to delete user %s: %s", email, err)
	}
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
	otherEmails := d.Get("other_emails")
	phoneNumbers := d.Get("phone_numbers")

	var addresses []platformclientv2.Contact
	if otherEmails != nil {
		addresses = append(addresses, buildSdkEmails(otherEmails.(*schema.Set))...)
	}
	if phoneNumbers != nil {
		sdkNums, err := buildSdkPhoneNumbers(phoneNumbers.(*schema.Set))
		if err != nil {
			return nil, err
		}
		addresses = append(addresses, sdkNums...)
	}
	return &addresses, nil
}

func flattenUserAddresses(addresses []platformclientv2.Contact) (emailSet *schema.Set, phoneNumSet *schema.Set) {
	emailSet = schema.NewSet(schema.HashResource(otherEmailResource), []interface{}{})
	phoneNumSet = schema.NewSet(phoneNumberHash, []interface{}{})

	for _, address := range addresses {
		if address.MediaType != nil {
			if *address.MediaType == "SMS" || *address.MediaType == "PHONE" {
				phoneNumber := make(map[string]interface{})
				phoneNumber["media_type"] = *address.MediaType

				if address.Address != nil {
					phoneNumber["number"] = *address.Address
				} else if address.Display != nil {
					// Some numbers are only returned in Display
					phoneNumber["number"] = *address.Display
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
			}
		}
	}
	return
}
