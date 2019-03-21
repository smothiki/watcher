package shared

import (
	"errors"
	"fmt"
	"net/http"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/go-playground/validator"
	"github.com/srelab/common/log"
	"github.com/srelab/common/slice"

	appsV1 "k8s.io/api/apps/v1"
	batchV1 "k8s.io/api/batch/v1"
	apiV1 "k8s.io/api/core/v1"
	extV1Beta1 "k8s.io/api/extensions/v1beta1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/labstack/echo"
	"github.com/srelab/watcher/pkg/g"
)

// humanized default time format
const defaultTimeFormat = "2006-01-02 15:04:05"

// custom time types
// Used to format time into a human-readable string
type Datetime struct {
	time.Time
}

func (d *Datetime) UnmarshalJSON(b []byte) (err error) {
	str := strings.Trim(string(b), "\"")
	if str == "null" || str == "" {
		d.Time = time.Time{}
		return
	}

	cst, _ := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		return fmt.Errorf("time.LoadLocation error: %s", err.Error())
	}

	d.Time, err = time.ParseInLocation(defaultTimeFormat, str, cst)
	if err != nil {
		// When time cannot be resolved using the default format, try RFC3339Nano
		if d.Time, err = time.ParseInLocation(time.RFC3339Nano, str, cst); err == nil {
			d.Time = d.Time.In(cst)
		}
	}

	return
}

func (d *Datetime) MarshalJSON() ([]byte, error) {
	if d.Time.UnixNano() == (time.Time{}).UnixNano() {
		return []byte("null"), nil
	}

	return []byte(fmt.Sprintf("\"%s\"", d.Time.Format(defaultTimeFormat))), nil
}

// empty path of the route
const EmptyPath = ""

// Handler is implemented by any handler.
// The Handle method is used to process event
type Handler interface {
	Name() string
	RoutePrefix() string

	Init(config *g.Configuration, handlers ...interface{}) error
	Created(event *Event)
	Deleted(event *Event)
	Updated(event *Event)
}

// Responder in order to unify the returned response structure
type Responder struct {
	Status     int         `json:"-"`
	Success    bool        `json:"success"`
	Result     interface{} `json:"result,omitempty"`
	Msg        interface{} `json:"msg"`
	Pagination interface{} `json:"pagination,omitempty"`
}

// sends a JSON response with status code.
func (r Responder) JSON(ctx echo.Context) error {
	if r.Msg == "" || r.Msg == nil {
		r.Msg = http.StatusText(r.Status)
	}

	if err, ok := r.Msg.(error); ok {
		r.Msg = err.Error()
	}

	return ctx.JSON(r.Status, r)
}

// Used to determine resource type
type ResourceType string

const (
	ResourceTypePod                   ResourceType = "Pod"
	ResourceTypeDaemonSet             ResourceType = "DaemonSet"
	ResourceTypeReplicaSet            ResourceType = "ReplicaSet"
	ResourceTypeService               ResourceType = "Service"
	ResourceTypeDeployment            ResourceType = "Deployment"
	ResourceTypeNamespace             ResourceType = "Namespace"
	ResourceTypeReplicationController ResourceType = "ReplicationController"
	ResourceTypeJob                   ResourceType = "Job"
	ResourceTypePersistentVolume      ResourceType = "PersistentVolume"
	ResourceTypeSecret                ResourceType = "Secret"
	ResourceTypeConfigMap             ResourceType = "ConfigMap"
	ResourceTypeIngress               ResourceType = "Ingress"
)

// Event indicate the informerEvent
type Event struct {
	Object       interface{}
	OldObject    interface{}
	Key          string
	Action       string
	Namespace    string
	ResourceType ResourceType
}

