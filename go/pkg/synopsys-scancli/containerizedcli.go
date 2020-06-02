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
package synopsys_scancli

import (
	"github.com/blackducksoftware/cerebros/go/pkg/scanqueue"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"time"
)

type ContainerizedCLI struct {
	stop    <-chan struct{}
	client  *scanqueue.Client
	scanner *Scanner
}

func NewContainerizedCLI(scanner *Scanner, client *scanqueue.Client, stop <-chan struct{}) *ContainerizedCLI {
	cc := &ContainerizedCLI{stop: stop, client: client, scanner: scanner}
	cc.start()
	return cc
}

func (cc *ContainerizedCLI) start() {
	log.Infof("starting job-poll goroutine")
	go func() {
		for {
			select {
			case <-cc.stop:
				break
			case <-time.After(10 * time.Second):
				err := cc.checkForAndRunScan()
				recordEvent("check_for_and_run_scan", err)
				if err != nil {
					log.Errorf("unable to checkForAndRunScan: %s", err)
				}
			}
		}
	}()
}

func (cc *ContainerizedCLI) checkForAndRunScan() error {
	log.Infof("getting next job")

	config := &ScanConfig{}
	err := cc.client.GetNextJob(config)
	if err != nil {
		return errors.WithMessagef(err, "unable to get next job")
	}
	if config.Key == "" && config.CodeLocation == nil && config.ScanType == nil {
		log.Infof("got nil scan job, so nothing to do")
		return nil
	}

	log.Infof("got job: %+v", config)
	return cc.scanner.Scan(config)
}
