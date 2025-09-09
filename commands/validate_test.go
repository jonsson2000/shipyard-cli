package commands

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/shipyard/shipyard-cli/pkg/client"
	"github.com/shipyard/shipyard-cli/pkg/requests"
	"github.com/shipyard/shipyard-cli/pkg/types"
)

func TestValidateCmd(t *testing.T) {

	tests := []struct {
		name           string
		args           []string
		fileContent    string
		serverResponse string
		serverStatus   int
		wantOutput     string
		wantErr        bool
	}{
		{
			name:        "valid compose file",
			args:        []string{"test-compose.yaml"},
			fileContent: "version: '3.8'\nservices:\n  web:\n    image: nginx",
			serverResponse: `{
				"valid": true,
				"message": "Docker Compose file is valid",
				"services": ["web"],
				"warnings": []
			}`,
			serverStatus: http.StatusOK,
			wantOutput:   "✅ Docker Compose file is valid\nServices found: [web]\n",
			wantErr:      false,
		},
		{
			name:        "invalid compose file with errors",
			args:        []string{"test-compose.yaml"},
			fileContent: "version: '3.8'\nservices:\n  web\n    image: nginx",
			serverResponse: `{
				"valid": false,
				"message": "Docker Compose file validation failed",
				"errors": [
					{
						"line": 3,
						"column": 6,
						"message": "Invalid YAML syntax: expected ':' but found newline",
						"code": "YAML_SYNTAX_ERROR"
					}
				],
				"warnings": []
			}`,
			serverStatus: http.StatusOK,
			wantOutput:   "❌ Docker Compose file validation failed\n\nErrors:\n  Line 3, Column 6: Invalid YAML syntax: expected ':' but found newline (YAML_SYNTAX_ERROR)\n",
			wantErr:      false,
		},
		{
			name:        "valid file with warnings",
			args:        []string{"test-compose.yaml"},
			fileContent: "version: '3.8'\nservices:\n  web:\n    image: nginx\n  db:\n    image: postgres",
			serverResponse: `{
				"valid": true,
				"message": "Docker Compose file is valid",
				"services": ["web", "db"],
				"warnings": [
					{
						"service": "db",
						"message": "No restart policy specified, defaulting to 'no'"
					}
				]
			}`,
			serverStatus: http.StatusOK,
			wantOutput:   "✅ Docker Compose file is valid\nServices found: [web db]\n\nWarnings:\n  Service 'db': No restart policy specified, defaulting to 'no'\n",
			wantErr:      false,
		},
		{
			name:        "json output flag",
			args:        []string{"test-compose.yaml", "--json"},
			fileContent: "version: '3.8'\nservices:\n  web:\n    image: nginx",
			serverResponse: `{
				"valid": true,
				"message": "Docker Compose file is valid",
				"services": ["web"],
				"warnings": []
			}`,
			serverStatus: http.StatusOK,
			wantOutput: `{
				"valid": true,
				"message": "Docker Compose file is valid",
				"services": ["web"],
				"warnings": []
			}`,
			wantErr: false,
		},
		{
			name:         "file not found",
			args:         []string{"nonexistent.yaml"},
			wantOutput:   "",
			wantErr:      true,
		},
		{
			name:        "server error",
			args:        []string{"test-compose.yaml"},
			fileContent: "version: '3.8'\nservices:\n  web:\n    image: nginx",
			serverResponse: `{
				"errors": [
					{
						"status": 500,
						"title": "Internal server error"
					}
				]
			}`,
			serverStatus: http.StatusInternalServerError,
			wantOutput:   "",
			wantErr:      true,
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {

			tmpDir := t.TempDir()
			
			var testFile string
			if test.fileContent != "" {
				testFile = filepath.Join(tmpDir, "test-compose.yaml")
				err := os.WriteFile(testFile, []byte(test.fileContent), 0644)
				if err != nil {
					t.Fatal(err)
				}
				test.args[0] = testFile
			}

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPost {
					t.Errorf("Expected POST request, got %s", r.Method)
					return
				}
				
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(test.serverStatus)
				w.Write([]byte(test.serverResponse))
			}))
			defer server.Close()

			viper.Set("api_url", server.URL)
			viper.Set("api_token", "test-token")
			c := client.New(requests.New(), func() string { return "test-token" })

			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w
			
			cmd := NewValidateCmd(c)
			cmd.SetArgs(test.args)

			err := cmd.Execute()
			
			w.Close()
			os.Stdout = oldStdout
			
			var buf bytes.Buffer
			buf.ReadFrom(r)
			got := buf.String()

			if (err != nil) != test.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, test.wantErr)
				return
			}

			if !test.wantErr {
				if strings.Contains(strings.Join(test.args, " "), "--json") {
					got = strings.TrimSpace(got)
					want := strings.TrimSpace(test.wantOutput)
					if !strings.Contains(got, `"valid": true`) && strings.Contains(want, `"valid": true`) {
						t.Errorf("JSON output mismatch:\ngot: %s\nwant: %s", got, want)
					}
				} else {
					if got != test.wantOutput {
						t.Errorf("Output mismatch:\ngot: %q\nwant: %q", got, test.wantOutput)
					}
				}
			}
		})
	}
}

