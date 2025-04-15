# function-jq-executor

[![CI](https://github.com/crossplane/function-template-go/actions/workflows/ci.yml/badge.svg)](https://github.com/marcosQuesada/function-jq-executor/actions/workflows/ci.yml)

A Crossplane Composite function that executes JSON Queries from Composed Resources field paths, designed to chain Http Requests and patch values processing responses.

Useful when there is a need of execute JSON Queries in the composition space. Implementation at **Proof of Concept** stage.

## Input Description
```shell
  - step: execute-json-query
    functionRef:
      name: function-jq-executor
    input:
      apiVersion: template.fn.crossplane.io/v1beta1
      kind: Input
      json-data-path: status.data
      json-query: ".id"
      response-path: status.value
```
Property description for this example:
- json-data-path: Path at Composite Resource where to find the JSON data to query, string expected. 
- json-query: Query to execute against JSON data.
- response-path: Path at Composite Resource to store query result.
Current implementation is pipelined with function-patch-and-transform to write JSON data input to Composite Resource Status and patches result values.

## Use Case Example
This composite function enables scenarios like this:
- Fire HTTP Request to a service (DisposableRequest from provider-http)
- a JSON HTTP Response body is stored at status.response.body
- function-jq-executor executes JSON Query defined at input.json-query field and publish results to Composite Resource at input.response-path path
- From Composite Resource status value we can patch another Composited Resource and chain previous result value

```shell
apiVersion: apiextensions.crossplane.io/v1
kind: Composition
metadata:
  name: function-jq-executor
spec:
  compositeTypeRef:
    apiVersion: example.crossplane.io/v1
    kind: XRegisterExample
  mode: Pipeline
  pipeline:
  - step: patch-and-transform-first-http-request
    functionRef:
      name: function-patch-and-transform
    input:
      apiVersion: pt.fn.crossplane.io/v1beta1
      kind: Resources
      resources:
        - name: first-http-request
          base:
            apiVersion: http.crossplane.io/v1alpha2
            kind: DisposableRequest
            metadata:
              name: fake
            spec:
              deletionPolicy: Orphan
              forProvider:
                url: http://172.17.0.1:8000/api/v1/pet/fake
                method: GET
                rollbackRetriesLimit: 5
              providerConfigRef:
                name: http-conf
          patches:
            - type: CombineFromComposite
              combine:
                variables:
                  - fromFieldPath: spec.name
                strategy: string
                string:
                  fmt: "first-request-%s"
              toFieldPath: metadata.name
            - type: CombineFromComposite
              combine:
                variables:
                  - fromFieldPath: spec.name
                strategy: string
                string:
                  fmt: "http://172.17.0.1:8080/api/v1/pet/%s"
              toFieldPath: spec.forProvider.url
            - type: ToCompositeFieldPath
              fromFieldPath: status.response.body
              toFieldPath: status.data
  - step: execute-json-query
    functionRef:
      name: function-jq-executor
    input:
      apiVersion: template.fn.crossplane.io/v1beta1
      kind: Input
      json-data-path: status.data
      json-query: ".id"
      response-path: status.value
  - step: patch-and-transform-second-request
    functionRef:
      name: function-patch-and-transform
    input:
      apiVersion: pt.fn.crossplane.io/v1beta1
      kind: Resources
      resources:
        - name: second-http-request
          base:
            apiVersion: http.crossplane.io/v1alpha2
            kind: DisposableRequest
            metadata:
              name: fake-second
            spec:
              deletionPolicy: Orphan
              forProvider:
                url: http://172.17.0.1:8000/api/v1/notify/fake
                method: GET
                rollbackRetriesLimit: 5
              providerConfigRef:
                name: http-conf
          patches:
            - type: CombineFromComposite
              combine:
                variables:
                  - fromFieldPath: spec.name
                strategy: string
                string:
                  fmt: "second-request-%s"
              toFieldPath: metadata.name
            - type: FromCompositeFieldPath
              fromFieldPath: status.value
              toFieldPath: metadata.labels["request-1-value"]
              policy:
                fromFieldPath: Required
            - type: CombineFromComposite
              combine:
                variables:
                  - fromFieldPath: status.value
                strategy: string
                string:
                  fmt: "http://172.17.0.1:8080/api/v1/notify/%s"
              toFieldPath: spec.forProvider.url
```

[functions]: https://docs.crossplane.io/latest/concepts/composition-functions
[go]: https://go.dev
[function guide]: https://docs.crossplane.io/knowledge-base/guides/write-a-composition-function-in-go
[package docs]: https://pkg.go.dev/github.com/crossplane/function-sdk-go
[docker]: https://www.docker.com
[cli]: https://docs.crossplane.io/latest/cli

