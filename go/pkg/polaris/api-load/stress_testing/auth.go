package stress_testing

import (
	"fmt"
	"github.com/blackducksoftware/cerebros/go/pkg/polaris/api"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"sync"
	"time"
)

type PostRoleAssignmentsSource struct {
	Name            string
	client          *api.Client
	userIds         []string
	userPage        int
	userPageSize    int
	projectIds      []string
	projectPage     int
	projectPageSize int
	orgId           string
	roleId          string
	mux             *sync.Mutex
}

func NewPostRoleAssignmentsSource(name string, client *api.Client) (*PostRoleAssignmentsSource, error) {
	pras := &PostRoleAssignmentsSource{
		Name:            name,
		client:          client,
		userIds:         []string{},
		userPage:        0,
		userPageSize:    10,
		projectIds:      []string{},
		projectPage:     0,
		projectPageSize: 10,
		mux:             &sync.Mutex{},
	}

	orgs, err := client.GetOrganizations()
	if err != nil {
		return nil, errors.WithMessagef(err, "unable to get organizations for PostRoleAssignmentsSource")
	}
	log.Infof("number of orgs: %d", len(orgs.Data))
	if len(orgs.Data) == 0 {
		return nil, errors.WithMessagef(err, "no organizations found for PostRoleAssignmentsSource")
	}
	pras.orgId = orgs.Data[0].Id

	roles, err := client.GetRoles()
	if err != nil {
		return nil, errors.WithMessagef(err, "unable to get roles for PostRoleAssignmentsSource")
	}
	log.Infof("number of roles: %d", len(roles.Data))
	if len(roles.Data) == 0 {
		return nil, errors.WithMessagef(err, "no roles found for PostRoleAssignmentsSource")
	}
	pras.roleId = roles.Data[0].Id

	return pras, nil
}

func (pras *PostRoleAssignmentsSource) getNextProjectPage() int {
	page := pras.projectPage
	recordEventGauge(fmt.Sprintf("%sProjectPage", pras.Name), page)
	pras.projectPage++
	return page
}

func (pras *PostRoleAssignmentsSource) resetProjectPage() {
	recordEvent(fmt.Sprintf("%sResetProjectPage", pras.Name), nil)
	pras.projectPage = 0
}

func (pras *PostRoleAssignmentsSource) getProjectId() string {
	pras.mux.Lock()
	defer pras.mux.Unlock()

	if len(pras.projectIds) == 0 {
		page := pras.getNextProjectPage()
		offset := page * pras.projectPageSize
		var projects *api.GetVinylV0ProjectsResponse
		var err error
		for {
			projects, err = pras.client.GetVinylV0Projects(offset, pras.projectPageSize)
			if err == nil {
				break
			}
			log.Errorf("unable to get vinyl projects: %+v", err)
			time.Sleep(500 * time.Millisecond)
		}
		recordEventGauge(fmt.Sprintf("%sProjectTotal", pras.Name), projects.Meta.Total)
		recordEventGauge(fmt.Sprintf("%sProjectLimit", pras.Name), projects.Meta.Limit)
		recordEventGauge(fmt.Sprintf("%sProjectOffset", pras.Name), projects.Meta.Offset)
		if (offset + pras.projectPageSize) >= projects.Meta.Total {
			pras.resetProjectPage()
		}
		for _, proj := range projects.Data {
			pras.projectIds = append(pras.projectIds, proj.Id)
		}
	}
	first, rest := pras.projectIds[0], pras.projectIds[1:]
	pras.projectIds = rest
	return first
}

func (pras *PostRoleAssignmentsSource) getNextUserPage() int {
	page := pras.userPage
	recordEventGauge(fmt.Sprintf("%sUserPage", pras.Name), page)
	pras.userPage++
	return page
}

func (pras *PostRoleAssignmentsSource) resetUserPage() {
	recordEvent(fmt.Sprintf("%sResetUserPage", pras.Name), nil)
	pras.userPage = 0
}

