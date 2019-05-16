package k8s

import (
	"net/http"

	"github.com/labstack/echo"
	"github.com/srelab/watcher/pkg/handlers/shared"
	coreV1 "k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// List objects of kind event
// The control list content can be filtered by the options payload.
func (h *Handler) getEvent(ctx echo.Context) error {
	var p = new(GetOptionsPayload)
	if err := ctx.Bind(p); err != nil {
		return shared.Responder{Status: http.StatusBadRequest, Success: false, Msg: err}.JSON(ctx)
	}

	events, err := h.handlers.kube.CoreV1().Events(ctx.Param("ns")).List(metaV1.ListOptions{
		FieldSelector: p.FieldSelector,
		LabelSelector: p.LabelSelector,
		Continue:      p.Continue,
		Limit:         p.Limit,
	})

	if err != nil {
		return shared.Responder{Status: http.StatusInternalServerError, Success: false, Msg: err}.JSON(ctx)
	}

	return shared.Responder{Status: http.StatusOK, Success: true, Result: events}.JSON(ctx)
}

// Create a Event
// The request needs to include a legal event configuration file, except that it is in json format.
func (h *Handler) createEvent(ctx echo.Context) error {
	event := new(coreV1.Event)
	if err := ctx.Bind(event); err != nil {
		return shared.Responder{Status: http.StatusBadRequest, Success: false, Msg: err}.JSON(ctx)
	}

	event, err := h.handlers.kube.CoreV1().Events(ctx.Param("ns")).Create(event)

	if err != nil {
		return shared.Responder{Status: http.StatusInternalServerError, Success: false, Msg: err}.JSON(ctx)
	}

	return shared.Responder{Status: http.StatusOK, Success: true, Result: event}.JSON(ctx)
}

// Update a Event
// The request needs to include a legal Event configuration file, except that it is in json format.
func (h *Handler) updateEvent(ctx echo.Context) error {
	event := new(coreV1.Pod)
	if err := ctx.Bind(event); err != nil {
		return shared.Responder{Status: http.StatusBadRequest, Success: false, Msg: err}.JSON(ctx)
	}

	event, err := h.handlers.kube.CoreV1().Pods(ctx.Param("ns")).Update(event)

	if err != nil {
		return shared.Responder{Status: http.StatusInternalServerError, Success: false, Msg: err}.JSON(ctx)
	}

	return shared.Responder{Status: http.StatusOK, Success: true, Result: event}.JSON(ctx)
}

// Delete a Event
// It is deleted by the name of the Event.
// The interface supports delete_policy to control the deletion mechanism.
// Unless you know how to use it, you should use the default value.
func (h *Handler) deleteEvent(ctx echo.Context) error {
	p := new(DeleteOptionsPayload)
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

	err := h.handlers.kube.CoreV1().Events(ctx.Param("ns")).Delete(p.Name, &metaV1.DeleteOptions{
		PropagationPolicy:  &deletePolicy,
		GracePeriodSeconds: p.GracePeriodSeconds,
	})

	if err != nil {
		return shared.Responder{Status: http.StatusInternalServerError, Success: false, Msg: err}.JSON(ctx)
	}

	return shared.Responder{Status: http.StatusOK, Success: true}.JSON(ctx)
}
