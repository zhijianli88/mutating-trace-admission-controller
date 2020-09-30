package server

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"mutating-trace-admission-controller/pkg/response"

	"k8s.io/api/admission/v1beta1"
	admissionregistrationv1beta1 "k8s.io/api/admissionregistration/v1beta1"
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
)

var (
	runtimeScheme = runtime.NewScheme()
	codecs        = serializer.NewCodecFactory(runtimeScheme)
	deserializer  = codecs.UniversalDeserializer()
)

func init() {
	_ = corev1.AddToScheme(runtimeScheme)
	_ = admissionregistrationv1beta1.AddToScheme(runtimeScheme)
	_ = v1.AddToScheme(runtimeScheme)
}

// WebhookServer is ...
type WebhookServer struct {
	Server *http.Server
}

// Serve http handler
func (whsvr *WebhookServer) Serve(w http.ResponseWriter, r *http.Request) {
	// verify the content type is accurate
	if r.Header.Get("Content-Type") != "application/json" {
		http.Error(w, "invalid Content-Type, expect `application/json`", http.StatusUnsupportedMediaType)
		return
	}

	// read request body
	body, err := ioutil.ReadAll(r.Body)
	if err != nil || len(body) == 0 {
		http.Error(w, "empty body", http.StatusBadRequest)
		return
	}

	// decode request body
	ar := v1beta1.AdmissionReview{}
	_, _, err = deserializer.Decode(body, nil, &ar)
	if err != nil {
		http.Error(w, fmt.Sprintf("could not encode response: %v", err), http.StatusBadRequest)
		return
	}

	// build response
	admissionReview := v1beta1.AdmissionReview{}
	admissionReview.Response = response.BuildResponse(r, &ar)

	// marshal respson
	resp, err := json.Marshal(admissionReview)
	if err != nil {
		http.Error(w, fmt.Sprintf("could not encode response: %v", err), http.StatusInternalServerError)
		return
	}

	// write respson
	_, err = w.Write(resp)
	if err != nil {
		http.Error(w, fmt.Sprintf("could not write response: %v", err), http.StatusInternalServerError)
	}

	return
}
