package api

import (
	"fmt"
	log "github.com/sirupsen/logrus"
)

type VinylV0Project struct {
	Type       string
	Id         string
	Attributes struct {
		Name string
	}
	Relationships map[string]struct {
		Links struct {
			Self    string
			Related string
		}
		Data *struct {
			Type string
			Id   string
		}
	}
}

type GetVinylV0ProjectsResponse struct {
	Data     []*VinylV0Project
	Included []struct {
		// TODO Type="branch" has different stuff than Type="entitlements"
		Type          string
		Id            string
		Attributes    map[string]interface{}
		Relationships map[string]interface{}
		Links         struct {
			Self struct {
				HRef string
			}
		}
		Meta *struct {
			ETag    string
			InTrash bool `json:"in-trash"`
		}
	}
	Meta struct {
		Offset int
		Limit  int
		Total  int
	}
}

func (client *Client) GetVinylV0Projects(offset int, limit int) (*GetVinylV0ProjectsResponse, error) {
	result := &GetVinylV0ProjectsResponse{}
	params := map[string]interface{}{
		"include[project][]": []string{"entitlements", "main-branch", "project-preference", "user-default-branch"},
		"page[offset]":       []string{fmt.Sprintf("%d", offset)},
		"page[limit]":        []string{fmt.Sprintf("%d", limit)},
	}
	json, err := client.GetJson(params, result, "api/vinyl/common/v0/projects")
	log.Tracef("vinyl json:\n%s\n\n", json)
	return result, err
}

type GetVinylV0ProjectsRelationshipsRunsResponse struct {
	Data []struct {
		Type string
		Id   string
	}
}

func (client *Client) GetVinylV0ProjectsRelationshipsRuns(projectId string) (*GetVinylV0ProjectsRelationshipsRunsResponse, error) {
	result := &GetVinylV0ProjectsRelationshipsRunsResponse{}
	params := map[string]interface{}{}
	//	"include[project][]": []string{"entitlements", "main-branch", "project-preference", "user-default-branch"},
	//	"page[offset]":       []string{fmt.Sprintf("%d", offset)},
	//	"page[limit]":        []string{fmt.Sprintf("%d", limit)},
	//}
	json, err := client.GetJson(params, result, "api/vinyl/common/v0/projects/%s/relationships/runs", projectId)
	log.Tracef("vinyl json:\n%s\n\n", json)
	return result, err
}

type GetVinylV0ProjectsRelatedRunsResponse struct {
	Data []struct {
		Type       string
		Id         string
		Attributes struct {
			CompletedDate string `json:"completed-date"`
			Segment       bool
			Status        string
			UploadId      string `json:"upload-id"`
			RunType       string `json:"run-type"`
			CreationDate  string `json:"creation-date"`
			Fingerprints  []interface{}
		}
	}
}

func (client *Client) GetVinylV0ProjectsRelatedRuns(projectId string) (*GetVinylV0ProjectsRelatedRunsResponse, error) {
	result := &GetVinylV0ProjectsRelatedRunsResponse{}
	params := map[string]interface{}{}
	json, err := client.GetJson(params, result, "api/vinyl/common/v0/projects/%s/related/runs", projectId)
	log.Tracef("vinyl json:\n%s\n\n", json)
	return result, err
}
