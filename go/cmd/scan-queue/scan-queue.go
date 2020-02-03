package main

import (
	"encoding/json"
	//"fmt"
	"os"

	scanqueue "github.com/blackducksoftware/cerebros/go/pkg/scanqueue"
	log "github.com/sirupsen/logrus"
)

func main() {
	var configPath string
	if len(os.Args) > 1 {
		configPath = os.Args[1]
	}
	log.Infof("Config path: %s", configPath)

	model := scanqueue.NewModel()
	bytes, err := json.Marshal(model)
	if err != nil {
		panic(err)
	}
	print(string(bytes))
	print("\n\n")
}