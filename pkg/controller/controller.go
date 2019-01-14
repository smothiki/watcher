package controller

import (
	"fmt"
	"time"

	"github.com/srelab/watcher/pkg/handlers"

	utilruntime "k8s.io/apimachinery/pkg/util/runtime"

	"github.com/srelab/common/log"
	"github.com/srelab/watcher/pkg/event"
	"github.com/srelab/watcher/pkg/util"

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
	handlers  []handlers.Handler
}

func New(client kubernetes.Interface, informer cache.SharedIndexInformer, resourceType event.ResourceType, handlers []handlers.Handler) *Controller {
	var e event.Event
	var err error

	queue := workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())
	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		// 当资源第一次加入到Informer的缓存后调用
		AddFunc: func(object interface{}) {
			e.Key, err = cache.MetaNamespaceKeyFunc(object)
			e.Action = "create"
			e.ResourceType = resourceType
			e.Namespace = util.GetObjectMetaData(object).Namespace
			e.Object = object

			if err == nil {
				queue.Add(e)
			}
		},

		// 当既有资源被修改时调用。oldObj 是资源的上一个状态，newObj 则是新状态
		// resync 时此方法也被调用，即使对象没有任何变化
		UpdateFunc: func(oldObject, object interface{}) {
			e.Key, err = cache.MetaNamespaceKeyFunc(oldObject)
			e.Action = "update"
			e.ResourceType = resourceType
			e.Namespace = util.GetObjectMetaData(object).Namespace
			e.Object = object
			e.OldObject = oldObject

			if err == nil {
				queue.Add(e)
			}
		},

		// 当既有资源被删除时调用，obj是对象的最后状态，如果最后状态未知则返回DeletedFinalStateUnknown
		DeleteFunc: func(object interface{}) {
			e.Key, err = cache.DeletionHandlingMetaNamespaceKeyFunc(object)
			e.Action = "delete"
			e.ResourceType = resourceType
			e.Namespace = util.GetObjectMetaData(object).Namespace
			e.Object = object

			if err == nil {
				queue.Add(e)
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
	e, quit := c.queue.Get()

	if quit {
		return false
	}
	defer c.queue.Done(e)

	err := c.processItem(e.(event.Event))
	if err == nil {
		// No error, reset the ratelimit counters
		c.queue.Forget(e)
	} else if c.queue.NumRequeues(e) < maxRetries {
		log.Errorf("error processing %s (will retry): %v", e.(event.Event).Key, err)
		c.queue.AddRateLimited(e)
	} else {
		// err != nil and too many retries
		log.Errorf("error processing %s (giving up): %v", e.(event.Event).Key, err)
		c.queue.Forget(e)
		utilruntime.HandleError(err)
	}

	return true
}

/* TODOs
- Enhance event creation using client-side cacheing machanisms - pending
- Enhance the processItem to classify events - done
- Send alerts correspoding to events - done
*/

func (c *Controller) processItem(e event.Event) error {
	obj, _, err := c.informer.GetIndexer().GetByKey(e.Key)
	if err != nil {
		return fmt.Errorf("error fetching object with key %s from store: %v", e.Key, err)
	}

	// get object's metedata
	objectMeta := util.GetObjectMetaData(obj)

	// process events based on its type
	switch e.Action {
	case "create":
		// compare CreationTimestamp and serverStartTime and alert only on latest events
		// Could be Replaced by using Delta or DeltaFIFO
		if objectMeta.CreationTimestamp.Sub(serverStartTime).Seconds() > 0 {
			for _, handler := range c.handlers {
				handler.Created(e)
			}
			return nil
		}
	case "update":
		for _, handler := range c.handlers {
			handler.Updated(e)
		}
		return nil
	case "delete":
		for _, handler := range c.handlers {
			handler.Deleted(e)
		}
		return nil
	}
	return nil
}
