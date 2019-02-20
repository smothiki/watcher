package k8s

import (
	"github.com/srelab/watcher/pkg/g"
	"github.com/srelab/watcher/pkg/handlers/shared"
	"k8s.io/client-go/kubernetes"
)

// kubernetes

// Default handler implements Handler interface,
// print each event with JSON format
type Handler struct {
	shared.Handler
	k8s struct {
		client kubernetes.Interface
	}
}

func (h *Handler) Name() string {
	return "k8s"
}

func (h *Handler) RoutePrefix() string {
	return "/" + h.Name()
}

// Init initializes handler configuration
// Do nothing for default handler
func (h *Handler) Init(config *g.Configuration, clientSet kubernetes.Interface) error {
	h.k8s.client = clientSet

	return nil
}

func (h *Handler) Created(e *shared.Event) {
}

func (h *Handler) Deleted(e *shared.Event) {
}

func (h *Handler) Updated(e *shared.Event) {
}
