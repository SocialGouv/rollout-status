package status

import (
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
	if !((sts.Status.ObservedGeneration == 0 || sts.Generation > sts.Status.ObservedGeneration) ||
		(sts.Spec.Replicas != nil && sts.Status.ReadyReplicas < *sts.Spec.Replicas) ||
		(sts.Spec.UpdateStrategy.Type == appsv1.RollingUpdateStatefulSetStrategyType && sts.Spec.UpdateStrategy.RollingUpdate != nil) ||
		(sts.Status.UpdateRevision != sts.Status.CurrentRevision)) {
		return aggr.Resolve()
	}

	status := TestStatefulSetStatus(wrapper, *sts)
	aggr.Add(status)
	if fatal := aggr.Fatal(); fatal != nil {
		return *fatal
	}

	return aggr.Resolve()
}
