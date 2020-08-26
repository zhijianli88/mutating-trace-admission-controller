package patch

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/golang/glog"
	apitrace "go.opentelemetry.io/otel/api/trace"
	"k8s.io/api/admission/v1beta1"

	"mutating-trace-admission-controller/pkg/util/trace"
)

const initTraceIDHeaderKey string = "Init-Traceid"

// avoid use char `/` in string
const initTraceIDAnnotationKey string = "trace.kubernetes.io.init"

// avoid use char `/` in string
const traceAnnotationKey string = "trace.kubernetes.io.context"

type patchOperation struct {
	Op    string      `json:"op"`
	Path  string      `json:"path"`
	Value interface{} `json:"value,omitempty"`
}

// InjectPatch inject the trace context into received object
func InjectPatch(r *http.Request, ar *v1beta1.AdmissionReview) (response *v1beta1.AdmissionResponse) {
	glog.V(3).Infof("AdmissionReview for Kind=%v, Namespace=%v Name=%v UID=%v patchOperation=%v UserInfo=%v", ar.Request.Kind, ar.Request.Namespace, ar.Request.Name, ar.Request.UID, ar.Request.Operation, ar.Request.UserInfo)

	fmt.Println("-------------------------------------")
	fmt.Println(r.Header)
	fmt.Println(ar.Request.Operation)
	fmt.Println(ar.Request.Kind.Kind)
	fmt.Println("-------------------------------------")

	// extract span context from request
	spanContext := trace.SpanContextFromRequestHeader(r)
	// get initTraceID from request header
	var initTraceID string = ""
	if len(r.Header[initTraceIDHeaderKey]) != 0 {
		initTraceID = r.Header[initTraceIDHeaderKey][0]
	}
	// if haven't initTraceID in header , copy from span trace id
	if initTraceID == "" && ar.Request.Operation == "CREATE" {
		initTraceID = spanContext.TraceID.String()
	}
	// build the annotations to patch
	patchAnnotations, err := buildPatchAnnotations(initTraceID, spanContext)
	if err != nil {
		return &v1beta1.AdmissionResponse{
			Allowed: true,
		}
	}

	switch ar.Request.Kind.Kind {
	case "Deployment":
		response = mutateDeployment(ar.Request.Object.Raw, patchAnnotations)
	case "ReplicaSet":
		response = mutateReplicaSet(ar.Request.Object.Raw, patchAnnotations)
	case "StatefulSet":
		response = mutateStatefulSet(ar.Request.Object.Raw, patchAnnotations)
	case "Pod":
		response = mutatePod(ar.Request.Object.Raw, patchAnnotations)
	}

	return
}

// buildPatchAnnotations create a annotation with initTraceID and span
func buildPatchAnnotations(initTraceID string, spanContext apitrace.SpanContext) (map[string]string, error) {
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

// createPatch creates a mutation patch for pod resource
func createPatch(annotations, patchAnnotations map[string]string) ([]byte, error) {
	var patch []patchOperation
	patch = append(patch, updateAnnotation(annotations, patchAnnotations)...)
	return json.Marshal(patch)
}

func updateAnnotation(target, added map[string]string) (patch []patchOperation) {
	var (
		patchAdd patchOperation = patchOperation{
			Op:    "add",
			Path:  "/metadata/annotations",
			Value: make(map[string]string, 0),
		}
		patchReplace []patchOperation = make([]patchOperation, 0)
	)

	for key, value := range added {
		if target == nil || target[key] == "" {
			patchAdd.Value.(map[string]string)[key] = value
		} else {
			patchReplace = append(patchReplace, patchOperation{
				Op:    "replace",
				Path:  "/metadata/annotations/" + key,
				Value: value,
			})
		}
	}

	if len(patchAdd.Value.(map[string]string)) != 0 {
		patch = append(patch, patchAdd)
	}
	if len(patchReplace) != 0 {
		patch = append(patch, patchReplace...)
	}

	return patch
}
