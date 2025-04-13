# function-jq-executor
[![CI](https://github.com/crossplane/function-template-go/actions/workflows/ci.yml/badge.svg)](https://github.com/marcosQuesada/function-jq-executor/actions/workflows/ci.yml)

A Composite function that executes JSON Queries from Composed Resources field paths. 

Useful when there is a need of execute JSON Queries in the composition space. Implementation at Proof of Concept stage.

As example, this composite function enables scenarios like this:
- Fire HTTP Request to a service (DisposableRequest from provider-http)
- a JSON HTTP Response body is stored at status.response.body
- function-jq-executor executes JSON Query defined at input.json-query field and publish results to Composite Resource at input.response-path path
- From Composite Resource status value we can patch another Composited Resource and chain previous result value

Check example folder to see full composition details.

[functions]: https://docs.crossplane.io/latest/concepts/composition-functions
[go]: https://go.dev
[function guide]: https://docs.crossplane.io/knowledge-base/guides/write-a-composition-function-in-go
[package docs]: https://pkg.go.dev/github.com/crossplane/function-sdk-go
[docker]: https://www.docker.com
[cli]: https://docs.crossplane.io/latest/cli

---
