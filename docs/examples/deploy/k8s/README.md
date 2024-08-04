# Kubernetes deployment example

## Prerequisites

* ingress-nginx
* openebs with hostpath enabled for PVC
* VictoriaMetrics or Prometheus operator for PodMonitors
* CloudNativePG for PostgreSQL
* External S3 service to store blobs

## Deploy

* Change all the fields marked as "CHANGEME" to appropriate values
* `kubectl apply -f .`
