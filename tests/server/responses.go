package server

import "github.com/shipyard/shipyard-cli/pkg/types"

//nolint:gochecknoglobals // OK for testing.
var store = map[string][]types.Environment{
	"default": {
		{
			Attributes: types.EnvironmentAttributes{
				URL:   "https://dev.example.com",
				Ready: true,
				Projects: []types.Project{
					{PullRequestNumber: 123, RepoName: "Repo1"},
					{PullRequestNumber: 456, RepoName: "Repo2"},
				},
				Services: []types.Service{
					{
						Name: "postgres",
					},
					{
						Name: "web",
					},
				},
			},
			ID: "default-1",
		},
		{
			Attributes: types.EnvironmentAttributes{
				URL:   "https://dev.example.com",
				Ready: true,
				Projects: []types.Project{
					{PullRequestNumber: 123, RepoName: "Repo1"},
					{PullRequestNumber: 456, RepoName: "Repo2"},
				},
			},
			ID: "default-2",
		},
	},
	"pugs": {
		{
			Attributes: types.EnvironmentAttributes{
				URL:   "https://prod.example.com",
				Ready: true,
				Projects: []types.Project{
					{PullRequestNumber: 900, RepoName: "pugs"},
				},
				Services: []types.Service{
					{
						Name: "mysql",
					},
					{
						Name: "nginx",
					},
				},
			},
			ID: "pug-1",
		},
		{
			Attributes: types.EnvironmentAttributes{
				URL:   "https://prod.example.com",
				Ready: true,
				Projects: []types.Project{
					{PullRequestNumber: 901, RepoName: "pugs"},
				},
			},
			ID: "pug-2",
		},
	},
}

var validationFixtures = map[string]types.ValidationResponse{
	"valid": {
		Valid:    true,
		Message:  "Docker Compose file is valid",
		Services: []string{"web", "database", "redis"},
		Warnings: []types.ValidationWarning{},
	},
	"invalid": {
		Valid:   false,
		Message: "Docker Compose file validation failed",
		Errors: []types.ValidationError{
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
		Warnings: []types.ValidationWarning{
			{
				Service: "database",
				Message: "No restart policy specified, defaulting to 'no'",
			},
		},
	},
}
