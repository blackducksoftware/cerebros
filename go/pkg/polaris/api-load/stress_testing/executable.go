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

	addr := fmt.Sprintf(":%d", config.Port)
	log.Infof("successfully instantiated issue %+v and auth %+v, serving on %s", islg, auth, addr)
	go func() {
		http.ListenAndServe(addr, nil)
	}()

	stop := make(chan struct{})
	<-stop
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
