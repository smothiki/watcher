package public

import (
	"github.com/srelab/watcher/pkg/g"
	"github.com/srelab/watcher/pkg/handlers/shared"
)

// Default handler implements Handler interface,
// print each event with JSON format
type Handler struct {
	shared.Handler
}

func (h *Handler) Name() string {
	return "default"
}

// Init initializes handler configuration
// Do nothing for default handler
func (h *Handler) Init(config *g.Configuration) error {
	return nil
}

func (h *Handler) Created(e *shared.Event) {
}

func (h *Handler) Deleted(e *shared.Event) {
}

func (h *Handler) Updated(e *shared.Event) {
}
