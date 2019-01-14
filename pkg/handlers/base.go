package handlers

import (
	"github.com/srelab/watcher/pkg/event"
	"github.com/srelab/watcher/pkg/g"
)

// Handler is implemented by any handler.
// The Handle method is used to process event
type Handler interface {
	Name() string

	Init(config *g.Configuration) error
	Created(event event.Event)
	Deleted(event event.Event)
	Updated(event event.Event)
}

// Default handler implements Handler interface,
// print each event with JSON format
type DefaultHandler struct {
	Handler
}

func (h *DefaultHandler) Name() string {
	return "default"
}

// Init initializes handler configuration
// Do nothing for default handler
func (h *DefaultHandler) Init(config *g.Configuration) error {
	return nil
}

func (h *DefaultHandler) Created(e event.Event) {
}

func (h *DefaultHandler) Deleted(e event.Event) {
}

func (h *DefaultHandler) Updated(e event.Event) {
}
