package commands

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/shipyard/shipyard-cli/pkg/client"
	"github.com/shipyard/shipyard-cli/pkg/types"
)

type MockRequester struct {
	mock.Mock
}

func (m *MockRequester) Do(method, uri, contentType string, body any) ([]byte, error) {
	args := m.Called(method, uri, contentType, body)
	return args.Get(0).([]byte), args.Error(1)
}

func TestNewCreateCmd(t *testing.T) {
	mockRequester := &MockRequester{}
	c := client.New(mockRequester, func() string { return "test-org" })

	cmd := NewCreateCmd(c)

	assert.Equal(t, "create", cmd.Use)
	assert.Equal(t, "Create resources", cmd.Short)
	assert.Len(t, cmd.Commands(), 1)
	assert.Equal(t, "application", cmd.Commands()[0].Use)
}

func TestNewCreateApplicationCmd(t *testing.T) {
	mockRequester := &MockRequester{}
	c := client.New(mockRequester, func() string { return "test-org" })

	cmd := NewCreateApplicationCmd(c)

	assert.Equal(t, "application", cmd.Use)
	assert.Contains(t, cmd.Aliases, "app")
	assert.Equal(t, "Create a new application", cmd.Short)
	assert.True(t, cmd.SilenceUsage)

	flags := cmd.Flags()
	assert.True(t, flags.Lookup("name") != nil)
	assert.True(t, flags.Lookup("repo-name") != nil)
	assert.True(t, flags.Lookup("branch") != nil)
	assert.True(t, flags.Lookup("description") != nil)
	assert.True(t, flags.Lookup("json") != nil)

	requiredFlags := []string{"name", "repo-name"}
	for _, flag := range requiredFlags {
		annotations := flags.Lookup(flag).Annotations
		assert.Contains(t, annotations, cobra.BashCompOneRequiredFlag)
	}
}

func TestHandleCreateApplicationCmd_Success_TableOutput(t *testing.T) {
	viper.Reset()
	defer viper.Reset()

	viper.Set("name", "test-app")
	viper.Set("repo-name", "test-repo")
	viper.Set("branch", "main")
	viper.Set("description", "Test application")
	viper.Set("json", false)

	mockResponse := types.CreateApplicationResponse{
		Data: types.Application{
			ID: "app-123",
			Attributes: types.ApplicationAttributes{
				Name:        "test-app",
				RepoName:    "test-repo",
				Branch:      "main",
				Description: "Test application",
				Status:      "created",
				CreatedAt:   "2023-01-01T00:00:00Z",
			},
		},
	}

	responseBytes, _ := json.Marshal(mockResponse)

	mockRequester := &MockRequester{}
	mockRequester.On("Do", http.MethodPost, mock.AnythingOfType("string"), "application/json", mock.AnythingOfType("map[string]interface {}")).Return(responseBytes, nil)

	c := client.New(mockRequester, func() string { return "test-org" })

	err := handleCreateApplicationCmd(c)

	assert.NoError(t, err)
	mockRequester.AssertExpectations(t)

	expectedBody := map[string]any{
		"name":        "test-app",
		"repo_name":   "test-repo",
		"branch":      "main",
		"description": "Test application",
	}
	mockRequester.AssertCalled(t, "Do", http.MethodPost, mock.AnythingOfType("string"), "application/json", expectedBody)
}

func TestHandleCreateApplicationCmd_Success_JSONOutput(t *testing.T) {
	viper.Reset()
	defer viper.Reset()

	viper.Set("name", "test-app")
	viper.Set("repo-name", "test-repo")
	viper.Set("branch", "main")
	viper.Set("json", true)

	mockResponse := types.CreateApplicationResponse{
		Data: types.Application{
			ID: "app-123",
			Attributes: types.ApplicationAttributes{
				Name:     "test-app",
				RepoName: "test-repo",
				Branch:   "main",
				Status:   "created",
			},
		},
	}

	responseBytes, _ := json.Marshal(mockResponse)

	mockRequester := &MockRequester{}
	mockRequester.On("Do", http.MethodPost, mock.AnythingOfType("string"), "application/json", mock.AnythingOfType("map[string]interface {}")).Return(responseBytes, nil)

	c := client.New(mockRequester, func() string { return "test-org" })

	var buf bytes.Buffer
	originalStdout := &buf

	err := handleCreateApplicationCmd(c)

	assert.NoError(t, err)
	mockRequester.AssertExpectations(t)
}

