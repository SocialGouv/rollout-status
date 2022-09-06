package status

import (
	"github.com/SocialGouv/rollout-status/pkg/client"
)

func TestRollout(wrapper client.Kubernetes, namespace, selector string) RolloutStatus {
	deployments, err := wrapper.ListAppsV1Deployments(namespace, selector)
	if err != nil {
		return RolloutFatal(err)
	}

	statefulsets, err := wrapper.ListAppsV1StatefulSets(namespace, selector)
	if err != nil {
		return RolloutFatal(err)
	}

	jobs, err := wrapper.ListBatchV1Jobs(namespace, selector)
	if err != nil {
		return RolloutFatal(err)
	}

	if (deployments == nil || len(deployments.Items) == 0) && (statefulsets == nil || len(statefulsets.Items) == 0) && (jobs == nil || len(jobs.Items) == 0) {
		err = MakeRolloutError(FailureNotFound, "Selector %q did not match any deployments, statefulsets or jobs in namespace %q", selector, namespace)
		return RolloutFatal(err)
	}

	aggr := Aggregator{}
	//https://github.com/kubernetes/kubernetes/blob/master/staging/src/k8s.io/kubectl/pkg/cmd/rollout/rollout_status.go
	//https://github.com/kubernetes/kubernetes/blob/47daccb272c1a98c7b005dc1c19a88dbb643a3ee/staging/src/k8s.io/kubectl/pkg/polymorphichelpers/rollout_status.go#L59
	if deployments != nil {
		for _, deployment := range deployments.Items {
			status := DeploymentStatus(wrapper, &deployment)
			aggr.Add(status)
			if fatal := aggr.Fatal(); fatal != nil {
				return *fatal
			}
		}
	}

	if statefulsets != nil {
		for _, statefulset := range statefulsets.Items {
			status := StatefulsetStatus(wrapper, &statefulset)
			aggr.Add(status)
			if fatal := aggr.Fatal(); fatal != nil {
				return *fatal
			}
		}
	}

	if jobs != nil {
		for _, job := range jobs.Items {
			status := JobStatus(wrapper, &job)
			aggr.Add(status)
			if fatal := aggr.Fatal(); fatal != nil {
				return *fatal
			}
		}
	}

	return aggr.Resolve()
}
