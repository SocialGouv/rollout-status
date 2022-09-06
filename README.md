Rollout Status
==============

An alternative to `kubectl rollout status`.

It is intended as a building block for your CI/CD pipeline. Unlike `kubectl` it does not rely *only* on `progressDeadlineSeconds` and fails fast when appropriate, such as when image does not exist or environment properties are misconfigured. 


Usage
-----

Docker Image is [available on GitHub packages](https://github.com/SocialGouv/rollout-status/pkgs/container/rollout-status).

```console
docker run -v "$HOME/.kube:/root/.kube:ro" ghcr.io/socialgouv/rollout-status:latest -namespace "$NAMESPACE" -selector "$SELECTOR"
```

Since the program is intended for a CI/CD pipeline, you will likely copy the binary to your own `linux-amd64` image:

```Dockerfile
# ...

COPY --from=ghcr.io/socialgouv/rollout-status:latest /rollout-status /opt/rollout-status 
```

Overview of handled states
--------------------------

Listed by `error.code`:

* `not-found`: Deployment or the ReplicaSet was not found in given `-namespace`. This is may be a pipeline setup error or indicate other deployment issues that prevented your manifests from being saved to Kubernetes.
* `invalid-config`: Includes `ErrImagePull`, `ImagePullBackOff`, `init:ErrImagePull`, `init:ImagePullBackOff`. Fails the deployment immediately as we expect the image to be built and published before we attempt application deployment. Also incldues `CreateContainerConfigError` and `init:CreateContainerConfigError` which happen when invalid envs are defined (such as mounting non-existent ConfigMap) etc.
* `resource-limits-exceeded`: Aborts deployment when ReplicaSet is unable to create pods. This may happen when available resources are limited with `ResourceQuota` or `LimitRange`.
* `process-crashing`: Application is scheduled and runs, but exits. This encompases `RunContainerError`, `CrashLoopBackOff` as well as init container errors `init:RunContainerError` and `init:CrashLoopBackOff` and termination `Error`.
* `scheduling`: Waits for `Pending` pods for (currently hardcoded) deadline. If not scheduled, deployment fails.
* `not-progressing`: Waits until [`.spec.progressDeadlineSeconds`](https://kubernetes.io/docs/concepts/workloads/controllers/deployment/#progress-deadline-seconds). If Pods do not become ready, deployment fails.

Transient blocking states. Multiple states such as waiting for scheduling, `PodInitializing`, `ContainerCreating` as well as waiting for pods to become ready block rollout-status. The final result is reported only after a pod fails or all pods transition to ready state and succeed. 


Example
-------

```console
go run github.com/SocialGouv/rollout-status/cmd -selector app=crashloop-backoff
```
```json
{
  "success": false,
  "error": {
    "code": "process-crashing",
    "message": "Container \"main\" is in \"CrashLoopBackOff\": back-off 5m0s restarting failed container=main pod=crashloop-backoff-7fd845849c-cfvqd_default(617f6364-5bd4-4e69-b19f-fcf3ce4c171a)",
    "type": "rollout",
    "log": "panic: invalid configuration: [unable to read client-key /Users/mikulas/.minikube/profiles/minikube/client.key for minikube due to open /Users/mikulas/.minikube/profiles/minikube/client.key: no such file or directory, unable to read certificate-authority /Users/mikulas/.minikube/ca.crt for minikube due to open /Users/mikulas/.minikube/ca.crt: no such file or directory]\n\ngoroutine 1 [running]:\nmain.makeClientset(0xc000041440, 0x12, 0x0)\n\t/src/cmd/main.go:57 +0xfb\nmain.main()\n\t/src/cmd/main.go:34 +0x21c\n"
  }
}
```


Output
------

See `schema.json` for details. The program blocks until rollout succeeds (as `success=true`) or is not progressing and effectively failed (`success=false`). In the latter case an `error` object is returned. It contains all details that should be neccessary to remediate the issue. Besides `error.mesage` your pipeline should also switch on `error.code` and print help specific to your pipeline.
