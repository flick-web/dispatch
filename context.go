package dispatch

import (
	"context"

	"github.com/aws/aws-lambda-go/events"
)

type contextPathVars struct{}
type contextLambdaRequest struct{}
type contextLambdaResponse struct{}

func SetContextPathVars(ctx context.Context, pathVars PathVars) context.Context {
	return context.WithValue(ctx, contextPathVars{}, pathVars)
}

func ContextPathVars(ctx context.Context) PathVars {
	pathVars, ok := ctx.Value(contextPathVars{}).(PathVars)
	if ok {
		return pathVars
	}
	return PathVars{}
}

func SetContextLambdaRequest(ctx context.Context, req *events.APIGatewayProxyRequest) context.Context {
	return context.WithValue(ctx, contextLambdaRequest{}, req)
}

func ContextLambdaRequest(ctx context.Context) *events.APIGatewayProxyRequest {
	req, ok := ctx.Value(contextLambdaRequest{}).(*events.APIGatewayProxyRequest)
	if ok {
		return req
	}
	return nil
}

func SetContextLambdaResponse(ctx context.Context, res *events.APIGatewayProxyResponse) context.Context {
	return context.WithValue(ctx, contextLambdaResponse{}, res)
}

func ContextLambdaResponse(ctx context.Context) *events.APIGatewayProxyResponse {
	res, ok := ctx.Value(contextLambdaResponse{}).(*events.APIGatewayProxyResponse)
	if ok {
		return res
	}
	return nil
}
