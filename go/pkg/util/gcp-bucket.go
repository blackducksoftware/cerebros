/*
Copyright (C) 2020 Synopsys, Inc.

Licensed to the Apache Software Foundation (ASF) under one
or more contributor license agreements. See the NOTICE file
distributed with this work for additional information
regarding copyright ownership. The ASF licenses this file
to you under the Apache License, Version 2.0 (the
"License"); you may not use this file except in compliance
with the License. You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing,
software distributed under the License is distributed on an
"AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
KIND, either express or implied. See the License for the
specific language governing permissions and limitations
under the License.
*/
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
