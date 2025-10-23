package user

import (
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	frameworkresource "github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
)

// Ensure test resources are initialized for Framework tests
func init() {
	if frameworkResources == nil || frameworkDataSources == nil {
		initTestResources()
	}
}

func TestAccFrameworkDataSourceUser(t *testing.T) {
	t.Parallel()
	var (
		userResourceLabel   = "test-user-resource"
		userDataSourceLabel = "test-user-data-source"
		randomString        = uuid.NewString()
		userEmail           = "framework_user_" + randomString + "@example.com"
		userName            = "Framework_User_" + randomString
	)

	resource.Test(t, resource.TestCase{
		PreCheck: func() { util.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: provider.GetMuxedProviderFactories(
			nil, // SDKv2 resources removed
			nil, // SDKv2 data sources removed
			map[string]func() frameworkresource.Resource{
				ResourceType: NewUserFrameworkResource,
			},
			map[string]func() datasource.DataSource{
				ResourceType: NewUserFrameworkDataSource,
			},
		),
		Steps: []resource.TestStep{
			{
				// Search by email
				Config: generateFrameworkUserResource(
					userResourceLabel,
					userEmail,
					userName,
					util.NullValue, // Active
					util.NullValue, // No title
					util.NullValue, // No department
					util.NullValue, // No manager
					util.NullValue, // Default acdAutoAnswer
				) + generateFrameworkUserDataSource(
					userDataSourceLabel,
					ResourceType+"."+userResourceLabel+".email",
					util.NullValue,
					ResourceType+"."+userResourceLabel,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data."+ResourceType+"."+userDataSourceLabel, "id", ResourceType+"."+userResourceLabel, "id"),
					resource.TestCheckResourceAttrPair("data."+ResourceType+"."+userDataSourceLabel, "name", ResourceType+"."+userResourceLabel, "name"),
					func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources[ResourceType+"."+userResourceLabel]
						if !ok {
							return fmt.Errorf("not found: %s", ResourceType+"."+userResourceLabel)
						}
						// Verify user ID is set
						if rs.Primary.ID == "" {
							return fmt.Errorf("user ID is empty")
						}
						return nil
					},
				),
			},
			{
				// Search by name
				Config: generateFrameworkUserResource(
					userResourceLabel,
					userEmail,
					userName,
					util.NullValue, // Active
					util.NullValue, // No title
					util.NullValue, // No department
					util.NullValue, // No manager
					util.NullValue, // Default acdAutoAnswer
				) + generateFrameworkUserDataSource(
					userDataSourceLabel,
					util.NullValue,
					ResourceType+"."+userResourceLabel+".name",
					ResourceType+"."+userResourceLabel,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data."+ResourceType+"."+userDataSourceLabel, "id", ResourceType+"."+userResourceLabel, "id"),
					resource.TestCheckResourceAttrPair("data."+ResourceType+"."+userDataSourceLabel, "name", ResourceType+"."+userResourceLabel, "name"),
					func(s *terraform.State) error {
						time.Sleep(30 * time.Second) // Wait for proper cleanup
						return nil
					},
				),
			},
		},
		CheckDestroy: func(state *terraform.State) error {
			time.Sleep(45 * time.Second)
			return testVerifyUsersDestroyed(state)
		},
	})
}

