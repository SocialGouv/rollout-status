package status

type ResourceType int

const (
	ResourceTypeDeployment ResourceType = iota
	ResourceTypeStatefulSet
	ResourceTypeJob
)
