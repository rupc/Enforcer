package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"

	"github.com/rupc/audit/server"
)

// The main function is a server entry point.
// After this, the AuditCore application will serve http request, as defined in server/handlers/***_handlers.go
func main() {
	var configFile string
	flag.StringVar(&configFile, "configFilePath", "../config.json", "Locate config file path")
	flag.Parse()

	src_json, err := ioutil.ReadFile(configFile)
	if err != nil {
		panic(err)
	}
	var m map[string]string

	err = json.Unmarshal(src_json, &m)
	if err != nil {
		panic(err)
	}

	listenAddress := "localhost" + ":" + m["port"]
	version := m["version"]

	o := server.Options{
		ListenAddress: listenAddress,
		Version:       version,
	}
	s := server.NewSystem(o)
	err = s.Start()
	if err != nil {
		fmt.Println("something wrong", err)
	}

	ch := make(chan string)
	ch <- ""
}
