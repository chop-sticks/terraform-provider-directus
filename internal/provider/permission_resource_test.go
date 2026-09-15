// Copyright IBM Corp. 2026

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccPermissionResource(t *testing.T) {
	name := fmt.Sprintf("tf_acc_perm_policy_%d", acctestSuffix())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccPermissionConfig(name, "read"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("directus_permission.test", "collection", "directus_users"),
					resource.TestCheckResourceAttr("directus_permission.test", "action", "read"),
					resource.TestCheckResourceAttrSet("directus_permission.test", "id"),
					resource.TestCheckResourceAttrSet("directus_permission.test", "policy"),
				),
			},
			{
				ResourceName:      "directus_permission.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccPermissionConfig(name, "update"),
				Check:  resource.TestCheckResourceAttr("directus_permission.test", "action", "update"),
			},
		},
	})
}

func testAccPermissionConfig(name, action string) string {
	return fmt.Sprintf(`
resource "directus_policy" "test" {
  name = %[1]q
}

resource "directus_permission" "test" {
  policy     = directus_policy.test.id
  collection = "directus_users"
  action     = %[2]q
  fields     = ["*"]
}
`, name, action)
}
