package dispatch

import (
	"context"
	"encoding/json"
	"log"
	"reflect"
	"runtime/debug"
)

// API is an object that holds all API methods and can dispatch them.
type API struct {
	Endpoints []*Endpoint
}

// MatchEndpoint matches a request to an endpoint, creating a map of path
// variables in the process.
func (api *API) MatchEndpoint(method, path string) (*Endpoint, PathVars) {
	for _, endpt := range api.Endpoints {
		pathVars, match := endpt.pathMatcher.Match(method, path)
		if match {
			return endpt, pathVars
		}
	}
	return nil, nil
}

// GetMethodsForPath returns the list of valid methods for a specified path
// (for use in OPTIONS requests).
func (api *API) GetMethodsForPath(path string) []string {
	methods := make([]string, 0)
	for _, endpt := range api.Endpoints {
		match := endpt.pathMatcher.MatchPath(path)
		if match {
			methods = append(methods, endpt.pathMatcher.Method)
		}
	}
	return methods
}

// Call sends the input to the endpoint and returns the result.
func (api *API) Call(ctx context.Context, method, path string, input json.RawMessage) (out interface{}, err error) {
	// Recover from any panics, and return an internal error in that case
	defer func() {
		if r := recover(); r != nil {
			log.Printf("API.Call panic: %v\n", r)
			debug.PrintStack()
			out = nil
			err = ErrInternal
		}
	}()

	endpoint, pathVars := api.MatchEndpoint(method, path)
	if endpoint == nil {
		return nil, ErrNotFound
	}
	ctx = SetContextPathVars(ctx, pathVars)

	for _, hook := range endpoint.PreRequestHooks {
		originalInput := &EndpointInput{method, path, ctx, input}
		modifiedInput, err := hook(originalInput)
		if err != nil {
			return nil, err
		}
		method = modifiedInput.Method
		path = modifiedInput.Path
		ctx = modifiedInput.Ctx
		input = modifiedInput.Input
	}

	handlerType := reflect.TypeOf(endpoint.Handler)
	if handlerType.Kind() != reflect.Func {
		log.Printf("Bad handler type for %s: %s\n", endpoint.Path, handlerType.Kind())
		return nil, ErrInternal
	}

	// Handler functions can take a custom value type and/or a context input
	if handlerType.NumIn() > 2 {
		log.Printf("Handler %s takes too many args\n", endpoint.Path)
		return nil, ErrInternal
	}
	var inputType reflect.Type
	var takesContext, takesCustom bool
	var ctxIndex, customIndex int
	for i := 0; i < handlerType.NumIn(); i++ {
		inType := handlerType.In(i)
		ctxType := reflect.TypeOf((*context.Context)(nil)).Elem()
		if inType.Implements(ctxType) {
			if takesContext {
				log.Printf("Handler %s takes multiple context inputs", endpoint.Path)
				return nil, ErrInternal
			}
			takesContext = true
			ctxIndex = i
		} else {
			if takesCustom {
				log.Printf("Handler %s takes multiple inputs", endpoint.Path)
				return nil, ErrInternal
			}
			takesCustom = true
			customIndex = i
			inputType = handlerType.In(i)
		}
	}

	handlerValue := reflect.ValueOf(endpoint.Handler)

	var resultValues []reflect.Value
	if takesCustom || takesContext {
		// Can return any interface and/or an error
		inputList := make([]reflect.Value, handlerType.NumIn())
		if takesContext {
			inputList[ctxIndex] = reflect.ValueOf(ctx)
		}
		if takesCustom {
			inputVal := reflect.New(inputType)
			inputInterface := inputVal.Interface()
			err = json.Unmarshal(input, inputInterface)
			if err != nil {
				return nil, err
			}
			directInput := reflect.Indirect(reflect.ValueOf(inputInterface))
			inputList[customIndex] = directInput
		}

		resultValues = handlerValue.Call(inputList)
	} else {
		resultValues = handlerValue.Call(nil)
	}

	switch len(resultValues) {
	case 0:
		return nil, nil

	case 1:
		// Function may return _either_ an error or a value
		retval := resultValues[0].Interface()
		// If nil, it doesn't matter
		if retval == nil {
			return nil, nil
		}
		// Otherwise, check if it can be asserted as an error
		returnErr, ok := retval.(error)
		if ok {
			return nil, returnErr
		}
		// Otherwise, assume it's data
		return retval, nil

	case 2:
		// If a value and error are returned, they must be in the order (out, error)
		out = resultValues[0].Interface()
		if errVal := resultValues[1].Interface(); errVal == nil {
			err = nil
		} else {
			err = errVal.(error)
		}
		return out, err

	default:
		log.Printf("Handler %s returned too many values\n", endpoint.Path)
		return nil, ErrInternal
	}
}
