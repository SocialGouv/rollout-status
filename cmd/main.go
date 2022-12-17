package main

import (
	"flag"
	"os"
	"time"

	"github.com/SocialGouv/rollout-status/pkg/client"
	"github.com/SocialGouv/rollout-status/pkg/config"
	"github.com/SocialGouv/rollout-status/pkg/output"
	"github.com/SocialGouv/rollout-status/pkg/status"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

func main() {
	namespace := flag.String("namespace", "", "Namespace to watch rollout in")
	selector := flag.String("selector", "", "Label selector to watch, kubectl format such as release=foo,component=frontend")
	kubecontext := flag.String("kubecontext", "", "Kubeconfig context to use")
	interval := flag.String("interval", "5s", "Interval between status checks")
	retryLimit := flag.Int64("retry-limit", 6, "Retry limit for deployments and statefulsets, default 6, -1 to disable")
	pendingDeadLineSeconds := flag.Int("pending-deadline", 180, "Pending deadLine in seconds, default 180, -1 to disable")

	ignoreSecretNotFound := flag.Bool("ignore-secret-not-found", false, "Ignore secret not found error")

	var kubeconfig *string
	kubeconfigEnv := os.Getenv("KUBECONFIG")
	if kubeconfigEnv != "" {
		kubeconfig = flag.String("kubeconfig", kubeconfigEnv, "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}

	flag.Parse()

	options := &config.Options{
		IgnoreSecretNotFound:   *ignoreSecretNotFound,
		RetryLimit:             int32(*retryLimit),
		PendingDeadLineSeconds: *pendingDeadLineSeconds,
	}

	clientset := makeClientset(*kubeconfig, *kubecontext)
	wrapper := client.FromClientset(clientset)

	intervalTimeDuration, err := time.ParseDuration(*interval)
	if err != nil {
		panic(err)
	}

	for {
		rollout := status.TestRollout(wrapper, *namespace, *selector, options)
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
		time.Sleep(intervalTimeDuration)
	}
}

func makeClientset(kubeconfigPath string, kubecontext string) *kubernetes.Clientset {
	var config *rest.Config
	var err error
	if kubeconfigPath != "" {
		config, err = clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
			&clientcmd.ClientConfigLoadingRules{ExplicitPath: kubeconfigPath},
			&clientcmd.ConfigOverrides{Context: clientcmdapi.Context{Cluster: kubecontext}}).ClientConfig()
	} else {
		config, err = clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
			clientcmd.NewDefaultClientConfigLoadingRules(),
			&clientcmd.ConfigOverrides{Context: clientcmdapi.Context{Cluster: kubecontext}}).ClientConfig()
	}

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
