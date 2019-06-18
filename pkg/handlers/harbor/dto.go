package harbor

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/srelab/watcher/pkg/handlers/shared"

	"github.com/google/go-querystring/query"
)

const (
	defaultPageSize int64 = 15
	maxPageSize     int64 = 500
)

type Cfg struct {
	Labels map[string]string `json:"labels"`
}

type Tag struct {
	Digest        string          `json:"digest"`
	Name          string          `json:"name"`
	Size          int64           `json:"size"`
	Architecture  string          `json:"architecture"`
	OS            string          `json:"os"`
	OSVersion     string          `json:"os.version"`
	DockerVersion string          `json:"docker_version"`
	Author        string          `json:"author"`
	Created       shared.Datetime `json:"created"`
	Config        *Cfg            `json:"config"`
}

type Tags []Tag

func (tags Tags) Len() int {
	return len(tags)
}

func (tags Tags) Less(i, j int) bool {
	return tags[i].Created.Before(tags[j].Created.Time)
}

func (tags Tags) Swap(i, j int) {
	tags[i], tags[j] = tags[j], tags[i]
}

// pagination information
type Pagination struct {
	Page     int64 `query:"page" url:"page,omitempty"`
	PageSize int64 `query:"page_size" url:"page_size,omitempty"`
}

type GetProjectPayload struct {
	Pagination
	Name string `query:"name" url:"name,omitempty"`
}

type GetProjectRepoPayload struct {
	Pagination
	ID string `url:"project_id,omitempty"`
}

type GetProjectRepoTagPayload struct {
	Limit int    `query:"limit"`
	Sort  string `query:"sort" validate:"in=desc;asc"`
}

type CreateProjectRepoTagPayload struct {
	Override bool `json:"override" validate:"required"`
	Src      struct {
		Tag     string `json:"tag" validate:"required"`
		Repo    string `json:"repo" validate:"required"`
		Project string `json:"project" validate:"required"`
		Digest  string `json:"digest" validate:"required"`
	} `json:"src" validate:"required"`
}

type CreateProjectPayload struct {
	Name   string `json:"name" validate:"required"`
	Public bool   `json:"public" validate:"required"`
}

// Convert GetProjectPayload struct to querystring
func (p *GetProjectPayload) ToQueryString() string {
	if p.PageSize < 1 || p.PageSize > maxPageSize {
		p.PageSize = defaultPageSize
	}

	values, _ := query.Values(p)
	return values.Encode()
}

// Convert GetProjectRepoPayload struct to querystring
func (p *GetProjectRepoPayload) ToQueryString() string {
	values, _ := query.Values(p)
	return values.Encode()
}

// Convert GetProjectRepoTagPayload struct to querystring
func (p *GetProjectRepoTagPayload) ToQueryString() string {
	values, _ := query.Values(p)
	return values.Encode()
}

// Convert CreateProjectPayload struct to JSON
func (p *CreateProjectPayload) ToJSON() string {
	jsonstr, _ := json.Marshal(map[string]interface{}{
		"project_name": p.Name,
		"metadata": map[string]string{
			"public": strconv.FormatBool(p.Public),
		},
	})

	return string(jsonstr)
}

// Convert CreateProjectRepoTagPayload struct to JSON
func (p *CreateProjectRepoTagPayload) ToJSON() string {
	jsonstr, _ := json.Marshal(map[string]interface{}{
		"override":  p.Override,
		"src_image": fmt.Sprintf("%s/%s:%s", p.Src.Project, p.Src.Repo, p.Src.Digest),
		"tag":       p.Src.Tag,
	})

	return string(jsonstr)
}
