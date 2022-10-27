package status

import (
	"fmt"

	"github.com/SocialGouv/rollout-status/pkg/client"
	"github.com/SocialGouv/rollout-status/pkg/config"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
)

// https://github.com/kubernetes/kubernetes/blob/dde6e8e7465468c32642659cb708a5cc922add64/staging/src/k8s.io/kubectl/pkg/util/deployment/deployment.go#L36
const RevisionAnnotation = "deployment.kubernetes.io/revision"

func DeploymentStatus(wrapper client.Kubernetes, deployment *appsv1.Deployment, options *config.Options) RolloutStatus {
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

		status := TestReplicaSetStatus(wrapper, replicaSet, options)
		aggr.Add(status)
		if fatal := aggr.Fatal(); fatal != nil {
			return *fatal
		}
	}

	for _, condition := range deployment.Status.Conditions {
		if condition.Type == appsv1.DeploymentProgressing {
			if condition.Status == v1.ConditionFalse {
				err := MakeRolloutError(FailureNotProgressing, "Deployment %q is not progressing: %v", deployment.Name, condition.Message)
				aggr.Add(RolloutFatal(err))
				break
			}
		}
	}

	if deployment.Spec.Replicas != nil && deployment.Status.UpdatedReplicas < *deployment.Spec.Replicas {
		err := fmt.Errorf("Waiting for deployment %q rollout to finish: %d out of %d new replicas have been updated...\n", deployment.Name, deployment.Status.UpdatedReplicas, *deployment.Spec.Replicas)
		aggr.Add(RolloutErrorProgressing(err))
	}
	if deployment.Status.Replicas > deployment.Status.UpdatedReplicas {
		err := fmt.Errorf("Waiting for deployment %q rollout to finish: %d old replicas are pending termination...\n", deployment.Name, deployment.Status.Replicas-deployment.Status.UpdatedReplicas)
		aggr.Add(RolloutErrorProgressing(err))
	}
	if deployment.Status.AvailableReplicas < deployment.Status.UpdatedReplicas {
		err := fmt.Errorf("Waiting for deployment %q rollout to finish: %d of %d updated replicas are available...\n", deployment.Name, deployment.Status.AvailableReplicas, deployment.Status.UpdatedReplicas)
		aggr.Add(RolloutErrorProgressing(err))
	}

	return aggr.Resolve()
}
