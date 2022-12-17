package config

type Options struct {
	IgnoreSecretNotFound   bool
	RetryLimit             int32
	PendingDeadLineSeconds int
}
