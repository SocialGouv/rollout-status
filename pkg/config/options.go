package config

type KindFilter string

const (
	NoKindFilter          KindFilter = ""
	DeploymentKindFilter  KindFilter = "deployment"
	JobKindFilter         KindFilter = "job"
	StatefulsetKindFilter KindFilter = "statefulset"
)

type Options struct {
	IgnoreSecretNotFound   bool
	RetryLimit             int32
	PendingDeadLineSeconds int
	KindFilter             KindFilter
}
