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
	"github.com/blackducksoftware/cerebros/go/pkg/blackduck/docker"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"os"
	"time"
)

func main() {
	configPath := os.Args[1]
	config, err := GetConfig(configPath)
	if err != nil {
		panic(err)
	}

	dir := "/tmp/images"
	err = os.MkdirAll(dir, 0755)
	if err != nil {
		panic(err)
	}
	skopeoDir := "/tmp/images-skopeo"
	err = os.MkdirAll(skopeoDir, 0755)
	if err != nil {
		panic(err)
	}

	dockerPuller := docker.NewImagePuller(dir, []docker.RegistryAuth{})
	pr, err := dockerPuller.PullImage(config.Image)
	if err != nil {
		log.Errorf("unable to pull docker image: %+v", err)
	} else {
		log.Infof("successfully pulled docker image %s to %+v", config.Image, pr)
	}

	skopeoPuller := docker.NewSkopeoImagePuller(skopeoDir, []docker.RegistryAuth{})
	spr, err := skopeoPuller.PullImage(config.Image)
	if err != nil {
		log.Errorf("unable to pull skopeo image: %+v", err)
	} else {
		log.Infof("successfully pulled skopeo image %s to %+v", config.Image, spr)
	}

	time.Sleep(time.Duration(config.PostPullWaitSeconds) * time.Second)
}

// Config ...
type Config struct {
	Image               string
	PostPullWaitSeconds int
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
