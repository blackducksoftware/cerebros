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
	"fmt"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"net/http"
	"strings"
	"time"
)

type Config struct {
	KubeConfigPath string
	Namespace      string
	LogLevel       string
	Port           int
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

func CollectMetrics(configPath string) error {
	config, err := GetConfig(configPath)
	if err != nil {
		return err
	}
	logLevel, err := config.GetLogLevel()
	if err != nil {
		return err
	}
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

	// get set up with kube
	kc, err := NewKubeClient(config.KubeConfigPath)
	if err != nil {
		return err
	}

	stop := make(chan struct{})

	// collect a bunch of data
	go func() {
		for {
			select {
			case <-stop:
				break
			default:
			}
			log.Infof("collecting node metrics")
			nodes := kc.Nodes()
			for name, node := range nodes {
				//str, err := json.Marshal(node)
				//if err != nil {
				//	log.Errorf("unable to serialize node metrics: %s", err)
				//} else {
				//	log.Infof("node metrics for %s: %s", name, str)
				//}
				recordNodeCPU(name, node.CPUMilli.Current, node.CPUMilli.Percentage())
				recordNodeMemory(name, node.MemoryBytes.Current/MB, node.MemoryBytes.Percentage())
			}
			//bytes, err := json.MarshalIndent(nodes, "", "  ")
			//if err != nil {
			//	return err
			//}
			//fmt.Printf("nodes:\n\n%s\n\n", string(bytes))
			//fmt.Printf("\n%s\n", PrettyPrint(nodes))
			time.Sleep(10 * time.Second)
			//break
		}
	}()

	go func() {
		for {
			select {
			case <-stop:
				break
			default:
			}
			log.Infof("collecting container metrics")
			namespaces := kc.Containers(config.Namespace)
			//containerBytes, err := json.MarshalIndent(namespace, "", "  ")
			//if err != nil {
			//	log.Errorf("unable to marshal containers: %s", err)
			//} else {
			//	log.Infof("containers: %s", containerBytes)
			//}
			for _, namespace := range namespaces {
				for _, pod := range namespace.Pods {
					recordPodMemory(namespace.Name, pod.Name, pod.MemoryBytes.Usage/MB, pod.MemoryBytes.LimitPercentage(), pod.MemoryBytes.RequestPercentage())
					recordPodCPU(namespace.Name, pod.Name, pod.CPUMilli.Usage, pod.CPUMilli.LimitPercentage(), pod.CPUMilli.RequestPercentage())
					for _, container := range pod.Containers {
						cpu := container.CPUMilli
						recordContainerCPU(namespace.Name, pod.Name, container.Name, cpu.Usage, cpu.LimitPercentage(), cpu.RequestPercentage())
						mem := container.MemoryBytes
						recordContainerMemory(namespace.Name, pod.Name, container.Name, mem.Usage/MB, mem.LimitPercentage(), mem.RequestPercentage())
					}
				}
			}
			//fmt.Printf("containers:\n\n%s\n\n", string(containerBytes))
			//fmt.Printf("\n%s\n", PrettyPrint(containers))
			time.Sleep(10 * time.Second)
		}
	}()

	<-stop

	return nil
}

func PrettyPrint(nodes map[string]*Node) string {
	header := strings.Join([]string{"-", "CPU used", "CPU total", "Memory used", "Memory allocatable", "Memory %"}, "\t")
	lines := []string{header}
	for name, node := range nodes {
		currentMBs := node.MemoryBytes.Current / MB
		allocatableMBs := node.MemoryBytes.Allocatable / MB
		line := []string{
			name[len(name)-7 : len(name)],
			fmt.Sprintf("%d", node.CPUMilli.Current),
			fmt.Sprintf("%d", node.CPUMilli.Total),
			fmt.Sprintf("%d", currentMBs),
			fmt.Sprintf("%d", allocatableMBs),
			fmt.Sprintf("%d", node.MemoryBytes.Percentage()),
		}
		lines = append(lines, strings.Join(line, "\t"))
	}
	return strings.Join(lines, "\n")
}
