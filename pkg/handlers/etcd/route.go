package etcd

import (
	"net/http"

	"github.com/labstack/echo"
	"github.com/srelab/watcher/pkg/handlers/shared"
)

func (h *Handler) AddRoutes(group *echo.Group) *echo.Group {
	group.GET("", h.getName)
	return group
}

func (h *Handler) getName(ctx echo.Context) error {
	return shared.Responder{Status: http.StatusOK, Success: true, Result: h.Name()}.JSON(ctx)
}
