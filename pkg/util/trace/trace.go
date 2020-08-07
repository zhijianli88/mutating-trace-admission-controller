package trace

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"

	"go.opencensus.io/plugin/ochttp/propagation/b3"
	"go.opencensus.io/trace"
	"go.opencensus.io/trace/propagation"
)

var defaultFormat propagation.HTTPFormat = &b3.HTTPFormat{}

// SpanContextFromRequestHeader get span context from http request header
func SpanContextFromRequestHeader(req *http.Request) (trace.SpanContext, error) {
	spanContext, ok := defaultFormat.SpanContextFromRequest(req)
	if !ok {
		return trace.SpanContext{}, fmt.Errorf("")
	}
	return spanContext, nil
}

// GenerateEncodedSpanContext takes a SpanContext and returns a serialized string
func GenerateEncodedSpanContext() string {
	_, span := trace.StartSpan(context.Background(), "")
	// should not be exported, purpose of this span is to retrieve OC compliant SpanContext
	spanContext := span.SpanContext()
	return EncodeSpanContext(spanContext)
}

// EncodeSpanContext encode span context to string
func EncodeSpanContext(spanContext trace.SpanContext) string {
	rawContextBytes := propagation.Binary(spanContext)
	encodedSpanContext := base64.StdEncoding.EncodeToString(rawContextBytes)
	return encodedSpanContext
}

// DecodeSpanContext encode span context from string
func DecodeSpanContext(encodedSpanContext string) (trace.SpanContext, error) {
	rawContextBytes := make([]byte, base64.StdEncoding.DecodedLen(len(encodedSpanContext)))
	l, err := base64.StdEncoding.Decode(rawContextBytes, []byte(encodedSpanContext))
	if err != nil {
		return trace.SpanContext{}, err
	}
	rawContextBytes = rawContextBytes[:l]
	spanContext, ok := propagation.FromBinary(rawContextBytes)
	if !ok {
		return trace.SpanContext{}, fmt.Errorf("has an unsupported version ID or contains no TraceID")
	}
	return spanContext, nil
}
