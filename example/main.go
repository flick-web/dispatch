package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/flick-web/dispatch"
)

func rootHandler(ctx context.Context) string {
	return fmt.Sprintf("Hello, %s!", dispatch.ContextPathVars(ctx)["name"])
}

func main() {
	api := &dispatch.API{}
	api.AddEndpoint("GET/{name}", rootHandler)
	http.HandleFunc("/", api.HTTPProxy)
	log.Fatal(http.ListenAndServe(":8000", nil))
}
