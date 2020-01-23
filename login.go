package main

import (
	"flag"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"os"
	"path/filepath"
)

func kubeconfig(method string) *rest.Config {
	config, err := login(method)
	if err != nil {
		panic(err)
	}
	return config
}

func login(method string) (*rest.Config, error) {
	if method == "INCLUSTER" {
		return rest.InClusterConfig()
	}
	var kubeconfig *string
	if home := homeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()
	return clientcmd.BuildConfigFromFlags("", *kubeconfig)
}

func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE")
}
