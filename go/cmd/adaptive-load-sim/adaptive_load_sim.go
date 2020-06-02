package main

import (
	"context"
	"fmt"
	resty2 "github.com/go-resty/resty/v2"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"golang.org/x/time/rate"
	"math/rand"
	"net/http"
	"os"
	"sync"
	"time"
)

// Config ...
type Config struct {
	Port int

	Client *struct {
		Workers   int
		ServerURL string
	}

	Server *struct {
		AverageResponseTimeMs int
		ResponseTimeRadiusMs  int
		Workers               int
	}

	LogLevel string
}

// GetLogLevel ...
func (config *Config) GetLogLevel() (log.Level, error) {
	return log.ParseLevel(config.LogLevel)
}

// GetConfig ...
func GetConfig(configPath string) (*Config, error) {
	var config *Config

	// avoid viper serialization problems caused by '.'s in keys
	v := viper.NewWithOptions(viper.KeyDelimiter("*"))

	v.SetConfigFile(configPath)
	err := v.ReadInConfig()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to ReadInConfig")
	}

	err = v.Unmarshal(&config)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to unmarshal config")
	}

	return config, nil
}

func doOrDie(err error) {
	if err != nil {
		log.Fatalf("%+v", err)
	}
}

func main() {
	configPath := os.Args[1]
	config, err := GetConfig(configPath)
	doOrDie(err)

	logLevel, err := config.GetLogLevel()
	doOrDie(err)

	log.SetLevel(logLevel)

	// TODO make this tear down client/server or something when it's closed
	stop := make(chan struct{})

	var clientOrServer interface{}
	if config.Client != nil {
		client := NewClient(config.Client.ServerURL, config.Client.Workers)
		clientOrServer = client
	} else if config.Server != nil {
		server := NewServer(config.Server.AverageResponseTimeMs, config.Server.ResponseTimeRadiusMs, config.Server.Workers)
		clientOrServer = server
	}

	http.Handle("/metrics", promhttp.Handler())

	addr := fmt.Sprintf(":%d", config.Port)
	log.Infof("successfully instantiated %+v, serving on %s", clientOrServer, addr)
	go func() {
		http.ListenAndServe(addr, nil)
	}()

	<-stop
}

type ClientWorker struct {
	WorkerId  string
	ServerURL string
	limiter   func()
	stopChan  chan struct{}
	handler   func(err error, code int)
	nextJobId int
	resty     *resty2.Client
}

func NewClientWorker(limiter func(), handler func(err error, code int), serverURL string, workerId string) *ClientWorker {
	cw := &ClientWorker{
		WorkerId:  workerId,
		ServerURL: serverURL,
		limiter:   limiter,
		stopChan:  make(chan struct{}),
		handler:   handler,
		nextJobId: 0,
		resty:     resty2.New()}
	cw.start()
	return cw
}

func (cw *ClientWorker) start() {
	go func() {
		for {
			select {
			case <-cw.stopChan:
				break
			default:
				cw.limiter()

				jobId := fmt.Sprintf("%s-%d", cw.WorkerId, cw.nextJobId)
				log.Tracef("worker %s issuing job id %s", cw.WorkerId, jobId)
				req := cw.resty.R()
				req.SetQueryParam("jobid", jobId)
				url := fmt.Sprintf("%s/%s", cw.ServerURL, "ping")

				start := time.Now()
				resp, err := req.Get(url)
				duration := time.Now().Sub(start)

				code := resp.StatusCode()
				if err != nil {
					log.Errorf("jobId %s failed with code %d: %+v", jobId, code, err)
					cw.handler(err, code)
				} else if code < 200 || code > 299 {
					log.Errorf("jobId %s failed with code %d", jobId, code)
					cw.handler(errors.New(fmt.Sprintf("bad status code: %d", code)), code)
				} else {
					log.Infof("worker %s finished job %s in %d ms: %s", cw.WorkerId, jobId, duration/time.Millisecond, resp)
					cw.handler(nil, code)
				}
				cw.nextJobId++
			}
		}
	}()
}

