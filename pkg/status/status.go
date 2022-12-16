package status

import (
	"fmt"
)

type Failure string

const (
	NoFailure                     Failure = ""
	FailureNotFound               Failure = "not-found"
	FailureProcessCrashing        Failure = "process-crashing"
	FailureInvalidConfig          Failure = "invalid-config"
	FailureResourceLimitsExceeded Failure = "resource-limits-exceeded"
	FailureScheduling             Failure = "scheduling"
	FailureNotProgressing         Failure = "not-progressing"
	FailureNotSupportedStratedy   Failure = "not-supported-strategy"
)

type RolloutStatus struct {
	Continue bool
	Error    error
}

// Rollout is not completed and it will not succeed with a high certainty.
// Also returned for program issues not directly relevant to the rollout.
func RolloutFatal(err error) RolloutStatus {
	return RolloutStatus{
		Continue: false,
		Error:    err,
	}
}

// Rollout is not completed but there is a chance it will eventually succeed.
// Returned for valid flow (ContainerCreating -> init containers -> not ready -> ready)
// as well as for error states under a deadline.
func RolloutErrorProgressing(err error) RolloutStatus {
	return RolloutStatus{
		Continue: true,
		Error:    err,
	}
}

// Rollout completed successfully
func RolloutOk() RolloutStatus {
	return RolloutStatus{
		Continue: false,
		Error:    nil,
	}
}

type RolloutError struct {
	Failure Failure
	Message string

	Namespace string
	Pod       string
	Container string
}

func (re RolloutError) Error() string {
	return re.Message
}

func MakeRolloutError(failure Failure, format string, args ...interface{}) RolloutError {
	return RolloutError{
		Failure: failure,
		Message: fmt.Sprintf(format, args...),
	}
}
