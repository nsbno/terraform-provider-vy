package provider

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"testing"
)

const testAccEnvironmentAccount = testAcc_ProviderConfig + `
resource "vy_environment_account" "test" {
	owner_account_id = "12345678901"
}
`

func TestAccEnvironmentAccount(t *testing.T) {
	expected_resource_name := "vy_environment_account.test"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccEnvironmentAccount,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(expected_resource_name, "owner_account_id"),
				),
			},
		},
	})
}
