package status

import (
	"time"

	"github.com/SocialGouv/rollout-status/pkg/config"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestPodStatus(pod *v1.Pod, options *config.Options, resourceType ResourceType) RolloutStatus {
	aggr := Aggregator{}
	for _, initStatus := range pod.Status.InitContainerStatuses {
		status := TestContainerStatus(&initStatus, options, resourceType)
		if status.Error != nil {
			if !status.Continue {
				if re, ok := status.Error.(RolloutError); ok {
					re.Namespace = pod.Namespace
					re.Pod = pod.Name
					re.Container = initStatus.Name
					status.Error = re
				}
			}
		}

		aggr.Add(status)
		if fatal := aggr.Fatal(); fatal != nil {
			return *fatal
		}
	}

	for _, containerStatus := range pod.Status.ContainerStatuses {
		status := TestContainerStatus(&containerStatus, options, resourceType)
		if status.Error != nil {
			if !status.Continue {
				if re, ok := status.Error.(RolloutError); ok {
					re.Namespace = pod.Namespace
					re.Pod = pod.Name
					re.Container = containerStatus.Name
					status.Error = re
				}
			}
		}

		aggr.Add(status)
		if fatal := aggr.Fatal(); fatal != nil {
			return *fatal
		}
	}
	status := aggr.Resolve()
	if status.Error != nil {
		return status
	}

	if pod.Status.Phase == v1.PodPending {
		for _, condition := range pod.Status.Conditions {
			// fail if the pod is pending for X time
			if condition.Type == v1.PodScheduled {
				err := MakeRolloutError(FailureScheduling, "Failed to schedule pod: %v", condition.Message)

				if options.PendingDeadLineSeconds == -1 {
					return RolloutErrorProgressing(err)
				}

				deadline := metav1.NewTime(time.Now().Add(time.Second * -1 * time.Duration(options.PendingDeadLineSeconds)))

				if condition.LastTransitionTime.Before(&deadline) {
					return RolloutFatal(err)
				} else {
					return RolloutErrorProgressing(err)
				}
			}
		}
	}

	return RolloutOk()
}
