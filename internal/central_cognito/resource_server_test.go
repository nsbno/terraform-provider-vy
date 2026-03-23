package central_cognito

import (
	"strings"
	"testing"
)

func TestCreateResourceServer_CreatesServerAndReadReturnsIt(t *testing.T) {
	api := &FakeCentralCognitoAPI{
		AppClients:      map[string]AppClient{},
		ResourceServers: map[string]ResourceServer{},
	}
	server, client := api.Start()
	defer server.Close()

	err := client.CreateResourceServer(ResourceServer{
		Identifier: "https://api.example.com",
		Name:       "Example API",
		Scopes: []Scope{
			{Name: "read", Description: "Read access"},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result ResourceServer
	err = client.ReadResourceServer("https://api.example.com", &result)
	if err != nil {
		t.Fatalf("unexpected error reading after create: %v", err)
	}
	if result.Identifier != "https://api.example.com" {
		t.Errorf("expected Identifier %q, got %q", "https://api.example.com", result.Identifier)
	}
	if result.Name != "Example API" {
		t.Errorf("expected Name %q, got %q", "Example API", result.Name)
	}
}

func TestReadResourceServer_ReturnsServerForMatchingIdentifier(t *testing.T) {
	api := &FakeCentralCognitoAPI{
		AppClients: map[string]AppClient{},
		ResourceServers: map[string]ResourceServer{
			"https://api.example.com": {
				Identifier: "https://api.example.com",
				Name:       "Example API",
				Scopes: []Scope{
					{Name: "read", Description: "Read access"},
					{Name: "write", Description: "Write access"},
				},
			},
		},
	}
	server, client := api.Start()
	defer server.Close()

	var result ResourceServer
	err := client.ReadResourceServer("https://api.example.com", &result)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Identifier != "https://api.example.com" {
		t.Errorf("expected Identifier %q, got %q", "https://api.example.com", result.Identifier)
	}
	if len(result.Scopes) != 2 {
		t.Errorf("expected 2 scopes, got %d", len(result.Scopes))
	}
}

func TestUpdateResourceServer_UpdatesNameAndScopesForExistingServer(t *testing.T) {
	api := &FakeCentralCognitoAPI{
		AppClients: map[string]AppClient{},
		ResourceServers: map[string]ResourceServer{
			"https://api.example.com": {
				Identifier: "https://api.example.com",
				Name:       "Old Name",
				Scopes:     []Scope{{Name: "read", Description: "Read"}},
			},
		},
	}
	server, client := api.Start()
	defer server.Close()

	err := client.UpdateResourceServer(ResourceServerUpdateRequest{
		Identifier: "https://api.example.com",
		Name:       "New Name",
		Scopes: []Scope{
			{Name: "read", Description: "Read"},
			{Name: "write", Description: "Write"},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result ResourceServer
	err = client.ReadResourceServer("https://api.example.com", &result)
	if err != nil {
		t.Fatalf("unexpected error reading after update: %v", err)
	}
	if result.Name != "New Name" {
		t.Errorf("expected Name %q, got %q", "New Name", result.Name)
	}
	if len(result.Scopes) != 2 {
		t.Errorf("expected 2 scopes after update, got %d", len(result.Scopes))
	}
}

func TestDeleteResourceServer_RemovesServerSoReadReturnsError(t *testing.T) {
	api := &FakeCentralCognitoAPI{
		AppClients: map[string]AppClient{},
		ResourceServers: map[string]ResourceServer{
			"https://api.example.com": {
				Identifier: "https://api.example.com",
				Name:       "Example API",
			},
		},
	}
	server, client := api.Start()
	defer server.Close()

	err := client.DeleteResourceServer("https://api.example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result ResourceServer
	err = client.ReadResourceServer("https://api.example.com", &result)
	if err == nil {
		t.Fatalf("expected error reading deleted resource server, got nil")
	}
}

func TestImportResourceServer_ReturnsServerForMatchingIdentifier(t *testing.T) {
	api := &FakeCentralCognitoAPI{
		AppClients: map[string]AppClient{},
		ResourceServers: map[string]ResourceServer{
			"https://api.example.com": {
				Identifier: "https://api.example.com",
				Name:       "Example API",
				Scopes:     []Scope{{Name: "admin", Description: "Admin access"}},
			},
		},
	}
	server, client := api.Start()
	defer server.Close()

	var result ResourceServer
	err := client.ImportResourceServer("https://api.example.com", &result)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Identifier != "https://api.example.com" {
		t.Errorf("expected Identifier %q, got %q", "https://api.example.com", result.Identifier)
	}
	if result.Name != "Example API" {
		t.Errorf("expected Name %q, got %q", "Example API", result.Name)
	}
}

func TestReadResourceServer_ReturnsErrorWhenServerDoesNotExist(t *testing.T) {
	api := &FakeCentralCognitoAPI{
		AppClients:      map[string]AppClient{},
		ResourceServers: map[string]ResourceServer{},
	}
	server, client := api.Start()
	defer server.Close()

	var result ResourceServer
	err := client.ReadResourceServer("nonexistent", &result)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "could not read resource") {
		t.Errorf("expected 'could not read resource' in error, got: %v", err)
	}
}

func TestDeleteResourceServer_ReturnsErrorWhenServerDoesNotExist(t *testing.T) {
	api := &FakeCentralCognitoAPI{
		AppClients:      map[string]AppClient{},
		ResourceServers: map[string]ResourceServer{},
	}
	server, client := api.Start()
	defer server.Close()

	err := client.DeleteResourceServer("nonexistent")
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "could not delete resource") {
		t.Errorf("expected 'could not delete resource' in error, got: %v", err)
	}
}

func TestImportResourceServer_ReturnsErrorWhenIdentifierDoesNotExist(t *testing.T) {
	api := &FakeCentralCognitoAPI{
		AppClients:      map[string]AppClient{},
		ResourceServers: map[string]ResourceServer{},
	}
	server, client := api.Start()
	defer server.Close()

	var result ResourceServer
	err := client.ImportResourceServer("nonexistent", &result)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "could not import resource") {
		t.Errorf("expected 'could not import resource' in error, got: %v", err)
	}
}

func TestCreateResourceServer_ReturnsErrorWhenServerAlreadyExists(t *testing.T) {
	api := &FakeCentralCognitoAPI{
		AppClients: map[string]AppClient{},
		ResourceServers: map[string]ResourceServer{
			"https://api.example.com": {
				Identifier: "https://api.example.com",
				Name:       "Existing",
			},
		},
	}
	server, client := api.Start()
	defer server.Close()

	err := client.CreateResourceServer(ResourceServer{
		Identifier: "https://api.example.com",
		Name:       "Duplicate",
	})
	if err == nil {
		t.Fatalf("expected conflict error, got nil")
	}
	if !strings.Contains(err.Error(), "could not create resource") {
		t.Errorf("expected 'could not create resource' in error, got: %v", err)
	}
}

func TestCreateResourceServer_AcceptsEmptyScopes(t *testing.T) {
	api := &FakeCentralCognitoAPI{
		AppClients:      map[string]AppClient{},
		ResourceServers: map[string]ResourceServer{},
	}
	server, client := api.Start()
	defer server.Close()

	err := client.CreateResourceServer(ResourceServer{
		Identifier: "https://api.example.com",
		Name:       "No Scopes API",
		Scopes:     []Scope{},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result ResourceServer
	err = client.ReadResourceServer("https://api.example.com", &result)
	if err != nil {
		t.Fatalf("unexpected error reading: %v", err)
	}
	if len(result.Scopes) != 0 {
		t.Errorf("expected empty scopes, got %v", result.Scopes)
	}
}

func TestReadResourceServer_HandlesUrlEncodedIdentifier(t *testing.T) {
	api := &FakeCentralCognitoAPI{
		AppClients: map[string]AppClient{},
		ResourceServers: map[string]ResourceServer{
			"https://api.example.com/my resource": {
				Identifier: "https://api.example.com/my resource",
				Name:       "Spaced API",
			},
		},
	}
	server, client := api.Start()
	defer server.Close()

	var result ResourceServer
	err := client.ReadResourceServer("https://api.example.com/my resource", &result)
	if err != nil {
		t.Fatalf("unexpected error reading URL-encoded identifier: %v", err)
	}
	if result.Identifier != "https://api.example.com/my resource" {
		t.Errorf("expected Identifier %q, got %q", "https://api.example.com/my resource", result.Identifier)
	}
}
