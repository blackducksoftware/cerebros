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

func (model *Model) getNextJob() interface{} {
	job, err := model.ScanQueue.Pop()
	if err != nil {
		log.Errorf("unable to get next job: %s", err)
		return nil
	}
	return job
}

func (model *Model) finishJob(key string, data interface{}, err error) error {
	// don't do anything for now
	// TODO do something later
	return errors.New("finishJob is not implemented")
}

// Public API

func (model *Model) AddJob(key string, data interface{}) error {
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
func (model *Model) GetNextJob() interface{} {
	done := make(chan interface{})
	model.actions <- &action{"getNextJob", func() error {
		log.Debugf("looking for next job")
		job := model.getNextJob()
		go func() {
			done <- job
		}()
		return nil
	}}
	return <-done
}

// FinishScanJob should be called when the scan client has finished.
func (model *Model) FinishJob(key string, data interface{}, err error) {
	log.Infof("finish job: %+v, %v", key, err)
	model.actions <- &action{"finishJob", func() error {
		return model.finishJob(key, data, err)
	}}
}

// GetModel ...
func (model *Model) GetModel() *[]byte {
	done := make(chan *[]byte)
	model.actions <- &action{"getModel", func() error {
		modelJson, err := json.Marshal(model)
		if err != nil {
			go func() {
				done <- nil
			}()
			return err
		}
		go func() {
			done <- &modelJson
		}()
		return nil
	}}
	return <-done
}
