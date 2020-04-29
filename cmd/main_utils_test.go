package main_test

import (
	"github.com/clusterise/rollout-status/pkg/client"
	"io/ioutil"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"path/filepath"
	"runtime"
	"sigs.k8s.io/yaml"
)

var (
	projectPath string
)

func init() {
	_, b, _, _ := runtime.Caller(0)
	projectPath = filepath.Dir(filepath.Dir(b))
}

func mockWrapperFromAssets(name string) client.Kubernetes {
	var deploymentList appsv1.DeploymentList
	var replicaSetList appsv1.ReplicaSetList
	var podList v1.PodList

	assetDir := filepath.Join(projectPath, "tests", "assets", name)
	assetPath := func(file string) string {
		return filepath.Join(assetDir, file)
	}

	unmarshallAsset(assetPath("deployments.yaml"), &deploymentList)
	unmarshallAsset(assetPath("replicasets.yaml"), &replicaSetList)
	unmarshallAsset(assetPath("pods.yaml"), &podList)

	return StaticClient{
		DeploymentList: &deploymentList,
		ReplicaSetList: &replicaSetList,
		PodList:        &podList,
	}
}

func unmarshallAsset(path string, o interface{}) {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		panic(err.Error())
	}
	err = yaml.Unmarshal(content, o)
	if err != nil {
		panic(err.Error())
	}
}

func mockWrapper(deployments []appsv1.Deployment, replicaSets []appsv1.ReplicaSet, pods []v1.Pod) client.Kubernetes {
	return StaticClient{
		DeploymentList: &appsv1.DeploymentList{Items: deployments},
		ReplicaSetList: &appsv1.ReplicaSetList{Items: replicaSets},
		PodList:        &v1.PodList{Items: pods},
	}
}

type StaticClient struct {
	DeploymentList *appsv1.DeploymentList
	ReplicaSetList *appsv1.ReplicaSetList
	PodList        *v1.PodList
}

func (client StaticClient) ListAppsV1Deployments(namespace, selector string) (*appsv1.DeploymentList, error) {
	return client.DeploymentList, nil
}

func (client StaticClient) ListAppsV1ReplicaSets(deployment *appsv1.Deployment) (*appsv1.ReplicaSetList, error) {
	return client.ReplicaSetList, nil
}

func (client StaticClient) ListV1Pods(replicasSet *appsv1.ReplicaSet) (*v1.PodList, error) {
	return client.PodList, nil
}

func (client StaticClient) TrailContainerLogs(namespace, pod, container string) ([]byte, error) {
	panic("not implemented")
}
