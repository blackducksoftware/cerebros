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
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	log "github.com/sirupsen/logrus"
)

// SetupHTTPServer .....
func SetupHTTPServer(responder Responder) {
	// state of the program
	http.HandleFunc("/model", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			modelJson, err := responder.GetModel()
			if err != nil {
				responder.Error(w, r, err, 500)
				return
			}
			header := w.Header()
			header.Set(http.CanonicalHeaderKey("content-type"), "application/json")
			fmt.Fprint(w, string(modelJson))
		} else {
			responder.NotFound(w, r)
		}
	})

	http.HandleFunc("/job", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "POST":
			body, err := ioutil.ReadAll(r.Body)
			if err != nil {
				log.Errorf("unable to read body for job POST: %s", err.Error())
				responder.Error(w, r, err, 400)
				return
			}
			var job Job
			err = json.Unmarshal(body, &job)
			if err != nil {
				log.Errorf("unable to ummarshal JSON for job POST: %s", err.Error())
				responder.Error(w, r, err, 400)
				return
			}
			err = responder.AddJob(job)
			if err != nil {
				log.Errorf("unable to add job: %s", err)
				responder.Error(w, r, err, 500)
				return
			}
			fmt.Fprint(w, "")
		default:
			responder.NotFound(w, r)
		}
	})

	http.HandleFunc("/nextjob", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			job, err := responder.GetNextJob()
			if err != nil {
				log.Errorf("unable to get next job: %s", err)
				responder.Error(w, r, err, 500)
			}
			jsonBytes, err := json.MarshalIndent(job, "", "  ")
			if err != nil {
				responder.Error(w, r, err, 500)
			} else {
				header := w.Header()
				header.Set(http.CanonicalHeaderKey("content-type"), "application/json")
				fmt.Fprint(w, string(jsonBytes))
			}
		} else {
			responder.NotFound(w, r)
		}
	})

	http.HandleFunc("/finishedjob", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			body, err := ioutil.ReadAll(r.Body)
			if err != nil {
				responder.Error(w, r, err, 400)
				return
			}
			var jobResult JobResult
			err = json.Unmarshal(body, &jobResult)
			if err != nil {
				responder.Error(w, r, err, 400)
				return
			}
			responder.PostFinishJob(jobResult)
			fmt.Fprint(w, "")
		} else {
			responder.NotFound(w, r)
		}
	})
}
