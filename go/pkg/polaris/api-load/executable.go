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
package api_load

import (
	"fmt"
	"github.com/blackducksoftware/cerebros/go/pkg/polaris/api"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	"net/http"
	"os"
	"time"
)

func doOrDie(err error) {
	if err != nil {
		log.Fatalf("%+v", err)
	}
}

func Run() {
	configPath := os.Args[1]
	config, err := GetConfig(configPath)
	doOrDie(err)

	log.Infof("config: \n%+v\n%+v\n%+v\n", config, config.DataSeeder, config.LoadGenerator)

	log.Infof("config: %+v", config)

	logLevel, err := config.GetLogLevel()
	doOrDie(err)
	log.SetLevel(logLevel)

	// Prometheus and http setup
	prometheus.Unregister(prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{}))
	prometheus.Unregister(prometheus.NewGoCollector())

	http.Handle("/metrics", promhttp.Handler())

	addr := fmt.Sprintf(":%d", config.Port)
	go func() {
		log.Infof("starting HTTP server on port %d", config.Port)
		http.ListenAndServe(addr, nil)
	}()

	stop := make(chan struct{})

	client := api.NewClient(config.PolarisURL, config.PolarisEmail, config.PolarisPassword)
	err = client.Authenticate()
	doOrDie(err)
	log.Infof("successfully authenticated")
	startRegularReauthentication(client, stop)

	if config.LoadGenerator != nil {
		loadGenerator := NewLoadGenerator(client, config.LoadGenerator.WorkerRequests, stop)
		log.Infof("starting load generator")
		loadGenerator.StartGeneratingLoad()
	}

	if config.DataSeeder != nil {
		log.Infof("instantiating data seeder")
		dataSeeder, err := NewDataSeeder(client, config.DataSeeder.UsersToCreate, config.DataSeeder.Concurrency)
		doOrDie(err)
		log.Infof("starting data seeder: create role assignments")
		dataSeeder.CreateRoleAssignments(stop)
	}

	<-stop
}

func startRegularReauthentication(client *api.Client, stop <-chan struct{}) {
	go func() {
	ForLoop:
		for {
			select {
			case <-stop:
				break ForLoop
			case <-time.After(15 * time.Minute):
			}
			log.Infof("attempting to authenticate into polaris")
			err := client.Authenticate()
			recordEvent("authenticate", err)
			if err != nil {
				log.Errorf("unable to re-authenticate with polaris: %s", err)
			} else {
				log.Infof("successfully re-authenticated into polaris")
			}
			log.Infof("waiting 15 minutes to re-authenticate into polaris")
		}
		log.Infof("exiting authentication goroutine")
	}()
}
