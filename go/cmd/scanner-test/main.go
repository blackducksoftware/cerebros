package main

import (
	"encoding/json"
	"github.com/blackducksoftware/cerebros/go/pkg/scancli/docker"
)

func main() {
	ra := docker.RegistryAuth{}
	print(json.Marshal(ra))
}
