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
			"s3_object_path":         "123456789012/lambda-artifacts/infrademo-demo-app/abc123/lambda.zip",
			"s3_object_version":      "abc123",
			"bucket_name":            "123456789012-deployment-delivery-artifacts",
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
					resource.TestCheckResourceAttr(expectedResourceName, "s3_object_path", "123456789012/lambda-artifacts/infrademo-demo-app/abc123/lambda.zip"),
					resource.TestCheckResourceAttr(expectedResourceName, "s3_object_version", "abc123"),
					resource.TestCheckResourceAttr(expectedResourceName, "s3_bucket_name", "123456789012-deployment-delivery-artifacts"),
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
		if r.URL.Path != "/v2/versions/infrademo-demo-app/lambda" {
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(fmt.Sprintf("Lambda Artifact not found: %s", r.URL.Path)))
			return
		}

		// Check for working_directory query parameter
		workingDir := r.URL.Query().Get("working_directory")
		if workingDir != "services/lambda-function" {
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(fmt.Sprintf("Lambda Artifact not found: working_directory=%s", workingDir)))
			return
		}

		// Return mock Lambda artifact data as JSON
		mockResponse := map[string]string{
			"github_repository_name": "infrademo-demo-app",
			"working_directory":      "services/lambda-function",
			"git_sha":                "def456",
			"branch":                 "main",
			"service_account_id":     "123456789012",
			"region":                 "eu-west-1",
			"s3_object_path":         "123456789012/lambda-artifacts/infrademo-demo-app/abc123/lambda.zip",
			"s3_object_version":      "abc123",
			"bucket_name":            "123456789012-deployment-delivery-artifacts",
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
					resource.TestCheckResourceAttr(expectedResourceName, "working_directory", "services/lambda-function"),
					resource.TestCheckResourceAttr(expectedResourceName, "git_sha", "def456"),
					resource.TestCheckResourceAttr(expectedResourceName, "branch", "main"),
					resource.TestCheckResourceAttr(expectedResourceName, "service_account_id", "123456789012"),
					resource.TestCheckResourceAttr(expectedResourceName, "region", "eu-west-1"),
					resource.TestCheckResourceAttr(expectedResourceName, "s3_object_path", "123456789012/lambda-artifacts/infrademo-demo-app/abc123/lambda.zip"),
					resource.TestCheckResourceAttr(expectedResourceName, "s3_object_version", "abc123"),
					resource.TestCheckResourceAttr(expectedResourceName, "s3_bucket_name", "123456789012-deployment-delivery-artifacts"),
				),
			},
		},
	})
}

func testLambdaArtifactWithECRConfig(mockServerHost string) string {
	return fmt.Sprintf(`
provider "vy" {
	environment = "test"
	version_handler_v2_base_url = "%s"
}

data "vy_lambda_artifact" "this" {
	github_repository_name = "infrademo-demo-app"
	ecr_repository_name    = "petstore-ecr"
}
`, mockServerHost)
}

func TestLambdaArtifact_ECR(t *testing.T) {
	// Create a mock HTTP server
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/versions/infrademo-demo-app/lambda" {
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(fmt.Sprintf("Lambda Artifact not found: %s", r.URL.Path)))
			return
		}

		// Check for ecr_repository_name query parameter
		ecrRepositoryName := r.URL.Query().Get("ecr_repository_name")
		if ecrRepositoryName != "petstore-ecr" {
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(fmt.Sprintf("Lambda Artifact not found: ecr_repository_name=%s", ecrRepositoryName)))
			return
		}

		// Return mock Lambda artifact data as JSON
		mockResponse := map[string]string{
			"github_repository_name": "infrademo-demo-app",
			"working_directory":      "",
			"git_sha":                "abc123",
			"branch":                 "main",
			"service_account_id":     "123456789012",
			"ecr_repository_uri":     "123456789012.dkr.ecr.eu-west-1.amazonaws.com/petstore-ecr",
			"region":                 "eu-west-1",
			"s3_object_path":         "",
			"s3_object_version":      "",
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
				Config: testLambdaArtifactWithECRConfig(mockServerHost),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(expectedResourceName, "github_repository_name", "infrademo-demo-app"),
					resource.TestCheckResourceAttr(expectedResourceName, "git_sha", "abc123"),
					resource.TestCheckResourceAttr(expectedResourceName, "branch", "main"),
					resource.TestCheckResourceAttr(expectedResourceName, "service_account_id", "123456789012"),
					resource.TestCheckResourceAttr(expectedResourceName, "ecr_repository_uri", "123456789012.dkr.ecr.eu-west-1.amazonaws.com/petstore-ecr"),
					resource.TestCheckResourceAttr(expectedResourceName, "region", "eu-west-1"),
					resource.TestCheckResourceAttr(expectedResourceName, "s3_object_path", ""),
					resource.TestCheckResourceAttr(expectedResourceName, "s3_object_version", ""),
				),
			},
		},
	})
}
