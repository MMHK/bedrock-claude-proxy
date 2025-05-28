package pkg

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type HttpConfig struct {
	Listen   string `json:"listen,omitempty"`
	WebRoot  string `json:"web_root,omitempty"`
	APIKey   string `json:"api_key,omitempty"`
	ZohoAuth bool   `json:"zoho_auth,omitempty"`
}

type HTTPService struct {
	conf          *Config
	bedrockClient *BedrockClient
	zohoAuth      *ZohoOAuth
	ApiStorage    APIKeyStore
	apiKeysMutex  sync.RWMutex
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
		FirstID: "",    // You may need to implement logic to determine these values
		HasMore: false, // You may need to implement pagination logic
		LastID:  "",    // You may need to implement logic to determine these values
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
	zohoConfig := LoadZohoConfigFromEnv()
	var cache APIKeyStore
	cache, err := NewCache()
	if err != nil {
		Log.Errorf("Failed to create cache: %v", err)
		cache = NewMemoryStore(24 * time.Hour) // Fallback to in-memory store if cache creation fails
	}

	return &HTTPService{
		conf:          conf,
		bedrockClient: bedrock,
		zohoAuth:      NewZohoOAuth(zohoConfig),
		ApiStorage:    cache,
	}
}

func (this *HTTPService) generateAPIKey() string {
	id := uuid.New()
	return id.String()
}

func (z *HTTPService) buildRedirectURIFromRequest(r *http.Request) string {
	scheme := "http"
	if r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https" {
		scheme = "https"
	}

	host := r.Host
	if forwardedHost := r.Header.Get("X-Forwarded-Host"); forwardedHost != "" {
		host = forwardedHost
	}

	return fmt.Sprintf("%s://%s/auth/callback", scheme, host)
}

func (this *HTTPService) HandleAuth(writer http.ResponseWriter, request *http.Request) {
	redirectURL := this.zohoAuth.GetAuthRedirectURLCustom(this.buildRedirectURIFromRequest(request))
	http.Redirect(writer, request, redirectURL, http.StatusTemporaryRedirect)
}

func (this *HTTPService) HandleAuthCallback(writer http.ResponseWriter, request *http.Request) {
	code := request.URL.Query().Get("code")
	if code == "" {
		this.ResponseError(fmt.Errorf("missing authorization code"), writer)
		return
	}

	email, err := this.zohoAuth.GetEmailCustom(code, this.buildRedirectURIFromRequest(request))
	if err != nil {
		this.ResponseError(fmt.Errorf("failed to get email: %v", err), writer)
		return
	}

	if !this.zohoAuth.IsEmailDomainAllowed(email) {
		this.ResponseError(fmt.Errorf("email domain not allowed"), writer)
		return
	}

	type ExistApiKey struct {
		APIKey    string `json:"api_key"`
		Email     string `json:"email"`
		ExpiredAt int64  `json:"expired_at,omitempty"`
	}

	emailExist, err := this.ApiStorage.HasAPIKey(email)
	if err == nil && emailExist {
		apiKeyJSON, err := this.ApiStorage.GetAPIKey(email)
		if err == nil {
			var existApiKey ExistApiKey
			err := json.Unmarshal([]byte(apiKeyJSON), &existApiKey)
			if err == nil {
				response := ExistApiKey{
					APIKey:    existApiKey.APIKey,
					Email:     existApiKey.Email,
					ExpiredAt: existApiKey.ExpiredAt,
				}

				this.ResponseJSON(response, writer)
				return
			}

		}
	}

	apiKey := this.generateAPIKey()

	this.apiKeysMutex.Lock()
	err = this.ApiStorage.SaveAPIKey(apiKey, email, 24*time.Hour)
	if err != nil {
		this.ResponseError(fmt.Errorf("failed to save API key: %v", err), writer)
		return
	}
	keyRecordsBin, err := json.Marshal(ExistApiKey{
		APIKey:    apiKey,
		Email:     email,
		ExpiredAt: time.Now().Add(24 * time.Hour).Unix(),
	})
	if err != nil {
		this.ResponseError(fmt.Errorf("failed to marshal API key record: %v", err), writer)
		return
	}
	err = this.ApiStorage.SaveAPIKey(email, string(keyRecordsBin), 24*time.Hour)
	if err != nil {
		this.ResponseError(fmt.Errorf("failed to save API key: %v", err), writer)
		return
	}
	this.apiKeysMutex.Unlock()

	response := ExistApiKey{
		APIKey:    apiKey,
		Email:     email,
		ExpiredAt: time.Now().Add(24 * time.Hour).Unix(),
	}

	this.ResponseJSON(response, writer)
}

func (this *HTTPService) RedirectSwagger(writer http.ResponseWriter, request *http.Request) {
	http.Redirect(writer, request, "/swagger/", 301)
}

func (this *HTTPService) RedirectLanding(writer http.ResponseWriter, request *http.Request) {
	http.Redirect(writer, request, "/landing.html", 301)
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

		userApiKeyExist, err := this.ApiStorage.HasAPIKey(apiKey)
		if err != nil {
			Log.Errorf("Failed to check API key existence: %v", err)
			userApiKeyExist = false
		}

		// 这里可以添加更多的 API Key 验证逻辑
		if apiKey != APIKey && !userApiKeyExist {
			this.ResponseError(fmt.Errorf("Invalid API key"), writer)
			return
		}

		next.ServeHTTP(writer, request)
	})
}

func (this *HTTPService) Start() {
	rHandler := mux.NewRouter()

	defer this.ApiStorage.Close()

	// Add auth routes
	rHandler.HandleFunc("/auth", this.HandleAuth)
	rHandler.HandleFunc("/auth/callback", this.HandleAuthCallback)

	// 需要 API Key 的路由
	apiRouter := rHandler.PathPrefix("/v1").Subrouter()
	apiRouter.Use(this.APIKeyMiddleware)

	apiRouter.HandleFunc("/messages", this.HandleMessageComplete)
	apiRouter.HandleFunc("/models", this.HandleListModels).Methods("GET")

	rHandler.HandleFunc("/", this.RedirectLanding)
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
