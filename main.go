package main

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	"os"
)

const MainContainerAnnotation = "k8s-job-sidekiller.killmeplz.github.com/main-container"

var LoginMethod = os.Getenv("AUTH_METHOD")
var Namespace = os.Getenv("POD_NAMESPACE")
var KubeConfig = kubeconfig(LoginMethod)

func main() {
	clientSet, err := kubernetes.NewForConfig(KubeConfig)
	if err != nil {
		panic(err.Error())
	}

	//channel to synchronize goroutines
	stopCh := make(chan struct{})
	defer close(stopCh)

	queue := workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())

	informer := cache.NewSharedIndexInformer(
		listWatchFuncs(clientSet),
		&v1.Pod{},
		0,
		cache.Indexers{},
	)
	informer.AddEventHandler(handlerFuncs(queue))

	listener := Listener{
		informer: informer,
		queue:    queue,
		handler:  NewShutdownKiller(),
	}
	listener.Run(stopCh)
}

func listWatchFuncs(c *kubernetes.Clientset) *cache.ListWatch {
	return &cache.ListWatch{
		ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
			return c.CoreV1().Pods(Namespace).List(options)
		},
		WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
			return c.CoreV1().Pods(Namespace).Watch(options)
		},
	}
}

func handlerFuncs(q workqueue.RateLimitingInterface) *cache.ResourceEventHandlerFuncs {
	return &cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(obj)
			if err == nil {
				q.Add(key)
			}
		},
		DeleteFunc: func(obj interface{}) {
			key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
			if err == nil {
				q.Add(key)
			}
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(newObj)
			if err == nil {
				q.Add(key)
			}
		},
	}
}