// Return a set of services from the pod's Containers
// Return an empty slice when Containers is empty
func (event *Event) GetPodServices(pod *apiV1.Pod) ([]*ServicePayload, error) {
	if event.ResourceType != ResourceTypePod {
		return nil, errors.New("invalid resource type, skipped")
	}

	if pod.Status.PodIP == "" {
		return nil, fmt.Errorf("pod[%s] has not yet obtained a valid IP, skipped", pod.Name)
	}

	requireKeys := []string{
		"SERVICE_NAME", "SERVICE_PORT", "SERVICE_PROTOCOL_TYPE",
		"DNS_FL_DOMAIN", "HEALTH_CHECK_URL", "HEALTH_CHECK_PORT",
	}

	services := make([]*ServicePayload, 0)
	for _, container := range pod.Spec.Containers {
		service := new(ServicePayload)
		if len(container.Env) < 5 {
			continue
		}

		service.Host = pod.Status.PodIP
		for _, env := range container.Env {
			if !slice.ContainsString(requireKeys, env.Name) {
				continue
			}

			switch env.Name {
			case "SERVICE_NAME":
				service.Name = env.Value
			case "SERVICE_PORT":
				port, err := strconv.Atoi(env.Value)
				if err != nil {
					continue
				}

				service.Port = port
			case "SERVICE_PROTOCOL_TYPE":
				service.Protocol = env.Value
			case "DNS_FL_DOMAIN":
				service.FLDomain = env.Value
				if service.FLDomain == "-" {
					service.FLDomain = ""
				}
			case "HEALTH_CHECK_URL":
				service.HealthCheck.Path = env.Value
			case "HEALTH_CHECK_PORT":
				port, err := strconv.Atoi(env.Value)
				if err != nil {
					continue
				}

				service.HealthCheck.Port = port
			}
		}

		service.Namespace = event.Namespace
		if err := validator.New().Struct(service); err != nil {
			if service.Name == "" {
				service.Name = container.Name
			}

			log.With("shared", "event").Infof("container[%s] variable is invalid: %s", service.Name, err)
		} else {
			services = append(services, service)
		}
	}

	return services, nil
}

// GetObjectMetaData returns metadata of a given k8s object
func (event *Event) GetObjectMetaData() metaV1.ObjectMeta {
	var objectMeta metaV1.ObjectMeta

	switch object := event.Object.(type) {
	case *appsV1.Deployment:
		objectMeta = object.ObjectMeta
	case *apiV1.ReplicationController:
		objectMeta = object.ObjectMeta
	case *appsV1.ReplicaSet:
		objectMeta = object.ObjectMeta
	case *appsV1.DaemonSet:
		objectMeta = object.ObjectMeta
	case *apiV1.Service:
		objectMeta = object.ObjectMeta
	case *apiV1.Pod:
		objectMeta = object.ObjectMeta
	case *batchV1.Job:
		objectMeta = object.ObjectMeta
	case *apiV1.PersistentVolume:
		objectMeta = object.ObjectMeta
	case *apiV1.Namespace:
		objectMeta = object.ObjectMeta
	case *apiV1.Secret:
		objectMeta = object.ObjectMeta
	case *extV1Beta1.Ingress:
		objectMeta = object.ObjectMeta
	}

	return objectMeta
}

// Message returns event message in standard format.
// included as a part of event packege to enhance code resuablity across handlers.
func (event *Event) Message() (msg string) {
	var kind string

	objectMeta := event.GetObjectMetaData()
	switch event.Object.(type) {
	case *extV1Beta1.DaemonSet:
		kind = "daemon set"
	case *appsV1.Deployment:
		kind = "deployment"
	case *batchV1.Job:
		kind = "job"
	case *apiV1.Namespace:
		kind = "namespace"
	case *extV1Beta1.Ingress:
		kind = "ingress"
	case *apiV1.PersistentVolume:
		kind = "persistent volume"
	case *apiV1.Pod:
		kind = "pod"
	case *apiV1.ReplicationController:
		kind = "replication controller"
	case *extV1Beta1.ReplicaSet:
		kind = "replica set"
	case *apiV1.Service:
		kind = "service"
	case *apiV1.Secret:
		kind = "secret"
	case *apiV1.ConfigMap:
		kind = "configmap"
	}

	switch kind {
	case "namespace":
		msg = fmt.Sprintf(
			"Kubernetes 集群事件\n"+
				"事件类别: namespace\n"+
				"事件描述: %s has been %s\n",
			objectMeta.Name,
			event.Action,
		)
	default:
		msg = fmt.Sprintf(
			"Kubernetes 集群事件\n"+
				"事件类别: %s\n"+
				"命名空间: %s\n"+
				"事件描述: %s has been %s\n",
			kind,
			event.Namespace,
			objectMeta.Name,
			event.Action,
		)
	}

	return msg
}

func (event *Event) CacheKey() string {
	return path.Join("/watcher/handlers/etcd/", event.Key, event.Action)
}
