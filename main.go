package main

import (
	"bedrock-claude-proxy/pkg"
	"flag"
	"runtime"
)

func main() {
	conf_path := flag.String("c", "conf.json", "config json file")
	flag.Parse()

	runtime.GOMAXPROCS(runtime.NumCPU())

	conf, err := pkg.NewConfigFromLocal(*conf_path)
	if err != nil {
		pkg.Log.Error(err)
		conf = &pkg.Config{}
	}

	conf.MarginWithENV()

	pkg.Log.Debug("show config detail:")
	pkg.Log.Debug(conf.ToJSON())

	service := pkg.NewHttpService(conf)
	service.Start()
}
