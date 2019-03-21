package harbor

import (
	"fmt"
	"net/http"
	"sort"

	// harbor models
	"github.com/goharbor/harbor/src/common/models"

	"github.com/labstack/echo"
	"github.com/srelab/watcher/pkg/handlers/shared"
)

func (h *Handler) AddRoutes(group *echo.Group) {
	group.GET(shared.EmptyPath, h.getName)
	group.GET("/projects", h.getProject)
	group.POST("/projects", h.createProject)
	group.DELETE("/projects/:id", h.deleteProject)
	group.GET("/projects/:id/repositories", h.getProjectRepo)

	group.GET("/repositories/:project/:repo/tags", h.getProjectRepoTag)

}

func (h *Handler) getName(ctx echo.Context) error {
	return shared.Responder{Status: http.StatusOK, Success: true, Result: h.Name()}.JSON(ctx)
}

// Get the list of harbor projects
// Filter and page query with GetProjectPayload
func (h *Handler) getProject(ctx echo.Context) error {
	p := new(GetProjectPayload)
	if err := ctx.Bind(p); err != nil {
		return shared.Responder{Status: http.StatusBadRequest, Success: false, Msg: err}.JSON(ctx)
	}

	res, err := h.Request().SetQueryString(p.ToQueryString()).SetResult([]models.Project{}).Get(h.URL("/api/projects"))
	if err != nil {
		return shared.Responder{Status: http.StatusInternalServerError, Success: false, Msg: err}.JSON(ctx)
	}

	if res.StatusCode() != http.StatusOK {
		err = fmt.Errorf("failed to get the list of projects, status code[%d]: %s", res.StatusCode(), res.Body())
		return shared.Responder{Status: http.StatusInternalServerError, Success: false, Msg: err}.JSON(ctx)
	}

	return shared.Responder{Status: http.StatusOK, Success: true, Result: res.Result()}.JSON(ctx)
}

// Create harbor project
func (h *Handler) createProject(ctx echo.Context) error {
	p := new(CreateProjectPayload)
	if err := ctx.Bind(p); err != nil {
		return shared.Responder{Status: http.StatusBadRequest, Success: false, Msg: err}.JSON(ctx)
	}

	res, err := h.Request().SetBody(p.ToJSON()).Post(h.URL("/api/projects"))
	if err != nil {
		return shared.Responder{Status: http.StatusInternalServerError, Success: false, Msg: err}.JSON(ctx)
	}

	if res.StatusCode() != http.StatusCreated {
		err = fmt.Errorf("failed to create project, status code[%d]: %s", res.StatusCode(), res.Body())
		return shared.Responder{Status: http.StatusInternalServerError, Success: false, Msg: err}.JSON(ctx)
	}

	return shared.Responder{Status: http.StatusOK, Success: true}.JSON(ctx)
}

// Delete harbor project
// params:
//     id -> project id
func (h *Handler) deleteProject(ctx echo.Context) error {
	id := ctx.Param("id")
	res, err := h.Request().Delete(h.URL(fmt.Sprintf("/api/projects/%s", id)))
	if err != nil {
		return shared.Responder{Status: http.StatusInternalServerError, Success: false, Msg: err}.JSON(ctx)
	}

	if res.StatusCode() != http.StatusOK {
		err = fmt.Errorf("failed to create project, status code[%d]: %s", res.StatusCode(), res.Body())
		return shared.Responder{Status: http.StatusInternalServerError, Success: false, Msg: err}.JSON(ctx)
	}

	return shared.Responder{Status: http.StatusOK, Success: true}.JSON(ctx)
}

// Get repositories via project id
// Filter and page query with GetProjectRepoPayload
func (h *Handler) getProjectRepo(ctx echo.Context) error {
	p := new(GetProjectRepoPayload)
	p.ID = ctx.Param("id")

	if err := ctx.Bind(p); err != nil {
		return shared.Responder{Status: http.StatusBadRequest, Success: false, Msg: err}.JSON(ctx)
	}

	res, err := h.Request().SetQueryString(p.ToQueryString()).SetResult([]interface{}{}).Get(h.URL("/api/repositories"))
	if err != nil {
		return shared.Responder{Status: http.StatusInternalServerError, Success: false, Msg: err}.JSON(ctx)
	}

	if res.StatusCode() != http.StatusOK {
		err = fmt.Errorf("failed to get the repo list of projects, status code[%d]: %s", res.StatusCode(), res.Body())
		return shared.Responder{Status: http.StatusInternalServerError, Success: false, Msg: err}.JSON(ctx)
	}

	return shared.Responder{Status: http.StatusOK, Success: true, Result: res.Result()}.JSON(ctx)
}

// Get the tag list of the project
// Use GetProjectRepoTagPayload to limit return entries and control sorting
func (h *Handler) getProjectRepoTag(ctx echo.Context) error {
	p := new(GetProjectRepoTagPayload)
	repoName := ctx.Param("repo")
	projectName := ctx.Param("project")

	if err := ctx.Bind(p); err != nil {
		return shared.Responder{Status: http.StatusBadRequest, Success: false, Msg: err}.JSON(ctx)
	}

	tags := make(Tags, 0)
	res, err := h.Request().SetQueryString(p.ToQueryString()).SetResult(&tags).Get(
		h.URL(fmt.Sprintf("/api/repositories/%s/%s/tags", projectName, repoName)),
	)

	if err != nil {
		return shared.Responder{Status: http.StatusInternalServerError, Success: false, Msg: err}.JSON(ctx)
	}

	if res.StatusCode() != http.StatusOK {
		err = fmt.Errorf("failed to get the tag list of repo name, status code[%d]: %s", res.StatusCode(), res.Body())
		return shared.Responder{Status: http.StatusInternalServerError, Success: false, Msg: err}.JSON(ctx)
	}

	if p.Sort == "desc" {
		sort.Sort(sort.Reverse(tags))
	} else {
		sort.Sort(tags)
	}

	if p.Limit > 0 {
		if p.Limit > len(tags) {
			p.Limit = len(tags)
		}

		tags = tags[0:p.Limit]
	}
	return shared.Responder{Status: http.StatusOK, Success: true, Result: tags}.JSON(ctx)
}
