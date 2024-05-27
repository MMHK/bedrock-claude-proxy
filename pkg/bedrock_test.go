package pkg

import "testing"

func GetBedrockTestConfig() (*BedrockConfig) {
	return LoadBedrockConfigWithEnv()
}

func TestBedrockClient_CompleteText(t *testing.T) {
	config := GetBedrockTestConfig()

	client := NewBedrockClient(config)

	events, err := client.CompleteText("anthropic.claude-v2:1", &ClaudeTextCompletionRequest{
		Prompt: "創作一首7言律詩",
		Temperature: 0.5,
		MaxTokensToSample: 2048,
		Stream: false,
	})

	if err != nil {
		t.Fatal(err)
		return
	}

	t.Logf("%+v", events.Completion())
	t.Log("PASS")
}
