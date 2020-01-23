package main

import (
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/remotecommand"
	"log"
	"os"
)

type Handler interface {
	Delete(obj interface{})
	Update(obj interface{})
}

const Command = "kill 1" //"bash&command=-c&command=kill+-s+TERM+1"

type ShutdownKiller struct {
}

func NewShutdownKiller() *ShutdownKiller {
	return &ShutdownKiller{}
}

func (s *ShutdownKiller) Update(obj interface{}) {
	mainContainer, exists := obj.(*v1.Pod).Annotations[MainContainerAnnotation]
	if !exists {
		return
	}
	var Containers2Kill []string
	for _, container := range obj.(*v1.Pod).Status.ContainerStatuses {
		if container.Name != mainContainer {
			Containers2Kill = append(Containers2Kill, container.Name)
		} else {
			if container.State.Terminated != nil && container.State.Terminated.Reason == "Completed" {
				defer sendShutdownSignal(obj.(*v1.Pod), &Containers2Kill)
			}
		}
	}
	log.Println("Got an update from: " + obj.(*v1.Pod).Name)
}

func (s *ShutdownKiller) Delete(obj interface{}) {
	log.Println("Deleted")
	return
}

func sendShutdownSignal(pod *v1.Pod, Containers2Kill *[]string) error {
	log.Println("Killing ", *Containers2Kill, " in ", pod.Name)
	clientSet, err := kubernetes.NewForConfig(KubeConfig)
	if err != nil {
		return err
	}
	req := clientSet.CoreV1().RESTClient().Post().Resource("pods").Name(pod.Name).
		Namespace(Namespace).SubResource("exec")
	for _, container := range *Containers2Kill {
		option := &v1.PodExecOptions{
			Container: container,
			Command: []string{
				"sh",
				"-c",
				Command,
			},
			Stdin:  true,
			Stdout: true,
			Stderr: true,
			TTY:    true,
		}
		req.VersionedParams(
			option,
			scheme.ParameterCodec,
		)
		exec, err := remotecommand.NewSPDYExecutor(KubeConfig, "POST", req.URL())
		if err != nil {
			return err
		}
		err = exec.Stream(remotecommand.StreamOptions{
			Stdin:  os.Stdin,
			Stdout: os.Stdout,
			Stderr: os.Stderr,
			Tty:    true,
		})
		if err != nil {
			return err
		}
	}
	return nil
}
