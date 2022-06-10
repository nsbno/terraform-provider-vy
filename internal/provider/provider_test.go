package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"testing"
)

var testAccProvider, _ = convertProviderType(New("test")())

const testAcc_ProviderConfig = `
provider "vy-cognito" {
	environment = "tm9ru6l46e"
	endpoint = "execute-api.eu-west-1.amazonaws.com/main"
}

`

// testAccProtoV6ProviderFactories are used to instantiate a provider during
// acceptance testing. The factory function will be invoked for every Terraform
// CLI command executed to create a provider server to which the CLI can
// reattach.
var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"vy-cognito": providerserver.NewProtocol6WithError(&testAccProvider),
}

func testAccPreCheck(t *testing.T) {
	// You can add code here to run prior to any test case execution, for example assertions
	// about the appropriate environment variables being set are common to see in a pre-check
	// function.

}
