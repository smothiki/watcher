package handlers

import (
	"context"
	"path/filepath"
	"strings"
	"time"

	"github.com/srelab/watcher/pkg/util"

	"github.com/pkg/errors"
	apiV1 "k8s.io/api/core/v1"

	"github.com/coreos/etcd/pkg/transport"
	"github.com/srelab/common/log"
	"github.com/srelab/watcher/pkg/event"
	"github.com/srelab/watcher/pkg/g"
	"go.etcd.io/etcd/clientv3"
)

type EtcdHandler struct {
	config *g.EtcdConfig
	client *clientv3.Client
	logger log.Logger
}

func (h *EtcdHandler) Name() string {
	return "etcd"
}

func (h *EtcdHandler) Init(config *g.Configuration) error {
	h.config = config.Handlers.EtcdConfig
	h.config.Prefix = strings.TrimRight(h.config.Prefix, "/")

	// simply judge whether prefix starts with "/" character
	if !strings.HasPrefix(h.config.Prefix, "/") {
		return errors.New("invalid coredns prefix")
	}

	tlsInfo := transport.TLSInfo{
		CertFile:      config.Handlers.EtcdConfig.CertPath,
		KeyFile:       config.Handlers.EtcdConfig.KeyPath,
		TrustedCAFile: config.Handlers.EtcdConfig.CAPath,
	}

	tlsConfig, err := tlsInfo.ClientConfig()
	if err != nil {
		return errors.Wrap(err, "generates a tls.Config error")
	}

	client, err := clientv3.New(clientv3.Config{
		Endpoints:   h.config.Endpoints,
		DialTimeout: h.config.Timeout * time.Second,
		TLS:         tlsConfig,
	})

	if err != nil {
		return errors.Wrap(err, "an error occurred while initializing handler")
	}

	h.client = client
	h.logger = log.With("handlers", h.Name())
	return nil
}

func (h *EtcdHandler) Close() {
	h.client.Close()
}

func (h *EtcdHandler) Created(e event.Event) {
	if e.ResourceType != event.ResourceTypePod {
		h.logger.Debug("invalid resource type, skipped")
		return
	}

	pod := e.Object.(*apiV1.Pod)
	h.logger.Debugf("pod[%s] did not do anything when created, skipped", pod.Name)
}

func (h *EtcdHandler) Deleted(e event.Event) {
	if e.ResourceType != event.ResourceTypePod {
		h.logger.Debug("invalid resource type, skipped")
		return
	}

	pod := e.Object.(*apiV1.Pod)
	h.logger.Debugf("pod[%s] did not do anything when deleted, skipped", pod.Name)
}

func (h *EtcdHandler) Updated(e event.Event) {
	if e.ResourceType != event.ResourceTypePod {
		h.logger.Debug("invalid resource type, skipped")
		return
	}

	pod := e.Object.(*apiV1.Pod)
	oldPod := e.OldObject.(*apiV1.Pod)

	if pod.Status.PodIP == "" {
		h.logger.Debugf("pod[%s] has not yet obtained a valid IP, skipped", pod.Name)
		return
	}

	if !isPodContainersReady(pod.Status.Conditions) {
		h.logger.Debugf("pod[%s] containers not ready, skipped", pod.Name)
		return
	}

	if pod.GetFinalizers() != nil {
		h.logger.Debugf("pod[%s] is about to be deleted", pod.Name)
		return
	}

	services := util.GetPodServices(pod)
	ctx, cancel := context.WithTimeout(context.Background(), h.config.Timeout*time.Second)

	if pod.GetDeletionTimestamp() != nil && oldPod.Status.Phase == apiV1.PodRunning {
		for _, s := range services {
			res, _ := h.client.Get(ctx, filepath.Join(h.config.Prefix, s.Name), clientv3.WithPrefix())
			for _, item := range res.Kvs {
				key := filepath.Join(h.config.Prefix, s.Name, strings.Replace(s.Host, ".", "-", -1))
				if string(item.Key) != key {
					continue
				}

				if _, err := h.client.Delete(ctx, key); err != nil {
					//删除时出现错误，需要联系管理员
					continue
				}
			}

			h.logger.Infof("pod[%s] - [%s] remove dns record successful", pod.Name, s.String())
		}
	} else if pod.Status.Phase == apiV1.PodRunning && !isPodContainersReady(oldPod.Status.Conditions) {
		for _, s := range services {
			key := filepath.Join(h.config.Prefix, s.Name, strings.Replace(s.Host, ".", "-", -1))

			h.client.Put(ctx, key, s.DNSRecord())
			h.logger.Infof("pod[%s] - [%s] add dns record successful", pod.Name, s.String())
		}
	} else {
		h.logger.Errorf("pod[%s] unknown event, need admin to handle", pod.Name)
	}

	cancel()
}
