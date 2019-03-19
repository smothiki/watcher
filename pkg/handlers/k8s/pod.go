package k8s

import (
	"net/http"

	"github.com/labstack/echo"
	"github.com/srelab/watcher/pkg/handlers/shared"
	coreV1 "k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// List objects of kind Pod
// The control list content can be filtered by the options payload.
func (h *Handler) getPod(ctx echo.Context) error {
	var p = new(getOptionsPayload)
	if err := ctx.Bind(p); err != nil {
		return shared.Responder{Status: http.StatusBadRequest, Success: false, Msg: err}.JSON(ctx)
	}

	pods, err := h.handlers.kube.CoreV1().Pods(ctx.Param("ns")).List(metaV1.ListOptions{
		FieldSelector: p.FieldSelector,
		LabelSelector: p.LabelSelector,
		Continue:      p.Continue,
		Limit:         p.Limit,
	})

	if err != nil {
		return shared.Responder{Status: http.StatusInternalServerError, Success: false, Msg: err}.JSON(ctx)
	}

	return shared.Responder{Status: http.StatusOK, Success: true, Result: pods}.JSON(ctx)
}

// Create a Pod
// The request needs to include a legal pod configuration file, except that it is in json format.
func (h *Handler) createPod(ctx echo.Context) error {
	pod := new(coreV1.Pod)
	if err := ctx.Bind(pod); err != nil {
		return shared.Responder{Status: http.StatusBadRequest, Success: false, Msg: err}.JSON(ctx)
	}

	pod, err := h.handlers.kube.CoreV1().Pods(ctx.Param("ns")).Create(pod)

	if err != nil {
		return shared.Responder{Status: http.StatusInternalServerError, Success: false, Msg: err}.JSON(ctx)
	}

	return shared.Responder{Status: http.StatusOK, Success: true, Result: pod}.JSON(ctx)
}

// Update a Pod
// The request needs to include a legal Pod configuration file, except that it is in json format.
// If the Pod needs to perform an update operation, at least ensure that the following attributes are present:
// metadata.name, metadata.resourceVersion  (the server will check for the latest) , metadata.labels
// spec.volumes, spec.containers, spec.serviceAccount, spec.nodeName
//
// It is recommended to always modify the submission based on the current latest pod config
func (h *Handler) updatePod(ctx echo.Context) error {
	pod := new(coreV1.Pod)
	if err := ctx.Bind(pod); err != nil {
		return shared.Responder{Status: http.StatusBadRequest, Success: false, Msg: err}.JSON(ctx)
	}

	pod, err := h.handlers.kube.CoreV1().Pods(ctx.Param("ns")).Update(pod)

	if err != nil {
		return shared.Responder{Status: http.StatusInternalServerError, Success: false, Msg: err}.JSON(ctx)
	}

	return shared.Responder{Status: http.StatusOK, Success: true, Result: pod}.JSON(ctx)
}

// Delete a Pod
// It is deleted by the name of the Pod.
// The interface supports delete_policy to control the deletion mechanism.
// Unless you know how to use it, you should use the default value.
func (h *Handler) deletePod(ctx echo.Context) error {
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

	err := h.handlers.kube.CoreV1().Pods(ctx.Param("ns")).Delete(p.Name, &metaV1.DeleteOptions{
		PropagationPolicy:  &deletePolicy,
		GracePeriodSeconds: p.GracePeriodSeconds,
	})

	if err != nil {
		return shared.Responder{Status: http.StatusInternalServerError, Success: false, Msg: err}.JSON(ctx)
	}

	return shared.Responder{Status: http.StatusOK, Success: true}.JSON(ctx)
}
