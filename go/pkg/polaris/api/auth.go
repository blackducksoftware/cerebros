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
	log "github.com/sirupsen/logrus"
	"time"
)

func (client *Client) Authenticate() error {
	path := "api/auth/authenticate"
	url := fmt.Sprintf("%s/%s", client.URL, path)
	log.Debugf("issuing POST request to %s", url)
	bodyParams := map[string]string{
		"email":    client.Email,
		"password": client.Password,
	}
	request := client.RestyClient.R().
		SetHeader("Content-Type", "application/x-www-form-urlencoded").
		SetHeader("Accept", "application/json").
		SetFormData(bodyParams)

	client.authMux.Lock()
	defer client.authMux.Unlock()
	start := time.Now()
	resp, err := request.Post(url)
	duration := time.Now().Sub(start)

	recordEvent("POST_"+path, err)

	if err != nil {
		return errors.Wrapf(err, "unable to POST to %s", url)
	}

	statusCode := resp.StatusCode()
	recordResponseTime("POST", path, duration, statusCode)
	recordResponseStatusCode("POST", path, statusCode)

	if statusCode < 200 || statusCode > 299 {
		return errors.New(fmt.Sprintf("bad response code: %d, %s", statusCode, resp.String()))
	}

	log.Tracef("resp from %s: %+v, %s", url, resp, resp.String())
	var accessToken string
	for _, cookie := range resp.Cookies() {
		log.Tracef("cookie: %s, \n%s\n\n", cookie.Name, cookie.Value)
		if cookie.Name == "access_token" {
			accessToken = cookie.Value
		}
	}
	if accessToken == "" {
		return errors.New(fmt.Sprintf("got status code %d, but did not find cookie access_token in response", statusCode))
	}
	client.AuthToken = accessToken
	return nil
}

type CreateAccessTokenResponse struct {
	Data struct {
		Attributes struct {
			AccessToken string `json:"access-token"`
		} `json:"attributes"`
	} `json:"data"`
}

func (client *Client) GetAccessToken(tokenName string) (*CreateAccessTokenResponse, error) {
	bodyParams := map[string]interface{}{
		"data": map[string]interface{}{
			"attributes": map[string]interface{}{
				"access-token": nil,
				"date-created": nil,
				"name":         tokenName,
				"revoked":      false,
			},
			"type": "apitokens",
		},
	}
	result := &CreateAccessTokenResponse{}
	_, err := client.PostJson(bodyParams, result, "api/auth/apitokens")
	return result, err
}

type GetEntitlementsForProjectResponse struct {
	Data []struct {
		Type       string
		Id         string
		Attributes struct {
			Allowed []string
		}
	}
	Meta struct {
		Limit  int
		Offset int
		Total  int
	}
}

func (client *Client) GetEntitlementsForProject(projectID string) (*GetEntitlementsForProjectResponse, error) {
	params := map[string]interface{}{
		"filter[entitlements][object][eq]": fmt.Sprintf("urn:x-swip:projects:%s", projectID),
	}
	result := &GetEntitlementsForProjectResponse{}
	_, err := client.GetJson(params, result, "api/auth/entitlements")
	return result, err
}

type GetEntitlementsForOrganizationResponse struct {
	Data []struct {
		Type       string
		Id         string
		Attributes struct {
			Allowed []string
		}
		// TODO relationships
		Links struct {
			Self struct {
				HRef string
			}
		}
	}
	// TODO included
}

func (client *Client) GetEntitlementsForOrganization(orgId string) (*GetEntitlementsForOrganizationResponse, error) {
	params := map[string]interface{}{
		"filter[entitlements][object][eq]": fmt.Sprintf("urn:x-swip:organizations:%s", orgId),
	}
	result := &GetEntitlementsForOrganizationResponse{}
	_, err := client.GetJson(params, result, "api/auth/entitlements")
	return result, err
}

