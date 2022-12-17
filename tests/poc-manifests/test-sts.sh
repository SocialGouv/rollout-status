#!/bin/bash

kubectl delete -f sts.yaml || true && kubectl apply -f sts.yaml
sleep 5
../../rollout-status -selector type=sts -retry-limit 3