func TestHandleCreateApplicationCmd_APIError(t *testing.T) {
	viper.Reset()
	defer viper.Reset()

	viper.Set("name", "test-app")
	viper.Set("repo-name", "test-repo")
	viper.Set("branch", "main")

	mockRequester := &MockRequester{}
	mockRequester.On("Do", http.MethodPost, mock.AnythingOfType("string"), "application/json", mock.AnythingOfType("map[string]interface {}")).Return([]byte{}, errors.New("API error"))

	c := client.New(mockRequester, func() string { return "test-org" })

	err := handleCreateApplicationCmd(c)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "API error")
	mockRequester.AssertExpectations(t)
}

func TestHandleCreateApplicationCmd_InvalidJSONResponse(t *testing.T) {
	viper.Reset()
	defer viper.Reset()

	viper.Set("name", "test-app")
	viper.Set("repo-name", "test-repo")
	viper.Set("branch", "main")
	viper.Set("json", false)

	invalidJSON := []byte(`{"invalid": json}`)

	mockRequester := &MockRequester{}
	mockRequester.On("Do", http.MethodPost, mock.AnythingOfType("string"), "application/json", mock.AnythingOfType("map[string]interface {}")).Return(invalidJSON, nil)

	c := client.New(mockRequester, func() string { return "test-org" })

	err := handleCreateApplicationCmd(c)

	assert.NoError(t, err)
	mockRequester.AssertExpectations(t)
}

func TestHandleCreateApplicationCmd_WithOrgParam(t *testing.T) {
	viper.Reset()
	defer viper.Reset()

	viper.Set("name", "test-app")
	viper.Set("repo-name", "test-repo")
	viper.Set("branch", "main")

	mockResponse := types.CreateApplicationResponse{
		Data: types.Application{
			ID: "app-123",
			Attributes: types.ApplicationAttributes{
				Name:     "test-app",
				RepoName: "test-repo",
				Branch:   "main",
				Status:   "created",
			},
		},
	}

	responseBytes, _ := json.Marshal(mockResponse)

	mockRequester := &MockRequester{}
	mockRequester.On("Do", http.MethodPost, mock.MatchedBy(func(uri string) bool {
		return uri != "" && uri != "https://shipyard.build/api/v1/application"
	}), "application/json", mock.AnythingOfType("map[string]interface {}")).Return(responseBytes, nil)

	c := client.New(mockRequester, func() string { return "custom-org" })

	err := handleCreateApplicationCmd(c)

	assert.NoError(t, err)
	mockRequester.AssertExpectations(t)
}

func TestHandleCreateApplicationCmd_EmptyDescription(t *testing.T) {
	viper.Reset()
	defer viper.Reset()

	viper.Set("name", "test-app")
	viper.Set("repo-name", "test-repo")
	viper.Set("branch", "main")
	viper.Set("description", "")

	mockResponse := types.CreateApplicationResponse{
		Data: types.Application{
			ID: "app-123",
			Attributes: types.ApplicationAttributes{
				Name:        "test-app",
				RepoName:    "test-repo",
				Branch:      "main",
				Description: "",
				Status:      "created",
			},
		},
	}

	responseBytes, _ := json.Marshal(mockResponse)

	mockRequester := &MockRequester{}
	mockRequester.On("Do", http.MethodPost, mock.AnythingOfType("string"), "application/json", mock.AnythingOfType("map[string]interface {}")).Return(responseBytes, nil)

	c := client.New(mockRequester, func() string { return "test-org" })

	err := handleCreateApplicationCmd(c)

	assert.NoError(t, err)
	mockRequester.AssertExpectations(t)

	expectedBody := map[string]any{
		"name":        "test-app",
		"repo_name":   "test-repo",
		"branch":      "main",
		"description": "",
	}
	mockRequester.AssertCalled(t, "Do", http.MethodPost, mock.AnythingOfType("string"), "application/json", expectedBody)
}
