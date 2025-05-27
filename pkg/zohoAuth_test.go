package pkg

import (
	"os"
	"testing"
)

func TestLoadZohoConfigFromEnv(t *testing.T) {
	// 設置測試環境變數
	os.Setenv("ZOHO_CLIENT_ID", "test_client_id")
	os.Setenv("ZOHO_CLIENT_SECRET", "test_client_secret")
	os.Setenv("ZOHO_SCOPES", "scope1,scope2, scope3")
	os.Setenv("ZOHO_ALLOW_DOMAINS", "domain1.com, domain2.com")
	os.Setenv("ZOHO_REDIRECT_URI", "http://localhost:8080/callback")

	config := LoadZohoConfigFromEnv()

	if config.ClientID != "test_client_id" {
		t.Errorf("Expected ClientID %s, got %s", "test_client_id", config.ClientID)
	}
	if config.ClientSecret != "test_client_secret" {
		t.Errorf("Expected ClientSecret %s, got %s", "test_client_secret", config.ClientSecret)
	}
	if len(config.Scopes) != 3 {
		t.Errorf("Expected 3 scopes, got %d", len(config.Scopes))
	}
	if len(config.AllowDomains) != 2 {
		t.Errorf("Expected 2 allowed domains, got %d", len(config.AllowDomains))
	}
}

func TestFilterNonEmpty(t *testing.T) {
	tests := []struct {
		input    []string
		expected []string
	}{
		{[]string{"a", "b", "c"}, []string{"a", "b", "c"}},
		{[]string{"a", "", "c"}, []string{"a", "c"}},
		{[]string{" ", "b", " "}, []string{"b"}},
		{[]string{}, []string{}},
	}

	for _, test := range tests {
		result := filterNonEmpty(test.input)
		if len(result) != len(test.expected) {
			t.Errorf("Expected length %d, got %d", len(test.expected), len(result))
		}
		for i := range result {
			if result[i] != test.expected[i] {
				t.Errorf("Expected %s at position %d, got %s", test.expected[i], i, result[i])
			}
		}
	}
}

func TestIsEmailDomainAllowed(t *testing.T) {
	config := ZohoConfig{
		AllowDomains: []string{"example.com", "test.com"},
	}
	oauth := NewZohoOAuth(config)

	tests := []struct {
		email    string
		expected bool
	}{
		{"user@example.com", true},
		{"user@test.com", true},
		{"user@other.com", false},
		{"invalid-email", false},
	}

	for _, test := range tests {
		result := oauth.IsEmailDomainAllowed(test.email)
		if result != test.expected {
			t.Errorf("For email %s, expected %v but got %v", test.email, test.expected, result)
		}
	}
}
