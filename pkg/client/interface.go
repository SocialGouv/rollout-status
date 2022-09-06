package client

import (
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
)

type Kubernetes interface {
	ListAppsV1Deployments(namespace, selector string) (*appsv1.DeploymentList, error)
	ListAppsV1StatefulSets(namespace, selector string) (*appsv1.StatefulSetList, error)
	ListAppsV1ReplicaSets(deployment *appsv1.Deployment) (*appsv1.ReplicaSetList, error)
	ListBatchV1Jobs(namespace, selector string) (*batchv1.JobList, error)
	ListV1Pods(replicasSet *appsv1.ReplicaSet) (*v1.PodList, error)
	ListV1StsPods(replicasSet *appsv1.StatefulSet) (*v1.PodList, error)
	ListV1JobPods(job *batchv1.Job) (*v1.PodList, error)
	TrailContainerLogs(namespace, pod, container string) ([]byte, error)
}
