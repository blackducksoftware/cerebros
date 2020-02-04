package main

import (
	"encoding/json"
	"fmt"
	"github.com/blackducksoftware/cerebros/go/pkg/jobrunner"
	"github.com/streadway/amqp"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
)

const queueName = "idir-scanner"

func main() {
	rabbitMQHost := getEnv("AMQP_URL", "amqp://guest:guest@localhost:5672/")
	serviceAccountPath := getEnv("GCP_SERVICE_ACCOUNT_PATH", "/Users/jeremyd/Downloads/polaris-dev-233821-b8a3ac17ca0f.json")
	polarisConfig := jobrunner.PolarisConfig{
		PolarisURL:      getEnv("POLARIS_URL", "https://onprem-dev.dev.polaris.synopsys.com"),
		PolarisEmailID:  getEnv("POLARIS_EMAIL", "jeremyd@synopsys.com"),
		PolarisPassword: getEnv("POLARIS_PASSWORD", "Synopsys123$"),
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
		true,   // durable
		false,  // delete when unused
		false,  // exclusive
		false,  // no-wait
		nil,    // arguments
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

	// TODO - Use go routines if we ever want to process multiple scans in parallel
	for d := range msgs {
		log.Printf("Received a message: %s", d.Body)

		var job jobrunner.IdirScanJob
		if err := json.Unmarshal(d.Body, &job); err != nil {
			d.Ack(false)
			continue
		}
		fmt.Printf("Starting Scan %s/%s\n", job.FromBucket, job.FromBucketPath)
		if err := jobrunner.ScanIdir(job, polarisConfig, serviceAccountPath); err != nil {
			log.Print(err)
		}
		d.Ack(false)
	}
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
