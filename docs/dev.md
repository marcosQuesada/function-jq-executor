# Development Notes

```shell
# Run code generation - see input/generate.go
$ go generate ./...

# Run tests - see fn_test.go
$ go test ./...

# Build the function's runtime image - see Dockerfile
$ docker build . --tag=runtime

# Build a function package - see package/crossplane.yaml
$ crossplane xpkg build -f package --embed-runtime-image=runtime
```
### Development steps
Create function skeleton:
```
crossplane xpkg init function-jq-executor function-template-go -d function-jq-executor 
```

Run function locally:
```shell
go run . --insecure --debug
```

## Build
```shell
docker build . --quiet --platform=linux/amd64 --tag runtime-amd64
crossplane xpkg build \
    --package-root=package \
    --embed-runtime-image=runtime-amd64 \
    --package-file=function-amd64.xpkg
    
crossplane xpkg push --package-files=function-amd64.xpkg  docker.io/marcosquesada/function-jq-executor:v0.0.5
 
```
---
```shell
kind create cluster
```
```shell
kubectl create namespace crossplane-system

helm repo add crossplane-stable https://charts.crossplane.io/stable
helm repo update

helm install crossplane --namespace crossplane-system crossplane-stable/crossplane
```

---
## Crossplane BUG?
```shell
2025-04-13T16:26:58Z	DEBUG	crossplane	cannot compose resources	{"controller": "defined/compositeresourcedefinition.apiextensions.crossplane.io", "controller": "composite/xregisterexamples.example.crossplane.io", "request": {"name":"example-xr"}, "uid": "9ba49091-1e94-4861-ac29-4a0f6cb3bdbe", "version": "47038", "name": "example-xr", "error": "cannot apply composite resource status: metadata.managedFields must be nil"}

2025-04-13T16:24:58Z	DEBUG	crossplane	Event(v1.ObjectReference{Kind:"XRegisterExample", Namespace:"", Name:"example-xr", UID:"9ba49091-1e94-4861-ac29-4a0f6cb3bdbe", APIVersion:"example.crossplane.io/v1", ResourceVersion:"47038", FieldPath:""}): type: 'Warning' reason: 'ComposeResources' cannot compose resources: cannot apply composite resource status: metadata.managedFields must be nil
```