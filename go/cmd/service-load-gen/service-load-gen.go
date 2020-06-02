package main

import (
	"github.com/blackducksoftware/cerebros/go/pkg/polaris/api-load/stress_testing"
	"os"
)

func main() {
	stress_testing.RunIssueServerLoadGenerator(os.Args[1])
}
