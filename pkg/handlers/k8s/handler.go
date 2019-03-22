package k8s

import (
	"fmt"

	"github.com/srelab/watcher/pkg/g"
	"github.com/srelab/watcher/pkg/handlers/shared"
	"k8s.io/client-go/kubernetes"
)

// Default handler implements Handler interface,
// print each event with JSON format
type Handler struct {
	handlers struct {
		kube kubernetes.Interface
	}
}

func (h *Handler) Name() string            { return "k8s" }
func (h *Handler) Handler() *Handler       { return h }
func (h *Handler) RoutePrefix() string     { return "/" + h.Name() }
func (h *Handler) Close()                  {}
func (h *Handler) Created(e *shared.Event) {}
func (h *Handler) Deleted(e *shared.Event) {}
func (h *Handler) Updated(e *shared.Event) {}

// Init initializes handler configuration
// Do nothing for default handler
func (h *Handler) Init(config *g.Configuration, objs ...interface{}) error {
	fmt.Println(objs)
	for _, obj := range objs {
		switch object := obj.(type) {
		case kubernetes.Interface:
			h.handlers.kube = object
		}
	}

	return nil
}
