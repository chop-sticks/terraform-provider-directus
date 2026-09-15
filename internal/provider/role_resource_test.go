// Copyright IBM Corp. 2026

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccRoleResource(t *testing.T) {
	name := fmt.Sprintf("tf_acc_role_%d", acctestSuffix())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRoleConfig(name, "verified_user", "acc role"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("directus_role.test", "name", name),
					resource.TestCheckResourceAttr("directus_role.test", "icon", "verified_user"),
					resource.TestCheckResourceAttr("directus_role.test", "description", "acc role"),
					resource.TestCheckResourceAttrSet("directus_role.test", "id"),
				),
			},
			{
				ResourceName:      "directus_role.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccRoleConfig(name+"_upd", "supervised_user_circle", "acc role updated"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("directus_role.test", "name", name+"_upd"),
					resource.TestCheckResourceAttr("directus_role.test", "icon", "supervised_user_circle"),
				),
			},
		},
	})
}

func testAccRoleConfig(name, icon, description string) string {
	return fmt.Sprintf(`
resource "directus_role" "test" {
  name        = %[1]q
  icon        = %[2]q
  description = %[3]q
}
`, name, icon, description)
}
