package core

import (
	"github.com/srelab/common/log"
	"github.com/srelab/watcher/pkg/g"
	"github.com/srelab/watcher/pkg/handlers/etcd"
	"github.com/srelab/watcher/pkg/handlers/gateway"
	"github.com/srelab/watcher/pkg/handlers/shared"
)

type Handler struct {
	handlers struct {
		// The core handler needs to use the handler for etcd and gateway
		etcd    *etcd.Handler
		gateway *gateway.Handler
	}
	logger log.Logger
}

func (h *Handler) Name() string            { return "core" }
func (h *Handler) RoutePrefix() string     { return "/" + h.Name() }
func (h *Handler) Close()                  {}
func (h *Handler) Created(e *shared.Event) {}
func (h *Handler) Deleted(e *shared.Event) {}
func (h *Handler) Updated(e *shared.Event) {}

// Initialize log and dependent handler
func (h *Handler) Init(config *g.Configuration, itfs ...interface{}) error {
	h.logger = log.With("handlers", h.Name())

	for _, itf := range itfs {
		switch object := itf.(type) {
		case *etcd.Handler:
			h.handlers.etcd = object
		case *gateway.Handler:
			h.handlers.gateway = object
		}

	}

	return nil
}
