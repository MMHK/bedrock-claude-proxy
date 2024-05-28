package pkg

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

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
	Prompt            string   `json:"prompt,omitempty"`
	MaxTokensToSample int      `json:"max_tokens_to_sample,omitempty"`
	Temperature       float64  `json:"temperature,omitempty"`
	StopSequences     []string `json:"stop_sequences,omitempty"`
	TopP              float64  `json:"top_p,omitempty"`
	TopK              int      `json:"top_k,omitempty"`
	Stream            bool     `json:"-"`
	Model             string   `json:"-"`
}

func (this *ClaudeTextCompletionRequest) UnmarshalJSON(data []byte) error {
	tmp := &struct {
		Stream bool   `json:"stream"`
		Model  string `json:"model"`
	}{
		Stream: false,
	}
	if err := json.Unmarshal(data, tmp); err != nil {
		return err
	}
	this.Stream = tmp.Stream
	this.Model = tmp.Model

	return nil
}

type ClaudeMessageCompletionRequestContentSource struct {
	Type      string `json:"type,omitempty"`
	MediaType string `json:"media_type,omitempty"`
	Data      string `json:"data,omitempty"`
}

type ClaudeMessageCompletionRequestContent struct {
	Type   string `json:"type,omitempty"`
	Text   string `json:"text,omitempty"`
	Source *ClaudeMessageCompletionRequestContentSource `json:"source,omitempty"`
}

type ClaudeMessageCompletionRequestMessage struct {
	Role    string `json:"role,omitempty"`
	Content []*ClaudeMessageCompletionRequestContent `json:"content,omitempty"`
}

type ClaudeMessageCompletionRequestMetadata struct {
	UserId string `json:"user_id,omitempty"`
}

type ClaudeMessageCompletionRequestInputSchema struct {
	Type       string `json:"type,omitempty"`
	Properties map[string]*ClaudeMessageCompletionRequestPropertiesItem `json:"properties,omitempty"`
	Required []string `json:"required,omitempty"`
}

type ClaudeMessageCompletionRequestPropertiesItem struct {
	Type        string `json:"type,omitempty"`
	Description string `json:"description,omitempty"`
}

type ClaudeMessageCompletionRequestTools struct {
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	InputSchema *ClaudeMessageCompletionRequestInputSchema `json:"input_schema,omitempty"`
}

type ClaudeMessageCompletionRequest struct {
	ClaudeTextCompletionRequest

	AnthropicVersion string `json:"anthropic_version,omitempty"`
	MaxToken         int    `json:"max_tokens,omitempty"`
	System           string `json:"system,omitempty"`
	Messages         []*ClaudeMessageCompletionRequestMessage `json:"messages,omitempty"`
	Metadata *ClaudeMessageCompletionRequestMetadata `json:"-"`
	Tools []*ClaudeMessageCompletionRequestTools `json:"-"`
}

func (this *ClaudeMessageCompletionRequest) UnmarshalJSON(data []byte) error {
	tmp := &struct {
		Stream   bool   `json:"stream"`
		Model    string `json:"model"`
		Metadata *ClaudeMessageCompletionRequestMetadata `json:"metadata"`
		Tools []*ClaudeMessageCompletionRequestTools `json:"tools"`
	}{
		Stream: false,
	}
	if err := json.Unmarshal(data, tmp); err != nil {
		return err
	}
	this.Stream = tmp.Stream
	this.Model = tmp.Model
	this.Metadata = tmp.Metadata
	this.Tools = tmp.Tools

	return nil
}

type ClaudeTextCompletionResponse struct {
	Completion string `json:"completion,omitempty"`
	StopReason string `json:"stop_reason,omitempty"`
	Stop       string `json:"stop,omitempty"`
	Id         string `json:"id,omitempty"`
	Model      string `json:"model,omitempty"`
}

type ClaudeMessageCompletionResponse struct {
	ClaudeMessageStop

	Id           string                       `json:"id,omitempty"`
	Model        string                       `json:"model,omitempty"`
	Type         string                       `json:"type,omitempty"`
	Role         string                       `json:"role,omitempty"`
	Content      []*ClaudeMessageContentBlock `json:"content,omitempty"`
	Usage        *ClaudeMessageUsage          `json:"usage,omitempty"`
}

