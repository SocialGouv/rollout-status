package output

import (
	"github.com/socialgouv/rollout-status/pkg/client"
	"io"
)

type Output struct {
	writer  io.Writer
	wrapper client.Kubernetes
}

func MakeOutput(writer io.Writer, wrapper client.Kubernetes) *Output {
	return &Output{
		writer:  writer,
		wrapper: wrapper,
	}
}
