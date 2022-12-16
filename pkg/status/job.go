package status

import (
	"github.com/SocialGouv/rollout-status/pkg/client"
	"github.com/SocialGouv/rollout-status/pkg/config"
	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
)

func TestJobStatus(wrapper client.Kubernetes, job batchv1.Job, options *config.Options) RolloutStatus {
	podList, err := wrapper.ListV1JobPods(&job)
	if err != nil {
		return RolloutFatal(err)
	}

	aggr := Aggregator{}
	for _, pod := range podList.Items {
		status := TestPodStatus(&pod, options, ResourceTypeJob)

		aggr.Add(status)
		if fatal := aggr.Fatal(); fatal != nil {
			return *fatal
		}
	}
	return aggr.Resolve()
}

func JobStatus(wrapper client.Kubernetes, job *batchv1.Job, options *config.Options) RolloutStatus {

	aggr := Aggregator{}

	for _, condition := range job.Status.Conditions {
		if condition.Type == batchv1.JobComplete && condition.Status == v1.ConditionTrue {
			aggr.Add(RolloutOk())
			return aggr.Resolve()
		}
		if condition.Type == batchv1.JobFailed && condition.Status == v1.ConditionTrue {
			status := TestJobStatus(wrapper, *job, options)
			aggr.Add(status)
			return aggr.Resolve()
		}
	}

	status := TestJobStatus(wrapper, *job, options)
	aggr.Add(status)

	if fatal := aggr.Fatal(); fatal != nil {
		return *fatal
	}

	return aggr.Resolve()
}
