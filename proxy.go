package dispatch

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/events"
)

// HTTPProxy is a handler function suitable for use in http.HandleFunc.
// For example:
//
//	http.HandleFunc("/", api.HTTPProxy)
//	log.Fatal(http.ListenAndServe(":8000", nil))
//
// The provided handler takes care of access control headers, CORS requests,
// JSON marshalling, and error handling.
func (api *API) HTTPProxy(w http.ResponseWriter, r *http.Request) {
	wroteHeader := 200
	wroteStatus := http.StatusText(200)
	startTime := time.Now()
	defer func() {
		fmt.Printf("%v %s%s - %d %s", time.Since(startTime), r.Method, r.URL.Path, wroteHeader, wroteStatus)
	}()
	writeError := func(w http.ResponseWriter, error string, code int) {
		wroteHeader = code
		wroteStatus = http.StatusText(code)
		http.Error(w, error, code)
	}
	w.Header().Set("Access-Control-Allow-Origin", "*")
	// w.Header().Set("Access-Control-Allow-Methods", "PUT, POST, GET, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	if r.Method == "OPTIONS" {
		validMethods := api.GetMethodsForPath(r.URL.Path)
		w.Header().Set("Access-Control-Allow-Methods", strings.Join(validMethods, ", "))
		w.WriteHeader(200)
		return
	}
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	ctx := &Context{Request: r, Writer: w}
	output, err := api.Call(r.Method, r.URL.Path, ctx, data)
	if err != nil {
		if err == ErrorNotFound {
			writeError(w, err.Error(), http.StatusNotFound)
		} else if err == ErrorBadRequest {
			writeError(w, err.Error(), http.StatusBadRequest)
		} else {
			writeError(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	outBytes, err := json.Marshal(output)
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(outBytes)
}

// LambdaProxy is a handler function suitable for use with github.com/aws/aws-lambda-go/lambda.
// For example:
//
//	import "github.com/aws/aws-lambda-go/lambda"
//	func main() {
//		lambda.Start(api.LambdaProxy)
//	}
//
// The provided handler takes care of access control headers, CORS requests,
// JSON marshalling, and error handling.
func (api *API) LambdaProxy(r *events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	response := &events.APIGatewayProxyResponse{
		Headers: make(map[string]string),
	}
	startTime := time.Now()
	defer func() {
		fmt.Printf("%v %s%s - %d %s\n", time.Since(startTime), r.HTTPMethod, r.Path, response.StatusCode, response.Body)
	}()
	writeError := func(err string, code int) {
		response.Body = err
		response.StatusCode = code
	}

	response.Headers["Access-Control-Allow-Origin"] = "*"
	response.Headers["Access-Control-Allow-Headers"] = "Authorization, Content-Type"

	if r.HTTPMethod == "OPTIONS" {
		validMethods := api.GetMethodsForPath(r.Path)
		response.Headers["Access-Control-Allow-Methods"] = strings.Join(validMethods, ", ")
		response.StatusCode = http.StatusOK
		return response, nil
	}

	data := []byte(r.Body)

	ctx := &Context{LambdaRequest: r, LambdaResponse: response}
	output, err := api.Call(r.HTTPMethod, r.Path, ctx, data)
	if err != nil {
		if apiErr, ok := err.(*APIError); ok {
			writeError(apiErr.Error(), apiErr.StatusCode)
		} else if err == ErrorNotFound {
			writeError(err.Error(), http.StatusNotFound)
		} else if err == ErrorBadRequest {
			writeError(err.Error(), http.StatusBadRequest)
		} else {
			writeError(err.Error(), http.StatusInternalServerError)
		}
		return response, nil
	}
	outBytes, err := json.Marshal(output)
	if err != nil {
		writeError(err.Error(), http.StatusInternalServerError)
		return response, nil
	}
	response.Headers["Content-Type"] = "application/json"
	response.Body = string(outBytes)
	response.StatusCode = http.StatusOK
	return response, nil
}
