# github.com/flick-web/dispatch

Dispatch is a not-too-complicated framework meant for creating super quick and easy JSON APIs. Error handling, CORS, JSON parsing, and more are all handled out of the box.

## Basic Usage

This program creates and serves an API with a single endpoint. The endpoint simply returns a string composed with the `{name}` path variable.

```go
package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/flick-web/dispatch"
)

func rootHandler(ctx *dispatch.Context) string {
	return fmt.Sprintf("Hello, %s!", ctx.PathVars["name"])
}

func main() {
	api := &dispatch.API{}
	api.AddEndpoint("GET/{name}", rootHandler)
	http.HandleFunc("/", api.GetHandler())
	log.Fatal(http.ListenAndServe(":8000", nil))
}
```

For a real-world example, check out [pjournal](https://github.com/olafal0/pjournal), a fully-fledged personal journaling web app.

## API Paths

Paths are expressed as a simple string, in the form:

`METHOD/path/{pathvar}`

Any path variables in curly braces will be automatically parsed and provided to handler functions in the `dispatch.Context.PathVars` map. Any path elements not in curly braces are treated as literals, and must be matched for the handler to be called.

## API Endpoints

Endpoints return JSON when used, but the handler functions themselves can accept and return any time, with certain restrictions.

A handler function's **input** signature can be any of these four types:

- `(none)`
- `(*dispatch.Context)`
- `(<AnyType>)`
- `(<AnyType>, *dispatch.Context)` (order does not matter)

If the handler accepts an input type other than `*dispatch.Context`, it can be anything—a string, a struct, or whatever else. Dispatch will automagically marshal any incoming JSON into your type for you.

A handler function's **output** signature is slightly more restricted:

- `(none)`
- `(error)`
- `(<AnyType>)`
- `(<AnyType>, error)` (order **does** matter)

If your function returns an error, the handler provided by the `api` package will automatically return an HTTP error. `dispatch.ErrorNotFound` and `dispatch.ErrorBadRequest` errors will also be accompianied by correct HTTP status codes. Otherwise, dispatch will simply return status 500 and the text of your error.

## Middleware

The `api.AddEndpoint` method also allows adding middleware hooks. These hooks are functions which will be called before the endpoint handler is called, and can choose to modify the method, path, context, or input of the endpoint before it is passed along. If the hook returns an error, execution of the endpoint will halt. This is useful for things like authentication checks, which must happen before the function is triggered, and must be able to return early if a call isn't authorized.

## Known Issues/Disclaimer

Access control headers allow a hardcoded value of `*` for the origin, and only specific content types.

Dispatch was created for a specific purpose, so there are many parts of the library that are too inflexible for many use cases.
