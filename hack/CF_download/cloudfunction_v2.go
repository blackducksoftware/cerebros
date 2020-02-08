// Package p contains a Pub/Sub Cloud Function.
package p

import (
	"cloud.google.com/go/storage"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

type PubSubMessage struct {
	Data []byte `json:"data"`
}

func Start(ctx context.Context, m PubSubMessage) error {
	log.Println(string(m.Data))
	var d struct {
		URL  string `json:"url"`
		Name string `json:"name"`
	}

	if err := json.Unmarshal(m.Data, &d); err != nil {
		return err
	}

	if d.URL == "" {
		return fmt.Errorf("No URL")
	}

	if d.Name == "" {

		return fmt.Errorf("No Name")
	}

	fmt.Printf("URL: %s | Name: %s\n", d.URL, d.Name)

	// Get the data
	resp, err := http.Get(d.URL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP error %d", resp.StatusCode)
	}

	client, err := storage.NewClient(ctx)
	if err != nil {
		log.Fatal(err)
	}

	wc := client.Bucket("XXXX").Object(fmt.Sprintf("%s/master.zip", d.Name)).NewWriter(ctx)
	if _, err = io.Copy(wc, resp.Body); err != nil {
		log.Fatal(err)
	}
	if err := wc.Close(); err != nil {
		log.Fatal(err)
	}

	return nil
}
