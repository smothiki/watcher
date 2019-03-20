package harbor

import (
	"github.com/srelab/watcher/pkg/g"
	"github.com/srelab/watcher/pkg/handlers/shared"
)

type Handler struct{}

func (h *Handler) Name() string            { return "harbor" }
func (h *Handler) Handler() *Handler       { return h }
func (h *Handler) RoutePrefix() string     { return "/" + h.Name() }
func (h *Handler) Created(e *shared.Event) {}
func (h *Handler) Deleted(e *shared.Event) {}
func (h *Handler) Updated(e *shared.Event) {}

func (h *Handler) Init(config *g.Configuration, handlers ...interface{}) error {
	return nil
}
