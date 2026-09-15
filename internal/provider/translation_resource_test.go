// Copyright IBM Corp. 2026

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccTranslationResource(t *testing.T) {
	key := fmt.Sprintf("tf_acc_key_%d", acctestSuffix())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccTranslationConfig("en-US", key, "Hello"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("directus_translation.test", "language", "en-US"),
					resource.TestCheckResourceAttr("directus_translation.test", "key", key),
					resource.TestCheckResourceAttr("directus_translation.test", "value", "Hello"),
					resource.TestCheckResourceAttrSet("directus_translation.test", "id"),
				),
			},
			{
				ResourceName:      "directus_translation.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccTranslationConfig("en-US", key, "Hello there"),
				Check:  resource.TestCheckResourceAttr("directus_translation.test", "value", "Hello there"),
			},
		},
	})
}

func testAccTranslationConfig(language, key, value string) string {
	return fmt.Sprintf(`
resource "directus_translation" "test" {
  language = %[1]q
  key      = %[2]q
  value    = %[3]q
}
`, language, key, value)
}