func (pras *PostRoleAssignmentsSource) getUserId() string {
	pras.mux.Lock()
	defer pras.mux.Unlock()

	if len(pras.userIds) == 0 {
		page := pras.getNextUserPage()
		offset := page * pras.userPageSize
		var users *api.GetUsersResponse
		var err error
		for {
			users, err = pras.client.GetUsers(offset, pras.userPageSize)
			if err == nil {
				break
			}
			log.Errorf("unable to get users: %+v", err)
			time.Sleep(500 * time.Millisecond)
		}
		log.Infof("got users: %+v", users.Meta)
		recordEventGauge(fmt.Sprintf("%sUsersTotal", pras.Name), users.Meta.Total)
		recordEventGauge(fmt.Sprintf("%sUsersLimit", pras.Name), users.Meta.Limit)
		recordEventGauge(fmt.Sprintf("%sUsersOffset", pras.Name), users.Meta.Offset)
		if (offset + pras.userPageSize) >= users.Meta.Total {
			pras.resetUserPage()
		}
		for _, user := range users.Data {
			pras.userIds = append(pras.userIds, user.Id)
		}
	}
	first, rest := pras.userIds[0], pras.userIds[1:]
	pras.userIds = rest
	return first
}

func (pras *PostRoleAssignmentsSource) RunJob() (string, error) {
	start := time.Now()
	projectId := pras.getProjectId()
	recordDuration(fmt.Sprintf("%sGetProjectId", pras.Name), time.Since(start))
	log.Infof("%s got project id: %s", pras.Name, projectId)

	userStart := time.Now()
	userId := pras.getUserId()
	recordDuration(fmt.Sprintf("%sGetUserId", pras.Name), time.Since(userStart))
	log.Infof("%s got user id: %s", pras.Name, projectId)

	ras, err := pras.client.CreateRoleAssignment(userId, pras.roleId, projectId, pras.orgId)
	if err != nil {
		log.Errorf("unable to create role assignment: %s, %s, %s, %s, %s, %+v", userId, pras.roleId, projectId, pras.orgId, ras, err)
	} else {
		log.Debugf("created role assignment: %s, %s, %s, %s, %s", userId, pras.roleId, projectId, pras.orgId, ras)
	}
	return pras.Name, err
}

type RoleAssignmentsPagerSource struct {
	Page     int
	mux      *sync.Mutex
	client   *api.Client
	Name     string
	PageSize int
}

func NewRoleAssignmentsPagerSource(name string, client *api.Client, start int, pageSize int) *RoleAssignmentsPagerSource {
	return &RoleAssignmentsPagerSource{
		Page:     start / pageSize,
		mux:      &sync.Mutex{},
		client:   client,
		Name:     name,
		PageSize: pageSize,
	}
}

func (raps *RoleAssignmentsPagerSource) getPage() int {
	raps.mux.Lock()
	defer raps.mux.Unlock()
	i := raps.Page
	raps.Page += 1
	return i
}

func (raps *RoleAssignmentsPagerSource) resetPage() {
	raps.mux.Lock()
	defer raps.mux.Unlock()
	raps.Page = 0
}

func (raps *RoleAssignmentsPagerSource) RunJob() (string, error) {
	page := raps.getPage()
	recordEventGauge(fmt.Sprintf("%sPage", raps.Name), page)
	offset := page * raps.PageSize
	ras, err := raps.client.GetRoleAssignments(offset, raps.PageSize)
	if err == nil {
		recordEventGauge(fmt.Sprintf("%sTotal", raps.Name), ras.Meta.Total)
		if offset >= ras.Meta.Total {
			recordEvent(fmt.Sprintf("%sResetPage", raps.Name), nil)
			raps.resetPage()
		}
	}
	return raps.Name, err
}

type RoleAssignmentsSingleProjectSource struct {
	projects *ProjectFetcher
	mux      *sync.Mutex
	client   *api.Client
	Index    int
}

func NewRoleAssignmentsSingleProjectSource(client *api.Client, projects *ProjectFetcher) *RoleAssignmentsSingleProjectSource {
	return &RoleAssignmentsSingleProjectSource{
		projects: projects,
		mux:      &sync.Mutex{},
		client:   client,
		Index:    0,
	}
}

func (raps *RoleAssignmentsSingleProjectSource) getNextIndex() (int, bool) {
	raps.mux.Lock()
	defer raps.mux.Unlock()

	projectsCount := raps.projects.ProjectsLength()
	if projectsCount == 0 {
		return -1, false
	}

	if raps.Index >= projectsCount {
		recordEvent("resetting rollupCounts index", nil)
		raps.Index = 0
	}
	log.Debugf("rollupCounts index %d", raps.Index)
	i := raps.Index
	recordProjectRollupCountsIndex(i)
	raps.Index++
	return i, true
}

