package pkg

import (
	"bedrock-claude-proxy/tests"
	"testing"
	"time"
)

func TestCache(t *testing.T) {
	// Initialize cache
	cache, err := NewCache()
	if err != nil {
		t.Fatalf("Failed to create cache: %v", err)
	}
	defer cache.Close()

	// Test SaveAPIKey and GetAPIKey
	t.Run("SaveAndGetAPIKey", func(t *testing.T) {
		email := "test@example.com"
		apiKey := "test-api-key-12345"

		// Save API key
		err := cache.SaveAPIKey(email, apiKey)
		if err != nil {
			t.Fatalf("Failed to save API key: %v", err)
		}

		// Get API key
		retrievedKey, err := cache.GetAPIKey(email)
		if err != nil {
			t.Fatalf("Failed to get API key: %v", err)
		}

		t.Log(tests.ToJSON(retrievedKey))

		if retrievedKey != apiKey {
			t.Errorf("Retrieved API key does not match. Got %s, want %s", retrievedKey, apiKey)
		}
	})

	// Test HasAPIKey
	t.Run("HasAPIKey", func(t *testing.T) {
		email := "exists@example.com"
		apiKey := "exists-api-key-12345"

		// Save API key
		err := cache.SaveAPIKey(email, apiKey)
		if err != nil {
			t.Fatalf("Failed to save API key: %v", err)
		}

		// Check if API key exists
		exists, err := cache.HasAPIKey(email)
		if err != nil {
			t.Fatalf("Failed to check API key existence: %v", err)
		}

		if !exists {
			t.Errorf("API key should exist but HasAPIKey returned false")
		}

		// Check non-existent key
		exists, err = cache.HasAPIKey("nonexistent@example.com")
		if err != nil {
			t.Fatalf("Failed to check API key existence: %v", err)
		}

		if exists {
			t.Errorf("API key should not exist but HasAPIKey returned true")
		}
	})

	// Test DeleteAPIKey
	t.Run("DeleteAPIKey", func(t *testing.T) {
		email := "delete@example.com"
		apiKey := "delete-api-key-12345"

		// Save API key
		err := cache.SaveAPIKey(email, apiKey)
		if err != nil {
			t.Fatalf("Failed to save API key: %v", err)
		}

		// Delete API key
		err = cache.DeleteAPIKey(email)
		if err != nil {
			t.Fatalf("Failed to delete API key: %v", err)
		}

		// Verify key is deleted
		exists, err := cache.HasAPIKey(email)
		if err != nil {
			t.Fatalf("Failed to check API key existence: %v", err)
		}

		if exists {
			t.Errorf("API key should have been deleted but still exists")
		}
	})

	// Test expiration
	t.Run("Expiration", func(t *testing.T) {
		email := "expire@example.com"
		apiKey := "expire-api-key-12345"

		// Save API key with short expiration
		err := cache.SaveAPIKey(email, apiKey, 1*time.Second)
		if err != nil {
			t.Fatalf("Failed to save API key: %v", err)
		}

		// Verify key exists immediately
		exists, err := cache.HasAPIKey(email)
		if err != nil {
			t.Fatalf("Failed to check API key existence: %v", err)
		}

		if !exists {
			t.Errorf("API key should exist immediately after saving")
		}

		// Wait for expiration
		time.Sleep(2 * time.Second)

		// Verify key is expired
		exists, err = cache.HasAPIKey(email)
		if err != nil {
			t.Fatalf("Failed to check API key existence: %v", err)
		}

		if exists {
			t.Errorf("API key should have expired but still exists")
		}
	})
}
