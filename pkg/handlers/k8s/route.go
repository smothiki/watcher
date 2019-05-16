package k8s

import (
	"net/http"

	"github.com/labstack/echo"
	"github.com/srelab/watcher/pkg/handlers/shared"
)

func (h *Handler) AddRoutes(group *echo.Group) {
	group.GET(shared.EmptyPath, h.getName)

	nodeGroup := group.Group("/nodes")
	nodeGroup.GET(shared.EmptyPath, h.getNode)

	nsGroup := group.Group("/namespaces")
	nsGroup.GET(shared.EmptyPath, h.getNamespace)
	nsGroup.POST(shared.EmptyPath, h.createNamespace)
	nsGroup.PUT(shared.EmptyPath, h.updateNamespace)
	nsGroup.DELETE(shared.EmptyPath, h.deleteNamespace)

	nsPodGroup := nsGroup.Group("/:ns/pods")
	nsPodGroup.GET(shared.EmptyPath, h.getPod)
	nsPodGroup.POST(shared.EmptyPath, h.createPod)
	nsPodGroup.PUT(shared.EmptyPath, h.updatePod)
	nsPodGroup.DELETE(shared.EmptyPath, h.deletePod)
	nsPodGroup.GET("/:name/logs", h.getPodLogs)

	nsEventGroup := nsGroup.Group("/:ns/events")
	nsEventGroup.GET(shared.EmptyPath, h.getEvent)
	nsEventGroup.POST(shared.EmptyPath, h.createEvent)
	nsEventGroup.PUT(shared.EmptyPath, h.updateEvent)
	nsEventGroup.DELETE(shared.EmptyPath, h.deleteEvent)

	nsSecretGroup := nsGroup.Group("/:ns/secrets")
	nsSecretGroup.GET(shared.EmptyPath, h.getSecret)
	nsSecretGroup.POST(shared.EmptyPath, h.createSecret)
	nsSecretGroup.PUT(shared.EmptyPath, h.updateSecret)
	nsSecretGroup.DELETE(shared.EmptyPath, h.deleteSecret)

	nsDaemonsetGroup := nsGroup.Group("/:ns/daemonsets")
	nsDaemonsetGroup.GET(shared.EmptyPath, h.getDaemonset)
	nsDaemonsetGroup.POST(shared.EmptyPath, h.createDaemonSet)
	nsDaemonsetGroup.PUT(shared.EmptyPath, h.updateDaemonset)
	nsDaemonsetGroup.DELETE(shared.EmptyPath, h.deleteDaemonset)

	nsDeploymentGroup := nsGroup.Group("/:ns/deployments")
	nsDeploymentGroup.GET(shared.EmptyPath, h.getDeployment)
	nsDeploymentGroup.POST(shared.EmptyPath, h.createDeployment)
	nsDeploymentGroup.PUT(shared.EmptyPath, h.updateDeployment)
	nsDeploymentGroup.DELETE(shared.EmptyPath, h.deleteDeployment)

	nsDeploymentGroup.GET("/:name/scale", h.getDeploymentScale)
	nsDeploymentGroup.PUT("/:name/scale/:replicas", h.updateDeploymentScale)
}

func (h *Handler) getName(ctx echo.Context) error {
	return shared.Responder{Status: http.StatusOK, Success: true, Result: h.Name()}.JSON(ctx)
}
