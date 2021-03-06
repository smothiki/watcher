package sa

import (
	"fmt"
	"time"

	"github.com/srelab/watcher/pkg/handlers/etcd"

	"git.srelab.cn/go/resty"

	"github.com/srelab/common/log"
	"github.com/srelab/watcher/pkg/g"
	"github.com/srelab/watcher/pkg/handlers/shared"
)

type Handler struct {
	handlers struct {
		etcd *etcd.Handler
	}

	config *g.SAConfig
	logger log.Logger
}

func (h *Handler) Name() string        { return "sa" }
func (h *Handler) RoutePrefix() string { return "/" + h.Name() }
func (h *Handler) Close()              {}

// The sa handler needs to use the etcd handler to ensure
// that messages are not sent repeatedly in a clustered environment.
func (h *Handler) Init(config *g.Configuration, itfs ...interface{}) error {
	h.config = config.Handlers.SAConfig
	h.logger = log.With("handlers", h.Name())

	for _, itf := range itfs {
		switch object := itf.(type) {
		case *etcd.Handler:
			h.handlers.etcd = object
		}
	}

	return nil
}

func (h *Handler) Created(e *shared.Event) {
	response, err := h.handlers.etcd.GetKey(e.CacheKey(), true, false, 1)
	if err == nil && response.Count > 0 {
		return
	}

	h.handlers.etcd.PutKey(e.CacheKey(), `{"success": true}`, 10)
	h.send(e.Message())
}

func (h *Handler) Deleted(e *shared.Event) {
	response, err := h.handlers.etcd.GetKey(e.CacheKey(), true, false, 1)
	if err == nil && response.Count > 0 {
		return
	}

	h.handlers.etcd.PutKey(e.CacheKey(), `{"success": true}`, 10)
	h.send(e.Message())
}

func (h *Handler) Updated(e *shared.Event) {
	response, err := h.handlers.etcd.GetKey(e.CacheKey(), true, false, 1)
	if err == nil && response.Count > 0 {
		return
	}

	h.handlers.etcd.PutKey(e.CacheKey(), `{"success": true}`, 10)
	h.send(e.Message())
}

func (h *Handler) request() *resty.Request {
	return resty.SetRetryCount(3).SetRetryWaitTime(5 * time.Second).SetRetryMaxWaitTime(20 * time.Second).R()
}

func (h *Handler) send(content string) {
	if !h.config.Notice.Enable {
		return
	}

	h.request().SetHeader("Host", "sa.wolaidai.com").
		SetHeader("Content-Type", "application/json").
		SetBasicAuth(h.config.Username, h.config.Password).
		SetBody(map[string]interface{}{
			"config": map[string]interface{}{"chat_id": h.config.Notice.ChatID, "content": content},
		}).Post(fmt.Sprintf("%s/api/tasks/wechat/push", h.config.Endpoint))
}
