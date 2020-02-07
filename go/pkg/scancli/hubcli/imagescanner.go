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
	"os"
	"strings"

	"github.com/blackducksoftware/cerebros/go/pkg/scancli/docker"
	"github.com/blackducksoftware/perceptor-scanner/pkg/common"
	"github.com/blackducksoftware/perceptor/pkg/api"
	"github.com/juju/errors"
	log "github.com/sirupsen/logrus"
)

type HubImageScan struct {
	Repository string
	Sha string
	HubURL string
	HubProjectName string
	HubProjectVersionName string
	HubScanName string
}

func (his *HubImageScan)pullSpec() string {
	return fmt.Sprintf("%s@sha256:%s", his.Repository, his.Sha);
}

func (his *HubImageScan) DockerTarFilePath() string {
	return fmt.Sprintf("%s/%s.tar", image.Directory, strings.Replace(image.PullSpec, "/", "_", -1))
}

// ImageScanner ...
type ImageScanner struct {
	ifClient       ImageFacadeClientInterface
	scanClient     ScanClientInterface
	imageDirectory string
	stop           <-chan struct{}
}

// NewImageScanner ...
func NewImageScanner(ifClient ImageFacadeClientInterface, scanClient ScanClientInterface, imageDirectory string, stop <-chan struct{}) *ImageScanner {
	return &ImageScanner{
		ifClient:       ifClient,
		scanClient:     scanClient,
		imageDirectory: imageDirectory,
		stop:           stop}
}

// ScanFullDockerImage runs the scan client on a full tar from 'docker export'
func (scanner *ImageScanner) ScanFullDockerImage(image *HubImageScan) error {
	pullSpec := image.pullSpec()
	image := common.NewImage(scanner.imageDirectory, pullSpec)
	err := scanner.ifClient.PullImage(image)
	if err != nil {
		return errors.Trace(err)
	}
	defer cleanUpFile(image.DockerTarFilePath())
	return scanner.ScanFile(apiImage.HubURL, image.DockerTarFilePath(), apiImage.HubProjectName, apiImage.HubProjectVersionName, apiImage.HubScanName)
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
