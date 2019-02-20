package k8s

import (
	"net/http"

	"github.com/labstack/echo"
	"github.com/srelab/watcher/pkg/handlers/shared"

	appsV1 "k8s.io/api/apps/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// List objects of kind DaemonSet
// The control list content can be filtered by the options payload.
func (h *Handler) getDaemonset(ctx echo.Context) error {
	var p = new(getOptionsPayload)
	if err := ctx.Bind(p); err != nil {
		return shared.Responder{Status: http.StatusBadRequest, Success: false, Msg: err}.JSON(ctx)
	}

	daemonsets, err := h.k8s.client.AppsV1().DaemonSets(ctx.Param("ns")).List(metaV1.ListOptions{
		FieldSelector: p.FieldSelector,
		LabelSelector: p.LabelSelector,
		Continue:      p.Continue,
		Limit:         p.Limit,
	})

	if err != nil {
		return shared.Responder{Status: http.StatusInternalServerError, Success: false, Msg: err}.JSON(ctx)
	}

	return shared.Responder{Status: http.StatusOK, Success: true, Result: daemonsets}.JSON(ctx)
}

// Create a DaemonSet
// The request needs to include a legal daemonset configuration file, except that it is in json format.
func (h *Handler) createDaemonSet(ctx echo.Context) error {
	daemonset := new(appsV1.DaemonSet)
	if err := ctx.Bind(daemonset); err != nil {
		return shared.Responder{Status: http.StatusBadRequest, Success: false, Msg: err}.JSON(ctx)
	}

	daemonset, err := h.k8s.client.AppsV1().DaemonSets(ctx.Param("ns")).Create(daemonset)
	if err != nil {
		return shared.Responder{Status: http.StatusInternalServerError, Success: false, Msg: err}.JSON(ctx)
	}

	return shared.Responder{Status: http.StatusOK, Success: true, Result: daemonset}.JSON(ctx)
}

// Update a DaemonSet
// The request needs to include a legal daemonset configuration file, except that it is in json format.
func (h *Handler) updateDaemonset(ctx echo.Context) error {
	daemonset := new(appsV1.DaemonSet)
	if err := ctx.Bind(daemonset); err != nil {
		return shared.Responder{Status: http.StatusBadRequest, Success: false, Msg: err}.JSON(ctx)
	}

	daemonset, err := h.k8s.client.AppsV1().DaemonSets(ctx.Param("ns")).Update(daemonset)
	if err != nil {
		return shared.Responder{Status: http.StatusInternalServerError, Success: false, Msg: err}.JSON(ctx)
	}

	return shared.Responder{Status: http.StatusOK, Success: true, Result: daemonset}.JSON(ctx)
}

// Delete a DaemonSet
// It is deleted by the name of the daemonset.
// The interface supports delete_policy to control the deletion mechanism.
// Unless you know how to use it, you should use the default value.
func (h *Handler) deleteDaemonset(ctx echo.Context) error {
	p := new(deleteOptionsPayload)
	if err := ctx.Bind(p); err != nil {
		return shared.Responder{Status: http.StatusBadRequest, Success: false, Msg: err}.JSON(ctx)
	}

	var deletePolicy metaV1.DeletionPropagation
	switch p.DeletePolicy {
	case "orphan":
		deletePolicy = metaV1.DeletePropagationOrphan
	case "background":
		deletePolicy = metaV1.DeletePropagationBackground
	case "foreground":
		deletePolicy = metaV1.DeletePropagationForeground
	default:
		deletePolicy = metaV1.DeletePropagationForeground
	}

	err := h.k8s.client.AppsV1().DaemonSets(ctx.Param("ns")).Delete(p.Name, &metaV1.DeleteOptions{
		PropagationPolicy:  &deletePolicy,
		GracePeriodSeconds: p.GracePeriodSeconds,
	})

	if err != nil {
		return shared.Responder{Status: http.StatusInternalServerError, Success: false, Msg: err}.JSON(ctx)
	}

	return shared.Responder{Status: http.StatusOK, Success: true}.JSON(ctx)
}
