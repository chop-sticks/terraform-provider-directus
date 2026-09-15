// Copyright IBM Corp. 2026

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccPolicyResource(t *testing.T) {
	name := fmt.Sprintf("tf_acc_policy_%d", acctestSuffix())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyConfig(name, "acc policy", true),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("directus_policy.test", "name", name),
					resource.TestCheckResourceAttr("directus_policy.test", "description", "acc policy"),
					resource.TestCheckResourceAttr("directus_policy.test", "app_access", "true"),
					resource.TestCheckResourceAttrSet("directus_policy.test", "id"),
				),
			},
			{
				ResourceName:      "directus_policy.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccPolicyConfig(name, "acc policy updated", false),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("directus_policy.test", "description", "acc policy updated"),
					resource.TestCheckResourceAttr("directus_policy.test", "app_access", "false"),
				),
			},
		},
	})
}

func testAccPolicyConfig(name, description string, appAccess bool) string {
	return fmt.Sprintf(`
resource "directus_policy" "test" {
  name        = %[1]q
  description = %[2]q
  app_access  = %[3]t
}
`, name, description, appAccess)
}
