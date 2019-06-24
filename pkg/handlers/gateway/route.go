package gateway

import (
	"fmt"
	"net/http"
	"strconv"

	"git.srelab.cn/go/resty"

	"github.com/labstack/echo"
	"github.com/srelab/watcher/pkg/handlers/shared"
)

func (h *Handler) AddRoutes(group *echo.Group) {
	group.GET(shared.EmptyPath, h.getName)
	group.GET("/namespaces", h.getNamespaces)
	group.GET("/namespaces/:namespace/upstreams", h.getUpstreams)
	group.GET("/namespaces/:namespace/upstreams/:upstream", h.getUpstreamsByName)
	group.POST("/namespaces/:namespace/upstreams/:upstream/register", h.registerServerToUpstream)
	group.POST("/namespaces/:namespace/upstreams/:upstream/unregister", h.unregisterServerFromUpstream)
}

func (h *Handler) getName(ctx echo.Context) error {
	return shared.Responder{Status: http.StatusOK, Success: true, Result: h.Name()}.JSON(ctx)
}

func (h *Handler) getNamespaces(ctx echo.Context) error {
	namespaces := make([]string, 0)
	for _, config := range h.configs {
		namespaces = append(namespaces, config.Namespace)
	}

	return shared.Responder{Status: http.StatusOK, Success: true, Result: namespaces}.JSON(ctx)
}

func (h *Handler) getUpstreams(ctx echo.Context) error {
	namespace := ctx.Param("namespace")
	url := h.URL(namespace, "/upstreams")
	if url == "" {
		return fmt.Errorf("namespace `%s` has no associated gateway config", namespace)
	}

	result := &SliceResult{}
	response, err := h.Request().SetResult(result).Get(url)
	if err != nil {
		return shared.Responder{Status: http.StatusInternalServerError, Success: false, Msg: err}.JSON(ctx)
	}

	if response.StatusCode() != http.StatusOK {
		err = fmt.Errorf("failed to get upstream list, status code[%d]: %s", response.StatusCode(), response.Body())
		return shared.Responder{Status: http.StatusInternalServerError, Success: false, Msg: err}.JSON(ctx)
	}

	return shared.Responder{Status: http.StatusOK, Success: true, Result: result.Data}.JSON(ctx)
}

func (h *Handler) getUpstreamsByName(ctx echo.Context) error {
	namespace := ctx.Param("namespace")
	upstream := ctx.Param("upstream")

	url := h.URL(namespace, fmt.Sprintf("/upstreams/%s", upstream))
	if url == "" {
		return fmt.Errorf("namespace `%s` has no associated gateway config", namespace)
	}

	result := &MapResult{}
	response, err := h.Request().SetResult(result).Get(url)
	if err != nil {
		return shared.Responder{Status: http.StatusInternalServerError, Success: false, Msg: err}.JSON(ctx)
	}

	if response.StatusCode() != http.StatusOK {
		err = fmt.Errorf("failed to get upstream, status code[%d]: %s", response.StatusCode(), response.Body())
		return shared.Responder{Status: http.StatusInternalServerError, Success: false, Msg: err}.JSON(ctx)
	}

	return shared.Responder{Status: http.StatusOK, Success: true, Result: result.Data}.JSON(ctx)
}

func (h *Handler) registerServerToUpstream(ctx echo.Context) error {
	var (
		p        RegisterServicePayload
		response *resty.Response
		err      error
	)

	namespace := ctx.Param("namespace")
	upstream := ctx.Param("upstream")

	// Get the URL of the handler in memory, when the `namespace` does not exist, skip the service
	url := h.URL(namespace, fmt.Sprintf("/upstreams/%s/register", upstream))
	if url == "" {
		err = fmt.Errorf("namespace `%s` has no associated gateway config, %s register skipped", namespace, upstream)
		return shared.Responder{Status: http.StatusBadRequest, Success: false, Msg: err}.JSON(ctx)
	}

	if err := ctx.Bind(&p); err != nil {
		return shared.Responder{Status: http.StatusBadRequest, Success: false, Msg: err}.JSON(ctx)
	}

	if p.Type == "http" {
		response, err = h.Request().SetBody(map[string]string{
			"host":    p.Host,
			"type":    p.Type,
			"port":    strconv.Itoa(p.Port),
			"hc_path": p.HcPath,
			"hc_port": strconv.Itoa(p.Port),
		}).Post(url)
	} else {
		response, err = h.Request().SetBody(map[string]string{
			"host": p.Host,
			"type": "general",
			"port": strconv.Itoa(p.Port),
		}).Post(url)
	}

	if err != nil {
		return shared.Responder{Status: http.StatusInternalServerError, Success: false, Msg: err}.JSON(ctx)
	}

	if response.StatusCode() != http.StatusOK {
		err = fmt.Errorf("failed to register upstream, status code[%d]: %s", response.StatusCode(), response.Body())
		return shared.Responder{Status: http.StatusInternalServerError, Success: false, Msg: err}.JSON(ctx)
	}

	return shared.Responder{Status: http.StatusOK, Success: true}.JSON(ctx)
}

func (h *Handler) unregisterServerFromUpstream(ctx echo.Context) error {
	var (
		p        UnRegisterServicePayload
		response *resty.Response
		err      error
	)

	namespace := ctx.Param("namespace")
	upstream := ctx.Param("upstream")

	// Get the URL of the handler in memory, when the `namespace` does not exist, skip the service
	url := h.URL(namespace, fmt.Sprintf("/upstreams/%s/unregister", upstream))
	if url == "" {
		err = fmt.Errorf("namespace `%s` has no associated gateway config, %s register skipped", namespace, upstream)
		return shared.Responder{Status: http.StatusBadRequest, Success: false, Msg: err}.JSON(ctx)
	}

	if err = ctx.Bind(&p); err != nil {
		return shared.Responder{Status: http.StatusBadRequest, Success: false, Msg: err}.JSON(ctx)
	}

	response, err = h.Request().SetBody(p).Post(url)
	if err != nil {
		return shared.Responder{Status: http.StatusInternalServerError, Success: false, Msg: err}.JSON(ctx)
	}

	if response.StatusCode() != http.StatusOK {
		err = fmt.Errorf("failed to unregister upstream, status code[%d]: %s", response.StatusCode(), response.Body())
		return shared.Responder{Status: http.StatusInternalServerError, Success: false, Msg: err}.JSON(ctx)
	}

	return shared.Responder{Status: http.StatusOK, Success: true}.JSON(ctx)
}
