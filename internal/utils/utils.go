package utils

import (
	"context"
	"strings"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel/trace"
)

func GetTraceIDFromContext(ctx context.Context) string {
	return trace.SpanContextFromContext(ctx).TraceID().String()
}

func ReplaceParamsValuesFromUrl(c *gin.Context) string {
	originalURL := c.Request.URL

	// Get the path parameters from the context
	params := c.Params

	// Replace path parameters with their names in the request URL
	replacedURL := originalURL.Path
	for _, param := range params {
		replacedURL = strings.Replace(replacedURL, param.Value, ":"+param.Key, 1)
	}

	return replacedURL

}

func FilterArray[T any](arr []T, predicate func(T) bool) (ret []T) {
	for _, elem := range arr {
		if predicate(elem) {
			ret = append(ret, elem)
		}
	}

	return
}
