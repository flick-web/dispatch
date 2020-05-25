package dispatch

import (
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

	_, err := api.Call("GET", "/test", nil, []byte("{\"foo\": \"hello\", \"Var2\": 42}"))
	if err != nil {
		t.Error(err)
	}

	_, err = api.Call("GET", "/test", nil, []byte("{\"Var1\":"))
	if !strings.Contains(err.Error(), "unexpected") {
		t.Error(err)
	}

	_, err = api.Call("POST", "/none", nil, nil)
	if !strings.Contains(err.Error(), "not found") {
		t.Error(err)
	}

	_, err = api.Call("GET", "/test", nil, []byte("{\"foo\": \"PANIC\", \"Var2\": 42}"))
	if err.Error() != "PANICKING" {
		t.Error(err)
	}

	_, err = api.Call("GET", "/apiErrorTest", nil, nil)
	if err.Error() != "I'm a teapot" {
		t.Error(err)
	}
}

func TestEndpointWithContext(t *testing.T) {
	api := API{}
	api.AddEndpoint("GET/user/{foo}", testPathVarHandler)

	result, err := api.Call("GET", "/user/abcde", nil, []byte("{}"))
	if result != "abcde" || err != nil {
		t.Error(err)
	}
}

func TestEndpointBadHandler(t *testing.T) {
	api := API{}
	api.AddEndpoint("GET/test", testBadHandler)

	_, err := api.Call("GET", "/test", nil, []byte("{\"foo\": \"TestAdmin\"}"))
	if err == nil {
		t.Error("Should have failed!")
	}
}

func TestMiddleware(t *testing.T) {
	api := API{}
	api.AddEndpoint("GET/test/{TestVar}", testEndpointHandler, middlewareHook)

	_, err := api.Call("GET", "/test/TestVar", nil, []byte("{}"))
	if err != nil {
		t.Error(err)
	}

	// Since the TestVar path variable is not "TestVar", the middleware should fail
	_, err = api.Call("GET", "/test/none", nil, []byte("{}"))
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

func testPathVarHandler(in1 testInputType, ctx *Context) (interface{}, error) {
	return ctx.PathVars["foo"], nil
}

func middlewareHook(input *EndpointInput) (*EndpointInput, error) {
	log.Println(string(input.Input))
	if input.Ctx.PathVars["TestVar"] != "TestVar" {
		return nil, errors.New("ERROR")
	}
	return input, nil
}

func testAPIErrors() *APIError {
	return NewAPIError(418, "I'm a teapot")
}
