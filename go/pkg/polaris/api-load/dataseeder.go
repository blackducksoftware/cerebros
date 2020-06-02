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
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"strings"
	"time"
)

type RoleAssignment struct {
	Organization   string
	User           string
	Role           string
	Object         string
	ToOrganization string
	ToProject      string
}

type DataSeeder struct {
	Client          *api.Client
	UsersToCreate   int
	Concurrency     int
	Projects        map[string]string
	Users           map[string]string
	UserNameToId    map[string]string
	Roles           map[string]string
	Organizations   map[string]string
	RoleAssignments map[string]*RoleAssignment
}

func NewDataSeeder(client *api.Client, usersToCreate int, concurrency int) (*DataSeeder, error) {
	ds := &DataSeeder{Client: client, UsersToCreate: usersToCreate, Concurrency: concurrency}
	err := ds.setupData()
	return ds, errors.WithMessagef(err, "unable to set up DataSeeder")
}

func (ds *DataSeeder) setupData() error {
	apiProjects, err := ds.Client.GetProjects(25)
	if err != nil {
		return err
	}
	log.Infof("found %d projects", len(apiProjects.Data))
	ds.Projects = map[string]string{}
	for _, project := range apiProjects.Data {
		id, name := project.Id, project.Attributes.Name
		log.Tracef("project id %s, name %s", id, name)
		ds.Projects[id] = name
	}

	// TODO what to do with entitlements?
	//entitlements, err := ds.Client.GetEntitlements("51eeb1ad-96fc-46a3-8671-7cd6501db9a2")
	//doOrDie(err)
	//log.Debugf("entitlements: %s", entitlements)

	// TODO what to do with groups?
	//groups, err := ds.Client.GetGroups()
	//doOrDie(err)
	//log.Debugf("groups: %s", groups)

	apiRoles, err := ds.Client.GetRoles()
	if err != nil {
		return err
	}
	log.Infof("found %d roles", len(apiRoles.Data))
	ds.Roles = map[string]string{}
	for _, role := range apiRoles.Data {
		id, name := role.Id, role.Attributes.RoleName
		ds.Roles[id] = name
	}

	apiUsers, err := ds.Client.GetUsers(0, ds.UsersToCreate*10)
	if err != nil {
		return err
	}
	log.Infof("found %d users", len(apiUsers.Data))
	ds.Users = map[string]string{}
	ds.UserNameToId = map[string]string{}
	for _, user := range apiUsers.Data {
		id, name := user.Id, user.Attributes.Name
		ds.Users[id] = name
		if _, ok := ds.UserNameToId[name]; ok {
			panic(fmt.Sprintf("found duplicate username: %s, ids %s and %s", name, id, ds.UserNameToId[name]))
		}
		ds.UserNameToId[name] = id
	}

	apiOrgs, err := ds.Client.GetOrganizations()
	if err != nil {
		return err
	}
	log.Infof("found %d orgs", len(apiOrgs.Data))
	ds.Organizations = map[string]string{}
	for _, org := range apiOrgs.Data {
		id, name := org.Id, org.Attributes.OrganizationName
		ds.Organizations[id] = name
	}

	apiRoleAssns, err := ds.Client.GetRoleAssignments(0, 100)
	if err != nil {
		return err
	}
	log.Infof("found %d roleAssignments", len(apiRoleAssns.Data))
	ds.RoleAssignments = map[string]*RoleAssignment{}
	for _, apiRoleAssn := range apiRoleAssns.Data {
		orgId := apiRoleAssn.Relationships["organization"].Data.Id
		roleId := apiRoleAssn.Relationships["role"].Data.Id
		userId := apiRoleAssn.Relationships["user"].Data.Id
		ra := &RoleAssignment{
			Organization: ds.Organizations[orgId],
			User:         ds.Users[userId],
			Role:         ds.Roles[roleId],
			Object:       apiRoleAssn.Attributes.Object,
		}
		pieces := strings.Split(apiRoleAssn.Attributes.Object, ":")
		if pieces[2] == "organizations" {
			ra.ToOrganization = ds.Organizations[pieces[3]]
		} else if pieces[2] == "projects" {
			ra.ToProject = ds.Projects[pieces[3]]
		} else {
			panic(fmt.Sprintf("unexpected object type %s from %s", pieces[2], apiRoleAssn.Attributes.Object))
		}
		ds.RoleAssignments[apiRoleAssn.Id] = ra
		log.Tracef("found role assignment: %+v", ra)
	}

	return nil
}

type CreateRoleAssignmentJob struct {
	JobId     int
	UserId    string
	RoleId    string
	ProjectId string
	OrgId     string
}

