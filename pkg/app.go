package pkg

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/srelab/watcher/pkg/event"

	"github.com/srelab/common/log"

	"github.com/srelab/watcher/pkg/handlers"

	appsV1Beta1 "k8s.io/api/apps/v1beta1"
	batchV1 "k8s.io/api/batch/v1"
	apiV1 "k8s.io/api/core/v1"
	extV1Beta1 "k8s.io/api/extensions/v1beta1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/srelab/watcher/pkg/controller"
	"github.com/srelab/watcher/pkg/g"
	"github.com/srelab/watcher/pkg/util"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
)

func Start() {
	var kubeClient kubernetes.Interface

	_, err := rest.InClusterConfig()
	if err != nil {
		kubeClient = util.GetClientOutOfCluster()
	} else {
		kubeClient = util.GetClient()
	}

	var defulatHandler = new(handlers.DefaultHandler)
	var saHandler = new(handlers.SAHandler)
	var gatewayHandler = new(handlers.GatewayHandler)
	var etcdHandler = new(handlers.EtcdHandler)

	// initialize all handler
	if err := defulatHandler.Init(g.Config()); err != nil {
		log.Panic("init default handler error: ", err)
	}

	if err := saHandler.Init(g.Config()); err != nil {
		log.Panic("init sa handler error: ", err)
	}

	if err := gatewayHandler.Init(g.Config()); err != nil {
		log.Panic("init gateway handler error: ", err)
	}

	if err := etcdHandler.Init(g.Config()); err != nil {
		log.Panic("init etcd handler error: ", err)
	}

	// close the etcd client
	// follow-up can implement public methods for the Close()
	defer etcdHandler.Close()

	if g.Config().Resource.Pod {
		informer := cache.NewSharedIndexInformer(
			&cache.ListWatch{
				ListFunc: func(options metaV1.ListOptions) (runtime.Object, error) {
					return kubeClient.CoreV1().Pods(g.Config().Kubernetes.Namespace).List(options)
				},
				WatchFunc: func(options metaV1.ListOptions) (watch.Interface, error) {
					return kubeClient.CoreV1().Pods(g.Config().Kubernetes.Namespace).Watch(options)
				},
			},
			&apiV1.Pod{},
			0, //Skip resync
			cache.Indexers{},
		)

		stopCh := make(chan struct{})
		defer close(stopCh)

		c := controller.New(kubeClient, informer, event.ResourceTypePod, []handlers.Handler{gatewayHandler, etcdHandler, saHandler})
		go c.Run(stopCh)
	}

	if g.Config().Resource.DaemonSet {
		informer := cache.NewSharedIndexInformer(
			&cache.ListWatch{
				ListFunc: func(options metaV1.ListOptions) (runtime.Object, error) {
					return kubeClient.ExtensionsV1beta1().DaemonSets(g.Config().Kubernetes.Namespace).List(options)
				},
				WatchFunc: func(options metaV1.ListOptions) (watch.Interface, error) {
					return kubeClient.ExtensionsV1beta1().DaemonSets(g.Config().Kubernetes.Namespace).Watch(options)
				},
			},
			&extV1Beta1.DaemonSet{},
			0, //Skip resync
			cache.Indexers{},
		)

		stopCh := make(chan struct{})
		defer close(stopCh)

		c := controller.New(kubeClient, informer, event.ResourceTypeDaemonSet, []handlers.Handler{})
		go c.Run(stopCh)
	}

	if g.Config().Resource.ReplicaSet {
		informer := cache.NewSharedIndexInformer(
			&cache.ListWatch{
				ListFunc: func(options metaV1.ListOptions) (runtime.Object, error) {
					return kubeClient.ExtensionsV1beta1().ReplicaSets(g.Config().Kubernetes.Namespace).List(options)
				},
				WatchFunc: func(options metaV1.ListOptions) (watch.Interface, error) {
					return kubeClient.ExtensionsV1beta1().ReplicaSets(g.Config().Kubernetes.Namespace).Watch(options)
				},
			},
			&extV1Beta1.ReplicaSet{},
			0, //Skip resync
			cache.Indexers{},
		)

		stopCh := make(chan struct{})
		defer close(stopCh)

		c := controller.New(kubeClient, informer, event.ResourceTypeReplicaSet, []handlers.Handler{})
		go c.Run(stopCh)
	}

	if g.Config().Resource.Services {
		informer := cache.NewSharedIndexInformer(
			&cache.ListWatch{
				ListFunc: func(options metaV1.ListOptions) (runtime.Object, error) {
					return kubeClient.CoreV1().Services(g.Config().Kubernetes.Namespace).List(options)
				},
				WatchFunc: func(options metaV1.ListOptions) (watch.Interface, error) {
					return kubeClient.CoreV1().Services(g.Config().Kubernetes.Namespace).Watch(options)
				},
			},
			&apiV1.Service{},
			0, //Skip resync
			cache.Indexers{},
		)

		stopCh := make(chan struct{})
		defer close(stopCh)

		c := controller.New(kubeClient, informer, event.ResourceTypeService, []handlers.Handler{})
		go c.Run(stopCh)
	}

	if g.Config().Resource.Deployment {
		informer := cache.NewSharedIndexInformer(
			&cache.ListWatch{
				ListFunc: func(options metaV1.ListOptions) (runtime.Object, error) {
					return kubeClient.AppsV1beta1().Deployments(g.Config().Kubernetes.Namespace).List(options)
				},
				WatchFunc: func(options metaV1.ListOptions) (watch.Interface, error) {
					return kubeClient.AppsV1beta1().Deployments(g.Config().Kubernetes.Namespace).Watch(options)
				},
			},
			&appsV1Beta1.Deployment{},
			0, //Skip resync
			cache.Indexers{},
		)

		stopCh := make(chan struct{})
		defer close(stopCh)

		c := controller.New(kubeClient, informer, event.ResourceTypeDeployment, []handlers.Handler{})
		go c.Run(stopCh)
	}

	if g.Config().Resource.Namespace {
		informer := cache.NewSharedIndexInformer(
			&cache.ListWatch{
				ListFunc: func(options metaV1.ListOptions) (runtime.Object, error) {
					return kubeClient.CoreV1().Namespaces().List(options)
				},
				WatchFunc: func(options metaV1.ListOptions) (watch.Interface, error) {
					return kubeClient.CoreV1().Namespaces().Watch(options)
				},
			},
			&apiV1.Namespace{},
			0, //Skip resync
			cache.Indexers{},
		)

		stopCh := make(chan struct{})
		defer close(stopCh)

		c := controller.New(kubeClient, informer, event.ResourceTypeNamespace, []handlers.Handler{})
		go c.Run(stopCh)
	}

	if g.Config().Resource.ReplicationController {
		informer := cache.NewSharedIndexInformer(
			&cache.ListWatch{
				ListFunc: func(options metaV1.ListOptions) (runtime.Object, error) {
					return kubeClient.CoreV1().ReplicationControllers(g.Config().Kubernetes.Namespace).List(options)
				},
				WatchFunc: func(options metaV1.ListOptions) (watch.Interface, error) {
					return kubeClient.CoreV1().ReplicationControllers(g.Config().Kubernetes.Namespace).Watch(options)
				},
			},
			&apiV1.ReplicationController{},
			0, //Skip resync
			cache.Indexers{},
		)

		stopCh := make(chan struct{})
		defer close(stopCh)

		c := controller.New(kubeClient, informer, event.ResourceTypeReplicationController, []handlers.Handler{})
		go c.Run(stopCh)
	}

	if g.Config().Resource.Job {
		informer := cache.NewSharedIndexInformer(
			&cache.ListWatch{
				ListFunc: func(options metaV1.ListOptions) (runtime.Object, error) {
					return kubeClient.BatchV1().Jobs(g.Config().Kubernetes.Namespace).List(options)
				},
				WatchFunc: func(options metaV1.ListOptions) (watch.Interface, error) {
					return kubeClient.BatchV1().Jobs(g.Config().Kubernetes.Namespace).Watch(options)
				},
			},
			&batchV1.Job{},
			0, //Skip resync
			cache.Indexers{},
		)

		stopCh := make(chan struct{})
		defer close(stopCh)

		c := controller.New(kubeClient, informer, event.ResourceTypeJob, []handlers.Handler{})
		go c.Run(stopCh)
	}

	if g.Config().Resource.PersistentVolume {
		informer := cache.NewSharedIndexInformer(
			&cache.ListWatch{
				ListFunc: func(options metaV1.ListOptions) (runtime.Object, error) {
					return kubeClient.CoreV1().PersistentVolumes().List(options)
				},
				WatchFunc: func(options metaV1.ListOptions) (watch.Interface, error) {
					return kubeClient.CoreV1().PersistentVolumes().Watch(options)
				},
			},
			&apiV1.PersistentVolume{},
			0, //Skip resync
			cache.Indexers{},
		)

		stopCh := make(chan struct{})
		defer close(stopCh)

		c := controller.New(kubeClient, informer, event.ResourceTypePersistentVolume, []handlers.Handler{})
		go c.Run(stopCh)
	}

	if g.Config().Resource.Secret {
		informer := cache.NewSharedIndexInformer(
			&cache.ListWatch{
				ListFunc: func(options metaV1.ListOptions) (runtime.Object, error) {
					return kubeClient.CoreV1().Secrets(g.Config().Kubernetes.Namespace).List(options)
				},
				WatchFunc: func(options metaV1.ListOptions) (watch.Interface, error) {
					return kubeClient.CoreV1().Secrets(g.Config().Kubernetes.Namespace).Watch(options)
				},
			},
			&apiV1.Secret{},
			0, //Skip resync
			cache.Indexers{},
		)

		stopCh := make(chan struct{})
		defer close(stopCh)

		c := controller.New(kubeClient, informer, event.ResourceTypeSecret, []handlers.Handler{})
		go c.Run(stopCh)
	}

	if g.Config().Resource.ConfigMap {
		informer := cache.NewSharedIndexInformer(
			&cache.ListWatch{
				ListFunc: func(options metaV1.ListOptions) (runtime.Object, error) {
					return kubeClient.CoreV1().ConfigMaps(g.Config().Kubernetes.Namespace).List(options)
				},
				WatchFunc: func(options metaV1.ListOptions) (watch.Interface, error) {
					return kubeClient.CoreV1().ConfigMaps(g.Config().Kubernetes.Namespace).Watch(options)
				},
			},
			&apiV1.ConfigMap{},
			0, //Skip resync
			cache.Indexers{},
		)

		stopCh := make(chan struct{})
		defer close(stopCh)

		c := controller.New(kubeClient, informer, event.ResourceTypeConfigMap, []handlers.Handler{})
		go c.Run(stopCh)
	}

	if g.Config().Resource.Ingress {
		informer := cache.NewSharedIndexInformer(
			&cache.ListWatch{
				ListFunc: func(options metaV1.ListOptions) (runtime.Object, error) {
					return kubeClient.ExtensionsV1beta1().Ingresses(g.Config().Kubernetes.Namespace).List(options)
				},
				WatchFunc: func(options metaV1.ListOptions) (watch.Interface, error) {
					return kubeClient.ExtensionsV1beta1().Ingresses(g.Config().Kubernetes.Namespace).Watch(options)
				},
			},
			&extV1Beta1.Ingress{},
			0, //Skip resync
			cache.Indexers{},
		)

		stopCh := make(chan struct{})
		defer close(stopCh)

		c := controller.New(kubeClient, informer, event.ResourceTypeIngress, []handlers.Handler{})
		go c.Run(stopCh)
	}

	// Open the built-in handler interface as http
	engine := handlers.NewServerEngine()
	go engine.Start(g.Config().Http.GetListenAddr())

	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, syscall.SIGTERM)
	signal.Notify(sigterm, syscall.SIGINT)
	<-sigterm

	// Wait for signal to gracefully shutdown the server with a timeout of 10 seconds.
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := engine.Shutdown(ctx); err != nil {
		log.Fatal(err)
	}
}
