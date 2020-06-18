package main

import (
	"encoding/json"
	"fmt"
	"github.com/blackducksoftware/cerebros/go/pkg/polaris/api"
	"github.com/blackducksoftware/cerebros/go/pkg/polaris/api-load/stress_testing"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"os"
	"time"
)

func doOrDie(err error) {
	if err != nil {
		log.Fatalf("%+v", err)
	}
}

type Config struct {
	LogLevel string

	PolarisURL      string
	PolarisEmail    string
	PolarisPassword string

	ServiceAccountName     string
	ServiceAccountPassword string

	ProjectLimit int
	RoleName     string
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

func main() {
	config, err := GetConfig(os.Args[1])
	doOrDie(err)

	logLevel, err := config.GetLogLevel()
	doOrDie(err)
	log.SetLevel(logLevel)

	url := config.PolarisURL
	email := config.PolarisEmail
	password := config.PolarisPassword
	serviceAccountName := config.ServiceAccountName
	serviceAccountEmail := fmt.Sprintf("%s@synopsys.com", config.ServiceAccountName)
	serviceAccountPassword := config.ServiceAccountPassword
	limit := config.ProjectLimit
	roleName := config.RoleName

	client := api.NewClient(url, email, password)
	doOrDie(client.Authenticate())

	// 1. prelim: get an orgId
	log.Infof("getting organizations")
	orgs, err := client.GetOrganizations()
	doOrDie(err)
	orgId := orgs.Data[0].Id
	log.Infof("found orgid %s from orgs meta %+v", orgId, orgs.Meta)

	// 2. prelim: get the 'Project Administrator' roleId
	roles, err := client.GetRoles()
	doOrDie(err)
	roleId := ""
	for _, role := range roles.Data {
		if role.Attributes.RoleName == roleName {
			roleId = role.Id
			break
		}
	}
	if roleId == "" {
		bytes, err := json.MarshalIndent(roles, "", "  ")
		doOrDie(err)
		log.Infof("struct: \n%s\n", bytes)
		log.Fatalf("unable to find role %s in %+v", roleName, roles)
	}
	log.Infof("found roleId %s for role %s", roleId, roleName)

	// 3. find projects that have main-branches and issue
	pf := stress_testing.NewProjectFetcher(client, 0, 1000000)
	pf.Start()

	for {
		if pf.MainBranchProjectsLength() >= limit || pf.IsDone() {
			pf.Stop()
			break
		}
		log.Infof("found %d out of %d projects, waiting ...", pf.MainBranchProjectsLength(), limit)
		time.Sleep(5 * time.Second)
	}
	log.Infof("finished fetching projects: found %d out of %d", pf.MainBranchProjectsLength(), limit)

	// 4. create service account and grab its id
	log.Infof("creating user %s, email %s", serviceAccountName, serviceAccountEmail)
	cru, err := client.CreateServiceAccount(serviceAccountEmail, serviceAccountName, orgId, serviceAccountPassword)
	doOrDie(err)
	userId := cru.Data.Id
	log.Infof("created user %s with id %s", serviceAccountName, userId)

	// 5. create role assignments
	for i := 0; i < pf.MainBranchProjectsLength() && i < limit; i++ {
		project := pf.GetMainBranchProject(i)
		_, err := client.CreateRoleAssignment(userId, roleId, project.ProjectId, orgId)
		doOrDie(err)
		log.Infof("created role assignment %d for user %s, role %s, project %s, org %s", i, userId, roleId, project.ProjectId, orgId)
	}
}
