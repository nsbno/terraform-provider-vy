package version_handler_v2

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestReadECSImage_ReturnsVersionForMatchingRepositoryAndECRName(t *testing.T) {
	api := &FakeVersionHandlerAPI{
		KnownECSVersions: []ECSVersion{
			{
				GitHubRepositoryName: "nsbno/my-service",
				ECRRepositoryName:    "my-service",
				ECRRepositoryURI:     "123456789.dkr.ecr.eu-west-1.amazonaws.com/my-service",
				GitSha:               "deadbeef",
			},
		},
	}
	server, client := api.Start()
	defer server.Close()

	var version ECSVersion
	err := client.ReadECSImage("nsbno/my-service", "my-service", "", &version)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if version.ECRRepositoryURI != "123456789.dkr.ecr.eu-west-1.amazonaws.com/my-service" {
		t.Errorf("ECRRepositoryURI = %q, want %q", version.ECRRepositoryURI, "123456789.dkr.ecr.eu-west-1.amazonaws.com/my-service")
	}
	if version.GitSha != "deadbeef" {
		t.Errorf("GitSha = %q, want %q", version.GitSha, "deadbeef")
	}
}

func TestReadECSImage_DistinguishesMonorepoServicesByWorkingDirectory(t *testing.T) {
	api := &FakeVersionHandlerAPI{
		KnownECSVersions: []ECSVersion{
			{
				GitHubRepositoryName: "nsbno/monorepo",
				ECRRepositoryName:    "api",
				WorkingDirectory:     "services/api",
				GitSha:               "aaa111",
			},
			{
				GitHubRepositoryName: "nsbno/monorepo",
				ECRRepositoryName:    "api",
				WorkingDirectory:     "services/worker",
				GitSha:               "bbb222",
			},
		},
	}
	server, client := api.Start()
	defer server.Close()

	var version ECSVersion
	err := client.ReadECSImage("nsbno/monorepo", "api", "services/worker", &version)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if version.GitSha != "bbb222" {
		t.Errorf("GitSha = %q, want %q", version.GitSha, "bbb222")
	}
}

func TestReadECSImage_ReturnsErrorWhenVersionDoesNotExist(t *testing.T) {
	api := &FakeVersionHandlerAPI{}
	server, client := api.Start()
	defer server.Close()

	var version ECSVersion
	err := client.ReadECSImage("nsbno/nonexistent", "no-such-ecr", "", &version)
	if err == nil {
		t.Fatal("expected an error for a missing ECS version, got nil")
	}
	if !strings.Contains(err.Error(), "404") {
		t.Errorf("error = %q, want it to contain HTTP status '404'", err.Error())
	}
}

func TestReadECSImage_PropagatesServerErrorWithStatusAndMessage(t *testing.T) {
	brokenServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(apiErrorPayload{
			Message:   "ECR registry unreachable",
			ErrorType: "INTERNAL_ERROR",
		})
	}))
	defer brokenServer.Close()

	client := &Client{
		BaseUrl:    strings.TrimPrefix(brokenServer.URL, "http://"),
		HTTPClient: brokenServer.Client(),
	}

	var version ECSVersion
	err := client.ReadECSImage("nsbno/my-service", "my-service", "", &version)
	if err == nil {
		t.Fatal("expected an error from a broken server, got nil")
	}
	if !strings.Contains(err.Error(), "500") {
		t.Errorf("error = %q, want it to contain HTTP status '500'", err.Error())
	}
	if !strings.Contains(err.Error(), "ECR registry unreachable") {
		t.Errorf("error = %q, want it to contain the API error message", err.Error())
	}
}
