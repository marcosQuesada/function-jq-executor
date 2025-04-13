package main

import (
	"context"
	"fmt"
	"github.com/crossplane/function-sdk-go/errors"
	"github.com/crossplane/function-sdk-go/logging"
	fnv1 "github.com/crossplane/function-sdk-go/proto/v1"
	"github.com/crossplane/function-sdk-go/request"
	"github.com/crossplane/function-sdk-go/response"
	"github.com/itchyny/gojq"
	"github.com/marcosQuesada/function-jq-executor/input/v1beta1"
	"k8s.io/apimachinery/pkg/util/json"
)

// Function returns whatever response you ask it to.
type Function struct {
	fnv1.UnimplementedFunctionRunnerServiceServer

	log logging.Logger
}

// RunFunction runs the Function.
func (f *Function) RunFunction(_ context.Context, req *fnv1.RunFunctionRequest) (*fnv1.RunFunctionResponse, error) {
	f.log.Info("Running function", "tag", req.GetMeta().GetTag())

	rsp := response.To(req, response.DefaultTTL)

	in := &v1beta1.Input{}
	if err := request.GetInput(req, in); err != nil {
		response.ConditionFalse(rsp, "FunctionSuccess", "InternalError").
			WithMessage("unable to get input.").
			TargetCompositeAndClaim()

		response.Fatal(rsp, errors.Wrapf(err, "cannot get Function input from %T", req))
		return rsp, nil
	}

	xr, err := request.GetObservedCompositeResource(req)
	if err != nil {
		response.ConditionFalse(rsp, "FunctionSuccess", "InternalError").
			WithMessage(fmt.Sprintf("cannot get observed composite resource , error: %s", err.Error())).
			TargetCompositeAndClaim()

		response.Fatal(rsp, errors.Wrapf(err, "cannot get observed composite resource from %T", req))
		return rsp, nil
	}

	desired, err := request.GetDesiredComposedResources(req)
	if err != nil {
		response.Fatal(rsp, errors.Wrapf(err, "cannot get desired resources from %T", req))
		return rsp, nil
	}

	f.log.Info("Running function", "tag", req.GetMeta().GetTag(), "total-desired", len(desired))

	data, err := xr.Resource.GetString(in.JSONDataPath)
	if err != nil {
		f.log.Info("cannot get value from JSON data path", "error", err)
		return rsp, nil
	}

	res, err := runJQuery(in.JSONQuery, data)
	if err != nil {
		f.log.Info("Error executing jQuery", "error", err)
		response.Fatal(rsp, errors.Wrapf(err, "unable to run query  %s", xr.Resource.GetKind()))
		return rsp, nil
	}

	if res == "" {
		f.log.Info("Empty results found", "resource", xr.Resource.GetKind())
		return rsp, nil
	}

	if err := xr.Resource.SetString(in.ResponsePath, res); err != nil {
		f.log.Info("cannot set response path", "error", err, "resource", xr.Resource.GetKind())
		response.Fatal(rsp, errors.Wrapf(err, "cannot set response path of %s", xr.Resource.GetKind()))
		return rsp, nil
	}

	if err := response.SetDesiredCompositeResource(rsp, xr); err != nil {
		response.Warning(rsp, fmt.Errorf("unable to set desired composite resource, error: %s", err)).TargetCompositeAndClaim()
		response.Fatal(rsp, errors.Wrapf(err, "unable to set desired composite resource %s", xr.Resource.GetKind()))
		return rsp, nil
	}

	response.ConditionTrue(rsp, "FunctionSuccess", "Success").TargetCompositeAndClaim()

	return rsp, nil
}

func runJQuery(jqQuery string, rawObj string) (string, error) {
	var obj any
	if err := json.Unmarshal([]byte(rawObj), &obj); err != nil {
		return "", errors.Errorf("cannot unmarshal raw object: %s", rawObj)
	}
	query, err := gojq.Parse(jqQuery)
	if err != nil {
		return "", errors.Errorf("cannot parse jq query: %s", jqQuery)
	}

	queryRes, ok := query.Run(obj).Next()
	if !ok {
		return "", errors.Errorf("unable to run query %s", fmt.Sprint(queryRes))
	}

	err, ok = queryRes.(error)
	if ok {
		return "", errors.Errorf("invalid query %s error %s", jqQuery, err.Error())
	}

	if queryRes == nil {
		return "", nil
	}

	q, ok := queryRes.(string)
	if !ok {
		return "", errors.Errorf("expected string response, got %T value %v", queryRes, queryRes)
	}
	return q, nil
}
