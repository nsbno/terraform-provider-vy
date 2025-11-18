package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

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

const testAccAppClient_FrontendAddedScope = testAcc_ProviderConfig + testAccAppClient_ResourceServer + `
resource "vy_app_client" "frontend" {
	name = "app_client_frontend.acceptancetest.io"
	type = "frontend"
	scopes = [
		"${vy_resource_server.test.identifier}/read",
		"${vy_resource_server.test.identifier}/modify",
	]
	callback_urls = ["https://example.com/callback"]
	logout_urls = ["https://example.com/logout"]
}
`

func TestAccAppClient_Frontend(t *testing.T) {
	expected_resource_name := "vy_app_client.frontend"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccAppClient_Frontend,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(expected_resource_name, "client_id"),
					resource.TestCheckNoResourceAttr(expected_resource_name, "client_secret"),
					resource.TestCheckResourceAttr(expected_resource_name, "generate_secret", "false"),
				),
			},
			{
				Config: testAccAppClient_FrontendAddedScope,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(expected_resource_name, "client_id"),
					resource.TestCheckNoResourceAttr(expected_resource_name, "client_secret"),
					resource.TestCheckResourceAttr(expected_resource_name, "generate_secret", "false"),
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

const testAccAppClient_BackendRemoveScope = testAcc_ProviderConfig + testAccAppClient_ResourceServer + `
resource "vy_app_client" "backend" {
	name = "app_client_backend.acceptancetest.io"
	type = "backend"
	scopes = [
		"${vy_resource_server.test.identifier}/read",
	]
}
`

func TestAccAppClient_Backend(t *testing.T) {
	expected_resource_name := "vy_app_client.backend"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccAppClient_Backend,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(expected_resource_name, "client_id"),
					resource.TestCheckResourceAttrSet(expected_resource_name, "client_secret"),
					resource.TestCheckResourceAttr(expected_resource_name, "generate_secret", "true"),
				),
			},
			{
				Config: testAccAppClient_BackendRemoveScope,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(expected_resource_name, "client_id"),
					resource.TestCheckResourceAttrSet(expected_resource_name, "client_secret"),
					resource.TestCheckResourceAttr(expected_resource_name, "generate_secret", "true"),
				),
			},
		},
	})
}

const testAccAppClient_Complex = testAcc_ProviderConfig + testAccAppClient_ResourceServer + `
resource "vy_app_client" "complex" {
  name = "app_client_complex.acceptancetest.io"

  type = "frontend"
  generate_secret = true

  callback_urls   = [
    "http://localhost:3000/auth/callback",
    "https://example.com/auth/callback"
  ]
  logout_urls = [
    "http://localhost:3000/logout",
    "https://example.com/logout"
  ]

  scopes = [
    "email",
    "openid",
    "phone",
    "profile"
  ]
}
`

func TestAccAppClient_Complex(t *testing.T) {
	expected_resource_name := "vy_app_client.complex"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccAppClient_Complex,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(expected_resource_name, "client_id"),
					resource.TestCheckResourceAttrSet(expected_resource_name, "client_secret"),
					resource.TestCheckResourceAttr(expected_resource_name, "generate_secret", "true"),
				),
			},
		},
	})
}
