package status

import (
	"github.com/SocialGouv/rollout-status/pkg/client"
	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
)

func TestJobStatus(wrapper client.Kubernetes, job batchv1.Job) RolloutStatus {
	podList, err := wrapper.ListV1JobPods(&job)
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

func JobStatus(wrapper client.Kubernetes, job *batchv1.Job) RolloutStatus {

	aggr := Aggregator{}

	for _, condition := range job.Status.Conditions {
		if condition.Type == batchv1.JobComplete && condition.Status == v1.ConditionTrue {
			aggr.Add(RolloutOk())
			return aggr.Resolve()
		}
	}

	status := TestJobStatus(wrapper, *job)
	if job.Status.Failed >= (*job.Spec.BackoffLimit + 1) {
		aggr.Add(status)
	} else if status.Error != nil {
		aggr.Add(RolloutErrorProgressing(status.Error))
	}
	if fatal := aggr.Fatal(); fatal != nil {
		return *fatal
	}

	return aggr.Resolve()
}
