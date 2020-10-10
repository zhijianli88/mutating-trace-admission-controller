package response

import (
	"reflect"
	"testing"

	apitrace "go.opentelemetry.io/otel/api/trace"
)

func TestBuildAnnotations(t *testing.T) {
	cases := []struct {
		name          string
		initTraceID   string
		spanContext   apitrace.SpanContext
		expected      map[string]string
		expectedError bool
	}{
		{
			name:        "both",
			initTraceID: "1234",
			spanContext: apitrace.SpanContext{
				TraceID:    [16]byte{1, 2, 3},
				SpanID:     [8]byte{4, 5},
				TraceFlags: 1,
			},
			expected: map[string]string{
				initialTraceIDAnnotationKey: "1234",
				spanContextAnnotationKey:    "AQIDAAAAAAAAAAAAAAAAAAQFAAAAAAAAAQ==",
			},
		},
		{
			name:          "only init trace id",
			initTraceID:   "1234",
			expectedError: true,
		},
		{
			name: "only span context",
			spanContext: apitrace.SpanContext{
				TraceID:    [16]byte{1, 2, 3},
				SpanID:     [8]byte{4, 5},
				TraceFlags: 1,
			},
			expected: map[string]string{
				spanContextAnnotationKey: "AQIDAAAAAAAAAAAAAAAAAAQFAAAAAAAAAQ==",
			},
		},
		{
			name:          "empty",
			expectedError: true,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := buildAnnotations(c.initTraceID, c.spanContext)
			if (err != nil) != c.expectedError {
				t.Errorf("got unexpected error: %v", err)
			}
			if !reflect.DeepEqual(c.expected, got) {
				t.Errorf("expected: %v, got: %v", c.expected, got)
			}
		})
	}
}
