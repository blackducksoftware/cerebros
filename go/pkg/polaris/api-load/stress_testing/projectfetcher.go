package stress_testing

import (
	"fmt"
	"github.com/blackducksoftware/cerebros/go/pkg/polaris/api"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"math/rand"
	"sync"
	"time"
)

type MainBranchProject struct {
	ProjectId    string
	MainBranchId string
}

type ProjectFetcher struct {
	StartIndex         int
	Limit              int
	isDone             bool
	client             *api.Client
	stopChan           chan struct{}
	projects           []*api.VinylV0Project
	mainBranchProjects []*MainBranchProject
	mux                *sync.Mutex
}

func NewProjectFetcher(client *api.Client, start int, limit int) *ProjectFetcher {
	pf := &ProjectFetcher{
		StartIndex:         start,
		Limit:              limit,
		isDone:             false,
		client:             client,
		stopChan:           make(chan struct{}),
		projects:           []*api.VinylV0Project{},
		mainBranchProjects: []*MainBranchProject{},
		mux:                &sync.Mutex{},
	}
	return pf
}

func NewProjectFetcherWithRandomStart(client *api.Client, limit int) *ProjectFetcher {
	// passing in limit: 1 lets us get the total number of projects without
	// getting a huge response back
	// no idea why limit: 0 doesn't work, but that gives a total of 0
	vinylInit, err := client.GetVinylV0Projects(0, 1)
	doOrDie(err)
	recordEvent("get initial projects", err)
	// TODO for now, let's just not worry about errors
	rand.Seed(time.Now().UTC().UnixNano())
	if vinylInit.Meta.Total == 0 {
		panic(errors.Errorf("unable to instantiate project fetcher: 0 projects found (meta %+v)", vinylInit.Meta))
	}
	initialOffset := rand.Intn(vinylInit.Meta.Total)
	return NewProjectFetcher(client, initialOffset, limit)
}

func (pf *ProjectFetcher) ProjectsLength() int {
	pf.mux.Lock()
	defer pf.mux.Unlock()
	return len(pf.projects)
}

func (pf *ProjectFetcher) MainBranchProjectsLength() int {
	pf.mux.Lock()
	defer pf.mux.Unlock()
	return len(pf.mainBranchProjects)
}

func (pf *ProjectFetcher) IsDone() bool {
	pf.mux.Lock()
	defer pf.mux.Unlock()
	return pf.isDone
}

func (pf *ProjectFetcher) GetProject(index int) *api.VinylV0Project {
	pf.mux.Lock()
	defer pf.mux.Unlock()
	if index < 0 || index >= len(pf.projects) {
		panic(fmt.Sprintf("invalid index %d, projects length %d", index, len(pf.projects)))
	}
	return pf.projects[index]
}

func (pf *ProjectFetcher) GetMainBranchProject(index int) *MainBranchProject {
	pf.mux.Lock()
	defer pf.mux.Unlock()
	if index < 0 || index >= len(pf.mainBranchProjects) {
		panic(fmt.Sprintf("invalid index %d, mainBranchProjects length %d", index, len(pf.mainBranchProjects)))
	}
	return pf.mainBranchProjects[index]
}

func (pf *ProjectFetcher) Stop() {
	close(pf.stopChan)
}

func (pf *ProjectFetcher) Start() {
	initialOffset := pf.StartIndex
	limit := pf.Limit
	pageSize := 10
	workerId := 0
	fetched := 0
	pageStart := initialOffset / pageSize
	go func() {
	ForLoop:
		for page := pageStart; fetched < limit; {
			select {
			case <-pf.stopChan:
				break ForLoop
			default:
			}
			recordEventGauge("projectPage", page)
			offset := page * pageSize
			log.Infof("projects worker %d: getting projects with offset %d, pagesize %d", workerId, offset, pageSize)
			vinyl, err := pf.client.GetVinylV0Projects(offset, pageSize)
			if err != nil {
				// TODO if there's an error, this will just spin and spin and spin ...
				log.Errorf("projects worker %d: unable to get vinyl projects: %+v", workerId, err)
				time.Sleep(1 * time.Second)
				continue
			}
			log.Infof("projects worker %d: vinyl meta: %+v", workerId, vinyl.Meta)

			go func() {
				pf.mux.Lock()

				defer pf.mux.Unlock()
				for _, project := range vinyl.Data {
					//start := time.Now()
					log.Infof("writing project %s", project.Id)
					pf.projects = append(pf.projects, project)
					//recordProjectWriteDuration(time.Since(start))
					if branch, ok := project.Relationships["main-branch"]; ok && branch.Data != nil {
						recordEvent("found main-branch", nil)
						pf.mainBranchProjects = append(pf.mainBranchProjects, &MainBranchProject{
							ProjectId:    project.Id,
							MainBranchId: project.Relationships["main-branch"].Data.Id,
						})
					} else {
						recordEvent("missing main-branch", nil)
					}
				}
			}()

			fetched += vinyl.Meta.Limit
			page++
			// wrap around // TODO should this send back to 0 instead?
			if offset >= vinyl.Meta.Total {
				break
			}
			// got back to where we started?  alright, we're done
			if page == pageStart {
				break
			}
		}
	}()

	pf.mux.Lock()
	pf.isDone = true
	pf.mux.Unlock()
}
