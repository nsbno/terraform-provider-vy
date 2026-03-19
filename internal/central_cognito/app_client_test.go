package central_cognito

import (
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

	err := client.UpdateAppClient(AppClientUpdateRequest{
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
	if !strings.Contains(err.Error(), "could not read resource") {
		t.Errorf("expected 'could not read resource' in error, got: %v", err)
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
	if !strings.Contains(err.Error(), "could not delete resource") {
		t.Errorf("expected 'could not delete resource' in error, got: %v", err)
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
	if !strings.Contains(err.Error(), "could not create resource") {
		t.Errorf("expected 'could not create resource' in error, got: %v", err)
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
