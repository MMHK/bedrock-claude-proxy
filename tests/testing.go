package tests

import (
	"bytes"
	"encoding/json"
	"github.com/joho/godotenv"
	"github.com/op/go-logging"
	"path/filepath"
	"runtime"
)

// preload config in testing
func init() {
	format := logging.MustStringFormatter(
		`Email2DB %{color} %{shortfunc} %{level:.4s} %{shortfile}
%{id:03x}%{color:reset} %{message}`,
	)
	logging.SetFormatter(format)
	log := logging.MustGetLogger("BedrockProxy")

	err := godotenv.Load(GetLocalPath("../.env"))
	if err != nil {
		log.Error("Error loading environment")
	}
}

func GetLocalPath(file string) string {
	_, filename, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(filename), file)
}

func ToJSON(target interface{}) string {
	str := new(bytes.Buffer)
	encoder := json.NewEncoder(str)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "    ")
	err := encoder.Encode(target)
	if err != nil {
		return err.Error()
	}

	return str.String()
}
