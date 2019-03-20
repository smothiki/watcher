package harbor

import (
	"encoding/json"
	"strconv"

	"github.com/google/go-querystring/query"
)

const (
	defaultPageSize int64 = 15
	maxPageSize     int64 = 500
)

type Pagination struct {
	Page     int64 `query:"page" url:"page,omitempty"`
	PageSize int64 `query:"page_size" url:"page_size,omitempty"`
}

type GetProjectPayload struct {
	Pagination
	Name string `query:"name" url:"name,omitempty"`
}

type CreateProjectPayload struct {
	Name   string `json:"name" validate:"required"`
	Public bool   `json:"public" validate:"required"`
}

func (p *GetProjectPayload) ToQueryString() string {
	if p.PageSize < 1 || p.PageSize > maxPageSize {
		p.PageSize = defaultPageSize
	}

	values, _ := query.Values(p)
	return values.Encode()
}

func (p *CreateProjectPayload) ToJSON() string {
	jsonstr, _ := json.Marshal(map[string]interface{}{
		"project_name": p.Name,
		"metadata": map[string]string{
			"public": strconv.FormatBool(p.Public),
		},
	})

	return string(jsonstr)
}
