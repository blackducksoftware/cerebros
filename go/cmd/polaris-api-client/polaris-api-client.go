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
	"github.com/blackducksoftware/cerebros/go/pkg/polaris/api"
	log "github.com/sirupsen/logrus"
)

func doOrDie(err error) {
	if err != nil {
		log.Fatalf("oops: %+v", err)
	}
}

func main() {
	url := "TODO"
	email := "TODO"
	password := "TODO"
	client := api.NewClient(url, email, password)

	log.SetLevel(log.DebugLevel)

	doOrDie(client.Authenticate())

	//j1, err := client.GetJson(map[string]interface{}{"page[limit]": "10"}, nil, "api/auth/role-assignments")
	//log.Infof("first 10 role assignments: \n\n%s\n\n", j1)
	//j2, err := client.GetJson(map[string]interface{}{"page[limit]": "10", "page[offset]": "100"}, nil, "api/auth/role-assignments")
	//log.Infof("role assignments 100-109: \n\n%s\n\n", j2)

	vinyl, err := client.GetVinylV0Projects(0, 1)
	log.Infof("vinyl: %+v", vinyl.Meta)

	projId := vinyl.Data[0].Id
	ras, err := client.GetRoleAssignmentsForProject(projId)
	doOrDie(err)
	log.Infof("role assignments for %s: \n\n%+v\n\n", projId, ras)

	orgs, err := client.GetOrganizations()
	doOrDie(err)
	log.Infof("orgs: %d", len(orgs.Data))
	orgId := orgs.Data[0].Id

	//params := map[string]interface{}{
	//	"filter[entitlements][object][eq]": fmt.Sprintf("urn:x-swip:organizations:%s", orgId),
	//}
	//json, err := client.GetJson(params, nil, "api/auth/entitlements")
	//doOrDie(err)
	//log.Infof("org entitlements:\n\n%s\n\n", json)

	//params := map[string]string{
	//	"page[limit]": "200000",
	//	"page[offset]": "100",
	//}
	//users, err := client.GetJson("api/auth/users", params, nil)
	//doOrDie(err)
	//fmt.Printf("users: \n\n%s\n\n", users)

	//users, err := client.GetUsers(0, 200000)
	//doOrDie(err)
	//log.Infof("found %d users, %+v", len(users.Data), users.Meta)

	users, err := client.GetUsers(0, 10)
	doOrDie(err)
	log.Infof("found %+v users", users.Meta) //.Meta)

	raCount, err := client.GetRoleAssignments(0, 1)
	doOrDie(err)
	log.Infof("found %d role assignments, %+v", raCount.Meta.Total, raCount.Meta)

	//roleCount, err := client.GetRoles(0)
	//doOrDie(err)
	//log.Infof("found %d roles", roleCount.Meta.Total)

	roles, err := client.GetRoles()
	doOrDie(err)
	log.Infof("roles: \n%+v\n", roles)

	for _, user := range users.Data {
		for _, role := range roles.Data {
			resp, err := client.CreateRoleAssignment(user.Id, role.Id, projId, orgId)
			if err != nil {
				log.Errorf("unable to create role assignment: %s, %s, %s, %s, %s, %+v", user.Id, role.Id, projId, orgId, resp, err)
			} else {
				log.Infof("created role assignment: %s, %s, %s, %s, %s", user.Id, role.Id, projId, orgId, resp)
			}
		}
	}
}
