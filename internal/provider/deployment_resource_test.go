package provider

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"testing"
)

const testAccDeploymentAccount = testAcc_ProviderConfig + `
resource "vy_deployment_account" "test" {
	slack_channel = "CMN2KHQL8"
}
`

func TestAccDeploymentAccount(t *testing.T) {
	expected_resource_name := "vy_deployment_account.test"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccDeploymentAccount,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(expected_resource_name, "slack_channel"),
				),
			},
		},
	})
}
