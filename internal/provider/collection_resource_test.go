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
				Config: testAccCollectionConfig(name, "box", "created by acc test", ""),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("directus_collection.test", "collection", name),
					resource.TestCheckResourceAttr("directus_collection.test", "meta.icon", "box"),
					resource.TestCheckResourceAttr("directus_collection.test", "meta.note", "created by acc test"),
					// collapse is unset in config; the schema default applies.
					resource.TestCheckResourceAttr("directus_collection.test", "meta.collapse", "open"),
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
				Config: testAccCollectionConfig(name, "database", "updated by acc test", "https://example.com/{{id}}"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("directus_collection.test", "meta.icon", "database"),
					resource.TestCheckResourceAttr("directus_collection.test", "meta.note", "updated by acc test"),
					resource.TestCheckResourceAttr("directus_collection.test", "meta.preview_url", "https://example.com/{{id}}"),
				),
			},
		},
	})
}

func testAccCollectionConfig(name, icon, note, previewURL string) string {
	previewLine := ""
	if previewURL != "" {
		previewLine = fmt.Sprintf("\n    preview_url = %q", previewURL)
	}
	return fmt.Sprintf(`
resource "directus_collection" "test" {
  collection = %[1]q
  meta = {
    icon = %[2]q
    note = %[3]q%[4]s
  }
}
`, name, icon, note, previewLine)
}
