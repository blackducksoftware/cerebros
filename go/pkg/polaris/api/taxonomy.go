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

type GetTaxonomiesResponse struct {
	Data []struct {
		TaxonomyType   string `json:"taxonomy-type"`
		Id             string
		OptimisticLock string `json:"optimistic-lock"`
		Taxonomy       struct {
			Taxa []struct {
				Id   string
				Name struct {
					En string
				}
				// TODO lots more stuff
			}
			Name struct {
				En string
			}
			Description struct {
				En string
			}
			Abbreviation struct {
				En string
			}
			// TODO extra
			// TODO depends-on
			RootTaxa []string `json:"root-taxa"`
		}
	}
	Meta struct {
		Total  int
		Offset int
		Limit  int
	}
}

func (client *Client) GetTaxonomies(tenantId *string, pageLimit int) (*GetTaxonomiesResponse, error) {
	result := &GetTaxonomiesResponse{}
	params := map[string]interface{}{
		"page[limit]": fmt.Sprintf("%d", pageLimit),
	}
	if tenantId != nil {
		params["tenant-id"] = *tenantId
	}
	_, err := client.GetRawJson(params, result, "api/taxonomy/v0/taxonomies")
	return result, err
}

func (client *Client) GetTaxonomyCount() (*GetTaxonomiesResponse, error) {
	return client.GetTaxonomies(nil, 0)
}
