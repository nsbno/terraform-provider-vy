package provider

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/nsbno/terraform-provider-central-cognito/internal/central_cognito"
	"testing"
)

func checkResourceServerDestroy(state *terraform.State) error {
	for _, resource_ := range state.RootModule().Resources {
		if resource_.Type != "vy_cognito_resource_server" {
			continue
		}

		err := resourceServerExists(resource_)
		if err != nil {
			return err
		}
	}

	return nil
}

func checkResourceServerExists(name string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		resource_, ok := state.RootModule().Resources[name]

		if !ok || resource_.Type != "vy-cognito_resource_server" {
			return fmt.Errorf("Resource Server '%s' not found", name)
		}

		return resourceServerExists(resource_)
	}
}

func resourceServerExists(resource_ *terraform.ResourceState) error {
	resource_server := central_cognito.ResourceServer{}
	err := testAccProvider.Client.ReadResourceServer(resource_.Primary.ID, &resource_server)
	if err != nil {
		return err
	}

	return nil
}

const testAccResourceServer_WithoutScopes = testAcc_ProviderConfig + `
resource "vy-cognito_resource_server" "test" {
	identifier = "basic.acceptancetest.io"
	name = "some service"
}
`

const testAccResourceServer_WithScopes = testAcc_ProviderConfig + `
resource "vy-cognito_resource_server" "test" {
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
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccResourceServer_WithoutScopes,
				Check: resource.ComposeTestCheckFunc(
					checkResourceServerExists("vy-cognito_resource_server.test"),
				),
			},
		},
	})
}

func TestAccResourceServer_WithScopes(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccResourceServer_WithScopes,
				Check: resource.ComposeTestCheckFunc(
					checkResourceServerExists("vy-cognito_resource_server.test"),
				),
			},
		},
	})
}
