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
	"fmt"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
)

// RunScanQueue ...
func RunScanQueue(configPath string, stop <-chan struct{}) {
	config, err := GetConfig(configPath)
	log.Warnf("unserialized config: %+v", config)
	if err != nil {
		panic(fmt.Errorf("Failed to load configuration: %v", err))
	}

	level, err := config.GetLogLevel()
	if err != nil {
		panic(err)
	}
	log.SetLevel(level)

	prometheus.Unregister(prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{}))
	prometheus.Unregister(prometheus.NewGoCollector())

	queue := NewModel()
	for key, data := range config.Jobs {
		log.Infof("adding key %s, job %+v", key, data)
		queue.AddJob(Job{Key: key, Data: data})
	}
	SetupHTTPServer(queue)

	http.Handle("/metrics", promhttp.Handler())

	addr := fmt.Sprintf(":%d", config.Port)
	log.Infof("successfully instantiated queue %+v, serving on %s", queue, addr)
	go func() {
		http.ListenAndServe(addr, nil)
	}()

	<-stop
}