func (cw *ClientWorker) stop() {
	close(cw.stopChan)
}

type Client struct {
	ServerURL                          string
	NextWorkerId                       int
	consecutiveSuccessfulRequests      int
	consecutiveSuccessfulRequestsMutex sync.Mutex
	limiter                            *rate.Limiter
	errorLimiter                       *rate.Limiter
	errorMutex                         sync.Mutex
	resultMutex                        sync.Mutex
	lastErrorRateDropTime              time.Time
	workers                            []*ClientWorker
	errorCount                         int
	successCount                       int
}

func NewClient(serverURL string, workersCount int) *Client {
	c := &Client{
		ServerURL:                     serverURL,
		NextWorkerId:                  0,
		consecutiveSuccessfulRequests: 0,
		limiter:                       rate.NewLimiter(10, workersCount),
		errorLimiter:                  rate.NewLimiter(2, 5),
		lastErrorRateDropTime:         time.Now(),
		workers:                       []*ClientWorker{},
		errorCount:                    0,
		successCount:                  0,
	}
	c.SetWorkerCount(workersCount)
	return c
}

func (c *Client) incrementConsecutiveSuccessfulRequests() {
	c.consecutiveSuccessfulRequestsMutex.Lock()
	defer c.consecutiveSuccessfulRequestsMutex.Unlock()
	c.consecutiveSuccessfulRequests++
	if c.consecutiveSuccessfulRequests > 25 {
		newLimit := c.limiter.Limit() + 1
		log.Infof("increasing rate limit by 1 to %f", newLimit)
		c.limiter.SetLimit(newLimit)
		c.consecutiveSuccessfulRequests = 0
	}
}

func (c *Client) setConsecutiveSuccessfulRequests(n int) {
	c.consecutiveSuccessfulRequestsMutex.Lock()
	defer c.consecutiveSuccessfulRequestsMutex.Unlock()
	log.Infof("setting consecutiveSuccessfulRequests to %d", n)
	c.consecutiveSuccessfulRequests = 0
}

func (c *Client) exhaustedErrors() {
	log.Infof("exhausted error limit")
	c.errorMutex.Lock()
	defer c.errorMutex.Unlock()
	if time.Now().Sub(c.lastErrorRateDropTime).Seconds() < 10 {
		log.Warnf("recently dropped rate due to errors, holding off on further drops for now")
		return
	}
	log.Warningf("hit error limit, reducing request rate limit by a factor of 2")
	c.limiter.SetLimit(c.limiter.Limit() / 2)
	log.Warnf("limit now set to %+v", c.limiter.Limit())
	c.lastErrorRateDropTime = time.Now()
}

func (c *Client) result(success bool) {
	c.resultMutex.Lock()
	defer c.resultMutex.Unlock()
	if success {
		c.successCount++
	} else {
		c.errorCount++
	}
	log.Infof("results: %d success, %d error", c.successCount, c.errorCount)
}

func (c *Client) SetWorkerCount(count int) {
	limiter := func() {
		err := c.limiter.Wait(context.TODO())
		doOrDie(err)
	}
	handler := func(err error, code int) {
		log.Tracef("handler: %d, %+v", code, err)
		if err != nil {
			c.setConsecutiveSuccessfulRequests(0)
			if !c.errorLimiter.Allow() {
				c.exhaustedErrors()
			}
		} else {
			c.incrementConsecutiveSuccessfulRequests()
		}
		c.result(err == nil)
	}
	workersCount := len(c.workers)
	if count > workersCount {
		for i := workersCount; i < count; i++ {
			workerId := fmt.Sprintf("%d", c.NextWorkerId)
			log.Infof("spinning up worker %s", workerId)

			c.workers = append(c.workers, NewClientWorker(limiter, handler, c.ServerURL, workerId))
			c.NextWorkerId++
		}
	} else if count < workersCount {
		removed, rest := c.workers[count:], c.workers[:count]
		c.workers = rest
		for _, worker := range removed {
			log.Infof("spinning down worker %s", worker.WorkerId)
			worker.stop()
		}
	} // else don't do anything if count == s.WorkersCount
}

