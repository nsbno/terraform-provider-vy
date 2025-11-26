package provider

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func testAccECSImageConfig(mockServerHost string) string {
	return fmt.Sprintf(`
provider "vy" {
	environment = "test"
	version_handler_v2_base_url = "%s"
}

data "vy_ecs_image" "this" {
	ecr_repository_name = "my-service"
}
`, mockServerHost)
}

func TestAccECSImage_Basic(t *testing.T) {
	// Create a mock HTTP server
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/ecs/versions/my-service" {
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(fmt.Sprintf("ECS repository not found: %s", r.URL.Path)))
			return
		}

		// Return mock ECR version data as JSON
		mockResponse := map[string]string{
			"ecr_repository_name": "my-service",
			"uri":                 "123456789012.dkr.ecr.eu-west-1.amazonaws.com/my-service:latest",
			"store":               "ecr",
			"path":                "latest",
			"version":             "sha256:abc123def456",
			"git_sha":             "abc123",
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(mockResponse)
	}))
	defer mockServer.Close()

	// Extract host from URL (strip http://)
	mockServerHost := mockServer.URL[7:] // Remove "http://" prefix

	expectedResourceName := "data.vy_ecs_image.this"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccECSImageConfig(mockServerHost),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(expectedResourceName, "ecr_repository_name", "my-service"),
					resource.TestCheckResourceAttr(expectedResourceName, "uri", "123456789012.dkr.ecr.eu-west-1.amazonaws.com/my-service:latest"),
					resource.TestCheckResourceAttr(expectedResourceName, "store", "ecr"),
					resource.TestCheckResourceAttr(expectedResourceName, "path", "latest"),
					resource.TestCheckResourceAttr(expectedResourceName, "version", "sha256:abc123def456"),
					resource.TestCheckResourceAttr(expectedResourceName, "git_sha", "abc123"),
				),
			},
		},
	})
}
