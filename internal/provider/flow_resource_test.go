// Copyright IBM Corp. 2026

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccFlowResource(t *testing.T) {
	name := fmt.Sprintf("tf_acc_flow_%d", acctestSuffix())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccFlowConfig(name, "active"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("directus_flow.test", "name", name),
					resource.TestCheckResourceAttr("directus_flow.test", "trigger", "manual"),
					resource.TestCheckResourceAttr("directus_flow.test", "status", "active"),
					resource.TestCheckResourceAttrSet("directus_flow.test", "id"),
				),
			},
			{
				ResourceName:      "directus_flow.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccFlowConfig(name, "inactive"),
				Check:  resource.TestCheckResourceAttr("directus_flow.test", "status", "inactive"),
			},
		},
	})
}

func testAccFlowConfig(name, status string) string {
	return fmt.Sprintf(`
resource "directus_flow" "test" {
  name    = %[1]q
  trigger = "manual"
  status  = %[2]q
  options = jsonencode({ collections = ["directus_users"] })
}
`, name, status)
}
