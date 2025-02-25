package pkg

import (
	"bufio"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	bedrock "github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime/types"
)

type BedrockConfig struct {
	AccessKey                string            `json:"access_key"`
	SecretKey                string            `json:"secret_key"`
	Region                   string            `json:"region"`
	AnthropicVersionMappings map[string]string `json:"anthropic_version_mappings"`
	ModelMappings            map[string]string `json:"model_mappings"`
	AnthropicDefaultModel    string            `json:"anthropic_default_model"`
	AnthropicDefaultVersion  string            `json:"anthropic_default_version"`
	EnableComputerUse        bool              `json:"enable_computer_use"`
	EnableOutputReason       bool              `json:"enable_output_reasoning"`
	ReasonBudgetTokens       int               `json:"reason_budget_tokens"`
	DEBUG 					 bool              `json:"debug,omitempty"`
}

func (this *BedrockConfig) GetInvokeEndpoint(modelId string) string {
	return fmt.Sprintf("bedrock-runtime.%s.amazonaws.com/model/%s/invoke", this.Region, modelId)
}

func (this *BedrockConfig) GetInvokeStreamEndpoint(modelId string, region string) string {
	return fmt.Sprintf("bedrock-runtime.%s.amazonaws.com/model/%s/invoke-with-response-stream", this.Region, modelId)
}

type ThinkingConfig struct {
	Type         string `json:"type"`
	BudgetTokens int    `json:"budget_tokens"`
}

func ParseMappingsFromStr(raw string) map[string]string {
	mappings := map[string]string{}
	pairs := strings.Split(raw, ",")
	// 遍历每个键值对
	for _, pair := range pairs {
		// 以等号分割键和值
		kv := strings.Split(pair, "=")
		if len(kv) == 2 {
			key := strings.TrimSpace(kv[0])
			value := strings.TrimSpace(kv[1])
			mappings[key] = value
		}
	}

	return mappings
}

func LoadBedrockConfigWithEnv() *BedrockConfig {
	config := &BedrockConfig{
		AccessKey:                os.Getenv("AWS_BEDROCK_ACCESS_KEY"),
		SecretKey:                os.Getenv("AWS_BEDROCK_SECRET_KEY"),
		Region:                   os.Getenv("AWS_BEDROCK_REGION"),
		ModelMappings:            ParseMappingsFromStr(os.Getenv("AWS_BEDROCK_MODEL_MAPPINGS")),
		AnthropicVersionMappings: ParseMappingsFromStr(os.Getenv("AWS_BEDROCK_ANTHROPIC_VERSION_MAPPINGS")),
		AnthropicDefaultModel:    os.Getenv("AWS_BEDROCK_ANTHROPIC_DEFAULT_MODEL"),
		AnthropicDefaultVersion:  os.Getenv("AWS_BEDROCK_ANTHROPIC_DEFAULT_VERSION"),
		EnableComputerUse:        os.Getenv("AWS_BEDROCK_ENABLE_COMPUTER_USE") == "true",
		EnableOutputReason:       os.Getenv("AWS_BEDROCK_ENABLE_OUTPUT_REASON") == "true",
		ReasonBudgetTokens:       1024,
	}

	budget := os.Getenv("AWS_BEDROCK_REASON_BUDGET_TOKENS")
	if len(budget) > 0 {
		if tokens, err := strconv.Atoi(budget); err == nil {
			config.ReasonBudgetTokens = tokens
		}
	}


	return config;
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
	type Alias ClaudeTextCompletionRequest
	tmp := &struct {
		*Alias

		Stream bool   `json:"stream"`
		Model  string `json:"model"`
	}{
		Stream: false,
		Alias:  (*Alias)(this),
	}
	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	this.Stream = tmp.Stream
	this.Model = tmp.Model

	//Log.Debug("ClaudeTextCompletionRequest UnmarshalJSON")
	//Log.Debug(tests.ToJSON(tmp))
	//Log.Debugf("%+v", this)

	return nil
}

type ClaudeMessageCompletionRequestContentSource struct {
	Type      string `json:"type,omitempty"`
	MediaType string `json:"media_type,omitempty"`
	Data      string `json:"data,omitempty"`
}

type ClaudeMessageCompletionRequestContent struct {
	Type      string                                       `json:"type,omitempty"`
	Name      string                                       `json:"name,omitempty"`
	Id        string                                       `json:"id,omitempty"`
	Text      string                                       `json:"text,omitempty"`
	ToolUseID string                                       `json:"tool_use_id,omitempty"`
	IsError   string                                       `json:"is_error,omitempty"`
	Source    *ClaudeMessageCompletionRequestContentSource `json:"source,omitempty"`
	Content   json.RawMessage                              `json:"content,omitempty"`
}

