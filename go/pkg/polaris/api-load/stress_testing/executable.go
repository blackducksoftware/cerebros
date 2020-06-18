package stress_testing

import (
	"fmt"
	"github.com/blackducksoftware/cerebros/go/pkg/polaris/api"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	"net/http"
)

func doOrDie(err error) {
	if err != nil {
		log.Fatalf("%+v", err)
	}
}

func RunIssueServerLoadGenerator(configPath string) {
	config, err := GetConfig(configPath)
	doOrDie(err)

	logLevel, err := config.GetLogLevel()
	doOrDie(err)
	log.SetLevel(logLevel)

	apiClient := api.NewClient(config.PolarisURL, config.PolarisEmail, config.PolarisPassword)

	doOrDie(apiClient.Authenticate())

	pf := NewProjectFetcherWithRandomStart(apiClient, config.LoadGenerator.Issue.FetchProjectsCount)
	pf.Start()

	err = RunLoginsForUsers(apiClient, config.LoadGenerator.Auth.PreRunLogins)
	doOrDie(err)

	islg := NewIssueServerLoadGenerator(apiClient, pf, config.LoadGenerator.Issue)
	auth := NewAuthLoadGenerator(pf, config.PolarisURL, config.PolarisEmail, config.PolarisPassword, config.LoadGenerator.Auth)

	prometheus.Unregister(prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{}))
	prometheus.Unregister(prometheus.NewGoCollector())

	http.Handle("/metrics", promhttp.Handler())

	addr := fmt.Sprintf(":%d", config.Port)
	log.Infof("successfully instantiated issue %+v and auth %+v, serving on %s", islg, auth, addr)
	go func() {
		http.ListenAndServe(addr, nil)
	}()

	stop := make(chan struct{})
	<-stop
}