func TestDisplayValidationResults(t *testing.T) {

	tests := []struct {
		name       string
		validation *types.ValidationResponse
	}{
		{
			name: "valid with services",
			validation: &types.ValidationResponse{
				Valid:    true,
				Message:  "Docker Compose file is valid",
				Services: []string{"web", "database"},
			},
		},
		{
			name: "valid without services",
			validation: &types.ValidationResponse{
				Valid:   true,
				Message: "Docker Compose file is valid",
			},
		},
		{
			name: "invalid with line-based error",
			validation: &types.ValidationResponse{
				Valid:   false,
				Message: "Validation failed",
				Errors: []types.ValidationError{
					{
						Line:    5,
						Column:  10,
						Message: "Syntax error",
						Code:    "SYNTAX_ERROR",
					},
				},
			},
		},
		{
			name: "invalid with service-based error",
			validation: &types.ValidationResponse{
				Valid:   false,
				Message: "Validation failed",
				Errors: []types.ValidationError{
					{
						Service: "web",
						Message: "Invalid port mapping",
						Code:    "PORT_ERROR",
					},
				},
			},
		},
		{
			name: "invalid with generic error",
			validation: &types.ValidationResponse{
				Valid:   false,
				Message: "Validation failed",
				Errors: []types.ValidationError{
					{
						Message: "Generic error",
					},
				},
			},
		},
		{
			name: "with warnings",
			validation: &types.ValidationResponse{
				Valid:   true,
				Message: "Valid with warnings",
				Warnings: []types.ValidationWarning{
					{
						Service: "db",
						Message: "No restart policy",
					},
					{
						Message: "Generic warning",
					},
				},
			},
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {

			var buf bytes.Buffer
			
			cmd := &cobra.Command{}
			cmd.SetOut(&buf)
			
			err := displayValidationResults(test.validation)
			if err != nil {
				t.Errorf("displayValidationResults() error = %v", err)
			}
			
		})
	}
}

func TestHandleValidateCmd_FileErrors(t *testing.T) {

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("Server should not be called for file errors")
	}))
	defer server.Close()

	viper.Set("api_url", server.URL)
	viper.Set("api_token", "test-token")
	c := client.New(requests.New(), func() string { return "test-token" })

	tests := []struct {
		name     string
		filename string
		wantErr  bool
	}{
		{
			name:     "nonexistent file",
			filename: "/nonexistent/file.yaml",
			wantErr:  true,
		},
		{
			name:     "empty filename",
			filename: "",
			wantErr:  true,
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			
			err := handleValidateCmd(c, test.filename)
			if (err != nil) != test.wantErr {
				t.Errorf("handleValidateCmd() error = %v, wantErr %v", err, test.wantErr)
			}
		})
	}
}
