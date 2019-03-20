package core

import (
	"net/http"

	"github.com/srelab/watcher/pkg/handlers/shared"

	"github.com/labstack/echo"
)

// Handling requests to create services
func (h *Handler) createService(ctx echo.Context) error {
	p := ctx.Get("payload")

	if err := h.etcdCreate(p); err != nil {
		h.logger.Error(err)
		return shared.Responder{Status: http.StatusInternalServerError, Success: false, Msg: err}.JSON(ctx)
	}

	if err := h.gatewayCreate(p); err != nil {
		h.logger.Error(err)
		return shared.Responder{Status: http.StatusInternalServerError, Success: false, Msg: err}.JSON(ctx)
	}

	return shared.Responder{Status: http.StatusOK, Success: true}.JSON(ctx)
}

// Handling requests to delete a service
func (h *Handler) deleteService(ctx echo.Context) error {
	p := ctx.Get("payload")

	if err := h.etcdDelete(p); err != nil {
		h.logger.Error(err)
		return shared.Responder{Status: http.StatusInternalServerError, Success: false, Msg: err}.JSON(ctx)
	}

	if err := h.gatewayDelete(p); err != nil {
		h.logger.Error(err)
		return shared.Responder{Status: http.StatusInternalServerError, Success: false, Msg: err}.JSON(ctx)
	}

	return nil
}
