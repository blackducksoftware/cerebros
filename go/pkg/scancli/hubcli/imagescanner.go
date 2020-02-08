/*
Copyright (C) 2018 Synopsys, Inc.

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

package hubcli

import (
	"fmt"
	"github.com/blackducksoftware/cerebros/go/pkg/scancli/docker"
	"github.com/juju/errors"
	log "github.com/sirupsen/logrus"
	"os"
)

type HubImageScanRequest struct {
	Repository            string
	Sha                   string
	HubURL                string
	HubProjectName        string
	HubProjectVersionName string
	HubScanName           string
}

func (hisr *HubImageScanRequest) pullSpec() string {
	return fmt.Sprintf("%s@sha256:%s", hisr.Repository, hisr.Sha)
}

func (hisr *HubImageScanRequest) dockerImage(directory string) *docker.Image {
	return docker.NewImage(directory, hisr.pullSpec())
}

// ImageScanner ...
type ImageScanner struct {
	imagePuller    docker.ImagePullerInterface
	scanClient     ScanClientInterface
	imageDirectory string
	stop           <-chan struct{}
}

// NewImageScanner ...
func NewImageScanner(imagePuller docker.ImagePullerInterface, scanClient ScanClientInterface, imageDirectory string, stop <-chan struct{}) *ImageScanner {
	return &ImageScanner{
		imagePuller:    imagePuller,
		scanClient:     scanClient,
		imageDirectory: imageDirectory,
		stop:           stop}
}

// ScanFullDockerImage runs the scan client on a full tar from 'docker export'
func (scanner *ImageScanner) ScanFullDockerImage(hisr *HubImageScanRequest) error {
	image := hisr.dockerImage(scanner.imageDirectory)
	err := scanner.imagePuller.PullImage(*image)
	if err != nil {
		return errors.Trace(err)
	}
	defer cleanUpFile(image.DockerTarFilePath())
	return scanner.ScanFile(hisr.HubURL, image.DockerTarFilePath(), hisr.HubProjectName, hisr.HubProjectVersionName, hisr.HubScanName)
}

// ScanFile runs the scan client against a single file
func (scanner *ImageScanner) ScanFile(host string, path string, hubProjectName string, hubVersionName string, hubScanName string) error {
	return scanner.scanClient.Scan(host, path, hubProjectName, hubVersionName, hubScanName)
}

func cleanUpFile(path string) {
	err := os.Remove(path)
	recordCleanUpFile(err == nil)
	if err != nil {
		log.Errorf("unable to remove file %s: %s", path, err.Error())
	} else {
		log.Infof("successfully cleaned up file %s", path)
	}
}
