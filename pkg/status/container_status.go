package status

import (
	"strings"

	"github.com/SocialGouv/rollout-status/pkg/config"
	v1 "k8s.io/api/core/v1"
)

func TestContainerStatus(status *v1.ContainerStatus, options *config.Options, resourceType ResourceType) RolloutStatus {
	// https://github.com/kubernetes/kubernetes/blob/4fda1207e347af92e649b59d60d48c7021ba0c54/pkg/kubelet/container/sync_result.go#L37
	if status.State.Waiting != nil {
		reason := status.State.Waiting.Reason
		switch reason {
		case "ContainerCreating":
			fallthrough
		case "PodInitializing":
			err := MakeRolloutError(NoFailure, "Container %q is in %q", status.Name, reason)
			return RolloutErrorProgressing(err)

		case "CrashLoopBackOff":
			fallthrough
		case "RunContainerError":
			err := MakeRolloutError(FailureProcessCrashing, "Container %q is in %q: %v", status.Name, reason, status.State.Waiting.Message)
			if (resourceType == ResourceTypeDeployment || resourceType == ResourceTypeStatefulSet) && status.RestartCount <= options.RetryLimit {
				return RolloutErrorProgressing(err)
			} else {
				return RolloutFatal(err)
			}

		case "ErrImagePull":
			fallthrough
		case "ImagePullBackOff":
			err := MakeRolloutError(FailureInvalidConfig, "Container %q is in %q: %v", status.Name, reason, status.State.Waiting.Message)
			return RolloutFatal(err)

		case "CreateContainerConfigError":
			err := MakeRolloutError(FailureInvalidConfig, "Container %q is in %q: %v", status.Name, reason, status.State.Waiting.Message)
			if options.IgnoreSecretNotFound {
				if strings.HasPrefix(status.State.Waiting.Message, "secret ") && strings.HasSuffix(status.State.Waiting.Message, " not found") {
					return RolloutErrorProgressing(err)
				}
			}
			return RolloutFatal(err)
		}
	}

	if status.State.Terminated != nil {
		reason := status.State.Terminated.Reason
		switch reason {
		case "Error":
			// TODO this should retry but have a deadline, all restarts fall to CrashLoopBackOff
			err := MakeRolloutError(FailureProcessCrashing, "Container %q is in %q", status.Name, reason)
			return RolloutErrorProgressing(err)
		case "OOMKilled":
			err := MakeRolloutError(FailureResourceLimitsExceeded, "Container %q is in %q", status.Name, reason)
			return RolloutFatal(err)
		}
	}

	return RolloutOk()
}
