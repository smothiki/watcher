package gateway

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/srelab/watcher/pkg/handlers/shared"

	"git.srelab.cn/go/resty"

	apiV1 "k8s.io/api/core/v1"

	"github.com/srelab/common/log"
	"github.com/srelab/watcher/pkg/g"
)

type Handler struct {
	logger  log.Logger
	configs []g.GatewayConfig
}

func (h *Handler) Name() string {
	return "gateway"
}

// initialize the gateway handler
// it will be responsible for handling kube events, regsiter and unregsiter pods
func (h *Handler) Init(config *g.Configuration) error {
	h.configs = config.Handlers.GatewayConfigs
	h.logger = log.With("handlers", h.Name())

	return nil
}

// handling the created event,
// since the kube only generates the created event when a new `deployment` is created,
// only the log is output here.
func (h *Handler) Created(e *shared.Event) {
	if e.ResourceType != shared.ResourceTypePod {
		h.logger.Debug("invalid resource type, skipped")
		return
	}

	pod := e.Object.(*apiV1.Pod)
	h.logger.Debugf("pod[%s] did not do anything when created, skipped", pod.Name)
}

func (h *Handler) Deleted(e *shared.Event) {
	if e.ResourceType != shared.ResourceTypePod {
		h.logger.Debug("invalid resource type, skipped")
		return
	}

	pod := e.Object.(*apiV1.Pod)
	h.logger.Debugf("pod[%s] did not do anything when deleted, skipped", pod.Name)
}

func (h *Handler) Updated(event *shared.Event) {
	// Convert the associated object in the event to pod
	pod, oldPod := event.Object.(*apiV1.Pod), event.OldObject.(*apiV1.Pod)

	// Get all valid services from the pod
	services, err := event.GetPodServices(pod)
	if err != nil {
		h.logger.Debug(err)
		return
	}

	if pod.GetDeletionTimestamp() != nil && oldPod.Status.Phase == apiV1.PodRunning {
		for _, s := range services {
			namespace := pod.GetNamespace()
			regURL := h.getURL(namespace, fmt.Sprintf("/upstreams/%s/unregister", s.Name))
			if regURL == "" {
				h.logger.Errorf("namespace `%s` has no associated gateway config, pod[%s] register skipped", namespace, s.String())
				continue
			}

			res, err := h.getRequest().SetBody(s).Post(regURL)
			if err != nil {
				h.logger.Errorf("pod[%s] - [%s] unregister error: %s", pod.Name, res.String(), err)
				continue
			}

			if res.StatusCode() != http.StatusOK {
				h.logger.Errorf("pod[%s] - [%s] unregister error: %d", pod.Name, res.String(), res.StatusCode())
				continue
			}

			h.logger.Infof("pod[%s] - [%s] unregister successful", pod.Name, s.String())
		}
	} else if pod.Status.Phase == apiV1.PodRunning && !shared.IsPodContainersReady(oldPod.Status.Conditions) {
		for _, s := range services {
			namespace := pod.GetNamespace()

			// Get the URL of the handler in memory, when the `namespace` does not exist, skip the service
			regURL := h.getURL(namespace, fmt.Sprintf("/upstreams/%s/register", s.Name))
			if regURL == "" {
				h.logger.Errorf("namespace `%s` has no associated gateway config, pod[%s] register skipped", namespace, s.String())
				continue
			}

			res, err := h.getRequest().SetBody(s).Post(regURL)
			if err != nil {
				h.logger.Errorf("pod[%s] - [%s] register error: %s", pod.Name, res.String(), err)
				continue
			}

			if res.StatusCode() != http.StatusOK {
				h.logger.Errorf("pod[%s] - [%s] register error: %d", pod.Name, res.String(), res.StatusCode())
				continue
			}

			h.logger.Infof("pod[%s] - [%s] register successful", pod.Name, s.String())
		}
	} else {
		j, _ := json.Marshal(pod)
		oj, _ := json.Marshal(oldPod)
		fmt.Println(string(j))
		fmt.Println(string(oj))

		h.logger.Errorf("pod[%s] unknown event, need a-e", pod.Name)
	}
}

func (h *Handler) getRequest() *resty.Request {
	r := resty.New().SetRetryCount(3).SetRetryWaitTime(5 * time.Second).SetRetryMaxWaitTime(10 * time.Second)
	return r.R().SetHeader("Content-Type", "application/json")
}

func (h *Handler) getURL(namespace, path string) string {
	for _, config := range h.configs {
		if config.Namespace == namespace {
			return fmt.Sprintf("http://%s:%s/%s", config.Host, config.Port, strings.TrimLeft(path, "/"))
		}
	}

	return ""
}
