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
package api

import (
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"time"
)

var eventCounter *prometheus.CounterVec
var statusCodeCounter *prometheus.CounterVec
var responseTimeHistogram *prometheus.HistogramVec

func recordEvent(event string, err error) {
	eventCounter.With(prometheus.Labels{"event": event, "iserror": fmt.Sprintf("%t", err != nil)}).Inc()
}

func recordResponseStatusCode(verb string, apiPath string, statusCode int) {
	statusCodeCounter.With(prometheus.Labels{"verb": verb, "apipath": apiPath, "code": fmt.Sprintf("%d", statusCode)}).Inc()
}

func recordResponseTime(verb string, apiPath string, duration time.Duration, statusCode int) {
	milliseconds := float64(duration / time.Millisecond)
	responseTimeHistogram.With(prometheus.Labels{
		"verb":    verb,
		"apipath": apiPath,
		"code":    fmt.Sprintf("%d", statusCode),
	}).Observe(milliseconds)
}

func init() {
	eventCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "cerebros",
		Subsystem: "polaris_api",
		Name:      "event_counter",
		Help:      "a count of events",
	}, []string{"event", "iserror"})
	prometheus.MustRegister(eventCounter)

	statusCodeCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "cerebros",
		Subsystem: "polaris_api",
		Name:      "status_code_counter",
		Help:      "a counter of status codes from http responses",
	}, []string{"verb", "apipath", "code"})
	prometheus.MustRegister(statusCodeCounter)

	responseTimeHistogram = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "cerebros",
		Subsystem: "polaris_api",
		Name:      "response_time_histogram",
		Help:      "a histogram of polaris api response times in milliseconds",
		Buckets:   prometheus.ExponentialBuckets(1, 2, 20),
	}, []string{"verb", "apipath", "code"})
	prometheus.MustRegister(responseTimeHistogram)
}
