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
var eventGauge *prometheus.GaugeVec
var namedEventGauge *prometheus.GaugeVec
var eventHistogram *prometheus.HistogramVec

func recordEvent(eventType string, err error) {
	eventCounter.With(prometheus.Labels{"type": eventType, "iserror": fmt.Sprintf("%t", err != nil)}).Inc()
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

// TODO
//func recordRate(rate int) {
//	eventGauge.With(prometheus.Labels{"type": "rateLimit"}).Set(float64(rate))
//}

func recordIssuePageJobProjectIndex(i int) {
	eventGauge.With(prometheus.Labels{"type": "issuePageJobProjectIndex"}).Set(float64(i))
}

func recordJobsInProgress(name string, count int) {
	namedEventGauge.With(prometheus.Labels{"type": "jobsInProgress", "name": name}).Set(float64(count))
}

func recordErrorCount(name string, count int) {
	namedEventGauge.With(prometheus.Labels{"type": "errorCount", "name": name}).Set(float64(count))
}

func recordSuccessCount(name string, count int) {
	namedEventGauge.With(prometheus.Labels{"type": "successCount", "name": name}).Set(float64(count))
}

func recordRoleAssignmentsSingleProjectIndex(index int) {
	eventGauge.With(prometheus.Labels{"type": "roleAssignmentsProjectIndex"}).Set(float64(index))
}

func recordProjectPage(page int) {
	eventGauge.With(prometheus.Labels{"type": "projectPage"}).Set(float64(page))
}

func recordErrorFraction(name string, fraction float64) {
	namedEventGauge.With(prometheus.Labels{"type": "errorFraction", "name": name}).Set(fraction)
}

//func recordProjectIssuesCount(count int) {
//	projectIssueCounter.With(prometheus.Labels{"count": fmt.Sprintf("%d", count)}).Inc()
//}

func recordProjectRollupCountsIndex(i int) {
	eventGauge.With(prometheus.Labels{"type": "projectRollupCountsIndex"}).Set(float64(i))
}

func init() {
	eventCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "cerebros",
		Subsystem: "polaris_api_load_issue_server",
		Name:      "event_counter",
		Help:      "a count of events",
	}, []string{"type", "iserror"})
	prometheus.MustRegister(eventCounter)

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
