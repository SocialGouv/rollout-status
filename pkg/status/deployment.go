package status

import (
	"fmt"
	"github.com/socialgouv/rollout-status/pkg/client"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
)

// https://github.com/kubernetes/kubernetes/blob/dde6e8e7465468c32642659cb708a5cc922add64/staging/src/k8s.io/kubectl/pkg/util/deployment/deployment.go#L36
const RevisionAnnotation = "deployment.kubernetes.io/revision"

func DeploymentStatus(wrapper client.Kubernetes, deployment *appsv1.Deployment) RolloutStatus {
	replicasSetList, err := wrapper.ListAppsV1ReplicaSets(deployment)
	if err != nil {
		return RolloutFatal(err)
	}

	lastRevision, ok := deployment.Annotations[RevisionAnnotation]
	if !ok {
		return RolloutFatal(fmt.Errorf("Missing annotation %q on deployment %q", RevisionAnnotation, deployment.Name))
	}

	aggr := Aggregator{}
	for _, replicaSet := range replicasSetList.Items {
		rsRevision, ok := replicaSet.Annotations[RevisionAnnotation]
		if !ok {
			return RolloutFatal(fmt.Errorf("Missing annotation %q on replicaset %q", RevisionAnnotation, replicaSet.Name))
		}

		if rsRevision != lastRevision {
			// RS that is live and we no longer want
			// TODO make sure there are no more pods running in this RS
			// or older RS or RS that never became live and was overwritten
			continue
		}

		status := TestReplicaSetStatus(wrapper, replicaSet)
		aggr.Add(status)
		if fatal := aggr.Fatal(); fatal != nil {
			return *fatal
		}
	}

	for _, condition := range deployment.Status.Conditions {
		if condition.Type == appsv1.DeploymentProgressing && condition.Status == v1.ConditionFalse {
			err := MakeRolloutError(FailureNotProgressing, "Deployment %q is not progressing: %v", deployment.Name, condition.Message)
			aggr.Add(RolloutFatal(err))
			break
		}
	}

	return aggr.Resolve()
}
