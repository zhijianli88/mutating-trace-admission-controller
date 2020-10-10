package config

import (
	"reflect"
	"testing"
)

func TestParseConfig(t *testing.T) {
	cases := []struct {
		path          string
		expected      Config
		expectedError bool
	}{
		{
			path:          "",
			expectedError: true,
		},
		{
			path: "config_test.yaml",
			expected: Config{
				Certificate: Certificate{
					CertPath: "/home/certs/cert.pem",
					KeyPath:  "/home/certs/key.pem",
				},
				Trace: Trace{
					SampleRate: 1.0,
				},
			},
		},
		{
			path: "config_test_empty.yaml",
			expected: Config{
				Certificate: Certificate{
					CertPath: "/etc/webhook/certs/cert.pem",
					KeyPath:  "/etc/webhook/certs/key.pem",
				},
			},
		},
		{
			path:          "config_test_error.yaml",
			expectedError: true,
		},
	}

	for _, c := range cases {
		t.Run(c.path, func(t *testing.T) {
			got, err := ParseConfig(c.path)
			if (err != nil) != c.expectedError {
				t.Errorf("got unexpected error: %v", err)
			}
			if !reflect.DeepEqual(c.expected, got) {
				t.Errorf("expected: %v, got: %v", c.expected, got)
			}
		})
	}
}
