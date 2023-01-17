package provider

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"testing"
)

const testAccResourceServer_WithoutScopes = testAcc_ProviderConfig + `
resource "vy_resource_server" "test" {
	identifier = "basic.acceptancetest.io"
	name = "some service"
}
`

const testAccResourceServer_WithScopes = testAcc_ProviderConfig + `
resource "vy_resource_server" "test" {
	identifier = "withscopes.acceptancetest.io"
	name = "some service"

	scopes = [
		{
			name = "read"
			description = "Allows for reading of stuff"	
		},
		{
			name = "modify"
			description = "Modify stuff"	
		}
	]
}
`

func TestAccResourceServer_Basic(t *testing.T) {
	expected_resource_name := "vy_resource_server.test"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccResourceServer_WithoutScopes,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(expected_resource_name, "scopes.#", "0"),
				),
			},
		},
	})
}

func TestAccResourceServer_WithScopes(t *testing.T) {
	expected_resource_name := "vy_resource_server.test"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccResourceServer_WithScopes,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(expected_resource_name, "scopes.#", "2"),
				),
			},
		},
	})
}
