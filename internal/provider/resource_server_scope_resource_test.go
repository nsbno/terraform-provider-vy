package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

const testAccResourceServerScope_ResourceServer = `
resource "vy_resource_server" "test" {
	identifier = "for-scope-basic.acceptancetest.io"
	name = "some service"
}
`

const testAccResourceServerScope_Basic = testAcc_ProviderConfig + testAccResourceServerScope_ResourceServer + `
data "vy_resource_server" "test" {
	identifier = vy_resource_server.test.identifier
}

resource "vy_resource_server_scope" "read" {
	resource_server = data.vy_resource_server.test.identifier

	namespace   = "fillrate"
	name        = "read"
	description = "Used for reading"
}
`

const testAccResourceServerScope_DescriptionUpdated = testAcc_ProviderConfig + testAccResourceServerScope_ResourceServer + `
data "vy_resource_server" "test" {
	identifier = vy_resource_server.test.identifier
}

resource "vy_resource_server_scope" "read" {
	resource_server = data.vy_resource_server.test.identifier

	namespace   = "fillrate"
	name        = "read"
	description = "Updated description"
}
`

func TestAccResourceServerScope_Basic(t *testing.T) {
	expected_resource_name := "vy_resource_server_scope.read"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccResourceServerScope_Basic,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(expected_resource_name, "namespace", "fillrate"),
					resource.TestCheckResourceAttr(expected_resource_name, "name", "read"),
					resource.TestCheckResourceAttr(expected_resource_name, "description", "Used for reading"),
					resource.TestCheckResourceAttr("data.vy_resource_server.test", "scopes.#", "1"),
				),
			},
			{
				Config: testAccResourceServerScope_DescriptionUpdated,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(expected_resource_name, "description", "Updated description"),
				),
			},
		},
	})
}
