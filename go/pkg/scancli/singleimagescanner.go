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
	"github.com/blackducksoftware/cerebros/go/pkg/scancli/docker"
	"github.com/blackducksoftware/cerebros/go/pkg/scancli/hubcli"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"os"
)


type singleImageScanConfig struct {
	HubURL string
	HubUsername string
	HubPassword string
	HubPort int

	HubProjectName string
	HubProjectVersionName string
	HubScanName string

	ImageRepository string
	ImageSha string

	ImageDirectory string

	LogLevel string
}

func (config *singleImageScanConfig) getLogLevel() (log.Level, error) {
	return log.ParseLevel(config.LogLevel)
}

func getSingleImageScanConfig(configPath string) (*singleImageScanConfig, error) {
	var config *singleImageScanConfig

	viper.SetConfigFile(configPath)
	err := viper.ReadInConfig()
	if err != nil {
		return nil, errors.Wrap(err, "failed to ReadInConfig")
	}

	err = viper.Unmarshal(&config)
	if err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal config")
	}

	return config, nil
}

func ScanImage(configPath string) error {
	config, err := getSingleImageScanConfig(configPath)
	if err != nil {
		return errors.Wrap(err, "Failed to load configuration")
	}

	level, err := config.getLogLevel()
	if err != nil {
		return errors.Wrap(err, "unable to get log level")
	}

	log.SetLevel(level)

	hisr := &hubcli.HubImageScanRequest{
		Repository:            config.ImageRepository,
		Sha:                   config.ImageSha,
		HubURL:                config.HubURL,
		HubProjectName:        config.HubProjectName,
		HubProjectVersionName: config.HubProjectVersionName,
		HubScanName:           config.HubScanName,
	}

	return runScan(hisr, config.HubUsername, config.HubPassword, config.HubPort, config.ImageDirectory)
}

func runScan(hisr *hubcli.HubImageScanRequest, hubUsername string, hubPassword string, hubPort int, imageDirectory string) error {
	imagePuller := docker.NewImagePuller([]docker.RegistryAuth{})
	stop := make(chan struct{})
	scanClient, err := hubcli.NewScanClient(hubUsername, hubPassword, hubPort, hubcli.OSTypeMac)
	if err != nil {
		return err
	}
	hubScanner := hubcli.NewImageScanner(imagePuller, scanClient, imageDirectory, stop)

	err = os.MkdirAll(imageDirectory, 0755)
	if err != nil {
		return errors.Wrapf(err, "unable to make dir %s", imageDirectory)
	}

	return hubScanner.ScanFullDockerImage(hisr)
}
