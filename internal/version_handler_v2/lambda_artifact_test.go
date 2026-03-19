package version_handler_v2

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestReadLambdaArtifact_ReturnsArtifactForMatchingRepository(t *testing.T) {
	api := &FakeVersionHandlerAPI{
		KnownLambdaArtifacts: []LambdaArtifact{
			{
				GitHubRepositoryName: "nsbno/my-service",
				S3ObjectPath:         "artifacts/lambda.zip",
				S3ObjectVersion:      "v42",
				S3BucketName:         "deploy-artifacts",
			},
		},
	}
	server, client := api.Start()
	defer server.Close()

	var artifact LambdaArtifact
	err := client.ReadLambdaArtifact("nsbno/my-service", "", "", "", &artifact)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if artifact.S3ObjectPath != "artifacts/lambda.zip" {
		t.Errorf("S3ObjectPath = %q, want %q", artifact.S3ObjectPath, "artifacts/lambda.zip")
	}
	if artifact.S3BucketName != "deploy-artifacts" {
		t.Errorf("S3BucketName = %q, want %q", artifact.S3BucketName, "deploy-artifacts")
	}
}

func TestReadLambdaArtifact_DistinguishesMonorepoServicesByWorkingDirectory(t *testing.T) {
	api := &FakeVersionHandlerAPI{
		KnownLambdaArtifacts: []LambdaArtifact{
			{
				GitHubRepositoryName: "nsbno/monorepo",
				WorkingDirectory:     "services/billing",
				S3ObjectPath:         "billing.zip",
			},
			{
				GitHubRepositoryName: "nsbno/monorepo",
				WorkingDirectory:     "services/notifications",
				S3ObjectPath:         "notifications.zip",
			},
		},
	}
	server, client := api.Start()
	defer server.Close()

	var artifact LambdaArtifact
	err := client.ReadLambdaArtifact("nsbno/monorepo", "", "services/notifications", "", &artifact)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if artifact.S3ObjectPath != "notifications.zip" {
		t.Errorf("S3ObjectPath = %q, want %q", artifact.S3ObjectPath, "notifications.zip")
	}
}

func TestReadLambdaArtifact_FiltersArtifactByPath(t *testing.T) {
	api := &FakeVersionHandlerAPI{
		KnownLambdaArtifacts: []LambdaArtifact{
			{
				GitHubRepositoryName: "nsbno/my-service",
				Path:                 "functions/authorizer",
				S3ObjectPath:         "authorizer.zip",
			},
		},
	}
	server, client := api.Start()
	defer server.Close()

	var artifact LambdaArtifact
	err := client.ReadLambdaArtifact("nsbno/my-service", "", "", "functions/authorizer", &artifact)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if artifact.S3ObjectPath != "authorizer.zip" {
		t.Errorf("S3ObjectPath = %q, want %q", artifact.S3ObjectPath, "authorizer.zip")
	}
}

func TestReadLambdaArtifact_ReturnsErrorWhenArtifactDoesNotExist(t *testing.T) {
	api := &FakeVersionHandlerAPI{} // no known artifacts
	server, client := api.Start()
	defer server.Close()

	var artifact LambdaArtifact
	err := client.ReadLambdaArtifact("nsbno/nonexistent", "", "", "", &artifact)
	if err == nil {
		t.Fatal("expected an error for a missing artifact, got nil")
	}
	if !strings.Contains(err.Error(), "404") {
		t.Errorf("error = %q, want it to contain HTTP status '404'", err.Error())
	}
}

func TestReadLambdaArtifact_PropagatesServerErrorWithStatusAndMessage(t *testing.T) {
	brokenServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(apiErrorPayload{
			Message:   "version database unavailable",
			ErrorType: "INTERNAL_ERROR",
		})
	}))
	defer brokenServer.Close()

	client := &Client{
		BaseUrl:    strings.TrimPrefix(brokenServer.URL, "http://"),
		HTTPClient: brokenServer.Client(),
	}

	var artifact LambdaArtifact
	err := client.ReadLambdaArtifact("nsbno/my-service", "", "", "", &artifact)
	if err == nil {
		t.Fatal("expected an error from a broken server, got nil")
	}
	if !strings.Contains(err.Error(), "500") {
		t.Errorf("error = %q, want it to contain HTTP status '500'", err.Error())
	}
	if !strings.Contains(err.Error(), "version database unavailable") {
		t.Errorf("error = %q, want it to contain the API error message", err.Error())
	}
}
