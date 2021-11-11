package dispatch

import (
	"context"
	"errors"
	"log"
	"strings"
	"testing"
)

func TestEndpoints(t *testing.T) {
	api := API{}
	api.AddEndpoint("GET/test", testEndpointHandler)
	api.AddEndpoint("PUT/test", testEndpointHandler)
	api.AddEndpoint("DELETE/tests", testEndpointHandler)
	api.AddEndpoint("DELETE/test/{id}", testEndpointHandler)
	api.AddEndpoint("GET/apiErrorTest", testAPIErrors)

	methods := api.GetMethodsForPath("/test")
	if (methods[0] != "GET" && methods[0] != "PUT") || (methods[1] != "GET" && methods[1] != "PUT") {
		t.Errorf("Expected GET, PUT, got %s\n", strings.Join(methods, ", "))
	}

	ctx := context.Background()
	_, err := api.Call(ctx, "GET", "/test", []byte("{\"foo\": \"hello\", \"Var2\": 42}"))
	if err != nil {
		t.Error(err)
	}

	_, err = api.Call(ctx, "GET", "/test", []byte("{\"Var1\":"))
	if !strings.Contains(err.Error(), "unexpected") {
		t.Error(err)
	}

	_, err = api.Call(ctx, "POST", "/none", nil)
	if !strings.Contains(err.Error(), "not found") {
		t.Error(err)
	}

	_, err = api.Call(ctx, "GET", "/test", []byte("{\"foo\": \"PANIC\", \"Var2\": 42}"))
	if err.Error() != "PANICKING" {
		t.Error(err)
	}

	_, err = api.Call(ctx, "GET", "/apiErrorTest", nil)
	if err.Error() != "I'm a teapot" {
		t.Error(err)
	}
}

func TestEndpointWithContext(t *testing.T) {
	api := API{}
	api.AddEndpoint("GET/user/{foo}", testPathVarHandler)

	ctx := context.Background()
	result, err := api.Call(ctx, "GET", "/user/abcde", []byte("{}"))
	if result != "abcde" || err != nil {
		t.Error(result)
		t.Error(err)
	}
}

func TestEndpointBadHandler(t *testing.T) {
	api := API{}
	api.AddEndpoint("GET/test", testBadHandler)

	ctx := context.Background()
	_, err := api.Call(ctx, "GET", "/test", []byte("{\"foo\": \"TestAdmin\"}"))
	if err == nil {
		t.Error("Should have failed!")
	}
}

func TestMiddleware(t *testing.T) {
	api := API{}
	api.AddEndpoint("GET/test/{TestVar}", testEndpointHandler, middlewareHook)

	ctx := context.Background()
	_, err := api.Call(ctx, "GET", "/test/TestVar", []byte("{}"))
	if err != nil {
		t.Error(err)
	}

	// Since the TestVar path variable is not "TestVar", the middleware should fail
	_, err = api.Call(ctx, "GET", "/test/none", []byte("{}"))
	if err == nil || err.Error() != "ERROR" {
		t.Error(err)
	}
}

type testInputType struct {
	Var1 string `json:"foo"`
	Var2 int
}

func testEndpointHandler(in testInputType) error {
	if in.Var1 == "PANIC" {
		return errors.New("PANICKING")
	}
	return nil
}

func testBadHandler(in1, in2 testInputType) (interface{}, error) {
	return "OK", nil
}

func testPathVarHandler(ctx context.Context, in1 testInputType) (interface{}, error) {
	pathVars := ContextPathVars(ctx)
	log.Println(pathVars)
	return pathVars["foo"], nil
}

func middlewareHook(input *EndpointInput) (*EndpointInput, error) {
	log.Println(string(input.Input))
	if ContextPathVars(input.Ctx)["TestVar"] != "TestVar" {
		return nil, errors.New("ERROR")
	}
	return input, nil
}

func testAPIErrors() *APIError {
	return NewAPIError(418, "I'm a teapot")
}
