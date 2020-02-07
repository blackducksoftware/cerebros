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

package scancli

import (
	"fmt"
	"github.com/blackducksoftware/cerebros/go/pkg/scancli/docker"
	"github.com/blackducksoftware/cerebros/go/pkg/scancli/hubcli"
	"os"
	"time"

	"github.com/blackducksoftware/perceptor/pkg/api"
	"github.com/juju/errors"
	log "github.com/sirupsen/logrus"
)

const (
	requestScanJobPause = 20 * time.Second
)

// Scanner ...
type Scanner struct {
	imageScanner         *hubcli.ImageScanner
	perceptorClient *PerceptorClient
	stop            <-chan struct{}
}

// NewScanner ...
func NewScanner(config *Config, stop <-chan struct{}) (*Scanner, error) {
	log.Infof("instantiating Scanner with config %+v", config)

	hubPassword, ok := os.LookupEnv(config.Hub.PasswordEnvVar)
	if !ok {
		return nil, fmt.Errorf("unable to get Hub password: environment variable %s not set", config.Hub.PasswordEnvVar)
	}

	imagePuller := docker.NewImagePuller(config.ImageFacade.PrivateDockerRegistries)
	scanClient, err := hubcli.NewScanClient(
		config.Hub.User,
		hubPassword,
		config.Hub.Port)
	if err != nil {
		return nil, errors.Annotatef(err, "unable to instantiate hub scan client")
	}

	return &Scanner{
		imageScanner:         hubcli.NewImageScanner(imagePuller, scanClient, config.Scanner.GetImageDirectory(), stop),
		perceptorClient: NewPerceptorClient(config.ScanQueue.Host, config.ScanQueue.Port),
		stop:            stop}, nil
}

// StartRequestingScanJobs will start asking for work
func (sm *Scanner) StartRequestingScanJobs() {
	log.Infof("starting to request scan jobs")
	go func() {
		for {
			select {
			case <-sm.stop:
				return
			case <-time.After(requestScanJobPause):
				sm.requestAndRunScanJob()
			}
		}
	}()
}

func (sm *Scanner) requestAndRunScanJob() {
	log.Debug("requesting scan job")
	nextImage, err := sm.perceptorClient.GetNextImage()
	if err != nil {
		log.Errorf("unable to request scan job: %s", err.Error())
		return
	}
	if nextImage.ImageSpec == nil {
		log.Debug("requested scan job, got nil")
		return
	}

	log.Infof("processing scan job %+v", nextImage)

	err = sm.imageScanner.ScanFullDockerImage(nextImage.ImageSpec)
	errorString := ""
	if err != nil {
		log.Errorf("scan error: %s", err.Error())
		errorString = err.Error()
	}

	finishedJob := api.FinishedScanClientJob{Err: errorString, ImageSpec: *nextImage.ImageSpec}
	log.Infof("about to finish job, going to send over %+v", finishedJob)
	sm.perceptorClient.PostFinishedScan(&finishedJob)
	if err != nil {
		log.Errorf("unable to finish scan job: %s", err.Error())
	}
}
