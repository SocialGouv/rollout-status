package main_test

import (
	"io/ioutil"
	"path/filepath"
	"runtime"

	"github.com/SocialGouv/rollout-status/pkg/client"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
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
	// unmarshallAsset(assetPath("statefulsets.yaml"), &statefulSetList)
	unmarshallAsset(assetPath("pods.yaml"), &podList)

	return StaticClient{
		DeploymentList: &deploymentList,
		// StatefulSetList: &statefulSetList,
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

func mockWrapper(deployments []appsv1.Deployment, sts []appsv1.StatefulSet, replicaSets []appsv1.ReplicaSet, pods []v1.Pod) client.Kubernetes {
	return StaticClient{
		DeploymentList:  &appsv1.DeploymentList{Items: deployments},
		StatefulSetList: &appsv1.StatefulSetList{Items: sts},
		ReplicaSetList:  &appsv1.ReplicaSetList{Items: replicaSets},
		PodList:         &v1.PodList{Items: pods},
	}
}

type StaticClient struct {
	DeploymentList  *appsv1.DeploymentList
	ReplicaSetList  *appsv1.ReplicaSetList
	StatefulSetList *appsv1.StatefulSetList
	JobList         *batchv1.JobList
	PodList         *v1.PodList
}

func (client StaticClient) ListAppsV1Deployments(namespace, selector string) (*appsv1.DeploymentList, error) {
	return client.DeploymentList, nil
}

func (client StaticClient) ListAppsV1StatefulSets(namespace, selector string) (*appsv1.StatefulSetList, error) {
	return client.StatefulSetList, nil
}

func (client StaticClient) ListAppsV1ReplicaSets(deployment *appsv1.Deployment) (*appsv1.ReplicaSetList, error) {
	return client.ReplicaSetList, nil
}

func (client StaticClient) ListBatchV1Jobs(namespace, selector string) (*batchv1.JobList, error) {
	return client.JobList, nil
}

func (client StaticClient) ListV1Pods(replicasSet *appsv1.ReplicaSet) (*v1.PodList, error) {
	if replicasSet == nil {
		return client.PodList, nil
	}

	// Filter pods to only include those owned by this ReplicaSet
	var filteredPods []v1.Pod
	for _, pod := range client.PodList.Items {
		for _, ownerRef := range pod.OwnerReferences {
			if ownerRef.Kind == "ReplicaSet" && ownerRef.Name == replicasSet.Name {
				filteredPods = append(filteredPods, pod)
				break
			}
		}
	}

	return &v1.PodList{Items: filteredPods}, nil
}

func (client StaticClient) ListV1StsPods(sts *appsv1.StatefulSet) (*v1.PodList, error) {
	return client.PodList, nil
}

func (client StaticClient) ListV1JobPods(job *batchv1.Job) (*v1.PodList, error) {
	return client.PodList, nil
}

func (client StaticClient) TrailContainerLogs(namespace, pod, container string) ([]byte, error) {
	panic("not implemented")
}
