package etcd

import (
	"context"
	"path/filepath"
	"strings"
	"time"

	"github.com/srelab/watcher/pkg/handlers/shared"

	"github.com/pkg/errors"
	apiV1 "k8s.io/api/core/v1"

	"github.com/coreos/etcd/pkg/transport"
	"github.com/srelab/common/log"
	"github.com/srelab/watcher/pkg/g"
	"go.etcd.io/etcd/clientv3"
	"go.etcd.io/etcd/etcdserver/api/v3rpc/rpctypes"
)

type Handler struct {
	config *g.EtcdConfig
	client *clientv3.Client
	logger log.Logger
}

func (h *Handler) Name() string {
	return "etcd"
}

func (h *Handler) RoutePrefix() string {
	return "/" + h.Name()
}

func (h *Handler) Init(config *g.Configuration) error {
	h.config = config.Handlers.EtcdConfig
	h.config.Prefix = strings.TrimRight(h.config.Prefix, "/")

	// simply judge whether prefix starts with "/" character
	if !strings.HasPrefix(h.config.Prefix, "/") {
		return errors.New("invalid coredns prefix")
	}

	tlsInfo := transport.TLSInfo{
		CertFile:      config.Handlers.EtcdConfig.CertFile,
		KeyFile:       config.Handlers.EtcdConfig.KeyFile,
		TrustedCAFile: config.Handlers.EtcdConfig.CAFile,
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

func (h *Handler) Close() {
	h.client.Close()
}

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
		for _, service := range services {
			res, err := h.eGet(filepath.Join(h.config.Prefix, service.DNSName()), false, true, 0)
			if err != nil {
				h.logger.Info("get error", err)
				// 获取 Key 出现错误，需要联系管理员
				continue
			}

			for _, item := range res.Kvs {
				key := filepath.Join(h.config.Prefix, service.DNSName(), strings.Replace(service.Host, ".", "-", -1))
				if string(item.Key) != key {
					continue
				}

				if _, err := h.eDelete(key, false); err != nil {
					h.logger.Info("delete error", err)
					continue
				}
			}

			h.logger.Infof("pod[%s] - [%s] remove dns record successful", pod.Name, service.String())
		}
	} else if pod.Status.Phase == apiV1.PodRunning && !shared.IsPodContainersReady(oldPod.Status.Conditions) {
		for _, service := range services {
			key := filepath.Join(h.config.Prefix, service.DNSName(), strings.Replace(service.Host, ".", "-", -1))
			if _, err := h.ePut(key, service.DNSRecord()); err != nil {
				//添加时出现错误，需要联系管理员
				h.logger.Info("add error", err)
				continue
			}

			h.logger.Infof("pod[%s] - [%s] add dns record successful", pod.Name, service.String())
		}
	} else {
		h.logger.Errorf("pod[%s] unknown event, need admin to handle", pod.Name)
	}
}

func (h *Handler) eGet(key string, keysOnly, prefix bool, limit int64) (*clientv3.GetResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), h.config.Timeout*time.Second)

	var options []clientv3.OpOption
	if prefix {
		options = append(options, clientv3.WithPrefix())
	}

	if keysOnly {
		options = append(options, clientv3.WithKeysOnly())
	}

	options = append(options, clientv3.WithLimit(limit))

	res, err := h.client.Get(ctx, key, options...)
	cancel()

	return res, h.eErrorHandling(ctx, err)
}

func (h *Handler) ePut(key, val string) (*clientv3.PutResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), h.config.Timeout*time.Second)

	res, err := h.client.Put(ctx, key, val)
	cancel()

	return res, h.eErrorHandling(ctx, err)
}

func (h *Handler) eDelete(key string, prefix bool) (*clientv3.DeleteResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), h.config.Timeout*time.Second)

	var options []clientv3.OpOption
	if prefix {
		options = append(options, clientv3.WithPrefix())
	}

	res, err := h.client.Delete(ctx, key, options...)
	cancel()

	return res, h.eErrorHandling(ctx, err)
}

func (h *Handler) eErrorHandling(ctx context.Context, err error) error {
	if err != nil {
		switch err {
		case context.Canceled:
			err = errors.Wrap(err, "ctx is canceled by another routine")
		case context.DeadlineExceeded:
			err = errors.Wrap(err, "ctx is attached with a deadline is exceeded")
		case rpctypes.ErrEmptyKey:
			err = errors.Wrap(err, "client-side error")
		default:
			err = errors.Wrap(err, "bad cluster endpoints, which are not etcd servers")
		}
	}

	return err
}
