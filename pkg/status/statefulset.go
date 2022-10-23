package status

import (
	"fmt"

	"github.com/SocialGouv/rollout-status/pkg/client"
	appsv1 "k8s.io/api/apps/v1"
)

func TestStatefulSetStatus(wrapper client.Kubernetes, statefulSet appsv1.StatefulSet) RolloutStatus {
	podList, err := wrapper.ListV1StsPods(&statefulSet)
	if err != nil {
		return RolloutFatal(err)
	}

	aggr := Aggregator{}
	for _, pod := range podList.Items {
		status := TestPodStatus(&pod)
		aggr.Add(status)
		if fatal := aggr.Fatal(); fatal != nil {
			return *fatal
		}
	}
	return aggr.Resolve()
}

func StatefulsetStatus(wrapper client.Kubernetes, sts *appsv1.StatefulSet) RolloutStatus {

	aggr := Aggregator{}

	if sts.Spec.UpdateStrategy.Type != appsv1.RollingUpdateStatefulSetStrategyType {
		err := MakeRolloutError(FailureNotSupportedStratedy, "rollout status is only available for %s strategy type", appsv1.RollingUpdateStatefulSetStrategyType)
		aggr.Add(RolloutFatal(err))
		if fatal := aggr.Fatal(); fatal != nil {
			return *fatal
		}
	}

	if sts.Spec.UpdateStrategy.Type != appsv1.RollingUpdateStatefulSetStrategyType {
		err := fmt.Errorf("rollout status is only available for %s strategy type", appsv1.RollingUpdateStatefulSetStrategyType)
		aggr.Add(RolloutErrorProgressing(err))
	}
	if sts.Status.ObservedGeneration == 0 || sts.Generation > sts.Status.ObservedGeneration {
		err := fmt.Errorf("Waiting for statefulset spec update to be observed...\n")
		aggr.Add(RolloutErrorProgressing(err))
	}
	if sts.Spec.Replicas != nil && sts.Status.ReadyReplicas < *sts.Spec.Replicas {
		err := fmt.Errorf("Waiting for %d pods to be ready...\n", *sts.Spec.Replicas-sts.Status.ReadyReplicas)
		aggr.Add(RolloutErrorProgressing(err))
	}
	if sts.Spec.UpdateStrategy.Type == appsv1.RollingUpdateStatefulSetStrategyType && sts.Spec.UpdateStrategy.RollingUpdate != nil {
		if sts.Spec.Replicas != nil && sts.Spec.UpdateStrategy.RollingUpdate.Partition != nil {
			if sts.Status.UpdatedReplicas < (*sts.Spec.Replicas - *sts.Spec.UpdateStrategy.RollingUpdate.Partition) {
				err := fmt.Errorf("Waiting for partitioned roll out to finish: %d out of %d new pods have been updated...\n",
					sts.Status.UpdatedReplicas, *sts.Spec.Replicas-*sts.Spec.UpdateStrategy.RollingUpdate.Partition)
				aggr.Add(RolloutErrorProgressing(err))
			}
		}
	} else if sts.Status.UpdateRevision != sts.Status.CurrentRevision {
		err := fmt.Errorf("waiting for statefulset rolling update to complete %d pods at revision %s...\n",
			sts.Status.UpdatedReplicas, sts.Status.UpdateRevision)
		aggr.Add(RolloutErrorProgressing(err))
	}

	status := TestStatefulSetStatus(wrapper, *sts)
	aggr.Add(status)
	if fatal := aggr.Fatal(); fatal != nil {
		return *fatal
	}

	return aggr.Resolve()
}
