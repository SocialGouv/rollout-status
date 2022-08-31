package main

import (
	"flag"
	"fmt"
	"github.com/socialgouv/rollout-status/pkg/client"
	"github.com/socialgouv/rollout-status/pkg/output"
	"github.com/socialgouv/rollout-status/pkg/status"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"os"
	"path/filepath"
	"time"
)

func main() {
	namespace := flag.String("namespace", "", "Namespace to watch rollout in")
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

	if *selector == "" {
		fmt.Fprintf(os.Stderr, "Missing value for flag selector\n")
		flag.Usage()
		os.Exit(2)
	}

	for {
		rollout := status.TestRollout(wrapper, *namespace, *selector)
		if !rollout.Continue {
			err := output.MakeOutput(os.Stdout, wrapper).PrintResult(rollout)
			if err != nil {
				panic(err)
			}

			if rollout.Error != nil {
				os.Exit(1)
			}
			os.Exit(0)
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
