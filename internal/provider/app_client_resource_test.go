package provider

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/nsbno/terraform-provider-vy/internal/central_cognito"
	"testing"
)

func checkAppClientExists(name string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		resource_, ok := state.RootModule().Resources[name]

		if !ok || resource_.Type != "vy_app_client" {
			return fmt.Errorf("Resource Server '%s' not found", name)
		}

		return appClientExists(resource_)
	}
}

func appClientExists(resource_ *terraform.ResourceState) error {
	app_client := central_cognito.AppClient{}
	err := testAccProvider.Client.ReadAppClient(resource_.Primary.ID, &app_client)
	if err != nil {
		return err
	}

	return nil
}

const testAccAppClient_ResourceServer = `
resource "vy_resource_server" "test" {
	identifier = "for-app-client-basic.acceptancetest.io"
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

const testAccAppClient_Frontend = testAcc_ProviderConfig + testAccAppClient_ResourceServer + `
resource "vy_app_client" "frontend" {
	name = "app_client_frontend.acceptancetest.io"
	type = "frontend"
	scopes = [
		"${vy_resource_server.test.identifier}/read"
	]
	callback_urls = ["https://example.com/callback"]
	logout_urls = ["https://example.com/logout"]
}
`

func TestAccAppClient_Frontend(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccAppClient_Frontend,
				Check: resource.ComposeTestCheckFunc(
					checkAppClientExists("vy_app_client.frontend"),
				),
			},
		},
	})
}

const testAccAppClient_Backend = testAcc_ProviderConfig + testAccAppClient_ResourceServer + `
resource "vy_app_client" "backend" {
	name = "app_client_backend.acceptancetest.io"
	type = "backend"
	scopes = [
		"${vy_resource_server.test.identifier}/read",
		"${vy_resource_server.test.identifier}/modify",
	]
}
`

func TestAccAppClient_Backend(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccAppClient_Backend,
				Check: resource.ComposeTestCheckFunc(
					checkAppClientExists("vy_app_client.backend"),
				),
			},
		},
	})
}
