/*
Copyright (C) 2020 Synopsys, Inc.

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
package api

import (
	"fmt"
	"github.com/go-resty/resty/v2"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"sync"
	"time"
)

type Client struct {
	URL         string
	Email       string
	Password    string
	AuthToken   string
	RestyClient *resty.Client
	// authMux needs to be used whenever AuthToken is touched
	authMux *sync.RWMutex
}

func NewClient(url string, email string, password string) *Client {
	return &Client{
		URL:         url,
		Email:       email,
		Password:    password,
		AuthToken:   "",
		RestyClient: resty.New(),
		authMux:     &sync.RWMutex{},
	}
}

func (client *Client) getJsonWithHeader(acceptHeader string, params map[string]interface{}, result interface{}, pathTemplate string, pathArgs []interface{}) (string, error) {
	path := fmt.Sprintf(pathTemplate, pathArgs...)
	url := fmt.Sprintf("%s/%s", client.URL, path)
	log.Debugf("issuing GET request to %s, params %+v", url, params)

	client.authMux.RLock()
	defer client.authMux.RUnlock()
	request := client.RestyClient.R().
		SetHeader("Accept", acceptHeader).
		SetHeader("Authorization", fmt.Sprintf("Bearer %s", client.AuthToken))

	for k, v := range params {
		switch t := v.(type) {
		case string:
			request = request.SetQueryParam(k, t)
		case []string:
			for _, v := range t {
				request = request.SetQueryParam(k, v)
			}
		default:
			return "", errors.New(fmt.Sprintf("expected string or []string for params value, found %T", v))
		}
	}

	if result != nil {
		request = request.SetResult(result)
	}

	start := time.Now()
	resp, err := request.Get(url)
	duration := time.Now().Sub(start)

	recordEvent("GET_"+pathTemplate, err)

	if err != nil {
		return "", errors.Wrapf(err, "unable to GET %s", url)
	}

	body, code := resp.String(), resp.StatusCode()
	recordResponseTime("GET", pathTemplate, duration, code)
	recordResponseStatusCode("GET", pathTemplate, code)

	if code < 200 || code > 299 {
		return body, errors.New(fmt.Sprintf("bad status code to url GET %s: %d, response %s", url, code, body))
	}
	return body, nil
}

func (client *Client) GetJson(params map[string]interface{}, result interface{}, pathTemplate string, pathArgs ...interface{}) (string, error) {
	return client.getJsonWithHeader("application/vnd.api+json", params, result, pathTemplate, pathArgs)
}

func (client *Client) GetRawJson(params map[string]interface{}, result interface{}, pathTemplate string, pathArgs ...interface{}) (string, error) {
	return client.getJsonWithHeader("application/json", params, result, pathTemplate, pathArgs)
}

func (client *Client) PostJson(bodyParams map[string]interface{}, result interface{}, pathTemplate string, pathArgs ...interface{}) (string, error) {
	path := fmt.Sprintf(pathTemplate, pathArgs...)
	url := fmt.Sprintf("%s/%s", client.URL, path)

	log.Debugf("issuing POST request to %s", url)
	client.authMux.RLock()
	defer client.authMux.RUnlock()
	request := client.RestyClient.R().
		SetHeader("Content-Type", "application/vnd.api+json").
		SetHeader("Accept", "application/vnd.api+json").
		SetHeader("Authorization", fmt.Sprintf("Bearer %s", client.AuthToken)).
		SetBody(bodyParams)

	if result != nil {
		request = request.SetResult(result)
	}

	start := time.Now()
	resp, err := request.Post(url)
	duration := time.Now().Sub(start)

	recordEvent("POST_"+pathTemplate, err)

	if err != nil {
		return "", errors.Wrapf(err, "unable to POST to url %s", url)
	}

	body, statusCode := resp.String(), resp.StatusCode()
	recordResponseStatusCode("POST", pathTemplate, statusCode)
	recordResponseTime("POST", pathTemplate, duration, statusCode)

	if statusCode < 200 || statusCode > 299 {
		return body, errors.New(fmt.Sprintf("bad response status code to POST %s: %d, body %s", url, resp.StatusCode(), body))
	}
	//log.Debugf("tokens response: %d", resp.StatusCode())
	return body, nil
}