type Job struct {
	Id   string
	Done chan<- bool
}

type ServerWorker struct {
	WorkerId string
	stopChan chan struct{}
	work     <-chan *Job
	latency  func() time.Duration
}

func NewServerWorker(latency func() time.Duration, workerId string, work <-chan *Job) *ServerWorker {
	sw := &ServerWorker{WorkerId: workerId, stopChan: make(chan struct{}), work: work, latency: latency}
	sw.start()
	return sw
}

func (sw *ServerWorker) start() {
	go func() {
		for {
			log.Infof("worker %s waiting for jobs ...", sw.WorkerId)
			select {
			case <-sw.stopChan:
				break
			case job := <-sw.work:
				log.Infof("worker %s handling id %s", sw.WorkerId, job.Id)
				lat := sw.latency()
				time.Sleep(lat)
				log.Infof("worker %s finished id %s in %d ms", sw.WorkerId, job.Id, lat/time.Millisecond)
				job.Done <- true
			}
		}
	}()
}

func (sw *ServerWorker) stop() {
	close(sw.stopChan)
}

type Server struct {
	AverageResponseTimeMs int
	ResponseTimeRadiusMs  int
	nextWorkerId          int
	workers               []*ServerWorker
	jobs                  chan *Job
	latency               func() time.Duration
}

func NewServer(averageResponseTimeMs int, responseTimeRadiusMs int, workers int) *Server {
	s := &Server{
		AverageResponseTimeMs: averageResponseTimeMs,
		ResponseTimeRadiusMs:  responseTimeRadiusMs,
		nextWorkerId:          0,
		workers:               []*ServerWorker{},
	}
	s.makeLatency()
	s.jobs = make(chan *Job)
	s.SetWorkerCount(workers)
	httpHandler := func(jobId string) (string, int) {
		done := make(chan bool)
		job := &Job{
			Id:   jobId,
			Done: done,
		}
		select {
		case s.jobs <- job:
			<-done
			return "success", 200
		default:
			log.Infof("backpressuring: no workers available")
			return "no workers available", 503
		}
	}
	SetupHTTPServer(httpHandler)
	return s
}

func (s *Server) makeLatency() {
	s.latency = func() time.Duration {
		return time.Duration(rand.Intn(s.ResponseTimeRadiusMs)+s.AverageResponseTimeMs) * time.Millisecond
	}
}

// TODO protect against concurrency
func (s *Server) SetWorkerCount(count int) {
	workersCount := len(s.workers)
	log.Infof("set worker count: have %d, want %d", workersCount, count)
	if count > workersCount {
		for i := workersCount; i < count; i++ {
			workerId := fmt.Sprintf("%d", s.nextWorkerId)
			log.Infof("spinning up worker %s", workerId)
			s.workers = append(s.workers, NewServerWorker(s.latency, workerId, s.jobs))
			s.nextWorkerId++
		}
	} else if count < workersCount {
		removed, rest := s.workers[count:], s.workers[:count]
		s.workers = rest
		for _, worker := range removed {
			log.Infof("spinning down worker %s", worker.WorkerId)
			worker.stop()
		}
	} // else don't do anything if count == s.WorkersCount
}

// SetupHTTPServer .....
func SetupHTTPServer(handle func(string) (string, int)) {
	http.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		log.Infof("received request to %s", r.URL.String())
		if r.Method == "GET" {
			vals, ok := r.URL.Query()["jobid"]
			if !ok {
				log.Errorf("missing jobid parameter")
				http.Error(w, "missing jobid parameter", 400)
				return
			}
			resp, code := handle(vals[0])
			header := w.Header()
			header.Set(http.CanonicalHeaderKey("content-type"), "application/json")
			w.WriteHeader(code)
			fmt.Fprint(w, resp)
		} else {
			http.NotFound(w, r)
		}
	})
}
