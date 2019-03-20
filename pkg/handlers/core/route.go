package core

import (
	"net/http"

	"github.com/labstack/echo"
	"github.com/srelab/watcher/pkg/handlers/shared"
)

func (h *Handler) AddRoutes(group *echo.Group) {
	group.GET(shared.EmptyPath, h.getName)

	serviceGroup := group.Group("/services")
	serviceGroup.Use(func() echo.MiddlewareFunc {
		return func(next echo.HandlerFunc) echo.HandlerFunc {
			return func(ctx echo.Context) error {
				var p = new(servicePayload)
				p.Name = ctx.Param("name")

				if err := ctx.Bind(p); err != nil {
					return shared.Responder{Status: http.StatusBadRequest, Success: false, Msg: err}.JSON(ctx)
				}

				ctx.Set("payload", p)
				return next(ctx)
			}
		}
	}())

	serviceGroup.PUT("/:name", h.createService)
	serviceGroup.DELETE("/:name", h.deleteService)
}

func (h *Handler) getName(ctx echo.Context) error {
	return shared.Responder{Status: http.StatusOK, Success: true, Result: h.Name()}.JSON(ctx)
}
