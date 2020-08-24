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
package synopsys_scancli

import (
	"encoding/json"
	"github.com/blackducksoftware/cerebros/go/pkg/blackduck/docker"
	"github.com/blackducksoftware/cerebros/go/pkg/blackduck/hubcli"
	polarisapi "github.com/blackducksoftware/cerebros/go/pkg/polaris/api"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
)

type ScanQueueConfig struct {
	Host string
	Port int
}

type BlackduckConfig struct {
	Host                 string
	Username             string
	Password             string
	Port                 int
	OSType               hubcli.OSType
	ClientTimeoutSeconds int
}

type ImageFacadeConfig struct {
	ImageDirectory          string
	ImagePullerType         string
	PrivateDockerRegistries []docker.RegistryAuth
	DeleteImageAfterScan    bool
}

type PolarisConfig struct {
	CLIPath  string
	URL      string
	Email    string
	Password string
	OSType   polarisapi.OSType
	JavaHome string
}

type Config struct {
	Blackduck   *BlackduckConfig
	ImageFacade *ImageFacadeConfig
	ScanQueue   *ScanQueueConfig
	//Scanner     *ScannerConfig // TODO do we need this for anything?
	Polaris *PolarisConfig

	LogLevel string
	Port     int
}

func (config *Config) getLogLevel() (log.Level, error) {
	return log.ParseLevel(config.LogLevel)
}

func GetConfig(configPath string) (*Config, error) {
	bytes, err := ioutil.ReadFile(configPath)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to read in config file %s", configPath)
	}

	var config Config
	err = json.Unmarshal(bytes, &config)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to unmarshal json from file %s", configPath)
	}

	// can't use viper if the config contains anything depending on UnmarshalJSON for its deserialization:
	// viper ignores those hooks
	// see: https://github.com/spf13/viper/issues/338
	//var config *Config
	//
	//viper.SetConfigFile(configPath)
	//err := viper.ReadInConfig()
	//if err != nil {
	//	return nil, errors.Wrap(err, "failed to ReadInConfig")
	//}
	//
	//err = viper.Unmarshal(&config)
	//if err != nil {
	//	return nil, errors.Wrap(err, "failed to unmarshal config")
	//}

	return &config, nil
}
