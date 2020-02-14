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
	"github.com/blackducksoftware/cerebros/go/pkg/jobrunner"
	"github.com/juju/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"os"
)

type Config struct {
	PolarisCLIPath string
	CapturePath    string
	PolarisURL     string
	PolarisToken   string
}

// GetConfig ...
func GetConfig(configPath string) (*Config, error) {
	var config *Config

	viper.SetConfigFile(configPath)
	err := viper.ReadInConfig()
	if err != nil {
		return nil, errors.Annotatef(err, "failed to ReadInConfig")
	}

	err = viper.Unmarshal(&config)
	if err != nil {
		return nil, errors.Annotatef(err, "failed to unmarshal config")
	}

	return config, nil
}

func main() {
	config, err := GetConfig(os.Args[1])
	scanner, err := jobrunner.NewPolarisScanner(config.PolarisCLIPath, jobrunner.PolarisConfig{
		PolarisURL:   config.PolarisURL,
		PolarisToken: config.PolarisToken,
	})
	if err != nil {
		log.Errorf("%+v", err)
		panic(err)
	}
	err = scanner.CaptureAndScan(config.CapturePath)
	if err != nil {
		panic(err)
	}
	log.Infof("successfully captured and scanned %s", config.CapturePath)

	//scanner.Scan()
}
