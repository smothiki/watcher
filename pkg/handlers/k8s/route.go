package k8s

import (
	"net/http"

	"github.com/labstack/echo"
	"github.com/srelab/watcher/pkg/handlers/shared"
)

func (h *Handler) AddRoutes(group *echo.Group) {
	group.GET("", h.getName)

	nodeGroup := group.Group("/nodes")
	nodeGroup.GET("", h.getNode)

	nsGroup := group.Group("/namespaces")
	nsGroup.GET("", h.getNamespace)
	nsGroup.POST("", h.createNamespace)
	nsGroup.PUT("", h.updateNamespace)
	nsGroup.DELETE("", h.deleteNamespace)

	nsPodGroup := nsGroup.Group("/:ns/pods")
	nsPodGroup.GET("", h.getPod)
	nsPodGroup.POST("", h.createPod)
	nsPodGroup.PUT("", h.updatePod)
	nsPodGroup.DELETE("", h.deletePod)

	nsSecretGroup := nsGroup.Group("/:ns/secrets")
	nsSecretGroup.GET("", h.getSecret)
	nsSecretGroup.POST("", h.createSecret)
	nsSecretGroup.PUT("", h.updateSecret)
	nsSecretGroup.DELETE("", h.deleteSecret)

	nsDaemonsetGroup := nsGroup.Group("/:ns/daemonsets")
	nsDaemonsetGroup.GET("", h.getDaemonset)
	nsDaemonsetGroup.POST("", h.createDaemonSet)
	nsDaemonsetGroup.PUT("", h.updateDaemonset)
	nsDaemonsetGroup.DELETE("", h.deleteDaemonset)

	nsDeploymentGroup := nsGroup.Group("/:ns/deployments")
	nsDeploymentGroup.GET("", h.getDeployment)
	nsDeploymentGroup.POST("", h.createDeployment)
	nsDeploymentGroup.PUT("", h.updateDeployment)
	nsDeploymentGroup.DELETE("", h.deleteDeployment)

	nsDeploymentGroup.GET("/:name/scale", h.getDeploymentScale)
	nsDeploymentGroup.PUT("/:name/scale/:replicas", h.updateDeploymentScale)
}

func (h *Handler) getName(ctx echo.Context) error {
	return shared.Responder{Status: http.StatusOK, Success: true, Result: h.Name()}.JSON(ctx)
}
