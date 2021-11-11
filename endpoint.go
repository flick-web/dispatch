package dispatch

import (
	"context"
	"encoding/json"
	"log"
)

// An Endpoint represents an API procedure.
type Endpoint struct {
	pathMatcher *APIPath

	// Path is the API path string that will be exposed as an API endpoint. Must
	// be unique.
	//
	// The format of Path is METHOD/path/{pathvar}. Any path variables in curly
	// brace notation will be parsed during API.Call and passed to Handler as
	// a Context struct value.
	Path string

	// Handler must be a function that receives any single input variable, an
	// input variable of type Context, neither, or both. It can return one
	// output variable of any time, an error, neither, or both in the order
	// (output, error).
	//
	// The input value for Handler, if not Context, will automatically be
	// unmarshalled from the input to API.Call.
	Handler interface{}

	// PreRequestHook is a middleware hook that runs before the handler. If the
	// hook returns an error, that error will be returned and the handler will
	// not be called.
	PreRequestHooks []MiddlewareHook
}

// EndpointInput represents the input to an endpoint call. These inputs can be
// modified by middleware hooks.
type EndpointInput struct {
	Method string
	Path   string
	Ctx    context.Context
	Input  json.RawMessage
}

// MiddlewareHook is a function type that is called for each request.
//
// These hooks are functions which will be called before the endpoint handler is
// called, and can choose to modify the method, path, context, or input of the
// endpoint before it is passed along. If the hook returns an error, execution
// of the endpoint will halt. This is useful for things like authentication
// checks, which must happen before the function is triggered, and must be able
// to return early if a call isn't authorized.
type MiddlewareHook func(*EndpointInput) (*EndpointInput, error)

// AddEndpoint registers an endpoint with this API. It also allows adding
// middleware hooks to the endpoint.
func (api *API) AddEndpoint(path string, handler interface{}, hooks ...MiddlewareHook) {
	if api.Endpoints == nil {
		api.Endpoints = make([]*Endpoint, 0)
	}

	endpoint := Endpoint{
		Path:    path,
		Handler: handler,
	}
	// Configure middleware hooks
	if len(hooks) >= 1 {
		endpoint.PreRequestHooks = hooks
	}

	var err error
	endpoint.pathMatcher, err = NewAPIPath(path)
	if err != nil {
		log.Fatal(err)
	}
	api.Endpoints = append(api.Endpoints, &endpoint)
}
