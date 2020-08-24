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
package main

import (
	"fmt"
	"github.com/blackducksoftware/cerebros/go/pkg/polaris"
	"github.com/blackducksoftware/cerebros/go/pkg/polaris/api"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"os"
)

type Config struct {
	PolarisEmail    string
	PolarisURL      string
	PolarisPassword string

	LogLevel string

	DownloadDir string
	OSType      string
	TokenName   string
	CapturePath string
}

// GetLogLevel ...
func (config *Config) GetLogLevel() (log.Level, error) {
	return log.ParseLevel(config.LogLevel)
}

// GetConfig ...
func GetConfig(configPath string) (*Config, error) {
	var config *Config

	viper.SetConfigFile(configPath)
	err := viper.ReadInConfig()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to ReadInConfig at %s", configPath)
	}

	err = viper.Unmarshal(&config)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to unmarshal config at %s", configPath)
	}

	return config, nil
}

func main() {
	configPath := os.Args[1]
	config, err := GetConfig(configPath)
	if err != nil {
		panic(err)
	}

	logLevel, err := config.GetLogLevel()
	if err != nil {
		panic(err)
	}
	log.SetLevel(logLevel)

	polarisClient := api.NewClient(config.PolarisURL, config.PolarisEmail, config.PolarisPassword)

	err = polarisClient.Authenticate()
	if err != nil {
		panic(err)
	}
	log.Infof("successfully authenticated")

	log.Infof("about to make dir %s", config.DownloadDir)
	err = os.MkdirAll(config.DownloadDir, 0755)
	if err != nil {
		panic(err)
	}

	osType, err := api.ParseOSType(config.OSType)
	if err != nil {
		panic(err)
	}
	unzipPath, err := polarisClient.DownloadCli(config.DownloadDir, osType)
	if err != nil {
		panic(err)
	}

	log.Infof("successfully downloaded polaris-cli to %s", unzipPath)

	cliPath := fmt.Sprintf("%s/bin", unzipPath)
	authResp, err := polarisClient.GetAccessToken(config.TokenName)
	if err != nil {
		panic(err)
	}
	log.Infof("successfully got access token %s", authResp.Data.Attributes.AccessToken)

	scanner, err := polaris.NewScanner(cliPath, config.PolarisURL, authResp.Data.Attributes.AccessToken, "")
	if err != nil {
		panic(err)
	}
	log.Infof("successfully instantiated and configured polaris scanner at %s", cliPath)

	err = scanner.CaptureAndScan(config.CapturePath, false)
	if err != nil {
		panic(err)
	}
	log.Infof("successfully captured/scanned %s", config.CapturePath)
}
