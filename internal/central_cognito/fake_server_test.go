package central_cognito

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
)

type FakeCentralCognitoAPI struct {
	AppClients      map[string]AppClient      // name → AppClient
	ResourceServers map[string]ResourceServer // identifier → ResourceServer
}

func (api *FakeCentralCognitoAPI) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	rawPath := r.URL.RawPath
	if rawPath == "" {
		rawPath = r.URL.Path
	}
	path := strings.TrimPrefix(rawPath, "/")
	segments := strings.SplitN(path, "/", 3)

	switch {
	case r.Method == http.MethodPost && path == "import/app-client":
		api.handleImportAppClient(w, r)

	case r.Method == http.MethodPost && path == "import/resource-server":
		api.handleImportResourceServer(w, r)

	case r.Method == http.MethodPost && path == "app-clients":
		api.handleCreateAppClient(w, r)

	case len(segments) == 2 && segments[0] == "app-clients":
		name, err := url.QueryUnescape(segments[1])
		if err != nil {
			respondWithError(w, http.StatusBadRequest, "invalid URL encoding", "BAD_REQUEST")
			return
		}
		switch r.Method {
		case http.MethodGet:
			api.handleReadAppClient(w, name)
		case http.MethodPut:
			api.handleUpdateAppClient(w, r, name)
		case http.MethodDelete:
			api.handleDeleteAppClient(w, name)
		default:
			respondWithError(w, http.StatusMethodNotAllowed, "method not allowed", "METHOD_NOT_ALLOWED")
		}

	case r.Method == http.MethodPost && path == "resource-servers":
		api.handleCreateResourceServer(w, r)

	case len(segments) == 2 && segments[0] == "resource-servers":
		identifier, err := url.QueryUnescape(segments[1])
		if err != nil {
			respondWithError(w, http.StatusBadRequest, "invalid URL encoding", "BAD_REQUEST")
			return
		}
		switch r.Method {
		case http.MethodGet:
			api.handleReadResourceServer(w, identifier)
		case http.MethodPut:
			api.handleUpdateResourceServer(w, r, identifier)
		case http.MethodDelete:
			api.handleDeleteResourceServer(w, identifier)
		default:
			respondWithError(w, http.StatusMethodNotAllowed, "method not allowed", "METHOD_NOT_ALLOWED")
		}

	default:
		respondWithError(w, http.StatusNotFound, "unknown endpoint: "+r.URL.Path, "NOT_FOUND")
	}
}

func (api *FakeCentralCognitoAPI) handleReadAppClient(w http.ResponseWriter, name string) {
	ac, ok := api.AppClients[name]
	if !ok {
		respondWithError(w, http.StatusNotFound, fmt.Sprintf("app client %q not found", name), "NOT_FOUND")
		return
	}
	respondWithJSON(w, http.StatusOK, ac)
}

func (api *FakeCentralCognitoAPI) handleCreateAppClient(w http.ResponseWriter, r *http.Request) {
	var ac AppClient
	if err := json.NewDecoder(r.Body).Decode(&ac); err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid request body: "+err.Error(), "BAD_REQUEST")
		return
	}

	if _, exists := api.AppClients[ac.Name]; exists {
		respondWithError(w, http.StatusConflict, fmt.Sprintf("app client %q already exists", ac.Name), "CONFLICT")
		return
	}

	clientID := "generated-client-id-" + ac.Name
	ac.ClientId = &clientID

	if ac.GenerateSecret != nil && *ac.GenerateSecret {
		secret := "generated-secret-for-" + ac.Name
		ac.ClientSecret = &secret
	}

	api.AppClients[ac.Name] = ac
	respondWithJSON(w, http.StatusCreated, ac)
}

func (api *FakeCentralCognitoAPI) handleUpdateAppClient(w http.ResponseWriter, r *http.Request, name string) {
	var req AppClientUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid request body: "+err.Error(), "BAD_REQUEST")
		return
	}

	existing, ok := api.AppClients[name]
	if !ok {
		respondWithError(w, http.StatusNotFound, fmt.Sprintf("app client %q not found", name), "NOT_FOUND")
		return
	}

	existing.Scopes = req.Scopes
	existing.CallbackUrls = req.CallbackUrls
	existing.LogoutUrls = req.LogoutUrls
	api.AppClients[name] = existing

	respondWithJSON(w, http.StatusOK, existing)
}

