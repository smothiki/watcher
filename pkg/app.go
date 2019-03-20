package pkg

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/srelab/watcher/pkg/handlers/harbor"

	"github.com/labstack/echo"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/srelab/common/log"
	"github.com/srelab/watcher/pkg/controller"
	"github.com/srelab/watcher/pkg/g"
	"github.com/srelab/watcher/pkg/handlers"
	"github.com/srelab/watcher/pkg/handlers/core"
	"github.com/srelab/watcher/pkg/handlers/etcd"
	"github.com/srelab/watcher/pkg/handlers/gateway"
	"github.com/srelab/watcher/pkg/handlers/k8s"
	"github.com/srelab/watcher/pkg/handlers/sa"
	"github.com/srelab/watcher/pkg/handlers/shared"

	appsV1 "k8s.io/api/apps/v1"
	batchV1 "k8s.io/api/batch/v1"
	apiV1 "k8s.io/api/core/v1"
	extV1Beta1 "k8s.io/api/extensions/v1beta1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
)

func Start() {
	var kubeClient kubernetes.Interface

	_, err := rest.InClusterConfig()
	if err != nil {
		kubeClient = GetClientOutOfCluster()
	} else {
		kubeClient = GetClient()
	}

	var gatewayHandler = new(gateway.Handler)
	var etcdHandler = new(etcd.Handler)
	var k8sHandler = new(k8s.Handler)
	var harborHandler = new(harbor.Handler)

	var saHandler = new(sa.Handler)
	var coreHandler = new(core.Handler)

	// initialize all handler
	if err := k8sHandler.Init(g.Config(), kubeClient); err != nil {
		log.Panic("init default handler error: ", err)
	}

	if err := gatewayHandler.Init(g.Config()); err != nil {
		log.Panic("init gateway handler error: ", err)
	}

	if err := etcdHandler.Init(g.Config()); err != nil {
		log.Panic("init etcd handler error: ", err)
	}

	if err := harborHandler.Init(g.Config()); err != nil {
		log.Panic("init sa handler error: ", err)
	}

	if err := saHandler.Init(g.Config(), etcdHandler); err != nil {
		log.Panic("init sa handler error: ", err)
	}

	if err := coreHandler.Init(g.Config(), etcdHandler, gatewayHandler); err != nil {
		log.Panic("init core handler error: ", err)
	}

	// close the etcd client
	// follow-up can implement public methods for the Close()
	defer etcdHandler.Close()

	informerHandlers := []shared.Handler{etcdHandler, gatewayHandler, saHandler}
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

		c := controller.New(kubeClient, informer, shared.ResourceTypePod, informerHandlers)
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

		c := controller.New(kubeClient, informer, shared.ResourceTypeDaemonSet, informerHandlers)
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

		c := controller.New(kubeClient, informer, shared.ResourceTypeReplicaSet, informerHandlers)
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

		c := controller.New(kubeClient, informer, shared.ResourceTypeService, informerHandlers)
		go c.Run(stopCh)
	}

	if g.Config().Resource.Deployment {
		informer := cache.NewSharedIndexInformer(
			&cache.ListWatch{
				ListFunc: func(options metaV1.ListOptions) (runtime.Object, error) {
					return kubeClient.AppsV1().Deployments(g.Config().Kubernetes.Namespace).List(options)
				},
				WatchFunc: func(options metaV1.ListOptions) (watch.Interface, error) {
					return kubeClient.AppsV1().Deployments(g.Config().Kubernetes.Namespace).Watch(options)
				},
			},
			&appsV1.Deployment{},
			0, //Skip resync
			cache.Indexers{},
		)

		stopCh := make(chan struct{})
		defer close(stopCh)

		c := controller.New(kubeClient, informer, shared.ResourceTypeDeployment, informerHandlers)
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

		c := controller.New(kubeClient, informer, shared.ResourceTypeNamespace, informerHandlers)
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

		c := controller.New(kubeClient, informer, shared.ResourceTypeReplicationController, informerHandlers)
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

		c := controller.New(kubeClient, informer, shared.ResourceTypeJob, informerHandlers)
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

		c := controller.New(kubeClient, informer, shared.ResourceTypePersistentVolume, informerHandlers)
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

		c := controller.New(kubeClient, informer, shared.ResourceTypeSecret, informerHandlers)
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

		c := controller.New(kubeClient, informer, shared.ResourceTypeConfigMap, informerHandlers)
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

		c := controller.New(kubeClient, informer, shared.ResourceTypeIngress, informerHandlers)
		go c.Run(stopCh)
	}

	// Open the built-in handler interface as http
	engine := handlers.NewHandlersEngine()
	engine.Use(handlers.NewMetric())
	engine.GET("/metrics", echo.WrapHandler(promhttp.Handler()))

	// Selectively add routes when some handler need to expose the interface
	handlersRoute := engine.Group("/handlers")
	gatewayHandler.AddRoutes(handlersRoute.Group(gatewayHandler.RoutePrefix()))
	etcdHandler.AddRoutes(handlersRoute.Group(etcdHandler.RoutePrefix()))
	harborHandler.AddRoutes(handlersRoute.Group(harborHandler.RoutePrefix()))
	k8sHandler.AddRoutes(handlersRoute.Group(k8sHandler.RoutePrefix()))
	coreHandler.AddRoutes(handlersRoute.Group(coreHandler.RoutePrefix()))

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
