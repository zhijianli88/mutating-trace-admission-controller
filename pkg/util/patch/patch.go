package patch

import (
	"encoding/json"
	"fmt"
	"mutating-trace-admission-controller/pkg/util/print"
	"mutating-trace-admission-controller/pkg/util/trace"
	"net/http"

	"github.com/golang/glog"
	"k8s.io/api/admission/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// TraceAnnotationKey is the annotation key the context should be injected at
const TraceAnnotationKey string = "trace.kubernetes.io/context"

type patchOperation struct {
	Op    string      `json:"op"`
	Path  string      `json:"path"`
	Value interface{} `json:"value,omitempty"`
}

// InjectPatch inject the trace context into received object
func InjectPatch(r *http.Request, ar *v1beta1.AdmissionReview) (response *v1beta1.AdmissionResponse) {
	glog.V(3).Infof("AdmissionReview for Kind=%v, Namespace=%v Name=%v UID=%v patchOperation=%v UserInfo=%v",
		ar.Request.Kind, ar.Request.Namespace, ar.Request.Name, ar.Request.UID, ar.Request.Operation, ar.Request.UserInfo)

	spanContext, err := trace.SpanContextFromRequestHeader(r)
	fmt.Println("-----------------------------")
	print.Request(r)
	fmt.Printf("%+v\n", spanContext)
	fmt.Println("-----------------------------")
	if err != nil {
		return &v1beta1.AdmissionResponse{
			Allowed: true,
		}
	}
	encodedSpanContext := trace.EncodeSpanContext(spanContext)

	switch ar.Request.Kind.Kind {
	case "Deployment":
		response = mutateDeployment(encodedSpanContext, ar.Request.Object.Raw)
	case "ReplicaSet":
		response = mutateReplicaSet(encodedSpanContext, ar.Request.Object.Raw)
	case "StatefulSet":
		response = mutateStatefulSet(encodedSpanContext, ar.Request.Object.Raw)
	case "Pod":
		response = mutatePod(encodedSpanContext, ar.Request.Object.Raw)
	}
	return
}

// mutationRequred checks whether we need to inject trace context into received pod
func mutationRequired(metadata *metav1.ObjectMeta) bool {
	// if already present, do not overwrite existing spanContext annotation
	if _, ok := metadata.Annotations[TraceAnnotationKey]; ok {
		glog.V(3).Infof("skipping mutation, spanContext annotation already exists")
		return false
	}

	return true
}

// createPatch creates a mutation patch for pod resource
func createPatch(annotations map[string]string, encodedSpanContext string) ([]byte, error) {
	var patch []patchOperation

	newAnnotations := map[string]string{TraceAnnotationKey: encodedSpanContext}
	patch = append(patch, updateAnnotation(annotations, newAnnotations)...)

	return json.Marshal(patch)
}

func updateAnnotation(target, added map[string]string) (patch []patchOperation) {
	for key, value := range added {
		if target == nil || target[key] == "" {
			target = map[string]string{}
			patch = append(patch, patchOperation{
				Op:   "add",
				Path: "/metadata/annotations",
				Value: map[string]string{
					key: value,
				},
			})
		} else {
			patch = append(patch, patchOperation{
				Op:    "replace",
				Path:  "/metadata/annotations/" + key,
				Value: value,
			})
		}
	}
	return patch
}
