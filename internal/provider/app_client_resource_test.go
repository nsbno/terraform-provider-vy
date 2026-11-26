package provider

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/path"
	fwresource "github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/nsbno/terraform-provider-vy/internal/central_cognito"
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

func TestAppClientToDomain_SendsEmptyScopesWhenScopesAreNull(t *testing.T) {
	model := AppClientResourceModel{
		Name: types.StringValue("my-app"),
		Type: types.StringValue("backend"),
	}

	var domain central_cognito.AppClient
	model.toDomain(&domain)

	if domain.Scopes == nil {
		t.Fatalf("expected Scopes to be an empty list, got nil")
	}

	body, err := json.Marshal(domain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(string(body), `"scopes":[]`) {
		t.Errorf("expected scopes to be sent as an empty list, got: %s", body)
	}
}

// appClientStateList stores the model in a Terraform state and returns the list attribute at name.
func appClientStateList(t *testing.T, model AppClientResourceModel, name string) types.List {
	t.Helper()
	ctx := context.Background()

	var schemaResponse fwresource.SchemaResponse
	(&AppClientResource{}).Schema(ctx, fwresource.SchemaRequest{}, &schemaResponse)

	state := tfsdk.State{
		Schema: schemaResponse.Schema,
		Raw:    tftypes.NewValue(schemaResponse.Schema.Type().TerraformType(ctx), nil),
	}
	if diags := state.Set(ctx, &model); diags.HasError() {
		t.Fatalf("unexpected error setting state: %v", diags)
	}

	var list types.List
	if diags := state.GetAttribute(ctx, path.Root(name), &list); diags.HasError() {
		t.Fatalf("unexpected error reading %s from state: %v", name, diags)
	}

	return list
}

func TestAppClientResourceDataFromDomain_EmptyUrlListsRoundTrip(t *testing.T) {
	testCases := []struct {
		name         string
		configUrls   []string
		expectedNull bool
	}{
		{name: "empty list in config stays an empty list", configUrls: []string{}, expectedNull: false},
		{name: "unset in config stays null", configUrls: nil, expectedNull: true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			config := AppClientResourceModel{
				Name:         types.StringValue("my-app"),
				Type:         types.StringValue("frontend"),
				Scopes:       []string{"openid"},
				CallbackUrls: tc.configUrls,
				LogoutUrls:   tc.configUrls,
			}

			var request central_cognito.AppClient
			config.toDomain(&request)

			// The remote answers with empty lists for both an empty list and an unset attribute.
			remote := request
			remote.CallbackUrls = []string{}
			remote.LogoutUrls = []string{}

			var created AppClientResourceModel
			appClientResourceDataFromDomain(remote, config, &created)

			var refreshed AppClientResourceModel
			appClientResourceDataFromDomain(remote, created, &refreshed)

			for step, model := range map[string]AppClientResourceModel{"create": created, "read": refreshed} {
				for _, attribute := range []string{"callback_urls", "logout_urls"} {
					list := appClientStateList(t, model, attribute)
					if list.IsNull() != tc.expectedNull {
						t.Errorf("%s: expected %s null=%v, got %s", step, attribute, tc.expectedNull, list)
					}
					if !list.IsNull() && len(list.Elements()) != 0 {
						t.Errorf("%s: expected %s to be empty, got %s", step, attribute, list)
					}
				}
			}
		})
	}
}

func TestAppClientResourceDataFromDomain_ImportGivesNullForEmptyUrlLists(t *testing.T) {
	remote := central_cognito.AppClient{
		Name:         "my-app",
		Type:         "frontend",
		Scopes:       []string{"openid"},
		CallbackUrls: []string{},
		LogoutUrls:   []string{},
	}

	var imported AppClientResourceModel
	appClientResourceDataFromDomain(remote, AppClientResourceModel{}, &imported)

	if imported.CallbackUrls != nil {
		t.Errorf("expected callback_urls to be null, got %v", imported.CallbackUrls)
	}
	if imported.LogoutUrls != nil {
		t.Errorf("expected logout_urls to be null, got %v", imported.LogoutUrls)
	}
}

func TestAppClientResourceDataFromDomain_UsesRemoteUrlsWhenSet(t *testing.T) {
	remote := central_cognito.AppClient{
		Name:         "my-app",
		Type:         "frontend",
		CallbackUrls: []string{"https://example.com/callback"},
		LogoutUrls:   []string{"https://example.com/logout"},
	}

	var state AppClientResourceModel
	appClientResourceDataFromDomain(remote, AppClientResourceModel{LogoutUrls: []string{}}, &state)

	if len(state.CallbackUrls) != 1 || state.CallbackUrls[0] != "https://example.com/callback" {
		t.Errorf("expected callback_urls from remote, got %v", state.CallbackUrls)
	}
	if len(state.LogoutUrls) != 1 || state.LogoutUrls[0] != "https://example.com/logout" {
		t.Errorf("expected logout_urls from remote, got %v", state.LogoutUrls)
	}
}
