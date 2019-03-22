package gateway

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"git.srelab.cn/go/resty"

	"github.com/srelab/common/log"
	"github.com/srelab/watcher/pkg/g"
	"github.com/srelab/watcher/pkg/handlers/shared"

	apiV1 "k8s.io/api/core/v1"
)

type Handler struct {
	logger  log.Logger
	configs []g.GatewayConfig
}

func (h *Handler) Name() string            { return "gateway" }
func (h *Handler) Handler() *Handler       { return h }
func (h *Handler) RoutePrefix() string     { return "/" + h.Name() }
func (h *Handler) Close()                  {}
func (h *Handler) Created(e *shared.Event) {}
func (h *Handler) Updated(e *shared.Event) {}

// Remove the service from the gateway when it detects that the pod is destroyed
func (h *Handler) Deleted(e *shared.Event) {
	switch object := e.Object.(type) {
	case *apiV1.Pod:
		services, err := e.GetPodServices(object)
		if err != nil {
			h.logger.Errorf("an error occurred while getting services: %s", err)
			return
		}

		for _, service := range services {
			if err := h.DeleteService(service); err != nil {
				h.logger.Errorf("an error occurred while deleting the service: %s", err)
			}
		}
	default:
		return
	}
}

// initialize the gateway handler
// it will be responsible for handling kube events, regsiter and unregsiter pods
func (h *Handler) Init(config *g.Configuration, objs ...interface{}) error {
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

// Write service information to the API Gateway
func (h *Handler) CreateService(service *shared.ServicePayload) error {
	// Get the URL of the handler in memory, when the `namespace` does not exist, skip the service
	regURL := h.URL(service.Namespace, fmt.Sprintf("/upstreams/%s/register", service.Name))
	if regURL == "" {
		return fmt.Errorf(
			"namespace `%s` has no associated gateway config, %s register skipped",
			service.Namespace, service.String(),
		)
	}

	res, err := h.Request().SetBody(map[string]string{
		"name":    service.Name,
		"host":    service.Host,
		"type":    service.Protocol,
		"port":    strconv.Itoa(service.Port),
		"hc_path": service.HealthCheck.Path,
		"hc_port": strconv.Itoa(service.HealthCheck.Port),
	}).Post(regURL)

	if err != nil {
		return fmt.Errorf("[%s] - [%s] register error: %s", service.Name, res.String(), err)
	}

	if res.StatusCode() != http.StatusOK {
		return fmt.Errorf("[%s] - [%s] register error: %d", service.Name, res.String(), res.StatusCode())
	}

	h.logger.Infof("[gateway][%s] - [%s] create successful", service.Name, service.String())
	return nil
}

// Remove service information from the API Gateway
func (h *Handler) DeleteService(service *shared.ServicePayload) error {
	// Get the URL of the handler in memory, when the `namespace` does not exist, skip the service
	regURL := h.URL(service.Namespace, fmt.Sprintf("/upstreams/%s/unregister", service.Name))
	if regURL == "" {
		return fmt.Errorf(
			"namespace `%s` has no associated gateway config, %s register skipped",
			service.Namespace, service.String(),
		)
	}

	res, err := h.Request().SetBody(service).Post(regURL)
	if err != nil {
		return fmt.Errorf("pod[%s] - [%s] unregister error: %s", service.Name, res.String(), err)
	}

	if res.StatusCode() != http.StatusOK {
		return fmt.Errorf("pod[%s] - [%s] unregister error: %d", service.Name, res.String(), res.StatusCode())
	}

	h.logger.Infof("[gateway][%s] - [%s] delete successful", service.Name, service.String())
	return nil
}
