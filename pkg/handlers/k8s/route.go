package k8s

import (
	"net/http"

	"github.com/labstack/echo"
	"github.com/srelab/watcher/pkg/handlers/shared"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (h *Handler) AddRoutes(group *echo.Group) {
	group.GET("", h.getName)

	nsGroup := group.Group("/namespaces")
	nsGroup.GET("", h.getNamespaces)

	nsPodGroup := group.Group("/:ns/pods")
	nsPodGroup.GET("", h.getPods)

	nsDeploymentGroup := nsGroup.Group("/:ns/deployments")
	nsDeploymentGroup.GET("", h.getDeployment)
	nsDeploymentGroup.POST("", h.createDeployment)
	nsDeploymentGroup.DELETE("", h.deleteDeployment)

	nsDeploymentGroup.GET("/:name/scale", h.getDeploymentScale)
	nsDeploymentGroup.PUT("/:name/scale/:replicas", h.updateDeploymentScale)
}

func (h *Handler) getName(ctx echo.Context) error {
	return shared.Responder{Status: http.StatusOK, Success: true, Result: h.Name()}.JSON(ctx)
}

func (h *Handler) getNamespaces(ctx echo.Context) error {
	namespaces, err := h.k8s.client.CoreV1().Namespaces().List(metaV1.ListOptions{})
	if err != nil {
		return shared.Responder{Status: http.StatusInternalServerError, Success: false, Msg: err}.JSON(ctx)
	}

	return shared.Responder{Status: http.StatusOK, Success: true, Result: namespaces}.JSON(ctx)
}