type ClaudeTextCompletionStreamEvent struct {
	Type       string `json:"type,omitempty"`
	StopReason string `json:"stop_reason,omitempty"`
	Model      string `json:"model,omitempty"`
	Completion string `json:"completion,omitempty"`
}

type ClaudeMessageUsage struct {
	InputTokens  int `json:"input_tokens,omitempty"`
	OutputTokens int `json:"output_tokens,omitempty"`
}

type ClaudeMessageStop struct {
	StopReason   string `json:"stop_reason,omitempty"`
	StopSequence string `json:"stop_sequence,omitempty"`
}
type ClaudeMessageInfo struct {
	ClaudeMessageStop

	Id      string              `json:"id,omitempty"`
	Type    string              `json:"type,omitempty"`
	Role    string              `json:"role,omitempty"`
	Content []string            `json:"content,omitempty"`
	Model   string              `json:"model,omitempty"`
	Usage   *ClaudeMessageUsage `json:"usage,omitempty"`
}
type ClaudeMessageContentBlock struct {
	Type  string      `json:"type,omitempty"`
	Text  string      `json:"text,omitempty"`
	Id    string      `json:"id,omitempty"`
	Name  string      `json:"name,omitempty"`
	Input interface{} `json:"input,omitempty"`
}
type ClaudeMessageDelta struct {
	ClaudeMessageStop

	Type         string `json:"type,omitempty"`
	Text         string `json:"text,omitempty"`
	PartialJson  string `json:"partial_json,omitempty"`
}

type ClaudeMessageCompletionStreamEvent struct {
	Type         string                     `json:"type,omitempty"`
	Model        string                     `json:"model,omitempty"`
	Completion   string                     `json:"completion,omitempty"`
	Message      *ClaudeMessageInfo         `json:"message,omitempty"`
	Usage        *ClaudeMessageUsage        `json:"usage,omitempty"`
	Index        int                        `json:"index,omitempty"`
	ContentBlock *ClaudeMessageContentBlock `json:"content_block,omitempty"`
	Delta        *ClaudeMessageDelta        `json:"delta,omitempty"`
}

type CompleteTextResponse struct {
	stream   bool
	Response *ClaudeTextCompletionResponse
	Events   <-chan *ClaudeTextCompletionStreamEvent
}

func NewStreamCompleteTextResponse(queue <-chan *ClaudeTextCompletionStreamEvent) *CompleteTextResponse {
	return &CompleteTextResponse{
		stream: true,
		Events: queue,
	}
}

func NewCompleteTextResponse(response *ClaudeTextCompletionResponse) *CompleteTextResponse {
	return &CompleteTextResponse{
		stream:   false,
		Response: response,
	}
}

func (this *CompleteTextResponse) IsStream() bool {
	return this.stream
}

func (this *CompleteTextResponse) GetResponse() *ClaudeTextCompletionResponse {
	return this.Response
}

func (this *CompleteTextResponse) GetEvents() <-chan *ClaudeTextCompletionStreamEvent {
	return this.Events
}

type MessageCompleteResponse struct {
	stream   bool
	Response *ClaudeMessageCompletionResponse
	Events   <-chan *ClaudeMessageCompletionStreamEvent
}

func NewStreamMessageCompleteResponse(queue <-chan *ClaudeMessageCompletionStreamEvent) *MessageCompleteResponse {
	return &MessageCompleteResponse{
		stream: true,
		Events: queue,
	}
}

func NewMessageCompleteResponse(response *ClaudeMessageCompletionResponse) *MessageCompleteResponse {
	return &MessageCompleteResponse{
		stream:   false,
		Response: response,
	}
}

func (this *MessageCompleteResponse) IsStream() bool {
	return this.stream
}

func (this *MessageCompleteResponse) GetResponse() *ClaudeMessageCompletionResponse {
	return this.Response
}

func (this *MessageCompleteResponse) GetEvents() <-chan *ClaudeMessageCompletionStreamEvent {
	return this.Events
}


func ClaudeTextCompletionStreamEventToSSE(event string, data string) string {
	return fmt.Sprintf("event: %s\ndata: %s\n", event, data)
}

type ClaudeTextCompletionStreamEventList []*ClaudeTextCompletionStreamEvent

