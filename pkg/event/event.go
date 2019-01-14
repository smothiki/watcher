/*
Copyright 2016 Skippbox, Ltd.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package event

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

var m = map[string]string{
	"created": "Normal",
	"deleted": "Danger",
	"updated": "Warning",
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
