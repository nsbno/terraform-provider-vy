package provider

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"testing"
)

func checkCognitoInfoExists(name string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		resource_, ok := state.RootModule().Resources[name]

		if !ok || resource_.Type != "vy_cognito_info" {
			return fmt.Errorf("Cognito Info '%s' not found", name)
		}

		return nil
	}
}

const testAccCognitoInfo = testAcc_ProviderConfig + `
data "vy_cognito_info" "this" {

}
`

func TestAccCognitoInfo(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccCognitoInfo,
				Check: resource.ComposeTestCheckFunc(
					checkCognitoInfoExists("data.vy_cognito_info.this"),
				),
			},
		},
	})
}
