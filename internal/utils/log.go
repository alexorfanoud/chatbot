package utils

import (
	"context"
	"fmt"
	"log"
)

func Log(ctx context.Context, msg string) {
	tid := GetTraceIDFromContext(ctx)
	prefix := fmt.Sprintf("[traceID=%s]", tid)
	log.Println(prefix, msg)
}