func (this *ClaudeTextCompletionStreamEventList) Completion() string {
	var completion string
	for _, event := range *this {
		completion += event.Completion
	}
	return completion
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

func (this *BedrockClient) CompleteText(req *ClaudeTextCompletionRequest) (*CompleteTextResponse, error) {
	modelId := req.Model

	if !strings.HasSuffix(req.Prompt, "Assistant:") {
		req.Prompt = fmt.Sprintf("\n\nHuman: %s\n\nAssistant:", req.Prompt)
	}
	body, err := json.Marshal(req)
	if err != nil {
		Log.Errorf("Couldn't marshal the request: ", err)
		return nil, err
	}

	if req.Stream {
		output, err := this.client.InvokeModelWithResponseStream(context.Background(), &bedrock.InvokeModelWithResponseStreamInput{
			Body:        body,
			ModelId:     aws.String(modelId),
			ContentType: aws.String("application/json"),
		})
		if err != nil {
			Log.Error(err)
			return nil, err
		}

		//Log.Debugf("Request: %+v", output)

		reader := output.GetStream()
		eventQueue := make(chan *ClaudeTextCompletionStreamEvent, 0)

		go func() {
			defer reader.Close()
			defer close(eventQueue)

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
					eventQueue <- &resp

				case *types.UnknownUnionMember:
					Log.Errorf("unknown tag:", v.Tag)
					continue
				default:
					Log.Errorf("union is nil or unknown type")
					continue
				}
			}
		}()

		return NewStreamCompleteTextResponse(eventQueue), nil
	}

	output, err := this.client.InvokeModel(context.Background(), &bedrock.InvokeModelInput{
		Body:        body,
		ModelId:     aws.String(modelId),
		ContentType: aws.String("application/json"),
	})
	if err != nil {
		Log.Error(err)
		return nil, err
	}

	if output.Body != nil {
		var resp ClaudeTextCompletionResponse
		err = json.NewDecoder(bytes.NewReader(output.Body)).Decode(&resp)
		if err != nil {
			Log.Error(err)
			return nil, err
		}
		//Log.Debug(resp)

		return NewCompleteTextResponse(&resp), nil
	}

	return nil, nil
}

func (this *BedrockClient) MessageCompletion(req *ClaudeMessageCompletionRequest) (*MessageCompleteResponse, error) {
	modelId := req.Model
	body, err := json.Marshal(req)
	if err != nil {
		Log.Errorf("Couldn't marshal the request: ", err)
		return nil, err
	}

	Log.Debugf("Request: %s", string(body))

	if req.Stream {
		output, err := this.client.InvokeModelWithResponseStream(context.Background(), &bedrock.InvokeModelWithResponseStreamInput{
			Body:        body,
			ModelId:     aws.String(modelId),
			ContentType: aws.String("application/json"),
		})
		if err != nil {
			Log.Error(err)
			return nil, err
		}


		reader := output.GetStream()
		eventQueue := make(chan *ClaudeMessageCompletionStreamEvent, 0)

		go func() {
			defer reader.Close()
			defer close(eventQueue)

			for event := range reader.Events() {
				switch v := event.(type) {
				case *types.ResponseStreamMemberChunk:

					//Log.Info("payload", string(v.Value.Bytes))

					var resp ClaudeMessageCompletionStreamEvent
					err := json.NewDecoder(bytes.NewReader(v.Value.Bytes)).Decode(&resp)
					if err != nil {
						Log.Error(err)
						continue
					}
					eventQueue <- &resp

				case *types.UnknownUnionMember:
					Log.Errorf("unknown tag:", v.Tag)
					continue
				default:
					Log.Errorf("union is nil or unknown type")
					continue
				}
			}
		}()

		return NewStreamMessageCompleteResponse(eventQueue), nil
	}

	output, err := this.client.InvokeModel(context.Background(), &bedrock.InvokeModelInput{
		Body:        body,
		ModelId:     aws.String(modelId),
		ContentType: aws.String("application/json"),
	})
	if err != nil {
		Log.Error(err)
		return nil, err
	}

	if output.Body != nil {
		var resp ClaudeMessageCompletionResponse
		err = json.NewDecoder(bytes.NewReader(output.Body)).Decode(&resp)
		if err != nil {
			Log.Error(err)
			return nil, err
		}
		//Log.Debug(resp)

		return NewMessageCompleteResponse(&resp), nil
	}

	return nil, nil
}
