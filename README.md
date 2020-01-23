# k8s-Job-SideKiller

A simple app to track k8s jobs that have additional side containers (istio-proxy, log-forward e.t.c)

# Quickstart

  - Run this app in your k8s cluster
  - add annotations in PodTemplate for jobs you want to track with main container as a value
  ```"k8s-job-sidekiller.killmeplz.github.com/main-container":"migrations"```
  - after ```migrations``` is completed any other containers in a pod will receive ```kill 1``` command.
 
# Example
CronJob
```
apiVersion: batch/v1beta1
kind: CronJob
metadata:
  name: example
spec:
  jobTemplate:
    spec:
      backoffLimit: 5
      template:
        metadata:
          annotations:
            k8s-job-sidekiller.killmeplz.github.com/main-container: busybox
          labels:
            app: busybox
        spec:
          containers:
          - name: busybox
            image: busybox:latest
            args: ["-c","exit 0"]
            command: ["/bin/sh"]
          - name: proxy
            image: nginx:alpine
  schedule: '* * * * *'
```