package util

import (
	"cloud.google.com/go/storage"
	"context"
	"io"
	"os"
)

func CopyToGSBucket(bucketName, serviceAccountPath, bucketFilePath, localFilePath string) error {
	err := os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", serviceAccountPath)
	if err != nil {
		return err
	}

	ctx := context.Background()

	f, err := os.Open(localFilePath)
	if err != nil {
		return err
	}
	defer f.Close()

	client, err := storage.NewClient(ctx)
	if err != nil {
		return err
	}

	wc := client.Bucket(bucketName).Object(bucketFilePath).NewWriter(ctx)
	if _, err = io.Copy(wc, f); err != nil {
		return err
	}
	if err := wc.Close(); err != nil {
		return err
	}

	return nil
}

func CopyFromGSBucket(bucketName, serviceAccountPath, bucketFilePath, localFilePath string) error {
	err := os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", serviceAccountPath)
	if err != nil {
		return err
	}

	ctx := context.Background()

	f, err := os.Create(localFilePath)
	if err != nil {
		return err
	}
	defer f.Close()

	client, err := storage.NewClient(ctx)
	if err != nil {
		return err
	}

	wc, err := client.Bucket(bucketName).Object(bucketFilePath).NewReader(ctx)
	if err != nil {
		return err
	}

	if _, err = io.Copy(f, wc); err != nil {
		return err
	}

	return nil
}
