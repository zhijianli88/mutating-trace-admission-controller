package patch

import (
	"reflect"
	"testing"
)

func TestAnnotationsPatch(t *testing.T) {
	cases := []struct {
		name     string
		new      map[string]string
		old      map[string]string
		expected []PatchOperation
	}{
		{
			name: "old is nil",
			new: map[string]string{
				"k1": "1",
				"k2": "2",
			},
			expected: []PatchOperation{
				{
					Op:   "add",
					Path: "/metadata/annotations",
					Value: map[string]string{
						"k1": "1",
						"k2": "2",
					},
				},
			},
		},
		{
			name: "old is empty",
			old:  map[string]string{},
			new: map[string]string{
				"k1": "1",
				"k2": "2",
			},
			expected: []PatchOperation{
				{
					Op:    "add",
					Path:  "/metadata/annotations/k1",
					Value: "1",
				},
				{
					Op:    "add",
					Path:  "/metadata/annotations/k2",
					Value: "2",
				},
			},
		},
		{
			name: "old have same key with new",
			old: map[string]string{
				"k1": "0",
				"k2": "2",
			},
			new: map[string]string{
				"k1": "1",
				"k2": "2",
			},
			expected: []PatchOperation{
				{
					Op:    "replace",
					Path:  "/metadata/annotations/k1",
					Value: "1",
				},
			},
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := BuildAnnotationsPatch(c.old, c.new)
			if !reflect.DeepEqual(c.expected, got) {
				t.Errorf("expected: %v, got: %v", c.expected, got)
			}
		})
	}
}
