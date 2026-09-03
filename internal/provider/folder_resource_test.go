// Copyright IBM Corp. 2026

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccFolderResource(t *testing.T) {
	name := fmt.Sprintf("tf_acc_folder_%d", acctestSuffix())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccFolderConfig(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("directus_folder.test", "name", name),
					resource.TestCheckResourceAttrSet("directus_folder.test", "id"),
				),
			},
			{
				ResourceName:      "directus_folder.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				// Renaming forces replacement (no update endpoint).
				Config: testAccFolderConfig(name + "_replaced"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("directus_folder.test", "name", name+"_replaced"),
				),
			},
		},
	})
}

func testAccFolderConfig(name string) string {
	return fmt.Sprintf(`
resource "directus_folder" "test" {
  name = %[1]q
}
`, name)
}
