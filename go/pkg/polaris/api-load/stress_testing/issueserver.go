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
	"github.com/blackducksoftware/cerebros/go/pkg/polaris/api"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"sync"
)

type RollupCountsSource struct {
	Projects *ProjectFetcher
	Index    int
	Limit    int
	client   *api.Client
	mux      *sync.Mutex
}

func NewRollupCountsSource(client *api.Client, projects *ProjectFetcher, start int, limit int) *RollupCountsSource {
	return &RollupCountsSource{
		client:   client,
		Projects: projects,
		Index:    start,
		Limit:    limit,
		mux:      &sync.Mutex{},
	}
}

func (rcs *RollupCountsSource) getNextIndex() (int, bool) {
	rcs.mux.Lock()
	defer rcs.mux.Unlock()

	projectsCount := rcs.Projects.MainBranchProjectsLength()
	if projectsCount == 0 {
		return -1, false
	}

	if rcs.Index >= projectsCount {
		recordEvent("resetting rollupCounts index", nil)
		rcs.Index = 0
	}
	log.Debugf("rollupCounts index %d", rcs.Index)
	i := rcs.Index
	recordEventGauge("projectRollupCountsIndex", i)
	rcs.Index++
	return i, true
}

func (rcs *RollupCountsSource) RunJob() (string, error) {
	index, ok := rcs.getNextIndex()
	if !ok {
		return "getV0RollUpCounts -- no project available", errors.New("no projects")
	}
	project := rcs.Projects.GetMainBranchProject(index)
	_, err := rcs.client.GetV0RollUpCounts(project.ProjectId, project.MainBranchId, rcs.Limit)
	return "getV0RollUpCounts", err
}

type issueJob struct {
	ProjectId string
	BranchId  string
	IssueId   string
}

type issuePageJob struct {
	ProjectId string
	BranchId  string
}

type IssuesSource struct {
	client   *api.Client
	Projects *ProjectFetcher
	Offset   int
	Index    int
	issues   chan *issueJob
	mux      *sync.Mutex
}

func NewIssuesSource(client *api.Client, projects *ProjectFetcher, start int) *IssuesSource {
	is := &IssuesSource{
		client:   client,
		Projects: projects,
		Offset:   start,
		Index:    0,
		issues:   make(chan *issueJob, 1000), // TODO what should the size be?  don't want writes to ever block
		mux:      &sync.Mutex{},
	}
	return is
}

func (is *IssuesSource) getIssueJob() *issueJob {
	select {
	case job := <-is.issues:
		return job
	default:
		return nil
	}
}

func (is *IssuesSource) getIssuePageJob() *issuePageJob {
	is.mux.Lock()
	defer is.mux.Unlock()

	projectCount := is.Projects.MainBranchProjectsLength()
	if projectCount == 0 {
		return nil
	}

	if is.Index >= projectCount {
		is.Index = 0
	}
	recordEventGauge("issuePageJobProjectIndex", is.Index)
	project := is.Projects.GetMainBranchProject(is.Index)
	is.Index++
	return &issuePageJob{
		ProjectId: project.ProjectId,
		BranchId:  project.MainBranchId,
	}
}

func (is *IssuesSource) RunJob() (string, error) {
	if job := is.getIssueJob(); job != nil {
		_, err := is.client.GetV1Issue(job.ProjectId, job.BranchId, job.IssueId)
		return "getV1Issue (single)", err
	} else if pageJob := is.getIssuePageJob(); pageJob != nil {
		offset := 0
		pageSize := 60
		v1Issues, err := is.client.GetV1Issues(pageJob.ProjectId, pageJob.BranchId, "", offset, pageSize)
		if err == nil {
			go func() {
				for _, issueResponse := range v1Issues.Data {
					issue := &issueJob{
						ProjectId: pageJob.ProjectId,
						BranchId:  pageJob.BranchId,
						IssueId:   issueResponse.Id,
					}
					is.issues <- issue
				}
			}()
		}
		return "getV1Issues", err
	} else {
		return "getV1Issues -- no project available", errors.New("no projects")
	}
}

type IssueServerLoadGenerator struct {
	polarisClient           *api.Client
	mux                     *sync.Mutex
	Config                  *IssueServerConfig
	issuesLoadManager       *LoadManager
	rollupCountsLoadManager *LoadManager
	stopChan                chan struct{}
}

func NewIssueServerLoadGenerator(polarisClient *api.Client, projects *ProjectFetcher, config *IssueServerConfig) *IssueServerLoadGenerator {
	c := &IssueServerLoadGenerator{
		polarisClient: polarisClient,
		mux:           &sync.Mutex{},
		Config:        config,
		stopChan:      make(chan struct{}),
		issuesLoadManager: NewLoadManager(
			"issues",
			NewIssuesSource(polarisClient, projects, 0),
			config.Issues.WorkersCount,
			config.Issues.Rate.MustRateLimiter("issues")),
		rollupCountsLoadManager: NewLoadManager(
			"rollupcounts",
			NewRollupCountsSource(polarisClient, projects, 0, config.RollupCounts.PageSize),
			config.RollupCounts.LoadConfig.WorkersCount,
			config.RollupCounts.LoadConfig.Rate.MustRateLimiter("rollupcounts")),
	}
	Reauthenticator(polarisClient, c.stopChan)
	return c
}

func (c *IssueServerLoadGenerator) Stop() {
	close(c.stopChan)
	if c.issuesLoadManager != nil {
		c.issuesLoadManager.stop()
	}
	if c.rollupCountsLoadManager != nil {
		c.rollupCountsLoadManager.stop()
	}
}
