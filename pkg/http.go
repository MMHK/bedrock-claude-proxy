package pkg

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

type HttpConfig struct {
	Listen  string `json:"listen,omitempty"`
	WebRoot string `json:"web_root,omitempty"`
	APIKey  string `json:"api_key,omitempty"`
}

type HTTPService struct {
	conf          *Config
	bedrockClient *BedrockClient
}

type APIModelInfo struct {
	CreatedAt   string `json:"created_at"`
	DisplayName string `json:"display_name"`
	ID          string `json:"id"`
	Type        string `json:"type"`
}

type ListModelsResponse struct {
	Data    []APIModelInfo `json:"data"`
	FirstID string         `json:"first_id"`
	HasMore bool           `json:"has_more"`
	LastID  string         `json:"last_id"`
}

func (this *HTTPService) HandleListModels(writer http.ResponseWriter, request *http.Request) {
	// Use the new merged model list that validates against Bedrock API
	models, err := this.bedrockClient.GetMergedModelList()
	if err != nil {
		Log.Errorf("Failed to get merged model list: %v", err)
		// Fall back to basic model list
		models = this.bedrockClient.ListModels()
	}

	response := ListModelsResponse{
		Data:    make([]APIModelInfo, 0, len(models)),
		FirstID: "<string>", // You may need to implement logic to determine these values
		HasMore: true,       // You may need to implement pagination logic
		LastID:  "<string>", // You may need to implement logic to determine these values
	}

	for _, model := range models {
		response.Data = append(response.Data, APIModelInfo{
			CreatedAt:   "2025-02-19T00:00:00Z", // You may need to implement logic to get the actual creation date
			DisplayName: model.Name,
			ID:          model.ID,
			Type:        "model",
		})
	}

	this.ResponseJSON(response, writer)
}

func (this *HTTPService) HandleValidateModels(writer http.ResponseWriter, request *http.Request) {
	// Get validation results from Bedrock client
	validationResults, err := this.bedrockClient.ValidateModelMappings()
	if err != nil {
		this.ResponseError(fmt.Errorf("failed to validate models: %v", err), writer)
		return
	}

	// Also get available models from Bedrock for additional info
	availableModels, err := this.bedrockClient.GetBedrockAvailableModels()
	if err != nil {
		Log.Warningf("Failed to get available models: %v", err)
	}

	// Create response structure
	response := map[string]interface{}{
		"validation_results":   validationResults,
		"total_configured":     len(validationResults),
		"available_count":      0,
		"unavailable_count":    0,
		"bedrock_models_count": len(availableModels),
	}

	// Count available vs unavailable
	for _, result := range validationResults {
		if result.Available {
			response["available_count"] = response["available_count"].(int) + 1
		} else {
			response["unavailable_count"] = response["unavailable_count"].(int) + 1
		}
	}

	this.ResponseJSON(response, writer)
}

type APIError struct {
	Type    string `json:"type,omitempty"`
	Message string `json:"message,omitempty"`
}

type APIStandardError struct {
	Type  string    `json:"type,omitempty"`
	Error *APIError `json:"error,omitempty"`
}

func NewHttpService(conf *Config) *HTTPService {
	bedrock := NewBedrockClient(conf.BedrockConfig)

	return &HTTPService{
		conf:          conf,
		bedrockClient: bedrock,
	}
}

func (this *HTTPService) RedirectSwagger(writer http.ResponseWriter, request *http.Request) {
	http.Redirect(writer, request, "/swagger/", 301)
}

func (this *HTTPService) NotFoundHandle(writer http.ResponseWriter, request *http.Request) {
	server_error := &APIStandardError{Type: "error", Error: &APIError{
		Type:    "error",
		Message: "not found",
	}}
	json_str, _ := json.Marshal(server_error)
	http.Error(writer, string(json_str), 404)
}

func (this *HTTPService) ResponseError(err error, writer http.ResponseWriter) {
	server_error := &APIStandardError{Type: "error", Error: &APIError{
		Type:    "invalid_request_error",
		Message: err.Error(),
	}}
	json_str, _ := json.Marshal(server_error)
	http.Error(writer, string(json_str), 200)
}

func (this *HTTPService) ResponseJSON(source interface{}, writer http.ResponseWriter) {
	encoder := json.NewEncoder(writer)
	encoder.SetEscapeHTML(false)

	writer.Header().Set("Content-Type", "application/json")
	err := encoder.Encode(source)
	if err != nil {
		this.ResponseError(err, writer)
	}
}

func (this *HTTPService) HandleMessageComplete(writer http.ResponseWriter, request *http.Request) {
	if request.Method != "POST" {
		this.ResponseError(fmt.Errorf("method not allowed"), writer)
		return
	}
	if request.Header.Get("Content-Type") != "application/json" {
		this.ResponseError(fmt.Errorf("invalid content type"), writer)
		return
	}

	this.bedrockClient.HandleProxy(writer, request)
}

// APIKeyMiddleware 验证 API Key 的中间件
func (this *HTTPService) APIKeyMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		Log.Debug("APIKeyMiddleware")
		APIKey := this.conf.APIKey
		Log.Debugf("APIKeyMiddleware: %s", APIKey)
		if APIKey == "" {
			next.ServeHTTP(writer, request)
			return
		}
		apiKey := request.Header.Get("x-api-key")
		Log.Debugf("API key in header: %s", apiKey)
		if apiKey == "" {
			this.ResponseError(fmt.Errorf("invalid api key"), writer)
			return
		}

		// 这里可以添加更多的 API Key 验证逻辑
		if apiKey != APIKey {
			this.ResponseError(fmt.Errorf("Invalid API key"), writer)
			return
		}

		next.ServeHTTP(writer, request)
	})
}

func (this *HTTPService) Start() {
	rHandler := mux.NewRouter()

	// 需要 API Key 的路由
	apiRouter := rHandler.PathPrefix("/v1").Subrouter()
	apiRouter.Use(this.APIKeyMiddleware)

	apiRouter.HandleFunc("/messages", this.HandleMessageComplete)
	apiRouter.HandleFunc("/models", this.HandleListModels).Methods("GET")
	apiRouter.HandleFunc("/models/validate", this.HandleValidateModels).Methods("GET")

	rHandler.HandleFunc("/", this.RedirectSwagger)
	rHandler.PathPrefix("/").Handler(http.StripPrefix("/",
		http.FileServer(http.Dir(fmt.Sprintf("%s", this.conf.WebRoot)))))
	rHandler.NotFoundHandler = http.HandlerFunc(this.NotFoundHandle)

	Log.Info("http service starting")
	Log.Infof("Please open http://%s\n", this.conf.Listen)
	err := http.ListenAndServe(this.conf.Listen, rHandler)
	if err != nil {
		Log.Error(err)
	}
}
