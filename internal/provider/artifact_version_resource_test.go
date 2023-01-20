package provider

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"testing"
)

const testAccArtifactVersion = testAcc_ProviderConfig + `
data "vy_artifact_version" "this" {
	application = "petstore-webapp"
}
`

func TestAccArtifactVersion_Basic(t *testing.T) {
	expected_resource_name := "data.vy_artifact_version.this"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },

		Steps: []resource.TestStep{
			{
				Config: testAccArtifactVersion,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(expected_resource_name, "application", "petstore-webapp"),
					resource.TestCheckResourceAttrSet(expected_resource_name, "uri"),
					resource.TestCheckResourceAttrSet(expected_resource_name, "store"),
					resource.TestCheckResourceAttrSet(expected_resource_name, "path"),
					resource.TestCheckResourceAttrSet(expected_resource_name, "version"),
				),
			},
		},
	})
}