type ClaudeMessageCompletionRequestMessage struct {
	Role    string          `json:"role,omitempty"`
	Content json.RawMessage `json:"content,omitempty"`
	Text    string          `json:"text,omitempty"`
}

type ClaudeMessageCompletionRequestMetadata struct {
	UserId string `json:"user_id,omitempty"`
}

type ClaudeMessageCompletionRequestInputSchema struct {
	Type       string                                                   `json:"type,omitempty"`
	Properties map[string]*ClaudeMessageCompletionRequestPropertiesItem `json:"properties,omitempty"`
	Required   []string                                                 `json:"required,omitempty"`
}

type ClaudeMessageCompletionRequestPropertiesItem struct {
	Type        string `json:"type,omitempty"`
	Description string `json:"description,omitempty"`
}

type ClaudeMessageCompletionRequestTools struct {
	Name        string                                     `json:"name,omitempty"`
	Description string                                     `json:"description,omitempty"`
	InputSchema *ClaudeMessageCompletionRequestInputSchema `json:"input_schema,omitempty"`
}

type ClaudeMessageCompletionRequest struct {
	Temperature      float64                                  `json:"temperature,omitempty"`
	StopSequences    []string                                 `json:"stop_sequences,omitempty"`
	TopP             float64                                  `json:"top_p,omitempty"`
	TopK             int                                      `json:"top_k,omitempty"`
	Stream           bool                                     `json:"-"`
	Model            string                                   `json:"-"`
	AnthropicVersion string                                   `json:"anthropic_version,omitempty"`
	MaxToken         int                                      `json:"max_tokens,omitempty"`
	System           string                                   `json:"system,omitempty"`
	Messages         []*ClaudeMessageCompletionRequestMessage `json:"messages,omitempty"`
	Metadata         *ClaudeMessageCompletionRequestMetadata  `json:"-"`
	Tools            []*ClaudeMessageCompletionRequestTools   `json:"tools,omitempty"`
}

func (this *ClaudeMessageCompletionRequest) UnmarshalJSON(data []byte) error {
	type Alias ClaudeMessageCompletionRequest
	tmp := &struct {
		*Alias

		Stream   bool                                    `json:"stream"`
		Model    string                                  `json:"model"`
		Metadata *ClaudeMessageCompletionRequestMetadata `json:"metadata"`
		Tools    []*ClaudeMessageCompletionRequestTools  `json:"tools"`
	}{
		Stream: false,
		Alias:  (*Alias)(this),
	}

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	this.Metadata = tmp.Metadata
	if tmp.Tools != nil {
		this.Tools = tmp.Tools
	} else {
		this.Tools = []*ClaudeMessageCompletionRequestTools{}
	}
	if this.TopK < 0 {
		this.TopK = 0
	}
	if this.TopP < 0 {
		this.TopP = 0.0
	}
	this.Stream = tmp.Stream
	this.Model = tmp.Model

	//Log.Debug("ClaudeMessageCompletionRequest UnmarshalJSON")
	//Log.Debug(tests.ToJSON(tmp))
	//Log.Debugf("%+v", this)

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

	Id      string                       `json:"id,omitempty"`
	Model   string                       `json:"model,omitempty"`
	Type    string                       `json:"type,omitempty"`
	Role    string                       `json:"role,omitempty"`
	Content []*ClaudeMessageContentBlock `json:"content,omitempty"`
	Usage   *ClaudeMessageUsage          `json:"usage,omitempty"`
}

type ISSEDecoder interface {
	GetBytes() []byte
	GetEvent() string
	GetText() string
}

type ClaudeTextCompletionStreamEvent struct {
	Type       string `json:"type,omitempty"`
	StopReason string `json:"stop_reason,omitempty"`
	Model      string `json:"model,omitempty"`
	Completion string `json:"completion,omitempty"`
	Raw        []byte `json:"-"`
}

func (this *ClaudeTextCompletionStreamEvent) GetBytes() []byte {
	return this.Raw
}

func (this *ClaudeTextCompletionStreamEvent) GetEvent() string {
	return this.Type
}

func (this *ClaudeTextCompletionStreamEvent) GetText() string {
	return this.Completion
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

	Type        string `json:"type,omitempty"`
	Text        string `json:"text,omitempty"`
	PartialJson string `json:"partial_json,omitempty"`
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
	Raw          []byte                     `json:"-"`
}

