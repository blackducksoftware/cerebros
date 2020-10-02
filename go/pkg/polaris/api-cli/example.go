package api_cli

import (
	"github.com/blackducksoftware/cerebros/go/pkg/polaris/api"
	log "github.com/sirupsen/logrus"
)

type ExampleArgs struct {
	PolarisURL string
	Email      string
	Password   string
}

func runExample(args *ExampleArgs) {
	client := api.NewClient(args.PolarisURL, args.Email, args.Password)

	log.SetLevel(log.DebugLevel)

	DoOrDie(client.Authenticate())

	//j1, err := client.GetJson(map[string]interface{}{"page[limit]": "10"}, nil, "api/auth/role-assignments")
	//log.Infof("first 10 role assignments: \n\n%s\n\n", j1)
	//j2, err := client.GetJson(map[string]interface{}{"page[limit]": "10", "page[offset]": "100"}, nil, "api/auth/role-assignments")
	//log.Infof("role assignments 100-109: \n\n%s\n\n", j2)

	vinyl, err := client.GetVinylV0Projects(0, 1)
	log.Infof("vinyl: %+v", vinyl.Meta)

	projId := vinyl.Data[0].Id
	ras, err := client.GetRoleAssignmentsForProject(projId)
	DoOrDie(err)
	log.Infof("role assignments for %s: \n\n%+v\n\n", projId, ras)

	orgs, err := client.GetOrganizations()
	DoOrDie(err)
	log.Infof("orgs: %d", len(orgs.Data))
	orgId := orgs.Data[0].Id

	//params := map[string]interface{}{
	//	"filter[entitlements][object][eq]": fmt.Sprintf("urn:x-swip:organizations:%s", orgId),
	//}
	//json, err := client.GetJson(params, nil, "api/auth/entitlements")
	//DoOrDie(err)
	//log.Infof("org entitlements:\n\n%s\n\n", json)

	//params := map[string]string{
	//	"page[limit]": "200000",
	//	"page[offset]": "100",
	//}
	//users, err := client.GetJson("api/auth/users", params, nil)
	//DoOrDie(err)
	//fmt.Printf("users: \n\n%s\n\n", users)

	//users, err := client.GetUsers(0, 200000)
	//DoOrDie(err)
	//log.Infof("found %d users, %+v", len(users.Data), users.Meta)

	users, err := client.GetUsers(0, 10)
	DoOrDie(err)
	log.Infof("found %+v users", users.Meta) //.Meta)

	raCount, err := client.GetRoleAssignments(0, 1)
	DoOrDie(err)
	log.Infof("found %d role assignments, %+v", raCount.Meta.Total, raCount.Meta)

	//roleCount, err := client.GetRoles(0)
	//DoOrDie(err)
	//log.Infof("found %d roles", roleCount.Meta.Total)

	roles, err := client.GetRoles()
	DoOrDie(err)
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
