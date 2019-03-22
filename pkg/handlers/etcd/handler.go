package etcd

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/coreos/etcd/pkg/transport"
	"github.com/pkg/errors"
	"github.com/srelab/common/log"
	"github.com/srelab/watcher/pkg/g"
	"github.com/srelab/watcher/pkg/handlers/shared"
	"go.etcd.io/etcd/clientv3"
	"go.etcd.io/etcd/etcdserver/api/v3rpc/rpctypes"
	apiV1 "k8s.io/api/core/v1"
)

type Handler struct {
	config *g.EtcdConfig

	client *clientv3.Client
	logger log.Logger
}

func (h *Handler) Name() string            { return "etcd" }
func (h *Handler) Handler() *Handler       { return h }
func (h *Handler) RoutePrefix() string     { return "/" + h.Name() }
func (h *Handler) DNSPrefix() string       { return h.config.DNSPrefix }
func (h *Handler) Close()                  { h.client.Close() }
func (h *Handler) Created(e *shared.Event) {}
func (h *Handler) Updated(e *shared.Event) {}

// Remove DNS resolution records from etcd when the pod is detected to be destroyed
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

// Initialize the Etcd client and log
func (h *Handler) Init(config *g.Configuration, objs ...interface{}) error {
	h.config = config.Handlers.EtcdConfig
	h.config.DNSPrefix = strings.TrimRight(h.config.DNSPrefix, "/")

	// simply judge whether prefix starts with "/" character
	if !strings.HasPrefix(h.config.DNSPrefix, "/") {
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

// Get key Val list from etcd
// key: Any string that returns empty when the matching key is not queried
// keysOnly: Return only key, no return value
// prefix: Match key based on the prefix
// limit: Limit the number of returns
func (h *Handler) GetKey(key string, keysOnly, prefix bool, limit int64) (*clientv3.GetResponse, error) {
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

	return res, h.eErrorHandling(err)
}

// Write Key Val to etcd
// key: Any string that returns empty when the matching key is not queried
// val: Only accept json string values
// ttl: key expire
func (h *Handler) PutKey(key, val string, ttl int64) (*clientv3.PutResponse, error) {
	if ttl > 0 {
		ctx, err := h.client.Grant(context.TODO(), ttl)
		if err != nil {
			return nil, h.eErrorHandling(err)
		}

		res, err := h.client.Put(context.TODO(), key, val, clientv3.WithLease(ctx.ID))
		return res, h.eErrorHandling(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), h.config.Timeout*time.Second)

	res, err := h.client.Put(ctx, key, val)
	cancel()

	return res, h.eErrorHandling(err)
}

// Delete Key Val from etcd
func (h *Handler) DeleteKey(key string, prefix bool) (*clientv3.DeleteResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), h.config.Timeout*time.Second)

	var options []clientv3.OpOption
	if prefix {
		options = append(options, clientv3.WithPrefix())
	}

	res, err := h.client.Delete(ctx, key, options...)
	cancel()

	return res, h.eErrorHandling(err)
}

// Formatting errors returned by etcd
func (h *Handler) eErrorHandling(err error) error {
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

// Remove DNS resolution records from CoreDNS
func (h *Handler) DeleteService(service *shared.ServicePayload) error {
	res, err := h.GetKey(
		filepath.Join(h.DNSPrefix(), service.DNSName()),
		false,
		true,
		0,
	)

	if err != nil {
		return fmt.Errorf("get key error: %s", err)
	}

	for _, item := range res.Kvs {
		key := filepath.Join(h.DNSPrefix(), service.DNSName(), service.DNSKey())
		if string(item.Key) != key {
			continue
		}

		if _, err := h.DeleteKey(key, false); err != nil {
			return fmt.Errorf("etcd key cannot be delete: %s", err)
		}
	}

	h.logger.Infof("[etcd][%s] - [%s] delete successful", service.Name, service.String())
	return nil
}

// Convert service to DNS resolution record and write to etcd for use by CoreDNS
func (h *Handler) CreateService(service *shared.ServicePayload) error {
	key := filepath.Join(h.DNSPrefix(), service.DNSName(), service.DNSKey())
	if _, err := h.PutKey(key, service.DNSRecord(), 0); err != nil {
		return fmt.Errorf("etcd key cannot be create: %s", err)
	}

	h.logger.Infof("[etcd][%s] - [%s] create successful", service.Name, service.String())
	return nil
}
