// Copyright IBM Corp. 2026

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDashboardResource(t *testing.T) {
	name := fmt.Sprintf("tf_acc_dash_%d", acctestSuffix())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDashboardConfig(name, "dashboard", "acc dashboard"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("directus_dashboard.test", "name", name),
					resource.TestCheckResourceAttr("directus_dashboard.test", "icon", "dashboard"),
					resource.TestCheckResourceAttr("directus_dashboard.test", "note", "acc dashboard"),
					resource.TestCheckResourceAttrSet("directus_dashboard.test", "id"),
				),
			},
			{
				ResourceName:      "directus_dashboard.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDashboardConfig(name, "insights", "acc dashboard updated"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("directus_dashboard.test", "icon", "insights"),
					resource.TestCheckResourceAttr("directus_dashboard.test", "note", "acc dashboard updated"),
				),
			},
		},
	})
}

func testAccDashboardConfig(name, icon, note string) string {
	return fmt.Sprintf(`
resource "directus_dashboard" "test" {
  name = %[1]q
  icon = %[2]q
  note = %[3]q
}
`, name, icon, note)
}
