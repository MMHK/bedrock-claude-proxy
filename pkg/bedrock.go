package pkg

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	bedrock "github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime/types"
)

type BedrockConfig struct {
	AccessKey string `json:"access_key"`
	SecretKey string `json:"secret_key"`
	Region    string `json:"region"`
}

func (this *BedrockConfig) GetInvokeEndpoint(modelId string) string {
	return fmt.Sprintf("bedrock-runtime.%s.amazonaws.com/model/%s/invoke", this.Region, modelId)
}

func (this *BedrockConfig) GetInvokeStreamEndpoint(modelId string, region string) string {
	return fmt.Sprintf("bedrock-runtime.%s.amazonaws.com/model/%s/invoke-with-response-stream", region, modelId)
}

func LoadBedrockConfigWithEnv() *BedrockConfig {
	return &BedrockConfig{
		AccessKey: os.Getenv("AWS_BEDROCK_ACCESS_KEY"),
		SecretKey: os.Getenv("AWS_BEDROCK_SECRET_KEY"),
		Region:    os.Getenv("AWS_BEDROCK_REGION"),
	}
}

type BedrockClient struct {
	config *BedrockConfig
	client *bedrock.Client
}

type ClaudeTextCompletionRequest struct {
	Prompt            string   `json:"prompt"`
	MaxTokensToSample int      `json:"max_tokens_to_sample"`
	Temperature       float64  `json:"temperature,omitempty"`
	StopSequences     []string `json:"stop_sequences,omitempty"`
	TopP              float64  `json:"top_p,omitempty"`
	TopK              int      `json:"top_k,omitempty"`
	Stream            bool     `json:"-"`
}

type ClaudeTextCompletionStreamEvent struct {
	Type       string `json:"type"`
	StopReason string `json:"stop_reason"`
	Model      string `json:"model"`
	Completion string `json:"completion"`
}

type ClaudeTextCompletionStreamEventList []*ClaudeTextCompletionStreamEvent

func (this *ClaudeTextCompletionStreamEventList) Completion() string {
	var completion string
	for _, event := range *this {
		completion += event.Completion
	}
	return completion
}

type ClaudeResponse struct {
	Completion string `json:"completion"`
}

func NewBedrockClient(config *BedrockConfig) *BedrockClient {
	staticProvider := credentials.NewStaticCredentialsProvider(config.AccessKey, config.SecretKey, "")

	cfg, err := awsConfig.LoadDefaultConfig(context.TODO(),
		awsConfig.WithRegion(config.Region),
	awsConfig.WithCredentialsProvider(staticProvider))
	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)
	}

	return &BedrockClient{
		config: config,
		client: bedrock.NewFromConfig(cfg),
	}
}

func (this *BedrockClient) CompleteText(modelId string, req *ClaudeTextCompletionRequest) (ClaudeTextCompletionStreamEventList, error) {
	req.Prompt = fmt.Sprintf("\n\nHuman: %s\n\nAssistant:", req.Prompt)
	body, err := json.Marshal(req)
	if err != nil {
		Log.Errorf("Couldn't marshal the request: ", err)
	}

	if req.Stream {
		output, err := this.client.InvokeModelWithResponseStream(context.Background(), &bedrock.InvokeModelWithResponseStreamInput{
			Body:        body,
			ModelId:     aws.String(modelId),
			ContentType: aws.String("application/json"),
		})
		if err != nil {
			Log.Errorf("Couldn't invoke the model: ", err)
			return nil, err
		}

		//Log.Debugf("Request: %+v", output)

		events, err := processStreamingOutput(output)
		if err != nil {
			Log.Errorf("Couldn't invoke the model: ", err)
			return nil, err
		}

		return events, nil;
	}

	output, err := this.client.InvokeModel(context.Background(), &bedrock.InvokeModelInput{
		Body:        body,
		ModelId:     aws.String(modelId),
		ContentType: aws.String("application/json"),
	})
	if err != nil {
		Log.Errorf("Couldn't invoke the model: ", err)
		return nil, err
	}

	if output.Body != nil {
		var resp ClaudeTextCompletionStreamEvent
		err = json.NewDecoder(bytes.NewReader(output.Body)).Decode(&resp)
		if err != nil {
			Log.Errorf("Couldn't decode the response: ", err)
			return nil, err
		}
		//Log.Debug(resp)

		return ClaudeTextCompletionStreamEventList{&resp}, nil
	}

	return nil, nil
}



func processStreamingOutput(output *bedrock.InvokeModelWithResponseStreamOutput) (ClaudeTextCompletionStreamEventList, error) {
	list := []*ClaudeTextCompletionStreamEvent{}

	reader := output.GetStream()
	defer reader.Close()

	for event := range reader.Events() {
		switch v := event.(type) {
		case *types.ResponseStreamMemberChunk:

			//Log.Info("payload", string(v.Value.Bytes))

			var resp ClaudeTextCompletionStreamEvent
			err := json.NewDecoder(bytes.NewReader(v.Value.Bytes)).Decode(&resp)
			if err != nil {
				Log.Error(err)
				continue
			}
			//Log.Debug(resp)

			list = append(list, &resp)

		case *types.UnknownUnionMember:
			Log.Errorf("unknown tag:", v.Tag)
			continue
		default:
			Log.Errorf("union is nil or unknown type")
			continue
		}
	}

	return list, nil
}