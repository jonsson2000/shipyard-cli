package server

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/shipyard/shipyard-cli/pkg/types"
)

func (handler) getAllEnvironments(w http.ResponseWriter, r *http.Request) {
	org := r.URL.Query().Get("org")
	envs, ok := store[org]
	if !ok {
		orgNotFound(w)
		return
	}
	resp := types.RespManyEnvs{Data: envs}
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(err.Error()))
	}
}

func (handler) getEnvironmentByID(w http.ResponseWriter, r *http.Request) {
	env := findEnvByID(w, r)
	if env != nil {
		resp := types.Response{Data: *env}
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(err.Error()))
		}
	}
}

func (handler) rebuildEnvironment(w http.ResponseWriter, r *http.Request) {
	_ = findEnvByID(w, r)
}

func findEnvByID(w http.ResponseWriter, r *http.Request) *types.Environment {
	org := r.URL.Query().Get("org")
	envs, ok := store[org]
	if !ok {
		orgNotFound(w)
		return nil
	}
	id := r.PathValue("id")
	for i := range envs {
		if envs[i].ID == id {
			return &envs[i]
		}
	}
	envNotFound(w)
	return nil
}

func orgNotFound(w http.ResponseWriter) {
	w.WriteHeader(http.StatusBadRequest)
	fmt.Fprintf(w, "user org not found")
}

func envNotFound(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNotFound)
	fmt.Fprintf(w, "environment not found")
}

func (handler) validateCompose(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "failed to read request body")
		return
	}
	defer r.Body.Close()

	content := string(body)
	var response types.ValidationResponse

	if len(content) == 0 {
		response = types.ValidationResponse{
			Valid:   false,
			Message: "Empty compose file",
			Errors: []types.ValidationError{
				{
					Message: "Compose file is empty",
					Code:    "EMPTY_FILE",
				},
			},
		}
	} else if !containsServices(content) {
		response = types.ValidationResponse{
			Valid:   false,
			Message: "No services found",
			Errors: []types.ValidationError{
				{
					Message: "No services section found in compose file",
					Code:    "NO_SERVICES",
				},
			},
		}
	} else if containsInvalidSyntax(content) {
		response = validationFixtures["invalid"]
	} else {
		response = validationFixtures["valid"]
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "failed to encode response")
	}
}

func containsServices(content string) bool {
	return len(content) > 0 && (strings.Contains(content, "services:") || strings.Contains(content, "services "))
}

func containsInvalidSyntax(content string) bool {
	return strings.Contains(content, "80:80:80") ||
		strings.Contains(content, "database\n    image:") ||
		strings.Contains(content, "web\n    image:")
}