func (api *FakeCentralCognitoAPI) handleDeleteAppClient(w http.ResponseWriter, name string) {
	if _, ok := api.AppClients[name]; !ok {
		respondWithError(w, http.StatusNotFound, fmt.Sprintf("app client %q not found", name), "NOT_FOUND")
		return
	}
	delete(api.AppClients, name)
	w.WriteHeader(http.StatusOK)
}

func (api *FakeCentralCognitoAPI) handleImportAppClient(w http.ResponseWriter, r *http.Request) {
	var req ImportAppClientRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid request body: "+err.Error(), "BAD_REQUEST")
		return
	}

	for _, ac := range api.AppClients {
		if ac.ClientId != nil && *ac.ClientId == req.ClientId {
			respondWithJSON(w, http.StatusOK, ac)
			return
		}
	}

	respondWithError(w, http.StatusNotFound, fmt.Sprintf("app client with client_id %q not found", req.ClientId), "NOT_FOUND")
}

func (api *FakeCentralCognitoAPI) handleReadResourceServer(w http.ResponseWriter, identifier string) {
	rs, ok := api.ResourceServers[identifier]
	if !ok {
		respondWithError(w, http.StatusNotFound, fmt.Sprintf("resource server %q not found", identifier), "NOT_FOUND")
		return
	}
	respondWithJSON(w, http.StatusOK, rs)
}

func (api *FakeCentralCognitoAPI) handleCreateResourceServer(w http.ResponseWriter, r *http.Request) {
	var rs ResourceServer
	if err := json.NewDecoder(r.Body).Decode(&rs); err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid request body: "+err.Error(), "BAD_REQUEST")
		return
	}

	if _, exists := api.ResourceServers[rs.Identifier]; exists {
		respondWithError(w, http.StatusConflict, fmt.Sprintf("resource server %q already exists", rs.Identifier), "CONFLICT")
		return
	}

	api.ResourceServers[rs.Identifier] = rs
	respondWithJSON(w, http.StatusCreated, rs)
}

func (api *FakeCentralCognitoAPI) handleUpdateResourceServer(w http.ResponseWriter, r *http.Request, identifier string) {
	var req ResourceServerUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid request body: "+err.Error(), "BAD_REQUEST")
		return
	}

	existing, ok := api.ResourceServers[identifier]
	if !ok {
		respondWithError(w, http.StatusNotFound, fmt.Sprintf("resource server %q not found", identifier), "NOT_FOUND")
		return
	}

	existing.Name = req.Name
	existing.Scopes = req.Scopes
	api.ResourceServers[identifier] = existing

	respondWithJSON(w, http.StatusOK, existing)
}

func (api *FakeCentralCognitoAPI) handleDeleteResourceServer(w http.ResponseWriter, identifier string) {
	if _, ok := api.ResourceServers[identifier]; !ok {
		respondWithError(w, http.StatusNotFound, fmt.Sprintf("resource server %q not found", identifier), "NOT_FOUND")
		return
	}
	delete(api.ResourceServers, identifier)
	w.WriteHeader(http.StatusOK)
}

func (api *FakeCentralCognitoAPI) handleImportResourceServer(w http.ResponseWriter, r *http.Request) {
	var req ImportResourceServerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid request body: "+err.Error(), "BAD_REQUEST")
		return
	}

	rs, ok := api.ResourceServers[req.Identifier]
	if !ok {
		respondWithError(w, http.StatusNotFound, fmt.Sprintf("resource server %q not found", req.Identifier), "NOT_FOUND")
		return
	}

	respondWithJSON(w, http.StatusOK, rs)
}

func (api *FakeCentralCognitoAPI) Start() (*httptest.Server, *Client) {
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
	respondWithJSON(w, statusCode, map[string]string{"message": message, "error_type": errorType})
}
