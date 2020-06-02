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

import "fmt"

type GetJobsResponse struct {
	Data []struct {
		Type       string
		Id         string
		Attributes struct {
			DateFinished string
			// TODO lots more stuff
			Status struct {
				State    string
				Progress int
			}
		}
		// TODO relationships
	}
	Meta struct {
		Total   int
		Offset  int
		Limit   int
		Filters map[string]string
	}
}

func (client *Client) GetJobs(limit int) (*GetJobsResponse, error) {
	params := map[string]interface{}{"page[limit]": fmt.Sprintf("%d", limit)}
	result := &GetJobsResponse{}
	_, err := client.GetJson(params, result, "api/jobs/jobs")
	return result, err
}
