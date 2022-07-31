package deployment

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-kit/kit/endpoint"
	kitHttp "github.com/go-kit/kit/transport/http"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	netHttp "net/http"
	"net/http/httputil"
	"time"
)

// JSONEncodeAPIResponse ...
func JSONEncodeAPIResponse(ctx context.Context, w netHttp.ResponseWriter, resp interface{}) error {
	fmt.Println(fmt.Sprintf("%+v",resp))
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(netHttp.StatusOK)
	return json.NewEncoder(w).Encode(resp)
}

// AddHandlers ...
func AddHandlers() netHttp.Handler {
	svc := NewGitHubAdapter()
	router := mux.NewRouter()
	options := []kitHttp.ServerOption{}
	var getEP endpoint.Endpoint
	{
		getEP = makeGetVMEndpoint(svc)
	}

	router.Methods("POST").Path("/githubapplication/v1/repo/createPullRequest").Handler(kitHttp.NewServer(
		getEP,
		decodeCreatePullRequest,
		JSONEncodeAPIResponse,
		append(options, kitHttp.ServerBefore(
			AddPathParametersToContext(),
		))...,
	))


	return router
}

// AddPathParametersToContext ...
func AddPathParametersToContext() kitHttp.RequestFunc {
	return func(ctx context.Context, r *netHttp.Request) context.Context {
		vars := mux.Vars(r)
		ctx, _ = context.WithDeadline(ctx, time.Now().Add(time.Minute * 10))
		tokenVal := r.Header.Get(token)
		ctx = context.WithValue(ctx, token, tokenVal)
		if vars == nil {
			return ctx
		}
		for key, val := range vars {
			ctx = context.WithValue(ctx, key, val)
		}

		return ctx
	}
}

func decodeCreatePullRequest (ctx context.Context, req *netHttp.Request) (interface{}, error) {
	_, er := httputil.DumpRequest(req, true)
	if er != nil {
		fmt.Println(er.Error())
		return nil, er
	}

	request := CreatePullRequest{}
	decoder := json.NewDecoder(req.Body)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&request)

	if err != nil {
		err = errors.Wrapf(err, "%v=%v msg=%s", "Layer", "Transport", "Failed in decodeCreatePullRequest")
		return nil, err
	}
	return request, nil
}

func makeGetVMEndpoint(svc GitHubInterface) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		fmt.Println("inside makeGetVMEndpoint before")
		response = CreatePullResponse{}
		req, ok := request.(CreatePullRequest)
		if !ok {
			return response, errors.New("Failed while type casting request")
		}
		response, err = svc.CreatePullRequest(ctx, &req)
		fmt.Println("inside makeGetVMEndpoint after")
		fmt.Println(fmt.Sprintf("Response: %+v",response))
		fmt.Println(fmt.Sprintf("Error: %+v",err))
		return
	}
}
