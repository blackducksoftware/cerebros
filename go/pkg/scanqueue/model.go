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
	"github.com/pkg/errors"
	"net/http"

	"github.com/blackducksoftware/cerebros/go/pkg/util"
	log "github.com/sirupsen/logrus"
)

const (
	actionChannelSize = 100
)

// Model ...
type Model struct {
	ScanQueue *util.PriorityQueue
	actions   chan *action
}

// NewModel .....
func NewModel() *Model {
	model := &Model{
		ScanQueue: util.NewPriorityQueue(),
		actions:   make(chan *action, actionChannelSize),
	}
	go func() {
		//stop := time.Now()
		for {
			select {
			case nextAction := <-model.actions:
				actionName := nextAction.name
				log.Debugf("processing model action of type %s", actionName)

				// metrics: how many messages are waiting?
				//recordNumberOfMessagesInQueue(len(model.actions))

				// metrics: log message type
				//recordMessageType(actionName)

				// metrics: how long idling since the last action finished processing?
				//start := time.Now()
				//recordReducerActivity(false, start.Sub(stop))

				// actually do the work
				err := nextAction.apply()
				if err != nil {
					log.Errorf("problem processing action %s: %v", actionName, err)
					//recordActionError(actionName)
				}

				// metrics: how long did the work take?
				//stop = time.Now()
				//recordReducerActivity(true, stop.Sub(start))
			}
		}
	}()
	return model
}

// Private API

func (model *Model) addJob(key string, data interface{}) error {
	return model.ScanQueue.Add(key, 0, data)
}

func (model *Model) getNextJob() (string, interface{}) {
	key, job, err := model.ScanQueue.Pop()
	if err != nil {
		log.Errorf("unable to get next job: %s", err)
		return "", nil
	}
	return key, job
}

func (model *Model) finishJob(key string, err string) error {
	// don't do anything for now
	// TODO do something later
	return errors.New("finishJob is not implemented")
}

// HTTP responder implementation -- Public API

func (model *Model) AddJob(job ApiJob) error {
	key := job.Key
	data := job.Data
	done := make(chan error)
	model.actions <- &action{"addJob", func() error {
		log.Debugf("adding job: key %s, data %+v", key, data)
		error := model.addJob(key, data)
		go func() {
			done <- error
		}()
		return error
	}}
	return <-done
}

// GetNextJob returns nil if no job was found
func (model *Model) GetNextJob() (ApiJob, error) {
	done := make(chan struct{})
	var apiJob ApiJob
	var err error
	model.actions <- &action{"getNextJob", func() error {
		log.Debugf("looking for next job")
		key, job := model.getNextJob()
		apiJob = ApiJob{
			Key:  key,
			Data: job,
		}
		close(done)
		return nil
	}}
	<-done
	return apiJob, err
}

// PostFinishJob ...
func (model *Model) PostFinishJob(jobResult ApiJobResult) error {
	log.Infof("finish job: %+v", jobResult)
	done := make(chan error)
	model.actions <- &action{"finishJob", func() error {
		err := model.finishJob(jobResult.Key, jobResult.Err)
		go func() {
			done <- err
		}()
		return err
	}}
	return <-done
}

// GetModel ...
func (model *Model) GetModel() ([]byte, error) {
	done := make(chan struct{})
	var modelJson []byte
	var err error
	model.actions <- &action{"getModel", func() error {
		modelJson, err = json.MarshalIndent(model, "", "  ")
		close(done)
		return err
	}}
	<-done
	return modelJson, err
}

// NotFound .....
func (model *Model) NotFound(w http.ResponseWriter, r *http.Request) {
	log.Errorf("HTTPResponder not found from request %+v", r)
	//recordHTTPNotFound(r)
	http.NotFound(w, r)
}

// Error .....
func (model *Model) Error(w http.ResponseWriter, r *http.Request, err error, statusCode int) {
	log.Errorf("HTTPResponder error %s with code %d from request %+v", err.Error(), statusCode, r)
	//recordHTTPError(r, err, statusCode)
	http.Error(w, err.Error(), statusCode)
}
