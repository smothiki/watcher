package sa

import (
	"github.com/srelab/watcher/pkg/g"
	"github.com/srelab/watcher/pkg/handlers/shared"
)

type Handler struct {
}

func (h *Handler) Name() string {
	return "sa"
}

func (h *Handler) RoutePrefix() string {
	return "/" + h.Name()
}

func (h *Handler) Init(config *g.Configuration) error {
	return nil
}

func (h *Handler) Created(e *shared.Event) {

}

func (h *Handler) Deleted(e *shared.Event) {

}

func (h *Handler) Updated(e *shared.Event) {

}
