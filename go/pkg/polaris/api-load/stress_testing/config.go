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
package stress_testing

import (
	"fmt"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"time"
)

type Config struct {
	PolarisURL      string
	PolarisEmail    string
	PolarisPassword string

	LogLevel string

	Port int

	LoadGenerator struct {
		Issue *IssueServerConfig
		Auth  *AuthConfig
	}
}

type RateConfig struct {
	RateChangePeriodSeconds float64
	Constant                *struct {
		Baseline float64
	}
	Sinusoid *struct {
		Baseline  float64
		Amplitude float64
		Period    float64
		Phase     float64
	}
	Spike *struct {
		Baseline          float64
		LowPeriodSeconds  float64
		Height            float64
		HighPeriodSeconds float64
		RampSeconds       float64
	}
}

func (rc *RateConfig) RateLimiter(name string) (*RateLimiter, error) {
	period := time.Duration(rc.RateChangePeriodSeconds) * time.Second
	if rc.Constant != nil {
		return NewRateLimiter(name, Const(rc.Constant.Baseline), period), nil
	}
	if rc.Sinusoid != nil {
		sin := rc.Sinusoid
		f := Sinusoid(sin.Baseline, sin.Amplitude, sin.Period, sin.Phase)
		return NewRateLimiter(name, f, period), nil
	}
	if rc.Spike != nil {
		s := rc.Spike
		f := Spike(s.Baseline, s.LowPeriodSeconds, s.Height, s.HighPeriodSeconds, s.RampSeconds)
		return NewRateLimiter(name, f, period), nil
	}
	return nil, errors.New(fmt.Sprintf("all RateConfig options nil"))
}

func (rc *RateConfig) MustRateLimiter(name string) *RateLimiter {
	rl, err := rc.RateLimiter(name)
	if err != nil {
		log.Fatalf("unable to instantiate RateLimiter from config: \n%+v\n", err)
		panic(err)
	}
	return rl
}

type LoadConfig struct {
	WorkersCount int
	Rate         *RateConfig
}

type RoleAssignmentsPager struct {
	LoadConfig *LoadConfig
	PageSize   int
}

type AuthConfig struct {
	Entitlements                 *LoadConfig
	Login                        *LoadConfig
	RoleAssignmentsPager         map[string]*RoleAssignmentsPager
	RoleAssignmentsSingleProject *LoadConfig
	CreateRoleAssignments        *LoadConfig
}

type RollupCountsPager struct {
	LoadConfig *LoadConfig
	PageSize   int
}

type IssueServerConfig struct {
	FetchProjectsCount int

	Issues       *LoadConfig
	RollupCounts *RollupCountsPager
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
