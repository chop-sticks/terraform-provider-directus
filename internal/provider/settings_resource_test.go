// Copyright IBM Corp. 2026

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccSettingsResource(t *testing.T) {
	descriptor := fmt.Sprintf("tf-acc-%d", acctestSuffix())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSettingsConfig(descriptor),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("directus_settings.test", "project_descriptor", descriptor),
					resource.TestCheckResourceAttrSet("directus_settings.test", "id"),
				),
			},
			{
				ResourceName:      "directus_settings.test",
				ImportState:       true,
				ImportStateId:     "1",
				ImportStateVerify: true,
			},
			{
				Config: testAccSettingsConfig(descriptor + "-updated"),
				Check:  resource.TestCheckResourceAttr("directus_settings.test", "project_descriptor", descriptor+"-updated"),
			},
		},
	})
}

func testAccSettingsConfig(descriptor string) string {
	return fmt.Sprintf(`
resource "directus_settings" "test" {
  project_descriptor = %[1]q
}
`, descriptor)
}
