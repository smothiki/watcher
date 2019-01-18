package shared

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	appsV1 "k8s.io/api/apps/v1"
	batchV1 "k8s.io/api/batch/v1"
	apiV1 "k8s.io/api/core/v1"
	extV1Beta1 "k8s.io/api/extensions/v1beta1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/go-playground/validator"
	"github.com/srelab/common/log"
	"github.com/srelab/common/slice"

	"github.com/labstack/echo"
	"github.com/srelab/watcher/pkg/g"
)

// Handler is implemented by any handler.
// The Handle method is used to process event
type Handler interface {
	Name() string

	Init(config *g.Configuration) error
	Created(event *Event)
	Deleted(event *Event)
	Updated(event *Event)
}

// Service struct is used to describe a service
type Service struct {
	Name   string `validate:"required" json:"-"`
	Host   string `validate:"required,ipv4" json:"host"`
	Port   int    `validate:"required,min=1,max=65535" json:"port"`
	Type   string `validate:"required" json:"type,omitempty"`
	HcURL  string `validate:"-" json:"hc_path,omitempty"`
	HcPort int    `validate:"required,min=1,max=65535" json:"hc_port,omitempty"`
	// FATHER LEVEL DOMAIN NAME
	FLDName string `validate:"-" json:"-"`
}

// Return a string consisting of host and port
func (s *Service) String() string {
	return fmt.Sprintf("%s:%d", s.Host, s.Port)
}

// Return a legal coredns parsing record
func (s *Service) DNSRecord() string {
	return fmt.Sprintf(`{"host":"%s"}`, s.Host)
}

// Return the name of Dns, which consists of FLD and name
func (s *Service) DNSName() string {
	if s.FLDName == "" {
		return s.Name
	}

	return s.FLDName + "/" + s.Name
}

// Responder in order to unify the returned response structure
type Responder struct {
	Status     int         `json:"-"`
	Success    bool        `json:"success"`
	Result     interface{} `json:"result,omitempty"`
	Msg        interface{} `json:"msg"`
	Pagination interface{} `json:"pagination,omitempty"`
}

func (r Responder) JSON(ctx echo.Context) error {
	if r.Msg == "" || r.Msg == nil {
		r.Msg = http.StatusText(r.Status)
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
func (event *Event) GetPodServices(pod *apiV1.Pod) ([]*Service, error) {
	if event.ResourceType != ResourceTypePod {
		return nil, errors.New("invalid resource type, skipped")
	}

	if pod.Status.PodIP == "" {
		return nil, fmt.Errorf("pod[%s] has not yet obtained a valid IP, skipped", pod.Name)
	}

	if !IsPodContainersReady(pod.Status.Conditions) {
		return nil, fmt.Errorf("pod[%s] containers not ready, skipped", pod.Name)
	}

	if pod.GetFinalizers() != nil {
		return nil, fmt.Errorf("pod[%s] is about to be deleted", pod.Name)
	}

	requireKeys := []string{
		"SERVICE_NAME", "SERVICE_PORT", "SERVICE_TYPE",
		"HEALTH_CHECK_URL", "HEALTH_CHECK_PORT", "DNS_FLD_NAME",
	}

	services := make([]*Service, 0)
	for _, container := range pod.Spec.Containers {
		service := new(Service)
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
			case "SERVICE_TYPE":
				service.Type = env.Value
			case "DNS_FLD_NAME":
				service.FLDName = env.Value
			case "HEALTH_CHECK_URL":
				service.HcURL = env.Value
			case "HEALTH_CHECK_PORT":
				port, err := strconv.Atoi(env.Value)
				if err != nil {
					continue
				}

				service.HcPort = port
			}
		}

		if err := validator.New().Struct(service); err != nil {
			if service.Name == "" {
				service.Name = container.Name
			}

			log.With("shared", "event").Info("container[%s] variable is invalid", service.Name)
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

func IsPodContainersReady(conditions []apiV1.PodCondition) bool {
	for _, condition := range conditions {
		if condition.Type != "ContainersReady" {
			continue
		}

		if condition.Status == apiV1.ConditionTrue {
			return true
		}
	}

	return false
}

//
//// New create new KubewatchEvent
//func New(obj interface{}, action string) Event {
//	var namespace, kind, component, host, reason, status, name string
//
//	objectMeta := util.GetObjectMetaData(obj)
//	namespace = objectMeta.Namespace
//	name = objectMeta.Name
//	reason = action
//	status = m[action]
//
//	switch object := obj.(type) {
//	case *extV1Beta1.DaemonSet:
//		kind = "daemon set"
//	case *appsV1Beta1.Deployment:
//		kind = "deployment"
//	case *batchV1.Job:
//		kind = "job"
//	case *apiV1.Namespace:
//		kind = "namespace"
//	case *extV1Beta1.Ingress:
//		kind = "ingress"
//	case *apiV1.PersistentVolume:
//		kind = "persistent volume"
//	case *apiV1.Pod:
//		kind = "pod"
//		host = object.Spec.NodeName
//	case *apiV1.ReplicationController:
//		kind = "replication controller"
//	case *extV1Beta1.ReplicaSet:
//		kind = "replica set"
//	case *apiV1.Service:
//		kind = "service"
//		component = string(object.Spec.Type)
//	case *apiV1.Secret:
//		kind = "secret"
//	case *apiV1.ConfigMap:
//		kind = "configmap"
//	case Event:
//		name = object.Name
//		kind = object.Kind
//		namespace = object.Namespace
//	}
//
//	kbEvent := Event{
//		Namespace: namespace,
//		Kind:      kind,
//		Component: component,
//		Host:      host,
//		Reason:    reason,
//		Status:    status,
//		Name:      name,
//	}
//	return kbEvent
//}
//
//// Message returns event message in standard format.
//// included as a part of event packege to enhance code resuablity across handlers.
//func (e *Event) Message() (msg string) {
//	// using switch over if..else, since the format could vary based on the kind of the object in future.
//	switch e.Kind {
//	case "namespace":
//		msg = fmt.Sprintf(
//			"A namespace `%s` has been `%s`",
//			e.Name,
//			e.Reason,
//		)
//	default:
//		msg = fmt.Sprintf(
//			"A `%s` in namespace `%s` has been `%s`:\n`%s`",
//			e.Kind,
//			e.Namespace,
//			e.Reason,
//			e.Name,
//		)
//	}
//	return msg
//}
