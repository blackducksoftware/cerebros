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
	"github.com/prometheus/client_golang/prometheus"
	"time"
)

var eventCounter *prometheus.CounterVec
var namedEventCounter *prometheus.CounterVec
var eventGauge *prometheus.GaugeVec
var namedEventGauge *prometheus.GaugeVec
var eventHistogram *prometheus.HistogramVec

func recordEvent(eventType string, err error) {
	eventCounter.With(prometheus.Labels{"type": eventType, "iserror": fmt.Sprintf("%t", err != nil)}).Inc()
}

func recordNamedEventBy(eventType string, name string, count int) {
	namedEventCounter.With(prometheus.Labels{"type": eventType, "name": name}).Add(float64(count))
}

func recordEventGauge(eventType string, value int) {
	eventGauge.With(prometheus.Labels{"type": eventType}).Set(float64(value))
}

func recordNamedEventGauge(eventType string, name string, value float64) {
	namedEventGauge.With(prometheus.Labels{"type": eventType, "name": name}).Set(value)
}

func recordDuration(typeName string, duration time.Duration) {
	milliseconds := float64(duration / time.Millisecond)
	eventHistogram.With(prometheus.Labels{"type": typeName}).Observe(milliseconds)
}

func init() {
	eventCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "cerebros",
		Subsystem: "polaris_api_load_issue_server",
		Name:      "event_counter",
		Help:      "a count of events",
	}, []string{"type", "iserror"})
	prometheus.MustRegister(eventCounter)

	namedEventCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "cerebros",
		Subsystem: "polaris_api_load_issue_server",
		Name:      "named_event_counter",
		Help:      "a count of named events",
	}, []string{"type", "name"})
	prometheus.MustRegister(namedEventCounter)

	eventGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "cerebros",
		Subsystem: "polaris_api_load_issue_server",
		Name:      "event_gauge",
		Help:      "gauges for events happening during load generation",
	}, []string{"type"})
	prometheus.MustRegister(eventGauge)

	namedEventGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "cerebros",
		Subsystem: "polaris_api_load_issue_server",
		Name:      "named_event_gauge",
		Help:      "gauges for named events happening during load generation",
	}, []string{"type", "name"})
	prometheus.MustRegister(namedEventGauge)

	eventHistogram = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "cerebros",
		Subsystem: "polaris_api_load_issue_server",
		Name:      "event_histogram",
		Help:      "durations for events happening in load generation",
		Buckets:   prometheus.ExponentialBuckets(1, 2, 20),
	}, []string{"type"})
	prometheus.MustRegister(eventHistogram)
}
