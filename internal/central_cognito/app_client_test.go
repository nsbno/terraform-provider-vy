package central_cognito

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCreateAppClient_ReturnsCreatedClientWithGeneratedId(t *testing.T) {
	api := &FakeCentralCognitoAPI{
		AppClients:      map[string]AppClient{},
		ResourceServers: map[string]ResourceServer{},
	}
	server, client := api.Start()
	defer server.Close()

	result, err := client.CreateAppClient(AppClient{
		Name:   "my-app",
		Scopes: []string{"read", "write"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatalf("expected non-nil result")
	}
	if result.ClientId == nil {
		t.Fatalf("expected ClientId to be set")
	}
	if *result.ClientId != "generated-client-id-my-app" {
		t.Errorf("expected ClientId %q, got %q", "generated-client-id-my-app", *result.ClientId)
	}
	if result.ClientSecret != nil {
		t.Errorf("expected no ClientSecret, got %q", *result.ClientSecret)
	}
	if result.Name != "my-app" {
		t.Errorf("expected Name %q, got %q", "my-app", result.Name)
	}
}

func TestCreateAppClient_ReturnsClientSecretWhenGenerateSecretIsTrue(t *testing.T) {
	api := &FakeCentralCognitoAPI{
		AppClients:      map[string]AppClient{},
		ResourceServers: map[string]ResourceServer{},
	}
	server, client := api.Start()
	defer server.Close()

	genSecret := true
	result, err := client.CreateAppClient(AppClient{
		Name:           "secret-app",
		Scopes:         []string{"admin"},
		GenerateSecret: &genSecret,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ClientSecret == nil {
		t.Fatalf("expected ClientSecret to be set")
	}
	if *result.ClientSecret != "generated-secret-for-secret-app" {
		t.Errorf("expected ClientSecret %q, got %q", "generated-secret-for-secret-app", *result.ClientSecret)
	}
}

func TestReadAppClient_ReturnsAppClientForMatchingName(t *testing.T) {
	clientID := "existing-client-id"
	api := &FakeCentralCognitoAPI{
		AppClients: map[string]AppClient{
			"existing-app": {
				Name:     "existing-app",
				Scopes:   []string{"read"},
				ClientId: &clientID,
			},
		},
		ResourceServers: map[string]ResourceServer{},
	}
	server, client := api.Start()
	defer server.Close()

	var result AppClient
	err := client.ReadAppClient("existing-app", &result)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Name != "existing-app" {
		t.Errorf("expected Name %q, got %q", "existing-app", result.Name)
	}
	if len(result.Scopes) != 1 || result.Scopes[0] != "read" {
		t.Errorf("expected Scopes [read], got %v", result.Scopes)
	}
}

func TestUpdateAppClient_UpdatesScopesForExistingClient(t *testing.T) {
	clientID := "cid"
	api := &FakeCentralCognitoAPI{
		AppClients: map[string]AppClient{
			"my-app": {
				Name:     "my-app",
				Scopes:   []string{"read"},
				ClientId: &clientID,
			},
		},
		ResourceServers: map[string]ResourceServer{},
	}
	server, client := api.Start()
	defer server.Close()

	err := client.UpdateAppClient("my-app", AppClientUpdateRequest{
		Name:   "my-app",
		Scopes: []string{"read", "write"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result AppClient
	err = client.ReadAppClient("my-app", &result)
	if err != nil {
		t.Fatalf("unexpected error reading after update: %v", err)
	}
	if len(result.Scopes) != 2 {
		t.Errorf("expected 2 scopes after update, got %v", result.Scopes)
	}
}

func TestUpdateAppClient_RenamesClientAndKeepsItReadableUnderNewName(t *testing.T) {
	clientID := "cid"
	api := &FakeCentralCognitoAPI{
		AppClients: map[string]AppClient{
			"old-name": {
				Name:     "old-name",
				Scopes:   []string{"read"},
				ClientId: &clientID,
			},
		},
		ResourceServers: map[string]ResourceServer{},
	}
	server, client := api.Start()
	defer server.Close()

	err := client.UpdateAppClient("old-name", AppClientUpdateRequest{
		Name:   "new-name",
		Scopes: []string{"read"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result AppClient
	err = client.ReadAppClient("new-name", &result)
	if err != nil {
		t.Fatalf("unexpected error reading renamed app client: %v", err)
	}
	if result.Name != "new-name" {
		t.Errorf("expected Name %q, got %q", "new-name", result.Name)
	}

	err = client.ReadAppClient("old-name", &result)
	if err == nil {
		t.Fatalf("expected error reading app client under old name after rename, got nil")
	}
}

func TestDeleteAppClient_RemovesClientSoReadReturnsError(t *testing.T) {
	clientID := "cid"
	api := &FakeCentralCognitoAPI{
		AppClients: map[string]AppClient{
			"to-delete": {
				Name:     "to-delete",
				ClientId: &clientID,
			},
		},
		ResourceServers: map[string]ResourceServer{},
	}
	server, client := api.Start()
	defer server.Close()

	err := client.DeleteAppClient("to-delete")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result AppClient
	err = client.ReadAppClient("to-delete", &result)
	if err == nil {
		t.Fatalf("expected error reading deleted app client, got nil")
	}
}

func TestImportAppClient_ReturnsAppClientForMatchingClientId(t *testing.T) {
	clientID := "import-client-id"
	api := &FakeCentralCognitoAPI{
		AppClients: map[string]AppClient{
			"importable-app": {
				Name:     "importable-app",
				Scopes:   []string{"admin"},
				ClientId: &clientID,
			},
		},
		ResourceServers: map[string]ResourceServer{},
	}
	server, client := api.Start()
	defer server.Close()

	var result AppClient
	err := client.ImportAppClient("import-client-id", &result)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Name != "importable-app" {
		t.Errorf("expected Name %q, got %q", "importable-app", result.Name)
	}
}

func TestReadAppClient_ReturnsErrorWhenClientDoesNotExist(t *testing.T) {
	api := &FakeCentralCognitoAPI{
		AppClients:      map[string]AppClient{},
		ResourceServers: map[string]ResourceServer{},
	}
	server, client := api.Start()
	defer server.Close()

	var result AppClient
	err := client.ReadAppClient("nonexistent", &result)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if err.Error() != `app client "nonexistent" not found` {
		t.Errorf("expected the message from the response body, got: %v", err)
	}
}

func TestDeleteAppClient_ReturnsErrorWhenClientDoesNotExist(t *testing.T) {
	api := &FakeCentralCognitoAPI{
		AppClients:      map[string]AppClient{},
		ResourceServers: map[string]ResourceServer{},
	}
	server, client := api.Start()
	defer server.Close()

	err := client.DeleteAppClient("nonexistent")
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if err.Error() != `app client "nonexistent" not found` {
		t.Errorf("expected the message from the response body, got: %v", err)
	}
}

func TestImportAppClient_ReturnsErrorWhenClientIdDoesNotExist(t *testing.T) {
	api := &FakeCentralCognitoAPI{
		AppClients:      map[string]AppClient{},
		ResourceServers: map[string]ResourceServer{},
	}
	server, client := api.Start()
	defer server.Close()

	var result AppClient
	err := client.ImportAppClient("no-such-client-id", &result)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "could not import resource") {
		t.Errorf("expected 'could not import resource' in error, got: %v", err)
	}
}

func TestCreateAppClient_ReturnsErrorWhenClientAlreadyExists(t *testing.T) {
	api := &FakeCentralCognitoAPI{
		AppClients: map[string]AppClient{
			"existing": {Name: "existing"},
		},
		ResourceServers: map[string]ResourceServer{},
	}
	server, client := api.Start()
	defer server.Close()

	_, err := client.CreateAppClient(AppClient{Name: "existing"})
	if err == nil {
		t.Fatalf("expected conflict error, got nil")
	}
	if err.Error() != `app client "existing" already exists` {
		t.Errorf("expected the message from the response body, got: %v", err)
	}
}

func TestCreateAppClient_AcceptsEmptyScopes(t *testing.T) {
	api := &FakeCentralCognitoAPI{
		AppClients:      map[string]AppClient{},
		ResourceServers: map[string]ResourceServer{},
	}
	server, client := api.Start()
	defer server.Close()

	result, err := client.CreateAppClient(AppClient{
		Name:   "no-scopes-app",
		Scopes: []string{},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Scopes) != 0 {
		t.Errorf("expected empty scopes, got %v", result.Scopes)
	}
}

func TestReadAppClient_HandlesUrlEncodedName(t *testing.T) {
	clientID := "cid"
	api := &FakeCentralCognitoAPI{
		AppClients: map[string]AppClient{
			"app with spaces": {
				Name:     "app with spaces",
				Scopes:   []string{"read"},
				ClientId: &clientID,
			},
		},
		ResourceServers: map[string]ResourceServer{},
	}
	server, client := api.Start()
	defer server.Close()

	var result AppClient
	err := client.ReadAppClient("app with spaces", &result)
	if err != nil {
		t.Fatalf("unexpected error reading URL-encoded name: %v", err)
	}
	if result.Name != "app with spaces" {
		t.Errorf("expected Name %q, got %q", "app with spaces", result.Name)
	}
}

func startAppClientErrorServer(t *testing.T, statusCode int, body string) *Client {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(statusCode)
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(server.Close)

	return &Client{
		BaseUrl:    strings.TrimPrefix(server.URL, "http://"),
		HTTPClient: server.Client(),
	}
}

func TestAppClient_ErrorMessageFromResponseBody(t *testing.T) {
	testCases := []struct {
		name       string
		statusCode int
		body       string
		expected   string
	}{
		{
			name:       "problem details uses detail",
			statusCode: http.StatusConflict,
			body:       `{"title":"AppClientAlreadyExistsException","status":409,"detail":"An app client already exists with this name."}`,
			expected:   "An app client already exists with this name.",
		},
		{
			name:       "message format uses message",
			statusCode: http.StatusNotFound,
			body:       `{"message":"App client not found","error_type":"NOT_FOUND"}`,
			expected:   "App client not found",
		},
		{
			name:       "non-JSON body is returned as is",
			statusCode: http.StatusBadGateway,
			body:       "upstream connect error\n",
			expected:   "upstream connect error",
		},
		{
			name:       "JSON without detail or message is returned as is",
			statusCode: http.StatusInternalServerError,
			body:       `{"title":"InternalError","status":500}`,
			expected:   `{"title":"InternalError","status":500}`,
		},
		{
			name:       "empty body gives the status code",
			statusCode: http.StatusInternalServerError,
			body:       "",
			expected:   "unexpected status code 500",
		},
	}

	operations := map[string]func(client *Client) error{
		"create": func(client *Client) error {
			_, err := client.CreateAppClient(AppClient{Name: "my-app"})
			return err
		},
		"read": func(client *Client) error {
			var result AppClient
			return client.ReadAppClient("my-app", &result)
		},
		"update": func(client *Client) error {
			return client.UpdateAppClient("my-app", AppClientUpdateRequest{Name: "new-name"})
		},
		"delete": func(client *Client) error {
			return client.DeleteAppClient("my-app")
		},
	}

	for _, tc := range testCases {
		for operationName, operation := range operations {
			t.Run(operationName+"/"+tc.name, func(t *testing.T) {
				client := startAppClientErrorServer(t, tc.statusCode, tc.body)

				err := operation(client)
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				if err.Error() != tc.expected {
					t.Errorf("expected error %q, got %q", tc.expected, err.Error())
				}
			})
		}
	}
}
