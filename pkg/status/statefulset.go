package status

import (
	"fmt"

	"github.com/SocialGouv/rollout-status/pkg/client"
	appsv1 "k8s.io/api/apps/v1"
)

// https://github.com/kubernetes/kubernetes/blob/dde6e8e7465468c32642659cb708a5cc922add64/staging/src/k8s.io/kubectl/pkg/util/statefulset/statefulset.go#L36
const StatefulsetRevisionAnnotation = "statefulset.kubernetes.io/revision"

func StatefulsetStatus(wrapper client.Kubernetes, sts *appsv1.StatefulSet) RolloutStatus {
	replicasSetList, err := wrapper.ListAppsV1StsReplicaSets(sts)
	if err != nil {
		return RolloutFatal(err)
	}

	lastRevision, ok := sts.Annotations[StatefulsetRevisionAnnotation]
	if !ok {
		return RolloutFatal(fmt.Errorf("Missing annotation %q on sts %q", StatefulsetRevisionAnnotation, sts.Name))
	}

	aggr := Aggregator{}
	for _, replicaSet := range replicasSetList.Items {
		rsRevision, ok := replicaSet.Annotations[StatefulsetRevisionAnnotation]
		if !ok {
			return RolloutFatal(fmt.Errorf("Missing annotation %q on replicaset %q", StatefulsetRevisionAnnotation, replicaSet.Name))
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

	return aggr.Resolve()
}
