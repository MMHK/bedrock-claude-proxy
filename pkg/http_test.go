package pkg

import (
	"bedrock-claude-proxy/tests"
	"testing"
)

func TestHTTPService_Start(t *testing.T) {
	conf, err := loadConfig()
	if err != nil {
		t.Fatal(err)
	}
	conf.MarginWithENV()

	t.Log(tests.ToJSON(conf))

	http := NewHttpService(conf)
	http.Start()
}
