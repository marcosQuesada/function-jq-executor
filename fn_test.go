package main

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"google.golang.org/protobuf/testing/protocmp"
	"google.golang.org/protobuf/types/known/durationpb"
	"k8s.io/utils/ptr"

	"github.com/crossplane/function-sdk-go/logging"
	fnv1 "github.com/crossplane/function-sdk-go/proto/v1"
	"github.com/crossplane/function-sdk-go/resource"
	"github.com/crossplane/function-sdk-go/response"
)

func TestRunFunction(t *testing.T) {

	type args struct {
		ctx context.Context
		req *fnv1.RunFunctionRequest
	}
	type want struct {
		rsp *fnv1.RunFunctionResponse
		err error
	}

	cases := map[string]struct {
		reason string
		args   args
		want   want
	}{
		"ResponseIsReturned": {
			reason: "The Function should return a fatal result if no input was specified",
			args: args{
				req: &fnv1.RunFunctionRequest{
					Meta: &fnv1.RequestMeta{Tag: "hello"},
					Input: resource.MustStructJSON(`{
						"apiVersion": "template.fn.crossplane.io/v1beta1",
						"kind": "Input",
						"example": "Hello, world"
					}`),
				},
			},
			want: want{
				rsp: &fnv1.RunFunctionResponse{
					Meta: &fnv1.ResponseMeta{Tag: "hello", Ttl: durationpb.New(response.DefaultTTL)},
					Results: []*fnv1.Result{
						{
							Severity: fnv1.Severity_SEVERITY_FATAL,
							Message:  "cannot get Function input from *v1.RunFunctionRequest: cannot get function input *v1beta1.Input from *v1.RunFunctionRequest: cannot unmarshal JSON from *structpb.Struct into *v1beta1.Input: json: cannot unmarshal Go value of type v1beta1.Input: unknown name \"example\"",
							Target:   fnv1.Target_TARGET_COMPOSITE.Enum(),
						},
					},
					Conditions: []*fnv1.Condition{
						{
							Type:    "FunctionSuccess",
							Status:  fnv1.Status_STATUS_CONDITION_FALSE,
							Reason:  "InternalError",
							Target:  fnv1.Target_TARGET_COMPOSITE_AND_CLAIM.Enum(),
							Message: ptr.To("unable to get input."),
						},
					},
				},
			},
		},
		"SuccessResponseIsReturned": {
			reason: "The Function should return result whit success",
			args: args{
				req: &fnv1.RunFunctionRequest{
					Meta: &fnv1.RequestMeta{Tag: "hello"},
					Input: resource.MustStructJSON(`{
						"apiVersion": "template.fn.crossplane.io/v1beta1",
						"kind": "Input",
					    "jsonDataPath": "status.data",
					    "jsonQuery": ".id",
					    "responsePath": "status.value"
					}`),
					Observed: &fnv1.State{
						Composite: &fnv1.Resource{
							Resource: resource.MustStructJSON(`{
							  "apiVersion": "example.crossplane.io/v1",
							  "kind": "XRegisterExample",
							  "metadata": {
								"name": "example-xr"
							  },
							  "spec": {
								"name": "foo"
							  },
                              "status": {
								"data": "{\"id\":\"9696986696876\"}"
							  }
							}`),
						},
					},
					Desired: &fnv1.State{
						Resources: map[string]*fnv1.Resource{
							"http-request": {
								Resource: resource.MustStructJSON(`{
								  "apiVersion": "http.crossplane.io/v1alpha2",
								  "kind": "DisposableRequest",
								  "metadata": {
									"name": "obtain-jwt-token"
								  },
								  "spec": {
									"deletionPolicy": "Orphan",
									"forProvider": {
									  "url": "http://localhost:8000/v1/login/",
									  "method": "GET",
									  "shouldLoopInfinitely": true,
									  "nextReconcile": "72h"
									},
									"providerConfigRef": {
									  "name": "http-conf"
									}
								  },
								  "status": {
									"response": {
									  "body": "{\n  \"id\":\"65565b69681e0b47dcea4464\",\n  \"key\":\"value\"\n}",
									  "headers": {
										"Content-Length": [
										  104
										],
										"Content-Type": [
										  "application/json"
										],
										"Date": [
										  "Thu, 16 Nov 2023 18:11:53 GMT"
										],
										"Server": [
										  "uvicorn"
										]
									  },
									  "statusCode": 200
									}
								  }
								}`),
							},
						},
					},
				},
			},
			want: want{
				rsp: &fnv1.RunFunctionResponse{
					Meta: &fnv1.ResponseMeta{Tag: "hello", Ttl: durationpb.New(response.DefaultTTL)},
					Desired: &fnv1.State{
						Composite: &fnv1.Resource{
							Resource: resource.MustStructJSON(`{
							  "apiVersion": "example.crossplane.io/v1",
							  "kind": "XRegisterExample",
							  "metadata": {
								"name": "example-xr"
							  },
							  "spec": {
								"name": "foo"
							  },
                              "status": {
								"data": "{\"id\":\"9696986696876\"}",
								"value": "9696986696876"
							  }
							}`),
						},
						Resources: map[string]*fnv1.Resource{
							"http-request": {
								Resource: resource.MustStructJSON(`{
								  "apiVersion": "http.crossplane.io/v1alpha2",
								  "kind": "DisposableRequest",
								  "metadata": {
									"name": "obtain-jwt-token"
								  },
								  "spec": {
									"deletionPolicy": "Orphan",
									"forProvider": {
									  "url": "http://localhost:8000/v1/login/",
									  "method": "GET",
									  "shouldLoopInfinitely": true,
									  "nextReconcile": "72h"
									},
									"providerConfigRef": {
									  "name": "http-conf"
									}
								  },
								  "status": {
									"response": {
									  "body": "{\n  \"id\":\"65565b69681e0b47dcea4464\",\n  \"key\":\"value\"\n}",
									  "headers": {
										"Content-Length": [
										  104
										],
										"Content-Type": [
										  "application/json"
										],
										"Date": [
										  "Thu, 16 Nov 2023 18:11:53 GMT"
										],
										"Server": [
										  "uvicorn"
										]
									  },
									  "statusCode": 200
									}
								  }
								}`),
							},
						},
					},
					Results: []*fnv1.Result{
						{
							Severity: fnv1.Severity_SEVERITY_NORMAL,
							Message:  "execution success!",
							Target:   fnv1.Target_TARGET_COMPOSITE.Enum(),
						},
					},
					Conditions: []*fnv1.Condition{
						{
							Type:   "FunctionSuccess",
							Status: fnv1.Status_STATUS_CONDITION_TRUE,
							Reason: "Success",
							Target: fnv1.Target_TARGET_COMPOSITE_AND_CLAIM.Enum(),
						},
					},
				},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			f := &Function{log: logging.NewNopLogger()}
			rsp, err := f.RunFunction(tc.args.ctx, tc.args.req)

			if diff := cmp.Diff(tc.want.rsp, rsp, protocmp.Transform()); diff != "" {
				t.Errorf("%s\nf.RunFunction(...): -want rsp, +got rsp:\n%s", tc.reason, diff)
			}

			if diff := cmp.Diff(tc.want.err, err, cmpopts.EquateErrors()); diff != "" {
				t.Errorf("%s\nf.RunFunction(...): -want err, +got err:\n%s", tc.reason, diff)
			}
		})
	}
}

func Test_runJQuery(t *testing.T) {
	type args struct {
		jqQuery string
		rawObj  string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "basic query with success",
			args: args{
				jqQuery: ".id",
				rawObj:  `{"id":"asddasd"}`,
			},
			want: "asddasd",
		},
		{
			name: "basic query with no result returns empty string",
			args: args{
				jqQuery: ".id",
				rawObj:  `{"foo":"asddasd"}`,
			},
			want: "",
		},
		{
			name: "broken query fails with compile error",
			args: args{
				jqQuery: "id",
				rawObj:  `{"foo":"asddasd"}`,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := runJQuery(tt.args.jqQuery, tt.args.rawObj)
			if (err != nil) != tt.wantErr {
				t.Errorf("runJQuery() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("runJQuery() got = %v, want %v", got, tt.want)
			}
		})
	}
}
