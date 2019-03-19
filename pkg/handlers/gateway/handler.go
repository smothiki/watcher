package gateway

import (
	"fmt"
	"strings"
	"time"

	"git.srelab.cn/go/resty"

	"github.com/srelab/common/log"
	"github.com/srelab/watcher/pkg/g"
	"github.com/srelab/watcher/pkg/handlers/shared"
)

type Handler struct {
	logger  log.Logger
	configs []g.GatewayConfig
}

func (h *Handler) Name() string            { return "gateway" }
func (h *Handler) Handler() *Handler       { return h }
func (h *Handler) RoutePrefix() string     { return "/" + h.Name() }
func (h *Handler) Created(e *shared.Event) {}
func (h *Handler) Deleted(e *shared.Event) {}
func (h *Handler) Updated(e *shared.Event) {}

// initialize the gateway handler
// it will be responsible for handling kube events, regsiter and unregsiter pods
func (h *Handler) Init(config *g.Configuration, handlers ...interface{}) error {
	h.configs = config.Handlers.GatewayConfigs
	h.logger = log.With("handlers", h.Name())

	return nil
}

// Return request client , default 5 seconds timeout and automatically retry
func (h *Handler) Request() *resty.Request {
	r := resty.New().SetRetryCount(3).SetRetryWaitTime(5 * time.Second).SetRetryMaxWaitTime(10 * time.Second)
	return r.R().SetHeader("Content-Type", "application/json")
}

// Returns the interface address of the gateway
// namespace: kubernetes namespace
// path: request path
func (h *Handler) URL(namespace, path string) string {
	for _, config := range h.configs {
		if config.Namespace == namespace {
			return fmt.Sprintf("http://%s:%s/%s", config.Host, config.Port, strings.TrimLeft(path, "/"))
		}
	}

	return ""
}
