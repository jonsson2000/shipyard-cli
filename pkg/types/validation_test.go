package types

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestUnmarshalValidation(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		input   []byte
		want    *ValidationResponse
		wantErr bool
	}{
		{
			name: "valid response with services",
			input: []byte(`{
				"valid": true,
				"message": "Docker Compose file is valid",
				"services": ["web", "database", "redis"],
				"warnings": []
			}`),
			want: &ValidationResponse{
				Valid:    true,
				Message:  "Docker Compose file is valid",
				Services: []string{"web", "database", "redis"},
				Warnings: []ValidationWarning{},
			},
			wantErr: false,
		},
		{
			name: "invalid response with errors and warnings",
			input: []byte(`{
				"valid": false,
				"message": "Docker Compose file validation failed",
				"errors": [
					{
						"line": 12,
						"column": 8,
						"message": "Invalid YAML syntax: expected ':' but found '-'",
						"code": "YAML_SYNTAX_ERROR"
					},
					{
						"service": "web",
						"field": "ports",
						"message": "Port mapping '80:80:80' is invalid",
						"code": "INVALID_PORT_MAPPING"
					}
				],
				"warnings": [
					{
						"service": "database",
						"message": "No restart policy specified, defaulting to 'no'"
					}
				]
			}`),
			want: &ValidationResponse{
				Valid:   false,
				Message: "Docker Compose file validation failed",
				Errors: []ValidationError{
					{
						Line:    12,
						Column:  8,
						Message: "Invalid YAML syntax: expected ':' but found '-'",
						Code:    "YAML_SYNTAX_ERROR",
					},
					{
						Service: "web",
						Field:   "ports",
						Message: "Port mapping '80:80:80' is invalid",
						Code:    "INVALID_PORT_MAPPING",
					},
				},
				Warnings: []ValidationWarning{
					{
						Service: "database",
						Message: "No restart policy specified, defaulting to 'no'",
					},
				},
			},
			wantErr: false,
		},
		{
			name:    "empty response",
			input:   []byte(`{}`),
			want:    &ValidationResponse{},
			wantErr: false,
		},
		{
			name:    "invalid JSON",
			input:   []byte(`{"valid": true, "message": "incomplete`),
			want:    nil,
			wantErr: true,
		},
		{
			name:    "nil input",
			input:   nil,
			want:    nil,
			wantErr: true,
		},
		{
			name: "minimal valid response",
			input: []byte(`{
				"valid": true,
				"message": "Valid"
			}`),
			want: &ValidationResponse{
				Valid:   true,
				Message: "Valid",
			},
			wantErr: false,
		},
		{
			name: "error without line/column info",
			input: []byte(`{
				"valid": false,
				"message": "Validation failed",
				"errors": [
					{
						"message": "Generic error",
						"code": "GENERIC_ERROR"
					}
				]
			}`),
			want: &ValidationResponse{
				Valid:   false,
				Message: "Validation failed",
				Errors: []ValidationError{
					{
						Message: "Generic error",
						Code:    "GENERIC_ERROR",
					},
				},
			},
			wantErr: false,
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			got, err := UnmarshalValidation(test.input)
			if (err != nil) != test.wantErr {
				t.Errorf("UnmarshalValidation() error = %v, wantErr %v", err, test.wantErr)
				return
			}
			if !cmp.Equal(got, test.want) {
				t.Error(cmp.Diff(got, test.want))
			}
		})
	}
}
