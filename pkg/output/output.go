package output

import (
	"encoding/json"
	"fmt"

	"github.com/SocialGouv/rollout-status/pkg/status"
)

type outputType struct {
	Success bool         `json:"success"`
	Error   *errorOutput `json:"error,omitempty"`
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

func (o Output) errorOutputFrom(err error) *errorOutput {
	if err == nil {
		return nil
	}

	if re, ok := err.(status.RolloutError); ok {
		errOut := errorOutput{
			Code:    re.Failure,
			Message: re.Message,
			Type:    ErrorTypeRollout,
		}

		if errOut.Code == status.FailureProcessCrashing && re.Namespace != "" && re.Pod != "" && re.Container != "" {
			logBytes, err := o.wrapper.TrailContainerLogs(re.Namespace, re.Pod, re.Container)
			if err != nil {
				return o.errorOutputFrom(err)
			}
			errOut.Log = string(logBytes)
		}

		return &errOut
	}
	return &errorOutput{
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
	o.writer.Write([]byte{'\n'})
	return err
}
