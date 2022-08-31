package output

import (
	"io"

	"github.com/SocialGouv/rollout-status/pkg/client"
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
