package provider

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func testLambdaArtifactConfig(mockServerHost string) string {
	return fmt.Sprintf(`
provider "vy" {
	environment = "test"
	version_handler_v2_base_url = "%s"
}

data "vy_lambda_artifact" "this" {
	github_repository_name = "infrademo-demo-app"
}
`, mockServerHost)
}

func TestLambdaArtifact_Basic(t *testing.T) {
	// Create a mock HTTP server
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/versions/infrademo-demo-app/lambda" {
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(fmt.Sprintf("Lambda Artifact not found: %s", r.URL.Path)))
			return
		}

		// Return mock Lambda artifact data as JSON
		mockResponse := map[string]string{
			"github_repository_name": "infrademo-demo-app",
			"working_directory":      "",
			"git_sha":                "abc123",
			"branch":                 "main",
			"service_account_id":     "123456789012",
			"region":                 "eu-west-1",
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(mockResponse)
	}))
	defer mockServer.Close()

	// Extract host from URL (strip http://)
	mockServerHost := mockServer.URL[7:] // Remove "http://" prefix

	expectedResourceName := "data.vy_lambda_artifact.this"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testLambdaArtifactConfig(mockServerHost),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(expectedResourceName, "github_repository_name", "infrademo-demo-app"),
					resource.TestCheckResourceAttr(expectedResourceName, "git_sha", "abc123"),
					resource.TestCheckResourceAttr(expectedResourceName, "branch", "main"),
					resource.TestCheckResourceAttr(expectedResourceName, "service_account_id", "123456789012"),
					resource.TestCheckResourceAttr(expectedResourceName, "region", "eu-west-1"),
				),
			},
		},
	})
}

func testLambdaArtifactConfigWithWorkingDirectory(mockServerHost string) string {
	return fmt.Sprintf(`
provider "vy" {
	environment = "test"
	version_handler_v2_base_url = "%s"
}

data "vy_lambda_artifact" "this" {
	github_repository_name = "infrademo-demo-app"
	working_directory = "services/lambda-function"
}
`, mockServerHost)
}

func TestLambdaArtifact_WithWorkingDirectory(t *testing.T) {
	// Create a mock HTTP server
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/versions/infrademo-demo-app/lambda/lambda-function" {
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(fmt.Sprintf("Lambda Artifact not found: %s", r.URL.Path)))
			return
		}

		// Return mock Lambda artifact data as JSON
		mockResponse := map[string]string{
			"github_repository_name": "infrademo-demo-app",
			"working_directory":      "lambda-function",
			"git_sha":                "def456",
			"branch":                 "main",
			"service_account_id":     "123456789012",
			"region":                 "eu-west-1",
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(mockResponse)
	}))
	defer mockServer.Close()

	// Extract host from URL (strip http://)
	mockServerHost := mockServer.URL[7:] // Remove "http://" prefix

	expectedResourceName := "data.vy_lambda_artifact.this"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testLambdaArtifactConfigWithWorkingDirectory(mockServerHost),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(expectedResourceName, "github_repository_name", "infrademo-demo-app"),
					resource.TestCheckResourceAttr(expectedResourceName, "working_directory", "lambda-function"),
					resource.TestCheckResourceAttr(expectedResourceName, "git_sha", "def456"),
					resource.TestCheckResourceAttr(expectedResourceName, "branch", "main"),
					resource.TestCheckResourceAttr(expectedResourceName, "service_account_id", "123456789012"),
					resource.TestCheckResourceAttr(expectedResourceName, "region", "eu-west-1"),
				),
			},
		},
	})
}
