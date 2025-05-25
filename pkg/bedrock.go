package pkg

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"

	"github.com/aws/aws-sdk-go-v2/aws/protocol/eventstream"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	bedrockRuntime "github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
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
	DEBUG                    bool              `json:"debug,omitempty"`
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
		DEBUG:                    os.Getenv("AWS_BEDROCK_DEBUG") == "true",
	}

	budget := os.Getenv("AWS_BEDROCK_REASON_BUDGET_TOKENS")
	if len(budget) > 0 {
		if tokens, err := strconv.Atoi(budget); err == nil {
			config.ReasonBudgetTokens = tokens
		}
	}

	return config
}

type BedrockClient struct {
	config *BedrockConfig
	client *bedrockRuntime.Client
}

type ModelInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	ID      string `json:"id"`
}

// BedrockFoundationModel represents a model from Bedrock API
type BedrockFoundationModel struct {
	ModelId                    string   `json:"modelId"`
	ModelName                  string   `json:"modelName"`
	ProviderName               string   `json:"providerName"`
	InputModalities            []string `json:"inputModalities"`
	OutputModalities           []string `json:"outputModalities"`
	ResponseStreamingSupported bool     `json:"responseStreamingSupported"`
}

// BedrockModelsResponse represents the response from Bedrock foundation models API
type BedrockModelsResponse struct {
	ModelSummaries []BedrockFoundationModel `json:"modelSummaries"`
}

// ModelValidationResult represents the validation result for a model mapping
type ModelValidationResult struct {
	ConfigModel    string `json:"config_model"`
	BedrockModelId string `json:"bedrock_model_id"`
	ModelName      string `json:"model_name,omitempty"` // Optional, can be used to store the model name if available
	IsValid        bool   `json:"is_valid"`
	Available      bool   `json:"available"`
}

func (this *BedrockClient) ListModels() []ModelInfo {
	models := make([]ModelInfo, 0, len(this.config.AnthropicVersionMappings))
	for name, version := range this.config.AnthropicVersionMappings {
		models = append(models, ModelInfo{ID: name, Version: version, Name: fmt.Sprintf("%s-%s", name, version)})
	}
	return models
}

// GetBedrockAvailableModels fetches available models from Bedrock API
func (this *BedrockClient) GetBedrockAvailableModels() ([]BedrockFoundationModel, error) {
	// Create the API endpoint URL - use bedrock service, not bedrock-runtime
	apiEndpoint := fmt.Sprintf("https://bedrock.%s.amazonaws.com/foundation-models", this.config.Region)

	// Create HTTP request
	req, err := http.NewRequest("GET", apiEndpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")

	// Sign the request using AWS v4 signature
	cfg, err := awsConfig.LoadDefaultConfig(context.TODO(),
		awsConfig.WithRegion(this.config.Region),
		awsConfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			this.config.AccessKey,
			this.config.SecretKey,
			"",
		)),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %v", err)
	}

	signer := v4.NewSigner()
	credentialList, err := cfg.Credentials.Retrieve(context.TODO())
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve credentials: %v", err)
	}

	// Sign the request
	hash := sha256.Sum256([]byte{})
	payloadHash := hex.EncodeToString(hash[:])
	err = signer.SignHTTP(context.TODO(), credentialList, req, payloadHash, "bedrock", cfg.Region, time.Now(), func(options *v4.SignerOptions) {
		if this.config.DEBUG {
			options.LogSigning = true
		}
	})
	if err != nil {
		return nil, fmt.Errorf("failed to sign request: %v", err)
	}

	// Execute the request
	httpClient := http.DefaultClient
	if this.config.DEBUG {
		httpClient = &http.Client{
			Transport: loggingRoundTripper{
				wrapped: http.DefaultTransport,
			},
		}
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	// Parse the response
	var modelsResponse BedrockModelsResponse
	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&modelsResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	return modelsResponse.ModelSummaries, nil
}

// ValidateModelMappings validates the configured model mappings against available Bedrock models
func (this *BedrockClient) ValidateModelMappings() ([]ModelValidationResult, error) {
	// Get available models from Bedrock
	availableModels, err := this.GetBedrockAvailableModels()
	if err != nil {
		return nil, fmt.Errorf("failed to get available models: %v", err)
	}

	// Create a map for quick lookup
	availableModelIds := make(map[string]string)
	for _, model := range availableModels {
		availableModelIds[model.ModelId] = model.ModelName
	}

	// Validate each mapping
	var results []ModelValidationResult
	for configModel, bedrockModelId := range this.config.ModelMappings {
		modelName, ok := availableModelIds[bedrockModelId]
		if ok {
			results = append(results, ModelValidationResult{
				ConfigModel:    configModel,
				BedrockModelId: bedrockModelId,
				IsValid:        ok,        // Mapping exists in config
				ModelName:      modelName, // Optional, can be filled if needed
				Available:      ok,
			})
		}
	}

	return results, nil
}

