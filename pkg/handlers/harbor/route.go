package harbor

import (
	"fmt"
	"net/http"

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

}

func (h *Handler) getName(ctx echo.Context) error {
	return shared.Responder{Status: http.StatusOK, Success: true, Result: h.Name()}.JSON(ctx)
}

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
