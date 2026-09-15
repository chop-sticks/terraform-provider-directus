// Copyright IBM Corp. 2026

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccRelationResource(t *testing.T) {
	suffix := acctestSuffix()
	articles := fmt.Sprintf("tf_acc_articles_%d", suffix)
	authors := fmt.Sprintf("tf_acc_authors_%d", suffix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRelationConfig(articles, authors, "SET NULL"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("directus_relation.test", "collection", articles),
					resource.TestCheckResourceAttr("directus_relation.test", "field", "author_id"),
					resource.TestCheckResourceAttr("directus_relation.test", "related_collection", authors),
				),
			},
			{
				ResourceName:                         "directus_relation.test",
				ImportState:                          true,
				ImportStateId:                        articles + "/author_id",
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "field",
				ImportStateVerifyIgnore:              []string{"meta", "schema"},
			},
		},
	})
}

func testAccRelationConfig(articles, authors, onDelete string) string {
	return fmt.Sprintf(`
resource "directus_collection" "authors" {
  collection = %[2]q
}

resource "directus_collection" "articles" {
  collection = %[1]q
  # Serialize schema mutations: creating two collections concurrently can race
  # Directus's schema cache and 403 the dependent field create.
  depends_on = [directus_collection.authors]
}

resource "directus_field" "author_id" {
  collection = directus_collection.articles.collection
  field      = "author_id"
  type       = "integer"
}

resource "directus_relation" "test" {
  collection         = directus_collection.articles.collection
  field              = directus_field.author_id.field
  related_collection = directus_collection.authors.collection
  schema = {
    on_delete = %[3]q
  }
}
`, articles, authors, onDelete)
}
