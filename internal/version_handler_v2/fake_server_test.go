package version_handler_v2

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
)

// FakeVersionHandlerAPI is an in-memory HTTP fake that emulates the version-handler v2 API.
// Populate it with known artifacts, then call Start() to get a running test server
// and a pre-configured Client.
type FakeVersionHandlerAPI struct {
	KnownLambdaArtifacts []LambdaArtifact
	KnownECSVersions     []ECSVersion
}

func (api *FakeVersionHandlerAPI) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Expected: /v2/versions/{repository}/lambda or /v2/versions/{repository}/ecs
	// Repository names may contain slashes (e.g. "nsbno/my-service"), so the path
	// can have more than 4 segments. The artifact kind is always the last segment,
	// and the repository name is everything between "versions/" and that last segment.
	segments := strings.Split(strings.TrimPrefix(r.URL.Path, "/"), "/")
	if len(segments) < 4 || segments[0] != "v2" || segments[1] != "versions" {
		respondWithError(w, http.StatusNotFound, "unknown endpoint", "NOT_FOUND")
		return
	}

	artifactKind := segments[len(segments)-1]
	repositoryName := strings.Join(segments[2:len(segments)-1], "/")
	queryParams := r.URL.Query()

	switch artifactKind {
	case "lambda":
		api.serveLambdaArtifact(w, repositoryName, queryParams)
	case "ecs":
		api.serveECSVersion(w, repositoryName, queryParams)
	default:
		respondWithError(w, http.StatusNotFound, "unknown artifact kind: "+artifactKind, "NOT_FOUND")
	}
}

func (api *FakeVersionHandlerAPI) serveLambdaArtifact(w http.ResponseWriter, repositoryName string, queryParams map[string][]string) {
	requestedECRName := firstQueryValue(queryParams, "ecr_repository_name")
	requestedWorkDir := firstQueryValue(queryParams, "working_directory")
	requestedPath := firstQueryValue(queryParams, "path")

	for _, artifact := range api.KnownLambdaArtifacts {
		if artifact.GitHubRepositoryName != repositoryName {
			continue
		}
		if requestedECRName != "" && artifact.ECRRepositoryName != requestedECRName {
			continue
		}
		if requestedWorkDir != "" && normalizePath(artifact.WorkingDirectory) != normalizePath(requestedWorkDir) {
			continue
		}
		if requestedPath != "" && normalizePath(artifact.Path) != normalizePath(requestedPath) {
			continue
		}
		respondWithJSON(w, http.StatusOK, artifact)
		return
	}

	respondWithError(w, http.StatusNotFound, "artifact not found", "NOT_FOUND")
}

func (api *FakeVersionHandlerAPI) serveECSVersion(w http.ResponseWriter, repositoryName string, queryParams map[string][]string) {
	requestedECRName := firstQueryValue(queryParams, "ecr_repository_name")
	requestedWorkDir := firstQueryValue(queryParams, "working_directory")

	for _, version := range api.KnownECSVersions {
		if version.GitHubRepositoryName != repositoryName {
			continue
		}
		if requestedECRName != "" && version.ECRRepositoryName != requestedECRName {
			continue
		}
		if requestedWorkDir != "" && normalizePath(version.WorkingDirectory) != normalizePath(requestedWorkDir) {
			continue
		}
		respondWithJSON(w, http.StatusOK, version)
		return
	}

	respondWithError(w, http.StatusNotFound, "artifact not found", "NOT_FOUND")
}

// Start launches an httptest.Server running the fake API and returns it alongside
// a Client pre-configured to talk to it. The caller must call server.Close() when done.
func (api *FakeVersionHandlerAPI) Start() (*httptest.Server, *Client) {
	server := httptest.NewServer(api)
	client := &Client{
		BaseUrl:    strings.TrimPrefix(server.URL, "http://"),
		HTTPClient: server.Client(),
	}
	return server, client
}

func respondWithJSON(w http.ResponseWriter, statusCode int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(payload)
}

func respondWithError(w http.ResponseWriter, statusCode int, message, errorType string) {
	respondWithJSON(w, statusCode, apiErrorPayload{Message: message, ErrorType: errorType})
}

func firstQueryValue(queryParams map[string][]string, key string) string {
	values, ok := queryParams[key]
	if !ok || len(values) == 0 {
		return ""
	}
	return values[0]
}
