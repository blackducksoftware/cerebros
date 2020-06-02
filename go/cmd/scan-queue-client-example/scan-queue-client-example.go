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
package main

import (
	"github.com/blackducksoftware/cerebros/go/pkg/scanqueue"
	log "github.com/sirupsen/logrus"
)

type Job struct {
	Name  string
	Value string
}

func main() {
	log.SetLevel(log.DebugLevel)

	client := scanqueue.NewClient("localhost", 4100)

	addJob(client, "job3", Job{"abc", "def"})
	addJob(client, "job4", Job{"qrs", "tuv"})

	getModel(client)

	nextJob := Job{}
	err := client.GetNextJob(&nextJob)
	if err != nil {
		panic(err)
	}
	log.Infof("next job: %+v", nextJob)

	getModel(client)
}

func addJob(client *scanqueue.Client, key string, job Job) {
	err := client.AddJob(key, job)
	if err != nil {
		panic(err)
	}
	log.Infof("add job succeeded")
}

func getModel(client *scanqueue.Client) {
	modelString, err := client.GetModel()
	if err != nil {
		panic(err)
	}
	log.Infof("model: %s", modelString)
}