func TestAccFrameworkDataSourceUserByEmail(t *testing.T) {
	t.Parallel()
	var (
		userResourceLabel   = "test-user-email-lookup"
		userDataSourceLabel = "test-user-email-data"
		randomString        = uuid.NewString()
		userEmail           = "email_lookup_" + randomString + "@example.com"
		userName            = "Email_Lookup_User_" + randomString
		title               = "Data Source Test User"
		department          = "Testing Department"
	)

	resource.Test(t, resource.TestCase{
		PreCheck: func() { util.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: provider.GetMuxedProviderFactories(
			nil, // SDKv2 resources removed
			nil, // SDKv2 data sources removed
			map[string]func() frameworkresource.Resource{
				ResourceType: NewUserFrameworkResource,
			},
			map[string]func() datasource.DataSource{
				ResourceType: NewUserFrameworkDataSource,
			},
		),
		Steps: []resource.TestStep{
			{
				// Create user and lookup by email
				Config: generateFrameworkUserResource(
					userResourceLabel,
					userEmail,
					userName,
					util.NullValue,                  // Active
					fmt.Sprintf(`"%s"`, title),      // Title
					fmt.Sprintf(`"%s"`, department), // Department
					util.NullValue,                  // No manager
					util.NullValue,                  // Default acdAutoAnswer
				) + generateFrameworkUserDataSource(
					userDataSourceLabel,
					fmt.Sprintf(`"%s"`, userEmail), // Direct email lookup
					util.NullValue,
					ResourceType+"."+userResourceLabel,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data."+ResourceType+"."+userDataSourceLabel, "id", ResourceType+"."+userResourceLabel, "id"),
					resource.TestCheckResourceAttr("data."+ResourceType+"."+userDataSourceLabel, "name", userName),
					resource.TestCheckResourceAttr("data."+ResourceType+"."+userDataSourceLabel, "email", userEmail),
				),
			},
		},
		CheckDestroy: testVerifyUsersDestroyed,
	})
}

func TestAccFrameworkDataSourceUserByName(t *testing.T) {
	t.Parallel()
	var (
		userResourceLabel   = "test-user-name-lookup"
		userDataSourceLabel = "test-user-name-data"
		randomString        = uuid.NewString()
		userEmail           = "name_lookup_" + randomString + "@example.com"
		userName            = "Name_Lookup_User_" + randomString
		title               = "Name Lookup Test User"
		department          = "Name Testing Department"
	)

	resource.Test(t, resource.TestCase{
		PreCheck: func() { util.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: provider.GetMuxedProviderFactories(
			nil, // SDKv2 resources removed
			nil, // SDKv2 data sources removed
			map[string]func() frameworkresource.Resource{
				ResourceType: NewUserFrameworkResource,
			},
			map[string]func() datasource.DataSource{
				ResourceType: NewUserFrameworkDataSource,
			},
		),
		Steps: []resource.TestStep{
			{
				// Create user and lookup by name
				Config: generateFrameworkUserResource(
					userResourceLabel,
					userEmail,
					userName,
					util.NullValue,                  // Active
					fmt.Sprintf(`"%s"`, title),      // Title
					fmt.Sprintf(`"%s"`, department), // Department
					util.NullValue,                  // No manager
					util.NullValue,                  // Default acdAutoAnswer
				) + generateFrameworkUserDataSource(
					userDataSourceLabel,
					util.NullValue,
					fmt.Sprintf(`"%s"`, userName), // Direct name lookup
					ResourceType+"."+userResourceLabel,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data."+ResourceType+"."+userDataSourceLabel, "id", ResourceType+"."+userResourceLabel, "id"),
					resource.TestCheckResourceAttr("data."+ResourceType+"."+userDataSourceLabel, "name", userName),
					resource.TestCheckResourceAttr("data."+ResourceType+"."+userDataSourceLabel, "email", userEmail),
				),
			},
		},
		CheckDestroy: testVerifyUsersDestroyed,
	})
}

func TestAccFrameworkDataSourceUserAttributeVerification(t *testing.T) {
	t.Parallel()
	var (
		userResourceLabel   = "test-user-attrs"
		userDataSourceLabel = "test-user-attrs-data"
		randomString        = uuid.NewString()
		userEmail           = "attrs_test_" + randomString + "@example.com"
		userName            = "Attrs_Test_User_" + randomString
		profileSkill        = "Framework Testing"
		certification       = "Framework Certification"
	)

	resource.Test(t, resource.TestCase{
		PreCheck: func() { util.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: provider.GetMuxedProviderFactories(
			nil, // SDKv2 resources removed
			nil, // SDKv2 data sources removed
			map[string]func() frameworkresource.Resource{
				ResourceType: NewUserFrameworkResource,
			},
			map[string]func() datasource.DataSource{
				ResourceType: NewUserFrameworkDataSource,
			},
		),
		Steps: []resource.TestStep{
			{
				// Create user with various attributes and verify data source returns them
				Config: generateFrameworkUserWithProfileAttrs(
					userResourceLabel,
					userEmail,
					userName,
					generateProfileSkills(profileSkill),
					generateCertifications(certification),
				) + generateFrameworkUserDataSource(
					userDataSourceLabel,
					fmt.Sprintf(`"%s"`, userEmail),
					util.NullValue,
					ResourceType+"."+userResourceLabel,
				),
				Check: resource.ComposeTestCheckFunc(
					// Verify basic attributes
					resource.TestCheckResourceAttrPair("data."+ResourceType+"."+userDataSourceLabel, "id", ResourceType+"."+userResourceLabel, "id"),
					resource.TestCheckResourceAttrPair("data."+ResourceType+"."+userDataSourceLabel, "name", ResourceType+"."+userResourceLabel, "name"),
					resource.TestCheckResourceAttrPair("data."+ResourceType+"."+userDataSourceLabel, "email", ResourceType+"."+userResourceLabel, "email"),
					resource.TestCheckResourceAttrPair("data."+ResourceType+"."+userDataSourceLabel, "state", ResourceType+"."+userResourceLabel, "state"),
					resource.TestCheckResourceAttrPair("data."+ResourceType+"."+userDataSourceLabel, "division_id", ResourceType+"."+userResourceLabel, "division_id"),

					// Verify computed attributes are set
					resource.TestCheckResourceAttrSet("data."+ResourceType+"."+userDataSourceLabel, "id"),
					resource.TestCheckResourceAttrSet("data."+ResourceType+"."+userDataSourceLabel, "division_id"),
				),
			},
		},
		CheckDestroy: testVerifyUsersDestroyed,
	})
}

// Helper function to generate Framework user data source configuration
func generateFrameworkUserDataSource(
	resourceLabel string,
	email string,
	name string,
	dependsOnResource string,
) string {
	return fmt.Sprintf(`data "%s" "%s" {
		email = %s
		name = %s
		depends_on = [%s]
	}`, ResourceType, resourceLabel, email, name, dependsOnResource)
}
