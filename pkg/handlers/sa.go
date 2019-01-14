package handlers

import (
	"github.com/srelab/watcher/pkg/event"
	"github.com/srelab/watcher/pkg/g"
)

type SAHandler struct {
}

func (h *SAHandler) Name() string {
	return "sa"
}

func (h *SAHandler) Init(config *g.Configuration) error {
	return nil
}

func (h *SAHandler) Created(e event.Event) {

}

func (h *SAHandler) Deleted(e event.Event) {

}

func (h *SAHandler) Updated(e event.Event) {

}
