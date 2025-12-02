package provider

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func testECSImageConfig(mockServerHost string) string {
	return fmt.Sprintf(`
provider "vy" {
	environment = "test"
	version_handler_v2_base_url = "%s"
}

data "vy_ecs_image" "this" {
	github_repository_name = "my-repo"
}
`, mockServerHost)
}

func TestECSImage_Basic(t *testing.T) {
	// Create a mock HTTP server
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/versions/my-repo/ecs" {
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(fmt.Sprintf("ECS image not found: %s", r.URL.Path)))
			return
		}

		// Return mock ECS version data as JSON
		mockResponse := map[string]string{
			"github_repository_name": "my-repo",
			"working_directory":      "",
			"git_sha":                "abc123",
			"branch":                 "main",
			"service_account_id":     "123456789012",
			"region":                 "eu-west-1",
			"ecr_repository_name":    "petstore-repo",
			"ecr_repository_uri":     "123456789012.dkr.ecr.eu-west-1.amazonaws.com/petstore-repo",
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
				Config: testECSImageConfig(mockServerHost),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(expectedResourceName, "github_repository_name", "my-repo"),
					resource.TestCheckResourceAttr(expectedResourceName, "working_directory", ""),
					resource.TestCheckResourceAttr(expectedResourceName, "git_sha", "abc123"),
					resource.TestCheckResourceAttr(expectedResourceName, "branch", "main"),
					resource.TestCheckResourceAttr(expectedResourceName, "service_account_id", "123456789012"),
					resource.TestCheckResourceAttr(expectedResourceName, "region", "eu-west-1"),
					resource.TestCheckResourceAttr(expectedResourceName, "ecr_repository_name", "petstore-repo"),
					resource.TestCheckResourceAttr(expectedResourceName, "ecr_repository_uri", "123456789012.dkr.ecr.eu-west-1.amazonaws.com/petstore-repo"),
				),
			},
		},
	})
}

func testECSImageConfigWithWorkingDirectory(mockServerHost string) string {
	return fmt.Sprintf(`
provider "vy" {
	environment = "test"
	version_handler_v2_base_url = "%s"
}

data "vy_ecs_image" "this" {
	github_repository_name = "my-repo"
	working_directory	   = "services/user-auth"
}
`, mockServerHost)
}

func TestECSImage_WithWorkingDirectory(t *testing.T) {
	// Create a mock HTTP server
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/versions/my-repo/ecs" {
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(fmt.Sprintf("ECS image not found: %s", r.URL.Path)))
			return
		}

		// Check for working_directory query parameter
		workingDir := r.URL.Query().Get("working_directory")
		if workingDir != "services/user-auth" {
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(fmt.Sprintf("ECS image not found: working_directory=%s", workingDir)))
			return
		}

		// Return mock ECS version data as JSON
		mockResponse := map[string]string{
			"github_repository_name": "my-repo",
			"working_directory":      "services/user-auth",
			"git_sha":                "abc123",
			"branch":                 "main",
			"service_account_id":     "123456789012",
			"region":                 "eu-west-1",
			"ecr_repository_name":    "petstore-repo",
			"ecr_repository_uri":     "123456789012.dkr.ecr.eu-west-1.amazonaws.com/petstore-repo",
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
				Config: testECSImageConfigWithWorkingDirectory(mockServerHost),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(expectedResourceName, "github_repository_name", "my-repo"),
					resource.TestCheckResourceAttr(expectedResourceName, "working_directory", "services/user-auth"),
					resource.TestCheckResourceAttr(expectedResourceName, "git_sha", "abc123"),
					resource.TestCheckResourceAttr(expectedResourceName, "branch", "main"),
					resource.TestCheckResourceAttr(expectedResourceName, "service_account_id", "123456789012"),
					resource.TestCheckResourceAttr(expectedResourceName, "region", "eu-west-1"),
					resource.TestCheckResourceAttr(expectedResourceName, "ecr_repository_name", "petstore-repo"),
					resource.TestCheckResourceAttr(expectedResourceName, "ecr_repository_uri", "123456789012.dkr.ecr.eu-west-1.amazonaws.com/petstore-repo"),
				),
			},
		},
	})
}

func testECSImageConfigWithECRRepoOverride(mockServerHost string) string {
	return fmt.Sprintf(`
provider "vy" {
	environment = "test"
	version_handler_v2_base_url = "%s"
}

data "vy_ecs_image" "this" {
	github_repository_name = "my-repo"
	
	ecr_repository_name = "my-override-repo"
}
`, mockServerHost)
}

func TestECSImage_WithECRRepoOverride(t *testing.T) {
	// Create a mock HTTP server
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/versions/my-repo/ecs" {
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(fmt.Sprintf("ECS image not found: %s", r.URL.Path)))
			return
		}

		// Return mock ECS version data as JSON
		mockResponse := map[string]string{
			"github_repository_name": "my-repo",
			"working_directory":      "",
			"git_sha":                "abc123",
			"branch":                 "main",
			"service_account_id":     "123456789012",
			"region":                 "eu-west-1",
			"ecr_repository_name":    "petstore-repo",
			"ecr_repository_uri":     "123456789012.dkr.ecr.eu-west-1.amazonaws.com/petstore-repo",
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
				Config: testECSImageConfigWithECRRepoOverride(mockServerHost),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(expectedResourceName, "github_repository_name", "my-repo"),
					resource.TestCheckResourceAttr(expectedResourceName, "working_directory", ""),
					resource.TestCheckResourceAttr(expectedResourceName, "git_sha", "abc123"),
					resource.TestCheckResourceAttr(expectedResourceName, "branch", "main"),
					resource.TestCheckResourceAttr(expectedResourceName, "service_account_id", "123456789012"),
					resource.TestCheckResourceAttr(expectedResourceName, "region", "eu-west-1"),
					resource.TestCheckResourceAttr(expectedResourceName, "ecr_repository_name", "my-override-repo"),
					resource.TestCheckResourceAttr(expectedResourceName, "ecr_repository_uri", "123456789012.dkr.ecr.eu-west-1.amazonaws.com/my-override-repo"),
				),
			},
		},
	})
}
