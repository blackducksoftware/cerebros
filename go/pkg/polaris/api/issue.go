package api

import (
	"fmt"
	"github.com/pkg/errors"
)

type IdAndType struct {
	Id   string
	Type string
}

type Data struct {
	Data IdAndType
}

type V1IssueResponse struct {
	Attributes struct {
		FindingKey string `json:"finding-key"`
		IssueKey   string `json:"issue-key"`
		SubTool    string `json:"sub-tool"`
		//Severity type?
	}
	Id            string
	Type          string
	Relationships struct {
		Path                Data
		ToolDomainService   Data `json:"tool-domain-service-data"`
		IssueType           Data `json:"issue-type"`
		Tool                Data
		LatestObservedOnRun Data `json:"latest-observed-on-run"`
		Transitions         struct {
			Data []IdAndType
		}
		RelatedTaxa struct {
			Data []interface{} // TODO type ?
		} `json:"related-taxa"`
		RelatedIndicators struct {
			Data []interface{} // TODO type ?
		} `json:"related-indicators"`
		IssueKind interface{} `json:"issue-kind"` // TODO type?
		Severity  interface{} // TODO type?
	}
	Links struct {
		Self struct {
			HRef string
		}
	}
}

type V1IssueIncluded struct {
	Id         string
	Type       string
	Attributes map[string]interface{}
}

type GetV1IssuesResponse struct {
	Data     []V1IssueResponse
	Included []V1IssueIncluded
	Meta     struct {
		Total    int
		Offset   int
		Limit    int
		Complete bool
		RunCount int `json:"run-count"`
	}
}

func (client *Client) GetV1Issues(projectId string, branchId string, runId string, offset int, limit int) (*GetV1IssuesResponse, error) {
	if branchId != "" && runId != "" {
		return nil, errors.New("only one of branchId and runId may be specified (both were non-empty)")
	}
	if branchId == "" && runId == "" {
		return nil, errors.New("one of branchId and runId may be specified (both were empty)")
	}
	result := &GetV1IssuesResponse{}
	params := map[string]interface{}{
		"project-id":   projectId,
		"page[offset]": fmt.Sprintf("%d", offset),
		"page[limit]":  fmt.Sprintf("%d", limit),
	}
	if branchId != "" {
		params["branch-id"] = branchId
	}
	if runId != "" {
		params["run-id[]"] = []string{runId}
	}
	_, err := client.GetJson(params, result, "api/query/v1/issues")
	return result, err
}

type GetV0RollUpCountsResponse struct {
	Data []struct {
		Id            string
		Type          string
		Relationships map[string]struct {
			Data *struct {
				Id   string
				Type string
			}
		}
		Attributes struct {
			Value int
		}
	}
	Included []struct {
		Id         string
		Type       string
		Attributes map[string]interface{} // TODO type?
	}
	Meta struct {
		Offset   int
		Limit    int
		Total    int
		Complete bool
		GroupBy  string `json:"group-by"`
		RunCount int    `json:"run-count"`
	}
}

func (client *Client) GetV0RollUpCounts(projectId string, branchId string, limit int) (*GetV0RollUpCountsResponse, error) {
	result := &GetV0RollUpCountsResponse{}
	params := map[string]interface{}{
		"project-id":  projectId,
		"branch-id":   branchId,
		"page[limit]": fmt.Sprintf("%d", limit),
	}
	_, err := client.GetJson(params, result, "api/query/v0/roll-up-counts")
	return result, err
}

type GetV1IssueResponse struct {
	Data     V1IssueResponse
	Included []V1IssueIncluded
}

func (client *Client) GetV1Issue(projectId string, branchId string, issueId string) (*GetV1IssueResponse, error) {
	result := &GetV1IssueResponse{}
	params := map[string]interface{}{
		"project-id": projectId,
		"branch-id":  branchId,
	}
	_, err := client.GetJson(params, result, "api/query/v1/issues/%s", issueId)
	return result, err
}

// TODO:
// /v1/roll-up-counts
// /v1/counts/status
