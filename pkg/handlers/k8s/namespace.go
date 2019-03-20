package k8s

import (
	"net/http"

	"github.com/labstack/echo"
	"github.com/srelab/watcher/pkg/handlers/shared"

	coreV1 "k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// List objects of kind Namespace
// The control list content can be filtered by the options payload.
func (h *Handler) getNamespace(ctx echo.Context) error {
	var p = new(GetOptionsPayload)
	if err := ctx.Bind(p); err != nil {
		return shared.Responder{Status: http.StatusBadRequest, Success: false, Msg: err}.JSON(ctx)
	}

	namespaces, err := h.handlers.kube.CoreV1().Namespaces().List(metaV1.ListOptions{
		FieldSelector: p.FieldSelector,
		LabelSelector: p.LabelSelector,
		Continue:      p.Continue,
		Limit:         p.Limit,
	})

	if err != nil {
		return shared.Responder{Status: http.StatusInternalServerError, Success: false, Msg: err}.JSON(ctx)
	}

	return shared.Responder{Status: http.StatusOK, Success: true, Result: namespaces}.JSON(ctx)
}

// Create a Namespace
// The request needs to include a legal namespace configuration file, except that it is in json format.
func (h *Handler) createNamespace(ctx echo.Context) error {
	namespace := new(coreV1.Namespace)
	if err := ctx.Bind(namespace); err != nil {
		return shared.Responder{Status: http.StatusBadRequest, Success: false, Msg: err}.JSON(ctx)
	}

	namespace, err := h.handlers.kube.CoreV1().Namespaces().Create(namespace)
	if err != nil {
		return shared.Responder{Status: http.StatusInternalServerError, Success: false, Msg: err}.JSON(ctx)
	}

	return shared.Responder{Status: http.StatusOK, Success: true, Result: namespace}.JSON(ctx)
}

// Update a Namespace
// The request needs to include a legal namespace configuration file, except that it is in json format.
// When the namespace is updated, only the state can be updated. Their values are:
//     "Active" means the namespace is available for use in the system
//     "Terminating" means the namespace is undergoing graceful termination
func (h *Handler) updateNamespace(ctx echo.Context) error {
	namespace := new(coreV1.Namespace)
	if err := ctx.Bind(namespace); err != nil {
		return shared.Responder{Status: http.StatusBadRequest, Success: false, Msg: err}.JSON(ctx)
	}

	namespace, err := h.handlers.kube.CoreV1().Namespaces().Update(namespace)
	if err != nil {
		return shared.Responder{Status: http.StatusInternalServerError, Success: false, Msg: err}.JSON(ctx)
	}

	return shared.Responder{Status: http.StatusOK, Success: true, Result: namespace}.JSON(ctx)
}

// Delete a Namespace
// It is deleted by the name of the namespace.
// The interface supports delete_policy to control the deletion mechanism.
// Unless you know how to use it, you should use the default value.
func (h *Handler) deleteNamespace(ctx echo.Context) error {
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

	err := h.handlers.kube.CoreV1().Namespaces().Delete(p.Name, &metaV1.DeleteOptions{
		PropagationPolicy:  &deletePolicy,
		GracePeriodSeconds: p.GracePeriodSeconds,
	})

	if err != nil {
		return shared.Responder{Status: http.StatusInternalServerError, Success: false, Msg: err}.JSON(ctx)
	}

	return shared.Responder{Status: http.StatusOK, Success: true}.JSON(ctx)
}
