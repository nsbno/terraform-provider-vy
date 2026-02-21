package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

const testAccDeploymentAccount = testAcc_ProviderConfig + `
resource "vy_deployment_account" "test" {
	slack_channel = "CMN2KHQL8"
}
`

// It's impossible for this test to work stably
// as it expects that any of the AWS accounts you have assumed during
// the test run, isn't already registered in enroll-accounts.
// But that only holds true at-most once
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
