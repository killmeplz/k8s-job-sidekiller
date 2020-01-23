package main

import (
	"fmt"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

type Listener struct {
	informer cache.SharedIndexInformer
	queue    workqueue.RateLimitingInterface
	handler  Handler
}

func (i *Listener) processNextItem() bool {
	key, quit := i.queue.Get()
	if quit {
		return false
	}
	defer i.queue.Done(key)
	var err error
	item, exists, err := i.informer.GetIndexer().GetByKey(key.(string))
	if err != nil {
		i.queue.Forget(key)
	}
	if !exists {
		i.queue.Forget(key)
		return true
	} else {
		i.handler.Update(item)
		i.queue.Forget(key)
	}
	return true
}

func (i *Listener) Run(stopper chan struct{}) {
	go i.informer.Run(stopper)
	if !cache.WaitForCacheSync(stopper, i.informer.HasSynced) {
		utilruntime.HandleError(fmt.Errorf("Error syncing cache"))
	}
	for i.processNextItem() {
	}
}
