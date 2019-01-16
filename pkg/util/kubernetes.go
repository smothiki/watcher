package util

import (
	"fmt"
	"os"
	"strconv"

	"github.com/go-playground/validator"
	"github.com/srelab/common/slice"

	"github.com/srelab/common/log"

	"github.com/srelab/watcher/pkg/g"

	appsV1 "k8s.io/api/apps/v1"
	batchV1 "k8s.io/api/batch/v1"
	apiV1 "k8s.io/api/core/v1"
	extV1Beta1 "k8s.io/api/extensions/v1beta1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type Service struct {
	Name   string `validate:"required" json:"-"`
	Host   string `validate:"required,ipv4" json:"host"`
	Port   int    `validate:"required,min=1,max=65535" json:"port"`
	Type   string `validate:"required" json:"type,omitempty"`
	HcURL  string `validate:"-" json:"hc_path,omitempty"`
	HcPort int    `validate:"required,min=1,max=65535" json:"hc_port,omitempty"`
}

func (s *Service) String() string {
	return fmt.Sprintf("%s:%d", s.Host, s.Port)
}

func (s *Service) DNSRecord() string {
	return fmt.Sprintf(`{"host":"%s"}`, s.Host)
}

func GetPodServices(pod *apiV1.Pod) (services []*Service) {
	requireKeys := []string{"SERVICE_NAME", "SERVICE_PORT", "SERVICE_TYPE", "HEALTH_CHECK_URL", "HEALTH_CHECK_PORT"}

	for _, container := range pod.Spec.Containers {
		s := new(Service)
		if len(container.Env) < 5 {
			continue
		}

		s.Host = pod.Status.PodIP
		for _, e := range container.Env {
			if !slice.ContainsString(requireKeys, e.Name) {
				continue
			}

			switch e.Name {
			case "SERVICE_NAME":
				s.Name = e.Value
			case "SERVICE_PORT":
				port, err := strconv.Atoi(e.Value)
				if err != nil {
					continue
				}

				s.Port = port
			case "SERVICE_TYPE":
				s.Type = e.Value
			case "HEALTH_CHECK_URL":
				s.HcURL = e.Value
			case "HEALTH_CHECK_PORT":
				port, err := strconv.Atoi(e.Value)
				if err != nil {
					continue
				}

				s.HcPort = port
			}
		}

		if err := validator.New().Struct(s); err != nil {
			if s.Name == "" {
				s.Name = container.Name
			}

			log.With("util", "kubernetes").Info("container[%s] variable is invalid", s.Name)
		} else {
			services = append(services, s)
		}
	}

	return
}

// GetClient returns a k8s clientset to the request from inside of cluster
func GetClient() kubernetes.Interface {
	config, err := rest.InClusterConfig()
	if err != nil {
		log.Fatalf("can not get watch config: %v", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalf("can not create watch client: %v", err)
	}

	return clientset
}

func buildOutOfClusterConfig() (*rest.Config, error) {
	kubeconfigPath := g.Config().Kubernetes.Config
	if kubeconfigPath == "" {
		kubeconfigPath = os.Getenv("HOME") + "/.kube/config"
	}

	return clientcmd.BuildConfigFromFlags("", kubeconfigPath)
}

// GetClientOutOfCluster returns a k8s clientset to the request from outside of cluster
func GetClientOutOfCluster() kubernetes.Interface {
	config, err := buildOutOfClusterConfig()
	if err != nil {
		log.Fatalf("Can not get kubernetes config: %v", err)
	}

	clientset, err := kubernetes.NewForConfig(config)

	return clientset
}

// GetObjectMetaData returns metadata of a given k8s object
func GetObjectMetaData(obj interface{}) metaV1.ObjectMeta {
	var objectMeta metaV1.ObjectMeta

	switch object := obj.(type) {
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
