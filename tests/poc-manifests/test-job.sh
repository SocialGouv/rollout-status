#!/bin/bash

kubectl delete -f job.yaml || true && kubectl apply -f job.yaml
sleep 5
../../rollout-status -selector type=job