# cerebros


## Go

```
cd go
```

### Project structure

`cmd/`: various executables, these shouldn't contain much code -- they're basically just entrypoints.  Each directory should have a `.go` file as the entrypoint, a `conf.json` as an example configuration file, and a `Dockerfile` to build an image.

`pkg/`: where the vast majority of the code is!  Functions, classes, etc. goes here.

`hack/deploy*`: scripts for deploying groups of containers on Kubernetes, including:
 - Blackduck image scanning
 - Polaris scanning
 - Polaris load generation, data seeding, scans, and prometheus metrics

`Makefile`: helps you build and test everything.

### Build an image

```
make polaris-cli
```

### Test

```
make vet
make test
```

### Format

```
make fmt
```

### Metrics

Add prometheus metrics to everything!  It's very easy, check out [these metrics](https://github.com/blackducksoftware/cerebros/blob/master/go/pkg/kube/metrics.go) for an example:
 - figure out how to model your metrics -- do you want gauges, counters, or histograms?
 - instantiate your metrics objects *paying attention to labels* -- labels provided at instantiation have to match labels used at collection, otherwise prometheus will get upset
 - [write a unit test to verify that the metrics instantiate and can be used successfully](https://github.com/blackducksoftware/cerebros/blob/master/go/pkg/kube/metrics_test.go)
 - set up an http server and serve the prometheus metrics -- this allows the prometheus master to scrape your metrics
    ```
    // Prometheus and http setup
    prometheus.Unregister(prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{}))
    prometheus.Unregister(prometheus.NewGoCollector())

    http.Handle("/metrics", promhttp.Handler())

    addr := fmt.Sprintf(":%d", config.Port)
    go func() {
        log.Infof("starting HTTP server on port %d", config.Port)
        http.ListenAndServe(addr, nil)
    }()
    ```
 
To collect metrics from your running container, you'll need to:
 - [run a prometheus master](https://github.com/blackducksoftware/cerebros/blob/474307bf0a2108a060cca93ed34a5a44288155aa/go/hack/deploy-polaris-load-generator/deploy.sh#L8)
 - [add metadata annotations to your pods](https://github.com/blackducksoftware/cerebros/blob/474307bf0a2108a060cca93ed34a5a44288155aa/go/hack/deploy-polaris-load-generator/polaris-cli.yaml#L16-L19)

## Python

TODO
