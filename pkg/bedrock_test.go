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