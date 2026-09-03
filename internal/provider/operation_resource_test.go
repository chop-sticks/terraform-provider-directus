// Copyright IBM Corp. 2026

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccOperationResource(t *testing.T) {
	name := fmt.Sprintf("tf_acc_op_%d", acctestSuffix())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccOperationConfig(name, "First log"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("directus_operation.test", "type", "log"),
					resource.TestCheckResourceAttr("directus_operation.test", "name", "First log"),
					resource.TestCheckResourceAttrSet("directus_operation.test", "id"),
					resource.TestCheckResourceAttrSet("directus_operation.test", "flow"),
				),
			},
			{
				ResourceName:      "directus_operation.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccOperationConfig(name, "Renamed log"),
				Check:  resource.TestCheckResourceAttr("directus_operation.test", "name", "Renamed log"),
			},
		},
	})
}

func testAccOperationConfig(name, opName string) string {
	return fmt.Sprintf(`
resource "directus_flow" "test" {
  name    = %[1]q
  trigger = "manual"
}

resource "directus_operation" "test" {
  flow       = directus_flow.test.id
  key        = "log_step"
  type       = "log"
  name       = %[2]q
  position_x = 20
  position_y = 1
  options    = jsonencode({ message = "hello" })
}
`, name, opName)
}
