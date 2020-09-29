package trace

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"net/http"
	"reflect"

	apitrace "go.opentelemetry.io/otel/api/trace"
	"go.opentelemetry.io/otel/propagators"
)

var propagator propagators.TraceContext

// SpanContextFromRequestHeader get span context from http request header
func SpanContextFromRequestHeader(req *http.Request) apitrace.SpanContext {
	ctx := propagator.Extract(req.Context(), req.Header)
	return apitrace.RemoteSpanContextFromContext(ctx)
}

// EncodedSpanContext encode span to string
func EncodedSpanContext(spanContext apitrace.SpanContext) (string, error) {
	if reflect.DeepEqual(spanContext, apitrace.SpanContext{}) {
		return "", fmt.Errorf("span context is nil")
	}
	// encode to byte
	buffer := new(bytes.Buffer)
	err := binary.Write(buffer, binary.LittleEndian, spanContext)
	if err != nil {
		return "", err
	}
	// encode to string
	return base64.StdEncoding.EncodeToString(buffer.Bytes()), nil
}

// DecodeSpanContext decode encodedSpanContext to spanContext
func DecodeSpanContext(encodedSpanContext string) (apitrace.SpanContext, error) {
	// decode to byte
	byteList := make([]byte, base64.StdEncoding.DecodedLen(len(encodedSpanContext)))
	l, err := base64.StdEncoding.Decode(byteList, []byte(encodedSpanContext))
	if err != nil {
		return apitrace.EmptySpanContext(), err
	}
	byteList = byteList[:l]
	// decode to span context
	buffer := bytes.NewBuffer(byteList)
	spanContext := apitrace.SpanContext{}
	err = binary.Read(buffer, binary.LittleEndian, &spanContext)
	if err != nil {
		return apitrace.EmptySpanContext(), err
	}
	return spanContext, nil
}
