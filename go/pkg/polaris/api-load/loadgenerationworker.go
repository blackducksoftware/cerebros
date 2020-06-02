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
package api_load

import (
	log "github.com/sirupsen/logrus"
	"time"
)

type LoadGenerationWorker struct {
	Id        string
	EventName string
	Func      func() error
	Stop      <-chan struct{}
}

func NewLoadGenerationWorker(id string, eventName string, f func() error, stop <-chan struct{}) *LoadGenerationWorker {
	return &LoadGenerationWorker{Id: id, EventName: eventName, Func: f, Stop: stop}
}

func (lgw *LoadGenerationWorker) Start() {
	go func() {
		log.Infof("starting LoadGenerationWorker goroutine %s:%s", lgw.Id, lgw.EventName)
		for {
			select {
			case <-lgw.Stop:
				break
			default:
			}
			log.Debugf("loadGen goroutine %s:%s issuing request", lgw.Id, lgw.EventName)
			start := time.Now()
			err := lgw.Func()
			duration := time.Now().Sub(start)
			recordEvent(lgw.EventName, err)
			if err != nil {
				log.Errorf("unable to %s: %+v", lgw.EventName, err)
			} else {
				log.Debugf("successfully %s in %d ms", lgw.EventName, duration.Milliseconds())
			}
		}
		log.Infof("exiting LoadGenerationWorker goroutine %s:%s", lgw.Id, lgw.EventName)
	}()
}
