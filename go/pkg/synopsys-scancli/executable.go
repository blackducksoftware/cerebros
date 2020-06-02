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
	"fmt"
	"github.com/blackducksoftware/cerebros/go/pkg/scanqueue"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	"net/http"
)

func doOrDie(err error) {
	if err != nil {
		log.Fatalf("unable to continue: %+v", err)
	}
}

func RunContainerizedCLI(configPath string) {
	config, err := GetConfig(configPath)
	doOrDie(err)

	configSerialized, err := json.MarshalIndent(config, "", "  ")
	doOrDie(err)
	log.Infof("got config: \n%+v\n%s\n", config, string(configSerialized))

	level, err := config.getLogLevel()
	doOrDie(err)
	log.SetLevel(level)
	log.Warnf("set log level to %s", level)

	prometheus.Unregister(prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{}))
	prometheus.Unregister(prometheus.NewGoCollector())

	http.Handle("/metrics", promhttp.Handler())

	addr := fmt.Sprintf(":%d", config.Port)
	log.Infof("serving on %s", addr)
	go func() {
		http.ListenAndServe(addr, nil)
	}()

	scanner, err := NewScannerFromConfig(config.Blackduck, config.Polaris, config.ImageFacade)
	doOrDie(err)

	stop := make(chan struct{})
	cc := NewContainerizedCLI(scanner, scanqueue.NewClient(config.ScanQueue.Host, config.ScanQueue.Port), stop)
	log.Infof("instantiated containerized cli: %+v", cc)

	<-stop
}