func (raps *RoleAssignmentsSingleProjectSource) RunJob() (string, error) {
	index, ok := raps.getNextIndex()
	if !ok {
		return "getRoleAssignmentsSingleProject -- no project available", errors.New(fmt.Sprintf("no project available"))
	}
	project := raps.projects.GetProject(index)
	recordRoleAssignmentsSingleProjectIndex(index)
	_, err := raps.client.GetRoleAssignmentsForProject(project.Id)
	return "getRoleAssignmentsSingleProject", err
}

type AuthLoadGenerator struct {
	mux                                     *sync.Mutex
	Config                                  *AuthConfig
	stopChan                                chan struct{}
	entitlementsLoadManager                 *LoadManager
	loginsLoadManager                       *LoadManager
	roleAssignmentsLoadManagers             map[string]*LoadManager
	roleAssignmentsSingleProjectLoadManager *LoadManager
	createRoleAssignmentsLoadManager        *LoadManager
}

func NewAuthLoadGenerator(projects *ProjectFetcher, url string, email string, password string, config *AuthConfig) *AuthLoadGenerator {
	alg := &AuthLoadGenerator{
		mux:                         &sync.Mutex{},
		Config:                      config,
		stopChan:                    make(chan struct{}),
		roleAssignmentsLoadManagers: map[string]*LoadManager{},
	}

	entitlementsClient := api.NewClient(url, email, password)
	err := entitlementsClient.Authenticate()
	doOrDie(err)
	Reauthenticator(entitlementsClient, alg.stopChan)
	orgs, err := entitlementsClient.GetOrganizations()
	doOrDie(err)
	log.Infof("found %d orgs", len(orgs.Data))
	// just crash if there's 0 orgs
	org := orgs.Data[0]

	entitlementsJob := func() (string, error) {
		entitlements, err := entitlementsClient.GetEntitlementsForOrganization(org.Id)
		log.Infof("found %d entitlements", len(entitlements.Data))
		return "entitlements", err
	}
	alg.entitlementsLoadManager = NewLoadManager(
		"entitlements",
		&FuncJobSource{function: entitlementsJob},
		config.Entitlements.WorkersCount,
		config.Entitlements.Rate.MustRateLimiter("entitlements"))

	loginJob := func() (string, error) {
		client := api.NewClient(url, email, password)
		err := client.Authenticate()
		return "login", err
	}
	alg.loginsLoadManager = NewLoadManager(
		"logins",
		&FuncJobSource{function: loginJob},
		config.Login.WorkersCount,
		config.Login.Rate.MustRateLimiter("logins"))

	roleAssignmentsClient := api.NewClient(url, email, password)
	err = roleAssignmentsClient.Authenticate()
	doOrDie(err)
	Reauthenticator(roleAssignmentsClient, alg.stopChan)
	rap := config.RoleAssignmentsPager
	for name, conf := range rap {
		debugName := fmt.Sprintf("role-assignments-pager-%s", name)
		alg.roleAssignmentsLoadManagers[name] = NewLoadManager(
			debugName,
			NewRoleAssignmentsPagerSource(debugName, roleAssignmentsClient, 0, conf.PageSize),
			conf.LoadConfig.WorkersCount,
			conf.LoadConfig.Rate.MustRateLimiter(debugName))
	}

	alg.roleAssignmentsSingleProjectLoadManager = NewLoadManager(
		"role-assignments-single-project",
		NewRoleAssignmentsSingleProjectSource(roleAssignmentsClient, projects),
		config.RoleAssignmentsSingleProject.WorkersCount,
		config.RoleAssignmentsSingleProject.Rate.MustRateLimiter("role-assignments-single-project"))

	pras, err := NewPostRoleAssignmentsSource("createRoleAssignments", roleAssignmentsClient)
	doOrDie(err)
	alg.createRoleAssignmentsLoadManager = NewLoadManager(
		"createRoleAssignments",
		pras,
		config.CreateRoleAssignments.WorkersCount,
		config.CreateRoleAssignments.Rate.MustRateLimiter("createRoleAssignments"))

	return alg
}

func (alg *AuthLoadGenerator) stop() {
	close(alg.stopChan)
	if alg.entitlementsLoadManager != nil {
		alg.entitlementsLoadManager.stop()
	}
	if alg.loginsLoadManager != nil {
		alg.loginsLoadManager.stop()
	}
	if alg.roleAssignmentsLoadManagers != nil {
		for _, lm := range alg.roleAssignmentsLoadManagers {
			lm.stop()
		}
	}
}
