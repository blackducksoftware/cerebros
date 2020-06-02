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
package main

import (
	"fmt"
	"github.com/blackducksoftware/hub-client-go/hubapi"
	"os"

	//"github.com/blackducksoftware/hub-client-go/hubapi"
	"github.com/blackducksoftware/hub-client-go/hubclient"
	log "github.com/sirupsen/logrus"
	"time"
)

func doOrDie(err error) {
	if err != nil {
		log.Fatalf("%+v", err)
	}
}

func main() {
	host, username, password := os.Args[1], os.Args[2], os.Args[3]
	port := 443
	httpTimeout := 120 * time.Second
	baseURL := fmt.Sprintf("https://%s:%d", host, port)
	rawClient, err := hubclient.NewWithSession(baseURL, hubclient.HubClientDebugTimings, httpTimeout)
	doOrDie(err)
	log.Infof("instantiated client")

	err = rawClient.Login(username, password)
	doOrDie(err)
	log.Infof("logged in")

	limit := 100000
	offset := 0
	clList, err := rawClient.ListAllCodeLocations(&hubapi.GetListOptions{
		Limit:  &limit,
		Offset: &offset,
	})
	doOrDie(err)
	//log.Infof("code locations list: %+v", clList)

	statusCounts := map[string]int{}

	for _, cl := range clList.Items {
		scanSummariesLink, err := cl.GetScanSummariesLink()
		doOrDie(err)

		log.Infof("scan summaries link for %s: %+v", cl.Name, scanSummariesLink)

		if scanSummariesLink == nil {
			log.Infof("no scan summaries link for %s", cl.Name)
			continue
		}

		scanSummaries, err := rawClient.ListScanSummaries(*scanSummariesLink)
		doOrDie(err)

		for _, scanSummary := range scanSummaries.Items {
			//log.Infof("found scan summary: %+v", scanSummary)
			statusCounts[scanSummary.Status]++
		}
	}

	for status, count := range statusCounts {
		log.Infof("status %s: %d scan summaries", status, count)
	}
}
