package response

import (
	"fmt"
	"net/http"

	"github.com/golang/glog"
	"k8s.io/api/admission/v1beta1"

	"mutating-trace-admission-controller/pkg/util/trace"

	apitrace "go.opentelemetry.io/otel/api/trace"
)

const initTraceIDHeaderKey string = "Init-Traceid"

// avoid use char `/` in string
const initTraceIDAnnotationKey string = "trace.kubernetes.io.init"

// avoid use char `/` in string
const traceAnnotationKey string = "trace.kubernetes.io.context"

// BuildResponse build the response to inject the trace context into received object
func BuildResponse(r *http.Request, ar *v1beta1.AdmissionReview) (response *v1beta1.AdmissionResponse) {
	glog.V(3).Infof("AdmissionReview for Kind=%v, Namespace=%v Name=%v UID=%v patchOperation=%v UserInfo=%v", ar.Request.Kind, ar.Request.Namespace, ar.Request.Name, ar.Request.UID, ar.Request.Operation, ar.Request.UserInfo)

	fmt.Println("-------------------------------------")
	fmt.Println(r.Header)
	fmt.Println(ar.Request.Operation)
	fmt.Println(ar.Request.Kind.Kind)
	fmt.Println("-------------------------------------")

	// extract span context from request
	var initTraceID string = ""
	spanContext := trace.SpanContextFromRequestHeader(r)

	// only when CREATE we need add initTraceID
	if ar.Request.Operation == v1beta1.Create {
		// get initTraceID from request header
		if len(r.Header[initTraceIDHeaderKey]) != 0 {
			initTraceID = r.Header[initTraceIDHeaderKey][0]
		} else {
			initTraceID = spanContext.TraceID.String()
		}
	}

	// build the annotations to patch
	patchAnnotations, err := buildAnnotations(initTraceID, spanContext)
	if len(patchAnnotations) == 0 || err != nil {
		return &v1beta1.AdmissionResponse{
			UID:     ar.Request.UID,
			Allowed: true,
		}
	}

	switch ar.Request.Kind.Kind {
	case "Deployment":
		response = buildDeploymentPatch(ar.Request.Object.Raw, patchAnnotations)
	case "ReplicaSet":
		response = buildReplicaSetPatch(ar.Request.Object.Raw, patchAnnotations)
	case "StatefulSet":
		response = buildStatefulSetPatch(ar.Request.Object.Raw, patchAnnotations)
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
func buildAnnotations(initTraceID string, spanContext apitrace.SpanContext) (map[string]string, error) {
	encodedSpanContext, err := trace.EncodedSpanContext(spanContext)
	if err != nil {
		return nil, err
	}
	if initTraceID == "" {
		return map[string]string{
			traceAnnotationKey: encodedSpanContext,
		}, nil
	}
	return map[string]string{
		initTraceIDAnnotationKey: initTraceID,
		traceAnnotationKey:       encodedSpanContext,
	}, nil
}
