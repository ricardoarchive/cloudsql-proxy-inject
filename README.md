# CloudSQL Proxy Inject (cloudsql-proxy-inject)

Inject a [CloudSQL Proxy](https://github.com/GoogleCloudPlatform/cloudsql-proxy/blob/master/Kubernetes.md) sidecar into a [Kubernetes deployment](https://kubernetes.io/docs/concepts/workloads/controllers/deployment/) manifest.

## Build
### Linux
```sh
$ make build-linux
```
### macos
```
$ make build-darwin
```

## Usage
```
usage: cloudsql-proxy-inject --path=PATH --instance=INSTANCE --region=REGION --project=PROJECT --verbose=false[<flags>]

Flags:
  --help                  Show context-sensitive help (also try --help-long and --help-man).
  --path=PATH             Deployment file path where to inject clousql proxy (eg. ./my-deploy-manifest.yaml)
  --instance=INSTANCE     CloudSQL instance (eg. my-clousql-instance=tcp:5432)
  --region=REGION         GCP region (eg. europe-west1)
  --project=PROJECT       GCP project ID (eg. ricardo)
  --cpu-request="5m"      CPU request of the sidecar container
  --memory-request="8Mi"  Memory request of the sidecar container
  --cpu-limit="100m"      CPU limit of the sidecar container
  --memory-limit="128Mi"  Memory limit of the sidecar container
  --proxy-version="1.11"  CloudSQL proxy version
  --verbose=VERBOSE       Verbose mode (eg. false)
```