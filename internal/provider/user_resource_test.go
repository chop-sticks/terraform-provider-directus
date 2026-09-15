// Copyright IBM Corp. 2026

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccUserResource(t *testing.T) {
	suffix := acctestSuffix()
	email := fmt.Sprintf("tf_acc_%d@example.com", suffix)
	role := fmt.Sprintf("tf_acc_user_role_%d", suffix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccUserConfig(email, role, "Ada"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("directus_user.test", "email", email),
					resource.TestCheckResourceAttr("directus_user.test", "first_name", "Ada"),
					resource.TestCheckResourceAttrSet("directus_user.test", "id"),
					resource.TestCheckResourceAttrSet("directus_user.test", "role"),
				),
			},
			{
				ResourceName:            "directus_user.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"password"},
			},
			{
				Config: testAccUserConfig(email, role, "Grace"),
				Check:  resource.TestCheckResourceAttr("directus_user.test", "first_name", "Grace"),
			},
		},
	})
}

func testAccUserConfig(email, role, firstName string) string {
	return fmt.Sprintf(`
resource "directus_role" "test" {
  name = %[2]q
}

resource "directus_user" "test" {
  email      = %[1]q
  password   = "s3cret-passw0rd"
  first_name = %[3]q
  role       = directus_role.test.id
}
`, email, role, firstName)
}
