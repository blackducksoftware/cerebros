package stress_testing

import (
	"fmt"
	"github.com/paulbellamy/ratecounter"
	log "github.com/sirupsen/logrus"
	"sync"
	"time"
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
	Name              string
	jobSource         JobSource
	workers           []*LoadWorker
	limiter           *RateLimiter
	startRateCounter  *ratecounter.RateCounter
	finishRateCounter *ratecounter.RateCounter
	errorCounter      *ratecounter.RateCounter
	successCount      int
	errorCount        int
	jobsInProgress    int
	mux               *sync.Mutex
	stopChan          chan struct{}
}

func NewLoadManager(name string, js JobSource, workerCount int, rateLimiter *RateLimiter) *LoadManager {
	lm := &LoadManager{
		Name:              name,
		jobSource:         js,
		workers:           []*LoadWorker{},
		limiter:           rateLimiter,
		startRateCounter:  ratecounter.NewRateCounter(1 * time.Minute),
		finishRateCounter: ratecounter.NewRateCounter(1 * time.Minute),
		errorCounter:      ratecounter.NewRateCounter(1 * time.Minute),
		successCount:      0,
		errorCount:        0,
		jobsInProgress:    0,
		mux:               &sync.Mutex{},
		stopChan:          make(chan struct{}),
	}
	lm.setWorkerCount(workerCount)
	go lm.recordMetrics()
	return lm
}

func (lm *LoadManager) errorFraction() float64 {
	errorRate := float64(lm.errorCounter.Rate())
	total := float64(lm.finishRateCounter.Rate())
	log.Debugf("error fraction: %s, %f, %f, %f", lm.Name, errorRate, total, errorRate/total)
	return errorRate / total
}

func (lm *LoadManager) recordMetrics() {
	for {
		select {
		case <-lm.stopChan:
			return
		case <-time.After(20 * time.Second):
			recordNamedEventGauge("errorFraction", lm.Name, lm.errorFraction())
			lm.mux.Lock()
			recordNamedEventGauge("jobsInProgress", lm.Name, float64(lm.jobsInProgress))
			recordNamedEventBy("errorCount", lm.Name, lm.errorCount)
			recordNamedEventBy("successCount", lm.Name, lm.successCount)
			lm.mux.Unlock()
		}
	}
}

func (lm *LoadManager) stop() {
	for _, worker := range lm.workers {
		worker.stop()
	}
	close(lm.stopChan)
}

func (lm *LoadManager) setWorkerCount(count int) {
	for i := 0; i < count; i++ {
		workerId := fmt.Sprintf("%s-%d", lm.Name, i)
		log.Infof("spinning up worker %s", workerId)
		lm.workers = append(lm.workers, NewLoadWorker(lm.runJob, workerId))
	}
}

func (lm *LoadManager) runJob() {
	waitName := fmt.Sprintf("%s-wait", lm.Name)
	start := time.Now()
	waitErr := lm.limiter.Wait()
	recordDuration(waitName, time.Since(start))
	doOrDie(waitErr)
	lm.startRateCounter.Incr(1)

	lm.mux.Lock()
	lm.jobsInProgress++
	lm.mux.Unlock()

	desc, err := lm.jobSource.RunJob()
	recordEvent(desc, err)

	lm.mux.Lock()
	defer lm.mux.Unlock()
	lm.jobsInProgress--
	lm.finishRateCounter.Incr(1)
	if err != nil {
		lm.errorCounter.Incr(1)
		lm.errorCount++
	} else {
		lm.successCount++
	}
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
