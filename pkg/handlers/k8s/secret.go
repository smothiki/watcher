package k8s

import (
	"net/http"

	"github.com/labstack/echo"
	"github.com/srelab/watcher/pkg/handlers/shared"

	coreV1 "k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// List objects of kind Secret
// The control list content can be filtered by the options payload.
func (h *Handler) getSecret(ctx echo.Context) error {
	var p = new(getOptionsPayload)
	if err := ctx.Bind(p); err != nil {
		return shared.Responder{Status: http.StatusBadRequest, Success: false, Msg: err}.JSON(ctx)
	}

	secrets, err := h.handlers.kube.CoreV1().Secrets(ctx.Param("ns")).List(metaV1.ListOptions{
		FieldSelector: p.FieldSelector,
		LabelSelector: p.LabelSelector,
		Continue:      p.Continue,
		Limit:         p.Limit,
	})

	if err != nil {
		return shared.Responder{Status: http.StatusInternalServerError, Success: false, Msg: err}.JSON(ctx)
	}

	return shared.Responder{Status: http.StatusOK, Success: true, Result: secrets}.JSON(ctx)
}

// Create a Secret
// The request needs to include a legal secret configuration file, except that it is in json format.
func (h *Handler) createSecret(ctx echo.Context) error {
	secret := new(coreV1.Secret)
	if err := ctx.Bind(secret); err != nil {
		return shared.Responder{Status: http.StatusBadRequest, Success: false, Msg: err}.JSON(ctx)
	}

	secret, err := h.handlers.kube.CoreV1().Secrets(ctx.Param("ns")).Create(secret)
	if err != nil {
		return shared.Responder{Status: http.StatusInternalServerError, Success: false, Msg: err}.JSON(ctx)
	}

	return shared.Responder{Status: http.StatusOK, Success: true, Result: secret}.JSON(ctx)
}

func (h *Handler) updateSecret(ctx echo.Context) error {
	secret := new(coreV1.Secret)
	if err := ctx.Bind(secret); err != nil {
		return shared.Responder{Status: http.StatusBadRequest, Success: false, Msg: err}.JSON(ctx)
	}

	secret, err := h.handlers.kube.CoreV1().Secrets(ctx.Param("ns")).Update(secret)
	if err != nil {
		return shared.Responder{Status: http.StatusInternalServerError, Success: false, Msg: err}.JSON(ctx)
	}

	return shared.Responder{Status: http.StatusOK, Success: true, Result: secret}.JSON(ctx)
}

func (h *Handler) deleteSecret(ctx echo.Context) error {
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

	err := h.handlers.kube.CoreV1().Secrets(ctx.Param("ns")).Delete(p.Name, &metaV1.DeleteOptions{
		PropagationPolicy:  &deletePolicy,
		GracePeriodSeconds: p.GracePeriodSeconds,
	})

	if err != nil {
		return shared.Responder{Status: http.StatusInternalServerError, Success: false, Msg: err}.JSON(ctx)
	}

	return shared.Responder{Status: http.StatusOK, Success: true}.JSON(ctx)
}
