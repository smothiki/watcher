package controller

import (
	"fmt"
	"time"

	"github.com/srelab/watcher/pkg/handlers/shared"

	utilruntime "k8s.io/apimachinery/pkg/util/runtime"

	"github.com/srelab/common/log"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

const maxRetries = 5

var serverStartTime time.Time

// Controller object
type Controller struct {
	logger    *log.Logger
	clientset kubernetes.Interface
	queue     workqueue.RateLimitingInterface
	informer  cache.SharedIndexInformer
	handlers  []shared.Handler
}

func New(client kubernetes.Interface, informer cache.SharedIndexInformer, resourceType shared.ResourceType, handlers []shared.Handler) *Controller {
	var event shared.Event
	var err error

	queue := workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())
	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		// 当资源第一次加入到 Informer 的缓存后调用
		AddFunc: func(object interface{}) {
			event.Key, err = cache.MetaNamespaceKeyFunc(object)
			event.Action = "create"
			event.ResourceType = resourceType
			event.Namespace = event.GetObjectMetaData().Namespace
			event.Object = object

			if err == nil {
				queue.Add(event)
			}
		},

		// 当既有资源被修改时调用。oldObj 是资源的上一个状态，newObj 则是新状态
		// resync 时此方法也被调用，即使对象没有任何变化
		UpdateFunc: func(oldObject, object interface{}) {
			event.Key, err = cache.MetaNamespaceKeyFunc(oldObject)
			event.Action = "update"
			event.ResourceType = resourceType
			event.Namespace = event.GetObjectMetaData().Namespace
			event.Object = object
			event.OldObject = oldObject

			if err == nil {
				queue.Add(event)
			}
		},

		// 当既有资源被删除时调用，obj是对象的最后状态，如果最后状态未知则返回 DeletedFinalStateUnknown
		DeleteFunc: func(object interface{}) {
			event.Key, err = cache.DeletionHandlingMetaNamespaceKeyFunc(object)
			event.Action = "delete"
			event.ResourceType = resourceType
			event.Namespace = event.GetObjectMetaData().Namespace
			event.Object = object

			if err == nil {
				queue.Add(event)
			}
		},
	})

	return &Controller{
		clientset: client,
		informer:  informer,
		queue:     queue,
		handlers:  handlers,
	}
}

// Run starts the watch controller
func (c *Controller) Run(stopCh <-chan struct{}) {
	defer utilruntime.HandleCrash()
	defer c.queue.ShutDown()

	log.Info("starting watch controller")
	serverStartTime = time.Now().Local()

	go c.informer.Run(stopCh)

	if !cache.WaitForCacheSync(stopCh, c.HasSynced) {
		utilruntime.HandleError(fmt.Errorf("timed out waiting for caches to sync"))
		return
	}

	log.Info("watch controller synced and ready")
	wait.Until(c.runWorker, time.Second, stopCh)
}

// HasSynced is required for the cache.Controller interface.
func (c *Controller) HasSynced() bool {
	return c.informer.HasSynced()
}

// LastSyncResourceVersion is required for the cache.Controller interface.
func (c *Controller) LastSyncResourceVersion() string {
	return c.informer.LastSyncResourceVersion()
}

func (c *Controller) runWorker() {
	for c.processNextItem() {
		// continue looping
	}
}

func (c *Controller) processNextItem() bool {
	item, quit := c.queue.Get()

	if quit {
		return false
	}
	defer c.queue.Done(item)

	// Convert the item obtained by queue to event
	event := item.(shared.Event)

	// Give the event to each handler
	err := c.processItem(&event)
	if err == nil {
		// No error, reset the ratelimit counters
		c.queue.Forget(event)
	} else if c.queue.NumRequeues(event) < maxRetries {
		log.Errorf("error processing %s (will retry): %v", event.Key, err)
		c.queue.AddRateLimited(event)
	} else {
		// err != nil and too many retries
		log.Errorf("error processing %s (giving up): %v", event.Key, err)
		c.queue.Forget(event)
		utilruntime.HandleError(err)
	}

	return true
}

func (c *Controller) processItem(event *shared.Event) error {
	//obj, _, err := c.informer.GetIndexer().GetByKey(event.Key)
	//if err != nil {
	//	return fmt.Errorf("error fetching object with key %s from store: %v", event.Key, err)
	//}

	// get object's metedata
	objectMeta := event.GetObjectMetaData()

	// process events based on its type
	switch event.Action {
	case "create":
		// compare CreationTimestamp and serverStartTime and alert only on latest events
		// Could be Replaced by using Delta or DeltaFIFO
		if objectMeta.CreationTimestamp.Sub(serverStartTime).Seconds() > 0 {
			for _, handler := range c.handlers {
				handler.Created(event)
			}
			return nil
		}
	case "update":
		for _, handler := range c.handlers {
			handler.Updated(event)
		}
		return nil
	case "delete":
		for _, handler := range c.handlers {
			handler.Deleted(event)
		}
		return nil
	}
	return nil
}