type GetGroupsResponse struct {
	Data []struct {
		Type       string
		Id         string
		Attributes struct {
			GroupName   string
			DateCreated string `json:"date-created"`
		}
		// TODO relationships
	}
	Meta struct {
		Limit  int
		Offset int
		Total  int
	}
}

func (client *Client) GetGroups() (*GetGroupsResponse, error) {
	result := &GetGroupsResponse{}
	_, err := client.GetJson(map[string]interface{}{}, result, "api/auth/groups")
	return result, err
}

func (client *Client) GetRoleAssignmentsForProject(projectId string) (*GetRoleAssignmentsResponse, error) {
	result := &GetRoleAssignmentsResponse{}
	params := map[string]interface{}{
		"filter[role-assignments][object][$eq]": fmt.Sprintf("urn:x-swip:projects:%s", projectId),
		"include[role-assignments][]":           []string{"role", "user", "group"},
	}
	_, err := client.GetJson(params, result, "api/auth/role-assignments")
	return result, err
}

func (client *Client) GetRoleAssignmentsForUser(email string, offset int, limit int, isServiceAccount bool) (*GetRoleAssignmentsResponse, error) {
	result := &GetRoleAssignmentsResponse{}
	params := map[string]interface{}{
		"filter[role-assignments][user][email][$eq]": email,
		"filter[role-assignments][user][automated]":  fmt.Sprintf("%t", isServiceAccount),
		"include[users][]":                           "roleassignments",
		"page[limit]":                                fmt.Sprintf("%d", limit),
		"page[offset]":                               fmt.Sprintf("%d", offset),
	}
	_, err := client.GetJson(params, result, "api/auth/role-assignments")
	return result, err
}

type GetRoleAssignmentsResponse struct {
	Data []struct {
		Type       string
		Id         string
		Attributes struct {
			ExpiresBy string `json:"expires-by"`
			Object    string
		}
		Relationships map[string]struct {
			Links struct {
				Self    string
				Related string
			}
			Data struct {
				Type string
				Id   string
			}
		}
	}
	Meta struct {
		Limit  int
		Offset int
		Total  int
	}
}

func (client *Client) GetRoleAssignments(offset int, limit int) (*GetRoleAssignmentsResponse, error) {
	result := &GetRoleAssignmentsResponse{}
	params := map[string]interface{}{
		"page[limit]":  fmt.Sprintf("%d", limit),
		"page[offset]": fmt.Sprintf("%d", offset),
	}
	_, err := client.GetJson(params, result, "api/auth/role-assignments")
	return result, err
}

type GetRolesResponse struct {
	Data []struct {
		Type       string
		Id         string
		Attributes struct {
			RoleName    string
			Permissions struct {
				Organization []string
				Project      []string
			}
		}
	}
	Meta struct {
		Limit  int
		Offset int
		Total  int
	}
}

// TODO encode this important data in types:
// Roles: Observer, Administrator, Contributor
// Permissions:
//    "ORGANIZATION" : [ "administer", "users.read", "users.readPrivate", "users.write", "groups.read", "groups.write", "projects.create" ],
//    "PROJECT" : [ "administer", "projects.read", "projects.write" ]

func (client *Client) GetRoles() (*GetRolesResponse, error) {
	result := &GetRolesResponse{}
	_, err := client.GetJson(map[string]interface{}{}, result, "api/auth/roles")
	return result, err
}

func (client *Client) CreateRoleAssignment(userId string, roleId string, projectId string, orgId string) (string, error) {
	objectId := fmt.Sprintf("urn:x-swip:projects:%s", projectId)
	bodyParams := map[string]interface{}{
		"data": map[string]interface{}{
			"attributes": map[string]interface{}{
				//"expires-by": "2029-01-01T14:14:14.141Z",
				"object": objectId,
			},
			"relationships": map[string]interface{}{
				//"group": map[string]interface{}{
				//	"data": map[string]interface{}{
				//		"type": "groups",
				//		"id": groupId,
				//	},
				//},
				"organization": map[string]interface{}{
					"data": map[string]interface{}{
						"type": "organizations",
						"id":   orgId,
					},
				},
				"role": map[string]interface{}{
					"data": map[string]interface{}{
						"type": "roles",
						"id":   roleId,
					},
				},
				"user": map[string]interface{}{
					"data": map[string]interface{}{
						"type": "users",
						"id":   userId,
					},
				},
			},
			"type": "role-assignments",
		},
	}
	return client.PostJson(bodyParams, nil, "api/auth/role-assignments")
}

