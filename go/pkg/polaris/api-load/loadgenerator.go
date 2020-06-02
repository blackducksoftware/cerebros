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
package api_load

import (
	"fmt"
	"github.com/blackducksoftware/cerebros/go/pkg/polaris/api"
	log "github.com/sirupsen/logrus"
)

type LoadGenerator struct {
	Client         *api.Client
	WorkerRequests map[string]int
	Workers        map[string]*LoadGenerationWorker
	Stop           <-chan struct{}
}

func NewLoadGenerator(client *api.Client, workerRequests map[string]int, stop <-chan struct{}) *LoadGenerator {
	loadGen := &LoadGenerator{Client: client, WorkerRequests: workerRequests, Workers: map[string]*LoadGenerationWorker{}, Stop: stop}
	return loadGen
}

// TODO add more types of load workers:
// - periodically measure what's in polaris -- number of users, groups, jobs, roles, roleassignments
//   and dump these out to prometheus as gauges
// - measure response times on various endpoints
//   do this say, once a minute -- don't overwhelm the api, just measure the response times
// - measure response times for common api calls coming in from popular ui pages
//   i.e look at the response times for api calls from: projects page, users page, etc.
// TODO !!!!

func (loadGen *LoadGenerator) StartGeneratingLoad() {
	log.Infof("starting loadgen workers")
	for workerType, count := range loadGen.WorkerRequests {
		for i := 0; i < count; i++ {
			idString := fmt.Sprintf("%d", i)
			var worker *LoadGenerationWorker
			switch workerType {
			case "Entitlements", "entitlements":
				//f := func() error {
				//	_, err := loadGen.Client.GetEntitlements(projectId)
				//	return err
				//}
				//worker = NewLoadGenerationWorker(idString, "entitlements", f, loadGen.Stop)
				panic("entitlements loadgen workers not supported at the moment")
				break
			case "Groups", "groups":
				f := func() error {
					groups, err := loadGen.Client.GetGroups()
					log.Debugf("got %d groups, total of %d", len(groups.Data), groups.Meta.Total)
					return err
				}
				worker = NewLoadGenerationWorker(idString, "groups", f, loadGen.Stop)
				break
			case "Jobs", "jobs":
				f := func() error {
					jobs, err := loadGen.Client.GetJobs(10)
					log.Debugf("found %d jobs, total of %d", len(jobs.Data), jobs.Meta.Total)
					return err
				}
				worker = NewLoadGenerationWorker(idString, "jobs", f, loadGen.Stop)
				break
			case "Projects", "projects":
				f := func() error {
					projects, err := loadGen.Client.GetProjects(10)
					log.Debugf("got %d projects, total of %d", len(projects.Data), projects.Meta.Total)
					return err
				}
				worker = NewLoadGenerationWorker(idString, "projects", f, loadGen.Stop)
				break
			case "RoleAssignments", "roleassignments":
				f := func() error {
					ras, err := loadGen.Client.GetRoleAssignments(0, 20)
					log.Debugf("found %d roleassignments, total of %d", len(ras.Data), ras.Meta.Total)
					return err
				}
				worker = NewLoadGenerationWorker(idString, "roleassignments", f, loadGen.Stop)
				break
			case "Taxonomies", "taxonomies":
				f := func() error {
					taxonomies, err := loadGen.Client.GetTaxonomies(nil, 10)
					log.Debugf("got %d projects, total of %d", len(taxonomies.Data), taxonomies.Meta.Total)
					return err
				}
				worker = NewLoadGenerationWorker(idString, "taxonomies", f, loadGen.Stop)
				break
			case "Login", "login":
				// need a separate client because every time we authenticate, the client gets a new token
				// which could stomp on what the other worker types are doing.
				// So: one client per login worker
				client := api.NewClient(loadGen.Client.URL, loadGen.Client.Email, loadGen.Client.Password)
				f := func() error {
					err := client.Authenticate()
					log.Debugf("login worker authenticate result: %+v", err)
					return err
				}
				worker = NewLoadGenerationWorker(idString, "login", f, loadGen.Stop)
				break
			default:
				log.Errorf("invalid worker type %s, skipping", workerType)
				continue
			}
			loadGen.Workers[workerType] = worker
			worker.Start()
		}
	}
}
