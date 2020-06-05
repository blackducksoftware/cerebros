package stress_testing

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"sync"
)

type JobSource interface {
	// RunJob must be synchronous and reentrant
	RunJob() (string, error)
}

type FuncJobSource struct {
	function func() (string, error)
}

func (fjs *FuncJobSource) RunJob() (string, error) {
	return fjs.function()
}

type LoadManager struct {
	Name      string
	jobSource JobSource
	workers   []*LoadWorker
	limiter   *RateLimiter
	mux       *sync.Mutex
	stopChan  chan struct{}
}

func NewLoadManager(name string, js JobSource, workerCount int, rateLimiter *RateLimiter) *LoadManager {
	lm := &LoadManager{
		Name:      name,
		jobSource: js,
		workers:   []*LoadWorker{},
		limiter:   rateLimiter,
		mux:       &sync.Mutex{},
		stopChan:  make(chan struct{}),
	}
	lm.setWorkerCount(workerCount)
	return lm
}

func (lm *LoadManager) stop() {
	for _, worker := range lm.workers {
		worker.stop()
	}
	close(lm.stopChan)
}

func (lm *LoadManager) setWorkerCount(count int) {
	for i := 0; i < count; i++ {
		recordNamedEventBy("createWorker", lm.Name, 1)
		workerId := fmt.Sprintf("%s-%d", lm.Name, i)
		log.Infof("spinning up worker %s", workerId)
		lm.workers = append(lm.workers, NewLoadWorker(lm.runJob, workerId))
	}
}

func (lm *LoadManager) runJob() {
	lm.limiter.Wait()
	lm.limiter.Finish(lm.jobSource.RunJob())
}

type LoadWorker struct {
	WorkerId  string
	stopChan  chan struct{}
	runJob    func()
	nextJobId int
}

func NewLoadWorker(runJob func(), workerId string) *LoadWorker {
	cw := &LoadWorker{
		WorkerId:  workerId,
		stopChan:  make(chan struct{}),
		runJob:    runJob,
		nextJobId: 0}
	cw.start()
	return cw
}

func (cw *LoadWorker) start() {
	go func() {
		for {
			select {
			case <-cw.stopChan:
				return
			default:
				log.Infof("worker running job %s-%d", cw.WorkerId, cw.nextJobId)
				cw.runJob()
				cw.nextJobId++
			}
		}
	}()
}

func (cw *LoadWorker) stop() {
	log.Infof("cleaning up worker %s", cw.WorkerId)
	close(cw.stopChan)
}