func (this *ClaudeMessageCompletionStreamEvent) GetBytes() []byte {
	return this.Raw
}

func (this *ClaudeMessageCompletionStreamEvent) GetEvent() string {
	return this.Type
}

func (this *ClaudeMessageCompletionStreamEvent) GetText() string {
	if this.Delta != nil {
		return this.Delta.Text
	}
	return this.Completion
}

type CompleteTextResponse struct {
	stream   bool
	Response *ClaudeTextCompletionResponse
	Events   <-chan ISSEDecoder
}

func NewStreamCompleteTextResponse(queue <-chan ISSEDecoder) *CompleteTextResponse {
	return &CompleteTextResponse{
		stream: true,
		Events: queue,
	}
}

type IStreamableResponse interface {
	IsStream() bool
	GetResponse() interface{}
	GetEvents() <-chan ISSEDecoder
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

func (this *CompleteTextResponse) GetResponse() interface{} {
	return this.Response
}

func (this *CompleteTextResponse) GetEvents() <-chan ISSEDecoder {
	return this.Events
}

type MessageCompleteResponse struct {
	stream   bool
	Response *ClaudeMessageCompletionResponse
	Events   <-chan ISSEDecoder
}

func NewStreamMessageCompleteResponse(queue <-chan ISSEDecoder) *MessageCompleteResponse {
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

func (this *MessageCompleteResponse) GetResponse() interface{} {
	return this.Response
}

func (this *MessageCompleteResponse) GetEvents() <-chan ISSEDecoder {
	return this.Events
}

func NewSSERaw(encoder ISSEDecoder) []byte {
	return []byte(fmt.Sprintf("event: %s\ndata: %s\n\n", encoder.GetEvent(), string(encoder.GetBytes())))
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

func (this *BedrockClient) GetModelMappings(source string) (string, error) {
	if len(this.config.ModelMappings) > 0 {
		if target, ok := this.config.ModelMappings[source]; ok {
			return target, nil
		}
	}

	return this.config.AnthropicDefaultModel, errors.New(fmt.Sprintf("model %s not found in model mappings", source))
}

func (this *BedrockClient) SignRequest(request *http.Request) (*http.Request, bool, error) {
	contentType := request.Header.Get("Content-Type")
	cloneReq := request
	isStream := false
	Model := ""
	var bodyBuff bytes.Buffer

	if strings.Contains(contentType, "json") {
		reader := io.TeeReader(request.Body, &bodyBuff)

		decoder := json.NewDecoder(reader)
		wrapper := make(map[string]interface{})
		err := decoder.Decode(&wrapper)
		if err != nil {
			Log.Error(err)
			return request, false, err
		}
		if srcModel, ok := wrapper["model"]; ok {
			if _model, ok := srcModel.(string); ok {
				Model = _model
			}
		}
		if srcStream, ok := wrapper["stream"]; ok {
			if _stream, ok := srcStream.(bool); ok {
				isStream = _stream
			}
		}

		wrapper["anthropic_version"] = this.config.AnthropicDefaultVersion

		if this.config.EnableComputerUse {
			wrapper["anthropic_beta"] = "computer-use-2024-10-22"
		}

		if _, ok := wrapper["thinking"]; !ok && this.config.EnableOutputReason {
			wrapper["thinking"] = &ThinkingConfig{
				Type:         "enabled",
				BudgetTokens: this.config.ReasonBudgetTokens,
			}
		}

		newBody, err := json.Marshal(wrapper)
		if err != nil {
			return request, false, err
		}

		cloneReq = &http.Request{
			Method: request.Method,
			URL:    request.URL,
			Proto:  request.Proto,
			Header: request.Header.Clone(),
			Body:   io.NopCloser(bytes.NewBuffer(newBody)),
		}
	}

	cfg, err := awsConfig.LoadDefaultConfig(context.TODO(),
		awsConfig.WithRegion(this.config.Region),
		awsConfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			this.config.AccessKey,
			this.config.SecretKey,
			"",
		)),
	)
	if err != nil {
		return nil, false, err
	}

	Model, err = this.GetModelMappings(Model)
	if err != nil {
		Log.Error(err)
	}

	bedrockRuntimeEndPoint := fmt.Sprintf(`https://bedrock-runtime.%s.amazonaws.com/model/%s/invoke`, this.config.Region, Model)
	if isStream {
		bedrockRuntimeEndPoint = fmt.Sprintf(`https://bedrock-runtime.%s.amazonaws.com/model/%s/invoke-with-response-stream`, this.config.Region, Model)
	}

	cloneReq, err = http.NewRequest("POST", bedrockRuntimeEndPoint, cloneReq.Body)
	if err != nil {
		Log.Error(err)
		return nil, false, err
	}

	cloneReq.Header.Set("Content-Type", "application/json")
	cloneReq.Header.Set("Accept", "application/json")

	signer := v4.NewSigner()

	// 获取凭证
	credentialList, err := cfg.Credentials.Retrieve(context.TODO())
	if err != nil {
		Log.Error(err)
		return nil, false, err
	}

	hash := sha256.Sum256(bodyBuff.Bytes())
	payloadHash := hex.EncodeToString(hash[:])
	// 签名请求
	err = signer.SignHTTP(context.TODO(), credentialList, cloneReq, payloadHash, "bedrock", cfg.Region, time.Now())
	if err != nil {
		Log.Error(err)
		return nil, false, err
	}

	return cloneReq, isStream, nil
}

func (this *BedrockClient) handleSSEStream(w http.ResponseWriter, res *http.Response) error {
	// 設置 SSE 相關的 headers
	for k, v := range res.Header {
		w.Header()[k] = v
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)
	if !ok {
		return fmt.Errorf("streaming unsupported")
	}

	if this.config.DEBUG {
		Log.Infof("handleSSEStream: %s", res.Header.Get("Content-Type"))
	}


	scanner := bufio.NewScanner(res.Body)
	for scanner.Scan() {
		line := scanner.Text()


		// 寫入修改後的行並立即刷新
		fmt.Fprintf(w, "%s\n", line)
		if this.config.DEBUG {
			Log.Infof("SSE: %s\n", line)
		}
		flusher.Flush()
	}

	if err := scanner.Err(); err != nil {
		Log.Error(err)
		return err
	}

	return nil
}


func (this *BedrockClient) HandleProxy(w http.ResponseWriter, r *http.Request) {
	cloneReq, isStream, err := this.SignRequest(r)
	if err != nil {
		Log.Error(err)
		w.Header().Set("Content-Type", "application/json")
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}

	if isStream {
		var (
			resp *http.Response
			err  error
		)
		// 發送請求到目標服務器
		if this.config.DEBUG {
			reqDump, _ := httputil.DumpRequestOut(cloneReq, true)
			Log.Infof("Request:\n%s", string(reqDump))
		}

		resp, err = http.DefaultClient.Do(cloneReq)
		if err != nil {
			Log.Error(err)
			w.Header().Set("Content-Type", "application/json")
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}
		defer resp.Body.Close()

		if err := this.handleSSEStream(w, resp); err != nil {
			Log.Error(err)
		}
		return
	}

	resp, err := http.DefaultClient.Do(cloneReq)
	if err != nil {
		Log.Error(err)
		w.Header().Set("Content-Type", "application/json")
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}

	// 寫入修改後的響應
	w.WriteHeader(resp.StatusCode)
	for k, v := range resp.Header {
		w.Header()[k] = v
	}
	_,err = io.Copy(w, resp.Body)
	if err != nil {
		Log.Error(err)
	}
}


func (this *BedrockClient) CompleteText(req *ClaudeTextCompletionRequest) (IStreamableResponse, error) {
	modelId := req.Model
	mappedModel, exist := this.config.ModelMappings[modelId]
	if exist {
		modelId = mappedModel
	}
	if len(modelId) == 0 {
		modelId = this.config.AnthropicDefaultModel
	}

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
		eventQueue := make(chan ISSEDecoder, 10)

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
					resp.Raw = v.Value.Bytes
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

func (this *BedrockClient) MessageCompletion(req *ClaudeMessageCompletionRequest) (IStreamableResponse, error) {
	modelId := req.Model
	mappedModel, exist := this.config.ModelMappings[modelId]
	if exist {
		modelId = mappedModel
	}
	if len(modelId) == 0 {
		modelId = this.config.AnthropicDefaultModel
	}
	apiVersion, exist := this.config.AnthropicVersionMappings[req.AnthropicVersion]
	if exist {
		req.AnthropicVersion = apiVersion
	}
	if len(req.AnthropicVersion) == 0 {
		req.AnthropicVersion = this.config.AnthropicDefaultVersion
	}

	body, err := json.Marshal(req)
	if err != nil {
		Log.Errorf("Couldn't marshal the request: ", err)
		return nil, err
	}

	Log.Debugf("Request: %s", string(body))
	Log.Debugf("Request Model ID: %s", modelId)

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
		eventQueue := make(chan ISSEDecoder, 10)

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
					resp.Raw = v.Value.Bytes
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