func (ds *DataSeeder) CreateRoleAssignments(stop <-chan struct{}) error {
	orgId, _, err := ds.getOrg()
	if err != nil {
		return err
	}
	roleId, _, err := ds.getRole()
	if err != nil {
		return err
	}

	userJobs := make(chan *CreateUserJob, ds.UsersToCreate)
	didCreateUser := make(chan string, ds.UsersToCreate)
	roleAssignmentJobs := make(chan *CreateRoleAssignmentJob, len(ds.Projects)*ds.UsersToCreate)

	for id := 0; id < ds.UsersToCreate; id++ {
		userJobs <- &CreateUserJob{
			JobId: id,
			Email: fmt.Sprintf("test-user-%d@synopsys.com", id),
			Name:  fmt.Sprintf("test-user-%d", id),
			OrgId: orgId,
		}
	}

	go func() {
		jobId := 0
		for {
			var userId string
			select {
			case <-stop:
				break
			case userId = <-didCreateUser:
			}
			for projectId, _ := range ds.Projects {
				roleAssignmentJobs <- &CreateRoleAssignmentJob{
					JobId:     jobId,
					UserId:    userId,
					RoleId:    roleId,
					ProjectId: projectId,
					OrgId:     orgId,
				}
				jobId++
			}
		}
	}()

	for workerId := 0; workerId < ds.Concurrency; workerId++ {
		go func(workerId int) {
			ds.createUsersHelper(workerId, userJobs, didCreateUser, stop)
		}(workerId)
		go func(workerId int) {
			ds.createRoleAssignmentsHelper(workerId, roleAssignmentJobs, stop)
		}(workerId)
	}

	return nil
}

type CreateUserJob struct {
	JobId int
	Email string
	Name  string
	OrgId string
}

func (ds *DataSeeder) createUsersHelper(workerId int, jobs <-chan *CreateUserJob, didCreateUser chan<- string, stop <-chan struct{}) {
	for {
		var job *CreateUserJob
		select {
		case <-stop:
			break
		case job = <-jobs:
		}

		userId, ok := ds.UserNameToId[job.Name]
		// user doesn't exist already?  let's create them
		if !ok {
			user, err := ds.Client.CreateUser(job.Email, job.Name, job.OrgId)
			recordEvent("create_user", err)
			log.Debugf("worker %d, job %d, create user %s, success? %t", workerId, job.JobId, job.Email, err == nil)
			if err == nil {
				didCreateUser <- user.Data.Id
			} else {
				log.Errorf("unable to create user: %s", err)
			}
		} else {
			log.Debugf("user %s, %s already found, skipping creation", userId, job.Name)
			didCreateUser <- userId
		}
		time.Sleep(500 * time.Millisecond)
	}
}

func (ds *DataSeeder) createRoleAssignmentsHelper(workerId int, jobs <-chan *CreateRoleAssignmentJob, stop <-chan struct{}) error {
	for {
		var job *CreateRoleAssignmentJob
		select {
		case <-stop:
			break
		case job = <-jobs:
		}

		_, err := ds.Client.CreateRoleAssignment(job.UserId, job.RoleId, job.ProjectId, job.OrgId)
		recordEvent("create_role_assignment", err)
		if err != nil {
			log.Errorf("worker %d: job %d, unable to create role assignment: data %+v, error %s", workerId, job.JobId, job, err)
		} else {
			log.Infof("worker %d: job %d, created role assignment: data %+v", workerId, job.JobId, job)
		}

		time.Sleep(500 * time.Millisecond)
	}
}

func (ds *DataSeeder) getRole() (string, string, error) {
	if len(ds.Roles) == 0 {
		return "", "", errors.New(fmt.Sprintf("unable to create role assignments: no roles found"))
	}
	var id, name string
	for id, name = range ds.Roles {
		break
	}
	return id, name, nil
}

func (ds *DataSeeder) getOrg() (string, string, error) {
	if len(ds.Organizations) == 0 {
		return "", "", errors.New(fmt.Sprintf("unable to create role assignments: no orgs found"))
	}
	var id, name string
	for id, name = range ds.Organizations {
		break
	}
	return id, name, nil
}

func (ds *DataSeeder) oldCreateRoleAssignmentsDeprecated() error {
	orgId, _, err := ds.getOrg()
	if err != nil {
		return err
	}

	for userId, user := range ds.Users {
		for roleId, role := range ds.Roles {
			for projectId, project := range ds.Projects {
				ra, err := ds.Client.CreateRoleAssignment(userId, roleId, projectId, orgId)
				log.Debugf("ra: %s", ra)
				if err != nil {
					log.Errorf("unable to add role %s to user %s for project %s", role, user, project)
				} else {
					log.Infof("successfully added role %s to user %s for project %s", role, user, project)
				}
			}
		}
	}

	return nil
}
