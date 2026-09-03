// Copyright IBM Corp. 2026

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccPresetResource(t *testing.T) {
	bookmark := fmt.Sprintf("tf_acc_preset_%d", acctestSuffix())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccPresetConfig(bookmark, "tabular"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("directus_preset.test", "bookmark", bookmark),
					resource.TestCheckResourceAttr("directus_preset.test", "collection", "directus_users"),
					resource.TestCheckResourceAttr("directus_preset.test", "layout", "tabular"),
					resource.TestCheckResourceAttrSet("directus_preset.test", "id"),
					resource.TestCheckResourceAttrSet("directus_preset.test", "filter"),
				),
			},
			{
				ResourceName:      "directus_preset.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccPresetConfig(bookmark, "cards"),
				Check:  resource.TestCheckResourceAttr("directus_preset.test", "layout", "cards"),
			},
		},
	})
}

func testAccPresetConfig(bookmark, layout string) string {
	return fmt.Sprintf(`
resource "directus_preset" "test" {
  bookmark   = %[1]q
  collection = "directus_users"
  layout     = %[2]q
  filter     = jsonencode({ status = { _eq = "active" } })
}
`, bookmark, layout)
}
