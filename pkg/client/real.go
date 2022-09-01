package client

import (
	"fmt"
	"io/ioutil"

	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const MaxLogLines = 20
const MaxLogBytes = MaxLogLines * 5000

type KubernetesImpl struct {
	clientset *kubernetes.Clientset
}

func (impl KubernetesImpl) ListAppsV1Deployments(namespace, selector string) (*appsv1.DeploymentList, error) {
	return impl.clientset.AppsV1().Deployments(namespace).List(metav1.ListOptions{LabelSelector: selector})
}

func (impl KubernetesImpl) ListAppsV1StatefulSets(namespace, selector string) (*appsv1.StatefulSetList, error) {
	return impl.clientset.AppsV1().StatefulSets(namespace).List(metav1.ListOptions{LabelSelector: selector})
}

func (impl KubernetesImpl) ListAppsV1ReplicaSets(deployment *appsv1.Deployment) (*appsv1.ReplicaSetList, error) {
	selector, err := metav1.LabelSelectorAsSelector(deployment.Spec.Selector)
	if err != nil {
		return nil, err
	}
	options := metav1.ListOptions{LabelSelector: selector.String()}
	return impl.clientset.AppsV1().ReplicaSets(deployment.Namespace).List(options)
}

func (impl KubernetesImpl) ListV1Pods(replicasSet *appsv1.ReplicaSet) (*v1.PodList, error) {
	selector, err := metav1.LabelSelectorAsSelector(replicasSet.Spec.Selector)
	if err != nil {
		return nil, err
	}
	options := metav1.ListOptions{LabelSelector: selector.String()}
	return impl.clientset.CoreV1().Pods(replicasSet.Namespace).List(options)
}

func (impl KubernetesImpl) ListV1StsPods(sts *appsv1.StatefulSet) (*v1.PodList, error) {
	selector, err := metav1.LabelSelectorAsSelector(sts.Spec.Selector)
	if err != nil {
		return nil, err
	}
	options := metav1.ListOptions{LabelSelector: selector.String()}
	return impl.clientset.CoreV1().Pods(sts.Namespace).List(options)
}

func (impl KubernetesImpl) TrailContainerLogs(namespace, pod, container string) ([]byte, error) {
	// https://github.com/kubernetes/kubernetes/blob/c2e90cd1549dff87db7941544ce15f4c8ad0ba4c/pkg/kubectl/cmd/log.go#L188
	req := impl.clientset.CoreV1().RESTClient().Get().
		Namespace(namespace).
		Name(pod).
		Resource("pods").
		SubResource("log").
		Param("container", container).
		Param("tailLines", fmt.Sprintf("%v", MaxLogLines)).
		Param("limitBytes", fmt.Sprintf("%v", MaxLogBytes))

	readCloser, err := req.Stream()
	if err != nil {
		return nil, err
	}

	defer readCloser.Close()
	return ioutil.ReadAll(readCloser)
}

func FromClientset(clientset *kubernetes.Clientset) *KubernetesImpl {
	return &KubernetesImpl{
		clientset: clientset,
	}
}
