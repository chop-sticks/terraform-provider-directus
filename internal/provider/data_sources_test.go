// Copyright IBM Corp. 2026

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccServerInfoDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `data "directus_server_info" "test" {}`,
				Check:  resource.TestCheckResourceAttrSet("data.directus_server_info.test", "info"),
			},
		},
	})
}

func TestAccCollectionDataSource(t *testing.T) {
	name := fmt.Sprintf("tf_acc_dscol_%d", acctestSuffix())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "directus_collection" "test" {
  collection = %[1]q
  meta = {
    icon = "box"
  }
}

data "directus_collection" "test" {
  collection = directus_collection.test.collection
}
`, name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.directus_collection.test", "collection", name),
					resource.TestCheckResourceAttr("data.directus_collection.test", "meta.icon", "box"),
				),
			},
		},
	})
}
