package response

import (
	"fmt"
	"net/http"

	"k8s.io/api/admission/v1beta1"

	"mutating-trace-admission-controller/pkg/util/trace"

	apitrace "go.opentelemetry.io/otel/api/trace"
)

// avoid use char `/` in string
const spanContextAnnotationKey string = "trace.kubernetes.io/context"

// BuildResponse build the response to inject the trace context into received object
func BuildResponse(r *http.Request, ar *v1beta1.AdmissionReview) (response *v1beta1.AdmissionResponse) {
	fmt.Println("-------------------------------------")
	fmt.Println(r.Header)
	fmt.Println(ar.Request.Operation)
	fmt.Println(ar.Request.Kind.Kind)
	fmt.Println("-------------------------------------")

	// extract span context from request
	spanContext := trace.SpanContextFromRequestHeader(r)

	// build the annotations to patch
	patchAnnotations, err := buildAnnotations(spanContext)
	if len(patchAnnotations) == 0 || err != nil {
		return &v1beta1.AdmissionResponse{
			UID:     ar.Request.UID,
			Allowed: true,
		}
	}

	switch ar.Request.Kind.Kind {
	case "Deployment":
		response = buildDeploymentPatch(ar.Request.Object.Raw, patchAnnotations)
	case "DeamonSet":
		response = buildDeamonSetPatch(ar.Request.Object.Raw, patchAnnotations)
	case "StatefulSet":
		response = buildStatefulSetPatch(ar.Request.Object.Raw, patchAnnotations)
	case "ReplicaSet":
		response = buildReplicaSetPatch(ar.Request.Object.Raw, patchAnnotations)
	case "Pod":
		response = buildPodPatch(ar.Request.Object.Raw, patchAnnotations)
	default:
		response = &v1beta1.AdmissionResponse{
			Allowed: true,
		}
	}
	response.UID = ar.Request.UID

	return
}

// buildAnnotations create a annotation with initTraceID and span
func buildAnnotations(spanContext apitrace.SpanContext) (map[string]string, error) {
	encodedSpanContext, err := trace.EncodedSpanContext(spanContext)
	if err != nil {
		return nil, err
	}
	return map[string]string{
		spanContextAnnotationKey: encodedSpanContext,
	}, nil
}
