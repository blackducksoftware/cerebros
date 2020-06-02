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
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"time"
)

var eventCounter *prometheus.CounterVec
var eventTimeHistogram *prometheus.HistogramVec

func recordEvent(eventType string, err error) {
	eventCounter.With(prometheus.Labels{"type": eventType, "iserror": fmt.Sprintf("%t", err != nil)}).Inc()
}

func recordEventTime(event string, duration time.Duration) {
	milliseconds := float64(duration / time.Millisecond)
	eventTimeHistogram.With(prometheus.Labels{"event": event}).Observe(milliseconds)
}

func init() {
	eventCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "cerebros",
		Subsystem: "synopsys_scancli",
		Name:      "event_counter",
		Help:      "a count of events",
	}, []string{"type", "iserror"})
	prometheus.MustRegister(eventCounter)

	eventTimeHistogram = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "cerebros",
		Subsystem: "synopsys_scancli",
		Name:      "response_time_histogram",
		Help:      "a histogram of polaris event times in milliseconds",
		Buckets:   prometheus.ExponentialBuckets(1, 2, 25),
	}, []string{"event"})
	prometheus.MustRegister(eventTimeHistogram)
}