// GetMergedModelList returns a combined list of configured and available models
func (this *BedrockClient) GetMergedModelList() ([]ModelInfo, error) {
	// Get validation results
	validationResults, err := this.ValidateModelMappings()
	if err != nil {
		Log.Errorf("Failed to validate model mappings: %v", err)
		// Fall back to config-only models
		return this.ListModels(), nil
	}

	Log.Debugf("Validation results: %v", validationResults)

	// Create model list from validation results
	var models []ModelInfo
	for _, result := range validationResults {
		if result.Available {
			// Use the config model name and extract version from mapping
			version := this.config.AnthropicDefaultVersion
			if mappedVersion, ok := this.config.AnthropicVersionMappings[result.ConfigModel]; ok {
				version = mappedVersion
			}

			models = append(models, ModelInfo{
				Name:    result.ModelName,
				Version: version,
				ID:      result.ConfigModel,
			})
		} else {
			Log.Warningf("Model %s (mapped to %s) is not available in Bedrock", result.ConfigModel, result.BedrockModelId)
		}
	}

	// If no valid models found, fall back to config models
	if len(models) == 0 {
		Log.Warning("No valid models found from Bedrock API, falling back to config models")
		return this.ListModels(), nil
	}

	return models, nil
}

// 自定义的 RoundTripper 用于记录请求和响应
type loggingRoundTripper struct {
	wrapped http.RoundTripper
}

func (l loggingRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	// 记录请求
	reqDump, _ := httputil.DumpRequestOut(req, true)
	Log.Infof("Request:\n%s", string(reqDump))

	// 发送请求
	resp, err := l.wrapped.RoundTrip(req)
	if err != nil {
		return nil, err
	}

	// 记录响应
	respDump, _ := httputil.DumpResponse(resp, true)
	Log.Infof("Response:\n%s", string(respDump))

	// 重要：我们需要重新创建响应体，因为 DumpResponse 会消耗它
	bodyBytes, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	resp.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	return resp, nil
}

