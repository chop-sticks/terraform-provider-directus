// Copyright IBM Corp. 2026

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccCollectionResource(t *testing.T) {
	name := fmt.Sprintf("tf_acc_col_%d", acctestSuffix())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCollectionConfig(name, "box", "created by acc test"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("directus_collection.test", "collection", name),
					resource.TestCheckResourceAttr("directus_collection.test", "meta.icon", "box"),
					resource.TestCheckResourceAttr("directus_collection.test", "meta.note", "created by acc test"),
				),
			},
			{
				ResourceName:                         "directus_collection.test",
				ImportState:                          true,
				ImportStateId:                        name,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "collection",
				// Optional nested blocks are only tracked when managed; on import
				// there is no prior state to indicate management, so skip verify.
				ImportStateVerifyIgnore: []string{"meta", "schema"},
			},
			{
				Config: testAccCollectionConfig(name, "database", "updated by acc test"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("directus_collection.test", "meta.icon", "database"),
					resource.TestCheckResourceAttr("directus_collection.test", "meta.note", "updated by acc test"),
				),
			},
		},
	})
}

func testAccCollectionConfig(name, icon, note string) string {
	return fmt.Sprintf(`
resource "directus_collection" "test" {
  collection = %[1]q
  meta = {
    icon = %[2]q
    note = %[3]q
  }
}
`, name, icon, note)
}
