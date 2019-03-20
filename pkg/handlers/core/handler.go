package core

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strconv"

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
func (h *Handler) Created(e *shared.Event) {}
func (h *Handler) Deleted(e *shared.Event) {}
func (h *Handler) Updated(e *shared.Event) {}

// Initialize log and dependent handler
func (h *Handler) Init(config *g.Configuration, handlers ...interface{}) error {
	h.logger = log.With("handlers", h.Name())

	for _, handler := range handlers {
		switch object := handler.(type) {
		case *etcd.Handler:
			h.handlers.etcd = object
		case *gateway.Handler:
			h.handlers.gateway = object
		}

	}

	return nil
}

// Convert service to DNS resolution record and write to etcd for use by CoreDNS
func (h *Handler) etcdCreate(payload interface{}) error {
	s := payload.(*ServicePayload)

	key := filepath.Join(h.handlers.etcd.DNSPrefix(), s.DNSName(), s.DNSKey())
	if _, err := h.handlers.etcd.PutKey(key, s.DNSRecord(), 0); err != nil {
		return fmt.Errorf("etcd key cannot be create: %s", err)
	}

	h.logger.Infof("[etcd][%s] - [%s] create successful", s.Name, s.String())
	return nil
}

// Remove DNS resolution records from CoreDNS
func (h *Handler) etcdDelete(payload interface{}) error {
	s := payload.(*ServicePayload)

	res, err := h.handlers.etcd.GetKey(
		filepath.Join(h.handlers.etcd.DNSPrefix(), s.DNSName()),
		false,
		true,
		0,
	)

	if err != nil {
		return fmt.Errorf("get key error: %s", err)
	}

	for _, item := range res.Kvs {
		key := filepath.Join(h.handlers.etcd.DNSPrefix(), s.DNSName(), s.DNSKey())
		if string(item.Key) != key {
			continue
		}

		if _, err := h.handlers.etcd.DeleteKey(key, false); err != nil {
			return fmt.Errorf("etcd key cannot be delete: %s", err)
		}
	}

	h.logger.Infof("[etcd][%s] - [%s] delete successful", s.Name, s.String())
	return nil
}

// Write service information to the API Gateway
func (h *Handler) gatewayCreate(payload interface{}) error {
	s := payload.(*ServicePayload)

	// Get the URL of the handler in memory, when the `namespace` does not exist, skip the service
	regURL := h.handlers.gateway.URL(s.Namespace, fmt.Sprintf("/upstreams/%s/register", s.Name))
	if regURL == "" {
		return fmt.Errorf(
			"namespace `%s` has no associated gateway config, %s register skipped",
			s.Namespace, s.String(),
		)
	}

	res, err := h.handlers.gateway.Request().SetBody(map[string]string{
		"name":    s.Name,
		"host":    s.Host,
		"type":    s.Protocol,
		"port":    strconv.Itoa(s.Port),
		"hc_path": s.HealthCheck.Path,
		"hc_port": strconv.Itoa(s.HealthCheck.Port),
	}).Post(regURL)

	if err != nil {
		return fmt.Errorf("[%s] - [%s] register error: %s", s.Name, res.String(), err)
	}

	if res.StatusCode() != http.StatusOK {
		return fmt.Errorf("[%s] - [%s] register error: %d", s.Name, res.String(), res.StatusCode())
	}

	h.logger.Infof("[gateway][%s] - [%s] create successful", s.Name, s.String())
	return nil
}

// Remove service information from the API Gateway
func (h *Handler) gatewayDelete(payload interface{}) error {
	s := payload.(*ServicePayload)

	// Get the URL of the handler in memory, when the `namespace` does not exist, skip the service
	regURL := h.handlers.gateway.URL(s.Namespace, fmt.Sprintf("/upstreams/%s/unregister", s.Name))
	if regURL == "" {
		return fmt.Errorf(
			"namespace `%s` has no associated gateway config, %s register skipped",
			s.Namespace, s.String(),
		)
	}

	res, err := h.handlers.gateway.Request().SetBody(s).Post(regURL)
	if err != nil {
		return fmt.Errorf("pod[%s] - [%s] unregister error: %s", s.Name, res.String(), err)
	}

	if res.StatusCode() != http.StatusOK {
		return fmt.Errorf("pod[%s] - [%s] unregister error: %d", s.Name, res.String(), res.StatusCode())
	}

	h.logger.Infof("[gateway][%s] - [%s] delete successful", s.Name, s.String())
	return nil
}
