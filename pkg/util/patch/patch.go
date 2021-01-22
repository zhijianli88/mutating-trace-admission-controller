package patch

import "encoding/json"

// PatchOperation ...
type PatchOperation struct {
	Op    string      `json:"op"`
	Path  string      `json:"path"`
	Value interface{} `json:"value,omitempty"`
}

// BuildAnnotationsPatch create patch for annotations
func BuildAnnotationsPatch(old, new map[string]string) (patch []PatchOperation) {
	for key, value := range new {
		if old == nil || old[key] == "" {
			patch = append(patch, PatchOperation{
				Op:   "add",
				Path: "/metadata/annotations",
				Value: map[string]string{
					key: value,
				},
			})
		} else if old[key] != value {
			patch = append(patch, PatchOperation{
				Op:    "replace",
				Path:  "/metadata/annotations/" + key,
				Value: value,
			})
		}
	}

	return
}

// EncodePatch encode patch by json
func EncodePatch(patch []PatchOperation) ([]byte, error) {
	return json.Marshal(patch)
}
