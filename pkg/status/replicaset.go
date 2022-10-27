package status

import (
	"github.com/SocialGouv/rollout-status/pkg/client"
	"github.com/SocialGouv/rollout-status/pkg/config"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
)

func TestReplicaSetStatus(wrapper client.Kubernetes, replicaSet appsv1.ReplicaSet, options *config.Options) RolloutStatus {
	for _, rsCondition := range replicaSet.Status.Conditions {
		if rsCondition.Type == appsv1.ReplicaSetReplicaFailure && rsCondition.Status == v1.ConditionTrue {
			err := MakeRolloutError(FailureResourceLimitsExceeded, "ReplicaSet %q failed to create pods: %v", replicaSet.Name, rsCondition.Message)
			return RolloutFatal(err)
		}
	}

	podList, err := wrapper.ListV1Pods(&replicaSet)
	if err != nil {
		return RolloutFatal(err)
	}

	aggr := Aggregator{}
	for _, pod := range podList.Items {
		status := TestPodStatus(&pod, options)
		aggr.Add(status)
		if fatal := aggr.Fatal(); fatal != nil {
			return *fatal
		}
	}
	return aggr.Resolve()
}
