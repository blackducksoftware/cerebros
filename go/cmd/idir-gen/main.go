package main

import (
	"encoding/json"
	"fmt"
	"github.com/blackducksoftware/cerebros/go/pkg/jobrunner"
	"github.com/blackducksoftware/cerebros/go/pkg/util"
	"github.com/streadway/amqp"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
)

const queueName = "idir-gen"

type IdirGenJob struct {
	FromBucket     string `json:"fromBucket"`
	FromBucketPath string `json:"fromBucketPath"`

	ToBucket     string `json:"toBucket"`
	ToBucketPath string `json:"toBucketPath"`
}

func main() {
	rabbitMQHost := getEnv("AMQP_URL", "amqp://guest:guest@localhost:5672/")
	serviceAccountPath := getEnv("GCP_SERVICE_ACCOUNT_PATH", "/Users/jeremyd/Downloads/polaris-dev-233821-b8a3ac17ca0f.json")
	polarisConfig := jobrunner.PolarisConfig{
		PolarisURL:   getEnv("POLARIS_URL", "https://onprem-dev.dev.polaris.synopsys.com"),
		PolarisToken: getEnv("POLARIS_TOKEN", ""),
	}

	env := os.Getenv("CA_PATH")
	if len(env) > 0 {
		input, err := ioutil.ReadFile(env)
		if err != nil {
			panic(err)
		}
		err = ioutil.WriteFile("/usr/local/share/ca-certificates/newCa.crt", input, 0644)
		if err != nil {
			panic(err)
		}
		if err := exec.Command("update-ca-certificates").Run(); err != nil {
			panic(err)
		}
	}

	fmt.Println("Worker starting...")
	conn, err := amqp.Dial(rabbitMQHost)
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	q, err := ch.QueueDeclare(
		queueName, // name
		true,      // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	failOnError(err, "Failed to declare a queue")

	err = ch.Qos(
		1,     // prefetch count
		0,     // prefetch size
		false, // global
	)
	failOnError(err, "Failed to set QoS")

	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		false,  // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	failOnError(err, "Failed to register a consumer")

	jb, err := jobrunner.NewPolarisScanner(polarisConfig)
	if err != nil {
		panic(err)
	}

	// TODO - Use go routines if we ever want to process multiple scans in parallel
	for d := range msgs {
		var job IdirGenJob
		if err := json.Unmarshal(d.Body, &job); err != nil {
			log.Print(err)
			d.Ack(false)
			continue
		}
		fmt.Printf("Starting IdirGenJob %s/%s\n", job.FromBucket, job.FromBucketPath)
		if err := process(job, serviceAccountPath, jb); err != nil {
			log.Print(err)
			d.Ack(false)
			continue
		}
		fmt.Printf("Completed Scan %s/%s\n", job.FromBucket, job.FromBucketPath)
		d.Ack(false)
	}
}

func process(job IdirGenJob, serviceAccountPath string, jb *jobrunner.PolarisScanner) error {
	// Download and extract
	tmpDir, err := downloadAndExtractToTmpDir(job, serviceAccountPath)
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpDir)

	// Capture
	idirPath, err := jb.Capture(tmpDir)
	if err != nil {
		return err
	}

	pathToIdirZip := filepath.Join(tmpDir, "idir.zip")
	if err := util.Zipit(idirPath, pathToIdirZip); err != nil {
		return err
	}

	//Upload
	if err := util.CopyToGSBucket(job.ToBucket, serviceAccountPath, job.ToBucketPath, pathToIdirZip); err != nil {
		return err
	}

	return nil
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

func getEnv(name string, defaultValue string) string {
	if value, ok := os.LookupEnv(name); ok {
		return value
	}
	return defaultValue
}

func downloadAndExtractToTmpDir(job IdirGenJob, svcPath string) (string, error) {
	tmpDir, err := ioutil.TempDir("/tmp", "scan")
	if err != nil {
		return "", err
	}

	fmt.Printf("Downloading from GS Bucket\n")

	// Download the zip archive
	tmpFile := path.Join(tmpDir, "master.zip")
	if err := util.CopyFromGSBucket(job.FromBucket, svcPath, job.FromBucketPath, tmpFile); err != nil {
		return "", err
	}

	// Unzip then remove
	if err := util.Unzip(tmpFile, tmpDir); err != nil {
		return "", err
	}
	return tmpDir, nil
}
