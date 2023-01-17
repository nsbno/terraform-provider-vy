package provider

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"testing"
)

const testAccCognitoInfo = testAcc_ProviderConfig + `
data "vy_cognito_info" "this" {

}
`

func TestAccCognitoInfo(t *testing.T) {
	expected_resource_name := "data.vy_cognito_info.this"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccCognitoInfo,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(expected_resource_name, "auth_url"),
					resource.TestCheckResourceAttrSet(expected_resource_name, "jwks_url"),
					resource.TestCheckResourceAttrSet(expected_resource_name, "open_id_url"),
				),
			},
		},
	})
}
