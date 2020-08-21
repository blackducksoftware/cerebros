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
	"github.com/pkg/errors"
)

type ProjectRelationshipsLinks struct {
	Self    string
	Related string
}

type ProjectRelationshipsData struct {
	Type string
	Id   string
}

type ProjectRelationshipsJustLinks struct {
	Links ProjectRelationshipsLinks
}

type ProjectRelationshipsLinksAndData struct {
	Links ProjectRelationshipsLinks
	Data  *ProjectRelationshipsData
}

type GetProjectsResponse struct {
	Data []struct {
		Type       string
		Id         string
		Attributes struct {
			// TODO properties
			Name string
		}
		Relationships struct {
			Branches          *ProjectRelationshipsJustLinks
			Runs              *ProjectRelationshipsJustLinks
			ProjectPreference *ProjectRelationshipsLinksAndData `json:"project-preference"`
			UserDefaultBranch *ProjectRelationshipsLinksAndData `json:"user-default-branch"`
		}
		Links *struct {
			Self *struct {
				HRef string
				Meta *struct {
					Durable string
				}
			}
		}
		Meta struct {
			Etag           string
			OrganizationId string `json:"organization-id"`
			InTrash        bool   `json:"in-trash"`
		}
	}
	// TODO included, links
	Meta struct {
		// Offset int
		Limit int
		Total int
	}
}

func (client *Client) GetProjects(limit int) (*GetProjectsResponse, error) {
	result := &GetProjectsResponse{}
	_, err := client.GetJson(map[string]interface{}{"page[limit]": fmt.Sprintf("%d", limit)}, result, "api/common/v0/projects")
	return result, err
}

type GetToolsResponse struct {
	Data []struct {
		Type       string
		Id         string
		Attributes struct {
			Name    string
			Version string
		}
		// TODO relationships
		// TODO links
		// TODO meta
	}
	// TODO included
	Meta struct {
		//Offset int
		Limit int
		Total int
	}
	// TODO links
}

func (client *Client) GetTools(limit int) (*GetToolsResponse, error) {
	if limit <= 0 || limit > 500 {
		return nil, errors.New(fmt.Sprintf("limit must be between 1 and 500, got %d", limit))
	}
	result := &GetToolsResponse{}
	params := map[string]interface{}{"page[limit]": fmt.Sprintf("%d", limit)}
	_, err := client.GetJson(params, result, "api/common/v0/tools")
	return result, err
}

// PostTools handles
//   https://sig-gitlab.internal.synopsys.com/clops/polaris-local/-/blob/master/README.md#temporary-workaround-if-needed-1
func (client *Client) PostTools() (string, error) {
	params := map[string]interface{}{
		"data": map[string]interface{}{
			"type": "tool",
			"attributes": map[string]string{
				"version": "2020.06",
				"name":    "Coverity",
			},
		},
	}
	return client.PostJson(params, nil, "api/common/v0/tools")
}

type GetV0BranchResponse struct {
	Data struct {
		Type       string
		Id         string
		Attributes struct {
			Name           string
			MainForProject bool `json:"main-for-project"`
		}
		// TODO Relationships Links Meta
	}
	// TODO Included
	// TODO Meta
	// TODO Links
}

func (client *Client) GetV0Branch(branchId string) (*GetV0BranchResponse, error) {
	result := &GetV0BranchResponse{}
	params := map[string]interface{}{}
	_, err := client.GetJson(params, result, "api/common/v0/branches/%s", branchId)
	return result, err
}

type GetV0RevisionsResponse struct {
	Data []struct {
		Type       string
		Id         string
		Attributes struct {
			Name                 string
			Timestamp            string
			ModifiedOffTheRecord bool `json:"modified-off-the-record"`
		}
		// TODO Relationships, Links, Meta
	}
	// TODO Included, Links
	Meta struct {
		Offset int
		Limit  int
		Total  int
	}
}

func (client *Client) GetV0Revisions(limit int) (*GetV0RevisionsResponse, error) {
	result := &GetV0RevisionsResponse{}
	params := map[string]interface{}{
		"page[limit]": fmt.Sprintf("%d", limit),
	}
	_, err := client.GetJson(params, result, "api/common/v0/revisions")
	return result, err
}

func (client *Client) GetV0RevisionsByBranch(branchId string, limit int) (*GetV0RevisionsResponse, error) {
	result := &GetV0RevisionsResponse{}
	params := map[string]interface{}{
		"filter[revision][branch][id][$eq]": branchId,
		"page[limit]":                       fmt.Sprintf("%d", limit),
	}
	_, err := client.GetJson(params, result, "api/common/v0/revisions")
	return result, err
}
