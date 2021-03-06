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

package scanqueue

import (
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// Config ...
type Config struct {
	Port int

	Jobs map[string]interface{}

	LogLevel string
}

// GetLogLevel ...
func (config *Config) GetLogLevel() (log.Level, error) {
	return log.ParseLevel(config.LogLevel)
}

// GetConfig ...
func GetConfig(configPath string) (*Config, error) {
	var config *Config

	// avoid viper serialization problems caused by '.'s in keys
	v := viper.NewWithOptions(viper.KeyDelimiter("*"))

	v.SetConfigFile(configPath)
	err := v.ReadInConfig()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to ReadInConfig")
	}

	err = v.Unmarshal(&config)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to unmarshal config")
	}

	return config, nil
}
