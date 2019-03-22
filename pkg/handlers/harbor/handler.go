package harbor

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"git.srelab.cn/go/resty"
	"github.com/srelab/common/log"
	"github.com/srelab/watcher/pkg/g"
	"github.com/srelab/watcher/pkg/handlers/shared"
)

var sid string

type Handler struct {
	config *g.HarborConfig
	logger log.Logger
}

func (h *Handler) Name() string            { return "harbor" }
func (h *Handler) Handler() *Handler       { return h }
func (h *Handler) RoutePrefix() string     { return "/" + h.Name() }
func (h *Handler) Close()                  {}
func (h *Handler) Created(e *shared.Event) {}
func (h *Handler) Deleted(e *shared.Event) {}
func (h *Handler) Updated(e *shared.Event) {}

func (h *Handler) Init(config *g.Configuration, objs ...interface{}) error {
	h.config = config.Handlers.HarborConfig
	h.logger = log.With("handlers", h.Name())

	_ = h.Request()
	if sid == "" {
		return errors.New("authentication failed, please check the configuration or contact the administrator")
	}

	return nil
}

// Return request client
// Ensure that sid is valid by requesting the /api/users/current interface
// and attempting to log in when the request fails
func (h *Handler) Request() *resty.Request {
	r := resty.New().SetRetryCount(3).SetRetryWaitTime(5 * time.Second).SetRetryMaxWaitTime(10 * time.Second)
	r.SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true}).SetCookie(&http.Cookie{
		Name: "sid", Value: sid, HttpOnly: true,
	}).SetHostURL(h.config.Endpoint)

	res, err := r.R().Get("/api/users/current")
	if res.StatusCode() != http.StatusOK || err != nil {
		// form data
		fd := map[string]string{"principal": h.config.Username, "password": h.config.Password}
		res, err = r.R().SetFormData(fd).Post("/c/login")

		// Check if the login is successful, and return an error if it fails.
		if res.StatusCode() == http.StatusOK {
			sid = ""
			for _, cookie := range res.Cookies() {
				if cookie.Name != "sid" {
					continue
				}

				sid = cookie.Value
				break
			}
		}
	}

	return r.R().SetHeader("Content-Type", "application/json").SetHeader("Cookie", "sid="+sid)
}

// Returns the interface address of the harbor
// path: request path
func (h *Handler) URL(path string) string {
	return fmt.Sprintf(
		"%s/%s",
		strings.TrimRight(h.config.Endpoint, "/"),
		strings.TrimLeft(path, "/"),
	)
}
