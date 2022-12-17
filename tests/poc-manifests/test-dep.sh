#!/bin/bash

kubectl delete -f dep.yaml || true && kubectl apply -f dep.yaml
sleep 5
../../rollout-status -selector type=dep -retry-limit 2