type GetUsersResponse struct {
	Data []struct {
		Type       string
		Id         string
		Attributes struct {
			Owner    bool
			Name     string
			Email    string
			Username string
		}
		// TODO relationships
	}
	Meta struct {
		Limit  int
		Offset int
		Total  int
	}
}

func (client *Client) GetUsers(offset int, limit int) (*GetUsersResponse, error) {
	result := &GetUsersResponse{}
	params := map[string]interface{}{
		"page[limit]":  fmt.Sprintf("%d", limit),
		"page[offset]": fmt.Sprintf("%d", offset),
	}
	_, err := client.GetJson(params, result, "api/auth/users")
	return result, err
}

func (client *Client) GetUserByEmail(email string) (*GetUsersResponse, error) {
	result := &GetUsersResponse{}
	params := map[string]interface{}{
		"filter[users][email][$eq]": email,
		"page[limit]":               "100",
		"page[offset]":              "0",
	}
	_, err := client.GetJson(params, result, "api/auth/users")
	return result, err
}

type CreateUserResponse struct {
	Data struct {
		Type       string
		Id         string
		Attributes struct {
			Name     string
			Email    string
			Username string
		}
	}
}

func (client *Client) CreateUser(email string, name string, orgId string) (*CreateUserResponse, error) {
	bodyParams := map[string]interface{}{
		"data": map[string]interface{}{
			"attributes": map[string]interface{}{
				"email":           email,
				"first-time":      nil,
				"enabled":         true,
				"owner":           false,
				"agreed-to-terms": false,
				"system":          false,
				"name":            name,
				"username":        fmt.Sprintf("%s-username", name),
			},
			"relationships": map[string]interface{}{
				"organization": map[string]interface{}{
					"data": map[string]interface{}{
						"type": "organizations",
						"id":   orgId,
					},
				},
			},
			"type": "users",
		},
	}
	result := &CreateUserResponse{}
	_, err := client.PostJson(bodyParams, result, "api/auth/users")
	return result, err
}

func (client *Client) CreateServiceAccount(email string, name string, orgId string, password string) (*CreateUserResponse, error) {
	bodyParams := map[string]interface{}{
		"data": map[string]interface{}{
			"attributes": map[string]interface{}{
				"email": email,
				"password-login": map[string]string{
					"password": password,
				},
				"first-time":      nil,
				"enabled":         true,
				"automated":       true,
				"owner":           false,
				"agreed-to-terms": false,
				"system":          false,
				"ip-whitelist": []string{
					"0.0.0.0/0",
				},
				"organization-admin": false,
				"name":               name,
				"username":           fmt.Sprintf("%s-username", name),
			},
			"relationships": map[string]interface{}{
				"organization": map[string]interface{}{
					"data": map[string]interface{}{
						"type": "organizations",
						"id":   orgId,
					},
				},
			},
			"type": "users",
		},
	}
	result := &CreateUserResponse{}
	_, err := client.PostJson(bodyParams, result, "api/auth/users")
	return result, err
}

type GetOrganizationsResponse struct {
	Data []struct {
		Type       string
		Id         string
		Attributes struct {
			OrganizationName string
			Description      string
		}
	}
	Meta struct {
		Limit  int
		Offset int
		Total  int
	}
}

func (client *Client) GetOrganizations() (*GetOrganizationsResponse, error) {
	result := &GetOrganizationsResponse{}
	_, err := client.GetJson(map[string]interface{}{}, result, "api/auth/organizations")
	return result, err
}
