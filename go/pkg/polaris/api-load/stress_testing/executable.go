package stress_testing

import (
	"fmt"
	"github.com/blackducksoftware/cerebros/go/pkg/polaris/api"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	"net/http"
	"time"
)

func doOrDie(err error) {
	if err != nil {
		log.Fatalf("%+v", err)
	}
}

func Run(configPath string) {
	config, err := GetConfig(configPath)
	doOrDie(err)

	logLevel, err := config.GetLogLevel()
	doOrDie(err)
	log.SetLevel(logLevel)

	islg, auth, err := RunLoadGenerator(config)
	doOrDie(err)

	prometheus.Unregister(prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{}))
	prometheus.Unregister(prometheus.NewGoCollector())

	http.Handle("/metrics", promhttp.Handler())

	//gatewayUrl:="http://polaris-monitoring-prometheus-pushgateway.monitoring:9091"
	//push.New(gatewayUrl, "cerebros").Collector(prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{})).Add()
	addr := fmt.Sprintf(":%d", config.Port)
	log.Infof("successfully instantiated issue %+v and auth %+v, serving on %s", islg, auth, addr)
	go func() {
		http.ListenAndServe(addr, nil)
	}()

	stop := make(chan struct{})
	<-stop

	<-time.After(time.Duration(config.MinutesToRun) * time.Minute)
	log.Infof("stopping auth and issue server load generators...")
	islg.Stop()
	auth.Stop()
	log.Infof("successfully stopped both the auth and issue server load generators")
}

func RunLoadGenerator(config *Config) (*IssueServerLoadGenerator, *AuthLoadGenerator, error) {
	apiClient := api.NewClient(config.PolarisURL, config.PolarisEmail, config.PolarisPassword)

	if err := apiClient.Authenticate(); err != nil {
		return nil, nil, err
	}

	pf := NewProjectFetcherWithRandomStart(apiClient, config.LoadGenerator.Issue.FetchProjectsCount)
	pf.Start()

	if err := RunLoginsForUsers(apiClient, config.LoadGenerator.Auth.PreRunLogins); err != nil {
		return nil, nil, err
	}

	islg := NewIssueServerLoadGenerator(apiClient, pf, config.LoadGenerator.Issue)
	auth := NewAuthLoadGenerator(pf, config.PolarisURL, config.PolarisEmail, config.PolarisPassword, config.LoadGenerator.Auth)

	return islg, auth, nil
}
