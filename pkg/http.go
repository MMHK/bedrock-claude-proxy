package pkg

import (
	"encoding/json"
	"fmt"
	"net/http"
	"github.com/gorilla/mux"
)

type HttpConfig struct {
	Listen  string `json:"listen"`
	WebRoot string `json:"web_root"`
}

type HTTPService struct {
	conf *Config
}

type APIStandardError struct {
	Status bool   `json:"status"`
	Error  string `json:"error"`
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
	server_error := &APIStandardError{Error: "handle not found!", Status: false}
	json_str, _ := json.Marshal(server_error)
	http.Error(writer, string(json_str), 404)
}

func (this *HTTPService) ResponseError(err error, writer http.ResponseWriter) {
	server_error := &APIStandardError{Error: err.Error(), Status: false}
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

func (this *HTTPService) Start() {
	rHandler := mux.NewRouter()

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
