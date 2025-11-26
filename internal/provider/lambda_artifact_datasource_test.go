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

func testLambdaArtifactConfigWithWorkingDirectory(mockServerHost string) string {
	return fmt.Sprintf(`
provider "vy" {
	environment = "test"
	version_handler_v2_base_url = "%s"
}

data "vy_lambda_artifact" "this" {
	github_repository_name = "infrademo-demo-app"
	working_directory = "lambda-function"
}
`, mockServerHost)
}

func TestLambdaArtifact_Basic(t *testing.T) {
	// Create a mock HTTP server
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/lambda/versions/infrademo-demo-app" {
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(fmt.Sprintf("Lambda Artifact not found: %s", r.URL.Path)))
			return
		}

		// Return mock S3 version data as JSON
		mockResponse := map[string]string{
			"uri":     "s3://123456789012-deployment-delivery-pipeline/infrademo-demo-app/abc123.zip",
			"store":   "s3",
			"path":    "latest",
			"version": "abc123def456",
			"git_sha": "abc123",
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
					resource.TestCheckResourceAttr(expectedResourceName, "uri", "s3://123456789012-deployment-delivery-pipeline/infrademo-demo-app/abc123.zip"),
					resource.TestCheckResourceAttr(expectedResourceName, "store", "s3"),
					resource.TestCheckResourceAttr(expectedResourceName, "path", "latest"),
					resource.TestCheckResourceAttr(expectedResourceName, "version", "abc123def456"),
					resource.TestCheckResourceAttr(expectedResourceName, "git_sha", "abc123"),
				),
			},
		},
	})
}

func TestLambdaArtifact_WithWorkingDirectory(t *testing.T) {
	// Create a mock HTTP server
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/lambda/versions/infrademo-demo-app/lambda-function" {
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(fmt.Sprintf("Lambda Artifact not found: %s", r.URL.Path)))
			return
		}

		// Return mock S3 version data as JSON
		mockResponse := map[string]string{
			"uri":     "s3://123456789012-deployment-delivery-pipeline/infrademo-demo-app/lambda-function/def456.zip",
			"store":   "s3",
			"path":    "lambda-function/latest",
			"version": "def456abc123",
			"git_sha": "def456",
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
					resource.TestCheckResourceAttr(expectedResourceName, "uri", "s3://123456789012-deployment-delivery-pipeline/infrademo-demo-app/lambda-function/def456.zip"),
					resource.TestCheckResourceAttr(expectedResourceName, "store", "s3"),
					resource.TestCheckResourceAttr(expectedResourceName, "path", "lambda-function/latest"),
					resource.TestCheckResourceAttr(expectedResourceName, "version", "def456abc123"),
					resource.TestCheckResourceAttr(expectedResourceName, "git_sha", "def456"),
				),
			},
		},
	})
}