func NewBedrockClient(config *BedrockConfig) *BedrockClient {
	staticProvider := credentials.NewStaticCredentialsProvider(config.AccessKey, config.SecretKey, "")

	opt := []func(*awsConfig.LoadOptions) error{
		awsConfig.WithRegion(config.Region),
		awsConfig.WithCredentialsProvider(staticProvider),
	}

	if config.DEBUG {
		httpClient := &http.Client{
			Transport: loggingRoundTripper{
				wrapped: http.DefaultTransport,
			},
		}
		opt = append(opt, awsConfig.WithHTTPClient(httpClient))
	}

	cfg, err := awsConfig.LoadDefaultConfig(context.TODO(), opt...)

	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)
	}

	return &BedrockClient{
		config: config,
		client: bedrockRuntime.NewFromConfig(cfg),
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
	reader := io.TeeReader(request.Body, &bodyBuff)

	if strings.Contains(contentType, "json") {
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

		Model, err = this.GetModelMappings(Model)
		if err != nil {
			Log.Error(err)
		} else {
			wrapper["model"] = Model
		}

		if srcStream, ok := wrapper["stream"]; ok {
			if _stream, ok := srcStream.(bool); ok {
				isStream = _stream
			}
		}

		wrapper["anthropic_version"] = this.config.AnthropicDefaultVersion
		delete(wrapper, "model")
		delete(wrapper, "stream")

		if this.config.EnableComputerUse {
			wrapper["anthropic_beta"] = "computer-use-2024-10-22"
		}

		if _, ok := wrapper["thinking"]; !ok && this.config.EnableOutputReason {
			wrapper["thinking"] = &ThinkingConfig{
				Type:         "enabled",
				BudgetTokens: this.config.ReasonBudgetTokens,
			}
		}

		if !this.config.EnableOutputReason {
			delete(wrapper, "thinking")
		}

		newBody, err := json.Marshal(wrapper)
		if err != nil {
			return request, false, err
		}

		bodyBuff = *bytes.NewBuffer(newBody)

		cloneReq = &http.Request{
			Method: request.Method,
			URL:    request.URL,
			Proto:  request.Proto,
			Header: request.Header.Clone(),
			Body:   io.NopCloser(bytes.NewBuffer(bodyBuff.Bytes())),
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

	bedrockRuntimeEndPoint := fmt.Sprintf(`https://bedrock-runtime.%s.amazonaws.com/model/%s/invoke`, this.config.Region, url.QueryEscape(Model))
	if isStream {
		bedrockRuntimeEndPoint = fmt.Sprintf(`https://bedrock-runtime.%s.amazonaws.com/model/%s/invoke-with-response-stream`, this.config.Region, url.QueryEscape(Model))
	}

	preSignReq, err := http.NewRequest("POST", bedrockRuntimeEndPoint, cloneReq.Body)
	if err != nil {
		Log.Error(err)
		return nil, false, err
	}
	preSignReq.Header.Set("Content-Type", contentType)
	preSignReq.ContentLength = int64(bodyBuff.Len())

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
	err = signer.SignHTTP(context.TODO(), credentialList, preSignReq, payloadHash, "bedrock", cfg.Region, time.Now(), func(options *v4.SignerOptions) {
		if this.config.DEBUG {
			options.LogSigning = true
		}
	})
	if err != nil {
		Log.Error(err)
		return nil, false, err
	}

	return preSignReq, isStream, nil
}

type RawAWSBedrockEvent struct {
	Bytes string `json:"bytes"`
	P     string `json:"p"`
}

func (this *RawAWSBedrockEvent) GetRawChunk() (string, string) {
	jsonRaw, err := base64.StdEncoding.DecodeString(this.Bytes)
	if err != nil {
		Log.Error(err)
	}

	type EventTypeWrapper struct {
		Type string `json:"type"`
	}

	var eventType EventTypeWrapper

	err = json.Unmarshal(jsonRaw, &eventType)
	if err != nil {
		Log.Error(err)
	}

	return eventType.Type, string(jsonRaw)
}
func AsClaudeEvent(line string) string {
	var rawEvent RawAWSBedrockEvent
	decode := json.NewDecoder(strings.NewReader(line))
	err := decode.Decode(&rawEvent)
	if err != nil {
		Log.Error(err)
		return ""
	}
	eventType, raw := rawEvent.GetRawChunk()
	return fmt.Sprintf("event: %s\ndata: %s\n", eventType, raw)
}

func (this *BedrockClient) handleBedrockStream(w http.ResponseWriter, res *http.Response) error {
	// 設置 SSE 相關的 headers
	for k, v := range res.Header {
		w.Header()[k] = v
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	StreamContentType := res.Header.Get("Content-Type")
	BedrockContentType := res.Header.Get("X-Amzn-Bedrock-Content-Type")
	isAWSEventstream := strings.Contains(StreamContentType, "amazon.eventstream")
	isJSONEncoded := strings.Contains(BedrockContentType, "json")

	if this.config.DEBUG {
		Log.Infof("handleBedrockStream: %s", res.Header.Get("Content-Type"))
		Log.Info("Response Header")
		for k, v := range res.Header {
			Log.Infof("%s: %s\n", k, v)
		}
	}

	if !isAWSEventstream {
		return nil
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		return fmt.Errorf("streaming unsupported")
	}

	decoder := eventstream.NewDecoder()

	// 创建缓冲读取器
	buf := make([]byte, 256*1024) // 256k 缓冲区，可根据需要调整

	for {
		// 读取事件头部 (前面的12字节包含总长度等信息)
		msg, err := decoder.Decode(res.Body, buf)
		if err != nil {
			if err == io.EOF {
				break
			}
			return fmt.Errorf("解码错误: %v\n", err)
		}

		//if this.config.DEBUG {
		//	Log.Info("New Message")
		//}

		//for _, header := range msg.Headers {
		//	Log.Infof("头部: %s = %v\n", header.Name, header.Value)
		//}

		if this.config.DEBUG {
			Log.Infof("handleBedrockStreamRaw: %s\n", string(msg.Payload))
		}

		if isJSONEncoded {
			// 查找事件类型和内容 (需要根据EventStream具体格式进一步解析)
			// 简化示例: 假设数据是JSON格式
			SSEEvent := AsClaudeEvent(string(msg.Payload))
			if this.config.DEBUG {
				Log.Infof("SSE: %s\n", SSEEvent)
			}
			// 寫入修改後的行並立即刷新
			fmt.Fprintf(w, "%s\n", SSEEvent)
			flusher.Flush()
		}
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

		if err := this.handleBedrockStream(w, resp); err != nil {
			Log.Error(err)
		}
		return
	}

	httpClient := http.DefaultClient
	if this.config.DEBUG {
		httpClient = &http.Client{
			Transport: loggingRoundTripper{
				wrapped: http.DefaultTransport,
			},
		}
	}

	resp, err := httpClient.Do(cloneReq)
	if err != nil {
		Log.Error(err)
		w.Header().Set("Content-Type", "application/json")
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}

	// 寫入修改後的響應
	for k, v := range resp.Header {
		w.Header()[k] = v
	}
	w.WriteHeader(resp.StatusCode)
	_, err = io.Copy(w, resp.Body)
	if err != nil {
		Log.Error(err)
	}
}
