package pkg

import (
	"bytes"
	"bedrock-claude-proxy/tests"
	"encoding/json"
	"testing"
)

func loadConfig() (*Config, error) {
	conf, err := NewConfigFromLocal(tests.GetLocalPath("../config.json"))
	if err != nil {
		return nil, err
	}
	return conf, nil
}

func TestNewConfigFromLocal(t *testing.T) {
	conf, err := NewConfigFromLocal(tests.GetLocalPath("../config.json"))
	if err != nil {
		t.Error(err)
		t.Fail()
		return
	}

	t.Logf("%+v", conf)
	t.Log("PASS")
}

func TestConfig_MarginWithENV(t *testing.T) {
	conf, err := NewConfigFromLocal(tests.GetLocalPath("../config.json"))
	if err != nil {
		t.Error(err)

		conf = &Config{}
	}

	conf.MarginWithENV()


	t.Logf("%+v", conf)
	jsonBin, err := json.Marshal(conf)
	if err != nil {
		t.Error(err)
	} else {
		var str bytes.Buffer
		_ = json.Indent(&str, jsonBin, "", "  ")
		t.Log(str.String())
	}
	t.Log("PASS")
}

func Test_SaveConfig(t *testing.T) {

	conf := &Config{}
	conf.MarginWithENV()

	configPath := tests.GetLocalPath("../config.json")
	err := conf.Save(configPath)
	if err != nil {
		t.Log(err)
		t.Fail()
	}

	t.Log("PASS")
}