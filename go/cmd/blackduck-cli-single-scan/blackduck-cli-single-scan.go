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
package main

import (
	"encoding/json"
	s "github.com/blackducksoftware/cerebros/go/pkg/synopsys-scancli"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"os"

	log "github.com/sirupsen/logrus"
)

type Config struct {
	Blackduck   *s.BlackduckConfig
	ImageFacade *s.ImageFacadeConfig
	Scan        *s.ScanConfig
}

func GetConfig(configPath string) (*Config, error) {
	var config *Config

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

func doOrDie(err error) {
	if err != nil {
		log.Fatalf("%+v", err)
	}
}

func main() {
	configPath := os.Args[1]
	config, err := GetConfig(configPath)
	doOrDie(err)

	configBytes, err := json.MarshalIndent(config, "", "  ")
	doOrDie(err)
	log.Infof("got config: \n%s\n", string(configBytes))

	scanner, err := s.NewScannerFromConfig(config.Blackduck, nil, config.ImageFacade)
	doOrDie(err)

	err = scanner.Scan(config.Scan)
	doOrDie(err)
}
