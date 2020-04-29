package output

import (
	"dite.pro/rollout-status/pkg/status"
	"encoding/json"
	"fmt"
)

type outputType struct {
	Success bool        `json:"success"`
	Error   errorOutput `json:"error"`
}

type errorType string

const (
	ErrorTypeProgram errorType = "program"
	ErrorTypeRollout errorType = "rollout"
)

type errorOutput struct {
	Code    status.Failure `json:"code"`
	Message string         `json:"message"`
	Type    errorType      `json:"type"`
	Log     string         `json:"log,omitempty"`
}

func (o Output) errorOutputFrom(err error) errorOutput {
	if re, ok := err.(status.RolloutError); ok {
		errOut := errorOutput{
			Code:    re.Failure,
			Message: re.Message,
			Type:    ErrorTypeRollout,
		}

		if errOut.Code == status.FailureProcessCrashing {
			errOut.Log, err = trailContainerLogs(re.Pod, re.Container)
			if err != nil {
			    return o.errorOutputFrom(err)
			}
		}

		return errOut
	}
	return errorOutput{
		Message: err.Error(),
		Type:    ErrorTypeProgram,
	}
}

func (o Output) PrintResult(rollout status.RolloutStatus) error {
	out := outputType{
		Success: rollout.Error == nil,
		Error:   o.errorOutputFrom(rollout.Error),
	}

	outBytes, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(o.writer, string(outBytes))
	return err
}