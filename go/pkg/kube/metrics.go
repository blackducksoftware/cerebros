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
package kube

import (
	"github.com/prometheus/client_golang/prometheus"
)

var containerMemoryUsageGauge *prometheus.GaugeVec
var containerMemoryPercentageGauge *prometheus.GaugeVec
var containerCpuUsageGauge *prometheus.GaugeVec
var containerCpuPercentageGauge *prometheus.GaugeVec

func recordContainerMemory(ns string, pod string, cont string, size int, limitPercentage int, requestPercentage int) {
	containerMemoryUsageGauge.With(prometheus.Labels{"namespace": ns, "pod": pod, "container": cont}).Set(float64(size))
	containerMemoryPercentageGauge.With(prometheus.Labels{"namespace": ns, "pod": pod, "container": cont, "type": "limit"}).Set(float64(limitPercentage))
	containerMemoryPercentageGauge.With(prometheus.Labels{"namespace": ns, "pod": pod, "container": cont, "type": "request"}).Set(float64(requestPercentage))
}

func recordContainerCPU(ns string, pod string, cont string, usage int, limitPercentage int, requestPercentage int) {
	containerCpuUsageGauge.With(prometheus.Labels{"namespace": ns, "pod": pod, "container": cont}).Set(float64(usage))
	containerCpuPercentageGauge.With(prometheus.Labels{"namespace": ns, "pod": pod, "container": cont, "type": "limit"}).Set(float64(limitPercentage))
	containerCpuPercentageGauge.With(prometheus.Labels{"namespace": ns, "pod": pod, "container": cont, "type": "request"}).Set(float64(requestPercentage))
}

var podMemoryUsageGauge *prometheus.GaugeVec
var podMemoryPercentageGauge *prometheus.GaugeVec
var podCpuUsageGauge *prometheus.GaugeVec
var podCpuPercentageGauge *prometheus.GaugeVec

func recordPodMemory(ns string, pod string, size int, limitPercentage int, requestPercentage int) {
	podMemoryUsageGauge.With(prometheus.Labels{"namespace": ns, "pod": pod}).Set(float64(size))
	podMemoryPercentageGauge.With(prometheus.Labels{"namespace": ns, "pod": pod, "type": "limit"}).Set(float64(limitPercentage))
	podMemoryPercentageGauge.With(prometheus.Labels{"namespace": ns, "pod": pod, "type": "request"}).Set(float64(requestPercentage))
}

func recordPodCPU(ns string, pod string, usage int, limitPercentage int, requestPercentage int) {
	podCpuUsageGauge.With(prometheus.Labels{"namespace": ns, "pod": pod}).Set(float64(usage))
	podCpuPercentageGauge.With(prometheus.Labels{"namespace": ns, "pod": pod, "type": "limit"}).Set(float64(limitPercentage))
	podCpuPercentageGauge.With(prometheus.Labels{"namespace": ns, "pod": pod, "type": "request"}).Set(float64(requestPercentage))
}

var nodeMemoryUsageGauge *prometheus.GaugeVec
var nodeMemoryPercentageGauge *prometheus.GaugeVec
var nodeCpuUsageGauge *prometheus.GaugeVec
var nodeCpuPercentageGauge *prometheus.GaugeVec

func recordNodeMemory(node string, size int, percentage int) {
	nodeMemoryUsageGauge.With(prometheus.Labels{"node": node}).Set(float64(size))
	nodeMemoryPercentageGauge.With(prometheus.Labels{"node": node}).Set(float64(percentage))
}

func recordNodeCPU(node string, usage int, percentage int) {
	nodeCpuUsageGauge.With(prometheus.Labels{"node": node}).Set(float64(usage))
	nodeCpuPercentageGauge.With(prometheus.Labels{"node": node}).Set(float64(percentage))
}

func init() {
	containerMemoryUsageGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "cerebros",
		Subsystem: "kube_metrics",
		Name:      "cont_memory_usage_gauge",
		Help:      "a gauge of container memory usage",
	}, []string{"namespace", "pod", "container"})
	prometheus.MustRegister(containerMemoryUsageGauge)

	containerMemoryPercentageGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "cerebros",
		Subsystem: "kube_metrics",
		Name:      "cont_percentage_memory_usage_gauge",
		Help:      "a gauge of container percent memory usage",
	}, []string{"namespace", "pod", "container", "type"})
	prometheus.MustRegister(containerMemoryPercentageGauge)

	containerCpuUsageGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "cerebros",
		Subsystem: "kube_metrics",
		Name:      "cont_cpu_usage_gauge",
		Help:      "a gauge of container cpu usage",
	}, []string{"namespace", "pod", "container"})
	prometheus.MustRegister(containerCpuUsageGauge)

	containerCpuPercentageGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "cerebros",
		Subsystem: "kube_metrics",
		Name:      "cont_percentage_cpu_usage_gauge",
		Help:      "a gauge of container percent cpu usage",
	}, []string{"namespace", "pod", "container", "type"})
	prometheus.MustRegister(containerCpuPercentageGauge)

	podMemoryUsageGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "cerebros",
		Subsystem: "kube_metrics",
		Name:      "pod_memory_usage_gauge",
		Help:      "a gauge of pod memory usage",
	}, []string{"namespace", "pod"})
	prometheus.MustRegister(podMemoryUsageGauge)

	podMemoryPercentageGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "cerebros",
		Subsystem: "kube_metrics",
		Name:      "pod_percentage_memory_usage_gauge",
		Help:      "a gauge of pod percent memory usage",
	}, []string{"namespace", "pod", "type"})
	prometheus.MustRegister(podMemoryPercentageGauge)

	podCpuUsageGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "cerebros",
		Subsystem: "kube_metrics",
		Name:      "pod_cpu_usage_gauge",
		Help:      "a gauge of pod cpu usage",
	}, []string{"namespace", "pod"})
	prometheus.MustRegister(podCpuUsageGauge)

	podCpuPercentageGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "cerebros",
		Subsystem: "kube_metrics",
		Name:      "pod_percentage_cpu_usage_gauge",
		Help:      "a gauge of pod percent cpu usage",
	}, []string{"namespace", "pod", "type"})
	prometheus.MustRegister(podCpuPercentageGauge)

	nodeMemoryUsageGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "cerebros",
		Subsystem: "kube_metrics",
		Name:      "node_memory_usage_gauge",
		Help:      "a gauge of node memory usage",
	}, []string{"node"})
	prometheus.MustRegister(nodeMemoryUsageGauge)

	nodeMemoryPercentageGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "cerebros",
		Subsystem: "kube_metrics",
		Name:      "node_percentage_memory_usage_gauge",
		Help:      "a gauge of node percent memory usage",
	}, []string{"node"})
	prometheus.MustRegister(nodeMemoryPercentageGauge)

	nodeCpuUsageGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "cerebros",
		Subsystem: "kube_metrics",
		Name:      "node_cpu_usage_gauge",
		Help:      "a gauge of node cpu usage",
	}, []string{"node"})
	prometheus.MustRegister(nodeCpuUsageGauge)

	nodeCpuPercentageGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "cerebros",
		Subsystem: "kube_metrics",
		Name:      "node_percentage_cpu_usage_gauge",
		Help:      "a gauge of node percent cpu usage",
	}, []string{"node"})
	prometheus.MustRegister(nodeCpuPercentageGauge)
}
