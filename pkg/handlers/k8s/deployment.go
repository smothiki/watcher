package k8s

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo"
	"github.com/srelab/watcher/pkg/handlers/shared"

	appsV1 "k8s.io/api/apps/v1"
	autoscalingV1 "k8s.io/api/autoscaling/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// List objects of kind Deployment
// The control list content can be filtered by the options payload.
func (h *Handler) getDeployment(ctx echo.Context) error {
	var p = new(optionsPayload)
	if err := ctx.Bind(p); err != nil {
		return shared.Responder{Status: http.StatusBadRequest, Success: false, Msg: err}.JSON(ctx)
	}

	deployments, err := h.k8s.client.AppsV1().Deployments(ctx.Param("ns")).List(metaV1.ListOptions{
		FieldSelector: p.FieldSelector,
		LabelSelector: p.LabelSelector,
		Continue:      p.Continue,
		Limit:         p.Limit,
	})

	if err != nil {
		return shared.Responder{Status: http.StatusInternalServerError, Success: false, Msg: err}.JSON(ctx)
	}

	return shared.Responder{Status: http.StatusOK, Success: true, Result: deployments}.JSON(ctx)
}

// Create a Deployment
// The request needs to include a legal deployment configuration file, except that it is in json format.
func (h *Handler) createDeployment(ctx echo.Context) error {
	namespace := ctx.Param("namespace")

	deployment := new(appsV1.Deployment)
	if err := ctx.Bind(deployment); err != nil {
		return shared.Responder{Status: http.StatusBadRequest, Success: false, Msg: err}.JSON(ctx)
	}

	deployment, err := h.k8s.client.AppsV1().Deployments(namespace).Create(deployment)
	if err != nil {
		return shared.Responder{Status: http.StatusInternalServerError, Success: false, Msg: err}.JSON(ctx)
	}

	return shared.Responder{Status: http.StatusOK, Success: true, Result: deployment}.JSON(ctx)
}

// Update a Deployment
// The request needs to include a legal deployment configuration file, except that it is in json format.
func (h *Handler) updateDeployment(ctx echo.Context) error {
	namespace := ctx.Param("namespace")

	deployment := new(appsV1.Deployment)
	if err := ctx.Bind(deployment); err != nil {
		return shared.Responder{Status: http.StatusBadRequest, Success: false, Msg: err}.JSON(ctx)
	}

	deployment, err := h.k8s.client.AppsV1().Deployments(namespace).Update(deployment)
	if err != nil {
		return shared.Responder{Status: http.StatusInternalServerError, Success: false, Msg: err}.JSON(ctx)
	}

	return shared.Responder{Status: http.StatusOK, Success: true, Result: deployment}.JSON(ctx)
}

// Delete a Deployment
// It is deleted by the name of the deployment.
// The interface supports delete_policy to control the deletion mechanism.
// Unless you know how to use it, you should use the default value.
func (h *Handler) deleteDeployment(ctx echo.Context) error {
	namespace := ctx.Param("namespace")

	p := new(deleteDeploymentPayload)
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

	err := h.k8s.client.AppsV1().Deployments(namespace).Delete(p.Name, &metaV1.DeleteOptions{
		PropagationPolicy:  &deletePolicy,
		GracePeriodSeconds: p.GracePeriodSeconds,
	})

	if err != nil {
		return shared.Responder{Status: http.StatusInternalServerError, Success: false, Msg: err}.JSON(ctx)
	}

	return shared.Responder{Status: http.StatusOK, Success: true}.JSON(ctx)
}

// List Scale info of kind Deployment
func (h *Handler) getDeploymentScale(ctx echo.Context) error {
	name := ctx.Param("name")

	scale, err := h.k8s.client.AppsV1().Deployments(ctx.Param("ns")).GetScale(name, metaV1.GetOptions{})
	if err != nil {
		return shared.Responder{Status: http.StatusInternalServerError, Success: false, Msg: err}.JSON(ctx)
	}

	return shared.Responder{Status: http.StatusOK, Success: true, Result: scale}.JSON(ctx)
}

// Update a Scale for Deployment
// Controlled by the URL's `replicas` parameter
func (h *Handler) updateDeploymentScale(ctx echo.Context) error {
	replicas, err := strconv.ParseInt(ctx.Param("replicas"), 10, 32)
	if err != nil {
		return shared.Responder{Status: http.StatusBadRequest, Success: false, Msg: err}.JSON(ctx)
	}

	scale := &autoscalingV1.Scale{
		ObjectMeta: metaV1.ObjectMeta{
			Name:      ctx.Param("name"),
			Namespace: ctx.Param("ns"),
		},
		Spec: autoscalingV1.ScaleSpec{Replicas: int32(replicas)},
	}

	scale, err = h.k8s.client.AppsV1().Deployments(scale.ObjectMeta.Namespace).UpdateScale(scale.ObjectMeta.Name, scale)
	if err != nil {
		return shared.Responder{Status: http.StatusInternalServerError, Success: false, Msg: err}.JSON(ctx)
	}

	return shared.Responder{Status: http.StatusOK, Success: true, Result: scale}.JSON(ctx)
}
