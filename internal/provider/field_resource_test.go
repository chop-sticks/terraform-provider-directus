// Copyright IBM Corp. 2026

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccFieldResource(t *testing.T) {
	collection := fmt.Sprintf("tf_acc_fcol_%d", acctestSuffix())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccFieldConfig(collection, "input"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("directus_field.test", "collection", collection),
					resource.TestCheckResourceAttr("directus_field.test", "field", "title"),
					resource.TestCheckResourceAttr("directus_field.test", "type", "string"),
					resource.TestCheckResourceAttr("directus_field.test", "meta.interface", "input"),
				),
			},
			{
				ResourceName:                         "directus_field.test",
				ImportState:                          true,
				ImportStateId:                        collection + "/title",
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "field",
				ImportStateVerifyIgnore:              []string{"meta", "schema"},
			},
			{
				Config: testAccFieldConfig(collection, "input-multiline"),
				Check:  resource.TestCheckResourceAttr("directus_field.test", "meta.interface", "input-multiline"),
			},
		},
	})
}

func testAccFieldConfig(collection, iface string) string {
	return fmt.Sprintf(`
resource "directus_collection" "test" {
  collection = %[1]q
}

resource "directus_field" "test" {
  collection = directus_collection.test.collection
  field      = "title"
  type       = "string"
  meta = {
    interface = %[2]q
    required  = true
  }
  schema = {
    is_nullable = false
    max_length  = 255
  }
}
`, collection, iface)
}
