package pkg

import (
	"bedrock-claude-proxy/tests"
	_ "bedrock-claude-proxy/tests"
	"bytes"
	"encoding/json"
	"io"
	"net/http/httptest"
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

	//t.Log(tests.ToJSON(config))

	config.DEBUG = true

	client := NewBedrockClient(config)

	prompt := "創作一首7言律詩"

	bin, err := json.Marshal([]*ClaudeMessageCompletionRequestContent{
		&ClaudeMessageCompletionRequestContent{
			Type: "text",
			Text: prompt,
		},
	})
	if err != nil {
		t.Fatal(err)
		return
	}

	resp, err := client.MessageCompletion(&ClaudeMessageCompletionRequest{
		Temperature:      0.5,
		TopP:             1,
		TopK:             5,
		Stream:           false,
		Model:            "claude-3-haiku-20240307",
		MaxToken:         2048,
		System:           "You are a helpful assistant.",
		AnthropicVersion: "bedrock-2023-05-31",
		Messages: []*ClaudeMessageCompletionRequestMessage{
			&ClaudeMessageCompletionRequestMessage{
				Role:    "user",
				Content: bin,
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

	bin, err := json.Marshal([]*ClaudeMessageCompletionRequestContent{
		&ClaudeMessageCompletionRequestContent{
			Type: "text",
			Text: prompt,
		},
	})
	if err != nil {
		t.Fatal(err)
		return
	}

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
				Content: bin,
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
	//raw := []byte("{\n    \"model\": \"claude-3-5-sonnet-20240620\",\n    \"max_tokens\": 1024,\n    \"messages\": [\n        {\"role\": \"user\", \"content\": \"Hello, world\"}\n    ]\n}")
	raw := []byte("{\"messages\":[{\"role\":\"user\",\"content\":\"This is a summary of the chat history as a recap: \\n\\n```json\\n{\\n  \\\"error\\\":{\\n    \\\"code\\\":\\\"DeploymentNotFound\\\",\\n    \\\"message\\\":\\\"The API deployment for this resource does not exist. If you created the deployment within the last 5 minutes, please wait a moment and try again.\\\"\\n  }\\n}\\n```\"},{\"role\":\"assistant\",\"content\":\";\"},{\"role\":\"user\",\"content\":\"你可以上网找资料，请比较 openai o1 模型与 gpt-4o 模型的区别\"},{\"role\":\"assistant\",\"content\":[{\"type\":\"tool_use\",\"id\":\"toolu_bdrk_017hPYarP3A6ka71MHHZmd5f\",\"name\":\"searchWeb\",\"input\":{\"q\":\"OpenAI GPT-4 vs GPT-4 Turbo (GPT-4-128k) differences\",\"format\":\"json\"}}]},{\"role\":\"user\",\"content\":[{\"type\":\"tool_result\",\"tool_use_id\":\"toolu_bdrk_017hPYarP3A6ka71MHHZmd5f\",\"content\":\"{\\\"query\\\":\\\"OpenAI GPT-4 vs GPT-4 Turbo (GPT-4-128k) differences\\\",\\\"number_of_results\\\":0,\\\"results\\\":[{\\\"url\\\":\\\"https://help.openai.com/en/articles/7127966-what-is-the-difference-between-the-gpt-4-model-versions\\\",\\\"title\\\":\\\"What is the difference between the GPT-4 model versions?\\\",\\\"content\\\":\\\"Context window (some models have as low as an 8k context window while some have an 128k context window) Knowledge cutoff (some models have been training on more up to date information which makes them better at certain tasks) Cost (the cost for models vary, our latest GPT-4 Turbo model is less expensive than previous GPT-4 model variants, you ...\\\",\\\"thumbnail\\\":\\\"\\\",\\\"engine\\\":\\\"google\\\",\\\"parsed_url\\\":[\\\"https\\\",\\\"help.openai.com\\\",\\\"/en/articles/7127966-what-is-the-difference-between-the-gpt-4-model-versions\\\",\\\"\\\",\\\"\\\",\\\"\\\"],\\\"template\\\":\\\"default.html\\\",\\\"engines\\\":[\\\"duckduckgo\\\",\\\"google\\\",\\\"brave\\\",\\\"qwant\\\"],\\\"positions\\\":[2,1,1,1],\\\"publishedDate\\\":null,\\\"score\\\":14,\\\"category\\\":\\\"general\\\"},{\\\"url\\\":\\\"https://tech.co/news/gpt-4-turbo-vs-gpt-4-openai-chatgpt\\\",\\\"title\\\":\\\"GPT-4 Turbo vs GPT-4: What Is OpenAI's ChatGPT Turbo?\\\",\\\"content\\\":\\\"January 29, 2024 - However, the drop-down menu that ChatGPT Plus has been using to switch between other OpenAI apps like DALLE-3, is being retired. Now, ChatGPT will work out what sort of output you need based on your prompts. GPT-4 Turbo also has an enlarged 128K context window, which helps it take prompts ...\\\",\\\"thumbnail\\\":\\\"\\\",\\\"engine\\\":\\\"google\\\",\\\"parsed_url\\\":[\\\"https\\\",\\\"tech.co\\\",\\\"/news/gpt-4-turbo-vs-gpt-4-openai-chatgpt\\\",\\\"\\\",\\\"\\\",\\\"\\\"],\\\"template\\\":\\\"default.html\\\",\\\"engines\\\":[\\\"duckduckgo\\\",\\\"google\\\",\\\"brave\\\",\\\"qwant\\\"],\\\"positions\\\":[5,2,2,2],\\\"publishedDate\\\":\\\"2024-01-29T00:00:00\\\",\\\"score\\\":6.8,\\\"category\\\":\\\"general\\\"},{\\\"url\\\":\\\"https://community.openai.com/t/gpt-4-vs-gpt-4o-which-is-the-better/746991\\\",\\\"title\\\":\\\"GPT-4 vs GPT-4o? Which is the better?\\\",\\\"content\\\":\\\"GPT-4 architecture is similar to GPT-4 Turbo. GPT-4o is for high interaction rates that compromise a bit of precision. GPT-4 architecture rarely hallucinates, while GPT4o seems to have more of these moments. You also have to understand you're now talking to a different \\\\\\\"brain\\\\\\\", different neural networks.\\\",\\\"thumbnail\\\":\\\"\\\",\\\"engine\\\":\\\"google\\\",\\\"parsed_url\\\":[\\\"https\\\",\\\"community.openai.com\\\",\\\"/t/gpt-4-vs-gpt-4o-which-is-the-better/746991\\\",\\\"\\\",\\\"\\\",\\\"\\\"],\\\"template\\\":\\\"default.html\\\",\\\"engines\\\":[\\\"duckduckgo\\\",\\\"google\\\",\\\"brave\\\"],\\\"positions\\\":[1,6,26],\\\"publishedDate\\\":\\\"2024-05-14T00:00:00\\\",\\\"score\\\":3.6153846153846154,\\\"category\\\":\\\"general\\\"},{\\\"url\\\":\\\"https://www.theverge.com/2023/11/6/23948426/openai-gpt4-turbo-generative-ai-new-model\\\",\\\"title\\\":\\\"OpenAI turbocharges GPT-4 and makes it cheaper - The Verge\\\",\\\"content\\\":\\\"November 6, 2023 - OpenAI plans to release a production-ready Turbo model in the next few weeks but did not give an exact date. ... GPT-4 Turbo will also “see” more data, with a 128K context window, which OpenAI says is “equivalent to more than 300 pages of text in a single prompt.” Generally, larger ...\\\",\\\"publishedDate\\\":\\\"2023-11-06T00:00:00\\\",\\\"thumbnail\\\":\\\"\\\",\\\"engine\\\":\\\"brave\\\",\\\"parsed_url\\\":[\\\"https\\\",\\\"www.theverge.com\\\",\\\"/2023/11/6/23948426/openai-gpt4-turbo-generative-ai-new-model\\\",\\\"\\\",\\\"\\\",\\\"\\\"],\\\"template\\\":\\\"default.html\\\",\\\"engines\\\":[\\\"duckduckgo\\\",\\\"brave\\\",\\\"qwant\\\"],\\\"positions\\\":[9,10,10],\\\"score\\\":0.9333333333333333,\\\"category\\\":\\\"general\\\"},{\\\"url\\\":\\\"https://community.openai.com/t/my-honest-take-on-gpt-4o-vs-gpt-4-turbo-2024-04-09-vs-gpt-4-1106/882676\\\",\\\"title\\\":\\\"My Honest Take on GPT-4o vs GPT-4-turbo-2024-04-09 vs ...\\\",\\\"content\\\":\\\"After extensively using these three versions of GPT-4, I'll share my findings. Firstly, it's important to highlight some facts: The original GPT-4, released in March 2023, is version 0314. It underwent several updates in 2023, with the latest being GPT-4-1106, launched at the OpenAI DevDay. This version is notable for its knowledge of events up until 2023 and an increased context window of ...\\\",\\\"thumbnail\\\":null,\\\"engine\\\":\\\"google\\\",\\\"parsed_url\\\":[\\\"https\\\",\\\"community.openai.com\\\",\\\"/t/my-honest-take-on-gpt-4o-vs-gpt-4-turbo-2024-04-09-vs-gpt-4-1106/882676\\\",\\\"\\\",\\\"\\\",\\\"\\\"],\\\"template\\\":\\\"default.html\\\",\\\"engines\\\":[\\\"duckduckgo\\\",\\\"google\\\"],\\\"positions\\\":[4,17],\\\"score\\\":0.6176470588235294,\\\"category\\\":\\\"general\\\"},{\\\"url\\\":\\\"https://learn.microsoft.com/en-us/azure/ai-services/openai/concepts/models\\\",\\\"title\\\":\\\"Azure OpenAI Service models - Azure OpenAI | Microsoft Learn\\\",\\\"content\\\":\\\"Expand table. Models. Description. GPT-4o & GPT-4o mini & GPT-4 Turbo. The latest most capable Azure OpenAI models with multimodal versions, which can accept both text and images as input. GPT-4. A set of models that improve on GPT-3.5 and can understand and generate natural language and code. GPT-3.5. A set of models that improve on GPT-3 and ...\\\",\\\"publishedDate\\\":null,\\\"thumbnail\\\":\\\"\\\",\\\"engine\\\":\\\"brave\\\",\\\"parsed_url\\\":[\\\"https\\\",\\\"learn.microsoft.com\\\",\\\"/en-us/azure/ai-services/openai/concepts/models\\\",\\\"\\\",\\\"\\\",\\\"\\\"],\\\"template\\\":\\\"default.html\\\",\\\"engines\\\":[\\\"duckduckgo\\\",\\\"brave\\\"],\\\"positions\\\":[5,14],\\\"score\\\":0.5428571428571429,\\\"category\\\":\\\"general\\\"},{\\\"url\\\":\\\"https://textcortex.com/post/gpt-4-turbo-vs-gpt-4\\\",\\\"title\\\":\\\"GPT-4 Turbo vs. GPT-4 - What's the Difference?\\\",\\\"content\\\":\\\"2 weeks ago - GPT-4 Turbo is OpenAI's latest generation model, launched in November 2023, only a few months after the launch of GPT-4. It has a more extensive knowledge base up to April 2023, which can provide more current information. Furthermore, it has a 128k context window, which equals 300 pages of ...\\\",\\\"thumbnail\\\":\\\"\\\",\\\"engine\\\":\\\"google\\\",\\\"parsed_url\\\":[\\\"https\\\",\\\"textcortex.com\\\",\\\"/post/gpt-4-turbo-vs-gpt-4\\\",\\\"\\\",\\\"\\\",\\\"\\\"],\\\"template\\\":\\\"default.html\\\",\\\"engines\\\":[\\\"google\\\",\\\"brave\\\"],\\\"positions\\\":[7,19],\\\"publishedDate\\\":null,\\\"score\\\":0.39097744360902253,\\\"category\\\":\\\"general\\\"},{\\\"url\\\":\\\"https://community.openai.com/t/at-the-end-is-gpt-4-turbo-128k-or-4k/648759\\\",\\\"title\\\":\\\"At the end, is GPT-4-turbo 128k or 4k? - ChatGPT - OpenAI Developer Forum\\\",\\\"content\\\":\\\"OpenAI limited the amount you can get back on the latest models as a response. That's where the 4k comes from. The max_tokens setting on API sets the response length. The slider shows how much you are able to dedicate to just getting the response. witch model in real life can generate larger response gpt-4-turbo or gpt-4.\\\",\\\"publishedDate\\\":\\\"2024-02-22T00:00:00\\\",\\\"thumbnail\\\":\\\"\\\",\\\"engine\\\":\\\"brave\\\",\\\"parsed_url\\\":[\\\"https\\\",\\\"community.openai.com\\\",\\\"/t/at-the-end-is-gpt-4-turbo-128k-or-4k/648759\\\",\\\"\\\",\\\"\\\",\\\"\\\"],\\\"template\\\":\\\"default.html\\\",\\\"engines\\\":[\\\"duckduckgo\\\",\\\"brave\\\"],\\\"positions\\\":[8,23],\\\"score\\\":0.33695652173913043,\\\"category\\\":\\\"general\\\"},{\\\"url\\\":\\\"https://www.reddit.com/r/LargeLanguageModels/comments/186kbyu/gpt4_vs_gpt4128k/\\\",\\\"title\\\":\\\"GPT-4 vs. GPT-4-128K? : r/LargeLanguageModels\\\",\\\"content\\\":\\\"I am wondering what are differences between those two models. What makes GPT-4-128K to be able to handle 128K tokens? Are there any available ...\\\",\\\"thumbnail\\\":null,\\\"engine\\\":\\\"google\\\",\\\"parsed_url\\\":[\\\"https\\\",\\\"www.reddit.com\\\",\\\"/r/LargeLanguageModels/comments/186kbyu/gpt4_vs_gpt4128k/\\\",\\\"\\\",\\\"\\\",\\\"\\\"],\\\"template\\\":\\\"default.html\\\",\\\"engines\\\":[\\\"google\\\"],\\\"positions\\\":[3],\\\"score\\\":0.3333333333333333,\\\"category\\\":\\\"general\\\"},{\\\"title\\\":\\\"What's the difference between GPT-3.5, 4, 4 Turbo, 4o? OpenAI LLMs ...\\\",\\\"url\\\":\\\"https://www.windowscentral.com/software-apps/windows-11/whats-the-difference-between-gpt-35-4-4-turbo-4o\\\",\\\"content\\\":\\\"GPT-4o is the latest version of the language model from OpenAI, which became available in May 2024. It's important to note that this is GPT-4\\\\\\\" o,\\\\\\\" not GPT-4 \\\\\\\"0\\\\\\\" or \\\\\\\"4.0.\\\\\\\" It's an \\\\\\\"o\\\\\\\" for \\\\\\\"omni ...\\\",\\\"engine\\\":\\\"qwant\\\",\\\"parsed_url\\\":[\\\"https\\\",\\\"www.windowscentral.com\\\",\\\"/software-apps/windows-11/whats-the-difference-between-gpt-35-4-4-turbo-4o\\\",\\\"\\\",\\\"\\\",\\\"\\\"],\\\"template\\\":\\\"default.html\\\",\\\"engines\\\":[\\\"duckduckgo\\\",\\\"qwant\\\"],\\\"positions\\\":[3,3],\\\"score\\\":1.3333333333333333,\\\"category\\\":\\\"general\\\"},{\\\"title\\\":\\\"GPT-4 Turbo in the OpenAI API - OpenAI Help Center\\\",\\\"url\\\":\\\"https://help.openai.com/en/articles/8555510-gpt-4-turbo-in-the-openai-api\\\",\\\"content\\\":\\\"GPT-4 Turbo is our latest generation model. It’s more capable, has an updated knowledge cutoff of April 2023 and introduces a 128k context window (the equivalent of 300 pages of text in a single prompt). The model is also 3X cheaper for input tokens and 2X cheaper for output tokens compared to the original GPT-4 model.\\\",\\\"engine\\\":\\\"qwant\\\",\\\"parsed_url\\\":[\\\"https\\\",\\\"help.openai.com\\\",\\\"/en/articles/8555510-gpt-4-turbo-in-the-openai-api\\\",\\\"\\\",\\\"\\\",\\\"\\\"],\\\"template\\\":\\\"default.html\\\",\\\"engines\\\":[\\\"duckduckgo\\\",\\\"qwant\\\"],\\\"positions\\\":[4,4],\\\"score\\\":1,\\\"category\\\":\\\"general\\\"},{\\\"title\\\":\\\"OpenAI Platform\\\",\\\"url\\\":\\\"https://platform.openai.com/docs/models/gpt-4-and-gpt-4-turbo\\\",\\\"content\\\":\\\"Explore resources, tutorials, API docs, and dynamic examples to get the most out of OpenAI's developer platform.\\\",\\\"engine\\\":\\\"qwant\\\",\\\"parsed_url\\\":[\\\"https\\\",\\\"platform.openai.com\\\",\\\"/docs/models/gpt-4-and-gpt-4-turbo\\\",\\\"\\\",\\\"\\\",\\\"\\\"],\\\"template\\\":\\\"default.html\\\",\\\"engines\\\":[\\\"duckduckgo\\\",\\\"qwant\\\"],\\\"positions\\\":[5,5],\\\"score\\\":0.8,\\\"category\\\":\\\"general\\\"},{\\\"title\\\":\\\"New models and developer products announced at DevDay - OpenAI\\\",\\\"url\\\":\\\"https://openai.com/index/new-models-and-developer-products-announced-at-devday/\\\",\\\"content\\\":\\\"New models and developer products announced at DevDay. GPT-4 Turbo with 128K context and lower prices, the new Assistants API, GPT-4 Turbo with Vision, DALL·E 3 API, and more. Update: We previously stated that applications using the _gpt-3.5-turbo_ name will automatically be upgraded to the new model version on December 11.\\\",\\\"engine\\\":\\\"qwant\\\",\\\"parsed_url\\\":[\\\"https\\\",\\\"openai.com\\\",\\\"/index/new-models-and-developer-products-announced-at-devday/\\\",\\\"\\\",\\\"\\\",\\\"\\\"],\\\"template\\\":\\\"default.html\\\",\\\"engines\\\":[\\\"duckduckgo\\\",\\\"qwant\\\"],\\\"positions\\\":[6,7],\\\"score\\\":0.6190476190476191,\\\"category\\\":\\\"general\\\"},{\\\"title\\\":\\\"GPT-4 | OpenAI\\\",\\\"url\\\":\\\"https://openai.com/index/gpt-4/\\\",\\\"content\\\":\\\"More on GPT-4. Research GPT-4 is the latest milestone in OpenAI’s effort in scaling up deep learning. View GPT-4 research. Infrastructure GPT-4 was trained on Microsoft Azure AI supercomputers. Azure’s AI-optimized infrastructure also allows us to deliver GPT-4 to users around the world.\\\",\\\"engine\\\":\\\"qwant\\\",\\\"parsed_url\\\":[\\\"https\\\",\\\"openai.com\\\",\\\"/index/gpt-4/\\\",\\\"\\\",\\\"\\\",\\\"\\\"],\\\"template\\\":\\\"default.html\\\",\\\"engines\\\":[\\\"duckduckgo\\\",\\\"qwant\\\"],\\\"positions\\\":[7,6],\\\"score\\\":0.6190476190476191,\\\"category\\\":\\\"general\\\"},{\\\"title\\\":\\\"GPT-4 - OpenAI\\\",\\\"url\\\":\\\"https://openai.com/index/gpt-4-research/\\\",\\\"content\\\":\\\"We’ve created GPT-4, the latest milestone in OpenAI’s effort in scaling up deep learning. GPT-4 is a large multimodal model (accepting image and text inputs, emitting text outputs) that, while less capable than humans in many real-world scenarios, exhibits human-level performance on various professional and academic benchmarks.\\\",\\\"engine\\\":\\\"qwant\\\",\\\"parsed_url\\\":[\\\"https\\\",\\\"openai.com\\\",\\\"/index/gpt-4-research/\\\",\\\"\\\",\\\"\\\",\\\"\\\"],\\\"template\\\":\\\"default.html\\\",\\\"engines\\\":[\\\"duckduckgo\\\",\\\"qwant\\\"],\\\"positions\\\":[8,9],\\\"score\\\":0.4722222222222222,\\\"category\\\":\\\"general\\\"},{\\\"title\\\":\\\"OpenAI Announce GPT-4 Turbo With Vision: What We Know So Far\\\",\\\"url\\\":\\\"https://www.datacamp.com/blog/gpt4-turbo\\\",\\\"content\\\":\\\"GPT-4 Turbo is an update to the existing GPT-4 large language model. It brings several improvements, including a greatly increased context window and access to more up-to-date knowledge. OpenAI has gradually been improving the capabilities of GPT-4 in ChatGPT with the addition of custom instructions, ChatGPT plugins, DALL-E 3, and Advanced Data ...\\\",\\\"engine\\\":\\\"qwant\\\",\\\"parsed_url\\\":[\\\"https\\\",\\\"www.datacamp.com\\\",\\\"/blog/gpt4-turbo\\\",\\\"\\\",\\\"\\\",\\\"\\\"],\\\"template\\\":\\\"default.html\\\",\\\"engines\\\":[\\\"duckduckgo\\\",\\\"qwant\\\"],\\\"positions\\\":[9,8],\\\"score\\\":0.4722222222222222,\\\"category\\\":\\\"general\\\"},{\\\"title\\\":\\\"Azure OpenAI Service Launches GPT-4 Turbo and GPT-3.5-Turbo-1106 Models ...\\\",\\\"content\\\":\\\"GPT-3.5 Turbo 1106 is generally available to all Azure OpenAI customers immediately. GPT-3.5 Turbo pricing is 3x most cost effective for input tokens and 2x more cost effective for output tokens compared to GPT-3.5 Turbo 16k. To deploy GPT-3.5-Turbo 1106 from the Studio UI, select \\\\\\\"gpt-35-turbo\\\\\\\" and then select version \\\\\\\"1106\\\\\\\" from the dropdown.\\\",\\\"url\\\":\\\"https://techcommunity.microsoft.com/t5/ai-azure-ai-services-blog/azure-openai-service-launches-gpt-4-turbo-and-gpt-3-5-turbo-1106/ba-p/3985962\\\",\\\"engine\\\":\\\"duckduckgo\\\",\\\"parsed_url\\\":[\\\"https\\\",\\\"techcommunity.microsoft.com\\\",\\\"/t5/ai-azure-ai-services-blog/azure-openai-service-launches-gpt-4-turbo-and-gpt-3-5-turbo-1106/ba-p/3985962\\\",\\\"\\\",\\\"\\\",\\\"\\\"],\\\"template\\\":\\\"default.html\\\",\\\"engines\\\":[\\\"duckduckgo\\\"],\\\"positions\\\":[13],\\\"score\\\":0.07692307692307693,\\\"category\\\":\\\"general\\\"},{\\\"title\\\":\\\"GPT Base, GPT-3.5 Turbo & GPT-4: What's the difference? | Pluralsight\\\",\\\"content\\\":\\\"Able to do complex tasks, but slower at giving answers. Currently used by ChatGPT Plus. GPT-3.5. Faster than GPT-4 and more flexible than GPT Base. The \\\\\\\"good enough\\\\\\\" model series for most tasks, whether chat or general. GPT-3.5 Turbo. The best model in the GPT-3.5 series. Currently used by the free version of ChatGPT. Cost effective and ...\\\",\\\"url\\\":\\\"https://www.pluralsight.com/resources/blog/ai-and-data/ai-gpt-models-differences\\\",\\\"engine\\\":\\\"duckduckgo\\\",\\\"parsed_url\\\":[\\\"https\\\",\\\"www.pluralsight.com\\\",\\\"/resources/blog/ai-and-data/ai-gpt-models-differences\\\",\\\"\\\",\\\"\\\",\\\"\\\"],\\\"template\\\":\\\"default.html\\\",\\\"engines\\\":[\\\"duckduckgo\\\"],\\\"positions\\\":[16],\\\"score\\\":0.0625,\\\"category\\\":\\\"general\\\"},{\\\"url\\\":\\\"https://help.openai.com/en/articles/7102672-how-can-i-access-gpt-4-gpt-4-turbo-gpt-4o-and-gpt-4o-mini\\\",\\\"title\\\":\\\"How can I access GPT-4, GPT-4 Turbo, GPT-4o, and GPT- ...\\\",\\\"content\\\":\\\"You can try sending us a message or logging in at https://beta.openai.com/\\\",\\\"publishedDate\\\":null,\\\"thumbnail\\\":\\\"\\\",\\\"engine\\\":\\\"brave\\\",\\\"parsed_url\\\":[\\\"https\\\",\\\"help.openai.com\\\",\\\"/en/articles/7102672-how-can-i-access-gpt-4-gpt-4-turbo-gpt-4o-and-gpt-4o-mini\\\",\\\"\\\",\\\"\\\",\\\"\\\"],\\\"template\\\":\\\"default.html\\\",\\\"engines\\\":[\\\"brave\\\"],\\\"positions\\\":[3],\\\"score\\\":0.3333333333333333,\\\"category\\\":\\\"general\\\"},{\\\"url\\\":\\\"https://community.openai.com/t/gpt-4-vs-gpt-4-turbo-preview/693031\\\",\\\"title\\\":\\\"Gpt-4 vs gpt-4-turbo-preview - Community - OpenAI Developer Forum\\\",\\\"content\\\":\\\"Some people (myself included) believe this: GPT-4-0314 is the best of the GPT-4 series. It performs the best at solving abstract problems. GPT-4-turbo (1106, 0125) struggles more with logic reasoning. However, GPT-4-0314 is more prone to hallucinations, and also more expensive. GPT-4-turbo, on the other hand, is much more stable, and less ...\\\",\\\"publishedDate\\\":\\\"2024-03-21T00:00:00\\\",\\\"thumbnail\\\":\\\"\\\",\\\"engine\\\":\\\"brave\\\",\\\"parsed_url\\\":[\\\"https\\\",\\\"community.openai.com\\\",\\\"/t/gpt-4-vs-gpt-4-turbo-preview/693031\\\",\\\"\\\",\\\"\\\",\\\"\\\"],\\\"template\\\":\\\"default.html\\\",\\\"engines\\\":[\\\"duckduckgo\\\",\\\"brave\\\"],\\\"positions\\\":[12,12],\\\"score\\\":0.2024-09-24T06:47:44.706861519Z 3333333333333333,\\\"category\\\":\\\"general\\\"},{\\\"url\\\":\\\"https://arstechnica.com/information-technology/2023/11/openai-introduces-gpt-4-turbo-larger-memory-lower-cost-new-knowledge/\\\",\\\"title\\\":\\\"OpenAI introduces GPT-4 Turbo: Larger memory, lower cost, new knowledge | Ars Technica\\\",\\\"content\\\":\\\"November 6, 2023 - According to OpenAI, one token corresponds roughly to about four characters of English text, or about three-quarters of a word. That means GPT-4 Turbo can consider around 96,000 words in one go, which is longer than many novels. Also, a 128K context length can lead to much longer conversations ...\\\",\\\"publishedDate\\\":\\\"2023-11-06T00:00:00\\\",\\\"thumbnail\\\":\\\"\\\",\\\"engine\\\":\\\"brave\\\",\\\"parsed_url\\\":[\\\"https\\\",\\\"arstechnica.com\\\",\\\"/information-technology/2023/11/openai-introduces-gpt-4-turbo-larger-memory-lower-cost-new-knowledge/\\\",\\\"\\\",\\\"\\\",\\\"\\\"],\\\"template\\\":\\\"default.html\\\",\\\"engines\\\":[\\\"duckduckgo\\\",\\\"brave\\\"],\\\"positions\\\":[17,11],\\\"score\\\":0.2994652406417112,\\\"category\\\":\\\"general\\\"},{\\\"url\\\":\\\"https://community.openai.com/t/gpt-4-turbo-and-gpt-4-o-benchmarks-released-they-do-well-compared-to-the-marketplace/744528\\\",\\\"title\\\":\\\"GPT-4-Turbo and GPT-4-O benchmarks released! They do well compared to the marketplace - Community - OpenAI Developer Forum\\\",\\\"content\\\":\\\"Hi all, I'm happy to say that the benchmarks on the gpt-4-turbo and gpt-4-o models were finally released by OpenAI and they both do pretty well. openai/simple-evals (github.com) Additionally, we have results on the LM… Hi all, I'm happy to say that the benchmarks on the gpt-4-turbo and gpt-4-o models were finally released by OpenAI and ...\\\",\\\"publishedDate\\\":\\\"2024-05-13T00:00:00\\\",\\\"thumbnail\\\":\\\"\\\",\\\"engine\\\":\\\"brave\\\",\\\"parsed_url\\\":[\\\"https\\\",\\\"community.openai.com\\\",\\\"/t/gpt-4-turbo-and-gpt-4-o-benchmarks-released-they-do-well-compared-to-the-marketplace/744528\\\",\\\"\\\",\\\"\\\",\\\"\\\"],\\\"template\\\":\\\"default.html\\\",\\\"engines\\\":[\\\"duckduckgo\\\",\\\"brave\\\"],\\\"positions\\\":[16,15],\\\"score\\\":0.2583333333333333,\\\"category\\\":\\\"general\\\"},{\\\"url\\\":\\\"https://writesonic.com/blog/gpt-4-turbo-vs-gpt-4\\\",\\\"title\\\":\\\"GPT-4 Turbo vs GPT-4: How Does OpenAI Turbo Charge GPT-4?\\\",\\\"content\\\":\\\"May 21, 2024 - GPT-4 Turbo is OpenAI's latest generation model, more capable than its predecessor, with an updated knowledge cutoff of April 2023. It introduces a 128k context window and is more cost-effective, with input tokens being 3X cheaper and output tokens 2X cheaper compared to the original GPT-4 model.\\\",\\\"publishedDate\\\":\\\"2024-05-21T00:00:00\\\",\\\"thumbnail\\\":\\\"\\\",\\\"engine\\\":\\\"brave\\\",\\\"parsed_url\\\":[\\\"https\\\",\\\"writesonic.com\\\",\\\"/blog/gpt-4-turbo-vs-gpt-4\\\",\\\"\\\",\\\"\\\",\\\"\\\"],\\\"template\\\":\\\"default.html\\\",\\\"engines\\\":[\\\"brave\\\"],\\\"positions\\\":[4],\\\"score\\\":0.25,\\\"category\\\":\\\"general\\\"},{\\\"url\\\":\\\"https://community.openai.com/t/difference-between-old-and-new-model/580481\\\",\\\"title\\\":\\\"Difference between old and new model - API\\\",\\\"content\\\":\\\"Jan 10, 2024 — Yes the models contain different fine tuning, different context sizes 4, 8 , 16, 32, 128 for example.\\\",\\\"thumbnail\\\":null,\\\"engine\\\":\\\"google\\\",\\\"parsed_url\\\":[\\\"https\\\",\\\"community.openai.com\\\",\\\"/t/difference-between-old-and-new-model/580481\\\",\\\"\\\",\\\"\\\",\\\"\\\"],\\\"template\\\":\\\"default.html\\\",\\\"engines\\\":[\\\"google\\\"],\\\"positions\\\":[6],\\\"score\\\":0.16666666666666666,\\\"category\\\":\\\"general\\\"},{\\\"url\\\":\\\"https://help.openai.com/en/articles/7102672-how-can-i-access-gpt-4-gpt-4-turbo-and-gpt-4o\\\",\\\"title\\\":\\\"How can I access GPT-4, GPT-4 Turbo and GPT-4o?\\\",\\\"content\\\":\\\"You can try sending us a message or logging in at https://beta.openai.com/\\\",\\\"publishedDate\\\":null,\\\"thumbnail\\\":\\\"\\\",\\\"engine\\\":\\\"brave\\\",\\\"parsed_url\\\":[\\\"https\\\",\\\"help.openai.com\\\",\\\"/en/articles/7102672-how-can-i-access-gpt-4-gpt-4-turbo-and-gpt-4o\\\",\\\"\\\",\\\"\\\",\\\"\\\"],\\\"template\\\":\\\"default.html\\\",\\\"engines\\\":[\\\"brave\\\"],\\\"positions\\\":[7],\\\"score\\\":0.14285714285714285,\\\"category\\\":\\\"general\\\"},{\\\"url\\\":\\\"https://community.openai.com/t/what-are-the-additional-risks-using-gpt-4-turbo-preview-over-gpt-4/705367\\\",\\\"title\\\":\\\"What are the additional risks using gpt-4-turbo-preview ...\\\",\\\"content\\\":\\\"Apr 2, 2024 — Potential Risks: OpenAI has outlined potential risks associated with using GPT-4 models, which include generating harmful content, societal ...\\\",\\\"thumbnail\\\":null,\\\"engine\\\":\\\"google\\\",\\\"parsed_url\\\":[\\\"https\\\",\\\"community.openai.com\\\",\\\"/t/what-are-the-additional-risks-using-gpt-4-turbo-preview-over-gpt-4/705367\\\",\\\"\\\",\\\"\\\",\\\"\\\"],\\\"template\\\":\\\"default.html\\\",\\\"engines\\\":[\\\"google\\\"],\\\"positions\\\":[8],\\\"score\\\":0.125,\\\"category\\\":\\\"general\\\"},{\\\"url\\\":\\\"https://www.vellum.ai/blog/analysis-gpt-4o-vs-gpt-4-turbo\\\",\\\"title\\\":\\\"Analysis: GPT-4o vs GPT-4 Turbo\\\",\\\"content\\\":\\\"Learn how GPT4o compares to GPT-4 Turbo on classification, reasoning and data extraction tasks.\\\",\\\"publishedDate\\\":null,\\\"thumbnail\\\":\\\"\\\",\\\"engine\\\":\\\"brave\\\",\\\"parsed_url\\\":[\\\"https\\\",\\\"www.vellum.ai\\\",\\\"/blog/analysis-gpt-4o-vs-gpt-4-turbo\\\",\\\"\\\",\\\"\\\",\\\"\\\"],\\\"template\\\":\\\"default.html\\\",\\\"engines\\\":[\\\"brave\\\"],\\\"positions\\\":[10],\\\"score\\\":0.1,\\\"category\\\":\\\"general\\\"},{\\\"url\\\":\\\"https://www.techtarget.com/searchenterpriseai/feature/GPT-4o-vs-GPT-4-How-do-they-compare\\\",\\\"title\\\":\\\"GPT-4o vs. GPT-4: How do they compare? | TechTarget\\\",\\\"content\\\":\\\"Explore the key differences between OpenAI's GPT-4 and GPT-4o, including enhancements in multimodal capabilities, performance and cost effectiveness.\\\",\\\"publishedDate\\\":null,\\\"thumbnail\\\":\\\"https://imgs.search.brave.com/7QzccOEgokp_qijZ_kxwkJVgeWMhI7WpBkZpwiYeTAk/rs:fit:200:200:1:0/g:ce/aHR0cHM6Ly9pLnl0/aW1nLmNvbS92aS93/VlBKQ1ZwamF5NC9t/cWRlZmF1bHQuanBn\\\",\\\"engine\\\":\\\"brave\\\",\\\"parsed_url\\\":[\\\"https\\\",\\\"www.techtarget.com\\\",\\\"/searchenterpriseai/feature/GPT-4o-vs-GPT-4-How-do-they-compare\\\",\\\"\\\",\\\"\\\",\\\"\\\"],\\\"template\\\":\\\"default.html\\\",\\\"engines\\\":[\\\"brave\\\"],\\\"positions\\\":[11],\\\"score\\\":0.09090909090909091,\\\"category\\\":\\\"general\\\"},{\\\"url\\\":\\\"https://community.openai.com/t/dev-guide-when-to-use-gpt-4o-vs-turbo/744750\\\",\\\"title\\\":\\\"Dev Guide: when to use GPT-4o vs Turbo? - API - OpenAI Developer Forum\\\",\\\"content\\\":\\\"May 13, 2024 - With 2x the speed and 1/2 cost of Turbo and higher scores on synthetic benchmarks and the chatbot arena ELO, it would seem appropriate to stop using Turbo altogether and just migrate to ‘o’. What’s the expectations for developers? I think it would be great if OpenAI produced a guide on ...\\\",\\\"publishedDate\\\":\\\"2024-05-13T00:00:00\\\",\\\"thumbnail\\\":\\\"\\\",\\\"engine\\\":\\\"brave\\\",\\\"parsed_url\\\":[\\\"https\\\",\\\"community.openai.com\\\",\\\"/t/dev-guide-when-to-use-gpt-4o-vs-turbo/744750\\\",\\\"\\\",\\\"\\\",\\\"\\\"],\\\"template\\\":\\\"default.html\\\",\\\"engines\\\":[\\\"brave\\\"],\\\"positions\\\":[13],\\\"score\\\":0.07692307692307693,\\\"category\\\":\\\"general\\\"},{\\\"url\\\":\\\"https://community.openai.com/t/gpt-4o-vs-gpt-4-turbo-2024-04-09-gpt-4o-loses/764328\\\",\\\"title\\\":\\\"GPT-4o vs. gpt-4-turbo-2024-04-09, gpt-4o loses - API - OpenAI Developer Forum\\\",\\\"content\\\":\\\"May 19, 2024 - Hi, We recently switched to using gpt-4o instead of gpt-4-turbo-2024-04-09, but the prompt working perfectly well with gpt-4-turbo-2024-04-09 doesn’t work with gpt-4o. It simply doesn’t follow the instructions properly. On the other hand gpt-4-turbo-2024-04-09 is stable, works perfectly well.\\\",\\\"publishedDate\\\":\\\"2024-05-19T00:00:00\\\",\\\"thumbnail\\\":\\\"\\\",\\\"engine\\\":\\\"brave\\\",\\\"parsed_url\\\":[\\\"https\\\",\\\"community.openai.com\\\",\\\"/t/gpt-4o-vs-gpt-4-turbo-2024-04-09-gpt-4o-loses/764328\\\",\\\"\\\",\\\"\\\",\\\"\\\"],\\\"template\\\":\\\"default.html\\\",\\\"engines\\\":[\\\"brave\\\"],\\\"positions\\\":[14],\\\"score\\\":0.07142857142857142,\\\"category\\\":\\\"general\\\"},{\\\"url\\\":\\\"https://community.openai.com/t/gpt-4-32k-vs-gpt-4-turbo-api-legal-advice-on-using/659991\\\",\\\"title\\\":\\\"Gpt-4 32k vs GPT-4 Turbo api + Legal advice on using - API - OpenAI Developer Forum\\\",\\\"content\\\":\\\"February 29, 2024 - We want to use gpt-4 to generate Question & Answers from a book series owned entirely by our Department to use as FAQ in our ChatBot , it has roughly 6 million tokens , and after prompt engineering our prompt is roughly 1000 tokens at bare minimum . the limits for these gpt4-32k & gpt4-turbo ...\\\",\\\"publishedDate\\\":\\\"2024-02-29T00:00:00\\\",\\\"thumbnail\\\":\\\"\\\",\\\"engine\\\":\\\"brave\\\",\\\"parsed_url\\\":[\\\"https\\\",\\\"community.openai.com\\\",\\\"/t/gpt-4-32k-vs-gpt-4-turbo-api-legal-advice-on-using/659991\\\",\\\"\\\",\\\"\\\",\\\"\\\"],\\\"template\\\":\\\"default.html\\\",\\\"engines\\\":[\\\"brave\\\"],\\\"positions\\\":[15],\\\"score\\\":0.06666666666666667,\\\"category\\\":\\\"general\\\"},{\\\"url\\\":\\\"https://textcortex.com/post/gpt-4o-vs-gpt-4\\\",\\\"title\\\":\\\"GPT-4o vs GPT-4: Which Model is Better?\\\",\\\"content\\\":\\\"2 weeks ago - See how OpenAI's GPT-4o is outpacing GPT-4 when compared on different benchmarks and learn all the new features GPT-4o brings with!\\\",\\\"publishedDate\\\":null,\\\"thumbnail\\\":\\\"\\\",\\\"engine\\\":\\\"brave\\\",\\\"parsed_url\\\":[\\\"https\\\",\\\"textcortex.com\\\",\\\"/post/gpt-4o-vs-gpt-4\\\",\\\"\\\",\\\"\\\",\\\"\\\"],\\\"template\\\":\\\"default.html\\\",\\\"engines\\\":[\\\"brave\\\"],\\\"positions\\\":[18],\\\"score\\\":0.05555555555555555,\\\"category\\\":\\\"general\\\"},{\\\"url\\\":\\\"https://www.pluralsight.com/resources/blog/data/ai-gpt-models-differences\\\",\\\"title\\\":\\\"GPT Base, GPT-3.5 Turbo & GPT-4: What's the difference?\\\",\\\"content\\\":\\\"A breakdown of OpenAI models, including their strengths, weaknesses, and cost. We also cover lesser-known AI models like Whisper and Embeddings.\\\",\\\"publishedDate\\\":null,\\\"thumbnail\\\":\\\"\\\",\\\"engine\\\":\\\"brave\\\",\\\"parsed_url\\\":[\\\"https\\\",\\\"www.pluralsight.com\\\",\\\"/resources/blog/data/ai-gpt-models-differences\\\",\\\"\\\",\\\"\\\",\\\"\\\"],\\\"template\\\":\\\"default.html\\\",\\\"engines\\\":[\\\"brave\\\"],\\\"positions\\\":[20],\\\"score\\\":0.05,\\\"category\\\":\\\"general\\\"},{\\\"title\\\":\\\"GPT-4 Turbo Preview: Exploring the 128k Context Window | Povio Blog\\\",\\\"content\\\":\\\"The GPT-4 Turbo Preview is not just an incremental update, but a substantial leap in the capabilities of AI language models. With a context window of 128k tokens, it stands head and shoulders above the existing GPT-4 models, which are limited to 8k and 32k tokens. This expansion isn't just about numbers; it represents a fundamental shift in how ...\\\",\\\"url\\\":\\\"https://povio.com/blog/gpt-4-turbo-preview-exploring-the-128k-context-window/\\\",\\\"engine\\\":\\\"duckduckgo\\\",\\\"parsed_url\\\":[\\\"https\\\",\\\"povio.com\\\",\\\"/blog/gpt-4-turbo-preview-exploring-the-128k-context-window/\\\",\\\"\\\",\\\"\\\",\\\"\\\"],\\\"template\\\":\\\"default.html\\\",\\\"engines\\\":[\\\"duckduckgo\\\"],\\\"positions\\\":[18],\\\"score\\\":0.05555555555555555,\\\"category\\\":\\\"general\\\"},{\\\"title\\\":\\\"Introducing OpenAI o1\\\",\\\"content\\\":\\\"In our tests, the next model update performs similarly to PhD students on challenging benchmark tasks in physics, chemistry, and biology. We also found that it excels in math and coding. In a qualifying exam for the International Mathematics Olympiad (IMO), GPT-4o correctly solved only 13% of problems, while the reasoning model scored 83%.\\\",\\\"url\\\":\\\"https://openai.com/index/introducing-openai-o1-preview/\\\",\\\"engine\\\":\\\"duckduckgo\\\",\\\"parsed_url\\\":[\\\"https\\\",\\\"openai.com\\\",\\\"/index/introducing-openai-o1-preview/\\\",\\\"\\\",\\\"\\\",\\\"\\\"],\\\"template\\\":\\\"default.html\\\",\\\"engines\\\":[\\\"duckduckgo\\\"],\\\"positions\\\":[19],\\\"score\\\":0.05263157894736842,\\\"category\\\":\\\"general\\\"},{\\\"title\\\":\\\"GPT-4 vs GPT-4-Turbo vs GPT-3.5-Turbo speed and ... | Geeky Gadgets\\\",\\\"content\\\":\\\"Picking the right OpenAI language model for your project can be crucial when it comes to performance, costs and implementation. OpenAI's suite, which includes the likes of GPT-3.5, GPT-4, and ...\\\",\\\"url\\\":\\\"https://www.geeky-gadgets.com/gpt-4-vs-gpt-4-turbo-vs-gpt-3-5-turbo/\\\",\\\"engine\\\":\\\"duckduckgo\\\",\\\"parsed_url\\\":[\\\"https\\\",\\\"www.geeky-gadgets.com\\\",\\\"/gpt-4-vs-gpt-4-turbo-vs-gpt-3-5-turbo/\\\",\\\"\\\",\\\"\\\",\\\"\\\"],\\\"template\\\":\\\"default.html\\\",\\\"engines\\\":[\\\"duckduckgo\\\"],\\\"positions\\\":[20],\\\"score\\\":0.05,\\\"category\\\":\\\"general\\\"},{\\\"title\\\":\\\"GPT-4 vs. GPT-4o vs. GPT-4o Mini: What's the Difference? | MUO\\\",\\\"content\\\":\\\"However, while GPT-4o mini is designed to be a multimodal model, its current ChatGPT version only supports text, without the ability to use vision or audio. Additionally, unlike GPT-4 and GPT-4o, ChatGPT does not allow GPT-4o mini to attach files. It is still unclear whether ChatGPT will allow multimodal capabilities in GPT-4o mini in the future.\\\",\\\"url\\\":\\\"https://www.makeuseof.com/gpt-4-vs-gpt-4-turbo-vs-gpt-4o-whats-the-difference/\\\",\\\"engine\\\":\\\"duckduckgo\\\",\\\"parsed_url\\\":[\\\"https\\\",\\\"www.makeuseof.com\\\",\\\"/gpt-4-vs-gpt-4-turbo-vs-gpt-4o-whats-the-difference/\\\",\\\"\\\",\\\"\\\",\\\"\\\"],\\\"template\\\":\\\"default.html\\\",\\\"engines\\\":[\\\"duckduckgo\\\"],\\\"positions\\\":[21],\\\"score\\\":0.047619047619047616,\\\"category\\\":\\\"general\\\"},{\\\"title\\\":\\\"GPT 3 vs. GPT 4: 10 Key Differences & How to Choose | Acorn\\\",\\\"content\\\":\\\"GPT 3 vs. GPT 4: Technical Differences. 1. Model Size. GPT-3, with its 175 billion parameters, marked a significant leap in the scale of language models at its release. This extensive number of parameters allowed for a richer and more nuanced understanding of language, enabling GPT-3 to generate highly coherent and contextually relevant text.\\\",\\\"url\\\":\\\"https://www.acorn.io/resources/learning-center/gpt3-vs-gpt4\\\",\\\"engine\\\":\\\"duckduckgo\\\",\\\"parsed_url\\\":[\\\"https\\\",\\\"www.acorn.io\\\",\\\"/resources/learning-center/gpt3-vs-gpt4\\\",\\\"\\\",\\\"\\\",\\\"\\\"],\\\"template\\\":\\\"default.html\\\",\\\"engines\\\":[\\\"duckduckgo\\\"],\\\"positions\\\":[22],\\\"score\\\":0.045454545454545456,\\\"category\\\":\\\"general\\\"},{\\\"title\\\":\\\"Using OpenAI o1 models and GPT-4o models on ChatGPT\\\",\\\"content\\\":\\\"The OpenAI o1-preview and o1-mini models are a new series of reasoning models for solving hard problems. This is a preview and we expect regular updates and improvements. While GPT-4o is still the best option for most prompts, the o1 series may be helpful for handling complex, problem-solving tasks in domains like research, strategy, coding ...\\\",\\\"url\\\":\\\"https://help.openai.com/en/articles/9824965-using-openai-o1-models-and-gpt-4o-models-on-chatgpt\\\",\\\"engine\\\":\\\"duckduckgo\\\",\\\"parsed_url\\\":[\\\"https\\\",\\\"help.openai.com\\\",\\\"/en/articles/9824965-using-openai-o1-models-and-gpt-4o-models-on-chatgpt\\\",\\\"\\\",\\\"\\\",\\\"\\\"],\\\"template\\\":\\\"default.html\\\",\\\"engines\\\":[\\\"duckduckgo\\\"],\\\"positions\\\":[24],\\\"score\\\":0.041666666666666664,\\\"category\\\":\\\"general\\\"},{\\\"title\\\":\\\"Evaluating GPT-4o's Performance in the Official European Board of ...\\\",\\\"content\\\":\\\"The higher accuracy rate may be attributed to the predominance of text-based questions in the MRQs and the use of the newer version of GPT-4. When questions from the Japanese radiology board exam were tested first with text-only using GPT-4 Turbo and then with both text and images using GPT-4 Turbo with Vision, the accuracy rates were similar.\\\",\\\"url\\\":\\\"https://www.academicradiology.org/article/S1076-6332(24)00653-6/fulltext\\\",\\\"engine\\\":\\\"duckduckgo\\\",\\\"parsed_url\\\":[\\\"https\\\",\\\"www.academicradiology.org\\\",\\\"/article/S1076-6332(24)00653-6/fulltext\\\",\\\"\\\",\\\"\\\",\\\"\\\"],\\\"template\\\":\\\"default.html\\\",\\\"engines\\\":[\\\"duckduckgo\\\"],\\\"positions\\\":[25],\\\"score\\\":0.04,\\\"category\\\":\\\"general\\\"},{\\\"title\\\":\\\"Generative artificial intelligence vs. law students: an empirical study ...\\\",\\\"content\\\":\\\"This study revealed significant differences in the performance of different AI models (GPT-4 vs. Google Bard) as well as the impact of prompt formulation. The study also demonstrated that if prompts are properly structured, AI can identify legal components that need to be proved as well as their relevant legal authorities (e.g. Paper 10).\\\",\\\"url\\\":\\\"https://www.tandfonline.com/doi/full/10.1080/17579961.2024.2392932\\\",\\\"engine\\\":\\\"duckduckgo\\\",\\\"parsed_url\\\":[\\\"https\\\",\\\"www.tandfonline.com\\\",\\\"/doi/full/10.1080/17579961.2024.2392932\\\",\\\"\\\",\\\"\\\",\\\"\\\"],\\\"template\\\":\\\"default.html\\\",\\\"engines\\\":[\\\"duckduckgo\\\"],\\\"positions\\\":[27],\\\"score\\\":0.037037037037037035,\\\"category\\\":\\\"general\\\"},{\\\"title\\\":\\\"Content variation issue with models GPT-4o,GPT-4o-mini and GPT-4-turbo\\\",\\\"content\\\":\\\"I have been testing models GPT-4o,GPT-4o-mini,GPT-4-turbo models and noticed a significant issue with the variation in its output. Multiple use cases I tested produced similar content with little to no variation, often resulting in responses that were nearly identical, with only minor adjustments. In contrast, the web version of GPT consistently generates diverse and varied responses ...\\\",\\\"url\\\":\\\"https://community.2024-09-24T06:47:44.706861519Z openai.com/t/content-variation-issue-with-models-gpt-4o-gpt-4o-mini-and-gpt-4-turbo/947710\\\",\\\"engine\\\":\\\"duckduckgo\\\",\\\"parsed_url\\\":[\\\"https\\\",\\\"community.openai.com\\\",\\\"/t/content-variation-issue-with-models-gpt-4o-gpt-4o-mini-and-gpt-4-turbo/947710\\\",\\\"\\\",\\\"\\\",\\\"\\\"],\\\"template\\\":\\\"default.html\\\",\\\"engines\\\":[\\\"duckduckgo\\\"],\\\"positions\\\":[28],\\\"score\\\":0.03571428571428571,\\\"category\\\":\\\"general\\\"}],\\\"answers\\\":[\\\"GPT-4 Turbo is our latest generation model. It's more capable, has an updated knowledge cutoff of April 2023 and introduces a 128k context window (the equivalent of 300 pages of text in a single prompt). The model is also 3X cheaper for input tokens and 2X cheaper for output tokens compared to the original GPT-4 model.\\\"],\\\"corrections\\\":[],\\\"infoboxes\\\":[],\\\"suggestions\\\":[\\\"How to use GPT-4 Turbo\\\",\\\"How to access GPT-4 Turbo\\\",\\\"GPT-4 Turbo release date\\\",\\\"GPT-4 Turbo API\\\",\\\"Chat GPT-4\\\",\\\"gpt-4 turbo price\\\",\\\"gpt-4 free\\\",\\\"Chatgpt\\\"],\\\"unresponsive_engines\\\":[[\\\"wolframalpha\\\",\\\"timeout\\\"]]}\"}]}],\"stream\":true,\"model\":\"claude-3-sonnet-20240229\",\"max_tokens\":4000,\"temperature\":0.5,\"top_p\":1,\"top_k\":5,\"tools\":[{\"name\":\"searchWeb\",\"description\":\"Perform a search query using SearXNG.\",\"input_schema\":{\"type\":\"object\",\"properties\":{\"q\":{\"type\":\"string\",\"description\":\"The search query string.\"},\"categories\":{\"type\":\"string\",\"description\":\"Comma-separated list of categories to search in (e.g., general, images, news).\"},\"language\":{\"type\":\"string\",\"description\":\"Language code for search results (e.g., en, fr, de).\"},\"format\":{\"type\":\"string\",\"description\":\"return format\"},\"safesearch\":{\"type\":\"integer\",\"description\":\"Safe search filter (0 = off, 1 = moderate, 2 = strict).\"}},\"required\":[\"q\",\"format\"]}},{\"name\":\"getSuggestions\",\"description\":\"Retrieve search suggestions based on a partial query.\",\"input_schema\":{\"type\":\"object\",\"properties\":{\"q\":{\"type\":\"string\",\"description\":\"The partial search query string.\"}},\"required\":[\"q\"]}}]}")

	req := ClaudeMessageCompletionRequest{}
	err := json.Unmarshal(raw, &req)
	if err != nil {
		t.Fatal(err)
		return
	}

	t.Log(tests.ToJSON(req))
}

func TestClaude_HandleProxyJSON(t *testing.T) {
	config := GetBedrockTestConfig()

	config.DEBUG = true

	tests.ToJSON(config)

	bedrock := NewBedrockClient(config)

	bodyJSON := `{
    "max_tokens": 1024,
    "messages": [{"role":"user","content":[{"type":"text","text":"創作一篇1000字小作文"}]}],
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
}

func TestClaude_HandleProxyStream(t *testing.T) {
	config := GetBedrockTestConfig()

	config.DEBUG = true

	tests.ToJSON(config)

	bedrock := NewBedrockClient(config)

	bodyJSON := `{
    "max_tokens": 1024,
    "messages": [{"role":"user","content":[{"type":"text","text":"創作一篇1000字小作文"}]}],
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