package main

import (
	"encoding/json"
	"fmt"
	"os"

	util "github.com/blackducksoftware/cerebros/go/pkg/util"
	log "github.com/sirupsen/logrus"
)

func main() {
	var configPath string
	if len(os.Args) > 1 {
		configPath = os.Args[1]
	}
	log.Infof("Config path: %s", configPath)
	pq := Queue{[]string{}}
	bytes, err := json.Marshal(pq)
	if err != nil {
		panic(err)
	}
	print(string(bytes))
	print("\n\n")

	pq.add("http://stsstore/qatest/swip-test-packages/scalability/cs/Dnn.zip")

	rootDownloadDir := "/tmp/downloaded"
	i := 0
	for !pq.isEmpty() {
		url, err := pq.pop()
		if err != nil {
			break
		}
		filepath := fmt.Sprintf("%s-%d.zip", rootDownloadDir, i)
		unzipDir := fmt.Sprintf("%s-unzipped-%d", rootDownloadDir, i)
		downloadAndScan(filepath, unzipDir, url)
		i += 1
	}

}

type Queue struct {
	items []string
}

func (pq *Queue) add(item string) {
	pq.items = append(pq.items, item)
}

func (pq *Queue) pop() (string, error) {
	if len(pq.items) == 0 {
		return "", fmt.Errorf("queue is empty")
	}
	first, rest := pq.items[0], pq.items[1:]
	pq.items = rest
	return first, nil
}

func (pq *Queue) isEmpty() bool {
	return len(pq.items) == 0
}

func downloadAndScan(filepath string, unzipDir string, url string) {
	log.Infof("downloading file from %s to %s", url, filepath)
	err := util.DownloadFile(filepath, url)
	if err != nil {
		panic(err)
	}
	log.Infof("unzipping file %s to %s", filepath, unzipDir)
	err = util.Unzip(filepath, unzipDir)
	if err != nil {
		panic(err)
	}
	log.Infof("successfully downloaded")
}
