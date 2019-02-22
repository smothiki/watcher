package k8s

import (
	"net/http"

	"github.com/labstack/echo"
	"github.com/srelab/watcher/pkg/handlers/shared"

	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// List objects of kind Namespace
// The control list content can be filtered by the options payload.
func (h *Handler) getNode(ctx echo.Context) error {
	var p = new(getOptionsPayload)
	if err := ctx.Bind(p); err != nil {
		return shared.Responder{Status: http.StatusBadRequest, Success: false, Msg: err}.JSON(ctx)
	}

	namespaces, err := h.k8s.client.CoreV1().Nodes().List(metaV1.ListOptions{
		FieldSelector: p.FieldSelector,
		LabelSelector: p.LabelSelector,
		Continue:      p.Continue,
		Limit:         p.Limit,
	})

	if err != nil {
		return shared.Responder{Status: http.StatusInternalServerError, Success: false, Msg: err}.JSON(ctx)
	}

	return shared.Responder{Status: http.StatusOK, Success: true, Result: namespaces}.JSON(ctx)
}
