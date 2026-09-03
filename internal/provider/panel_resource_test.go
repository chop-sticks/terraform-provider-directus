// Copyright IBM Corp. 2026

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccPanelResource(t *testing.T) {
	name := fmt.Sprintf("tf_acc_panel_%d", acctestSuffix())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccPanelConfig(name, "hello"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("directus_panel.test", "name", name),
					resource.TestCheckResourceAttr("directus_panel.test", "type", "label"),
					resource.TestCheckResourceAttrSet("directus_panel.test", "id"),
					resource.TestCheckResourceAttrSet("directus_panel.test", "dashboard"),
				),
			},
			{
				ResourceName:      "directus_panel.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccPanelConfig(name, "updated"),
				Check:  resource.TestCheckResourceAttrSet("directus_panel.test", "options"),
			},
		},
	})
}

func testAccPanelConfig(name, text string) string {
	return fmt.Sprintf(`
resource "directus_dashboard" "test" {
  name = "%[1]s_dash"
}

resource "directus_panel" "test" {
  dashboard  = directus_dashboard.test.id
  name       = %[1]q
  type       = "label"
  position_x = 1
  position_y = 1
  width      = 4
  height     = 4
  options    = jsonencode({ text = %[2]q })
}
`, name, text)
}
