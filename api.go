package dispatch

import (
	"encoding/json"
	"log"
	"net/http"
	"reflect"
	"runtime/debug"

	"github.com/aws/aws-lambda-go/events"
)

// Context represents data about the endpoint call, such as path variables, the
// calling user, and requests and responses.
type Context struct {
	// Request is the original http request.
	Request *http.Request
	// Writer is the original response writer.
	Writer http.ResponseWriter

	// LambdaRequest is the original request for LambdaProxy requests.
	LambdaRequest *events.APIGatewayProxyRequest
	// LambdaResponse is the constructed response object for LambdaProxy requests.
	LambdaResponse *events.APIGatewayProxyResponse

	// PathVars is the map of path variable names to values.
	PathVars PathVars
}

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
func (api *API) Call(method, path string, ctx *Context, input json.RawMessage) (out interface{}, err error) {
	// Recover from any panics, and return an internal error in that case
	defer func() {
		if r := recover(); r != nil {
			log.Printf("API.Call panic: %v\n", r)
			debug.PrintStack()
			out = nil
			err = ErrInternal
		}
	}()

	if ctx == nil {
		ctx = &Context{}
	}

	endpoint, pathVars := api.MatchEndpoint(method, path)
	if endpoint == nil {
		return nil, ErrNotFound
	}
	ctx.PathVars = pathVars

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
	var takesContext, takesCustom, ctxPointer bool
	var ctxIndex, customIndex int
	for i := 0; i < handlerType.NumIn(); i++ {
		inType := handlerType.In(i)
		if inType == reflect.TypeOf(Context{}) {
			takesContext = true
			ctxIndex = i
		} else if inType == reflect.TypeOf(&Context{}) {
			takesContext = true
			ctxPointer = true
			ctxIndex = i
		} else {
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
			if ctxPointer {
				inputList[ctxIndex] = reflect.ValueOf(ctx)
			} else {
				inputList[ctxIndex] = reflect.ValueOf(*ctx)
			}
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
