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
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/metrics/pkg/client/clientset/versioned"
	"math"
)

const (
	MB = 1048576
)

type Node struct {
	Name         string
	MemoryBytes  *CTA
	CPUMilli     *CTA
	StorageBytes *CTA
}

type CTA struct {
	Current     int
	Total       int
	Allocatable int
}

func (cta *CTA) Available() int {
	return cta.Allocatable - cta.Current
}

func (cta *CTA) Percentage() int {
	return Percentage(cta.Current, cta.Allocatable)
}

func Percentage(top int, bottom int) int {
	if bottom == 0 {
		return 0
	}
	return int(math.Floor((float64(top) / float64(bottom)) * 100))
}

type Namespace struct {
	Name string
	Pods map[string]*Pod
}

type Pod struct {
	Name         string
	Containers   map[string]*Container
	MemoryBytes  *RLU
	CPUMilli     *RLU
	StorageBytes *RLU
}

type Container struct {
	Name         string
	MemoryBytes  *RLU
	CPUMilli     *RLU
	StorageBytes *RLU
}

type RLU struct {
	Request int
	Limit   int
	Usage   int
}

func (rlu *RLU) LimitPercentage() int {
	return Percentage(rlu.Usage, rlu.Limit)
}

func (rlu *RLU) RequestPercentage() int {
	return Percentage(rlu.Usage, rlu.Request)
}

type KubeClient struct {
	client  *kubernetes.Clientset
	vClient *versioned.Clientset
}

func NewKubeClient(kubeConfigPath string) (*KubeClient, error) {
	kubeConfig, err := clientcmd.BuildConfigFromFlags("", kubeConfigPath)
	//		kubeConfig, err := rest.InClusterConfig()
	if err != nil {
		return nil, errors.Wrapf(err, "unable to build config from flags")
	}
	client, err := kubernetes.NewForConfig(kubeConfig)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to instantiate client")
	}
	log.Infof("client set: %+v", client)
	vClient, err := versioned.NewForConfig(kubeConfig)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to instantiate versioned client")
	}
	return &KubeClient{
		client:  client,
		vClient: vClient,
	}, nil
}

func (kc *KubeClient) Nodes() map[string]*Node {
	nodeMetrics, err := kc.vClient.MetricsV1beta1().NodeMetricses().List(metav1.ListOptions{})
	if err != nil {
		panic(err)
	}
	nodeList, err := kc.client.CoreV1().Nodes().List(metav1.ListOptions{})
	if err != nil {
		panic(err)
	}
	nodes := map[string]*Node{}
	for _, node := range nodeList.Items {
		//log.Infof("node: %+v", node)
		nodes[node.Name] = &Node{
			Name: node.Name,
			CPUMilli: &CTA{
				Total:       int(node.Status.Capacity.Cpu().MilliValue()),
				Allocatable: int(node.Status.Allocatable.Cpu().MilliValue()),
			},
			MemoryBytes: &CTA{
				Total:       int(node.Status.Capacity.Memory().Value()),
				Allocatable: int(node.Status.Allocatable.Memory().Value()),
			},
			StorageBytes: &CTA{
				Total:       int(node.Status.Capacity.StorageEphemeral().Value()),
				Allocatable: int(node.Status.Allocatable.StorageEphemeral().Value()),
			},
		}
	}
	for _, nm := range nodeMetrics.Items {
		nodes[nm.Name].CPUMilli.Current = int(nm.Usage.Cpu().MilliValue())
		nodes[nm.Name].MemoryBytes.Current = int(nm.Usage.Memory().Value())
		nodes[nm.Name].StorageBytes.Current = int(nm.Usage.StorageEphemeral().Value())
	}
	return nodes
}

func (kc *KubeClient) Containers(namespace string) map[string]*Namespace {
	podMetrics, err := kc.vClient.MetricsV1beta1().PodMetricses(namespace).List(metav1.ListOptions{})
	if err != nil {
		panic(err)
	}
	pods, err := kc.client.CoreV1().Pods(namespace).List(metav1.ListOptions{})
	if err != nil {
		panic(err)
	}
	namespaces := make(map[string]*Namespace)
	for _, pod := range pods.Items {
		podNs := pod.Namespace
		if _, ok := namespaces[podNs]; !ok {
			namespaces[podNs] = &Namespace{
				Name: podNs,
				Pods: map[string]*Pod{},
			}
		}
		ns := namespaces[podNs]
		ns.Pods[pod.Name] = &Pod{
			Name:         pod.Name,
			Containers:   map[string]*Container{},
			MemoryBytes:  &RLU{},
			CPUMilli:     &RLU{},
			StorageBytes: &RLU{},
		}
		memoryReq := 0
		cpuReq := 0
		ephReq := 0
		memoryLimit := 0
		cpuLimit := 0
		ephLimit := 0
		for _, cont := range pod.Spec.Containers {
			limits := cont.Resources.Limits
			reqs := cont.Resources.Requests
			memoryLimit += int(limits.Memory().Value())
			cpuLimit += int(limits.Cpu().MilliValue())
			ephLimit += int(limits.StorageEphemeral().Value())
			memoryReq += int(reqs.Memory().Value())
			cpuReq += int(reqs.Cpu().MilliValue())
			ephReq += int(reqs.StorageEphemeral().Value())
			ns.Pods[pod.Name].Containers[cont.Name] = &Container{
				Name: cont.Name,
				MemoryBytes: &RLU{
					Request: int(reqs.Memory().Value()),
					Limit:   int(limits.Memory().Value()),
				},
				CPUMilli: &RLU{
					Request: int(reqs.Cpu().MilliValue()),
					Limit:   int(limits.Cpu().MilliValue()),
				},
				StorageBytes: &RLU{
					Request: int(reqs.StorageEphemeral().Value()),
					Limit:   int(limits.StorageEphemeral().Value()),
				},
			}
		}
		ns.Pods[pod.Name].MemoryBytes.Limit = memoryLimit
		ns.Pods[pod.Name].MemoryBytes.Request = memoryReq
		ns.Pods[pod.Name].CPUMilli.Limit = cpuLimit
		ns.Pods[pod.Name].CPUMilli.Request = cpuReq
		ns.Pods[pod.Name].StorageBytes.Limit = ephLimit
		ns.Pods[pod.Name].StorageBytes.Request = ephReq
	}

	for _, pm := range podMetrics.Items {
		memTotal := 0
		cpuTotal := 0
		ephTotal := 0
		ns := namespaces[pm.Namespace]
		for _, cont := range pm.Containers {
			mem := int(cont.Usage.Memory().Value())
			cpu := int(cont.Usage.Cpu().MilliValue())
			eph := int(cont.Usage.StorageEphemeral().Value())
			memTotal += mem
			cpuTotal += cpu
			ephTotal += eph
			ns.Pods[pm.Name].Containers[cont.Name].MemoryBytes.Usage = mem
			ns.Pods[pm.Name].Containers[cont.Name].CPUMilli.Usage = cpu
			ns.Pods[pm.Name].Containers[cont.Name].StorageBytes.Usage = eph
		}
		ns.Pods[pm.Name].MemoryBytes.Usage = memTotal
		ns.Pods[pm.Name].CPUMilli.Usage = cpuTotal
		ns.Pods[pm.Name].StorageBytes.Usage = ephTotal
	}
	return namespaces
}
