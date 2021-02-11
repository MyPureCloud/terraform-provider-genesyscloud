package genesyscloud

import (
	"context"
	"log"

	"github.com/MyPureCloud/platform-client-sdk-go/platformclientv2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
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
				Description: "User's email and username.",
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
			"division_id": {
				Description: "The division to which this user will belong. If not set, the home division will be used.",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
			},
		},
	}
}

func createUser(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	email := d.Get("email").(string)
	name := d.Get("name").(string)
	password := d.Get("password").(string)

	usersAPI := platformclientv2.NewUsersApi()
	authAPI := platformclientv2.NewAuthorizationApi()

	divisionID, diagErr := getDivisionOrDefault(d, authAPI)
	if diagErr != nil {
		return diagErr
	}

	createUser := platformclientv2.Createuser{
		Email:      &email,
		Name:       &name,
		DivisionId: &divisionID,
	}

	if password != "" {
		createUser.Password = &password
	}

	log.Printf("Creating user %s", email)
	user, _, err := usersAPI.PostUsers(createUser)
	if err != nil {
		// TODO: Handle restoring previously deleted users with the same email
		return diag.Errorf("Failed to create user %s: %s", email, err)
	}

	// Set ID
	d.SetId(*user.Id)

	// Set optional attribute defaults
	d.Set("division_id", createUser.DivisionId)

	return nil
}

func readUser(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	usersAPI := platformclientv2.NewUsersApi()

	currentUser, _, getErr := usersAPI.GetUser(d.Id(), nil, "active")
	if getErr != nil {
		return diag.Errorf("Failed to read user %s: %s", d.Id(), getErr)
	}

	d.Set("name", *currentUser.Name)
	d.Set("email", *currentUser.Email)
	d.Set("division_id", *currentUser.Division.Id)

	return nil
}

func updateUser(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)
	email := d.Get("email").(string)

	usersAPI := platformclientv2.NewUsersApi()
	authAPI := platformclientv2.NewAuthorizationApi()

	divisionID, diagErr := getDivisionOrDefault(d, authAPI)
	if diagErr != nil {
		return diagErr
	}

	// Need the current user version for patches
	currentUser, _, getErr := usersAPI.GetUser(d.Id(), nil, "active")
	if getErr != nil {
		return diag.Errorf("Failed to read user %s: %s", d.Id(), getErr)
	}

	log.Printf("Updating user %s", email)
	_, _, patchErr := usersAPI.PatchUser(d.Id(), platformclientv2.Updateuser{
		Name:    &name,
		Email:   &email,
		Version: currentUser.Version,
	})
	if patchErr != nil {
		return diag.Errorf("Failed to update user %s: %s", email, patchErr)
	}

	if currentUser.Division == nil || *currentUser.Division.Id != divisionID {
		log.Printf("Updating division for user %s to %s", email, divisionID)
		_, divErr := authAPI.PostAuthorizationDivisionObject(divisionID, "USER", []string{d.Id()})
		if divErr != nil {
			return diag.Errorf("Failed to update division for user %s: %s", email, divErr)
		}
		d.Set("division_id", divisionID)
	}

	return nil
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

func getDivisionOrDefault(d *schema.ResourceData, authAPI *platformclientv2.AuthorizationApi) (string, diag.Diagnostics) {
	divisionID := d.Get("division_id").(string)
	if divisionID != "" {
		return divisionID, nil
	}
	// Use home division as default
	homeDiv, _, err := authAPI.GetAuthorizationDivisionsHome()
	if err != nil {
		return "", diag.Errorf("Failed to query home division: %s", err)
	}
	return *homeDiv.Id, nil
}
