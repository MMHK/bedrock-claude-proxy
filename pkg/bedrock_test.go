package pkg

import (
	"bedrock-claude-proxy/tests"
	_ "bedrock-claude-proxy/tests"
	"encoding/json"
	"testing"
)

func GetBedrockTestConfig() *BedrockConfig {
	return LoadBedrockConfigWithEnv()
}

func TestBedrockClient_CompleteTextWithStream(t *testing.T) {
	config := GetBedrockTestConfig()

	client := NewBedrockClient(config)

	prompt := "創作一首7言律詩"

	resp, err := client.CompleteText(&ClaudeTextCompletionRequest{
		Prompt:            prompt,
		Temperature:       0.5,
		MaxTokensToSample: 2048,
		Stream:            true,
		Model:             "anthropic.claude-v2:1",
	})

	if err != nil {
		t.Fatal(err)
		return
	}

	if resp.IsStream() {
		buffer := ""
		for event := range resp.GetEvents() {
			t.Log(tests.ToJSON(event))
			if event.GetEvent() == "completion" {
				buffer += event.GetText()
			}
		}
		t.Log(buffer)
	} else {
		t.Logf("%+v", resp.GetResponse())
	}
	t.Log("PASS")
}

func TestBedrockClient_CompleteTextWithoutStream(t *testing.T) {
	config := GetBedrockTestConfig()

	client := NewBedrockClient(config)

	prompt := "創作一首7言律詩"

	resp, err := client.CompleteText(&ClaudeTextCompletionRequest{
		Prompt:            prompt,
		Temperature:       0.5,
		MaxTokensToSample: 2048,
		Stream:            false,
		Model:             "anthropic.claude-v2:1",
	})

	if err != nil {
		t.Fatal(err)
		return
	}

	if resp.IsStream() {
		for event := range resp.GetEvents() {
			t.Logf("%+v", event)
		}
	} else {
		t.Logf("%+v", resp.GetResponse())
	}
	t.Log("PASS")
}

func TestBedrockClient_MessageCompletionWithoutStream(t *testing.T) {
	config := GetBedrockTestConfig()

	client := NewBedrockClient(config)

	prompt := "創作一首7言律詩"

	resp, err := client.MessageCompletion(&ClaudeMessageCompletionRequest{
		Temperature:      0.5,
		Stream:           false,
		Model:            "anthropic.claude-v2:1",
		MaxToken:         2048,
		System:           "You are a helpful assistant.",
		AnthropicVersion: "bedrock-2023-05-31",
		Messages: []*ClaudeMessageCompletionRequestMessage{
			&ClaudeMessageCompletionRequestMessage{
				Role: "user",
				Content: []*ClaudeMessageCompletionRequestContent{
					&ClaudeMessageCompletionRequestContent{
						Type: "text",
						Text: prompt,
					},
				},
			},
		},
	})

	if err != nil {
		t.Fatal(err)
		return
	}

	if resp.IsStream() {
		for event := range resp.GetEvents() {
			t.Logf("%+v", event)
		}
	} else {
		t.Log(tests.ToJSON(resp.GetResponse()))
	}
	t.Log("PASS")
}

func TestBedrockClient_MessageCompletionWithStream(t *testing.T) {
	config := GetBedrockTestConfig()

	client := NewBedrockClient(config)

	prompt := "創作一首7言律詩"

	resp, err := client.MessageCompletion(&ClaudeMessageCompletionRequest{
		Temperature:      0.5,
		Stream:           true,
		Model:            "anthropic.claude-v2:1",
		MaxToken:         2048,
		System:           "You are a helpful assistant.",
		AnthropicVersion: "bedrock-2023-05-31",
		Messages: []*ClaudeMessageCompletionRequestMessage{
			&ClaudeMessageCompletionRequestMessage{
				Role: "user",
				Content: []*ClaudeMessageCompletionRequestContent{
					&ClaudeMessageCompletionRequestContent{
						Type: "text",
						Text: prompt,
					},
				},
			},
		},
	})

	if err != nil {
		t.Fatal(err)
		return
	}

	if resp.IsStream() {
		buffer := ""
		for event := range resp.GetEvents() {
			t.Log(tests.ToJSON(event))
			if event.GetEvent() == "content_block_delta" {
				buffer += event.GetText()
			}
		}
		t.Log(buffer)
	} else {
		t.Log(tests.ToJSON(resp.GetResponse()))
	}
	t.Log("PASS")
}

func TestClaudeMessageCompletionRequest_UnmarshalJSON(t *testing.T) {
	raw := []byte("{\n    \"model\": \"claude-3-5-sonnet-20240620\",\n    \"max_tokens\": 1024,\n    \"messages\": [\n        {\"role\": \"user\", \"content\": \"Hello, world\"}\n    ]\n}")

	req := ClaudeMessageCompletionRequest{}
	err := json.Unmarshal(raw, &req)
	if err != nil {
		t.Fatal(err)
		return
	}

	t.Log(tests.ToJSON(req))
}
