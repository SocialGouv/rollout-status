package main

import (
	"dite.pro/rollout-status/pkg/client"
	"dite.pro/rollout-status/pkg/status"
	"flag"
	"fmt"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"os"
	"path/filepath"
	"time"
)

func main() {
	namespace := flag.String("namespace", "", "Namespace to watch")
	selector := flag.String("selector", "", "Label selector to watch, kubectl format such as release=foo,component=frontend")

	var kubeconfig *string
	if home := homeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}

	flag.Parse()

	clientset := makeClientset(*kubeconfig)
	wrapper := client.FromClientset(clientset)

	for {
		rollout := status.TestRollout(wrapper, *namespace, *selector)
		if !rollout.Continue {
			if rollout.Error == nil {
				fmt.Println("Rollout successfully completed")

			} else if re, ok := rollout.Error.(status.RolloutError); ok {
				fmt.Printf("Rollout failed: %v\n", re)
				os.Exit(1)

			} else {
				fmt.Printf("Program failure: %v\n", rollout.Error)
				os.Exit(2)
			}
		}
		time.Sleep(10 * time.Second) // TODO configure
	}
}

func makeClientset(kubeconfigPath string) *kubernetes.Clientset {
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	if err != nil {
		panic(err.Error())
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
	return clientset
}

func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE") // windows
}
