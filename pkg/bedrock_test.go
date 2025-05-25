package pkg

import (
	"bedrock-claude-proxy/tests"
	_ "bedrock-claude-proxy/tests"
	"bytes"
	"io"
	"net/http/httptest"
	"testing"
)

func GetBedrockTestConfig() *BedrockConfig {
	return LoadBedrockConfigWithEnv()
}

func TestBedrockClient_HandleProxyJSON(t *testing.T) {
	config := GetBedrockTestConfig()

	config.DEBUG = true

	tests.ToJSON(config)

	bedrock := NewBedrockClient(config)

	bodyJSON := `{
    "max_tokens": 1024,
    "messages": [{"role":"user","content":[{"type":"text","text":"創作一首7言律詩"}]}],
	"temperature":0.5,
	"top_p":1,"top_k":5,"system":"You are a helpful assistant.",
    "model": "claude-3-haiku-20240307",
    "stream": false
}`
	// 創建一個測試請求
	req := httptest.NewRequest("POST", "https://api.anthropic.com/v1/messages", bytes.NewBufferString(bodyJSON))
	req.Header.Set("Content-Type", "application/json")
	//req.Header.Set("Accept", "application/json")

	// 創建一個響應記錄器
	w := httptest.NewRecorder()

	bedrock.HandleProxy(w, req)

	response := w.Result()
	if response.StatusCode != 200 {
		t.Errorf("Expected status code 200, got %d\n", response.StatusCode)
	} else {
		t.Logf("status: %d\n", response.StatusCode)
		t.Logf("header: %s\n", tests.ToJSON(response.Header))
		respData, err := io.ReadAll(response.Body)
		if err == nil {
			t.Logf("body: %s", string(respData))
		}
	}
}

func TestBedrockClient_HandleProxyStream(t *testing.T) {
	config := GetBedrockTestConfig()

	config.DEBUG = true

	tests.ToJSON(config)

	bedrock := NewBedrockClient(config)

	bodyJSON := `{
    "max_tokens": 1024,
    "messages": [{"role":"user","content":[{"type":"text","text":"創作1篇7言律诗"}]}],
	"temperature":0.5,
	"top_p":1,"top_k":5,"system":"You are a helpful assistant.",
    "model": "claude-3-haiku-20240307",
    "stream": true
}`
	// 創建一個測試請求
	req := httptest.NewRequest("POST", "https://api.anthropic.com/v1/messages", bytes.NewBufferString(bodyJSON))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "text/event-stream")

	// 創建一個響應記錄器
	w := httptest.NewRecorder()

	bedrock.HandleProxy(w, req)

	resp := w.Result()
	if resp.StatusCode != 200 {
		t.Errorf("Expected status code 200, got %d", resp.StatusCode)
	} else {
		t.Logf("status: %d", resp.StatusCode)
		t.Logf("headers: %v", resp.Header)
		respData, err := io.ReadAll(resp.Body)
		if err == nil {
			t.Logf("body: %s", string(respData))
		}
	}
}

func TestBedrockClient_GetMergedModelList(t *testing.T) {
	config := GetBedrockTestConfig()
	// config.DEBUG = true

	// Set up test model mappings and version mappings
	if config.ModelMappings == nil {
		config.ModelMappings = make(map[string]string)
	}
	if config.AnthropicVersionMappings == nil {
		config.AnthropicVersionMappings = make(map[string]string)
	}

	t.Logf("Test config: %s", tests.ToJSON(config))

	bedrock := NewBedrockClient(config)

	// Test GetMergedModelList
	models, err := bedrock.GetMergedModelList()

	if err != nil {
		t.Logf("GetMergedModelList returned error (this may be expected if AWS credentials are not configured): %v", err)
		// Even with error, it should fall back to config models
		if len(models) == 0 {
			t.Errorf("Expected fallback to config models, but got empty list")
		}
	}

	if len(models) == 0 {
		t.Errorf("Expected at least some models, got empty list")
	} else {
		t.Logf("Found %d models:", len(models))
		for _, model := range models {
			t.Log(tests.ToJSON(model))
		}
	}

	// Verify that models have required fields
	for _, model := range models {
		if model.Name == "" {
			t.Errorf("Model name should not be empty")
		}
		if model.Version == "" {
			t.Errorf("Model version should not be empty for model %s", model.Name)
		}
	}
}

func TestBedrockClient_GetMergedModelList_FallbackToConfig(t *testing.T) {
	config := GetBedrockTestConfig()
	config.DEBUG = true

	// Set up test configuration with invalid AWS credentials to force fallback
	config.AccessKey = "invalid-access-key"
	config.SecretKey = "invalid-secret-key"
	config.Region = "us-east-1"

	// Set up model mappings
	if config.ModelMappings == nil {
		config.ModelMappings = make(map[string]string)
	}
	if config.AnthropicVersionMappings == nil {
		config.AnthropicVersionMappings = make(map[string]string)
	}

	config.ModelMappings["test-model-1"] = "anthropic.claude-3-haiku-20240307-v1:0"
	config.ModelMappings["test-model-2"] = "anthropic.claude-3-sonnet-20240229-v1:0"
	config.AnthropicVersionMappings["test-model-1"] = "2023-06-01"
	config.AnthropicVersionMappings["test-model-2"] = "2023-06-01"
	config.AnthropicDefaultVersion = "2023-06-01"

	bedrock := NewBedrockClient(config)

	// Test GetMergedModelList - should fall back to config models
	models, err := bedrock.GetMergedModelList()

	// Error is expected due to invalid credentials, but should still return models
	if err != nil {
		t.Logf("Expected error due to invalid credentials: %v", err)
	}

	// Should fall back to config models
	if len(models) == 0 {
		t.Errorf("Expected fallback to config models, but got empty list")
	} else {
		t.Logf("Fallback successful, found %d models:", len(models))
		for i, model := range models {
			t.Logf("  Model %d: Name=%s, Version=%s", i+1, model.Name, model.Version)
		}
	}

	// Verify that we get the expected models from config
	expectedModels := map[string]string{
		"test-model-1": "2023-06-01",
		"test-model-2": "2023-06-01",
	}

	for _, model := range models {
		if expectedVersion, exists := expectedModels[model.Name]; exists {
			if model.Version != expectedVersion {
				t.Errorf("Expected version %s for model %s, got %s", expectedVersion, model.Name, model.Version)
			}
		}
	}
}
