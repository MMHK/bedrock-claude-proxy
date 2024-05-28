package pkg

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"io"
	"net/http"
)

type HttpConfig struct {
	Listen  string `json:"listen,omitempty"`
	WebRoot string `json:"web_root,omitempty"`
}

type HTTPService struct {
	conf *Config
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
	return &HTTPService{
		conf: conf,
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

func (this *HTTPService) HandleComplete(writer http.ResponseWriter, request *http.Request) {
	if request.Method != "POST" {
		this.ResponseError(fmt.Errorf("method not allowed"), writer)
		return
	}
	if request.Header.Get("Content-Type") != "application/json" {
		this.ResponseError(fmt.Errorf("invalid content type"), writer)
		return
	}
	defer request.Body.Close()
	// json decode request body
	var req *ClaudeTextCompletionRequest
	err := json.NewDecoder(request.Body).Decode(&req)
	if err != nil {
		this.ResponseError(err, writer)
		return
	}
	// get anthropic-version,x-api-key from request
	//anthropicVersion := request.Header.Get("anthropic-version")
	//anthropicKey := request.Header.Get("x-api-key")

	bedrockClient := NewBedrockClient(this.conf.BedrockConfig)
	response, err := bedrockClient.CompleteText(req)
	if err != nil {
		this.ResponseError(err, writer)
		return
	}

	if response.IsStream() {
		// output & flush SSE
		flusher, ok := writer.(http.Flusher)
		if !ok {
			this.ResponseError(fmt.Errorf("streaming not supported"), writer)
			return
		}
		writer.Header().Set("Content-Type", "text/event-stream")
		writer.Header().Set("Cache-Control", "no-cache")
		writer.Header().Set("Connection", "keep-alive")


		for event := range response.GetEvents() {
			_, err = writer.Write([]byte(ClaudeTextCompletionStreamEventToSSE(event.Type, event.Completion)))
			if err != nil {
				Log.Error(err)
				continue
			}
			flusher.Flush()
		}
		return
	}

	this.ResponseJSON(response.GetResponse(), writer)
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

	// 读取请求 body
	body, err := io.ReadAll(request.Body)
	if err != nil {
		this.ResponseError(fmt.Errorf("Error reading request body"), writer)
		return
	}
	defer request.Body.Close()


	// json decode request body
	var req ClaudeMessageCompletionRequest
	err = json.Unmarshal(body, &req)
	if err != nil {
		this.ResponseError(err, writer)
		return
	}
	// get anthropic-version,x-api-key from request
	//anthropicVersion := request.Header.Get("anthropic-version")
	//anthropicKey := request.Header.Get("x-api-key")

	req.AnthropicVersion = "bedrock-2023-05-31"

	Log.Debug(string(body))
	Log.Debugf("%+v", req)

	bedrockClient := NewBedrockClient(this.conf.BedrockConfig)
	response, err := bedrockClient.MessageCompletion(&req)
	if err != nil {
		this.ResponseError(err, writer)
		return
	}

	if response.IsStream() {
		// output & flush SSE
		flusher, ok := writer.(http.Flusher)
		if !ok {
			this.ResponseError(fmt.Errorf("streaming not supported"), writer)
			return
		}
		writer.Header().Set("Content-Type", "text/event-stream")
		writer.Header().Set("Cache-Control", "no-cache")
		writer.Header().Set("Connection", "keep-alive")

		for event := range response.GetEvents() {
			_, err = writer.Write([]byte(ClaudeTextCompletionStreamEventToSSE(event.Type, string(event.Raw))))
			if err != nil {
				Log.Error(err)
				continue
			}
			flusher.Flush()
		}
		return
	}

	this.ResponseJSON(response.GetResponse(), writer)
}

func (this *HTTPService) Start() {
	rHandler := mux.NewRouter()

	rHandler.HandleFunc("/v1/complete", this.HandleComplete)
	rHandler.HandleFunc("/v1/messages", this.HandleMessageComplete)
	rHandler.HandleFunc("/", this.RedirectSwagger)
	rHandler.PathPrefix("/").Handler(http.StripPrefix("/",
		http.FileServer(http.Dir(fmt.Sprintf("%s", this.conf.WebRoot)))))
	rHandler.NotFoundHandler = http.HandlerFunc(this.NotFoundHandle)

	Log.Info("http service starting")
	Log.Infof("Please open http://%this\n", this.conf.Listen)
	err := http.ListenAndServe(this.conf.Listen, rHandler)
	if err != nil {
		Log.Error(err)
	}
}
