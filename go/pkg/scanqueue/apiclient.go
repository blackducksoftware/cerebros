/*
Copyright (C) 2018 Synopsys, Inc.

Licensed to the Apache Software Foundation (ASF) under one
or more contributor license agreements. See the NOTICE file
distributed with this work for additional information
regarding copyright ownership. The ASF licenses this file
to you under the Apache License, Version 2.0 (the
"License"); you may not use this file except in compliance
with the License. You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing,
software distributed under the License is distributed on an
"AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
KIND, either express or implied. See the License for the
specific language governing permissions and limitations
under the License.
*/

package scanqueue

import (
	"fmt"
	"time"

	resty "github.com/go-resty/resty/v2"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

const (
	addJobPath      = "job"
	nextJobPath     = "nextjob"
	finishedJobPath = "finishedjob"
	modelPath       = "model"
)

// ApiClientInterface ...
type ApiClientInterface interface {
	GetNextJob() error
	PostFinishedScan(key string) error
}

// ApiJob ...
type ApiJob struct {
	Key  string
	Data interface{}
}

// ApiJobResult ...
type ApiJobResult struct {
	Key string
	Err string
}

// ApiClient ...
type ApiClient struct {
	Resty *resty.Client
	Host  string
	Port  int
}

// NewApiClient ...
func NewApiClient(host string, port int) *ApiClient {
	restyClient := resty.New()
	restyClient.SetRetryCount(3)
	restyClient.SetRetryWaitTime(500 * time.Millisecond)
	restyClient.SetTimeout(time.Duration(5 * time.Second))
	return &ApiClient{
		Resty: restyClient,
		Host:  host,
		Port:  port,
	}
}

func (ac *ApiClient) url(path string) string {
	return fmt.Sprintf("http://%s:%d/%s", ac.Host, ac.Port, path)
}

// AddJob ...
func (ac *ApiClient) AddJob(key string, data interface{}) error {
	url := ac.url(addJobPath)
	log.Debugf("about to issue post request to url %s", url)
	job := &ApiJob{Key: key, Data: data}
	resp, err := ac.Resty.R().SetBody(job).Post(url)
	log.Debugf("received resp %+v, status code %d, error %+v from url %s", resp, resp.StatusCode(), err, url)
	//recordHTTPStats(addJobPath, resp.StatusCode())
	if err != nil {
		//recordScannerError("unable to add job")
		return errors.Wrapf(err, "unable to add job")
	} else if (resp.StatusCode() < 200) || (resp.StatusCode() >= 300) {
		//recordScannerError("unable to add job -- bad status code")
		return fmt.Errorf("unable to add job; body %s and status code %d", string(resp.Body()), resp.StatusCode())
	}
	return nil
}

// GetNextJob ...
func (ac *ApiClient) GetNextJob(res interface{}) error {
	url := ac.url(nextJobPath)
	log.Debugf("about to issue post request to url %s", url)
	resp, err := ac.Resty.R().
		SetHeader("Content-Type", "application/json").
		SetResult(&res).
		Post(url)
	log.Debugf("received resp %+v and error %+v from url %s", resp, err, url)
	//recordHTTPStats(nextImagePath, resp.StatusCode())
	if err != nil {
		//recordScannerError("unable to get next job")
		return errors.Wrapf(err, "unable to get next job")
	} else if (resp.StatusCode() < 200) || (resp.StatusCode() >= 300) {
		//recordScannerError("unable to get next job -- bad status code")
		return fmt.Errorf("unable to get next job; body %s and status code %d", string(resp.Body()), resp.StatusCode())
	}
	return nil
}

// GetModel ...
func (ac *ApiClient) GetModel() (string, error) {
	url := ac.url(modelPath)
	log.Debugf("about to issue post request to url %s", url)
	//var modelString string
	resp, err := ac.Resty.R().
		SetHeader("Content-Type", "application/json").
		//SetResult(&modelString).
		Get(url)
	log.Debugf("received resp %+v and error %+v from url %s", resp, err, url)
	//recordHTTPStats(nextImagePath, resp.StatusCode())
	if err != nil {
		//recordScannerError("unable to get next job")
		return "", errors.Wrapf(err, "unable to get next job")
	} else if (resp.StatusCode() < 200) || (resp.StatusCode() >= 300) {
		//recordScannerError("unable to get next job -- bad status code")
		return "", fmt.Errorf("unable to get next job; body %s and status code %d", string(resp.Body()), resp.StatusCode())
	}
	return string(resp.String()), nil
}

// PostFinishedJob ...
func (ac *ApiClient) PostFinishedJob(key string, errString string) error {
	url := ac.url(finishedJobPath)
	jobResult := &ApiJobResult{
		Key: key,
		Err: errString,
	}
	log.Debugf("about to issue post request %+v to url %s", jobResult, url)
	resp, err := ac.Resty.R().SetBody(jobResult).Post(url)
	log.Debugf("received resp %+v, status code %d, error %+v from url %s", resp, resp.StatusCode(), err, url)
	//recordHTTPStats(finishedScanPath, resp.StatusCode())
	if err != nil {
		//recordScannerError("unable to post finished scan")
		return errors.Wrapf(err, "unable to post finished scan")
	} else if (resp.StatusCode() < 200) || (resp.StatusCode() >= 300) {
		//recordScannerError("unable to post finished scan -- bad status code")
		return fmt.Errorf("unable to post finished scan; body %s and status code %d", string(resp.Body()), resp.StatusCode())
	}
	return nil
}
